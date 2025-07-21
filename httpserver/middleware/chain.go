// Package middleware provides framework-agnostic middleware implementations.
package middleware

import (
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Chain implements the MiddlewareChain interface.
type Chain struct {
	middlewares []interfaces.Middleware
}

// NewChain creates a new middleware chain.
func NewChain(middlewares ...interfaces.Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Add adds middleware to the chain.
func (c *Chain) Add(middleware interfaces.Middleware) interfaces.MiddlewareChain {
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Then wraps the given handler with all middleware in the chain.
func (c *Chain) Then(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}

	// Sort by priority (lower numbers first)
	middlewares := make([]interfaces.Middleware, len(c.middlewares))
	copy(middlewares, c.middlewares)

	for i := 0; i < len(middlewares); i++ {
		for j := i + 1; j < len(middlewares); j++ {
			if middlewares[i].Priority() > middlewares[j].Priority() {
				middlewares[i], middlewares[j] = middlewares[j], middlewares[i]
			}
		}
	}

	// Build chain from last to first
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i].Wrap(h)
	}

	return h
}

// Context keys for middleware data.
type contextKey string

const (
	// CorrelationIDKey is the context key for correlation ID.
	CorrelationIDKey contextKey = "correlation_id"
	// RequestStartTimeKey is the context key for request start time.
	RequestStartTimeKey contextKey = "request_start_time"
	// BulkheadResourceKey is the context key for bulkhead resource.
	BulkheadResourceKey contextKey = "bulkhead_resource"
)

// ResponseWriter wraps http.ResponseWriter to capture response data.
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

// NewResponseWriter creates a new ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code.
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.Size += n
	return n, err
}

// Config represents common middleware configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths to skip for this middleware.
	SkipPaths []string
	// Headers contains custom headers to add.
	Headers map[string]string
}

// IsEnabled returns whether the middleware is enabled.
func (c *Config) IsEnabled() bool {
	return c.Enabled
}

// ShouldSkip checks if a path should be skipped.
func (c *Config) ShouldSkip(path string) bool {
	for _, skipPath := range c.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// RetryConfig represents retry configuration.
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffMultiple float64
	RetryableStatus []int
}

// BulkheadConfig represents bulkhead configuration.
type BulkheadConfig struct {
	MaxConcurrent int
	QueueSize     int
	Timeout       time.Duration
}

// TimeoutConfig represents timeout configuration.
type TimeoutConfig struct {
	Timeout time.Duration
	Message string
}
