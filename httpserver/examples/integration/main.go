package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
)

func main() {
	log.Println("🚀 Complete Integration Example: Custom Hooks + Middleware")
	log.Println("=========================================================")

	// Setup integrated system
	server := setupIntegratedServer()

	// Start server
	go server.start()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🌐 Server running on http://localhost:8080")
	log.Println("📋 Available endpoints:")
	log.Println("   • GET  /api/users        - List users")
	log.Println("   • POST /api/users        - Create user")
	log.Println("   • GET  /api/users/123    - Get user")
	log.Println("   • GET  /health           - Health check")
	log.Println("   • GET  /public/info      - Public info")
	log.Println("\n💡 Try making requests to see hooks and middleware in action!")
	log.Println("   curl http://localhost:8080/api/users")
	log.Println("   curl -X POST http://localhost:8080/api/users -H 'Content-Type: application/json'")
	log.Println("\nPress Ctrl+C to stop the server...")

	<-sigChan
	log.Println("\n🛑 Shutdown signal received, stopping server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.shutdown(ctx); err != nil {
		log.Printf("❌ Server shutdown error: %v", err)
	} else {
		log.Println("✅ Server shutdown completed")
	}
}

type IntegratedServer struct {
	httpServer   *http.Server
	hookRegistry interfaces.HookRegistry
}

func setupIntegratedServer() *IntegratedServer {
	// 1. Setup Custom Hooks
	log.Println("🎣 Setting up custom hooks...")
	hookFactory := hooks.NewCustomHookFactory()

	// Request Logger Hook
	requestLoggerHook := hookFactory.NewSimpleHook(
		"request-logger",
		[]interfaces.HookEvent{
			interfaces.HookEventRequestStart,
			interfaces.HookEventRequestEnd,
		},
		100,
		func(ctx *interfaces.HookContext) error {
			if ctx.Event == interfaces.HookEventRequestStart {
				log.Printf("🎣 [HOOK] [%s] %s %s - Request started",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path)
			} else if ctx.Event == interfaces.HookEventRequestEnd {
				log.Printf("🎣 [HOOK] [%s] %s %s - %d (%v) - Request completed",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path,
					ctx.StatusCode, ctx.Duration)
			}
			return nil
		},
	)

	// API Security Hook
	apiSecurityHook := hookFactory.NewConditionalHook(
		"api-security",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		200,
		func(ctx *interfaces.HookContext) bool {
			return strings.HasPrefix(ctx.Request.URL.Path, "/api/")
		},
		func(ctx *interfaces.HookContext) error {
			log.Printf("🎣 [HOOK] [%s] API Security check for %s",
				ctx.TraceID, ctx.Request.URL.Path)

			authHeader := ctx.Request.Header.Get("Authorization")
			if authHeader == "" && ctx.Request.Method != "GET" {
				log.Printf("🎣 [HOOK] [%s] ⚠️  Unauthenticated %s request",
					ctx.TraceID, ctx.Request.Method)
			}
			return nil
		},
	)

	// Performance Monitor Hook (Async)
	performanceHook := hookFactory.NewAsyncHook(
		"performance-monitor",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		300,
		10,
		5*time.Second,
		func(ctx *interfaces.HookContext) error {
			if ctx.Duration > 100*time.Millisecond {
				log.Printf("🎣 [HOOK] [%s] ⚠️  Slow request detected: %v",
					ctx.TraceID, ctx.Duration)
			}
			return nil
		},
	)

	// 2. Setup Custom Middleware
	log.Println("🔧 Setting up custom middleware...")
	middlewareFactory := middleware.NewCustomMiddlewareFactory()

	// Request Logger Middleware
	requestLoggerMiddleware := middlewareFactory.NewSimpleMiddleware(
		"request-logger",
		100,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				traceID := generateTraceID()
				r.Header.Set("X-Trace-ID", traceID)

				log.Printf("🔧 [MIDDLEWARE] [%s] %s %s - Middleware processing",
					traceID, r.Method, r.URL.Path)

				start := time.Now()
				next.ServeHTTP(w, r)
				duration := time.Since(start)

				log.Printf("🔧 [MIDDLEWARE] [%s] %s %s - Middleware completed (%v)",
					traceID, r.Method, r.URL.Path, duration)
			})
		},
	)

	// Security Headers Middleware
	securityMiddleware := middlewareFactory.NewSimpleMiddleware(
		"security-headers",
		200,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Printf("🔧 [MIDDLEWARE] [%s] Adding security headers",
					r.Header.Get("X-Trace-ID"))

				// Add security headers
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("Strict-Transport-Security", "max-age=31536000")

				next.ServeHTTP(w, r)
			})
		},
	)

	// CORS Middleware
	corsMiddleware := middlewareFactory.NewConditionalMiddleware(
		"cors",
		300,
		func(path string) bool {
			// Skip CORS for health checks
			return strings.HasPrefix(path, "/health")
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Printf("🔧 [MIDDLEWARE] [%s] Adding CORS headers",
					r.Header.Get("X-Trace-ID"))

				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}

				next.ServeHTTP(w, r)
			})
		},
	)

	// 3. Create an integrated handler that triggers hooks
	integrationHandler := &IntegrationHandler{
		hooks: []interfaces.Hook{
			requestLoggerHook,
			apiSecurityHook,
			performanceHook,
		},
	}

	// 4. Setup middleware chain
	middlewares := []interfaces.Middleware{
		requestLoggerMiddleware,
		securityMiddleware,
		corsMiddleware,
	}

	// Build handler chain
	var handler http.Handler = integrationHandler
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].Wrap(handler)
	}

	// 5. Create HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &IntegratedServer{
		httpServer: server,
	}
}

type IntegrationHandler struct {
	hooks []interfaces.Hook
}

func (h *IntegrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	if traceID == "" {
		traceID = generateTraceID()
	}

	start := time.Now()

	// Create hook context
	hookCtx := &interfaces.HookContext{
		Event:         interfaces.HookEventRequestStart,
		ServerName:    "integration-server",
		Timestamp:     start,
		Request:       r,
		Response:      w,
		TraceID:       traceID,
		CorrelationID: "corr-" + traceID,
	}

	// Execute pre-request hooks
	for _, hook := range h.hooks {
		if hook.ShouldExecute(hookCtx) {
			if err := hook.Execute(hookCtx); err != nil {
				log.Printf("❌ Hook error: %v", err)
			}
		}
	}

	// Handle the actual request
	statusCode := h.handleRequest(w, r)
	duration := time.Since(start)

	// Execute post-request hooks
	hookCtx.Event = interfaces.HookEventRequestEnd
	hookCtx.StatusCode = statusCode
	hookCtx.Duration = duration

	for _, hook := range h.hooks {
		if hook.ShouldExecute(hookCtx) {
			// Execute async hooks asynchronously
			if asyncHook, ok := hook.(interfaces.AsyncHook); ok {
				go func(ah interfaces.AsyncHook, ctx *interfaces.HookContext) {
					errChan := ah.ExecuteAsync(ctx)
					if err := <-errChan; err != nil {
						log.Printf("❌ Async hook error: %v", err)
					}
				}(asyncHook, hookCtx)
			} else {
				if err := hook.Execute(hookCtx); err != nil {
					log.Printf("❌ Hook error: %v", err)
				}
			}
		}
	}
}

func (h *IntegrationHandler) handleRequest(w http.ResponseWriter, r *http.Request) int {
	traceID := r.Header.Get("X-Trace-ID")

	log.Printf("🎯 [HANDLER] [%s] Processing %s %s", traceID, r.Method, r.URL.Path)

	// Route handling
	switch {
	case r.URL.Path == "/health":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
		return http.StatusOK

	case r.URL.Path == "/public/info":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"integration-example","version":"1.0.0"}`))
		return http.StatusOK

	case strings.HasPrefix(r.URL.Path, "/api/users"):
		return h.handleUsersAPI(w, r)

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"endpoint not found"}`))
		return http.StatusNotFound
	}
}

func (h *IntegrationHandler) handleUsersAPI(w http.ResponseWriter, r *http.Request) int {
	traceID := r.Header.Get("X-Trace-ID")

	// Simulate some processing time
	time.Sleep(50 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		if r.URL.Path == "/api/users" {
			log.Printf("🎯 [HANDLER] [%s] Listing users", traceID)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}],"total":2}`))
			return http.StatusOK
		} else {
			// GET /api/users/123
			log.Printf("🎯 [HANDLER] [%s] Getting user %s", traceID, strings.TrimPrefix(r.URL.Path, "/api/users/"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":123,"name":"John Doe","email":"john@example.com"}`))
			return http.StatusOK
		}

	case "POST":
		auth := r.Header.Get("Authorization")
		if auth == "" {
			log.Printf("🎯 [HANDLER] [%s] Unauthorized POST request", traceID)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"authorization required"}`))
			return http.StatusUnauthorized
		}

		log.Printf("🎯 [HANDLER] [%s] Creating new user", traceID)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":456,"name":"New User","created":true}`))
		return http.StatusCreated

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"method not allowed"}`))
		return http.StatusMethodNotAllowed
	}
}

func (s *IntegratedServer) start() {
	log.Printf("🚀 Starting integrated server on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ Server error: %v", err)
	}
}

func (s *IntegratedServer) shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func generateTraceID() string {
	return fmt.Sprintf("trace-%d", time.Now().UnixNano()%1000000)
}
