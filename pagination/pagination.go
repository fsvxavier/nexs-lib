// Package pagination provides comprehensive pagination functionality.
package pagination

import (
	"encoding/json"
	"net/url"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/fsvxavier/nexs-lib/pagination/providers"
)

// PaginationService orchestrates pagination functionality using dependency injection
type PaginationService struct {
	config     *config.Config
	parser     interfaces.RequestParser
	validator  interfaces.Validator
	builder    interfaces.QueryBuilder
	calculator interfaces.PaginationCalculator
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

	return &PaginationService{
		config:     cfg,
		parser:     providers.NewStandardRequestParser(cfg),
		validator:  providers.NewStandardValidator(cfg),
		builder:    providers.NewStandardQueryBuilder(),
		calculator: providers.NewStandardPaginationCalculator(),
	}
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

	return &PaginationService{
		config:     cfg,
		parser:     parser,
		validator:  validator,
		builder:    builder,
		calculator: calculator,
	}
}

// ParseRequest parses pagination parameters from URL values
func (s *PaginationService) ParseRequest(params url.Values, sortableFields ...string) (*interfaces.PaginationParams, error) {
	// Parse parameters
	paginationParams, err := s.parser.ParsePaginationParams(params)
	if err != nil {
		return nil, err
	}

	// Validate parameters
	if err := s.validator.ValidateParams(paginationParams, sortableFields); err != nil {
		return nil, err
	}

	return paginationParams, nil
}

// BuildQuery builds a paginated SQL query
func (s *PaginationService) BuildQuery(baseQuery string, params *interfaces.PaginationParams) string {
	return s.builder.BuildQuery(baseQuery, params)
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
