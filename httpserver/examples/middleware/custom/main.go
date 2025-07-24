package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
)

func main() {
	log.Println("🔧 Custom Middleware Example")
	log.Println("==========================")

	// Setup custom middleware examples
	demonstrateCustomMiddleware()
}

func demonstrateCustomMiddleware() {
	middlewareFactory := middleware.NewCustomMiddlewareFactory()

	// 1. Simple Request Logger Middleware
	log.Println("📝 Creating Simple Request Logger Middleware...")
	requestLoggerMiddleware := middlewareFactory.NewSimpleMiddleware(
		"request-logger",
		100, // priority
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				log.Printf("📥 [%s] %s %s - Middleware BEFORE processing",
					r.Header.Get("X-Trace-ID"), r.Method, r.URL.Path)

				next.ServeHTTP(w, r)

				duration := time.Since(start)
				log.Printf("📤 [%s] %s %s - Middleware AFTER processing (%v)",
					r.Header.Get("X-Trace-ID"), r.Method, r.URL.Path, duration)
			})
		},
	)

	// 2. API Authentication Middleware with Skip Logic
	log.Println("🔐 Creating API Authentication Middleware...")
	authMiddleware := middlewareFactory.NewConditionalMiddleware(
		"api-auth",
		200,
		func(path string) bool {
			// Skip authentication for health checks and public endpoints
			return strings.HasPrefix(path, "/health") ||
				strings.HasPrefix(path, "/public") ||
				path == "/favicon.ico"
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Printf("🔍 [%s] API Auth: Checking authentication for %s %s",
					r.Header.Get("X-Trace-ID"), r.Method, r.URL.Path)

				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					log.Printf("⚠️  [%s] Missing Authorization header", r.Header.Get("X-Trace-ID"))
				} else {
					log.Printf("✅ [%s] Authorization header present", r.Header.Get("X-Trace-ID"))
				}

				next.ServeHTTP(w, r)
			})
		},
	)

	// 3. Rate Limiting Middleware using Builder Pattern
	log.Println("⏱️ Creating Rate Limiting Middleware...")
	rateLimitMiddleware, err := middleware.NewCustomMiddlewareBuilder().
		WithName("rate-limiter").
		WithPriority(50). // High priority
		WithWrapFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Skip rate limiting for health checks
				if strings.HasPrefix(r.URL.Path, "/health") {
					next.ServeHTTP(w, r)
					return
				}

				clientIP := r.Header.Get("X-Real-IP")
				if clientIP == "" {
					clientIP = r.RemoteAddr
				}

				log.Printf("🚦 [%s] Rate Limit Check: %s from %s",
					r.Header.Get("X-Trace-ID"), r.URL.Path, clientIP)

				// Here you would check rate limits against a store (Redis, memory, etc.)
				// For demo purposes, just log
				log.Printf("✅ [%s] Rate limit OK for %s", r.Header.Get("X-Trace-ID"), clientIP)

				next.ServeHTTP(w, r)
			})
		}).
		Build()

	if err != nil {
		log.Printf("Error creating rate limit middleware: %v", err)
		return
	}

	// 4. Security Headers Middleware
	log.Println("🛡️ Creating Security Headers Middleware...")
	securityMiddleware := middlewareFactory.NewSimpleMiddleware(
		"security-headers",
		300,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Printf("🔒 [%s] Security: Adding security headers for %s %s",
					r.Header.Get("X-Trace-ID"), r.Method, r.URL.Path)

				// Add security headers to response
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("Strict-Transport-Security", "max-age=31536000")

				log.Printf("   Added security headers")

				next.ServeHTTP(w, r)
			})
		},
	)

	// 5. Performance Monitoring Middleware
	log.Println("📊 Creating Performance Monitoring Middleware...")
	performanceMiddleware, err := middleware.NewCustomMiddlewareBuilder().
		WithName("performance-monitor").
		WithPriority(400).
		WithWrapFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Monitor all requests except static files
				path := r.URL.Path
				if strings.HasSuffix(path, ".css") ||
					strings.HasSuffix(path, ".js") ||
					strings.HasSuffix(path, ".ico") ||
					strings.HasSuffix(path, ".png") ||
					strings.HasSuffix(path, ".jpg") {
					next.ServeHTTP(w, r)
					return
				}

				start := time.Now()
				log.Printf("⏱️  [%s] Performance: Starting timer for %s %s",
					r.Header.Get("X-Trace-ID"), r.Method, r.URL.Path)

				next.ServeHTTP(w, r)

				duration := time.Since(start)
				log.Printf("📈 [%s] Performance: %s %s completed in %v",
					r.Header.Get("X-Trace-ID"), r.Method, r.URL.Path, duration)

				// Performance alerts
				if duration > 1*time.Second {
					log.Printf("🐌 [%s] SLOW REQUEST ALERT: %v", r.Header.Get("X-Trace-ID"), duration)
				} else if duration > 500*time.Millisecond {
					log.Printf("⚠️  [%s] Performance Warning: %v", r.Header.Get("X-Trace-ID"), duration)
				}
			})
		}).
		Build()

	if err != nil {
		log.Printf("Error creating performance middleware: %v", err)
		return
	}

	// 6. Business Context Middleware with Complex Logic
	log.Println("💼 Creating Business Context Middleware...")
	businessContextMiddleware, err := middleware.NewCustomMiddlewareBuilder().
		WithName("business-context").
		WithPriority(150).
		WithWrapFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Only apply to business API endpoints
				if !strings.HasPrefix(r.URL.Path, "/api/") {
					next.ServeHTTP(w, r)
					return
				}

				log.Printf("💼 [%s] Business Context: Setting up for %s %s",
					r.Header.Get("X-Trace-ID"), r.Method, r.URL.Path)

				// Extract business context from headers/tokens
				userID := r.Header.Get("X-User-ID")
				tenantID := r.Header.Get("X-Tenant-ID")

				if userID != "" {
					log.Printf("   User ID: %s", userID)
				}
				if tenantID != "" {
					log.Printf("   Tenant ID: %s", tenantID)
				}

				// In real implementation, you would:
				// - Load user permissions
				// - Set up tenant context
				// - Initialize audit trail
				// - Prepare business rules

				next.ServeHTTP(w, r)

				log.Printf("💼 [%s] Business Context: Cleanup for %s %s",
					r.Header.Get("X-Trace-ID"), r.Method, r.URL.Path)
			})
		}).
		Build()

	if err != nil {
		log.Printf("Error creating business context middleware: %v", err)
		return
	}

	// 7. Demonstrate middleware execution with simulated requests
	log.Println("\n🧪 Testing middleware with simulated requests...")

	// Create a list of middleware in execution order (by priority)
	middlewares := []interfaces.Middleware{
		rateLimitMiddleware,       // Priority 50
		requestLoggerMiddleware,   // Priority 100
		businessContextMiddleware, // Priority 150
		authMiddleware,            // Priority 200
		securityMiddleware,        // Priority 300
		performanceMiddleware,     // Priority 400
	}

	// Create mock HTTP requests
	requests := []*http.Request{
		createMockRequest("GET", "/", "user-123", "tenant-456"),
		createMockRequest("GET", "/api/users", "user-123", "tenant-456"),
		createMockRequest("POST", "/api/users", "user-456", "tenant-789"),
		createMockRequest("GET", "/health", "", ""),
		createMockRequest("GET", "/public/info", "", ""),
		createMockRequest("GET", "/favicon.ico", "", ""),
	}

	// Simulate middleware execution for each request
	for i, req := range requests {
		log.Printf("\n--- Simulating Request %d: %s %s ---", i+1, req.Method, req.URL.Path)

		// Add trace ID for tracking
		req.Header.Set("X-Trace-ID", fmt.Sprintf("trace-%d", i+1))

		// Create a test handler chain
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate request processing
			time.Sleep(time.Duration(50+i*20) * time.Millisecond)
			w.WriteHeader(200)
			w.Write([]byte("OK"))
		})

		// Wrap handler with middleware chain
		for j := len(middlewares) - 1; j >= 0; j-- {
			mw := middlewares[j]
			handler = mw.Wrap(handler)
		}

		// Create mock response writer
		w := &MockResponseWriter{headers: make(http.Header)}

		// Execute the handler chain
		handler.ServeHTTP(w, req)
	}

	log.Println("\n✅ All custom middleware demonstrated successfully!")
	log.Println("📋 Summary of demonstrated features:")
	log.Println("   • Simple middleware creation with factory")
	log.Println("   • Conditional middleware with skip logic")
	log.Println("   • Builder pattern for complex middleware configuration")
	log.Println("   • Path and method filtering")
	log.Println("   • Priority-based execution ordering")
	log.Println("   • Standard http.Handler interface compatibility")
	log.Println("   • Type-safe middleware implementations")
	log.Println("   • Integration with business logic")
}

// MockResponseWriter implements http.ResponseWriter for testing
type MockResponseWriter struct {
	headers    http.Header
	statusCode int
	written    bool
}

func (w *MockResponseWriter) Header() http.Header {
	return w.headers
}

func (w *MockResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.statusCode = 200
		w.written = true
	}
	return len(data), nil
}

func (w *MockResponseWriter) WriteHeader(statusCode int) {
	if !w.written {
		w.statusCode = statusCode
		w.written = true
	}
}

func createMockRequest(method, path, userID, tenantID string) *http.Request {
	req, _ := http.NewRequest(method, "http://localhost:8080"+path, nil)
	req.Header.Set("User-Agent", "CustomMiddlewareExample/1.0")
	req.RemoteAddr = "192.168.1.100:45678"

	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	if tenantID != "" {
		req.Header.Set("X-Tenant-ID", tenantID)
	}
	if strings.HasPrefix(path, "/api/") && userID != "" {
		req.Header.Set("Authorization", "Bearer token-for-"+userID)
	}

	return req
}
