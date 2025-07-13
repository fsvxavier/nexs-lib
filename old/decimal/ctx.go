package decimal

import (
	"github.com/cockroachdb/apd/v3"
)

const (
	MAX_PRECISION    = 21 // total number of digits, before and after decimal points
	MAX_EXPONENT     = 13 // total number of digits, after decimal points
	MIN_EXPONENT     = -8 // total number of digits, before decimal points
	DEFAULT_ROUNDING = apd.RoundDown
)

type context struct {
	*apd.Context
}

var apdCtx = &apd.Context{
	Precision:   MAX_PRECISION,
	MaxExponent: MAX_EXPONENT,
	MinExponent: MIN_EXPONENT,
	Rounding:    DEFAULT_ROUNDING,
	Traps:       apd.DefaultTraps,
}

var Ctx = &context{apdCtx}

// Abs sets d to |x| (the absolute value of x).
// If x is NaN, d will be set to NaN.
// If x is negative, d will be set to -x.
// If x is positive, d will be set to x.
// If x is zero, d will be set to zero.
// If x is infinite, d will be set to infinity.
// If x is NaN, d will be set to NaN.
// If x is zero, d will be set to zero.
// If x is infinite, d will be set to infinity.
func (c *context) Abs(d, x *Decimal) error {
	_, err := apdCtx.Abs(&d.Decimal, &x.Decimal)

	return err
}

// Add sets d to the sum x+y.
// If x or y is NaN, d will be set to NaN.
// If x or y is infinite, d will be set to infinity.
// If x or y is zero, d will be set to the other value.
// If both x and y are zero, d will be set to zero.
// If both x and y are NaN, d will be set to NaN.
// If both x and y are infinite, d will be set to infinity.
func (c *context) Add(d, x, y *Decimal) error {
	_, err := apdCtx.Add(&d.Decimal, &x.Decimal, &y.Decimal)

	return err
}

// Sub sets d to the difference x-y.
func (c *context) Sub(d, x, y *Decimal) error {
	_, err := apdCtx.Sub(&d.Decimal, &x.Decimal, &y.Decimal)

	return err
}

// Mul sets d to the product x*y.
func (c *context) Mul(d, x, y *Decimal) error {
	_, err := apdCtx.Mul(&d.Decimal, &x.Decimal, &y.Decimal)

	return err
}

// Quo sets d to the quotient x/y for y != 0. c.Precision must be > 0.
// If an exact division is required, use a context with high precision and verify it was exact by checking the Inexact flag on the return Condition.
func (c *context) Quo(d, x, y *Decimal) error {
	_, err := apdCtx.Quo(&d.Decimal, &x.Decimal, &y.Decimal)

	return err
}

// Neg sets d to -x.
func (c *context) Neg(d, x *Decimal) error {
	_, err := apdCtx.Neg(&d.Decimal, &x.Decimal)

	return err
}

// NewFromString creates a new decimal from string s.
// The returned Decimal has its exponents restricted by the context and its value rounded if it contains more digits than the context's precision.
// If you need to convert a string to Decimal, use this function.
// Note that the conversion from string to Decimal may lose precision if the string represents a number with more digits than the context's precision.
func (c *context) NewFromString(s string) *Decimal {
	dec := &Decimal{}
	dec.SetString(s)

	return dec
}

// NewFromInt creates a new decimal from int64 i.
// The returned Decimal has its exponents restricted by the context and its value rounded if it contains more digits than the context's precision.
// If you need to convert an int64 to Decimal, use this function.
// Note that the conversion from int64 to Decimal is exact, so it is safe to use this function for integer values.
func (c *context) NewFromInt(i int64) *Decimal {
	dec := &Decimal{}
	dec.SetInt64(i)

	return dec
}

// NewFromFloat creates a new decimal from float64 f.
// The returned Decimal has its exponents restricted by the context and its value rounded if it contains more digits than the context's precision.
// Note that the conversion from float64 to Decimal may lose precision, so it is recommended to use NewFromString for exact decimal representations.
// If you need to convert a float64 to Decimal, use this function.
func (c *context) NewFromFloat(f float64) *Decimal {
	dec := &Decimal{}
	dec.SetFloat64(f)

	return dec
}

// IsEqual returns true if d and x are equal.
// It compares the two Decimals and returns true if they are equal, false otherwise.
// If either d or x is NaN, it returns false.
// If either d or x is infinite, it returns false.
// If either d or x is zero, it returns true if the other value is also zero.
// If both d and x are NaN, it returns false.
// If both d and x are infinite, it returns false.
func (c *context) IsEqual(d, x *Decimal) bool {
	return d.Cmp(&x.Decimal) == 0
}

// IsGreaterThan returns true if d is greater than x.
// It compares the two Decimals and returns true if d is greater than x, false otherwise.
// If either d or x is NaN, it returns false.
// If either d or x is infinite, it returns true if d is greater than x, false otherwise.
func (c *context) IsGreaterThan(d, x *Decimal) bool {
	return d.Cmp(&x.Decimal) > 0
}

// IsLessThan returns true if d is less than x.
// It compares the two Decimals and returns true if d is less than x, false otherwise.
// If either d or x is NaN, it returns false.
// If either d or x is infinite, it returns true if d is less than x, false otherwise.
func (c *context) IsLessThan(d, x *Decimal) bool {
	return d.Cmp(&x.Decimal) < 0
}

// IsGreaterThanOrEqual returns true if d is greater than or equal to x.
// It compares the two Decimals and returns true if d is greater than or equal to x, false otherwise.
// If either d or x is NaN, it returns false.
// If either d or x is infinite, it returns true if d is greater than or equal to x, false otherwise.
// If either d or x is zero, it returns true if the other value is also zero.
func (c *context) IsGreaterThanOrEqual(d, x *Decimal) bool {
	return d.Cmp(&x.Decimal) >= 0
}

// IsLessThanOrEqual returns true if d is less than or equal to x.
// It compares the two Decimals and returns true if d is less than or equal to x, false otherwise.
// If either d or x is NaN, it returns false.
// If either d or x is infinite, it returns true if d is less than or equal to x, false otherwise.
// If either d or x is zero, it returns true if the other value is also zero.
func (c *context) IsLessThanOrEqual(d, x *Decimal) bool {
	return d.Cmp(&x.Decimal) <= 0
}
