package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func main() {
	log.Println("🎣 Custom Hooks Example")
	log.Println("=====================")

	// Setup custom hooks examples
	demonstrateCustomHooks()
}

func demonstrateCustomHooks() {
	hookFactory := hooks.NewCustomHookFactory()

	// 1. Simple Request Logger Hook
	log.Println("📝 Creating Simple Request Logger Hook...")
	requestLogger := hookFactory.NewSimpleHook(
		"request-logger",
		[]interfaces.HookEvent{
			interfaces.HookEventRequestStart,
			interfaces.HookEventRequestEnd,
		},
		100, // priority
		func(ctx *interfaces.HookContext) error {
			if ctx.Event == interfaces.HookEventRequestStart {
				log.Printf("📥 [%s] %s %s - Request started",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path)
			} else if ctx.Event == interfaces.HookEventRequestEnd {
				log.Printf("📤 [%s] %s %s - %d (%v)",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path,
					ctx.StatusCode, ctx.Duration)
			}
			return nil
		},
	)

	// 2. API-Only Security Monitor Hook (Conditional)
	log.Println("🔒 Creating API Security Monitor Hook...")
	apiSecurityHook := hookFactory.NewConditionalHook(
		"api-security-monitor",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		200,
		func(ctx *interfaces.HookContext) bool {
			// Only monitor API endpoints
			return strings.HasPrefix(ctx.Request.URL.Path, "/api/")
		},
		func(ctx *interfaces.HookContext) error {
			// Check for security headers
			authHeader := ctx.Request.Header.Get("Authorization")
			userAgent := ctx.Request.Header.Get("User-Agent")

			log.Printf("🔍 [%s] API Security Check: %s %s",
				ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path)
			log.Printf("   Auth: %t, User-Agent: %s",
				authHeader != "", userAgent)

			// Log suspicious activity
			if authHeader == "" && ctx.Request.Method != "GET" {
				log.Printf("⚠️  [%s] Suspicious: Unauthenticated %s request to %s",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path)
			}

			return nil
		},
	)

	// 3. Performance Monitor Hook (Filtered)
	log.Println("⚡ Creating Performance Monitor Hook...")
	performanceHook := hookFactory.NewFilteredHook(
		"performance-monitor",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		300,
		func(path string) bool {
			// Monitor all paths except health checks
			return !strings.HasPrefix(path, "/health")
		},
		func(method string) bool {
			// Monitor all methods except OPTIONS
			return method != "OPTIONS"
		},
		func(ctx *interfaces.HookContext) error {
			if ctx.Duration > 1*time.Second {
				log.Printf("🐌 [%s] SLOW REQUEST ALERT: %s %s took %v (Status: %d)",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path,
					ctx.Duration, ctx.StatusCode)
			} else if ctx.Duration > 500*time.Millisecond {
				log.Printf("⏱️  [%s] Performance Warning: %s %s took %v",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path, ctx.Duration)
			}
			return nil
		},
	)

	// 4. Async Analytics Hook
	log.Println("📊 Creating Async Analytics Hook...")
	analyticsHook := hookFactory.NewAsyncHook(
		"analytics-processor",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		400,
		10,            // buffer size
		5*time.Second, // timeout
		func(ctx *interfaces.HookContext) error {
			// Simulate analytics processing (this runs asynchronously)
			time.Sleep(50 * time.Millisecond)

			analytics := map[string]interface{}{
				"method":     ctx.Request.Method,
				"path":       ctx.Request.URL.Path,
				"status":     ctx.StatusCode,
				"duration":   ctx.Duration.Milliseconds(),
				"user_agent": ctx.Request.Header.Get("User-Agent"),
				"timestamp":  ctx.Timestamp,
			}

			log.Printf("📈 [%s] Analytics processed: %+v", ctx.TraceID, analytics)
			return nil
		},
	)

	// 5. Complex Error Handler Hook using Builder Pattern
	log.Println("🚨 Creating Complex Error Handler Hook...")
	_, err := hooks.NewCustomHookBuilder().
		WithName("error-handler").
		WithEvents(interfaces.HookEventRequestError, interfaces.HookEventRequestEnd).
		WithPriority(50). // High priority for error handling
		WithPathFilter(func(path string) bool {
			// Handle errors for all paths except health
			return !strings.HasPrefix(path, "/health")
		}).
		WithCondition(func(ctx *interfaces.HookContext) bool {
			// Handle errors and 4xx/5xx status codes
			return ctx.Error != nil || ctx.StatusCode >= 400
		}).
		WithAsyncExecution(5, 3*time.Second).
		WithExecuteFunc(func(ctx *interfaces.HookContext) error {
			if ctx.Error != nil {
				log.Printf("❌ [%s] REQUEST ERROR: %s %s - %v",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path, ctx.Error)

				// Could send to error tracking service like Sentry
				// sentry.CaptureException(ctx.Error)
			} else if ctx.StatusCode >= 500 {
				log.Printf("💥 [%s] SERVER ERROR: %s %s - Status %d",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path, ctx.StatusCode)
			} else if ctx.StatusCode >= 400 {
				log.Printf("⚠️  [%s] CLIENT ERROR: %s %s - Status %d",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path, ctx.StatusCode)
			}

			return nil
		}).
		Build()

	if err != nil {
		log.Printf("Error creating error handler hook: %v", err)
		return
	}

	// 6. Business Logic Hook with Advanced Filtering
	log.Println("💼 Creating Business Logic Hook...")
	businessLogicHook, err := hooks.NewCustomHookBuilder().
		WithName("business-logic").
		WithEvents(interfaces.HookEventRequestStart).
		WithPriority(250).
		WithPathFilter(func(path string) bool {
			// Only business endpoints
			return strings.HasPrefix(path, "/api/users")
		}).
		WithMethodFilter(func(method string) bool {
			// Only modification operations
			return method == "POST" || method == "PUT" || method == "DELETE"
		}).
		WithHeaderFilter(func(headers http.Header) bool {
			// Only requests with content type
			contentType := headers.Get("Content-Type")
			return contentType != ""
		}).
		WithExecuteFunc(func(ctx *interfaces.HookContext) error {
			log.Printf("💼 [%s] Business Logic: Processing %s %s",
				ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path)

			// Could implement business rules here:
			// - Rate limiting per user
			// - Data validation
			// - Audit logging
			// - Compliance checks

			return nil
		}).
		Build()

	if err != nil {
		log.Printf("Error creating business logic hook: %v", err)
		return
	}

	// 7. Demonstrate hook execution with simulated requests
	log.Println("🧪 Testing hooks with simulated requests...")

	// Create mock HTTP requests
	requests := []*http.Request{
		createMockRequest("GET", "/", ""),
		createMockRequest("GET", "/api/users", ""),
		createMockRequest("POST", "/api/users", "application/json"),
		createMockRequest("GET", "/health", ""),
		createMockRequest("DELETE", "/api/users/123", ""),
	}

	// Simulate hook execution for each request
	for i, req := range requests {
		log.Printf("\n--- Simulating Request %d: %s %s ---", i+1, req.Method, req.URL.Path)

		// Create hook context
		ctx := &interfaces.HookContext{
			Event:         interfaces.HookEventRequestStart,
			ServerName:    "example-server",
			Timestamp:     time.Now(),
			Request:       req,
			TraceID:       fmt.Sprintf("trace-%d", i+1),
			CorrelationID: fmt.Sprintf("corr-%d", i+1),
		}

		// Execute hooks that should run for this request
		executeHookIfShould(requestLogger, ctx)
		executeHookIfShould(apiSecurityHook, ctx)
		executeHookIfShould(businessLogicHook, ctx)

		// Simulate request processing
		time.Sleep(10 * time.Millisecond)

		// Simulate request end
		ctx.Event = interfaces.HookEventRequestEnd
		ctx.StatusCode = 200
		ctx.Duration = time.Since(ctx.Timestamp)

		executeHookIfShould(requestLogger, ctx)
		executeHookIfShould(performanceHook, ctx)

		// Execute analytics hook asynchronously
		if analyticsHook.ShouldExecute(ctx) {
			errChan := analyticsHook.ExecuteAsync(ctx)
			go func() {
				if err := <-errChan; err != nil {
					log.Printf("Analytics hook error: %v", err)
				}
			}()
		}
	}

	// Wait a bit for async hooks to complete
	time.Sleep(200 * time.Millisecond)

	log.Println("\n✅ All custom hooks demonstrated successfully!")
	log.Println("📋 Summary of demonstrated features:")
	log.Println("   • Simple hook creation with factory")
	log.Println("   • Conditional hooks with custom logic")
	log.Println("   • Filtered hooks with path/method/header filters")
	log.Println("   • Async hooks with timeout and buffer configuration")
	log.Println("   • Builder pattern for complex hook configuration")
	log.Println("   • Advanced filtering combinations")
	log.Println("   • Hook execution prioritization")
	log.Println("   • Type-safe hook implementations")
}

func createMockRequest(method, path, contentType string) *http.Request {
	req, _ := http.NewRequest(method, "http://localhost:8080"+path, nil)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("User-Agent", "CustomHooksExample/1.0")
	return req
}

func executeHookIfShould(hook interfaces.Hook, ctx *interfaces.HookContext) {
	if hook.ShouldExecute(ctx) {
		if err := hook.Execute(ctx); err != nil {
			log.Printf("Hook execution error: %v", err)
		}
	}
}
