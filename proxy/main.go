package main

import (
	"context"
	"io"
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
	upstreamURLStr = os.Getenv("UPSTREAM_URL") // Env var for upstream URL
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

	logRequest(requestID, r, startTime, upstreamURL)

	wrappedWriter := &statusRecorder{ResponseWriter: w}

	// Explicitly call the proxy's ServeHTTP
	proxy.ServeHTTP(wrappedWriter, r)

	endTime := time.Now()
	logResponse(requestID, r, startTime, endTime, wrappedWriter)
}

func logRequest(requestID string, r *http.Request, startTime time.Time, upstreamURL *url.URL) {
	requestLog := requestLog{
		ID:          requestID,
		Timestamp:   startTime,
		Method:      r.Method,
		RequestURI:  r.RequestURI,
		UpstreamURL: upstreamURL.String(),
		RequestSize: r.ContentLength,
	}
	_, err := firestoreDB.Collection("requests").Doc(requestID).Set(context.Background(), requestLog)
	if err != nil {
		log.Printf("Failed to log request: %v", err)
	}
}

func logResponse(requestID string, r *http.Request, startTime time.Time, endTime time.Time, w *statusRecorder) {
	responseSize, _ := io.Copy(io.Discard, r.Body)
	responseLog := map[string]interface{}{
		"responseStatus": w.status,
		"responseSize":   responseSize,
		"latency":        endTime.Sub(startTime).Milliseconds(),
	}
	_, err := firestoreDB.Collection("requests").Doc(requestID).Update(context.Background(), []firestore.Update{
		{
			Path:  "",
			Value: responseLog,
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
