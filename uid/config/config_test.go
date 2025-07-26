package config

import (
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestDefaultProviderConfig(t *testing.T) {
	tests := []struct {
		name    string
		uidType interfaces.UIDType
	}{
		{
			name:    "ULID config",
			uidType: interfaces.UIDTypeULID,
		},
		{
			name:    "UUID v4 config",
			uidType: interfaces.UIDTypeUUIDV4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultProviderConfig(tt.uidType)

			assert.Equal(t, tt.uidType, config.Type)
			assert.Equal(t, string(tt.uidType)+"-provider", config.Name)
			assert.True(t, config.ThreadSafe)
			assert.Equal(t, 0, config.CacheSize)
			assert.True(t, config.ValidationEnabled)
			assert.False(t, config.MetricsEnabled)
		})
	}
}

func TestDefaultULIDConfig(t *testing.T) {
	config := DefaultULIDConfig()

	assert.Equal(t, interfaces.UIDTypeULID, config.Type)
	assert.Equal(t, "ulid-provider", config.Name)
	assert.True(t, config.ThreadSafe)
	assert.True(t, config.MonotonicMode)
	assert.False(t, config.CustomEntropy)
	assert.Equal(t, 10, config.EntropySize)
}

func TestDefaultUUIDV4Config(t *testing.T) {
	config := DefaultUUIDV4Config()

	assert.Equal(t, interfaces.UIDTypeUUIDV4, config.Type)
	assert.Equal(t, "uuid_v4-provider", config.Name)
	assert.Equal(t, 4, config.Version)
	assert.False(t, config.EnableSorting)
}

func TestDefaultUUIDV7Config(t *testing.T) {
	config := DefaultUUIDV7Config()

	assert.Equal(t, interfaces.UIDTypeUUIDV7, config.Type)
	assert.Equal(t, "uuid_v7-provider", config.Name)
	assert.Equal(t, 7, config.Version)
	assert.True(t, config.EnableSorting)
}

func TestDefaultFactoryConfig(t *testing.T) {
	config := DefaultFactoryConfig()

	assert.Equal(t, interfaces.UIDTypeULID, config.DefaultType)
	assert.True(t, config.EnableCaching)
	assert.Equal(t, 30*time.Minute, config.CacheTimeout)
	assert.Equal(t, 100, config.MaxCacheSize)
	assert.False(t, config.EnableMetrics)
	assert.NotNil(t, config.ProviderConfigs)
}

func TestProviderConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *ProviderConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &ProviderConfig{
				Type:              interfaces.UIDTypeULID,
				Name:              "test-provider",
				ThreadSafe:        true,
				CacheSize:         10,
				ValidationEnabled: true,
			},
			expectError: false,
		},
		{
			name: "empty type",
			config: &ProviderConfig{
				Name:              "test-provider",
				ThreadSafe:        true,
				CacheSize:         10,
				ValidationEnabled: true,
			},
			expectError: true,
		},
		{
			name: "empty name",
			config: &ProviderConfig{
				Type:              interfaces.UIDTypeULID,
				ThreadSafe:        true,
				CacheSize:         10,
				ValidationEnabled: true,
			},
			expectError: true,
		},
		{
			name: "negative cache size",
			config: &ProviderConfig{
				Type:              interfaces.UIDTypeULID,
				Name:              "test-provider",
				ThreadSafe:        true,
				CacheSize:         -1,
				ValidationEnabled: true,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestULIDConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *ULIDConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &ULIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeULID,
					Name: "ulid-provider",
				},
				MonotonicMode: true,
				EntropySize:   10,
			},
			expectError: false,
		},
		{
			name: "wrong type",
			config: &ULIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV4,
					Name: "ulid-provider",
				},
				MonotonicMode: true,
				EntropySize:   10,
			},
			expectError: true,
		},
		{
			name: "entropy size too small",
			config: &ULIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeULID,
					Name: "ulid-provider",
				},
				MonotonicMode: true,
				EntropySize:   0,
			},
			expectError: true,
		},
		{
			name: "entropy size too large",
			config: &ULIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeULID,
					Name: "ulid-provider",
				},
				MonotonicMode: true,
				EntropySize:   17,
			},
			expectError: true,
		},
		{
			name: "invalid provider config",
			config: &ULIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeULID,
					// Missing name
				},
				MonotonicMode: true,
				EntropySize:   10,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUUIDConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *UUIDConfig
		expectError bool
	}{
		{
			name: "valid config v4",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV4,
					Name: "uuid-provider",
				},
				Version: 4,
			},
			expectError: false,
		},
		{
			name: "valid config v1 with node ID",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV1,
					Name: "uuid-provider",
				},
				Version: 1,
				NodeID:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			},
			expectError: false,
		},
		{
			name: "valid config with clock sequence",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV1,
					Name: "uuid-provider",
				},
				Version:       1,
				ClockSequence: intPtr(1000),
			},
			expectError: false,
		},
		{
			name: "invalid version too low",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV4,
					Name: "uuid-provider",
				},
				Version: 0,
			},
			expectError: true,
		},
		{
			name: "invalid version too high",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV4,
					Name: "uuid-provider",
				},
				Version: 8,
			},
			expectError: true,
		},
		{
			name: "invalid node ID length",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV1,
					Name: "uuid-provider",
				},
				Version: 1,
				NodeID:  []byte{0x01, 0x02, 0x03},
			},
			expectError: true,
		},
		{
			name: "invalid clock sequence too low",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV1,
					Name: "uuid-provider",
				},
				Version:       1,
				ClockSequence: intPtr(-1),
			},
			expectError: true,
		},
		{
			name: "invalid clock sequence too high",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV1,
					Name: "uuid-provider",
				},
				Version:       1,
				ClockSequence: intPtr(16384),
			},
			expectError: true,
		},
		{
			name: "invalid provider config",
			config: &UUIDConfig{
				ProviderConfig: ProviderConfig{
					Type: interfaces.UIDTypeUUIDV4,
					// Missing name
				},
				Version: 4,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFactoryConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *FactoryConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &FactoryConfig{
				DefaultType:     interfaces.UIDTypeULID,
				EnableCaching:   true,
				CacheTimeout:    30 * time.Minute,
				MaxCacheSize:    100,
				EnableMetrics:   false,
				ProviderConfigs: make(map[interfaces.UIDType]interface{}),
			},
			expectError: false,
		},
		{
			name: "empty default type",
			config: &FactoryConfig{
				EnableCaching:   true,
				CacheTimeout:    30 * time.Minute,
				MaxCacheSize:    100,
				EnableMetrics:   false,
				ProviderConfigs: make(map[interfaces.UIDType]interface{}),
			},
			expectError: true,
		},
		{
			name: "negative cache timeout",
			config: &FactoryConfig{
				DefaultType:     interfaces.UIDTypeULID,
				EnableCaching:   true,
				CacheTimeout:    -time.Minute,
				MaxCacheSize:    100,
				EnableMetrics:   false,
				ProviderConfigs: make(map[interfaces.UIDType]interface{}),
			},
			expectError: true,
		},
		{
			name: "negative max cache size",
			config: &FactoryConfig{
				DefaultType:     interfaces.UIDTypeULID,
				EnableCaching:   true,
				CacheTimeout:    30 * time.Minute,
				MaxCacheSize:    -1,
				EnableMetrics:   false,
				ProviderConfigs: make(map[interfaces.UIDType]interface{}),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
