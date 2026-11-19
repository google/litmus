// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/logging"
)

// startProxy runs handleRequest in front of upstream. It returns the URL of
// the proxy and a function that waits for the next request log entry, which is
// recorded instead of being sent to Cloud Logging.
func startProxy(t *testing.T, upstream http.HandlerFunc) (proxyURL string, nextLogEntry func() requestLog) {
	t.Helper()

	entries := make(chan requestLog, 1024)
	realWriteLogEntry := writeLogEntry
	writeLogEntry = func(_ context.Context, entry logging.Entry) error {
		entries <- entry.Payload.(requestLog)
		return nil
	}
	// Runs after the servers below have shut down and no handler is left.
	t.Cleanup(func() { writeLogEntry = realWriteLogEntry })

	backend := httptest.NewServer(upstream)
	t.Cleanup(backend.Close)
	upstreamURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, proxy, upstreamURL)
	}))
	t.Cleanup(frontend.Close)

	return frontend.URL, func() requestLog {
		t.Helper()
		select {
		case entry := <-entries:
			return entry
		case <-time.After(30 * time.Second):
			t.Fatal("timed out waiting for the request log entry")
			return requestLog{}
		}
	}
}

// jsonString returns a JSON document of exactly size bytes: one long string.
func jsonString(size int) []byte {
	return []byte(`"` + strings.Repeat("a", size-2) + `"`)
}

// zeros is an endless stream of zero bytes.
type zeros struct{}

func (zeros) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

// drain reads r to the end and returns its length and SHA-256 hash, without
// holding on to the data.
func drain(r io.Reader) (int64, [sha256.Size]byte) {
	h := sha256.New()
	n, _ := io.Copy(h, r)
	return n, [sha256.Size]byte(h.Sum(nil))
}

func TestProxiesAndLogsBodies(t *testing.T) {
	tests := []struct {
		name            string
		requestBody     string
		responseBody    string
		wantRequestLog  interface{}
		wantResponseLog interface{}
	}{
		{
			name:            "JSON",
			requestBody:     `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`,
			responseBody:    `{"usageMetadata":{"totalTokenCount":3}}`,
			wantRequestLog:  map[string]interface{}{"contents": []interface{}{map[string]interface{}{"role": "user", "parts": []interface{}{map[string]interface{}{"text": "hi"}}}}},
			wantResponseLog: map[string]interface{}{"usageMetadata": map[string]interface{}{"totalTokenCount": float64(3)}},
		},
		{
			name:            "not JSON",
			requestBody:     "plain request",
			responseBody:    "plain response",
			wantRequestLog:  "plain request",
			wantResponseLog: "plain response",
		},
		{
			name:            "empty",
			wantRequestLog:  "",
			wantResponseLog: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gotPath, gotBody string
			proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				gotPath, gotBody = r.URL.Path, string(body)
				io.WriteString(w, test.responseBody)
			})

			resp, err := http.Post(proxyURL+"/litmus-context-run1/v1/models/m:generateContent", "application/json", strings.NewReader(test.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			responseBody, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK || string(responseBody) != test.responseBody {
				t.Errorf("client got %d %q, want 200 %q", resp.StatusCode, responseBody, test.responseBody)
			}
			if gotPath != "/v1/models/m:generateContent" || gotBody != test.requestBody {
				t.Errorf("upstream got %s %q, want /v1/models/m:generateContent %q", gotPath, gotBody, test.requestBody)
			}

			entry := nextLogEntry()
			if entry.LitmusContext != "run1" {
				t.Errorf("logged litmusContext = %q, want %q", entry.LitmusContext, "run1")
			}
			if !reflect.DeepEqual(entry.RequestBody, test.wantRequestLog) || entry.RequestSize != int64(len(test.requestBody)) || entry.RequestBodyOmitted {
				t.Errorf("logged request = %#v (size %d, omitted %t), want %#v (size %d)", entry.RequestBody, entry.RequestSize, entry.RequestBodyOmitted, test.wantRequestLog, len(test.requestBody))
			}
			if !reflect.DeepEqual(entry.ResponseBody, test.wantResponseLog) || entry.ResponseSize != int64(len(test.responseBody)) || entry.ResponseBodyOmitted {
				t.Errorf("logged response = %#v (size %d, omitted %t), want %#v (size %d)", entry.ResponseBody, entry.ResponseSize, entry.ResponseBodyOmitted, test.wantResponseLog, len(test.responseBody))
			}
		})
	}
}

// A request body is logged up to and including maxLoggedBodyBytes. Beyond
// that it still reaches the upstream intact, but only its size is logged.
func TestRequestBodyLoggingLimit(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		wantOmitted bool
	}{
		{"at the limit", maxLoggedBodyBytes, false},
		{"over the limit", maxLoggedBodyBytes + 1, true},
		{"far over the limit", 20 * maxLoggedBodyBytes, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := jsonString(test.size)
			var gotSize int64
			var gotHash [sha256.Size]byte
			proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
				gotSize, gotHash = drain(r.Body)
				io.WriteString(w, `{"ok":true}`)
			})

			resp, err := http.Post(proxyURL+"/v1/x", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("client got status %d, want 200", resp.StatusCode)
			}
			if gotSize != int64(len(body)) || gotHash != sha256.Sum256(body) {
				t.Errorf("upstream got a different body (%d bytes) than the %d bytes that were sent", gotSize, len(body))
			}

			entry := nextLogEntry()
			if entry.RequestSize != int64(len(body)) {
				t.Errorf("logged requestSize = %d, want %d", entry.RequestSize, len(body))
			}
			if entry.RequestBodyOmitted != test.wantOmitted {
				t.Errorf("logged requestBodyOmitted = %t, want %t", entry.RequestBodyOmitted, test.wantOmitted)
			}
			var want interface{}
			if !test.wantOmitted {
				want = string(body[1 : len(body)-1])
			}
			if !reflect.DeepEqual(entry.RequestBody, want) {
				t.Errorf("logged requestBody is not the expected value (omitted: %t)", test.wantOmitted)
			}
			if want := map[string]interface{}{"ok": true}; !reflect.DeepEqual(entry.ResponseBody, want) || entry.ResponseBodyOmitted {
				t.Errorf("logged response = %#v (omitted %t), want %#v", entry.ResponseBody, entry.ResponseBodyOmitted, want)
			}
		})
	}
}

// The proxy must pass a large request body on while it is still arriving. If
// it waited for the end of the body first, it would have to hold all of it in
// memory, and a client could send more than the proxy has.
func TestLargeRequestBodyIsStreamed(t *testing.T) {
	body := bytes.Repeat([]byte("0123456789abcdef"), 4<<20/16)
	start, rest := body[:2*maxLoggedBodyBytes], body[2*maxLoggedBodyBytes:]

	var gotSize int64
	var gotHash [sha256.Size]byte
	arrived := make(chan struct{}) // closed when the upstream has seen the start of the body
	proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
		first := make([]byte, 1)
		if _, err := io.ReadFull(r.Body, first); err == nil {
			close(arrived)
		}
		gotSize, gotHash = drain(io.MultiReader(bytes.NewReader(first), r.Body))
		io.WriteString(w, `{"ok":true}`)
	})

	// Send the start of the body, then hold the rest back until the start has
	// come out at the other end of the proxy.
	errNotStreamed := errors.New("the start of the request body did not reach the upstream before the end of the body was sent")
	pr, pw := io.Pipe()
	go func() {
		if _, err := pw.Write(start); err != nil {
			return
		}
		select {
		case <-arrived:
			_, err := pw.Write(rest)
			pw.CloseWithError(err)
		case <-time.After(10 * time.Second):
			pw.CloseWithError(errNotStreamed)
		}
	}()

	resp, err := http.Post(proxyURL+"/v1/x", "application/octet-stream", pr)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("client got status %d, want 200", resp.StatusCode)
	}
	if gotSize != int64(len(body)) || gotHash != sha256.Sum256(body) {
		t.Errorf("upstream got a different body (%d bytes) than the %d bytes that were sent", gotSize, len(body))
	}
	entry := nextLogEntry()
	if entry.RequestBody != nil || !entry.RequestBodyOmitted || entry.RequestSize != int64(len(body)) {
		t.Errorf("logged request = %T (size %d, omitted %t), want no body (size %d, omitted true)", entry.RequestBody, entry.RequestSize, entry.RequestBodyOmitted, len(body))
	}
}

// A response body is logged up to and including maxLoggedBodyBytes. Beyond
// that it still reaches the client intact, but only its size is logged.
func TestResponseBodyLoggingLimit(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		wantOmitted bool
	}{
		{"at the limit", maxLoggedBodyBytes, false},
		{"over the limit", maxLoggedBodyBytes + 1, true},
		{"far over the limit", 20 * maxLoggedBodyBytes, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := jsonString(test.size)
			proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
				w.Write(body)
			})

			resp, err := http.Get(proxyURL + "/v1/x")
			if err != nil {
				t.Fatal(err)
			}
			gotSize, gotHash := drain(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK || gotSize != int64(len(body)) || gotHash != sha256.Sum256(body) {
				t.Errorf("client got status %d and a body of %d bytes, want 200 and the %d bytes that the upstream sent", resp.StatusCode, gotSize, len(body))
			}

			entry := nextLogEntry()
			if entry.ResponseSize != int64(len(body)) {
				t.Errorf("logged responseSize = %d, want %d", entry.ResponseSize, len(body))
			}
			if entry.ResponseBodyOmitted != test.wantOmitted {
				t.Errorf("logged responseBodyOmitted = %t, want %t", entry.ResponseBodyOmitted, test.wantOmitted)
			}
			var want interface{}
			if !test.wantOmitted {
				want = string(body[1 : len(body)-1])
			}
			if !reflect.DeepEqual(entry.ResponseBody, want) {
				t.Errorf("logged responseBody is not the expected value (omitted: %t)", test.wantOmitted)
			}
		})
	}
}

func gzipped(t *testing.T, r io.Reader) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := io.Copy(zw, r); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// A gzip response is passed to the client as it is and decompressed for the
// log only, and only if that takes no more than maxLoggedBodyBytes.
func TestGzipResponse(t *testing.T) {
	const document = `{"usageMetadata":{"totalTokenCount":3}}`
	tests := []struct {
		name     string
		body     []byte
		wantLog  interface{}
		wantSize int64 // 0 means the size of body
	}{
		{
			name:     "small",
			body:     gzipped(t, strings.NewReader(document)),
			wantLog:  map[string]interface{}{"usageMetadata": map[string]interface{}{"totalTokenCount": float64(3)}},
			wantSize: int64(len(document)),
		},
		{
			name:     "inflates to the limit",
			body:     gzipped(t, bytes.NewReader(jsonString(maxLoggedBodyBytes))),
			wantLog:  strings.Repeat("a", maxLoggedBodyBytes-2),
			wantSize: maxLoggedBodyBytes,
		},
		{
			name: "inflates beyond the limit",
			body: gzipped(t, bytes.NewReader(jsonString(maxLoggedBodyBytes+1))),
		},
		{
			// 64 MiB of zeros compress to about 64 KiB.
			name: "decompression bomb",
			body: gzipped(t, io.LimitReader(zeros{}, 64<<20)),
		},
		{
			name: "not gzip at all",
			body: []byte("this is not gzip"),
		},
		{
			name: "cut short",
			body: gzipped(t, strings.NewReader(document))[:20],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Encoding", "gzip")
				w.Write(test.body)
			})

			req, err := http.NewRequest(http.MethodGet, proxyURL+"/v1/x", nil)
			if err != nil {
				t.Fatal(err)
			}
			// Asking for gzip explicitly makes the client return the
			// compressed bytes instead of decompressing them itself.
			req.Header.Set("Accept-Encoding", "gzip")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			got, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK || !bytes.Equal(got, test.body) {
				t.Errorf("client got status %d and %d bytes, want 200 and the %d bytes that the upstream sent", resp.StatusCode, len(got), len(test.body))
			}

			entry := nextLogEntry()
			wantOmitted := test.wantLog == nil
			wantSize := test.wantSize
			if wantSize == 0 {
				wantSize = int64(len(test.body))
			}
			if !reflect.DeepEqual(entry.ResponseBody, test.wantLog) || entry.ResponseBodyOmitted != wantOmitted || entry.ResponseSize != wantSize {
				t.Errorf("logged response = %T (size %d, omitted %t), want %T (size %d, omitted %t)", entry.ResponseBody, entry.ResponseSize, entry.ResponseBodyOmitted, test.wantLog, wantSize, wantOmitted)
			}
		})
	}
}

// allocated returns the number of bytes that f made the process allocate.
func allocated(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// The memory that a request costs the proxy must not grow with the size of
// the bodies: anyone who can reach the proxy decides how large they are.
func TestMemoryUseDoesNotGrowWithBodySize(t *testing.T) {
	const (
		bodySize = 64 << 20
		// What handling one request may allocate, here and in the test's
		// own client and upstream. Buffering a body takes several times
		// bodySize.
		maxAllocated = bodySize / 4
	)

	t.Run("request body", func(t *testing.T) {
		var gotSize int64
		proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
			gotSize, _ = io.Copy(io.Discard, r.Body)
		})
		got := allocated(func() {
			resp, err := http.Post(proxyURL+"/v1/x", "application/octet-stream", io.LimitReader(zeros{}, bodySize))
			if err != nil {
				t.Fatal(err)
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			nextLogEntry()
		})
		if gotSize != bodySize {
			t.Errorf("upstream got %d bytes, want %d", gotSize, bodySize)
		}
		t.Logf("allocated %d KiB", got>>10)
		if got > maxAllocated {
			t.Errorf("proxying a %d MiB request body allocated %d MiB, want at most %d MiB", bodySize>>20, got>>20, maxAllocated>>20)
		}
	})

	t.Run("response body", func(t *testing.T) {
		proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, io.LimitReader(zeros{}, bodySize))
		})
		var gotSize int64
		got := allocated(func() {
			resp, err := http.Get(proxyURL + "/v1/x")
			if err != nil {
				t.Fatal(err)
			}
			gotSize, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			nextLogEntry()
		})
		if gotSize != bodySize {
			t.Errorf("client got %d bytes, want %d", gotSize, bodySize)
		}
		t.Logf("allocated %d KiB", got>>10)
		if got > maxAllocated {
			t.Errorf("proxying a %d MiB response body allocated %d MiB, want at most %d MiB", bodySize>>20, got>>20, maxAllocated>>20)
		}
	})

	t.Run("gzip response body", func(t *testing.T) {
		bomb := gzipped(t, io.LimitReader(zeros{}, bodySize))
		proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Encoding", "gzip")
			w.Write(bomb)
		})
		got := allocated(func() {
			req, err := http.NewRequest(http.MethodGet, proxyURL+"/v1/x", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Accept-Encoding", "gzip")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			nextLogEntry()
		})
		t.Logf("allocated %d KiB", got>>10)
		if got > maxAllocated {
			t.Errorf("proxying a %d KiB response that inflates to %d MiB allocated %d MiB, want at most %d MiB", len(bomb)>>10, bodySize>>20, got>>20, maxAllocated>>20)
		}
	})
}

// Bodies are only turned into log entries while their combined size is within
// maxConcurrentLoggedBytes, however many requests there are.
func TestBodiesBeingLoggedAreLimited(t *testing.T) {
	const requests = 16
	body := jsonString(maxLoggedBodyBytes * 3 / 4)

	var mu sync.Mutex
	var beingLogged, mostBeingLogged int64
	omitted := 0

	proxyURL, _ := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
	})
	// Take long enough over each entry for the entries to pile up.
	writeLogEntry = func(_ context.Context, entry logging.Entry) error {
		payload := entry.Payload.(requestLog)
		mu.Lock()
		if payload.RequestBodyOmitted {
			omitted++
		}
		beingLogged += payload.RequestSize
		mostBeingLogged = max(mostBeingLogged, beingLogged)
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		beingLogged -= payload.RequestSize
		mu.Unlock()
		return nil
	}

	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Post(proxyURL+"/v1/x", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Error(err)
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if mostBeingLogged > maxConcurrentLoggedBytes {
		t.Errorf("%d bytes of bodies were being logged at the same time, want at most %d", mostBeingLogged, maxConcurrentLoggedBytes)
	}
	if omitted != 0 {
		t.Errorf("%d of %d requests were logged without their body, want all bodies logged", omitted, requests)
	}
}

// A request that cannot get capacity to log its bodies within logQueueTimeout
// is answered all the same, and logged without them.
func TestLogsWithoutBodiesWhenCapacityIsExhausted(t *testing.T) {
	realTimeout := logQueueTimeout
	logQueueTimeout = 10 * time.Millisecond
	t.Cleanup(func() { logQueueTimeout = realTimeout })

	if !logCapacity.TryAcquire(maxConcurrentLoggedBytes) {
		t.Fatal("the logging capacity is in use before the test has started")
	}
	t.Cleanup(func() { logCapacity.Release(maxConcurrentLoggedBytes) })

	const requestBody, responseBody = `{"prompt":"hi"}`, `{"answer":"hello"}`
	proxyURL, nextLogEntry := startProxy(t, func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		io.WriteString(w, responseBody)
	})

	resp, err := http.Post(proxyURL+"/v1/x", "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(got) != responseBody {
		t.Errorf("client got %d %q, want 200 %q", resp.StatusCode, got, responseBody)
	}

	entry := nextLogEntry()
	if entry.RequestBody != nil || !entry.RequestBodyOmitted || entry.RequestSize != int64(len(requestBody)) {
		t.Errorf("logged request = %#v (size %d, omitted %t), want no body (size %d, omitted true)", entry.RequestBody, entry.RequestSize, entry.RequestBodyOmitted, len(requestBody))
	}
	if entry.ResponseBody != nil || !entry.ResponseBodyOmitted || entry.ResponseSize != int64(len(responseBody)) {
		t.Errorf("logged response = %#v (size %d, omitted %t), want no body (size %d, omitted true)", entry.ResponseBody, entry.ResponseSize, entry.ResponseBodyOmitted, len(responseBody))
	}

	// A request without bodies needs no capacity and is logged as usual.
	resp, err = http.Get(proxyURL + "/v1/empty")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	entry = nextLogEntry()
	if entry.RequestBody != "" || entry.RequestBodyOmitted {
		t.Errorf("logged request = %#v (omitted %t), want an empty body", entry.RequestBody, entry.RequestBodyOmitted)
	}
}

func TestLimitedBuffer(t *testing.T) {
	tests := []struct {
		name     string
		writes   []string
		wantData string
		wantSize int64
		wantOmit bool
	}{
		{name: "nothing written"},
		{name: "below the limit", writes: []string{"ab", "c"}, wantData: "abc", wantSize: 3},
		{name: "at the limit", writes: []string{"abcd", "efgh"}, wantData: "abcdefgh", wantSize: 8},
		{name: "over the limit in one write", writes: []string{"abcdefghi"}, wantSize: 9, wantOmit: true},
		{name: "over the limit in a later write", writes: []string{"abcdefgh", "i", "jkl"}, wantSize: 12, wantOmit: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := limitedBuffer{limit: 8}
			for _, w := range test.writes {
				if n, err := buf.Write([]byte(w)); n != len(w) || err != nil {
					t.Fatalf("Write(%q) = %d, %v, want %d, nil", w, n, err, len(w))
				}
			}
			got := buf.body()
			if string(got.data) != test.wantData || got.size != test.wantSize || got.omitted != test.wantOmit {
				t.Errorf("body() = {data: %q, size: %d, omitted: %t}, want {data: %q, size: %d, omitted: %t}", got.data, got.size, got.omitted, test.wantData, test.wantSize, test.wantOmit)
			}
			if test.wantOmit && buf.data != nil {
				t.Errorf("the buffer still holds %d bytes after going over its limit", len(buf.data))
			}
		})
	}
}

func TestGunzip(t *testing.T) {
	for _, size := range []int{0, 1, 100, 101, 5000} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			const limit = 100
			plain := bytes.Repeat([]byte("x"), size)
			got, err := gunzip(gzipped(t, bytes.NewReader(plain)), limit)
			if size > limit {
				if err == nil {
					t.Errorf("gunzip returned %d bytes for data that inflates to %d bytes, want an error above %d bytes", len(got), size, limit)
				}
				return
			}
			if err != nil || !bytes.Equal(got, plain) {
				t.Errorf("gunzip returned %d bytes and error %v, want the %d bytes that were compressed", len(got), err, size)
			}
		})
	}
}
