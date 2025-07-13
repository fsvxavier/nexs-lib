package checks

import (
	"encoding/json"
	"math/big"

	dcm "github.com/dock-tech/isis-golang-lib/decimal"
)

func NewDecimalByFactor8() Decimal {
	return Decimal{
		decimalFactor:  -8,
		validateFactor: true,
	}
}

func NewDecimal() Decimal {
	return Decimal{
		decimalFactor:  0,
		validateFactor: false,
	}
}

type Decimal struct {
	decimalFactor  int32
	validateFactor bool
}

func (d Decimal) IsFormat(input interface{}) bool {
	decimalValue := GetDecimalValue(input)

	if !d.validateFactor {
		return true
	}

	return !(decimalValue.Exponent < d.decimalFactor)
}

func GetDecimalValue(input interface{}) *dcm.Decimal {
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
	decimalValue := dcm.NewFromFloat(inputAsFloat)
	return decimalValue
}
