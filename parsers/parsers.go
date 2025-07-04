// Package parsers provides a comprehensive set of parsing utilities for Go applications.
// It includes parsers for datetime, duration, and environment variables with modern
// patterns and best practices.
//
// The library is designed with the following principles:
// - Type safety with interfaces and generics where possible
// - Context support for cancellation and timeouts
// - Detailed error reporting with suggestions
// - Flexible configuration via options pattern
// - High performance with caching strategies
// - Comprehensive testing and benchmarking
//
// Example usage:
//
//	// DateTime parsing
//	parser := datetime.NewParser()
//	date, err := parser.Parse(ctx, "2023-01-15T10:30:45Z")
//
//	// Duration parsing with extended units
//	duration, err := duration.Parse("1w2d3h")
//
//	// Environment variable parsing
//	env := environment.NewParser(
//		environment.WithPrefix("APP"),
//		environment.WithDefaults(map[string]string{
//			"PORT": "8080",
//			"HOST": "localhost",
//		}),
//	)
//	port := env.GetInt("PORT")
package parsers

// Version information
const (
	Version = "1.0.0"
	Name    = "nexs-lib/parsers"
)

// Library provides access to all parsers
type Library struct {
	DateTime    DateTimeParser
	Duration    DurationParser
	Environment EnvironmentParser
}

// New creates a new Library instance with default parsers
func New(opts ...Option) *Library {
	config := DefaultConfig()
	for _, opt := range opts {
		opt.Apply(config)
	}

	return &Library{
		DateTime:    newDateTimeParser(config),
		Duration:    newDurationParser(config),
		Environment: newEnvironmentParser(config),
	}
}

// These functions will be implemented by importing the subpackages
// to avoid circular dependencies

func newDateTimeParser(config *Config) DateTimeParser {
	// This would typically import and use datetime.NewParser
	// For now, return nil to avoid circular dependency
	return nil
}

func newDurationParser(config *Config) DurationParser {
	// This would typically import and use duration.NewParser
	// For now, return nil to avoid circular dependency
	return nil
}

func newEnvironmentParser(config *Config) EnvironmentParser {
	// This would typically import and use environment.NewParser
	// For now, return nil to avoid circular dependency
	return nil
}
