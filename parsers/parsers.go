// Package parsers provides a comprehensive set of parsing utilities.
package parsers

import (
	"context"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers/datetime"
	"github.com/fsvxavier/nexs-lib/parsers/duration"
	"github.com/fsvxavier/nexs-lib/parsers/env"
	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
	"github.com/fsvxavier/nexs-lib/parsers/json"
	"github.com/fsvxavier/nexs-lib/parsers/url"
)

// Version of the parsers package.
const Version = "1.0.0"

// Factory provides a centralized way to create parsers.
type Factory struct {
	config *interfaces.ParserConfig
}

// NewFactory creates a new parser factory with default configuration.
func NewFactory() *Factory {
	return &Factory{
		config: interfaces.DefaultConfig(),
	}
}

// NewFactoryWithConfig creates a new parser factory with custom configuration.
func NewFactoryWithConfig(config *interfaces.ParserConfig) *Factory {
	return &Factory{
		config: config,
	}
}

// JSON creates a new JSON parser.
func (f *Factory) JSON() *json.Parser {
	return json.NewParserWithConfig(f.config)
}

// URL creates a new URL parser.
func (f *Factory) URL() *url.Parser {
	return url.NewParserWithConfig(f.config)
}

// Datetime creates a new datetime parser.
func (f *Factory) Datetime() *datetime.Parser {
	return datetime.NewParserWithConfig(f.config)
}

// Duration creates a new duration parser.
func (f *Factory) Duration() *duration.Parser {
	return duration.NewParserWithConfig(f.config)
}

// Env creates a new environment variable parser.
func (f *Factory) Env() *env.Parser {
	return env.NewParserWithConfig(f.config)
}

// Manager provides high-level parsing operations with context and metadata.
type Manager struct {
	factory *Factory
}

// NewManager creates a new parser manager.
func NewManager() *Manager {
	return &Manager{
		factory: NewFactory(),
	}
}

// NewManagerWithConfig creates a new parser manager with custom configuration.
func NewManagerWithConfig(config *interfaces.ParserConfig) *Manager {
	return &Manager{
		factory: NewFactoryWithConfig(config),
	}
}

// ParseJSON parses JSON data and returns result with metadata.
func (m *Manager) ParseJSON(ctx context.Context, data []byte) (*interfaces.Result[interface{}], error) {
	parser := m.factory.JSON()

	startTime := time.Now()
	result, err := parser.ParseBytes(ctx, data)
	duration := time.Since(startTime)

	if err != nil {
		return nil, err
	}

	metadata := &interfaces.Metadata{
		ParsedAt:       time.Now(),
		Duration:       duration,
		BytesProcessed: int64(len(data)),
		ParserType:     "json",
		Version:        Version,
	}

	return &interfaces.Result[interface{}]{
		Data:     &result,
		Metadata: metadata,
	}, nil
}

// ParseURL parses URL and returns result with metadata.
func (m *Manager) ParseURL(ctx context.Context, urlStr string) (*interfaces.Result[*url.ParsedURL], error) {
	parser := m.factory.URL()

	startTime := time.Now()
	result, err := parser.ParseString(ctx, urlStr)
	duration := time.Since(startTime)

	if err != nil {
		return nil, err
	}

	metadata := &interfaces.Metadata{
		ParsedAt:       time.Now(),
		Duration:       duration,
		BytesProcessed: int64(len(urlStr)),
		ParserType:     "url",
		Version:        Version,
	}

	return &interfaces.Result[*url.ParsedURL]{
		Data:     &result,
		Metadata: metadata,
	}, nil
}

// Validator provides validation utilities for parsed data.
type Validator struct{}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateJSON validates JSON data without parsing.
func (v *Validator) ValidateJSON(data []byte) error {
	return json.ValidateJSONString(string(data))
}

// ValidateJSONString validates JSON string without parsing.
func (v *Validator) ValidateJSONString(input string) error {
	return json.ValidateJSONString(input)
}

// ValidateURL validates URL string without parsing.
func (v *Validator) ValidateURL(input string) error {
	if !url.IsValidURL(input) {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "invalid URL format",
		}
	}
	return nil
}

// Transformer provides data transformation utilities.
type Transformer struct{}

// NewTransformer creates a new transformer.
func NewTransformer() *Transformer {
	return &Transformer{}
}

// CompactJSON compacts JSON by removing whitespace.
func (t *Transformer) CompactJSON(input string) (string, error) {
	return json.CompactJSON(input)
}

// PrettyJSON formats JSON with indentation.
func (t *Transformer) PrettyJSON(input string) (string, error) {
	return json.PrettyJSON(input, "  ")
}

// Convenience functions for common operations.

// ParseJSON parses JSON data from interface{}.
func ParseJSON(data interface{}) (interface{}, error) {
	return json.ParseJSON(data)
}

// ParseJSONString parses JSON string.
func ParseJSONString(input string) (interface{}, error) {
	return json.ParseJSONString(input)
}

// ParseJSONBytes parses JSON bytes.
func ParseJSONBytes(data []byte) (interface{}, error) {
	return json.ParseJSONBytes(data)
}

// ParseURLString parses URL string.
func ParseURLString(urlStr string) (*url.ParsedURL, error) {
	parser := url.NewParser()
	return parser.ParseString(context.Background(), urlStr)
}

// ValidateJSONData validates JSON data.
func ValidateJSONData(data interface{}) error {
	return json.ValidateJSONData(data)
}

// ValidateJSONStr validates JSON string.
func ValidateJSONStr(input string) error {
	return json.ValidateJSONString(input)
}

// ValidateURLStr validates URL string.
func ValidateURLStr(input string) error {
	validator := NewValidator()
	return validator.ValidateURL(input)
}
