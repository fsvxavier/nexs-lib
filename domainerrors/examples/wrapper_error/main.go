package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Enhanced Wrapper Functions Examples ===\n")

	// Example 1: Basic wrapper with type preservation
	fmt.Println("1. Basic wrapper with type preservation:")
	originalValidation := domainerrors.NewValidationError("VAL001", "Invalid input", nil)
	originalValidation.WithMetadata("field", "email")
	originalValidation.WithMetadata("reason", "invalid format")

	wrappedValidation := domainerrors.Wrap("WRAP001", "Processing failed", originalValidation)

	fmt.Printf("   Original Type: %s\n", originalValidation.Type)
	fmt.Printf("   Wrapped Type: %s (preserved!)\n", wrappedValidation.Type)
	fmt.Printf("   Metadata preserved: %v\n", wrappedValidation.Metadata)
	fmt.Printf("   Error chain: %s\n\n", wrappedValidation.Error())

	// Example 2: Context-aware wrapper
	fmt.Println("2. Context-aware wrapper:")
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "trace_id", "trace-456")

	baseError := errors.New("database connection failed")
	wrappedWithContext := domainerrors.WrapWithContext(ctx, "DB001", "Database operation failed", baseError)

	fmt.Printf("   Context values added: %v\n", wrappedWithContext.Metadata)
	fmt.Printf("   Error: %s\n\n", wrappedWithContext.Error())

	// Example 3: Wrapper with specific type
	fmt.Println("3. Wrapper with specific type:")
	networkError := errors.New("network timeout")
	wrappedAsTimeout := domainerrors.WrapWithType("NET001", "Network operation failed", domainerrors.ErrorTypeTimeout, networkError)

	fmt.Printf("   Original error: %s\n", networkError.Error())
	fmt.Printf("   Wrapped type: %s\n", wrappedAsTimeout.Type)
	fmt.Printf("   HTTP Status: %d\n\n", wrappedAsTimeout.HTTPStatus())

	// Example 4: Chain wrapper with layers
	fmt.Println("4. Chain wrapper with layers:")
	businessErr := domainerrors.NewBusinessError("BUS001", "Business rule violated")
	chainedError := domainerrors.WrapChain("CHAIN001", "Multi-layer failure", businessErr, "service", "controller", "handler")

	fmt.Printf("   Original Type: %s (preserved)\n", chainedError.Type)
	fmt.Printf("   Layers: %v\n", chainedError.Metadata)
	fmt.Printf("   Error: %s\n\n", chainedError.Error())

	// Example 5: Complex error chain
	fmt.Println("5. Complex error chain:")

	// Start with a database error
	dbErr := domainerrors.NewDatabaseError("DB001", "Connection failed", errors.New("network unreachable"))
	dbErr.WithQuery("SELECT * FROM users")

	// Wrap it as a service error
	serviceErr := domainerrors.Wrap("SVC001", "User service failed", dbErr)

	// Add context wrapper
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	finalErr := domainerrors.WrapWithContext(ctx2, "API001", "API request failed", serviceErr)

	fmt.Printf("   Final error type: %s\n", finalErr.Type)
	fmt.Printf("   Metadata: %v\n", finalErr.Metadata)
	fmt.Printf("   Full error chain: %s\n", finalErr.Error())

	// Demonstrate unwrapping
	fmt.Println("\n   Unwrapping chain:")
	current := finalErr
	level := 1
	for current != nil {
		fmt.Printf("   Level %d: %s (Type: %s)\n", level, current.Message, current.Type)
		if wrapped := errors.Unwrap(current); wrapped != nil {
			if domainErr, ok := wrapped.(*domainerrors.DomainError); ok {
				current = domainErr
			} else {
				fmt.Printf("   Level %d: %s (Standard error)\n", level+1, wrapped.Error())
				break
			}
		} else {
			break
		}
		level++
	}
}
