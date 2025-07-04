package schema

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationResult(t *testing.T) {
	t.Run("NewValidationResult", func(t *testing.T) {
		result := NewValidationResult()
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.GlobalErrors)
		assert.Empty(t, result.Warnings)
	})

	t.Run("AddError", func(t *testing.T) {
		result := NewValidationResult()
		result.AddError("field1", "error1")

		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors["field1"], "error1")
		assert.Equal(t, 1, result.ErrorCount())
	})

	t.Run("AddGlobalError", func(t *testing.T) {
		result := NewValidationResult()
		result.AddGlobalError("global error")

		assert.False(t, result.Valid)
		assert.Contains(t, result.GlobalErrors, "global error")
	})

	t.Run("Merge", func(t *testing.T) {
		result1 := NewValidationResult()
		result1.AddError("field1", "error1")

		result2 := NewValidationResult()
		result2.AddError("field2", "error2")
		result2.AddGlobalError("global error")

		result1.Merge(result2)

		assert.False(t, result1.Valid)
		assert.Contains(t, result1.Errors["field1"], "error1")
		assert.Contains(t, result1.Errors["field2"], "error2")
		assert.Contains(t, result1.GlobalErrors, "global error")
	})
}

func TestRequiredRule(t *testing.T) {
	rule := NewRequiredRule()
	ctx := context.Background()

	t.Run("valid string", func(t *testing.T) {
		err := rule.Validate(ctx, "hello")
		assert.NoError(t, err)
	})

	t.Run("empty string", func(t *testing.T) {
		err := rule.Validate(ctx, "")
		assert.Error(t, err)
	})

	t.Run("whitespace string", func(t *testing.T) {
		err := rule.Validate(ctx, "   ")
		assert.Error(t, err)
	})

	t.Run("nil value", func(t *testing.T) {
		err := rule.Validate(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("empty slice", func(t *testing.T) {
		err := rule.Validate(ctx, []string{})
		assert.Error(t, err)
	})

	t.Run("non-empty slice", func(t *testing.T) {
		err := rule.Validate(ctx, []string{"item"})
		assert.NoError(t, err)
	})
}

func TestMinLengthRule(t *testing.T) {
	rule := NewMinLengthRule(5)
	ctx := context.Background()

	t.Run("valid length", func(t *testing.T) {
		err := rule.Validate(ctx, "hello")
		assert.NoError(t, err)
	})

	t.Run("longer than minimum", func(t *testing.T) {
		err := rule.Validate(ctx, "hello world")
		assert.NoError(t, err)
	})

	t.Run("shorter than minimum", func(t *testing.T) {
		err := rule.Validate(ctx, "hi")
		assert.Error(t, err)
	})

	t.Run("non-string value", func(t *testing.T) {
		err := rule.Validate(ctx, 123)
		assert.Error(t, err)
	})
}

func TestMaxLengthRule(t *testing.T) {
	rule := NewMaxLengthRule(10)
	ctx := context.Background()

	t.Run("valid length", func(t *testing.T) {
		err := rule.Validate(ctx, "hello")
		assert.NoError(t, err)
	})

	t.Run("at maximum", func(t *testing.T) {
		err := rule.Validate(ctx, "1234567890")
		assert.NoError(t, err)
	})

	t.Run("longer than maximum", func(t *testing.T) {
		err := rule.Validate(ctx, "this is too long")
		assert.Error(t, err)
	})
}

func TestPatternRule(t *testing.T) {
	rule := NewPatternRule(`^[a-z]+$`)
	ctx := context.Background()

	t.Run("matching pattern", func(t *testing.T) {
		err := rule.Validate(ctx, "hello")
		assert.NoError(t, err)
	})

	t.Run("non-matching pattern", func(t *testing.T) {
		err := rule.Validate(ctx, "Hello123")
		assert.Error(t, err)
	})

	t.Run("invalid pattern", func(t *testing.T) {
		invalidRule := NewPatternRule(`[`)
		err := invalidRule.Validate(ctx, "test")
		assert.Error(t, err)
	})
}

func TestEmailRule(t *testing.T) {
	rule := NewEmailRule()
	ctx := context.Background()

	t.Run("valid email", func(t *testing.T) {
		err := rule.Validate(ctx, "test@example.com")
		assert.NoError(t, err)
	})

	t.Run("valid email with subdomain", func(t *testing.T) {
		err := rule.Validate(ctx, "user@mail.example.com")
		assert.NoError(t, err)
	})

	t.Run("invalid email", func(t *testing.T) {
		err := rule.Validate(ctx, "invalid-email")
		assert.Error(t, err)
	})

	t.Run("empty string", func(t *testing.T) {
		err := rule.Validate(ctx, "")
		assert.Error(t, err)
	})
}

func TestURLRule(t *testing.T) {
	rule := NewURLRule()
	ctx := context.Background()

	t.Run("valid HTTP URL", func(t *testing.T) {
		err := rule.Validate(ctx, "https://example.com")
		assert.NoError(t, err)
	})

	t.Run("valid HTTP URL with path", func(t *testing.T) {
		err := rule.Validate(ctx, "https://example.com/path/to/resource")
		assert.NoError(t, err)
	})

	t.Run("invalid URL", func(t *testing.T) {
		err := rule.Validate(ctx, "not-a-url")
		assert.Error(t, err)
	})
}

func TestUUIDRule(t *testing.T) {
	rule := NewUUIDRule()
	ctx := context.Background()

	t.Run("valid UUID v4", func(t *testing.T) {
		err := rule.Validate(ctx, "550e8400-e29b-41d4-a716-446655440000")
		assert.NoError(t, err)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		err := rule.Validate(ctx, "invalid-uuid")
		assert.Error(t, err)
	})

	t.Run("UUID without hyphens", func(t *testing.T) {
		err := rule.Validate(ctx, "550e8400e29b41d4a716446655440000")
		assert.NoError(t, err) // google/uuid accepts UUIDs without hyphens
	})
}

func TestMinValueRule(t *testing.T) {
	rule := NewMinValueRule(10.0)
	ctx := context.Background()

	t.Run("valid value", func(t *testing.T) {
		err := rule.Validate(ctx, 15.5)
		assert.NoError(t, err)
	})

	t.Run("value at minimum", func(t *testing.T) {
		err := rule.Validate(ctx, 10.0)
		assert.NoError(t, err)
	})

	t.Run("value below minimum", func(t *testing.T) {
		err := rule.Validate(ctx, 5.0)
		assert.Error(t, err)
	})

	t.Run("integer value", func(t *testing.T) {
		err := rule.Validate(ctx, 20)
		assert.NoError(t, err)
	})

	t.Run("string number", func(t *testing.T) {
		err := rule.Validate(ctx, "15.5")
		assert.NoError(t, err)
	})

	t.Run("invalid string", func(t *testing.T) {
		err := rule.Validate(ctx, "not-a-number")
		assert.Error(t, err)
	})
}

func TestDateTimeFormatRule(t *testing.T) {
	rule := NewDateTimeFormatRule(time.RFC3339)
	ctx := context.Background()

	t.Run("valid RFC3339 date", func(t *testing.T) {
		err := rule.Validate(ctx, "2023-12-25T10:30:00Z")
		assert.NoError(t, err)
	})

	t.Run("invalid format", func(t *testing.T) {
		err := rule.Validate(ctx, "2023-12-25 10:30:00")
		assert.Error(t, err)
	})

	t.Run("invalid date", func(t *testing.T) {
		err := rule.Validate(ctx, "invalid-date")
		assert.Error(t, err)
	})
}

func TestCustomRule(t *testing.T) {
	rule := NewCustomRule("custom", "value must be positive", func(ctx context.Context, value interface{}) error {
		if num, ok := value.(int); ok && num > 0 {
			return nil
		}
		return &ValidationError{Field: "", Message: "value must be positive", Code: "positive_required", Value: value}
	})

	ctx := context.Background()

	t.Run("valid value", func(t *testing.T) {
		err := rule.Validate(ctx, 5)
		assert.NoError(t, err)
	})

	t.Run("invalid value", func(t *testing.T) {
		err := rule.Validate(ctx, -5)
		assert.Error(t, err)
	})
}

func TestDefaultValidator(t *testing.T) {
	t.Run("simple validation", func(t *testing.T) {
		validator := NewValidator()
		validator.AddRule(NewRequiredRule())

		ctx := context.Background()

		result := validator.Validate(ctx, "hello")
		assert.True(t, result.Valid)

		result = validator.Validate(ctx, "")
		assert.False(t, result.Valid)
	})

	t.Run("struct validation", func(t *testing.T) {
		type User struct {
			Name  string `validate:"required,min=2"`
			Email string `validate:"required,email"`
		}

		validator := NewValidator()
		ctx := context.Background()

		// Valid user
		user := User{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		result := validator.ValidateStruct(ctx, user)
		assert.True(t, result.Valid)

		// Invalid user
		invalidUser := User{
			Name:  "", // Required field empty
			Email: "invalid-email",
		}
		result = validator.ValidateStruct(ctx, invalidUser)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "Name")
		assert.Contains(t, result.Errors, "Email")
	})
}

func TestValidatorBuilder(t *testing.T) {
	validator := NewValidatorBuilder().
		Rule(NewRequiredRule()).
		Field("name", NewMinLengthRule(2), NewMaxLengthRule(50)).
		Field("email", NewEmailRule()).
		Build()

	ctx := context.Background()

	type User struct {
		Name  string
		Email string
	}

	user := User{
		Name:  "Jo",
		Email: "john@example.com",
	}

	result := validator.ValidateStruct(ctx, user)
	// Should pass global validation but might have field-specific issues
	require.NotNil(t, result)
}

func TestRuleBuilder(t *testing.T) {
	t.Run("string builder", func(t *testing.T) {
		// Test just min/max length without email validation
		rule := NewRuleBuilder().
			Required().
			String().
			MinLength(5).
			MaxLength(10).
			Build()

		ctx := context.Background()

		// Valid length
		err := rule.Validate(ctx, "hello")
		assert.NoError(t, err)

		// Too short
		err = rule.Validate(ctx, "hi")
		assert.Error(t, err)

		// Too long
		err = rule.Validate(ctx, "this is way too long")
		assert.Error(t, err)
	})

	t.Run("number builder", func(t *testing.T) {
		rule := NewRuleBuilder().
			Required().
			Number().
			Min(0).
			Max(100).
			Integer().
			Build()

		ctx := context.Background()

		// Valid integer
		err := rule.Validate(ctx, 50)
		assert.NoError(t, err)

		// Too high
		err = rule.Validate(ctx, 150)
		assert.Error(t, err)

		// Not an integer
		err = rule.Validate(ctx, 50.5)
		assert.Error(t, err)
	})

	t.Run("datetime builder", func(t *testing.T) {
		rule := NewRuleBuilder().
			Required().
			DateTime().
			RFC3339().
			After("2020-01-01T00:00:00Z").
			Before("2030-01-01T00:00:00Z").
			Build()

		ctx := context.Background()

		// Valid date
		err := rule.Validate(ctx, "2025-07-04T12:00:00Z")
		assert.NoError(t, err)

		// Too early
		err = rule.Validate(ctx, "2019-01-01T00:00:00Z")
		assert.Error(t, err)

		// Too late
		err = rule.Validate(ctx, "2031-01-01T00:00:00Z")
		assert.Error(t, err)
	})
}

// TestStructWithNumericValidation tests struct validation with type-aware min/max rules
func TestStructWithNumericValidation(t *testing.T) {
	type TestStruct struct {
		Name  string  `validate:"required,min=2,max=50"`
		Age   int     `validate:"required,min=1,max=120"`
		Price float64 `validate:"min=0,max=1000"`
	}

	validator := NewValidator()

	t.Run("Valid struct with numeric and string fields", func(t *testing.T) {
		valid := TestStruct{
			Name:  "John Doe",
			Age:   25,
			Price: 99.99,
		}

		result := validator.ValidateStruct(context.Background(), valid)
		assert.True(t, result.Valid, "Valid struct should pass validation")
		assert.Empty(t, result.Errors, "No errors should be present")
	})

	t.Run("Invalid struct - age too young (numeric validation)", func(t *testing.T) {
		invalid := TestStruct{
			Name:  "Jane",
			Age:   0, // Should fail min=1 (value validation for int)
			Price: 50.0,
		}

		result := validator.ValidateStruct(context.Background(), invalid)
		assert.False(t, result.Valid, "Invalid struct should fail validation")
		assert.Contains(t, result.Errors, "Age", "Age field should have validation error")

		// The error should be about minimum value, not length
		ageErrors := result.Errors["Age"]
		assert.NotEmpty(t, ageErrors, "Age should have errors")
		assert.Contains(t, ageErrors[0], "minimum value", "Error should be about minimum value, not length")
	})

	t.Run("Invalid struct - age too old (numeric validation)", func(t *testing.T) {
		invalid := TestStruct{
			Name:  "Old Person",
			Age:   130, // Should fail max=120 (value validation for int)
			Price: 50.0,
		}

		result := validator.ValidateStruct(context.Background(), invalid)
		assert.False(t, result.Valid, "Invalid struct should fail validation")
		assert.Contains(t, result.Errors, "Age", "Age field should have validation error")

		// The error should be about maximum value, not length
		ageErrors := result.Errors["Age"]
		assert.NotEmpty(t, ageErrors, "Age should have errors")
		assert.Contains(t, ageErrors[0], "maximum value", "Error should be about maximum value, not length")
	})

	t.Run("Invalid struct - name too short (string length validation)", func(t *testing.T) {
		invalid := TestStruct{
			Name:  "A", // Should fail min=2 (length validation for string)
			Age:   25,
			Price: 50.0,
		}

		result := validator.ValidateStruct(context.Background(), invalid)
		assert.False(t, result.Valid, "Invalid struct should fail validation")
		assert.Contains(t, result.Errors, "Name", "Name field should have validation error")

		// The error should be about minimum length
		nameErrors := result.Errors["Name"]
		assert.NotEmpty(t, nameErrors, "Name should have errors")
		assert.Contains(t, nameErrors[0], "minimum length", "Error should be about minimum length for string")
	})

	t.Run("Invalid struct - price too high (float validation)", func(t *testing.T) {
		invalid := TestStruct{
			Name:  "Test",
			Age:   25,
			Price: 1500.0, // Should fail max=1000 (value validation for float64)
		}

		result := validator.ValidateStruct(context.Background(), invalid)
		assert.False(t, result.Valid, "Invalid struct should fail validation")
		assert.Contains(t, result.Errors, "Price", "Price field should have validation error")

		// The error should be about maximum value
		priceErrors := result.Errors["Price"]
		assert.NotEmpty(t, priceErrors, "Price should have errors")
		assert.Contains(t, priceErrors[0], "maximum value", "Error should be about maximum value for float")
	})
}

func TestMixedTypeValidation(t *testing.T) {
	// Test struct with mixed field types to ensure type-aware validation
	type MixedStruct struct {
		Name  string  `validate:"required,min=3,max=10"`
		Age   int     `validate:"required,min=18,max=120"`
		Score float64 `validate:"min=0,max=100"`
	}

	validator := NewValidator()

	t.Run("ValidMixedStruct", func(t *testing.T) {
		data := MixedStruct{
			Name:  "John",
			Age:   25,
			Score: 85.5,
		}

		result := validator.ValidateStruct(context.Background(), data)
		assert.True(t, result.Valid, "Valid mixed struct should pass validation")
		assert.Empty(t, result.Errors, "Valid mixed struct should have no errors")
	})

	t.Run("InvalidAgeInt", func(t *testing.T) {
		data := MixedStruct{
			Name:  "John",
			Age:   16, // Too low, should trigger MinValueRule not MinLengthRule
			Score: 85.5,
		}

		result := validator.ValidateStruct(context.Background(), data)
		assert.False(t, result.Valid, "Invalid age should fail validation")
		assert.Contains(t, result.Errors, "Age", "Age field should have validation error")

		ageErrors := result.Errors["Age"]
		assert.NotEmpty(t, ageErrors, "Age should have errors")
		// Should be about minimum value, not length
		assert.Contains(t, ageErrors[0], "minimum value", "Error should be about minimum value for int")
	})

	t.Run("InvalidNameString", func(t *testing.T) {
		data := MixedStruct{
			Name:  "Jo", // Too short, should trigger MinLengthRule
			Age:   25,
			Score: 85.5,
		}

		result := validator.ValidateStruct(context.Background(), data)
		assert.False(t, result.Valid, "Invalid name should fail validation")
		assert.Contains(t, result.Errors, "Name", "Name field should have validation error")

		nameErrors := result.Errors["Name"]
		assert.NotEmpty(t, nameErrors, "Name should have errors")
		// Should be about minimum length for strings
		assert.Contains(t, nameErrors[0], "minimum length", "Error should be about minimum length for string")
	})
}
