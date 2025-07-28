// Package pagination provides comprehensive pagination functionality.
package pagination

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/fsvxavier/nexs-lib/pagination/providers"
	"github.com/fsvxavier/nexs-lib/pagination/schema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema"
	jsonschemaconfig "github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
)

// PaginationService orchestrates pagination functionality using dependency injection
type PaginationService struct {
	config         *config.Config
	parser         interfaces.RequestParser
	validator      interfaces.Validator
	builder        interfaces.QueryBuilder
	calculator     interfaces.PaginationCalculator
	hooks          *interfaces.PaginationHooks
	builderPool    interfaces.QueryBuilderPool
	jsonValidator  *jsonschema.JSONSchemaValidator
	lazyValidators map[string]interfaces.LazyValidator
	validatorMutex sync.RWMutex
	poolEnabled    bool
}

// NewPaginationService creates a new pagination service with default providers
func NewPaginationService(cfg *config.Config) *PaginationService {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		// Use default config if validation fails
		cfg = config.NewDefaultConfig()
	}

	// Initialize JSON Schema validator with local schema
	jsonCfg := jsonschemaconfig.NewConfig()
	jsonCfg.Provider = jsonschemaconfig.GoJSONSchemaProvider
	jsonValidator, _ := jsonschema.NewValidator(jsonCfg)

	// Create query builder pool
	builderPool := NewDefaultQueryBuilderPool(10) // Start with 10 builders

	service := &PaginationService{
		config:         cfg,
		parser:         providers.NewStandardRequestParser(cfg),
		validator:      providers.NewStandardValidator(cfg),
		builder:        providers.NewStandardQueryBuilder(),
		calculator:     providers.NewStandardPaginationCalculator(),
		hooks:          &interfaces.PaginationHooks{},
		builderPool:    builderPool,
		jsonValidator:  jsonValidator,
		lazyValidators: make(map[string]interfaces.LazyValidator),
		poolEnabled:    true,
	}

	return service
}

// NewPaginationServiceWithProviders creates a service with custom providers
func NewPaginationServiceWithProviders(
	cfg *config.Config,
	parser interfaces.RequestParser,
	validator interfaces.Validator,
	builder interfaces.QueryBuilder,
	calculator interfaces.PaginationCalculator,
) *PaginationService {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Initialize JSON Schema validator with local schema
	jsonCfg := jsonschemaconfig.NewConfig()
	jsonCfg.Provider = jsonschemaconfig.GoJSONSchemaProvider
	jsonValidator, _ := jsonschema.NewValidator(jsonCfg)

	// Create query builder pool
	builderPool := NewDefaultQueryBuilderPool(10) // Start with 10 builders

	return &PaginationService{
		config:         cfg,
		parser:         parser,
		validator:      validator,
		builder:        builder,
		calculator:     calculator,
		hooks:          &interfaces.PaginationHooks{},
		builderPool:    builderPool,
		jsonValidator:  jsonValidator,
		lazyValidators: make(map[string]interfaces.LazyValidator),
		poolEnabled:    true,
	}
}

// ParseRequest parses pagination parameters from URL values with hooks support
func (s *PaginationService) ParseRequest(params url.Values, sortableFields ...string) (*interfaces.PaginationParams, error) {
	ctx := context.Background()

	// Execute pre-validation hooks
	if err := s.executeHooks(ctx, s.hooks.PreValidation, params); err != nil {
		return nil, domainerrors.NewValidationError("PRE_VALIDATION_HOOK_FAILED", "pre-validation hook failed", err)
	}

	// Parse parameters
	paginationParams, err := s.parser.ParsePaginationParams(params)
	if err != nil {
		return nil, err
	}

	// Validate parameters with standard validator first (for specific error codes)
	if err := s.validator.ValidateParams(paginationParams, sortableFields); err != nil {
		return nil, err
	}

	// Validate with JSON Schema using local schema (additional validation)
	if s.jsonValidator != nil {
		if err := s.validateWithJSONSchema(paginationParams, sortableFields); err != nil {
			return nil, err
		}
	}

	// Execute post-validation hooks
	if err := s.executeHooks(ctx, s.hooks.PostValidation, paginationParams); err != nil {
		return nil, domainerrors.NewValidationError("POST_VALIDATION_HOOK_FAILED", "post-validation hook failed", err)
	}

	return paginationParams, nil
}

// BuildQuery builds a paginated SQL query using pool if enabled
func (s *PaginationService) BuildQuery(baseQuery string, params *interfaces.PaginationParams) string {
	ctx := context.Background()

	// Execute pre-query hooks
	if err := s.executeHooks(ctx, s.hooks.PreQuery, map[string]interface{}{
		"baseQuery": baseQuery,
		"params":    params,
	}); err != nil {
		// Log error but continue with query building
	}

	var result string
	if s.poolEnabled && s.builderPool != nil {
		// Use pooled query builder
		builder := s.builderPool.Get()
		defer s.builderPool.Put(builder)
		result = builder.BuildQuery(baseQuery, params)
	} else {
		// Use standard query builder
		result = s.builder.BuildQuery(baseQuery, params)
	}

	// Execute post-query hooks
	if err := s.executeHooks(ctx, s.hooks.PostQuery, map[string]interface{}{
		"baseQuery": baseQuery,
		"params":    params,
		"result":    result,
	}); err != nil {
		// Log error but continue
	}

	return result
}

// BuildCountQuery builds a count query for total records
func (s *PaginationService) BuildCountQuery(baseQuery string) string {
	return s.builder.BuildCountQuery(baseQuery)
}

// CreateResponse creates a paginated response with metadata
func (s *PaginationService) CreateResponse(content interface{}, params *interfaces.PaginationParams, totalRecords int) *interfaces.PaginatedResponse {
	// Ensure content is not nil for JSON serialization
	if content == nil {
		content = make([]interface{}, 0)
	}

	// Handle empty or null JSON content
	if b, err := json.Marshal(content); err != nil || len(b) == 0 || string(b) == "null" {
		content = make([]interface{}, 0)
	}

	metadata := s.calculator.CalculateMetadata(params, totalRecords)

	return &interfaces.PaginatedResponse{
		Content:  content,
		Metadata: metadata,
	}
}

// ValidatePageNumber validates if the requested page is valid
func (s *PaginationService) ValidatePageNumber(params *interfaces.PaginationParams, totalRecords int) error {
	metadata := s.calculator.CalculateMetadata(params, totalRecords)

	if params.Page > metadata.TotalPages && metadata.TotalPages > 0 {
		return domainerrors.NewValidationError("INVALID_PAGE", "page number exceeds total pages", nil).
			WithField("page", "INVALID_VALUE")
	}

	return nil
}

// GetConfig returns the current configuration
func (s *PaginationService) GetConfig() *config.Config {
	return s.config
}

// SetParser sets a custom request parser
func (s *PaginationService) SetParser(parser interfaces.RequestParser) {
	s.parser = parser
}

// SetValidator sets a custom validator
func (s *PaginationService) SetValidator(validator interfaces.Validator) {
	s.validator = validator
}

// SetQueryBuilder sets a custom query builder
func (s *PaginationService) SetQueryBuilder(builder interfaces.QueryBuilder) {
	s.builder = builder
}

// SetCalculator sets a custom pagination calculator
func (s *PaginationService) SetCalculator(calculator interfaces.PaginationCalculator) {
	s.calculator = calculator
}

// AddHook adds a hook to specific pagination stage
func (s *PaginationService) AddHook(stage string, hook interfaces.Hook) {
	if s.hooks == nil {
		s.hooks = &interfaces.PaginationHooks{}
	}

	switch stage {
	case "pre-validation":
		s.hooks.PreValidation = append(s.hooks.PreValidation, hook)
	case "post-validation":
		s.hooks.PostValidation = append(s.hooks.PostValidation, hook)
	case "pre-query":
		s.hooks.PreQuery = append(s.hooks.PreQuery, hook)
	case "post-query":
		s.hooks.PostQuery = append(s.hooks.PostQuery, hook)
	case "pre-response":
		s.hooks.PreResponse = append(s.hooks.PreResponse, hook)
	case "post-response":
		s.hooks.PostResponse = append(s.hooks.PostResponse, hook)
	}
}

// SetPoolEnabled enables or disables query builder pool
func (s *PaginationService) SetPoolEnabled(enabled bool) {
	s.poolEnabled = enabled
}

// GetPoolStats returns query builder pool statistics
func (s *PaginationService) GetPoolStats() map[string]interface{} {
	if s.builderPool == nil {
		return map[string]interface{}{"enabled": false}
	}

	return map[string]interface{}{
		"enabled": s.poolEnabled,
		"size":    s.builderPool.Size(),
	}
}

// RegisterLazyValidator registers a lazy validator for specific context
func (s *PaginationService) RegisterLazyValidator(context string, validator interfaces.LazyValidator) {
	s.validatorMutex.Lock()
	defer s.validatorMutex.Unlock()

	s.lazyValidators[context] = validator
}

// GetLazyValidator retrieves a lazy validator for specific context
func (s *PaginationService) GetLazyValidator(context string) interfaces.LazyValidator {
	s.validatorMutex.RLock()
	defer s.validatorMutex.RUnlock()

	return s.lazyValidators[context]
}

// executeHooks executes a list of hooks
func (s *PaginationService) executeHooks(ctx context.Context, hooks []interfaces.Hook, data interface{}) error {
	for _, hook := range hooks {
		if err := hook.Execute(ctx, data); err != nil {
			return err
		}
	}
	return nil
}

// validateWithJSONSchema validates parameters using the local pagination schema
func (s *PaginationService) validateWithJSONSchema(params *interfaces.PaginationParams, sortableFields []string) error {
	if s.jsonValidator == nil {
		return nil
	}

	// Create validation data structure matching the schema format
	validationData := map[string]interface{}{
		"page":  params.Page,
		"limit": params.Limit,
		"sort":  params.SortField,
		"order": params.SortOrder,
	}

	// Get the local pagination schema as bytes
	paginationSchema := []byte(schema.GetPaginationSchema())

	// Validate using JSON Schema
	errors, err := s.jsonValidator.ValidateFromBytes(paginationSchema, validationData)
	if err != nil {
		return domainerrors.NewValidationError("SCHEMA_VALIDATION_FAILED", "failed to validate against pagination schema", err)
	}

	if len(errors) > 0 {
		// Convert JSON schema errors to domain errors
		validationErrors := make(map[string]string)
		for _, schemaErr := range errors {
			field := extractFieldFromPath(schemaErr.Field)
			validationErrors[field] = schemaErr.Description
		}

		// Use WithField for each error individually since WithFields might not exist
		validationError := domainerrors.NewValidationError("SCHEMA_VALIDATION_FAILED", "parameters validation failed", nil)
		for field, message := range validationErrors {
			validationError = validationError.WithField(field, message)
		}
		return validationError
	}

	return nil
}

// extractFieldFromPath extracts field name from JSON pointer path
func extractFieldFromPath(path string) string {
	// Remove leading slash and return the field name
	if strings.HasPrefix(path, "/") {
		return strings.TrimPrefix(path, "/")
	}
	return path
}
