// Package main demonstrates the IP library with Gin framework
package main

import (
	"fmt"
	"net/http"

	"github.com/fsvxavier/nexs-lib/ip"
)

// Since we can't import gin-gonic/gin without adding it as a dependency,
// this example shows how the IP library would be used with Gin.
// To run with real Gin, uncomment the import and main function at the bottom.

// simulateGinRequest simulates a Gin request for demonstration
func simulateGinRequest() {
	fmt.Println("🚀 Gin IP Library Example (Simulation)")
	fmt.Println("=====================================")

	// Show supported frameworks
	frameworks := ip.GetSupportedFrameworks()
	fmt.Printf("📦 Supported Frameworks: %v\n", frameworks)

	fmt.Println("\n💡 This is a simulation showing how to use the IP library with Gin.")
	fmt.Println("📋 To use with real Gin, see the commented code at the bottom of this file.")

	// Simulate different scenarios using standard http.Request
	scenarios := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
	}{
		{
			name: "Cloudflare CDN Request",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.45",
				"X-Forwarded-For":  "203.0.113.45, 198.51.100.10",
			},
			remoteAddr: "198.51.100.10:443",
		},
		{
			name: "Load Balancer Request",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.100, 172.31.0.5",
				"X-Real-IP":       "198.51.100.100",
			},
			remoteAddr: "172.31.0.5:8080",
		},
		{
			name: "Private Network Request",
			headers: map[string]string{
				"X-Real-IP": "10.0.0.1",
			},
			remoteAddr: "10.0.0.1:3000",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n🔍 Scenario %d: %s\n", i+1, scenario.name)
		fmt.Println("─────────────────────────────────────")

		// Create mock request (this would be c.Request in real Gin)
		req := &http.Request{
			Header:     make(http.Header),
			RemoteAddr: scenario.remoteAddr,
		}
		for key, value := range scenario.headers {
			req.Header.Set(key, value)
		}

		// Extract IP information using our library
		clientIP := ip.GetRealIP(req)
		ipInfo := ip.GetRealIPInfo(req)
		ipChain := ip.GetIPChain(req)

		fmt.Printf("   Client IP: %s\n", clientIP)
		if ipInfo != nil {
			fmt.Printf("   IP Type: %s\n", ipInfo.Type.String())
			fmt.Printf("   Is Public: %v | Is Private: %v\n", ipInfo.IsPublic, ipInfo.IsPrivate)
			fmt.Printf("   Source: %s\n", ipInfo.Source)
		}
		fmt.Printf("   IP Chain: %v\n", ipChain)

		// Simulate middleware logic
		if ipInfo != nil && ipInfo.IsPrivate {
			fmt.Printf("   🚫 Security: Private IP detected - would restrict access\n")
		} else {
			fmt.Printf("   ✅ Security: Public IP - access granted\n")
		}
	}

	fmt.Println("\n📋 Example Gin Middleware Implementation:")
	fmt.Println("========================================")
	fmt.Printf(`
func IPMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract real client IP using our library
        clientIP := ip.GetRealIP(c.Request)
        ipInfo := ip.GetRealIPInfo(c.Request)
        ipChain := ip.GetIPChain(c.Request)

        // Store in Gin context
        c.Set("clientIP", clientIP)
        c.Set("ipInfo", ipInfo)
        c.Set("ipChain", ipChain)

        // Log IP information
        log.Printf("Client IP: %%s, Type: %%s", clientIP, ipInfo.Type.String())

        c.Next()
    }
}

func handleAPI(c *gin.Context) {
    // Get IP information from context
    clientIP := c.GetString("clientIP")
    ipInfoInterface, _ := c.Get("ipInfo")
    ipInfo := ipInfoInterface.(*ip.IPInfo)

    // Security check
    if ipInfo.IsPrivate {
        c.JSON(403, gin.H{
            "error": "Access denied for private IP",
            "clientIP": clientIP,
        })
        return
    }

    // Process request
    c.JSON(200, gin.H{
        "clientIP": clientIP,
        "ipType":   ipInfo.Type.String(),
        "isPublic": ipInfo.IsPublic,
        "data":     []string{"item1", "item2"},
    })
}
`)
}

func main() {
	simulateGinRequest()
}

/*
// Real Gin implementation (uncomment and add gin dependency to use):

package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "github.com/fsvxavier/nexs-lib/ip"
)

func main() {
    r := gin.Default()

    // IP extraction middleware
    r.Use(func(c *gin.Context) {
        clientIP := ip.GetRealIP(c.Request)
        ipInfo := ip.GetRealIPInfo(c.Request)
        ipChain := ip.GetIPChain(c.Request)

        // Store in context
        c.Set("clientIP", clientIP)
        c.Set("ipInfo", ipInfo)
        c.Set("ipChain", ipChain)

        // Log
        log.Printf("Client IP: %s, Type: %s", clientIP, ipInfo.Type.String())

        c.Next()
    })

    // Basic endpoint
    r.GET("/", func(c *gin.Context) {
        clientIP := c.GetString("clientIP")
        ipInfoInterface, _ := c.Get("ipInfo")
        ipInfo := ipInfoInterface.(*ip.IPInfo)

        c.JSON(200, gin.H{
            "clientIP":  clientIP,
            "ipType":    ipInfo.Type.String(),
            "isPublic":  ipInfo.IsPublic,
            "framework": "gin",
        })
    })

    // API endpoint with IP validation
    r.GET("/api/data", func(c *gin.Context) {
        clientIP := c.GetString("clientIP")
        ipInfoInterface, _ := c.Get("ipInfo")
        ipInfo := ipInfoInterface.(*ip.IPInfo)

        // Block private IPs
        if ipInfo.IsPrivate {
            c.JSON(403, gin.H{
                "error":    "Access denied",
                "clientIP": clientIP,
                "reason":   "Private IP not allowed",
            })
            return
        }

        c.JSON(200, gin.H{
            "clientIP": clientIP,
            "data":     []string{"item1", "item2", "item3"},
        })
    })

    // Health check
    r.GET("/health", func(c *gin.Context) {
        clientIP := c.GetString("clientIP")
        c.JSON(200, gin.H{
            "status":   "healthy",
            "clientIP": clientIP,
        })
    })

    log.Println("Gin server starting on :8080")
    r.Run(":8080")
}
*/
