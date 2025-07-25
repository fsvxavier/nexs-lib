package checks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONNumberChecker_IsFormat(t *testing.T) {
	checker := NewJSONNumberChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid json.Number", json.Number("123"), true},
		{"valid json.Number decimal", json.Number("123.45"), true},
		{"regular int", 123, false},
		{"regular float", 123.45, false},
		{"string number", "123", false},
		{"non-numeric", "abc", false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNumericChecker_IsFormat(t *testing.T) {
	checker := NewNumericChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		// Integer types
		{"int", 123, true},
		{"int8", int8(123), true},
		{"int16", int16(123), true},
		{"int32", int32(123), true},
		{"int64", int64(123), true},
		{"uint", uint(123), true},
		{"uint8", uint8(123), true},
		{"uint16", uint16(123), true},
		{"uint32", uint32(123), true},
		{"uint64", uint64(123), true},

		// Float types
		{"float32", float32(123.45), true},
		{"float64", float64(123.45), true},

		// JSON Number
		{"json.Number", json.Number("123.45"), true},

		// String numbers
		{"string number", "123.45", true},
		{"string integer", "123", true},

		// Invalid types
		{"string non-number", "abc", false},
		{"bool", true, false},
		{"nil", nil, false},

		// Zero value
		{"zero", 0, true},        // AllowZero is true by default
		{"negative", -123, true}, // AllowNegative is true by default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNumericChecker_Constraints(t *testing.T) {
	t.Run("no zero allowed", func(t *testing.T) {
		checker := NewNumericChecker()
		checker.AllowZero = false

		assert.False(t, checker.IsFormat(0))
		assert.True(t, checker.IsFormat(1))
		assert.True(t, checker.IsFormat(-1))
	})

	t.Run("no negative allowed", func(t *testing.T) {
		checker := NewNumericChecker()
		checker.AllowNegative = false

		assert.True(t, checker.IsFormat(0))
		assert.True(t, checker.IsFormat(1))
		assert.False(t, checker.IsFormat(-1))
	})

	t.Run("with range", func(t *testing.T) {
		checker := NewNumericChecker().WithRange(10, 20)

		assert.False(t, checker.IsFormat(5))
		assert.True(t, checker.IsFormat(10))
		assert.True(t, checker.IsFormat(15))
		assert.True(t, checker.IsFormat(20))
		assert.False(t, checker.IsFormat(25))
	})

	t.Run("with min value", func(t *testing.T) {
		checker := NewNumericChecker().WithMinValue(10)

		assert.False(t, checker.IsFormat(5))
		assert.True(t, checker.IsFormat(10))
		assert.True(t, checker.IsFormat(15))
	})

	t.Run("with max value", func(t *testing.T) {
		checker := NewNumericChecker().WithMaxValue(20)

		assert.True(t, checker.IsFormat(15))
		assert.True(t, checker.IsFormat(20))
		assert.False(t, checker.IsFormat(25))
	})
}

func TestDecimalChecker_IsFormat(t *testing.T) {
	checker := NewDecimalChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"float64", 123.45, true},
		{"float32", float32(123.45), true},
		{"int", 123, true},
		{"int64", int64(123), true},
		{"string number", "123.45", true},
		{"json.Number", json.Number("123.45"), true},
		{"zero", 0.0, true}, // AllowZero is true by default
		{"string non-number", "abc", false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecimalChecker_Factor8(t *testing.T) {
	checker := NewDecimalCheckerByFactor8()

	// With factor validation, it should still accept numeric values
	// but the actual factor validation is simplified in this implementation
	assert.True(t, checker.IsFormat(123.45))
	assert.True(t, checker.IsFormat("123.45"))
	assert.False(t, checker.IsFormat("abc"))
}

func TestDecimalChecker_NoZero(t *testing.T) {
	checker := NewDecimalChecker()
	checker.AllowZero = false

	assert.False(t, checker.IsFormat(0))
	assert.False(t, checker.IsFormat(0.0))
	assert.True(t, checker.IsFormat(0.1))
}

func TestIntegerChecker_IsFormat(t *testing.T) {
	checker := NewIntegerChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		// Pure integers
		{"int", 123, true},
		{"int8", int8(123), true},
		{"int16", int16(123), true},
		{"int32", int32(123), true},
		{"int64", int64(123), true},
		{"uint", uint(123), true},
		{"uint8", uint8(123), true},
		{"uint16", uint16(123), true},
		{"uint32", uint32(123), true},

		// Floats that are actually integers
		{"float32 integer", float32(123), true},
		{"float64 integer", float64(123), true},

		// Floats with decimals
		{"float32 decimal", float32(123.45), false},
		{"float64 decimal", float64(123.45), false},

		// String integers
		{"string integer", "123", true},
		{"string decimal", "123.45", false},

		// JSON Number
		{"json.Number integer", json.Number("123"), true},
		{"json.Number decimal", json.Number("123.45"), false},

		// Invalid
		{"string non-number", "abc", false},
		{"nil", nil, false},

		// Edge cases
		{"zero", 0, true},
		{"negative", -123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegerChecker_Constraints(t *testing.T) {
	t.Run("no zero allowed", func(t *testing.T) {
		checker := NewIntegerChecker()
		checker.AllowZero = false

		assert.False(t, checker.IsFormat(0))
		assert.True(t, checker.IsFormat(1))
		assert.True(t, checker.IsFormat(-1))
	})

	t.Run("no negative allowed", func(t *testing.T) {
		checker := NewIntegerChecker()
		checker.AllowNegative = false

		assert.True(t, checker.IsFormat(0))
		assert.True(t, checker.IsFormat(1))
		assert.False(t, checker.IsFormat(-1))
	})

	t.Run("with range", func(t *testing.T) {
		checker := NewIntegerChecker().WithRange(10, 20)

		assert.False(t, checker.IsFormat(5))
		assert.True(t, checker.IsFormat(10))
		assert.True(t, checker.IsFormat(15))
		assert.True(t, checker.IsFormat(20))
		assert.False(t, checker.IsFormat(25))
	})
}

func TestNumericChecker_Check(t *testing.T) {
	checker := NewNumericChecker()

	// Valid data
	errors := checker.Check(123)
	assert.Empty(t, errors)

	// Invalid data
	errors = checker.Check("abc")
	require.Len(t, errors, 1)
	assert.Equal(t, "numeric", errors[0].Field)
	assert.Equal(t, "INVALID_NUMERIC_VALUE", errors[0].ErrorType)
}
