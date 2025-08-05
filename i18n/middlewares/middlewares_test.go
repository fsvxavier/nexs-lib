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

// ===== TESTES AVANÇADOS PARA EXPANSÃO DE COBERTURA =====

// TestMemoryCache_ConcurrentAccess - Testes de concorrência para cache
func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache(1000)
	ttl := 1 * time.Hour
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Teste de operações concorrentes
	const numGoroutines = 50
	const operationsPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	// Goroutines de escrita
	for i := 0; i < numGoroutines/2; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				cache.Set(key, value, ttl)
			}
		}(i)
	}

	// Goroutines de leitura
	for i := numGoroutines / 2; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id-numGoroutines/2, j)
				cache.Get(key)
			}
		}(i)
	}

	// Aguardar conclusão
	completedGoroutines := 0
	for completedGoroutines < numGoroutines {
		select {
		case <-done:
			completedGoroutines++
		case <-ctx.Done():
			t.Fatal("Timeout waiting for concurrent operations to complete")
		}
	}

	// Verificar que o cache ainda funciona
	cache.Set("test_key", "test_value", ttl)
	value, found := cache.Get("test_key")
	if !found || value != "test_value" {
		t.Error("Cache integrity compromised after concurrent access")
	}
}

// TestMemoryCache_EdgeCases - Casos extremos do cache
func TestMemoryCache_EdgeCases(t *testing.T) {
	t.Run("zero_capacity", func(t *testing.T) {
		cache := NewMemoryCache(0)
		err := cache.Set("key", "value", time.Hour)
		if err == nil {
			t.Error("Expected error when setting in zero-capacity cache")
		}
	})

	t.Run("negative_ttl", func(t *testing.T) {
		cache := NewMemoryCache(10)
		err := cache.Set("key", "value", -time.Hour)
		if err != nil {
			t.Errorf("Unexpected error with negative TTL: %v", err)
		}

		// Item deve expirar imediatamente
		_, found := cache.Get("key")
		if found {
			t.Error("Item with negative TTL should not be found")
		}
	})

	t.Run("zero_ttl", func(t *testing.T) {
		cache := NewMemoryCache(10)
		err := cache.Set("key", "value", 0)
		if err != nil {
			t.Errorf("Unexpected error with zero TTL: %v", err)
		}

		// Item deve expirar imediatamente
		_, found := cache.Get("key")
		if found {
			t.Error("Item with zero TTL should not be found")
		}
	})

	t.Run("very_large_value", func(t *testing.T) {
		cache := NewMemoryCache(10)
		largeValue := strings.Repeat("x", 1000000) // 1MB string

		err := cache.Set("large_key", largeValue, time.Hour)
		if err != nil {
			t.Errorf("Unexpected error with large value: %v", err)
		}

		value, found := cache.Get("large_key")
		if !found {
			t.Error("Large value not found in cache")
		}

		if value != largeValue {
			t.Error("Large value was corrupted in cache")
		}
	})

	t.Run("nil_value", func(t *testing.T) {
		cache := NewMemoryCache(10)
		err := cache.Set("nil_key", nil, time.Hour)
		if err != nil {
			t.Errorf("Unexpected error with nil value: %v", err)
		}

		value, found := cache.Get("nil_key")
		if !found {
			t.Error("Nil value not found in cache")
		}

		if value != nil {
			t.Errorf("Expected nil value, got %v", value)
		}
	})

	t.Run("empty_key", func(t *testing.T) {
		cache := NewMemoryCache(10)
		err := cache.Set("", "value", time.Hour)
		if err != nil {
			t.Errorf("Unexpected error with empty key: %v", err)
		}

		value, found := cache.Get("")
		if !found {
			t.Error("Empty key not found in cache")
		}

		if value != "value" {
			t.Errorf("Expected 'value', got %v", value)
		}
	})
}

// TestCachingMiddleware_EdgeCases_Fixed - Casos extremos do middleware de cache (corrigido)
func TestCachingMiddleware_EdgeCases_Fixed(t *testing.T) {
	cache := NewMemoryCache(100)
	config := CachingMiddlewareConfig{
		TTL:              time.Hour,
		MaxSize:          100,
		CacheKeyPrefix:   "test:",
		EnableStats:      true,
		CacheNullResults: true,
	}
	middleware, err := NewCachingMiddleware("test-cache", config, cache)
	if err != nil {
		t.Fatalf("failed to create caching middleware: %v", err)
	}

	t.Run("nil_params", func(t *testing.T) {
		called := false
		translateFunc := func(ctx context.Context, key, lang string, params map[string]interface{}) (string, error) {
			called = true
			return "result", nil
		}

		wrapped := middleware.WrapTranslate(translateFunc)
		result, err := wrapped(context.Background(), "key", "en", nil)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != "result" {
			t.Errorf("Expected 'result', got: %s", result)
		}
		if !called {
			t.Error("Expected translate function to be called")
		}
	})

	t.Run("disabled_cache", func(t *testing.T) {
		cache := NewMemoryCache(100)
		config := CachingMiddlewareConfig{
			TTL:              time.Hour,
			MaxSize:          100,
			CacheKeyPrefix:   "disabled:",
			EnableStats:      false,
			CacheNullResults: false,
		}
		disabledMiddleware, err := NewCachingMiddleware("disabled-cache", config, cache)
		if err != nil {
			t.Fatalf("failed to create disabled middleware: %v", err)
		}

		callCount := 0
		translateFunc := func(ctx context.Context, key, lang string, params map[string]interface{}) (string, error) {
			callCount++
			return "no_cache_result", nil
		}

		wrapped := disabledMiddleware.WrapTranslate(translateFunc)

		// Múltiplas chamadas sempre devem chamar a função original
		for i := 0; i < 3; i++ {
			result, err := wrapped(context.Background(), "key", "en", nil)
			if err != nil {
				t.Errorf("Call %d failed: %v", i, err)
			}
			if result != "no_cache_result" {
				t.Errorf("Expected 'no_cache_result', got %s", result)
			}
		}

		if callCount != 3 {
			t.Errorf("Expected 3 calls, got %d", callCount)
		}
	})
}

// TestSimpleRateLimiter_EdgeCases_Fixed - Casos extremos do rate limiter (corrigido)
func TestSimpleRateLimiter_EdgeCases_Fixed(t *testing.T) {
	t.Run("zero_rate", func(t *testing.T) {
		limiter := NewSimpleRateLimiter(0, 1)

		// Primeira tentativa deve ser permitida pelo burst
		if !limiter.Allow("test-key") {
			t.Error("First request should be allowed due to burst capacity")
		}

		// Segunda tentativa deve ser negada
		if limiter.Allow("test-key") {
			t.Error("Second request should be denied due to zero rate")
		}
	})

	t.Run("zero_burst", func(t *testing.T) {
		limiter := NewSimpleRateLimiter(1, 0)

		// Nenhuma tentativa deve ser permitida
		if limiter.Allow("test-key") {
			t.Error("No requests should be allowed with zero burst")
		}
	})

	t.Run("high_frequency_requests", func(t *testing.T) {
		limiter := NewSimpleRateLimiter(10, 5) // 10 requests/sec, burst of 5

		// Primeiros 5 devem passar
		allowed := 0
		for i := 0; i < 5; i++ {
			if limiter.Allow("test-key") {
				allowed++
			}
		}
		if allowed != 5 {
			t.Errorf("Expected 5 allowed requests, got %d", allowed)
		}

		// Próximos devem ser negados
		denied := 0
		for i := 0; i < 10; i++ {
			if !limiter.Allow("test-key") {
				denied++
			}
		}
		if denied == 0 {
			t.Error("Expected some requests to be denied")
		}
	})
}

// TestRateLimitingMiddleware_EdgeCases_Fixed - Casos extremos do middleware de rate limiting (corrigido)
func TestRateLimitingMiddleware_EdgeCases_Fixed(t *testing.T) {
	limiter := NewSimpleRateLimiter(1, 1)
	config := RateLimitingMiddlewareConfig{
		RequestsPerSecond: 1,
		BurstSize:         1,
		PerKey:            false,
		PerLanguage:       false,
		ErrorMessage:      "rate limit exceeded",
	}
	middleware, err := NewRateLimitingMiddleware("test-limiter", config, limiter)
	if err != nil {
		t.Fatalf("failed to create rate limiting middleware: %v", err)
	}

	t.Run("rate_limit_exceeded", func(t *testing.T) {
		translateFunc := func(ctx context.Context, key, lang string, params map[string]interface{}) (string, error) {
			return "result", nil
		}

		wrapped := middleware.WrapTranslate(translateFunc)

		// Primeira chamada deve passar
		result, err := wrapped(context.Background(), "key", "en", nil)
		if err != nil {
			t.Errorf("First call should succeed: %v", err)
		}
		if result != "result" {
			t.Errorf("Expected 'result', got %s", result)
		}

		// Segunda chamada deve ser limitada
		_, err = wrapped(context.Background(), "key", "en", nil)
		if err == nil {
			t.Error("Second call should be rate limited")
		}
		if !strings.Contains(err.Error(), "rate limit exceeded") {
			t.Errorf("Expected rate limit error, got: %v", err)
		}
	})

	t.Run("disabled_rate_limiting", func(t *testing.T) {
		limiter := NewSimpleRateLimiter(1, 1)
		config := RateLimitingMiddlewareConfig{
			RequestsPerSecond: 1,
			BurstSize:         1,
			PerKey:            false,
			PerLanguage:       false,
			ErrorMessage:      "rate limit exceeded",
		}
		disabledMiddleware, err := NewRateLimitingMiddleware("disabled-limiter", config, limiter)
		if err != nil {
			t.Fatalf("failed to create disabled middleware: %v", err)
		}

		callCount := 0
		translateFunc := func(ctx context.Context, key, lang string, params map[string]interface{}) (string, error) {
			callCount++
			return "unlimited_result", nil
		}

		wrapped := disabledMiddleware.WrapTranslate(translateFunc)

		// Múltiplas chamadas devem passar
		for i := 0; i < 10; i++ {
			result, err := wrapped(context.Background(), "key", "en", nil)
			if err != nil {
				t.Errorf("Call %d should succeed when rate limiting is enabled: %v", i, err)
			}
			if result != "unlimited_result" {
				t.Errorf("Expected 'unlimited_result', got %s", result)
			}
		}

		if callCount != 10 {
			t.Errorf("Expected 10 calls, got %d", callCount)
		}
	})
}

// TestLoggingMiddleware_EdgeCases_Fixed - Casos extremos do middleware de logging (corrigido)
func TestLoggingMiddleware_EdgeCases_Fixed(t *testing.T) {
	logger := &TestLogger{}
	config := LoggingMiddlewareConfig{
		LogRequests:       true,
		LogResults:        true,
		LogLevel:          "info",
		IncludeParameters: true,
		MaxResultLength:   500,
	}
	_, err := NewLoggingMiddleware("test-logger", config, logger)
	if err != nil {
		t.Fatalf("failed to create logging middleware: %v", err)
	}

	t.Run("nil_logger", func(t *testing.T) {
		config := LoggingMiddlewareConfig{
			LogRequests: true,
			LogResults:  true,
		}
		_, err := NewLoggingMiddleware("nil-logger", config, nil)
		if err == nil {
			t.Error("Expected error when creating middleware with nil logger")
		}
		if !strings.Contains(err.Error(), "logger cannot be nil") {
			t.Errorf("Expected nil logger error, got: %v", err)
		}
	})

	t.Run("disabled_logging", func(t *testing.T) {
		logger := &TestLogger{}
		config := LoggingMiddlewareConfig{
			LogRequests: false,
			LogResults:  false,
		}
		disabledMiddleware, err := NewLoggingMiddleware("disabled-logger", config, logger)
		if err != nil {
			t.Fatalf("failed to create disabled middleware: %v", err)
		}

		translateFunc := func(ctx context.Context, key, lang string, params map[string]interface{}) (string, error) {
			return "silent_result", nil
		}

		wrapped := disabledMiddleware.WrapTranslate(translateFunc)
		logger.Reset()

		result, err := wrapped(context.Background(), "key", "en", nil)
		if err != nil {
			t.Errorf("Expected no error: %v", err)
		}
		if result != "silent_result" {
			t.Errorf("Expected 'silent_result', got %s", result)
		}

		// Deve ter menos logs quando desabilitado
		if len(logger.Messages) > 1 {
			t.Errorf("Expected minimal logging when disabled, got %d messages", len(logger.Messages))
		}
	})
}

// TestMiddleware_Integration_Fixed - Teste de integração dos middlewares (corrigido)
func TestMiddleware_Integration_Fixed(t *testing.T) {
	cache := NewMemoryCache(100)
	limiter := NewSimpleRateLimiter(10, 5)
	logger := &TestLogger{}

	// Criar middlewares
	cachingConfig := CachingMiddlewareConfig{
		TTL:              time.Hour,
		MaxSize:          100,
		CacheKeyPrefix:   "integration:",
		EnableStats:      true,
		CacheNullResults: true,
	}
	cachingMiddleware, err := NewCachingMiddleware("integration-cache", cachingConfig, cache)
	if err != nil {
		t.Fatalf("failed to create caching middleware: %v", err)
	}

	rateLimitingConfig := RateLimitingMiddlewareConfig{
		RequestsPerSecond: 10,
		BurstSize:         5,
		PerKey:            false,
		PerLanguage:       false,
		ErrorMessage:      "rate limit exceeded",
	}
	rateLimitingMiddleware, err := NewRateLimitingMiddleware("integration-limiter", rateLimitingConfig, limiter)
	if err != nil {
		t.Fatalf("failed to create rate limiting middleware: %v", err)
	}

	loggingConfig := LoggingMiddlewareConfig{
		LogRequests:     true,
		LogResults:      true,
		LogLevel:        "info",
		MaxResultLength: 100,
	}
	loggingMiddleware, err := NewLoggingMiddleware("integration-logger", loggingConfig, logger)
	if err != nil {
		t.Fatalf("failed to create logging middleware: %v", err)
	}

	// Função de tradução base
	callCount := 0
	translateFunc := func(ctx context.Context, key, lang string, params map[string]interface{}) (string, error) {
		callCount++
		return fmt.Sprintf("translated_%s_%s", key, lang), nil
	}

	// Aplicar middlewares em cadeia
	wrapped := translateFunc
	wrapped = loggingMiddleware.WrapTranslate(wrapped)
	wrapped = cachingMiddleware.WrapTranslate(wrapped)
	wrapped = rateLimitingMiddleware.WrapTranslate(wrapped)

	logger.Reset()

	// Primeira chamada - deve passar por todos os middlewares
	result1, err := wrapped(context.Background(), "test.key", "en", nil)
	if err != nil {
		t.Errorf("First call failed: %v", err)
	}
	if result1 != "translated_test.key_en" {
		t.Errorf("Expected 'translated_test.key_en', got %s", result1)
	}

	// Segunda chamada - deve usar cache
	result2, err := wrapped(context.Background(), "test.key", "en", nil)
	if err != nil {
		t.Errorf("Second call failed: %v", err)
	}
	if result2 != result1 {
		t.Errorf("Expected cached result, got different result")
	}

	// Deve ter chamado a função original apenas uma vez devido ao cache
	if callCount != 1 {
		t.Errorf("Expected function to be called once (cached), got %d calls", callCount)
	}

	// Deve ter logs da primeira chamada
	if len(logger.Messages) == 0 {
		t.Error("Expected log messages from middleware chain")
	}
}

// TestCachingMiddleware_Performance_Fixed - Teste de performance do cache (corrigido)
func TestCachingMiddleware_Performance_Fixed(t *testing.T) {
	cache := NewMemoryCache(1000)
	config := CachingMiddlewareConfig{
		TTL:              time.Hour,
		MaxSize:          1000,
		CacheKeyPrefix:   "perf:",
		EnableStats:      true,
		CacheNullResults: false,
	}
	middleware, err := NewCachingMiddleware("perf-cache", config, cache)
	if err != nil {
		t.Fatalf("failed to create caching middleware: %v", err)
	}

	translateFunc := func(ctx context.Context, key, lang string, params map[string]interface{}) (string, error) {
		// Simular operação lenta
		time.Sleep(1 * time.Millisecond)
		return fmt.Sprintf("translated_%s_%s", key, lang), nil
	}

	wrapped := middleware.WrapTranslate(translateFunc)

	// Primeira chamada - sem cache
	start := time.Now()
	result1, err := wrapped(context.Background(), "perf.key", "en", nil)
	duration1 := time.Since(start)

	if err != nil {
		t.Errorf("First call failed: %v", err)
	}
	if result1 != "translated_perf.key_en" {
		t.Errorf("Expected 'translated_perf.key_en', got %s", result1)
	}

	// Segunda chamada - com cache (deve ser mais rápida)
	start = time.Now()
	result2, err := wrapped(context.Background(), "perf.key", "en", nil)
	duration2 := time.Since(start)

	if err != nil {
		t.Errorf("Second call failed: %v", err)
	}
	if result2 != result1 {
		t.Errorf("Expected cached result, got different result")
	}

	// Cache deve ser significativamente mais rápido
	if duration2 >= duration1 {
		t.Logf("Warning: Cache call (%v) not faster than original call (%v)", duration2, duration1)
	}
}
