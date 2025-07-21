// Package main demonstrates CORS middleware usage.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/middleware"
	"github.com/fsvxavier/nexs-lib/httpserver/middleware/cors"
)

func main() {
	// Create middleware chain
	chain := middleware.NewChain()

	// Configure CORS middleware
	corsConfig := cors.Config{
		Enabled:   true,
		SkipPaths: []string{"/health"}, // Health endpoint doesn't need CORS
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8000",
			"https://mydomain.com",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-User-ID",
		},
		ExposedHeaders: []string{
			"X-Total-Count",
			"X-Rate-Limit-Remaining",
		},
		AllowCredentials:   true,
		MaxAge:             12 * time.Hour, // Cache preflight for 12 hours
		OptionsPassthrough: false,          // Handle OPTIONS requests
	}

	// Add CORS middleware
	chain.Add(cors.NewMiddleware(corsConfig))

	// Setup routes
	mux := http.NewServeMux()

	// Health endpoint (no CORS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API endpoints with CORS
	apiHandler := chain.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add custom headers that will be exposed
		w.Header().Set("X-Total-Count", "100")
		w.Header().Set("X-Rate-Limit-Remaining", "50")
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			response := map[string]interface{}{
				"message": "GET request successful",
				"origin":  r.Header.Get("Origin"),
				"method":  r.Method,
				"path":    r.URL.Path,
				"headers": getRequestHeaders(r),
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)

		case http.MethodPost:
			response := map[string]interface{}{
				"message": "POST request successful",
				"origin":  r.Header.Get("Origin"),
				"method":  r.Method,
				"path":    r.URL.Path,
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)

		case http.MethodPut:
			response := map[string]interface{}{
				"message": "PUT request successful",
				"origin":  r.Header.Get("Origin"),
				"method":  r.Method,
				"path":    r.URL.Path,
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)

		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			// No response body for DELETE

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Method not allowed",
			})
		}
	}))

	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))

	// Create a separate endpoint with different CORS settings (allow all origins)
	publicCorsConfig := cors.Config{
		Enabled:        true,
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
		},
		AllowedHeaders: []string{
			"Content-Type",
		},
		AllowCredentials: false, // Can't use credentials with wildcard origin
		MaxAge:           1 * time.Hour,
	}

	publicChain := middleware.NewChain()
	publicChain.Add(cors.NewMiddleware(publicCorsConfig))

	publicHandler := publicChain.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":    "Public endpoint with open CORS policy",
			"timestamp":  time.Now().Format(time.RFC3339),
			"origin":     r.Header.Get("Origin"),
			"user_agent": r.Header.Get("User-Agent"),
		}

		json.NewEncoder(w).Encode(response)
	}))

	mux.Handle("/public", publicHandler)

	fmt.Println("CORS server starting on :8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /health      - Health check (no CORS)")
	fmt.Println("  GET  /api/test    - API endpoint (restricted CORS)")
	fmt.Println("  POST /api/test    - API endpoint (restricted CORS)")
	fmt.Println("  PUT  /api/test    - API endpoint (restricted CORS)")
	fmt.Println("  DEL  /api/test    - API endpoint (restricted CORS)")
	fmt.Println("  GET  /public      - Public endpoint (open CORS)")
	fmt.Println("")
	fmt.Println("Test with different origins:")
	fmt.Println("  curl -H 'Origin: http://localhost:3000' http://localhost:8080/api/test")
	fmt.Println("  curl -H 'Origin: https://unauthorized.com' http://localhost:8080/api/test")
	fmt.Println("  curl -H 'Origin: https://unauthorized.com' http://localhost:8080/public")
	fmt.Println("")
	fmt.Println("Test preflight request:")
	fmt.Println("  curl -X OPTIONS -H 'Origin: http://localhost:3000' \\")
	fmt.Println("       -H 'Access-Control-Request-Method: POST' \\")
	fmt.Println("       -H 'Access-Control-Request-Headers: Content-Type' \\")
	fmt.Println("       http://localhost:8080/api/test")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// getRequestHeaders extracts relevant headers from the request.
func getRequestHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	relevantHeaders := []string{
		"Origin",
		"Referer",
		"User-Agent",
		"Authorization",
		"X-Requested-With",
		"X-User-ID",
	}

	for _, header := range relevantHeaders {
		if value := r.Header.Get(header); value != "" {
			headers[header] = value
		}
	}

	return headers
}
