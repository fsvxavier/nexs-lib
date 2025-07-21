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
	"github.com/fsvxavier/nexs-lib/httpserver/graceful"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func main() {
	log.Println("🚀 Graceful Shutdown Example - Multiple Providers")
	log.Println("===============================================")

	// Check if running in test mode
	testMode := len(os.Args) > 1 && os.Args[1] == "test"

	// Create graceful manager
	manager := graceful.NewManager()

	// Configure timeouts
	manager.SetDrainTimeout(5 * time.Second)
	manager.SetShutdownTimeout(10 * time.Second)

	// Add health checks
	manager.AddHealthCheck("database", func() interfaces.HealthCheck {
		return interfaces.HealthCheck{
			Status:    "healthy",
			Message:   "Database connection OK",
			Duration:  5 * time.Millisecond,
			Timestamp: time.Now(),
		}
	})

	manager.AddHealthCheck("redis", func() interfaces.HealthCheck {
		return interfaces.HealthCheck{
			Status:    "healthy",
			Message:   "Redis connection OK",
			Duration:  2 * time.Millisecond,
			Timestamp: time.Now(),
		}
	})

	// Add shutdown hooks
	manager.AddPreShutdownHook(func() error {
		log.Println("📝 Pre-shutdown: Saving application state...")
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	manager.AddPostShutdownHook(func() error {
		log.Println("🧹 Post-shutdown: Cleanup completed")
		return nil
	})

	// Create servers using all available providers
	servers := createAllServers(testMode)

	// Add servers to graceful manager
	for name, server := range servers {
		manager.RegisterServer(name, server)
		log.Printf("✅ Registered server: %s", name)
	}

	// Start all servers
	log.Println("🚀 Starting all servers...")
	startAllServers(servers)

	// Display server information
	displayServerInfo(servers)

	// Setup graceful shutdown
	setupGracefulShutdown(manager, testMode)

	// Keep main goroutine alive
	select {}
}

// createAllServers creates servers using all available providers
func createAllServers(testMode bool) map[string]interfaces.HTTPServer {
	servers := make(map[string]interfaces.HTTPServer)
	basePort := 8080

	if testMode {
		basePort = 8090 // Use different ports in test mode to avoid conflicts
	}

	// Provider configurations
	providers := []struct {
		name     string
		provider string
		port     int
		path     string
	}{
		{"nethttp-api", "nethttp", basePort, "/api"},
		{"gin-web", "gin", basePort + 1, "/web"},
		{"fiber-admin", "fiber", basePort + 2, "/admin"},
		{"echo-service", "echo", basePort + 3, "/service"},
	}

	// Create servers for each provider
	for _, p := range providers {
		cfg := config.DefaultConfig().
			WithHost("localhost").
			WithPort(p.port).
			WithReadTimeout(10 * time.Second).
			WithWriteTimeout(10 * time.Second)

		server, err := httpserver.Create(p.provider, cfg)
		if err != nil {
			log.Printf("⚠️  Failed to create %s server: %v", p.name, err)
			continue
		}

		// Set handler for this server
		server.SetHandler(createHandler(p.name, p.path))
		servers[p.name] = server

		log.Printf("✅ Created %s server on port %d", p.name, p.port)
	}

	return servers
}

// startAllServers starts all servers in separate goroutines
func startAllServers(servers map[string]interfaces.HTTPServer) {
	for name, server := range servers {
		go func(serverName string, srv interfaces.HTTPServer) {
			log.Printf("🚀 Starting %s on %s", serverName, srv.GetAddr())
			if err := srv.Start(); err != nil {
				log.Printf("❌ %s error: %v", serverName, err)
			}
		}(name, server)
	}

	// Give servers time to start
	time.Sleep(500 * time.Millisecond)
}

// displayServerInfo shows information about all running servers
func displayServerInfo(servers map[string]interfaces.HTTPServer) {
	log.Println("\n📊 Server Information:")
	log.Println("=====================")

	for name, server := range servers {
		log.Printf("• %s: http://%s", name, server.GetAddr())
	}

	log.Println("\n🌐 Available endpoints:")
	log.Println("• GET  /api/health    - NetHTTP API health")
	log.Println("• GET  /web/health    - Gin Web health")
	log.Println("• GET  /admin/health  - Fiber Admin health")
	log.Println("• GET  /service/health - Echo Service health")
	log.Println("• GET  :9090/health   - Overall system health")

	if len(os.Args) > 1 && os.Args[1] == "test" {
		log.Println("\n⏱️  Test mode: Auto-shutdown in 3 seconds...")
	} else {
		log.Println("\nPress Ctrl+C for graceful shutdown...")
	}
}

// createHandler creates a handler for a specific server with its path prefix
func createHandler(serverName, pathPrefix string) http.Handler {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc(pathPrefix+"/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
	"status": "healthy",
	"server": "%s",
	"timestamp": "%s",
	"path": "%s"
}`, serverName, time.Now().Format(time.RFC3339), pathPrefix)
	})

	// Data endpoint
	mux.HandleFunc(pathPrefix+"/data", func(w http.ResponseWriter, r *http.Request) {
		// Simulate some work
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
	"data": "Sample data from %s",
	"server": "%s", 
	"timestamp": "%s",
	"method": "%s"
}`, serverName, serverName, time.Now().Format(time.RFC3339), r.Method)
	})

	// Root endpoint
	mux.HandleFunc(pathPrefix+"/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != pathPrefix+"/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
	"message": "Welcome to %s",
	"server": "%s",
	"endpoints": ["%s/health", "%s/data"],
	"timestamp": "%s"
}`, serverName, serverName, pathPrefix, pathPrefix, time.Now().Format(time.RFC3339))
	})

	return mux
}

func setupGracefulShutdown(manager *graceful.Manager, testMode bool) {
	// Create channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// If in test mode, set a timer to automatically shutdown after 3 seconds
	if testMode {
		go func() {
			time.Sleep(3 * time.Second)
			log.Println("⏰ Test mode: Auto-shutting down after 3 seconds")
			sigChan <- syscall.SIGTERM
		}()
	}

	go func() {
		// Wait for signal
		sig := <-sigChan
		log.Printf("📨 Received signal: %v", sig)

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Println("🛑 Initiating graceful shutdown...")

		// Perform graceful shutdown
		if err := manager.GracefulShutdown(ctx); err != nil {
			log.Printf("❌ Error during graceful shutdown: %v", err)
			os.Exit(1)
		}

		log.Println("✅ Application shut down successfully")
		os.Exit(0)
	}()

	// Setup health status endpoint on port 9090
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

		log.Println("🏥 Health status endpoint available at :9090/health")
		if err := http.ListenAndServe(":9090", healthMux); err != nil {
			log.Printf("Health server error: %v", err)
		}
	}()
}
