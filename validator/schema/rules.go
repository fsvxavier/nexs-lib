package schema

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// baseRule provides common functionality for all rules
type baseRule struct {
	name    string
	message string
}

// Name returns the rule name
func (br *baseRule) Name() string {
	return br.name
}

// Message returns the error message template
func (br *baseRule) Message() string {
	return br.message
}

// RequiredRule validates that a value is not nil/empty
type RequiredRule struct {
	*baseRule
}

// NewRequiredRule creates a new required rule
func NewRequiredRule() *RequiredRule {
	return &RequiredRule{
		baseRule: &baseRule{
			name:    "required",
			message: "field is required",
		},
	}
}

// Validate checks if the value is not nil/empty
func (r *RequiredRule) Validate(ctx context.Context, value interface{}) error {
	if value == nil {
		return errors.New(r.message)
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		if strings.TrimSpace(v.String()) == "" {
			return errors.New(r.message)
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if v.Len() == 0 {
			return errors.New(r.message)
		}
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return errors.New(r.message)
		}
	}

	return nil
}

// MinLengthRule validates minimum string length
type MinLengthRule struct {
	*baseRule
	minLength int
}

// NewMinLengthRule creates a new minimum length rule
func NewMinLengthRule(minLength int) *MinLengthRule {
	return &MinLengthRule{
		baseRule: &baseRule{
			name:    "min_length",
			message: fmt.Sprintf("minimum length is %d", minLength),
		},
		minLength: minLength,
	}
}

// Validate checks minimum string length
func (r *MinLengthRule) Validate(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	if len(str) < r.minLength {
		return errors.New(r.message)
	}

	return nil
}

// MaxLengthRule validates maximum string length
type MaxLengthRule struct {
	*baseRule
	maxLength int
}

// NewMaxLengthRule creates a new maximum length rule
func NewMaxLengthRule(maxLength int) *MaxLengthRule {
	return &MaxLengthRule{
		baseRule: &baseRule{
			name:    "max_length",
			message: fmt.Sprintf("maximum length is %d", maxLength),
		},
		maxLength: maxLength,
	}
}

// Validate checks maximum string length
func (r *MaxLengthRule) Validate(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	if len(str) > r.maxLength {
		return errors.New(r.message)
	}

	return nil
}

// PatternRule validates string against regex pattern
type PatternRule struct {
	*baseRule
	pattern *regexp.Regexp
}

// NewPatternRule creates a new pattern rule
func NewPatternRule(pattern string) *PatternRule {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		// Return a rule that always fails if pattern is invalid
		return &PatternRule{
			baseRule: &baseRule{
				name:    "pattern",
				message: fmt.Sprintf("invalid pattern: %s", err.Error()),
			},
			pattern: nil,
		}
	}

	return &PatternRule{
		baseRule: &baseRule{
			name:    "pattern",
			message: fmt.Sprintf("value must match pattern: %s", pattern),
		},
		pattern: regex,
	}
}

// Validate checks if string matches pattern
func (r *PatternRule) Validate(ctx context.Context, value interface{}) error {
	if r.pattern == nil {
		return errors.New(r.message)
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	if !r.pattern.MatchString(str) {
		return errors.New(r.message)
	}

	return nil
}

// EmailRule validates email format
type EmailRule struct {
	*baseRule
}

// NewEmailRule creates a new email rule
func NewEmailRule() *EmailRule {
	return &EmailRule{
		baseRule: &baseRule{
			name:    "email",
			message: "invalid email format",
		},
	}
}

// Validate checks email format
func (r *EmailRule) Validate(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	_, err := mail.ParseAddress(str)
	if err != nil {
		return errors.New(r.message)
	}

	return nil
}

// URLRule validates URL format
type URLRule struct {
	*baseRule
}

// NewURLRule creates a new URL rule
func NewURLRule() *URLRule {
	return &URLRule{
		baseRule: &baseRule{
			name:    "url",
			message: "invalid URL format",
		},
	}
}

// Validate checks URL format
func (r *URLRule) Validate(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	_, err := url.ParseRequestURI(str)
	if err != nil {
		return errors.New(r.message)
	}

	return nil
}

// UUIDRule validates UUID format
type UUIDRule struct {
	*baseRule
}

// NewUUIDRule creates a new UUID rule
func NewUUIDRule() *UUIDRule {
	return &UUIDRule{
		baseRule: &baseRule{
			name:    "uuid",
			message: "invalid UUID format",
		},
	}
}

// Validate checks UUID format
func (r *UUIDRule) Validate(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	_, err := uuid.Parse(str)
	if err != nil {
		return errors.New(r.message)
	}

	return nil
}

// MinValueRule validates minimum numeric value
type MinValueRule struct {
	*baseRule
	minValue float64
}

// NewMinValueRule creates a new minimum value rule
func NewMinValueRule(minValue float64) *MinValueRule {
	return &MinValueRule{
		baseRule: &baseRule{
			name:    "min_value",
			message: fmt.Sprintf("minimum value is %g", minValue),
		},
		minValue: minValue,
	}
}

// Validate checks minimum numeric value
func (r *MinValueRule) Validate(ctx context.Context, value interface{}) error {
	num, err := toFloat64(value)
	if err != nil {
		return fmt.Errorf("value must be numeric")
	}

	if num < r.minValue {
		return errors.New(r.message)
	}

	return nil
}

// MaxValueRule validates maximum numeric value
type MaxValueRule struct {
	*baseRule
	maxValue float64
}

// NewMaxValueRule creates a new maximum value rule
func NewMaxValueRule(maxValue float64) *MaxValueRule {
	return &MaxValueRule{
		baseRule: &baseRule{
			name:    "max_value",
			message: fmt.Sprintf("maximum value is %g", maxValue),
		},
		maxValue: maxValue,
	}
}

// Validate checks maximum numeric value
func (r *MaxValueRule) Validate(ctx context.Context, value interface{}) error {
	num, err := toFloat64(value)
	if err != nil {
		return fmt.Errorf("value must be numeric")
	}

	if num > r.maxValue {
		return errors.New(r.message)
	}

	return nil
}

// DateTimeFormatRule validates date/time format
type DateTimeFormatRule struct {
	*baseRule
	format string
}

// NewDateTimeFormatRule creates a new date/time format rule
func NewDateTimeFormatRule(format string) *DateTimeFormatRule {
	return &DateTimeFormatRule{
		baseRule: &baseRule{
			name:    "datetime_format",
			message: fmt.Sprintf("invalid date/time format, expected: %s", format),
		},
		format: format,
	}
}

// Validate checks date/time format
func (r *DateTimeFormatRule) Validate(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	_, err := time.Parse(r.format, str)
	if err != nil {
		return errors.New(r.message)
	}

	return nil
}

// CustomRule allows for custom validation logic
type CustomRule struct {
	*baseRule
	validator func(context.Context, interface{}) error
}

// NewCustomRule creates a new custom rule
func NewCustomRule(name, message string, validator func(context.Context, interface{}) error) *CustomRule {
	return &CustomRule{
		baseRule: &baseRule{
			name:    name,
			message: message,
		},
		validator: validator,
	}
}

// Validate executes custom validation logic
func (r *CustomRule) Validate(ctx context.Context, value interface{}) error {
	return r.validator(ctx, value)
}

// Helper function to convert various numeric types to float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}
