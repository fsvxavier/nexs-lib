package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// LoggingMiddleware logs request and response details
type LoggingMiddleware struct {
	logger *log.Logger
}

func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	start := time.Now()
	m.logger.Printf("📤 [REQUEST] %s %s", req.Method, req.URL)

	// Add request ID header
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	req.Headers["X-Request-ID"] = fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Call next middleware/handler
	resp, err := next(ctx, req)

	duration := time.Since(start)
	if err != nil {
		m.logger.Printf("❌ [ERROR] %s %s - %v (took %v)", req.Method, req.URL, err, duration)
	} else if resp != nil {
		m.logger.Printf("📥 [RESPONSE] %s %s - %d (took %v)", req.Method, req.URL, resp.StatusCode, duration)
	}

	return resp, err
}

// AuthMiddleware adds authentication headers
type AuthMiddleware struct {
	apiKey string
}

func NewAuthMiddleware(apiKey string) *AuthMiddleware {
	return &AuthMiddleware{apiKey: apiKey}
}

func (m *AuthMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	// Add authentication header
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	req.Headers["Authorization"] = "Bearer " + m.apiKey
	req.Headers["X-API-Version"] = "v1"

	return next(ctx, req)
}

// RateLimitMiddleware implements simple rate limiting
type RateLimitMiddleware struct {
	requests    chan struct{}
	maxRequests int
}

func NewRateLimitMiddleware(maxRequests int, window time.Duration) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		requests:    make(chan struct{}, maxRequests),
		maxRequests: maxRequests,
	}

	// Fill the channel initially
	for i := 0; i < maxRequests; i++ {
		m.requests <- struct{}{}
	}

	// Refill the channel periodically
	go func() {
		ticker := time.NewTicker(window / time.Duration(maxRequests))
		defer ticker.Stop()

		for range ticker.C {
			select {
			case m.requests <- struct{}{}:
			default:
				// Channel is full, skip
			}
		}
	}()

	return m
}

func (m *RateLimitMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	// Wait for rate limit token
	select {
	case <-m.requests:
		// Token acquired, proceed
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return next(ctx, req)
}

func main() {
	fmt.Printf("🚀 Middleware Example\n")
	fmt.Printf("=====================\n\n")

	// Create HTTP client
	client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create middlewares
	logger := log.New(log.Writer(), "[HTTP] ", log.LstdFlags)
	loggingMiddleware := NewLoggingMiddleware(logger)
	authMiddleware := NewAuthMiddleware("demo-api-key-12345")
	rateLimitMiddleware := NewRateLimitMiddleware(5, 10*time.Second)

	// Add middlewares to client (they will be executed in reverse order)
	client.
		AddMiddleware(rateLimitMiddleware). // Executed first (rate limiting)
		AddMiddleware(authMiddleware).      // Executed second (auth)
		AddMiddleware(loggingMiddleware)    // Executed third (logging)

	fmt.Println("📋 Middlewares added:")
	fmt.Println("  1. Rate Limit Middleware (5 requests per 10s)")
	fmt.Println("  2. Auth Middleware (adds Bearer token)")
	fmt.Println("  3. Logging Middleware (logs requests/responses)")
	fmt.Println()

	// Make multiple requests to demonstrate middleware functionality
	ctx := context.Background()

	// Request 1: Simple GET
	fmt.Println("1️⃣ Making GET request to /get...")
	resp1, err := client.Get(ctx, "/get")
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("   ✅ Status: %d\n", resp1.StatusCode)
	}
	fmt.Println()

	// Request 2: POST with data
	fmt.Println("2️⃣ Making POST request to /post...")
	postData := map[string]interface{}{
		"message":   "Hello from middleware example!",
		"timestamp": time.Now().Unix(),
	}
	resp2, err := client.Post(ctx, "/post", postData)
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("   ✅ Status: %d\n", resp2.StatusCode)
	}
	fmt.Println()

	// Request 3: Test rate limiting with rapid requests
	fmt.Println("3️⃣ Testing rate limiting with rapid requests...")
	for i := 1; i <= 7; i++ {
		start := time.Now()
		resp, err := client.Get(ctx, fmt.Sprintf("/get?request=%d", i))
		duration := time.Since(start)

		if err != nil {
			log.Printf("   Request %d failed: %v", i, err)
		} else {
			fmt.Printf("   Request %d: Status %d (took %v)\n", i, resp.StatusCode, duration)
		}

		// Small delay to see rate limiting in action
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println()

	// Demonstrate middleware removal
	fmt.Println("4️⃣ Removing rate limit middleware...")
	client.RemoveMiddleware(rateLimitMiddleware)

	fmt.Println("Making request without rate limiting...")
	resp4, err := client.Get(ctx, "/get?no-rate-limit=true")
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("   ✅ Status: %d (should be faster)\n", resp4.StatusCode)
	}

	fmt.Println("\n🎉 Middleware example completed!")
	fmt.Println("\n💡 Key Features Demonstrated:")
	fmt.Println("  • Chaining multiple middlewares")
	fmt.Println("  • Request/response logging")
	fmt.Println("  • Authentication header injection")
	fmt.Println("  • Rate limiting with token bucket")
	fmt.Println("  • Dynamic middleware removal")
	fmt.Println("  • Context propagation")
}
