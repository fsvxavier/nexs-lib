package bodyvalidator

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestConfig_ShouldSkipMethod(t *testing.T) {
	config := Config{
		SkipMethods: []string{"GET", "HEAD", "OPTIONS"},
	}

	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{
			name:     "should skip GET method",
			method:   "GET",
			expected: true,
		},
		{
			name:     "should skip HEAD method",
			method:   "HEAD",
			expected: true,
		},
		{
			name:     "should skip OPTIONS method",
			method:   "OPTIONS",
			expected: true,
		},
		{
			name:     "should not skip POST method",
			method:   "POST",
			expected: false,
		},
		{
			name:     "should skip case insensitive",
			method:   "get",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.ShouldSkipMethod(tt.method); got != tt.expected {
				t.Errorf("Config.ShouldSkipMethod() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.Enabled {
		t.Error("Expected default config to be enabled")
	}

	if !config.RequireJSON {
		t.Error("Expected default config to require JSON")
	}

	if config.MaxBodySize != 1024*1024 {
		t.Errorf("Expected default max body size to be 1MB, got %d", config.MaxBodySize)
	}

	expectedContentTypes := []string{
		"application/json",
		"application/json; charset=utf-8",
	}

	if len(config.AllowedContentTypes) != len(expectedContentTypes) {
		t.Errorf("Expected %d content types, got %d", len(expectedContentTypes), len(config.AllowedContentTypes))
	}
}

func TestMiddleware_Name(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Name() != "bodyvalidator" {
		t.Errorf("Expected middleware name to be 'bodyvalidator', got '%s'", middleware.Name())
	}
}

func TestMiddleware_Priority(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Priority() != 200 {
		t.Errorf("Expected middleware priority to be 200, got %d", middleware.Priority())
	}
}

func TestMiddleware_Wrap_Disabled(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = false
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "value"}`))
	req.Header.Set("Content-Type", "application/json")
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/health", strings.NewReader(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMiddleware_Wrap_SkipMethod(t *testing.T) {
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
}

func TestMiddleware_Wrap_InvalidContentType(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "value"}`))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status %d, got %d", http.StatusUnsupportedMediaType, rr.Code)
	}
}

func TestMiddleware_Wrap_ValidJSON(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"test": "value"}` {
			t.Error("Body was not properly restored")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "value"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMiddleware_Wrap_InvalidJSON(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMiddleware_Wrap_NonObjectJSON(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`["array", "not", "object"]`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMiddleware_Wrap_MaxBodySizeExceeded(t *testing.T) {
	config := DefaultConfig()
	config.MaxBodySize = 10 // Very small limit
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	largeBody := strings.Repeat("a", 20)
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"data": "`+largeBody+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status %d, got %d", http.StatusRequestEntityTooLarge, rr.Code)
	}
}

func TestMiddleware_Wrap_EmptyBody(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/test", &bytes.Buffer{})
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestNewMiddleware(t *testing.T) {
	config := DefaultConfig()
	middleware := NewMiddleware(config)

	if middleware == nil {
		t.Error("Expected middleware to be created")
	}

	if middleware.config.Enabled != config.Enabled {
		t.Error("Expected config to be set correctly")
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
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "value"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
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

		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "value"}`))
		req.Header.Set("Content-Type", "application/json")
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
