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
	fmt.Println("=== Fiber Provider Example ===")

	// Example 1: Simple client creation
	simpleExample()

	// Example 2: JSON operations
	jsonExample()

	// Example 3: Performance testing
	performanceExample()

	// Example 4: Concurrent requests
	concurrentExample()
}

func simpleExample() {
	fmt.Println("\n1. Simple Fiber Client Example")

	// Create a simple Fiber client
	client, err := httpclient.New(interfaces.ProviderFiber, "https://httpbin.org")
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
	fmt.Printf("Is Error: %t\n", resp.IsError)
}

func jsonExample() {
	fmt.Println("\n2. JSON Operations Example")

	// Create Fiber client with JSON configuration
	cfg := config.NewBuilder().
		WithBaseURL("https://httpbin.org").
		WithTimeout(15*time.Second).
		WithHeader("Content-Type", "application/json").
		WithHeader("Accept", "application/json").
		WithHeader("User-Agent", "Fiber-Example/1.0").
		Build()

	client, err := httpclient.NewWithConfig(interfaces.ProviderFiber, cfg)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Test different data structures
	testData := []interface{}{
		map[string]interface{}{
			"string_field":  "test value",
			"number_field":  42,
			"boolean_field": true,
			"array_field":   []string{"a", "b", "c"},
		},
		struct {
			Name   string   `json:"name"`
			Age    int      `json:"age"`
			Tags   []string `json:"tags"`
			Active bool     `json:"active"`
		}{
			Name:   "Fiber Test",
			Age:    25,
			Tags:   []string{"fiber", "http", "client"},
			Active: true,
		},
	}

	for i, data := range testData {
		resp, err := client.Post(ctx, "/post", data)
		if err != nil {
			log.Printf("POST request %d failed: %v", i+1, err)
			continue
		}

		fmt.Printf("POST %d Status: %d\n", i+1, resp.StatusCode)
		fmt.Printf("POST %d Response length: %d bytes\n", i+1, len(resp.Body))
	}
}

func performanceExample() {
	fmt.Println("\n3. Performance Testing Example")

	// Create high-performance Fiber client
	cfg := config.NewBuilder().
		WithBaseURL("https://httpbin.org").
		WithTimeout(5 * time.Second).
		WithMetricsEnabled(true).
		WithTracingEnabled(false). // Disable for performance
		Build()

	client, err := httpclient.NewWithConfig(interfaces.ProviderFiber, cfg)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Perform multiple requests to measure performance
	const numRequests = 10
	start := time.Now()

	for i := 0; i < numRequests; i++ {
		resp, err := client.Get(ctx, "/get")
		if err != nil {
			log.Printf("Request %d failed: %v", i+1, err)
			continue
		}

		if i == 0 {
			fmt.Printf("First request status: %d\n", resp.StatusCode)
		}
	}

	totalTime := time.Since(start)
	fmt.Printf("Completed %d requests in %v\n", numRequests, totalTime)
	fmt.Printf("Average time per request: %v\n", totalTime/numRequests)

	// Display metrics
	metrics := client.GetProvider().GetMetrics()
	fmt.Printf("Provider metrics:\n")
	fmt.Printf("  Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Successful: %d\n", metrics.SuccessfulRequests)
	fmt.Printf("  Failed: %d\n", metrics.FailedRequests)
	fmt.Printf("  Average latency: %v\n", metrics.AverageLatency)
}

func concurrentExample() {
	fmt.Println("\n4. Concurrent Requests Example")

	client, err := httpclient.New(interfaces.ProviderFiber, "https://httpbin.org")
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	const numWorkers = 5
	const requestsPerWorker = 3

	resultChan := make(chan result, numWorkers*requestsPerWorker)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go worker(i, client, requestsPerWorker, resultChan)
	}

	// Collect results
	var successful, failed int
	var totalLatency time.Duration

	for i := 0; i < numWorkers*requestsPerWorker; i++ {
		res := <-resultChan
		if res.err != nil {
			failed++
			fmt.Printf("Worker %d request failed: %v\n", res.workerID, res.err)
		} else {
			successful++
			totalLatency += res.latency
		}
	}

	fmt.Printf("Concurrent execution results:\n")
	fmt.Printf("  Successful: %d\n", successful)
	fmt.Printf("  Failed: %d\n", failed)
	if successful > 0 {
		fmt.Printf("  Average latency: %v\n", totalLatency/time.Duration(successful))
	}

	// Final provider metrics
	metrics := client.GetProvider().GetMetrics()
	fmt.Printf("Final provider metrics:\n")
	fmt.Printf("  Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Success rate: %.2f%%\n",
		float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
}

type result struct {
	workerID int
	latency  time.Duration
	err      error
}

func worker(workerID int, client interfaces.Client, numRequests int, resultChan chan<- result) {
	ctx := context.Background()

	for i := 0; i < numRequests; i++ {
		start := time.Now()

		// Alternate between different endpoints
		endpoint := "/get"
		if i%2 == 1 {
			endpoint = "/user-agent"
		}

		_, err := client.Get(ctx, endpoint)
		latency := time.Since(start)

		resultChan <- result{
			workerID: workerID,
			latency:  latency,
			err:      err,
		}

		// Small delay between requests from the same worker
		time.Sleep(100 * time.Millisecond)
	}
}
