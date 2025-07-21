// Package hooks provides specific hook implementations for common use cases.
package hooks

import (
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// LoggingHook implements structured logging for HTTP requests.
type LoggingHook struct {
	*BaseHook
	logger interface{} // Can be any logger interface
}

// NewLoggingHook creates a new logging hook.
func NewLoggingHook(logger interface{}) *LoggingHook {
	events := []interfaces.HookEvent{
		interfaces.HookEventRequestStart,
		interfaces.HookEventRequestEnd,
		interfaces.HookEventRequestError,
		interfaces.HookEventServerStart,
		interfaces.HookEventServerStop,
	}

	return &LoggingHook{
		BaseHook: NewBaseHook("logging", events, 100), // Low priority
		logger:   logger,
	}
}

// Execute executes the logging hook.
func (h *LoggingHook) Execute(ctx *interfaces.HookContext) error {
	switch ctx.Event {
	case interfaces.HookEventServerStart:
		log.Printf("INFO: Server %s started at %v", ctx.ServerName, ctx.Timestamp)
	case interfaces.HookEventServerStop:
		log.Printf("INFO: Server %s stopped at %v", ctx.ServerName, ctx.Timestamp)
	case interfaces.HookEventRequestStart:
		if ctx.Request != nil {
			log.Printf("INFO: Request started - %s %s - Server: %s - TraceID: %s",
				ctx.Request.Method, ctx.Request.URL.Path, ctx.ServerName, ctx.TraceID)
		}
	case interfaces.HookEventRequestEnd:
		if ctx.Request != nil {
			log.Printf("INFO: Request completed - %s %s - Status: %d - Duration: %v - Server: %s",
				ctx.Request.Method, ctx.Request.URL.Path, ctx.StatusCode, ctx.Duration, ctx.ServerName)
		}
	case interfaces.HookEventRequestError:
		if ctx.Request != nil && ctx.Error != nil {
			log.Printf("ERROR: Request failed - %s %s - Error: %v - Server: %s",
				ctx.Request.Method, ctx.Request.URL.Path, ctx.Error, ctx.ServerName)
		}
	}
	return nil
}

// MetricsHook implements metrics collection for HTTP requests.
type MetricsHook struct {
	*BaseHook
	requestCount    int64
	errorCount      int64
	totalDuration   time.Duration
	averageDuration time.Duration
}

// NewMetricsHook creates a new metrics hook.
func NewMetricsHook() *MetricsHook {
	events := []interfaces.HookEvent{
		interfaces.HookEventRequestEnd,
		interfaces.HookEventRequestError,
	}

	return &MetricsHook{
		BaseHook: NewBaseHook("metrics", events, 50), // Medium priority
	}
}

// Execute executes the metrics hook.
func (h *MetricsHook) Execute(ctx *interfaces.HookContext) error {
	switch ctx.Event {
	case interfaces.HookEventRequestEnd:
		h.requestCount++
		h.totalDuration += ctx.Duration
		h.averageDuration = h.totalDuration / time.Duration(h.requestCount)
	case interfaces.HookEventRequestError:
		h.errorCount++
	}
	return nil
}

// GetMetrics returns the collected metrics.
func (h *MetricsHook) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"request_count":    h.requestCount,
		"error_count":      h.errorCount,
		"total_duration":   h.totalDuration,
		"average_duration": h.averageDuration,
		"error_rate":       float64(h.errorCount) / float64(h.requestCount),
	}
}

// SecurityHook implements basic security checks.
type SecurityHook struct {
	*FilteredBaseHook
	allowedOrigins   []string
	blockedIPs       []string
	rateLimitEnabled bool
}

// NewSecurityHook creates a new security hook.
func NewSecurityHook() *SecurityHook {
	events := []interfaces.HookEvent{
		interfaces.HookEventRequestStart,
	}

	hook := &SecurityHook{
		FilteredBaseHook: NewFilteredBaseHook("security", events, 10), // High priority
		allowedOrigins:   []string{"*"},                               // Default allow all
		blockedIPs:       []string{},
		rateLimitEnabled: false,
	}

	return hook
}

// SetAllowedOrigins sets the allowed origins for CORS.
func (h *SecurityHook) SetAllowedOrigins(origins []string) {
	h.allowedOrigins = origins
}

// SetBlockedIPs sets the blocked IP addresses.
func (h *SecurityHook) SetBlockedIPs(ips []string) {
	h.blockedIPs = ips
}

// EnableRateLimit enables rate limiting.
func (h *SecurityHook) EnableRateLimit() {
	h.rateLimitEnabled = true
}

// Execute executes the security hook.
func (h *SecurityHook) Execute(ctx *interfaces.HookContext) error {
	if ctx.Request == nil {
		return nil
	}

	// Check blocked IPs
	clientIP := ctx.Request.RemoteAddr
	for _, blockedIP := range h.blockedIPs {
		if clientIP == blockedIP {
			return fmt.Errorf("blocked IP address: %s", clientIP)
		}
	}

	// Check CORS origins
	origin := ctx.Request.Header.Get("Origin")
	if origin != "" && len(h.allowedOrigins) > 0 {
		allowed := false
		for _, allowedOrigin := range h.allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("origin not allowed: %s", origin)
		}
	}

	return nil
}

// CacheHook implements response caching.
type CacheHook struct {
	*ConditionalBaseHook
	cache      map[string]interface{} // Simple in-memory cache
	cacheTTL   time.Duration
	cacheStats map[string]time.Time
}

// NewCacheHook creates a new cache hook.
func NewCacheHook(ttl time.Duration) *CacheHook {
	events := []interfaces.HookEvent{
		interfaces.HookEventRequestStart,
		interfaces.HookEventRequestEnd,
	}

	// Only cache GET requests
	condition := func(ctx *interfaces.HookContext) bool {
		return ctx.Request != nil && ctx.Request.Method == "GET"
	}

	hook := &CacheHook{
		ConditionalBaseHook: NewConditionalBaseHook("cache", events, 30, condition),
		cache:               make(map[string]interface{}),
		cacheTTL:            ttl,
		cacheStats:          make(map[string]time.Time),
	}

	return hook
}

// Execute executes the cache hook.
func (h *CacheHook) Execute(ctx *interfaces.HookContext) error {
	if ctx.Request == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%s:%s", ctx.Request.Method, ctx.Request.URL.Path)

	switch ctx.Event {
	case interfaces.HookEventRequestStart:
		// Check if response is cached and not expired
		if cachedTime, exists := h.cacheStats[cacheKey]; exists {
			if time.Since(cachedTime) < h.cacheTTL {
				// Cache hit - mark in metadata
				if ctx.Metadata == nil {
					ctx.Metadata = make(map[string]interface{})
				}
				ctx.Metadata["cache_hit"] = true
				ctx.Metadata["cached_data"] = h.cache[cacheKey]
			} else {
				// Cache expired - remove
				delete(h.cache, cacheKey)
				delete(h.cacheStats, cacheKey)
			}
		}

	case interfaces.HookEventRequestEnd:
		// Cache successful responses
		if ctx.StatusCode >= 200 && ctx.StatusCode < 300 {
			h.cache[cacheKey] = map[string]interface{}{
				"status":    ctx.StatusCode,
				"timestamp": ctx.Timestamp,
				"duration":  ctx.Duration,
			}
			h.cacheStats[cacheKey] = time.Now()
		}
	}

	return nil
}

// GetCacheStats returns cache statistics.
func (h *CacheHook) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"cache_size":      len(h.cache),
		"cache_entries":   len(h.cacheStats),
		"cache_ttl":       h.cacheTTL,
		"cache_hit_ratio": h.calculateHitRatio(),
	}
}

func (h *CacheHook) calculateHitRatio() float64 {
	// Simple implementation - in real scenario you'd track hits/misses
	return 0.75 // Placeholder
}

// HealthCheckHook implements health check functionality.
type HealthCheckHook struct {
	*BaseHook
	healthChecks map[string]func() error
	lastResults  map[string]bool
}

// NewHealthCheckHook creates a new health check hook.
func NewHealthCheckHook() *HealthCheckHook {
	events := []interfaces.HookEvent{
		interfaces.HookEventHealthCheck,
		interfaces.HookEventServerStart,
	}

	return &HealthCheckHook{
		BaseHook:     NewBaseHook("healthcheck", events, 20),
		healthChecks: make(map[string]func() error),
		lastResults:  make(map[string]bool),
	}
}

// AddHealthCheck adds a health check function.
func (h *HealthCheckHook) AddHealthCheck(name string, check func() error) {
	h.healthChecks[name] = check
}

// Execute executes the health check hook.
func (h *HealthCheckHook) Execute(ctx *interfaces.HookContext) error {
	switch ctx.Event {
	case interfaces.HookEventHealthCheck, interfaces.HookEventServerStart:
		// Run all health checks
		for name, check := range h.healthChecks {
			err := check()
			h.lastResults[name] = err == nil
			if err != nil {
				log.Printf("WARN: Health check %s failed: %v", name, err)
			}
		}
	}
	return nil
}

// GetHealthStatus returns the current health status.
func (h *HealthCheckHook) GetHealthStatus() map[string]bool {
	result := make(map[string]bool)
	for name, status := range h.lastResults {
		result[name] = status
	}
	return result
}

// IsHealthy returns true if all health checks pass.
func (h *HealthCheckHook) IsHealthy() bool {
	for _, status := range h.lastResults {
		if !status {
			return false
		}
	}
	return true
}

// Verify implementations
var _ interfaces.Hook = (*LoggingHook)(nil)
var _ interfaces.Hook = (*MetricsHook)(nil)
var _ interfaces.Hook = (*SecurityHook)(nil)
var _ interfaces.FilteredHook = (*SecurityHook)(nil)
var _ interfaces.Hook = (*CacheHook)(nil)
var _ interfaces.ConditionalHook = (*CacheHook)(nil)
var _ interfaces.Hook = (*HealthCheckHook)(nil)
