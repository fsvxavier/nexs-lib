package schema

import (
	"context"
)

// Validator defines the main validation interface
type Validator interface {
	// Validate validates the given value against configured rules
	Validate(ctx context.Context, value interface{}) *ValidationResult

	// ValidateStruct validates a struct with field-level validations
	ValidateStruct(ctx context.Context, s interface{}) *ValidationResult

	// AddRule adds a validation rule to the validator
	AddRule(rule Rule) Validator

	// AddFieldRule adds a field-specific validation rule for struct validation
	AddFieldRule(fieldName string, rule Rule) Validator
}

// Rule defines a validation rule interface
type Rule interface {
	// Validate performs the validation logic
	Validate(ctx context.Context, value interface{}) error

	// Name returns the rule name for error reporting
	Name() string

	// Message returns the error message template
	Message() string
}

// FieldRule defines field-specific validation rules for structs
type FieldRule interface {
	Rule

	// FieldName returns the field name this rule applies to
	FieldName() string

	// IsRequired returns true if the field is required
	IsRequired() bool
}

// ConditionalRule defines rules that may or may not apply based on conditions
type ConditionalRule interface {
	Rule

	// ShouldApply determines if this rule should be applied
	ShouldApply(ctx context.Context, value interface{}) bool
}

// FormatValidator defines interface for format validation (like JSON Schema formats)
type FormatValidator interface {
	// IsFormat checks if the input matches the expected format
	IsFormat(input interface{}) bool

	// FormatName returns the name of the format
	FormatName() string
}

// SchemaValidator defines interface for JSON Schema validation
type SchemaValidator interface {
	// ValidateSchema validates data against a JSON schema
	ValidateSchema(ctx context.Context, data interface{}, schema string) *ValidationResult

	// AddCustomFormat adds a custom format validator
	AddCustomFormat(name string, validator FormatValidator)

	// RegisterFormatValidator registers a format validator by name
	RegisterFormatValidator(name string, formatFunc func(interface{}) bool)
}

// Sanitizer defines interface for data sanitization
type Sanitizer interface {
	// Sanitize cleans/normalizes the input value
	Sanitize(value interface{}) interface{}

	// Name returns the sanitizer name
	Name() string
}

// ValidationResult holds the result of a validation operation
type ValidationResult struct {
	// Valid indicates if validation passed
	Valid bool

	// Errors contains field-specific validation errors
	Errors map[string][]string

	// GlobalErrors contains validation errors not tied to specific fields
	GlobalErrors []string

	// Warnings contains non-fatal validation warnings
	Warnings map[string][]string
}

// RuleBuilder provides a fluent interface for building validation rules
type RuleBuilder interface {
	// Required marks the field as required
	Required() RuleBuilder

	// Optional marks the field as optional
	Optional() RuleBuilder

	// String adds string validation rules
	String() StringRuleBuilder

	// Number adds numeric validation rules
	Number() NumberRuleBuilder

	// DateTime adds date/time validation rules
	DateTime() DateTimeRuleBuilder

	// Custom adds a custom validation rule
	Custom(rule Rule) RuleBuilder

	// Build creates the final rule
	Build() Rule
}

// StringRuleBuilder provides string-specific validation rules
type StringRuleBuilder interface {
	// MinLength sets minimum string length
	MinLength(min int) StringRuleBuilder

	// MaxLength sets maximum string length
	MaxLength(max int) StringRuleBuilder

	// Pattern validates against a regex pattern
	Pattern(pattern string) StringRuleBuilder

	// Email validates email format
	Email() StringRuleBuilder

	// URL validates URL format
	URL() StringRuleBuilder

	// UUID validates UUID format
	UUID() StringRuleBuilder

	// Custom adds custom string validation
	Custom(validator func(string) error) StringRuleBuilder

	// Build creates the final string rule
	Build() Rule
}

// NumberRuleBuilder provides numeric validation rules
type NumberRuleBuilder interface {
	// Min sets minimum value
	Min(min float64) NumberRuleBuilder

	// Max sets maximum value
	Max(max float64) NumberRuleBuilder

	// Range sets min and max values
	Range(min, max float64) NumberRuleBuilder

	// Positive ensures the number is positive
	Positive() NumberRuleBuilder

	// NonNegative ensures the number is >= 0
	NonNegative() NumberRuleBuilder

	// Integer ensures the number is an integer
	Integer() NumberRuleBuilder

	// Decimal validates decimal with specific precision
	Decimal(precision int) NumberRuleBuilder

	// Custom adds custom numeric validation
	Custom(validator func(float64) error) NumberRuleBuilder

	// Build creates the final number rule
	Build() Rule
}

// DateTimeRuleBuilder provides date/time validation rules
type DateTimeRuleBuilder interface {
	// Format validates against specific date format
	Format(format string) DateTimeRuleBuilder

	// RFC3339 validates RFC3339 format
	RFC3339() DateTimeRuleBuilder

	// ISO8601 validates ISO8601 format
	ISO8601() DateTimeRuleBuilder

	// Before ensures date is before specified date
	Before(date string) DateTimeRuleBuilder

	// After ensures date is after specified date
	After(date string) DateTimeRuleBuilder

	// Range ensures date is within specified range
	Range(start, end string) DateTimeRuleBuilder

	// Custom adds custom date/time validation
	Custom(validator func(string) error) DateTimeRuleBuilder

	// Build creates the final date/time rule
	Build() Rule
}
