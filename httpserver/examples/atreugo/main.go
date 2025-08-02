package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/savsgio/atreugo/v11"

	"github.com/fsvxavier/nexs-lib/httpserver"
	"github.com/fsvxavier/nexs-lib/httpserver/config"
)

// LoggingObserver logs all server events
type LoggingObserver struct{}

func (o *LoggingObserver) OnStart(ctx context.Context, addr string) error {
	log.Printf("🚀 Atreugo server started on %s", addr)
	return nil
}

func (o *LoggingObserver) OnStop(ctx context.Context) error {
	log.Printf("🛑 Atreugo server stopped")
	return nil
}

func (o *LoggingObserver) OnError(ctx context.Context, err error) error {
	log.Printf("❌ Atreugo server error: %v", err)
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
	// Create server configuration with Atreugo-specific setup
	cfg, err := config.NewBuilder().
		Apply(
			config.WithAddr("0.0.0.0"),
			config.WithPort(8082),
			config.WithObserver(&LoggingObserver{}),
		).
		Build()
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	// Create Atreugo server
	server, err := httpserver.CreateServerWithConfig("atreugo", cfg)
	if err != nil {
		log.Fatalf("Failed to create Atreugo server: %v", err)
	}

	// Register Atreugo middleware
	corsMiddleware := func(ctx *atreugo.RequestCtx) error {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(http.StatusOK)
			return nil
		}

		return ctx.Next()
	}

	if err := server.RegisterMiddleware(corsMiddleware); err != nil {
		log.Fatalf("Failed to register CORS middleware: %v", err)
	}

	// Register routes with Atreugo handlers
	helloHandler := func(ctx *atreugo.RequestCtx) error {
		response := map[string]interface{}{
			"message":  "Hello from Atreugo!",
			"provider": "atreugo",
			"time":     time.Now().Format(time.RFC3339),
		}
		return ctx.JSONResponse(response, http.StatusOK)
	}

	userHandler := func(ctx *atreugo.RequestCtx) error {
		id := ctx.UserValue("id").(string)
		response := map[string]interface{}{
			"user_id":  id,
			"name":     fmt.Sprintf("User %s", id),
			"provider": "atreugo",
		}
		return ctx.JSONResponse(response, http.StatusOK)
	}

	jsonHandler := func(ctx *atreugo.RequestCtx) error {
		var data map[string]interface{}
		if err := json.Unmarshal(ctx.PostBody(), &data); err != nil {
			return ctx.JSONResponse(map[string]string{"error": err.Error()}, http.StatusBadRequest)
		}

		response := map[string]interface{}{
			"received":  data,
			"provider":  "atreugo",
			"timestamp": time.Now().Unix(),
		}
		return ctx.JSONResponse(response, http.StatusOK)
	}

	healthHandler := func(ctx *atreugo.RequestCtx) error {
		response := map[string]interface{}{
			"status":   "healthy",
			"provider": "atreugo",
			"uptime":   time.Now().Format(time.RFC3339),
		}
		return ctx.JSONResponse(response, http.StatusOK)
	}

	performanceHandler := func(ctx *atreugo.RequestCtx) error {
		// Demonstrating Atreugo's performance features
		start := time.Now()

		// Simulate some work
		time.Sleep(1 * time.Millisecond)

		response := map[string]interface{}{
			"provider":     "atreugo",
			"framework":    "FastHTTP-based",
			"performance":  "high",
			"process_time": time.Since(start).String(),
			"memory_pool":  "enabled",
		}
		return ctx.JSONResponse(response, http.StatusOK)
	}

	// Register routes
	routes := []struct {
		method  string
		path    string
		handler atreugo.View
	}{
		{"GET", "/", helloHandler},
		{"GET", "/hello", helloHandler},
		{"GET", "/user/{id}", userHandler},
		{"POST", "/data", jsonHandler},
		{"GET", "/health", healthHandler},
		{"GET", "/performance", performanceHandler},
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

	log.Printf("🎯 Atreugo server running on %s", server.GetAddr())
	log.Printf("📊 Stats: %+v", server.GetStats())
	log.Println("💡 Try these endpoints:")
	log.Println("   GET  http://localhost:8082/")
	log.Println("   GET  http://localhost:8082/hello")
	log.Println("   GET  http://localhost:8082/user/123")
	log.Println("   POST http://localhost:8082/data (with JSON body)")
	log.Println("   GET  http://localhost:8082/health")
	log.Println("   GET  http://localhost:8082/performance")
	log.Println("🚀 Atreugo uses FastHTTP for maximum performance!")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	select {}
}
