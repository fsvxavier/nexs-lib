// Package main demonstrates JSON Schema validation with custom checks
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/checks"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
)

func main() {
	fmt.Println("=== JSON Schema Validation with Custom Checks ===\n")

	// Example 1: Required fields check
	fmt.Println("1. Required Fields Check:")
	requiredFieldsExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 2: Enum constraints check
	fmt.Println("2. Enum Constraints Check:")
	enumConstraintsExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 3: Date validation check
	fmt.Println("3. Date Validation Check:")
	dateValidationExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 4: Combined checks
	fmt.Println("4. Combined Checks:")
	combinedChecksExample()
}

func requiredFieldsExample() {
	// Create configuration with required fields check
	cfg := config.NewConfig()

	// Add required fields check
	requiredCheck := &checks.RequiredFieldsCheck{
		RequiredFields: []string{"user_id", "email", "full_name"},
	}
	cfg.AddCheck(requiredCheck)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Schema that doesn't enforce these as required
	schema := []byte(`{
		"type": "object",
		"properties": {
			"user_id": {"type": "integer"},
			"email": {"type": "string"},
			"full_name": {"type": "string"},
			"age": {"type": "integer"},
			"bio": {"type": "string"}
		}
	}`)

	// Test complete data
	completeData := map[string]interface{}{
		"user_id":   123,
		"email":     "john@example.com",
		"full_name": "John Doe",
		"age":       30,
	}

	fmt.Println("Testing complete data:")
	errors, err := validator.ValidateFromBytes(schema, completeData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Complete data passed validation!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}

	// Test incomplete data
	incompleteData := map[string]interface{}{
		"user_id": 123,
		// Missing email and full_name
		"age": 30,
	}

	fmt.Println("\nTesting incomplete data (missing required fields):")
	errors, err = validator.ValidateFromBytes(schema, incompleteData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Data passed validation!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func enumConstraintsExample() {
	// Create configuration with enum constraints check
	cfg := config.NewConfig()

	// Add enum constraints check
	enumCheck := &checks.EnumConstraintsCheck{
		Constraints: map[string][]interface{}{
			"status":   {"active", "inactive", "pending", "suspended"},
			"role":     {"admin", "user", "guest", "moderator"},
			"priority": {1, 2, 3, 4, 5}, // Numeric enum
		},
	}
	cfg.AddCheck(enumCheck)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Schema without enum constraints (will be enforced by check)
	schema := []byte(`{
		"type": "object",
		"properties": {
			"status": {"type": "string"},
			"role": {"type": "string"},
			"priority": {"type": "integer"},
			"name": {"type": "string"}
		}
	}`)

	// Valid data
	validData := map[string]interface{}{
		"status":   "active",
		"role":     "user",
		"priority": 3,
		"name":     "Task 1",
	}

	fmt.Println("Testing valid enum values:")
	errors, err := validator.ValidateFromBytes(schema, validData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Valid enum data passed validation!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}

	// Invalid data
	invalidData := map[string]interface{}{
		"status":   "archived",   // Not in enum
		"role":     "superadmin", // Not in enum
		"priority": 10,           // Not in enum
		"name":     "Task 2",
	}

	fmt.Println("\nTesting invalid enum values:")
	errors, err = validator.ValidateFromBytes(schema, invalidData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Data passed validation!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s (%s)\n", ve.Field, ve.Message, ve.ErrorType)
		}
	}
}

func dateValidationExample() {
	// Create configuration with date validation check
	cfg := config.NewConfig()

	// Add date validation check
	dateCheck := &checks.DateValidationCheck{
		DateFields:  []string{"birth_date", "created_at", "expires_at"},
		AllowFuture: true, // Allow future dates for expires_at
		AllowPast:   true, // Allow past dates for birth_date
	}
	cfg.AddCheck(dateCheck)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Schema with date fields as strings
	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"birth_date": {"type": "string"},
			"created_at": {"type": "string"},
			"expires_at": {"type": "string"}
		}
	}`)

	// Valid data with proper dates
	validData := map[string]interface{}{
		"name":       "John Doe",
		"birth_date": "1990-05-15T00:00:00Z",
		"created_at": time.Now().Format(time.RFC3339),
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	fmt.Println("Testing valid date data:")
	errors, err := validator.ValidateFromBytes(schema, validData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Valid date data passed validation!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}

	// Invalid data with malformed dates
	invalidData := map[string]interface{}{
		"name":       "Jane Doe",
		"birth_date": "not-a-date", // Invalid date format
		"created_at": "2024-13-45", // Invalid date values
		"expires_at": "yesterday",  // Invalid date format
	}

	fmt.Println("\nTesting invalid date data:")
	errors, err = validator.ValidateFromBytes(schema, invalidData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Data passed validation!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func combinedChecksExample() {
	// Create configuration with multiple checks
	cfg := config.NewConfig()

	// Add required fields check
	cfg.AddCheck(&checks.RequiredFieldsCheck{
		RequiredFields: []string{"id", "title", "status", "created_at"},
	})

	// Add enum constraints check
	cfg.AddCheck(&checks.EnumConstraintsCheck{
		Constraints: map[string][]interface{}{
			"status":   {"draft", "published", "archived"},
			"category": {"news", "blog", "tutorial", "announcement"},
		},
	})

	// Add date validation check
	cfg.AddCheck(&checks.DateValidationCheck{
		DateFields:  []string{"created_at", "updated_at", "published_at"},
		AllowFuture: false, // Don't allow future dates
		AllowPast:   true,
	})

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Schema for article/post validation
	schema := []byte(`{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"title": {"type": "string"},
			"content": {"type": "string"},
			"status": {"type": "string"},
			"category": {"type": "string"},
			"created_at": {"type": "string"},
			"updated_at": {"type": "string"},
			"published_at": {"type": "string"}
		}
	}`)

	// Valid article data
	validArticle := map[string]interface{}{
		"id":           1,
		"title":        "Getting Started with JSON Schema",
		"content":      "This is a comprehensive guide...",
		"status":       "published",
		"category":     "tutorial",
		"created_at":   "2024-01-15T10:00:00Z",
		"updated_at":   "2024-01-16T14:30:00Z",
		"published_at": "2024-01-16T15:00:00Z",
	}

	fmt.Println("Testing valid article with all checks:")
	errors, err := validator.ValidateFromBytes(schema, validArticle)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Valid article passed all checks!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s (%s)\n", ve.Field, ve.Message, ve.ErrorType)
		}
	}

	// Invalid article with multiple issues
	invalidArticle := map[string]interface{}{
		"id":    1,
		"title": "Invalid Article",
		// Missing required "status" and "created_at"
		"category":     "invalid_category",                                  // Not in enum
		"updated_at":   "not-a-date",                                        // Invalid date
		"published_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339), // Future date
	}

	fmt.Println("\nTesting invalid article with multiple issues:")
	errors, err = validator.ValidateFromBytes(schema, invalidArticle)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Article passed validation!")
	} else {
		fmt.Printf("❌ Article validation failed with %d errors:\n", len(errors))
		for i, ve := range errors {
			fmt.Printf("%d. %s: %s (%s)\n", i+1, ve.Field, ve.Message, ve.ErrorType)
		}
	}

	fmt.Println("\nAll checks were executed:")
	fmt.Println("- RequiredFieldsCheck: Verified required fields presence")
	fmt.Println("- EnumConstraintsCheck: Validated enum values")
	fmt.Println("- DateValidationCheck: Validated date formats and constraints")
}
