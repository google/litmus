// Copyright 2024 Google LLC
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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"

	"cloud.google.com/go/logging"
	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
)

const (
	// maxLoggedBodyBytes is the largest request or response body that the proxy
	// holds in memory and writes to its request log. Larger bodies are still
	// proxied, but they are streamed and the log only records their size.
	// Cloud Logging does not accept log entries above 256 KiB, so a larger body
	// could not be logged in any case.
	maxLoggedBodyBytes = 256 << 10

	// maxConcurrentLoggedBytes limits the combined size of the bodies that are
	// being turned into log entries at any one time. A structured log payload
	// can take over two hundred times the memory of the JSON document it was
	// parsed from (every "{" becomes a map, several times over), so the
	// bodies of concurrent requests must not all be converted at once. The
	// limit leaves room for the two bodies of one request.
	maxConcurrentLoggedBytes = 2 * maxLoggedBodyBytes
)

var (
	projectID      = os.Getenv("PROJECT_ID")
	logger         *logging.Logger
	upstreamURLStr = "https://" + os.Getenv("UPSTREAM_URL")
	tracingHeader  = "X-Litmus-Request" // Customizable tracing header name
	// Default to NOT logging the Authorization header for security reasons
	logAuthorizationHeader, _ = strconv.ParseBool(os.Getenv("LOG_AUTHORIZATION_HEADER"))
	// Regex to match /litmus-context-<random-string>/ path prefix
	contextPathRegex = regexp.MustCompile(`^/?(litmus-context-[a-zA-Z0-9\-]+)?(/.*)?$`)

	// logCapacity enforces maxConcurrentLoggedBytes.
	logCapacity = semaphore.NewWeighted(maxConcurrentLoggedBytes)
	// logQueueTimeout is how long a request waits for logCapacity before it is
	// logged without its bodies instead.
	logQueueTimeout = 5 * time.Second
	// writeLogEntry sends an entry to Cloud Logging. Tests replace it.
	writeLogEntry = func(ctx context.Context, entry logging.Entry) error {
		return logger.LogSync(ctx, entry)
	}
)

type requestLog struct {
	ID             string      `json:"id"`
	TracingID      string      `json:"tracingID"`
	LitmusContext  string      `json:"litmusContext"`
	Timestamp      time.Time   `json:"timestamp"`
	Method         string      `json:"method"`
	RequestURI     string      `json:"requestURI"`
	UpstreamURL    string      `json:"upstreamURL"`
	RequestHeaders http.Header `json:"requestHeaders"`
	RequestBody    interface{} `json:"requestBody"`
	RequestSize    int64       `json:"requestSize"`
	ResponseStatus int         `json:"responseStatus"`
	ResponseBody   interface{} `json:"responseBody"`
	ResponseSize   int64       `json:"responseSize"`
	Latency        int64       `json:"latency"`
	// Set when a body is left out of the log, see loggedBody.
	RequestBodyOmitted  bool `json:"requestBodyOmitted,omitempty"`
	ResponseBodyOmitted bool `json:"responseBodyOmitted,omitempty"`
}

// loggedBody is a request or response body as it goes into the request log.
type loggedBody struct {
	data []byte // the complete body, nil if omitted
	size int64  // the number of body bytes that went through the proxy
	// omitted reports that the body is not logged, because it is larger than
	// maxLoggedBodyBytes, could not be decoded, or had to give way to other
	// requests (see logQueueTimeout).
	omitted bool
}

// logValue returns the body in the form in which it is logged: parsed if it is
// JSON, as a string if it is not, and nil if it is omitted.
func (b loggedBody) logValue() interface{} {
	if b.omitted {
		return nil
	}
	var value interface{}
	if err := json.Unmarshal(b.data, &value); err != nil {
		// If unmarshaling fails, keep the raw string
		return string(b.data)
	}
	return value
}

// withoutData returns the body with its content left out of the log.
func (b loggedBody) withoutData() loggedBody {
	if len(b.data) == 0 {
		return b
	}
	return loggedBody{size: b.size, omitted: true}
}

// limitedBuffer is an io.Writer that keeps what is written to it only while
// the total stays within limit. Once the limit is exceeded it drops what it
// held and merely counts, so it never holds more than limit bytes however much
// is written to it.
type limitedBuffer struct {
	limit int64
	data  []byte
	size  int64
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	b.size += int64(len(p))
	if b.size > b.limit {
		b.data = nil
	} else {
		b.data = append(b.data, p...)
	}
	return len(p), nil
}

func (b *limitedBuffer) body() loggedBody {
	return loggedBody{data: b.data, size: b.size, omitted: b.size > b.limit}
}

// countingReader counts the bytes read through it. The count is atomic because
// the HTTP transport may still be reading a request body when the round trip
// has already returned.
type countingReader struct {
	io.Reader
	n atomic.Int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.Reader.Read(p)
	c.n.Add(int64(n))
	return n, err
}

// gunzip decompresses data, which must not inflate to more than limit bytes.
// The limit keeps a small compressed body from being expanded into an
// arbitrarily large buffer.
func gunzip(data []byte, limit int64) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	decoded, err := io.ReadAll(io.LimitReader(gr, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(decoded)) > limit {
		return nil, fmt.Errorf("decompressed body is larger than %d bytes", limit)
	}
	return decoded, nil
}

func main() {
	// Initialize Cloud Logging client
	ctx := context.Background()
	logClient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Cloud Logging client: %v", err)
	}
	defer logClient.Close()
	logger = logClient.Logger("litmus-proxy-log")

	// Validate UPSTREAM_URL
	if upstreamURLStr == "" {
		log.Fatal("UPSTREAM_URL environment variable is not set")
	}
	upstreamURL, err := url.Parse(upstreamURLStr)
	if err != nil {
		log.Fatalf("Invalid UPSTREAM_URL: %v", err)
	}

	// Explicitly create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)

	// Custom handler to wrap the proxy
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, proxy, upstreamURL)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request, proxy *httputil.ReverseProxy, upstreamURL *url.URL) {
	startTime := time.Now()
	requestID := uuid.New().String()
	tracingID := r.Header.Get(tracingHeader)
	if tracingID == "" {
		tracingID = uuid.New().String()
	}

	// Extract Litmus Context from path
	litmusContext, newPath := extractLitmusContext(r.URL.Path)
	r.URL.Path = newPath

	// If no context is found in the path, use the tracingID as the context
	if litmusContext == "" {
		litmusContext = tracingID
	}

	// Ensure Correct Protocol Scheme
	if r.URL.Scheme == "" {
		r.URL.Scheme = upstreamURL.Scheme
	}

	if r.URL.Host == "" {
		r.URL.Host = upstreamURL.Host
	}

	// Read the request body, but hold no more than maxLoggedBodyBytes of it in
	// memory. Reading all of it would let a client make the proxy allocate as
	// much memory as the client cares to send.
	requestBody, err := io.ReadAll(io.LimitReader(r.Body, maxLoggedBodyBytes+1))
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	request := loggedBody{data: requestBody, size: int64(len(requestBody))}
	var streamed *countingReader
	if request.size <= maxLoggedBodyBytes {
		// Reset the request body for the proxy using the bytes read
		r.Body = io.NopCloser(bytes.NewReader(requestBody))
	} else {
		// The body is too large to be logged. Pass on what has been read
		// followed by the rest of the body, which is streamed, not buffered.
		request.data, request.omitted = nil, true
		streamed = &countingReader{Reader: r.Body}
		r.Body = struct {
			io.Reader
			io.Closer
		}{io.MultiReader(bytes.NewReader(requestBody), streamed), r.Body}
	}

	// Set the Host header to the upstream URL
	r.Host = upstreamURL.Host

	// Add tracing ID to the request header for propagation
	r.Header.Set(tracingHeader, tracingID)

	// Copy request headers, potentially filtering out Authorization
	sanitizedHeaders := make(http.Header)
	for name, values := range r.Header {
		if name == "Authorization" && !logAuthorizationHeader {
			continue
		}
		sanitizedHeaders[name] = values
	}

	wrappedWriter := &statusRecorder{ResponseWriter: w, body: limitedBuffer{limit: maxLoggedBodyBytes}}

	// Explicitly call the proxy's ServeHTTP
	proxy.ServeHTTP(wrappedWriter, r)

	endTime := time.Now()

	if streamed != nil {
		request.size += streamed.n.Load()
	}

	// Handle gzip encoded response. The response has been sent to the client
	// by now, so a body that cannot be decoded is only left out of the log.
	response := wrappedWriter.body.body()
	if len(response.data) > 0 && wrappedWriter.Header().Get("Content-Encoding") == "gzip" {
		decoded, err := gunzip(response.data, maxLoggedBodyBytes)
		if err != nil {
			log.Printf("Failed to decompress response body: %v", err)
			response = response.withoutData()
		} else {
			response = loggedBody{data: decoded, size: int64(len(decoded))}
		}
	}

	// Log the combined request and response details
	logRequestAndResponse(requestID, tracingID, litmusContext, r, startTime, endTime, upstreamURL, request, response, sanitizedHeaders)
}

func logRequestAndResponse(requestID, tracingID, litmusContext string, r *http.Request, startTime time.Time, endTime time.Time, upstreamURL *url.URL, request loggedBody, response loggedBody, sanitizedHeaders http.Header) {

	// Turning the bodies into a log entry takes far more memory than the
	// bodies themselves, so wait for capacity first. If there is none for too
	// long because many large bodies are being logged, log this request
	// without its bodies rather than hold it up.
	if weight := int64(len(request.data) + len(response.data)); weight > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), logQueueTimeout)
		err := logCapacity.Acquire(ctx, weight)
		cancel()
		if err != nil {
			log.Printf("Logging request %s without bodies: too many bodies are being logged", requestID)
			request, response = request.withoutData(), response.withoutData()
		} else {
			defer logCapacity.Release(weight)
		}
	}

	// Use the unmarshalled or raw bodies, nil for a body that is omitted
	requestBodyJSON := request.logValue()
	responseBodyJSON := response.logValue()

	requestLog := requestLog{
		ID:             requestID,
		TracingID:      tracingID,
		LitmusContext:  litmusContext,
		Timestamp:      startTime,
		Method:         r.Method,
		RequestURI:     r.RequestURI,
		UpstreamURL:    upstreamURL.String(),
		RequestHeaders: sanitizedHeaders, // Log the potentially filtered headers
		RequestBody:    requestBodyJSON,  // Use the unmarshalled or raw request body
		RequestSize:    request.size,
		ResponseStatus: 0,                // Placeholder - will be updated below
		ResponseBody:   responseBodyJSON, // Use the unmarshalled or raw response body
		ResponseSize:   response.size,
		Latency:        endTime.Sub(startTime).Milliseconds(),

		RequestBodyOmitted:  request.omitted,
		ResponseBodyOmitted: response.omitted,
	}

	// Update ResponseStatus now that we have it
	if rec, ok := r.Context().Value("statusRecorder").(*statusRecorder); ok {
		requestLog.ResponseStatus = rec.status
	}

	// Log the combined entry
	if err := writeLogEntry(context.Background(), logging.Entry{
		Payload: requestLog,
	}); err != nil {
		log.Printf("Failed to log request and response: %v", err)
	}
}

// statusRecorder modified to capture the response body
type statusRecorder struct {
	http.ResponseWriter
	status int
	body   limitedBuffer
}

// Write reimplements the necessary methods to capture the response body
func (rec *statusRecorder) Write(b []byte) (int, error) {
	rec.body.Write(b)
	// Flush the buffer after writing
	return rec.ResponseWriter.Write(b)
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func extractLitmusContext(path string) (string, string) {
	matches := contextPathRegex.FindStringSubmatch(path)
	// If there is a context
	if len(matches) == 3 {
		// Extract the context value and update the path
		context := matches[1]
		newPath := matches[2]

		// Remove the "litmus-context-" prefix
		if len(context) > 15 && context[:15] == "litmus-context-" {
			context = context[15:]
		}

		return context, newPath
	}
	// If there is no context
	if len(matches) == 2 {
		newPath := matches[1]
		return "", newPath
	}
	return "", path // Return empty string if no match
}
