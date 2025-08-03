package middlewares

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewHealthCheckMiddleware(t *testing.T) {
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
			got := NewHealthCheckMiddleware(tt.priority)
			if got == nil {
				t.Error("NewHealthCheckMiddleware() returned nil")
				return
			}

			if got.Name() != "health_check" {
				t.Errorf("NewHealthCheckMiddleware().Name() = %v, want %v", got.Name(), "health_check")
			}

			if got.Priority() != tt.priority {
				t.Errorf("NewHealthCheckMiddleware().Priority() = %v, want %v", got.Priority(), tt.priority)
			}

			// Verify default config is applied
			defaultConfig := DefaultHealthCheckConfig()
			if !reflect.DeepEqual(got.config, defaultConfig) {
				t.Error("NewHealthCheckMiddleware() did not apply default config")
			}

			// Verify checkers map is initialized
			if got.checkers == nil {
				t.Error("NewHealthCheckMiddleware() did not initialize checkers map")
			}
		})
	}
}

func TestNewHealthCheckMiddlewareWithConfig(t *testing.T) {
	customConfig := HealthCheckConfig{
		HealthPath:        "/custom/health",
		LivenessPath:      "/custom/live",
		ReadinessPath:     "/custom/ready",
		CheckInterval:     time.Minute,
		CheckTimeout:      time.Second * 5,
		CacheTimeout:      time.Second * 10,
		SuccessStatusCode: 200,
		FailureStatusCode: 500,
		DetailedResponse:  false,
		IncludeMetrics:    false,
		IncludeVersion:    false,
		Dependencies:      []string{"db", "redis"},
		CriticalChecks:    []string{"db"},
		FailFast:          true,
		ParallelChecks:    false,
		GracefulShutdown:  false,
		CustomData:        map[string]interface{}{"env": "test"},
		Version:           "2.0.0",
		ServiceName:       "test-service",
		Environment:       "test",
	}

	tests := []struct {
		name     string
		priority int
		config   HealthCheckConfig
	}{
		{
			name:     "Create with custom config",
			priority: 3,
			config:   customConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHealthCheckMiddlewareWithConfig(tt.priority, tt.config)
			if got == nil {
				t.Error("NewHealthCheckMiddlewareWithConfig() returned nil")
				return
			}

			if !reflect.DeepEqual(got.config, tt.config) {
				t.Error("NewHealthCheckMiddlewareWithConfig() did not apply custom config correctly")
			}
		})
	}
}

func TestDefaultHealthCheckConfig(t *testing.T) {
	config := DefaultHealthCheckConfig()

	// Test endpoint paths
	if config.HealthPath != "/health" {
		t.Errorf("DefaultHealthCheckConfig().HealthPath = %v, want %v", config.HealthPath, "/health")
	}
	if config.LivenessPath != "/health/live" {
		t.Errorf("DefaultHealthCheckConfig().LivenessPath = %v, want %v", config.LivenessPath, "/health/live")
	}
	if config.ReadinessPath != "/health/ready" {
		t.Errorf("DefaultHealthCheckConfig().ReadinessPath = %v, want %v", config.ReadinessPath, "/health/ready")
	}

	// Test timeouts
	if config.CheckInterval != time.Second*30 {
		t.Errorf("DefaultHealthCheckConfig().CheckInterval = %v, want %v", config.CheckInterval, time.Second*30)
	}
	if config.CheckTimeout != time.Second*10 {
		t.Errorf("DefaultHealthCheckConfig().CheckTimeout = %v, want %v", config.CheckTimeout, time.Second*10)
	}
	if config.CacheTimeout != time.Second*5 {
		t.Errorf("DefaultHealthCheckConfig().CacheTimeout = %v, want %v", config.CacheTimeout, time.Second*5)
	}

	// Test status codes
	if config.SuccessStatusCode != 200 {
		t.Errorf("DefaultHealthCheckConfig().SuccessStatusCode = %v, want %v", config.SuccessStatusCode, 200)
	}
	if config.FailureStatusCode != 503 {
		t.Errorf("DefaultHealthCheckConfig().FailureStatusCode = %v, want %v", config.FailureStatusCode, 503)
	}

	// Test boolean settings
	if !config.DetailedResponse {
		t.Error("DefaultHealthCheckConfig().DetailedResponse should be true")
	}
	if !config.IncludeMetrics {
		t.Error("DefaultHealthCheckConfig().IncludeMetrics should be true")
	}
	if !config.IncludeVersion {
		t.Error("DefaultHealthCheckConfig().IncludeVersion should be true")
	}
	if config.FailFast {
		t.Error("DefaultHealthCheckConfig().FailFast should be false")
	}
	if !config.ParallelChecks {
		t.Error("DefaultHealthCheckConfig().ParallelChecks should be true")
	}
	if !config.GracefulShutdown {
		t.Error("DefaultHealthCheckConfig().GracefulShutdown should be true")
	}

	// Test service info
	if config.Version != "1.0.0" {
		t.Errorf("DefaultHealthCheckConfig().Version = %v, want %v", config.Version, "1.0.0")
	}
	if config.ServiceName != "http-server" {
		t.Errorf("DefaultHealthCheckConfig().ServiceName = %v, want %v", config.ServiceName, "http-server")
	}
	if config.Environment != "development" {
		t.Errorf("DefaultHealthCheckConfig().Environment = %v, want %v", config.Environment, "development")
	}

	// Test custom data is initialized
	if config.CustomData == nil {
		t.Error("DefaultHealthCheckConfig().CustomData should be initialized")
	}
}

func TestHealthCheckMiddleware_GetConfig(t *testing.T) {
	customConfig := HealthCheckConfig{
		HealthPath:  "/custom/health",
		Version:     "2.0.0",
		ServiceName: "test-service",
		Environment: "test",
		FailFast:    true,
		CustomData:  map[string]interface{}{"test": "value"},
	}

	hcm := NewHealthCheckMiddlewareWithConfig(1, customConfig)
	got := hcm.GetConfig()

	if !reflect.DeepEqual(got, customConfig) {
		t.Error("GetConfig() did not return the correct configuration")
	}
}

func TestHealthCheckMiddleware_SetConfig(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)
	newConfig := HealthCheckConfig{
		HealthPath:       "/new/health",
		CheckTimeout:     time.Second * 15,
		DetailedResponse: false,
		ParallelChecks:   false,
		Version:          "3.0.0",
		CustomData:       map[string]interface{}{"new": "config"},
	}

	hcm.SetConfig(newConfig)
	got := hcm.GetConfig()

	if !reflect.DeepEqual(got, newConfig) {
		t.Error("SetConfig() did not update the configuration correctly")
	}
}

func TestHealthCheckMiddleware_extractPath(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)

	tests := []struct {
		name string
		req  interface{}
		want string
	}{
		{
			name: "Valid HTTP request with path",
			req: map[string]interface{}{
				"path": "/health",
			},
			want: "/health",
		},
		{
			name: "HTTP request with complex path",
			req: map[string]interface{}{
				"path": "/api/v1/health/ready",
			},
			want: "/api/v1/health/ready",
		},
		{
			name: "Request without path",
			req: map[string]interface{}{
				"method": "GET",
			},
			want: "",
		},
		{
			name: "Invalid request type",
			req:  "invalid",
			want: "",
		},
		{
			name: "Request with non-string path",
			req: map[string]interface{}{
				"path": 123,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hcm.extractPath(tt.req)
			if got != tt.want {
				t.Errorf("extractPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealthCheckMiddleware_isHealthCheckPath(t *testing.T) {
	config := DefaultHealthCheckConfig()
	config.HealthPath = "/health"
	config.LivenessPath = "/health/live"
	config.ReadinessPath = "/health/ready"

	hcm := NewHealthCheckMiddlewareWithConfig(1, config)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Health check path",
			path: "/health",
			want: true,
		},
		{
			name: "Liveness check path",
			path: "/health/live",
			want: true,
		},
		{
			name: "Readiness check path",
			path: "/health/ready",
			want: true,
		},
		{
			name: "API path",
			path: "/api/data",
			want: false,
		},
		{
			name: "Root path",
			path: "/",
			want: false,
		},
		{
			name: "Empty path",
			path: "",
			want: false,
		},
		{
			name: "Similar but different path",
			path: "/health/status",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hcm.isHealthCheckPath(tt.path)
			if got != tt.want {
				t.Errorf("isHealthCheckPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealthCheckMiddleware_AddChecker(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)

	tests := []struct {
		name    string
		checker HealthChecker
		wantErr bool
	}{
		{
			name:    "Add valid checker",
			checker: NewDatabaseHealthChecker("db", true),
			wantErr: false,
		},
		{
			name:    "Add nil checker",
			checker: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hcm.AddChecker(tt.checker)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddChecker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checker != nil {
				// Verify checker was added
				_, err := hcm.GetChecker(tt.checker.Name())
				if err != nil {
					t.Errorf("AddChecker() did not add checker properly: %v", err)
				}
			}
		})
	}
}

func TestHealthCheckMiddleware_RemoveChecker(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)
	checker := NewDatabaseHealthChecker("db", true)

	// Add a checker first
	err := hcm.AddChecker(checker)
	if err != nil {
		t.Fatalf("Failed to add checker: %v", err)
	}

	tests := []struct {
		name        string
		checkerName string
		wantErr     bool
	}{
		{
			name:        "Remove existing checker",
			checkerName: "db",
			wantErr:     false,
		},
		{
			name:        "Remove non-existent checker",
			checkerName: "nonexistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hcm.RemoveChecker(tt.checkerName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveChecker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify checker was removed
				_, err := hcm.GetChecker(tt.checkerName)
				if err == nil {
					t.Error("RemoveChecker() did not remove checker properly")
				}
			}
		})
	}
}

func TestHealthCheckMiddleware_GetChecker(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)
	checker := NewDatabaseHealthChecker("db", true)

	// Add a checker first
	err := hcm.AddChecker(checker)
	if err != nil {
		t.Fatalf("Failed to add checker: %v", err)
	}

	tests := []struct {
		name        string
		checkerName string
		wantErr     bool
	}{
		{
			name:        "Get existing checker",
			checkerName: "db",
			wantErr:     false,
		},
		{
			name:        "Get non-existent checker",
			checkerName: "nonexistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hcm.GetChecker(tt.checkerName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChecker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Error("GetChecker() returned nil checker")
				}
				if got.Name() != tt.checkerName {
					t.Errorf("GetChecker() returned wrong checker: got %v, want %v", got.Name(), tt.checkerName)
				}
			}
		})
	}
}

func TestHealthCheckMiddleware_Reset(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)

	// Simulate some metrics
	hcm.totalChecks = 10
	hcm.successfulChecks = 8
	hcm.failedChecks = 2
	hcm.lastCheckTime = time.Now().Unix()

	// Set some cached data
	hcm.cachedResult = &HealthResult{Status: HealthStatusHealthy}
	hcm.cacheExpiry = time.Now().Add(time.Minute)

	hcm.Reset()

	metrics := hcm.GetMetrics()
	if metrics["total_checks"].(int64) != 0 {
		t.Error("Reset() did not reset total_checks")
	}
	if metrics["successful_checks"].(int64) != 0 {
		t.Error("Reset() did not reset successful_checks")
	}
	if metrics["failed_checks"].(int64) != 0 {
		t.Error("Reset() did not reset failed_checks")
	}

	// Check cache was cleared
	if hcm.getCachedResult() != nil {
		t.Error("Reset() did not clear cached result")
	}
}

func TestHealthCheckMiddleware_GetMetrics(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)

	// Initial metrics should be zero
	metrics := hcm.GetMetrics()

	expectedKeys := []string{
		"total_checks", "successful_checks", "failed_checks",
		"success_rate", "time_since_last_check", "active_checkers", "uptime",
	}

	for _, key := range expectedKeys {
		if _, exists := metrics[key]; !exists {
			t.Errorf("GetMetrics() missing key: %s", key)
		}
	}

	// Test with some data
	hcm.totalChecks = 100
	hcm.successfulChecks = 95
	hcm.failedChecks = 5
	hcm.lastCheckTime = time.Now().Unix()

	// Add some checkers
	hcm.AddChecker(NewDatabaseHealthChecker("db", true))
	hcm.AddChecker(NewServiceHealthChecker("api", "http://api.example.com", false))

	metrics = hcm.GetMetrics()

	if metrics["total_checks"].(int64) != 100 {
		t.Errorf("GetMetrics() total_checks = %v, want %v", metrics["total_checks"], 100)
	}

	successRate := metrics["success_rate"].(float64)
	if successRate != 95.0 {
		t.Errorf("GetMetrics() success_rate = %v, want %v", successRate, 95.0)
	}

	activeCheckers := metrics["active_checkers"].(int)
	if activeCheckers != 2 {
		t.Errorf("GetMetrics() active_checkers = %v, want %v", activeCheckers, 2)
	}
}

func TestHealthCheckMiddleware_determineOverallStatus(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)

	tests := []struct {
		name   string
		checks map[string]HealthCheckResult
		want   HealthStatus
	}{
		{
			name:   "No checks - healthy",
			checks: map[string]HealthCheckResult{},
			want:   HealthStatusHealthy,
		},
		{
			name: "All healthy",
			checks: map[string]HealthCheckResult{
				"db":  {Status: HealthStatusHealthy},
				"api": {Status: HealthStatusHealthy},
			},
			want: HealthStatusHealthy,
		},
		{
			name: "One unhealthy - overall unhealthy",
			checks: map[string]HealthCheckResult{
				"db":  {Status: HealthStatusHealthy},
				"api": {Status: HealthStatusUnhealthy},
			},
			want: HealthStatusUnhealthy,
		},
		{
			name: "One degraded - overall degraded",
			checks: map[string]HealthCheckResult{
				"db":  {Status: HealthStatusHealthy},
				"api": {Status: HealthStatusDegraded},
			},
			want: HealthStatusDegraded,
		},
		{
			name: "Unknown status treated as degraded",
			checks: map[string]HealthCheckResult{
				"db":  {Status: HealthStatusHealthy},
				"api": {Status: HealthStatusUnknown},
			},
			want: HealthStatusDegraded,
		},
		{
			name: "Unhealthy takes precedence over degraded",
			checks: map[string]HealthCheckResult{
				"db":    {Status: HealthStatusUnhealthy},
				"api":   {Status: HealthStatusDegraded},
				"cache": {Status: HealthStatusHealthy},
			},
			want: HealthStatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hcm.determineOverallStatus(tt.checks)
			if got != tt.want {
				t.Errorf("determineOverallStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealthCheckMiddleware_createHealthResponse(t *testing.T) {
	config := DefaultHealthCheckConfig()
	config.SuccessStatusCode = 200
	config.FailureStatusCode = 503
	config.DetailedResponse = true

	hcm := NewHealthCheckMiddlewareWithConfig(1, config)

	tests := []struct {
		name         string
		result       *HealthResult
		wantStatus   int
		wantDetailed bool
	}{
		{
			name: "Healthy response",
			result: &HealthResult{
				Status:    HealthStatusHealthy,
				Timestamp: time.Now(),
				Version:   "1.0.0",
				Checks:    map[string]HealthCheckResult{},
			},
			wantStatus:   200,
			wantDetailed: true,
		},
		{
			name: "Unhealthy response",
			result: &HealthResult{
				Status:    HealthStatusUnhealthy,
				Timestamp: time.Now(),
				Version:   "1.0.0",
				Checks:    map[string]HealthCheckResult{},
			},
			wantStatus:   503,
			wantDetailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := hcm.createHealthResponse(tt.result)

			httpResp, ok := resp.(map[string]interface{})
			if !ok {
				t.Error("createHealthResponse() did not return map[string]interface{}")
				return
			}

			if httpResp["status_code"] != tt.wantStatus {
				t.Errorf("createHealthResponse() status_code = %v, want %v", httpResp["status_code"], tt.wantStatus)
			}

			// Check headers
			if headers, exists := httpResp["headers"]; exists {
				if h, ok := headers.(map[string]string); ok {
					if h["Content-Type"] != "application/json" {
						t.Error("createHealthResponse() missing or incorrect Content-Type header")
					}
					if h["Cache-Control"] != "no-cache, no-store, must-revalidate" {
						t.Error("createHealthResponse() missing or incorrect Cache-Control header")
					}
				}
			}

			// Check body is valid JSON
			if body, exists := httpResp["body"]; exists {
				if bodyStr, ok := body.(string); ok {
					var jsonData interface{}
					if err := json.Unmarshal([]byte(bodyStr), &jsonData); err != nil {
						t.Errorf("createHealthResponse() body is not valid JSON: %v", err)
					}

					if tt.wantDetailed {
						// Check if detailed response contains expected fields
						if !strings.Contains(bodyStr, "status") || !strings.Contains(bodyStr, "timestamp") {
							t.Error("createHealthResponse() detailed response missing expected fields")
						}
					}
				}
			}
		})
	}
}

func TestHealthCheckMiddleware_getCachedResult(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)

	// Test no cached result
	result := hcm.getCachedResult()
	if result != nil {
		t.Error("getCachedResult() should return nil when no cache exists")
	}

	// Test valid cached result
	testResult := &HealthResult{Status: HealthStatusHealthy}
	hcm.setCachedResult(testResult)

	result = hcm.getCachedResult()
	if result == nil {
		t.Error("getCachedResult() should return cached result")
	}
	if result.Status != HealthStatusHealthy {
		t.Error("getCachedResult() returned incorrect cached result")
	}

	// Test expired cache
	hcm.cacheExpiry = time.Now().Add(-time.Hour) // Set to past

	result = hcm.getCachedResult()
	if result != nil {
		t.Error("getCachedResult() should return nil for expired cache")
	}
}

func TestHealthCheckMiddleware_setCachedResult(t *testing.T) {
	hcm := NewHealthCheckMiddleware(1)

	testResult := &HealthResult{
		Status:    HealthStatusHealthy,
		Timestamp: time.Now(),
	}

	hcm.setCachedResult(testResult)

	// Verify cache was set
	hcm.cacheMutex.RLock()
	if hcm.cachedResult == nil {
		t.Error("setCachedResult() did not set cached result")
	}
	if hcm.cachedResult.Status != HealthStatusHealthy {
		t.Error("setCachedResult() set incorrect cached result")
	}
	if hcm.cacheExpiry.IsZero() {
		t.Error("setCachedResult() did not set cache expiry")
	}
	hcm.cacheMutex.RUnlock()
}

func TestHealthCheckMiddleware_Process(t *testing.T) {
	tests := []struct {
		name           string
		config         HealthCheckConfig
		req            interface{}
		expectDisabled bool
		expectHealth   bool
		expectError    bool
	}{
		{
			name:   "Process with disabled middleware",
			config: DefaultHealthCheckConfig(),
			req: map[string]interface{}{
				"path": "/health",
			},
			expectDisabled: true,
		},
		{
			name:   "Process health check request",
			config: DefaultHealthCheckConfig(),
			req: map[string]interface{}{
				"path": "/health",
			},
			expectHealth: true,
		},
		{
			name:   "Process non-health request",
			config: DefaultHealthCheckConfig(),
			req: map[string]interface{}{
				"path": "/api/data",
			},
		},
		{
			name:   "Process liveness check",
			config: DefaultHealthCheckConfig(),
			req: map[string]interface{}{
				"path": "/health/live",
			},
			expectHealth: true,
		},
		{
			name:   "Process readiness check",
			config: DefaultHealthCheckConfig(),
			req: map[string]interface{}{
				"path": "/health/ready",
			},
			expectHealth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hcm := NewHealthCheckMiddlewareWithConfig(1, tt.config)

			if tt.expectDisabled {
				hcm.SetEnabled(false)
			} else {
				hcm.SetEnabled(true)
			}

			ctx := context.Background()
			called := false
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				called = true
				return map[string]interface{}{"body": "test"}, nil
			}

			resp, err := hcm.Process(ctx, tt.req, next)

			if (err != nil) != tt.expectError {
				t.Errorf("Process() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectHealth {
				// Should not call next middleware for health check requests
				if called {
					t.Error("Process() should not call next middleware for health check requests")
				}

				// Verify response structure
				if resp == nil {
					t.Error("Process() returned nil response for health check")
					return
				}

				if httpResp, ok := resp.(map[string]interface{}); ok {
					if _, exists := httpResp["status_code"]; !exists {
						t.Error("Process() health response missing status_code")
					}
					if _, exists := httpResp["headers"]; !exists {
						t.Error("Process() health response missing headers")
					}
					if _, exists := httpResp["body"]; !exists {
						t.Error("Process() health response missing body")
					}
				} else {
					t.Error("Process() health response not in expected format")
				}
			} else if !tt.expectDisabled {
				// Should call next middleware for non-health requests
				if !called {
					t.Error("Process() should call next middleware for non-health requests")
				}
			}
		})
	}
}

// Test HealthChecker implementations

func TestNewDatabaseHealthChecker(t *testing.T) {
	tests := []struct {
		name     string
		dbName   string
		critical bool
	}{
		{
			name:     "Critical database checker",
			dbName:   "primary_db",
			critical: true,
		},
		{
			name:     "Non-critical database checker",
			dbName:   "cache_db",
			critical: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewDatabaseHealthChecker(tt.dbName, tt.critical)

			if checker.Name() != tt.dbName {
				t.Errorf("NewDatabaseHealthChecker().Name() = %v, want %v", checker.Name(), tt.dbName)
			}

			if checker.IsCritical() != tt.critical {
				t.Errorf("NewDatabaseHealthChecker().IsCritical() = %v, want %v", checker.IsCritical(), tt.critical)
			}

			if checker.GetTimeout() <= 0 {
				t.Error("NewDatabaseHealthChecker().GetTimeout() should be positive")
			}
		})
	}
}

func TestDatabaseHealthChecker_Check(t *testing.T) {
	checker := NewDatabaseHealthChecker("test_db", true)
	ctx := context.Background()

	result := checker.Check(ctx)

	if result.Name != "test_db" {
		t.Errorf("DatabaseHealthChecker.Check().Name = %v, want %v", result.Name, "test_db")
	}

	if result.Status != HealthStatusHealthy {
		t.Errorf("DatabaseHealthChecker.Check().Status = %v, want %v", result.Status, HealthStatusHealthy)
	}

	if result.Duration <= 0 {
		t.Error("DatabaseHealthChecker.Check().Duration should be positive")
	}

	if result.Timestamp.IsZero() {
		t.Error("DatabaseHealthChecker.Check().Timestamp should be set")
	}

	if result.Metadata == nil {
		t.Error("DatabaseHealthChecker.Check().Metadata should be set")
	}
}

func TestNewServiceHealthChecker(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		url         string
		critical    bool
	}{
		{
			name:        "Critical service checker",
			serviceName: "auth_service",
			url:         "http://auth.example.com",
			critical:    true,
		},
		{
			name:        "Non-critical service checker",
			serviceName: "analytics",
			url:         "http://analytics.example.com",
			critical:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewServiceHealthChecker(tt.serviceName, tt.url, tt.critical)

			if checker.Name() != tt.serviceName {
				t.Errorf("NewServiceHealthChecker().Name() = %v, want %v", checker.Name(), tt.serviceName)
			}

			if checker.IsCritical() != tt.critical {
				t.Errorf("NewServiceHealthChecker().IsCritical() = %v, want %v", checker.IsCritical(), tt.critical)
			}

			if checker.GetTimeout() <= 0 {
				t.Error("NewServiceHealthChecker().GetTimeout() should be positive")
			}
		})
	}
}

func TestServiceHealthChecker_Check(t *testing.T) {
	checker := NewServiceHealthChecker("test_service", "http://test.example.com", false)
	ctx := context.Background()

	result := checker.Check(ctx)

	if result.Name != "test_service" {
		t.Errorf("ServiceHealthChecker.Check().Name = %v, want %v", result.Name, "test_service")
	}

	// Status can be healthy or degraded due to simulation
	if result.Status != HealthStatusHealthy && result.Status != HealthStatusDegraded {
		t.Errorf("ServiceHealthChecker.Check().Status = %v, want healthy or degraded", result.Status)
	}

	if result.Duration <= 0 {
		t.Error("ServiceHealthChecker.Check().Duration should be positive")
	}

	if result.Metadata == nil {
		t.Error("ServiceHealthChecker.Check().Metadata should be set")
	}
}

func TestNewMemoryHealthChecker(t *testing.T) {
	tests := []struct {
		name      string
		threshold float64
		critical  bool
	}{
		{
			name:      "Critical memory checker",
			threshold: 0.8,
			critical:  true,
		},
		{
			name:      "Non-critical memory checker",
			threshold: 0.9,
			critical:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewMemoryHealthChecker(tt.name, tt.threshold, tt.critical)

			if checker.Name() != tt.name {
				t.Errorf("NewMemoryHealthChecker().Name() = %v, want %v", checker.Name(), tt.name)
			}

			if checker.IsCritical() != tt.critical {
				t.Errorf("NewMemoryHealthChecker().IsCritical() = %v, want %v", checker.IsCritical(), tt.critical)
			}

			if checker.GetTimeout() <= 0 {
				t.Error("NewMemoryHealthChecker().GetTimeout() should be positive")
			}
		})
	}
}

func TestMemoryHealthChecker_Check(t *testing.T) {
	checker := NewMemoryHealthChecker("memory", 0.8, true)
	ctx := context.Background()

	result := checker.Check(ctx)

	if result.Name != "memory" {
		t.Errorf("MemoryHealthChecker.Check().Name = %v, want %v", result.Name, "memory")
	}

	// Status should be healthy since simulated usage (0.65) is below threshold (0.8)
	if result.Status != HealthStatusHealthy {
		t.Errorf("MemoryHealthChecker.Check().Status = %v, want %v", result.Status, HealthStatusHealthy)
	}

	if result.Duration <= 0 {
		t.Error("MemoryHealthChecker.Check().Duration should be positive")
	}

	if result.Metadata == nil {
		t.Error("MemoryHealthChecker.Check().Metadata should be set")
	}

	// Check metadata contains expected fields
	if usage, exists := result.Metadata["usage"]; !exists {
		t.Error("MemoryHealthChecker.Check().Metadata should contain usage")
	} else if usage.(float64) <= 0 {
		t.Error("MemoryHealthChecker.Check().Metadata usage should be positive")
	}
}

// Benchmark tests
func BenchmarkHealthCheckMiddleware_Process_NonHealth(b *testing.B) {
	hcm := NewHealthCheckMiddleware(1)
	hcm.SetEnabled(true)

	ctx := context.Background()
	req := map[string]interface{}{
		"path": "/api/data",
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"body": "test"}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hcm.Process(ctx, req, next)
	}
}

func BenchmarkHealthCheckMiddleware_Process_HealthCheck(b *testing.B) {
	hcm := NewHealthCheckMiddleware(1)
	hcm.SetEnabled(true)

	ctx := context.Background()
	req := map[string]interface{}{
		"path": "/health",
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"body": "test"}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hcm.Process(ctx, req, next)
	}
}

func BenchmarkHealthChecker_DatabaseCheck(b *testing.B) {
	checker := NewDatabaseHealthChecker("test_db", true)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(ctx)
	}
}
