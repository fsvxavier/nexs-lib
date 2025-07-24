package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/nethttp"
)

func main() {
	log.Println("🔧 HTTPServer Providers with Custom Hooks & Middleware")
	log.Println("=====================================================")

	// Create and start server with integrated custom functionality
	setupProviderServer()
}

func setupProviderServer() {
	// 1. Setup Custom Hooks
	log.Println("🎣 Setting up custom hooks...")
	hookFactory := hooks.NewCustomHookFactory()

	// Request lifecycle hook
	lifecycleHook := hookFactory.NewSimpleHook(
		"request-lifecycle",
		[]interfaces.HookEvent{
			interfaces.HookEventRequestStart,
			interfaces.HookEventRequestEnd,
		},
		100,
		func(ctx *interfaces.HookContext) error {
			if ctx.Event == interfaces.HookEventRequestStart {
				log.Printf("🎣 [%s] Request started: %s %s",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path)
			} else {
				log.Printf("🎣 [%s] Request completed: %s %s (%d) in %v",
					ctx.TraceID, ctx.Request.Method, ctx.Request.URL.Path,
					ctx.StatusCode, ctx.Duration)
			}
			return nil
		},
	)

	// Performance monitoring hook
	perfHook := hookFactory.NewFilteredHook(
		"performance-monitor",
		[]interfaces.HookEvent{interfaces.HookEventRequestEnd},
		200,
		func(path string) bool {
			// Monitor all paths except health
			return path != "/health"
		},
		func(method string) bool {
			// Monitor all methods
			return true
		},
		func(ctx *interfaces.HookContext) error {
			if ctx.Duration > 500*time.Millisecond {
				log.Printf("🎣 [%s] ⚠️  Slow request detected: %v for %s %s",
					ctx.TraceID, ctx.Duration, ctx.Request.Method, ctx.Request.URL.Path)
			}
			return nil
		},
	)

	// 2. Setup Custom Middleware
	log.Println("🔧 Setting up custom middleware...")
	middlewareFactory := middleware.NewCustomMiddlewareFactory()

	// Request tracking middleware
	trackingMiddleware := middlewareFactory.NewSimpleMiddleware(
		"request-tracking",
		100,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				traceID := generateTraceID()
				r.Header.Set("X-Trace-ID", traceID)

				log.Printf("🔧 [%s] Middleware: Processing %s %s",
					traceID, r.Method, r.URL.Path)

				next.ServeHTTP(w, r)
			})
		},
	)

	// CORS middleware
	corsMiddleware := middlewareFactory.NewConditionalMiddleware(
		"cors",
		200,
		func(path string) bool {
			// Skip CORS for health endpoint
			return path == "/health"
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				origin := r.Header.Get("Origin")
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				}

				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}

				next.ServeHTTP(w, r)
			})
		},
	)

	// 3. Create NetHTTP Provider Server (you can switch to other providers)
	log.Println("🌐 Creating NetHTTP server with hooks and middleware...")

	config := config.DefaultConfig().
		WithHost("localhost").
		WithPort(8080).
		WithReadTimeout(30 * time.Second).
		WithWriteTimeout(30 * time.Second).
		WithIdleTimeout(60 * time.Second)

	// Create server
	server, err := nethttp.NewServer(config)
	if err != nil {
		log.Fatalf("❌ Failed to create server: %v", err)
	}

	// 4. Setup routes with integrated handler
	integratedHandler := &IntegratedHandler{
		hooks: []interfaces.Hook{lifecycleHook, perfHook},
	}

	// 5. Apply middleware to handler
	var handler http.Handler = integratedHandler
	middlewares := []interfaces.Middleware{trackingMiddleware, corsMiddleware}

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].Wrap(handler)
	}

	// 6. Set handler on server
	if nethttpServer, ok := server.(*nethttp.Server); ok {
		nethttpServer.SetHandler(handler)
	}

	// 7. Setup routes (if server supports routing)
	// Note: NetHTTP uses a single handler, so we handle routing in the handler itself

	log.Println("✅ Server setup complete!")
	log.Println("\n📋 Available endpoints:")
	log.Println("   • GET  /api/users        - List users")
	log.Println("   • POST /api/users        - Create user")
	log.Println("   • GET  /api/posts        - List posts")
	log.Println("   • GET  /health           - Health check")
	log.Println("   • GET  /                 - Welcome message")

	log.Println("\n💡 Test the integrated system:")
	log.Println("   curl http://localhost:8080/api/users")
	log.Println("   curl -X POST http://localhost:8080/api/users -d '{\"name\":\"John\"}' -H 'Content-Type: application/json'")
	log.Println("   curl http://localhost:8080/health")

	log.Println("\n🚀 Starting server on http://localhost:8080")

	// 8. Start server
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// IntegratedHandler demonstrates hooks integration with HTTP handlers
type IntegratedHandler struct {
	hooks []interfaces.Hook
}

func (h *IntegratedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	start := time.Now()

	// Create hook context
	hookCtx := &interfaces.HookContext{
		Event:         interfaces.HookEventRequestStart,
		ServerName:    "provider-server",
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
				log.Printf("❌ Pre-request hook error: %v", err)
			}
		}
	}

	// Handle the request and capture status code
	statusCode := h.handleRequest(w, r)
	duration := time.Since(start)

	// Execute post-request hooks
	hookCtx.Event = interfaces.HookEventRequestEnd
	hookCtx.StatusCode = statusCode
	hookCtx.Duration = duration

	for _, hook := range h.hooks {
		if hook.ShouldExecute(hookCtx) {
			if err := hook.Execute(hookCtx); err != nil {
				log.Printf("❌ Post-request hook error: %v", err)
			}
		}
	}
}

func (h *IntegratedHandler) handleRequest(w http.ResponseWriter, r *http.Request) int {
	traceID := r.Header.Get("X-Trace-ID")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", traceID)

	log.Printf("🎯 [%s] Handling %s %s", traceID, r.Method, r.URL.Path)

	// Route handling
	switch r.URL.Path {
	case "/":
		return h.handleWelcome(w, r)
	case "/health":
		return h.handleHealth(w, r)
	case "/api/users":
		return h.handleUsers(w, r)
	case "/api/posts":
		return h.handlePosts(w, r)
	default:
		return h.handleNotFound(w, r)
	}
}

func (h *IntegratedHandler) handleWelcome(w http.ResponseWriter, r *http.Request) int {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("🏠 [%s] Welcome page", traceID)

	response := map[string]interface{}{
		"message":    "Welcome to HTTPServer with Custom Hooks & Middleware!",
		"version":    "2.0.0",
		"provider":   "nethttp",
		"features":   []string{"custom-hooks", "custom-middleware", "integrated-system"},
		"request_id": traceID,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return http.StatusOK
}

func (h *IntegratedHandler) handleHealth(w http.ResponseWriter, r *http.Request) int {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("🏥 [%s] Health check", traceID)

	// Simulate health check processing
	time.Sleep(10 * time.Millisecond)

	response := map[string]interface{}{
		"status":     "healthy",
		"uptime":     "1h 30m",
		"hooks":      "enabled",
		"middleware": "enabled",
		"request_id": traceID,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return http.StatusOK
}

func (h *IntegratedHandler) handleUsers(w http.ResponseWriter, r *http.Request) int {
	traceID := r.Header.Get("X-Trace-ID")

	switch r.Method {
	case "GET":
		log.Printf("👥 [%s] Listing users", traceID)

		// Simulate database query
		time.Sleep(100 * time.Millisecond)

		users := []map[string]interface{}{
			{"id": 1, "name": "Alice Johnson", "email": "alice@example.com", "active": true},
			{"id": 2, "name": "Bob Smith", "email": "bob@example.com", "active": true},
			{"id": 3, "name": "Carol Brown", "email": "carol@example.com", "active": false},
		}

		response := map[string]interface{}{
			"users":      users,
			"total":      len(users),
			"page":       1,
			"page_size":  10,
			"request_id": traceID,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return http.StatusOK

	case "POST":
		log.Printf("👥 [%s] Creating user", traceID)

		var newUser map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			log.Printf("❌ [%s] Invalid JSON: %v", traceID, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":      "invalid JSON",
				"request_id": traceID,
			})
			return http.StatusBadRequest
		}

		// Simulate user creation
		time.Sleep(150 * time.Millisecond)

		newUser["id"] = 4
		newUser["active"] = true
		newUser["created_at"] = time.Now().Format(time.RFC3339)
		newUser["request_id"] = traceID

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newUser)
		return http.StatusCreated

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error":      "method not allowed",
			"request_id": traceID,
		})
		return http.StatusMethodNotAllowed
	}
}

func (h *IntegratedHandler) handlePosts(w http.ResponseWriter, r *http.Request) int {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("📝 [%s] Listing posts", traceID)

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error":      "only GET allowed",
			"request_id": traceID,
		})
		return http.StatusMethodNotAllowed
	}

	// Simulate database query with some delay
	time.Sleep(200 * time.Millisecond)

	posts := []map[string]interface{}{
		{"id": 1, "title": "Getting Started with HTTPServer", "author": "Alice", "published": true},
		{"id": 2, "title": "Advanced Hooks and Middleware", "author": "Bob", "published": true},
		{"id": 3, "title": "Provider Integration Patterns", "author": "Carol", "published": false},
	}

	response := map[string]interface{}{
		"posts":      posts,
		"total":      len(posts),
		"request_id": traceID,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return http.StatusOK
}

func (h *IntegratedHandler) handleNotFound(w http.ResponseWriter, r *http.Request) int {
	traceID := r.Header.Get("X-Trace-ID")
	log.Printf("❌ [%s] Not found: %s", traceID, r.URL.Path)

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error":      "endpoint not found",
		"path":       r.URL.Path,
		"request_id": traceID,
	})
	return http.StatusNotFound
}

func generateTraceID() string {
	return fmt.Sprintf("prov-%d", time.Now().UnixNano()%1000000)
}
