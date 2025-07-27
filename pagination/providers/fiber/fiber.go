// Package fiber provides Fiber-specific implementations for pagination.
package fiber

import (
	"net/url"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/fsvxavier/nexs-lib/pagination/providers"
	"github.com/gofiber/fiber/v2"
)

// FiberRequestParser implements pagination parameter parsing for Fiber contexts
type FiberRequestParser struct {
	standardParser *providers.StandardRequestParser
}

// NewFiberRequestParser creates a new Fiber request parser
func NewFiberRequestParser(cfg *config.Config) *FiberRequestParser {
	return &FiberRequestParser{
		standardParser: providers.NewStandardRequestParser(cfg),
	}
}

// ParseFromFiberContext extracts pagination parameters from Fiber context
func (p *FiberRequestParser) ParseFromFiberContext(c *fiber.Ctx) (*interfaces.PaginationParams, error) {
	// Convert Fiber query params to url.Values
	params := make(url.Values)

	// Extract query parameters
	if page := c.Query("page"); page != "" {
		params.Set("page", page)
	}
	if limit := c.Query("limit"); limit != "" {
		params.Set("limit", limit)
	}
	if sort := c.Query("sort"); sort != "" {
		params.Set("sort", sort)
	}
	if order := c.Query("order"); order != "" {
		params.Set("order", order)
	}

	return p.standardParser.ParsePaginationParams(params)
}

// ParsePaginationParams implements the RequestParser interface
func (p *FiberRequestParser) ParsePaginationParams(params url.Values) (*interfaces.PaginationParams, error) {
	return p.standardParser.ParsePaginationParams(params)
}

// FiberPaginationService provides Fiber-specific pagination functionality
type FiberPaginationService struct {
	service *pagination.PaginationService
	parser  *FiberRequestParser
}

// NewFiberPaginationService creates a new Fiber pagination service
func NewFiberPaginationService(cfg *config.Config) *FiberPaginationService {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	parser := NewFiberRequestParser(cfg)

	// Create service with Fiber parser
	service := pagination.NewPaginationServiceWithProviders(
		cfg,
		parser,
		providers.NewStandardValidator(cfg),
		providers.NewStandardQueryBuilder(),
		providers.NewStandardPaginationCalculator(),
	)

	return &FiberPaginationService{
		service: service,
		parser:  parser,
	}
}

// ParseFromFiber parses pagination parameters from Fiber context
func (fs *FiberPaginationService) ParseFromFiber(c *fiber.Ctx, sortableFields ...string) (*interfaces.PaginationParams, error) {
	// Convert Fiber query params to url.Values for standard processing
	urlParams := make(url.Values)
	if c.Query("page") != "" {
		urlParams.Set("page", c.Query("page"))
	}
	if c.Query("limit") != "" {
		urlParams.Set("limit", c.Query("limit"))
	}
	if c.Query("sort") != "" {
		urlParams.Set("sort", c.Query("sort"))
	}
	if c.Query("order") != "" {
		urlParams.Set("order", c.Query("order"))
	}

	return fs.service.ParseRequest(urlParams, sortableFields...)
}

// BuildQuery builds a paginated SQL query
func (fs *FiberPaginationService) BuildQuery(baseQuery string, params *interfaces.PaginationParams) string {
	return fs.service.BuildQuery(baseQuery, params)
}

// BuildCountQuery builds a count query for total records
func (fs *FiberPaginationService) BuildCountQuery(baseQuery string) string {
	return fs.service.BuildCountQuery(baseQuery)
}

// CreateResponse creates a paginated response with metadata
func (fs *FiberPaginationService) CreateResponse(content interface{}, params *interfaces.PaginationParams, totalRecords int) *interfaces.PaginatedResponse {
	return fs.service.CreateResponse(content, params, totalRecords)
}

// ValidatePageNumber validates if the requested page is valid
func (fs *FiberPaginationService) ValidatePageNumber(params *interfaces.PaginationParams, totalRecords int) error {
	return fs.service.ValidatePageNumber(params, totalRecords)
}

// Convenience functions for backward compatibility

// ParseMetadata parses pagination metadata from Fiber context (legacy compatibility)
func ParseMetadata(c *fiber.Ctx, sortableFields ...string) (*interfaces.PaginationParams, error) {
	service := NewFiberPaginationService(nil)
	return service.ParseFromFiber(c, sortableFields...)
}

// NewPaginatedOutput creates a paginated output (legacy compatibility)
func NewPaginatedOutput(content interface{}, params *interfaces.PaginationParams) *interfaces.PaginatedResponse {
	service := NewFiberPaginationService(nil)
	return service.CreateResponse(content, params, 0)
}

// NewPaginatedOutputWithTotal creates a paginated output with total count (legacy compatibility)
func NewPaginatedOutputWithTotal(content interface{}, totalRecords int, params *interfaces.PaginationParams) (*interfaces.PaginatedResponse, error) {
	service := NewFiberPaginationService(nil)

	// Validate page number
	if err := service.ValidatePageNumber(params, totalRecords); err != nil {
		return nil, err
	}

	return service.CreateResponse(content, params, totalRecords), nil
}
