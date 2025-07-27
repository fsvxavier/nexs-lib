// Package interfaces defines the contracts for pagination functionality.
package interfaces

import "net/url"

// RequestParser extracts pagination parameters from HTTP requests
type RequestParser interface {
	// ParsePaginationParams extracts pagination parameters from the request
	ParsePaginationParams(params url.Values) (*PaginationParams, error)
}

// Validator validates pagination parameters
type Validator interface {
	// ValidateParams validates pagination parameters
	ValidateParams(params *PaginationParams, sortableFields []string) error
}

// QueryBuilder builds database queries with pagination
type QueryBuilder interface {
	// BuildQuery constructs the final query with ORDER BY and LIMIT/OFFSET
	BuildQuery(baseQuery string, params *PaginationParams) string

	// BuildCountQuery constructs a count query for total records
	BuildCountQuery(baseQuery string) string
}

// PaginationCalculator calculates pagination metadata
type PaginationCalculator interface {
	// CalculateMetadata calculates navigation metadata (next, previous, total pages)
	CalculateMetadata(params *PaginationParams, totalRecords int) *PaginationMetadata
}

// PaginationParams represents parsed pagination parameters
type PaginationParams struct {
	Page             int    `json:"page"`
	Limit            int    `json:"limit"`
	SortField        string `json:"sort_field"`
	SortOrder        string `json:"sort_order"`
	MaxLimit         int    `json:"-"` // Internal limit control
	DefaultLimit     int    `json:"-"` // Default when not specified
	DefaultSortField string `json:"-"` // Default sort field
	DefaultSortOrder string `json:"-"` // Default sort order
}

// PaginationMetadata contains pagination navigation information
type PaginationMetadata struct {
	CurrentPage    int    `json:"current_page"`
	RecordsPerPage int    `json:"records_per_page"`
	TotalPages     int    `json:"total_pages"`
	TotalRecords   int    `json:"total_records"`
	Previous       *int   `json:"previous,omitempty"`
	Next           *int   `json:"next,omitempty"`
	SortField      string `json:"sort_field,omitempty"`
	SortOrder      string `json:"sort_order,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Content  interface{}         `json:"content"`
	Metadata *PaginationMetadata `json:"metadata"`
}
