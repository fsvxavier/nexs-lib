package checks

import (
	"encoding/json"
	"math/big"

	dec "github.com/fsvxavier/nexs-lib/decimal"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
	"github.com/fsvxavier/nexs-lib/decimal/providers/apd"
)

// DecimalChecker validates decimal format with optional factor validation
type DecimalChecker struct {
	validateFactor bool
	decimalFactor  int32
}

// NewDecimalByFactor8 creates a decimal checker with factor validation
func NewDecimalByFactor8() *DecimalChecker {
	return &DecimalChecker{
		decimalFactor:  -8,
		validateFactor: true,
	}
}

// NewDecimal creates a basic decimal checker
func NewDecimal() *DecimalChecker {
	return &DecimalChecker{
		decimalFactor:  0,
		validateFactor: false,
	}
}

// IsFormat validates if input is a valid decimal
func (d *DecimalChecker) IsFormat(input interface{}) bool {
	decimalValue := d.getDecimalValue(input)
	if decimalValue == nil {
		return false
	}

	if !d.validateFactor {
		return true
	}

	// Use type assertion to access the underlying decimal's methods
	if apdDecimal, ok := decimalValue.(*apd.Decimal); ok {
		// Use string representation to check decimal places
		str := apdDecimal.String()
		decimalPlaces := getDecimalPlaces(str)
		return !(decimalPlaces > -d.decimalFactor)
	}

	// Fallback: if we can't access exponent, just validate it's a valid decimal
	return true
}

// FormatName returns the name of this format checker
func (d *DecimalChecker) FormatName() string {
	if d.validateFactor {
		return "decimal_by_factor_of_8"
	}
	return "decimal"
}

// getDecimalValue converts input to decimal value
func (d *DecimalChecker) getDecimalValue(input interface{}) interfaces.Decimal {
	var inputAsFloat float64
	var err error

	switch inputAsNumber := input.(type) {
	case *big.Rat:
		inputAsFloat, _ = inputAsNumber.Float64()

	case json.Number:
		inputAsFloat, err = inputAsNumber.Float64()
		if err != nil {
			return nil
		}
	default:
		return nil
	}

	provider := dec.NewProvider(dec.APD)
	decimalValue, err := provider.NewFromFloat(inputAsFloat)
	if err != nil {
		return nil
	}
	return decimalValue
}

// getDecimalPlaces counts the number of decimal places in a decimal string
func getDecimalPlaces(str string) int32 {
	dotIndex := -1
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 {
		return 0
	}

	return int32(len(str) - dotIndex - 1)
}
