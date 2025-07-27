package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Service demonstrates dependency injection pattern
type APIService struct {
	httpClient interfaces.Client
}

// NewAPIService creates a new API service with injected HTTP client
func NewAPIService(client interfaces.Client) *APIService {
	return &APIService{
		httpClient: client,
	}
}

// GetUserData demonstrates using the injected client
func (s *APIService) GetUserData(ctx context.Context, userID string) ([]byte, error) {
	resp, err := s.httpClient.Get(ctx, fmt.Sprintf("/users/%s", userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}
	return resp.Body, nil
}

// Application demonstrates the main application setup
type Application struct {
	apiService    *APIService
	clientManager *httpclient.ClientManager
}

// NewApplication creates a new application with all dependencies
func NewApplication() (*Application, error) {
	// Get the global client manager (singleton)
	manager := httpclient.GetManager()

	// Create or get a reusable HTTP client for the API service
	// This client will be reused across all instances
	apiClient, err := httpclient.NewNamed(
		"jsonplaceholder-api", // Named client for dependency injection
		interfaces.ProviderNetHTTP,
		"https://jsonplaceholder.typicode.com",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// Configure the client for optimal connection reuse
	apiClient.SetTimeout(30 * time.Second)
	apiClient.SetHeaders(map[string]string{
		"User-Agent": "nexs-lib-example/1.0",
		"Accept":     "application/json",
	})

	// Create the API service with the injected client
	apiService := NewAPIService(apiClient)

	return &Application{
		apiService:    apiService,
		clientManager: manager,
	}, nil
}

// Run demonstrates the application running with connection reuse
func (app *Application) Run() error {
	ctx := context.Background()

	fmt.Println("=== Dependency Injection with Connection Reuse Example ===")

	// Demonstrate reusing the same client across multiple operations
	fmt.Println("1. Multiple API calls using the same injected client:")

	for i := 1; i <= 3; i++ {
		userData, err := app.apiService.GetUserData(ctx, fmt.Sprintf("%d", i))
		if err != nil {
			log.Printf("Error getting user %d: %v", i, err)
			continue
		}
		fmt.Printf("User %d data length: %d bytes\n", i, len(userData))
	}

	// Demonstrate getting the same client from different parts of the application
	fmt.Println("\n2. Retrieving the same client from different locations:")

	// Get the same named client from anywhere in the application
	sameClient, exists := httpclient.GetNamedClient("jsonplaceholder-api")
	if !exists {
		return fmt.Errorf("named client not found")
	}

	// Verify it's the same client by comparing IDs
	originalID := app.apiService.httpClient.GetID()
	retrievedID := sameClient.GetID()

	fmt.Printf("Original client ID: %s\n", originalID)
	fmt.Printf("Retrieved client ID: %s\n", retrievedID)
	fmt.Printf("Same client instance: %t\n", originalID == retrievedID)

	// Demonstrate connection reuse metrics
	fmt.Println("\n3. Connection reuse metrics:")
	metrics := sameClient.GetMetrics()
	fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("Successful requests: %d\n", metrics.SuccessfulRequests)
	fmt.Printf("Average latency: %v\n", metrics.AverageLatency)

	// Demonstrate multiple named clients for different services
	fmt.Println("\n4. Multiple named clients for different services:")

	// Create another client for a different service
	githubClient, err := httpclient.NewNamed(
		"github-api",
		interfaces.ProviderNetHTTP,
		"https://api.github.com",
	)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// List all managed clients
	clientNames := app.clientManager.ListClients()
	fmt.Printf("Managed clients: %v\n", clientNames)

	// Test the GitHub client
	resp, err := githubClient.Get(ctx, "/users/octocat")
	if err != nil {
		log.Printf("Error calling GitHub API: %v", err)
	} else {
		fmt.Printf("GitHub API response status: %d\n", resp.StatusCode)
	}

	// Demonstrate client health checks
	fmt.Println("\n5. Client health checks:")
	fmt.Printf("API client healthy: %t\n", app.apiService.httpClient.IsHealthy())
	fmt.Printf("GitHub client healthy: %t\n", githubClient.IsHealthy())

	return nil
}

// Shutdown demonstrates proper cleanup
func (app *Application) Shutdown() error {
	fmt.Println("\n6. Shutting down application and cleaning up connections...")
	return app.clientManager.Shutdown()
}

func main() {
	// Create the application with dependency injection
	app, err := NewApplication()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}

	// Shutdown gracefully
	if err := app.Shutdown(); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	fmt.Println("Application finished successfully!")
}
