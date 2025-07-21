// Package examples provides usage examples for custom hooks and middleware.
package examples

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
)

// ExampleCustomHooks demonstrates how to create and use custom hooks.
func ExampleCustomHooks() {
	// Create a custom hook factory
	hookFactory := hooks.NewCustomHookFactory()

	// Example 1: Simple logging hook
	loggingHook := hookFactory.NewSimpleHook(
		"request-logger",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart, interfaces.HookEventRequestEnd},
		100, // priority
		func(ctx *interfaces.HookContext) error {
			if ctx.Event == interfaces.HookEventRequestStart {
				log.Printf("Request started: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
			} else if ctx.Event == interfaces.HookEventRequestEnd {
				log.Printf("Request completed: %s %s (Status: %d, Duration: %v)",
					ctx.Request.Method, ctx.Request.URL.Path, ctx.StatusCode, ctx.Duration)
			}
			return nil
		},
	)

	// Example 2: Conditional hook for API endpoints only
	apiOnlyHook := hookFactory.NewConditionalHook(
		"api-only-hook",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		200,
		func(ctx *interfaces.HookContext) bool {
			return strings.HasPrefix(ctx.Request.URL.Path, "/api/")
		},
		func(ctx *interfaces.HookContext) error {
			log.Printf("API request: %s", ctx.Request.URL.Path)
			return nil
		},
	)

	// Example 3: Filtered hook for specific methods and paths
	filteredHook := hookFactory.NewFilteredHook(
		"post-delete-hook",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		300,
		func(path string) bool { return strings.Contains(path, "/users/") },        // path filter
		func(method string) bool { return method == "POST" || method == "DELETE" }, // method filter
		func(ctx *interfaces.HookContext) error {
			log.Printf("User modification request: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
			return nil
		},
	)

	// Example 4: Async hook for heavy operations
	asyncHook := hookFactory.NewAsyncHook(
		"analytics-hook",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		400,
		10,            // buffer size
		5*time.Second, // timeout
		func(ctx *interfaces.HookContext) error {
			// Simulate analytics processing
			time.Sleep(100 * time.Millisecond)
			log.Printf("Analytics processed for: %s", ctx.Request.URL.Path)
			return nil
		},
	)

	// Example 5: Complex hook using builder pattern
	complexHook, err := hooks.NewCustomHookBuilder().
		WithName("complex-hook").
		WithEvents(interfaces.HookEventRequestStart, interfaces.HookEventRequestError).
		WithPriority(150).
		WithPathFilter(func(path string) bool {
			return !strings.HasPrefix(path, "/health")
		}).
		WithMethodFilter(func(method string) bool {
			return method != "OPTIONS"
		}).
		WithCondition(func(ctx *interfaces.HookContext) bool {
			return ctx.Request.Header.Get("X-Monitor") == "true"
		}).
		WithAsyncExecution(5, 3*time.Second).
		WithExecuteFunc(func(ctx *interfaces.HookContext) error {
			log.Printf("Complex hook executed for monitored request: %s %s",
				ctx.Request.Method, ctx.Request.URL.Path)
			return nil
		}).
		Build()

	if err != nil {
		log.Printf("Error building complex hook: %v", err)
		return
	}

	// Use the hooks (this would typically be done in your HTTP server setup)
	log.Printf("Created hooks: %s, %s, %s, %s, %s",
		loggingHook.Name(),
		apiOnlyHook.Name(),
		filteredHook.Name(),
		asyncHook.Name(),
		complexHook.Name(),
	)
}

// ExampleCustomMiddleware demonstrates how to create and use custom middleware.
func ExampleCustomMiddleware() {
	// Create a custom middleware factory
	middlewareFactory := middleware.NewCustomMiddlewareFactory()

	// Example 1: Simple request ID middleware
	requestIDMiddleware := middlewareFactory.NewSimpleMiddleware(
		"request-id",
		100, // priority
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
				w.Header().Set("X-Request-ID", requestID)
				r.Header.Set("X-Request-ID", requestID)
				next.ServeHTTP(w, r)
			})
		},
	)

	// Example 2: Conditional middleware for admin routes
	adminOnlyMiddleware := middlewareFactory.NewConditionalMiddleware(
		"admin-auth",
		200,
		func(path string) bool { return !strings.HasPrefix(path, "/admin/") }, // skip non-admin paths
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check admin authentication
				if r.Header.Get("X-Admin-Token") == "" {
					http.Error(w, "Admin access required", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
			})
		},
	)

	// Example 3: Timing middleware
	timingMiddleware := middlewareFactory.NewTimingMiddleware(
		"request-timing",
		50,
		func(duration time.Duration, path string) {
			if duration > 1*time.Second {
				log.Printf("Slow request detected: %s took %v", path, duration)
			}
		},
	)

	// Example 4: Logging middleware
	loggingMiddleware := middlewareFactory.NewLoggingMiddleware(
		"request-logger",
		75,
		func(method, path string, statusCode int, duration time.Duration) {
			log.Printf("%s %s - %d (%v)", method, path, statusCode, duration)
		},
	)

	// Example 5: Complex middleware using builder pattern
	complexMiddleware, err := middleware.NewCustomMiddlewareBuilder().
		WithName("security-middleware").
		WithPriority(25).
		WithSkipPaths("/health", "/metrics").
		WithSkipFunc(func(path string) bool {
			return strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js")
		}).
		WithBeforeFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
		}).
		WithAfterFunc(func(w http.ResponseWriter, r *http.Request, statusCode int, duration time.Duration) {
			// Log security events
			if statusCode >= 400 {
				log.Printf("Security alert: %s %s returned %d", r.Method, r.URL.Path, statusCode)
			}
		}).
		WithWrapFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Rate limiting check (simplified)
				if r.Header.Get("X-Rate-Limit-Bypass") != "true" {
					// Simulate rate limit check
					next.ServeHTTP(w, r)
				} else {
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				}
			})
		}).
		Build()

	if err != nil {
		log.Printf("Error building complex middleware: %v", err)
		return
	}

	// Example 6: Custom CORS middleware with specific configuration
	corsMiddleware, err := middleware.NewCustomMiddlewareBuilder().
		WithName("custom-cors").
		WithPriority(10).
		WithWrapFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				origin := r.Header.Get("Origin")
				allowedOrigins := []string{"https://example.com", "https://app.example.com"}

				// Check if origin is allowed
				for _, allowedOrigin := range allowedOrigins {
					if origin == allowedOrigin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}

				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")

				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}

				next.ServeHTTP(w, r)
			})
		}).
		Build()

	if err != nil {
		log.Printf("Error building CORS middleware: %v", err)
		return
	}

	// Use the middleware (this would typically be done in your HTTP server setup)
	log.Printf("Created middleware: %s, %s, %s, %s, %s, %s",
		requestIDMiddleware.Name(),
		adminOnlyMiddleware.Name(),
		timingMiddleware.Name(),
		loggingMiddleware.Name(),
		complexMiddleware.Name(),
		corsMiddleware.Name(),
	)
}

// ExampleIntegration demonstrates how to integrate custom hooks and middleware.
func ExampleIntegration() {
	log.Println("=== Custom Hooks and Middleware Integration Example ===")

	// Create factories
	hookFactory := hooks.NewCustomHookFactory()
	middlewareFactory := middleware.NewCustomMiddlewareFactory()

	// Create a monitoring hook that works with timing middleware
	monitoringHook := hookFactory.NewSimpleHook(
		"performance-monitor",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		500,
		func(ctx *interfaces.HookContext) error {
			if ctx.Duration > 2*time.Second {
				log.Printf("PERFORMANCE ALERT: Request %s %s took %v (Status: %d)",
					ctx.Request.Method, ctx.Request.URL.Path, ctx.Duration, ctx.StatusCode)
			}
			return nil
		},
	)

	// Create middleware that adds metadata for hooks
	metadataMiddleware := middlewareFactory.NewSimpleMiddleware(
		"metadata-enricher",
		50,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Add metadata that hooks can use
				r.Header.Set("X-Start-Time", fmt.Sprintf("%d", time.Now().UnixNano()))
				r.Header.Set("X-User-Agent-Type", detectUserAgentType(r.UserAgent()))

				next.ServeHTTP(w, r)
			})
		},
	)

	// Create a hook that uses the metadata added by middleware
	enrichedHook := hookFactory.NewConditionalHook(
		"enriched-analytics",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		600,
		func(ctx *interfaces.HookContext) bool {
			// Only process for non-bot traffic
			return ctx.Request.Header.Get("X-User-Agent-Type") != "bot"
		},
		func(ctx *interfaces.HookContext) error {
			userAgentType := ctx.Request.Header.Get("X-User-Agent-Type")
			log.Printf("Analytics: %s request from %s user agent (Duration: %v)",
				ctx.Request.URL.Path, userAgentType, ctx.Duration)
			return nil
		},
	)

	log.Printf("Integrated components: %s hook, %s middleware, %s hook",
		monitoringHook.Name(),
		metadataMiddleware.Name(),
		enrichedHook.Name(),
	)
}

// detectUserAgentType is a helper function to categorize user agents
func detectUserAgentType(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "bot") || strings.Contains(userAgent, "crawler") {
		return "bot"
	}
	if strings.Contains(userAgent, "mobile") {
		return "mobile"
	}
	if strings.Contains(userAgent, "tablet") {
		return "tablet"
	}
	return "desktop"
}

// RunAllExamples runs all the examples
func RunAllExamples() {
	log.Println("=== Running Custom Hooks and Middleware Examples ===")

	log.Println("\n1. Custom Hooks Examples:")
	ExampleCustomHooks()

	log.Println("\n2. Custom Middleware Examples:")
	ExampleCustomMiddleware()

	log.Println("\n3. Integration Examples:")
	ExampleIntegration()

	log.Println("\n=== Examples completed ===")
}
