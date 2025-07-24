package decimal

import (
	"bytes"
	"encoding/json"
	j "encoding/json"
	"strings"

	"github.com/cockroachdb/apd/v3"
	jsoniter "github.com/json-iterator/go"
)

var jsonSTD = jsoniter.ConfigCompatibleWithStandardLibrary

// Decimal is a wrapper around apd.Decimal that implements the json.Marshaler and json.Unmarshaler interfaces.
type Decimal struct {
	apd.Decimal
}

// NewFromString creates a new Decimal from a string value.
// The returned Decimal has its exponents restricted by the context and its value rounded if it contains more digits than the context's precision.
// If you need to convert a string to Decimal, use this function.
// Note that the conversion from string to Decimal may lose precision, so it is recommended to use NewFromString for exact decimal representations.
func NewFromString(numString string) *Decimal {
	dec := &Decimal{}
	dec.SetString(numString)

	return dec
}

// NewFromFloat creates a new Decimal from a float64 value.
// The returned Decimal has its exponents restricted by the context and its value rounded if it contains more digits than the context's precision.
// If you need to convert a float64 to Decimal, use this function.
// Note that the conversion from float64 to Decimal may lose precision, so it is recommended to use NewFromString for exact decimal representations.
func NewFromFloat(numFloat float64) *Decimal {
	dec := &Decimal{}
	dec.SetFloat64(numFloat)

	return dec
}

// NewFromInt creates a new Decimal from an int64 value.
// The returned Decimal has its exponents restricted by the context and its value rounded if it contains more digits than the context's precision.
// If you need to convert an int64 to Decimal, use this function.
// Note that the conversion from int64 to Decimal is exact, so it is safe to use this function for integer values.
func NewFromInt(numInt int64) *Decimal {
	dec := &Decimal{}
	dec.SetInt64(numInt)

	return dec
}

func Truncate(nDecimal *Decimal, minExponent int32) *Decimal {
	c := apd.BaseContext.WithPrecision(MAX_PRECISION)
	c.Rounding = DEFAULT_ROUNDING
	output := Decimal{}

	// if the exponent of the number to be truncated is less than -8, assign minExponent to numExponent.
	// numExponent := nDecimal.Exponent
	// if numExponent < minExponent {
	// 	numExponent = minExponent
	// }
	numExponent := max(nDecimal.Exponent, minExponent)

	c.Quantize(&output.Decimal, &nDecimal.Decimal, numExponent)

	return &output
}

// Truncate returns a new Decimal with the precision and exponent truncated.
func TrimZerosRight(nDecimal *Decimal) *Decimal {
	/*
		will remove trailing zeros and then to remove the decimal point if there are no more digits after it.
		For example, the number 10.500 is formatted as a string and then the decimal zeros are stripped, resulting in 10.5
	*/
	formatDecimal := nDecimal.String()
	if nDecimal.Exponent < 0 {
		formatDecimal = strings.TrimRight(formatDecimal, "0")
		formatDecimal = strings.TrimRight(formatDecimal, ".")
	}

	return NewFromString(formatDecimal)
}

// UnmarshalJSON sets d to a copy of data with the Decimal type.
func (d *Decimal) UnmarshalJSON(b []byte) error {
	_, _, err := d.SetString(string(b))

	return err
}

// MarshalJSON returns d as the JSON encoding of d which is json.Number.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return jsonSTD.Marshal(j.Number(d.Truncate(MAX_PRECISION, MIN_EXPONENT, DEFAULT_ROUNDING).Text('f')))
}

// IsEqual returns true if d and x are equal.
func (d *Decimal) IsEqual(x *Decimal) bool {
	return d.Cmp(&x.Decimal) == 0
}

// IsGreaterThan returns true if d is greater than x.
func (d *Decimal) IsGreaterThan(x *Decimal) bool {
	return d.Cmp(&x.Decimal) > 0
}

// IsLessThan returns true if d is less than x.
func (d *Decimal) IsLessThan(x *Decimal) bool {
	return d.Cmp(&x.Decimal) < 0
}

// IsGreaterThanOrEqual returns true if d is greater than or equal to x.
func (d *Decimal) IsGreaterThanOrEqual(x *Decimal) bool {
	return d.Cmp(&x.Decimal) >= 0
}

// IsLessThanOrEqual returns true if d is less than or equal to x.
func (d *Decimal) IsLessThanOrEqual(x *Decimal) bool {
	return d.Cmp(&x.Decimal) <= 0
}

// Truncate returns a new Decimal with the precision and exponent truncated.
func (d *Decimal) Truncate(precision uint32, minExponent int32, rounding apd.Rounder) *Decimal {
	o := &apd.Decimal{}
	c := apd.BaseContext.WithPrecision(precision)
	c.Rounding = rounding

	// if the exponent of the number to be truncated is less than -8, assign minExponent to numExponent.
	// numExponent := d.Exponent
	// if numExponent < minExponent {
	// 	numExponent = minExponent
	// }
	numExponent := max(d.Exponent, minExponent)
	c.Quantize(o, &d.Decimal, numExponent)

	return &Decimal{*o}
}

// Remove trailing zeros and then to remove the decimal point if there are no more digits after it.
// For example, the number 10.500 is formatted as a string and then the decimal zeros are stripped, resulting in 10.5
func (d *Decimal) TrimZerosRight() *Decimal {

	formatDecimal := d.Decimal.String()
	if d.Exponent < 0 {
		formatDecimal = strings.TrimRight(formatDecimal, "0")
		formatDecimal = strings.TrimRight(formatDecimal, ".")
	}

	return NewFromString(formatDecimal)
}

func Marshal(i any) []byte {
	b, _ := json.Marshal(i)
	return b
}

func Unmarshal(b []byte, i any) error {
	r := bytes.NewReader(b)

	dec := json.NewDecoder(r)
	dec.UseNumber()

	return dec.Decode(&i)
}
