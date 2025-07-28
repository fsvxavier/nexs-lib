package providers_test

import (
	"net/url"
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/fsvxavier/nexs-lib/pagination/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStandardRequestParser_ParsePaginationParams(t *testing.T) {
	t.Parallel()

	parser := providers.NewStandardRequestParser(nil)

	tests := []struct {
		name          string
		params        url.Values
		expectedPage  int
		expectedLimit int
		expectedField string
		expectedOrder string
		expectError   bool
	}{
		{
			name:          "default values",
			params:        url.Values{},
			expectedPage:  1,
			expectedLimit: 50,
			expectedField: "id",
			expectedOrder: "asc",
			expectError:   false,
		},
		{
			name: "all valid parameters",
			params: url.Values{
				"page":  []string{"3"},
				"limit": []string{"25"},
				"sort":  []string{"name"},
				"order": []string{"desc"},
			},
			expectedPage:  3,
			expectedLimit: 25,
			expectedField: "name",
			expectedOrder: "desc",
			expectError:   false,
		},
		{
			name: "invalid page",
			params: url.Values{
				"page": []string{"invalid"},
			},
			expectError: true,
		},
		{
			name: "zero page",
			params: url.Values{
				"page": []string{"0"},
			},
			expectError: true,
		},
		{
			name: "negative page",
			params: url.Values{
				"page": []string{"-1"},
			},
			expectError: true,
		},
		{
			name: "invalid limit",
			params: url.Values{
				"limit": []string{"invalid"},
			},
			expectError: true,
		},
		{
			name: "zero limit",
			params: url.Values{
				"limit": []string{"0"},
			},
			expectError: true,
		},
		{
			name: "limit exceeds maximum",
			params: url.Values{
				"limit": []string{"200"},
			},
			expectedPage:  1,
			expectedLimit: 150, // should be capped
			expectedField: "id",
			expectedOrder: "asc",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := parser.ParsePaginationParams(tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedPage, result.Page)
				assert.Equal(t, tt.expectedLimit, result.Limit)
				assert.Equal(t, tt.expectedField, result.SortField)
				assert.Equal(t, tt.expectedOrder, result.SortOrder)
			}
		})
	}
}

func TestStandardValidator_ValidateParams(t *testing.T) {
	t.Parallel()

	validator := providers.NewStandardValidator(nil)

	tests := []struct {
		name           string
		params         *interfaces.PaginationParams
		sortableFields []string
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid parameters",
			params: &interfaces.PaginationParams{
				Page:      1,
				Limit:     10,
				SortField: "name",
				SortOrder: "asc",
			},
			sortableFields: []string{"id", "name", "created_at"},
			expectError:    false,
		},
		{
			name: "invalid sort order",
			params: &interfaces.PaginationParams{
				Page:      1,
				Limit:     10,
				SortField: "name",
				SortOrder: "invalid",
			},
			sortableFields: []string{"id", "name"},
			expectError:    true,
			errorContains:  "INVALID_SORT_ORDER",
		},
		{
			name: "invalid sort field",
			params: &interfaces.PaginationParams{
				Page:      1,
				Limit:     10,
				SortField: "invalid_field",
				SortOrder: "asc",
			},
			sortableFields: []string{"id", "name"},
			expectError:    true,
			errorContains:  "INVALID_SORT_FIELD",
		},
		{
			name: "empty sortable fields allows any field",
			params: &interfaces.PaginationParams{
				Page:      1,
				Limit:     10,
				SortField: "any_field",
				SortOrder: "asc",
			},
			sortableFields: []string{},
			expectError:    false,
		},
		{
			name: "empty sort field is valid",
			params: &interfaces.PaginationParams{
				Page:      1,
				Limit:     10,
				SortField: "",
				SortOrder: "asc",
			},
			sortableFields: []string{"id", "name"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.ValidateParams(tt.params, tt.sortableFields)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStandardQueryBuilder_BuildQuery(t *testing.T) {
	t.Parallel()

	builder := providers.NewStandardQueryBuilder()

	tests := []struct {
		name        string
		baseQuery   string
		params      *interfaces.PaginationParams
		expectedSQL string
	}{
		{
			name:      "full query with all parameters",
			baseQuery: "SELECT * FROM users",
			params: &interfaces.PaginationParams{
				Page:      2,
				Limit:     10,
				SortField: "name",
				SortOrder: "desc",
			},
			expectedSQL: "SELECT * FROM users ORDER BY name DESC LIMIT 10 OFFSET 10",
		},
		{
			name:      "query without sort parameters",
			baseQuery: "SELECT id, name FROM users WHERE active = true",
			params: &interfaces.PaginationParams{
				Page:  1,
				Limit: 5,
			},
			expectedSQL: "SELECT id, name FROM users WHERE active = true LIMIT 5 OFFSET 0",
		},
		{
			name:      "query with only sort field",
			baseQuery: "SELECT * FROM products",
			params: &interfaces.PaginationParams{
				Page:      3,
				Limit:     20,
				SortField: "price",
				SortOrder: "",
			},
			expectedSQL: "SELECT * FROM products LIMIT 20 OFFSET 40",
		},
		{
			name:      "query with whitespace in base query",
			baseQuery: "  SELECT * FROM users  ",
			params: &interfaces.PaginationParams{
				Page:      1,
				Limit:     10,
				SortField: "id",
				SortOrder: "asc",
			},
			expectedSQL: "SELECT * FROM users ORDER BY id ASC LIMIT 10 OFFSET 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := builder.BuildQuery(tt.baseQuery, tt.params)
			assert.Equal(t, tt.expectedSQL, result)
		})
	}
}

func TestStandardQueryBuilder_BuildCountQuery(t *testing.T) {
	t.Parallel()

	builder := providers.NewStandardQueryBuilder()

	tests := []struct {
		name        string
		baseQuery   string
		expectedSQL string
	}{
		{
			name:        "simple select query",
			baseQuery:   "SELECT * FROM users",
			expectedSQL: "SELECT COUNT(*) FROM (SELECT * FROM users) AS count_query",
		},
		{
			name:        "complex join query",
			baseQuery:   "SELECT u.id, u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id",
			expectedSQL: "SELECT COUNT(*) FROM (SELECT u.id, u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id) AS count_query",
		},
		{
			name:        "query with whitespace",
			baseQuery:   "  SELECT * FROM products WHERE active = true  ",
			expectedSQL: "SELECT COUNT(*) FROM (SELECT * FROM products WHERE active = true) AS count_query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := builder.BuildCountQuery(tt.baseQuery)
			assert.Equal(t, tt.expectedSQL, result)
		})
	}
}

func TestStandardPaginationCalculator_CalculateMetadata(t *testing.T) {
	t.Parallel()

	calculator := providers.NewStandardPaginationCalculator()

	tests := []struct {
		name            string
		params          *interfaces.PaginationParams
		totalRecords    int
		expectedTotal   int
		expectedPrev    *int
		expectedNext    *int
		expectedCurrent int
	}{
		{
			name: "first page with more pages available",
			params: &interfaces.PaginationParams{
				Page:      1,
				Limit:     10,
				SortField: "id",
				SortOrder: "asc",
			},
			totalRecords:    25,
			expectedTotal:   3,
			expectedPrev:    nil,
			expectedNext:    intPtr(2),
			expectedCurrent: 1,
		},
		{
			name: "middle page",
			params: &interfaces.PaginationParams{
				Page:      2,
				Limit:     10,
				SortField: "name",
				SortOrder: "desc",
			},
			totalRecords:    25,
			expectedTotal:   3,
			expectedPrev:    intPtr(1),
			expectedNext:    intPtr(3),
			expectedCurrent: 2,
		},
		{
			name: "last page",
			params: &interfaces.PaginationParams{
				Page:      3,
				Limit:     10,
				SortField: "created_at",
				SortOrder: "desc",
			},
			totalRecords:    25,
			expectedTotal:   3,
			expectedPrev:    intPtr(2),
			expectedNext:    nil,
			expectedCurrent: 3,
		},
		{
			name: "single page",
			params: &interfaces.PaginationParams{
				Page:  1,
				Limit: 20,
			},
			totalRecords:    10,
			expectedTotal:   1,
			expectedPrev:    nil,
			expectedNext:    nil,
			expectedCurrent: 1,
		},
		{
			name: "no records",
			params: &interfaces.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			totalRecords:    0,
			expectedTotal:   0,
			expectedPrev:    nil,
			expectedNext:    nil,
			expectedCurrent: 1,
		},
		{
			name: "exact page boundary",
			params: &interfaces.PaginationParams{
				Page:  2,
				Limit: 10,
			},
			totalRecords:    20,
			expectedTotal:   2,
			expectedPrev:    intPtr(1),
			expectedNext:    nil,
			expectedCurrent: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := calculator.CalculateMetadata(tt.params, tt.totalRecords)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedCurrent, result.CurrentPage)
			assert.Equal(t, tt.params.Limit, result.RecordsPerPage)
			assert.Equal(t, tt.totalRecords, result.TotalRecords)
			assert.Equal(t, tt.expectedTotal, result.TotalPages)
			assert.Equal(t, tt.params.SortField, result.SortField)
			assert.Equal(t, tt.params.SortOrder, result.SortOrder)

			if tt.expectedPrev != nil {
				require.NotNil(t, result.Previous)
				assert.Equal(t, *tt.expectedPrev, *result.Previous)
			} else {
				assert.Nil(t, result.Previous)
			}

			if tt.expectedNext != nil {
				require.NotNil(t, result.Next)
				assert.Equal(t, *tt.expectedNext, *result.Next)
			} else {
				assert.Nil(t, result.Next)
			}
		})
	}
}

func TestProvidersWithCustomConfig(t *testing.T) {
	t.Parallel()

	customConfig := &config.Config{
		DefaultLimit:      25,
		MaxLimit:          100,
		DefaultSortField:  "created_at",
		DefaultSortOrder:  "desc",
		AllowedSortOrders: []string{"asc", "desc"},
		ValidationEnabled: true,
		StrictMode:        true,
	}

	parser := providers.NewStandardRequestParser(customConfig)
	validator := providers.NewStandardValidator(customConfig)

	t.Run("parser uses custom config", func(t *testing.T) {
		params, err := parser.ParsePaginationParams(url.Values{})
		require.NoError(t, err)

		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 25, params.Limit)
		assert.Equal(t, "created_at", params.SortField)
		assert.Equal(t, "desc", params.SortOrder)
		assert.Equal(t, 100, params.MaxLimit)
	})

	t.Run("validator uses custom config", func(t *testing.T) {
		params := &interfaces.PaginationParams{
			SortOrder: "DESC", // This should be invalid with custom config
		}

		err := validator.ValidateParams(params, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "INVALID_SORT_ORDER")
	})
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
