package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/ip"
)

func main() {
	// Example 1: Basic IP extraction
	fmt.Println("=== Example 1: Basic IP Extraction ===")

	// Simulate a request with proxy headers
	req1 := &http.Request{
		Header:     make(http.Header),
		RemoteAddr: "192.168.1.100:54321",
	}
	req1.Header.Set("X-Forwarded-For", "203.0.113.45, 198.51.100.10")
	req1.Header.Set("X-Real-IP", "203.0.113.45")

	realIP := ip.GetRealIP(req1)
	fmt.Printf("Real IP: %s\n", realIP)

	// Example 2: Detailed IP information
	fmt.Println("\n=== Example 2: Detailed IP Information ===")

	ipInfo := ip.GetRealIPInfo(req1)
	if ipInfo != nil {
		fmt.Printf("IP: %s\n", ipInfo.IP.String())
		fmt.Printf("Type: %s\n", ipInfo.Type.String())
		fmt.Printf("Is IPv4: %v\n", ipInfo.IsIPv4)
		fmt.Printf("Is IPv6: %v\n", ipInfo.IsIPv6)
		fmt.Printf("Is Public: %v\n", ipInfo.IsPublic)
		fmt.Printf("Is Private: %v\n", ipInfo.IsPrivate)
		fmt.Printf("Original: %s\n", ipInfo.Original)
		fmt.Printf("Source: %s\n", ipInfo.Source)
	}

	// Example 3: IP Chain analysis
	fmt.Println("\n=== Example 3: IP Chain Analysis ===")

	chain := ip.GetIPChain(req1)
	fmt.Printf("IP Chain (%d hops):\n", len(chain))
	for i, ipStr := range chain {
		ipInfo := ip.ParseIP(ipStr)
		if ipInfo != nil {
			fmt.Printf("  %d. %s (%s)\n", i+1, ipInfo.IP.String(), ipInfo.Type.String())
		} else {
			fmt.Printf("  %d. %s (invalid)\n", i+1, ipStr)
		}
	}

	// Example 4: IP validation and classification
	fmt.Println("\n=== Example 4: IP Validation and Classification ===")

	testIPs := []string{
		"8.8.8.8",     // Public IPv4
		"192.168.1.1", // Private IPv4
		"127.0.0.1",   // Loopback IPv4
		"2001:db8::1", // Public IPv6
		"::1",         // Loopback IPv6
		"fe80::1",     // Link-local IPv6
		"invalid.ip",  // Invalid IP
	}

	for _, testIP := range testIPs {
		fmt.Printf("\nTesting IP: %s\n", testIP)

		if ip.IsValidIP(testIP) {
			info := ip.ParseIP(testIP)
			if info != nil {
				fmt.Printf("  Valid: true\n")
				fmt.Printf("  Type: %s\n", info.Type.String())
				fmt.Printf("  Is Public: %v\n", ip.IsPublicIP(info.IP))
				fmt.Printf("  Is Private: %v\n", ip.IsPrivateIP(info.IP))
				fmt.Printf("  Is IPv4: %v\n", info.IsIPv4)
				fmt.Printf("  Is IPv6: %v\n", info.IsIPv6)
			}
		} else {
			fmt.Printf("  Valid: false\n")
		}
	}

	// Example 5: IPv4 to IPv6 conversion
	fmt.Println("\n=== Example 5: IPv4 to IPv6 Conversion ===")

	ipv4Examples := []string{"192.168.1.1", "8.8.8.8", "127.0.0.1"}
	for _, ipv4Str := range ipv4Examples {
		info := ip.ParseIP(ipv4Str)
		if info != nil && info.IsIPv4 {
			ipv6 := ip.ConvertIPv4ToIPv6(info.IP)
			fmt.Printf("IPv4: %s -> IPv6: %s\n", ipv4Str, ipv6.String())
		}
	}

	// Example 6: Working with HTTP server
	fmt.Println("\n=== Example 6: HTTP Server Integration ===")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		realIP := ip.GetRealIP(r)
		ipInfo := ip.GetRealIPInfo(r)

		var response string
		if ipInfo != nil {
			response = fmt.Sprintf(`
Client IP Information:
- Real IP: %s
- IP Type: %s
- Is Public: %v
- Is Private: %v
- Is IPv4: %v
- Is IPv6: %v
- Source Header: %s
- Remote Address: %s

Request Headers:
- X-Forwarded-For: %s
- X-Real-IP: %s
- CF-Connecting-IP: %s
- True-Client-IP: %s

Supported Frameworks: %v
`,
				realIP,
				ipInfo.Type.String(),
				ipInfo.IsPublic,
				ipInfo.IsPrivate,
				ipInfo.IsIPv4,
				ipInfo.IsIPv6,
				ipInfo.Source,
				r.RemoteAddr,
				r.Header.Get("X-Forwarded-For"),
				r.Header.Get("X-Real-IP"),
				r.Header.Get("CF-Connecting-IP"),
				r.Header.Get("True-Client-IP"),
				ip.GetSupportedFrameworks(),
			)
		} else {
			response = "Could not extract IP information\n"
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})

	fmt.Println("Starting HTTP server on :8080...")
	fmt.Println("Visit http://localhost:8080 to see IP extraction in action")
	fmt.Println("Try setting headers like:")
	fmt.Println("  curl -H 'X-Forwarded-For: 8.8.8.8, 192.168.1.1' http://localhost:8080")
	fmt.Println("  curl -H 'CF-Connecting-IP: 203.0.113.100' http://localhost:8080")
	fmt.Println("  curl -H 'X-Real-IP: 198.51.100.42' http://localhost:8080")
	fmt.Println("\nThis example demonstrates the new multi-framework IP library!")
	fmt.Println("The same code works with net/http, Fiber, Gin, Echo, FastHTTP, and Atreugo!")

	// Auto-stop server after 3 seconds for demonstration
	fmt.Println("🚀 Server will stop automatically in 3 seconds for demonstration...")

	server := &http.Server{Addr: ":8080"}
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("✅ Example completed successfully - Server stopped automatically")
		server.Shutdown(context.Background())
	}()

	server.ListenAndServe()
}
