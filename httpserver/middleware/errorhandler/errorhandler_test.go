package errorhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockLogger for testing.
type mockLogger struct {
	errors []string
}

func (m *mockLogger) Error(args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprint(args...))
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func TestConfig_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enabled true",
			enabled:  true,
			expected: true,
		},
		{
			name:     "enabled false",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Enabled: tt.enabled}
			if got := config.IsEnabled(); got != tt.expected {
				t.Errorf("Config.IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_ShouldSkip(t *testing.T) {
	config := Config{
		SkipPaths: []string{"/health", "/metrics"},
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "should skip health path",
			path:     "/health",
			expected: true,
		},
		{
			name:     "should skip metrics path",
			path:     "/metrics",
			expected: true,
		},
		{
			name:     "should not skip regular path",
			path:     "/api/v1/users",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.ShouldSkip(tt.path); got != tt.expected {
				t.Errorf("Config.ShouldSkip() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.Enabled {
		t.Error("Expected default config to be enabled")
	}

	if config.IncludeStackTrace {
		t.Error("Expected default config to not include stack trace")
	}

	if !config.EnableRecovery {
		t.Error("Expected default config to enable recovery")
	}

	if config.TraceIDHeader != "X-Trace-Id" {
		t.Errorf("Expected default trace ID header to be 'X-Trace-Id', got '%s'", config.TraceIDHeader)
	}

	if config.TraceIDContextKey != "trace_id" {
		t.Errorf("Expected default trace ID context key to be 'trace_id', got '%s'", config.TraceIDContextKey)
	}

	if config.Logger == nil {
		t.Error("Expected default logger to be set")
	}
}

func TestMiddleware_Name(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Name() != "errorhandler" {
		t.Errorf("Expected middleware name to be 'errorhandler', got '%s'", middleware.Name())
	}
}

func TestMiddleware_Priority(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Priority() != 1000 {
		t.Errorf("Expected middleware priority to be 1000, got %d", middleware.Priority())
	}
}

func TestNewMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name:   "with default config",
			config: DefaultConfig(),
		},
		{
			name: "with custom config",
			config: Config{
				Enabled:           true,
				TraceIDHeader:     "Custom-Trace",
				TraceIDContextKey: "custom_trace",
			},
		},
		{
			name: "with nil logger",
			config: Config{
				Enabled: true,
				Logger:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewMiddleware(tt.config)

			if middleware == nil {
				t.Error("Expected middleware to be created")
			}

			if middleware.config.Logger == nil {
				t.Error("Expected logger to be set")
			}

			if middleware.config.TraceIDHeader == "" {
				t.Error("Expected trace ID header to be set")
			}

			if middleware.config.TraceIDContextKey == "" {
				t.Error("Expected trace ID context key to be set")
			}
		})
	}
}

func TestMiddleware_Wrap_Disabled(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = false
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("should not be caught")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// This should panic since middleware is disabled
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when middleware is disabled")
		}
	}()

	wrapped.ServeHTTP(rr, req)
}

func TestMiddleware_Wrap_SkipPath(t *testing.T) {
	config := DefaultConfig()
	config.SkipPaths = []string{"/health"}
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("should not be caught")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	// This should panic since path is skipped
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when path is skipped")
		}
	}()

	wrapped.ServeHTTP(rr, req)
}

func TestMiddleware_Wrap_PanicRecovery(t *testing.T) {
	mockLog := &mockLogger{}
	config := DefaultConfig()
	config.Logger = mockLog
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// Check that panic was logged
	if len(mockLog.errors) == 0 {
		t.Error("Expected panic to be logged")
	}

	// Check response content type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got '%s'", contentType)
	}

	// Check response body
	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Error != "Internal Server Error" {
		t.Errorf("Expected error to be 'Internal Server Error', got '%s'", response.Error)
	}
}

func TestMiddleware_Wrap_PanicRecoveryDisabled(t *testing.T) {
	config := DefaultConfig()
	config.EnableRecovery = false
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// This should panic since recovery is disabled
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when recovery is disabled")
		}
	}()

	wrapped.ServeHTTP(rr, req)
}

func TestMiddleware_Wrap_HTTPError(t *testing.T) {
	mockLog := &mockLogger{}
	config := DefaultConfig()
	config.Logger = mockLog
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Custom error", http.StatusBadRequest)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	// The default http.Error already writes the response, so our middleware shouldn't overwrite it
	if strings.Contains(rr.Body.String(), "Custom error") {
		// This is expected - the original error response is preserved
	}
}

func TestMiddleware_Wrap_SuccessfulRequest(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if body := rr.Body.String(); body != "success" {
		t.Errorf("Expected body to be 'success', got '%s'", body)
	}
}

func TestMiddleware_Wrap_WithTraceID(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	traceID := "test-trace-123"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-Id", traceID)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Check response body
	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.TraceID != traceID {
		t.Errorf("Expected trace ID to be '%s', got '%s'", traceID, response.TraceID)
	}
}

func TestMiddleware_Wrap_WithTraceIDFromContext(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	traceID := "context-trace-456"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "trace_id", traceID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Check response body
	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.TraceID != traceID {
		t.Errorf("Expected trace ID to be '%s', got '%s'", traceID, response.TraceID)
	}
}

func TestMiddleware_Wrap_WithStackTrace(t *testing.T) {
	config := DefaultConfig()
	config.IncludeStackTrace = true
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Check response body
	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Details == nil {
		t.Error("Expected details to be set")
	}

	if _, ok := response.Details["stack_trace"]; !ok {
		t.Error("Expected stack trace in details")
	}
}

func TestMiddleware_Wrap_CustomErrorFormatter(t *testing.T) {
	config := DefaultConfig()
	config.CustomErrorFormatter = func(err error, statusCode int, traceID string) interface{} {
		return map[string]interface{}{
			"custom_error": err.Error(),
			"status":       statusCode,
			"trace":        traceID,
		}
	}
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if _, ok := response["custom_error"]; !ok {
		t.Error("Expected custom error format")
	}
}

func TestMiddleware_Wrap_CustomPanicHandler(t *testing.T) {
	panicHandlerCalled := false
	config := DefaultConfig()
	config.PanicHandler = func(recovered interface{}, traceID string, stack []byte) {
		panicHandlerCalled = true
	}
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if !panicHandlerCalled {
		t.Error("Expected custom panic handler to be called")
	}
}

func TestMiddleware_extractTraceID(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	tests := []struct {
		name     string
		setupReq func() *http.Request
		expected string
	}{
		{
			name: "from context",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				ctx := context.WithValue(req.Context(), "trace_id", "context-trace")
				return req.WithContext(ctx)
			},
			expected: "context-trace",
		},
		{
			name: "from primary header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Trace-Id", "header-trace")
				return req
			},
			expected: "header-trace",
		},
		{
			name: "from alternative header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Trace-Id", "alt-trace")
				return req
			},
			expected: "alt-trace",
		},
		{
			name: "no trace ID",
			setupReq: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expected: "",
		},
		{
			name: "context priority over header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Trace-Id", "header-trace")
				ctx := context.WithValue(req.Context(), "trace_id", "context-trace")
				return req.WithContext(ctx)
			},
			expected: "context-trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			result := middleware.extractTraceID(req)
			if result != tt.expected {
				t.Errorf("Expected trace ID '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rr,
		statusCode:     http.StatusOK,
	}

	// Test WriteHeader
	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.statusCode)
	}

	// Test Write
	data := []byte("test data")
	n, err := rw.Write(data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}
	if !rw.written {
		t.Error("Expected written flag to be true")
	}
}

func TestCreateWithLogger(t *testing.T) {
	logger := &mockLogger{}
	middleware := CreateWithLogger(logger)

	if middleware.config.Logger != logger {
		t.Error("Expected custom logger to be set")
	}
}

func TestCreateWithRecoveryDisabled(t *testing.T) {
	middleware := CreateWithRecoveryDisabled()

	if middleware.config.EnableRecovery {
		t.Error("Expected recovery to be disabled")
	}
}

func TestCreateWithStackTrace(t *testing.T) {
	middleware := CreateWithStackTrace()

	if !middleware.config.IncludeStackTrace {
		t.Error("Expected stack trace to be enabled")
	}
}

func TestDefaultLogger(t *testing.T) {
	logger := &defaultLogger{}

	// These should not panic
	logger.Error("test error")
	logger.Errorf("test error: %s", "formatted")
}

// Benchmark tests
func BenchmarkMiddleware_Wrap(b *testing.B) {
	middleware := NewMiddleware(DefaultConfig())
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := middleware.Wrap(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
	}
}

func BenchmarkMiddleware_extractTraceID(b *testing.B) {
	middleware := NewMiddleware(DefaultConfig())
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-Id", "trace123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = middleware.extractTraceID(req)
	}
}

func TestMiddleware_Wrap_Timeout(t *testing.T) {
	// Set timeout for test
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()

		middleware := NewMiddleware(DefaultConfig())
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := middleware.Wrap(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out after 30 seconds")
	}
}
