package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient"
	"github.com/fsvxavier/nexs-lib/httpclient/config"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

func main() {
	fmt.Println("=== FastHTTP Provider Example ===")

	// Example 1: Simple client creation
	simpleExample()

	// Example 2: High-performance configuration
	highPerformanceExample()

	// Example 3: Stress testing
	stressTestExample()

	// Example 4: Memory efficiency demonstration
	memoryEfficiencyExample()
}

func simpleExample() {
	fmt.Println("\n1. Simple FastHTTP Client Example")

	// Create a simple FastHTTP client
	client, err := httpclient.New(interfaces.ProviderFastHTTP, "https://httpbin.org")
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
	fmt.Printf("Headers count: %d\n", len(resp.Headers))

	// Display some response headers
	for key, value := range resp.Headers {
		if key == "Content-Type" || key == "Content-Length" {
			fmt.Printf("Header %s: %s\n", key, value)
		}
	}
}

func highPerformanceExample() {
	fmt.Println("\n2. High-Performance Configuration Example")

	// Create FastHTTP client optimized for performance
	cfg := config.NewBuilder().
		WithBaseURL("https://httpbin.org").
		WithTimeout(3*time.Second).
		WithMaxIdleConns(200).
		WithIdleConnTimeout(60*time.Second).
		WithMetricsEnabled(true).
		WithTracingEnabled(false). // Disable for maximum performance
		WithHeader("User-Agent", "FastHTTP-Performance/1.0").
		WithHeader("Accept-Encoding", "gzip, deflate").
		Build()

	client, err := httpclient.NewWithConfig(interfaces.ProviderFastHTTP, cfg)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Test different HTTP methods for performance
	methods := []struct {
		name string
		fn   func() (*interfaces.Response, error)
	}{
		{"GET", func() (*interfaces.Response, error) {
			return client.Get(ctx, "/get")
		}},
		{"POST", func() (*interfaces.Response, error) {
			return client.Post(ctx, "/post", map[string]string{"test": "data"})
		}},
		{"PUT", func() (*interfaces.Response, error) {
			return client.Put(ctx, "/put", map[string]interface{}{
				"update":    true,
				"timestamp": time.Now().Unix(),
			})
		}},
		{"DELETE", func() (*interfaces.Response, error) {
			return client.Delete(ctx, "/delete")
		}},
	}

	for _, method := range methods {
		start := time.Now()
		resp, err := method.fn()
		latency := time.Since(start)

		if err != nil {
			log.Printf("%s request failed: %v", method.name, err)
			continue
		}

		fmt.Printf("%s: Status %d, Latency %v, Size %d bytes\n",
			method.name, resp.StatusCode, latency, len(resp.Body))
	}

	// Display performance metrics
	metrics := client.GetProvider().GetMetrics()
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("  Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Success rate: %.2f%%\n",
		float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
	fmt.Printf("  Average latency: %v\n", metrics.AverageLatency)
}

func stressTestExample() {
	fmt.Println("\n3. Stress Testing Example")

	// Create client optimized for high concurrency
	cfg := config.NewBuilder().
		WithBaseURL("https://httpbin.org").
		WithTimeout(2 * time.Second).
		WithMaxIdleConns(500).
		WithMetricsEnabled(true).
		Build()

	client, err := httpclient.NewWithConfig(interfaces.ProviderFastHTTP, cfg)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	const (
		numWorkers        = 20
		requestsPerWorker = 10
		totalRequests     = numWorkers * requestsPerWorker
	)

	fmt.Printf("Starting stress test: %d workers, %d requests each (%d total)\n",
		numWorkers, requestsPerWorker, totalRequests)

	var wg sync.WaitGroup
	resultChan := make(chan StressResult, totalRequests)

	start := time.Now()

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go stressWorker(i, client, requestsPerWorker, resultChan, &wg)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)

	totalTime := time.Since(start)

	// Collect and analyze results
	var successful, failed int
	var totalLatency time.Duration
	var minLatency, maxLatency time.Duration = time.Hour, 0

	for result := range resultChan {
		if result.err != nil {
			failed++
		} else {
			successful++
			totalLatency += result.latency
			if result.latency < minLatency {
				minLatency = result.latency
			}
			if result.latency > maxLatency {
				maxLatency = result.latency
			}
		}
	}

	fmt.Printf("\nStress Test Results:\n")
	fmt.Printf("  Total time: %v\n", totalTime)
	fmt.Printf("  Requests/second: %.2f\n", float64(totalRequests)/totalTime.Seconds())
	fmt.Printf("  Successful requests: %d\n", successful)
	fmt.Printf("  Failed requests: %d\n", failed)
	fmt.Printf("  Success rate: %.2f%%\n", float64(successful)/float64(totalRequests)*100)

	if successful > 0 {
		avgLatency := totalLatency / time.Duration(successful)
		fmt.Printf("  Average latency: %v\n", avgLatency)
		fmt.Printf("  Min latency: %v\n", minLatency)
		fmt.Printf("  Max latency: %v\n", maxLatency)
	}

	// Final metrics from provider
	metrics := client.GetProvider().GetMetrics()
	fmt.Printf("\nProvider Metrics:\n")
	fmt.Printf("  Provider total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Provider avg latency: %v\n", metrics.AverageLatency)
}

func memoryEfficiencyExample() {
	fmt.Println("\n4. Memory Efficiency Demonstration")

	client, err := httpclient.New(interfaces.ProviderFastHTTP, "https://httpbin.org")
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Test with different payload sizes
	payloadSizes := []int{100, 1000, 10000, 100000}

	for _, size := range payloadSizes {
		// Create payload of specified size
		payload := make(map[string]interface{})
		payload["data"] = generateString(size)
		payload["size"] = size
		payload["timestamp"] = time.Now().Unix()

		start := time.Now()
		resp, err := client.Post(ctx, "/post", payload)
		latency := time.Since(start)

		if err != nil {
			log.Printf("Request with payload size %d failed: %v", size, err)
			continue
		}

		fmt.Printf("Payload %d bytes: Status %d, Latency %v, Response %d bytes\n",
			size, resp.StatusCode, latency, len(resp.Body))
	}

	// Test rapid successive requests (memory reuse)
	fmt.Printf("\nTesting memory reuse with rapid requests:\n")
	const rapidRequests = 50

	start := time.Now()
	for i := 0; i < rapidRequests; i++ {
		_, err := client.Get(ctx, "/get")
		if err != nil {
			log.Printf("Rapid request %d failed: %v", i+1, err)
		}
	}
	rapidTime := time.Since(start)

	fmt.Printf("Completed %d rapid requests in %v\n", rapidRequests, rapidTime)
	fmt.Printf("Average time per rapid request: %v\n", rapidTime/rapidRequests)

	metrics := client.GetProvider().GetMetrics()
	fmt.Printf("Final metrics: %d total requests, avg latency %v\n",
		metrics.TotalRequests, metrics.AverageLatency)
}

type StressResult struct {
	workerID int
	latency  time.Duration
	err      error
}

func stressWorker(workerID int, client interfaces.Client, numRequests int,
	resultChan chan<- StressResult, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx := context.Background()

	for i := 0; i < numRequests; i++ {
		start := time.Now()

		// Vary endpoints to test different scenarios
		endpoints := []string{"/get", "/user-agent", "/headers", "/ip"}
		endpoint := endpoints[i%len(endpoints)]

		_, err := client.Get(ctx, endpoint)
		latency := time.Since(start)

		resultChan <- StressResult{
			workerID: workerID,
			latency:  latency,
			err:      err,
		}
	}
}

func generateString(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
