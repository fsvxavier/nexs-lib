package pagination

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/fsvxavier/nexs-lib/pagination/schema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema"
	jsonschemaconfig "github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
)

// LazyValidator implements lazy loading validation
type LazyValidator struct {
	config          *config.Config
	loadedFields    map[string]bool
	validationRules map[string]interface{}
	jsonValidator   *jsonschema.JSONSchemaValidator
	mu              sync.RWMutex
}

// NewLazyValidator creates a new lazy validator
func NewLazyValidator(cfg *config.Config) *LazyValidator {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Configure JSON Schema validator
	jsonCfg := jsonschemaconfig.NewConfig()
	jsonCfg.Provider = jsonschemaconfig.JSONSchemaProvider

	jsonValidator, _ := jsonschema.NewValidator(jsonCfg)

	return &LazyValidator{
		config:          cfg,
		loadedFields:    make(map[string]bool),
		validationRules: make(map[string]interface{}),
		jsonValidator:   jsonValidator,
	}
}

// ValidateParams validates pagination parameters using lazy-loaded rules
func (v *LazyValidator) ValidateParams(params *interfaces.PaginationParams, sortableFields []string) error {
	if !v.config.ValidationEnabled {
		return nil
	}

	// Load validation rules for the required fields if not already loaded
	if err := v.LoadValidator(sortableFields); err != nil {
		return fmt.Errorf("failed to load validator: %w", err)
	}

	// Perform JSON Schema validation if enabled
	if v.jsonValidator != nil {
		if err := v.validateWithJSONSchema(params, sortableFields); err != nil {
			return err
		}
	}

	// Perform standard business rules validation
	return v.validateBusinessRules(params, sortableFields)
}

// LoadValidator loads validation rules for specific fields
func (v *LazyValidator) LoadValidator(fields []string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, field := range fields {
		if !v.loadedFields[field] {
			// Simulate loading validation rules for the field
			// In a real implementation, this would load from a config file or database
			v.validationRules[field] = map[string]interface{}{
				"type":        "string",
				"maxLength":   100,
				"pattern":     "^[a-zA-Z0-9_]+$",
				"description": fmt.Sprintf("Validation rules for field %s", field),
			}
			v.loadedFields[field] = true
		}
	}

	return nil
}

// IsLoaded checks if validator is loaded for given fields
func (v *LazyValidator) IsLoaded(fields []string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	for _, field := range fields {
		if !v.loadedFields[field] {
			return false
		}
	}
	return true
}

// validateWithJSONSchema performs JSON Schema validation
func (v *LazyValidator) validateWithJSONSchema(params *interfaces.PaginationParams, sortableFields []string) error {
	// Create validation data structure
	validationData := map[string]interface{}{
		"page":  params.Page,
		"limit": params.Limit,
		"sort":  params.SortField,
		"order": params.SortOrder,
	}

	// Create a basic schema for pagination parameters using local schema
	paginationSchemaStr := schema.GetPaginationSchema()
	paginationSchema := []byte(paginationSchemaStr)

	// Validate using JSON Schema
	errors, err := v.jsonValidator.ValidateFromBytes(paginationSchema, validationData)
	if err != nil {
		return domainerrors.NewValidationError("SCHEMA_VALIDATION_FAILED", "failed to validate against schema", err)
	}

	if len(errors) > 0 {
		// Convert JSON schema errors to domain errors
		validationError := domainerrors.NewValidationError("VALIDATION_FAILED", "parameters validation failed", nil)
		for _, schemaErr := range errors {
			field := extractFieldFromPath(schemaErr.Field)
			validationError = validationError.WithField(field, schemaErr.Description)
		}
		return validationError
	}

	return nil
}

// validateBusinessRules performs standard business validation
func (v *LazyValidator) validateBusinessRules(params *interfaces.PaginationParams, sortableFields []string) error {
	// Validate page number
	if params.Page < 1 {
		return domainerrors.NewValidationError("INVALID_PAGE", "page must be greater than 0", nil).
			WithField("page", "INVALID_VALUE")
	}

	// Validate limit
	if params.Limit < 1 {
		return domainerrors.NewValidationError("INVALID_LIMIT", "limit must be greater than 0", nil).
			WithField("limit", "INVALID_VALUE")
	}

	if params.Limit > v.config.MaxLimit {
		return domainerrors.NewValidationError("LIMIT_EXCEEDED", "limit exceeds maximum allowed", nil).
			WithField("limit", "EXCEEDS_MAXIMUM").
			WithField("max_limit", strconv.Itoa(v.config.MaxLimit))
	}

	// Validate sort field
	if params.SortField != "" && len(sortableFields) > 0 {
		validField := false
		for _, field := range sortableFields {
			if field == params.SortField {
				validField = true
				break
			}
		}

		if !validField {
			return domainerrors.NewValidationError("INVALID_SORT_FIELD", "invalid sort field", nil).
				WithField("sort_field", "INVALID_VALUE").
				WithField("allowed_fields", strings.Join(sortableFields, ","))
		}
	}

	// Validate sort order
	if params.SortOrder != "" {
		upperOrder := strings.ToUpper(params.SortOrder)
		if upperOrder != "ASC" && upperOrder != "DESC" {
			return domainerrors.NewValidationError("INVALID_SORT_ORDER", "sort order must be ASC or DESC", nil).
				WithField("sort_order", "INVALID_VALUE")
		}
	}

	return nil
}

// GetLoadedFields returns the list of loaded fields
func (v *LazyValidator) GetLoadedFields() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var fields []string
	for field := range v.loadedFields {
		fields = append(fields, field)
	}
	return fields
}

// ClearCache clears the loaded validation rules
func (v *LazyValidator) ClearCache() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.loadedFields = make(map[string]bool)
	v.validationRules = make(map[string]interface{})
}

// ExecuteHook executes a hook with context
type BasicHook struct {
	name     string
	executor func(ctx context.Context, data interface{}) error
}

// NewBasicHook creates a new basic hook
func NewBasicHook(name string, executor func(ctx context.Context, data interface{}) error) *BasicHook {
	return &BasicHook{
		name:     name,
		executor: executor,
	}
}

// Execute runs the hook
func (h *BasicHook) Execute(ctx context.Context, data interface{}) error {
	if h.executor == nil {
		return nil
	}
	return h.executor(ctx, data)
}
