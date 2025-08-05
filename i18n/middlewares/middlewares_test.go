package middlewares

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	cache := NewMemoryCache(100)

	if cache.Size() != 0 {
		t.Errorf("expected empty cache, got size %d", cache.Size())
	}

	stats := cache.Stats()
	if stats.MaxSize != 100 {
		t.Errorf("expected max size 100, got %d", stats.MaxSize)
	}
}

func TestMemoryCacheSetGet(t *testing.T) {
	cache := NewMemoryCache(10)
	ttl := 1 * time.Hour

	// Test setting and getting a value
	err := cache.Set("key1", "value1", ttl)
	if err != nil {
		t.Errorf("expected no error setting value, got: %v", err)
	}

	value, found := cache.Get("key1")
	if !found {
		t.Error("expected to find key1, but it was not found")
	}

	if value != "value1" {
		t.Errorf("expected value 'value1', got '%v'", value)
	}

	// Test getting non-existent key
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("expected not to find nonexistent key, but it was found")
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	cache := NewMemoryCache(10)
	shortTTL := 10 * time.Millisecond

	// Set a value with short TTL
	err := cache.Set("expiring", "value", shortTTL)
	if err != nil {
		t.Errorf("expected no error setting value, got: %v", err)
	}

	// Should be available immediately
	value, found := cache.Get("expiring")
	if !found || value != "value" {
		t.Error("expected to find value immediately after setting")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should not be available after expiration
	_, found = cache.Get("expiring")
	if found {
		t.Error("expected value to be expired, but it was still found")
	}
}

func TestMemoryCacheEviction(t *testing.T) {
	cache := NewMemoryCache(2) // Small cache for testing eviction
	ttl := 1 * time.Hour

	// Fill the cache
	cache.Set("key1", "value1", ttl)
	cache.Set("key2", "value2", ttl)

	if cache.Size() != 2 {
		t.Errorf("expected cache size 2, got %d", cache.Size())
	}

	// Add one more item, should trigger eviction
	cache.Set("key3", "value3", ttl)

	if cache.Size() != 2 {
		t.Errorf("expected cache size to remain 2 after eviction, got %d", cache.Size())
	}
}

func TestMemoryCacheDelete(t *testing.T) {
	cache := NewMemoryCache(10)
	ttl := 1 * time.Hour

	cache.Set("key1", "value1", ttl)
	cache.Set("key2", "value2", ttl)

	if cache.Size() != 2 {
		t.Errorf("expected cache size 2, got %d", cache.Size())
	}

	err := cache.Delete("key1")
	if err != nil {
		t.Errorf("expected no error deleting key, got: %v", err)
	}

	if cache.Size() != 1 {
		t.Errorf("expected cache size 1 after delete, got %d", cache.Size())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("expected key1 to be deleted, but it was still found")
	}

	_, found = cache.Get("key2")
	if !found {
		t.Error("expected key2 to still exist after deleting key1")
	}
}

func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache(10)
	ttl := 1 * time.Hour

	cache.Set("key1", "value1", ttl)
	cache.Set("key2", "value2", ttl)
	cache.Set("key3", "value3", ttl)

	if cache.Size() != 3 {
		t.Errorf("expected cache size 3, got %d", cache.Size())
	}

	err := cache.Clear()
	if err != nil {
		t.Errorf("expected no error clearing cache, got: %v", err)
	}

	if cache.Size() != 0 {
		t.Errorf("expected cache size 0 after clear, got %d", cache.Size())
	}

	stats := cache.Stats()
	if stats.HitCount != 0 || stats.MissCount != 0 {
		t.Errorf("expected stats to be reset after clear, got hit=%d, miss=%d", stats.HitCount, stats.MissCount)
	}
}

func TestMemoryCacheStats(t *testing.T) {
	cache := NewMemoryCache(10)
	ttl := 1 * time.Hour

	cache.Set("key1", "value1", ttl)
	cache.Set("key2", "value2", ttl)

	// Hit
	cache.Get("key1")
	// Miss
	cache.Get("nonexistent")

	stats := cache.Stats()
	if stats.HitCount != 1 {
		t.Errorf("expected 1 hit, got %d", stats.HitCount)
	}

	if stats.MissCount != 1 {
		t.Errorf("expected 1 miss, got %d", stats.MissCount)
	}

	if stats.Size != 2 {
		t.Errorf("expected size 2, got %d", stats.Size)
	}

	if stats.MaxSize != 10 {
		t.Errorf("expected max size 10, got %d", stats.MaxSize)
	}
}

func TestNewCachingMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		mwName      string
		config      CachingMiddlewareConfig
		cache       Cache
		expectError bool
		errorMsg    string
	}{
		{
			name:   "valid middleware with custom cache",
			mwName: "test-caching",
			config: CachingMiddlewareConfig{
				TTL:              30 * time.Minute,
				MaxSize:          500,
				CacheKeyPrefix:   "test:",
				EnableStats:      true,
				CacheNullResults: false,
			},
			cache:       NewMemoryCache(500),
			expectError: false,
		},
		{
			name:   "valid middleware with default cache",
			mwName: "test-caching",
			config: CachingMiddlewareConfig{
				MaxSize: 200,
			},
			cache:       nil, // should create default cache
			expectError: false,
		},
		{
			name:        "empty middleware name",
			mwName:      "",
			config:      CachingMiddlewareConfig{},
			cache:       nil,
			expectError: true,
			errorMsg:    "middleware name cannot be empty",
		},
		{
			name:   "config with defaults",
			mwName: "test-caching",
			config: CachingMiddlewareConfig{
				// Empty config should get defaults
			},
			cache:       nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, err := NewCachingMiddleware(tt.mwName, tt.config, tt.cache)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}

				if mw == nil {
					t.Error("expected middleware to be created, but got nil")
					return
				}

				if mw.Name() != tt.mwName {
					t.Errorf("expected middleware name '%s', got '%s'", tt.mwName, mw.Name())
				}
			}
		})
	}
}

func TestCachingMiddlewareWrapTranslate(t *testing.T) {
	cache := NewMemoryCache(100)
	config := CachingMiddlewareConfig{
		TTL:              1 * time.Hour,
		MaxSize:          100,
		CacheKeyPrefix:   "test:",
		EnableStats:      true,
		CacheNullResults: false,
	}

	mw, err := NewCachingMiddleware("test-caching", config, cache)
	if err != nil {
		t.Fatalf("failed to create caching middleware: %v", err)
	}

	// Mock translation function
	callCount := 0
	mockTranslate := func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		callCount++
		return fmt.Sprintf("translated_%s_%s", key, lang), nil
	}

	wrappedTranslate := mw.WrapTranslate(mockTranslate)

	ctx := context.Background()

	// First call should hit the mock function and cache the result
	result1, err := wrappedTranslate(ctx, "test.key", "en", nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result1 != "translated_test.key_en" {
		t.Errorf("expected 'translated_test.key_en', got '%s'", result1)
	}
	if callCount != 1 {
		t.Errorf("expected mock function to be called once, got %d calls", callCount)
	}

	// Second call with same parameters should use cache
	result2, err := wrappedTranslate(ctx, "test.key", "en", nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result2 != result1 {
		t.Errorf("expected cached result '%s', got '%s'", result1, result2)
	}
	if callCount != 1 {
		t.Errorf("expected mock function to still be called once (cached), got %d calls", callCount)
	}

	// Call with different parameters should hit the mock function again
	result3, err := wrappedTranslate(ctx, "test.key", "es", nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result3 != "translated_test.key_es" {
		t.Errorf("expected 'translated_test.key_es', got '%s'", result3)
	}
	if callCount != 2 {
		t.Errorf("expected mock function to be called twice, got %d calls", callCount)
	}
}

func TestCachingMiddlewareCacheNullResults(t *testing.T) {
	cache := NewMemoryCache(100)
	config := CachingMiddlewareConfig{
		TTL:              1 * time.Hour,
		CacheNullResults: true,
	}

	mw, err := NewCachingMiddleware("test-caching", config, cache)
	if err != nil {
		t.Fatalf("failed to create caching middleware: %v", err)
	}

	callCount := 0
	mockTranslate := func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		callCount++
		return "", nil // Return empty result
	}

	wrappedTranslate := mw.WrapTranslate(mockTranslate)
	ctx := context.Background()

	// First call should cache empty result
	result1, err := wrappedTranslate(ctx, "empty.key", "en", nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result1 != "" {
		t.Errorf("expected empty result, got '%s'", result1)
	}
	if callCount != 1 {
		t.Errorf("expected mock function to be called once, got %d calls", callCount)
	}

	// Second call should use cached empty result
	result2, err := wrappedTranslate(ctx, "empty.key", "en", nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result2 != "" {
		t.Errorf("expected cached empty result, got '%s'", result2)
	}
	if callCount != 1 {
		t.Errorf("expected mock function to still be called once (cached), got %d calls", callCount)
	}
}

func TestNewSimpleRateLimiter(t *testing.T) {
	limiter := NewSimpleRateLimiter(10, 20)

	// Test that it allows requests initially
	allowed := limiter.Allow("test-key")
	if !allowed {
		t.Error("expected rate limiter to allow initial request")
	}

	stats := limiter.GetStats("test-key")
	if stats.RequestCount != 1 {
		t.Errorf("expected request count 1, got %d", stats.RequestCount)
	}
	if stats.AllowedCount != 1 {
		t.Errorf("expected allowed count 1, got %d", stats.AllowedCount)
	}
	if stats.BlockedCount != 0 {
		t.Errorf("expected blocked count 0, got %d", stats.BlockedCount)
	}
}

func TestSimpleRateLimiterBurstLimit(t *testing.T) {
	limiter := NewSimpleRateLimiter(1, 2) // 1 req/sec, burst of 2

	// Should allow burst requests
	allowed1 := limiter.Allow("test-key")
	allowed2 := limiter.Allow("test-key")

	if !allowed1 || !allowed2 {
		t.Error("expected rate limiter to allow burst requests")
	}

	// Third request should be blocked
	allowed3 := limiter.Allow("test-key")
	if allowed3 {
		t.Error("expected rate limiter to block request after burst limit")
	}

	stats := limiter.GetStats("test-key")
	if stats.AllowedCount != 2 {
		t.Errorf("expected allowed count 2, got %d", stats.AllowedCount)
	}
	if stats.BlockedCount != 1 {
		t.Errorf("expected blocked count 1, got %d", stats.BlockedCount)
	}
}

func TestSimpleRateLimiterReset(t *testing.T) {
	limiter := NewSimpleRateLimiter(1, 1)

	// Use up the tokens
	limiter.Allow("test-key")
	allowed := limiter.Allow("test-key")
	if allowed {
		t.Error("expected request to be blocked")
	}

	// Reset the limiter
	err := limiter.Reset("test-key")
	if err != nil {
		t.Errorf("expected no error resetting limiter, got: %v", err)
	}

	// Should allow request after reset
	allowed = limiter.Allow("test-key")
	if !allowed {
		t.Error("expected request to be allowed after reset")
	}
}

func TestNewRateLimitingMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		mwName      string
		config      RateLimitingMiddlewareConfig
		limiter     RateLimiter
		expectError bool
		errorMsg    string
	}{
		{
			name:   "valid middleware with custom limiter",
			mwName: "test-ratelimit",
			config: RateLimitingMiddlewareConfig{
				RequestsPerSecond: 10,
				BurstSize:         20,
				PerKey:            true,
				PerLanguage:       false,
				ErrorMessage:      "custom rate limit exceeded",
			},
			limiter:     NewSimpleRateLimiter(10, 20),
			expectError: false,
		},
		{
			name:   "valid middleware with default limiter",
			mwName: "test-ratelimit",
			config: RateLimitingMiddlewareConfig{
				RequestsPerSecond: 5,
				BurstSize:         10,
			},
			limiter:     nil, // should create default limiter
			expectError: false,
		},
		{
			name:        "empty middleware name",
			mwName:      "",
			config:      RateLimitingMiddlewareConfig{},
			limiter:     nil,
			expectError: true,
			errorMsg:    "middleware name cannot be empty",
		},
		{
			name:   "config with defaults",
			mwName: "test-ratelimit",
			config: RateLimitingMiddlewareConfig{
				// Empty config should get defaults
			},
			limiter:     nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, err := NewRateLimitingMiddleware(tt.mwName, tt.config, tt.limiter)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}

				if mw == nil {
					t.Error("expected middleware to be created, but got nil")
					return
				}

				if mw.Name() != tt.mwName {
					t.Errorf("expected middleware name '%s', got '%s'", tt.mwName, mw.Name())
				}
			}
		})
	}
}

func TestRateLimitingMiddlewareWrapTranslate(t *testing.T) {
	limiter := NewSimpleRateLimiter(1, 1) // Very restrictive for testing
	config := RateLimitingMiddlewareConfig{
		RequestsPerSecond: 1,
		BurstSize:         1,
		PerKey:            false,
		PerLanguage:       false,
		ErrorMessage:      "rate limit exceeded",
	}

	mw, err := NewRateLimitingMiddleware("test-ratelimit", config, limiter)
	if err != nil {
		t.Fatalf("failed to create rate limiting middleware: %v", err)
	}

	mockTranslate := func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		return "translated_result", nil
	}

	wrappedTranslate := mw.WrapTranslate(mockTranslate)
	ctx := context.Background()

	// First request should be allowed
	result1, err := wrappedTranslate(ctx, "test.key", "en", nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result1 != "translated_result" {
		t.Errorf("expected 'translated_result', got '%s'", result1)
	}

	// Second request should be rate limited
	_, err = wrappedTranslate(ctx, "test.key", "en", nil)
	if err == nil {
		t.Error("expected rate limiting error, but got none")
	}
	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("expected rate limiting error message, got: %v", err)
	}
}

func TestNewLoggingMiddleware(t *testing.T) {
	logger := &TestLogger{}

	tests := []struct {
		name        string
		mwName      string
		config      LoggingMiddlewareConfig
		logger      Logger
		expectError bool
		errorMsg    string
	}{
		{
			name:   "valid middleware",
			mwName: "test-logging",
			config: LoggingMiddlewareConfig{
				LogRequests:       true,
				LogResults:        true,
				LogLevel:          "info",
				IncludeParameters: true,
				MaxResultLength:   500,
			},
			logger:      logger,
			expectError: false,
		},
		{
			name:        "empty middleware name",
			mwName:      "",
			config:      LoggingMiddlewareConfig{},
			logger:      logger,
			expectError: true,
			errorMsg:    "middleware name cannot be empty",
		},
		{
			name:        "nil logger",
			mwName:      "test-logging",
			config:      LoggingMiddlewareConfig{},
			logger:      nil,
			expectError: true,
			errorMsg:    "logger cannot be nil",
		},
		{
			name:   "config with defaults",
			mwName: "test-logging",
			config: LoggingMiddlewareConfig{
				// Empty config should get defaults
			},
			logger:      logger,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, err := NewLoggingMiddleware(tt.mwName, tt.config, tt.logger)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error, but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}

				if mw == nil {
					t.Error("expected middleware to be created, but got nil")
					return
				}

				if mw.Name() != tt.mwName {
					t.Errorf("expected middleware name '%s', got '%s'", tt.mwName, mw.Name())
				}
			}
		})
	}
}

func TestLoggingMiddlewareWrapTranslate(t *testing.T) {
	logger := &TestLogger{}
	config := LoggingMiddlewareConfig{
		LogRequests:       true,
		LogResults:        true,
		LogLevel:          "info",
		IncludeParameters: true,
		MaxResultLength:   100,
	}

	mw, err := NewLoggingMiddleware("test-logging", config, logger)
	if err != nil {
		t.Fatalf("failed to create logging middleware: %v", err)
	}

	mockTranslate := func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		return "translated_result", nil
	}

	wrappedTranslate := mw.WrapTranslate(mockTranslate)
	ctx := context.Background()
	params := map[string]interface{}{"name": "John"}

	// Reset logger messages
	logger.Reset()

	result, err := wrappedTranslate(ctx, "test.key", "en", params)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result != "translated_result" {
		t.Errorf("expected 'translated_result', got '%s'", result)
	}

	// Check that request and result were logged
	if len(logger.Messages) < 2 {
		t.Errorf("expected at least 2 log messages, got %d", len(logger.Messages))
	}

	// Check request log
	requestLog := logger.Messages[0]
	if !strings.Contains(requestLog.Message, "Translation request") {
		t.Errorf("expected request log message, got: %s", requestLog.Message)
	}
	if !strings.Contains(requestLog.Message, "key=test.key") {
		t.Errorf("expected key in request log, got: %s", requestLog.Message)
	}
	if !strings.Contains(requestLog.Message, "lang=en") {
		t.Errorf("expected language in request log, got: %s", requestLog.Message)
	}
	if !strings.Contains(requestLog.Message, "params=") {
		t.Errorf("expected parameters in request log, got: %s", requestLog.Message)
	}

	// Check result log
	resultLog := logger.Messages[1]
	if !strings.Contains(resultLog.Message, "Translation result") {
		t.Errorf("expected result log message, got: %s", resultLog.Message)
	}
	if !strings.Contains(resultLog.Message, "result=translated_result") {
		t.Errorf("expected result in result log, got: %s", resultLog.Message)
	}
	if !strings.Contains(resultLog.Message, "duration=") {
		t.Errorf("expected duration in result log, got: %s", resultLog.Message)
	}
}

func TestLoggingMiddlewareError(t *testing.T) {
	logger := &TestLogger{}
	config := LoggingMiddlewareConfig{
		LogRequests: true,
		LogResults:  true,
	}

	mw, err := NewLoggingMiddleware("test-logging", config, logger)
	if err != nil {
		t.Fatalf("failed to create logging middleware: %v", err)
	}

	mockError := fmt.Errorf("translation failed")
	mockTranslate := func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		return "", mockError
	}

	wrappedTranslate := mw.WrapTranslate(mockTranslate)
	ctx := context.Background()

	// Reset logger messages
	logger.Reset()

	_, err = wrappedTranslate(ctx, "test.key", "en", nil)
	if err == nil {
		t.Error("expected error, but got none")
	}
	if err != mockError {
		t.Errorf("expected mock error, got: %v", err)
	}

	// Check that error was logged
	found := false
	for _, msg := range logger.Messages {
		if msg.Level == "error" && strings.Contains(msg.Message, "Translation error") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected error to be logged")
	}
}

func TestLoggingMiddlewareResultTruncation(t *testing.T) {
	logger := &TestLogger{}
	config := LoggingMiddlewareConfig{
		LogResults:      true,
		LogLevel:        "info",
		MaxResultLength: 10, // Very short for testing
	}

	mw, err := NewLoggingMiddleware("test-logging", config, logger)
	if err != nil {
		t.Fatalf("failed to create logging middleware: %v", err)
	}

	longResult := "this is a very long translation result that should be truncated"
	mockTranslate := func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		return longResult, nil
	}

	wrappedTranslate := mw.WrapTranslate(mockTranslate)
	ctx := context.Background()

	// Reset logger messages
	logger.Reset()

	result, err := wrappedTranslate(ctx, "test.key", "en", nil)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result != longResult {
		t.Errorf("expected full result, got '%s'", result)
	}

	// Check that result was truncated in log
	found := false
	for _, msg := range logger.Messages {
		if strings.Contains(msg.Message, "Translation result") {
			if strings.Contains(msg.Message, "...") {
				found = true
			}
			break
		}
	}

	if !found {
		t.Error("expected result to be truncated in log")
	}
}

// TestLogger is a test implementation of the Logger interface
type TestLogger struct {
	Messages []LogMessage
}

type LogMessage struct {
	Level   string
	Message string
	Args    []interface{}
}

func (l *TestLogger) Debug(msg string, args ...interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "debug", Message: msg, Args: args})
}

func (l *TestLogger) Info(msg string, args ...interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "info", Message: msg, Args: args})
}

func (l *TestLogger) Warn(msg string, args ...interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "warn", Message: msg, Args: args})
}

func (l *TestLogger) Error(msg string, args ...interface{}) {
	l.Messages = append(l.Messages, LogMessage{Level: "error", Message: msg, Args: args})
}

func (l *TestLogger) Reset() {
	l.Messages = nil
}
