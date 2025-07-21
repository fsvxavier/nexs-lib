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
)

func main() {
	// Providers are auto-registered via init() - no manual registration needed!

	// Check if running in test mode
	testMode := len(os.Args) > 1 && os.Args[1] == "test"

	// Create logging observer
	loggingObserver := hooks.NewLoggingObserver(log.Default())
	httpserver.AttachObserver(loggingObserver)

	// Create metrics observer
	metricsObserver := hooks.NewMetricsObserver()
	httpserver.AttachObserver(metricsObserver)

	// Use a different port if specified, otherwise default to 8080
	port := 8080
	if len(os.Args) > 2 {
		if p, err := fmt.Sscanf(os.Args[2], "%d", &port); err != nil || p != 1 {
			port = 8080
		}
	}

	// Create configuration
	cfg := config.DefaultConfig().
		WithHost("localhost").
		WithPort(port).
		WithReadTimeout(30 * time.Second).
		WithWriteTimeout(30 * time.Second)

	// Create server
	server, err := httpserver.Create("gin", cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create a simple handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello from Gin HTTP Server!", "framework": "gin", "path": "%s"}`, r.URL.Path)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "framework": "gin", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"users": [{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}], "framework": "gin"}`)
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"message": "User created", "framework": "gin"}`)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `{"error": "Method not allowed", "framework": "gin"}`)
		}
	})

	// Set the handler
	server.SetHandler(mux)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting Gin HTTP server on %s\n", server.GetAddr())
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Create channel for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// If in test mode, set a timer to automatically shutdown after 3 seconds
	if testMode {
		go func() {
			time.Sleep(3 * time.Second)
			log.Println("Test mode: Auto-shutting down after 3 seconds")
			quit <- syscall.SIGTERM
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the server
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
