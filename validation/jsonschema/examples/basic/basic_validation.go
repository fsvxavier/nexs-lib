// Package main demonstrates basic JSON Schema validation usage
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
)

func main() {
	fmt.Println("=== Basic JSON Schema Validation Example ===\n")

	// Example 1: Simple validation with default config
	fmt.Println("1. Simple Validation:")
	basicValidation()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 2: Validation with custom provider
	fmt.Println("2. Custom Provider Validation:")
	customProviderValidation()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 3: Schema from file
	fmt.Println("3. Schema from File:")
	schemaFromFileValidation()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 4: Multiple validations with same schema
	fmt.Println("4. Multiple Validations:")
	multipleValidations()
}

func basicValidation() {
	// Create validator with default configuration
	validator, err := jsonschema.NewValidator(nil)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Define a simple schema
	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"minLength": 2,
				"maxLength": 50
			},
			"age": {
				"type": "integer",
				"minimum": 0,
				"maximum": 120
			},
			"email": {
				"type": "string",
				"format": "email"
			}
		},
		"required": ["name", "email"],
		"additionalProperties": false
	}`)

	// Valid data
	validData := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
	}

	fmt.Println("Validating valid data:")
	errors, err := validator.ValidateFromBytes(schema, validData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Valid data passed validation!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}

	// Invalid data
	invalidData := map[string]interface{}{
		"name": "J", // Too short
		"age":  150, // Too high
		// Missing required email
		"extra": "not allowed", // Additional property
	}

	fmt.Println("\nValidating invalid data:")
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

func customProviderValidation() {
	// Configure with specific provider for compatibility
	cfg := config.NewConfig().WithProvider(config.GoJSONSchemaProvider)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	schema := []byte(`{
		"type": "object",
		"properties": {
			"product": {
				"type": "string"
			},
			"price": {
				"type": "number",
				"minimum": 0
			},
			"tags": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"minItems": 1
			}
		},
		"required": ["product", "price"]
	}`)

	data := map[string]interface{}{
		"product": "Laptop",
		"price":   999.99,
		"tags":    []interface{}{"electronics", "computers"},
	}

	fmt.Printf("Using provider: %s\n", config.GoJSONSchemaProvider)
	errors, err := validator.ValidateFromBytes(schema, data)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Product data is valid!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func schemaFromFileValidation() {
	// This would work if we had a schema file
	// For demo purposes, we'll create the validator but skip file validation
	_, err := jsonschema.NewValidator(nil)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	fmt.Println("Note: This example would validate against a schema file.")
	fmt.Println("Usage: errors, err := validator.ValidateFromFile(\"schema.json\", data)")
	fmt.Println("Make sure the schema file exists and is readable.")

	// Example data that would be validated
	userData := map[string]interface{}{
		"username": "johndoe",
		"profile": map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"age":       25,
		},
	}

	dataJSON, _ := json.MarshalIndent(userData, "", "  ")
	fmt.Printf("Sample data to validate:\n%s\n", string(dataJSON))
}

func multipleValidations() {
	// Create validator with registered schema
	cfg := config.NewConfig()

	// Register a schema for reuse
	userSchema := []byte(`{
		"type": "object",
		"properties": {
			"id": {
				"type": "integer",
				"minimum": 1
			},
			"username": {
				"type": "string",
				"pattern": "^[a-zA-Z0-9_]+$",
				"minLength": 3,
				"maxLength": 20
			},
			"role": {
				"type": "string",
				"enum": ["admin", "user", "guest"]
			}
		},
		"required": ["id", "username", "role"]
	}`)

	cfg.RegisterSchema("user", userSchema)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Test multiple users
	users := []map[string]interface{}{
		{
			"id":       1,
			"username": "admin_user",
			"role":     "admin",
		},
		{
			"id":       2,
			"username": "john_doe",
			"role":     "user",
		},
		{
			"id":       0,              // Invalid: minimum 1
			"username": "a",            // Invalid: too short
			"role":     "invalid_role", // Invalid: not in enum
		},
	}

	for i, user := range users {
		fmt.Printf("Validating user %d:\n", i+1)
		errors, err := validator.ValidateFromStruct("user", user)
		if err != nil {
			log.Printf("Validation error: %v", err)
			continue
		}

		if len(errors) == 0 {
			fmt.Printf("✅ User %d is valid\n", i+1)
		} else {
			fmt.Printf("❌ User %d validation failed with %d errors:\n", i+1, len(errors))
			for _, ve := range errors {
				fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
			}
		}
		fmt.Println()
	}
}
