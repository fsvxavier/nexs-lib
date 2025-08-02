package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fsvxavier/nexs-lib/httpserver"
)

// ExampleObserver demonstrates observing server events
type ExampleObserver struct{}

func (o *ExampleObserver) OnStart(ctx context.Context, addr string) error {
	fmt.Printf("✅ FastHTTP Server started on %s\n", addr)
	return nil
}

func (o *ExampleObserver) OnStop(ctx context.Context) error {
	fmt.Println("🛑 FastHTTP Server stopped")
	return nil
}

func (o *ExampleObserver) OnError(ctx context.Context, err error) error {
	fmt.Printf("❌ FastHTTP Server error: %v\n", err)
	return nil
}

func (o *ExampleObserver) OnRequest(ctx context.Context, req interface{}) error {
	if fastReq, ok := req.(*fasthttp.RequestCtx); ok {
		fmt.Printf("📥 Request: %s %s\n",
			string(fastReq.Method()),
			string(fastReq.Path()))
	}
	return nil
}

func (o *ExampleObserver) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if fastReq, ok := req.(*fasthttp.RequestCtx); ok {
		fmt.Printf("📤 Response: %d (took %v)\n",
			fastReq.Response.StatusCode(),
			duration)
	}
	return nil
}

func (o *ExampleObserver) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	fmt.Printf("🚀 Entering route: %s %s\n", method, path)
	return nil
}

func (o *ExampleObserver) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	fmt.Printf("🏁 Exiting route: %s %s (took %v)\n", method, path, duration)
	return nil
}

func main() {
	// Create FastHTTP server with default configuration (port 8080)
	server, err := httpserver.CreateServer("fasthttp")
	if err != nil {
		log.Fatalf("Failed to create FastHTTP server: %v", err)
	}

	// Attach observer for monitoring
	observer := &ExampleObserver{}
	err = server.AttachObserver(observer)
	if err != nil {
		log.Fatalf("Failed to attach observer: %v", err)
	}

	// Add global middleware for request logging
	loggingMiddleware := func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		fmt.Printf("🔄 Processing %s %s\n",
			string(ctx.Method()),
			string(ctx.Path()))

		// Set a custom header
		ctx.Response.Header.Set("X-Powered-By", "FastHTTP-Nexs-Lib")

		// Add processing time to context
		ctx.SetUserValue("start_time", start)
	}

	err = server.RegisterMiddleware(loggingMiddleware)
	if err != nil {
		log.Fatalf("Failed to register middleware: %v", err)
	}

	// CORS middleware
	corsMiddleware := func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusOK)
		}
	}

	err = server.RegisterMiddleware(corsMiddleware)
	if err != nil {
		log.Fatalf("Failed to register CORS middleware: %v", err)
	}

	// Register routes

	// Health check endpoint
	healthHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(fasthttp.StatusOK)

		response := `{
			"status": "healthy",
			"provider": "fasthttp",
			"timestamp": "%s"
		}`

		fmt.Fprintf(ctx, response, time.Now().Format(time.RFC3339))
	}

	err = server.RegisterRoute("GET", "/health", healthHandler)
	if err != nil {
		log.Fatalf("Failed to register health route: %v", err)
	}

	// JSON API endpoint
	jsonHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(fasthttp.StatusOK)

		// Get start time from middleware
		startTime, _ := ctx.UserValue("start_time").(time.Time)
		processingTime := time.Since(startTime)

		response := `{
			"message": "Hello from FastHTTP!",
			"method": "%s",
			"path": "%s",
			"processing_time": "%v",
			"headers": {
				"user_agent": "%s",
				"content_type": "%s"
			}
		}`

		fmt.Fprintf(ctx, response,
			string(ctx.Method()),
			string(ctx.Path()),
			processingTime,
			string(ctx.Request.Header.UserAgent()),
			string(ctx.Request.Header.ContentType()),
		)
	}

	err = server.RegisterRoute("GET", "/api/hello", jsonHandler)
	if err != nil {
		log.Fatalf("Failed to register JSON route: %v", err)
	}

	// POST endpoint that echoes request body
	echoHandler := func(ctx *fasthttp.RequestCtx) {
		body := ctx.PostBody()

		ctx.SetContentType("application/json")
		ctx.SetStatusCode(fasthttp.StatusOK)

		response := `{
			"echo": %s,
			"content_length": %d,
			"method": "%s"
		}`

		if len(body) > 0 {
			fmt.Fprintf(ctx, response, string(body), len(body), string(ctx.Method()))
		} else {
			fmt.Fprintf(ctx, response, `"empty body"`, 0, string(ctx.Method()))
		}
	}

	err = server.RegisterRoute("POST", "/api/echo", echoHandler)
	if err != nil {
		log.Fatalf("Failed to register echo route: %v", err)
	}

	// Performance test endpoint
	perfHandler := func(ctx *fasthttp.RequestCtx) {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		ctx.SetContentType("text/plain")
		ctx.SetStatusCode(fasthttp.StatusOK)

		fmt.Fprintf(ctx, "FastHTTP Performance Test - Request ID: %d",
			ctx.ID())
	}

	err = server.RegisterRoute("GET", "/perf", perfHandler)
	if err != nil {
		log.Fatalf("Failed to register performance route: %v", err)
	}

	// Start server
	ctx := context.Background()

	fmt.Println("🚀 Starting FastHTTP server...")
	fmt.Printf("Provider: %s\n", "fasthttp")
	fmt.Printf("Address: %s\n", server.GetAddr())
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /health      - Health check")
	fmt.Println("  GET  /api/hello   - JSON API example")
	fmt.Println("  POST /api/echo    - Echo request body")
	fmt.Println("  GET  /perf        - Performance test")
	fmt.Println("\nExample requests:")
	fmt.Printf("  curl http://%s/health\n", server.GetAddr())
	fmt.Printf("  curl http://%s/api/hello\n", server.GetAddr())
	fmt.Printf("  curl -X POST -d '{\"message\":\"test\"}' http://%s/api/echo\n", server.GetAddr())
	fmt.Printf("  curl http://%s/perf\n", server.GetAddr())
	fmt.Println("\nPress Ctrl+C to stop...")

	// Start the server
	err = server.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	// Test the server with some requests
	fmt.Println("\n🧪 Testing server endpoints...")

	// Test health endpoint
	testEndpoint("GET", fmt.Sprintf("http://%s/health", server.GetAddr()))

	// Test JSON API
	testEndpoint("GET", fmt.Sprintf("http://%s/api/hello", server.GetAddr()))

	// Print server stats
	stats := server.GetStats()
	fmt.Printf("\n📊 Server Statistics:\n")
	fmt.Printf("  Provider: %s\n", stats.Provider)
	fmt.Printf("  Requests: %d\n", stats.RequestCount)
	fmt.Printf("  Running: %v\n", server.IsRunning())

	// Keep server running for demonstration
	fmt.Println("\n⏰ Server will run for 30 seconds for testing...")
	time.Sleep(30 * time.Second)

	// Stop server gracefully
	fmt.Println("\n🛑 Stopping server...")
	err = server.Stop(ctx)
	if err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	fmt.Println("✅ Example completed!")
}

func testEndpoint(method, url string) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ Failed to test %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ %s %s - Status: %d\n", method, url, resp.StatusCode)
}
