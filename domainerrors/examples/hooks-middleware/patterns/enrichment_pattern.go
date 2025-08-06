// Enrichment Pattern Example
// This example demonstrates how to enrich errors with contextual information using middleware
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Enrichment Pattern Example ===")

	// Middleware 1: Add request context
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		fmt.Println("🔧 Middleware 1: Adding request context")

		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		// Simulate request context
		err.Metadata["request"] = map[string]interface{}{
			"id":     "req_12345",
			"method": "POST",
			"path":   "/api/v1/users",
			"ip":     "192.168.1.100",
		}

		return next(err)
	})

	// Middleware 2: Add service information
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		fmt.Println("🔧 Middleware 2: Adding service information")

		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		err.Metadata["service"] = map[string]interface{}{
			"name":        "user-management-api",
			"version":     "2.1.0",
			"environment": "production",
			"hostname":    "api-server-03",
			"region":      "us-east-1",
		}

		return next(err)
	})

	// Middleware 3: Add timing information
	domainerrors.RegisterMiddleware(func(ctx context.Context, err *domainerrors.DomainError, next func(*domainerrors.DomainError) *domainerrors.DomainError) *domainerrors.DomainError {
		fmt.Println("🔧 Middleware 3: Adding timing information")

		if err.Metadata == nil {
			err.Metadata = make(map[string]interface{})
		}

		now := time.Now()
		err.Metadata["timing"] = map[string]interface{}{
			"timestamp":       now.Format(time.RFC3339),
			"unix_timestamp":  now.Unix(),
			"timezone":        now.Location().String(),
			"processing_time": "127ms", // Simulated
		}

		return next(err)
	})

	fmt.Println()
	fmt.Println("Creating error - watch the enrichment pipeline...")
	fmt.Println()

	// Create an error that will go through the enrichment pipeline
	businessErr := domainerrors.NewWithType(
		"USER_CREATION_FAILED",
		"Failed to create user account",
		domainerrors.ErrorTypeBusinessRule,
	)

	// Add some business-specific metadata
	businessErr.WithMetadata("user_email", "john.doe@example.com")
	businessErr.WithMetadata("validation_errors", []string{"email_already_exists", "weak_password"})

	fmt.Println()
	fmt.Println("=== Final Enriched Error ===")
	fmt.Printf("Code: %s\n", businessErr.Code)
	fmt.Printf("Type: %s\n", businessErr.Type)
	fmt.Printf("Message: %s\n", businessErr.Message)
	fmt.Println()
	fmt.Println("Enriched Metadata:")
	printMetadata(businessErr.Metadata, "")
}

func printMetadata(metadata map[string]interface{}, indent string) {
	for key, value := range metadata {
		switch v := value.(type) {
		case map[string]interface{}:
			fmt.Printf("%s%s:\n", indent, key)
			printMetadata(v, indent+"  ")
		case []string:
			fmt.Printf("%s%s: [%s]\n", indent, key, joinStrings(v, ", "))
		default:
			fmt.Printf("%s%s: %v\n", indent, key, value)
		}
	}
}

func joinStrings(slice []string, sep string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
