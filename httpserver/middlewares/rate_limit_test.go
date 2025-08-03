package middlewares

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewRateLimitMiddleware(t *testing.T) {
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
			got := NewRateLimitMiddleware(tt.priority)
			if got == nil {
				t.Error("NewRateLimitMiddleware() returned nil")
				return
			}

			if got.Name() != "rate_limit" {
				t.Errorf("NewRateLimitMiddleware().Name() = %v, want %v", got.Name(), "rate_limit")
			}

			if got.Priority() != tt.priority {
				t.Errorf("NewRateLimitMiddleware().Priority() = %v, want %v", got.Priority(), tt.priority)
			}

			// Verify default config is applied
			defaultConfig := DefaultRateLimitConfig()
			if !reflect.DeepEqual(got.config, defaultConfig) {
				t.Error("NewRateLimitMiddleware() did not apply default config")
			}

			// Verify limiters map is initialized
			if got.limiters == nil {
				t.Error("NewRateLimitMiddleware() did not initialize limiters map")
			}

			// Clean up
			got.Stop()
		})
	}
}

func TestNewRateLimitMiddlewareWithConfig(t *testing.T) {
	customConfig := RateLimitConfig{
		Strategy:          FixedWindow,
		RequestsPerSecond: 5.0,
		RequestsPerMinute: 300,
		RequestsPerHour:   18000,
		RequestsPerDay:    432000,
		BurstSize:         10,
		IdentifyByIP:      false,
		IdentifyByUser:    true,
		IdentifyByHeader:  "X-API-Key",
		IdentifyByQuery:   "token",
		PathLimits:        map[string]RateLimit{"/api/upload": {RequestsPerSecond: 1.0, BurstSize: 2}},
		MethodLimits:      map[string]RateLimit{"POST": {RequestsPerSecond: 2.0, BurstSize: 5}},
		SkipPaths:         []string{"/public"},
		SkipMethods:       []string{"HEAD"},
		SkipIPs:           []string{"192.168.1.1"},
		SkipUsers:         []string{"admin"},
		BlockDuration:     time.Hour,
		CleanupInterval:   time.Minute * 5,
		MemoryLimit:       2097152,
		RetryAfterHeader:  false,
		IncludeHeaders:    false,
		ErrorMessage:      "Too many requests",
		ErrorStatusCode:   503,
		SlidingWindow:     true,
		DistributedMode:   true,
		WhitelistMode:     true,
	}

	tests := []struct {
		name     string
		priority int
		config   RateLimitConfig
	}{
		{
			name:     "Create with custom config",
			priority: 3,
			config:   customConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRateLimitMiddlewareWithConfig(tt.priority, tt.config)
			if got == nil {
				t.Error("NewRateLimitMiddlewareWithConfig() returned nil")
				return
			}

			if !reflect.DeepEqual(got.config, tt.config) {
				t.Error("NewRateLimitMiddlewareWithConfig() did not apply custom config correctly")
			}

			// Clean up
			got.Stop()
		})
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	// Test strategy
	if config.Strategy != TokenBucket {
		t.Errorf("DefaultRateLimitConfig().Strategy = %v, want %v", config.Strategy, TokenBucket)
	}

	// Test rate limits
	if config.RequestsPerSecond != 10.0 {
		t.Errorf("DefaultRateLimitConfig().RequestsPerSecond = %v, want %v", config.RequestsPerSecond, 10.0)
	}
	if config.RequestsPerMinute != 600 {
		t.Errorf("DefaultRateLimitConfig().RequestsPerMinute = %v, want %v", config.RequestsPerMinute, 600)
	}
	if config.RequestsPerHour != 36000 {
		t.Errorf("DefaultRateLimitConfig().RequestsPerHour = %v, want %v", config.RequestsPerHour, 36000)
	}
	if config.RequestsPerDay != 864000 {
		t.Errorf("DefaultRateLimitConfig().RequestsPerDay = %v, want %v", config.RequestsPerDay, 864000)
	}
	if config.BurstSize != 20 {
		t.Errorf("DefaultRateLimitConfig().BurstSize = %v, want %v", config.BurstSize, 20)
	}

	// Test identification settings
	if !config.IdentifyByIP {
		t.Error("DefaultRateLimitConfig().IdentifyByIP should be true")
	}
	if config.IdentifyByUser {
		t.Error("DefaultRateLimitConfig().IdentifyByUser should be false")
	}

	// Test skip lists
	expectedSkipPaths := []string{"/health", "/metrics"}
	if !reflect.DeepEqual(config.SkipPaths, expectedSkipPaths) {
		t.Errorf("DefaultRateLimitConfig().SkipPaths = %v, want %v", config.SkipPaths, expectedSkipPaths)
	}

	expectedSkipMethods := []string{"OPTIONS"}
	if !reflect.DeepEqual(config.SkipMethods, expectedSkipMethods) {
		t.Errorf("DefaultRateLimitConfig().SkipMethods = %v, want %v", config.SkipMethods, expectedSkipMethods)
	}

	expectedSkipIPs := []string{"127.0.0.1", "::1"}
	if !reflect.DeepEqual(config.SkipIPs, expectedSkipIPs) {
		t.Errorf("DefaultRateLimitConfig().SkipIPs = %v, want %v", config.SkipIPs, expectedSkipIPs)
	}

	// Test timeouts
	if config.BlockDuration != time.Minute*5 {
		t.Errorf("DefaultRateLimitConfig().BlockDuration = %v, want %v", config.BlockDuration, time.Minute*5)
	}
	if config.CleanupInterval != time.Minute*10 {
		t.Errorf("DefaultRateLimitConfig().CleanupInterval = %v, want %v", config.CleanupInterval, time.Minute*10)
	}

	// Test error settings
	if config.ErrorStatusCode != 429 {
		t.Errorf("DefaultRateLimitConfig().ErrorStatusCode = %v, want %v", config.ErrorStatusCode, 429)
	}
	if config.ErrorMessage != "Rate limit exceeded" {
		t.Errorf("DefaultRateLimitConfig().ErrorMessage = %v, want %v", config.ErrorMessage, "Rate limit exceeded")
	}

	// Test header settings
	if !config.RetryAfterHeader {
		t.Error("DefaultRateLimitConfig().RetryAfterHeader should be true")
	}
	if !config.IncludeHeaders {
		t.Error("DefaultRateLimitConfig().IncludeHeaders should be true")
	}

	// Test feature flags
	if config.SlidingWindow {
		t.Error("DefaultRateLimitConfig().SlidingWindow should be false")
	}
	if config.DistributedMode {
		t.Error("DefaultRateLimitConfig().DistributedMode should be false")
	}
	if config.WhitelistMode {
		t.Error("DefaultRateLimitConfig().WhitelistMode should be false")
	}
}

func TestRateLimitMiddleware_GetConfig(t *testing.T) {
	customConfig := RateLimitConfig{
		Strategy:          SlidingWindow,
		RequestsPerSecond: 15.0,
		BurstSize:         30,
		IdentifyByUser:    true,
		ErrorStatusCode:   503,
	}

	rlm := NewRateLimitMiddlewareWithConfig(1, customConfig)
	defer rlm.Stop()

	got := rlm.GetConfig()

	if !reflect.DeepEqual(got, customConfig) {
		t.Error("GetConfig() did not return the correct configuration")
	}
}

func TestRateLimitMiddleware_SetConfig(t *testing.T) {
	rlm := NewRateLimitMiddleware(1)
	defer rlm.Stop()

	newConfig := RateLimitConfig{
		Strategy:          FixedWindow,
		RequestsPerSecond: 25.0,
		BurstSize:         50,
		IdentifyByHeader:  "X-Client-ID",
		ErrorStatusCode:   503,
		ErrorMessage:      "Request quota exceeded",
	}

	rlm.SetConfig(newConfig)
	got := rlm.GetConfig()

	if !reflect.DeepEqual(got, newConfig) {
		t.Error("SetConfig() did not update the configuration correctly")
	}
}

func TestRateLimitMiddleware_extractRequestInfo(t *testing.T) {
	rlm := NewRateLimitMiddleware(1)
	defer rlm.Stop()

	tests := []struct {
		name string
		req  interface{}
		want *RateLimitRequestInfo
	}{
		{
			name: "Valid HTTP request",
			req: map[string]interface{}{
				"method":  "POST",
				"path":    "/api/data",
				"ip":      "192.168.1.100",
				"user_id": "user123",
				"headers": map[string]string{
					"X-API-Key":  "abc123",
					"User-Agent": "test-client",
				},
				"query": map[string]string{
					"token": "xyz789",
					"limit": "10",
				},
			},
			want: &RateLimitRequestInfo{
				Method: "POST",
				Path:   "/api/data",
				IP:     "192.168.1.100",
				UserID: "user123",
				Headers: map[string]string{
					"X-API-Key":  "abc123",
					"User-Agent": "test-client",
				},
				Query: map[string]string{
					"token": "xyz789",
					"limit": "10",
				},
			},
		},
		{
			name: "Minimal request",
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/",
			},
			want: &RateLimitRequestInfo{
				Method:  "GET",
				Path:    "/",
				Headers: map[string]string{},
				Query:   map[string]string{},
			},
		},
		{
			name: "Invalid request type",
			req:  "invalid",
			want: &RateLimitRequestInfo{
				Headers: map[string]string{},
				Query:   map[string]string{},
			},
		},
		{
			name: "Request with non-string values",
			req: map[string]interface{}{
				"method":  123,
				"path":    456,
				"ip":      789,
				"user_id": 101112,
			},
			want: &RateLimitRequestInfo{
				Headers: map[string]string{},
				Query:   map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rlm.extractRequestInfo(tt.req)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractRequestInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimitMiddleware_shouldSkipRateLimit(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.SkipPaths = []string{"/health", "/metrics"}
	config.SkipMethods = []string{"OPTIONS", "HEAD"}
	config.SkipIPs = []string{"127.0.0.1", "192.168.1.10"}
	config.SkipUsers = []string{"admin", "system"}

	rlm := NewRateLimitMiddlewareWithConfig(1, config)
	defer rlm.Stop()

	tests := []struct {
		name    string
		reqInfo *RateLimitRequestInfo
		want    bool
	}{
		{
			name: "Skip path in skip list",
			reqInfo: &RateLimitRequestInfo{
				Path: "/health",
			},
			want: true,
		},
		{
			name: "Skip method in skip list",
			reqInfo: &RateLimitRequestInfo{
				Method: "OPTIONS",
				Path:   "/api/data",
			},
			want: true,
		},
		{
			name: "Skip IP in skip list",
			reqInfo: &RateLimitRequestInfo{
				Method: "GET",
				Path:   "/api/data",
				IP:     "127.0.0.1",
			},
			want: true,
		},
		{
			name: "Skip user in skip list",
			reqInfo: &RateLimitRequestInfo{
				Method: "GET",
				Path:   "/api/data",
				IP:     "192.168.1.100",
				UserID: "admin",
			},
			want: true,
		},
		{
			name: "Do not skip regular request",
			reqInfo: &RateLimitRequestInfo{
				Method: "GET",
				Path:   "/api/data",
				IP:     "192.168.1.100",
				UserID: "user123",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rlm.shouldSkipRateLimit(tt.reqInfo)
			if got != tt.want {
				t.Errorf("shouldSkipRateLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimitMiddleware_generateKey(t *testing.T) {
	tests := []struct {
		name    string
		config  RateLimitConfig
		reqInfo *RateLimitRequestInfo
		want    string
	}{
		{
			name: "Identify by IP",
			config: RateLimitConfig{
				IdentifyByIP: true,
			},
			reqInfo: &RateLimitRequestInfo{
				IP: "192.168.1.100",
			},
			want: "ip:192.168.1.100",
		},
		{
			name: "Identify by User",
			config: RateLimitConfig{
				IdentifyByUser: true,
			},
			reqInfo: &RateLimitRequestInfo{
				UserID: "user123",
			},
			want: "user:user123",
		},
		{
			name: "Identify by Header",
			config: RateLimitConfig{
				IdentifyByHeader: "X-API-Key",
			},
			reqInfo: &RateLimitRequestInfo{
				Headers: map[string]string{
					"X-API-Key": "abc123",
				},
			},
			want: "header:abc123",
		},
		{
			name: "Identify by Query",
			config: RateLimitConfig{
				IdentifyByQuery: "token",
			},
			reqInfo: &RateLimitRequestInfo{
				Query: map[string]string{
					"token": "xyz789",
				},
			},
			want: "query:xyz789",
		},
		{
			name: "Multiple identification methods",
			config: RateLimitConfig{
				IdentifyByIP:     true,
				IdentifyByUser:   true,
				IdentifyByHeader: "X-API-Key",
			},
			reqInfo: &RateLimitRequestInfo{
				IP:     "192.168.1.100",
				UserID: "user123",
				Headers: map[string]string{
					"X-API-Key": "abc123",
				},
			},
			want: "ip:192.168.1.100|user:user123|header:abc123",
		},
		{
			name: "Default to IP when no methods configured",
			config: RateLimitConfig{
				IdentifyByIP:   false,
				IdentifyByUser: false,
			},
			reqInfo: &RateLimitRequestInfo{
				IP: "192.168.1.100",
			},
			want: "ip:192.168.1.100",
		},
		{
			name: "Anonymous when no identification possible",
			config: RateLimitConfig{
				IdentifyByIP:   false,
				IdentifyByUser: false,
			},
			reqInfo: &RateLimitRequestInfo{},
			want:    "anonymous",
		},
		{
			name: "Custom key function",
			config: RateLimitConfig{
				KeyFunc: func(ctx context.Context, req interface{}) string {
					return "custom:key"
				},
			},
			reqInfo: &RateLimitRequestInfo{
				IP: "192.168.1.100",
			},
			want: "custom:key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rlm := NewRateLimitMiddlewareWithConfig(1, tt.config)
			defer rlm.Stop()

			ctx := context.Background()
			req := map[string]interface{}{}

			got := rlm.generateKey(ctx, req, tt.reqInfo)
			if got != tt.want {
				t.Errorf("generateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimitMiddleware_getRateLimitForRequest(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.PathLimits = map[string]RateLimit{
		"/api/upload": {RequestsPerSecond: 1.0, BurstSize: 2},
	}
	config.MethodLimits = map[string]RateLimit{
		"POST": {RequestsPerSecond: 5.0, BurstSize: 10},
	}

	rlm := NewRateLimitMiddlewareWithConfig(1, config)
	defer rlm.Stop()

	tests := []struct {
		name    string
		reqInfo *RateLimitRequestInfo
		want    RateLimit
	}{
		{
			name: "Path-specific limit",
			reqInfo: &RateLimitRequestInfo{
				Path:   "/api/upload",
				Method: "POST",
			},
			want: RateLimit{RequestsPerSecond: 1.0, BurstSize: 2},
		},
		{
			name: "Method-specific limit",
			reqInfo: &RateLimitRequestInfo{
				Path:   "/api/data",
				Method: "POST",
			},
			want: RateLimit{RequestsPerSecond: 5.0, BurstSize: 10},
		},
		{
			name: "Default limit",
			reqInfo: &RateLimitRequestInfo{
				Path:   "/api/data",
				Method: "GET",
			},
			want: RateLimit{
				RequestsPerSecond: config.RequestsPerSecond,
				RequestsPerMinute: config.RequestsPerMinute,
				RequestsPerHour:   config.RequestsPerHour,
				RequestsPerDay:    config.RequestsPerDay,
				BurstSize:         config.BurstSize,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rlm.getRateLimitForRequest(tt.reqInfo)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRateLimitForRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimitMiddleware_Reset(t *testing.T) {
	rlm := NewRateLimitMiddleware(1)
	defer rlm.Stop()

	// Simulate some metrics and limiters
	rlm.totalRequests = 100
	rlm.allowedRequests = 80
	rlm.blockedRequests = 20

	// Add a limiter
	reqInfo := &RateLimitRequestInfo{IP: "192.168.1.100"}
	limiter := rlm.getLimiter("test", reqInfo)
	_ = limiter // Use the limiter

	rlm.Reset()

	metrics := rlm.GetMetrics()
	if metrics["total_requests"].(int64) != 0 {
		t.Error("Reset() did not reset total_requests")
	}
	if metrics["allowed_requests"].(int64) != 0 {
		t.Error("Reset() did not reset allowed_requests")
	}
	if metrics["blocked_requests"].(int64) != 0 {
		t.Error("Reset() did not reset blocked_requests")
	}
	if metrics["active_limiters"].(int) != 0 {
		t.Error("Reset() did not clear limiters")
	}
	if metrics["reset_count"].(int64) != 1 {
		t.Error("Reset() did not increment reset_count")
	}
}

func TestRateLimitMiddleware_GetMetrics(t *testing.T) {
	rlm := NewRateLimitMiddleware(1)
	defer rlm.Stop()

	// Initial metrics should be zero
	metrics := rlm.GetMetrics()

	expectedKeys := []string{
		"total_requests", "allowed_requests", "blocked_requests",
		"allowed_rate", "blocked_rate", "active_limiters", "reset_count", "uptime",
	}

	for _, key := range expectedKeys {
		if _, exists := metrics[key]; !exists {
			t.Errorf("GetMetrics() missing key: %s", key)
		}
	}

	// Test with some data
	rlm.totalRequests = 100
	rlm.allowedRequests = 90
	rlm.blockedRequests = 10

	// Add some limiters
	reqInfo1 := &RateLimitRequestInfo{IP: "192.168.1.100"}
	reqInfo2 := &RateLimitRequestInfo{IP: "192.168.1.101"}
	rlm.getLimiter("test1", reqInfo1)
	rlm.getLimiter("test2", reqInfo2)

	metrics = rlm.GetMetrics()

	if metrics["total_requests"].(int64) != 100 {
		t.Errorf("GetMetrics() total_requests = %v, want %v", metrics["total_requests"], 100)
	}

	allowedRate := metrics["allowed_rate"].(float64)
	if allowedRate != 90.0 {
		t.Errorf("GetMetrics() allowed_rate = %v, want %v", allowedRate, 90.0)
	}

	blockedRate := metrics["blocked_rate"].(float64)
	if blockedRate != 10.0 {
		t.Errorf("GetMetrics() blocked_rate = %v, want %v", blockedRate, 10.0)
	}

	activeLimiters := metrics["active_limiters"].(int)
	if activeLimiters != 2 {
		t.Errorf("GetMetrics() active_limiters = %v, want %v", activeLimiters, 2)
	}
}

func TestRateLimitMiddleware_Process(t *testing.T) {
	tests := []struct {
		name           string
		config         RateLimitConfig
		req            interface{}
		expectDisabled bool
		expectSkipped  bool
		expectBlocked  bool
		expectError    bool
	}{
		{
			name:   "Process with disabled middleware",
			config: DefaultRateLimitConfig(),
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
				"ip":     "192.168.1.100",
			},
			expectDisabled: true,
		},
		{
			name:   "Process skip for health check path",
			config: DefaultRateLimitConfig(),
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/health",
				"ip":     "192.168.1.100",
			},
			expectSkipped: true,
		},
		{
			name:   "Process skip for OPTIONS method",
			config: DefaultRateLimitConfig(),
			req: map[string]interface{}{
				"method": "OPTIONS",
				"path":   "/api/data",
				"ip":     "192.168.1.100",
			},
			expectSkipped: true,
		},
		{
			name:   "Process regular request",
			config: DefaultRateLimitConfig(),
			req: map[string]interface{}{
				"method": "GET",
				"path":   "/api/data",
				"ip":     "192.168.1.100",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rlm := NewRateLimitMiddlewareWithConfig(1, tt.config)
			defer rlm.Stop()

			if tt.expectDisabled {
				rlm.SetEnabled(false)
			} else {
				rlm.SetEnabled(true)
			}

			ctx := context.Background()
			called := false
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				called = true
				return map[string]interface{}{"body": "test"}, nil
			}

			resp, err := rlm.Process(ctx, tt.req, next)

			if (err != nil) != tt.expectError {
				t.Errorf("Process() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if tt.expectDisabled || tt.expectSkipped {
				if !called {
					t.Error("Process() should call next middleware for disabled/skipped requests")
				}
			} else if !tt.expectBlocked {
				if !called {
					t.Error("Process() should call next middleware for allowed requests")
				}
			}

			if resp == nil {
				t.Error("Process() returned nil response")
			}

			// Verify response structure for non-error cases
			if !tt.expectError && resp != nil {
				if httpResp, ok := resp.(map[string]interface{}); ok {
					if tt.expectBlocked {
						// Should be an error response
						if statusCode, exists := httpResp["status_code"]; exists {
							if statusCode != tt.config.ErrorStatusCode {
								t.Errorf("Process() blocked response status_code = %v, want %v", statusCode, tt.config.ErrorStatusCode)
							}
						}
					}
				}
			}
		})
	}
}

// Test RateLimiter implementations

func TestNewRateLimiter(t *testing.T) {
	config := RateLimit{
		RequestsPerSecond: 10.0,
		RequestsPerMinute: 600,
		BurstSize:         20,
	}

	tests := []struct {
		name     string
		key      string
		config   RateLimit
		strategy RateLimitStrategy
	}{
		{
			name:     "Token bucket limiter",
			key:      "test:key",
			config:   config,
			strategy: TokenBucket,
		},
		{
			name:     "Fixed window limiter",
			key:      "test:key",
			config:   config,
			strategy: FixedWindow,
		},
		{
			name:     "Sliding window limiter",
			key:      "test:key",
			config:   config,
			strategy: SlidingWindow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.key, tt.config, tt.strategy)

			if limiter.key != tt.key {
				t.Errorf("NewRateLimiter().key = %v, want %v", limiter.key, tt.key)
			}

			if !reflect.DeepEqual(limiter.config, tt.config) {
				t.Error("NewRateLimiter() did not set config correctly")
			}

			if limiter.strategy != tt.strategy {
				t.Errorf("NewRateLimiter().strategy = %v, want %v", limiter.strategy, tt.strategy)
			}

			// Verify initial state
			if limiter.tokens != float64(tt.config.BurstSize) {
				t.Errorf("NewRateLimiter().tokens = %v, want %v", limiter.tokens, float64(tt.config.BurstSize))
			}

			if limiter.requests == nil {
				t.Error("NewRateLimiter() did not initialize requests slice")
			}
		})
	}
}

func TestRateLimiter_checkTokenBucket(t *testing.T) {
	config := RateLimit{
		RequestsPerSecond: 2.0, // 2 requests per second
		BurstSize:         5,   // Allow burst of 5
	}

	limiter := NewRateLimiter("test", config, TokenBucket)

	// Test initial requests (should be allowed due to burst capacity)
	for i := 0; i < 5; i++ {
		ctx := limiter.checkLimit()
		if !ctx.Allowed {
			t.Errorf("Request %d should be allowed (burst capacity)", i+1)
		}
	}

	// Next request should be blocked (burst capacity exhausted)
	ctx := limiter.checkLimit()
	if ctx.Allowed {
		t.Error("Request should be blocked after burst capacity exhausted")
	}

	// Wait and test token refill
	time.Sleep(time.Millisecond * 600) // Wait for ~1.2 tokens to be added
	ctx = limiter.checkLimit()
	if !ctx.Allowed {
		t.Error("Request should be allowed after token refill")
	}
}

func TestRateLimiter_checkFixedWindow(t *testing.T) {
	config := RateLimit{
		RequestsPerSecond: 2.0, // 2 requests per second
	}

	limiter := NewRateLimiter("test", config, FixedWindow)

	// Test requests within window
	ctx1 := limiter.checkLimit()
	if !ctx1.Allowed {
		t.Error("First request should be allowed")
	}

	ctx2 := limiter.checkLimit()
	if !ctx2.Allowed {
		t.Error("Second request should be allowed")
	}

	// Third request should be blocked (limit reached)
	ctx3 := limiter.checkLimit()
	if ctx3.Allowed {
		t.Error("Third request should be blocked")
	}

	// Check context values - remaining is calculated before incrementing
	if ctx1.Remaining != 2 { // Initially 2 requests available, after first request = 2
		t.Errorf("First request remaining = %v, want %v", ctx1.Remaining, 2)
	}

	if ctx2.Remaining != 1 { // After first request, 1 remaining, after second = 1
		t.Errorf("Second request remaining = %v, want %v", ctx2.Remaining, 1)
	}

	if ctx3.Remaining != 0 {
		t.Errorf("Third request remaining = %v, want %v", ctx3.Remaining, 0)
	}
}

func TestRateLimiter_checkSlidingWindow(t *testing.T) {
	config := RateLimit{
		RequestsPerMinute: 3, // 3 requests per minute
	}

	limiter := NewRateLimiter("test", config, SlidingWindow)

	// Test requests within limit
	for i := 0; i < 3; i++ {
		ctx := limiter.checkLimit()
		if !ctx.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Fourth request should be blocked
	ctx := limiter.checkLimit()
	if ctx.Allowed {
		t.Error("Fourth request should be blocked")
	}

	if ctx.Remaining != 0 {
		t.Errorf("Blocked request remaining = %v, want %v", ctx.Remaining, 0)
	}
}

func TestRateLimitMiddleware_createRateLimitErrorResponse(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.ErrorStatusCode = 429
	config.ErrorMessage = "Too many requests"
	config.RetryAfterHeader = true
	config.IncludeHeaders = true

	rlm := NewRateLimitMiddlewareWithConfig(1, config)
	defer rlm.Stop()

	rateLimitCtx := &RateLimitContext{
		Allowed:    false,
		Remaining:  0,
		ResetTime:  time.Now().Add(time.Minute),
		RetryAfter: time.Second * 30,
	}

	resp, err := rlm.createRateLimitErrorResponse(rateLimitCtx)
	if err != nil {
		t.Errorf("createRateLimitErrorResponse() error = %v, want nil", err)
		return
	}

	httpResp, ok := resp.(map[string]interface{})
	if !ok {
		t.Error("createRateLimitErrorResponse() did not return map[string]interface{}")
		return
	}

	if httpResp["status_code"] != 429 {
		t.Errorf("createRateLimitErrorResponse() status_code = %v, want %v", httpResp["status_code"], 429)
	}

	if headers, exists := httpResp["headers"]; exists {
		if h, ok := headers.(map[string]string); ok {
			if h["Content-Type"] != "application/json" {
				t.Error("createRateLimitErrorResponse() missing or incorrect Content-Type header")
			}
			if h["Retry-After"] != "30" {
				t.Errorf("createRateLimitErrorResponse() Retry-After = %v, want %v", h["Retry-After"], "30")
			}
			if _, exists := h["X-RateLimit-Limit"]; !exists {
				t.Error("createRateLimitErrorResponse() missing X-RateLimit-Limit header")
			}
		}
	}

	body := httpResp["body"].(string)
	if !strings.Contains(body, "Too many requests") {
		t.Error("createRateLimitErrorResponse() body does not contain error message")
	}
}

func TestRateLimitMiddleware_addRateLimitHeaders(t *testing.T) {
	config := DefaultRateLimitConfig()
	rlm := NewRateLimitMiddlewareWithConfig(1, config)
	defer rlm.Stop()

	rateLimitCtx := &RateLimitContext{
		Remaining: 5,
		ResetTime: time.Unix(1609459200, 0), // Fixed timestamp for testing
	}

	tests := []struct {
		name string
		resp interface{}
		want map[string]string
	}{
		{
			name: "Add headers to response with existing headers",
			resp: map[string]interface{}{
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
				"body": "test",
			},
			want: map[string]string{
				"Content-Type":          "application/json",
				"X-RateLimit-Limit":     "10",
				"X-RateLimit-Remaining": "5",
				"X-RateLimit-Reset":     "1609459200",
			},
		},
		{
			name: "Add headers to response without existing headers",
			resp: map[string]interface{}{
				"body": "test",
			},
			want: map[string]string{
				"X-RateLimit-Limit":     "10",
				"X-RateLimit-Remaining": "5",
				"X-RateLimit-Reset":     "1609459200",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rlm.addRateLimitHeaders(tt.resp, rateLimitCtx)

			if httpResp, ok := tt.resp.(map[string]interface{}); ok {
				if headers, exists := httpResp["headers"]; exists {
					if h, ok := headers.(map[string]string); ok {
						for key, expectedValue := range tt.want {
							if actualValue, exists := h[key]; !exists || actualValue != expectedValue {
								t.Errorf("addRateLimitHeaders() header %s = %v, want %v", key, actualValue, expectedValue)
							}
						}
					} else {
						t.Error("addRateLimitHeaders() headers not in expected format")
					}
				} else {
					t.Error("addRateLimitHeaders() did not add headers")
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test min function
	if result := min(5.0, 3.0); result != 3.0 {
		t.Errorf("min(5.0, 3.0) = %v, want %v", result, 3.0)
	}
	if result := min(2.0, 7.0); result != 2.0 {
		t.Errorf("min(2.0, 7.0) = %v, want %v", result, 2.0)
	}

	// Test max function
	if result := max(5, 3); result != 5 {
		t.Errorf("max(5, 3) = %v, want %v", result, 5)
	}
	if result := max(2, 7); result != 7 {
		t.Errorf("max(2, 7) = %v, want %v", result, 7)
	}
}

// Benchmark tests
func BenchmarkRateLimitMiddleware_Process_Allowed(b *testing.B) {
	config := DefaultRateLimitConfig()
	config.RequestsPerSecond = 1000.0 // High limit to avoid blocking
	config.BurstSize = 1000

	rlm := NewRateLimitMiddlewareWithConfig(1, config)
	defer rlm.Stop()
	rlm.SetEnabled(true)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/api/data",
		"ip":     "192.168.1.100",
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"body": "test"}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rlm.Process(ctx, req, next)
	}
}

func BenchmarkRateLimitMiddleware_Process_Skipped(b *testing.B) {
	rlm := NewRateLimitMiddleware(1)
	defer rlm.Stop()
	rlm.SetEnabled(true)

	ctx := context.Background()
	req := map[string]interface{}{
		"method": "GET",
		"path":   "/health", // Skip path
		"ip":     "192.168.1.100",
	}

	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		return map[string]interface{}{"body": "test"}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rlm.Process(ctx, req, next)
	}
}

func BenchmarkRateLimiter_TokenBucket(b *testing.B) {
	config := RateLimit{
		RequestsPerSecond: 1000.0,
		BurstSize:         1000,
	}

	limiter := NewRateLimiter("benchmark", config, TokenBucket)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.checkLimit()
	}
}

func BenchmarkRateLimiter_FixedWindow(b *testing.B) {
	config := RateLimit{
		RequestsPerSecond: 1000.0,
	}

	limiter := NewRateLimiter("benchmark", config, FixedWindow)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.checkLimit()
	}
}
