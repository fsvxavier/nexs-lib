// Package ratelimit provides rate limiting middleware implementations.
package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Algorithm represents the rate limiting algorithm.
type Algorithm string

const (
	// TokenBucket uses token bucket algorithm.
	TokenBucket Algorithm = "token_bucket"
	// SlidingWindow uses sliding window algorithm.
	SlidingWindow Algorithm = "sliding_window"
	// FixedWindow uses fixed window algorithm.
	FixedWindow Algorithm = "fixed_window"
)

// Config represents rate limiting configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths that should bypass rate limiting.
	SkipPaths []string
	// Limit is the number of requests allowed per time window.
	Limit int
	// Window is the time window duration.
	Window time.Duration
	// Algorithm is the rate limiting algorithm to use.
	Algorithm Algorithm
	// KeyGenerator generates the rate limiting key from request.
	KeyGenerator func(*http.Request) string
	// OnLimitExceeded is called when rate limit is exceeded.
	OnLimitExceeded func(http.ResponseWriter, *http.Request)
}

// IsEnabled returns true if the middleware is enabled.
func (c Config) IsEnabled() bool {
	return c.Enabled
}

// ShouldSkip returns true if the given path should be skipped.
func (c Config) ShouldSkip(path string) bool {
	for _, skipPath := range c.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// DefaultConfig returns a default rate limiting configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:         true,
		Limit:           100,
		Window:          time.Minute,
		Algorithm:       TokenBucket,
		KeyGenerator:    defaultKeyGenerator,
		OnLimitExceeded: defaultOnLimitExceeded,
	}
}

// Middleware implements rate limiting middleware.
type Middleware struct {
	config  Config
	limiter interfaces.RateLimiter
}

// NewMiddleware creates a new rate limiting middleware.
func NewMiddleware(config Config) *Middleware {
	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultKeyGenerator
	}
	if config.OnLimitExceeded == nil {
		config.OnLimitExceeded = defaultOnLimitExceeded
	}

	var limiter interfaces.RateLimiter
	switch config.Algorithm {
	case TokenBucket:
		limiter = NewTokenBucketLimiter(config.Limit, config.Window)
	case SlidingWindow:
		limiter = NewSlidingWindowLimiter(config.Limit, config.Window)
	case FixedWindow:
		limiter = NewFixedWindowLimiter(config.Limit, config.Window)
	default:
		limiter = NewTokenBucketLimiter(config.Limit, config.Window)
	}

	return &Middleware{
		config:  config,
		limiter: limiter,
	}
}

// Wrap implements the interfaces.Middleware interface.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		if m.config.ShouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		key := m.config.KeyGenerator(r)
		if !m.limiter.Allow(key) {
			// Add rate limit headers
			w.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(m.limiter.GetLimit()))
			w.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(m.limiter.GetRemaining(key)))
			w.Header().Set("X-Rate-Limit-Reset", strconv.FormatInt(time.Now().Add(m.config.Window).Unix(), 10))

			m.config.OnLimitExceeded(w, r)
			return
		}

		// Add rate limit headers for successful requests
		w.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(m.limiter.GetLimit()))
		w.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(m.limiter.GetRemaining(key)))

		next.ServeHTTP(w, r)
	})
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "ratelimit"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 200 // Rate limiting should happen early
}

// defaultKeyGenerator generates rate limiting key from client IP.
func defaultKeyGenerator(r *http.Request) string {
	// Try to get real IP from headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// defaultOnLimitExceeded sends 429 Too Many Requests response.
func defaultOnLimitExceeded(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
}

// TokenBucketLimiter implements token bucket rate limiting.
type TokenBucketLimiter struct {
	mu       sync.RWMutex
	limit    int
	refill   time.Duration
	buckets  map[string]*bucket
	stopChan chan struct{}
}

type bucket struct {
	tokens   int
	lastSeen time.Time
}

// NewTokenBucketLimiter creates a new token bucket rate limiter.
func NewTokenBucketLimiter(limit int, window time.Duration) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		limit:    limit,
		refill:   window / time.Duration(limit),
		buckets:  make(map[string]*bucket),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request is allowed.
func (l *TokenBucketLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.buckets[key]
	if !exists {
		b = &bucket{
			tokens:   l.limit - 1,
			lastSeen: now,
		}
		l.buckets[key] = b
		return true
	}

	// Add tokens based on elapsed time
	elapsed := now.Sub(b.lastSeen)
	tokensToAdd := int(elapsed / l.refill)
	b.tokens += tokensToAdd
	if b.tokens > l.limit {
		b.tokens = l.limit
	}
	b.lastSeen = now

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

// Reset resets the rate limit for a key.
func (l *TokenBucketLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// GetLimit returns the current limit.
func (l *TokenBucketLimiter) GetLimit() int {
	return l.limit
}

// GetRemaining returns remaining requests for a key.
func (l *TokenBucketLimiter) GetRemaining(key string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	b, exists := l.buckets[key]
	if !exists {
		return l.limit
	}

	// Calculate current tokens
	now := time.Now()
	elapsed := now.Sub(b.lastSeen)
	tokensToAdd := int(elapsed / l.refill)
	tokens := b.tokens + tokensToAdd
	if tokens > l.limit {
		tokens = l.limit
	}

	return tokens
}

// cleanup removes expired buckets.
func (l *TokenBucketLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for key, b := range l.buckets {
				if now.Sub(b.lastSeen) > time.Hour {
					delete(l.buckets, key)
				}
			}
			l.mu.Unlock()
		case <-l.stopChan:
			return
		}
	}
}

// Close stops the cleanup goroutine.
func (l *TokenBucketLimiter) Close() {
	close(l.stopChan)
}

// SlidingWindowLimiter implements sliding window rate limiting.
type SlidingWindowLimiter struct {
	mu      sync.RWMutex
	limit   int
	window  time.Duration
	windows map[string]*slidingWindow
}

type slidingWindow struct {
	requests []time.Time
	lastSeen time.Time
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter.
func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:   limit,
		window:  window,
		windows: make(map[string]*slidingWindow),
	}
}

// Allow checks if a request is allowed.
func (l *SlidingWindowLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, exists := l.windows[key]
	if !exists {
		w = &slidingWindow{
			requests: []time.Time{now},
			lastSeen: now,
		}
		l.windows[key] = w
		return true
	}

	// Remove expired requests
	cutoff := now.Add(-l.window)
	validRequests := w.requests[:0]
	for _, req := range w.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	w.requests = validRequests
	w.lastSeen = now

	if len(w.requests) >= l.limit {
		return false
	}

	w.requests = append(w.requests, now)
	return true
}

// Reset resets the rate limit for a key.
func (l *SlidingWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.windows, key)
}

// GetLimit returns the current limit.
func (l *SlidingWindowLimiter) GetLimit() int {
	return l.limit
}

// GetRemaining returns remaining requests for a key.
func (l *SlidingWindowLimiter) GetRemaining(key string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	w, exists := l.windows[key]
	if !exists {
		return l.limit
	}

	// Count valid requests
	now := time.Now()
	cutoff := now.Add(-l.window)
	validCount := 0
	for _, req := range w.requests {
		if req.After(cutoff) {
			validCount++
		}
	}

	remaining := l.limit - validCount
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// FixedWindowLimiter implements fixed window rate limiting.
type FixedWindowLimiter struct {
	mu      sync.RWMutex
	limit   int
	window  time.Duration
	windows map[string]*fixedWindow
}

type fixedWindow struct {
	count     int
	windowEnd time.Time
}

// NewFixedWindowLimiter creates a new fixed window rate limiter.
func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:   limit,
		window:  window,
		windows: make(map[string]*fixedWindow),
	}
}

// Allow checks if a request is allowed.
func (l *FixedWindowLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, exists := l.windows[key]
	if !exists || now.After(w.windowEnd) {
		w = &fixedWindow{
			count:     1,
			windowEnd: now.Add(l.window),
		}
		l.windows[key] = w
		return true
	}

	if w.count >= l.limit {
		return false
	}

	w.count++
	return true
}

// Reset resets the rate limit for a key.
func (l *FixedWindowLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.windows, key)
}

// GetLimit returns the current limit.
func (l *FixedWindowLimiter) GetLimit() int {
	return l.limit
}

// GetRemaining returns remaining requests for a key.
func (l *FixedWindowLimiter) GetRemaining(key string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	w, exists := l.windows[key]
	if !exists {
		return l.limit
	}

	now := time.Now()
	if now.After(w.windowEnd) {
		return l.limit
	}

	remaining := l.limit - w.count
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}
