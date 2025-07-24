package tenantid

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

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.Enabled {
		t.Error("Expected default config to be enabled")
	}

	if config.HeaderName != "X-Tenant-Id" {
		t.Errorf("Expected default header name to be 'X-Tenant-Id', got '%s'", config.HeaderName)
	}

	if config.ContextKey != "tenant_id" {
		t.Errorf("Expected default context key to be 'tenant_id', got '%s'", config.ContextKey)
	}

	if config.QueryParam != "tenant_id" {
		t.Errorf("Expected default query param to be 'tenant_id', got '%s'", config.QueryParam)
	}

	if config.Required {
		t.Error("Expected default config to not require tenant ID")
	}

	if !config.CaseSensitive {
		t.Error("Expected default config to be case sensitive")
	}

	expectedAltHeaders := []string{
		"Client-Id",
		"client-id",
		"Client-ID",
		"X-Client-Id",
		"x-client-id",
		"Tenant-Id",
		"tenant-id",
	}

	if len(config.AlternativeHeaders) != len(expectedAltHeaders) {
		t.Errorf("Expected %d alternative headers, got %d", len(expectedAltHeaders), len(config.AlternativeHeaders))
	}
}

func TestMiddleware_Name(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Name() != "tenantid" {
		t.Errorf("Expected middleware name to be 'tenantid', got '%s'", middleware.Name())
	}
}

func TestMiddleware_Priority(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())
	if middleware.Priority() != 100 {
		t.Errorf("Expected middleware priority to be 100, got %d", middleware.Priority())
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
				HeaderName: "Custom-Tenant",
				ContextKey: "custom_tenant",
			},
		},
		{
			name: "with empty header name",
			config: Config{
				Enabled:    true,
				HeaderName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewMiddleware(tt.config)

			if middleware == nil {
				t.Error("Expected middleware to be created")
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
		tenantID := GetTenantIDFromContext(r.Context(), "tenant_id")
		if tenantID != "" {
			t.Error("Expected no tenant ID when disabled")
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-Id", "tenant123")
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
		tenantID := GetTenantIDFromContext(r.Context(), "tenant_id")
		if tenantID != "" {
			t.Error("Expected no tenant ID for skipped path")
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Tenant-Id", "tenant123")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestMiddleware_Wrap_ExtractFromHeader(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	expectedTenantID := "tenant123"
	var capturedTenantID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantID = GetTenantIDFromContext(r.Context(), "tenant_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-Id", expectedTenantID)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if capturedTenantID != expectedTenantID {
		t.Errorf("Expected tenant ID '%s', got '%s'", expectedTenantID, capturedTenantID)
	}
}

func TestMiddleware_Wrap_ExtractFromAlternativeHeader(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	expectedTenantID := "tenant456"
	var capturedTenantID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantID = GetTenantIDFromContext(r.Context(), "tenant_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Client-Id", expectedTenantID) // Alternative header
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if capturedTenantID != expectedTenantID {
		t.Errorf("Expected tenant ID '%s', got '%s'", expectedTenantID, capturedTenantID)
	}
}

func TestMiddleware_Wrap_ExtractFromQueryParam(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	expectedTenantID := "tenant789"
	var capturedTenantID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantID = GetTenantIDFromContext(r.Context(), "tenant_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test?tenant_id="+expectedTenantID, nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if capturedTenantID != expectedTenantID {
		t.Errorf("Expected tenant ID '%s', got '%s'", expectedTenantID, capturedTenantID)
	}
}

func TestMiddleware_Wrap_UseDefaultTenant(t *testing.T) {
	config := DefaultConfig()
	config.DefaultTenant = "default-tenant"
	middleware := NewMiddleware(config)

	var capturedTenantID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantID = GetTenantIDFromContext(r.Context(), "tenant_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if capturedTenantID != config.DefaultTenant {
		t.Errorf("Expected tenant ID '%s', got '%s'", config.DefaultTenant, capturedTenantID)
	}
}

func TestMiddleware_Wrap_RequiredTenantMissing(t *testing.T) {
	config := DefaultConfig()
	config.Required = true
	middleware := NewMiddleware(config)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when tenant ID is required but missing")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMiddleware_Wrap_CaseInsensitive(t *testing.T) {
	config := DefaultConfig()
	config.CaseSensitive = false
	middleware := NewMiddleware(config)

	var capturedTenantID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantID = GetTenantIDFromContext(r.Context(), "tenant_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-Id", "TENANT123")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	expectedTenantID := "tenant123" // Should be lowercase
	if capturedTenantID != expectedTenantID {
		t.Errorf("Expected tenant ID '%s', got '%s'", expectedTenantID, capturedTenantID)
	}
}

func TestMiddleware_Wrap_TrimWhitespace(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	var capturedTenantID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTenantID = GetTenantIDFromContext(r.Context(), "tenant_id")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-Id", "  tenant123  ")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	expectedTenantID := "tenant123"
	if capturedTenantID != expectedTenantID {
		t.Errorf("Expected tenant ID '%s', got '%s'", expectedTenantID, capturedTenantID)
	}
}

func TestGetTenantIDFromContext(t *testing.T) {
	ctx := context.Background()
	tenantID := "test-tenant-id"

	// Test with tenant ID in context
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	result := GetTenantIDFromContext(ctx, "tenant_id")
	if result != tenantID {
		t.Errorf("Expected tenant ID '%s', got '%s'", tenantID, result)
	}

	// Test with empty context
	emptyCtx := context.Background()
	result = GetTenantIDFromContext(emptyCtx, "tenant_id")
	if result != "" {
		t.Errorf("Expected empty tenant ID, got '%s'", result)
	}

	// Test with default context key
	result = GetTenantIDFromContext(ctx, "")
	if result != tenantID {
		t.Errorf("Expected tenant ID '%s' with default key, got '%s'", tenantID, result)
	}
}

func TestGetTenantIDFromRequest(t *testing.T) {
	config := DefaultConfig()
	tenantID := "test-tenant-id"

	// Test with tenant ID in context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
	req = req.WithContext(ctx)

	result := GetTenantIDFromRequest(req, config)
	if result != tenantID {
		t.Errorf("Expected tenant ID '%s' from context, got '%s'", tenantID, result)
	}

	// Test with tenant ID in header
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Tenant-Id", tenantID)

	result = GetTenantIDFromRequest(req2, config)
	if result != tenantID {
		t.Errorf("Expected tenant ID '%s' from header, got '%s'", tenantID, result)
	}

	// Test with tenant ID in alternative header
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("Client-Id", tenantID)

	result = GetTenantIDFromRequest(req3, config)
	if result != tenantID {
		t.Errorf("Expected tenant ID '%s' from alternative header, got '%s'", tenantID, result)
	}

	// Test with tenant ID in query parameter
	req4 := httptest.NewRequest("GET", "/test?tenant_id="+tenantID, nil)
	result = GetTenantIDFromRequest(req4, config)
	if result != tenantID {
		t.Errorf("Expected tenant ID '%s' from query param, got '%s'", tenantID, result)
	}

	// Test with no tenant ID
	req5 := httptest.NewRequest("GET", "/test", nil)
	result = GetTenantIDFromRequest(req5, config)
	if result != "" {
		t.Errorf("Expected empty tenant ID, got '%s'", result)
	}
}

func TestMiddleware_extractTenantID(t *testing.T) {
	middleware := NewMiddleware(DefaultConfig())

	tests := []struct {
		name     string
		setupReq func() *http.Request
		expected string
	}{
		{
			name: "from primary header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Tenant-Id", "primary-tenant")
				return req
			},
			expected: "primary-tenant",
		},
		{
			name: "from alternative header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Client-Id", "alt-tenant")
				return req
			},
			expected: "alt-tenant",
		},
		{
			name: "from query parameter",
			setupReq: func() *http.Request {
				return httptest.NewRequest("GET", "/test?tenant_id=query-tenant", nil)
			},
			expected: "query-tenant",
		},
		{
			name: "no tenant ID",
			setupReq: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expected: "",
		},
		{
			name: "header priority over query",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test?tenant_id=query-tenant", nil)
				req.Header.Set("X-Tenant-Id", "header-tenant")
				return req
			},
			expected: "header-tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			result := middleware.extractTenantID(req)
			if result != tt.expected {
				t.Errorf("Expected tenant ID '%s', got '%s'", tt.expected, result)
			}
		})
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
		req.Header.Set("X-Tenant-Id", "tenant123")
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
	}
}

func BenchmarkMiddleware_extractTenantID(b *testing.B) {
	middleware := NewMiddleware(DefaultConfig())
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-Id", "tenant123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = middleware.extractTenantID(req)
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
		req.Header.Set("X-Tenant-Id", "tenant123")
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
