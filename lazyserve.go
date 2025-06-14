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
	MaxRetries       int           `json:"maxRetries,omitempty"`       // Number of retry attempts
	RetryDelay       time.Duration `json:"retryDelay,omitempty"`       // Delay between retries
	RetryStatusCodes []int         `json:"retryStatusCodes,omitempty"` // List of HTTP status codes to retry on
}

// CreateConfig returns the default configuration values.
func CreateConfig() *Config {
	return &Config{
		MaxRetries:       5,
		RetryDelay:       1000,
		RetryStatusCodes: []int{502, 503, 504},
	}
}

// LazyServe represents the middleware instance.
type LazyServe struct {
	next             http.Handler  // Next handler in the chain
	name             string        // Name of the middleware
	maxRetries       int           // Number of times to retry
	retryDelay       time.Duration // Delay between retries
	retryStatusCodes map[int]bool  // Set of HTTP status codes to retry on
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

	// Convert the slice into a map for faster lookup
	retryCodeMap := make(map[int]bool)
	for _, code := range config.RetryStatusCodes {
		retryCodeMap[code] = true
	}

	log.SetOutput(os.Stdout)

	log.Printf("[lazyserve] Initializing middleware '%s' with maxRetries=%d and retryDelay=%s", name, config.MaxRetries, retryDuration)

	return &LazyServe{
		next:             next,
		name:             name,
		maxRetries:       config.MaxRetries,
		retryDelay:       retryDuration,
		retryStatusCodes: retryCodeMap,
	}, nil
}

// ServeHTTP attempts to handle the request, retrying if the backend responds with a server error.
func (m *LazyServe) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var attempt int
	var recorder *responseRecorder

	for attempt = 1; attempt <= m.maxRetries; attempt++ {
		recorder = &responseRecorder{ResponseWriter: rw, statusCode: 0, header: http.Header{}}
		m.next.ServeHTTP(recorder, req)

		// If not a retryable status code, treat it as successful
		if !m.retryStatusCodes[recorder.statusCode] {
			log.Printf("[lazyserve] Request %s %s completed on attempt %d with status %d",
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
