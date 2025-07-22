package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/ip"
)

// IPLoggingMiddleware logs requests with real client IP information
func IPLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extract real IP information
		ipInfo := ip.GetRealIPInfo(r)
		realIP := "unknown"
		ipType := "unknown"

		if ipInfo != nil {
			realIP = ipInfo.IP.String()
			ipType = ipInfo.Type.String()
		}

		// Create a wrapped response writer to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(wrappedWriter, r)

		// Log request information
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %d %v - Real IP: %s (%s) - Remote: %s - XFF: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			wrappedWriter.statusCode,
			duration,
			realIP,
			ipType,
			r.RemoteAddr,
			r.Header.Get("X-Forwarded-For"),
		)
	})
}

// SecurityMiddleware adds security headers and blocks suspicious IPs
func SecurityMiddleware(next http.Handler) http.Handler {
	// Example blocked IPs (in production, this might come from a database or external service)
	blockedIPs := map[string]bool{
		"192.0.2.1":   true, // Example blocked IP
		"203.0.113.1": true, // Another example
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipInfo := ip.GetRealIPInfo(r)
		if ipInfo == nil {
			http.Error(w, "Unable to determine client IP", http.StatusBadRequest)
			return
		}

		realIP := ipInfo.IP.String()

		// Check if IP is blocked
		if blockedIPs[realIP] {
			log.Printf("Blocked request from IP: %s", realIP)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Add client IP information to response headers (for debugging)
		w.Header().Set("X-Real-Client-IP", realIP)
		w.Header().Set("X-Client-IP-Type", ipInfo.Type.String())

		next.ServeHTTP(w, r)
	})
}

// GeoLocationMiddleware simulates adding geographic information based on IP
func GeoLocationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipInfo := ip.GetRealIPInfo(r)
		if ipInfo != nil {
			realIP := ipInfo.IP.String()

			// Simulate geo-location lookup (in production, use a real service)
			var country, city string
			if ipInfo.IsPrivate {
				country = "Private Network"
				city = "Internal"
			} else if ipInfo.Type == ip.IPTypeLoopback {
				country = "Localhost"
				city = "Local"
			} else {
				// This is a simulation - in production you'd use a real geo-IP service
				switch {
				case realIP == "8.8.8.8":
					country = "United States"
					city = "Mountain View"
				case realIP == "1.1.1.1":
					country = "Australia"
					city = "Sydney"
				default:
					country = "Unknown"
					city = "Unknown"
				}
			}

			// Add geo information to request context or headers
			w.Header().Set("X-Client-Country", country)
			w.Header().Set("X-Client-City", city)

			log.Printf("IP %s geolocated to: %s, %s", realIP, city, country)
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware implements simple rate limiting based on client IP
func RateLimitMiddleware(next http.Handler) http.Handler {
	// Simple in-memory rate limiter (in production, use Redis or similar)
	requests := make(map[string][]time.Time)
	const maxRequests = 10
	const timeWindow = time.Minute

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipInfo := ip.GetRealIPInfo(r)
		if ipInfo == nil {
			http.Error(w, "Unable to determine client IP", http.StatusBadRequest)
			return
		}

		realIP := ipInfo.IP.String()
		now := time.Now()

		// Clean old requests
		if reqTimes, exists := requests[realIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range reqTimes {
				if now.Sub(reqTime) < timeWindow {
					validRequests = append(validRequests, reqTime)
				}
			}
			requests[realIP] = validRequests
		}

		// Check rate limit
		if len(requests[realIP]) >= maxRequests {
			log.Printf("Rate limit exceeded for IP: %s", realIP)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
			w.Header().Set("X-RateLimit-Window", timeWindow.String())
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Add current request
		requests[realIP] = append(requests[realIP], now)

		// Add rate limit headers
		remaining := maxRequests - len(requests[realIP])
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		w.Header().Set("X-RateLimit-Window", timeWindow.String())
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// API handlers
func homeHandler(w http.ResponseWriter, r *http.Request) {
	ipInfo := ip.GetRealIPInfo(r)

	response := fmt.Sprintf(`
Welcome to the IP Middleware Demo!

Your Connection Information:
- Real IP: %s
- IP Type: %s
- Is Public: %v
- Is IPv4: %v
- Remote Address: %s

Proxy Chain:
`,
		ipInfo.IP.String(),
		ipInfo.Type.String(),
		ipInfo.IsPublic,
		ipInfo.IsIPv4,
		r.RemoteAddr,
	)

	chain := ip.GetIPChain(r)
	for i, chainIP := range chain {
		chainInfo := ip.ParseIP(chainIP)
		if chainInfo != nil {
			response += fmt.Sprintf("  %d. %s (%s)\n", i+1, chainIP, chainInfo.Type.String())
		} else {
			response += fmt.Sprintf("  %d. %s (invalid)\n", i+1, chainIP)
		}
	}

	response += fmt.Sprintf(`
Request Headers:
- X-Forwarded-For: %s
- X-Real-IP: %s
- CF-Connecting-IP: %s
- Forwarded: %s

Rate Limiting:
- Limit: %s
- Window: %s  
- Remaining: %s

Geo Location:
- Country: %s
- City: %s
`,
		r.Header.Get("X-Forwarded-For"),
		r.Header.Get("X-Real-IP"),
		r.Header.Get("CF-Connecting-IP"),
		r.Header.Get("Forwarded"),
		w.Header().Get("X-RateLimit-Limit"),
		w.Header().Get("X-RateLimit-Window"),
		w.Header().Get("X-RateLimit-Remaining"),
		w.Header().Get("X-Client-Country"),
		w.Header().Get("X-Client-City"),
	)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	ipInfo := ip.GetRealIPInfo(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON response (in production, use json.Marshal)
	fmt.Fprintf(w, `{
  "success": true,
  "client_ip": {
    "address": "%s",
    "type": "%s",
    "is_public": %v,
    "is_ipv4": %v
  },
  "rate_limit": {
    "limit": "%s",
    "remaining": "%s"
  }
}`,
		ipInfo.IP.String(),
		ipInfo.Type.String(),
		ipInfo.IsPublic,
		ipInfo.IsIPv4,
		w.Header().Get("X-RateLimit-Limit"),
		w.Header().Get("X-RateLimit-Remaining"),
	)
}

func main() {
	// Create a new mux
	mux := http.NewServeMux()

	// Add routes
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/info", apiHandler)

	// Build middleware chain
	var handler http.Handler = mux

	// Apply middleware in reverse order (last applied = first executed)
	handler = RateLimitMiddleware(handler)
	handler = GeoLocationMiddleware(handler)
	handler = SecurityMiddleware(handler)
	handler = IPLoggingMiddleware(handler)

	fmt.Println("Starting HTTP server with IP middleware on :8080...")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET / - Home page with detailed IP information")
	fmt.Println("  GET /api/info - JSON API with IP data")
	fmt.Println("\nMiddleware features:")
	fmt.Println("  ✓ Real IP extraction and logging")
	fmt.Println("  ✓ Security headers and IP blocking")
	fmt.Println("  ✓ Simulated geo-location")
	fmt.Println("  ✓ Rate limiting (10 requests/minute per IP)")
	fmt.Println("\nTest with different headers:")
	fmt.Println("  curl -H 'X-Forwarded-For: 8.8.8.8' http://localhost:8080")
	fmt.Println("  curl -H 'CF-Connecting-IP: 1.1.1.1' http://localhost:8080/api/info")
	fmt.Println("\nPress Ctrl+C to stop the server")

	log.Fatal(http.ListenAndServe(":8080", handler))
}
