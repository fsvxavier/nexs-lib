// Package main demonstrates the IP library with Atreugo framework
package main

import (
	"fmt"
	"net/http"

	"github.com/fsvxavier/nexs-lib/ip"
)

// simulateAtreugoRequest simulates an Atreugo request for demonstration
func simulateAtreugoRequest() {
	fmt.Println("🚀 Atreugo IP Library Example")
	fmt.Println("=============================")

	// Show supported frameworks
	frameworks := ip.GetSupportedFrameworks()
	fmt.Printf("📦 Supported Frameworks: %v\n", frameworks)

	fmt.Println("\n💡 This example shows how to use the IP library with Atreugo.")
	fmt.Println("📋 For real Atreugo usage, see the commented code at the bottom.")

	// Simulate different scenarios
	scenarios := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
	}{
		{
			name: "High-Performance REST API",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.45, 198.51.100.10",
				"Accept":          "application/json",
				"Content-Type":    "application/json",
			},
			remoteAddr: "198.51.100.10:443",
		},
		{
			name: "GraphQL Endpoint",
			headers: map[string]string{
				"X-Real-IP":     "198.51.100.42",
				"Content-Type":  "application/graphql",
				"Authorization": "Bearer token123",
			},
			remoteAddr: "172.16.0.1:8080",
		},
		{
			name: "WebSocket Connection",
			headers: map[string]string{
				"CF-Connecting-IP":  "203.0.113.100",
				"Upgrade":           "websocket",
				"Connection":        "Upgrade",
				"Sec-WebSocket-Key": "x3JJHMbDL1EzLkh9GBhXDw==",
			},
			remoteAddr: "172.17.0.5:80",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n🔍 Scenario %d: %s\n", i+1, scenario.name)
		fmt.Println("─────────────────────────────────────")

		// Create mock request (this would be ctx.Request in real Atreugo)
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

		// Simulate routing decisions based on IP
		if ipInfo != nil && ipInfo.IsPrivate {
			fmt.Printf("   🛣️  Routing: Internal service path\n")
		} else {
			fmt.Printf("   🛣️  Routing: Public API path\n")
		}
	}

	fmt.Println("\n📋 Atreugo Usage Example:")
	fmt.Println("=========================")
	fmt.Print(`
// Real Atreugo implementation:
func atreugoHandler(ctx *atreugo.RequestCtx) error {
    clientIP := ip.GetRealIP(ctx)
    ipInfo := ip.GetRealIPInfo(ctx)

    if ipInfo.IsPrivate {
        return ctx.JSONResponse(map[string]interface{}{
            "error": "Access denied",
            "clientIP": clientIP,
        }, atreugo.StatusForbidden)
    }

    return ctx.JSONResponse(map[string]interface{}{
        "clientIP":  clientIP,
        "ipType":    ipInfo.Type.String(),
        "isPublic":  ipInfo.IsPublic,
        "framework": "atreugo",
    }, atreugo.StatusOK)
}

func main() {
    config := atreugo.Config{
        Host: "0.0.0.0",
        Port: 8080,
    }
    
    server := atreugo.New(config)
    server.GET("/", atreugoHandler)
    server.GET("/health", healthHandler)

    log.Fatal(server.ListenAndServe())
}
`)
}

func main() {
	simulateAtreugoRequest()
}

/*
// Real Atreugo implementation (add atreugo dependency to use):

package main

import (
    "log"
    "github.com/savsgio/atreugo/v11"
    "github.com/fsvxavier/nexs-lib/ip"
)

func main() {
    config := atreugo.Config{
        Host: "0.0.0.0",
        Port: 8080,
    }

    server := atreugo.New(config)

    // IP extraction middleware
    server.UseBefore(func(ctx *atreugo.RequestCtx) error {
        clientIP := ip.GetRealIP(ctx)
        ipInfo := ip.GetRealIPInfo(ctx)
        ipChain := ip.GetIPChain(ctx)

        // Store in context
        ctx.SetUserValue("clientIP", clientIP)
        ctx.SetUserValue("ipInfo", ipInfo)
        ctx.SetUserValue("ipChain", ipChain)

        // Log
        log.Printf("Client IP: %s, Type: %s", clientIP, ipInfo.Type.String())

        return ctx.Next()
    })

    // Basic endpoint
    server.GET("/", func(ctx *atreugo.RequestCtx) error {
        clientIP := ctx.UserValue("clientIP").(string)
        ipInfo := ctx.UserValue("ipInfo").(*ip.IPInfo)

        return ctx.JSONResponse(map[string]interface{}{
            "clientIP":  clientIP,
            "ipType":    ipInfo.Type.String(),
            "isPublic":  ipInfo.IsPublic,
            "framework": "atreugo",
        }, atreugo.StatusOK)
    })

    // API endpoint with IP validation
    server.GET("/api/data", func(ctx *atreugo.RequestCtx) error {
        clientIP := ctx.UserValue("clientIP").(string)
        ipInfo := ctx.UserValue("ipInfo").(*ip.IPInfo)

        // Block private IPs
        if ipInfo.IsPrivate {
            return ctx.JSONResponse(map[string]interface{}{
                "error":    "Access denied",
                "clientIP": clientIP,
                "reason":   "Private IP not allowed",
            }, atreugo.StatusForbidden)
        }

        return ctx.JSONResponse(map[string]interface{}{
            "clientIP": clientIP,
            "data":     []string{"item1", "item2", "item3"},
        }, atreugo.StatusOK)
    })

    // Health check
    server.GET("/health", func(ctx *atreugo.RequestCtx) error {
        clientIP := ctx.UserValue("clientIP").(string)
        return ctx.JSONResponse(map[string]interface{}{
            "status":   "healthy",
            "clientIP": clientIP,
        }, atreugo.StatusOK)
    })

    log.Println("Atreugo server starting on :8080")
    log.Fatal(server.ListenAndServe())
}
*/
