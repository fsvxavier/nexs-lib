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
	"github.com/fsvxavier/nexs-lib/httpserver/providers/atreugo"
)

func main() {
	// Register the Atreugo provider
	err := httpserver.Register("atreugo", atreugo.Factory)
	if err != nil {
		log.Fatalf("Failed to register atreugo provider: %v", err)
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
		WithPort(8084).
		WithReadTimeout(30 * time.Second).
		WithWriteTimeout(30 * time.Second)

	// Create server
	server, err := httpserver.Create("atreugo", cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create a simple handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello from Atreugo HTTP Server!", "framework": "atreugo", "path": "%s"}`, r.URL.Path)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "framework": "atreugo", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"services": [{"id": 1, "name": "API Gateway"}, {"id": 2, "name": "Auth Service"}], "framework": "atreugo"}`)
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"message": "Service created", "framework": "atreugo"}`)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `{"error": "Method not allowed", "framework": "atreugo"}`)
		}
	})

	mux.HandleFunc("/middleware", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Atreugo middleware integration!", "framework": "atreugo", "features": ["FastHTTP based", "High performance", "Middleware support"]}`)
	})

	mux.HandleFunc("/routing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Advanced routing with Atreugo!", "framework": "atreugo", "method": "%s"}`, r.Method)
	})

	// Set the handler
	server.SetHandler(mux)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting Atreugo HTTP server on %s\n", server.GetAddr())
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
