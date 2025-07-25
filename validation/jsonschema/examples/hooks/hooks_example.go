// Package main demonstrates advanced JSON Schema validation with hooks
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/hooks"
)

func main() {
	fmt.Println("=== Advanced JSON Schema Validation with Hooks ===\n")

	// Example 1: Data normalization hooks
	fmt.Println("1. Data Normalization Hook:")
	dataNormalizationExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 2: Logging hooks
	fmt.Println("2. Logging Hook:")
	loggingHookExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 3: Error enrichment hooks
	fmt.Println("3. Error Enrichment Hook:")
	errorEnrichmentExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 4: Combined hooks pipeline
	fmt.Println("4. Combined Hooks Pipeline:")
	combinedHooksExample()
}

func dataNormalizationExample() {
	// Create configuration with data normalization hook
	cfg := config.NewConfig()

	// Add pre-validation hook for data normalization
	normalizationHook := &hooks.DataNormalizationHook{
		TrimStrings:   true,
		LowerCaseKeys: false, // Keep original key casing
	}
	cfg.AddPreValidationHook(normalizationHook)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"minLength": 3
			},
			"description": {
				"type": "string",
				"maxLength": 100
			}
		},
		"required": ["name"]
	}`)

	// Data with extra whitespace that will be normalized
	dataWithWhitespace := map[string]interface{}{
		"name":        "  John Doe  ",                         // Will be trimmed
		"description": "  A sample description with spaces  ", // Will be trimmed
	}

	fmt.Println("Original data has whitespace around values")
	fmt.Printf("Name: '%s'\n", dataWithWhitespace["name"])
	fmt.Printf("Description: '%s'\n", dataWithWhitespace["description"])

	errors, err := validator.ValidateFromBytes(schema, dataWithWhitespace)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Data validation passed after normalization!")
		fmt.Println("Strings were automatically trimmed by the normalization hook")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func loggingHookExample() {
	// Create configuration with logging hooks
	cfg := config.NewConfig()

	// Add logging hook for pre-validation
	loggingHook := &hooks.LoggingHook{
		LogData: true, // Log the data being validated
	}
	cfg.AddPreValidationHook(loggingHook)

	// Add validation summary hook for post-validation
	summaryHook := &hooks.ValidationSummaryHook{
		LogSummary: true, // Log summary of validation results
	}
	cfg.AddPostValidationHook(summaryHook)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	schema := []byte(`{
		"type": "object",
		"properties": {
			"user_id": {
				"type": "integer",
				"minimum": 1
			},
			"status": {
				"type": "string",
				"enum": ["active", "inactive"]
			}
		},
		"required": ["user_id", "status"]
	}`)

	testData := map[string]interface{}{
		"user_id": 42,
		"status":  "active",
	}

	fmt.Println("Validation with logging enabled:")

	errors, err := validator.ValidateFromBytes(schema, testData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Validation completed successfully!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func errorEnrichmentExample() {
	// Create configuration with error enrichment
	cfg := config.NewConfig()

	// Add error enrichment hook
	enrichmentHook := &hooks.ErrorEnrichmentHook{
		AddContext:     true, // Add contextual information to errors
		AddSuggestions: true, // Add suggestions for fixing errors
	}
	cfg.AddPostValidationHook(enrichmentHook)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	schema := []byte(`{
		"type": "object",
		"properties": {
			"email": {
				"type": "string",
				"format": "email"
			},
			"age": {
				"type": "integer",
				"minimum": 18,
				"maximum": 100
			},
			"role": {
				"type": "string",
				"enum": ["admin", "user", "guest"]
			}
		},
		"required": ["email", "age", "role"]
	}`)

	// Intentionally invalid data to trigger enriched errors
	invalidData := map[string]interface{}{
		"email": "not-an-email", // Invalid email format
		"age":   15,             // Below minimum
		"role":  "superuser",    // Not in enum
	}

	fmt.Println("Validating data with errors to see enrichment:")

	errors, err := validator.ValidateFromBytes(schema, invalidData)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if len(errors) == 0 {
		fmt.Println("✅ Validation passed!")
	} else {
		fmt.Printf("❌ Validation failed with %d enriched errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
			if ve.Description != "" {
				fmt.Printf("    Description: %s\n", ve.Description)
			}
		}
	}
}

func combinedHooksExample() {
	// Create configuration with multiple hooks
	cfg := config.NewConfig()

	// Pre-validation hooks
	cfg.AddPreValidationHook(&hooks.DataNormalizationHook{
		TrimStrings:   true,
		LowerCaseKeys: false,
	})

	cfg.AddPreValidationHook(&hooks.LoggingHook{
		LogData: false, // Don't log data for this example
	})

	// Post-validation hooks
	cfg.AddPostValidationHook(&hooks.ErrorEnrichmentHook{
		AddContext:     true,
		AddSuggestions: true,
	})

	cfg.AddPostValidationHook(&hooks.ValidationSummaryHook{
		LogSummary: true,
	})

	// Error hooks
	cfg.AddErrorHook(&hooks.ErrorFilterHook{
		IgnoreFields: []string{"debug_info", "internal_id"},
		MaxErrors:    5, // Limit to 5 errors max
	})

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	schema := []byte(`{
		"type": "object",
		"properties": {
			"product_name": {
				"type": "string",
				"minLength": 3,
				"maxLength": 50
			},
			"price": {
				"type": "number",
				"minimum": 0
			},
			"category": {
				"type": "string",
				"enum": ["electronics", "clothing", "books", "home"]
			},
			"tags": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"minItems": 1,
				"maxItems": 10
			},
			"debug_info": {
				"type": "string"
			}
		},
		"required": ["product_name", "price", "category"]
	}`)

	// Test data with some issues
	testData := map[string]interface{}{
		"product_name": "  Smart Phone  ",    // Has whitespace (will be normalized)
		"price":        -10,                  // Invalid: negative price
		"category":     "gadgets",            // Invalid: not in enum
		"tags":         []interface{}{},      // Invalid: empty array
		"debug_info":   "should be filtered", // Will be filtered by error hook
	}

	fmt.Println("Running validation with complete hooks pipeline:")
	fmt.Println("- Data normalization (pre-validation)")
	fmt.Println("- Logging (pre-validation)")
	fmt.Println("- Error enrichment (post-validation)")
	fmt.Println("- Validation summary (post-validation)")
	fmt.Println("- Error filtering (error hook)")
	fmt.Println()

	start := time.Now()
	errors, err := validator.ValidateFromBytes(schema, testData)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	fmt.Printf("⏱️  Validation completed in %v\n", duration)

	if len(errors) == 0 {
		fmt.Println("✅ All validations passed!")
	} else {
		fmt.Printf("❌ Validation failed with %d processed errors:\n", len(errors))
		for i, ve := range errors {
			fmt.Printf("%d. Field: %s\n", i+1, ve.Field)
			fmt.Printf("   Message: %s\n", ve.Message)
			fmt.Printf("   Type: %s\n", ve.ErrorType)
			if ve.Description != "" {
				fmt.Printf("   Description: %s\n", ve.Description)
			}
			if ve.Value != nil {
				fmt.Printf("   Value: %v\n", ve.Value)
			}
			fmt.Println()
		}
	}

	fmt.Println("Note: Hooks were executed in the following order:")
	fmt.Println("1. DataNormalizationHook (trimmed whitespace)")
	fmt.Println("2. LoggingHook (logged validation attempt)")
	fmt.Println("3. Validation executed")
	fmt.Println("4. ErrorEnrichmentHook (added context and suggestions)")
	fmt.Println("5. ValidationSummaryHook (logged summary)")
	fmt.Println("6. ErrorFilterHook (filtered and limited errors)")
}
