// Package middlewares provides implementation of middlewares that wrap and enhance i18n operations.
// Middlewares implement the Middleware interface and can be registered to add functionality
// like caching, rate limiting, validation, and monitoring to translation operations.
package middlewares

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TranslateFunc is a function type for translation operations.
// Used by middlewares to wrap translation calls.
type TranslateFunc func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error)

// CachingMiddleware provides caching functionality for translation operations.
// It implements the Middleware interface and caches translation results to improve performance.
type CachingMiddleware struct {
	name   string
	cache  Cache
	config CachingMiddlewareConfig
	mu     sync.RWMutex
}

// CachingMiddlewareConfig contains configuration options for the caching middleware.
type CachingMiddlewareConfig struct {
	// TTL is the time-to-live for cached translations
	TTL time.Duration `json:"ttl" yaml:"ttl"`

	// MaxSize is the maximum number of cached translations
	MaxSize int `json:"max_size" yaml:"max_size"`

	// CacheKeyPrefix is the prefix for cache keys
	CacheKeyPrefix string `json:"cache_key_prefix" yaml:"cache_key_prefix"`

	// EnableStats determines if cache statistics should be collected
	EnableStats bool `json:"enable_stats" yaml:"enable_stats"`

	// CacheNullResults determines if null/empty results should be cached
	CacheNullResults bool `json:"cache_null_results" yaml:"cache_null_results"`
}

// Cache defines the interface for cache operations.
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Size() int
	Stats() CacheStats
}

// CacheStats contains cache statistics.
type CacheStats struct {
	HitCount  int64 `json:"hit_count"`
	MissCount int64 `json:"miss_count"`
	Size      int   `json:"size"`
	MaxSize   int   `json:"max_size"`
}

// MemoryCache is a simple in-memory cache implementation.
type MemoryCache struct {
	data    map[string]cacheItem
	stats   CacheStats
	maxSize int
	mu      sync.RWMutex
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache with the specified maximum size.
func NewMemoryCache(maxSize int) *MemoryCache {
	return &MemoryCache{
		data:    make(map[string]cacheItem),
		maxSize: maxSize,
		stats:   CacheStats{MaxSize: maxSize},
	}
}

// Get retrieves a value from the cache.
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[key]
	if !exists {
		c.stats.MissCount++
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		// Item has expired, remove it
		delete(c.data, key)
		c.stats.MissCount++
		c.stats.Size = len(c.data)
		return nil, false
	}

	c.stats.HitCount++
	return item.value, true
}

// Set stores a value in the cache with the specified TTL.
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if cache has zero capacity
	if c.maxSize <= 0 {
		return fmt.Errorf("cannot set value in cache with zero or negative capacity")
	}

	// Check if we need to make room
	if len(c.data) >= c.maxSize {
		// Simple eviction: remove the first item (not optimal, but sufficient for this example)
		for k := range c.data {
			delete(c.data, k)
			break
		}
	}

	c.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	c.stats.Size = len(c.data)
	return nil
}

// Delete removes a value from the cache.
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	c.stats.Size = len(c.data)
	return nil
}

// Clear removes all values from the cache.
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]cacheItem)
	c.stats.Size = 0
	c.stats.HitCount = 0
	c.stats.MissCount = 0
	return nil
}

// Size returns the current number of items in the cache.
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Stats returns cache statistics.
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.data)
	return stats
}

// NewCachingMiddleware creates a new caching middleware with the specified configuration.
func NewCachingMiddleware(name string, config CachingMiddlewareConfig, cache Cache) (*CachingMiddleware, error) {
	if name == "" {
		return nil, fmt.Errorf("middleware name cannot be empty")
	}

	if cache == nil {
		// Create default in-memory cache
		maxSize := config.MaxSize
		if maxSize <= 0 {
			maxSize = 1000 // default max size
		}
		cache = NewMemoryCache(maxSize)
	}

	// Set defaults if not provided
	if config.TTL == 0 {
		config.TTL = 1 * time.Hour
	}
	if config.CacheKeyPrefix == "" {
		config.CacheKeyPrefix = "i18n:"
	}

	return &CachingMiddleware{
		name:   name,
		cache:  cache,
		config: config,
	}, nil
}

// Name returns the unique name of the middleware.
func (m *CachingMiddleware) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

// WrapTranslate wraps a translation operation with caching functionality.
func (m *CachingMiddleware) WrapTranslate(next TranslateFunc) TranslateFunc {
	return func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		m.mu.RLock()
		cacheKey := m.buildCacheKey(key, lang, params)
		m.mu.RUnlock()

		// Try to get from cache first
		if cached, found := m.cache.Get(cacheKey); found {
			if result, ok := cached.(string); ok {
				return result, nil
			}
		}

		// Call the next function in the chain
		result, err := next(ctx, key, lang, params)
		if err != nil {
			return "", err
		}

		// Cache the result if it's not empty or if we cache null results
		if result != "" || m.config.CacheNullResults {
			m.cache.Set(cacheKey, result, m.config.TTL)
		}

		return result, nil
	}
}

// OnStart is called when a translation provider starts.
func (m *CachingMiddleware) OnStart(ctx context.Context, providerName string) error {
	// Nothing to do on start
	return nil
}

// OnStop is called when a translation provider stops.
func (m *CachingMiddleware) OnStop(ctx context.Context, providerName string) error {
	// Optionally clear cache on stop
	return nil
}

// OnError is called when a translation provider encounters an error.
func (m *CachingMiddleware) OnError(ctx context.Context, providerName string, err error) error {
	// Nothing to do on error
	return nil
}

// OnTranslate is called when a translation is performed.
func (m *CachingMiddleware) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	// Nothing to do - caching is handled in WrapTranslate
	return nil
}

// buildCacheKey builds a cache key from the translation parameters.
func (m *CachingMiddleware) buildCacheKey(key string, lang string, params map[string]interface{}) string {
	baseKey := fmt.Sprintf("%s%s:%s", m.config.CacheKeyPrefix, lang, key)

	// If there are parameters, include them in the cache key
	if len(params) > 0 {
		// Simple approach: append parameter count to key
		// In a real implementation, you might want to hash the parameters
		baseKey += fmt.Sprintf(":%d", len(params))
	}

	return baseKey
}

// GetStats returns cache statistics if stats are enabled.
func (m *CachingMiddleware) GetStats() CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config.EnableStats {
		return m.cache.Stats()
	}

	return CacheStats{}
}

// ClearCache clears all cached translations.
func (m *CachingMiddleware) ClearCache() error {
	return m.cache.Clear()
}

// RateLimitingMiddleware provides rate limiting functionality for translation operations.
// It implements the Middleware interface and limits the number of translation requests per time period.
type RateLimitingMiddleware struct {
	name    string
	limiter RateLimiter
	config  RateLimitingMiddlewareConfig
	mu      sync.RWMutex
}

// RateLimitingMiddlewareConfig contains configuration options for the rate limiting middleware.
type RateLimitingMiddlewareConfig struct {
	// RequestsPerSecond is the maximum number of requests allowed per second
	RequestsPerSecond int `json:"requests_per_second" yaml:"requests_per_second"`

	// BurstSize is the maximum number of requests allowed in a burst
	BurstSize int `json:"burst_size" yaml:"burst_size"`

	// PerKey determines if rate limiting should be applied per translation key
	PerKey bool `json:"per_key" yaml:"per_key"`

	// PerLanguage determines if rate limiting should be applied per language
	PerLanguage bool `json:"per_language" yaml:"per_language"`

	// ErrorMessage is the message returned when rate limit is exceeded
	ErrorMessage string `json:"error_message" yaml:"error_message"`
}

// RateLimiter defines the interface for rate limiting operations.
type RateLimiter interface {
	Allow(key string) bool
	Reset(key string) error
	GetStats(key string) RateLimitStats
}

// RateLimitStats contains rate limiting statistics.
type RateLimitStats struct {
	RequestCount int64     `json:"request_count"`
	AllowedCount int64     `json:"allowed_count"`
	BlockedCount int64     `json:"blocked_count"`
	LastRequest  time.Time `json:"last_request"`
}

// SimpleRateLimiter is a basic rate limiter implementation using token bucket algorithm.
type SimpleRateLimiter struct {
	requestsPerSecond int
	burstSize         int
	buckets           map[string]*tokenBucket
	mu                sync.RWMutex
}

type tokenBucket struct {
	tokens       int
	lastRefill   time.Time
	requestCount int64
	allowedCount int64
	blockedCount int64
}

// NewSimpleRateLimiter creates a new simple rate limiter.
func NewSimpleRateLimiter(requestsPerSecond, burstSize int) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		requestsPerSecond: requestsPerSecond,
		burstSize:         burstSize,
		buckets:           make(map[string]*tokenBucket),
	}
}

// Allow checks if a request is allowed for the given key.
func (rl *SimpleRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     rl.burstSize,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * rl.requestsPerSecond

	if tokensToAdd > 0 {
		bucket.tokens += tokensToAdd
		if bucket.tokens > rl.burstSize {
			bucket.tokens = rl.burstSize
		}
		bucket.lastRefill = now
	}

	bucket.requestCount++

	if bucket.tokens > 0 {
		bucket.tokens--
		bucket.allowedCount++
		return true
	}

	bucket.blockedCount++
	return false
}

// Reset resets the rate limiter state for the given key.
func (rl *SimpleRateLimiter) Reset(key string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.buckets, key)
	return nil
}

// GetStats returns rate limiting statistics for the given key.
func (rl *SimpleRateLimiter) GetStats(key string) RateLimitStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	bucket, exists := rl.buckets[key]
	if !exists {
		return RateLimitStats{}
	}

	return RateLimitStats{
		RequestCount: bucket.requestCount,
		AllowedCount: bucket.allowedCount,
		BlockedCount: bucket.blockedCount,
		LastRequest:  bucket.lastRefill,
	}
}

// NewRateLimitingMiddleware creates a new rate limiting middleware with the specified configuration.
func NewRateLimitingMiddleware(name string, config RateLimitingMiddlewareConfig, limiter RateLimiter) (*RateLimitingMiddleware, error) {
	if name == "" {
		return nil, fmt.Errorf("middleware name cannot be empty")
	}

	if limiter == nil {
		// Create default rate limiter
		requestsPerSecond := config.RequestsPerSecond
		if requestsPerSecond <= 0 {
			requestsPerSecond = 100 // default: 100 requests per second
		}

		burstSize := config.BurstSize
		if burstSize <= 0 {
			burstSize = requestsPerSecond * 2 // default: 2x requests per second
		}

		limiter = NewSimpleRateLimiter(requestsPerSecond, burstSize)
	}

	// Set defaults if not provided
	if config.ErrorMessage == "" {
		config.ErrorMessage = "rate limit exceeded"
	}

	return &RateLimitingMiddleware{
		name:    name,
		limiter: limiter,
		config:  config,
	}, nil
}

// Name returns the unique name of the middleware.
func (m *RateLimitingMiddleware) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

// WrapTranslate wraps a translation operation with rate limiting functionality.
func (m *RateLimitingMiddleware) WrapTranslate(next TranslateFunc) TranslateFunc {
	return func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		m.mu.RLock()
		rateLimitKey := m.buildRateLimitKey(key, lang)
		m.mu.RUnlock()

		if !m.limiter.Allow(rateLimitKey) {
			return "", fmt.Errorf("%s", m.config.ErrorMessage)
		}

		return next(ctx, key, lang, params)
	}
}

// OnStart is called when a translation provider starts.
func (m *RateLimitingMiddleware) OnStart(ctx context.Context, providerName string) error {
	// Nothing to do on start
	return nil
}

// OnStop is called when a translation provider stops.
func (m *RateLimitingMiddleware) OnStop(ctx context.Context, providerName string) error {
	// Nothing to do on stop
	return nil
}

// OnError is called when a translation provider encounters an error.
func (m *RateLimitingMiddleware) OnError(ctx context.Context, providerName string, err error) error {
	// Nothing to do on error
	return nil
}

// OnTranslate is called when a translation is performed.
func (m *RateLimitingMiddleware) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	// Nothing to do - rate limiting is handled in WrapTranslate
	return nil
}

// buildRateLimitKey builds a rate limit key based on the configuration.
func (m *RateLimitingMiddleware) buildRateLimitKey(key string, lang string) string {
	var parts []string

	if m.config.PerLanguage {
		parts = append(parts, lang)
	}

	if m.config.PerKey {
		parts = append(parts, key)
	}

	if len(parts) == 0 {
		return "global" // global rate limiting
	}

	return fmt.Sprintf("%s", parts[0])
}

// GetStats returns rate limiting statistics.
func (m *RateLimitingMiddleware) GetStats(key string) RateLimitStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rateLimitKey := key
	if key == "" {
		rateLimitKey = "global"
	}

	return m.limiter.GetStats(rateLimitKey)
}

// LoggingMiddleware provides logging functionality for translation operations.
// It implements the Middleware interface and logs translation requests and results.
type LoggingMiddleware struct {
	name   string
	logger Logger
	config LoggingMiddlewareConfig
	mu     sync.RWMutex
}

// LoggingMiddlewareConfig contains configuration options for the logging middleware.
type LoggingMiddlewareConfig struct {
	// LogRequests determines if translation requests should be logged
	LogRequests bool `json:"log_requests" yaml:"log_requests"`

	// LogResults determines if translation results should be logged
	LogResults bool `json:"log_results" yaml:"log_results"`

	// LogLevel is the log level to use ("debug", "info", "warn", "error")
	LogLevel string `json:"log_level" yaml:"log_level"`

	// IncludeParameters determines if translation parameters should be logged
	IncludeParameters bool `json:"include_parameters" yaml:"include_parameters"`

	// MaxResultLength is the maximum length of results to log
	MaxResultLength int `json:"max_result_length" yaml:"max_result_length"`
}

// Logger defines the interface for logging operations.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// NewLoggingMiddleware creates a new logging middleware with the specified configuration.
func NewLoggingMiddleware(name string, config LoggingMiddlewareConfig, logger Logger) (*LoggingMiddleware, error) {
	if name == "" {
		return nil, fmt.Errorf("middleware name cannot be empty")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Set defaults if not provided
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.MaxResultLength == 0 {
		config.MaxResultLength = 1000
	}

	return &LoggingMiddleware{
		name:   name,
		logger: logger,
		config: config,
	}, nil
}

// Name returns the unique name of the middleware.
func (m *LoggingMiddleware) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

// WrapTranslate wraps a translation operation with logging functionality.
func (m *LoggingMiddleware) WrapTranslate(next TranslateFunc) TranslateFunc {
	return func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		start := time.Now()

		if m.config.LogRequests {
			msg := fmt.Sprintf("Translation request: key=%s, lang=%s", key, lang)
			if m.config.IncludeParameters && len(params) > 0 {
				msg += fmt.Sprintf(", params=%v", params)
			}
			m.logAtLevel(m.config.LogLevel, msg)
		}

		result, err := next(ctx, key, lang, params)
		duration := time.Since(start)

		if err != nil {
			m.logger.Error("Translation error: key=%s, lang=%s, error=%v, duration=%v", key, lang, err, duration)
		} else if m.config.LogResults {
			resultToLog := result
			if len(result) > m.config.MaxResultLength {
				resultToLog = result[:m.config.MaxResultLength] + "..."
			}
			msg := fmt.Sprintf("Translation result: key=%s, lang=%s, result=%s, duration=%v", key, lang, resultToLog, duration)
			m.logAtLevel(m.config.LogLevel, msg)
		}

		return result, err
	}
}

// OnStart is called when a translation provider starts.
func (m *LoggingMiddleware) OnStart(ctx context.Context, providerName string) error {
	m.logger.Info("Translation provider started: %s", providerName)
	return nil
}

// OnStop is called when a translation provider stops.
func (m *LoggingMiddleware) OnStop(ctx context.Context, providerName string) error {
	m.logger.Info("Translation provider stopped: %s", providerName)
	return nil
}

// OnError is called when a translation provider encounters an error.
func (m *LoggingMiddleware) OnError(ctx context.Context, providerName string, err error) error {
	m.logger.Error("Translation provider error: %s, error=%v", providerName, err)
	return nil
}

// OnTranslate is called when a translation is performed.
func (m *LoggingMiddleware) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	// Nothing to do - logging is handled in WrapTranslate
	return nil
}

// logAtLevel logs a message at the specified level.
func (m *LoggingMiddleware) logAtLevel(level string, msg string) {
	switch level {
	case "debug":
		m.logger.Debug(msg)
	case "info":
		m.logger.Info(msg)
	case "warn":
		m.logger.Warn(msg)
	case "error":
		m.logger.Error(msg)
	default:
		m.logger.Info(msg)
	}
}
