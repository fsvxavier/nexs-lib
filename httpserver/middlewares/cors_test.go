package middlewares

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestNewCORSMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		priority int
	}{
		{
			name:     "Create with priority 5",
			priority: 5,
		},
		{
			name:     "Create with priority 0",
			priority: 0,
		},
		{
			name:     "Create with negative priority",
			priority: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCORSMiddleware(tt.priority)
			if got == nil {
				t.Error("NewCORSMiddleware() returned nil")
				return
			}

			if got.Name() != "cors" {
				t.Errorf("NewCORSMiddleware().Name() = %v, want %v", got.Name(), "cors")
			}

			if got.Priority() != tt.priority {
				t.Errorf("NewCORSMiddleware().Priority() = %v, want %v", got.Priority(), tt.priority)
			}

			// Verify default config is applied
			defaultConfig := DefaultCORSConfig()
			if !reflect.DeepEqual(got.config, defaultConfig) {
				t.Error("NewCORSMiddleware() did not apply default config")
			}
		})
	}
}

func TestNewCORSMiddlewareWithConfig(t *testing.T) {
	customConfig := CORSConfig{
		AllowedOrigins:      []string{"https://example.com", "https://api.example.com"},
		AllowAllOrigins:     false,
		AllowCredentials:    true,
		AllowedMethods:      []string{"GET", "POST"},
		AllowedHeaders:      []string{"Content-Type", "Authorization"},
		ExposedHeaders:      []string{"X-Custom-Header"},
		MaxAge:              3600,
		AllowPrivateNetwork: true,
		SkipPaths:           []string{"/public"},
		OptionsPassthrough:  true,
		VaryByOrigin:        false,
		Debug:               true,
	}

	tests := []struct {
		name     string
		priority int
		config   CORSConfig
	}{
		{
			name:     "Create with custom config",
			priority: 3,
			config:   customConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCORSMiddlewareWithConfig(tt.priority, tt.config)
			if got == nil {
				t.Error("NewCORSMiddlewareWithConfig() returned nil")
				return
			}

			if !reflect.DeepEqual(got.config, tt.config) {
				t.Error("NewCORSMiddlewareWithConfig() did not apply custom config correctly")
			}
		})
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	// Test basic CORS settings
	expectedOrigins := []string{"*"}
	if !reflect.DeepEqual(config.AllowedOrigins, expectedOrigins) {
		t.Errorf("DefaultCORSConfig().AllowedOrigins = %v, want %v", config.AllowedOrigins, expectedOrigins)
	}

	if config.AllowAllOrigins {
		t.Error("DefaultCORSConfig().AllowAllOrigins should be false")
	}

	if config.AllowCredentials {
		t.Error("DefaultCORSConfig().AllowCredentials should be false")
	}

	// Test allowed methods
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"}
	if !reflect.DeepEqual(config.AllowedMethods, expectedMethods) {
		t.Errorf("DefaultCORSConfig().AllowedMethods = %v, want %v", config.AllowedMethods, expectedMethods)
	}

	// Test allowed headers
	expectedHeaders := []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}
	if !reflect.DeepEqual(config.AllowedHeaders, expectedHeaders) {
		t.Errorf("DefaultCORSConfig().AllowedHeaders = %v, want %v", config.AllowedHeaders, expectedHeaders)
	}

	// Test MaxAge
	if config.MaxAge != 86400 {
		t.Errorf("DefaultCORSConfig().MaxAge = %v, want %v", config.MaxAge, 86400)
	}

	// Test other default values
	if config.AllowPrivateNetwork {
		t.Error("DefaultCORSConfig().AllowPrivateNetwork should be false")
	}

	if config.OptionsPassthrough {
		t.Error("DefaultCORSConfig().OptionsPassthrough should be false")
	}

	if !config.VaryByOrigin {
		t.Error("DefaultCORSConfig().VaryByOrigin should be true")
	}

	if config.Debug {
		t.Error("DefaultCORSConfig().Debug should be false")
	}
}

func TestCORSMiddleware_GetConfig(t *testing.T) {
	customConfig := CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	cm := NewCORSMiddlewareWithConfig(1, customConfig)
	got := cm.GetConfig()

	if !reflect.DeepEqual(got, customConfig) {
		t.Error("GetConfig() did not return the correct configuration")
	}
}

func TestCORSMiddleware_SetConfig(t *testing.T) {
	cm := NewCORSMiddleware(1)
	newConfig := CORSConfig{
		AllowedOrigins:   []string{"https://test.com"},
		AllowCredentials: false,
		MaxAge:           7200,
		Debug:            true,
	}

	cm.SetConfig(newConfig)
	got := cm.GetConfig()

	if !reflect.DeepEqual(got, newConfig) {
		t.Error("SetConfig() did not update the configuration correctly")
	}
}

func TestCORSMiddleware_extractCORSRequest(t *testing.T) {
	cm := NewCORSMiddleware(1)

	tests := []struct {
		name string
		req  interface{}
		want *CORSRequest
	}{
		{
			name: "Simple CORS request",
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
				"headers": map[string]string{
					"Origin": "https://example.com",
				},
			},
			want: &CORSRequest{
				Origin:        "https://example.com",
				Method:        "GET",
				Path:          "/api/data",
				IsCORSRequest: true,
			},
		},
		{
			name: "Preflight request",
			req: map[string]interface{}{
				"method": "OPTIONS",
				"path":   "/api/data",
				"headers": map[string]string{
					"Origin":                         "https://example.com",
					"Access-Control-Request-Method":  "POST",
					"Access-Control-Request-Headers": "Content-Type, Authorization",
				},
			},
			want: &CORSRequest{
				Origin:           "https://example.com",
				Method:           "OPTIONS",
				Path:             "/api/data",
				RequestedHeaders: []string{"Content-Type", "Authorization"},
				IsPreflightReq:   true,
				IsCORSRequest:    true,
			},
		},
		{
			name: "Non-CORS request",
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
				"headers": map[string]string{
					"User-Agent": "test-agent",
				},
			},
			want: &CORSRequest{
				Method:        "GET",
				Path:          "/api/data",
				IsCORSRequest: false,
			},
		},
		{
			name: "OPTIONS without CORS headers",
			req: map[string]interface{}{
				"method": "OPTIONS",
				"path":   "/api/health",
				"headers": map[string]string{
					"User-Agent": "test-agent",
				},
			},
			want: &CORSRequest{
				Method:        "OPTIONS",
				Path:          "/api/health",
				IsCORSRequest: false,
			},
		},
		{
			name: "Invalid request type",
			req:  "invalid",
			want: &CORSRequest{},
		},
		{
			name: "Request without headers",
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
			},
			want: &CORSRequest{
				Method: "GET",
				Path:   "/api/data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.extractCORSRequest(tt.req)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractCORSRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware_shouldSkipCORS(t *testing.T) {
	config := DefaultCORSConfig()
	config.SkipPaths = []string{"/health", "/metrics", "/public"}

	cm := NewCORSMiddlewareWithConfig(1, config)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Skip health path",
			path: "/health",
			want: true,
		},
		{
			name: "Skip metrics path",
			path: "/metrics",
			want: true,
		},
		{
			name: "Skip public path",
			path: "/public",
			want: true,
		},
		{
			name: "Do not skip API path",
			path: "/api/data",
			want: false,
		},
		{
			name: "Do not skip root path",
			path: "/",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.shouldSkipCORS(tt.path)
			if got != tt.want {
				t.Errorf("shouldSkipCORS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware_isOriginAllowed(t *testing.T) {
	tests := []struct {
		name   string
		config CORSConfig
		origin string
		want   bool
	}{
		{
			name: "Allow all origins enabled",
			config: CORSConfig{
				AllowAllOrigins: true,
			},
			origin: "https://example.com",
			want:   true,
		},
		{
			name: "Exact origin match",
			config: CORSConfig{
				AllowedOrigins: []string{"https://example.com", "https://api.example.com"},
			},
			origin: "https://example.com",
			want:   true,
		},
		{
			name: "Wildcard match",
			config: CORSConfig{
				AllowedOrigins: []string{"*"},
			},
			origin: "https://example.com",
			want:   true,
		},
		{
			name: "Subdomain wildcard match",
			config: CORSConfig{
				AllowedOrigins: []string{"*.example.com"},
			},
			origin: "https://api.example.com",
			want:   true,
		},
		{
			name: "Subdomain wildcard no match",
			config: CORSConfig{
				AllowedOrigins: []string{"*.example.com"},
			},
			origin: "https://different.com",
			want:   false,
		},
		{
			name: "Origin not in allowed list",
			config: CORSConfig{
				AllowedOrigins: []string{"https://allowed.com"},
			},
			origin: "https://blocked.com",
			want:   false,
		},
		{
			name: "Custom origin function allows",
			config: CORSConfig{
				AllowOriginFunc: func(origin string) bool {
					return strings.HasSuffix(origin, ".trusted.com")
				},
			},
			origin: "https://api.trusted.com",
			want:   true,
		},
		{
			name: "Custom origin function blocks",
			config: CORSConfig{
				AllowOriginFunc: func(origin string) bool {
					return strings.HasSuffix(origin, ".trusted.com")
				},
			},
			origin: "https://malicious.com",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewCORSMiddlewareWithConfig(1, tt.config)
			got := cm.isOriginAllowed(tt.origin)
			if got != tt.want {
				t.Errorf("isOriginAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware_matchOrigin(t *testing.T) {
	cm := NewCORSMiddleware(1)

	tests := []struct {
		name          string
		origin        string
		allowedOrigin string
		want          bool
	}{
		{
			name:          "Wildcard match",
			origin:        "https://example.com",
			allowedOrigin: "*",
			want:          true,
		},
		{
			name:          "Exact match",
			origin:        "https://example.com",
			allowedOrigin: "https://example.com",
			want:          true,
		},
		{
			name:          "Subdomain wildcard match",
			origin:        "https://api.example.com",
			allowedOrigin: "*.example.com",
			want:          true,
		},
		{
			name:          "Subdomain wildcard no match - different domain",
			origin:        "https://api.different.com",
			allowedOrigin: "*.example.com",
			want:          false,
		},
		{
			name:          "Subdomain wildcard no match - exact domain",
			origin:        "https://example.com",
			allowedOrigin: "*.example.com",
			want:          false,
		},
		{
			name:          "No match",
			origin:        "https://blocked.com",
			allowedOrigin: "https://allowed.com",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.matchOrigin(tt.origin, tt.allowedOrigin)
			if got != tt.want {
				t.Errorf("matchOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware_isMethodAllowed(t *testing.T) {
	config := CORSConfig{
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}
	cm := NewCORSMiddlewareWithConfig(1, config)

	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{
			name:   "Allowed method GET",
			method: "GET",
			want:   true,
		},
		{
			name:   "Allowed method POST",
			method: "POST",
			want:   true,
		},
		{
			name:   "Allowed method lowercase",
			method: "get",
			want:   true,
		},
		{
			name:   "Not allowed method",
			method: "PATCH",
			want:   false,
		},
		{
			name:   "Not allowed method OPTIONS",
			method: "OPTIONS",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.isMethodAllowed(tt.method)
			if got != tt.want {
				t.Errorf("isMethodAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware_isHeaderAllowed(t *testing.T) {
	config := CORSConfig{
		AllowedHeaders: []string{"Authorization", "X-Custom-Header"},
	}
	cm := NewCORSMiddlewareWithConfig(1, config)

	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{
			name:   "Simple header Accept",
			header: "Accept",
			want:   true,
		},
		{
			name:   "Simple header Content-Type",
			header: "Content-Type",
			want:   true,
		},
		{
			name:   "Allowed custom header",
			header: "Authorization",
			want:   true,
		},
		{
			name:   "Allowed custom header case insensitive",
			header: "AUTHORIZATION",
			want:   true,
		},
		{
			name:   "Allowed custom header with spaces",
			header: " X-Custom-Header ",
			want:   true,
		},
		{
			name:   "Not allowed header",
			header: "X-Forbidden-Header",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.isHeaderAllowed(tt.header)
			if got != tt.want {
				t.Errorf("isHeaderAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware_areHeadersAllowed(t *testing.T) {
	config := CORSConfig{
		AllowedHeaders: []string{"Authorization", "X-Custom-Header"},
	}
	cm := NewCORSMiddlewareWithConfig(1, config)

	tests := []struct {
		name             string
		requestedHeaders []string
		want             bool
	}{
		{
			name:             "All headers allowed",
			requestedHeaders: []string{"Accept", "Authorization"},
			want:             true,
		},
		{
			name:             "Some headers not allowed",
			requestedHeaders: []string{"Authorization", "X-Forbidden"},
			want:             false,
		},
		{
			name:             "Empty headers list",
			requestedHeaders: []string{},
			want:             true,
		},
		{
			name:             "Only simple headers",
			requestedHeaders: []string{"Accept", "Content-Type"},
			want:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.areHeadersAllowed(tt.requestedHeaders)
			if got != tt.want {
				t.Errorf("areHeadersAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware_parseHeaderList(t *testing.T) {
	cm := NewCORSMiddleware(1)

	tests := []struct {
		name      string
		headerStr string
		want      []string
	}{
		{
			name:      "Empty string",
			headerStr: "",
			want:      []string{},
		},
		{
			name:      "Single header",
			headerStr: "Content-Type",
			want:      []string{"Content-Type"},
		},
		{
			name:      "Multiple headers",
			headerStr: "Content-Type, Authorization, X-Custom",
			want:      []string{"Content-Type", "Authorization", "X-Custom"},
		},
		{
			name:      "Headers with spaces",
			headerStr: " Content-Type , Authorization , X-Custom ",
			want:      []string{"Content-Type", "Authorization", "X-Custom"},
		},
		{
			name:      "Headers with empty segments",
			headerStr: "Content-Type,, Authorization, , X-Custom",
			want:      []string{"Content-Type", "Authorization", "X-Custom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.parseHeaderList(tt.headerStr)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseHeaderList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware_createCORSResponse(t *testing.T) {
	tests := []struct {
		name    string
		config  CORSConfig
		corsReq *CORSRequest
		check   func(*CORSResponse) bool
	}{
		{
			name: "Simple CORS request",
			config: CORSConfig{
				AllowedOrigins:   []string{"https://example.com"},
				AllowCredentials: false,
				ExposedHeaders:   []string{"X-Custom"},
				VaryByOrigin:     true,
			},
			corsReq: &CORSRequest{
				Origin:        "https://example.com",
				IsCORSRequest: true,
			},
			check: func(resp *CORSResponse) bool {
				return resp.AllowOrigin == "https://example.com" &&
					resp.ExposeHeaders == "X-Custom" &&
					resp.AllowCredentials == "" &&
					len(resp.VaryHeaders) == 1 && resp.VaryHeaders[0] == "Origin"
			},
		},
		{
			name: "Preflight request",
			config: CORSConfig{
				AllowedOrigins:   []string{"https://example.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT"},
				AllowedHeaders:   []string{"Content-Type", "Authorization"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			corsReq: &CORSRequest{
				Origin:         "https://example.com",
				IsPreflightReq: true,
				IsCORSRequest:  true,
			},
			check: func(resp *CORSResponse) bool {
				return resp.AllowOrigin == "https://example.com" &&
					resp.AllowMethods == "GET, POST, PUT" &&
					resp.AllowHeaders == "Content-Type, Authorization" &&
					resp.AllowCredentials == "true" &&
					resp.MaxAge == "3600"
			},
		},
		{
			name: "Allow all origins without credentials",
			config: CORSConfig{
				AllowAllOrigins:  true,
				AllowCredentials: false,
			},
			corsReq: &CORSRequest{
				Origin:        "https://example.com",
				IsCORSRequest: true,
			},
			check: func(resp *CORSResponse) bool {
				return resp.AllowOrigin == "*"
			},
		},
		{
			name: "Allow all origins with credentials",
			config: CORSConfig{
				AllowAllOrigins:  true,
				AllowCredentials: true,
			},
			corsReq: &CORSRequest{
				Origin:        "https://example.com",
				IsCORSRequest: true,
			},
			check: func(resp *CORSResponse) bool {
				return resp.AllowOrigin == "https://example.com" &&
					resp.AllowCredentials == "true"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewCORSMiddlewareWithConfig(1, tt.config)
			got := cm.createCORSResponse(tt.corsReq)
			if !tt.check(got) {
				t.Errorf("createCORSResponse() failed check for %s", tt.name)
			}
		})
	}
}

func TestCORSMiddleware_corsResponseToHeaders(t *testing.T) {
	cm := NewCORSMiddleware(1)

	corsResp := &CORSResponse{
		AllowOrigin:      "https://example.com",
		AllowMethods:     "GET, POST, PUT",
		AllowHeaders:     "Content-Type, Authorization",
		ExposeHeaders:    "X-Custom",
		AllowCredentials: "true",
		MaxAge:           "3600",
		VaryHeaders:      []string{"Origin", "Accept"},
	}

	got := cm.corsResponseToHeaders(corsResp)

	expected := map[string]string{
		"Access-Control-Allow-Origin":      "https://example.com",
		"Access-Control-Allow-Methods":     "GET, POST, PUT",
		"Access-Control-Allow-Headers":     "Content-Type, Authorization",
		"Access-Control-Expose-Headers":    "X-Custom",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "3600",
		"Vary":                             "Origin, Accept",
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("corsResponseToHeaders() = %v, want %v", got, expected)
	}
}

func TestCORSMiddleware_createCORSErrorResponse(t *testing.T) {
	cm := NewCORSMiddleware(1)

	resp, err := cm.createCORSErrorResponse("Test error message")
	if err != nil {
		t.Errorf("createCORSErrorResponse() error = %v, want nil", err)
		return
	}

	httpResp, ok := resp.(map[string]interface{})
	if !ok {
		t.Error("createCORSErrorResponse() did not return map[string]interface{}")
		return
	}

	if httpResp["status_code"] != 403 {
		t.Errorf("createCORSErrorResponse() status_code = %v, want %v", httpResp["status_code"], 403)
	}

	if headers, exists := httpResp["headers"]; exists {
		if h, ok := headers.(map[string]string); ok {
			if h["Content-Type"] != "application/json" {
				t.Errorf("createCORSErrorResponse() Content-Type = %v, want %v", h["Content-Type"], "application/json")
			}
		}
	}

	body := httpResp["body"].(string)
	if !strings.Contains(body, "Test error message") {
		t.Error("createCORSErrorResponse() body does not contain error message")
	}
}

func TestCORSMiddleware_Reset(t *testing.T) {
	cm := NewCORSMiddleware(1)

	// Simulate some metrics
	cm.preflightRequests = 10
	cm.corsRequests = 50
	cm.corsAllowed = 45
	cm.corsBlocked = 5
	cm.originChecks = 100

	cm.Reset()

	metrics := cm.GetMetrics()
	if metrics["preflight_requests"].(int64) != 0 {
		t.Error("Reset() did not reset preflight_requests")
	}
	if metrics["cors_requests"].(int64) != 0 {
		t.Error("Reset() did not reset cors_requests")
	}
	if metrics["cors_allowed"].(int64) != 0 {
		t.Error("Reset() did not reset cors_allowed")
	}
	if metrics["cors_blocked"].(int64) != 0 {
		t.Error("Reset() did not reset cors_blocked")
	}
	if metrics["origin_checks"].(int64) != 0 {
		t.Error("Reset() did not reset origin_checks")
	}
}

func TestCORSMiddleware_GetMetrics(t *testing.T) {
	cm := NewCORSMiddleware(1)

	// Initial metrics should be zero
	metrics := cm.GetMetrics()

	expectedKeys := []string{
		"preflight_requests", "cors_requests", "cors_allowed",
		"cors_blocked", "origin_checks", "allowed_rate", "uptime",
	}

	for _, key := range expectedKeys {
		if _, exists := metrics[key]; !exists {
			t.Errorf("GetMetrics() missing key: %s", key)
		}
	}

	// Test with some data
	cm.preflightRequests = 10
	cm.corsRequests = 50
	cm.corsAllowed = 45
	cm.corsBlocked = 5

	metrics = cm.GetMetrics()

	if metrics["preflight_requests"].(int64) != 10 {
		t.Errorf("GetMetrics() preflight_requests = %v, want %v", metrics["preflight_requests"], 10)
	}

	allowedRate := metrics["allowed_rate"].(float64)
	if allowedRate != 90.0 {
		t.Errorf("GetMetrics() allowed_rate = %v, want %v", allowedRate, 90.0)
	}
}

func TestCORSMiddleware_AddAllowedOrigin(t *testing.T) {
	cm := NewCORSMiddleware(1)
	originalCount := len(cm.GetConfig().AllowedOrigins)

	cm.AddAllowedOrigin("https://new.example.com")

	config := cm.GetConfig()
	if len(config.AllowedOrigins) != originalCount+1 {
		t.Errorf("AddAllowedOrigin() did not add origin, got %d origins, want %d", len(config.AllowedOrigins), originalCount+1)
	}

	found := false
	for _, origin := range config.AllowedOrigins {
		if origin == "https://new.example.com" {
			found = true
			break
		}
	}

	if !found {
		t.Error("AddAllowedOrigin() did not add the specified origin")
	}
}

func TestCORSMiddleware_RemoveAllowedOrigin(t *testing.T) {
	cm := NewCORSMiddleware(1)

	// Add an origin first
	cm.AddAllowedOrigin("https://test.example.com")
	originalCount := len(cm.GetConfig().AllowedOrigins)

	// Remove it
	cm.RemoveAllowedOrigin("https://test.example.com")

	config := cm.GetConfig()
	if len(config.AllowedOrigins) != originalCount-1 {
		t.Errorf("RemoveAllowedOrigin() did not remove origin, got %d origins, want %d", len(config.AllowedOrigins), originalCount-1)
	}

	// Verify it's not in the list
	for _, origin := range config.AllowedOrigins {
		if origin == "https://test.example.com" {
			t.Error("RemoveAllowedOrigin() did not remove the specified origin")
		}
	}

	// Test removing non-existent origin
	originalCount = len(cm.GetConfig().AllowedOrigins)
	cm.RemoveAllowedOrigin("https://nonexistent.com")

	if len(cm.GetConfig().AllowedOrigins) != originalCount {
		t.Error("RemoveAllowedOrigin() should not change list when removing non-existent origin")
	}
}

func TestCORSMiddleware_SetAllowAllOrigins(t *testing.T) {
	cm := NewCORSMiddleware(1)

	// Test enabling
	cm.SetAllowAllOrigins(true)
	if !cm.GetConfig().AllowAllOrigins {
		t.Error("SetAllowAllOrigins(true) did not enable allow all origins")
	}

	// Test disabling
	cm.SetAllowAllOrigins(false)
	if cm.GetConfig().AllowAllOrigins {
		t.Error("SetAllowAllOrigins(false) did not disable allow all origins")
	}
}

func TestCORSMiddleware_SetAllowCredentials(t *testing.T) {
	cm := NewCORSMiddleware(1)

	// Test enabling
	cm.SetAllowCredentials(true)
	if !cm.GetConfig().AllowCredentials {
		t.Error("SetAllowCredentials(true) did not enable credentials support")
	}

	// Test disabling
	cm.SetAllowCredentials(false)
	if cm.GetConfig().AllowCredentials {
		t.Error("SetAllowCredentials(false) did not disable credentials support")
	}
}

func TestCORSMiddleware_Process(t *testing.T) {
	tests := []struct {
		name           string
		config         CORSConfig
		req            interface{}
		expectDisabled bool
		expectSkipped  bool
		expectError    bool
	}{
		{
			name:   "Process with disabled middleware",
			config: DefaultCORSConfig(),
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
				"headers": map[string]string{
					"Origin": "https://example.com",
				},
			},
			expectDisabled: true,
		},
		{
			name: "Process skip for health check path",
			config: CORSConfig{
				SkipPaths: []string{"/health"},
			},
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/health",
				"headers": map[string]string{
					"Origin": "https://example.com",
				},
			},
			expectSkipped: true,
		},
		{
			name:   "Process non-CORS request",
			config: DefaultCORSConfig(),
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
				"headers": map[string]string{
					"User-Agent": "test-agent",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewCORSMiddlewareWithConfig(1, tt.config)

			if tt.expectDisabled {
				cm.SetEnabled(false)
			} else {
				cm.SetEnabled(true)
			}

			ctx := context.Background()
			called := false
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				called = true
				return map[string]interface{}{"body": "test"}, nil
			}

			resp, err := cm.Process(ctx, tt.req, next)

			if (err != nil) != tt.expectError {
				t.Errorf("Process() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !called {
				t.Error("Process() did not call next middleware")
			}

			if resp == nil {
				t.Error("Process() returned nil response")
			}
		})
	}
}

// Benchmark tests
func BenchmarkCORSMiddleware_isOriginAllowed(b *testing.B) {
	config := CORSConfig{
		AllowedOrigins: []string{"https://example.com", "https://api.example.com", "*.test.com"},
	}
	cm := NewCORSMiddlewareWithConfig(1, config)
	origin := "https://api.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.isOriginAllowed(origin)
	}
}

func BenchmarkCORSMiddleware_parseHeaderList(b *testing.B) {
	cm := NewCORSMiddleware(1)
	headerStr := "Content-Type, Authorization, X-Custom-Header, Accept"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.parseHeaderList(headerStr)
	}
}

func BenchmarkCORSMiddleware_extractCORSRequest(b *testing.B) {
	cm := NewCORSMiddleware(1)
	req := map[string]interface{}{
		"method": "OPTIONS",
		"path":   "/api/data",
		"headers": map[string]string{
			"Origin":                         "https://example.com",
			"Access-Control-Request-Method":  "POST",
			"Access-Control-Request-Headers": "Content-Type, Authorization",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.extractCORSRequest(req)
	}
}
