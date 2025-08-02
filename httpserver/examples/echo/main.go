package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/fsvxavier/nexs-lib/httpserver"
	"github.com/fsvxavier/nexs-lib/httpserver/config"
)

// LoggingObserver logs all server events
type LoggingObserver struct{}

func (o *LoggingObserver) OnStart(ctx context.Context, addr string) error {
	log.Printf("🚀 Echo server started on %s", addr)
	return nil
}

func (o *LoggingObserver) OnStop(ctx context.Context) error {
	log.Printf("🛑 Echo server stopped")
	return nil
}

func (o *LoggingObserver) OnError(ctx context.Context, err error) error {
	log.Printf("❌ Echo server error: %v", err)
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
	// Create server configuration with Echo-specific setup
	cfg, err := config.NewBuilder().
		Apply(
			config.WithAddr("0.0.0.0"),
			config.WithPort(8081),
			config.WithObserver(&LoggingObserver{}),
		).
		Build()
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	// Create Echo server
	server, err := httpserver.CreateServerWithConfig("echo", cfg)
	if err != nil {
		log.Fatalf("Failed to create Echo server: %v", err)
	}

	// Register Echo middleware
	corsMiddleware := echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusOK)
			}

			return next(c)
		}
	})

	if err := server.RegisterMiddleware(corsMiddleware); err != nil {
		log.Fatalf("Failed to register CORS middleware: %v", err)
	}

	// Register routes with Echo handlers
	helloHandler := echo.HandlerFunc(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":  "Hello from Echo!",
			"provider": "echo",
			"time":     time.Now().Format(time.RFC3339),
		})
	})

	userHandler := echo.HandlerFunc(func(c echo.Context) error {
		id := c.Param("id")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id":  id,
			"name":     fmt.Sprintf("User %s", id),
			"provider": "echo",
		})
	})

	jsonHandler := echo.HandlerFunc(func(c echo.Context) error {
		var data map[string]interface{}
		if err := c.Bind(&data); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"received":  data,
			"provider":  "echo",
			"timestamp": time.Now().Unix(),
		})
	})

	healthHandler := echo.HandlerFunc(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":   "healthy",
			"provider": "echo",
			"uptime":   time.Now().Format(time.RFC3339),
		})
	})

	// Register routes
	routes := []struct {
		method  string
		path    string
		handler echo.HandlerFunc
	}{
		{"GET", "/", helloHandler},
		{"GET", "/hello", helloHandler},
		{"GET", "/user/:id", userHandler},
		{"POST", "/data", jsonHandler},
		{"GET", "/health", healthHandler},
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

	log.Printf("🎯 Echo server running on %s", server.GetAddr())
	log.Printf("📊 Stats: %+v", server.GetStats())
	log.Println("💡 Try these endpoints:")
	log.Println("   GET  http://localhost:8081/")
	log.Println("   GET  http://localhost:8081/hello")
	log.Println("   GET  http://localhost:8081/user/123")
	log.Println("   POST http://localhost:8081/data (with JSON body)")
	log.Println("   GET  http://localhost:8081/health")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	select {}
}
