// Package main demonstrates enhanced middleware patterns with custom implementations.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
)

func main() {
	log.Println("🎯 Enhanced Middleware Patterns Example")
	log.Println("======================================")

	setupEnhancedMiddlewareServer()
}

func setupEnhancedMiddlewareServer() {
	// Create custom middleware factory
	customFactory := middleware.NewCustomMiddlewareFactory()

	// 1. Request Context Middleware (highest priority)
	contextMiddleware := customFactory.NewSimpleMiddleware(
		"request-context",
		50,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				traceID := generateTraceID()
				start := time.Now()

				// Add context headers
				r.Header.Set("X-Trace-ID", traceID)
				r.Header.Set("X-Start-Time", start.Format(time.RFC3339Nano))

				log.Printf("🆔 [%s] Request context established for %s %s",
					traceID, r.Method, r.URL.Path)

				next.ServeHTTP(w, r)
			})
		},
	)

	// 2. Enhanced CORS Middleware
	corsMiddleware := customFactory.NewConditionalMiddleware(
		"enhanced-cors",
		100,
		func(path string) bool {
			// Skip CORS for health and internal endpoints
			return strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/internal")
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				traceID := r.Header.Get("X-Trace-ID")
				origin := r.Header.Get("Origin")

				// Define allowed origins
				allowedOrigins := map[string]bool{
					"http://localhost:3000":     true, // React dev
					"http://localhost:8000":     true, // Vue dev
					"http://localhost:4200":     true, // Angular dev
					"https://app.example.com":   true, // Production app
					"https://admin.example.com": true, // Admin app
				}

				// Handle CORS
				if origin != "" {
					if allowedOrigins[origin] {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Access-Control-Allow-Credentials", "true")
						log.Printf("✅ [%s] CORS allowed for origin: %s", traceID, origin)
					} else {
						log.Printf("❌ [%s] CORS blocked for origin: %s", traceID, origin)
					}

					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers",
						"Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, X-User-ID, X-Tenant-ID")
					w.Header().Set("Access-Control-Expose-Headers",
						"X-Total-Count, X-Rate-Limit-Remaining, X-Request-ID")
					w.Header().Set("Access-Control-Max-Age", "86400")
				}

				// Handle preflight
				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}

				next.ServeHTTP(w, r)
			})
		},
	)

	// 3. Security Headers Middleware
	securityMiddleware := customFactory.NewSimpleMiddleware(
		"security-headers",
		150,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				traceID := r.Header.Get("X-Trace-ID")

				// Add comprehensive security headers
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
				w.Header().Set("Content-Security-Policy", "default-src 'self'")
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
				w.Header().Set("X-Request-ID", traceID)

				log.Printf("🛡️  [%s] Security headers applied", traceID)
				next.ServeHTTP(w, r)
			})
		},
	)

	// 4. Rate Limiting Middleware with Enhanced Logic
	rateLimitMiddleware, err := middleware.NewCustomMiddlewareBuilder().
		WithName("enhanced-rate-limit").
		WithPriority(200).
		WithWrapFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				traceID := r.Header.Get("X-Trace-ID")

				// Skip rate limiting for certain paths
				if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
					next.ServeHTTP(w, r)
					return
				}

				clientIP := getClientIP(r)
				userID := r.Header.Get("X-User-ID")

				// Different rate limits for different endpoints
				var limit, remaining int
				switch {
				case strings.HasPrefix(r.URL.Path, "/api/auth"):
					limit, remaining = 5, 4 // Strict for auth endpoints
				case strings.HasPrefix(r.URL.Path, "/api/"):
					limit, remaining = 100, 95 // Standard for API
				default:
					limit, remaining = 200, 195 // Generous for others
				}

				log.Printf("⏱️  [%s] Rate limit check - IP: %s, User: %s, Limit: %d",
					traceID, clientIP, userID, limit)

				// Add rate limit headers
				w.Header().Set("X-Rate-Limit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-Rate-Limit-Remaining", fmt.Sprintf("%d", remaining))
				w.Header().Set("X-Rate-Limit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))

				// Simulate rate limit check (in real implementation, use Redis/memory store)
				if remaining <= 0 {
					log.Printf("🚫 [%s] Rate limit exceeded for %s", traceID, clientIP)
					w.WriteHeader(http.StatusTooManyRequests)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":       "rate limit exceeded",
						"retry_after": 3600,
					})
					return
				}

				log.Printf("✅ [%s] Rate limit OK - %d/%d remaining", traceID, remaining, limit)
				next.ServeHTTP(w, r)
			})
		}).
		Build()

	if err != nil {
		log.Fatalf("❌ Error creating rate limit middleware: %v", err)
	}

	// 5. Request/Response Logging Middleware
	loggingMiddleware := customFactory.NewSimpleMiddleware(
		"enhanced-logging",
		250,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				traceID := r.Header.Get("X-Trace-ID")
				start := time.Now()

				// Log request details
				log.Printf("📥 [%s] REQUEST: %s %s from %s (UA: %s)",
					traceID, r.Method, r.URL.Path, getClientIP(r),
					r.Header.Get("User-Agent"))

				// Wrap response writer to capture status code
				ww := &responseWriter{ResponseWriter: w, statusCode: 200}

				next.ServeHTTP(ww, r)

				duration := time.Since(start)

				// Log response details
				log.Printf("📤 [%s] RESPONSE: %d (%v) - %s %s",
					traceID, ww.statusCode, duration, r.Method, r.URL.Path)

				// Performance warnings
				if duration > 1*time.Second {
					log.Printf("🐌 [%s] SLOW REQUEST WARNING: %v", traceID, duration)
				}
			})
		},
	)

	// 6. Business Logic Middleware
	businessMiddleware := customFactory.NewConditionalMiddleware(
		"business-logic",
		300,
		func(path string) bool {
			// Only apply to API endpoints
			return !strings.HasPrefix(path, "/api/")
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				traceID := r.Header.Get("X-Trace-ID")
				userID := r.Header.Get("X-User-ID")
				tenantID := r.Header.Get("X-Tenant-ID")

				log.Printf("💼 [%s] Business context - User: %s, Tenant: %s",
					traceID, userID, tenantID)

				// Add business context headers
				w.Header().Set("X-Service-Version", "2.1.0")
				w.Header().Set("X-Environment", "production")

				// In real implementation:
				// - Load user permissions
				// - Set tenant context
				// - Initialize audit trail
				// - Apply business rules

				next.ServeHTTP(w, r)
			})
		},
	)

	// 7. Error Recovery Middleware (lowest priority)
	recoveryMiddleware := customFactory.NewSimpleMiddleware(
		"panic-recovery",
		1000,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if err := recover(); err != nil {
						traceID := r.Header.Get("X-Trace-ID")
						log.Printf("💥 [%s] PANIC RECOVERED: %v", traceID, err)

						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(map[string]interface{}{
							"error":      "internal server error",
							"request_id": traceID,
						})
					}
				}()

				next.ServeHTTP(w, r)
			})
		},
	)

	// 8. Build middleware chain
	middlewares := []interfaces.Middleware{
		contextMiddleware,   // 50
		corsMiddleware,      // 100
		securityMiddleware,  // 150
		rateLimitMiddleware, // 200
		loggingMiddleware,   // 250
		businessMiddleware,  // 300
		recoveryMiddleware,  // 1000
	}

	// Create handler chain
	var handler http.Handler = &EnhancedAPIHandler{}
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].Wrap(handler)
	}

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("🚀 Enhanced middleware server starting on :8080")
	log.Println("\n📋 Available endpoints:")
	log.Println("   • GET  /api/users        - Get users (rate limited)")
	log.Println("   • POST /api/users        - Create user (business logic)")
	log.Println("   • GET  /api/auth/login   - Login (strict rate limit)")
	log.Println("   • GET  /api/posts        - Get posts")
	log.Println("   • GET  /health           - Health check (no middleware)")
	log.Println("   • GET  /metrics          - Metrics (no rate limit)")
	log.Println("   • GET  /panic            - Test panic recovery")

	log.Println("\n💡 Test the middleware stack:")
	log.Println("   # Standard API request")
	log.Println("   curl -H 'X-User-ID: user123' -H 'X-Tenant-ID: tenant456' \\")
	log.Println("        http://localhost:8080/api/users -v")

	log.Println("\n   # CORS preflight")
	log.Println("   curl -X OPTIONS -H 'Origin: http://localhost:3000' \\")
	log.Println("        -H 'Access-Control-Request-Method: POST' \\")
	log.Println("        http://localhost:8080/api/users -v")

	log.Println("\n   # Test panic recovery")
	log.Println("   curl http://localhost:8080/panic -v")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// EnhancedAPIHandler demonstrates various middleware effects
type EnhancedAPIHandler struct{}

func (h *EnhancedAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/health":
		h.handleHealth(w, r)
	case "/metrics":
		h.handleMetrics(w, r)
	case "/api/users":
		h.handleUsers(w, r)
	case "/api/posts":
		h.handlePosts(w, r)
	case "/api/auth/login":
		h.handleLogin(w, r)
	case "/panic":
		h.handlePanic(w, r)
	default:
		h.handleNotFound(w, r)
	}
}

func (h *EnhancedAPIHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("🏥 [%s] Health check", traceID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now().Format(time.RFC3339),
		"request_id": traceID,
		"middleware": "minimal",
	})
}

func (h *EnhancedAPIHandler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("📊 [%s] Metrics request", traceID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests_total": 12345,
		"errors_total":   23,
		"latency_avg_ms": 45.6,
		"uptime_seconds": 86400,
		"request_id":     traceID,
	})
}

func (h *EnhancedAPIHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	userID := r.Header.Get("X-User-ID")
	tenantID := r.Header.Get("X-Tenant-ID")

	log.Printf("👥 [%s] Users API - User: %s, Tenant: %s", traceID, userID, tenantID)

	// Simulate some processing time
	time.Sleep(50 * time.Millisecond)

	users := []map[string]interface{}{
		{"id": 1, "name": "Alice", "tenant": tenantID, "active": true},
		{"id": 2, "name": "Bob", "tenant": tenantID, "active": true},
		{"id": 3, "name": "Carol", "tenant": tenantID, "active": false},
	}

	w.Header().Set("X-Total-Count", "3")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":   users,
		"total":   3,
		"user_id": userID,
		"tenant":  tenantID,
	})
}

func (h *EnhancedAPIHandler) handlePosts(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("📝 [%s] Posts API", traceID)

	posts := []map[string]interface{}{
		{"id": 1, "title": "Middleware Patterns", "author": "Alice"},
		{"id": 2, "title": "CORS Best Practices", "author": "Bob"},
		{"id": 3, "title": "Rate Limiting Strategies", "author": "Carol"},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts": posts,
		"total": 3,
	})
}

func (h *EnhancedAPIHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("🔐 [%s] Auth login request", traceID)

	// Simulate login processing
	time.Sleep(100 * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      "jwt-token-here",
		"expires_in": 3600,
		"user_id":    "user123",
	})
}

func (h *EnhancedAPIHandler) handlePanic(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("💥 [%s] Panic endpoint called - this will trigger recovery", traceID)

	// This will trigger the panic recovery middleware
	panic("simulated panic for testing recovery middleware")
}

func (h *EnhancedAPIHandler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("❌ [%s] Not found: %s", traceID, r.URL.Path)

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "endpoint not found",
		"path":  r.URL.Path,
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func generateTraceID() string {
	return fmt.Sprintf("enh-%d", time.Now().UnixNano()%1000000)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
