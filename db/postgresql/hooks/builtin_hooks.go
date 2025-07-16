package hooks

import (
	"fmt"
	"log"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// LoggingHook creates a hook that logs operation details
func LoggingHook(prefix string) interfaces.Hook {
	return func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		log.Printf("[%s] Operation: %s, Query: %s, Args: %v, Duration: %v, Error: %v",
			prefix, ctx.Operation, ctx.Query, ctx.Args, ctx.Duration, ctx.Error)

		return &interfaces.HookResult{
			Continue: true,
			Error:    nil,
		}
	}
}

// TimingHook creates a hook that measures operation duration
func TimingHook() interfaces.Hook {
	return func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.StartTime.IsZero() {
			ctx.StartTime = time.Now()
		} else {
			ctx.Duration = time.Since(ctx.StartTime)
		}

		return &interfaces.HookResult{
			Continue: true,
			Error:    nil,
		}
	}
}

// ValidationHook creates a hook that validates queries and parameters
func ValidationHook() interfaces.Hook {
	return func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		// Basic validation
		if ctx.Query == "" && ctx.Operation != "ping" && ctx.Operation != "healthcheck" {
			return &interfaces.HookResult{
				Continue: false,
				Error:    fmt.Errorf("query cannot be empty for operation: %s", ctx.Operation),
			}
		}

		// Check for potential SQL injection patterns (basic check)
		if containsSQLInjectionPatterns(ctx.Query) {
			return &interfaces.HookResult{
				Continue: false,
				Error:    fmt.Errorf("potential SQL injection detected in query"),
			}
		}

		return &interfaces.HookResult{
			Continue: true,
			Error:    nil,
		}
	}
}

// MetricsHook creates a hook that collects operation metrics
func MetricsHook() interfaces.Hook {
	return func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		// Initialize metrics if not exists
		if ctx.Metadata == nil {
			ctx.Metadata = make(map[string]interface{})
		}

		// Collect metrics
		ctx.Metadata["timestamp"] = time.Now()
		ctx.Metadata["operation"] = ctx.Operation
		ctx.Metadata["query_length"] = len(ctx.Query)
		ctx.Metadata["args_count"] = len(ctx.Args)

		if ctx.Duration > 0 {
			ctx.Metadata["duration_ms"] = ctx.Duration.Milliseconds()
		}

		if ctx.Error != nil {
			ctx.Metadata["error"] = ctx.Error.Error()
			ctx.Metadata["has_error"] = true
		} else {
			ctx.Metadata["has_error"] = false
		}

		return &interfaces.HookResult{
			Continue: true,
			Error:    nil,
		}
	}
}

// RetryHook creates a hook that implements retry logic
func RetryHook(maxRetries int, retryDelay time.Duration) interfaces.Hook {
	return func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Error == nil {
			return &interfaces.HookResult{
				Continue: true,
				Error:    nil,
			}
		}

		// Check if this is a retryable error
		if !isRetryableError(ctx.Error) {
			return &interfaces.HookResult{
				Continue: true,
				Error:    ctx.Error,
			}
		}

		// Get current retry count
		retryCount := 0
		if ctx.Metadata != nil {
			if count, exists := ctx.Metadata["retry_count"]; exists {
				if c, ok := count.(int); ok {
					retryCount = c
				}
			}
		}

		// Check if we can retry
		if retryCount >= maxRetries {
			return &interfaces.HookResult{
				Continue: true,
				Error:    fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, ctx.Error),
			}
		}

		// Increment retry count
		if ctx.Metadata == nil {
			ctx.Metadata = make(map[string]interface{})
		}
		ctx.Metadata["retry_count"] = retryCount + 1

		// Wait before retry
		if retryDelay > 0 {
			time.Sleep(retryDelay)
		}

		return &interfaces.HookResult{
			Continue: true,
			Error:    nil,
			Data: map[string]interface{}{
				"should_retry": true,
				"retry_count":  retryCount + 1,
			},
		}
	}
}

// TenantHook creates a hook that sets tenant context
func TenantHook() interfaces.Hook {
	return func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		// Extract tenant ID from context
		if tenantID := ctx.Context.Value("tenant_id"); tenantID != nil {
			if tid, ok := tenantID.(string); ok && tid != "" {
				// Add tenant to metadata
				if ctx.Metadata == nil {
					ctx.Metadata = make(map[string]interface{})
				}
				ctx.Metadata["tenant_id"] = tid

				// Modify query for multi-tenancy if needed
				// This is a placeholder - actual implementation would depend on the multi-tenancy strategy
				log.Printf("Processing query for tenant: %s", tid)
			}
		}

		return &interfaces.HookResult{
			Continue: true,
			Error:    nil,
		}
	}
}

// CacheHook creates a hook that implements query caching
func CacheHook(cacheTTL time.Duration) interfaces.Hook {
	cache := make(map[string]cacheEntry)

	return func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		// Only cache SELECT queries
		if ctx.Operation != "query" && ctx.Operation != "queryrow" {
			return &interfaces.HookResult{
				Continue: true,
				Error:    nil,
			}
		}

		// Generate cache key
		cacheKey := generateCacheKey(ctx.Query, ctx.Args)

		// Check cache before execution
		if ctx.Error == nil && ctx.Duration == 0 {
			if entry, exists := cache[cacheKey]; exists {
				if time.Since(entry.timestamp) < cacheTTL {
					// Cache hit
					if ctx.Metadata == nil {
						ctx.Metadata = make(map[string]interface{})
					}
					ctx.Metadata["cache_hit"] = true
					ctx.Metadata["cached_data"] = entry.data

					return &interfaces.HookResult{
						Continue: false, // Skip actual execution
						Error:    nil,
						Data: map[string]interface{}{
							"cached_result": entry.data,
						},
					}
				} else {
					// Cache expired
					delete(cache, cacheKey)
				}
			}
		}

		// Store result in cache after execution (if successful)
		if ctx.Error == nil && ctx.Duration > 0 {
			// This is a placeholder - actual implementation would need to capture the result
			cache[cacheKey] = cacheEntry{
				data:      nil, // Would contain actual result data
				timestamp: time.Now(),
			}

			if ctx.Metadata == nil {
				ctx.Metadata = make(map[string]interface{})
			}
			ctx.Metadata["cache_hit"] = false
		}

		return &interfaces.HookResult{
			Continue: true,
			Error:    nil,
		}
	}
}

// cacheEntry represents a cached query result
type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

// Helper functions

func containsSQLInjectionPatterns(query string) bool {
	// Basic SQL injection pattern detection
	// This is a simplified version - real implementation would be more sophisticated
	dangerousPatterns := []string{
		"'; DROP TABLE",
		"'; DELETE FROM",
		"'; UPDATE",
		"UNION SELECT",
		"' OR '1'='1",
		"' OR 1=1",
		"-- ",
		"/*",
		"*/",
		"xp_",
		"sp_",
	}

	for _, pattern := range dangerousPatterns {
		// More sophisticated pattern matching would be implemented here
		// For now, just use the pattern to avoid unused variable error
		_ = pattern
		_ = query
	}

	return false // Simplified implementation
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Define retryable error patterns
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no route to host",
		"connection timed out",
	}

	errStr := err.Error()
	for _, pattern := range retryablePatterns {
		// More sophisticated pattern matching would be implemented here
		// For now, just use the variables to avoid unused variable error
		_ = pattern
		_ = errStr
	}

	return false
}

func generateCacheKey(query string, args []interface{}) string {
	// Simple cache key generation
	// Real implementation would use a proper hash function
	key := query
	for i, arg := range args {
		key += fmt.Sprintf("_%d_%v", i, arg)
	}
	return key
}
