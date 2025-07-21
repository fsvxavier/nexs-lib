package traceid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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

func TestGenerateULID(t *testing.T) {
	traceID := generateULID()

	if traceID == "" {
		t.Error("Expected trace ID to be generated")
	}

	// Should be 32 characters for hex representation of 16 bytes
	if len(traceID) != 32 {
		t.Errorf("Expected trace ID length to be 32, got %d", len(traceID))
	}

	// Should be different each time
	traceID2 := generateULID()
	if traceID == traceID2 {
		t.Error("Expected different trace IDs")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.Enabled {
		t.Error("Expected default config to be enabled")
	}

	if config.HeaderName != "X-Trace-Id" {
		t.Errorf("Expected default header name to be 'X-Trace-Id', got '%s'", config.HeaderName)
	}

	if config.ContextKey != "trace_id" {
		t.Errorf("Expected default context key to be 'trace_id', got '%s'", config.ContextKey)
	}

	if config.Generator == nil {
		t.Error("Expected default generator to be set")
	}

	expectedAltHeaders := []string{
		"Trace-Id",
		"trace-id",
		"Trace-ID",
		"X-Trace-ID",
		"x-trace-id",
	}

	if len(config.AlternativeHeaders) != len(expectedAltHeaders) {
		t.Errorf("Expected %d alternative headers, got %d", len(expectedAltHeaders), len(config.AlternativeHeaders))
	}
}

func TestMiddleware_Name(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Name() != "traceid" {
		t.Errorf("Expected middleware name to be 'traceid', got '%s'", middleware.Name())
	}
}

func TestMiddleware_Priority(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Priority() != 50 {
		t.Errorf("Expected middleware priority to be 50, got %d", middleware.Priority())
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
				Enabled:    true,
				HeaderName: "Custom-Trace",
				ContextKey: "custom_trace",
			},
		},
		{
			name: "with nil generator",
			config: Config{
				Enabled:   true,
				Generator: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewMiddleware(tt.config)

			if middleware == nil {
				t.Error("Expected middleware to be created")
			}

			if middleware.config.Generator == nil {
				t.Error("Expected generator to be set")
			}

			if middleware.config.HeaderName == "" {
				t.Error("Expected header name to be set")
			}

			if middleware.config.ContextKey == "" {
				t.Error("Expected context key to be set")
			}
		})
	}
}

func TestMiddleware_Wrap_Disabled(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = false
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := GetTraceIDFromRequest(r, config)
		if traceID != "" {
			t.Error("Expected no trace ID when disabled")
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMiddleware_Wrap_SkipPath(t *testing.T) {
	config := DefaultConfig()
	config.SkipPaths = []string{"/health"}
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := GetTraceIDFromRequest(r, config)
		if traceID != "" {
			t.Error("Expected no trace ID for skipped path")
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMiddleware_Wrap_GenerateNewTraceID(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	var capturedTraceID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = GetTraceIDFromContext(r.Context(), "trace_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if capturedTraceID == "" {
		t.Error("Expected trace ID to be generated")
	}

	// Check response header
	responseTraceID := rr.Header().Get("X-Trace-Id")
	if responseTraceID != capturedTraceID {
		t.Errorf("Expected response trace ID to match context, got '%s' vs '%s'", responseTraceID, capturedTraceID)
	}
}

func TestMiddleware_Wrap_UseExistingTraceID(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	existingTraceID := "existing-trace-id-123"
	var capturedTraceID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = GetTraceIDFromContext(r.Context(), "trace_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-Id", existingTraceID)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if capturedTraceID != existingTraceID {
		t.Errorf("Expected trace ID to be '%s', got '%s'", existingTraceID, capturedTraceID)
	}

	// Check response header
	responseTraceID := rr.Header().Get("X-Trace-Id")
	if responseTraceID != existingTraceID {
		t.Errorf("Expected response trace ID to be '%s', got '%s'", existingTraceID, responseTraceID)
	}
}

func TestMiddleware_Wrap_UseAlternativeHeader(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	existingTraceID := "alt-trace-id-456"
	var capturedTraceID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = GetTraceIDFromContext(r.Context(), "trace_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Trace-Id", existingTraceID) // Alternative header
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if capturedTraceID != existingTraceID {
		t.Errorf("Expected trace ID to be '%s', got '%s'", existingTraceID, capturedTraceID)
	}
}

func TestMiddleware_Wrap_EmptyTraceIDGenerated(t *testing.T) {
	config := DefaultConfig()
	config.Generator = func() string { return "" } // Always return empty
	middleware := NewMiddleware(config)

	var capturedTraceID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = GetTraceIDFromContext(r.Context(), "trace_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-Id", "   ") // Whitespace only
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Should fall back to default generator
	if capturedTraceID == "" {
		t.Error("Expected trace ID to be generated by fallback")
	}
}

func TestGetTraceIDFromContext(t *testing.T) {
	ctx := context.Background()
	traceID := "test-trace-id"

	// Test with trace ID in context
	ctx = context.WithValue(ctx, "trace_id", traceID)
	result := GetTraceIDFromContext(ctx, "trace_id")
	if result != traceID {
		t.Errorf("Expected trace ID '%s', got '%s'", traceID, result)
	}

	// Test with empty context
	emptyCtx := context.Background()
	result = GetTraceIDFromContext(emptyCtx, "trace_id")
	if result != "" {
		t.Errorf("Expected empty trace ID, got '%s'", result)
	}

	// Test with default context key
	result = GetTraceIDFromContext(ctx, "")
	if result != traceID {
		t.Errorf("Expected trace ID '%s' with default key, got '%s'", traceID, result)
	}
}

func TestGetTraceIDFromRequest(t *testing.T) {
	config := DefaultConfig()
	traceID := "test-trace-id"

	// Test with trace ID in context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "trace_id", traceID)
	req = req.WithContext(ctx)

	result := GetTraceIDFromRequest(req, config)
	if result != traceID {
		t.Errorf("Expected trace ID '%s' from context, got '%s'", traceID, result)
	}

	// Test with trace ID in header
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Trace-Id", traceID)

	result = GetTraceIDFromRequest(req2, config)
	if result != traceID {
		t.Errorf("Expected trace ID '%s' from header, got '%s'", traceID, result)
	}

	// Test with trace ID in alternative header
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("Trace-Id", traceID)

	result = GetTraceIDFromRequest(req3, config)
	if result != traceID {
		t.Errorf("Expected trace ID '%s' from alternative header, got '%s'", traceID, result)
	}

	// Test with no trace ID
	req4 := httptest.NewRequest("GET", "/test", nil)
	result = GetTraceIDFromRequest(req4, config)
	if result != "" {
		t.Errorf("Expected empty trace ID, got '%s'", result)
	}
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

func BenchmarkGenerateULID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateULID()
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
