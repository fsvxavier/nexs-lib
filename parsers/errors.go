package parsers

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of parsing error
type ErrorType string

const (
	ErrorTypeInvalidFormat   ErrorType = "invalid_format"
	ErrorTypeInvalidValue    ErrorType = "invalid_value"
	ErrorTypeUnsupportedType ErrorType = "unsupported_type"
	ErrorTypeNotFound        ErrorType = "not_found"
	ErrorTypeValidation      ErrorType = "validation"
	ErrorTypeTimeout         ErrorType = "timeout"
	ErrorTypeInternal        ErrorType = "internal"
)

// ParseError represents a parsing error with additional context
type ParseError struct {
	Type        ErrorType
	Input       string
	Field       string
	Message     string
	Cause       error
	Suggestions []string
	Context     map[string]interface{}
}

func (e *ParseError) Error() string {
	var parts []string

	if e.Field != "" {
		parts = append(parts, fmt.Sprintf("field '%s'", e.Field))
	}

	if e.Input != "" {
		parts = append(parts, fmt.Sprintf("input '%s'", e.Input))
	}

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("caused by: %v", e.Cause))
	}

	result := fmt.Sprintf("parse error (%s)", e.Type)
	if len(parts) > 0 {
		result += ": " + strings.Join(parts, ", ")
	}

	if len(e.Suggestions) > 0 {
		result += fmt.Sprintf(" | suggestions: %s", strings.Join(e.Suggestions, ", "))
	}

	return result
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

func (e *ParseError) Is(target error) bool {
	t, ok := target.(*ParseError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

func (e *ParseError) WithContext(key string, value interface{}) *ParseError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func (e *ParseError) WithSuggestion(suggestion string) *ParseError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// NewParseError creates a new ParseError
func NewParseError(errorType ErrorType, input, message string) *ParseError {
	return &ParseError{
		Type:    errorType,
		Input:   input,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// NewInvalidFormatError creates an error for invalid format
func NewInvalidFormatError(input, expectedFormat string) *ParseError {
	return NewParseError(
		ErrorTypeInvalidFormat,
		input,
		fmt.Sprintf("expected format: %s", expectedFormat),
	)
}

// NewInvalidValueError creates an error for invalid value
func NewInvalidValueError(input, reason string) *ParseError {
	return NewParseError(
		ErrorTypeInvalidValue,
		input,
		reason,
	)
}

// NewNotFoundError creates an error for missing values
func NewNotFoundError(field string) *ParseError {
	return &ParseError{
		Type:    ErrorTypeNotFound,
		Field:   field,
		Message: "required field not found",
		Context: make(map[string]interface{}),
	}
}

// NewValidationError creates an error for validation failures
func NewValidationError(field, message string) *ParseError {
	return &ParseError{
		Type:    ErrorTypeValidation,
		Field:   field,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// WrapError wraps an existing error as a ParseError
func WrapError(err error, errorType ErrorType, input, message string) *ParseError {
	return &ParseError{
		Type:    errorType,
		Input:   input,
		Message: message,
		Cause:   err,
		Context: make(map[string]interface{}),
	}
}

// MultiError represents multiple parsing errors
type MultiError struct {
	Errors []error
}

func (me *MultiError) Error() string {
	if len(me.Errors) == 0 {
		return "no errors"
	}

	if len(me.Errors) == 1 {
		return me.Errors[0].Error()
	}

	var parts []string
	for i, err := range me.Errors {
		parts = append(parts, fmt.Sprintf("[%d] %v", i+1, err))
	}

	return fmt.Sprintf("multiple errors (%d): %s", len(me.Errors), strings.Join(parts, "; "))
}

func (me *MultiError) Unwrap() []error {
	return me.Errors
}

func (me *MultiError) Add(err error) {
	if err != nil {
		me.Errors = append(me.Errors, err)
	}
}

func (me *MultiError) HasErrors() bool {
	return len(me.Errors) > 0
}

// NewMultiError creates a new MultiError
func NewMultiError() *MultiError {
	return &MultiError{
		Errors: make([]error, 0),
	}
}
