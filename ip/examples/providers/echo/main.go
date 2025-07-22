// Package main demonstrates the IP library with Echo framework
package main

import (
	"fmt"
	"net/http"

	"github.com/fsvxavier/nexs-lib/ip"
)

// Since we can't import labstack/echo without adding it as a dependency,
// this example shows how the IP library would be used with Echo.
// To run with real Echo, uncomment the import and main function at the bottom.

// simulateEchoRequest simulates an Echo request for demonstration
func simulateEchoRequest() {
	fmt.Println("🚀 Echo IP Library Example (Simulation)")
	fmt.Println("=======================================")

	// Show supported frameworks
	frameworks := ip.GetSupportedFrameworks()
	fmt.Printf("📦 Supported Frameworks: %v\n", frameworks)

	fmt.Println("\n💡 This is a simulation showing how to use the IP library with Echo.")
	fmt.Println("📋 To use with real Echo, see the commented code at the bottom of this file.")

	// Simulate different scenarios using standard http.Request
	scenarios := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
	}{
		{
			name: "AWS ALB Request",
			headers: map[string]string{
				"X-Forwarded-For":   "198.51.100.100, 172.31.0.5",
				"X-Forwarded-Proto": "https",
				"X-Real-IP":         "198.51.100.100",
			},
			remoteAddr: "172.31.0.5:8080",
		},
		{
			name: "Nginx Proxy Request",
			headers: map[string]string{
				"X-Real-IP":       "203.0.113.195",
				"X-Forwarded-For": "203.0.113.195",
			},
			remoteAddr: "192.168.1.1:80",
		},
		{
			name: "Internal Service Request",
			headers: map[string]string{
				"X-Internal-Service": "true",
			},
			remoteAddr: "172.16.0.100:8080",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n🔍 Scenario %d: %s\n", i+1, scenario.name)
		fmt.Println("─────────────────────────────────────")

		// Create mock request (this would be c.Request() in real Echo)
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

		// Simulate geolocation logic
		if ipInfo != nil && ipInfo.IsPublic {
			fmt.Printf("   🌍 Geolocation: Enabled (public IP)\n")
		} else {
			fmt.Printf("   🌍 Geolocation: Disabled (private/local IP)\n")
		}
	}

	fmt.Println("\n📋 Example Echo Middleware Implementation:")
	fmt.Println("=========================================")
	fmt.Printf(`
func IPMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Extract real client IP using our library
            clientIP := ip.GetRealIP(c.Request())
            ipInfo := ip.GetRealIPInfo(c.Request())
            ipChain := ip.GetIPChain(c.Request())

            // Store in Echo context
            c.Set("clientIP", clientIP)
            c.Set("ipInfo", ipInfo)
            c.Set("ipChain", ipChain)

            // Log IP information
            log.Printf("Client IP: %%s, Type: %%s", clientIP, ipInfo.Type.String())

            return next(c)
        }
    }
}

func handleAPI(c echo.Context) error {
    // Get IP information from context
    clientIP := c.Get("clientIP").(string)
    ipInfo := c.Get("ipInfo").(*ip.IPInfo)

    // Security check
    if ipInfo.IsPrivate {
        return c.JSON(http.StatusForbidden, map[string]interface{}{
            "error":    "Access denied for private IP",
            "clientIP": clientIP,
        })
    }

    // Process request
    return c.JSON(http.StatusOK, map[string]interface{}{
        "clientIP": clientIP,
        "ipType":   ipInfo.Type.String(),
        "isPublic": ipInfo.IsPublic,
        "data":     []string{"item1", "item2"},
    })
}
`)
}

func main() {
	simulateEchoRequest()
}

/*
// Real Echo implementation (uncomment and add echo dependency to use):

package main

import (
    "log"
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/fsvxavier/nexs-lib/ip"
)

func main() {
    e := echo.New()

    // Built-in middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // IP extraction middleware
    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            clientIP := ip.GetRealIP(c.Request())
            ipInfo := ip.GetRealIPInfo(c.Request())
            ipChain := ip.GetIPChain(c.Request())

            // Store in context
            c.Set("clientIP", clientIP)
            c.Set("ipInfo", ipInfo)
            c.Set("ipChain", ipChain)

            // Log
            log.Printf("Client IP: %s, Type: %s", clientIP, ipInfo.Type.String())

            return next(c)
        }
    })

    // Basic endpoint
    e.GET("/", func(c echo.Context) error {
        clientIP := c.Get("clientIP").(string)
        ipInfo := c.Get("ipInfo").(*ip.IPInfo)

        return c.JSON(http.StatusOK, map[string]interface{}{
            "clientIP":  clientIP,
            "ipType":    ipInfo.Type.String(),
            "isPublic":  ipInfo.IsPublic,
            "framework": "echo",
        })
    })

    // API endpoint with IP validation
    e.GET("/api/data", func(c echo.Context) error {
        clientIP := c.Get("clientIP").(string)
        ipInfo := c.Get("ipInfo").(*ip.IPInfo)

        // Block private IPs
        if ipInfo.IsPrivate {
            return c.JSON(http.StatusForbidden, map[string]interface{}{
                "error":    "Access denied",
                "clientIP": clientIP,
                "reason":   "Private IP not allowed",
            })
        }

        return c.JSON(http.StatusOK, map[string]interface{}{
            "clientIP": clientIP,
            "data":     []string{"item1", "item2", "item3"},
        })
    })

    // Health check
    e.GET("/health", func(c echo.Context) error {
        clientIP := c.Get("clientIP").(string)
        return c.JSON(http.StatusOK, map[string]interface{}{
            "status":   "healthy",
            "clientIP": clientIP,
        })
    })

    log.Println("Echo server starting on :1323")
    e.Logger.Fatal(e.Start(":1323"))
}
*/
