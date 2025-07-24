// Package main demonstrates all middleware working together.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/bulkhead"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/compression"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/cors"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/health"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/logging"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/ratelimit"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/retry"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/timeout"
)

func main() {
	fmt.Println("🚀 Starting Complete Middleware Demo Server")
	fmt.Println("=" + fmt.Sprintf("%*s", 50, "="))

	// Setup health checks
	healthRegistry := setupHealthChecks()
	healthHandler := health.NewHandler(healthRegistry)

	// Setup complete middleware chain
	middlewareChain := setupCompleteMiddleware()

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoints (no middleware)
	mux.Handle("/health", healthHandler.HealthHandler())
	mux.Handle("/health/live", healthHandler.LivenessHandler())
	mux.Handle("/health/ready", healthHandler.ReadinessHandler())

	// API endpoints with full middleware stack
	apiHandler := middlewareChain.Then(setupAPIRoutes())
	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))

	// Start server
	fmt.Println("\n📡 Server starting on :8080")
	printEndpoints()

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}

func setupHealthChecks() *health.Registry {
	registry := health.NewRegistry()

	// Simple ping check
	registry.Register("ping", func(ctx context.Context) interfaces.HealthCheckResult {
		return interfaces.HealthCheckResult{
			Status:  "healthy",
			Message: "Ping successful",
		}
	}, health.WithType(health.CheckTypeLiveness))

	// Database simulation check
	registry.Register("database", func(ctx context.Context) interfaces.HealthCheckResult {
		select {
		case <-time.After(50 * time.Millisecond):
			return interfaces.HealthCheckResult{
				Status:  "healthy",
				Message: "Database connection active",
				Metadata: map[string]interface{}{
					"connections": 5,
					"latency_ms":  50,
				},
			}
		case <-ctx.Done():
			return interfaces.HealthCheckResult{
				Status:  "unhealthy",
				Message: "Database check timeout",
			}
		}
	}, health.WithType(health.CheckTypeReadiness), health.WithCritical(true))

	// External service check
	registry.Register("external-api", health.URLCheck("https://httpbin.org/status/200"),
		health.WithType(health.CheckTypeLiveness),
		health.WithCritical(false))

	return registry
}

func setupCompleteMiddleware() *middleware.Chain {
	chain := middleware.NewChain()

	// 1. CORS - First, handle cross-origin requests
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowedOrigins = []string{"*"}
	corsConfig.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowedHeaders = []string{"Content-Type", "Authorization", "X-User-ID"}
	chain.Add(cors.NewMiddleware(corsConfig))

	// 2. Request/Response Logging - Log all requests
	loggingConfig := logging.DefaultConfig()
	loggingConfig.Logger = func(entry logging.LogEntry) {
		fmt.Printf("📝 [%s] %s %s -> %d (%v)\n",
			entry.Timestamp.Format("15:04:05"),
			entry.Method,
			entry.Path,
			entry.StatusCode,
			entry.Duration)
	}
	chain.Add(logging.NewMiddleware(loggingConfig))

	// 3. Timeout Management
	timeoutConfig := timeout.DefaultConfig()
	timeoutConfig.Timeout = 30 * time.Second
	chain.Add(timeout.NewMiddleware(timeoutConfig))

	// 4. Rate Limiting
	rateLimitConfig := ratelimit.DefaultConfig()
	rateLimitConfig.Limit = 20
	rateLimitConfig.Window = time.Minute
	rateLimitConfig.SkipPaths = []string{"/health"}
	chain.Add(ratelimit.NewMiddleware(rateLimitConfig))

	// 5. Bulkhead Pattern
	bulkheadConfig := bulkhead.DefaultConfig()
	bulkheadConfig.MaxConcurrent = 10
	bulkheadConfig.QueueSize = 20
	bulkheadConfig.ResourceKey = func(r *http.Request) string {
		switch {
		case r.URL.Path == "/api/heavy":
			return "heavy-operations"
		case r.URL.Path == "/api/users":
			return "user-service"
		default:
			return "default"
		}
	}
	chain.Add(bulkhead.NewMiddleware(bulkheadConfig))

	// 6. Retry Policies
	retryConfig := retry.DefaultConfig()
	retryConfig.MaxRetries = 2
	retryConfig.InitialDelay = 100 * time.Millisecond
	retryConfig.OnRetry = func(r *http.Request, attempt int, delay time.Duration) {
		fmt.Printf("🔄 Retrying %s (attempt %d) after %v\n", r.URL.Path, attempt, delay)
	}
	chain.Add(retry.NewMiddleware(retryConfig))

	// 7. Response Compression - Last, compress the final response
	compressionConfig := compression.DefaultConfig()
	compressionConfig.Level = 6
	compressionConfig.MinSize = 512
	chain.Add(compression.NewMiddleware(compressionConfig))

	return chain
}

func setupAPIRoutes() http.Handler {
	mux := http.NewServeMux()

	// Simple test endpoint
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":    "All middleware working!",
			"timestamp":  time.Now().Format(time.RFC3339),
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_agent": r.Header.Get("User-Agent"),
			"middleware_features": []string{
				"CORS enabled",
				"Request logged",
				"Timeout protected",
				"Rate limited",
				"Bulkhead isolated",
				"Retry enabled",
				"Response compressed",
			},
		}

		json.NewEncoder(w).Encode(response)
	})

	// Heavy operation endpoint (separate bulkhead)
	mux.HandleFunc("/heavy", func(w http.ResponseWriter, r *http.Request) {
		// Simulate heavy operation
		time.Sleep(2 * time.Second)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":   "Heavy operation completed",
			"duration":  "2 seconds",
			"bulkhead":  "heavy-operations",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		json.NewEncoder(w).Encode(response)
	})

	// Users endpoint (separate bulkhead)
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Generate some user data
		users := make([]map[string]interface{}, 0, 10)
		for i := 1; i <= 10; i++ {
			users = append(users, map[string]interface{}{
				"id":      i,
				"name":    fmt.Sprintf("User %d", i),
				"email":   fmt.Sprintf("user%d@example.com", i),
				"active":  i%2 == 0,
				"created": time.Now().Add(-time.Duration(i*24) * time.Hour).Format(time.RFC3339),
			})
		}

		response := map[string]interface{}{
			"users":     users,
			"total":     len(users),
			"bulkhead":  "user-service",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		json.NewEncoder(w).Encode(response)
	})

	// Error endpoint for testing retry
	errorCount := 0
	mux.HandleFunc("/flaky", func(w http.ResponseWriter, r *http.Request) {
		errorCount++

		w.Header().Set("Content-Type", "application/json")

		// Fail first 2 attempts, succeed on 3rd
		if errorCount%3 != 0 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Service temporarily unavailable",
				"attempt": errorCount,
				"message": "This endpoint fails 2 out of 3 times to demo retry",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   "Success after retry!",
			"attempt":   errorCount,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Large response for compression testing
	mux.HandleFunc("/large", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Generate large dataset
		data := make([]map[string]interface{}, 0, 1000)
		for i := 0; i < 1000; i++ {
			data = append(data, map[string]interface{}{
				"id":          i,
				"title":       fmt.Sprintf("Item %d", i),
				"description": fmt.Sprintf("This is a detailed description for item %d. It contains repetitive text that compresses well.", i),
				"category":    fmt.Sprintf("Category %d", i%10),
				"tags":        []string{"tag1", "tag2", "tag3"},
				"metadata": map[string]interface{}{
					"created": time.Now().Add(-time.Duration(i) * time.Minute).Format(time.RFC3339),
					"active":  i%2 == 0,
				},
			})
		}

		response := map[string]interface{}{
			"message":   "Large dataset (compression demo)",
			"count":     len(data),
			"data":      data,
			"size_info": "This response is large and will be compressed",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		json.NewEncoder(w).Encode(response)
	})

	return mux
}

func printEndpoints() {
	fmt.Println("\n🌐 Available Endpoints:")
	fmt.Println("──────────────────────")

	fmt.Println("\n🏥 Health Checks:")
	fmt.Println("  GET /health         - Overall health status")
	fmt.Println("  GET /health/live    - Liveness probe")
	fmt.Println("  GET /health/ready   - Readiness probe")

	fmt.Println("\n🔧 API Endpoints (Full Middleware Stack):")
	fmt.Println("  GET /api/test       - Simple test endpoint")
	fmt.Println("  GET /api/heavy      - Heavy operation (2s delay)")
	fmt.Println("  GET /api/users      - User list (bulkhead: user-service)")
	fmt.Println("  GET /api/flaky      - Flaky endpoint (demos retry)")
	fmt.Println("  GET /api/large      - Large response (demos compression)")

	fmt.Println("\n🧪 Test Commands:")
	fmt.Println("──────────────────")
	fmt.Println("  # Basic test")
	fmt.Println("  curl http://localhost:8080/api/test")
	fmt.Println("")
	fmt.Println("  # Test CORS")
	fmt.Println("  curl -H 'Origin: http://localhost:3000' http://localhost:8080/api/test")
	fmt.Println("")
	fmt.Println("  # Test compression")
	fmt.Println("  curl -H 'Accept-Encoding: gzip' http://localhost:8080/api/large")
	fmt.Println("")
	fmt.Println("  # Test rate limiting (run multiple times quickly)")
	fmt.Println("  for i in {1..25}; do curl http://localhost:8080/api/test; done")
	fmt.Println("")
	fmt.Println("  # Test retry (may need multiple attempts)")
	fmt.Println("  curl http://localhost:8080/api/flaky")
	fmt.Println("")
	fmt.Println("  # Test health checks")
	fmt.Println("  curl http://localhost:8080/health")

	fmt.Println("\n🔍 Middleware Features Active:")
	fmt.Println("──────────────────────────────")
	fmt.Println("  ✅ CORS - Cross-origin requests allowed")
	fmt.Println("  ✅ Logging - All requests logged to console")
	fmt.Println("  ✅ Timeout - 30 second request timeout")
	fmt.Println("  ✅ Rate Limiting - 20 requests per minute")
	fmt.Println("  ✅ Bulkhead - Resource isolation by endpoint")
	fmt.Println("  ✅ Retry - Up to 2 retries for failed requests")
	fmt.Println("  ✅ Compression - Gzip/deflate compression")
	fmt.Println("  ✅ Health Checks - Liveness and readiness probes")
}
