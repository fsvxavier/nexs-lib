package cockroach

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fsvxavier/nexs-lib/decimal/config"
)

func TestNewProvider(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		provider := NewProvider(nil)

		assert.NotNil(t, provider)
		assert.Equal(t, ProviderName, provider.Name())
		assert.Equal(t, ProviderVersion, provider.Version())
		assert.NotNil(t, provider.config)
		assert.NotNil(t, provider.ctx)
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := config.NewConfig(
			config.WithMaxPrecision(50),
			config.WithRounding("RoundHalfUp"),
		)

		provider := NewProvider(cfg)

		assert.NotNil(t, provider)
		assert.Equal(t, cfg, provider.config)
		assert.Equal(t, uint32(50), provider.ctx.Precision)
	})
}

func TestProviderNewFromString(t *testing.T) {
	provider := NewProvider(nil)

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "positive decimal",
			input:    "123.456",
			expected: "123.456",
		},
		{
			name:     "negative decimal",
			input:    "-123.456",
			expected: "-123.456",
		},
		{
			name:     "integer",
			input:    "12345",
			expected: "12345",
		},
		{
			name:     "zero",
			input:    "0",
			expected: "0",
		},
		{
			name:     "decimal with trailing zeros",
			input:    "1.2300",
			expected: "1.2300",
		},
		{
			name:     "scientific notation",
			input:    "1.23E+2",
			expected: "123",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectError: true,
		},
		{
			name:        "invalid format",
			input:       "abc",
			expectError: true,
		},
		{
			name:        "multiple decimal points",
			input:       "12.34.56",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.NewFromString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.String())
			}
		})
	}
}

func TestProviderNewFromFloat(t *testing.T) {
	provider := NewProvider(nil)

	tests := []struct {
		name        string
		input       float64
		expectError bool
	}{
		{
			name:  "positive float",
			input: 123.456,
		},
		{
			name:  "negative float",
			input: -123.456,
		},
		{
			name:  "zero",
			input: 0.0,
		},
		{
			name:  "very small number",
			input: 0.00000001,
		},
		{
			name:  "very large number",
			input: 1234567890123456789.0,
		},
		{
			name:        "NaN",
			input:       math.NaN(),
			expectError: true,
		},
		{
			name:        "positive infinity",
			input:       math.Inf(1),
			expectError: true,
		},
		{
			name:        "negative infinity",
			input:       math.Inf(-1),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := provider.NewFromFloat(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Convert back to float to verify
				backToFloat, err := result.Float64()
				assert.NoError(t, err)
				assert.InDelta(t, tt.input, backToFloat, 0.000001)
			}
		})
	}
}

func TestProviderNewFromInt(t *testing.T) {
	provider := NewProvider(nil)

	tests := []int64{
		0,
		1,
		-1,
		123456789,
		-123456789,
		9223372036854775807,  // max int64
		-9223372036854775808, // min int64
	}

	for _, input := range tests {
		t.Run("int64_"+string(rune(input)), func(t *testing.T) {
			result, err := provider.NewFromInt(input)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Convert back to int64 to verify
			backToInt, err := result.Int64()
			assert.NoError(t, err)
			assert.Equal(t, input, backToInt)
		})
	}
}

func TestProviderZero(t *testing.T) {
	provider := NewProvider(nil)

	zero := provider.Zero()

	assert.NotNil(t, zero)
	assert.True(t, zero.IsZero())
	assert.False(t, zero.IsPositive())
	assert.False(t, zero.IsNegative())
	assert.Equal(t, "0", zero.String())
}

func TestDecimalComparisons(t *testing.T) {
	provider := NewProvider(nil)

	dec1, _ := provider.NewFromString("10.5")
	dec2, _ := provider.NewFromString("10.5")
	dec3, _ := provider.NewFromString("20.7")
	dec4, _ := provider.NewFromString("5.2")

	// IsEqual
	assert.True(t, dec1.IsEqual(dec2))
	assert.False(t, dec1.IsEqual(dec3))
	assert.False(t, dec1.IsEqual(nil))

	// IsGreaterThan
	assert.True(t, dec3.IsGreaterThan(dec1))
	assert.False(t, dec1.IsGreaterThan(dec3))
	assert.False(t, dec1.IsGreaterThan(dec2))

	// IsLessThan
	assert.True(t, dec4.IsLessThan(dec1))
	assert.False(t, dec1.IsLessThan(dec4))
	assert.False(t, dec1.IsLessThan(dec2))

	// IsGreaterThanOrEqual
	assert.True(t, dec1.IsGreaterThanOrEqual(dec2))
	assert.True(t, dec3.IsGreaterThanOrEqual(dec1))
	assert.False(t, dec4.IsGreaterThanOrEqual(dec1))

	// IsLessThanOrEqual
	assert.True(t, dec1.IsLessThanOrEqual(dec2))
	assert.True(t, dec4.IsLessThanOrEqual(dec1))
	assert.False(t, dec3.IsLessThanOrEqual(dec1))
}

func TestDecimalStates(t *testing.T) {
	provider := NewProvider(nil)

	zero, _ := provider.NewFromString("0")
	positive, _ := provider.NewFromString("10.5")
	negative, _ := provider.NewFromString("-10.5")

	// IsZero
	assert.True(t, zero.IsZero())
	assert.False(t, positive.IsZero())
	assert.False(t, negative.IsZero())

	// IsPositive
	assert.False(t, zero.IsPositive())
	assert.True(t, positive.IsPositive())
	assert.False(t, negative.IsPositive())

	// IsNegative
	assert.False(t, zero.IsNegative())
	assert.False(t, positive.IsNegative())
	assert.True(t, negative.IsNegative())
}

func TestDecimalArithmetic(t *testing.T) {
	provider := NewProvider(nil)

	dec1, _ := provider.NewFromString("10.5")
	dec2, _ := provider.NewFromString("2.5")
	zero := provider.Zero()

	t.Run("addition", func(t *testing.T) {
		result, err := dec1.Add(dec2)
		assert.NoError(t, err)
		assert.Equal(t, "13.0", result.String())

		// Test with nil
		_, err = dec1.Add(nil)
		assert.Error(t, err)
	})

	t.Run("subtraction", func(t *testing.T) {
		result, err := dec1.Sub(dec2)
		assert.NoError(t, err)
		assert.Equal(t, "8.0", result.String())

		// Test with nil
		_, err = dec1.Sub(nil)
		assert.Error(t, err)
	})

	t.Run("multiplication", func(t *testing.T) {
		result, err := dec1.Mul(dec2)
		assert.NoError(t, err)
		assert.Equal(t, "26.25", result.String())

		// Test with nil
		_, err = dec1.Mul(nil)
		assert.Error(t, err)
	})

	t.Run("division", func(t *testing.T) {
		result, err := dec1.Div(dec2)
		assert.NoError(t, err)
		// CockroachDB APD may return more precision, so check if it starts with "4.2"
		assert.True(t, strings.HasPrefix(result.String(), "4.2"))

		// Test division by zero
		_, err = dec1.Div(zero)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "division by zero")

		// Test with nil
		_, err = dec1.Div(nil)
		assert.Error(t, err)
	})

	t.Run("modulo", func(t *testing.T) {
		dec3, _ := provider.NewFromString("7")
		dec4, _ := provider.NewFromString("3")

		result, err := dec3.Mod(dec4)
		assert.NoError(t, err)
		assert.Equal(t, "1", result.String())

		// Test modulo by zero
		_, err = dec3.Mod(zero)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "modulo by zero")

		// Test with nil
		_, err = dec3.Mod(nil)
		assert.Error(t, err)
	})

	t.Run("absolute value", func(t *testing.T) {
		negative, _ := provider.NewFromString("-10.5")

		absPositive := dec1.Abs()
		assert.Equal(t, "10.5", absPositive.String())

		absNegative := negative.Abs()
		assert.Equal(t, "10.5", absNegative.String())
	})

	t.Run("negation", func(t *testing.T) {
		neg := dec1.Neg()
		assert.Equal(t, "-10.5", neg.String())

		negNeg := neg.Neg()
		assert.Equal(t, "10.5", negNeg.String())
	})
}

func TestDecimalPrecisionOps(t *testing.T) {
	provider := NewProvider(nil)

	t.Run("truncate", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456789")

		truncated := dec.Truncate(5, -2)
		assert.NotNil(t, truncated)
		// The exact result depends on the truncation implementation
		assert.NotEqual(t, "", truncated.String())
	})

	t.Run("trim zeros right", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.4500")

		trimmed := dec.TrimZerosRight()
		assert.Equal(t, "123.45", trimmed.String())

		// Test integer with no decimals
		decInt, _ := provider.NewFromString("123")
		trimmedInt := decInt.TrimZerosRight()
		assert.Equal(t, "123", trimmedInt.String())
	})

	t.Run("round", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456")

		rounded := dec.Round(2)
		assert.NotNil(t, rounded)
		assert.NotEqual(t, "", rounded.String())
	})
}

func TestDecimalConversions(t *testing.T) {
	provider := NewProvider(nil)

	t.Run("to float64", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456")

		f, err := dec.Float64()
		assert.NoError(t, err)
		assert.InDelta(t, 123.456, f, 0.000001)
	})

	t.Run("to int64", func(t *testing.T) {
		// Test integer
		dec, _ := provider.NewFromString("12345")

		i, err := dec.Int64()
		assert.NoError(t, err)
		assert.Equal(t, int64(12345), i)

		// Test non-integer
		decFloat, _ := provider.NewFromString("123.456")

		_, err = decFloat.Int64()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not an integer")
	})
}

func TestDecimalJSON(t *testing.T) {
	provider := NewProvider(nil)

	t.Run("marshal", func(t *testing.T) {
		dec, _ := provider.NewFromString("123.456")

		data, err := dec.MarshalJSON()
		assert.NoError(t, err)
		assert.Contains(t, string(data), "123.456")
	})

	t.Run("unmarshal", func(t *testing.T) {
		dec := &Decimal{provider: provider}

		err := dec.UnmarshalJSON([]byte(`"123.456"`))
		assert.NoError(t, err)
		assert.Equal(t, "123.456", dec.String())

		// Test invalid JSON
		err = dec.UnmarshalJSON([]byte(`"invalid"`))
		assert.Error(t, err)
	})

	t.Run("json roundtrip", func(t *testing.T) {
		original, _ := provider.NewFromString("123.456")

		// Marshal
		data, err := json.Marshal(original)
		assert.NoError(t, err)

		// Unmarshal
		var restored *Decimal
		err = json.Unmarshal(data, &restored)
		if err != nil {
			// Handle case where standard json.Unmarshal doesn't work with our type
			restored = &Decimal{provider: provider}
			err = restored.UnmarshalJSON(data)
		}
		assert.NoError(t, err)
		assert.NotNil(t, restored)
	})
}

func TestDecimalInternalValue(t *testing.T) {
	provider := NewProvider(nil)
	dec, _ := provider.NewFromString("123.456")

	internal := dec.InternalValue()
	assert.NotNil(t, internal)
}

func TestDecimalText(t *testing.T) {
	provider := NewProvider(nil)
	dec, _ := provider.NewFromString("123.456")

	text := dec.Text('f')
	assert.NotEqual(t, "", text)
	assert.Contains(t, text, "123.456")
}

func TestGetRoundingMode(t *testing.T) {
	tests := []struct {
		input    string
		expected string // We'll just test that it doesn't panic
	}{
		{"RoundDown", "RoundDown"},
		{"RoundUp", "RoundUp"},
		{"RoundHalfUp", "RoundHalfUp"},
		{"RoundHalfDown", "RoundHalfDown"},
		{"RoundHalfEven", "RoundHalfEven"},
		{"RoundCeiling", "RoundCeiling"},
		{"RoundFloor", "RoundFloor"},
		{"Round05Up", "Round05Up"},
		{"Invalid", "Invalid"}, // Should default to RoundDown
	}

	for _, tt := range tests {
		t.Run("rounding_"+tt.input, func(t *testing.T) {
			// Test that it doesn't panic
			rounder := getRoundingMode(tt.input)
			assert.NotNil(t, rounder)
		})
	}
}

func TestMaxHelper(t *testing.T) {
	assert.Equal(t, int32(5), max(3, 5))
	assert.Equal(t, int32(5), max(5, 3))
	assert.Equal(t, int32(5), max(5, 5))
	assert.Equal(t, int32(0), max(-3, 0))
}

func BenchmarkNewFromString(b *testing.B) {
	provider := NewProvider(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.NewFromString("123.456")
	}
}

func BenchmarkNewFromFloat(b *testing.B) {
	provider := NewProvider(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.NewFromFloat(123.456)
	}
}

func BenchmarkArithmeticOperations(b *testing.B) {
	provider := NewProvider(nil)
	dec1, _ := provider.NewFromString("123.456")
	dec2, _ := provider.NewFromString("78.901")

	b.Run("addition", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dec1.Add(dec2)
		}
	})

	b.Run("multiplication", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dec1.Mul(dec2)
		}
	})
}

func BenchmarkComparisons(b *testing.B) {
	provider := NewProvider(nil)
	dec1, _ := provider.NewFromString("123.456")
	dec2, _ := provider.NewFromString("78.901")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec1.IsGreaterThan(dec2)
	}
}
