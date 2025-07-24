package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Domain Errors - Basic Examples ===")
	fmt.Println()

	// Example 1: Basic error creation
	basicErrorExample()

	// Example 2: Error with cause
	errorWithCauseExample()

	// Example 3: Error with metadata
	errorWithMetadataExample()

	// Example 4: Error chaining
	errorChainingExample()

	// Example 5: HTTP status mapping
	httpStatusMappingExample()

	// Example 6: Error type checking
	demonstrateErrorChecking()

	// Example 7: Standard error interface
	demonstrateStandardErrorInterface()

	// Example 8: Specific error types
	specificErrorTypesExample()
}

func basicErrorExample() {
	fmt.Println("1. Basic Error Creation:")

	// Create a simple domain error
	err := domainerrors.New("USR_001", "User validation failed")

	fmt.Printf("   Error: %s\n", err.Error())
	fmt.Printf("   Code: %s\n", err.Code)
	fmt.Printf("   Type: %s\n", err.Type)
	fmt.Printf("   HTTP Status: %d\n", err.HTTPStatus())
	fmt.Println()
}

func errorWithCauseExample() {
	fmt.Println("2. Error with Cause:")

	// Create an error with an underlying cause
	originalErr := errors.New("database connection timeout")
	err := domainerrors.NewWithCause("DB_001", "Failed to save user", originalErr)

	fmt.Printf("   Error: %s\n", err.Error())
	fmt.Printf("   Root Cause: %v\n", err.Unwrap())
	fmt.Println()
}

func errorWithMetadataExample() {
	fmt.Println("3. Error with Metadata:")

	// Create error with additional metadata
	err := domainerrors.New("API_001", "Request processing failed")
	err.WithMetadata("request_id", "req-12345")
	err.WithMetadata("user_id", "user-789")
	err.WithMetadata("timestamp", "2025-01-18T10:30:00Z")

	fmt.Printf("   Error: %s\n", err.Error())
	fmt.Printf("   Metadata: %+v\n", err.Metadata)
	fmt.Println()
}

func errorChainingExample() {
	fmt.Println("4. Error Chaining:")

	// Create a chain of errors
	dbErr := errors.New("connection refused")
	infraErr := domainerrors.NewWithCause("INFRA_001", "Database unavailable", dbErr)

	serviceErr := domainerrors.New("SVC_001", "User service failed")
	serviceErr.Wrap("processing user request", infraErr)

	fmt.Printf("   Error: %s\n", serviceErr.Error())
	fmt.Printf("   Root Cause: %v\n", serviceErr.Unwrap())
	fmt.Printf("   Stack Trace:\n%s\n", serviceErr.StackTrace())
	fmt.Println()
}

func httpStatusMappingExample() {
	fmt.Println("5. HTTP Status Mapping:")

	// Different error types map to different HTTP status codes
	errors := []struct {
		name string
		err  *domainerrors.DomainError
	}{
		{
			name: "Validation Error",
			err:  domainerrors.NewWithType("VAL_001", "Invalid input", domainerrors.ErrorTypeValidation),
		},
		{
			name: "Not Found Error",
			err:  domainerrors.NewWithType("NF_001", "Resource not found", domainerrors.ErrorTypeNotFound),
		},
		{
			name: "Authentication Error",
			err:  domainerrors.NewWithType("AUTH_001", "Invalid credentials", domainerrors.ErrorTypeAuthentication),
		},
		{
			name: "Authorization Error",
			err:  domainerrors.NewWithType("AUTHZ_001", "Insufficient permissions", domainerrors.ErrorTypeAuthorization),
		},
		{
			name: "Rate Limit Error",
			err:  domainerrors.NewWithType("RATE_001", "Too many requests", domainerrors.ErrorTypeRateLimit),
		},
		{
			name: "Server Error",
			err:  domainerrors.NewWithType("SRV_001", "Internal server error", domainerrors.ErrorTypeServer),
		},
	}

	for _, errInfo := range errors {
		fmt.Printf("   %s: %d\n", errInfo.name, errInfo.err.HTTPStatus())
	}

	fmt.Println()
}

// Helper function to demonstrate error checking
func demonstrateErrorChecking() {
	fmt.Println("6. Error Type Checking:")

	validationErr := domainerrors.NewWithType("VAL_001", "Invalid data", domainerrors.ErrorTypeValidation)

	// Check if error is of specific type
	if domainerrors.IsType(validationErr, domainerrors.ErrorTypeValidation) {
		fmt.Println("   ✓ Error is a validation error")
	}

	if !domainerrors.IsType(validationErr, domainerrors.ErrorTypeNotFound) {
		fmt.Println("   ✓ Error is NOT a not found error")
	}

	// Map to HTTP status
	status := domainerrors.MapHTTPStatus(validationErr)
	fmt.Printf("   HTTP Status: %d\n", status)

	fmt.Println()
}

// Example of using errors.Is and errors.As
func demonstrateStandardErrorInterface() {
	fmt.Println("7. Standard Error Interface Compatibility:")

	originalErr := errors.New("original error")
	wrappedErr := domainerrors.NewWithCause("WRAP_001", "Wrapped error", originalErr)

	// Using errors.Is
	if errors.Is(wrappedErr, originalErr) {
		fmt.Println("   ✓ errors.Is works correctly")
	}

	// Using errors.As
	var domainErr *domainerrors.DomainError
	if errors.As(wrappedErr, &domainErr) {
		fmt.Printf("   ✓ errors.As works correctly, code: %s\n", domainErr.Code)
	}

	fmt.Println()
}

// Example of specific error types with custom codes
func specificErrorTypesExample() {
	fmt.Println("8. Specific Error Types with Custom Codes:")

	// Validation error with custom code
	validationErr := domainerrors.NewValidationError("USER_EMAIL_INVALID", "Invalid email format", nil)
	validationErr.WithField("email", "must be a valid email address")
	fmt.Printf("   Validation Error: %s (Code: %s)\n", validationErr.Message, validationErr.Code)

	// Business error with custom code
	businessErr := domainerrors.NewBusinessError("INSUFFICIENT_BALANCE", "Account balance too low")
	businessErr.WithRule("Minimum balance of $10 required")
	fmt.Printf("   Business Error: %s (Code: %s)\n", businessErr.Message, businessErr.Code)

	// Database error with custom code
	dbErr := domainerrors.NewDatabaseError("DB_CONNECTION_FAILED", "Failed to connect to database", errors.New("connection timeout"))
	dbErr.WithOperation("SELECT", "users")
	fmt.Printf("   Database Error: %s (Code: %s)\n", dbErr.Message, dbErr.Code)

	// External service error with custom code
	extErr := domainerrors.NewExternalServiceError("PAYMENT_API_ERROR", "payment-service", "Payment processing failed", errors.New("timeout"))
	extErr.WithEndpoint("/api/v1/payments")
	extErr.WithResponse(503, "Service Unavailable")
	fmt.Printf("   External Service Error: %s (Code: %s)\n", extErr.Message, extErr.Code)

	fmt.Println()
}

func init() {
	log.SetFlags(0) // Remove timestamp from logs for cleaner output
}
