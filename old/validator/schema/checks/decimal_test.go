package checks

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDecimalByFactor8(t *testing.T) {
	decimal := NewDecimalByFactor8()

	assert.Equal(t, int32(-8), decimal.decimalFactor)
	assert.True(t, decimal.validateFactor)
}

func TestNewDecimal(t *testing.T) {
	decimal := NewDecimal()

	assert.Equal(t, int32(0), decimal.decimalFactor)
	assert.False(t, decimal.validateFactor)
}

func TestIsFormat(t *testing.T) {
	tests := []struct {
		name          string
		decimal       Decimal
		input         interface{}
		expectedValid bool
	}{
		{
			name: "when validateFactor is false should return true",
			decimal: Decimal{
				decimalFactor:  0,
				validateFactor: false,
			},
			input:         json.Number("123.456"),
			expectedValid: true,
		},
		{
			name: "when number has valid factor should return true",
			decimal: Decimal{
				decimalFactor:  -2,
				validateFactor: true,
			},
			input:         json.Number("123.45"),
			expectedValid: true,
		},
		{
			name: "when number has invalid factor should return false",
			decimal: Decimal{
				decimalFactor:  -2,
				validateFactor: true,
			},
			input:         json.Number("123.456"),
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.decimal.IsFormat(tt.input)
			assert.Equal(t, tt.expectedValid, result)
		})
	}
}

func TestGetDecimalValue(t *testing.T) {
	t.Run("with big.Rat input", func(t *testing.T) {
		rat := big.NewRat(12345, 100) // 123.45
		result := GetDecimalValue(rat)
		assert.NotNil(t, result)
	})

	t.Run("with json.Number input", func(t *testing.T) {
		number := json.Number("123.45")
		result := GetDecimalValue(number)
		assert.NotNil(t, result)
	})

	t.Run("with invalid json.Number", func(t *testing.T) {
		number := json.Number("invalid")
		result := GetDecimalValue(number)
		assert.Nil(t, result)
	})

	t.Run("with unsupported type", func(t *testing.T) {
		result := GetDecimalValue("123.45")
		assert.Nil(t, result)
	})
}
