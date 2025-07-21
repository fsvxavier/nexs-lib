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

	"github.com/fsvxavier/nexs-lib/httpserver/graceful"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/echo"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/gin"
)

func main() {
	// Create graceful manager
	manager := graceful.NewManager()

	// Configure timeouts
	manager.SetDrainTimeout(30 * time.Second)
	manager.SetShutdownTimeout(60 * time.Second)

	// Add health checks
	manager.AddHealthCheck("database", func() interfaces.HealthCheck {
		// Simulate database check
		return interfaces.HealthCheck{
			Status:    "healthy",
			Message:   "Database connection OK",
			Duration:  5 * time.Millisecond,
			Timestamp: time.Now(),
		}
	})

	manager.AddHealthCheck("redis", func() interfaces.HealthCheck {
		// Simulate Redis check
		return interfaces.HealthCheck{
			Status:    "healthy",
			Message:   "Redis connection OK",
			Duration:  2 * time.Millisecond,
			Timestamp: time.Now(),
		}
	})

	// Add shutdown hooks
	manager.AddPreShutdownHook(func() error {
		log.Println("Pre-shutdown: Saving application state...")
		time.Sleep(100 * time.Millisecond) // Simulate saving state
		return nil
	})

	manager.AddPreShutdownHook(func() error {
		log.Println("Pre-shutdown: Notifying external services...")
		time.Sleep(50 * time.Millisecond) // Simulate notification
		return nil
	})

	manager.AddPostShutdownHook(func() error {
		log.Println("Post-shutdown: Cleaning up resources...")
		time.Sleep(100 * time.Millisecond) // Simulate cleanup
		return nil
	})

	// Create and start multiple servers
	ginServer, err := gin.NewServer(":8080")
	if err != nil {
		log.Fatalf("Failed to create Gin server: %v", err)
	}

	echoServer, err := echo.NewServer(":8081")
	if err != nil {
		log.Fatalf("Failed to create Echo server: %v", err)
	}

	// Set up basic handlers
	ginServer.SetHandler(createGinHandler())
	echoServer.SetHandler(createEchoHandler())

	// Register servers with graceful manager
	manager.RegisterServer("gin-api", ginServer)
	manager.RegisterServer("echo-admin", echoServer)

	// Start servers
	go func() {
		log.Println("Starting Gin server on :8080...")
		if err := ginServer.Start(); err != nil {
			log.Printf("Gin server error: %v", err)
		}
	}()

	go func() {
		log.Println("Starting Echo server on :8081...")
		if err := echoServer.Start(); err != nil {
			log.Printf("Echo server error: %v", err)
		}
	}()

	// Setup graceful shutdown handling
	setupGracefulShutdown(manager)

	// Keep main goroutine alive
	select {}
}

func createGinHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"gin-api","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		// Simulate some work
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"data":"Sample data from Gin","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	return mux
}

func createEchoHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/admin/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"echo-admin","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/admin/status", func(w http.ResponseWriter, r *http.Request) {
		// Simulate some work
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"admin":"status","uptime":"1h30m","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	return mux
}

func setupGracefulShutdown(manager *graceful.Manager) {
	// Create channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		// Wait for signal
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		log.Println("Initiating graceful shutdown...")

		// Perform graceful shutdown
		if err := manager.GracefulShutdown(ctx); err != nil {
			log.Printf("Error during graceful shutdown: %v", err)
			os.Exit(1)
		}

		log.Println("Application shut down successfully")
		os.Exit(0)
	}()

	// Also setup health status endpoint
	go func() {
		healthMux := http.NewServeMux()
		healthMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			status := manager.GetHealthStatus()
			w.Header().Set("Content-Type", "application/json")

			if status.Status == "healthy" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}

			fmt.Fprintf(w, `{
	"status": "%s",
	"version": "%s",
	"timestamp": "%s",
	"uptime": "%s",
	"connections": %d,
	"checks": {`,
				status.Status,
				status.Version,
				status.Timestamp.Format(time.RFC3339),
				status.Uptime.String(),
				status.Connections)

			first := true
			for name, check := range status.Checks {
				if !first {
					fmt.Fprint(w, ",")
				}
				first = false
				fmt.Fprintf(w, `
		"%s": {
			"status": "%s",
			"message": "%s",
			"duration": "%s",
			"timestamp": "%s"
		}`, name, check.Status, check.Message, check.Duration.String(), check.Timestamp.Format(time.RFC3339))
			}

			fmt.Fprint(w, `
	}
}`)
		})

		log.Println("Health status endpoint available at :9090/health")
		if err := http.ListenAndServe(":9090", healthMux); err != nil {
			log.Printf("Health server error: %v", err)
		}
	}()
}
