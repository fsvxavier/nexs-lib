package schema

import (
	"context"
	"fmt"
	"strings"
)

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:        true,
		Errors:       make(map[string][]string),
		GlobalErrors: make([]string, 0),
		Warnings:     make(map[string][]string),
	}
}

// AddError adds a field-specific error
func (vr *ValidationResult) AddError(field, message string) {
	vr.Valid = false
	if vr.Errors == nil {
		vr.Errors = make(map[string][]string)
	}
	vr.Errors[field] = append(vr.Errors[field], message)
}

// AddGlobalError adds a global validation error
func (vr *ValidationResult) AddGlobalError(message string) {
	vr.Valid = false
	vr.GlobalErrors = append(vr.GlobalErrors, message)
}

// AddWarning adds a field-specific warning
func (vr *ValidationResult) AddWarning(field, message string) {
	if vr.Warnings == nil {
		vr.Warnings = make(map[string][]string)
	}
	vr.Warnings[field] = append(vr.Warnings[field], message)
}

// Merge combines multiple validation results
func (vr *ValidationResult) Merge(other *ValidationResult) {
	if !other.Valid {
		vr.Valid = false
	}

	// Merge errors
	for field, errors := range other.Errors {
		for _, err := range errors {
			vr.AddError(field, err)
		}
	}

	// Merge global errors
	for _, err := range other.GlobalErrors {
		vr.AddGlobalError(err)
	}

	// Merge warnings
	for field, warnings := range other.Warnings {
		for _, warning := range warnings {
			vr.AddWarning(field, warning)
		}
	}
}

// HasErrors returns true if there are any validation errors
func (vr *ValidationResult) HasErrors() bool {
	return !vr.Valid
}

// HasWarnings returns true if there are any warnings
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// ErrorCount returns the total number of errors
func (vr *ValidationResult) ErrorCount() int {
	count := len(vr.GlobalErrors)
	for _, errors := range vr.Errors {
		count += len(errors)
	}
	return count
}

// String returns a human-readable representation of the validation result
func (vr *ValidationResult) String() string {
	if vr.Valid {
		return "Validation passed"
	}

	var sb strings.Builder
	sb.WriteString("Validation failed:\n")

	// Global errors
	for _, err := range vr.GlobalErrors {
		sb.WriteString(fmt.Sprintf("  - %s\n", err))
	}

	// Field errors
	for field, errors := range vr.Errors {
		for _, err := range errors {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", field, err))
		}
	}

	return sb.String()
}

// FirstError returns the first error message found, or empty string if none
func (vr *ValidationResult) FirstError() string {
	if len(vr.GlobalErrors) > 0 {
		return vr.GlobalErrors[0]
	}

	for _, errors := range vr.Errors {
		if len(errors) > 0 {
			return errors[0]
		}
	}

	return ""
}

// FieldErrors returns all errors for a specific field
func (vr *ValidationResult) FieldErrors(field string) []string {
	if errors, exists := vr.Errors[field]; exists {
		return errors
	}
	return []string{}
}

// AllErrors returns all errors as a flat slice
func (vr *ValidationResult) AllErrors() []string {
	var allErrors []string

	// Add global errors
	allErrors = append(allErrors, vr.GlobalErrors...)

	// Add field errors
	for field, errors := range vr.Errors {
		for _, err := range errors {
			allErrors = append(allErrors, fmt.Sprintf("%s: %s", field, err))
		}
	}

	return allErrors
}

// ToMap converts the validation result to a map for JSON serialization
func (vr *ValidationResult) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"valid":         vr.Valid,
		"errors":        vr.Errors,
		"global_errors": vr.GlobalErrors,
		"warnings":      vr.Warnings,
		"error_count":   vr.ErrorCount(),
	}

	return result
}

// ValidationError represents a validation error that can be returned from validators
type ValidationError struct {
	Field   string
	Message string
	Code    string
	Value   interface{}
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	if ve.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s", ve.Field, ve.Message)
	}
	return fmt.Sprintf("validation error: %s", ve.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message, code string, value interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
		Value:   value,
	}
}

// ValidationContext provides context for validation operations
type ValidationContext struct {
	context.Context
	FailFast bool
	Strict   bool
	Tags     map[string]string
}

// NewValidationContext creates a new validation context
func NewValidationContext(ctx context.Context) *ValidationContext {
	if ctx == nil {
		ctx = context.Background()
	}

	return &ValidationContext{
		Context:  ctx,
		FailFast: false,
		Strict:   false,
		Tags:     make(map[string]string),
	}
}

// WithFailFast returns a new context with fail-fast enabled
func (vc *ValidationContext) WithFailFast(failFast bool) *ValidationContext {
	newCtx := *vc
	newCtx.FailFast = failFast
	return &newCtx
}

// WithStrict returns a new context with strict mode enabled
func (vc *ValidationContext) WithStrict(strict bool) *ValidationContext {
	newCtx := *vc
	newCtx.Strict = strict
	return &newCtx
}

// WithTag adds a tag to the validation context
func (vc *ValidationContext) WithTag(key, value string) *ValidationContext {
	newCtx := *vc
	newCtx.Tags = make(map[string]string)
	for k, v := range vc.Tags {
		newCtx.Tags[k] = v
	}
	newCtx.Tags[key] = value
	return &newCtx
}

// GetTag returns a tag value from the validation context
func (vc *ValidationContext) GetTag(key string) (string, bool) {
	value, exists := vc.Tags[key]
	return value, exists
}
