// Package main demonstrates rate limiting middleware usage.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/ratelimit"
)

func main() {
	// Create middleware chain
	chain := middleware.NewChain()

	// Configure rate limiting with different algorithms
	tokenBucketConfig := ratelimit.Config{
		Enabled:   true,
		SkipPaths: []string{"/health"},
		Limit:     10, // 10 requests per minute
		Window:    time.Minute,
		Algorithm: ratelimit.TokenBucket,
		KeyGenerator: func(r *http.Request) string {
			// Rate limit by IP address
			return r.RemoteAddr
		},
		OnLimitExceeded: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests, please try again later",
				"retry_after": 60,
			})
		},
	}

	// Add rate limiting middleware
	chain.Add(ratelimit.NewMiddleware(tokenBucketConfig))

	// Setup routes
	mux := http.NewServeMux()

	// Health endpoint (not rate limited)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API endpoint with rate limiting
	apiHandler := chain.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":   "Hello from rate-limited API!",
			"timestamp": time.Now().Format(time.RFC3339),
			"method":    r.Method,
			"path":      r.URL.Path,
			"client_ip": r.RemoteAddr,
		}

		json.NewEncoder(w).Encode(response)
	}))

	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))

	// Create another endpoint with different rate limiting
	slidingWindowConfig := ratelimit.Config{
		Enabled:   true,
		Limit:     5, // 5 requests per minute
		Window:    time.Minute,
		Algorithm: ratelimit.SlidingWindow,
		KeyGenerator: func(r *http.Request) string {
			// Rate limit by user (if authenticated) or IP
			userID := r.Header.Get("X-User-ID")
			if userID != "" {
				return "user:" + userID
			}
			return "ip:" + r.RemoteAddr
		},
	}

	strictChain := middleware.NewChain()
	strictChain.Add(ratelimit.NewMiddleware(slidingWindowConfig))

	strictHandler := strictChain.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":    "This is a strictly rate-limited endpoint",
			"timestamp":  time.Now().Format(time.RFC3339),
			"rate_limit": "5 requests per minute",
			"algorithm":  "sliding_window",
		}

		json.NewEncoder(w).Encode(response)
	}))

	mux.Handle("/strict", strictHandler)

	fmt.Println("Rate limiting server starting on :8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /health       - Health check (no rate limit)")
	fmt.Println("  GET /api/test     - API endpoint (10 req/min, token bucket)")
	fmt.Println("  GET /strict       - Strict endpoint (5 req/min, sliding window)")
	fmt.Println("")
	fmt.Println("Test with:")
	fmt.Println("  curl http://localhost:8080/api/test")
	fmt.Println("  curl http://localhost:8080/strict")
	fmt.Println("  curl -H 'X-User-ID: user123' http://localhost:8080/strict")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
