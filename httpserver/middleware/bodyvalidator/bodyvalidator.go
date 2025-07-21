// Package bodyvalidator provides HTTP body validation middleware implementation.
package bodyvalidator

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// Config represents body validator configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass body validation.
	SkipPaths []string
	// AllowedContentTypes contains allowed content types.
	AllowedContentTypes []string
	// RequireJSON indicates if JSON validation is required.
	RequireJSON bool
	// MaxBodySize limits the maximum body size in bytes (0 means no limit).
	MaxBodySize int64
	// SkipMethods contains HTTP methods that should skip validation.
	SkipMethods []string
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

// ShouldSkipMethod returns true if the given method should be skipped.
func (c Config) ShouldSkipMethod(method string) bool {
	for _, skipMethod := range c.SkipMethods {
		if strings.EqualFold(method, skipMethod) {
			return true
		}
	}
	return false
}

// DefaultConfig returns a default body validator configuration.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		AllowedContentTypes: []string{
			"application/json",
			"application/json; charset=utf-8",
		},
		RequireJSON: true,
		MaxBodySize: 1024 * 1024, // 1MB
		SkipMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
			http.MethodDelete,
		},
	}
}

// Middleware implements body validation middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new body validator middleware.
func NewMiddleware(config Config) *Middleware {
	return &Middleware{
		config: config,
	}
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "bodyvalidator"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 200
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

		if m.config.ShouldSkipMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		// Check content type for methods that have body
		contentType := r.Header.Get("Content-Type")
		if contentType != "" && len(m.config.AllowedContentTypes) > 0 {
			allowed := false
			for _, allowedType := range m.config.AllowedContentTypes {
				if strings.Contains(strings.ToLower(contentType), strings.ToLower(allowedType)) {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
				return
			}
		}

		// Read and validate body if required
		if m.config.RequireJSON && r.ContentLength > 0 {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Bad Request: Unable to read body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			// Check max body size
			if m.config.MaxBodySize > 0 && int64(len(body)) > m.config.MaxBodySize {
				http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
				return
			}

			// Validate JSON if body is not empty
			if len(body) > 0 {
				if !json.Valid(body) {
					http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
					return
				}

				sBody := strings.TrimSpace(string(body))
				if !strings.HasPrefix(sBody, "{") || !strings.HasSuffix(sBody, "}") {
					http.Error(w, "Bad Request: JSON must be an object", http.StatusBadRequest)
					return
				}
			}

			// Restore body for next handler
			r.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		next.ServeHTTP(w, r)
	})
}
