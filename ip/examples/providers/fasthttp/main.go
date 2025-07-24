// Package main demonstrates the IP library with FastHTTP framework
package main

import (
	"fmt"
	"net/http"

	"github.com/fsvxavier/nexs-lib/ip"
)

// simulateFastHTTPRequest simulates a FastHTTP request for demonstration
func simulateFastHTTPRequest() {
	fmt.Println("🚀 FastHTTP IP Library Example")
	fmt.Println("==============================")

	// Show supported frameworks
	frameworks := ip.GetSupportedFrameworks()
	fmt.Printf("📦 Supported Frameworks: %v\n", frameworks)

	fmt.Println("\n💡 This example shows how to use the IP library with FastHTTP.")
	fmt.Println("📋 For real FastHTTP usage, see the commented code at the bottom.")

	// Simulate different scenarios
	scenarios := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
	}{
		{
			name: "High-Performance API Request",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.45, 198.51.100.10",
				"User-Agent":      "High-Performance-Client/1.0",
			},
			remoteAddr: "198.51.100.10:443",
		},
		{
			name: "Microservice Communication",
			headers: map[string]string{
				"X-Service-Name": "user-service",
				"X-Real-IP":      "172.17.0.5",
			},
			remoteAddr: "172.17.0.5:8080",
		},
		{
			name: "WebSocket Upgrade Request",
			headers: map[string]string{
				"CF-Connecting-IP": "198.51.100.42",
				"Upgrade":          "websocket",
				"Connection":       "Upgrade",
			},
			remoteAddr: "172.16.0.1:80",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n🔍 Scenario %d: %s\n", i+1, scenario.name)
		fmt.Println("─────────────────────────────────────")

		// Create mock request (this would be ctx.Request in real FastHTTP)
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

		// Simulate performance optimizations
		if ipInfo != nil && ipInfo.IsPrivate {
			fmt.Printf("   ⚡ Performance: Internal request - skip CDN cache\n")
		} else {
			fmt.Printf("   ⚡ Performance: External request - enable CDN cache\n")
		}
	}

	fmt.Println("\n📋 FastHTTP Usage Example:")
	fmt.Println("==========================")
	codeExample := `
// Real FastHTTP implementation:
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
    clientIP := ip.GetRealIP(ctx)
    ipInfo := ip.GetRealIPInfo(ctx)

    // High-performance response
    ctx.SetContentType("application/json")
    fmt.Fprintf(ctx, ` + "`" + `{
        "clientIP": "%s",
        "ipType": "%s",
        "isPublic": %t,
        "framework": "fasthttp"
    }` + "`" + `, clientIP, ipInfo.Type.String(), ipInfo.IsPublic)
}

func main() {
    handler := func(ctx *fasthttp.RequestCtx) {
        switch string(ctx.Path()) {
        case "/":
            fastHTTPHandler(ctx)
        case "/health":
            healthHandler(ctx)
        }
    }

    log.Println("FastHTTP server starting on :8080")
    log.Fatal(fasthttp.ListenAndServe(":8080", handler))
}
`
	fmt.Print(codeExample)
}

func main() {
	simulateFastHTTPRequest()
}

/*
// Real FastHTTP implementation (add fasthttp dependency to use):

package main

import (
    "fmt"
    "log"
    "github.com/valyala/fasthttp"
    "github.com/fsvxavier/nexs-lib/ip"
)

func main() {
    handler := func(ctx *fasthttp.RequestCtx) {
        clientIP := ip.GetRealIP(ctx)
        ipInfo := ip.GetRealIPInfo(ctx)

        // Log for performance monitoring
        log.Printf("FastHTTP - Client IP: %s, Type: %s", clientIP, ipInfo.Type.String())

        switch string(ctx.Path()) {
        case "/":
            // Basic endpoint
            ctx.SetStatusCode(fasthttp.StatusOK)
            ctx.SetContentType("application/json")
            fmt.Fprintf(ctx, `{
                "clientIP": "%s",
                "ipType": "%s",
                "isPublic": %t,
                "framework": "fasthttp"
            }`, clientIP, ipInfo.Type.String(), ipInfo.IsPublic)

        case "/api/data":
            // API endpoint with IP validation
            if ipInfo.IsPrivate {
                ctx.SetStatusCode(fasthttp.StatusForbidden)
                ctx.SetContentType("application/json")
                fmt.Fprintf(ctx, `{
                    "error": "Access denied",
                    "clientIP": "%s",
                    "reason": "Private IP not allowed"
                }`, clientIP)
                return
            }

            ctx.SetStatusCode(fasthttp.StatusOK)
            ctx.SetContentType("application/json")
            fmt.Fprintf(ctx, `{
                "clientIP": "%s",
                "data": ["item1", "item2", "item3"]
            }`, clientIP)

        case "/health":
            // Health check
            ctx.SetStatusCode(fasthttp.StatusOK)
            ctx.SetContentType("application/json")
            fmt.Fprintf(ctx, `{
                "status": "healthy",
                "clientIP": "%s"
            }`, clientIP)

        default:
            ctx.Error("Not found", fasthttp.StatusNotFound)
        }
    }

    log.Println("FastHTTP server starting on :8080")
    log.Fatal(fasthttp.ListenAndServe(":8080", handler))
}
*/

/*
// Real FastHTTP implementation (add fasthttp dependency to use):

package main

import (
    "fmt"
    "log"
    "github.com/valyala/fasthttp"
    "github.com/fsvxavier/nexs-lib/ip"
)

func main() {
    handler := func(ctx *fasthttp.RequestCtx) {
        clientIP := ip.GetRealIP(ctx)
        ipInfo := ip.GetRealIPInfo(ctx)

        // Log for performance monitoring
        log.Printf("FastHTTP - Client IP: %s, Type: %s", clientIP, ipInfo.Type.String())

        switch string(ctx.Path()) {
        case "/":
            // Basic endpoint
            ctx.SetStatusCode(fasthttp.StatusOK)
            ctx.SetContentType("application/json")
            fmt.Fprintf(ctx, `{
                "clientIP": "%s",
                "ipType": "%s",
                "isPublic": %t,
                "framework": "fasthttp"
            }`, clientIP, ipInfo.Type.String(), ipInfo.IsPublic)

        case "/api/data":
            // API endpoint with IP validation
            if ipInfo.IsPrivate {
                ctx.SetStatusCode(fasthttp.StatusForbidden)
                ctx.SetContentType("application/json")
                fmt.Fprintf(ctx, `{
                    "error": "Access denied",
                    "clientIP": "%s",
                    "reason": "Private IP not allowed"
                }`, clientIP)
                return
            }

            ctx.SetStatusCode(fasthttp.StatusOK)
            ctx.SetContentType("application/json")
            fmt.Fprintf(ctx, `{
                "clientIP": "%s",
                "data": ["item1", "item2", "item3"]
            }`, clientIP)

        case "/health":
            // Health check
            ctx.SetStatusCode(fasthttp.StatusOK)
            ctx.SetContentType("application/json")
            fmt.Fprintf(ctx, `{
                "status": "healthy",
                "clientIP": "%s"
            }`, clientIP)

        default:
            ctx.Error("Not found", fasthttp.StatusNotFound)
        }
    }

    log.Println("FastHTTP server starting on :8080")
    log.Fatal(fasthttp.ListenAndServe(":8080", handler))
}
*/
