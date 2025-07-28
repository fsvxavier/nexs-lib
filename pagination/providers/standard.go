// Package providers contains concrete implementations of pagination interfaces.
package providers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
)

// StandardRequestParser implements pagination parameter parsing
type StandardRequestParser struct {
	config *config.Config
}

// NewStandardRequestParser creates a new standard request parser
func NewStandardRequestParser(cfg *config.Config) *StandardRequestParser {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	return &StandardRequestParser{config: cfg}
}

// ParsePaginationParams extracts and parses pagination parameters from URL values
func (p *StandardRequestParser) ParsePaginationParams(params url.Values) (*interfaces.PaginationParams, error) {
	result := &interfaces.PaginationParams{
		Page:             1,
		Limit:            p.config.DefaultLimit,
		SortField:        p.config.DefaultSortField,
		SortOrder:        p.config.DefaultSortOrder,
		MaxLimit:         p.config.MaxLimit,
		DefaultLimit:     p.config.DefaultLimit,
		DefaultSortField: p.config.DefaultSortField,
		DefaultSortOrder: p.config.DefaultSortOrder,
	}

	// Parse page parameter
	if pageStr := params.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, domainerrors.NewValidationError("INVALID_PAGE", "page must be a valid integer", err).
				WithField("page", "INVALID_DATA_TYPE")
		}
		if page <= 0 {
			return nil, domainerrors.NewValidationError("INVALID_PAGE", "page must be greater than 0", nil).
				WithField("page", "INVALID_VALUE")
		}
		result.Page = page
	}

	// Parse limit parameter
	if limitStr := params.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, domainerrors.NewValidationError("INVALID_LIMIT", "limit must be a valid integer", err).
				WithField("limit", "INVALID_DATA_TYPE")
		}
		if limit <= 0 {
			return nil, domainerrors.NewValidationError("INVALID_LIMIT", "limit must be greater than 0", nil).
				WithField("limit", "INVALID_VALUE")
		}
		if limit > p.config.MaxLimit {
			limit = p.config.MaxLimit
		}
		result.Limit = limit
	}

	// Parse sort field
	if sortField := params.Get("sort"); sortField != "" {
		result.SortField = strings.TrimSpace(sortField)
	}

	// Parse sort order
	if sortOrder := params.Get("order"); sortOrder != "" {
		result.SortOrder = strings.TrimSpace(sortOrder)
	}

	return result, nil
}

// StandardValidator implements pagination parameter validation
type StandardValidator struct {
	config *config.Config
}

// NewStandardValidator creates a new standard validator
func NewStandardValidator(cfg *config.Config) *StandardValidator {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	return &StandardValidator{config: cfg}
}

// ValidateParams validates pagination parameters
func (v *StandardValidator) ValidateParams(params *interfaces.PaginationParams, sortableFields []string) error {
	if !v.config.ValidationEnabled {
		return nil
	}

	// Validate sort order
	if params.SortOrder != "" {
		validOrder := false
		for _, allowed := range v.config.AllowedSortOrders {
			if params.SortOrder == allowed {
				validOrder = true
				break
			}
		}
		if !validOrder {
			return domainerrors.NewValidationError("INVALID_SORT_ORDER",
				fmt.Sprintf("sort order must be one of: %v", v.config.AllowedSortOrders), nil).
				WithField("order", "INVALID_VALUE")
		}
	}

	// Validate sort field against allowed fields
	if params.SortField != "" && len(sortableFields) > 0 {
		validField := false
		for _, field := range sortableFields {
			if params.SortField == field {
				validField = true
				break
			}
		}
		if !validField {
			return domainerrors.NewValidationError("INVALID_SORT_FIELD",
				fmt.Sprintf("sort field must be one of: %v", sortableFields), nil).
				WithField("sort", "INVALID_VALUE")
		}
	}

	return nil
}

// StandardQueryBuilder implements SQL query building with pagination
type StandardQueryBuilder struct{}

// NewStandardQueryBuilder creates a new standard query builder
func NewStandardQueryBuilder() *StandardQueryBuilder {
	return &StandardQueryBuilder{}
}

// BuildQuery constructs the final query with ORDER BY and LIMIT/OFFSET
func (q *StandardQueryBuilder) BuildQuery(baseQuery string, params *interfaces.PaginationParams) string {
	query := strings.TrimSpace(baseQuery)

	// Add ORDER BY clause
	if params.SortField != "" && params.SortOrder != "" {
		query += fmt.Sprintf(" ORDER BY %s %s", params.SortField, strings.ToUpper(params.SortOrder))
	}

	// Add LIMIT and OFFSET
	if params.Limit > 0 && params.Page >= 1 {
		offset := (params.Page - 1) * params.Limit
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", params.Limit, offset)
	}

	return query
}

// BuildCountQuery constructs a count query for total records
func (q *StandardQueryBuilder) BuildCountQuery(baseQuery string) string {
	// Remove ORDER BY, LIMIT, OFFSET clauses and wrap in COUNT
	query := strings.TrimSpace(baseQuery)

	// Simple approach: wrap the base query
	return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_query", query)
}

// StandardPaginationCalculator implements pagination metadata calculation
type StandardPaginationCalculator struct{}

// NewStandardPaginationCalculator creates a new standard calculator
func NewStandardPaginationCalculator() *StandardPaginationCalculator {
	return &StandardPaginationCalculator{}
}

// CalculateMetadata calculates navigation metadata (next, previous, total pages)
func (c *StandardPaginationCalculator) CalculateMetadata(params *interfaces.PaginationParams, totalRecords int) *interfaces.PaginationMetadata {
	metadata := &interfaces.PaginationMetadata{
		CurrentPage:    params.Page,
		RecordsPerPage: params.Limit,
		TotalRecords:   totalRecords,
		SortField:      params.SortField,
		SortOrder:      params.SortOrder,
	}

	// Calculate total pages
	if totalRecords > 0 && params.Limit > 0 {
		metadata.TotalPages = (totalRecords + params.Limit - 1) / params.Limit
	}

	// Calculate previous page
	if params.Page > 1 {
		prev := params.Page - 1
		metadata.Previous = &prev
	}

	// Calculate next page
	if params.Page < metadata.TotalPages {
		next := params.Page + 1
		metadata.Next = &next
	}

	return metadata
}
