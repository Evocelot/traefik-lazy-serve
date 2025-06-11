package lazyserve

import (
	"context"
	"errors"
	"net/http"
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
		RetryDelay: 2 * time.Second,
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

	return &LazyServe{
		next:       next,
		name:       name,
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
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
			recorder.WriteToOriginal(rw)
			return
		}

		// Wait before retrying
		time.Sleep(m.retryDelay)
	}

	// After all retries, write the last response
	recorder.WriteToOriginal(rw)
}
