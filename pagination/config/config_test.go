package config_test

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefaultConfig()
	require.NotNil(t, cfg)

	assert.Equal(t, 50, cfg.DefaultLimit)
	assert.Equal(t, 150, cfg.MaxLimit)
	assert.Equal(t, "id", cfg.DefaultSortField)
	assert.Equal(t, "asc", cfg.DefaultSortOrder)
	assert.Equal(t, []string{"asc", "desc", "ASC", "DESC"}, cfg.AllowedSortOrders)
	assert.True(t, cfg.ValidationEnabled)
	assert.False(t, cfg.StrictMode)
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         *config.Config
		expectedConfig *config.Config
	}{
		{
			name: "valid config unchanged",
			config: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          100,
				DefaultSortField:  "name",
				DefaultSortOrder:  "desc",
				AllowedSortOrders: []string{"asc", "desc"},
				ValidationEnabled: true,
				StrictMode:        true,
			},
			expectedConfig: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          100,
				DefaultSortField:  "name",
				DefaultSortOrder:  "desc",
				AllowedSortOrders: []string{"asc", "desc"},
				ValidationEnabled: true,
				StrictMode:        true,
			},
		},
		{
			name: "zero default limit gets default",
			config: &config.Config{
				DefaultLimit: 0,
				MaxLimit:     100,
			},
			expectedConfig: &config.Config{
				DefaultLimit:      50,
				MaxLimit:          100,
				DefaultSortField:  "id",
				DefaultSortOrder:  "asc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
		{
			name: "negative default limit gets default",
			config: &config.Config{
				DefaultLimit: -10,
				MaxLimit:     100,
			},
			expectedConfig: &config.Config{
				DefaultLimit:      50,
				MaxLimit:          100,
				DefaultSortField:  "id",
				DefaultSortOrder:  "asc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
		{
			name: "zero max limit gets default",
			config: &config.Config{
				DefaultLimit: 25,
				MaxLimit:     0,
			},
			expectedConfig: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          150,
				DefaultSortField:  "id",
				DefaultSortOrder:  "asc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
		{
			name: "negative max limit gets default",
			config: &config.Config{
				DefaultLimit: 25,
				MaxLimit:     -100,
			},
			expectedConfig: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          150,
				DefaultSortField:  "id",
				DefaultSortOrder:  "asc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
		{
			name: "default limit exceeds max limit",
			config: &config.Config{
				DefaultLimit: 200,
				MaxLimit:     100,
			},
			expectedConfig: &config.Config{
				DefaultLimit:      100, // Should be capped at MaxLimit
				MaxLimit:          100,
				DefaultSortField:  "id",
				DefaultSortOrder:  "asc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
		{
			name: "empty default sort field gets default",
			config: &config.Config{
				DefaultLimit:     25,
				MaxLimit:         100,
				DefaultSortField: "",
			},
			expectedConfig: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          100,
				DefaultSortField:  "id",
				DefaultSortOrder:  "asc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
		{
			name: "empty default sort order gets default",
			config: &config.Config{
				DefaultLimit:     25,
				MaxLimit:         100,
				DefaultSortField: "name",
				DefaultSortOrder: "",
			},
			expectedConfig: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          100,
				DefaultSortField:  "name",
				DefaultSortOrder:  "asc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
		{
			name: "empty allowed sort orders gets default",
			config: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          100,
				DefaultSortField:  "name",
				DefaultSortOrder:  "desc",
				AllowedSortOrders: []string{},
			},
			expectedConfig: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          100,
				DefaultSortField:  "name",
				DefaultSortOrder:  "desc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
		{
			name: "nil allowed sort orders gets default",
			config: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          100,
				DefaultSortField:  "name",
				DefaultSortOrder:  "desc",
				AllowedSortOrders: nil,
			},
			expectedConfig: &config.Config{
				DefaultLimit:      25,
				MaxLimit:          100,
				DefaultSortField:  "name",
				DefaultSortOrder:  "desc",
				AllowedSortOrders: []string{"asc", "desc", "ASC", "DESC"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedConfig.DefaultLimit, tt.config.DefaultLimit)
			assert.Equal(t, tt.expectedConfig.MaxLimit, tt.config.MaxLimit)
			assert.Equal(t, tt.expectedConfig.DefaultSortField, tt.config.DefaultSortField)
			assert.Equal(t, tt.expectedConfig.DefaultSortOrder, tt.config.DefaultSortOrder)
			assert.Equal(t, tt.expectedConfig.AllowedSortOrders, tt.config.AllowedSortOrders)
		})
	}
}

func TestConfig_ValidationScenarios(t *testing.T) {
	t.Parallel()

	t.Run("all zero values get defaults", func(t *testing.T) {
		cfg := &config.Config{}
		err := cfg.Validate()
		assert.NoError(t, err)

		assert.Equal(t, 50, cfg.DefaultLimit)
		assert.Equal(t, 150, cfg.MaxLimit)
		assert.Equal(t, "id", cfg.DefaultSortField)
		assert.Equal(t, "asc", cfg.DefaultSortOrder)
		assert.Equal(t, []string{"asc", "desc", "ASC", "DESC"}, cfg.AllowedSortOrders)
	})

	t.Run("preserve valid boolean flags", func(t *testing.T) {
		cfg := &config.Config{
			ValidationEnabled: true,
			StrictMode:        true,
		}
		err := cfg.Validate()
		assert.NoError(t, err)

		assert.True(t, cfg.ValidationEnabled)
		assert.True(t, cfg.StrictMode)
	})

	t.Run("preserve false boolean flags", func(t *testing.T) {
		cfg := &config.Config{
			ValidationEnabled: false,
			StrictMode:        false,
		}
		err := cfg.Validate()
		assert.NoError(t, err)

		assert.False(t, cfg.ValidationEnabled)
		assert.False(t, cfg.StrictMode)
	})
}

func TestConfig_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("very large values", func(t *testing.T) {
		cfg := &config.Config{
			DefaultLimit: 1000000,
			MaxLimit:     500000,
		}
		err := cfg.Validate()
		assert.NoError(t, err)

		// DefaultLimit should be capped at MaxLimit
		assert.Equal(t, 500000, cfg.DefaultLimit)
		assert.Equal(t, 500000, cfg.MaxLimit)
	})

	t.Run("custom sort orders", func(t *testing.T) {
		customOrders := []string{"ascending", "descending", "RANDOM"}
		cfg := &config.Config{
			AllowedSortOrders: customOrders,
		}
		err := cfg.Validate()
		assert.NoError(t, err)

		assert.Equal(t, customOrders, cfg.AllowedSortOrders)
	})

	t.Run("single character sort field", func(t *testing.T) {
		cfg := &config.Config{
			DefaultSortField: "a",
		}
		err := cfg.Validate()
		assert.NoError(t, err)

		assert.Equal(t, "a", cfg.DefaultSortField)
	})
}

func BenchmarkConfig_Validate(b *testing.B) {
	cfg := &config.Config{
		DefaultLimit:      25,
		MaxLimit:          100,
		DefaultSortField:  "name",
		DefaultSortOrder:  "desc",
		AllowedSortOrders: []string{"asc", "desc"},
		ValidationEnabled: true,
		StrictMode:        false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}

func BenchmarkNewDefaultConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.NewDefaultConfig()
	}
}
