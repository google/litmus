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
)

type requestLog struct {
	ID             string    `firestore:"id"`
	Timestamp      time.Time `firestore:"timestamp"`
	Method         string    `firestore:"method"`
	RequestURI     string    `firestore:"requestURI"`
	UpstreamURL    string    `firestore:"upstreamURL"`
	RequestSize    int64     `firestore:"requestSize"`
	ResponseStatus int       `firestore:"responseStatus"`
	ResponseSize   int64     `firestore:"responseSize"`
	Latency        int64     `firestore:"latency"`
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

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request, proxy *httputil.ReverseProxy, upstreamURL *url.URL) {
	startTime := time.Now()
	requestID := uuid.New().String()

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
	// Reset the request body for the proxy using io.NopCloser
	r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	logRequest(requestID, r, startTime, upstreamURL, requestBody)

	wrappedWriter := &statusRecorder{ResponseWriter: w}

	// Explicitly call the proxy's ServeHTTP
	proxy.ServeHTTP(wrappedWriter, r)

	endTime := time.Now()
	logResponse(requestID, r, startTime, endTime, wrappedWriter, requestBody)
}

func logRequest(requestID string, r *http.Request, startTime time.Time, upstreamURL *url.URL, requestBody []byte) {
	requestLog := requestLog{
		ID:          requestID,
		Timestamp:   startTime,
		Method:      r.Method,
		RequestURI:  r.RequestURI,
		UpstreamURL: upstreamURL.String(),
		RequestSize: int64(len(requestBody)), // Calculate size from the stored body
	}
	_, err := firestoreDB.Collection("requests").Doc(requestID).Set(context.Background(), requestLog)
	if err != nil {
		log.Printf("Failed to log request: %v", err)
	}
}

func logResponse(requestID string, r *http.Request, startTime time.Time, endTime time.Time, w *statusRecorder, requestBody []byte) {
	responseSize := int64(len(requestBody))
	responseLog := map[string]interface{}{
		"responseStatus": w.status,
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

// statusRecorder remains the same
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
