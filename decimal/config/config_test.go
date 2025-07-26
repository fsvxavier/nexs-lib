package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	assert.Equal(t, uint32(DefaultMaxPrecision), cfg.GetMaxPrecision())
	assert.Equal(t, int32(DefaultMaxExponent), cfg.GetMaxExponent())
	assert.Equal(t, int32(DefaultMinExponent), cfg.GetMinExponent())
	assert.Equal(t, DefaultRounding, cfg.GetDefaultRounding())
	assert.Equal(t, DefaultProvider, cfg.GetProviderName())
	assert.False(t, cfg.IsHooksEnabled())
	assert.Equal(t, DefaultTimeout, cfg.GetTimeout())
	assert.NotNil(t, cfg.ProviderConfig)
}

func TestNewConfigWithOptions(t *testing.T) {
	cfg := NewConfig(
		WithMaxPrecision(50),
		WithMaxExponent(20),
		WithMinExponent(-10),
		WithRounding("RoundHalfUp"),
		WithProvider("shopspring"),
		WithHooksEnabled(true),
		WithTimeout(60),
		WithProviderConfig("key1", "value1"),
	)

	assert.Equal(t, uint32(50), cfg.GetMaxPrecision())
	assert.Equal(t, int32(20), cfg.GetMaxExponent())
	assert.Equal(t, int32(-10), cfg.GetMinExponent())
	assert.Equal(t, "RoundHalfUp", cfg.GetDefaultRounding())
	assert.Equal(t, "shopspring", cfg.GetProviderName())
	assert.True(t, cfg.IsHooksEnabled())
	assert.Equal(t, 60, cfg.GetTimeout())
	assert.Equal(t, "value1", cfg.ProviderConfig["key1"])
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		configFunc  func() *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid default config",
			configFunc: func() *Config {
				return NewDefaultConfig()
			},
			expectError: false,
		},
		{
			name: "zero max precision",
			configFunc: func() *Config {
				return NewConfig(WithMaxPrecision(0))
			},
			expectError: true,
			errorMsg:    "max_precision must be greater than 0",
		},
		{
			name: "max precision too high",
			configFunc: func() *Config {
				return NewConfig(WithMaxPrecision(1001))
			},
			expectError: true,
			errorMsg:    "max_precision cannot exceed 1000",
		},
		{
			name: "max exponent not greater than min",
			configFunc: func() *Config {
				return NewConfig(
					WithMaxExponent(5),
					WithMinExponent(10),
				)
			},
			expectError: true,
			errorMsg:    "max_exponent must be greater than min_exponent",
		},
		{
			name: "invalid rounding mode",
			configFunc: func() *Config {
				return NewConfig(WithRounding("InvalidRounding"))
			},
			expectError: true,
			errorMsg:    "invalid rounding mode: InvalidRounding",
		},
		{
			name: "invalid provider",
			configFunc: func() *Config {
				return NewConfig(WithProvider("invalid"))
			},
			expectError: true,
			errorMsg:    "invalid provider: invalid",
		},
		{
			name: "zero timeout",
			configFunc: func() *Config {
				return NewConfig(WithTimeout(0))
			},
			expectError: true,
			errorMsg:    "timeout must be greater than 0",
		},
		{
			name: "timeout too high",
			configFunc: func() *Config {
				return NewConfig(WithTimeout(301))
			},
			expectError: true,
			errorMsg:    "timeout cannot exceed 300 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.configFunc()
			err := cfg.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigClone(t *testing.T) {
	original := NewConfig(
		WithMaxPrecision(50),
		WithProvider("shopspring"),
		WithProviderConfig("key1", "value1"),
		WithProviderConfig("key2", "value2"),
	)

	clone := original.Clone()

	// Verify clone has same values
	assert.Equal(t, original.GetMaxPrecision(), clone.GetMaxPrecision())
	assert.Equal(t, original.GetProviderName(), clone.GetProviderName())
	assert.Equal(t, original.ProviderConfig["key1"], clone.ProviderConfig["key1"])
	assert.Equal(t, original.ProviderConfig["key2"], clone.ProviderConfig["key2"])

	// Verify they are different objects
	assert.NotSame(t, original, clone)
	// Verify provider configs are different instances (but don't use assert.NotSame on maps)
	originalAddr := fmt.Sprintf("%p", original.ProviderConfig)
	cloneAddr := fmt.Sprintf("%p", clone.ProviderConfig)
	assert.NotEqual(t, originalAddr, cloneAddr)

	// Modify clone and verify original is unchanged
	clone.MaxPrecision = 100
	clone.ProviderConfig["key1"] = "modified"

	assert.Equal(t, uint32(50), original.GetMaxPrecision())
	assert.Equal(t, "value1", original.ProviderConfig["key1"])
}

func TestConfigGetTimeoutDuration(t *testing.T) {
	cfg := NewConfig(WithTimeout(45))
	duration := cfg.GetTimeoutDuration()

	assert.Equal(t, 45*time.Second, duration)
}

func TestConfigString(t *testing.T) {
	cfg := NewConfig(
		WithMaxPrecision(25),
		WithMaxExponent(15),
		WithMinExponent(-5),
		WithRounding("RoundHalfUp"),
		WithProvider("shopspring"),
		WithHooksEnabled(true),
		WithTimeout(45),
	)

	str := cfg.String()
	expected := "Config{MaxPrecision: 25, MaxExponent: 15, MinExponent: -5, Rounding: RoundHalfUp, Provider: shopspring, HooksEnabled: true, Timeout: 45}"

	assert.Equal(t, expected, str)
}

func TestConfigProviderConfig(t *testing.T) {
	cfg := NewConfig()

	// Test initial state
	assert.NotNil(t, cfg.ProviderConfig)
	assert.Empty(t, cfg.ProviderConfig)

	// Test adding provider config
	cfg = NewConfig(
		WithProviderConfig("precision", 30),
		WithProviderConfig("scale", 8),
		WithProviderConfig("currency", "USD"),
	)

	assert.Equal(t, 30, cfg.ProviderConfig["precision"])
	assert.Equal(t, 8, cfg.ProviderConfig["scale"])
	assert.Equal(t, "USD", cfg.ProviderConfig["currency"])
}

func TestValidRoundingModes(t *testing.T) {
	validModes := []string{
		"RoundDown",
		"RoundUp",
		"RoundHalfUp",
		"RoundHalfDown",
		"RoundHalfEven",
		"RoundCeiling",
		"RoundFloor",
		"Round05Up",
	}

	for _, mode := range validModes {
		t.Run("rounding_mode_"+mode, func(t *testing.T) {
			cfg := NewConfig(WithRounding(mode))
			err := cfg.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestValidProviders(t *testing.T) {
	validProviders := []string{
		"cockroach",
		"shopspring",
	}

	for _, provider := range validProviders {
		t.Run("provider_"+provider, func(t *testing.T) {
			cfg := NewConfig(WithProvider(provider))
			err := cfg.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestConfigEdgeCases(t *testing.T) {
	t.Run("nil provider config", func(t *testing.T) {
		cfg := &Config{
			MaxPrecision:    DefaultMaxPrecision,
			MaxExponent:     DefaultMaxExponent,
			MinExponent:     DefaultMinExponent,
			DefaultRounding: DefaultRounding,
			ProviderName:    DefaultProvider,
			Timeout:         DefaultTimeout,
			ProviderConfig:  nil,
		}

		clone := cfg.Clone()
		if cfg.ProviderConfig == nil {
			assert.Empty(t, clone.ProviderConfig)
		} else {
			assert.NotNil(t, clone.ProviderConfig)
			assert.Empty(t, clone.ProviderConfig)
		}
	})

	t.Run("boundary precision values", func(t *testing.T) {
		// Test minimum valid precision
		cfg := NewConfig(WithMaxPrecision(1))
		err := cfg.Validate()
		assert.NoError(t, err)

		// Test maximum valid precision
		cfg = NewConfig(WithMaxPrecision(1000))
		err = cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("boundary timeout values", func(t *testing.T) {
		// Test minimum valid timeout
		cfg := NewConfig(WithTimeout(1))
		err := cfg.Validate()
		assert.NoError(t, err)

		// Test maximum valid timeout
		cfg = NewConfig(WithTimeout(300))
		err = cfg.Validate()
		assert.NoError(t, err)
	})
}

func BenchmarkNewDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewDefaultConfig()
	}
}

func BenchmarkConfigValidate(b *testing.B) {
	cfg := NewDefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cfg.Validate()
	}
}

func BenchmarkConfigClone(b *testing.B) {
	cfg := NewConfig(
		WithProviderConfig("key1", "value1"),
		WithProviderConfig("key2", "value2"),
		WithProviderConfig("key3", "value3"),
	)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cfg.Clone()
	}
}
