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
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

var (
	projectID      = os.Getenv("PROJECT_ID")
	firestoreDB    *firestore.Client
	upstreamURLStr = "https://" + os.Getenv("UPSTREAM_URL")
	tracingHeader  = "X-Litmus-Request" // Customizable tracing header name
)

type requestLog struct {
	ID             string      `firestore:"id"`
	TracingID      string      `firestore:"tracingID"` // Store the tracing ID
	Timestamp      time.Time   `firestore:"timestamp"`
	Method         string      `firestore:"method"`
	RequestURI     string      `firestore:"requestURI"`
	UpstreamURL    string      `firestore:"upstreamURL"`
	RequestHeaders http.Header `firestore:"requestHeaders"`
	RequestBody    string      `firestore:"requestBody"`
	RequestSize    int64       `firestore:"requestSize"`
	ResponseStatus int         `firestore:"responseStatus"`
	ResponseBody   string      `firestore:"responseBody"`
	ResponseSize   int64       `firestore:"responseSize"`
	Latency        int64       `firestore:"latency"`
}

func main() {
	// Initialize Firestore client
	ctx := context.Background()
	var err error
	firestoreDB, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreDB.Close()

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

	log.Fatal(http.ListenAndServe(":8082", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request, proxy *httputil.ReverseProxy, upstreamURL *url.URL) {
	startTime := time.Now()
	requestID := uuid.New().String()
	tracingID := r.Header.Get(tracingHeader)
	if tracingID == "" {
		tracingID = uuid.New().String() // Generate if not provided
	}

	// Ensure Correct Protocol Scheme
	if r.URL.Scheme == "" {
		r.URL.Scheme = upstreamURL.Scheme
	}

	if r.URL.Host == "" {
		r.URL.Host = upstreamURL.Host
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Reset the request body for the proxy
	r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	// Set the Host header to the upstream URL
	r.Host = upstreamURL.Host

	// Add tracing ID to the request header for propagation
	r.Header.Set(tracingHeader, tracingID)

	logRequest(requestID, tracingID, r, startTime, upstreamURL, requestBody)

	wrappedWriter := &statusRecorder{ResponseWriter: w}

	// Print request going to the upstream server
	reqDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("Error dumping request: %v", err)
	} else {
		log.Printf("Upstream Request:\n%s", string(reqDump))
	}

	// Explicitly call the proxy's ServeHTTP
	proxy.ServeHTTP(wrappedWriter, r)

	// Read the response body from the buffer
	responseBody := wrappedWriter.buf.Bytes()

	endTime := time.Now()
	logResponse(requestID, startTime, endTime, wrappedWriter, requestBody, responseBody)

}

func logRequest(requestID, tracingID string, r *http.Request, startTime time.Time, upstreamURL *url.URL, requestBody []byte) {
	// Store headers as a map for easier querying in Firestore
	headerMap := make(map[string][]string)
	for k, v := range r.Header {
		headerMap[k] = v
	}

	requestLog := requestLog{
		ID:             requestID,
		TracingID:      tracingID, // Store the tracing ID
		Timestamp:      startTime,
		Method:         r.Method,
		RequestURI:     r.RequestURI,
		UpstreamURL:    upstreamURL.String(),
		RequestHeaders: headerMap,
		RequestBody:    string(requestBody),
		RequestSize:    int64(len(requestBody)),
	}
	_, err := firestoreDB.Collection("requests").Doc(requestID).Set(context.Background(), requestLog)
	if err != nil {
		log.Printf("Failed to log request: %v", err)
	}
}

func logResponse(requestID string, startTime time.Time, endTime time.Time, w *statusRecorder, requestBody []byte, responseBody []byte) {
	responseSize := int64(len(requestBody))
	responseLog := map[string]interface{}{
		"responseStatus": w.status,
		"responseBody":   string(responseBody),
		"responseSize":   responseSize,
		"latency":        endTime.Sub(startTime).Milliseconds(),
	}
	// Correct Firestore Update
	_, err := firestoreDB.Collection("requests").Doc(requestID).Update(context.Background(), []firestore.Update{
		{
			Path:  "responseStatus",
			Value: responseLog["responseStatus"],
		},
		{
			Path:  "responseBody",
			Value: responseLog["responseBody"],
		},
		{
			Path:  "responseSize",
			Value: responseLog["responseSize"],
		},
		{
			Path:  "latency",
			Value: responseLog["latency"],
		},
	})
	if err != nil {
		log.Printf("Failed to log response: %v", err)
	}
}

// statusRecorder modified to capture the response body
type statusRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

// Write reimplements the necessary methods to capture the response body
func (rec *statusRecorder) Write(b []byte) (int, error) {
	rec.buf.Write(b)
	return rec.ResponseWriter.Write(b)
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
