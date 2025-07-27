package pagination_test

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/providers"
	"github.com/stretchr/testify/assert"
)

func TestNewPaginationService_EdgeCases(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		cfg := config.NewDefaultConfig()
		service := pagination.NewPaginationService(cfg)
		assert.NotNil(t, service)
		assert.Equal(t, cfg, service.GetConfig())
	})
}

func TestNewPaginationServiceWithProviders_EdgeCases(t *testing.T) {
	t.Run("with all providers", func(t *testing.T) {
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
	})

	t.Run("with nil config gets default", func(t *testing.T) {
		parser := providers.NewStandardRequestParser(config.NewDefaultConfig())
		validator := providers.NewStandardValidator(config.NewDefaultConfig())
		queryBuilder := providers.NewStandardQueryBuilder()
		calculator := providers.NewStandardPaginationCalculator()

		service := pagination.NewPaginationServiceWithProviders(
			nil,
			parser,
			validator,
			queryBuilder,
			calculator,
		)

		assert.NotNil(t, service)
		retrievedConfig := service.GetConfig()
		assert.NotNil(t, retrievedConfig)
		assert.Equal(t, 50, retrievedConfig.DefaultLimit)
	})
}
