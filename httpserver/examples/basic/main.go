// Package main demonstrates simple usage of the HTTP server library.
package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/fsvxavier/nexs-lib/httpserver"
	"github.com/fsvxavier/nexs-lib/httpserver/config"
)

func main() {
	// Create a simple server configuration
	cfg := config.NewBaseConfig()

	// Create server with Fiber provider (default)
	server, err := httpserver.CreateServerWithConfig("fiber", cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Register a simple route
	handler := func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, World!",
			"status":  "success",
		})
	}

	if err := server.RegisterRoute("GET", "/", handler); err != nil {
		log.Fatalf("Failed to register route: %v", err)
	}

	// Start server
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("Server running at: %s", server.GetAddr())
	log.Println("Try: curl http://localhost:8080/")

	// In a real application, you would add signal handling here
	select {} // Keep the server running
}
