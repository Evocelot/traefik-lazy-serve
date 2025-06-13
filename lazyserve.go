package traefik_lazy_serve

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"
)

// Config defines the plugin's configurable parameters.
type Config struct {
	MaxRetries int           `json:"maxRetries,omitempty"` // Number of retry attempts
	RetryDelay time.Duration `json:"retryDelay,omitempty"` // Delay between retries
}

// CreateConfig returns the default configuration values.
func CreateConfig() *Config {
	return &Config{
		MaxRetries: 3,
		RetryDelay: 2000,
	}
}

// LazyServe represents the middleware instance.
type LazyServe struct {
	next       http.Handler  // Next handler in the chain
	name       string        // Name of the middleware
	maxRetries int           // Number of times to retry
	retryDelay time.Duration // Delay between retries
}

// New creates a new instance of the middleware with the provided configuration.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.MaxRetries < 1 {
		return nil, errors.New("maxRetries must be >= 1")
	}
	if config.RetryDelay <= 0 {
		return nil, errors.New("retryDelay must be > 0")
	}

	retryDuration := time.Duration(config.RetryDelay) * time.Millisecond

	log.SetOutput(os.Stdout)

	log.Printf("[lazyserve] Initializing middleware '%s' with maxRetries=%d and retryDelay=%s", name, config.MaxRetries, retryDuration)

	return &LazyServe{
		next:       next,
		name:       name,
		maxRetries: config.MaxRetries,
		retryDelay: retryDuration,
	}, nil
}

// ServeHTTP attempts to handle the request, retrying if the backend responds with a server error.
func (m *LazyServe) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var attempt int
	var recorder *responseRecorder

	for attempt = 0; attempt < m.maxRetries; attempt++ {
		recorder = &responseRecorder{ResponseWriter: rw, statusCode: 0, header: http.Header{}}
		m.next.ServeHTTP(recorder, req)

		// If the response is successful (not a 5xx), forward it
		if recorder.statusCode < 500 && recorder.statusCode != 0 {
			log.Printf("[lazyserve] Request %s %s succeeded on attempt %d with status %d",
				req.Method, req.URL.Path, attempt, recorder.statusCode)
			recorder.WriteToOriginal(rw)
			return
		}

		log.Printf("[lazyserve] Attempt %d failed for request %s %s (status: %d), retrying in %s...",
			attempt, req.Method, req.URL.Path, recorder.statusCode, m.retryDelay)

		// Wait before retrying
		time.Sleep(m.retryDelay)
	}

	log.Printf("[lazyserve] All %d attempts failed for request %s %s. Responding with last received status: %d",
		m.maxRetries, req.Method, req.URL.Path, recorder.statusCode)

	// After all retries, write the last response
	recorder.WriteToOriginal(rw)
}
