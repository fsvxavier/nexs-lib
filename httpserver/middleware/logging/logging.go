// Package logging provides request/response logging middleware implementation.
package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// Context key for storing correlation ID
const correlationIDKey = "logging:correlationID"

// Config represents logging configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass logging.
	SkipPaths []string
	// Headers contains headers to log.
	Headers map[string]bool
	// IncludeRequestBody indicates whether to log request body.
	IncludeRequestBody bool
	// IncludeResponseBody indicates whether to log response body.
	IncludeResponseBody bool
	// MaxBodySize is the maximum size of body to log (in bytes).
	MaxBodySize int64
	// Logger is the function to log request/response information.
	Logger func(LogEntry)
	// CorrelationIDHeader is the header name for correlation ID.
	CorrelationIDHeader string
	// GenerateCorrelationID generates a correlation ID if not present.
	GenerateCorrelationID bool
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

// LogEntry represents a log entry for a request/response.
type LogEntry struct {
	CorrelationID   string            `json:"correlation_id"`
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	Query           string            `json:"query,omitempty"`
	RemoteAddr      string            `json:"remote_addr"`
	UserAgent       string            `json:"user_agent,omitempty"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty"`
	RequestBody     string            `json:"request_body,omitempty"`
	StatusCode      int               `json:"status_code"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	ResponseBody    string            `json:"response_body,omitempty"`
	ResponseSize    int               `json:"response_size"`
	Duration        time.Duration     `json:"duration"`
	Timestamp       time.Time         `json:"timestamp"`
	Error           string            `json:"error,omitempty"`
}

// DefaultConfig returns a default logging configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:               true,
		Headers:               make(map[string]bool),
		IncludeRequestBody:    false,
		IncludeResponseBody:   false,
		MaxBodySize:           64 * 1024, // 64KB
		Logger:                defaultLogger,
		CorrelationIDHeader:   "X-Correlation-ID",
		GenerateCorrelationID: true,
	}
}

// Middleware implements logging middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new logging middleware.
func NewMiddleware(config Config) *Middleware {
	if config.Logger == nil {
		config.Logger = defaultLogger
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

		start := time.Now()

		// Handle correlation ID
		correlationID := r.Header.Get(m.config.CorrelationIDHeader)
		if correlationID == "" && m.config.GenerateCorrelationID {
			correlationID = generateCorrelationID()
		}

		// Add correlation ID to response header
		if correlationID != "" {
			w.Header().Set(m.config.CorrelationIDHeader, correlationID)
			// Add to request context
			ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)
			r = r.WithContext(ctx)
		}

		// Capture request data
		entry := LogEntry{
			CorrelationID: correlationID,
			Method:        r.Method,
			Path:          r.URL.Path,
			Query:         r.URL.RawQuery,
			RemoteAddr:    getClientIP(r),
			UserAgent:     r.Header.Get("User-Agent"),
			Timestamp:     start,
		}

		// Capture request headers if needed
		if len(m.config.Headers) > 0 {
			entry.RequestHeaders = make(map[string]string)
			for header := range m.config.Headers {
				if value := r.Header.Get(header); value != "" {
					entry.RequestHeaders[header] = value
				}
			}
		}

		// Wrap response writer to capture response data
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Complete log entry
		entry.StatusCode = rw.statusCode
		entry.ResponseSize = rw.size
		entry.Duration = time.Since(start)

		// Capture response headers if needed
		if len(m.config.Headers) > 0 {
			entry.ResponseHeaders = make(map[string]string)
			for header := range m.config.Headers {
				if value := w.Header().Get(header); value != "" {
					entry.ResponseHeaders[header] = value
				}
			}
		}

		// Log the entry
		m.config.Logger(entry)
	})
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "logging"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 50 // Logging should happen relatively early
}

// responseWriter wraps http.ResponseWriter to capture response data.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// generateCorrelationID generates a random correlation ID.
func generateCorrelationID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(bytes)
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Try to get real IP from headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// defaultLogger is a no-op logger.
func defaultLogger(entry LogEntry) {
	// Default implementation does nothing
	// Users should provide their own logger
}
