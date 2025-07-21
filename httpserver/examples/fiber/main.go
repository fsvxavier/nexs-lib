package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver"
	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/hooks"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/fiber"
)

func main() {
	// Register the Fiber provider
	err := httpserver.Register("fiber", fiber.Factory)
	if err != nil {
		log.Fatalf("Failed to register fiber provider: %v", err)
	}

	// Create logging observer
	loggingObserver := hooks.NewLoggingObserver(log.Default())
	httpserver.AttachObserver(loggingObserver)

	// Create metrics observer
	metricsObserver := hooks.NewMetricsObserver()
	httpserver.AttachObserver(metricsObserver)

	// Create configuration
	cfg := config.DefaultConfig().
		WithHost("localhost").
		WithPort(8082).
		WithReadTimeout(30 * time.Second).
		WithWriteTimeout(30 * time.Second)

	// Create server
	server, err := httpserver.Create("fiber", cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create a simple handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello from Fiber HTTP Server!", "framework": "fiber", "path": "%s"}`, r.URL.Path)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "framework": "fiber", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"orders": [{"id": 1, "total": 99.99}, {"id": 2, "total": 149.99}], "framework": "fiber"}`)
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"message": "Order created", "framework": "fiber"}`)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `{"error": "Method not allowed", "framework": "fiber"}`)
		}
	})

	mux.HandleFunc("/performance", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Fiber is known for high performance!", "framework": "fiber", "benchmark": "fastest Go web framework"}`)
	})

	// Set the handler
	server.SetHandler(mux)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting Fiber HTTP server on %s\n", server.GetAddr())
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Stop(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped")
}
