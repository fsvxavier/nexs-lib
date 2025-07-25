// Package main demonstrates JSON Schema validation with different providers
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
)

// PhoneFormatChecker implements interfaces.FormatChecker for phone validation
type PhoneFormatChecker struct{}

func (p *PhoneFormatChecker) IsFormat(input interface{}) bool {
	if str, ok := input.(string); ok {
		// Simple phone validation - requires at least 10 chars and contains dash
		return len(str) >= 10 && strings.Contains(str, "-")
	}
	return false
}

// SSNFormatChecker implements interfaces.FormatChecker for SSN validation
type SSNFormatChecker struct{}

func (s *SSNFormatChecker) IsFormat(input interface{}) bool {
	if str, ok := input.(string); ok {
		// Simple SSN validation (XXX-XX-XXXX)
		parts := strings.Split(str, "-")
		return len(parts) == 3 && len(parts[0]) == 3 && len(parts[1]) == 2 && len(parts[2]) == 4
	}
	return false
}

func main() {
	fmt.Println("=== JSON Schema Validation with Different Providers ===\n")

	// Example 1: Default provider (kaptinlin)
	fmt.Println("1. Default Provider (kaptinlin/jsonschema):")
	defaultProviderExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 2: GoJSONSchema provider for compatibility
	fmt.Println("2. GoJSONSchema Provider (xeipuuv/gojsonschema):")
	gojsonschemaProviderExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 3: Santhosh provider
	fmt.Println("3. Santhosh Provider (santhosh-tekuri/jsonschema):")
	santhoshProviderExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 4: Provider comparison
	fmt.Println("4. Provider Performance Comparison:")
	providerComparisonExample()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// Example 5: Custom formats with providers
	fmt.Println("5. Custom Formats with Different Providers:")
	customFormatsExample()
}

func defaultProviderExample() {
	// Use default configuration (kaptinlin provider)
	validator, err := jsonschema.NewValidator(nil)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	schema := []byte(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"minLength": 2
			},
			"email": {
				"type": "string",
				"format": "email"
			},
			"age": {
				"type": "integer",
				"minimum": 0,
				"maximum": 150
			}
		},
		"required": ["name", "email"]
	}`)

	data := map[string]interface{}{
		"name":  "Alice Johnson",
		"email": "alice@example.com",
		"age":   28,
	}

	fmt.Println("Using kaptinlin/jsonschema provider (default)")
	start := time.Now()
	errors, err := validator.ValidateFromBytes(schema, data)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	fmt.Printf("⏱️  Validation took: %v\n", duration)

	if len(errors) == 0 {
		fmt.Println("✅ Data validation passed!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func gojsonschemaProviderExample() {
	// Configure with GoJSONSchema provider for compatibility
	cfg := config.NewConfig().WithProvider(config.GoJSONSchemaProvider)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Schema with nested objects
	schema := []byte(`{
		"type": "object",
		"properties": {
			"user": {
				"type": "object",
				"properties": {
					"id": {"type": "integer"},
					"profile": {
						"type": "object",
						"properties": {
							"firstName": {"type": "string"},
							"lastName": {"type": "string"},
							"age": {"type": "integer", "minimum": 18}
						},
						"required": ["firstName", "lastName"]
					}
				},
				"required": ["id", "profile"]
			},
			"metadata": {
				"type": "object",
				"additionalProperties": true
			}
		},
		"required": ["user"]
	}`)

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id": 123,
			"profile": map[string]interface{}{
				"firstName": "Bob",
				"lastName":  "Smith",
				"age":       25,
			},
		},
		"metadata": map[string]interface{}{
			"source":    "api",
			"timestamp": "2024-01-15T10:00:00Z",
		},
	}

	fmt.Println("Using xeipuuv/gojsonschema provider (compatibility)")
	start := time.Now()
	errors, err := validator.ValidateFromBytes(schema, data)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	fmt.Printf("⏱️  Validation took: %v\n", duration)

	if len(errors) == 0 {
		fmt.Println("✅ Nested data validation passed!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func santhoshProviderExample() {
	// Configure with SchemaJSON provider (santhosh implementation)
	cfg := config.NewConfig().WithProvider(config.SchemaJSONProvider)

	validator, err := jsonschema.NewValidator(cfg)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	// Schema with arrays and complex validation
	schema := []byte(`{
		"type": "object",
		"properties": {
			"products": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"price": {"type": "number", "minimum": 0},
						"tags": {
							"type": "array",
							"items": {"type": "string"},
							"uniqueItems": true
						}
					},
					"required": ["name", "price"]
				},
				"minItems": 1
			},
			"total": {"type": "number", "minimum": 0}
		},
		"required": ["products"]
	}`)

	data := map[string]interface{}{
		"products": []interface{}{
			map[string]interface{}{
				"name":  "Laptop",
				"price": 999.99,
				"tags":  []interface{}{"electronics", "computers"},
			},
			map[string]interface{}{
				"name":  "Mouse",
				"price": 29.99,
				"tags":  []interface{}{"electronics", "accessories"},
			},
		},
		"total": 1029.98,
	}

	fmt.Println("Using santhosh-tekuri/jsonschema provider")
	start := time.Now()
	errors, err := validator.ValidateFromBytes(schema, data)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	fmt.Printf("⏱️  Validation took: %v\n", duration)

	if len(errors) == 0 {
		fmt.Println("✅ Array data validation passed!")
	} else {
		fmt.Printf("❌ Validation failed with %d errors:\n", len(errors))
		for _, ve := range errors {
			fmt.Printf("  - %s: %s\n", ve.Field, ve.Message)
		}
	}
}

func providerComparisonExample() {
	// Prepare test data and schema
	schema := []byte(`{
		"type": "object",
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string", "minLength": 1},
			"email": {"type": "string", "format": "email"},
			"scores": {
				"type": "array",
				"items": {"type": "number"},
				"minItems": 1
			}
		},
		"required": ["id", "name", "email"]
	}`)

	data := map[string]interface{}{
		"id":     1,
		"name":   "Test User",
		"email":  "test@example.com",
		"scores": []interface{}{85.5, 92.0, 78.5},
	}

	providers := []struct {
		name     string
		provider config.ProviderType
	}{
		{"kaptinlin/jsonschema", config.JSONSchemaProvider},
		{"xeipuuv/gojsonschema", config.GoJSONSchemaProvider},
		{"santhosh-tekuri/jsonschema", config.SchemaJSONProvider},
	}

	fmt.Println("Comparing validation performance across providers:")
	fmt.Println()

	for _, p := range providers {
		cfg := config.NewConfig().WithProvider(p.provider)
		validator, err := jsonschema.NewValidator(cfg)
		if err != nil {
			log.Printf("Failed to create validator for %s: %v", p.name, err)
			continue
		}

		// Run validation multiple times for better measurement
		iterations := 1000
		start := time.Now()

		for i := 0; i < iterations; i++ {
			_, err := validator.ValidateFromBytes(schema, data)
			if err != nil {
				log.Printf("Validation error with %s: %v", p.name, err)
				break
			}
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		fmt.Printf("📊 %s:\n", p.name)
		fmt.Printf("   Total time (%d iterations): %v\n", iterations, duration)
		fmt.Printf("   Average per validation: %v\n", avgDuration)
		fmt.Printf("   Validations per second: %.0f\n", float64(iterations)/duration.Seconds())
		fmt.Println()
	}
}

func customFormatsExample() {
	fmt.Println("Testing custom format validation across providers:")
	fmt.Println()

	// Schema with custom format
	schema := []byte(`{
		"type": "object",
		"properties": {
			"phone": {
				"type": "string",
				"format": "phone"
			},
			"ssn": {
				"type": "string",
				"format": "ssn"
			}
		}
	}`)

	// Test data
	validData := map[string]interface{}{
		"phone": "+1-555-123-4567",
		"ssn":   "123-45-6789",
	}

	invalidData := map[string]interface{}{
		"phone": "invalid-phone",
		"ssn":   "invalid-ssn",
	}

	providers := []struct {
		name     string
		provider config.ProviderType
	}{
		{"kaptinlin", config.JSONSchemaProvider},
		{"gojsonschema", config.GoJSONSchemaProvider},
		{"santhosh", config.SchemaJSONProvider},
	}

	for _, p := range providers {
		fmt.Printf("Testing with %s provider:\n", p.name)

		cfg := config.NewConfig().WithProvider(p.provider)

		// Add custom formats using interface implementation
		phoneChecker := &PhoneFormatChecker{}
		ssnChecker := &SSNFormatChecker{}

		cfg.AddCustomFormat("phone", phoneChecker)
		cfg.AddCustomFormat("ssn", ssnChecker)

		validator, err := jsonschema.NewValidator(cfg)
		if err != nil {
			log.Printf("Failed to create validator for %s: %v", p.name, err)
			continue
		}

		// Test valid data
		errors, err := validator.ValidateFromBytes(schema, validData)
		if err != nil {
			log.Printf("Validation error with %s: %v", p.name, err)
			continue
		}

		if len(errors) == 0 {
			fmt.Printf("  ✅ Valid custom formats passed\n")
		} else {
			fmt.Printf("  ❌ Valid data failed (%d errors)\n", len(errors))
		}

		// Test invalid data
		errors, err = validator.ValidateFromBytes(schema, invalidData)
		if err != nil {
			log.Printf("Validation error with %s: %v", p.name, err)
			continue
		}

		if len(errors) > 0 {
			fmt.Printf("  ✅ Invalid custom formats correctly rejected (%d errors)\n", len(errors))
		} else {
			fmt.Printf("  ❌ Invalid data incorrectly passed\n")
		}

		fmt.Println()
	}

	fmt.Println("Note: Custom format support may vary between providers.")
	fmt.Println("Some providers may not support runtime custom format registration.")
}
