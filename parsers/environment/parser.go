package environment

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers"
)

// Parser implements the EnvironmentParser interface
type Parser struct {
	prefix   string
	defaults map[string]string
	required []string
	cache    map[string]interface{}
}

// NewParser creates a new environment parser
func NewParser(opts ...Option) *Parser {
	p := &Parser{
		defaults: make(map[string]string),
		cache:    make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt.Apply(p)
	}

	return p
}

// Option represents a configuration option for the environment parser
type Option interface {
	Apply(*Parser)
}

// OptionFunc is a function that implements Option
type OptionFunc func(*Parser)

func (f OptionFunc) Apply(p *Parser) {
	f(p)
}

// WithPrefix sets a prefix for all environment variable lookups
func WithPrefix(prefix string) Option {
	return OptionFunc(func(p *Parser) {
		p.prefix = prefix
	})
}

// WithDefaults sets default values for environment variables
func WithDefaults(defaults map[string]string) Option {
	return OptionFunc(func(p *Parser) {
		for k, v := range defaults {
			p.defaults[k] = v
		}
	})
}

// WithRequired marks environment variables as required
func WithRequired(keys ...string) Option {
	return OptionFunc(func(p *Parser) {
		p.required = append(p.required, keys...)
	})
}

// GetString retrieves a string value from environment variables
func (p *Parser) GetString(key string, defaultValue ...string) string {
	value := p.getValue(key)
	if value != "" {
		return value
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

// GetInt retrieves an integer value from environment variables
func (p *Parser) GetInt(key string, defaultValue ...int) int {
	value := p.getValue(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetInt32 retrieves an int32 value from environment variables
func (p *Parser) GetInt32(key string, defaultValue ...int32) int32 {
	value := p.getValue(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	if parsed, err := strconv.ParseInt(value, 10, 32); err == nil {
		return int32(parsed)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetInt64 retrieves an int64 value from environment variables
func (p *Parser) GetInt64(key string, defaultValue ...int64) int64 {
	value := p.getValue(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
		return parsed
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetFloat64 retrieves a float64 value from environment variables
func (p *Parser) GetFloat64(key string, defaultValue ...float64) float64 {
	value := p.getValue(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		return parsed
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetBool retrieves a boolean value from environment variables
func (p *Parser) GetBool(key string, defaultValue ...bool) bool {
	value := p.getValue(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}

	// Handle common boolean representations
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on", "enabled":
		return true
	case "false", "0", "no", "off", "disabled":
		return false
	default:
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// GetDuration retrieves a duration value from environment variables
func (p *Parser) GetDuration(key string, defaultValue ...time.Duration) time.Duration {
	value := p.getValue(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	if parsed, err := time.ParseDuration(value); err == nil {
		return parsed
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetSlice retrieves a slice of strings from environment variables
func (p *Parser) GetSlice(key, separator string, defaultValue ...[]string) []string {
	value := p.getValue(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	if separator == "" {
		separator = ","
	}

	parts := strings.Split(value, separator)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return result
}

// GetMap retrieves a map from environment variables
func (p *Parser) GetMap(key, separator, kvSeparator string) map[string]string {
	value := p.getValue(key)
	if value == "" {
		return make(map[string]string)
	}

	if separator == "" {
		separator = ","
	}
	if kvSeparator == "" {
		kvSeparator = "="
	}

	result := make(map[string]string)
	pairs := strings.Split(value, separator)

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, kvSeparator, 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				result[key] = value
			}
		}
	}

	return result
}

// GetValue retrieves a raw value using a custom parser function
func (p *Parser) GetValue(key string, parser func(string) (interface{}, error), defaultValue ...interface{}) interface{} {
	value := p.getValue(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	if parsed, err := parser(value); err == nil {
		return parsed
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return nil
}

// MustGetValue retrieves a value using a custom parser function and panics on error
func (p *Parser) MustGetValue(key string, parser func(string) (interface{}, error)) interface{} {
	value := p.getValue(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable '%s' not found", p.getFullKey(key)))
	}

	parsed, err := parser(value)
	if err != nil {
		panic(fmt.Sprintf("failed to parse environment variable '%s': %v", p.getFullKey(key), err))
	}

	return parsed
}

// Validate checks that all required environment variables are present and valid
func (p *Parser) Validate() error {
	multiErr := parsers.NewMultiError()

	for _, key := range p.required {
		if !p.IsSet(key) {
			multiErr.Add(parsers.NewNotFoundError(p.getFullKey(key)))
		}
	}

	if multiErr.HasErrors() {
		return multiErr
	}

	return nil
}

// IsSet checks if an environment variable is set (not empty)
func (p *Parser) IsSet(key string) bool {
	value := p.getValue(key)
	return value != ""
}

// WithPrefix returns a new parser with the given prefix added
func (p *Parser) WithPrefix(prefix string) parsers.EnvironmentParser {
	newPrefix := prefix
	if p.prefix != "" {
		newPrefix = p.prefix + "_" + prefix
	}

	return &Parser{
		prefix:   newPrefix,
		defaults: p.defaults,
		required: p.required,
		cache:    make(map[string]interface{}),
	}
}

// WithDefaults returns a new parser with additional default values
func (p *Parser) WithDefaults(defaults map[string]string) parsers.EnvironmentParser {
	newDefaults := make(map[string]string)

	// Copy existing defaults
	for k, v := range p.defaults {
		newDefaults[k] = v
	}

	// Add new defaults
	for k, v := range defaults {
		newDefaults[k] = v
	}

	return &Parser{
		prefix:   p.prefix,
		defaults: newDefaults,
		required: p.required,
		cache:    make(map[string]interface{}),
	}
}

// getValue retrieves the raw string value for a key
func (p *Parser) getValue(key string) string {
	fullKey := p.getFullKey(key)

	// Check cache first
	if cached, exists := p.cache[fullKey]; exists {
		if str, ok := cached.(string); ok {
			return str
		}
	}

	// Get from environment
	value := os.Getenv(fullKey)

	// Fall back to default if not found
	if value == "" {
		if defaultValue, exists := p.defaults[key]; exists {
			value = defaultValue
		}
	}

	// Cache the result
	p.cache[fullKey] = value

	return value
}

// getFullKey returns the full environment variable key with prefix
func (p *Parser) getFullKey(key string) string {
	if p.prefix == "" {
		return key
	}
	return p.prefix + "_" + key
}

// Advanced parsing methods

// GetStringPtr retrieves a string pointer from environment variables
func (p *Parser) GetStringPtr(key string) *string {
	value := p.getValue(key)
	if value == "" {
		return nil
	}
	return &value
}

// GetIntPtr retrieves an integer pointer from environment variables
func (p *Parser) GetIntPtr(key string) *int {
	value := p.getValue(key)
	if value == "" {
		return nil
	}

	if parsed, err := strconv.Atoi(value); err == nil {
		return &parsed
	}
	return nil
}

// GetInt32Ptr retrieves an int32 pointer from environment variables
func (p *Parser) GetInt32Ptr(key string) *int32 {
	value := p.getValue(key)
	if value == "" {
		return nil
	}

	if parsed, err := strconv.ParseInt(value, 10, 32); err == nil {
		val := int32(parsed)
		return &val
	}
	return nil
}

// GetInt64Ptr retrieves an int64 pointer from environment variables
func (p *Parser) GetInt64Ptr(key string) *int64 {
	value := p.getValue(key)
	if value == "" {
		return nil
	}

	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
		return &parsed
	}
	return nil
}

// GetBoolPtr retrieves a boolean pointer from environment variables
func (p *Parser) GetBoolPtr(key string) *bool {
	value := p.getValue(key)
	if value == "" {
		return nil
	}

	parsed := p.GetBool(key)
	return &parsed
}

// GetDurationPtr retrieves a duration pointer from environment variables
func (p *Parser) GetDurationPtr(key string) *time.Duration {
	value := p.getValue(key)
	if value == "" {
		return nil
	}

	if parsed, err := time.ParseDuration(value); err == nil {
		return &parsed
	}
	return nil
}

// Config represents environment configuration
type Config struct {
	parser *Parser
}

// NewConfig creates a new environment configuration
func NewConfig(opts ...Option) *Config {
	return &Config{
		parser: NewParser(opts...),
	}
}

// Bind binds environment variables to a struct
func (c *Config) Bind(target interface{}) error {
	// This would use reflection to bind environment variables to struct fields
	// For now, we'll return a not implemented error
	return fmt.Errorf("bind method not implemented yet")
}

// Package-level convenience functions

// defaultParser is the default environment parser instance
var defaultParser = NewParser()

// GetString retrieves a string value using the default parser
func GetString(key string, defaultValue ...string) string {
	return defaultParser.GetString(key, defaultValue...)
}

// GetInt retrieves an integer value using the default parser
func GetInt(key string, defaultValue ...int) int {
	return defaultParser.GetInt(key, defaultValue...)
}

// GetInt32 retrieves an int32 value using the default parser
func GetInt32(key string, defaultValue ...int32) int32 {
	return defaultParser.GetInt32(key, defaultValue...)
}

// GetInt64 retrieves an int64 value using the default parser
func GetInt64(key string, defaultValue ...int64) int64 {
	return defaultParser.GetInt64(key, defaultValue...)
}

// GetFloat64 retrieves a float64 value using the default parser
func GetFloat64(key string, defaultValue ...float64) float64 {
	return defaultParser.GetFloat64(key, defaultValue...)
}

// GetBool retrieves a boolean value using the default parser
func GetBool(key string, defaultValue ...bool) bool {
	return defaultParser.GetBool(key, defaultValue...)
}

// GetDuration retrieves a duration value using the default parser
func GetDuration(key string, defaultValue ...time.Duration) time.Duration {
	return defaultParser.GetDuration(key, defaultValue...)
}

// GetSlice retrieves a slice of strings using the default parser
func GetSlice(key, separator string, defaultValue ...[]string) []string {
	return defaultParser.GetSlice(key, separator, defaultValue...)
}

// GetMap retrieves a map using the default parser
func GetMap(key, separator, kvSeparator string) map[string]string {
	return defaultParser.GetMap(key, separator, kvSeparator)
}

// IsSet checks if an environment variable is set using the default parser
func IsSet(key string) bool {
	return defaultParser.IsSet(key)
}
