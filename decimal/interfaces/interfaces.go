package interfaces

import (
	"context"
)

// DecimalProvider defines the contract for decimal operations
type DecimalProvider interface {
	// Creation methods
	NewFromString(value string) (Decimal, error)
	NewFromFloat(value float64) (Decimal, error)
	NewFromInt(value int64) (Decimal, error)
	Zero() Decimal

	// Provider information
	Name() string
	Version() string
}

// Decimal represents a decimal number with various operations
type Decimal interface {
	// String representation
	String() string
	Text(format byte) string

	// Comparison operations
	IsEqual(other Decimal) bool
	IsGreaterThan(other Decimal) bool
	IsLessThan(other Decimal) bool
	IsGreaterThanOrEqual(other Decimal) bool
	IsLessThanOrEqual(other Decimal) bool
	IsZero() bool
	IsPositive() bool
	IsNegative() bool

	// Arithmetic operations
	Add(other Decimal) (Decimal, error)
	Sub(other Decimal) (Decimal, error)
	Mul(other Decimal) (Decimal, error)
	Div(other Decimal) (Decimal, error)
	Mod(other Decimal) (Decimal, error)
	Abs() Decimal
	Neg() Decimal

	// Precision and rounding operations
	Truncate(precision uint32, minExponent int32) Decimal
	TrimZerosRight() Decimal
	Round(places int32) Decimal

	// Conversion methods
	Float64() (float64, error)
	Int64() (int64, error)

	// JSON marshaling
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error

	// Internal representation access
	InternalValue() interface{}
}

// DecimalManager provides high-level decimal operations with provider management
type DecimalManager interface {
	// Provider management
	SetProvider(provider DecimalProvider)
	GetProvider() DecimalProvider
	SwitchProvider(providerName string) error

	// Factory methods with current provider
	NewFromString(value string) (Decimal, error)
	NewFromFloat(value float64) (Decimal, error)
	NewFromInt(value int64) (Decimal, error)
	Zero() Decimal

	// Batch operations
	Sum(decimals ...Decimal) (Decimal, error)
	Average(decimals ...Decimal) (Decimal, error)
	Max(decimals ...Decimal) (Decimal, error)
	Min(decimals ...Decimal) (Decimal, error)

	// Utility operations
	Parse(value interface{}) (Decimal, error)
	MarshalJSON(decimal Decimal) ([]byte, error)
	UnmarshalJSON(data []byte) (Decimal, error)
}

// PreHook defines operations to execute before decimal operations
type PreHook interface {
	Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error)
}

// PostHook defines operations to execute after decimal operations
type PostHook interface {
	Execute(ctx context.Context, operation string, result interface{}, err error) error
}

// ErrorHook defines operations to execute when errors occur
type ErrorHook interface {
	Execute(ctx context.Context, operation string, err error) error
}

// HookManager manages all hook operations
type HookManager interface {
	RegisterPreHook(hook PreHook)
	RegisterPostHook(hook PostHook)
	RegisterErrorHook(hook ErrorHook)

	ExecutePreHooks(ctx context.Context, operation string, args ...interface{}) (interface{}, error)
	ExecutePostHooks(ctx context.Context, operation string, result interface{}, err error) error
	ExecuteErrorHooks(ctx context.Context, operation string, err error) error

	ClearHooks()
}

// Config defines configuration options for decimal operations
type Config interface {
	GetMaxPrecision() uint32
	GetMaxExponent() int32
	GetMinExponent() int32
	GetDefaultRounding() string
	GetProviderName() string
	IsHooksEnabled() bool
	GetTimeout() int

	Validate() error
}
