// Package main demonstrates basic middleware usage.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/cors"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/logging"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/ratelimit"
)

func main() {
	// Create middleware chain
	chain := middleware.NewChain()

	// Add CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowedOrigins = []string{"*"} // Allow all origins for demo
	chain.Add(cors.NewMiddleware(corsConfig))

	// Add logging middleware
	loggingConfig := logging.DefaultConfig()
	loggingConfig.Logger = func(entry logging.LogEntry) {
		fmt.Printf("[%s] %s %s - %d (%v)\n",
			entry.Timestamp.Format("15:04:05"),
			entry.Method,
			entry.Path,
			entry.StatusCode,
			entry.Duration)
	}
	chain.Add(logging.NewMiddleware(loggingConfig))

	// Add rate limiting middleware
	rateLimitConfig := ratelimit.Config{
		Enabled:   true,
		Limit:     10, // 10 requests per minute for demo
		Window:    time.Minute,
		Algorithm: ratelimit.TokenBucket,
	}
	chain.Add(ratelimit.NewMiddleware(rateLimitConfig))

	// Simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello World", "middleware": "working"}`))
	})

	// Apply middleware chain
	server := &http.Server{
		Addr:    ":8080",
		Handler: chain.Then(handler),
	}

	fmt.Println("Simple middleware example starting on :8080")
	fmt.Println("Try: curl http://localhost:8080")
	fmt.Println("Rate limit: 10 requests per minute")

	log.Fatal(server.ListenAndServe())
}
