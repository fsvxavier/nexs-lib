package pagination_test

import (
	"net/url"
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginationService_ParseRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		queryParams    url.Values
		sortableFields []string
		expectedParams *interfaces.PaginationParams
		expectedError  string
	}{
		{
			name:        "default values when no params provided",
			queryParams: url.Values{},
			expectedParams: &interfaces.PaginationParams{
				Page:      1,
				Limit:     50,
				SortField: "id",
				SortOrder: "asc",
			},
		},
		{
			name: "valid pagination parameters",
			queryParams: url.Values{
				"page":  []string{"2"},
				"limit": []string{"25"},
				"sort":  []string{"name"},
				"order": []string{"desc"},
			},
			sortableFields: []string{"id", "name", "created_at"},
			expectedParams: &interfaces.PaginationParams{
				Page:      2,
				Limit:     25,
				SortField: "name",
				SortOrder: "desc",
			},
		},
		{
			name: "invalid page parameter",
			queryParams: url.Values{
				"page": []string{"invalid"},
			},
			expectedError: "INVALID_PAGE",
		},
		{
			name: "negative page parameter",
			queryParams: url.Values{
				"page": []string{"-1"},
			},
			expectedError: "INVALID_PAGE",
		},
		{
			name: "invalid limit parameter",
			queryParams: url.Values{
				"limit": []string{"invalid"},
			},
			expectedError: "INVALID_LIMIT",
		},
		{
			name: "limit exceeds maximum",
			queryParams: url.Values{
				"limit": []string{"200"},
			},
			expectedParams: &interfaces.PaginationParams{
				Page:      1,
				Limit:     150, // Should be capped at max limit
				SortField: "id",
				SortOrder: "asc",
			},
		},
		{
			name: "invalid sort field",
			queryParams: url.Values{
				"sort": []string{"invalid_field"},
			},
			sortableFields: []string{"id", "name"},
			expectedError:  "INVALID_SORT_FIELD",
		},
		{
			name: "invalid sort order",
			queryParams: url.Values{
				"order": []string{"invalid"},
			},
			expectedError: "INVALID_SORT_ORDER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := pagination.NewPaginationService(nil)
			params, err := service.ParseRequest(tt.queryParams, tt.sortableFields...)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, params)
			} else {
				require.NoError(t, err)
				require.NotNil(t, params)
				assert.Equal(t, tt.expectedParams.Page, params.Page)
				assert.Equal(t, tt.expectedParams.Limit, params.Limit)
				assert.Equal(t, tt.expectedParams.SortField, params.SortField)
				assert.Equal(t, tt.expectedParams.SortOrder, params.SortOrder)
			}
		})
	}
}

func TestPaginationService_BuildQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		baseQuery   string
		params      *interfaces.PaginationParams
		expectedSQL string
	}{
		{
			name:      "basic query with pagination",
			baseQuery: "SELECT * FROM users",
			params: &interfaces.PaginationParams{
				Page:      1,
				Limit:     10,
				SortField: "id",
				SortOrder: "asc",
			},
			expectedSQL: "SELECT * FROM users ORDER BY id asc LIMIT 10 OFFSET 0",
		},
		{
			name:      "query with second page",
			baseQuery: "SELECT * FROM users WHERE active = true",
			params: &interfaces.PaginationParams{
				Page:      3,
				Limit:     20,
				SortField: "created_at",
				SortOrder: "desc",
			},
			expectedSQL: "SELECT * FROM users WHERE active = true ORDER BY created_at desc LIMIT 20 OFFSET 40",
		},
		{
			name:      "query without sort parameters",
			baseQuery: "SELECT * FROM users",
			params: &interfaces.PaginationParams{
				Page:  1,
				Limit: 5,
			},
			expectedSQL: "SELECT * FROM users LIMIT 5 OFFSET 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := pagination.NewPaginationService(nil)
			sql := service.BuildQuery(tt.baseQuery, tt.params)
			assert.Equal(t, tt.expectedSQL, sql)
		})
	}
}

func TestPaginationService_BuildCountQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		baseQuery   string
		expectedSQL string
	}{
		{
			name:        "simple count query",
			baseQuery:   "SELECT * FROM users",
			expectedSQL: "SELECT COUNT(*) FROM (SELECT * FROM users) AS count_query",
		},
		{
			name:        "complex count query",
			baseQuery:   "SELECT u.*, p.name FROM users u JOIN profiles p ON u.id = p.user_id",
			expectedSQL: "SELECT COUNT(*) FROM (SELECT u.*, p.name FROM users u JOIN profiles p ON u.id = p.user_id) AS count_query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := pagination.NewPaginationService(nil)
			sql := service.BuildCountQuery(tt.baseQuery)
			assert.Equal(t, tt.expectedSQL, sql)
		})
	}
}

func TestPaginationService_CreateResponse(t *testing.T) {
	t.Parallel()

	service := pagination.NewPaginationService(nil)

	tests := []struct {
		name         string
		content      interface{}
		params       *interfaces.PaginationParams
		totalRecords int
		expectNext   *int
		expectPrev   *int
	}{
		{
			name:    "first page with next available",
			content: []string{"item1", "item2"},
			params: &interfaces.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			totalRecords: 25,
			expectNext:   intPtr(2),
			expectPrev:   nil,
		},
		{
			name:    "middle page with both next and previous",
			content: []string{"item1", "item2"},
			params: &interfaces.PaginationParams{
				Page:  2,
				Limit: 10,
			},
			totalRecords: 25,
			expectNext:   intPtr(3),
			expectPrev:   intPtr(1),
		},
		{
			name:    "last page with only previous",
			content: []string{"item1", "item2"},
			params: &interfaces.PaginationParams{
				Page:  3,
				Limit: 10,
			},
			totalRecords: 25,
			expectNext:   nil,
			expectPrev:   intPtr(2),
		},
		{
			name:    "nil content should be converted to empty array",
			content: nil,
			params: &interfaces.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			totalRecords: 0,
			expectNext:   nil,
			expectPrev:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			response := service.CreateResponse(tt.content, tt.params, tt.totalRecords)

			require.NotNil(t, response)
			require.NotNil(t, response.Metadata)

			if tt.content == nil {
				assert.Equal(t, make([]interface{}, 0), response.Content)
			} else {
				assert.Equal(t, tt.content, response.Content)
			}

			assert.Equal(t, tt.params.Page, response.Metadata.CurrentPage)
			assert.Equal(t, tt.params.Limit, response.Metadata.RecordsPerPage)
			assert.Equal(t, tt.totalRecords, response.Metadata.TotalRecords)

			if tt.expectNext != nil {
				assert.Equal(t, *tt.expectNext, *response.Metadata.Next)
			} else {
				assert.Nil(t, response.Metadata.Next)
			}

			if tt.expectPrev != nil {
				assert.Equal(t, *tt.expectPrev, *response.Metadata.Previous)
			} else {
				assert.Nil(t, response.Metadata.Previous)
			}
		})
	}
}

func TestPaginationService_ValidatePageNumber(t *testing.T) {
	t.Parallel()

	service := pagination.NewPaginationService(nil)

	tests := []struct {
		name         string
		params       *interfaces.PaginationParams
		totalRecords int
		expectError  bool
	}{
		{
			name: "valid page number",
			params: &interfaces.PaginationParams{
				Page:  2,
				Limit: 10,
			},
			totalRecords: 25,
			expectError:  false,
		},
		{
			name: "page exceeds total pages",
			params: &interfaces.PaginationParams{
				Page:  5,
				Limit: 10,
			},
			totalRecords: 25,
			expectError:  true,
		},
		{
			name: "edge case: exactly on last page",
			params: &interfaces.PaginationParams{
				Page:  3,
				Limit: 10,
			},
			totalRecords: 25,
			expectError:  false,
		},
		{
			name: "no records",
			params: &interfaces.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			totalRecords: 0,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := service.ValidatePageNumber(tt.params, tt.totalRecords)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaginationService_WithCustomConfig(t *testing.T) {
	t.Parallel()

	customConfig := &config.Config{
		DefaultLimit:      25,
		MaxLimit:          100,
		DefaultSortField:  "created_at",
		DefaultSortOrder:  "desc",
		ValidationEnabled: true,
		StrictMode:        true,
	}

	service := pagination.NewPaginationService(customConfig)

	// Test with empty parameters to verify defaults
	params, err := service.ParseRequest(url.Values{})
	require.NoError(t, err)

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 25, params.Limit)
	assert.Equal(t, "created_at", params.SortField)
	assert.Equal(t, "desc", params.SortOrder)
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
