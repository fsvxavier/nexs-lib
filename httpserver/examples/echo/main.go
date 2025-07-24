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
	// Check if running in test mode
	testMode := len(os.Args) > 1 && os.Args[1] == "test"

	// Create logging observer
	loggingObserver := hooks.NewLoggingObserver(log.Default())
	httpserver.AttachObserver(loggingObserver)

	// Create metrics observer
	metricsObserver := hooks.NewMetricsObserver()
	httpserver.AttachObserver(metricsObserver)

	// Use a different port if specified, otherwise default to 8081
	port := 8081
	if len(os.Args) > 2 {
		if p, err := fmt.Sscanf(os.Args[2], "%d", &port); err != nil || p != 1 {
			port = 8081
		}
	}

	// Create configuration
	cfg := config.DefaultConfig().
		WithHost("localhost").
		WithPort(port).
		WithReadTimeout(30 * time.Second).
		WithWriteTimeout(30 * time.Second)

	// Create server
	server, err := httpserver.Create("echo", cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create a simple handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello from Echo HTTP Server!", "framework": "echo", "path": "%s"}`, r.URL.Path)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "framework": "echo", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"products": [{"id": 1, "name": "Laptop"}, {"id": 2, "name": "Phone"}], "framework": "echo"}`)
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"message": "Product created", "framework": "echo"}`)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `{"error": "Method not allowed", "framework": "echo"}`)
		}
	})

	mux.HandleFunc("/middleware-test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Echo middleware working!", "headers": "%v"}`, r.Header)
	})

	// Set the handler
	server.SetHandler(mux)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting Echo HTTP server on %s\n", server.GetAddr())
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
