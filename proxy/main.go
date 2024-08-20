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
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/logging"
	"github.com/google/uuid"
)

var (
	projectID      = os.Getenv("PROJECT_ID")
	logger         *logging.Logger
	upstreamURLStr = "https://" + os.Getenv("UPSTREAM_URL")
	tracingHeader  = "X-Litmus-Request" // Customizable tracing header name
)

type requestLog struct {
	ID             string      `json:"id"`
	TracingID      string      `json:"tracingID"`
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
}

func main() {
	// Initialize Cloud Logging client
	ctx := context.Background()
	logClient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Cloud Logging client: %v", err)
	}
	defer logClient.Close()
	logger = logClient.Logger("my-proxy-log") // Use a meaningful log name

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
		tracingID = uuid.New().String() // Generate if not provided
	}

	// Ensure Correct Protocol Scheme
	if r.URL.Scheme == "" {
		r.URL.Scheme = upstreamURL.Scheme
	}

	if r.URL.Host == "" {
		r.URL.Host = upstreamURL.Host
	}

	// Create a new buffer to hold the request body
	requestBodyBuffer := bytes.NewBuffer(nil)
	// Copy the request body to the buffer
	if _, err := io.Copy(requestBodyBuffer, r.Body); err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get the byte slice from the buffer
	requestBody := requestBodyBuffer.Bytes()

	// Reset the request body for the proxy using the buffer
	r.Body = io.NopCloser(requestBodyBuffer)

	// Set the Host header to the upstream URL
	r.Host = upstreamURL.Host

	// Add tracing ID to the request header for propagation
	r.Header.Set(tracingHeader, tracingID)

	wrappedWriter := &statusRecorder{ResponseWriter: w}

	// Explicitly call the proxy's ServeHTTP
	proxy.ServeHTTP(wrappedWriter, r)

	endTime := time.Now()

	// Log the combined request and response details
	logRequestAndResponse(requestID, tracingID, r, startTime, endTime, upstreamURL, requestBody, wrappedWriter.buf.Bytes())
}

func logRequestAndResponse(requestID, tracingID string, r *http.Request, startTime time.Time, endTime time.Time, upstreamURL *url.URL, requestBody []byte, responseBody []byte) {

	// Attempt to unmarshal the request body
	var requestBodyJSON interface{}
	if err := json.Unmarshal(requestBody, &requestBodyJSON); err != nil {
		// If unmarshaling fails, keep the raw string
		requestBodyJSON = string(requestBody)
	}

	// Attempt to unmarshal the response body
	var responseBodyJSON interface{}
	if err := json.Unmarshal(responseBody, &responseBodyJSON); err != nil {
		// If unmarshaling fails, keep the raw string
		responseBodyJSON = string(responseBody)
	}

	requestLog := requestLog{
		ID:             requestID,
		TracingID:      tracingID,
		Timestamp:      startTime,
		Method:         r.Method,
		RequestURI:     r.RequestURI,
		UpstreamURL:    upstreamURL.String(),
		RequestHeaders: r.Header,
		RequestBody:    requestBodyJSON, // Use the unmarshalled or raw request body
		RequestSize:    int64(len(requestBody)),
		ResponseStatus: 0,                // Placeholder - will be updated below
		ResponseBody:   responseBodyJSON, // Use the unmarshalled or raw response body
		ResponseSize:   int64(len(responseBody)),
		Latency:        endTime.Sub(startTime).Milliseconds(),
	}

	// Update ResponseStatus now that we have it
	if rec, ok := r.Context().Value("statusRecorder").(*statusRecorder); ok {
		requestLog.ResponseStatus = rec.status
	}

	// Log the combined entry
	if err := logger.LogSync(context.Background(), logging.Entry{
		Payload: requestLog,
	}); err != nil {
		log.Printf("Failed to log request and response: %v", err)
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
	// Flush the buffer after writing
	return rec.ResponseWriter.Write(b)
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
