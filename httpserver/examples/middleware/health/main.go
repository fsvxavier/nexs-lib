// Package main demonstrates health check middleware usage.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/health"
)

func main() {
	// Create health check registry
	registry := health.NewRegistry()

	// Register a simple ping check
	registry.Register("ping", simpleCheck("Ping check passed"),
		health.WithType(health.CheckTypeLiveness))

	// Register database connection check (simulated)
	registry.Register("database", databaseCheck(),
		health.WithType(health.CheckTypeReadiness),
		health.WithCritical(true))

	// Register external service check
	registry.Register("external-api", health.URLCheck("https://httpbin.org/status/200"),
		health.WithType(health.CheckTypeLiveness),
		health.WithCritical(false))

	// Register memory usage check
	registry.Register("memory", health.MemoryCheck(1024), // 1GB limit
		health.WithType(health.CheckTypeLiveness))

	// Register disk space check
	registry.Register("disk", health.DiskSpaceCheck("/tmp", 1), // 1GB minimum
		health.WithType(health.CheckTypeLiveness))

	// Create health handler
	healthHandler := health.NewHandler(registry)

	// Setup routes
	mux := http.NewServeMux()

	// Health endpoints
	mux.Handle("/health", healthHandler.HealthHandler())
	mux.Handle("/health/live", healthHandler.LivenessHandler())
	mux.Handle("/health/ready", healthHandler.ReadinessHandler())
	mux.Handle("/health/startup", healthHandler.StartupHandler())

	// Simple API endpoint
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Hello, World!", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	fmt.Println("Health check server starting on :8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /health         - Overall health status")
	fmt.Println("  GET /health/live    - Liveness probe")
	fmt.Println("  GET /health/ready   - Readiness probe")
	fmt.Println("  GET /health/startup - Startup probe")
	fmt.Println("  GET /api/hello      - Simple API endpoint")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// databaseCheck simulates a database connection check.
func databaseCheck() health.Check {
	return func(ctx context.Context) interfaces.HealthCheckResult {
		// Simulate database connection check
		select {
		case <-time.After(100 * time.Millisecond):
			// Simulate successful connection
			return interfaces.HealthCheckResult{
				Status:  "healthy",
				Message: "Database connection successful",
				Metadata: map[string]interface{}{
					"connection_time":    "100ms",
					"active_connections": 10,
					"max_connections":    100,
				},
			}
		case <-ctx.Done():
			return interfaces.HealthCheckResult{
				Status:  "unhealthy",
				Message: "Database check timeout",
				Metadata: map[string]interface{}{
					"error": ctx.Err().Error(),
				},
			}
		}
	}
}

// simpleCheck creates a simple health check that always passes.
func simpleCheck(message string) health.Check {
	return func(ctx context.Context) interfaces.HealthCheckResult {
		return interfaces.HealthCheckResult{
			Status:  "healthy",
			Message: message,
		}
	}
}
