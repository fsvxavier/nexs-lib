// Package main demonstrates the IP library with net/http framework
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/ip"
)

// IPResponse represents the JSON response structure
type IPResponse struct {
	ClientIP   string   `json:"clientIP"`
	IPType     string   `json:"ipType"`
	IsPublic   bool     `json:"isPublic"`
	IsPrivate  bool     `json:"isPrivate"`
	IsIPv4     bool     `json:"isIPv4"`
	IsIPv6     bool     `json:"isIPv6"`
	Source     string   `json:"source"`
	IPChain    []string `json:"ipChain"`
	RemoteAddr string   `json:"remoteAddr"`
	Framework  string   `json:"framework"`
}

// IPMiddleware extracts and logs client IP information
func IPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := ip.GetRealIP(r)
		ipInfo := ip.GetRealIPInfo(r)
		ipChain := ip.GetIPChain(r)

		log.Printf("[IP Middleware] Client: %s, Type: %s, Chain: %v",
			clientIP,
			func() string {
				if ipInfo != nil {
					return ipInfo.Type.String()
				} else {
					return "unknown"
				}
			}(),
			ipChain,
		)

		// Call next handler
		next(w, r)
	}
}

// handleRoot demonstrates basic IP extraction
func handleRoot(w http.ResponseWriter, r *http.Request) {
	clientIP := ip.GetRealIP(r)
	ipInfo := ip.GetRealIPInfo(r)
	ipChain := ip.GetIPChain(r)

	response := IPResponse{
		ClientIP:   clientIP,
		IPChain:    ipChain,
		RemoteAddr: r.RemoteAddr,
		Framework:  "net/http",
	}

	if ipInfo != nil {
		response.IPType = ipInfo.Type.String()
		response.IsPublic = ipInfo.IsPublic
		response.IsPrivate = ipInfo.IsPrivate
		response.IsIPv4 = ipInfo.IsIPv4
		response.IsIPv6 = ipInfo.IsIPv6
		response.Source = ipInfo.Source
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAPI demonstrates API endpoint with IP validation
func handleAPI(w http.ResponseWriter, r *http.Request) {
	clientIP := ip.GetRealIP(r)
	ipInfo := ip.GetRealIPInfo(r)

	// Security check: block private IPs for public API
	if ipInfo != nil && ipInfo.IsPrivate {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":     "Access denied",
			"reason":    "Private IP addresses not allowed for public API",
			"clientIP":  clientIP,
			"framework": "net/http",
		})
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "API access granted",
		"clientIP":  clientIP,
		"timestamp": "2025-07-22T00:00:00Z",
		"data":      []string{"item1", "item2", "item3"},
		"framework": "net/http",
	})
}

// handleHealth demonstrates health check with IP logging
func handleHealth(w http.ResponseWriter, r *http.Request) {
	clientIP := ip.GetRealIP(r)
	ipInfo := ip.GetRealIPInfo(r)

	status := "healthy"
	if ipInfo != nil {
		if ipInfo.IsPrivate {
			status = "healthy-internal"
		} else if ipInfo.IsPublic {
			status = "healthy-external"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    status,
		"clientIP":  clientIP,
		"framework": "net/http",
		"version":   "1.0.0",
	})
}

// handleAdmin demonstrates admin endpoint with strict IP checking
func handleAdmin(w http.ResponseWriter, r *http.Request) {
	clientIP := ip.GetRealIP(r)
	ipInfo := ip.GetRealIPInfo(r)

	// Only allow specific IPs for admin
	allowedIPs := []string{"127.0.0.1", "::1", "192.168.1.100"}
	allowed := false
	for _, allowedIP := range allowedIPs {
		if clientIP == allowedIP {
			allowed = true
			break
		}
	}

	if !allowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":     "Unauthorized",
			"reason":    "IP not in whitelist",
			"clientIP":  clientIP,
			"framework": "net/http",
		})
		return
	}

	// Admin access granted
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Admin access granted",
		"clientIP": clientIP,
		"ipType": func() string {
			if ipInfo != nil {
				return ipInfo.Type.String()
			} else {
				return "unknown"
			}
		}(),
		"adminLevel": "full",
		"framework":  "net/http",
	})
}

func main() {
	fmt.Println("🚀 Net/HTTP IP Library Example")
	fmt.Println("==============================")

	// Show supported frameworks
	frameworks := ip.GetSupportedFrameworks()
	fmt.Printf("📦 Supported Frameworks: %v\n", frameworks)

	// Setup routes with middleware
	http.HandleFunc("/", IPMiddleware(handleRoot))
	http.HandleFunc("/api/data", IPMiddleware(handleAPI))
	http.HandleFunc("/health", IPMiddleware(handleHealth))
	http.HandleFunc("/admin", IPMiddleware(handleAdmin))

	fmt.Println("\n📡 Server starting on http://localhost:8080")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("   GET  /           - Basic IP information")
	fmt.Println("   GET  /api/data   - API with IP validation")
	fmt.Println("   GET  /health     - Health check")
	fmt.Println("   GET  /admin      - Admin endpoint (IP whitelist)")

	fmt.Println("\n🔧 Test with different headers:")
	fmt.Println("   curl http://localhost:8080")
	fmt.Println("   curl -H 'X-Forwarded-For: 8.8.8.8' http://localhost:8080")
	fmt.Println("   curl -H 'CF-Connecting-IP: 203.0.113.100' http://localhost:8080/api/data")
	fmt.Println("   curl -H 'X-Real-IP: 192.168.1.1' http://localhost:8080/api/data")
	fmt.Println("   curl http://localhost:8080/admin")

	fmt.Println("\n✨ The same IP library works across ALL supported frameworks!")

	// Auto-stop server after 3 seconds for demonstration
	fmt.Println("\n� Server will stop automatically in 3 seconds for demonstration...")

	server := &http.Server{Addr: ":8080"}
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("✅ Net/HTTP provider example completed successfully - Server stopped automatically")
		server.Shutdown(context.Background())
	}()

	server.ListenAndServe()
}
