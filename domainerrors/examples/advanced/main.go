package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Domain Errors - Advanced Examples ===")
	fmt.Println()

	// Example 1: Different error types with specific information
	specificErrorTypesExample()

	// Example 2: Error wrapping with context
	errorWrappingExample()

	// Example 3: Complex error chains
	complexErrorChainsExample()

	// Example 4: Error handling in different layers
	layeredErrorHandlingExample()

	// Example 5: Custom metadata and serialization
	customMetadataExample()
}

func specificErrorTypesExample() {
	fmt.Println("1. Specific Error Types:")

	// Validation Error
	validationErr := domainerrors.NewValidationError("USER_VALIDATION_FAILED", "User validation failed", nil)
	validationErr.WithField("email", "invalid format")
	validationErr.WithField("age", "must be positive")

	fmt.Printf("   Validation Error: %s\n", validationErr.Error())
	fmt.Printf("   Fields: %+v\n", validationErr.Fields)

	// Business Error
	businessErr := domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Account balance too low")
	businessErr.WithRule("minimum balance of $10 required")
	businessErr.WithRule("account must be active")

	fmt.Printf("   Business Error: %s\n", businessErr.Error())
	fmt.Printf("   Rules: %+v\n", businessErr.Rules)

	// Database Error
	dbErr := domainerrors.NewDatabaseError("DB_QUERY_FAILED", "Query failed", errors.New("connection timeout"))
	dbErr.WithOperation("SELECT", "users")
	dbErr.WithQuery("SELECT * FROM users WHERE id = ?")

	fmt.Printf("   Database Error: %s\n", dbErr.Error())
	fmt.Printf("   Operation: %s on %s\n", dbErr.Operation, dbErr.Table)

	// External Service Error
	extErr := domainerrors.NewExternalServiceError("PAYMENT_GATEWAY_ERROR", "payment-gateway", "Payment processing failed", errors.New("timeout"))
	extErr.WithEndpoint("/api/v1/charge")
	extErr.WithResponse(503, "Service temporarily unavailable")

	fmt.Printf("   External Service Error: %s\n", extErr.Error())
	fmt.Printf("   Service: %s, Endpoint: %s, Status: %d\n", extErr.Service, extErr.Endpoint, extErr.StatusCode)

	fmt.Println()
}

func errorWrappingExample() {
	fmt.Println("2. Error Wrapping with Context:")

	ctx := context.Background()

	// Create base error
	originalErr := errors.New("network connection failed")

	// Wrap with infraestructure error
	infraErr := domainerrors.NewinfraestructureError("REDIS_CONNECTION_FAILED", "redis", "Cache operation failed", originalErr)
	infraErr.WithAction("SET")

	// Wrap with service error using context
	serviceErr := domainerrors.NewServerError("USER_SERVICE_FAILED", "User service failed", infraErr)
	serviceErr.WithContext(ctx, "processing user registration")
	serviceErr.WithRequestInfo("req-12345", "corr-67890")

	fmt.Printf("   Final Error: %s\n", serviceErr.Error())
	fmt.Printf("   Request ID: %s\n", serviceErr.RequestID)
	fmt.Printf("   Correlation ID: %s\n", serviceErr.CorrelationID)
	fmt.Printf("   Stack Trace:\n%s\n", serviceErr.StackTrace())

	fmt.Println()
}

func complexErrorChainsExample() {
	fmt.Println("3. Complex Error Chains:")

	// Simulate a complex error scenario

	// 1. Database connection fails
	dbConnErr := errors.New("connection pool exhausted")

	// 2. Repository layer handles it
	repoErr := domainerrors.NewDatabaseError("DB_FETCH_FAILED", "Failed to fetch user", dbConnErr)
	repoErr.WithOperation("SELECT", "users")
	repoErr.WithQuery("SELECT * FROM users WHERE email = ?")

	// 3. Service layer wraps it
	serviceErr := domainerrors.NewServerError("USER_LOOKUP_FAILED", "User lookup failed", repoErr)
	serviceErr.WithComponent("user-service")
	serviceErr.WithRequestInfo("req-abc123", "trace-def456")

	// 4. Controller layer adds final context
	controllerErr := domainerrors.New("API_001", "Request processing failed")
	controllerErr.WithMetadata("endpoint", "/api/v1/users/search")
	controllerErr.WithMetadata("method", "GET")
	controllerErr.WithMetadata("user_agent", "curl/7.68.0")
	controllerErr.Wrap("handling user search request", serviceErr)

	fmt.Printf("   Complex Error Chain: %s\n", controllerErr.Error())
	fmt.Printf("   Metadata: %+v\n", controllerErr.Metadata)

	// Walk through the error chain
	fmt.Println("   Error Chain:")
	current := controllerErr
	level := 1
	for current != nil {
		fmt.Printf("     %d. %s (Type: %s)\n", level, current.Message, current.Type)
		if unwrapped := current.Unwrap(); unwrapped != nil {
			if domainErr, ok := unwrapped.(*domainerrors.DomainError); ok {
				current = domainErr
				level++
			} else {
				fmt.Printf("     %d. %s (Type: standard error)\n", level+1, unwrapped.Error())
				break
			}
		} else {
			break
		}
	}

	fmt.Println()
}

func layeredErrorHandlingExample() {
	fmt.Println("4. Layered Error Handling:")

	// Simulate different layers of an application

	// Data layer
	dataErr := simulateDataLayerError()
	fmt.Printf("   Data Layer Error: %s\n", dataErr.Error())

	// Service layer
	serviceErr := simulateServiceLayerError(dataErr)
	fmt.Printf("   Service Layer Error: %s\n", serviceErr.Error())

	// API layer
	apiErr := simulateAPILayerError(serviceErr)
	fmt.Printf("   API Layer Error: %s\n", apiErr.Error())
	fmt.Printf("   Final HTTP Status: %d\n", domainerrors.MapHTTPStatus(apiErr))

	fmt.Println()
}

func simulateDataLayerError() error {
	// Simulate a timeout in data layer
	return domainerrors.NewTimeoutError("DB_QUERY_TIMEOUT", "database-query", "Query timeout", errors.New("context deadline exceeded")).
		WithDuration(5*time.Second, 3*time.Second)
}

func simulateServiceLayerError(dataErr error) error {
	// Service layer wraps data layer error
	err := domainerrors.NewServerError("USER_SERVICE_FAILED", "User service operation failed", dataErr)
	err.WithComponent("user-service")
	err.WithMetadata("operation", "getUserByEmail")
	err.WithMetadata("retries", 3)
	return err
}

func simulateAPILayerError(serviceErr error) error {
	// API layer adds final context
	err := domainerrors.New("API_TIMEOUT", "Request timeout")
	err.WithMetadata("endpoint", "/api/v1/users/profile")
	err.WithMetadata("method", "GET")
	err.WithMetadata("client_ip", "192.168.1.100")
	err.Wrap("processing user profile request", serviceErr)

	// Check if it's a timeout error and adjust type
	if domainerrors.IsType(serviceErr, domainerrors.ErrorTypeTimeout) {
		err.Type = domainerrors.ErrorTypeTimeout
	}

	return err
}

func customMetadataExample() {
	fmt.Println("5. Custom Metadata and Serialization:")

	// Create error with rich metadata
	err := domainerrors.NewUnprocessableEntityError("USER_VALIDATION_FAILED", "User entity validation failed")
	err.WithEntityInfo("User", "user-12345")
	err.WithValidationErrors(map[string][]string{
		"email": {"invalid format", "already exists"},
		"age":   {"must be 18 or older"},
		"phone": {"invalid country code"},
	})
	err.WithBusinessRuleViolation("User must be verified before activation")
	err.WithBusinessRuleViolation("Premium features require subscription")

	// Add custom metadata
	err.WithMetadata("validation_engine", "v2.1.0")
	err.WithMetadata("rules_version", "2025.01")
	err.WithMetadata("request_source", "web-app")

	fmt.Printf("   Entity Error: %s\n", err.Error())
	fmt.Printf("   Entity Type: %s, Entity ID: %s\n", err.EntityType, err.EntityID)
	fmt.Printf("   Validation Errors: %+v\n", err.ValidationErrors)
	fmt.Printf("   Business Rules: %+v\n", err.BusinessRules)
	fmt.Printf("   Custom Metadata: %+v\n", err.Metadata)

	// Demonstrate error analysis
	fmt.Println("\n   Error Analysis:")
	fmt.Printf("   - Is Validation Error: %v\n", domainerrors.IsType(err.DomainError, domainerrors.ErrorTypeValidation))
	fmt.Printf("   - Is Unprocessable Error: %v\n", domainerrors.IsType(err.DomainError, domainerrors.ErrorTypeUnprocessable))
	fmt.Printf("   - HTTP Status: %d\n", err.HTTPStatus())
	fmt.Printf("   - Error Code: %s\n", err.Code)
	fmt.Printf("   - Timestamp: %s\n", err.Timestamp.Format(time.RFC3339))

	fmt.Println()
}

// Helper function to demonstrate error recovery
func demonstrateErrorRecovery() {
	fmt.Println("6. Error Recovery Patterns:")

	// Retry pattern
	var err error
	for i := 0; i < 3; i++ {
		err = simulateUnreliableOperation()
		if err == nil {
			fmt.Printf("   ✓ Operation succeeded on attempt %d\n", i+1)
			break
		}

		// Check if error is retryable
		if domainerrors.IsType(err, domainerrors.ErrorTypeTimeout) ||
			domainerrors.IsType(err, domainerrors.ErrorTypeExternalService) {
			fmt.Printf("   ⚠ Retryable error on attempt %d: %s\n", i+1, err.Error())
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		} else {
			fmt.Printf("   ✗ Non-retryable error: %s\n", err.Error())
			break
		}
	}

	if err != nil {
		fmt.Printf("   ✗ Operation failed after all retries: %s\n", err.Error())
	}

	fmt.Println()
}

func simulateUnreliableOperation() error {
	// Simulate an operation that might fail
	// This is just for demonstration
	return domainerrors.NewTimeoutError("API_CALL_TIMEOUT", "external-api", "API call timeout", errors.New("connection timeout"))
}

func init() {
	log.SetFlags(0) // Remove timestamp from logs for cleaner output
}
