// Package traceid provides trace ID middleware implementation for request tracing.
package traceid

import (
	"context"
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"strings"
)

// Config represents trace ID configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass trace ID generation.
	SkipPaths []string
	// HeaderName is the name of the header to use for trace ID.
	HeaderName string
	// ContextKey is the key to use when storing trace ID in context.
	ContextKey string
	// Generator is a function that generates new trace IDs.
	Generator func() string
	// AlternativeHeaders are alternative header names to check for existing trace ID.
	AlternativeHeaders []string
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

// generateULID generates a ULID-like string for trace ID.
func generateULID() string {
	// Generate 16 random bytes
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to simple random string
		return fmt.Sprintf("trace-%d", mathrand.Int63())
	}

	// Convert to hex string (32 characters)
	return fmt.Sprintf("%x", b)
}

// DefaultConfig returns a default trace ID configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:    true,
		HeaderName: "X-Trace-Id",
		ContextKey: "trace_id",
		Generator:  generateULID,
		AlternativeHeaders: []string{
			"Trace-Id",
			"trace-id",
			"Trace-ID",
			"X-Trace-ID",
			"x-trace-id",
		},
	}
}

// Middleware implements trace ID middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new trace ID middleware.
func NewMiddleware(config Config) *Middleware {
	if config.Generator == nil {
		config.Generator = generateULID
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-Trace-Id"
	}
	if config.ContextKey == "" {
		config.ContextKey = "trace_id"
	}

	return &Middleware{
		config: config,
	}
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "traceid"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 50 // Very high priority, should run early
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

		// Try to get existing trace ID from headers
		traceID := r.Header.Get(m.config.HeaderName)

		// If not found, try alternative headers
		if traceID == "" {
			for _, altHeader := range m.config.AlternativeHeaders {
				if traceID = r.Header.Get(altHeader); traceID != "" {
					break
				}
			}
		}

		// Generate new trace ID if not found
		if traceID == "" {
			traceID = m.config.Generator()
		}

		// Clean and validate trace ID
		traceID = strings.TrimSpace(traceID)
		if traceID == "" {
			// Try the configured generator first
			traceID = m.config.Generator()
			// If still empty, use default generator as fallback
			if traceID == "" {
				traceID = generateULID()
			}
		}

		// Set trace ID in request and response headers
		r.Header.Set(m.config.HeaderName, traceID)
		w.Header().Set(m.config.HeaderName, traceID)

		// Add trace ID to request context
		ctx := context.WithValue(r.Context(), m.config.ContextKey, traceID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetTraceIDFromContext extracts trace ID from context.
func GetTraceIDFromContext(ctx context.Context, contextKey string) string {
	if contextKey == "" {
		contextKey = "trace_id"
	}

	if traceID, ok := ctx.Value(contextKey).(string); ok {
		return traceID
	}
	return ""
}

// GetTraceIDFromRequest extracts trace ID from request context or headers.
func GetTraceIDFromRequest(r *http.Request, config Config) string {
	// Try context first
	if traceID := GetTraceIDFromContext(r.Context(), config.ContextKey); traceID != "" {
		return traceID
	}

	// Try headers
	if traceID := r.Header.Get(config.HeaderName); traceID != "" {
		return traceID
	}

	// Try alternative headers
	for _, altHeader := range config.AlternativeHeaders {
		if traceID := r.Header.Get(altHeader); traceID != "" {
			return traceID
		}
	}

	return ""
}
