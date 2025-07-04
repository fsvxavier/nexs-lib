package schema

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// defaultValidator is the default implementation of the Validator interface
type defaultValidator struct {
	rules      []Rule
	fieldRules map[string][]Rule
	mu         sync.RWMutex
}

// NewValidator creates a new validator instance
func NewValidator() Validator {
	return &defaultValidator{
		rules:      make([]Rule, 0),
		fieldRules: make(map[string][]Rule),
	}
}

// Validate validates the given value against configured rules
func (v *defaultValidator) Validate(ctx context.Context, value interface{}) *ValidationResult {
	result := NewValidationResult()

	vCtx := NewValidationContext(ctx)

	v.mu.RLock()
	rules := make([]Rule, len(v.rules))
	copy(rules, v.rules)
	v.mu.RUnlock()

	for _, rule := range rules {
		if err := rule.Validate(vCtx, value); err != nil {
			result.AddGlobalError(err.Error())

			// Check if we should fail fast
			if vCtx.FailFast {
				break
			}
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			result.AddGlobalError("validation cancelled")
			return result
		default:
		}
	}

	return result
}

// ValidateStruct validates a struct with field-level validations
func (v *defaultValidator) ValidateStruct(ctx context.Context, s interface{}) *ValidationResult {
	result := NewValidationResult()

	if s == nil {
		result.AddGlobalError("cannot validate nil struct")
		return result
	}

	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			result.AddGlobalError("cannot validate nil struct")
			return result
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	// Must be a struct
	if val.Kind() != reflect.Struct {
		result.AddGlobalError("value must be a struct")
		return result
	}

	vCtx := NewValidationContext(ctx)

	v.mu.RLock()
	fieldRules := make(map[string][]Rule)
	for field, rules := range v.fieldRules {
		fieldRules[field] = make([]Rule, len(rules))
		copy(fieldRules[field], rules)
	}
	v.mu.RUnlock()

	// Validate each field
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name

		// Check for validation tag
		if tag := field.Tag.Get("validate"); tag != "" {
			// Parse tag and create rules
			tagRules := v.parseValidationTag(fieldName, tag, field.Type)
			fieldRules[fieldName] = append(fieldRules[fieldName], tagRules...)
		}

		// Apply field-specific rules
		if rules, exists := fieldRules[fieldName]; exists {
			for _, rule := range rules {
				if err := rule.Validate(vCtx, fieldValue.Interface()); err != nil {
					result.AddError(fieldName, err.Error())

					if vCtx.FailFast {
						return result
					}
				}
			}
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			result.AddGlobalError("validation cancelled")
			return result
		default:
		}
	}

	return result
}

// AddRule adds a validation rule to the validator
func (v *defaultValidator) AddRule(rule Rule) Validator {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.rules = append(v.rules, rule)
	return v
}

// AddFieldRule adds a field-specific validation rule for struct validation
func (v *defaultValidator) AddFieldRule(fieldName string, rule Rule) Validator {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.fieldRules[fieldName] == nil {
		v.fieldRules[fieldName] = make([]Rule, 0)
	}
	v.fieldRules[fieldName] = append(v.fieldRules[fieldName], rule)
	return v
}

// parseValidationTag parses validation tags and creates rules
func (v *defaultValidator) parseValidationTag(fieldName, tag string, fieldType reflect.Type) []Rule {
	var rules []Rule

	// Split by comma for multiple rules
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch {
		case part == "required":
			rules = append(rules, NewRequiredRule())
		case strings.HasPrefix(part, "min="):
			if value := strings.TrimPrefix(part, "min="); value != "" {
				// Check if field is numeric type
				if isNumericType(fieldType) {
					if minVal, err := strconv.ParseFloat(value, 64); err == nil {
						rules = append(rules, NewMinValueRule(minVal))
					}
				} else {
					// For strings and other types, use length validation
					rules = append(rules, NewMinLengthRule(parseInt(value)))
				}
			}
		case strings.HasPrefix(part, "max="):
			if value := strings.TrimPrefix(part, "max="); value != "" {
				// Check if field is numeric type
				if isNumericType(fieldType) {
					if maxVal, err := strconv.ParseFloat(value, 64); err == nil {
						rules = append(rules, NewMaxValueRule(maxVal))
					}
				} else {
					// For strings and other types, use length validation
					rules = append(rules, NewMaxLengthRule(parseInt(value)))
				}
			}
		case strings.HasPrefix(part, "pattern="):
			if pattern := strings.TrimPrefix(part, "pattern="); pattern != "" {
				rules = append(rules, NewPatternRule(pattern))
			}
		case part == "email":
			rules = append(rules, NewEmailRule())
		case part == "url":
			rules = append(rules, NewURLRule())
		case part == "uuid":
			rules = append(rules, NewUUIDRule())
		}
	}

	return rules
}

// parseInt safely parses an integer from string
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// isNumericType checks if a reflect.Type represents a numeric type
func isNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// ValidatorBuilder provides a fluent interface for building validators
type ValidatorBuilder struct {
	validator *defaultValidator
}

// NewValidatorBuilder creates a new validator builder
func NewValidatorBuilder() *ValidatorBuilder {
	return &ValidatorBuilder{
		validator: NewValidator().(*defaultValidator),
	}
}

// Rule adds a global rule to the validator
func (vb *ValidatorBuilder) Rule(rule Rule) *ValidatorBuilder {
	vb.validator.AddRule(rule)
	return vb
}

// Field adds field-specific rules to the validator
func (vb *ValidatorBuilder) Field(fieldName string, rules ...Rule) *ValidatorBuilder {
	for _, rule := range rules {
		vb.validator.AddFieldRule(fieldName, rule)
	}
	return vb
}

// Build returns the configured validator
func (vb *ValidatorBuilder) Build() Validator {
	return vb.validator
}
