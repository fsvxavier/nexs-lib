package pagination_test

import (
	"testing"

	"github.com/dock-tech/isis-golang-lib/pagination"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestPagination_CalculateNextPreviousPage(t *testing.T) {
	p := &pagination.Pagination{
		CurrentPage:    1,
		RecordsPerPage: 10,
	}

	totalData := 25
	p.CalculateNextPreviousPage(&totalData)
	assert.Equal(t, 3, p.TotalPages)
	assert.Equal(t, 2, p.Next)
	assert.Equal(t, 0, p.Previous)

	p.CurrentPage = 2
	p.CalculateNextPreviousPage(&totalData)
	assert.Equal(t, 3, p.TotalPages)
	assert.Equal(t, 3, p.Next)
	assert.Equal(t, 1, p.Previous)

	p.CurrentPage = 3
	p.CalculateNextPreviousPage(&totalData)
	assert.Equal(t, 3, p.TotalPages)
	assert.Equal(t, 0, p.Next)
	assert.Equal(t, 2, p.Previous)
}

func TestParseMetadata_InvalidPage(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	ctx.Request().SetRequestURI("/?page=invalid&limit=10&sort=id&order=asc")

	metadata, err := pagination.ParseMetadata(ctx, "id", "name")
	assert.Error(t, err)
	assert.Nil(t, metadata)
}

func TestParseMetadata_InvalidLimit(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	ctx.Request().SetRequestURI("/?page=1&limit=invalid&sort=id&order=asc")

	metadata, err := pagination.ParseMetadata(ctx, "id", "name")
	assert.Error(t, err)
	assert.Nil(t, metadata)
}

func TestParseMetadata_InvalidSort(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	ctx.Request().SetRequestURI("/?page=1&limit=10&sort=invalid&order=asc")

	metadata, err := pagination.ParseMetadata(ctx, "id", "name")
	assert.Error(t, err)
	assert.Nil(t, metadata)
}

func TestParseMetadata_InvalidOrder(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	ctx.Request().SetRequestURI("/?page=1&limit=10&sort=id&order=invalid")

	metadata, err := pagination.ParseMetadata(ctx, "id", "name")
	assert.Error(t, err)
	assert.Nil(t, metadata)
}

func TestParseMetadata_ValidRequest(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	ctx.Request().SetRequestURI("/?page=1&limit=10&sort=id&order=asc")

	metadata, err := pagination.ParseMetadata(ctx, "id", "name")
	assert.NoError(t, err)
	assert.Equal(t, 1, metadata.Pagination.CurrentPage)
	assert.Equal(t, 10, metadata.Pagination.RecordsPerPage)
	assert.Equal(t, "id", metadata.Sort.Field)
	assert.Equal(t, "asc", metadata.Sort.Order)
}

func TestParseMetadata_LimitExceedsMax(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	ctx.Request().SetRequestURI("/?page=1&limit=200&sort=id&order=asc")

	metadata, err := pagination.ParseMetadata(ctx, "id", "name")
	assert.NoError(t, err)
	assert.Equal(t, 1, metadata.Pagination.CurrentPage)
	assert.Equal(t, 150, metadata.Pagination.RecordsPerPage)
	assert.Equal(t, "id", metadata.Sort.Field)
	assert.Equal(t, "asc", metadata.Sort.Order)
}

func TestParseMetadata_EmptyQueryParams(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	ctx.Request().SetRequestURI("/?page=0&limit=200&sort=id&order=asc")

	_, err := pagination.ParseMetadata(ctx, "id", "name")
	assert.Error(t, err)
}

func TestSetQuery(t *testing.T) {
	metadata := &pagination.Metadata{}
	query := "SELECT * FROM table"
	pagination.SetQuery(metadata, query)
	assert.Equal(t, query, pagination.GetQuery(metadata))
}
func TestNewMetadata(t *testing.T) {
	tests := []struct {
		name      string
		sortField string
		sortOrder string
		wantField string
		wantOrder string
		page      int
		limit     int
		wantPage  int
		wantLimit int
	}{
		{
			name:      "Valid input",
			page:      1,
			limit:     10,
			sortField: "id",
			sortOrder: "asc",
			wantPage:  1,
			wantLimit: 10,
			wantField: "id",
			wantOrder: "asc",
		},
		{
			name:      "Limit exceeds max",
			page:      1,
			limit:     200,
			sortField: "id",
			sortOrder: "asc",
			wantPage:  1,
			wantLimit: 150,
			wantField: "id",
			wantOrder: "asc",
		},
		{
			name:      "Zero limit",
			page:      1,
			limit:     0,
			sortField: "id",
			sortOrder: "asc",
			wantPage:  1,
			wantLimit: 150,
			wantField: "id",
			wantOrder: "asc",
		},
		{
			name:      "Empty sort field",
			page:      1,
			limit:     10,
			sortField: "",
			sortOrder: "asc",
			wantPage:  1,
			wantLimit: 10,
			wantField: "id",
			wantOrder: "asc",
		},
		{
			name:      "Empty sort order",
			page:      1,
			limit:     10,
			sortField: "id",
			sortOrder: "",
			wantPage:  1,
			wantLimit: 10,
			wantField: "id",
			wantOrder: "asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := pagination.NewMetadata(tt.page, tt.limit, tt.sortField, tt.sortOrder)
			assert.Equal(t, tt.wantPage, metadata.Pagination.CurrentPage)
			assert.Equal(t, tt.wantLimit, metadata.Pagination.RecordsPerPage)
			assert.Equal(t, tt.wantField, metadata.Sort.Field)
			assert.Equal(t, tt.wantOrder, metadata.Sort.Order)
		})
	}
}
func TestPreparePagination(t *testing.T) {
	tests := []struct {
		name     string
		metadata *pagination.Metadata
		want     string
	}{
		{
			name: "Valid sort and pagination",
			metadata: &pagination.Metadata{
				Sort: pagination.Sort{
					Field: "id",
					Order: "asc",
				},
				Pagination: pagination.Pagination{
					CurrentPage:    1,
					RecordsPerPage: 10,
				},
			},
			want: " ORDER BY id asc LIMIT 10 OFFSET 0",
		},
		{
			name: "No sort, valid pagination",
			metadata: &pagination.Metadata{
				Pagination: pagination.Pagination{
					CurrentPage:    2,
					RecordsPerPage: 10,
				},
			},
			want: " LIMIT 10 OFFSET 1",
		},
		{
			name: "Valid sort, no pagination",
			metadata: &pagination.Metadata{
				Sort: pagination.Sort{
					Field: "name",
					Order: "desc",
				},
			},
			want: " ORDER BY name desc",
		},
		{
			name:     "No sort, no pagination",
			metadata: &pagination.Metadata{},
			want:     "",
		},
		{
			name: "Pagination with zero records per page",
			metadata: &pagination.Metadata{
				Pagination: pagination.Pagination{
					CurrentPage:    1,
					RecordsPerPage: 0,
				},
			},
			want: "",
		},
		{
			name: "Pagination with negative current page",
			metadata: &pagination.Metadata{
				Pagination: pagination.Pagination{
					CurrentPage:    -1,
					RecordsPerPage: 10,
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pagination.PreparePagination(tt.metadata)
			assert.Equal(t, tt.want, pagination.GetQuery(got))
		})
	}
}
func TestNewPaginatedOutput(t *testing.T) {
	tests := []struct {
		body       any
		wantBody   any
		pagination *pagination.Metadata
		name       string
	}{
		{
			name:       "Valid body and pagination",
			body:       []string{"item1", "item2"},
			pagination: &pagination.Metadata{},
			wantBody:   []string{"item1", "item2"},
		},
		{
			name:       "Empty body",
			body:       "",
			pagination: &pagination.Metadata{},
			wantBody:   "",
		},
		{
			name:       "Nil body",
			body:       nil,
			pagination: &pagination.Metadata{},
			wantBody:   []any{},
		},
		{
			name:       "Invalid body",
			body:       make(chan int),
			pagination: &pagination.Metadata{},
			wantBody:   []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := pagination.NewPaginatedOutput(tt.body, tt.pagination)
			assert.Equal(t, tt.wantBody, output.Content)
			assert.Equal(t, tt.pagination, output.Metadata)
		})
	}
}
func TestNewPaginatedOutputWithTotal(t *testing.T) {
	tests := []struct {
		body       any
		wantBody   any
		total      *int
		pagination *pagination.Metadata
		name       string
		wantErr    bool
	}{
		{
			name:       "Valid body and total",
			body:       []string{"item1", "item2"},
			total:      intPtr(20),
			pagination: &pagination.Metadata{Pagination: pagination.Pagination{CurrentPage: 1, RecordsPerPage: 10}},
			wantBody:   []string{"item1", "item2"},
			wantErr:    false,
		},
		{
			name:       "Empty body",
			body:       "",
			total:      intPtr(20),
			pagination: &pagination.Metadata{Pagination: pagination.Pagination{CurrentPage: 1, RecordsPerPage: 10}},
			wantBody:   "",
			wantErr:    false,
		},
		{
			name:       "Nil body",
			body:       nil,
			total:      intPtr(20),
			pagination: &pagination.Metadata{Pagination: pagination.Pagination{CurrentPage: 1, RecordsPerPage: 10}},
			wantBody:   []any{},
			wantErr:    false,
		},
		{
			name:       "Invalid body",
			body:       make(chan int),
			total:      intPtr(20),
			pagination: &pagination.Metadata{Pagination: pagination.Pagination{CurrentPage: 1, RecordsPerPage: 10}},
			wantBody:   []any{},
			wantErr:    false,
		},
		{
			name:       "Current page exceeds total pages",
			body:       []string{"item1", "item2"},
			total:      intPtr(10),
			pagination: &pagination.Metadata{Pagination: pagination.Pagination{CurrentPage: 2, RecordsPerPage: 10}},
			wantBody:   []string{"item1", "item2"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := pagination.NewPaginatedOutputWithTotal(tt.body, tt.total, tt.pagination)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.wantBody, output.Content)
				assert.Equal(t, tt.pagination, output.Metadata)
			}
		})
	}
}

func TestPagination_CalculationTotalPage(t *testing.T) {
	tests := []struct {
		pagination *pagination.Pagination
		totalData  *int
		name       string
		want       int
	}{
		{
			name:       "Nil totalData",
			pagination: &pagination.Pagination{RecordsPerPage: 10},
			totalData:  nil,
			want:       0,
		},
		{
			name:       "Zero totalData",
			pagination: &pagination.Pagination{RecordsPerPage: 10},
			totalData:  intPtr(0),
			want:       0,
		},
		{
			name:       "Exact division",
			pagination: &pagination.Pagination{RecordsPerPage: 10},
			totalData:  intPtr(20),
			want:       2,
		},
		{
			name:       "Non-exact division",
			pagination: &pagination.Pagination{RecordsPerPage: 10},
			totalData:  intPtr(25),
			want:       3,
		},
		{
			name:       "RecordsPerPage greater than totalData",
			pagination: &pagination.Pagination{RecordsPerPage: 10},
			totalData:  intPtr(5),
			want:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pagination.CalculationTotalPage(tt.totalData)
			assert.Equal(t, tt.want, got)
		})
	}
}

func intPtr(i int) *int {
	return &i
}
