// Package env provides environment variable parsing functionality.
package env

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

// ParsedEnv represents a parsed environment variable with metadata.
type ParsedEnv struct {
	Key      string
	Value    interface{}
	RawValue string
	Type     string
	Found    bool
}

// Parser implements environment variable parsing with type conversion.
type Parser struct {
	config *interfaces.ParserConfig
}

// NewParser creates a new environment variable parser.
func NewParser() *Parser {
	return &Parser{
		config: interfaces.DefaultConfig(),
	}
}

// NewParserWithConfig creates a new environment variable parser with custom configuration.
func NewParserWithConfig(config *interfaces.ParserConfig) *Parser {
	return &Parser{
		config: config,
	}
}

// Parse implements interfaces.Parser - parses env var from key=value format.
func (p *Parser) Parse(ctx context.Context, data []byte) (*ParsedEnv, error) {
	return p.ParseString(ctx, string(data))
}

// ParseString parses an environment variable from key=value format.
func (p *Parser) ParseString(ctx context.Context, input string) (*ParsedEnv, error) {
	if err := p.validateInput(input); err != nil {
		return nil, err
	}

	parts := strings.SplitN(input, "=", 2)
	if len(parts) != 2 {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: "input must be in key=value format",
		}
	}

	key := strings.TrimSpace(parts[0])
	value := parts[1] // Don't trim value as it might be intentional

	return &ParsedEnv{
		Key:      key,
		Value:    value,
		RawValue: value,
		Type:     "string",
		Found:    true,
	}, nil
}

// ParseEnvVar parses an environment variable by key name.
func (p *Parser) ParseEnvVar(key string) *ParsedEnv {
	value, found := os.LookupEnv(key)
	if !found {
		return &ParsedEnv{
			Key:      key,
			Value:    nil,
			RawValue: "",
			Type:     "undefined",
			Found:    false,
		}
	}

	return &ParsedEnv{
		Key:      key,
		Value:    value,
		RawValue: value,
		Type:     "string",
		Found:    true,
	}
}

// ParseInt32 parses an environment variable as int32.
func (p *Parser) ParseInt32(key string) *ParsedEnv {
	result := p.ParseEnvVar(key)
	if !result.Found {
		return result
	}

	if value, err := strconv.ParseInt(result.RawValue, 10, 32); err == nil {
		val := int32(value)
		result.Value = &val
		result.Type = "int32"
	} else {
		result.Value = nil
		result.Type = "error"
	}

	return result
}

// ParseInt parses an environment variable as int.
func (p *Parser) ParseInt(key string) *ParsedEnv {
	result := p.ParseEnvVar(key)
	if !result.Found {
		return result
	}

	if value, err := strconv.ParseInt(result.RawValue, 10, 0); err == nil {
		val := int(value)
		result.Value = &val
		result.Type = "int"
	} else {
		result.Value = nil
		result.Type = "error"
	}

	return result
}

// ParseInt64 parses an environment variable as int64.
func (p *Parser) ParseInt64(key string) *ParsedEnv {
	result := p.ParseEnvVar(key)
	if !result.Found {
		return result
	}

	if value, err := strconv.ParseInt(result.RawValue, 10, 64); err == nil {
		result.Value = &value
		result.Type = "int64"
	} else {
		result.Value = nil
		result.Type = "error"
	}

	return result
}

// ParseFloat64 parses an environment variable as float64.
func (p *Parser) ParseFloat64(key string) *ParsedEnv {
	result := p.ParseEnvVar(key)
	if !result.Found {
		return result
	}

	if value, err := strconv.ParseFloat(result.RawValue, 64); err == nil {
		result.Value = &value
		result.Type = "float64"
	} else {
		result.Value = nil
		result.Type = "error"
	}

	return result
}

// ParseBool parses an environment variable as bool.
func (p *Parser) ParseBool(key string) *ParsedEnv {
	result := p.ParseEnvVar(key)
	if !result.Found {
		return result
	}

	if value, err := strconv.ParseBool(result.RawValue); err == nil {
		result.Value = &value
		result.Type = "bool"
	} else {
		result.Value = nil
		result.Type = "error"
	}

	return result
}

// ParseDuration parses an environment variable as time.Duration.
func (p *Parser) ParseDuration(key string) *ParsedEnv {
	result := p.ParseEnvVar(key)
	if !result.Found {
		return result
	}

	if value, err := time.ParseDuration(result.RawValue); err == nil {
		result.Value = &value
		result.Type = "duration"
	} else {
		result.Value = nil
		result.Type = "error"
	}

	return result
}

// ParseSliceString parses an environment variable as []string with custom separator.
func (p *Parser) ParseSliceString(key, separator string) *ParsedEnv {
	result := p.ParseEnvVar(key)
	if !result.Found {
		return result
	}

	if separator == "" {
		separator = ","
	}

	values := strings.Split(result.RawValue, separator)
	// Trim whitespace from each value
	for i, v := range values {
		values[i] = strings.TrimSpace(v)
	}

	result.Value = &values
	result.Type = "[]string"
	return result
}

// validateInput validates the input string.
func (p *Parser) validateInput(input string) error {
	if len(input) == 0 {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "input environment variable string is empty",
		}
	}

	if p.config.MaxSize > 0 && int64(len(input)) > p.config.MaxSize {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSize,
			Message: fmt.Sprintf("env var string length %d exceeds maximum %d", len(input), p.config.MaxSize),
		}
	}

	return nil
}

// Formatter implements environment variable formatting.
type Formatter struct{}

// NewFormatter creates a new environment variable formatter.
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format implements interfaces.Formatter.
func (f *Formatter) Format(ctx context.Context, data *ParsedEnv) ([]byte, error) {
	if data == nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "data cannot be nil",
		}
	}

	formatted := fmt.Sprintf("%s=%s", data.Key, data.RawValue)
	return []byte(formatted), nil
}

// FormatString implements interfaces.Formatter.
func (f *Formatter) FormatString(ctx context.Context, data *ParsedEnv) (string, error) {
	result, err := f.Format(ctx, data)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// FormatWriter is not commonly used for env vars, returns error.
func (f *Formatter) FormatWriter(ctx context.Context, data *ParsedEnv, writer interface{}) error {
	return &interfaces.ParseError{
		Type:    interfaces.ErrorTypeValidation,
		Message: "FormatWriter not supported for env formatter",
	}
}

// Utility functions (backward compatibility with _old/parse)

// ParseEnvInt32 parses an environment variable as *int32.
func ParseEnvInt32(env string) *int32 {
	parser := NewParser()
	result := parser.ParseInt32(env)
	if result.Value != nil {
		return result.Value.(*int32)
	}
	return nil
}

// ParseEnvInt parses an environment variable as *int.
func ParseEnvInt(env string) *int {
	parser := NewParser()
	result := parser.ParseInt(env)
	if result.Value != nil {
		return result.Value.(*int)
	}
	return nil
}

// ParseEnvInt64 parses an environment variable as *int64.
func ParseEnvInt64(env string) *int64 {
	parser := NewParser()
	result := parser.ParseInt64(env)
	if result.Value != nil {
		return result.Value.(*int64)
	}
	return nil
}

// ParseEnvFloat64 parses an environment variable as *float64.
func ParseEnvFloat64(env string) *float64 {
	parser := NewParser()
	result := parser.ParseFloat64(env)
	if result.Value != nil {
		return result.Value.(*float64)
	}
	return nil
}

// ParseEnvDuration parses an environment variable as *time.Duration.
func ParseEnvDuration(env string) *time.Duration {
	parser := NewParser()
	result := parser.ParseDuration(env)
	if result.Value != nil {
		return result.Value.(*time.Duration)
	}
	return nil
}

// ParseEnvBool parses an environment variable as *bool.
func ParseEnvBool(env string) *bool {
	parser := NewParser()
	result := parser.ParseBool(env)
	if result.Value != nil {
		return result.Value.(*bool)
	}
	return nil
}

// ParseEnvSliceString parses an environment variable as *[]string.
func ParseEnvSliceString(env, separation string) *[]string {
	parser := NewParser()
	result := parser.ParseSliceString(env, separation)
	if result.Value != nil {
		return result.Value.(*[]string)
	}
	return nil
}

// GetEnv gets an environment variable with a default value.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvInt gets an environment variable as int with a default value.
func GetEnvInt(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetEnvBool gets an environment variable as bool with a default value.
func GetEnvBool(key string, defaultValue bool) bool {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// SetEnv sets an environment variable with type conversion.
func SetEnv(key string, value interface{}) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case int, int8, int16, int32, int64:
		strValue = fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		strValue = fmt.Sprintf("%d", v)
	case float32, float64:
		strValue = fmt.Sprintf("%g", v)
	case bool:
		strValue = strconv.FormatBool(v)
	case time.Duration:
		strValue = v.String()
	default:
		strValue = fmt.Sprintf("%v", v)
	}

	return os.Setenv(key, strValue)
}
