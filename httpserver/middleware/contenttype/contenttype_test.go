package contenttype

import (
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

func TestConfig_ShouldValidateMethod(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		method   string
		expected bool
	}{
		{
			name: "method in restrict list",
			config: Config{
				RestrictMethods: []string{"POST", "PUT"},
			},
			method:   "POST",
			expected: true,
		},
		{
			name: "method not in restrict list",
			config: Config{
				RestrictMethods: []string{"POST", "PUT"},
			},
			method:   "GET",
			expected: false,
		},
		{
			name: "method in ignore list",
			config: Config{
				IgnoreMethods: []string{"GET", "HEAD"},
			},
			method:   "GET",
			expected: false,
		},
		{
			name: "empty restrict methods validates all except ignored",
			config: Config{
				RestrictMethods: []string{},
				IgnoreMethods:   []string{"GET"},
			},
			method:   "POST",
			expected: true,
		},
		{
			name: "case insensitive method matching",
			config: Config{
				RestrictMethods: []string{"POST"},
			},
			method:   "post",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.ShouldValidateMethod(tt.method); got != tt.expected {
				t.Errorf("Config.ShouldValidateMethod() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.Enabled {
		t.Error("Expected default config to be enabled")
	}

	if !config.RequireContentType {
		t.Error("Expected default config to require content type")
	}

	if config.CaseSensitive {
		t.Error("Expected default config to be case insensitive")
	}

	if config.StrictMatching {
		t.Error("Expected default config to not use strict matching")
	}

	expectedContentTypes := []string{
		"application/json",
		"application/json; charset=utf-8",
		"application/xml",
		"application/xml; charset=utf-8",
		"text/xml",
		"text/xml; charset=utf-8",
	}

	if len(config.AllowedContentTypes) != len(expectedContentTypes) {
		t.Errorf("Expected %d content types, got %d", len(expectedContentTypes), len(config.AllowedContentTypes))
	}

	expectedRestrictMethods := []string{"POST", "PUT", "PATCH"}
	if len(config.RestrictMethods) != len(expectedRestrictMethods) {
		t.Errorf("Expected %d restrict methods, got %d", len(expectedRestrictMethods), len(config.RestrictMethods))
	}

	expectedIgnoreMethods := []string{"GET", "HEAD", "OPTIONS", "DELETE"}
	if len(config.IgnoreMethods) != len(expectedIgnoreMethods) {
		t.Errorf("Expected %d ignore methods, got %d", len(expectedIgnoreMethods), len(config.IgnoreMethods))
	}
}

func TestMiddleware_Name(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Name() != "contenttype" {
		t.Errorf("Expected middleware name to be 'contenttype', got '%s'", middleware.Name())
	}
}

func TestMiddleware_Priority(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Priority() != 150 {
		t.Errorf("Expected middleware priority to be 150, got %d", middleware.Priority())
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
	req.Header.Set("Content-Type", "text/plain") // Invalid content type
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

	req := httptest.NewRequest("POST", "/health", strings.NewReader(`{"test": "value"}`))
	req.Header.Set("Content-Type", "text/plain") // Invalid content type
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMiddleware_Wrap_IgnoreMethod(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil) // GET is in ignore methods
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMiddleware_Wrap_ValidContentType(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

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

func TestMiddleware_Wrap_MissingContentType(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "value"}`))
	// No Content-Type header
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMiddleware_Wrap_ContentTypeNotRequired(t *testing.T) {
	config := DefaultConfig()
	config.RequireContentType = false
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "value"}`))
	// No Content-Type header
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMiddleware_isAllowedContentType(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		contentType string
		expected    bool
	}{
		{
			name: "exact match",
			config: Config{
				AllowedContentTypes: []string{"application/json"},
				StrictMatching:      true,
				CaseSensitive:       false,
			},
			contentType: "application/json",
			expected:    true,
		},
		{
			name: "partial match with charset",
			config: Config{
				AllowedContentTypes: []string{"application/json"},
				StrictMatching:      false,
				CaseSensitive:       false,
			},
			contentType: "application/json; charset=utf-8",
			expected:    true,
		},
		{
			name: "case insensitive match",
			config: Config{
				AllowedContentTypes: []string{"APPLICATION/JSON"},
				StrictMatching:      false,
				CaseSensitive:       false,
			},
			contentType: "application/json",
			expected:    true,
		},
		{
			name: "case sensitive no match",
			config: Config{
				AllowedContentTypes: []string{"APPLICATION/JSON"},
				StrictMatching:      false,
				CaseSensitive:       true,
			},
			contentType: "application/json",
			expected:    false,
		},
		{
			name: "strict matching no match with charset",
			config: Config{
				AllowedContentTypes: []string{"application/json"},
				StrictMatching:      true,
				CaseSensitive:       false,
			},
			contentType: "application/json; charset=utf-8",
			expected:    false,
		},
		{
			name: "empty allowed types allows all",
			config: Config{
				AllowedContentTypes: []string{},
				StrictMatching:      false,
				CaseSensitive:       false,
			},
			contentType: "text/plain",
			expected:    true,
		},
		{
			name: "whitespace handling",
			config: Config{
				AllowedContentTypes: []string{"  application/json  "},
				StrictMatching:      true,
				CaseSensitive:       false,
			},
			contentType: "  application/json  ",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewMiddleware(tt.config)
			result := middleware.isAllowedContentType(tt.contentType)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCreateForMethods(t *testing.T) {
	allowedTypes := []string{"application/json"}
	methods := []string{"POST", "PUT"}

	middleware := CreateForMethods(allowedTypes, methods...)

	if len(middleware.config.AllowedContentTypes) != len(allowedTypes) {
		t.Errorf("Expected %d allowed content types, got %d", len(allowedTypes), len(middleware.config.AllowedContentTypes))
	}

	if len(middleware.config.RestrictMethods) != len(methods) {
		t.Errorf("Expected %d restrict methods, got %d", len(methods), len(middleware.config.RestrictMethods))
	}

	if len(middleware.config.IgnoreMethods) != 0 {
		t.Errorf("Expected 0 ignore methods, got %d", len(middleware.config.IgnoreMethods))
	}
}

func TestCreateJSONOnly(t *testing.T) {
	methods := []string{"POST", "PUT"}
	middleware := CreateJSONOnly(methods...)

	expectedTypes := []string{
		"application/json",
		"application/json; charset=utf-8",
	}

	if len(middleware.config.AllowedContentTypes) != len(expectedTypes) {
		t.Errorf("Expected %d allowed content types, got %d", len(expectedTypes), len(middleware.config.AllowedContentTypes))
	}

	if len(middleware.config.RestrictMethods) != len(methods) {
		t.Errorf("Expected %d restrict methods, got %d", len(methods), len(middleware.config.RestrictMethods))
	}
}

func TestCreateXMLOnly(t *testing.T) {
	methods := []string{"POST", "PUT"}
	middleware := CreateXMLOnly(methods...)

	expectedTypes := []string{
		"application/xml",
		"application/xml; charset=utf-8",
		"text/xml",
		"text/xml; charset=utf-8",
	}

	if len(middleware.config.AllowedContentTypes) != len(expectedTypes) {
		t.Errorf("Expected %d allowed content types, got %d", len(expectedTypes), len(middleware.config.AllowedContentTypes))
	}

	if len(middleware.config.RestrictMethods) != len(methods) {
		t.Errorf("Expected %d restrict methods, got %d", len(methods), len(middleware.config.RestrictMethods))
	}
}

func TestMiddleware_Wrap_JSONOnlyIntegration(t *testing.T) {
	middleware := CreateJSONOnly("POST")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrapped := middleware.Wrap(handler)

	// Test valid JSON content type
	req1 := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "value"}`))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()

	wrapped.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Expected status %d for JSON, got %d", http.StatusOK, rr1.Code)
	}

	// Test invalid XML content type
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader(`<test>value</test>`))
	req2.Header.Set("Content-Type", "application/xml")
	rr2 := httptest.NewRecorder()

	wrapped.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status %d for XML, got %d", http.StatusUnsupportedMediaType, rr2.Code)
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

func BenchmarkMiddleware_isAllowedContentType(b *testing.B) {
	middleware := NewMiddleware(DefaultConfig())
	contentType := "application/json; charset=utf-8"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = middleware.isAllowedContentType(contentType)
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
