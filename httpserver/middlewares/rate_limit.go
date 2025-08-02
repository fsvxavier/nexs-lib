package middlewares

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimitMiddleware provides rate limiting functionality.
type RateLimitMiddleware struct {
	*BaseMiddleware

	// Configuration
	config RateLimitConfig

	// Rate limiters
	limiters map[string]*RateLimiter
	mu       sync.RWMutex

	// Metrics
	totalRequests   int64
	allowedRequests int64
	blockedRequests int64
	resetCount      int64

	// Internal state
	startTime     time.Time
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// RateLimitConfig defines configuration options for the rate limiting middleware.
type RateLimitConfig struct {
	// Rate limiting strategy
	Strategy RateLimitStrategy

	// Rate limits
	RequestsPerSecond float64
	RequestsPerMinute int64
	RequestsPerHour   int64
	RequestsPerDay    int64
	BurstSize         int

	// Identification
	KeyFunc          func(ctx context.Context, req interface{}) string
	IdentifyByIP     bool
	IdentifyByUser   bool
	IdentifyByHeader string
	IdentifyByQuery  string

	// Path and method specific limits
	PathLimits   map[string]RateLimit
	MethodLimits map[string]RateLimit

	// Skip configuration
	SkipPaths   []string
	SkipMethods []string
	SkipIPs     []string
	SkipUsers   []string

	// Behavior configuration
	BlockDuration   time.Duration
	CleanupInterval time.Duration
	MemoryLimit     int64

	// Response configuration
	RetryAfterHeader bool
	IncludeHeaders   bool
	ErrorMessage     string
	ErrorStatusCode  int

	// Advanced features
	SlidingWindow   bool
	DistributedMode bool
	WhitelistMode   bool
}

// RateLimitStrategy defines different rate limiting strategies.
type RateLimitStrategy int

const (
	TokenBucket RateLimitStrategy = iota
	FixedWindow
	SlidingWindow
	LeakyBucket
)

// RateLimit defines a specific rate limit configuration.
type RateLimit struct {
	RequestsPerSecond float64
	RequestsPerMinute int64
	RequestsPerHour   int64
	RequestsPerDay    int64
	BurstSize         int
}

// RateLimiter implements different rate limiting algorithms.
type RateLimiter struct {
	key      string
	config   RateLimit
	strategy RateLimitStrategy

	// Token bucket fields
	tokens     float64
	lastRefill time.Time

	// Fixed/sliding window fields
	windowStart    time.Time
	windowRequests int64

	// Request history for sliding window
	requests []time.Time

	// Metrics
	totalRequests   int64
	allowedRequests int64
	blockedRequests int64
	lastReset       time.Time

	mu sync.Mutex
}

// RateLimitContext represents rate limiting context for a request.
type RateLimitContext struct {
	Key              string
	Allowed          bool
	Remaining        int64
	ResetTime        time.Time
	RetryAfter       time.Duration
	WindowStart      time.Time
	RequestsInWindow int64
}

// NewRateLimitMiddleware creates a new rate limiting middleware with default configuration.
func NewRateLimitMiddleware(priority int) *RateLimitMiddleware {
	middleware := &RateLimitMiddleware{
		BaseMiddleware: NewBaseMiddleware("rate_limit", priority),
		config:         DefaultRateLimitConfig(),
		limiters:       make(map[string]*RateLimiter),
		startTime:      time.Now(),
		stopCleanup:    make(chan bool, 1),
	}

	middleware.startCleanupRoutine()
	return middleware
}

// NewRateLimitMiddlewareWithConfig creates a new rate limiting middleware with custom configuration.
func NewRateLimitMiddlewareWithConfig(priority int, config RateLimitConfig) *RateLimitMiddleware {
	middleware := &RateLimitMiddleware{
		BaseMiddleware: NewBaseMiddleware("rate_limit", priority),
		config:         config,
		limiters:       make(map[string]*RateLimiter),
		startTime:      time.Now(),
		stopCleanup:    make(chan bool, 1),
	}

	middleware.startCleanupRoutine()
	return middleware
}

// DefaultRateLimitConfig returns a default rate limiting configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Strategy:          TokenBucket,
		RequestsPerSecond: 10.0,
		RequestsPerMinute: 600,
		RequestsPerHour:   36000,
		RequestsPerDay:    864000,
		BurstSize:         20,
		IdentifyByIP:      true,
		IdentifyByUser:    false,
		PathLimits:        make(map[string]RateLimit),
		MethodLimits:      make(map[string]RateLimit),
		SkipPaths:         []string{"/health", "/metrics"},
		SkipMethods:       []string{"OPTIONS"},
		SkipIPs:           []string{"127.0.0.1", "::1"},
		SkipUsers:         []string{},
		BlockDuration:     time.Minute * 5,
		CleanupInterval:   time.Minute * 10,
		MemoryLimit:       1048576, // 1MB
		RetryAfterHeader:  true,
		IncludeHeaders:    true,
		ErrorMessage:      "Rate limit exceeded",
		ErrorStatusCode:   429, // Too Many Requests
		SlidingWindow:     false,
		DistributedMode:   false,
		WhitelistMode:     false,
	}
}

// Process implements the Middleware interface for rate limiting.
func (rlm *RateLimitMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	if !rlm.IsEnabled() {
		return next(ctx, req)
	}

	atomic.AddInt64(&rlm.totalRequests, 1)

	// Extract request information
	reqInfo := rlm.extractRequestInfo(req)

	// Check if request should skip rate limiting
	if rlm.shouldSkipRateLimit(reqInfo) {
		return next(ctx, req)
	}

	// Generate rate limit key
	key := rlm.generateKey(ctx, req, reqInfo)

	// Get or create rate limiter for this key
	limiter := rlm.getLimiter(key, reqInfo)

	// Check rate limit
	rateLimitCtx := limiter.checkLimit()
	rateLimitCtx.Key = key

	if !rateLimitCtx.Allowed {
		atomic.AddInt64(&rlm.blockedRequests, 1)
		rlm.GetLogger().Warn("Rate limit exceeded for key: %s", key)
		return rlm.createRateLimitErrorResponse(rateLimitCtx)
	}

	atomic.AddInt64(&rlm.allowedRequests, 1)
	rlm.GetLogger().Debug("Rate limit check passed for key: %s (remaining: %d)", key, rateLimitCtx.Remaining)

	// Add rate limit context to request context
	ctxWithRateLimit := context.WithValue(ctx, "rate_limit", rateLimitCtx)

	// Process request
	resp, err := next(ctxWithRateLimit, req)
	if err != nil {
		return resp, err
	}

	// Add rate limit headers to response if configured
	if rlm.config.IncludeHeaders {
		rlm.addRateLimitHeaders(resp, rateLimitCtx)
	}

	return resp, nil
}

// GetConfig returns the current rate limiting configuration.
func (rlm *RateLimitMiddleware) GetConfig() RateLimitConfig {
	return rlm.config
}

// SetConfig updates the rate limiting configuration.
func (rlm *RateLimitMiddleware) SetConfig(config RateLimitConfig) {
	rlm.config = config
	rlm.GetLogger().Info("Rate limit middleware configuration updated")
}

// GetMetrics returns rate limiting metrics.
func (rlm *RateLimitMiddleware) GetMetrics() map[string]interface{} {
	totalRequests := atomic.LoadInt64(&rlm.totalRequests)
	allowedRequests := atomic.LoadInt64(&rlm.allowedRequests)
	blockedRequests := atomic.LoadInt64(&rlm.blockedRequests)

	var allowedRate float64
	if totalRequests > 0 {
		allowedRate = float64(allowedRequests) / float64(totalRequests) * 100.0
	}

	var blockedRate float64
	if totalRequests > 0 {
		blockedRate = float64(blockedRequests) / float64(totalRequests) * 100.0
	}

	rlm.mu.RLock()
	activeLimiters := len(rlm.limiters)
	rlm.mu.RUnlock()

	return map[string]interface{}{
		"total_requests":   totalRequests,
		"allowed_requests": allowedRequests,
		"blocked_requests": blockedRequests,
		"allowed_rate":     allowedRate,
		"blocked_rate":     blockedRate,
		"active_limiters":  activeLimiters,
		"reset_count":      atomic.LoadInt64(&rlm.resetCount),
		"uptime":           time.Since(rlm.startTime),
	}
}

// RateLimitRequestInfo holds request information for rate limiting.
type RateLimitRequestInfo struct {
	Method  string
	Path    string
	IP      string
	UserID  string
	Headers map[string]string
	Query   map[string]string
}

// extractRequestInfo extracts relevant information from the request.
func (rlm *RateLimitMiddleware) extractRequestInfo(req interface{}) *RateLimitRequestInfo {
	info := &RateLimitRequestInfo{
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	if httpReq, ok := req.(map[string]interface{}); ok {
		if method, exists := httpReq["method"]; exists {
			if m, ok := method.(string); ok {
				info.Method = strings.ToUpper(m)
			}
		}

		if path, exists := httpReq["path"]; exists {
			if p, ok := path.(string); ok {
				info.Path = p
			}
		}

		if ip, exists := httpReq["ip"]; exists {
			if i, ok := ip.(string); ok {
				info.IP = i
			}
		}

		if userID, exists := httpReq["user_id"]; exists {
			if u, ok := userID.(string); ok {
				info.UserID = u
			}
		}

		if headers, exists := httpReq["headers"]; exists {
			if h, ok := headers.(map[string]string); ok {
				info.Headers = h
			}
		}

		if query, exists := httpReq["query"]; exists {
			if q, ok := query.(map[string]string); ok {
				info.Query = q
			}
		}
	}

	return info
}

// shouldSkipRateLimit determines if rate limiting should be skipped.
func (rlm *RateLimitMiddleware) shouldSkipRateLimit(reqInfo *RateLimitRequestInfo) bool {
	// Check skip paths
	for _, skipPath := range rlm.config.SkipPaths {
		if reqInfo.Path == skipPath {
			return true
		}
	}

	// Check skip methods
	for _, skipMethod := range rlm.config.SkipMethods {
		if reqInfo.Method == skipMethod {
			return true
		}
	}

	// Check skip IPs
	for _, skipIP := range rlm.config.SkipIPs {
		if reqInfo.IP == skipIP {
			return true
		}
	}

	// Check skip users
	for _, skipUser := range rlm.config.SkipUsers {
		if reqInfo.UserID == skipUser {
			return true
		}
	}

	return false
}

// generateKey generates a unique key for rate limiting.
func (rlm *RateLimitMiddleware) generateKey(ctx context.Context, req interface{}, reqInfo *RateLimitRequestInfo) string {
	// Use custom key function if provided
	if rlm.config.KeyFunc != nil {
		return rlm.config.KeyFunc(ctx, req)
	}

	var keyParts []string

	// Add IP to key if configured
	if rlm.config.IdentifyByIP && reqInfo.IP != "" {
		keyParts = append(keyParts, "ip:"+reqInfo.IP)
	}

	// Add user ID to key if configured
	if rlm.config.IdentifyByUser && reqInfo.UserID != "" {
		keyParts = append(keyParts, "user:"+reqInfo.UserID)
	}

	// Add header value to key if configured
	if rlm.config.IdentifyByHeader != "" {
		if headerValue, exists := reqInfo.Headers[rlm.config.IdentifyByHeader]; exists {
			keyParts = append(keyParts, "header:"+headerValue)
		}
	}

	// Add query parameter value to key if configured
	if rlm.config.IdentifyByQuery != "" {
		if queryValue, exists := reqInfo.Query[rlm.config.IdentifyByQuery]; exists {
			keyParts = append(keyParts, "query:"+queryValue)
		}
	}

	// Default to IP if no other identification method is configured
	if len(keyParts) == 0 && reqInfo.IP != "" {
		keyParts = append(keyParts, "ip:"+reqInfo.IP)
	}

	if len(keyParts) == 0 {
		keyParts = append(keyParts, "anonymous")
	}

	return strings.Join(keyParts, "|")
}

// getLimiter gets or creates a rate limiter for the given key.
func (rlm *RateLimitMiddleware) getLimiter(key string, reqInfo *RateLimitRequestInfo) *RateLimiter {
	rlm.mu.RLock()
	limiter, exists := rlm.limiters[key]
	rlm.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter
	rateLimit := rlm.getRateLimitForRequest(reqInfo)
	limiter = NewRateLimiter(key, rateLimit, rlm.config.Strategy)

	rlm.mu.Lock()
	rlm.limiters[key] = limiter
	rlm.mu.Unlock()

	rlm.GetLogger().Debug("Created new rate limiter for key: %s", key)
	return limiter
}

// getRateLimitForRequest gets the appropriate rate limit for a request.
func (rlm *RateLimitMiddleware) getRateLimitForRequest(reqInfo *RateLimitRequestInfo) RateLimit {
	// Check path-specific limits
	if pathLimit, exists := rlm.config.PathLimits[reqInfo.Path]; exists {
		return pathLimit
	}

	// Check method-specific limits
	if methodLimit, exists := rlm.config.MethodLimits[reqInfo.Method]; exists {
		return methodLimit
	}

	// Return default limit
	return RateLimit{
		RequestsPerSecond: rlm.config.RequestsPerSecond,
		RequestsPerMinute: rlm.config.RequestsPerMinute,
		RequestsPerHour:   rlm.config.RequestsPerHour,
		RequestsPerDay:    rlm.config.RequestsPerDay,
		BurstSize:         rlm.config.BurstSize,
	}
}

// createRateLimitErrorResponse creates an error response for rate limit exceeded.
func (rlm *RateLimitMiddleware) createRateLimitErrorResponse(rateLimitCtx *RateLimitContext) (interface{}, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	if rlm.config.RetryAfterHeader && rateLimitCtx.RetryAfter > 0 {
		headers["Retry-After"] = fmt.Sprintf("%.0f", rateLimitCtx.RetryAfter.Seconds())
	}

	if rlm.config.IncludeHeaders {
		headers["X-RateLimit-Limit"] = fmt.Sprintf("%.0f", rlm.config.RequestsPerSecond)
		headers["X-RateLimit-Remaining"] = fmt.Sprintf("%d", rateLimitCtx.Remaining)
		headers["X-RateLimit-Reset"] = fmt.Sprintf("%d", rateLimitCtx.ResetTime.Unix())
	}

	response := map[string]interface{}{
		"status_code": rlm.config.ErrorStatusCode,
		"headers":     headers,
		"body":        fmt.Sprintf(`{"error": "%s", "retry_after": %.0f}`, rlm.config.ErrorMessage, rateLimitCtx.RetryAfter.Seconds()),
	}

	return response, nil
}

// addRateLimitHeaders adds rate limit headers to the response.
func (rlm *RateLimitMiddleware) addRateLimitHeaders(resp interface{}, rateLimitCtx *RateLimitContext) {
	if httpResp, ok := resp.(map[string]interface{}); ok {
		headers, exists := httpResp["headers"]
		if !exists {
			headers = make(map[string]string)
			httpResp["headers"] = headers
		}

		if h, ok := headers.(map[string]string); ok {
			h["X-RateLimit-Limit"] = fmt.Sprintf("%.0f", rlm.config.RequestsPerSecond)
			h["X-RateLimit-Remaining"] = fmt.Sprintf("%d", rateLimitCtx.Remaining)
			h["X-RateLimit-Reset"] = fmt.Sprintf("%d", rateLimitCtx.ResetTime.Unix())
		}
	}
}

// startCleanupRoutine starts the cleanup routine for expired limiters.
func (rlm *RateLimitMiddleware) startCleanupRoutine() {
	if rlm.config.CleanupInterval > 0 {
		rlm.cleanupTicker = time.NewTicker(rlm.config.CleanupInterval)
		go rlm.cleanupRoutine()
	}
}

// cleanupRoutine periodically cleans up expired limiters.
func (rlm *RateLimitMiddleware) cleanupRoutine() {
	for {
		select {
		case <-rlm.cleanupTicker.C:
			rlm.cleanup()
		case <-rlm.stopCleanup:
			rlm.cleanupTicker.Stop()
			return
		}
	}
}

// cleanup removes expired limiters.
func (rlm *RateLimitMiddleware) cleanup() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	now := time.Now()
	expiredKeys := []string{}

	for key, limiter := range rlm.limiters {
		limiter.mu.Lock()
		lastActivity := limiter.lastReset
		if lastActivity.IsZero() {
			lastActivity = now
		}
		limiter.mu.Unlock()

		// Remove limiters that haven't been used recently
		if now.Sub(lastActivity) > rlm.config.CleanupInterval*2 {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(rlm.limiters, key)
	}

	if len(expiredKeys) > 0 {
		rlm.GetLogger().Debug("Cleaned up %d expired rate limiters", len(expiredKeys))
	}
}

// Reset resets all metrics and clears all limiters.
func (rlm *RateLimitMiddleware) Reset() {
	atomic.StoreInt64(&rlm.totalRequests, 0)
	atomic.StoreInt64(&rlm.allowedRequests, 0)
	atomic.StoreInt64(&rlm.blockedRequests, 0)
	atomic.AddInt64(&rlm.resetCount, 1)

	rlm.mu.Lock()
	rlm.limiters = make(map[string]*RateLimiter)
	rlm.mu.Unlock()

	rlm.startTime = time.Now()
	rlm.GetLogger().Info("Rate limit middleware metrics and limiters reset")
}

// Stop stops the cleanup routine.
func (rlm *RateLimitMiddleware) Stop() {
	if rlm.cleanupTicker != nil {
		select {
		case rlm.stopCleanup <- true:
		default:
		}
	}
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(key string, config RateLimit, strategy RateLimitStrategy) *RateLimiter {
	return &RateLimiter{
		key:         key,
		config:      config,
		strategy:    strategy,
		tokens:      float64(config.BurstSize),
		lastRefill:  time.Now(),
		windowStart: time.Now(),
		requests:    make([]time.Time, 0),
		lastReset:   time.Now(),
	}
}

// checkLimit checks if a request is allowed based on the rate limit.
func (rl *RateLimiter) checkLimit() *RateLimitContext {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	switch rl.strategy {
	case TokenBucket:
		return rl.checkTokenBucket(now)
	case FixedWindow:
		return rl.checkFixedWindow(now)
	case SlidingWindow:
		return rl.checkSlidingWindow(now)
	default:
		return rl.checkTokenBucket(now)
	}
}

// checkTokenBucket implements token bucket algorithm.
func (rl *RateLimiter) checkTokenBucket(now time.Time) *RateLimitContext {
	// Refill tokens
	elapsed := now.Sub(rl.lastRefill).Seconds()
	tokensToAdd := elapsed * rl.config.RequestsPerSecond
	rl.tokens = min(rl.tokens+tokensToAdd, float64(rl.config.BurstSize))
	rl.lastRefill = now

	ctx := &RateLimitContext{
		Remaining: int64(rl.tokens),
		ResetTime: now.Add(time.Duration(float64(time.Second) / rl.config.RequestsPerSecond)),
	}

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		rl.allowedRequests++
		ctx.Allowed = true
	} else {
		rl.blockedRequests++
		ctx.Allowed = false
		ctx.RetryAfter = time.Duration(float64(time.Second) / rl.config.RequestsPerSecond)
	}

	rl.totalRequests++
	return ctx
}

// checkFixedWindow implements fixed window algorithm.
func (rl *RateLimiter) checkFixedWindow(now time.Time) *RateLimitContext {
	windowDuration := time.Second
	if rl.config.RequestsPerMinute > 0 {
		windowDuration = time.Minute
	}

	// Check if we need a new window
	if now.Sub(rl.windowStart) >= windowDuration {
		rl.windowStart = now
		rl.windowRequests = 0
	}

	limit := int64(rl.config.RequestsPerSecond)
	if windowDuration == time.Minute {
		limit = rl.config.RequestsPerMinute
	}

	ctx := &RateLimitContext{
		WindowStart:      rl.windowStart,
		RequestsInWindow: rl.windowRequests,
		Remaining:        max(0, limit-rl.windowRequests),
		ResetTime:        rl.windowStart.Add(windowDuration),
	}

	if rl.windowRequests < limit {
		rl.windowRequests++
		rl.allowedRequests++
		ctx.Allowed = true
	} else {
		rl.blockedRequests++
		ctx.Allowed = false
		ctx.RetryAfter = rl.windowStart.Add(windowDuration).Sub(now)
	}

	rl.totalRequests++
	return ctx
}

// checkSlidingWindow implements sliding window algorithm.
func (rl *RateLimiter) checkSlidingWindow(now time.Time) *RateLimitContext {
	windowDuration := time.Minute

	// Remove old requests outside the window
	cutoff := now.Add(-windowDuration)
	validRequests := []time.Time{}
	for _, reqTime := range rl.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	rl.requests = validRequests

	limit := rl.config.RequestsPerMinute

	ctx := &RateLimitContext{
		RequestsInWindow: int64(len(rl.requests)),
		Remaining:        max(0, limit-int64(len(rl.requests))),
		ResetTime:        now.Add(windowDuration),
	}

	if int64(len(rl.requests)) < limit {
		rl.requests = append(rl.requests, now)
		rl.allowedRequests++
		ctx.Allowed = true
	} else {
		rl.blockedRequests++
		ctx.Allowed = false
		if len(rl.requests) > 0 {
			ctx.RetryAfter = rl.requests[0].Add(windowDuration).Sub(now)
		}
	}

	rl.totalRequests++
	return ctx
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
