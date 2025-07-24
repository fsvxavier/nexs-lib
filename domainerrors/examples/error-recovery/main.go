package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

// RetryConfig defines retry configuration
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	BackoffType BackoffType
}

// BackoffType defines backoff strategy
type BackoffType int

const (
	BackoffLinear BackoffType = iota
	BackoffExponential
	BackoffJitter
)

// RetryableError interface for errors that can be retried
type RetryableError interface {
	error
	IsRetryable() bool
}

// ExternalAPIClient simulates an external API client
type ExternalAPIClient struct {
	failureRate float64
	timeoutRate float64
}

// NewExternalAPIClient creates a new external API client
func NewExternalAPIClient(failureRate, timeoutRate float64) *ExternalAPIClient {
	return &ExternalAPIClient{
		failureRate: failureRate,
		timeoutRate: timeoutRate,
	}
}

// FetchData simulates fetching data from external API
func (c *ExternalAPIClient) FetchData(ctx context.Context, id string) (string, error) {
	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", domainerrors.NewTimeoutError("API_TIMEOUT", "fetch-data",
			"Request cancelled or timed out", ctx.Err()).
			WithDuration(time.Since(time.Now()), 5*time.Second)
	default:
	}

	// Simulate timeout
	if rand.Float64() < c.timeoutRate {
		return "", domainerrors.NewTimeoutError("API_TIMEOUT", "fetch-data",
			"API request timed out", errors.New("timeout")).
			WithDuration(5*time.Second, 5*time.Second)
	}

	// Simulate failure
	if rand.Float64() < c.failureRate {
		statusCode := []int{500, 502, 503, 504}[rand.Intn(4)]
		return "", domainerrors.NewExternalServiceError("API_ERROR", "external-api",
			fmt.Sprintf("API returned status %d", statusCode),
			errors.New("service unavailable")).
			WithEndpoint("/api/data/"+id).
			WithResponse(statusCode, "Service temporarily unavailable")
	}

	return fmt.Sprintf("data-%s", id), nil
}

// DatabaseClient simulates a database client
type DatabaseClient struct {
	connectionFailureRate float64
	queryFailureRate      float64
}

// NewDatabaseClient creates a new database client
func NewDatabaseClient(connectionFailureRate, queryFailureRate float64) *DatabaseClient {
	return &DatabaseClient{
		connectionFailureRate: connectionFailureRate,
		queryFailureRate:      queryFailureRate,
	}
}

// SaveData simulates saving data to database
func (c *DatabaseClient) SaveData(ctx context.Context, id, data string) error {
	// Simulate processing delay
	time.Sleep(50 * time.Millisecond)

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return domainerrors.NewTimeoutError("DB_TIMEOUT", "save-data",
			"Database operation cancelled", ctx.Err()).
			WithDuration(time.Since(time.Now()), 3*time.Second)
	default:
	}

	// Simulate connection failure
	if rand.Float64() < c.connectionFailureRate {
		return domainerrors.NewDatabaseError("DB_CONNECTION_FAILED", "Connection to database failed",
			errors.New("connection refused")).
			WithOperation("CONNECT", "users").
			WithQuery("CONNECT TO DATABASE")
	}

	// Simulate query failure
	if rand.Float64() < c.queryFailureRate {
		return domainerrors.NewDatabaseError("DB_QUERY_FAILED", "Query execution failed",
			errors.New("deadlock detected")).
			WithOperation("INSERT", "users").
			WithQuery("INSERT INTO users (id, data) VALUES (?, ?)")
	}

	return nil
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failureCount    int
	lastFailureTime time.Time
	state           CircuitState
}

// CircuitState represents circuit breaker state
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.failureCount = 0
		} else {
			return domainerrors.NewExternalServiceError("CIRCUIT_BREAKER_OPEN", "circuit-breaker",
				"Circuit breaker is open", errors.New("too many failures")).
				WithEndpoint("circuit-breaker").
				WithResponse(503, "Service temporarily unavailable")
		}
	}

	err := fn()
	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.failureCount >= cb.maxFailures {
			cb.state = CircuitOpen
		}

		return err
	}

	// Reset on success
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
	}
	cb.failureCount = 0

	return nil
}

// RetryWithBackoff implements retry with backoff strategy
func RetryWithBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastError error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastError = err

		// Check if error is retryable
		if retryableErr, ok := err.(RetryableError); ok && !retryableErr.IsRetryable() {
			return err
		}

		// Don't retry on the last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Calculate delay
		delay := calculateDelay(config, attempt)

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return domainerrors.NewTimeoutError("RETRY_TIMEOUT", "retry-operation",
				"Retry operation cancelled", ctx.Err()).
				WithDuration(time.Since(time.Now()), delay)
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	// Return wrapped error with retry information
	return domainerrors.NewServerError("RETRY_EXHAUSTED", "Maximum retry attempts exceeded",
		lastError).
		WithRequestInfo(fmt.Sprintf("attempts-%d", config.MaxAttempts), "retry-context")
}

// calculateDelay calculates delay based on backoff strategy
func calculateDelay(config RetryConfig, attempt int) time.Duration {
	switch config.BackoffType {
	case BackoffLinear:
		delay := time.Duration(attempt) * config.BaseDelay
		if delay > config.MaxDelay {
			return config.MaxDelay
		}
		return delay

	case BackoffExponential:
		delay := config.BaseDelay * time.Duration(1<<uint(attempt-1))
		if delay > config.MaxDelay {
			return config.MaxDelay
		}
		return delay

	case BackoffJitter:
		baseDelay := config.BaseDelay * time.Duration(1<<uint(attempt-1))
		if baseDelay > config.MaxDelay {
			baseDelay = config.MaxDelay
		}
		jitter := time.Duration(rand.Float64() * float64(baseDelay) * 0.1)
		return baseDelay + jitter

	default:
		return config.BaseDelay
	}
}

// DataProcessor demonstrates error recovery patterns
type DataProcessor struct {
	apiClient      *ExternalAPIClient
	dbClient       *DatabaseClient
	circuitBreaker *CircuitBreaker
}

// NewDataProcessor creates a new data processor
func NewDataProcessor(apiClient *ExternalAPIClient, dbClient *DatabaseClient) *DataProcessor {
	return &DataProcessor{
		apiClient:      apiClient,
		dbClient:       dbClient,
		circuitBreaker: NewCircuitBreaker(3, 30*time.Second),
	}
}

// ProcessData processes data with error recovery
func (p *DataProcessor) ProcessData(ctx context.Context, id string) error {
	// Retry configuration
	retryConfig := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		BackoffType: BackoffExponential,
	}

	// Fetch data with retry and circuit breaker
	var data string
	err := RetryWithBackoff(ctx, retryConfig, func() error {
		return p.circuitBreaker.Execute(func() error {
			fetchedData, err := p.apiClient.FetchData(ctx, id)
			if err != nil {
				return err
			}
			data = fetchedData
			return nil
		})
	})

	if err != nil {
		// Try fallback strategy
		fallbackData, fallbackErr := p.getFallbackData(ctx, id)
		if fallbackErr != nil {
			return domainerrors.NewExternalServiceError("DATA_FETCH_FAILED", "data-processing",
				"Failed to fetch data from primary and fallback sources", err).
				WithEndpoint("/api/data/"+id).
				WithResponse(0, "Both primary and fallback failed")
		}
		data = fallbackData
		fmt.Printf("Used fallback data for ID: %s\n", id)
	}

	// Save data with retry
	saveRetryConfig := RetryConfig{
		MaxAttempts: 2,
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    500 * time.Millisecond,
		BackoffType: BackoffLinear,
	}

	err = RetryWithBackoff(ctx, saveRetryConfig, func() error {
		return p.dbClient.SaveData(ctx, id, data)
	})

	if err != nil {
		// Try alternative storage
		alternativeErr := p.saveToAlternativeStorage(ctx, id, data)
		if alternativeErr != nil {
			return domainerrors.NewDatabaseError("DATA_SAVE_FAILED", "Failed to save data",
				err).
				WithOperation("SAVE", "users").
				WithQuery("Multiple storage attempts failed")
		}
		fmt.Printf("Used alternative storage for ID: %s\n", id)
	}

	return nil
}

// getFallbackData provides fallback data when primary source fails
func (p *DataProcessor) getFallbackData(ctx context.Context, id string) (string, error) {
	// Simulate fallback data source (cache, secondary API, etc.)
	time.Sleep(50 * time.Millisecond)

	// Simulate fallback failure rate
	if rand.Float64() < 0.2 {
		return "", domainerrors.NewExternalServiceError("FALLBACK_FAILED", "fallback-api",
			"Fallback data source failed", errors.New("fallback unavailable")).
			WithEndpoint("/fallback/data/"+id).
			WithResponse(503, "Fallback service unavailable")
	}

	return fmt.Sprintf("fallback-data-%s", id), nil
}

// saveToAlternativeStorage saves data to alternative storage
func (p *DataProcessor) saveToAlternativeStorage(ctx context.Context, id, data string) error {
	// Simulate alternative storage (file system, message queue, etc.)
	time.Sleep(30 * time.Millisecond)

	// Simulate alternative storage failure rate
	if rand.Float64() < 0.1 {
		return domainerrors.NewServerError("ALTERNATIVE_STORAGE_FAILED", "Alternative storage failed",
			errors.New("file system error")).
			WithRequestInfo("alt-storage-"+id, "alternative-context")
	}

	return nil
}

// ErrorRecoveryExample demonstrates various error recovery patterns
func ErrorRecoveryExample() {
	fmt.Println("=== Error Recovery Patterns Example ===")
	fmt.Println()

	// Initialize components with failure rates
	apiClient := NewExternalAPIClient(0.4, 0.2) // 40% failure, 20% timeout
	dbClient := NewDatabaseClient(0.3, 0.2)     // 30% connection failure, 20% query failure
	processor := NewDataProcessor(apiClient, dbClient)

	// Process multiple data items
	ids := []string{"user1", "user2", "user3", "user4", "user5"}

	for _, id := range ids {
		fmt.Printf("Processing data for ID: %s\n", id)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		start := time.Now()
		err := processor.ProcessData(ctx, id)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ Failed to process %s (took %v): %v\n", id, duration, err)

			// Print detailed error information
			if domainErr, ok := err.(*domainerrors.DomainError); ok {
				fmt.Printf("   Error Code: %s\n", domainErr.Code)
				fmt.Printf("   Error Type: %s\n", domainErr.Type)
				fmt.Printf("   HTTP Status: %d\n", domainErr.HTTPStatus())
				if domainErr.Cause != nil {
					fmt.Printf("   Cause: %v\n", domainErr.Cause)
				}
			}
		} else {
			fmt.Printf("✅ Successfully processed %s (took %v)\n", id, duration)
		}

		cancel()
		fmt.Println()
	}
}

// BulkOperationExample demonstrates bulk operations with error recovery
func BulkOperationExample() {
	fmt.Println("=== Bulk Operation Error Recovery Example ===")
	fmt.Println()

	apiClient := NewExternalAPIClient(0.3, 0.1)
	dbClient := NewDatabaseClient(0.2, 0.1)
	processor := NewDataProcessor(apiClient, dbClient)

	// Bulk process data
	ids := []string{"bulk1", "bulk2", "bulk3", "bulk4", "bulk5", "bulk6", "bulk7", "bulk8"}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	successCount := 0
	failureCount := 0
	var failures []string

	for _, id := range ids {
		err := processor.ProcessData(ctx, id)
		if err != nil {
			failureCount++
			failures = append(failures, id)
			fmt.Printf("❌ Failed: %s - %v\n", id, err)
		} else {
			successCount++
			fmt.Printf("✅ Success: %s\n", id)
		}
	}

	fmt.Printf("\n=== Bulk Operation Results ===\n")
	fmt.Printf("Total items: %d\n", len(ids))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failureCount)
	fmt.Printf("Success rate: %.1f%%\n", float64(successCount)/float64(len(ids))*100)

	if len(failures) > 0 {
		fmt.Printf("Failed items: %v\n", failures)
	}
}

func main() {
	// Initialize random number generator (deprecated rand.Seed is not needed in Go 1.20+)
	// Modern Go versions automatically seed the global random number generator

	// Run error recovery examples
	ErrorRecoveryExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	BulkOperationExample()

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("- Retry with exponential backoff")
	fmt.Println("- Circuit breaker pattern")
	fmt.Println("- Fallback mechanisms")
	fmt.Println("- Alternative storage options")
	fmt.Println("- Bulk operation error handling")
	fmt.Println("- Context-based timeouts")
	fmt.Println("- Comprehensive error recovery strategies")
}
