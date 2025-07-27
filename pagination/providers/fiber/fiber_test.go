package fiber_test

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	fiberPagination "github.com/fsvxavier/nexs-lib/pagination/providers/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestFiberPaginationService_ParseFromFiber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		queryString    string
		sortableFields []string
		expectedPage   int
		expectedLimit  int
		expectedSort   string
		expectedOrder  string
		expectError    bool
		errorContains  string
	}{
		{
			name:          "default values",
			queryString:   "",
			expectedPage:  1,
			expectedLimit: 50,
			expectedSort:  "id",
			expectedOrder: "asc",
			expectError:   false,
		},
		{
			name:           "valid parameters",
			queryString:    "page=2&limit=25&sort=name&order=desc",
			sortableFields: []string{"id", "name", "created_at"},
			expectedPage:   2,
			expectedLimit:  25,
			expectedSort:   "name",
			expectedOrder:  "desc",
			expectError:    false,
		},
		{
			name:          "invalid page",
			queryString:   "page=invalid",
			expectError:   true,
			errorContains: "INVALID_PAGE",
		},
		{
			name:          "negative page",
			queryString:   "page=-1",
			expectError:   true,
			errorContains: "INVALID_PAGE",
		},
		{
			name:          "invalid limit",
			queryString:   "limit=invalid",
			expectError:   true,
			errorContains: "INVALID_LIMIT",
		},
		{
			name:          "limit exceeds maximum",
			queryString:   "limit=200",
			expectedPage:  1,
			expectedLimit: 150,
			expectedSort:  "id",
			expectedOrder: "asc",
			expectError:   false,
		},
		{
			name:           "invalid sort field",
			queryString:    "sort=invalid_field",
			sortableFields: []string{"id", "name"},
			expectError:    true,
			errorContains:  "INVALID_SORT_FIELD",
		},
		{
			name:          "invalid sort order",
			queryString:   "order=invalid",
			expectError:   true,
			errorContains: "INVALID_SORT_ORDER",
		},
		{
			name:          "case sensitive sort order",
			queryString:   "order=DESC",
			expectedPage:  1,
			expectedLimit: 50,
			expectedSort:  "id",
			expectedOrder: "DESC",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create Fiber app and context
			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(ctx)

			// Set query string
			ctx.Request().SetRequestURI("/?" + tt.queryString)

			// Create service and parse
			service := fiberPagination.NewFiberPaginationService(nil)
			params, err := service.ParseFromFiber(ctx, tt.sortableFields...)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, params)
			} else {
				require.NoError(t, err)
				require.NotNil(t, params)
				assert.Equal(t, tt.expectedPage, params.Page)
				assert.Equal(t, tt.expectedLimit, params.Limit)
				assert.Equal(t, tt.expectedSort, params.SortField)
				assert.Equal(t, tt.expectedOrder, params.SortOrder)
			}
		})
	}
}

func TestFiberPaginationService_BuildQuery(t *testing.T) {
	t.Parallel()

	service := fiberPagination.NewFiberPaginationService(nil)

	tests := []struct {
		name        string
		baseQuery   string
		page        int
		limit       int
		sortField   string
		sortOrder   string
		expectedSQL string
	}{
		{
			name:        "basic pagination query",
			baseQuery:   "SELECT * FROM users",
			page:        1,
			limit:       10,
			sortField:   "id",
			sortOrder:   "asc",
			expectedSQL: "SELECT * FROM users ORDER BY id asc LIMIT 10 OFFSET 0",
		},
		{
			name:        "second page query",
			baseQuery:   "SELECT * FROM products",
			page:        3,
			limit:       20,
			sortField:   "price",
			sortOrder:   "desc",
			expectedSQL: "SELECT * FROM products ORDER BY price desc LIMIT 20 OFFSET 40",
		},
		{
			name:        "query without sort",
			baseQuery:   "SELECT * FROM orders",
			page:        1,
			limit:       5,
			expectedSQL: "SELECT * FROM orders LIMIT 5 OFFSET 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := createPaginationParams(tt.page, tt.limit, tt.sortField, tt.sortOrder)
			sql := service.BuildQuery(tt.baseQuery, params)
			assert.Equal(t, tt.expectedSQL, sql)
		})
	}
}

func TestFiberPaginationService_CreateResponse(t *testing.T) {
	t.Parallel()

	service := fiberPagination.NewFiberPaginationService(nil)

	tests := []struct {
		name         string
		content      interface{}
		page         int
		limit        int
		totalRecords int
		expectNext   *int
		expectPrev   *int
	}{
		{
			name:         "first page with next",
			content:      []string{"item1", "item2"},
			page:         1,
			limit:        10,
			totalRecords: 25,
			expectNext:   intPtr(2),
			expectPrev:   nil,
		},
		{
			name:         "middle page",
			content:      []string{"item1", "item2"},
			page:         2,
			limit:        10,
			totalRecords: 25,
			expectNext:   intPtr(3),
			expectPrev:   intPtr(1),
		},
		{
			name:         "last page",
			content:      []string{"item1", "item2"},
			page:         3,
			limit:        10,
			totalRecords: 25,
			expectNext:   nil,
			expectPrev:   intPtr(2),
		},
		{
			name:         "empty content",
			content:      nil,
			page:         1,
			limit:        10,
			totalRecords: 0,
			expectNext:   nil,
			expectPrev:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := createPaginationParams(tt.page, tt.limit, "id", "asc")
			response := service.CreateResponse(tt.content, params, tt.totalRecords)

			require.NotNil(t, response)
			require.NotNil(t, response.Metadata)

			if tt.content == nil {
				assert.Equal(t, make([]interface{}, 0), response.Content)
			} else {
				assert.Equal(t, tt.content, response.Content)
			}

			assert.Equal(t, tt.page, response.Metadata.CurrentPage)
			assert.Equal(t, tt.limit, response.Metadata.RecordsPerPage)
			assert.Equal(t, tt.totalRecords, response.Metadata.TotalRecords)

			if tt.expectNext != nil {
				require.NotNil(t, response.Metadata.Next)
				assert.Equal(t, *tt.expectNext, *response.Metadata.Next)
			} else {
				assert.Nil(t, response.Metadata.Next)
			}

			if tt.expectPrev != nil {
				require.NotNil(t, response.Metadata.Previous)
				assert.Equal(t, *tt.expectPrev, *response.Metadata.Previous)
			} else {
				assert.Nil(t, response.Metadata.Previous)
			}
		})
	}
}

func TestFiberPaginationService_ValidatePageNumber(t *testing.T) {
	t.Parallel()

	service := fiberPagination.NewFiberPaginationService(nil)

	tests := []struct {
		name         string
		page         int
		limit        int
		totalRecords int
		expectError  bool
	}{
		{
			name:         "valid page",
			page:         2,
			limit:        10,
			totalRecords: 25,
			expectError:  false,
		},
		{
			name:         "page exceeds total",
			page:         5,
			limit:        10,
			totalRecords: 25,
			expectError:  true,
		},
		{
			name:         "last valid page",
			page:         3,
			limit:        10,
			totalRecords: 25,
			expectError:  false,
		},
		{
			name:         "no records",
			page:         1,
			limit:        10,
			totalRecords: 0,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := createPaginationParams(tt.page, tt.limit, "id", "asc")
			err := service.ValidatePageNumber(params, tt.totalRecords)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test legacy compatibility functions
func TestLegacyCompatibilityFunctions(t *testing.T) {
	t.Parallel()

	t.Run("ParseMetadata compatibility", func(t *testing.T) {
		app := fiber.New(fiber.Config{
			DisableStartupMessage: true,
		})
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(ctx)

		ctx.Request().SetRequestURI("/?page=2&limit=25&sort=name&order=desc")

		params, err := fiberPagination.ParseMetadata(ctx, "id", "name", "created_at")
		require.NoError(t, err)
		require.NotNil(t, params)

		assert.Equal(t, 2, params.Page)
		assert.Equal(t, 25, params.Limit)
		assert.Equal(t, "name", params.SortField)
		assert.Equal(t, "desc", params.SortOrder)
	})

	t.Run("NewPaginatedOutput compatibility", func(t *testing.T) {
		content := []string{"item1", "item2"}
		params := createPaginationParams(1, 10, "id", "asc")

		response := fiberPagination.NewPaginatedOutput(content, params)
		require.NotNil(t, response)
		assert.Equal(t, content, response.Content)
		assert.Equal(t, 1, response.Metadata.CurrentPage)
	})

	t.Run("NewPaginatedOutputWithTotal compatibility", func(t *testing.T) {
		content := []string{"item1", "item2"}
		params := createPaginationParams(2, 10, "id", "asc")
		totalRecords := 25

		response, err := fiberPagination.NewPaginatedOutputWithTotal(content, totalRecords, params)
		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, content, response.Content)
		assert.Equal(t, 2, response.Metadata.CurrentPage)
		assert.Equal(t, 25, response.Metadata.TotalRecords)
		assert.Equal(t, 3, response.Metadata.TotalPages)
	})

	t.Run("NewPaginatedOutputWithTotal page validation", func(t *testing.T) {
		content := []string{"item1", "item2"}
		params := createPaginationParams(5, 10, "id", "asc") // Page 5 is too high
		totalRecords := 25                                   // Only 3 pages available

		response, err := fiberPagination.NewPaginatedOutputWithTotal(content, totalRecords, params)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "INVALID_PAGE")
	})
}

// Helper functions
func createPaginationParams(page, limit int, sortField, sortOrder string) *interfaces.PaginationParams {
	return &interfaces.PaginationParams{
		Page:      page,
		Limit:     limit,
		SortField: sortField,
		SortOrder: sortOrder,
	}
}

func intPtr(i int) *int {
	return &i
}
