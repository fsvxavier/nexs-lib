// Package main demonstrates advanced middleware usage with all features.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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
	// Setup health checks
	healthRegistry := setupHealthChecks()
	healthHandler := health.NewHandler(healthRegistry)

	// Setup middleware chain
	middlewareChain := setupMiddleware()

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoints
	mux.Handle("/health", healthHandler.HealthHandler())
	mux.Handle("/health/live", healthHandler.LivenessHandler())
	mux.Handle("/health/ready", healthHandler.ReadinessHandler())

	// API endpoints with middleware
	apiHandler := middlewareChain.Then(setupAPIRoutes())
	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))

	// Static files without heavy middleware
	staticChain := middleware.NewChain()
	staticChain.Add(cors.NewMiddleware(cors.DefaultConfig()))
	staticChain.Add(compression.NewMiddleware(compression.DefaultConfig()))
	staticHandler := staticChain.Then(http.FileServer(http.Dir("./static/")))
	mux.Handle("/static/", http.StripPrefix("/static", staticHandler))

	fmt.Println("Server starting on :8080")
	fmt.Println("Health checks available at:")
	fmt.Println("  http://localhost:8080/health")
	fmt.Println("  http://localhost:8080/health/live")
	fmt.Println("  http://localhost:8080/health/ready")
	fmt.Println("API endpoints:")
	fmt.Println("  http://localhost:8080/api/users")
	fmt.Println("  http://localhost:8080/api/orders")
	fmt.Println("  http://localhost:8080/api/heavy-task")

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

// setupHealthChecks configures health check registry.
func setupHealthChecks() *health.Registry {
	registry := health.NewRegistry()

	// Database health check (simulated)
	registry.Register("database", health.DatabaseCheck(func(ctx context.Context) error {
		// Simulate database ping
		time.Sleep(10 * time.Millisecond)
		return nil // Always healthy for demo
	}), health.WithType(health.CheckTypeReadiness))

	// External API health check
	registry.Register("external-api", health.URLCheck("https://httpbin.org/status/200"),
		health.WithType(health.CheckTypeLiveness),
		health.WithCritical(false)) // Non-critical for demo

	// Memory check
	registry.Register("memory", health.MemoryCheck(1024),
		health.WithType(health.CheckTypeLiveness))

	// Disk space check
	registry.Register("disk", health.DiskSpaceCheck("/tmp", 1),
		health.WithType(health.CheckTypeLiveness))

	return registry
}

// setupMiddleware configures the complete middleware chain.
func setupMiddleware() *middleware.Chain {
	chain := middleware.NewChain()

	// 1. CORS - Allow all origins for demo
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowedOrigins = []string{"*"}
	corsConfig.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	chain.Add(cors.NewMiddleware(corsConfig))

	// 2. Request/Response Logging
	loggingConfig := logging.DefaultConfig()
	loggingConfig.Logger = func(entry logging.LogEntry) {
		logData, _ := json.MarshalIndent(entry, "", "  ")
		fmt.Printf("HTTP Request Log:\n%s\n\n", logData)
	}
	chain.Add(logging.NewMiddleware(loggingConfig))

	// 3. Timeout Management
	timeoutConfig := timeout.DefaultConfig()
	timeoutConfig.Timeout = 30 * time.Second
	chain.Add(timeout.NewMiddleware(timeoutConfig))

	// 4. Rate Limiting
	rateLimitConfig := ratelimit.Config{
		Enabled:   true,
		SkipPaths: []string{"/health", "/health/live", "/health/ready"},
		Limit:     100,
		Window:    time.Minute,
		Algorithm: ratelimit.TokenBucket,
	}
	chain.Add(ratelimit.NewMiddleware(rateLimitConfig))

	// 5. Bulkhead Pattern for resource isolation
	bulkheadConfig := bulkhead.DefaultConfig()
	bulkheadConfig.MaxConcurrent = 20
	bulkheadConfig.QueueSize = 50
	bulkheadConfig.ResourceKey = func(r *http.Request) string {
		// Isolate by endpoint type
		switch {
		case r.URL.Path == "/api/heavy-task":
			return "heavy-tasks"
		case r.URL.Path == "/api/users":
			return "user-service"
		case r.URL.Path == "/api/orders":
			return "order-service"
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
		fmt.Printf("Retrying request %s (attempt %d) after %v\n", r.URL.Path, attempt, delay)
	}
	chain.Add(retry.NewMiddleware(retryConfig))

	// 7. Response Compression
	compressionConfig := compression.DefaultConfig()
	compressionConfig.Level = -1 // Default compression
	compressionConfig.MinSize = 1024
	chain.Add(compression.NewMiddleware(compressionConfig))

	return chain
}

// setupAPIRoutes configures API routes.
func setupAPIRoutes() http.Handler {
	mux := http.NewServeMux()

	// Users endpoint
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			users := []map[string]interface{}{
				{"id": 1, "name": "John Doe", "email": "john@example.com"},
				{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"users": users,
				"total": len(users),
			})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Orders endpoint
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			orders := []map[string]interface{}{
				{"id": 1, "user_id": 1, "amount": 99.99, "status": "completed"},
				{"id": 2, "user_id": 2, "amount": 149.99, "status": "pending"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"orders": orders,
				"total":  len(orders),
			})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Heavy task endpoint (demonstrates bulkhead isolation)
	mux.HandleFunc("/heavy-task", func(w http.ResponseWriter, r *http.Request) {
		// Simulate heavy processing
		time.Sleep(2 * time.Second)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "Heavy task completed",
			"duration":   "2s",
			"timestamp":  time.Now(),
			"request_id": r.Header.Get("X-Correlation-ID"),
		})
	})

	// Error endpoint (demonstrates retry behavior)
	errorCount := 0
	mux.HandleFunc("/flaky-endpoint", func(w http.ResponseWriter, r *http.Request) {
		errorCount++
		if errorCount%3 == 0 {
			// Success every 3rd request
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Success after retries",
				"attempt": errorCount,
			})
		} else {
			// Fail with retryable status
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	return mux
}
