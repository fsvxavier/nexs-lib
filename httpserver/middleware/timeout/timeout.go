// Package timeout provides HTTP request timeout middleware implementation.
package timeout

import (
	"context"
	"net/http"
	"time"
)

// Config represents timeout configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass timeout.
	SkipPaths []string
	// Timeout is the request timeout duration.
	Timeout time.Duration
	// Message is the timeout error message.
	Message string
	// ErrorHandler handles timeout errors.
	ErrorHandler func(http.ResponseWriter, *http.Request)
}

// IsEnabled returns true if the middleware is enabled.
func (c Config) IsEnabled() bool {
	return c.Enabled
}

// ShouldSkip returns true if the given path should be skipped.
func (c Config) ShouldSkip(path string) bool {
	for _, skipPath := range c.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// DefaultConfig returns a default timeout configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:      true,
		Timeout:      30 * time.Second,
		Message:      "Request timeout",
		ErrorHandler: defaultErrorHandler,
	}
}

// Middleware implements timeout middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new timeout middleware.
func NewMiddleware(config Config) *Middleware {
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}
	return &Middleware{
		config: config,
	}
}

// Wrap implements the interfaces.Middleware interface.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		if m.config.ShouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Create timeout context
		ctx, cancel := context.WithTimeout(r.Context(), m.config.Timeout)
		defer cancel()

		// Create a channel to signal completion
		done := make(chan struct{})

		// Create a new request with timeout context
		r = r.WithContext(ctx)

		go func() {
			defer close(done)
			next.ServeHTTP(w, r)
		}()

		select {
		case <-done:
			// Request completed normally
			return
		case <-ctx.Done():
			// Request timed out
			if ctx.Err() == context.DeadlineExceeded {
				m.config.ErrorHandler(w, r)
			}
			return
		}
	})
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "timeout"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 150 // Timeout should happen relatively early
}

// defaultErrorHandler sends a timeout error response.
func defaultErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Request timeout", http.StatusRequestTimeout)
}
