package main

import (
	"context"
	"fmt"
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
	projectID   = os.Getenv("PROJECT_ID")
	firestoreDB *firestore.Client
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

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Generate unique request ID
	requestID := uuid.New().String()

	// Log incoming request details
	logRequest(requestID, r, startTime)

	// Proxy the request
	upstreamURL := getUpstreamURL(r.Host)
	if upstreamURL == nil {
		http.Error(w, "Invalid upstream URL", http.StatusInternalServerError)
		return
	}

	// Wrap the ResponseWriter to capture status code
	wrappedWriter := &statusRecorder{ResponseWriter: w}

	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
	proxy.ServeHTTP(wrappedWriter, r) // Use the wrapped writer

	// Log outgoing response details
	endTime := time.Now()
	logResponse(requestID, r, startTime, endTime, wrappedWriter) // Pass the wrapped writer
}

func getUpstreamURL(host string) *url.URL {
	// Replace this with your actual logic to determine the upstream URL
	// based on the request host.
	upstreamHost := fmt.Sprintf("%s-aiplatform.googleapis.com", host)
	upstreamURL, err := url.Parse(fmt.Sprintf("https://%s", upstreamHost))
	if err != nil {
		return nil
	}
	return upstreamURL
}

func logRequest(requestID string, r *http.Request, startTime time.Time) {
	// Log request details to Firestore
	requestLog := requestLog{
		ID:          requestID,
		Timestamp:   startTime,
		Method:      r.Method,
		RequestURI:  r.RequestURI,
		UpstreamURL: r.Host,
		RequestSize: r.ContentLength,
	}
	_, err := firestoreDB.Collection("requests").Doc(requestID).Set(context.Background(), requestLog)
	if err != nil {
		log.Printf("Failed to log request: %v", err)
	}
}

func logResponse(requestID string, r *http.Request, startTime time.Time, endTime time.Time, w *statusRecorder) {
	// Log response details to Firestore
	responseSize, _ := io.Copy(io.Discard, r.Body) // Handle the error if needed
	responseLog := map[string]interface{}{
		"responseStatus": w.status, // Get the captured status
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

// statusRecorder captures the status code from WriteHeader
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader implements the ResponseWriter interface
func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
