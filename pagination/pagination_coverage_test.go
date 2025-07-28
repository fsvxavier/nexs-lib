package pagination_test

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/fsvxavier/nexs-lib/pagination/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPaginationServiceWithProviders(t *testing.T) {
	cfg := config.NewDefaultConfig()
	parser := providers.NewStandardRequestParser(cfg)
	validator := providers.NewStandardValidator(cfg)
	queryBuilder := providers.NewStandardQueryBuilder()
	calculator := providers.NewStandardPaginationCalculator()

	service := pagination.NewPaginationServiceWithProviders(
		cfg,
		parser,
		validator,
		queryBuilder,
		calculator,
	)

	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.GetConfig())
}

func TestPaginationService_GetConfigMethod(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.DefaultLimit = 25
	cfg.MaxLimit = 200

	service := pagination.NewPaginationService(cfg)

	retrievedConfig := service.GetConfig()
	assert.Equal(t, cfg.DefaultLimit, retrievedConfig.DefaultLimit)
	assert.Equal(t, cfg.MaxLimit, retrievedConfig.MaxLimit)
}

func TestPaginationService_SetParserMethod(t *testing.T) {
	service := pagination.NewPaginationService(nil)

	cfg := config.NewDefaultConfig()
	customParser := providers.NewStandardRequestParser(cfg)

	service.SetParser(customParser)

	// Test that the parser was set by calling ParseRequest
	params := map[string][]string{
		"page":  {"2"},
		"limit": {"15"},
	}

	result, err := service.ParseRequest(params)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 15, result.Limit)
}

func TestPaginationService_SetValidatorMethod(t *testing.T) {
	service := pagination.NewPaginationService(nil)

	cfg := config.NewDefaultConfig()
	customValidator := providers.NewStandardValidator(cfg)

	service.SetValidator(customValidator)

	// Test validation through ValidatePageNumber with proper params
	params := &interfaces.PaginationParams{Page: 1, Limit: 10}
	err := service.ValidatePageNumber(params, 100)
	assert.NoError(t, err)

	params = &interfaces.PaginationParams{Page: 101, Limit: 10}
	err = service.ValidatePageNumber(params, 100)
	assert.Error(t, err)
}

func TestPaginationService_SetQueryBuilderMethod(t *testing.T) {
	service := pagination.NewPaginationService(nil)

	customQueryBuilder := providers.NewStandardQueryBuilder()
	service.SetQueryBuilder(customQueryBuilder)

	// Test query building
	params := &interfaces.PaginationParams{
		Page:      2,
		Limit:     10,
		SortField: "name",
		SortOrder: "asc",
	}

	query := service.BuildQuery("SELECT * FROM users", params)
	assert.Contains(t, query, "ORDER BY name ASC")
	assert.Contains(t, query, "LIMIT 10 OFFSET 10")
}

func TestPaginationService_SetCalculatorMethod(t *testing.T) {
	service := pagination.NewPaginationService(nil)

	customCalculator := providers.NewStandardPaginationCalculator()
	service.SetCalculator(customCalculator)

	// Test calculation through CreateResponse
	params := &interfaces.PaginationParams{
		Page:  1,
		Limit: 10,
	}

	content := []string{"item1", "item2"}
	response := service.CreateResponse(content, params, 25)

	assert.Equal(t, 1, response.Metadata.CurrentPage)
	assert.Equal(t, 3, response.Metadata.TotalPages)
	assert.Equal(t, 25, response.Metadata.TotalRecords)
	assert.NotNil(t, response.Metadata.Next)
	assert.Nil(t, response.Metadata.Previous)
}

func TestPaginationService_CreateResponseNilContent(t *testing.T) {
	service := pagination.NewPaginationService(nil)

	params := &interfaces.PaginationParams{
		Page:  1,
		Limit: 10,
	}

	// Test with nil content
	response := service.CreateResponse(nil, params, 0)

	assert.NotNil(t, response.Content)
	assert.Equal(t, 0, len(response.Content.([]interface{})))
	assert.Equal(t, 1, response.Metadata.CurrentPage)
	assert.Equal(t, 0, response.Metadata.TotalPages)
	assert.Equal(t, 0, response.Metadata.TotalRecords)
	assert.Nil(t, response.Metadata.Next)
	assert.Nil(t, response.Metadata.Previous)
}

func TestPaginationService_CreateResponseEmptyContent(t *testing.T) {
	service := pagination.NewPaginationService(nil)

	params := &interfaces.PaginationParams{
		Page:  1,
		Limit: 10,
	}

	// Test with empty slice
	var emptySlice []string
	response := service.CreateResponse(emptySlice, params, 0)

	assert.NotNil(t, response.Content)
	// The service converts empty slice to empty interface slice
	contentSlice, ok := response.Content.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0, len(contentSlice))
	assert.Equal(t, 1, response.Metadata.CurrentPage)
	assert.Equal(t, 0, response.Metadata.TotalPages)
	assert.Equal(t, 0, response.Metadata.TotalRecords)
}

func TestNewPaginationServiceNilConfig(t *testing.T) {
	service := pagination.NewPaginationService(nil)
	assert.NotNil(t, service)

	// Should use default config
	config := service.GetConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 50, config.DefaultLimit)
	assert.Equal(t, 150, config.MaxLimit)
}

func TestNewPaginationServiceCustomConfig(t *testing.T) {
	cfg := &config.Config{
		DefaultLimit: 25,
		MaxLimit:     200,
	}

	service := pagination.NewPaginationService(cfg)
	assert.NotNil(t, service)

	retrievedConfig := service.GetConfig()
	assert.Equal(t, 25, retrievedConfig.DefaultLimit)
	assert.Equal(t, 200, retrievedConfig.MaxLimit)
}
