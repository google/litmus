package tunnel

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/litmus/cli/utils"
	"golang.org/x/net/context"
)

// CreateTunnel creates a tunnel to the Litmus service URL.
func CreateTunnel(cloudRunEndpoint string, localPort int, quiet bool, projectID string) error {

	endpointURL, err := url.Parse(cloudRunEndpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if cloudRunEndpoint == "" { 
		return fmt.Errorf("service URL is empty")
	}

	proxy := httputil.NewSingleHostReverseProxy(endpointURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = endpointURL.Scheme
		req.URL.Host = endpointURL.Host
		req.Host = endpointURL.Host
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
	}

	username, password, err := utils.GetAuthCredentials(projectID)
	if err != nil {
		return fmt.Errorf("error getting auth credentials: %w", err)
	}

	authProxy := &authMiddleware{
		username: username,
		password: password,
		next:     proxy,
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", localPort),
		Handler: authProxy,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	fmt.Printf("Tunnel created: Access Litmus at http://localhost:%d\n", localPort)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server ListenAndServe: %w", err)
	}

	<-idleConnsClosed
	if !quiet {
		log.Println("Tunnel closed")
	}
	return nil
}

// authMiddleware handles basic authentication for the tunnel.
type authMiddleware struct {
	username string
	password string
	next     http.Handler
}

// ServeHTTP handles the HTTP request, performing basic auth.
func (h *authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()

	if !ok || user != h.username || pass != h.password {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.next.ServeHTTP(w, r)
}