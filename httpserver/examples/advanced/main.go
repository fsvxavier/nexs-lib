// Package main demonstrates basic usage of the HTTP server library with Fiber provider.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/fsvxavier/nexs-lib/httpserver"
	"github.com/fsvxavier/nexs-lib/httpserver/config"
)

// LoggingObserver demonstrates how to implement a custom observer
type LoggingObserver struct{}

func (l *LoggingObserver) OnStart(ctx context.Context, addr string) error {
	log.Printf("🚀 Server started on %s", addr)
	return nil
}

func (l *LoggingObserver) OnStop(ctx context.Context) error {
	log.Println("🛑 Server stopped")
	return nil
}

func (l *LoggingObserver) OnError(ctx context.Context, err error) error {
	log.Printf("❌ Server error: %v", err)
	return nil
}

func (l *LoggingObserver) OnRequest(ctx context.Context, req interface{}) error {
	log.Printf("📨 Request received")
	return nil
}

func (l *LoggingObserver) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	log.Printf("📤 Response sent in %v", duration)
	return nil
}

func (l *LoggingObserver) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	log.Printf("🔄 Route: %s %s", method, path)
	return nil
}

func (l *LoggingObserver) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	log.Printf("✅ Route completed: %s %s (%v)", method, path, duration)
	return nil
}

func main() {
	log.Println("🔧 Starting HTTP Server Example with Fiber provider")

	// Create configuration using the builder pattern
	cfg, err := config.NewBuilder().
		Apply(config.WithAddr("localhost")).
		Apply(config.WithPort(8080)).
		Apply(config.WithProvider("fiber")).
		Apply(config.WithObserver(&LoggingObserver{})).
		Build()
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	// Create server using the default manager (Fiber provider is default)
	server, err := httpserver.CreateServerWithConfig("fiber", cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Register middleware
	loggingMiddleware := func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Printf("[%s] %s %d - %v",
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			time.Since(start))
		return err
	}

	if err := server.RegisterMiddleware(loggingMiddleware); err != nil {
		log.Fatalf("Failed to register middleware: %v", err)
	}

	// Register routes
	helloHandler := func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from Fiber provider!",
			"time":    time.Now().Format(time.RFC3339),
		})
	}

	if err := server.RegisterRoute("GET", "/hello", helloHandler); err != nil {
		log.Fatalf("Failed to register route: %v", err)
	}

	userHandler := func(c *fiber.Ctx) error {
		name := c.Params("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Name parameter is required",
			})
		}
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Hello, %s!", name),
			"user":    name,
			"time":    time.Now().Format(time.RFC3339),
		})
	}

	if err := server.RegisterRoute("GET", "/hello/:name", userHandler); err != nil {
		log.Fatalf("Failed to register parameterized route: %v", err)
	}

	healthHandler := func(c *fiber.Ctx) error {
		stats := server.GetStats()
		return c.JSON(fiber.Map{
			"status":       "healthy",
			"provider":     stats.Provider,
			"uptime":       time.Since(stats.StartTime).String(),
			"requests":     stats.RequestCount,
			"errors":       stats.ErrorCount,
			"avg_response": stats.AverageResponseTime.String(),
		})
	}

	if err := server.RegisterRoute("GET", "/health", healthHandler); err != nil {
		log.Fatalf("Failed to register health route: %v", err)
	}

	// Start server
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Log server info
	log.Printf("🌟 Server running at: %s", server.GetAddr())
	log.Println("📋 Available endpoints:")
	log.Println("   GET /hello")
	log.Println("   GET /hello/:name")
	log.Println("   GET /health")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🔄 Shutting down server...")

	// Stop server gracefully
	if err := server.Stop(ctx); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	log.Println("👋 Server stopped successfully")
}
