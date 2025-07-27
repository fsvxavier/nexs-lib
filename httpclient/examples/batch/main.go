package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// User represents a user object from JSONPlaceholder API
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
}

// Post represents a post object from JSONPlaceholder API
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// Comment represents a comment object from JSONPlaceholder API
type Comment struct {
	PostID int    `json:"postId"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

// BatchResult contains the result of a batch operation
type BatchResult struct {
	Users    []User
	Posts    []Post
	Comments []Comment
	Errors   []error
}

func main() {
	fmt.Printf("📦 Batch Operations Example\n")
	fmt.Printf("===========================\n\n")

	// Create HTTP client
	client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://jsonplaceholder.typicode.com")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Simple batch requests
	fmt.Println("1️⃣ Simple batch requests...")
	simpleBatchExample(ctx, client)
	fmt.Println()

	// Example 2: Complex batch with different endpoints
	fmt.Println("2️⃣ Complex batch with different endpoints...")
	complexBatchExample(ctx, client)
	fmt.Println()

	// Example 3: Batch with custom request objects
	fmt.Println("3️⃣ Batch with custom request objects...")
	customRequestBatchExample(ctx, client)
	fmt.Println()

	// Example 4: Parallel vs Sequential batch execution
	fmt.Println("4️⃣ Parallel vs Sequential execution comparison...")
	performanceComparisonExample(ctx, client)
	fmt.Println()

	// Example 5: Error handling in batch operations
	fmt.Println("5️⃣ Error handling in batch operations...")
	errorHandlingExample(ctx, client)
	fmt.Println()

	// Example 6: Large batch operations with timeout
	fmt.Println("6️⃣ Large batch operations with timeout...")
	largeTimeoutBatchExample(ctx, client)

	fmt.Println("\n🎉 Batch operations example completed!")
	fmt.Println("\n💡 Key Features Demonstrated:")
	fmt.Println("  • Simple batch requests")
	fmt.Println("  • Complex multi-endpoint batches")
	fmt.Println("  • Custom request objects")
	fmt.Println("  • Parallel vs sequential execution")
	fmt.Println("  • Error handling and resilience")
	fmt.Println("  • Large batch operations with timeout")
	fmt.Println("  • Performance optimization")
}

func simpleBatchExample(ctx context.Context, client interfaces.Client) {
	start := time.Now()

	// Create a batch request for multiple users
	batch := client.Batch().
		Add("GET", "/users/1", nil).
		Add("GET", "/users/2", nil).
		Add("GET", "/users/3", nil).
		Add("GET", "/users/4", nil).
		Add("GET", "/users/5", nil)

	// Execute batch
	results, err := batch.Execute(ctx)
	if err != nil {
		log.Printf("Batch execution failed: %v", err)
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("⏱️  Batch execution took: %v\n", elapsed)
	fmt.Printf("📊 Executed %d requests successfully\n", len(results))

	// Process results
	for i, response := range results {
		if response == nil {
			fmt.Printf("❌ Request %d failed: no response\n", i+1)
			continue
		}

		if response.StatusCode >= 400 {
			fmt.Printf("❌ Request %d failed with status: %d\n", i+1, response.StatusCode)
			continue
		}

		var user User
		if err := json.Unmarshal(response.Body, &user); err != nil {
			fmt.Printf("❌ Failed to parse user %d: %v\n", i+1, err)
			continue
		}

		fmt.Printf("👤 User %d: %s (%s)\n", user.ID, user.Name, user.Email)
	}
}

func complexBatchExample(ctx context.Context, client interfaces.Client) {
	start := time.Now()

	// Create a complex batch with different endpoints
	batch := client.Batch().
		Add("GET", "/users", nil).    // Get all users
		Add("GET", "/posts", nil).    // Get all posts
		Add("GET", "/comments", nil). // Get all comments
		Add("GET", "/albums", nil).   // Get all albums
		Add("GET", "/photos", nil)    // Get all photos

	// Execute batch
	results, err := batch.Execute(ctx)
	if err != nil {
		log.Printf("Complex batch execution failed: %v", err)
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("⏱️  Complex batch execution took: %v\n", elapsed)

	// Process results
	endpoints := []string{"users", "posts", "comments", "albums", "photos"}
	for i, response := range results {
		if response == nil {
			fmt.Printf("❌ %s request failed: no response\n", endpoints[i])
			continue
		}

		if response.StatusCode >= 400 {
			fmt.Printf("❌ %s request failed with status: %d\n", endpoints[i], response.StatusCode)
			continue
		}

		// Parse as generic array to count items
		var items []interface{}
		if err := json.Unmarshal(response.Body, &items); err != nil {
			fmt.Printf("❌ Failed to parse %s: %v\n", endpoints[i], err)
			continue
		}

		fmt.Printf("📋 %s: %d items (response size: %d bytes)\n",
			endpoints[i], len(items), len(response.Body))
	}
}

func customRequestBatchExample(ctx context.Context, client interfaces.Client) {
	start := time.Now()

	// Create custom requests with specific configurations
	batch := client.Batch()

	// Add requests with custom request objects
	requests := []*interfaces.Request{
		{
			Method:  "GET",
			URL:     "/users/1",
			Headers: map[string]string{"Accept": "application/json"},
		},
		{
			Method:  "POST",
			URL:     "/posts",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body: map[string]interface{}{
				"title":  "Batch Created Post",
				"body":   "This post was created via batch operation",
				"userId": 1,
			},
		},
		{
			Method:  "PUT",
			URL:     "/posts/1",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body: map[string]interface{}{
				"id":     1,
				"title":  "Updated Post Title",
				"body":   "This post was updated via batch operation",
				"userId": 1,
			},
		},
		{
			Method: "DELETE",
			URL:    "/posts/1",
		},
	}

	// Add custom requests to batch
	for _, req := range requests {
		batch.AddRequest(req)
	}

	// Execute batch
	results, err := batch.Execute(ctx)
	if err != nil {
		log.Printf("Custom batch execution failed: %v", err)
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("⏱️  Custom batch execution took: %v\n", elapsed)

	// Process results
	operations := []string{"GET user", "POST post", "PUT post", "DELETE post"}
	for i, response := range results {
		if response == nil {
			fmt.Printf("❌ %s failed: no response\n", operations[i])
			continue
		}

		if response.StatusCode >= 400 {
			fmt.Printf("❌ %s failed with status: %d\n", operations[i], response.StatusCode)
			continue
		}

		fmt.Printf("✅ %s: Status %d (response size: %d bytes)\n",
			operations[i], response.StatusCode, len(response.Body))
	}
}

func performanceComparisonExample(ctx context.Context, client interfaces.Client) {
	userIDs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Sequential execution
	fmt.Println("🐌 Sequential execution:")
	start := time.Now()
	for _, id := range userIDs {
		resp, err := client.Get(ctx, fmt.Sprintf("/users/%d", id))
		if err != nil {
			fmt.Printf("Sequential request %d failed: %v\n", id, err)
			continue
		}
		_ = resp // Process response
	}
	sequentialTime := time.Since(start)
	fmt.Printf("⏱️  Sequential time: %v\n", sequentialTime)

	// Batch execution (parallel)
	fmt.Println("🚀 Batch execution (parallel):")
	start = time.Now()
	batch := client.Batch()
	for _, id := range userIDs {
		batch.Add("GET", fmt.Sprintf("/users/%d", id), nil)
	}

	results, err := batch.Execute(ctx)
	if err != nil {
		log.Printf("Batch execution failed: %v", err)
		return
	}
	batchTime := time.Since(start)
	fmt.Printf("⏱️  Batch time: %v\n", batchTime)

	// Calculate improvement
	improvement := float64(sequentialTime) / float64(batchTime)
	fmt.Printf("📈 Performance improvement: %.2fx faster\n", improvement)
	fmt.Printf("📊 Successfully executed: %d/%d requests\n", len(results), len(userIDs))
}

func errorHandlingExample(ctx context.Context, client interfaces.Client) {
	// Create batch with some invalid endpoints
	batch := client.Batch().
		Add("GET", "/users/1", nil).          // Valid
		Add("GET", "/users/999999", nil).     // Valid but returns 404
		Add("GET", "/invalid-endpoint", nil). // Invalid endpoint
		Add("GET", "/users/2", nil).          // Valid
		Add("GET", "/users/3", nil)           // Valid

	// Execute batch
	results, err := batch.Execute(ctx)
	if err != nil {
		log.Printf("Error handling batch execution failed: %v", err)
		return
	}

	// Analyze results
	successful := 0
	clientErrors := 0
	serverErrors := 0
	networkErrors := 0

	for i, response := range results {
		if response == nil {
			networkErrors++
			fmt.Printf("❌ Request %d network error: no response\n", i+1)
			continue
		}

		status := response.StatusCode
		switch {
		case status >= 200 && status < 300:
			successful++
			fmt.Printf("✅ Request %d successful: %d\n", i+1, status)
		case status >= 400 && status < 500:
			clientErrors++
			fmt.Printf("⚠️  Request %d client error: %d\n", i+1, status)
		case status >= 500:
			serverErrors++
			fmt.Printf("🔥 Request %d server error: %d\n", i+1, status)
		}
	}

	fmt.Printf("\n📊 Error Analysis:\n")
	fmt.Printf("  ✅ Successful: %d\n", successful)
	fmt.Printf("  ⚠️  Client errors (4xx): %d\n", clientErrors)
	fmt.Printf("  🔥 Server errors (5xx): %d\n", serverErrors)
	fmt.Printf("  ❌ Network errors: %d\n", networkErrors)
	fmt.Printf("  📈 Success rate: %.1f%%\n",
		float64(successful)/float64(len(results))*100)
}

func largeTimeoutBatchExample(ctx context.Context, client interfaces.Client) {
	// Create a large batch with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	batch := client.Batch()

	// Add 20 requests with varying delays
	for i := 1; i <= 20; i++ {
		if i%5 == 0 {
			// Every 5th request has a 2-second delay
			batch.Add("GET", "/delay/2", nil)
		} else {
			batch.Add("GET", fmt.Sprintf("/users/%d", i%10+1), nil)
		}
	}

	fmt.Printf("⏱️  Executing 20 requests with 5-second timeout...\n")
	start := time.Now()

	results, err := batch.Execute(timeoutCtx)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Large batch failed: %v (after %v)\n", err, elapsed)
		return
	}

	fmt.Printf("✅ Large batch completed in %v\n", elapsed)

	// Count successful vs failed
	successful := 0
	failed := 0
	for _, response := range results {
		if response == nil || response.StatusCode >= 400 {
			failed++
		} else {
			successful++
		}
	}

	fmt.Printf("📊 Results: %d successful, %d failed\n", successful, failed)
	fmt.Printf("⚡ Average time per request: %v\n", elapsed/time.Duration(len(results)))
}
