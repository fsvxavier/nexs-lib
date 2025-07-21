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
	"github.com/fsvxavier/nexs-lib/httpserver/providers/nethttp"
)

func main() {
	// Register the net/http provider
	err := httpserver.Register("nethttp", nethttp.Factory)
	if err != nil {
		log.Fatalf("Failed to register nethttp provider: %v", err)
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
		WithPort(8080).
		WithReadTimeout(30 * time.Second).
		WithWriteTimeout(30 * time.Second)

	// Create server
	server, err := httpserver.Create("nethttp", cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create a simple handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello from nexs-lib httpserver!\nPath: %s\nMethod: %s\n", r.URL.Path, r.Method)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","server":"nethttp"}`)
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := metricsObserver.GetMetrics("nethttp")
		if metrics == nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "No metrics available")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"request_count": %d,
			"error_count": %d,
			"active_requests": %d,
			"start_time": "%s",
			"last_request_time": "%s"
		}`,
			metrics.RequestCount,
			metrics.ErrorCount,
			metrics.ActiveRequests,
			metrics.StartTime.Format(time.RFC3339),
			metrics.GetLastRequestTime().Format(time.RFC3339),
		)
	})

	// Set the handler
	server.SetHandler(mux)

	// Start the server
	log.Printf("Starting server on %s", server.GetAddr())
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server exited gracefully")
	}
}
