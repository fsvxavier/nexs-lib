package schema

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/xeipuuv/gojsonschema"
)

// jsonSchemaValidator implements SchemaValidator interface
type jsonSchemaValidator struct {
	customFormats map[string]FormatValidator
	mu            sync.RWMutex
}

// NewJSONSchemaValidator creates a new JSON schema validator
func NewJSONSchemaValidator() SchemaValidator {
	validator := &jsonSchemaValidator{
		customFormats: make(map[string]FormatValidator),
	}

	// Register default format validators
	validator.registerDefaultFormats()

	return validator
}

// ValidateSchema validates data against a JSON schema
func (jsv *jsonSchemaValidator) ValidateSchema(ctx context.Context, data interface{}, schema string) *ValidationResult {
	result := NewValidationResult()

	jsonLoader := gojsonschema.NewGoLoader(data)
	schemaLoader := gojsonschema.NewStringLoader(schema)

	gojsonResult, err := gojsonschema.Validate(schemaLoader, jsonLoader)
	if err != nil {
		result.AddGlobalError(fmt.Sprintf("schema validation error: %s", err.Error()))
		return result
	}

	if gojsonResult.Valid() {
		return result
	}

	// Convert gojsonschema errors to our format
	for _, err := range gojsonResult.Errors() {
		field := err.Field()
		if field == "(root)" {
			if property, found := err.Details()["property"]; found {
				field = property.(string)
			}
		}

		message := jsv.getErrorMessage(err.Type())
		result.AddError(field, message)
	}

	return result
}

// AddCustomFormat adds a custom format validator
func (jsv *jsonSchemaValidator) AddCustomFormat(name string, validator FormatValidator) {
	jsv.mu.Lock()
	defer jsv.mu.Unlock()

	jsv.customFormats[name] = validator

	// Register with gojsonschema
	gojsonschema.FormatCheckers.Add(name, &formatChecker{
		validator: validator,
	})
}

// RegisterFormatValidator registers a format validator by name
func (jsv *jsonSchemaValidator) RegisterFormatValidator(name string, formatFunc func(interface{}) bool) {
	validator := &funcFormatValidator{
		name:         name,
		validateFunc: formatFunc,
	}

	jsv.AddCustomFormat(name, validator)
}

// registerDefaultFormats registers the default format validators using the checks
func (jsv *jsonSchemaValidator) registerDefaultFormats() {
	// DateTime formats
	jsv.AddCustomFormat("date_time", NewDateTimeFormatValidator())
	jsv.AddCustomFormat("iso_8601_date", NewISO8601DateFormatValidator())

	// Text formats
	jsv.AddCustomFormat("text_match", NewTextMatchFormatValidator())
	jsv.AddCustomFormat("text_match_with_number", NewTextMatchWithNumberFormatValidator())
	jsv.AddCustomFormat("strong_name", NewStrongNameFormatValidator())

	// Number formats
	jsv.AddCustomFormat("json_number", NewJSONNumberFormatValidator())
	jsv.AddCustomFormat("decimal", NewDecimalFormatValidator())
	jsv.AddCustomFormat("decimal_by_factor_of_8", NewDecimalByFactor8FormatValidator())

	// String formats
	jsv.AddCustomFormat("empty_string", NewEmptyStringFormatValidator())
	jsv.AddCustomFormat("string", NewStringFormatValidator())
}

// getErrorMessage returns a user-friendly error message for schema validation errors
func (jsv *jsonSchemaValidator) getErrorMessage(errorType string) string {
	errorMap := map[string]string{
		"required":                        "REQUIRED_ATTRIBUTE_MISSING",
		"invalid_type":                    "INVALID_DATA_TYPE",
		"number_any_of":                   "INVALID_DATA_TYPE",
		"number_one_of":                   "INVALID_DATA_TYPE",
		"number_all_of":                   "INVALID_DATA_TYPE",
		"number_not":                      "INVALID_DATA_TYPE",
		"missing_dependency":              "INVALID_DATA_TYPE",
		"internal":                        "INVALID_DATA_TYPE",
		"const":                           "INVALID_DATA_TYPE",
		"enum":                            "INVALID_VALUE",
		"array_no_additional_items":       "INVALID_DATA_TYPE",
		"array_min_items":                 "INVALID_DATA_TYPE",
		"array_max_items":                 "INVALID_DATA_TYPE",
		"unique":                          "INVALID_DATA_TYPE",
		"contains":                        "INVALID_DATA_TYPE",
		"array_min_properties":            "INVALID_DATA_TYPE",
		"array_max_properties":            "INVALID_DATA_TYPE",
		"additional_property_not_allowed": "INVALID_DATA_TYPE",
		"invalid_property_pattern":        "INVALID_DATA_TYPE",
		"invalid_property_name":           "INVALID_DATA_TYPE",
		"string_gte":                      "INVALID_LENGTH",
		"string_lte":                      "INVALID_LENGTH",
		"pattern":                         "INVALID_DATA_TYPE",
		"multiple_of":                     "INVALID_DATA_TYPE",
		"number_gte":                      "INVALID_VALUE",
		"number_gt":                       "INVALID_VALUE",
		"number_lte":                      "INVALID_VALUE",
		"number_lt":                       "INVALID_VALUE",
		"condition_then":                  "INVALID_DATA_TYPE",
		"condition_else":                  "INVALID_DATA_TYPE",
		"format":                          "INVALID_FORMAT",
	}

	if message, exists := errorMap[errorType]; exists {
		return message
	}

	return "VALIDATION_ERROR"
}

// formatChecker adapts our FormatValidator to gojsonschema's FormatChecker interface
type formatChecker struct {
	validator FormatValidator
}

func (fc *formatChecker) IsFormat(input interface{}) bool {
	return fc.validator.IsFormat(input)
}

// funcFormatValidator implements FormatValidator using a function
type funcFormatValidator struct {
	name         string
	validateFunc func(interface{}) bool
}

func (ffv *funcFormatValidator) IsFormat(input interface{}) bool {
	return ffv.validateFunc(input)
}

func (ffv *funcFormatValidator) FormatName() string {
	return ffv.name
}

// SchemaValidationError converts validation results to domain errors
type SchemaValidationError struct {
	*domainerrors.InvalidSchemaError
}

// NewSchemaValidationError creates a new schema validation error from validation result
func NewSchemaValidationError(result *ValidationResult) *SchemaValidationError {
	dae := &domainerrors.InvalidSchemaError{}
	dae.Details = make(map[string][]string)

	// Convert validation result to error domain format
	for field, errors := range result.Errors {
		dae.Details[field] = errors
	}

	// Add global errors to a special field
	if len(result.GlobalErrors) > 0 {
		dae.Details["_global"] = result.GlobalErrors
	}

	return &SchemaValidationError{
		InvalidSchemaError: dae,
	}
}

// ValidateWithDomainError validates and returns domain error if validation fails
func (jsv *jsonSchemaValidator) ValidateWithDomainError(ctx context.Context, data interface{}, schema string) error {
	result := jsv.ValidateSchema(ctx, data, schema)

	if result.Valid {
		return nil
	}

	return NewSchemaValidationError(result)
}

// AddCustomFormatByRegex adds a custom format validator using regex
func AddCustomFormatByRegex(name, pattern string) {
	validator := &regexFormatValidator{
		name:    name,
		pattern: pattern,
	}

	gojsonschema.FormatCheckers.Add(name, &formatChecker{
		validator: validator,
	})
}

// regexFormatValidator implements format validation using regex
type regexFormatValidator struct {
	name    string
	pattern string
}

func (rfv *regexFormatValidator) IsFormat(input interface{}) bool {
	str, ok := input.(string)
	if !ok {
		return false
	}

	rule := NewPatternRule(rfv.pattern)
	err := rule.Validate(context.Background(), str)
	return err == nil
}

func (rfv *regexFormatValidator) FormatName() string {
	return rfv.name
}
