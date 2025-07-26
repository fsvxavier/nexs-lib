package cockroach

import (
	"fmt"
	"math"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/fsvxavier/nexs-lib/decimal/config"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

const (
	ProviderName    = "cockroach"
	ProviderVersion = "v3.0.0"
)

// Provider implements DecimalProvider using CockroachDB APD
type Provider struct {
	config *config.Config
	ctx    *apd.Context
}

// NewProvider creates a new CockroachDB decimal provider
func NewProvider(cfg *config.Config) *Provider {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Create APD context based on config
	apdCtx := &apd.Context{
		Precision:   cfg.GetMaxPrecision(),
		MaxExponent: cfg.GetMaxExponent(),
		MinExponent: cfg.GetMinExponent(),
		Rounding:    getRoundingMode(cfg.GetDefaultRounding()),
		Traps:       apd.DefaultTraps,
	}

	return &Provider{
		config: cfg,
		ctx:    apdCtx,
	}
}

func (p *Provider) Name() string {
	return ProviderName
}

func (p *Provider) Version() string {
	return ProviderVersion
}

func (p *Provider) NewFromString(value string) (interfaces.Decimal, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty string cannot be converted to decimal")
	}

	dec := &Decimal{
		provider: p,
	}

	_, _, err := dec.decimal.SetString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid decimal string '%s': %w", value, err)
	}

	return dec, nil
}

func (p *Provider) NewFromFloat(value float64) (interfaces.Decimal, error) {
	if math.IsNaN(value) {
		return nil, fmt.Errorf("NaN cannot be converted to decimal")
	}
	if math.IsInf(value, 0) {
		return nil, fmt.Errorf("infinity cannot be converted to decimal")
	}

	dec := &Decimal{
		provider: p,
	}

	_, err := dec.decimal.SetFloat64(value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert float64 %f to decimal: %w", value, err)
	}

	return dec, nil
}

func (p *Provider) NewFromInt(value int64) (interfaces.Decimal, error) {
	dec := &Decimal{
		provider: p,
	}

	dec.decimal.SetInt64(value)
	return dec, nil
}

func (p *Provider) Zero() interfaces.Decimal {
	dec := &Decimal{
		provider: p,
	}
	dec.decimal.SetInt64(0)
	return dec
}

// Decimal represents a decimal number using CockroachDB APD
type Decimal struct {
	decimal  apd.Decimal
	provider *Provider
}

func (d *Decimal) String() string {
	return d.decimal.String()
}

func (d *Decimal) Text(format byte) string {
	return d.decimal.Text(format)
}

// Comparison operations
func (d *Decimal) IsEqual(other interfaces.Decimal) bool {
	if other == nil {
		return false
	}

	otherCockroach, ok := other.(*Decimal)
	if !ok {
		// Try to convert using string representation
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return false
		}
		otherCockroach = otherDec.(*Decimal)
	}

	return d.decimal.Cmp(&otherCockroach.decimal) == 0
}

func (d *Decimal) IsGreaterThan(other interfaces.Decimal) bool {
	if other == nil {
		return !d.IsZero()
	}

	otherCockroach, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return false
		}
		otherCockroach = otherDec.(*Decimal)
	}

	return d.decimal.Cmp(&otherCockroach.decimal) > 0
}

func (d *Decimal) IsLessThan(other interfaces.Decimal) bool {
	if other == nil {
		return d.IsNegative()
	}

	otherCockroach, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return false
		}
		otherCockroach = otherDec.(*Decimal)
	}

	return d.decimal.Cmp(&otherCockroach.decimal) < 0
}

func (d *Decimal) IsGreaterThanOrEqual(other interfaces.Decimal) bool {
	return d.IsGreaterThan(other) || d.IsEqual(other)
}

func (d *Decimal) IsLessThanOrEqual(other interfaces.Decimal) bool {
	return d.IsLessThan(other) || d.IsEqual(other)
}

func (d *Decimal) IsZero() bool {
	return d.decimal.IsZero()
}

func (d *Decimal) IsPositive() bool {
	return d.decimal.Sign() > 0
}

func (d *Decimal) IsNegative() bool {
	return d.decimal.Sign() < 0
}

// Arithmetic operations
func (d *Decimal) Add(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot add to nil decimal")
	}

	otherCockroach, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherCockroach = otherDec.(*Decimal)
	}

	result := &Decimal{provider: d.provider}
	_, err := d.provider.ctx.Add(&result.decimal, &d.decimal, &otherCockroach.decimal)
	if err != nil {
		return nil, fmt.Errorf("addition failed: %w", err)
	}

	return result, nil
}

func (d *Decimal) Sub(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot subtract nil decimal")
	}

	otherCockroach, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherCockroach = otherDec.(*Decimal)
	}

	result := &Decimal{provider: d.provider}
	_, err := d.provider.ctx.Sub(&result.decimal, &d.decimal, &otherCockroach.decimal)
	if err != nil {
		return nil, fmt.Errorf("subtraction failed: %w", err)
	}

	return result, nil
}

func (d *Decimal) Mul(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot multiply by nil decimal")
	}

	otherCockroach, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherCockroach = otherDec.(*Decimal)
	}

	result := &Decimal{provider: d.provider}
	_, err := d.provider.ctx.Mul(&result.decimal, &d.decimal, &otherCockroach.decimal)
	if err != nil {
		return nil, fmt.Errorf("multiplication failed: %w", err)
	}

	return result, nil
}

func (d *Decimal) Div(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot divide by nil decimal")
	}

	if other.IsZero() {
		return nil, fmt.Errorf("division by zero")
	}

	otherCockroach, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherCockroach = otherDec.(*Decimal)
	}

	result := &Decimal{provider: d.provider}
	_, err := d.provider.ctx.Quo(&result.decimal, &d.decimal, &otherCockroach.decimal)
	if err != nil {
		return nil, fmt.Errorf("division failed: %w", err)
	}

	return result, nil
}

func (d *Decimal) Mod(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot mod by nil decimal")
	}

	if other.IsZero() {
		return nil, fmt.Errorf("modulo by zero")
	}

	otherCockroach, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherCockroach = otherDec.(*Decimal)
	}

	result := &Decimal{provider: d.provider}
	_, err := d.provider.ctx.Rem(&result.decimal, &d.decimal, &otherCockroach.decimal)
	if err != nil {
		return nil, fmt.Errorf("modulo failed: %w", err)
	}

	return result, nil
}

func (d *Decimal) Abs() interfaces.Decimal {
	result := &Decimal{provider: d.provider}
	d.provider.ctx.Abs(&result.decimal, &d.decimal)
	return result
}

func (d *Decimal) Neg() interfaces.Decimal {
	result := &Decimal{provider: d.provider}
	result.decimal.Neg(&d.decimal)
	return result
}

// Precision and rounding operations
func (d *Decimal) Truncate(precision uint32, minExponent int32) interfaces.Decimal {
	result := &Decimal{provider: d.provider}

	ctx := d.provider.ctx
	if precision > 0 {
		ctx = &apd.Context{
			Precision:   precision,
			MaxExponent: ctx.MaxExponent,
			MinExponent: ctx.MinExponent,
			Rounding:    ctx.Rounding,
			Traps:       ctx.Traps,
		}
	}

	numExponent := max(d.decimal.Exponent, minExponent)
	ctx.Quantize(&result.decimal, &d.decimal, numExponent)

	return result
}

func (d *Decimal) TrimZerosRight() interfaces.Decimal {
	formatDecimal := d.decimal.String()
	if d.decimal.Exponent < 0 {
		formatDecimal = strings.TrimRight(formatDecimal, "0")
		formatDecimal = strings.TrimRight(formatDecimal, ".")
	}

	result, _ := d.provider.NewFromString(formatDecimal)
	return result
}

func (d *Decimal) Round(places int32) interfaces.Decimal {
	result := &Decimal{provider: d.provider}
	d.provider.ctx.Quantize(&result.decimal, &d.decimal, -places)
	return result
}

// Conversion methods
func (d *Decimal) Float64() (float64, error) {
	f, err := d.decimal.Float64()
	if err != nil {
		return 0, fmt.Errorf("failed to convert decimal to float64: %w", err)
	}
	return f, nil
}

func (d *Decimal) Int64() (int64, error) {
	// Check if it's an integer by checking if exponent >= 0 or all decimal places are zero
	if d.decimal.Exponent < 0 {
		// Create a copy and truncate to check if it's effectively an integer
		truncated := &apd.Decimal{}
		d.provider.ctx.Quantize(truncated, &d.decimal, 0)
		if truncated.Cmp(&d.decimal) != 0 {
			return 0, fmt.Errorf("decimal %s is not an integer", d.String())
		}
	}

	i, err := d.decimal.Int64()
	if err != nil {
		return 0, fmt.Errorf("failed to convert decimal to int64: %w", err)
	}
	return i, nil
}

// JSON marshaling
func (d *Decimal) MarshalJSON() ([]byte, error) {
	truncated := d.Truncate(d.provider.config.GetMaxPrecision(), d.provider.config.GetMinExponent())
	return []byte(`"` + truncated.Text('f') + `"`), nil
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = strings.Trim(str, `"`)

	_, _, err := d.decimal.SetString(str)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to decimal: %w", err)
	}

	return nil
}

// InternalValue returns the internal APD decimal
func (d *Decimal) InternalValue() interface{} {
	return &d.decimal
}

// Helper function to convert string rounding mode to APD rounding mode
func getRoundingMode(mode string) apd.Rounder {
	switch mode {
	case "RoundDown":
		return apd.RoundDown
	case "RoundUp":
		return apd.RoundUp
	case "RoundHalfUp":
		return apd.RoundHalfUp
	case "RoundHalfDown":
		return apd.RoundHalfDown
	case "RoundHalfEven":
		return apd.RoundHalfEven
	case "RoundCeiling":
		return apd.RoundCeiling
	case "RoundFloor":
		return apd.RoundFloor
	case "Round05Up":
		return apd.Round05Up
	default:
		return apd.RoundDown
	}
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
