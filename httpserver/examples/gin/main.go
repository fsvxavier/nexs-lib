package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fsvxavier/nexs-lib/httpserver"
	"github.com/fsvxavier/nexs-lib/httpserver/config"
)

// LoggingObserver logs all server events
type LoggingObserver struct{}

func (o *LoggingObserver) OnStart(ctx context.Context, addr string) error {
	log.Printf("🚀 Gin server started on %s", addr)
	return nil
}

func (o *LoggingObserver) OnStop(ctx context.Context) error {
	log.Printf("🛑 Gin server stopped")
	return nil
}

func (o *LoggingObserver) OnError(ctx context.Context, err error) error {
	log.Printf("❌ Gin server error: %v", err)
	return nil
}

func (o *LoggingObserver) OnRequest(ctx context.Context, req interface{}) error {
	log.Printf("📥 Request received")
	return nil
}

func (o *LoggingObserver) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	log.Printf("📤 Response sent in %v", duration)
	return nil
}

func (o *LoggingObserver) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	log.Printf("🔀 Route entered: %s %s", method, path)
	return nil
}

func (o *LoggingObserver) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	log.Printf("🔚 Route exited: %s %s in %v", method, path, duration)
	return nil
}

func main() {
	// Create server configuration with Gin-specific setup
	cfg, err := config.NewBuilder().
		Apply(
			config.WithAddr("0.0.0.0"),
			config.WithPort(8080),
			config.WithObserver(&LoggingObserver{}),
		).
		Build()
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	// Create Gin server
	server, err := httpserver.CreateServerWithConfig("gin", cfg)
	if err != nil {
		log.Fatalf("Failed to create Gin server: %v", err)
	}

	// Register Gin middleware
	corsMiddleware := gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	if err := server.RegisterMiddleware(corsMiddleware); err != nil {
		log.Fatalf("Failed to register CORS middleware: %v", err)
	}

	// Register routes with Gin handlers
	helloHandler := gin.HandlerFunc(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":  "Hello from Gin!",
			"provider": "gin",
			"time":     time.Now().Format(time.RFC3339),
		})
	})

	userHandler := gin.HandlerFunc(func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"user_id":  id,
			"name":     fmt.Sprintf("User %s", id),
			"provider": "gin",
		})
	})

	jsonHandler := gin.HandlerFunc(func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"received":  data,
			"provider":  "gin",
			"timestamp": time.Now().Unix(),
		})
	})

	// Register routes
	routes := []struct {
		method  string
		path    string
		handler gin.HandlerFunc
	}{
		{"GET", "/", helloHandler},
		{"GET", "/hello", helloHandler},
		{"GET", "/user/:id", userHandler},
		{"POST", "/data", jsonHandler},
	}

	for _, route := range routes {
		if err := server.RegisterRoute(route.method, route.path, route.handler); err != nil {
			log.Fatalf("Failed to register route %s %s: %v", route.method, route.path, err)
		}
	}

	// Start the server
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("🎯 Gin server running on %s", server.GetAddr())
	log.Printf("📊 Stats: %+v", server.GetStats())
	log.Println("💡 Try these endpoints:")
	log.Println("   GET  http://localhost:8080/")
	log.Println("   GET  http://localhost:8080/hello")
	log.Println("   GET  http://localhost:8080/user/123")
	log.Println("   POST http://localhost:8080/data (with JSON body)")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	select {}
}
