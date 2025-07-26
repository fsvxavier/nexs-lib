package decimal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/decimal/config"
	"github.com/fsvxavier/nexs-lib/decimal/hooks"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

func TestNewManager(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		manager := NewManager(nil)

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.config)
		assert.NotNil(t, manager.currentProvider)
		assert.Equal(t, "cockroach", manager.currentProvider.Name())
		assert.Nil(t, manager.hookManager) // hooks disabled by default
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := config.NewConfig(
			config.WithProvider("shopspring"),
			config.WithHooksEnabled(true),
		)

		manager := NewManager(cfg)

		assert.NotNil(t, manager)
		assert.Equal(t, cfg, manager.config)
		assert.Equal(t, "shopspring", manager.currentProvider.Name())
		assert.NotNil(t, manager.hookManager)
	})

	t.Run("with invalid provider fallback", func(t *testing.T) {
		cfg := config.NewConfig(config.WithProvider("invalid"))

		manager := NewManager(cfg)

		assert.NotNil(t, manager)
		assert.Equal(t, "cockroach", manager.currentProvider.Name()) // fallback
	})
}

func TestNewManagerWithProvider(t *testing.T) {
	cfg := config.NewConfig()
	provider := &mockProvider{name: "test"}

	manager := NewManagerWithProvider(provider, cfg)

	assert.NotNil(t, manager)
	assert.Equal(t, provider, manager.currentProvider)
	assert.Equal(t, cfg, manager.config)
}

func TestManagerProviderSwitching(t *testing.T) {
	manager := NewManager(nil)

	// Test switching to shopspring
	err := manager.SwitchProvider("shopspring")
	assert.NoError(t, err)
	assert.Equal(t, "shopspring", manager.GetProvider().Name())

	// Test switching back to cockroach
	err = manager.SwitchProvider("cockroach")
	assert.NoError(t, err)
	assert.Equal(t, "cockroach", manager.GetProvider().Name())

	// Test switching to invalid provider
	err = manager.SwitchProvider("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider 'invalid' not found")
}

func TestManagerFactoryMethods(t *testing.T) {
	manager := NewManager(nil)

	t.Run("NewFromString", func(t *testing.T) {
		dec, err := manager.NewFromString("123.456")
		assert.NoError(t, err)
		assert.NotNil(t, dec)
		assert.Equal(t, "123.456", dec.String())

		// Test invalid string
		_, err = manager.NewFromString("invalid")
		assert.Error(t, err)
	})

	t.Run("NewFromFloat", func(t *testing.T) {
		dec, err := manager.NewFromFloat(123.456)
		assert.NoError(t, err)
		assert.NotNil(t, dec)

		f, err := dec.Float64()
		assert.NoError(t, err)
		assert.InDelta(t, 123.456, f, 0.000001)
	})

	t.Run("NewFromInt", func(t *testing.T) {
		dec, err := manager.NewFromInt(12345)
		assert.NoError(t, err)
		assert.NotNil(t, dec)

		i, err := dec.Int64()
		assert.NoError(t, err)
		assert.Equal(t, int64(12345), i)
	})

	t.Run("Zero", func(t *testing.T) {
		zero := manager.Zero()
		assert.NotNil(t, zero)
		assert.True(t, zero.IsZero())
	})
}

func TestManagerBatchOperations(t *testing.T) {
	manager := NewManager(nil)

	dec1, _ := manager.NewFromString("10.5")
	dec2, _ := manager.NewFromString("20.7")
	dec3, _ := manager.NewFromString("5.2")

	t.Run("Sum", func(t *testing.T) {
		// Test normal sum
		sum, err := manager.Sum(dec1, dec2, dec3)
		assert.NoError(t, err)
		assert.NotNil(t, sum)

		expected, _ := manager.NewFromString("36.4")
		assert.True(t, sum.IsEqual(expected))

		// Test empty sum
		sum, err = manager.Sum()
		assert.NoError(t, err)
		assert.True(t, sum.IsZero())
	})

	t.Run("Average", func(t *testing.T) {
		avg, err := manager.Average(dec1, dec2, dec3)
		assert.NoError(t, err)
		assert.NotNil(t, avg)

		// Test empty average
		_, err = manager.Average()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot calculate average of empty slice")
	})

	t.Run("Max", func(t *testing.T) {
		max, err := manager.Max(dec1, dec2, dec3)
		assert.NoError(t, err)
		assert.True(t, max.IsEqual(dec2)) // 20.7 is the largest

		// Test empty max
		_, err = manager.Max()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find max of empty slice")
	})

	t.Run("Min", func(t *testing.T) {
		min, err := manager.Min(dec1, dec2, dec3)
		assert.NoError(t, err)
		assert.True(t, min.IsEqual(dec3)) // 5.2 is the smallest

		// Test empty min
		_, err = manager.Min()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot find min of empty slice")
	})
}

func TestManagerParse(t *testing.T) {
	manager := NewManager(nil)

	tests := []struct {
		name     string
		input    interface{}
		expected string
		hasError bool
	}{
		{
			name:     "string",
			input:    "123.456",
			expected: "123.456",
		},
		{
			name:     "float64",
			input:    float64(123.456),
			expected: "123.456",
		},
		{
			name:     "float32",
			input:    float32(123.456),
			expected: "123.456",
		},
		{
			name:     "int",
			input:    int(123),
			expected: "123",
		},
		{
			name:     "int32",
			input:    int32(123),
			expected: "123",
		},
		{
			name:     "int64",
			input:    int64(123),
			expected: "123",
		},
		{
			name:     "unsupported type",
			input:    struct{}{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.Parse(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Contains(t, result.String(), tt.expected[:3]) // Approximate match
			}
		})
	}

	t.Run("parse existing decimal", func(t *testing.T) {
		original, _ := manager.NewFromString("123.456")
		parsed, err := manager.Parse(original)

		assert.NoError(t, err)
		assert.Equal(t, original, parsed)
	})
}

func TestManagerJSON(t *testing.T) {
	manager := NewManager(nil)

	t.Run("MarshalJSON", func(t *testing.T) {
		dec, _ := manager.NewFromString("123.456")

		data, err := manager.MarshalJSON(dec)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "123.456")
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		data := []byte(`"123.456"`)

		dec, err := manager.UnmarshalJSON(data)
		assert.NoError(t, err)
		assert.NotNil(t, dec)
		assert.Equal(t, "123.456", dec.String())

		// Test without quotes
		data = []byte(`123.456`)
		dec, err = manager.UnmarshalJSON(data)
		assert.NoError(t, err)
		assert.NotNil(t, dec)
		assert.Equal(t, "123.456", dec.String())
	})
}

func TestManagerHooks(t *testing.T) {
	cfg := config.NewConfig(config.WithHooksEnabled(true))
	manager := NewManager(cfg)

	require.NotNil(t, manager.hookManager)

	// Create a simple logging hook
	var logMessages []string
	loggingHook := hooks.NewBasicLoggingHook(func(msg string) {
		logMessages = append(logMessages, msg)
	})

	manager.hookManager.RegisterPreHook(loggingHook)

	// Perform an operation
	_, err := manager.NewFromString("123.456")
	assert.NoError(t, err)

	// Check that hook was called
	assert.Len(t, logMessages, 1)
	assert.Contains(t, logMessages[0], "NewFromString")
}

func TestManagerConfigManagement(t *testing.T) {
	manager := NewManager(nil)

	// Test getting config
	cfg := manager.GetConfig()
	assert.NotNil(t, cfg)

	// Test updating config
	newCfg := config.NewConfig(
		config.WithProvider("shopspring"),
		config.WithMaxPrecision(50),
	)

	err := manager.UpdateConfig(newCfg)
	assert.NoError(t, err)
	assert.Equal(t, "shopspring", manager.GetProvider().Name())

	// Test invalid config
	invalidCfg := config.NewConfig(config.WithMaxPrecision(0))
	err = manager.UpdateConfig(invalidCfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration")
}

func TestDefaultManager(t *testing.T) {
	// Test getting default manager
	defaultMgr := GetDefaultManager()
	assert.NotNil(t, defaultMgr)

	// Test that it's a singleton
	defaultMgr2 := GetDefaultManager()
	assert.Same(t, defaultMgr, defaultMgr2)

	// Test setting a new default manager
	customMgr := NewManager(config.NewConfig(config.WithProvider("shopspring")))
	SetDefaultManager(customMgr)

	newDefault := GetDefaultManager()
	assert.Same(t, customMgr, newDefault)

	// Reset for other tests
	SetDefaultManager(NewManager(nil))
}

func TestConvenienceFunctions(t *testing.T) {
	// Reset default manager to ensure clean state
	SetDefaultManager(NewManager(nil))

	t.Run("NewFromString", func(t *testing.T) {
		dec, err := NewFromString("123.456")
		assert.NoError(t, err)
		assert.Equal(t, "123.456", dec.String())
	})

	t.Run("NewFromFloat", func(t *testing.T) {
		dec, err := NewFromFloat(123.456)
		assert.NoError(t, err)
		assert.NotNil(t, dec)
	})

	t.Run("NewFromInt", func(t *testing.T) {
		dec, err := NewFromInt(123)
		assert.NoError(t, err)
		assert.Equal(t, "123", dec.String())
	})

	t.Run("Zero", func(t *testing.T) {
		zero := Zero()
		assert.True(t, zero.IsZero())
	})

	t.Run("batch operations", func(t *testing.T) {
		dec1, _ := NewFromString("10")
		dec2, _ := NewFromString("20")
		dec3, _ := NewFromString("30")

		sum, err := Sum(dec1, dec2, dec3)
		assert.NoError(t, err)
		expected, _ := NewFromString("60")
		assert.True(t, sum.IsEqual(expected))

		avg, err := Average(dec1, dec2, dec3)
		assert.NoError(t, err)
		assert.NotNil(t, avg)

		max, err := Max(dec1, dec2, dec3)
		assert.NoError(t, err)
		assert.True(t, max.IsEqual(dec3))

		min, err := Min(dec1, dec2, dec3)
		assert.NoError(t, err)
		assert.True(t, min.IsEqual(dec1))
	})

	t.Run("Parse", func(t *testing.T) {
		dec, err := Parse("123.456")
		assert.NoError(t, err)
		assert.Equal(t, "123.456", dec.String())

		dec, err = Parse(123)
		assert.NoError(t, err)
		assert.Equal(t, "123", dec.String())
	})
}

func BenchmarkManagerOperations(b *testing.B) {
	manager := NewManager(nil)

	b.Run("NewFromString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			manager.NewFromString("123.456")
		}
	})

	b.Run("NewFromFloat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			manager.NewFromFloat(123.456)
		}
	})

	b.Run("NewFromInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			manager.NewFromInt(123)
		}
	})
}

func BenchmarkBatchOperations(b *testing.B) {
	manager := NewManager(nil)

	decimals := make([]interfaces.Decimal, 100)
	for i := 0; i < 100; i++ {
		decimals[i], _ = manager.NewFromInt(int64(i))
	}

	b.Run("Sum", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			manager.Sum(decimals...)
		}
	})

	b.Run("Average", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			manager.Average(decimals...)
		}
	})
}

// Mock provider for testing
type mockProvider struct {
	name string
}

func (m *mockProvider) NewFromString(value string) (interfaces.Decimal, error) {
	return &mockDecimal{value: value}, nil
}

func (m *mockProvider) NewFromFloat(value float64) (interfaces.Decimal, error) {
	return &mockDecimal{value: "float"}, nil
}

func (m *mockProvider) NewFromInt(value int64) (interfaces.Decimal, error) {
	return &mockDecimal{value: "int"}, nil
}

func (m *mockProvider) Zero() interfaces.Decimal {
	return &mockDecimal{value: "0"}
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Version() string {
	return "test"
}

// Mock decimal for testing
type mockDecimal struct {
	value string
}

func (m *mockDecimal) String() string                                     { return m.value }
func (m *mockDecimal) Text(format byte) string                            { return m.value }
func (m *mockDecimal) IsEqual(other interfaces.Decimal) bool              { return false }
func (m *mockDecimal) IsGreaterThan(other interfaces.Decimal) bool        { return false }
func (m *mockDecimal) IsLessThan(other interfaces.Decimal) bool           { return false }
func (m *mockDecimal) IsGreaterThanOrEqual(other interfaces.Decimal) bool { return false }
func (m *mockDecimal) IsLessThanOrEqual(other interfaces.Decimal) bool    { return false }
func (m *mockDecimal) IsZero() bool                                       { return m.value == "0" }
func (m *mockDecimal) IsPositive() bool                                   { return false }
func (m *mockDecimal) IsNegative() bool                                   { return false }
func (m *mockDecimal) Add(other interfaces.Decimal) (interfaces.Decimal, error) {
	return m, nil
}
func (m *mockDecimal) Sub(other interfaces.Decimal) (interfaces.Decimal, error) {
	return m, nil
}
func (m *mockDecimal) Mul(other interfaces.Decimal) (interfaces.Decimal, error) {
	return m, nil
}
func (m *mockDecimal) Div(other interfaces.Decimal) (interfaces.Decimal, error) {
	return m, nil
}
func (m *mockDecimal) Mod(other interfaces.Decimal) (interfaces.Decimal, error) {
	return m, nil
}
func (m *mockDecimal) Abs() interfaces.Decimal                   { return m }
func (m *mockDecimal) Neg() interfaces.Decimal                   { return m }
func (m *mockDecimal) Truncate(uint32, int32) interfaces.Decimal { return m }
func (m *mockDecimal) TrimZerosRight() interfaces.Decimal        { return m }
func (m *mockDecimal) Round(int32) interfaces.Decimal            { return m }
func (m *mockDecimal) Float64() (float64, error)                 { return 0, nil }
func (m *mockDecimal) Int64() (int64, error)                     { return 0, nil }
func (m *mockDecimal) MarshalJSON() ([]byte, error)              { return []byte(`"` + m.value + `"`), nil }
func (m *mockDecimal) UnmarshalJSON([]byte) error                { return nil }
func (m *mockDecimal) InternalValue() interface{}                { return m.value }
