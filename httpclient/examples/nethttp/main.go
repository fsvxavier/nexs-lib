package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient"
	"github.com/fsvxavier/nexs-lib/httpclient/config"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

func main() {
	fmt.Println("=== NetHTTP Provider Example ===")

	// Example 1: Simple client creation
	simpleExample()

	// Example 2: Advanced configuration
	advancedExample()

	// Example 3: Error handling and retries
	errorHandlingExample()

	// Example 4: Custom headers and timeout
	customConfigExample()
}

func simpleExample() {
	fmt.Println("\n1. Simple NetHTTP Client Example")

	// Create a simple client
	client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Perform a GET request
	resp, err := client.Get(ctx, "/get")
	if err != nil {
		log.Printf("GET request failed: %v", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response length: %d bytes\n", len(resp.Body))
	fmt.Printf("Latency: %v\n", resp.Latency)
}

func advancedExample() {
	fmt.Println("\n2. Advanced NetHTTP Configuration Example")

	// Create configuration with custom settings
	cfg := config.NewBuilder().
		WithBaseURL("https://httpbin.org").
		WithTimeout(10*time.Second).
		WithMaxIdleConns(50).
		WithHeader("User-Agent", "NetHTTP-Example/1.0").
		WithHeader("Accept", "application/json").
		WithTracingEnabled(true).
		WithMetricsEnabled(true).
		Build()

	client, err := httpclient.NewWithConfig(interfaces.ProviderNetHTTP, cfg)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// POST request with JSON body
	requestBody := map[string]interface{}{
		"name":    "NetHTTP Example",
		"version": "1.0.0",
		"tags":    []string{"http", "client", "go"},
	}

	resp, err := client.Post(ctx, "/post", requestBody)
	if err != nil {
		log.Printf("POST request failed: %v", err)
		return
	}

	fmt.Printf("POST Status: %d\n", resp.StatusCode)
	fmt.Printf("Response length: %d bytes\n", len(resp.Body))

	// Check provider metrics
	metrics := client.GetProvider().GetMetrics()
	fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("Successful requests: %d\n", metrics.SuccessfulRequests)
	fmt.Printf("Average latency: %v\n", metrics.AverageLatency)
}

func errorHandlingExample() {
	fmt.Println("\n3. Error Handling and Retry Example")

	// Configuration with retry settings
	cfg := config.NewBuilder().
		WithBaseURL("https://httpbin.org").
		WithMaxRetries(3).
		WithRetryInterval(1 * time.Second).
		Build()

	client, err := httpclient.NewWithConfig(interfaces.ProviderNetHTTP, cfg)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	// Set custom error handler
	client.SetErrorHandler(func(resp *interfaces.Response) error {
		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(resp.Body))
		}
		return nil
	})

	ctx := context.Background()

	// Test with endpoint that returns an error
	resp, err := client.Get(ctx, "/status/500")
	if err != nil {
		fmt.Printf("Expected error received: %v\n", err)
	} else {
		fmt.Printf("Unexpected success: %d\n", resp.StatusCode)
	}

	// Test with successful endpoint
	resp, err = client.Get(ctx, "/status/200")
	if err != nil {
		log.Printf("Unexpected error: %v", err)
	} else {
		fmt.Printf("Success status: %d\n", resp.StatusCode)
	}
}

func customConfigExample() {
	fmt.Println("\n4. Custom Headers and Timeout Example")

	client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	// Set custom headers
	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-API-Version": "v1",
		"X-Client-Type": "nethttp-example",
	}
	client.SetHeaders(headers)

	// Set timeout
	client.SetTimeout(5 * time.Second)

	ctx := context.Background()

	// Test headers endpoint
	resp, err := client.Get(ctx, "/headers")
	if err != nil {
		log.Printf("Headers request failed: %v", err)
		return
	}

	fmt.Printf("Headers response status: %d\n", resp.StatusCode)
	fmt.Printf("Response contains headers info: %t\n",
		len(resp.Body) > 0 && resp.StatusCode == 200)

	// Test different HTTP methods
	methods := []struct {
		name string
		fn   func(context.Context, string) (*interfaces.Response, error)
	}{
		{"DELETE", client.Delete},
		{"HEAD", client.Head},
		{"OPTIONS", client.Options},
	}

	for _, method := range methods {
		resp, err := method.fn(ctx, "/")
		if err != nil {
			log.Printf("%s request failed: %v", method.name, err)
			continue
		}
		fmt.Printf("%s status: %d\n", method.name, resp.StatusCode)
	}
}
