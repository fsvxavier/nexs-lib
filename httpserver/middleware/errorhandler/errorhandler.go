// Package errorhandler provides error handling middleware implementation.
package errorhandler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	Code      string                 `json:"code,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Logger interface for custom logging.
type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// Config represents error handler configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass error handling.
	SkipPaths []string
	// IncludeStackTrace indicates if stack traces should be included in responses.
	IncludeStackTrace bool
	// Logger is the logger instance to use for error logging.
	Logger Logger
	// TraceIDHeader is the header name to extract trace ID from.
	TraceIDHeader string
	// TraceIDContextKey is the context key to extract trace ID from.
	TraceIDContextKey string
	// EnableRecovery indicates if panic recovery should be enabled.
	EnableRecovery bool
	// CustomErrorFormatter allows custom error response formatting.
	CustomErrorFormatter func(err error, statusCode int, traceID string) interface{}
	// PanicHandler is called when a panic occurs (if recovery is enabled).
	PanicHandler func(recovered interface{}, traceID string, stack []byte)
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

// DefaultConfig returns a default error handler configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:              true,
		IncludeStackTrace:    false,
		TraceIDHeader:        "X-Trace-Id",
		TraceIDContextKey:    "trace_id",
		EnableRecovery:       true,
		Logger:               &defaultLogger{},
		CustomErrorFormatter: nil,
		PanicHandler:         nil,
	}
}

// defaultLogger is a simple logger implementation.
type defaultLogger struct{}

func (l *defaultLogger) Error(args ...interface{}) {
	log.Println(args...)
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// Middleware implements error handling middleware.
type Middleware struct {
	config Config
}

// NewMiddleware creates a new error handler middleware.
func NewMiddleware(config Config) *Middleware {
	if config.Logger == nil {
		config.Logger = &defaultLogger{}
	}
	if config.TraceIDHeader == "" {
		config.TraceIDHeader = "X-Trace-Id"
	}
	if config.TraceIDContextKey == "" {
		config.TraceIDContextKey = "trace_id"
	}

	return &Middleware{
		config: config,
	}
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "errorhandler"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 1000 // Very low priority, should wrap everything
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

		// Setup panic recovery if enabled
		if m.config.EnableRecovery {
			defer func() {
				if recovered := recover(); recovered != nil {
					m.handlePanic(w, r, recovered)
				}
			}()
		}

		// Wrap response writer to capture status code
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter, r)

		// Handle HTTP error status codes
		if wrappedWriter.statusCode >= 400 {
			m.handleHTTPError(wrappedWriter, r, wrappedWriter.statusCode)
		}
	})
}

// handlePanic handles panics that occur during request processing.
func (m *Middleware) handlePanic(w http.ResponseWriter, r *http.Request, recovered interface{}) {
	stack := debug.Stack()
	traceID := m.extractTraceID(r)

	// Log the panic
	m.config.Logger.Errorf("Panic recovered: %v\nTrace ID: %s\nPath: %s\nStack: %s",
		recovered, traceID, r.URL.Path, string(stack))

	// Call custom panic handler if configured
	if m.config.PanicHandler != nil {
		m.config.PanicHandler(recovered, traceID, stack)
	}

	// Create error response
	var response interface{}
	if m.config.CustomErrorFormatter != nil {
		response = m.config.CustomErrorFormatter(
			fmt.Errorf("internal server error: %v", recovered),
			http.StatusInternalServerError,
			traceID,
		)
	} else {
		response = m.createErrorResponse(
			fmt.Errorf("internal server error"),
			http.StatusInternalServerError,
			traceID,
			stack,
		)
	}

	// Send error response
	m.sendErrorResponse(w, http.StatusInternalServerError, response)
}

// handleHTTPError handles HTTP error status codes.
func (m *Middleware) handleHTTPError(w *responseWriter, r *http.Request, statusCode int) {
	// Only handle if response body is empty (no custom error response was set)
	if w.written {
		return
	}

	traceID := m.extractTraceID(r)
	err := fmt.Errorf("HTTP %d: %s", statusCode, http.StatusText(statusCode))

	// Log the error
	m.config.Logger.Errorf("HTTP Error: %d - %s - Trace ID: %s - Path: %s",
		statusCode, http.StatusText(statusCode), traceID, r.URL.Path)

	// Create error response
	var response interface{}
	if m.config.CustomErrorFormatter != nil {
		response = m.config.CustomErrorFormatter(err, statusCode, traceID)
	} else {
		response = m.createErrorResponse(err, statusCode, traceID, nil)
	}

	// Send error response
	m.sendErrorResponse(w, statusCode, response)
}

// createErrorResponse creates a standardized error response.
func (m *Middleware) createErrorResponse(err error, statusCode int, traceID string, stack []byte) ErrorResponse {
	response := ErrorResponse{
		Error:     http.StatusText(statusCode),
		Message:   err.Error(),
		Code:      fmt.Sprintf("%d", statusCode),
		Timestamp: time.Now().UTC(),
		TraceID:   traceID,
	}

	// Add stack trace if enabled and available
	if m.config.IncludeStackTrace && stack != nil {
		if response.Details == nil {
			response.Details = make(map[string]interface{})
		}
		response.Details["stack_trace"] = string(stack)
	}

	return response
}

// sendErrorResponse sends an error response as JSON.
func (m *Middleware) sendErrorResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		m.config.Logger.Errorf("Failed to encode error response: %v", err)
		// Fallback to plain text
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Internal Server Error")
	}
}

// extractTraceID extracts trace ID from request headers or context.
func (m *Middleware) extractTraceID(r *http.Request) string {
	// Try context first
	if traceID, ok := r.Context().Value(m.config.TraceIDContextKey).(string); ok && traceID != "" {
		return traceID
	}

	// Try header
	if traceID := r.Header.Get(m.config.TraceIDHeader); traceID != "" {
		return traceID
	}

	// Try alternative headers
	alternatives := []string{
		"Trace-Id", "trace-id", "Trace-ID", "X-Trace-ID", "x-trace-id",
	}

	for _, header := range alternatives {
		if traceID := r.Header.Get(header); traceID != "" {
			return traceID
		}
	}

	return ""
}

// responseWriter wraps http.ResponseWriter to capture status code and write status.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.written = true
	return rw.ResponseWriter.Write(b)
}

// CreateWithLogger creates an error handler middleware with a custom logger.
func CreateWithLogger(logger Logger) *Middleware {
	config := DefaultConfig()
	config.Logger = logger
	return NewMiddleware(config)
}

// CreateWithRecoveryDisabled creates an error handler middleware with panic recovery disabled.
func CreateWithRecoveryDisabled() *Middleware {
	config := DefaultConfig()
	config.EnableRecovery = false
	return NewMiddleware(config)
}

// CreateWithStackTrace creates an error handler middleware that includes stack traces.
func CreateWithStackTrace() *Middleware {
	config := DefaultConfig()
	config.IncludeStackTrace = true
	return NewMiddleware(config)
}
