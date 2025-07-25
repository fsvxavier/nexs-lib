// Package main demonstrates migration from _old/validator to new jsonschema module
package main

import (
	"fmt"
	"log"
	"strings"

	// Old validator import (simulated)
	// old "github.com/fsvxavier/nexs-lib/_old/validator"

	// New jsonschema imports
	"github.com/fsvxavier/nexs-lib/validation/jsonschema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
)

func main() {
	fmt.Println("=== Migration Example: Old Validator to New JSON Schema ===\n")

	// Example 1: Legacy function usage (still works)
	fmt.Println("1. Legacy Function Compatibility:")
	legacyFunctionExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 2: Migration to new API
	fmt.Println("2. Migrated to New API:")
	migratedAPIExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 3: Enhanced features with new API
	fmt.Println("3. Enhanced Features (New API Only):")
	enhancedFeaturesExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 4: Side-by-side comparison
	fmt.Println("4. Side-by-Side Comparison:")
	comparisonExample()
}

func legacyFunctionExample() {
	// This simulates the old validator usage pattern
	// In reality, these functions would be imported from _old/validator

	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["name"]
	}`

	data := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	// This would be the old way (commented out since _old/validator might not exist)
	// err := old.Validate(data, schema)

	// But the new module provides legacy compatibility functions
	fmt.Println("Using legacy compatibility function:")
	err := jsonschema.Validate(data, schema)

	if err != nil {
		fmt.Printf("❌ Legacy validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Legacy validation passed!")
	}

	fmt.Println("Note: This uses the same function signature as the old validator")
}

func migratedAPIExample() {
	// The same validation using the new, more powerful API

	// Configure with gojsonschema for maximum compatibility with old behavior
	cfg := config.NewConfig().WithProvider(config.GoJSONSchemaProvider)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["name"]
	}`)

	data := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	fmt.Println("Using new API with detailed error information:")
	errors, err := validator.ValidateFromBytes(schema, data)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ New API validation passed!")
		fmt.Println("Benefits: More detailed error information, configurable providers")
	} else {
		fmt.Printf("❌ New API validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - Field: %s, Message: %s, Type: %s\n", ve.Field, ve.Message, ve.ErrorType)
		}
	}
}

func enhancedFeaturesExample() {
	// Features that are only available in the new API

	cfg := config.NewConfig()

	// Enhanced feature 1: Multiple provider support
	cfg.WithProvider(config.JSONSchemaProvider) // Use kaptinlin for better performance

	// Enhanced feature 2: Schema registration for reuse
	userSchema := []byte(`{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"username": {"type": "string", "minLength": 3},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["id", "username", "email"]
	}`)

	cfg.RegisterSchema("user", userSchema)

	// Enhanced feature 3: Custom error mapping
	cfg.SetErrorMapping(map[string]string{
		"required":     "CAMPO_OBRIGATORIO",
		"invalid_type": "TIPO_INVALIDO",
		"format":       "FORMATO_INVALIDO",
	})

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Test with registered schema
	userData := map[string]interface{}{
		"id":       123,
		"username": "ab", // Too short
		// Missing email
	}

	fmt.Println("Using enhanced features:")
	fmt.Println("- Schema registration and reuse")
	fmt.Println("- Custom error type mapping")
	fmt.Println("- Performance-optimized provider")

	errors, err := validator.ValidateFromStruct("user", userData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Enhanced validation passed!")
	} else {
		fmt.Printf("❌ Enhanced validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - Field: %s, Type: %s, Message: %s\n", ve.Field, ve.ErrorType, ve.Message)
		}
	}
}

func comparisonExample() {
	schema := `{
		"type": "object",
		"properties": {
			"email": {"type": "string", "format": "email"},
			"age": {"type": "integer", "minimum": 18}
		},
		"required": ["email", "age"]
	}`

	// Data with validation errors
	invalidData := map[string]interface{}{
		"email": "not-an-email",
		"age":   15, // Below minimum
	}

	fmt.Println("Old API Style (legacy compatibility):")
	fmt.Println("===================================")

	// Legacy function - simple boolean result
	err := jsonschema.Validate(invalidData, schema)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Validation passed")
	}

	fmt.Println("\nNew API Style (detailed errors):")
	fmt.Println("=================================")

	// New API - detailed error information
	validator, err := jsonschema.NewValidator(nil)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	errors, err := validator.ValidateFromBytes([]byte(schema), invalidData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Validation passed")
	} else {
		fmt.Printf("❌ Validation failed with %d specific errors:\n", len(errors))
		for i, ve := range errors {
			fmt.Printf("  %d. Field: '%s'\n", i+1, ve.Field)
			fmt.Printf("     Error: %s\n", ve.Message)
			fmt.Printf("     Type: %s\n", ve.ErrorType)
			if ve.Value != nil {
				fmt.Printf("     Invalid value: %v\n", ve.Value)
			}
			fmt.Println()
		}
	}

	fmt.Println("Migration Benefits:")
	fmt.Println("==================")
	fmt.Println("✅ Backward compatibility - old code continues to work")
	fmt.Println("✅ Enhanced error reporting - field-level error details")
	fmt.Println("✅ Multiple validation engines - choose best for your use case")
	fmt.Println("✅ Schema reuse - register schemas once, use multiple times")
	fmt.Println("✅ Extensibility - hooks and custom checks")
	fmt.Println("✅ Performance options - different providers for different needs")
	fmt.Println("✅ Custom error mapping - localization support")
}
