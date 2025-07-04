package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/validator/schema"
)

// User represents a user in our system
type User struct {
	ID       string    `validate:"required,uuid"`
	Name     string    `validate:"required,min=2,max=100"`
	Email    string    `validate:"required,email"`
	Age      int       `validate:"min=18,max=120"`
	Website  string    `validate:"url"`
	JoinedAt time.Time `json:"joined_at"`
}

// Product represents a product
type Product struct {
	Name        string  `validate:"required,min=1,max=200"`
	Description string  `validate:"max=1000"`
	Price       float64 `validate:"min=0"`
	Category    string  `validate:"required,pattern=^[a-z_]+$"`
}

func main() {
	fmt.Println("=== Nexs-Lib Validator Examples ===")

	// Example 1: Basic validation with rules
	fmt.Println("1. Basic Validation Examples")
	basicValidationExamples()

	// Example 2: Struct validation with tags
	fmt.Println("\n2. Struct Validation Examples")
	structValidationExamples()

	// Example 3: Fluent builder API
	fmt.Println("\n3. Fluent Builder API Examples")
	builderAPIExamples()

	// Example 4: JSON Schema validation
	fmt.Println("\n4. JSON Schema Validation Examples")
	jsonSchemaExamples()

	// Example 5: Custom validation rules
	fmt.Println("\n5. Custom Validation Rules Examples")
	customValidationExamples()

	// Example 6: All Format Validators
	fmt.Println("\n6. Format Validators Examples")
	formatValidatorsExamples()

	// Example 7: Advanced Schema Validation
	fmt.Println("\n7. Advanced Schema Validation")
	advancedSchemaExamples()

	// Example 8: Context and Performance
	fmt.Println("\n8. Context and Performance Examples")
	contextAndPerformanceExamples()

	// Example 9: Domain Error Integration
	fmt.Println("\n9. Domain Error Integration Examples")
	domainErrorExamples()

	// Example 10: Batch Validation
	fmt.Println("\n10. Batch Validation Examples")
	batchValidationExamples()
}

func basicValidationExamples() {
	ctx := context.Background()

	// Required validation
	fmt.Println("Required validation:")
	requiredRule := schema.NewRequiredRule()

	if err := requiredRule.Validate(ctx, "hello"); err != nil {
		fmt.Printf("  ❌ Error: %s\n", err)
	} else {
		fmt.Println("  ✅ Valid: 'hello'")
	}

	if err := requiredRule.Validate(ctx, ""); err != nil {
		fmt.Printf("  ❌ Error: %s\n", err)
	} else {
		fmt.Println("  ✅ Valid: ''")
	}

	// Email validation
	fmt.Println("\nEmail validation:")
	emailRule := schema.NewEmailRule()

	emails := []string{"test@example.com", "invalid-email", "user@domain.co.uk"}
	for _, email := range emails {
		if err := emailRule.Validate(ctx, email); err != nil {
			fmt.Printf("  ❌ Invalid: %s - %s\n", email, err)
		} else {
			fmt.Printf("  ✅ Valid: %s\n", email)
		}
	}

	// Length validation
	fmt.Println("\nLength validation:")
	minLengthRule := schema.NewMinLengthRule(5)

	strings := []string{"short", "long enough", "hi"}
	for _, str := range strings {
		if err := minLengthRule.Validate(ctx, str); err != nil {
			fmt.Printf("  ❌ Too short: '%s' - %s\n", str, err)
		} else {
			fmt.Printf("  ✅ Valid length: '%s'\n", str)
		}
	}
}

func structValidationExamples() {
	ctx := context.Background()
	v := schema.NewValidator()

	// Valid user
	fmt.Println("Valid user:")
	validUser := User{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    "John Doe",
		Email:   "john@example.com",
		Age:     30,
		Website: "https://johndoe.com",
	}

	result := v.ValidateStruct(ctx, validUser)
	if result.Valid {
		fmt.Println("  ✅ User validation passed")
	} else {
		fmt.Printf("  ❌ User validation failed: %s\n", result.String())
	}

	// Invalid user
	fmt.Println("\nInvalid user:")
	invalidUser := User{
		ID:      "invalid-uuid",
		Name:    "J", // Too short
		Email:   "invalid-email",
		Age:     15, // Too young
		Website: "not-a-url",
	}

	result = v.ValidateStruct(ctx, invalidUser)
	if result.Valid {
		fmt.Println("  ✅ User validation passed")
	} else {
		fmt.Printf("  ❌ User validation failed:\n")
		for field, errors := range result.Errors {
			for _, err := range errors {
				fmt.Printf("    %s: %s\n", field, err)
			}
		}
	}
}

func builderAPIExamples() {
	ctx := context.Background()

	fmt.Println("Building complex validation rules:")

	// Complex string validation
	stringRule := schema.NewRuleBuilder().
		Required().
		String().
		MinLength(5).
		MaxLength(50).
		Email().
		Build()

	testEmails := []string{
		"test@example.com", // Valid
		"a@b.c",            // Too short
		"verylongemailthatexceedsthelimit@example.com", // Too long
		"", // Required but empty
	}

	for _, email := range testEmails {
		if err := stringRule.Validate(ctx, email); err != nil {
			fmt.Printf("  ❌ '%s': %s\n", email, err)
		} else {
			fmt.Printf("  ✅ '%s': Valid\n", email)
		}
	}

	// Number validation
	fmt.Println("\nNumber validation with range:")
	numberRule := schema.NewRuleBuilder().
		Required().
		Number().
		Range(18, 65).
		Integer().
		Build()

	testNumbers := []interface{}{25, 17, 70, 25.5, "30"}
	for _, num := range testNumbers {
		if err := numberRule.Validate(ctx, num); err != nil {
			fmt.Printf("  ❌ %v: %s\n", num, err)
		} else {
			fmt.Printf("  ✅ %v: Valid\n", num)
		}
	}

	// DateTime validation
	fmt.Println("\nDateTime validation with range:")
	dateRule := schema.NewRuleBuilder().
		Required().
		DateTime().
		RFC3339().
		After("2020-01-01T00:00:00Z").
		Before("2030-01-01T00:00:00Z").
		Build()

	testDates := []string{
		"2025-07-04T12:00:00Z", // Valid
		"2019-12-31T23:59:59Z", // Too early
		"2031-01-01T00:00:00Z", // Too late
		"invalid-date",         // Invalid format
	}

	for _, date := range testDates {
		if err := dateRule.Validate(ctx, date); err != nil {
			fmt.Printf("  ❌ '%s': %s\n", date, err)
		} else {
			fmt.Printf("  ✅ '%s': Valid\n", date)
		}
	}
}

func jsonSchemaExamples() {
	ctx := context.Background()
	schemaValidator := schema.NewJSONSchemaValidator()

	fmt.Println("JSON Schema validation:")

	// Product schema
	productSchema := `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"minLength": 1,
				"maxLength": 200
			},
			"description": {
				"type": "string",
				"maxLength": 1000
			},
			"price": {
				"type": "number",
				"minimum": 0
			},
			"category": {
				"type": "string",
				"format": "strong_name"
			},
			"tags": {
				"type": "array",
				"items": {"type": "string"},
				"minItems": 1
			}
		},
		"required": ["name", "price", "category"]
	}`

	// Valid product
	validProduct := map[string]interface{}{
		"name":        "Awesome Product",
		"description": "This is a great product",
		"price":       29.99,
		"category":    "electronics",
		"tags":        []string{"new", "popular"},
	}

	result := schemaValidator.ValidateSchema(ctx, validProduct, productSchema)
	if result.Valid {
		fmt.Println("  ✅ Product schema validation passed")
	} else {
		fmt.Printf("  ❌ Product validation failed: %s\n", result.String())
	}

	// Invalid product
	invalidProduct := map[string]interface{}{
		"name":     "",                  // Empty name
		"price":    -10,                 // Negative price
		"category": "invalid category!", // Invalid format
		"tags":     []string{},          // Empty array
	}

	result = schemaValidator.ValidateSchema(ctx, invalidProduct, productSchema)
	if result.Valid {
		fmt.Println("  ✅ Invalid product somehow passed")
	} else {
		fmt.Println("  ❌ Invalid product correctly failed:")
		for field, errors := range result.Errors {
			for _, err := range errors {
				fmt.Printf("    %s: %s\n", field, err)
			}
		}
	}
}

func customValidationExamples() {
	ctx := context.Background()

	fmt.Println("Custom validation rules:")

	// Custom rule for even numbers
	evenNumberRule := schema.NewCustomRule(
		"even_number",
		"number must be even",
		func(ctx context.Context, value interface{}) error {
			if num, ok := value.(int); ok {
				if num%2 == 0 {
					return nil
				}
				return &schema.ValidationError{
					Field:   "",
					Message: "number must be even",
					Code:    "even_required",
					Value:   value,
				}
			}
			return &schema.ValidationError{
				Field:   "",
				Message: "value must be an integer",
				Code:    "type_error",
				Value:   value,
			}
		},
	)

	testNumbers := []interface{}{2, 3, 4, 5, "not-a-number"}
	for _, num := range testNumbers {
		if err := evenNumberRule.Validate(ctx, num); err != nil {
			fmt.Printf("  ❌ %v: %s\n", num, err)
		} else {
			fmt.Printf("  ✅ %v: Valid (even number)\n", num)
		}
	}

	// Custom format validator
	fmt.Println("\nCustom format validation:")
	schemaValidator := schema.NewJSONSchemaValidator()

	// Add custom credit card format
	schemaValidator.RegisterFormatValidator("credit-card", func(input interface{}) bool {
		if str, ok := input.(string); ok {
			// Simple credit card format: XXXX-XXXX-XXXX-XXXX
			if len(str) == 19 && str[4] == '-' && str[9] == '-' && str[14] == '-' {
				for i, char := range str {
					if i == 4 || i == 9 || i == 14 {
						continue // Skip hyphens
					}
					if char < '0' || char > '9' {
						return false
					}
				}
				return true
			}
		}
		return false
	})

	paymentSchema := `{
		"type": "object",
		"properties": {
			"card_number": {
				"type": "string",
				"format": "credit-card"
			}
		},
		"required": ["card_number"]
	}`

	testCards := []map[string]interface{}{
		{"card_number": "1234-5678-9012-3456"}, // Valid
		{"card_number": "1234567890123456"},    // No hyphens
		{"card_number": "1234-5678-9012-345a"}, // Contains letter
	}

	for i, card := range testCards {
		result := schemaValidator.ValidateSchema(ctx, card, paymentSchema)
		if result.Valid {
			fmt.Printf("  ✅ Card %d: Valid format\n", i+1)
		} else {
			fmt.Printf("  ❌ Card %d: Invalid format\n", i+1)
		}
	}
}

// formatValidatorsExamples demonstrates all built-in format validators
func formatValidatorsExamples() {
	ctx := context.Background()
	validator := schema.NewJSONSchemaValidator()

	fmt.Println("Testing all built-in format validators:")

	// Test date_time format
	fmt.Println("\n📅 DateTime Format Validation:")
	dateTimeSchema := `{
		"type": "object",
		"properties": {
			"timestamp": {"type": "string", "format": "date_time"}
		},
		"required": ["timestamp"]
	}`

	dateTimeTests := []map[string]interface{}{
		{"timestamp": "2025-07-04T12:00:00Z"},      // RFC3339
		{"timestamp": "2025-07-04T12:00:00-03:00"}, // With timezone
		{"timestamp": "2025-07-04"},                // Date only
		{"timestamp": "12:00:00"},                  // Time only
		{"timestamp": "invalid-date"},              // Invalid
	}

	for _, test := range dateTimeTests {
		result := validator.ValidateSchema(ctx, test, dateTimeSchema)
		if result.Valid {
			fmt.Printf("  ✅ %s: Valid\n", test["timestamp"])
		} else {
			fmt.Printf("  ❌ %s: Invalid\n", test["timestamp"])
		}
	}

	// Test iso_8601_date format
	fmt.Println("\n📆 ISO 8601 Date Format Validation:")
	isoDateSchema := `{
		"type": "object", 
		"properties": {
			"date": {"type": "string", "format": "iso_8601_date"}
		},
		"required": ["date"]
	}`

	isoDateTests := []map[string]interface{}{
		{"date": "2025-07-04"}, // Valid ISO date
		{"date": "2024-02-29"}, // Leap year
		{"date": "2025-13-01"}, // Invalid month
		{"date": "2025-07-32"}, // Invalid day
		{"date": "25-07-04"},   // Wrong format
	}

	for _, test := range isoDateTests {
		result := validator.ValidateSchema(ctx, test, isoDateSchema)
		if result.Valid {
			fmt.Printf("  ✅ %s: Valid\n", test["date"])
		} else {
			fmt.Printf("  ❌ %s: Invalid\n", test["date"])
		}
	}

	// Test text_match format
	fmt.Println("\n📝 Text Match Format Validation:")
	textMatchSchema := `{
		"type": "object",
		"properties": {
			"text": {"type": "string", "format": "text_match"}
		},
		"required": ["text"]
	}`

	textMatchTests := []map[string]interface{}{
		{"text": "Hello World"},          // Valid text
		{"text": "Text_with_underscore"}, // Valid with underscore
		{"text": ""},                     // Empty string
		{"text": "Text123"},              // Invalid with numbers
		{"text": "Text@special"},         // Invalid with special chars
	}

	for _, test := range textMatchTests {
		result := validator.ValidateSchema(ctx, test, textMatchSchema)
		if result.Valid {
			fmt.Printf("  ✅ '%s': Valid\n", test["text"])
		} else {
			fmt.Printf("  ❌ '%s': Invalid\n", test["text"])
		}
	}

	// Test text_match_with_number format
	fmt.Println("\n🔢 Text Match With Number Format Validation:")
	textNumberSchema := `{
		"type": "object",
		"properties": {
			"text": {"type": "string", "format": "text_match_with_number"}
		},
		"required": ["text"]
	}`

	textNumberTests := []map[string]interface{}{
		{"text": "Hello World"},      // Valid text
		{"text": "Text123"},          // Valid with numbers
		{"text": "User_ID_123"},      // Valid with underscore and numbers
		{"text": "Text with spaces"}, // Valid with spaces
		{"text": "Text@special"},     // Invalid with special chars
	}

	for _, test := range textNumberTests {
		result := validator.ValidateSchema(ctx, test, textNumberSchema)
		if result.Valid {
			fmt.Printf("  ✅ '%s': Valid\n", test["text"])
		} else {
			fmt.Printf("  ❌ '%s': Invalid\n", test["text"])
		}
	}

	// Test strong_name format
	fmt.Println("\n💪 Strong Name Format Validation:")
	strongNameSchema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string", "format": "strong_name"}
		},
		"required": ["name"]
	}`

	strongNameTests := []map[string]interface{}{
		{"name": "validName"},    // Valid identifier
		{"name": "validName123"}, // Valid with numbers
		{"name": "valid_name"},   // Valid with underscore
		{"name": "valid-name"},   // Valid with hyphen
		{"name": "123invalid"},   // Invalid - starts with number
		{"name": "_invalid"},     // Invalid - starts with underscore
		{"name": "-invalid"},     // Invalid - starts with hyphen
		{"name": ""},             // Invalid - empty
		{"name": "invalid name"}, // Invalid - contains space
	}

	for _, test := range strongNameTests {
		result := validator.ValidateSchema(ctx, test, strongNameSchema)
		if result.Valid {
			fmt.Printf("  ✅ '%s': Valid\n", test["name"])
		} else {
			fmt.Printf("  ❌ '%s': Invalid\n", test["name"])
		}
	}

	// Test json_number format
	fmt.Println("\n🔢 JSON Number Format Validation:")
	jsonNumberSchema := `{
		"type": "object",
		"properties": {
			"number": {"type": "string", "format": "json_number"}
		},
		"required": ["number"]
	}`

	jsonNumberTests := []map[string]interface{}{
		{"number": "123.45"},       // Valid decimal
		{"number": "123"},          // Valid integer
		{"number": "-123.45"},      // Valid negative
		{"number": "0"},            // Valid zero
		{"number": "not-a-number"}, // Invalid
	}

	for _, test := range jsonNumberTests {
		result := validator.ValidateSchema(ctx, test, jsonNumberSchema)
		if result.Valid {
			fmt.Printf("  ✅ %s: Valid\n", test["number"])
		} else {
			fmt.Printf("  ❌ %s: Invalid\n", test["number"])
		}
	}

	// Test decimal format
	fmt.Println("\n💰 Decimal Format Validation:")
	decimalSchema := `{
		"type": "object",
		"properties": {
			"price": {"type": "string", "format": "decimal"}
		},
		"required": ["price"]
	}`

	decimalTests := []map[string]interface{}{
		{"price": "29.99"},   // Valid decimal
		{"price": "100"},     // Valid integer
		{"price": "0.01"},    // Valid small decimal
		{"price": "-50.25"},  // Valid negative
		{"price": "invalid"}, // Invalid
	}

	for _, test := range decimalTests {
		result := validator.ValidateSchema(ctx, test, decimalSchema)
		if result.Valid {
			fmt.Printf("  ✅ %s: Valid\n", test["price"])
		} else {
			fmt.Printf("  ❌ %s: Invalid\n", test["price"])
		}
	}

	// Test decimal_by_factor_of_8 format
	fmt.Println("\n🎯 Decimal by Factor of 8 Format Validation:")
	decimalFactorSchema := `{
		"type": "object",
		"properties": {
			"precise_value": {"type": "string", "format": "decimal_by_factor_of_8"}
		},
		"required": ["precise_value"]
	}`

	factorTests := []map[string]interface{}{
		{"precise_value": "10.00000000"},  // 8 decimal places - Valid
		{"precise_value": "25.12345678"},  // 8 decimal places - Valid
		{"precise_value": "100"},          // Integer - Valid
		{"precise_value": "10.123456789"}, // 9 decimal places - Invalid
		{"precise_value": "invalid"},      // Invalid
	}

	for _, test := range factorTests {
		result := validator.ValidateSchema(ctx, test, decimalFactorSchema)
		if result.Valid {
			fmt.Printf("  ✅ %s: Valid\n", test["precise_value"])
		} else {
			fmt.Printf("  ❌ %s: Invalid\n", test["precise_value"])
		}
	}

	// Test empty_string format
	fmt.Println("\n📭 Empty String Format Validation:")
	emptyStringSchema := `{
		"type": "object",
		"properties": {
			"optional_field": {"type": "string", "format": "empty_string"}
		},
		"required": ["optional_field"]
	}`

	emptyStringTests := []map[string]interface{}{
		{"optional_field": ""},          // Valid empty string
		{"optional_field": "   "},       // Valid whitespace only
		{"optional_field": "not empty"}, // Invalid - not empty
	}

	for _, test := range emptyStringTests {
		result := validator.ValidateSchema(ctx, test, emptyStringSchema)
		if result.Valid {
			fmt.Printf("  ✅ '%s': Valid\n", test["optional_field"])
		} else {
			fmt.Printf("  ❌ '%s': Invalid\n", test["optional_field"])
		}
	}
}

// advancedSchemaExamples demonstrates complex JSON Schema features
func advancedSchemaExamples() {
	ctx := context.Background()
	validator := schema.NewJSONSchemaValidator()

	fmt.Println("Advanced JSON Schema validation features:")

	// Conditional validation
	fmt.Println("\n🔀 Conditional Validation (if/then/else):")
	conditionalSchema := `{
		"type": "object",
		"properties": {
			"type": {"type": "string", "enum": ["admin", "user", "guest"]},
			"permissions": {"type": "array", "items": {"type": "string"}},
			"email": {"type": "string", "format": "email"},
			"temporary": {"type": "boolean"}
		},
		"required": ["type"],
		"if": {
			"properties": {"type": {"const": "admin"}}
		},
		"then": {
			"required": ["permissions", "email"]
		},
		"else": {
			"if": {
				"properties": {"type": {"const": "guest"}}
			},
			"then": {
				"required": ["temporary"]
			}
		}
	}`

	conditionalTests := []map[string]interface{}{
		{
			"type":        "admin",
			"permissions": []string{"read", "write", "delete"},
			"email":       "admin@example.com",
		}, // Valid admin
		{
			"type":  "admin",
			"email": "admin@example.com",
			// Missing permissions - Invalid
		},
		{
			"type":      "guest",
			"temporary": true,
		}, // Valid guest
		{
			"type": "user",
		}, // Valid user (no extra requirements)
	}

	for i, test := range conditionalTests {
		result := validator.ValidateSchema(ctx, test, conditionalSchema)
		if result.Valid {
			fmt.Printf("  ✅ Test %d (%s): Valid\n", i+1, test["type"])
		} else {
			fmt.Printf("  ❌ Test %d (%s): Invalid - %s\n", i+1, test["type"], result.String())
		}
	}

	// Complex nested objects
	fmt.Println("\n🏗️ Complex Nested Object Validation:")
	nestedSchema := `{
		"type": "object",
		"properties": {
			"user": {
				"type": "object",
				"properties": {
					"profile": {
						"type": "object",
						"properties": {
							"name": {"type": "string", "minLength": 1},
							"age": {"type": "integer", "minimum": 0, "maximum": 150},
							"addresses": {
								"type": "array",
								"items": {
									"type": "object",
									"properties": {
										"street": {"type": "string", "minLength": 1},
										"city": {"type": "string", "minLength": 1},
										"zipcode": {"type": "string", "pattern": "^[0-9]{5}(-[0-9]{4})?$"}
									},
									"required": ["street", "city", "zipcode"]
								},
								"minItems": 1
							}
						},
						"required": ["name", "age", "addresses"]
					}
				},
				"required": ["profile"]
			}
		},
		"required": ["user"]
	}`

	nestedTest := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "John Doe",
				"age":  30,
				"addresses": []interface{}{
					map[string]interface{}{
						"street":  "123 Main St",
						"city":    "New York",
						"zipcode": "10001",
					},
					map[string]interface{}{
						"street":  "456 Oak Ave",
						"city":    "Boston",
						"zipcode": "02101-1234",
					},
				},
			},
		},
	}

	result := validator.ValidateSchema(ctx, nestedTest, nestedSchema)
	if result.Valid {
		fmt.Println("  ✅ Complex nested object: Valid")
	} else {
		fmt.Printf("  ❌ Complex nested object: Invalid - %s\n", result.String())
	}

	// Array validation with different constraints
	fmt.Println("\n📊 Array Validation with Constraints:")
	arraySchema := `{
		"type": "object",
		"properties": {
			"tags": {
				"type": "array",
				"items": {"type": "string", "minLength": 1},
				"minItems": 2,
				"maxItems": 5,
				"uniqueItems": true
			},
			"scores": {
				"type": "array",
				"items": {"type": "number", "minimum": 0, "maximum": 100},
				"minItems": 3
			}
		},
		"required": ["tags", "scores"]
	}`

	arrayTests := []map[string]interface{}{
		{
			"tags":   []string{"important", "urgent", "business"},
			"scores": []float64{85.5, 92.0, 78.5},
		}, // Valid
		{
			"tags":   []string{"tag1"},
			"scores": []float64{85.5, 92.0, 78.5},
		}, // Invalid - not enough tags
		{
			"tags":   []string{"tag1", "tag2", "tag1"},
			"scores": []float64{85.5, 92.0, 78.5},
		}, // Invalid - duplicate tags
	}

	for i, test := range arrayTests {
		result := validator.ValidateSchema(ctx, test, arraySchema)
		if result.Valid {
			fmt.Printf("  ✅ Array test %d: Valid\n", i+1)
		} else {
			fmt.Printf("  ❌ Array test %d: Invalid - %s\n", i+1, result.String())
		}
	}
}

// contextAndPerformanceExamples demonstrates context usage and performance considerations
func contextAndPerformanceExamples() {
	fmt.Println("Context and performance examples:")

	// Timeout validation
	fmt.Println("\n⏰ Validation with Timeout:")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	validator := schema.NewJSONSchemaValidator()
	testSchema := `{"type": "object", "properties": {"name": {"type": "string"}}}`
	data := map[string]interface{}{"name": "John"}

	result := validator.ValidateSchema(timeoutCtx, data, testSchema)
	if result.Valid {
		fmt.Println("  ✅ Validation completed within timeout")
	} else {
		fmt.Printf("  ❌ Validation failed or timed out: %s\n", result.String())
	}

	// Context cancellation
	fmt.Println("\n🚫 Validation with Cancellation:")
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	// Simulate cancellation after a short time
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancelFunc()
	}()

	result = validator.ValidateSchema(cancelCtx, data, testSchema)
	if result.Valid {
		fmt.Println("  ✅ Validation completed before cancellation")
	} else {
		fmt.Println("  ⚠️ Validation may have been cancelled or had errors")
	}

	// Validator reuse (performance best practice)
	fmt.Println("\n🚀 Validator Reuse for Performance:")
	// Reusing the same validator instance is more efficient
	reusableValidator := schema.NewJSONSchemaValidator()

	// Add custom formats once
	reusableValidator.RegisterFormatValidator("custom-id", func(input interface{}) bool {
		if str, ok := input.(string); ok {
			return len(str) == 10 && str[:3] == "ID_"
		}
		return false
	})

	customTestSchema := `{
		"type": "object",
		"properties": {
			"id": {"type": "string", "format": "custom-id"},
			"name": {"type": "string"}
		},
		"required": ["id", "name"]
	}`

	testCases := []map[string]interface{}{
		{"id": "ID_1234567", "name": "Item 1"},
		{"id": "ID_9876543", "name": "Item 2"},
		{"id": "INVALID123", "name": "Item 3"},
	}

	for i, testCase := range testCases {
		start := time.Now()
		result := reusableValidator.ValidateSchema(context.Background(), testCase, customTestSchema)
		duration := time.Since(start)

		if result.Valid {
			fmt.Printf("  ✅ Test %d: Valid (took %v)\n", i+1, duration)
		} else {
			fmt.Printf("  ❌ Test %d: Invalid (took %v)\n", i+1, duration)
		}
	}
}

// domainErrorExamples demonstrates integration with nexs-lib error domain
func domainErrorExamples() {
	ctx := context.Background()

	fmt.Println("Domain error integration examples:")

	// Standard validation with detailed error handling
	fmt.Println("\n🚨 Detailed Error Handling:")
	validator := schema.NewJSONSchemaValidator()

	schemaStr := `{
		"type": "object",
		"properties": {
			"email": {"type": "string", "format": "email"},
			"age": {"type": "integer", "minimum": 18}
		},
		"required": ["email", "age"]
	}`

	validData := map[string]interface{}{
		"email": "valid@example.com",
		"age":   25,
	}

	invalidData := map[string]interface{}{
		"email": "invalid-email",
		"age":   15,
	}

	// Test with valid data
	result := validator.ValidateSchema(ctx, validData, schemaStr)
	if result.Valid {
		fmt.Println("  ✅ Valid data passed validation")
	} else {
		fmt.Printf("  ❌ Valid data failed: %s\n", result.String())
	}

	// Test with invalid data
	result = validator.ValidateSchema(ctx, invalidData, schemaStr)
	if !result.Valid {
		fmt.Println("  ✅ Invalid data correctly failed validation:")
		for field, errors := range result.Errors {
			for _, err := range errors {
				fmt.Printf("    - %s: %s\n", field, err)
			}
		}
	} else {
		fmt.Println("  ❌ Invalid data unexpectedly passed")
	}

	// Regular validation result details
	fmt.Println("\n📋 Validation Result Analysis:")

	result = validator.ValidateSchema(ctx, invalidData, schemaStr)
	fmt.Printf("  📄 Result - Valid: %t, Error Count: %d\n", result.Valid, result.ErrorCount())
	fmt.Printf("  📄 Has Errors: %t\n", result.HasErrors())
	fmt.Printf("  📄 First Error: %s\n", result.FirstError())
	fmt.Printf("  📄 All Errors: %v\n", result.AllErrors())
}

// batchValidationExamples demonstrates batch processing and bulk validation
func batchValidationExamples() {
	ctx := context.Background()

	fmt.Println("Batch validation examples:")

	// Batch validation of multiple users
	fmt.Println("\n👥 Batch User Validation:")
	validator := schema.NewJSONSchemaValidator()

	userSchema := `{
		"type": "object",
		"properties": {
			"id": {"type": "integer", "minimum": 1},
			"name": {"type": "string", "minLength": 2, "maxLength": 50},
			"email": {"type": "string", "format": "email"},
			"age": {"type": "integer", "minimum": 18, "maximum": 120}
		},
		"required": ["id", "name", "email", "age"]
	}`

	users := []map[string]interface{}{
		{"id": 1, "name": "John Doe", "email": "john@example.com", "age": 30},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com", "age": 25},
		{"id": 3, "name": "J", "email": "invalid-email", "age": 15}, // Invalid
		{"id": 4, "name": "Bob Johnson", "email": "bob@example.com", "age": 45},
		{"id": 5, "name": "", "email": "empty@example.com", "age": 35}, // Invalid
	}

	validCount := 0
	invalidCount := 0
	invalidUsers := make(map[int]*schema.ValidationResult)

	for i, user := range users {
		result := validator.ValidateSchema(ctx, user, userSchema)
		if result.Valid {
			validCount++
			fmt.Printf("  ✅ User %d (%s): Valid\n", i+1, user["name"])
		} else {
			invalidCount++
			invalidUsers[i] = result
			fmt.Printf("  ❌ User %d (%s): Invalid\n", i+1, user["name"])
		}
	}

	fmt.Printf("\n📊 Batch Results: %d valid, %d invalid out of %d total\n",
		validCount, invalidCount, len(users))

	// Show detailed errors for invalid users
	fmt.Println("\n🔍 Detailed Errors for Invalid Users:")
	for userIndex, result := range invalidUsers {
		fmt.Printf("  User %d errors:\n", userIndex+1)
		for field, errors := range result.Errors {
			for _, err := range errors {
				fmt.Printf("    - %s: %s\n", field, err)
			}
		}
	}

	// Performance measurement for batch validation
	fmt.Println("\n⚡ Batch Performance Measurement:")
	start := time.Now()

	// Validate a larger batch
	largeBatch := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeBatch[i] = map[string]interface{}{
			"id":    i + 1,
			"name":  fmt.Sprintf("User %d", i+1),
			"email": fmt.Sprintf("user%d@example.com", i+1),
			"age":   25 + (i % 50),
		}
	}

	validatedCount := 0
	for _, user := range largeBatch {
		result := validator.ValidateSchema(ctx, user, userSchema)
		if result.Valid {
			validatedCount++
		}
	}

	duration := time.Since(start)
	fmt.Printf("  📈 Validated %d users in %v (%.2f users/ms)\n",
		len(largeBatch), duration, float64(len(largeBatch))/float64(duration.Milliseconds()))
}
