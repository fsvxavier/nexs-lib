package parsers

import (
	"context"
	"time"
)

// Parser defines the common interface for all parsers
type Parser[T any] interface {
	Parse(ctx context.Context, input string) (T, error)
	ParseWithOptions(ctx context.Context, input string, opts ...Option) (T, error)
	MustParse(ctx context.Context, input string) T
	TryParse(ctx context.Context, input string) (T, bool)
}

// DateTimeParser defines the interface for datetime parsing
type DateTimeParser interface {
	Parser[time.Time]
	ParseInLocation(ctx context.Context, input string, loc *time.Location) (time.Time, error)
	ParseToUTC(ctx context.Context, input string) (time.Time, error)
	SetDefaultLocation(loc *time.Location)
	GetSupportedFormats() []string
}

// DurationParser defines the interface for duration parsing
type DurationParser interface {
	Parser[time.Duration]
	ParseExtended(ctx context.Context, input string) (time.Duration, error)
	GetSupportedUnits() map[string]time.Duration
}

// EnvironmentParser defines the interface for environment variable parsing
type EnvironmentParser interface {
	GetString(key string, defaultValue ...string) string
	GetInt(key string, defaultValue ...int) int
	GetInt32(key string, defaultValue ...int32) int32
	GetInt64(key string, defaultValue ...int64) int64
	GetFloat64(key string, defaultValue ...float64) float64
	GetBool(key string, defaultValue ...bool) bool
	GetDuration(key string, defaultValue ...time.Duration) time.Duration
	GetSlice(key, separator string, defaultValue ...[]string) []string
	GetMap(key, separator, kvSeparator string) map[string]string

	// Generic methods will be implemented as concrete methods
	GetValue(key string, parser func(string) (interface{}, error), defaultValue ...interface{}) interface{}
	MustGetValue(key string, parser func(string) (interface{}, error)) interface{}

	// Validation
	Validate() error
	IsSet(key string) bool

	// Configuration
	WithPrefix(prefix string) EnvironmentParser
	WithDefaults(defaults map[string]string) EnvironmentParser
}

// Option represents a configuration option for parsers
type Option interface {
	Apply(config *Config)
}

// Config holds configuration for parsers
type Config struct {
	DefaultLocation *time.Location
	StrictMode      bool
	CustomFormats   []string
	IgnoreCase      bool
	AllowPartial    bool
	Timezone        string
	DateOrder       DateOrder
	CustomUnits     map[string]time.Duration
}

// DateOrder represents the preferred date order when ambiguous
type DateOrder int

const (
	DateOrderMDY DateOrder = iota // Month-Day-Year (US style)
	DateOrderDMY                  // Day-Month-Year (European style)
	DateOrderYMD                  // Year-Month-Day (ISO style)
)

// OptionFunc is a function that implements Option
type OptionFunc func(*Config)

func (f OptionFunc) Apply(config *Config) {
	f(config)
}

// Common options
func WithLocation(loc *time.Location) Option {
	return OptionFunc(func(c *Config) {
		c.DefaultLocation = loc
	})
}

func WithStrictMode(strict bool) Option {
	return OptionFunc(func(c *Config) {
		c.StrictMode = strict
	})
}

func WithCustomFormats(formats ...string) Option {
	return OptionFunc(func(c *Config) {
		c.CustomFormats = append(c.CustomFormats, formats...)
	})
}

func WithIgnoreCase(ignore bool) Option {
	return OptionFunc(func(c *Config) {
		c.IgnoreCase = ignore
	})
}

func WithDateOrder(order DateOrder) Option {
	return OptionFunc(func(c *Config) {
		c.DateOrder = order
	})
}

func WithCustomUnits(units map[string]time.Duration) Option {
	return OptionFunc(func(c *Config) {
		if c.CustomUnits == nil {
			c.CustomUnits = make(map[string]time.Duration)
		}
		for k, v := range units {
			c.CustomUnits[k] = v
		}
	})
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultLocation: time.UTC,
		StrictMode:      false,
		IgnoreCase:      true,
		AllowPartial:    false,
		DateOrder:       DateOrderMDY,
		CustomUnits:     make(map[string]time.Duration),
	}
}
