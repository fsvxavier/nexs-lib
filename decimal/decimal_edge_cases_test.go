package decimal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

func TestEdgeCasesArithmetic(t *testing.T) {
	manager := NewManager(nil)

	t.Run("very_small_numbers", func(t *testing.T) {
		small1, err := manager.NewFromString("0.001")
		require.NoError(t, err)
		small2, err := manager.NewFromString("0.002")
		require.NoError(t, err)

		// Addition
		sum, err := small1.Add(small2)
		assert.NoError(t, err)
		expected, _ := manager.NewFromString("0.003")
		assert.True(t, sum.IsEqual(expected))

		// Subtraction
		diff, err := small2.Sub(small1)
		assert.NoError(t, err)
		expected, _ = manager.NewFromString("0.001")
		assert.True(t, diff.IsEqual(expected))

		// Multiplication
		product, err := small1.Mul(small2)
		assert.NoError(t, err)
		assert.True(t, product.IsPositive())
		assert.False(t, product.IsZero())
	})

	t.Run("very_large_numbers", func(t *testing.T) {
		large1, err := manager.NewFromString("123456789.123456")
		require.NoError(t, err)
		large2, err := manager.NewFromString("987654321.654321")
		require.NoError(t, err)

		// Addition should work
		sum, err := large1.Add(large2)
		assert.NoError(t, err)
		assert.True(t, sum.IsPositive())
		assert.True(t, sum.IsGreaterThan(large1))
		assert.True(t, sum.IsGreaterThan(large2))

		// Subtraction
		diff, err := large1.Sub(large2)
		assert.NoError(t, err)
		assert.True(t, diff.IsNegative()) // large2 is bigger than large1
	})

	t.Run("precision_edge_cases", func(t *testing.T) {
		// Test precision limits
		precise1, err := manager.NewFromString("1.12345678901234567890")
		require.NoError(t, err)
		precise2, err := manager.NewFromString("2.98765432109876543210")
		require.NoError(t, err)

		sum, err := precise1.Add(precise2)
		assert.NoError(t, err)
		assert.True(t, sum.IsPositive())

		// Test rounding behavior
		product, err := precise1.Mul(precise2)
		assert.NoError(t, err)
		assert.True(t, product.IsPositive())
	})

	t.Run("zero_operations", func(t *testing.T) {
		zero := manager.Zero()
		positive, _ := manager.NewFromString("123.456")
		negative, _ := manager.NewFromString("-789.012")

		// Adding zero
		result, err := positive.Add(zero)
		assert.NoError(t, err)
		assert.True(t, result.IsEqual(positive))

		result, err = zero.Add(negative)
		assert.NoError(t, err)
		assert.True(t, result.IsEqual(negative))

		// Subtracting zero
		result, err = positive.Sub(zero)
		assert.NoError(t, err)
		assert.True(t, result.IsEqual(positive))

		result, err = zero.Sub(positive)
		assert.NoError(t, err)
		expected, _ := manager.NewFromString("-123.456")
		assert.True(t, result.IsEqual(expected))

		// Multiplying by zero
		result, err = positive.Mul(zero)
		assert.NoError(t, err)
		assert.True(t, result.IsZero())

		result, err = zero.Mul(negative)
		assert.NoError(t, err)
		assert.True(t, result.IsZero())
	})

	t.Run("negative_operations", func(t *testing.T) {
		positive, _ := manager.NewFromString("123.456")
		negative, _ := manager.NewFromString("-123.456")

		// Adding positive and negative
		result, err := positive.Add(negative)
		assert.NoError(t, err)
		assert.True(t, result.IsZero())

		// Subtracting negative (double negative)
		result, err = positive.Sub(negative)
		assert.NoError(t, err)
		expected, _ := manager.NewFromString("246.912")
		assert.True(t, result.IsEqual(expected))

		// Multiplying negative
		result, err = positive.Mul(negative)
		assert.NoError(t, err)
		assert.True(t, result.IsNegative())
		expectedNeg, _ := manager.NewFromString("-15241.383936")
		assert.True(t, result.IsEqual(expectedNeg))
	})

	t.Run("division_edge_cases", func(t *testing.T) {
		dividend, _ := manager.NewFromString("100")
		divisor, _ := manager.NewFromString("3")

		// Division that results in repeating decimal
		result, err := dividend.Div(divisor)
		assert.NoError(t, err)
		assert.True(t, result.IsPositive())
		// Should be approximately 33.333...
		threshold33, _ := manager.NewFromString("33")
		threshold34, _ := manager.NewFromString("34")
		assert.True(t, result.IsGreaterThan(threshold33))
		assert.True(t, result.IsLessThan(threshold34))

		// Division by small number (but not too small to avoid subnormal)
		small, _ := manager.NewFromString("0.1")
		result, err = dividend.Div(small)
		assert.NoError(t, err)
		assert.True(t, result.IsPositive())
		assert.True(t, result.IsGreaterThan(dividend))

		// Division of small by large
		large, _ := manager.NewFromString("1000")
		smallDividend, _ := manager.NewFromString("1")
		result, err = smallDividend.Div(large)
		assert.NoError(t, err)
		assert.True(t, result.IsPositive())
		assert.True(t, result.IsLessThan(smallDividend))
	})
}

func TestEdgeCasesBatchOperations(t *testing.T) {
	manager := NewManager(nil)

	t.Run("mixed_sign_numbers", func(t *testing.T) {
		positive, _ := manager.NewFromString("100")
		negative, _ := manager.NewFromString("-50")
		zero := manager.Zero()
		decimals := []interfaces.Decimal{positive, negative, zero}

		// Sum should be 50
		sum, err := manager.SumSlice(decimals)
		assert.NoError(t, err)
		expected, _ := manager.NewFromString("50")
		assert.True(t, sum.IsEqual(expected))

		// Max should be positive
		max, err := manager.MaxSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, max.IsEqual(positive))

		// Min should be negative
		min, err := manager.MinSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, min.IsEqual(negative))

		// Batch processor
		processor := manager.NewBatchProcessor()
		result, err := processor.ProcessSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, result.Sum.IsEqual(expected))
		assert.True(t, result.Max.IsEqual(positive))
		assert.True(t, result.Min.IsEqual(negative))
	})

	t.Run("all_same_numbers", func(t *testing.T) {
		value, _ := manager.NewFromString("42.42")
		decimals := []interfaces.Decimal{value, value, value, value, value}

		sum, err := manager.SumSlice(decimals)
		assert.NoError(t, err)
		expected, _ := manager.NewFromString("212.1")
		assert.True(t, sum.IsEqual(expected))

		avg, err := manager.AverageSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, avg.IsEqual(value))

		max, err := manager.MaxSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, max.IsEqual(value))

		min, err := manager.MinSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, min.IsEqual(value))
	})

	t.Run("single_element", func(t *testing.T) {
		single, _ := manager.NewFromString("123.456")
		decimals := []interfaces.Decimal{single}

		sum, err := manager.SumSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, sum.IsEqual(single))

		avg, err := manager.AverageSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, avg.IsEqual(single))

		max, err := manager.MaxSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, max.IsEqual(single))

		min, err := manager.MinSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, min.IsEqual(single))

		// Batch processor
		processor := manager.NewBatchProcessor()
		result, err := processor.ProcessSlice(decimals)
		assert.NoError(t, err)
		assert.True(t, result.Sum.IsEqual(single))
		assert.True(t, result.Average.IsEqual(single))
		assert.True(t, result.Max.IsEqual(single))
		assert.True(t, result.Min.IsEqual(single))
		assert.Equal(t, 1, result.Count)
	})

	t.Run("large_dataset", func(t *testing.T) {
		size := 1000
		decimals := make([]interfaces.Decimal, size)

		// Create decimals from 1 to 1000
		for i := 0; i < size; i++ {
			val, err := manager.NewFromInt(int64(i + 1))
			require.NoError(t, err)
			decimals[i] = val
		}

		// Test batch processor efficiency
		processor := manager.NewBatchProcessor()
		result, err := processor.ProcessSlice(decimals)
		assert.NoError(t, err)

		// Sum should be 1+2+...+1000 = 500500
		expectedSum, _ := manager.NewFromString("500500")
		assert.True(t, result.Sum.IsEqual(expectedSum))

		// Average should be 500.5
		expectedAvg, _ := manager.NewFromString("500.5")
		assert.True(t, result.Average.IsEqual(expectedAvg))

		// Max should be 1000
		expectedMax, _ := manager.NewFromInt(1000)
		assert.True(t, result.Max.IsEqual(expectedMax))

		// Min should be 1
		expectedMin, _ := manager.NewFromInt(1)
		assert.True(t, result.Min.IsEqual(expectedMin))

		assert.Equal(t, size, result.Count)
	})
}

func TestEdgeCasesStringConversion(t *testing.T) {
	manager := NewManager(nil)

	t.Run("scientific_notation", func(t *testing.T) {
		// Test various scientific notation formats
		tests := []struct {
			input    string
			hasError bool
		}{
			{"1e5", false},    // 100000
			{"1E5", false},    // 100000
			{"1.5e3", false},  // 1500
			{"1.5E-3", false}, // 0.0015
			{"-2.5e2", false}, // -250
			{"1e", true},      // invalid
			{"e5", true},      // invalid
			{"1.5e", true},    // invalid
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result, err := manager.NewFromString(tt.input)
				if tt.hasError {
					assert.Error(t, err)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
			})
		}
	})

	t.Run("leading_trailing_zeros", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"000123.456000", "123.456"},
			{"0.0100", "0.01"},
			{"100.000", "100"},
			{"0000.0000", "0"},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result, err := manager.NewFromString(tt.input)
				assert.NoError(t, err)
				assert.NotNil(t, result)

				expected, err := manager.NewFromString(tt.expected)
				assert.NoError(t, err)
				assert.True(t, result.IsEqual(expected))
			})
		}
	})

	t.Run("invalid_strings", func(t *testing.T) {
		invalidStrings := []string{
			"",
			"abc",
			"12.34.56",
			"12,34",
			"12.34e",
			"12.34.e5",
			"++12.34",
			"--12.34",
			"12..34",
			"12.34..",
		}

		for _, invalid := range invalidStrings {
			t.Run(invalid, func(t *testing.T) {
				result, err := manager.NewFromString(invalid)
				assert.Error(t, err)
				assert.Nil(t, result)
			})
		}
	})
}

func TestEdgeCasesTypeConversion(t *testing.T) {
	manager := NewManager(nil)

	t.Run("float_edge_cases", func(t *testing.T) {
		tests := []struct {
			name  string
			input float64
		}{
			{"small_positive", 1e-10},
			{"small_negative", -1e-10},
			{"large_positive", 1e10},
			{"large_negative", -1e10},
			{"precise_decimal", 0.1 + 0.2}, // Known floating point precision issue
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := manager.NewFromFloat(tt.input)
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Convert back to float and verify it's close
				backToFloat, err := result.Float64()
				assert.NoError(t, err)
				assert.InDelta(t, tt.input, backToFloat, 1e-15)
			})
		}
	})

	t.Run("int_boundary_values", func(t *testing.T) {
		tests := []struct {
			name  string
			input int64
		}{
			{"max_int64", 9223372036854775807},
			{"min_int64", -9223372036854775808},
			{"zero", 0},
			{"positive_one", 1},
			{"negative_one", -1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := manager.NewFromInt(tt.input)
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Convert back to int64 and verify
				backToInt, err := result.Int64()
				assert.NoError(t, err)
				assert.Equal(t, tt.input, backToInt)
			})
		}
	})

	t.Run("parse_interface_edge_cases", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			hasError bool
		}{
			{"nil", nil, true},
			{"empty_string", "", true},
			{"bool_true", true, true},
			{"bool_false", false, true},
			{"slice", []int{1, 2, 3}, true},
			{"map", map[string]int{"a": 1}, true},
			{"channel", make(chan int), true},
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
				}
			})
		}
	})
}

// Benchmarks for edge cases
func BenchmarkEdgeCases(b *testing.B) {
	manager := NewManager(nil)

	b.Run("VeryLargeNumbers", func(b *testing.B) {
		large, _ := manager.NewFromString("123456789.123456")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			large.Add(large)
		}
	})

	b.Run("VerySmallNumbers", func(b *testing.B) {
		small, _ := manager.NewFromString("0.0001")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			small.Mul(small)
		}
	})

	b.Run("MixedSignBatch", func(b *testing.B) {
		positive, _ := manager.NewFromString("100")
		negative, _ := manager.NewFromString("-50")
		zero := manager.Zero()
		decimals := []interfaces.Decimal{positive, negative, zero}

		processor := manager.NewBatchProcessor()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			processor.ProcessSlice(decimals)
		}
	})
}
