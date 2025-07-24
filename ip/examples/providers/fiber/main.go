// Package main demonstrates the IP library with Fiber framework
package main

import (
	"fmt"
	"net/http"

	"github.com/fsvxavier/nexs-lib/ip"
)

// Since we can't import gofiber/fiber without adding it as a dependency,
// this example shows how the IP library would be used with Fiber.
// To run with real Fiber, uncomment the import and main function at the bottom.

// simulateFiberRequest simulates a Fiber request for demonstration
func simulateFiberRequest() {
	fmt.Println("🚀 Fiber IP Library Example (Simulation)")
	fmt.Println("========================================")

	// Show supported frameworks
	frameworks := ip.GetSupportedFrameworks()
	fmt.Printf("📦 Supported Frameworks: %v\n", frameworks)

	fmt.Println("\n💡 This is a simulation showing how to use the IP library with Fiber.")
	fmt.Println("📋 To use with real Fiber, see the commented code at the bottom of this file.")

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
			name:       "Direct Connection",
			headers:    map[string]string{},
			remoteAddr: "8.8.8.8:54321",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n🔍 Scenario %d: %s\n", i+1, scenario.name)
		fmt.Println("─────────────────────────────────────")

		// Create mock request (this would be c.Context() in real Fiber)
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

		// Simulate rate limiting based on IP
		if ipInfo != nil && ipInfo.IsPublic {
			fmt.Printf("   🚦 Rate Limiting: Applied (public IP)\n")
		} else {
			fmt.Printf("   🚦 Rate Limiting: Relaxed (private/local IP)\n")
		}
	}

	fmt.Println("\n📋 Example Fiber Middleware Implementation:")
	fmt.Println("==========================================")
	fmt.Printf(`
func IPMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Extract real client IP using our library
        clientIP := ip.GetRealIP(c.Context())
        ipInfo := ip.GetRealIPInfo(c.Context())
        ipChain := ip.GetIPChain(c.Context())

        // Store in Fiber context
        c.Locals("clientIP", clientIP)
        c.Locals("ipInfo", ipInfo)
        c.Locals("ipChain", ipChain)

        // Log IP information
        log.Printf("Client IP: %%s, Type: %%s", clientIP, ipInfo.Type.String())

        return c.Next()
    }
}

func handleAPI(c *fiber.Ctx) error {
    // Get IP information from context
    clientIP := c.Locals("clientIP").(string)
    ipInfo := c.Locals("ipInfo").(*ip.IPInfo)

    // Security check
    if ipInfo.IsPrivate {
        return c.Status(403).JSON(fiber.Map{
            "error":    "Access denied for private IP",
            "clientIP": clientIP,
        })
    }

    // Process request
    return c.JSON(fiber.Map{
        "clientIP": clientIP,
        "ipType":   ipInfo.Type.String(),
        "isPublic": ipInfo.IsPublic,
        "data":     []string{"item1", "item2"},
    })
}
`)
}

func main() {
	simulateFiberRequest()
}

/*
// Real Fiber implementation (uncomment and add fiber dependency to use):

package main

import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/fsvxavier/nexs-lib/ip"
)

func main() {
    app := fiber.New()

    // IP extraction middleware
    app.Use(func(c *fiber.Ctx) error {
        clientIP := ip.GetRealIP(c.Context())
        ipInfo := ip.GetRealIPInfo(c.Context())
        ipChain := ip.GetIPChain(c.Context())

        // Store in context
        c.Locals("clientIP", clientIP)
        c.Locals("ipInfo", ipInfo)
        c.Locals("ipChain", ipChain)

        // Log
        log.Printf("Client IP: %s, Type: %s", clientIP, ipInfo.Type.String())

        return c.Next()
    })

    // Basic endpoint
    app.Get("/", func(c *fiber.Ctx) error {
        clientIP := c.Locals("clientIP").(string)
        ipInfo := c.Locals("ipInfo").(*ip.IPInfo)

        return c.JSON(fiber.Map{
            "clientIP":  clientIP,
            "ipType":    ipInfo.Type.String(),
            "isPublic":  ipInfo.IsPublic,
            "framework": "fiber",
        })
    })

    // API endpoint with IP validation
    app.Get("/api/data", func(c *fiber.Ctx) error {
        clientIP := c.Locals("clientIP").(string)
        ipInfo := c.Locals("ipInfo").(*ip.IPInfo)

        // Block private IPs
        if ipInfo.IsPrivate {
            return c.Status(403).JSON(fiber.Map{
                "error":    "Access denied",
                "clientIP": clientIP,
                "reason":   "Private IP not allowed",
            })
        }

        return c.JSON(fiber.Map{
            "clientIP": clientIP,
            "data":     []string{"item1", "item2", "item3"},
        })
    })

    // Health check
    app.Get("/health", func(c *fiber.Ctx) error {
        clientIP := c.Locals("clientIP").(string)
        return c.JSON(fiber.Map{
            "status":   "healthy",
            "clientIP": clientIP,
        })
    })

    log.Println("Fiber server starting on :3000")
    log.Fatal(app.Listen(":3000"))
}
*/
