package shopspring

import (
	"fmt"
	"math"
	"strings"

	"github.com/fsvxavier/nexs-lib/decimal/config"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
	"github.com/shopspring/decimal"
)

const (
	ProviderName    = "shopspring"
	ProviderVersion = "v1.4.0"
)

// Provider implements DecimalProvider using Shopspring decimal
type Provider struct {
	config *config.Config
}

// NewProvider creates a new Shopspring decimal provider
func NewProvider(cfg *config.Config) *Provider {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	return &Provider{
		config: cfg,
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

	dec, err := decimal.NewFromString(value)
	if err != nil {
		return nil, fmt.Errorf("invalid decimal string '%s': %w", value, err)
	}

	return &Decimal{
		decimal:  dec,
		provider: p,
	}, nil
}

func (p *Provider) NewFromFloat(value float64) (interfaces.Decimal, error) {
	if math.IsNaN(value) {
		return nil, fmt.Errorf("NaN cannot be converted to decimal")
	}
	if math.IsInf(value, 0) {
		return nil, fmt.Errorf("infinity cannot be converted to decimal")
	}

	dec := decimal.NewFromFloat(value)
	return &Decimal{
		decimal:  dec,
		provider: p,
	}, nil
}

func (p *Provider) NewFromInt(value int64) (interfaces.Decimal, error) {
	dec := decimal.NewFromInt(value)
	return &Decimal{
		decimal:  dec,
		provider: p,
	}, nil
}

func (p *Provider) Zero() interfaces.Decimal {
	return &Decimal{
		decimal:  decimal.Zero,
		provider: p,
	}
}

// Decimal represents a decimal number using Shopspring decimal
type Decimal struct {
	decimal  decimal.Decimal
	provider *Provider
}

func (d *Decimal) String() string {
	return d.decimal.String()
}

func (d *Decimal) Text(format byte) string {
	switch format {
	case 'f':
		return d.decimal.String()
	case 'e', 'E':
		return d.decimal.StringFixed(int32(d.provider.config.GetMaxPrecision()))
	default:
		return d.decimal.String()
	}
}

// Comparison operations
func (d *Decimal) IsEqual(other interfaces.Decimal) bool {
	if other == nil {
		return false
	}

	otherShopspring, ok := other.(*Decimal)
	if !ok {
		// Try to convert using string representation
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return false
		}
		otherShopspring = otherDec.(*Decimal)
	}

	return d.decimal.Equal(otherShopspring.decimal)
}

func (d *Decimal) IsGreaterThan(other interfaces.Decimal) bool {
	if other == nil {
		return !d.IsZero()
	}

	otherShopspring, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return false
		}
		otherShopspring = otherDec.(*Decimal)
	}

	return d.decimal.GreaterThan(otherShopspring.decimal)
}

func (d *Decimal) IsLessThan(other interfaces.Decimal) bool {
	if other == nil {
		return d.IsNegative()
	}

	otherShopspring, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return false
		}
		otherShopspring = otherDec.(*Decimal)
	}

	return d.decimal.LessThan(otherShopspring.decimal)
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
	return d.decimal.IsPositive()
}

func (d *Decimal) IsNegative() bool {
	return d.decimal.IsNegative()
}

// Arithmetic operations
func (d *Decimal) Add(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot add to nil decimal")
	}

	otherShopspring, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherShopspring = otherDec.(*Decimal)
	}

	result := d.decimal.Add(otherShopspring.decimal)
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}, nil
}

func (d *Decimal) Sub(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot subtract nil decimal")
	}

	otherShopspring, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherShopspring = otherDec.(*Decimal)
	}

	result := d.decimal.Sub(otherShopspring.decimal)
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}, nil
}

func (d *Decimal) Mul(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot multiply by nil decimal")
	}

	otherShopspring, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherShopspring = otherDec.(*Decimal)
	}

	result := d.decimal.Mul(otherShopspring.decimal)
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}, nil
}

func (d *Decimal) Div(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot divide by nil decimal")
	}

	if other.IsZero() {
		return nil, fmt.Errorf("division by zero")
	}

	otherShopspring, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherShopspring = otherDec.(*Decimal)
	}

	result := d.decimal.Div(otherShopspring.decimal)
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}, nil
}

func (d *Decimal) Mod(other interfaces.Decimal) (interfaces.Decimal, error) {
	if other == nil {
		return nil, fmt.Errorf("cannot mod by nil decimal")
	}

	if other.IsZero() {
		return nil, fmt.Errorf("modulo by zero")
	}

	otherShopspring, ok := other.(*Decimal)
	if !ok {
		otherStr := other.String()
		otherDec, err := d.provider.NewFromString(otherStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert other decimal: %w", err)
		}
		otherShopspring = otherDec.(*Decimal)
	}

	result := d.decimal.Mod(otherShopspring.decimal)
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}, nil
}

func (d *Decimal) Abs() interfaces.Decimal {
	result := d.decimal.Abs()
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}
}

func (d *Decimal) Neg() interfaces.Decimal {
	result := d.decimal.Neg()
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}
}

// Precision and rounding operations
func (d *Decimal) Truncate(precision uint32, minExponent int32) interfaces.Decimal {
	// Shopspring doesn't have direct support for minExponent, so we'll use places
	places := max(0, -minExponent)
	result := d.decimal.Truncate(places)

	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}
}

func (d *Decimal) TrimZerosRight() interfaces.Decimal {
	// Convert to string, trim zeros, then back to decimal
	str := d.decimal.String()
	if strings.Contains(str, ".") {
		str = strings.TrimRight(str, "0")
		str = strings.TrimRight(str, ".")
	}

	result, _ := decimal.NewFromString(str)
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}
}

func (d *Decimal) Round(places int32) interfaces.Decimal {
	result := d.decimal.Round(places)
	return &Decimal{
		decimal:  result,
		provider: d.provider,
	}
}

// Conversion methods
func (d *Decimal) Float64() (float64, error) {
	f, exact := d.decimal.Float64()
	if !exact {
		return f, fmt.Errorf("conversion to float64 lost precision")
	}
	return f, nil
}

func (d *Decimal) Int64() (int64, error) {
	if !d.decimal.IsInteger() {
		return 0, fmt.Errorf("decimal %s is not an integer", d.String())
	}

	// Check if it fits in int64
	if d.decimal.GreaterThan(decimal.NewFromInt(math.MaxInt64)) {
		return 0, fmt.Errorf("decimal %s too large for int64", d.String())
	}
	if d.decimal.LessThan(decimal.NewFromInt(math.MinInt64)) {
		return 0, fmt.Errorf("decimal %s too small for int64", d.String())
	}

	return d.decimal.IntPart(), nil
}

// JSON marshaling
func (d *Decimal) MarshalJSON() ([]byte, error) {
	// Apply truncation based on config
	truncated := d.Truncate(d.provider.config.GetMaxPrecision(), d.provider.config.GetMinExponent())
	return []byte(`"` + truncated.String() + `"`), nil
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = strings.Trim(str, `"`)

	dec, err := decimal.NewFromString(str)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to decimal: %w", err)
	}

	d.decimal = dec
	return nil
}

// InternalValue returns the internal Shopspring decimal
func (d *Decimal) InternalValue() interface{} {
	return d.decimal
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
