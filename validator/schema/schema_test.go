package schema

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSchemaValidator(t *testing.T) {
	validator := NewJSONSchemaValidator()
	ctx := context.Background()

	t.Run("valid object", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
		}

		schema := `{
			"type": "object",
			"properties": {
				"name": {"type": "string", "minLength": 1},
				"age": {"type": "integer", "minimum": 0},
				"email": {"type": "string", "format": "email"}
			},
			"required": ["name", "age"]
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("invalid object - missing required field", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "John Doe",
			// missing age
		}

		schema := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"age": {"type": "integer", "minimum": 0}
			},
			"required": ["name", "age"]
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("invalid object - wrong type", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "John Doe",
			"age":  "thirty", // should be integer
		}

		schema := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"age": {"type": "integer"}
			},
			"required": ["name", "age"]
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "age")
	})

	t.Run("invalid schema", func(t *testing.T) {
		data := map[string]interface{}{"name": "John"}
		invalidSchema := `{"type": "invalid"}`

		result := validator.ValidateSchema(ctx, data, invalidSchema)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.GlobalErrors)
	})

	t.Run("complex nested object", func(t *testing.T) {
		data := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "John Doe",
				"contacts": map[string]interface{}{
					"email": "john@example.com",
					"phone": "+1234567890",
				},
			},
			"preferences": []string{"email", "sms"},
		}

		schema := `{
			"type": "object",
			"properties": {
				"user": {
					"type": "object",
					"properties": {
						"name": {"type": "string", "minLength": 1},
						"contacts": {
							"type": "object",
							"properties": {
								"email": {"type": "string", "format": "email"},
								"phone": {"type": "string", "pattern": "^\\+\\d+$"}
							},
							"required": ["email"]
						}
					},
					"required": ["name", "contacts"]
				},
				"preferences": {
					"type": "array",
					"items": {"type": "string"},
					"minItems": 1
				}
			},
			"required": ["user"]
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.True(t, result.Valid)
	})
}

func TestJSONSchemaValidatorCustomFormats(t *testing.T) {
	validator := NewJSONSchemaValidator()
	ctx := context.Background()

	t.Run("custom format validator", func(t *testing.T) {
		// Add a custom format for credit card numbers
		validator.AddCustomFormat("credit-card", &customRegexFormatValidator{
			name:    "credit-card",
			pattern: mustCompileRegex(`^\d{4}-\d{4}-\d{4}-\d{4}$`),
		})

		data := map[string]interface{}{
			"card_number": "1234-5678-9012-3456",
		}

		schema := `{
			"type": "object",
			"properties": {
				"card_number": {"type": "string", "format": "credit-card"}
			},
			"required": ["card_number"]
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.True(t, result.Valid)

		// Test with invalid format
		invalidData := map[string]interface{}{
			"card_number": "1234567890123456", // missing hyphens
		}

		result = validator.ValidateSchema(ctx, invalidData, schema)
		assert.False(t, result.Valid)
	})

	t.Run("register format validator function", func(t *testing.T) {
		validator.RegisterFormatValidator("positive-number", func(input interface{}) bool {
			if str, ok := input.(string); ok {
				if num, err := parseFloat(str); err == nil {
					return num > 0
				}
			}
			return false
		})

		data := map[string]interface{}{
			"amount": "123.45",
		}

		schema := `{
			"type": "object",
			"properties": {
				"amount": {"type": "string", "format": "positive-number"}
			}
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.True(t, result.Valid)

		// Test with negative number
		negativeData := map[string]interface{}{
			"amount": "-123.45",
		}

		result = validator.ValidateSchema(ctx, negativeData, schema)
		assert.False(t, result.Valid)
	})
}

func TestJSONSchemaValidatorDefaultFormats(t *testing.T) {
	validator := NewJSONSchemaValidator()
	ctx := context.Background()

	t.Run("date_time format", func(t *testing.T) {
		data := map[string]interface{}{
			"created_at": "2023-12-25T10:30:00Z",
		}

		schema := `{
			"type": "object",
			"properties": {
				"created_at": {"type": "string", "format": "date_time"}
			}
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.True(t, result.Valid)
	})

	t.Run("iso_8601_date format", func(t *testing.T) {
		data := map[string]interface{}{
			"birth_date": "1990-05-15",
		}

		schema := `{
			"type": "object",
			"properties": {
				"birth_date": {"type": "string", "format": "iso_8601_date"}
			}
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.True(t, result.Valid)
	})

	t.Run("strong_name format", func(t *testing.T) {
		data := map[string]interface{}{
			"variable_name": "myVariable123",
		}

		schema := `{
			"type": "object",
			"properties": {
				"variable_name": {"type": "string", "format": "strong_name"}
			}
		}`

		result := validator.ValidateSchema(ctx, data, schema)
		assert.True(t, result.Valid)
	})
}

func TestSchemaValidationError(t *testing.T) {
	result := NewValidationResult()
	result.AddError("field1", "error1")
	result.AddError("field1", "error2")
	result.AddError("field2", "error3")
	result.AddGlobalError("global error")

	schemaError := NewSchemaValidationError(result)

	require.NotNil(t, schemaError)
	assert.Contains(t, schemaError.Details["field1"], "error1")
	assert.Contains(t, schemaError.Details["field1"], "error2")
	assert.Contains(t, schemaError.Details["field2"], "error3")
	assert.Contains(t, schemaError.Details["_global"], "global error")
}

func TestValidateWithDomainError(t *testing.T) {
	validator := NewJSONSchemaValidator()
	ctx := context.Background()

	t.Run("valid data returns no error", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "John",
		}

		schema := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"}
			}
		}`

		// Cast to concrete type to access the method
		if jsv, ok := validator.(*jsonSchemaValidator); ok {
			err := jsv.ValidateWithDomainError(ctx, data, schema)
			assert.NoError(t, err)
		}
	})

	t.Run("invalid data returns domain error", func(t *testing.T) {
		data := map[string]interface{}{
			"name": 123, // should be string
		}

		schema := `{
			"type": "object",
			"properties": {
				"name": {"type": "string"}
			}
		}`

		// Cast to concrete type to access the method
		if jsv, ok := validator.(*jsonSchemaValidator); ok {
			err := jsv.ValidateWithDomainError(ctx, data, schema)
			assert.Error(t, err)

			schemaErr, ok := err.(*SchemaValidationError)
			assert.True(t, ok)
			assert.NotEmpty(t, schemaErr.Details)
		}
	})
}

func TestAddCustomFormatByRegex(t *testing.T) {
	// This is a global function that adds to gojsonschema
	AddCustomFormatByRegex("test-format", `^test-\d+$`)

	validator := NewJSONSchemaValidator()
	ctx := context.Background()

	data := map[string]interface{}{
		"test_field": "test-123",
	}

	schema := `{
		"type": "object",
		"properties": {
			"test_field": {"type": "string", "format": "test-format"}
		}
	}`

	result := validator.ValidateSchema(ctx, data, schema)
	assert.True(t, result.Valid)

	// Test invalid format
	invalidData := map[string]interface{}{
		"test_field": "invalid-format",
	}

	result = validator.ValidateSchema(ctx, invalidData, schema)
	assert.False(t, result.Valid)
}

// Helper functions for tests
func mustCompileRegex(pattern string) *regexp.Regexp {
	r, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	return r
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
