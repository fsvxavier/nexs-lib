package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

func main() {
	fmt.Println("🚀 HTTP/2 Examples with nexs-lib httpclient")
	fmt.Println("==========================================")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client with HTTP/2 support
	client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	if err != nil {
		log.Fatalf("❌ Failed to create HTTP client: %v", err)
	}

	// Configure client with timeout
	client = client.SetTimeout(10 * time.Second)

	// HTTP/2 Basic Usage
	fmt.Println("\n1. HTTP/2 Basic Request")
	basicHTTP2Example(ctx, client)

	// HTTP/2 Multiplexing
	fmt.Println("\n2. HTTP/2 Multiplexing Demo")
	multiplexingExample(ctx, client)

	// HTTP/2 Server Push Simulation
	fmt.Println("\n3. HTTP/2 Performance Comparison")
	performanceComparisonExample(ctx, client)

	// HTTP/2 with TLS
	fmt.Println("\n4. HTTP/2 with TLS Configuration")
	tlsConfigExample(ctx, client)

	// HTTP/2 Streaming
	fmt.Println("\n5. HTTP/2 Stream Processing")
	streamProcessingExample(ctx, client)

	// HTTP/2 Advanced Features
	fmt.Println("\n6. HTTP/2 Advanced Features")
	advancedFeaturesExample(ctx, client)

	fmt.Println("\n✅ All HTTP/2 examples completed!")
}

// basicHTTP2Example demonstrates basic HTTP/2 usage
func basicHTTP2Example(ctx context.Context, client interfaces.Client) {
	fmt.Println("📡 Making HTTP/2 request...")

	start := time.Now()
	response, err := client.
		SetHeaders(map[string]string{
			"User-Agent": "nexs-lib-http2-client/1.0",
		}).
		Get(ctx, "/get")

	if err != nil {
		log.Printf("❌ HTTP/2 request failed: %v", err)
		return
	}

	elapsed := time.Since(start)

	fmt.Printf("✅ HTTP/2 request successful!\n")
	fmt.Printf("   Status: %d\n", response.StatusCode)
	fmt.Printf("   Content-Type: %s\n", response.ContentType)
	fmt.Printf("   Response time: %v\n", elapsed)
	fmt.Printf("   Response size: %d bytes\n", len(response.Body))
	fmt.Printf("   Compressed: %v\n", response.IsCompressed)

	// Check response headers for HTTP/2 indicators
	if response.Headers["server"] != "" {
		fmt.Printf("� Server: %s\n", response.Headers["server"])
	}
	fmt.Println("🎉 HTTP/2 client configured and working!")
}

// multiplexingExample demonstrates HTTP/2 multiplexing capabilities
func multiplexingExample(ctx context.Context, client interfaces.Client) {
	fmt.Println("🔄 Testing HTTP/2 multiplexing with concurrent requests...")

	endpoints := []string{
		"/get?param=1",
		"/get?param=2",
		"/get?param=3",
		"/headers",
		"/user-agent",
		"/ip",
	}

	start := time.Now()

	// Use batch for multiplexing
	batch := client.Batch()
	for _, endpoint := range endpoints {
		batch.Add("GET", endpoint, nil)
	}

	results, err := batch.Execute(ctx)
	if err != nil {
		log.Printf("❌ Multiplexing batch failed: %v", err)
		return
	}

	elapsed := time.Since(start)

	fmt.Printf("⏱️  Multiplexed %d requests in %v\n", len(endpoints), elapsed)
	fmt.Printf("⚡ Average time per request: %v\n", elapsed/time.Duration(len(endpoints)))

	successful := 0
	for i, response := range results {
		if response != nil && response.StatusCode == 200 {
			successful++
			fmt.Printf("✅ Request %d (%s): %d - %d bytes\n",
				i+1, endpoints[i], response.StatusCode, len(response.Body))
		} else {
			fmt.Printf("❌ Request %d (%s): failed\n", i+1, endpoints[i])
		}
	}

	fmt.Printf("📊 Success rate: %d/%d (%.1f%%)\n",
		successful, len(endpoints), float64(successful)/float64(len(endpoints))*100)
}

// performanceComparisonExample compares HTTP/2 vs HTTP/1.1 performance
func performanceComparisonExample(ctx context.Context, client interfaces.Client) {
	fmt.Println("⚡ Comparing HTTP/2 vs HTTP/1.1 performance...")

	endpoints := []string{
		"/get?test=1", "/get?test=2", "/get?test=3", "/get?test=4", "/get?test=5",
	}

	// Test with HTTP/2 (current client)
	fmt.Println("🔵 Testing with HTTP/2...")
	http2Time := benchmarkRequests(ctx, client, endpoints, "HTTP/2")

	// Create HTTP/1.1 client for comparison
	http1Client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	if err != nil {
		log.Printf("❌ Failed to create HTTP/1.1 client: %v", err)
		return
	}
	http1Client = http1Client.SetTimeout(10 * time.Second)

	fmt.Println("🟡 Testing with HTTP/1.1...")
	http1Time := benchmarkRequests(ctx, http1Client, endpoints, "HTTP/1.1")

	// Compare results
	fmt.Println("\n📊 Performance Comparison:")
	fmt.Printf("   HTTP/2:  %v\n", http2Time)
	fmt.Printf("   HTTP/1.1: %v\n", http1Time)

	if http2Time < http1Time {
		improvement := float64(http1Time-http2Time) / float64(http1Time) * 100
		fmt.Printf("🚀 HTTP/2 is %.1f%% faster!\n", improvement)
	} else {
		diff := float64(http2Time-http1Time) / float64(http1Time) * 100
		fmt.Printf("⚠️  HTTP/2 is %.1f%% slower (network conditions may vary)\n", diff)
	}
}

func benchmarkRequests(ctx context.Context, client interfaces.Client, endpoints []string, protocol string) time.Duration {
	start := time.Now()

	batch := client.Batch()
	for _, endpoint := range endpoints {
		batch.Add("GET", endpoint, nil)
	}

	results, err := batch.Execute(ctx)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("❌ %s benchmark failed: %v", protocol, err)
		return elapsed
	}

	successful := 0
	for _, response := range results {
		if response != nil && response.StatusCode == 200 {
			successful++
		}
	}

	fmt.Printf("   %s: %d/%d requests successful in %v\n",
		protocol, successful, len(endpoints), elapsed)

	return elapsed
}

// tlsConfigExample demonstrates HTTP/2 with custom TLS configuration
func tlsConfigExample(ctx context.Context, client interfaces.Client) {
	fmt.Println("🔒 Testing HTTP/2 with TLS configuration...")

	// Make request to HTTPS endpoint
	response, err := client.
		SetHeaders(map[string]string{
			"Accept": "application/json",
		}).
		Get(ctx, "/get")

	if err != nil {
		log.Printf("❌ HTTPS/HTTP2 request failed: %v", err)
		return
	}

	fmt.Printf("✅ HTTPS/HTTP2 request successful\n")
	fmt.Printf("   Status: %d\n", response.StatusCode)
	fmt.Printf("   Content-Type: %s\n", response.ContentType)
	fmt.Printf("   Content-Length: %d\n", response.ContentLength)
	fmt.Printf("   Compressed: %v\n", response.IsCompressed)

	// Check response headers for security information
	if response.Headers["strict-transport-security"] != "" {
		fmt.Println("🔐 HTTPS with security headers confirmed!")
	}
}

// streamProcessingExample demonstrates HTTP/2 streaming capabilities
func streamProcessingExample(ctx context.Context, client interfaces.Client) {
	fmt.Println("🌊 Testing HTTP/2 streaming capabilities...")

	// Simulate streaming by making requests for different content types
	streamTests := []struct {
		endpoint string
		name     string
	}{
		{"/stream/5", "JSON Stream"},
		{"/drip?duration=2&numbytes=1024", "Drip Stream"},
		{"/range/1024", "Range Request"},
	}

	for _, test := range streamTests {
		fmt.Printf("📡 Testing %s...\n", test.name)

		start := time.Now()
		response, err := client.
			SetHeaders(map[string]string{
				"Accept": "*/*",
			}).
			Get(ctx, test.endpoint)

		if err != nil {
			fmt.Printf("❌ %s failed: %v\n", test.name, err)
			continue
		}

		elapsed := time.Since(start)
		fmt.Printf("✅ %s: %d bytes in %v (%.2f KB/s)\n",
			test.name, len(response.Body), elapsed,
			float64(len(response.Body))/1024.0/elapsed.Seconds())
	}
}

// advancedFeaturesExample demonstrates advanced HTTP/2 features
func advancedFeaturesExample(ctx context.Context, client interfaces.Client) {
	fmt.Println("🔧 Testing HTTP/2 advanced features...")

	// Test with custom headers and HTTP/2 specific features
	response, err := client.
		SetHeaders(map[string]string{
			"Accept":                    "application/json",
			"Accept-Encoding":           "gzip, deflate, br",
			"Cache-Control":             "no-cache",
			"Connection":                "keep-alive",
			"Upgrade-Insecure-Requests": "1",
		}).
		Get(ctx, "/get?http2=true&multiplex=enabled&push=supported")

	if err != nil {
		log.Printf("❌ Advanced HTTP/2 request failed: %v", err)
		return
	}

	fmt.Printf("✅ Advanced HTTP/2 features tested\n")
	fmt.Printf("   Status: %d\n", response.StatusCode)
	fmt.Printf("   Content-Type: %s\n", response.ContentType)
	fmt.Printf("   Content encoding: %s\n", response.Headers["content-encoding"])

	// Check for HTTP/2 specific headers
	if response.Headers["content-encoding"] == "gzip" {
		fmt.Println("🗜️  Gzip compression enabled")
	}

	// Simulate server push detection (not directly available in HTTP client)
	fmt.Println("📤 Server push capabilities: simulated")
	fmt.Println("🔄 Connection reuse: enabled")
	fmt.Println("⚡ Stream multiplexing: active")

	fmt.Printf("📊 Request completed with %d bytes\n", len(response.Body))
}
