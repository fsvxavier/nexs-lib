// Package json provides JSON parsing functionality.
package json

import (
	"context"
	"fmt"

	jsoniter "github.com/json-iterator/go"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

var jsonInstance = jsoniter.ConfigCompatibleWithStandardLibrary

// Parser implements JSON parsing functionality.
type Parser struct {
	config *interfaces.ParserConfig
}

// NewParser creates a new JSON parser with default configuration.
func NewParser() *Parser {
	return &Parser{
		config: interfaces.DefaultConfig(),
	}
}

// NewParserWithConfig creates a new JSON parser with custom configuration.
func NewParserWithConfig(config *interfaces.ParserConfig) *Parser {
	return &Parser{
		config: config,
	}
}

// ParseJSONToType parses JSON data into the specified type
func ParseJSONToType[T any](data interface{}) (T, error) {
	var result T

	// Convert the data to JSON bytes first
	jsonBytes, err := jsonInstance.Marshal(data)
	if err != nil {
		return result, err
	}

	// Unmarshal into the target type
	err = jsonInstance.Unmarshal(jsonBytes, &result)
	return result, err
}

// ParseJSON parses JSON data from interface{} to interface{}
func (p *Parser) ParseJSON(ctx context.Context, data interface{}) (interface{}, error) {
	if data == nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "input data is nil",
		}
	}

	// If data is already a map or slice, return as is
	switch v := data.(type) {
	case map[string]interface{}, []interface{}:
		return v, nil
	case string:
		var result interface{}
		if err := jsonInstance.UnmarshalFromString(v, &result); err != nil {
			return nil, &interfaces.ParseError{
				Type:    interfaces.ErrorTypeSyntax,
				Message: fmt.Sprintf("failed to parse JSON string: %v", err),
				Cause:   err,
			}
		}
		return result, nil
	case []byte:
		var result interface{}
		if err := jsonInstance.Unmarshal(v, &result); err != nil {
			return nil, &interfaces.ParseError{
				Type:    interfaces.ErrorTypeSyntax,
				Message: fmt.Sprintf("failed to parse JSON bytes: %v", err),
				Cause:   err,
			}
		}
		return result, nil
	default:
		// Try to marshal and unmarshal for type conversion
		jsonBytes, err := jsonInstance.Marshal(data)
		if err != nil {
			return nil, &interfaces.ParseError{
				Type:    interfaces.ErrorTypeSyntax,
				Message: fmt.Sprintf("failed to marshal data: %v", err),
				Cause:   err,
			}
		}

		var result interface{}
		if err := jsonInstance.Unmarshal(jsonBytes, &result); err != nil {
			return nil, &interfaces.ParseError{
				Type:    interfaces.ErrorTypeSyntax,
				Message: fmt.Sprintf("failed to unmarshal data: %v", err),
				Cause:   err,
			}
		}
		return result, nil
	}
}

// ParseString parses JSON string into interface{}
func (p *Parser) ParseString(ctx context.Context, input string) (interface{}, error) {
	if input == "" {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "input string is empty",
		}
	}

	var result interface{}
	if err := jsonInstance.UnmarshalFromString(input, &result); err != nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: fmt.Sprintf("failed to parse JSON string: %v", err),
			Cause:   err,
		}
	}
	return result, nil
}

// ParseBytes parses JSON bytes into interface{}
func (p *Parser) ParseBytes(ctx context.Context, data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "input data is empty",
		}
	}

	var result interface{}
	if err := jsonInstance.Unmarshal(data, &result); err != nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: fmt.Sprintf("failed to parse JSON bytes: %v", err),
			Cause:   err,
		}
	}
	return result, nil
}

// ValidateJSON validates if the given data is valid JSON
func (p *Parser) ValidateJSON(ctx context.Context, data interface{}) error {
	switch v := data.(type) {
	case string:
		if !jsonInstance.Valid([]byte(v)) {
			return &interfaces.ParseError{
				Type:    interfaces.ErrorTypeValidation,
				Message: "invalid JSON string",
			}
		}
		return nil
	case []byte:
		if !jsonInstance.Valid(v) {
			return &interfaces.ParseError{
				Type:    interfaces.ErrorTypeValidation,
				Message: "invalid JSON bytes",
			}
		}
		return nil
	case map[string]interface{}, []interface{}:
		// Already valid JSON structures
		return nil
	default:
		// Try to marshal to check validity
		_, err := jsonInstance.Marshal(data)
		if err != nil {
			return &interfaces.ParseError{
				Type:    interfaces.ErrorTypeValidation,
				Message: fmt.Sprintf("invalid JSON data: %v", err),
				Cause:   err,
			}
		}
		return nil
	}
}

// Formatter implements JSON formatting functionality.
type Formatter struct {
	indent string
}

// NewFormatter creates a new JSON formatter.
func NewFormatter() *Formatter {
	return &Formatter{}
}

// NewFormatterWithIndent creates a new JSON formatter with custom indentation.
func NewFormatterWithIndent(indent string) *Formatter {
	return &Formatter{
		indent: indent,
	}
}

// Format formats data to JSON bytes.
func (f *Formatter) Format(ctx context.Context, data interface{}) ([]byte, error) {
	if data == nil {
		return []byte("null"), nil
	}

	if f.indent != "" {
		return jsonInstance.MarshalIndent(data, "", f.indent)
	}
	return jsonInstance.Marshal(data)
}

// FormatString formats data to JSON string.
func (f *Formatter) FormatString(ctx context.Context, data interface{}) (string, error) {
	bytes, err := f.Format(ctx, data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CompactJSON compacts JSON string by removing whitespace
func CompactJSON(jsonStr string) (string, error) {
	var compact interface{}
	if err := jsonInstance.UnmarshalFromString(jsonStr, &compact); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}

	compactBytes, err := jsonInstance.Marshal(compact)
	if err != nil {
		return "", fmt.Errorf("failed to compact JSON: %v", err)
	}

	return string(compactBytes), nil
}

// PrettyJSON formats JSON string with indentation
func PrettyJSON(jsonStr string, indent string) (string, error) {
	var data interface{}
	if err := jsonInstance.UnmarshalFromString(jsonStr, &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}

	// jsoniter only accepts spaces for indentation, convert tabs to spaces
	if indent == "\t" {
		indent = "  " // Convert tab to 2 spaces
	}

	prettyBytes, err := jsonInstance.MarshalIndent(data, "", indent)
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %v", err)
	}

	return string(prettyBytes), nil
} // Utility functions for convenience

// ParseJSON parses JSON data from interface{} to interface{}
func ParseJSON(data interface{}) (interface{}, error) {
	parser := NewParser()
	return parser.ParseJSON(context.Background(), data)
}

// ParseJSONString parses JSON string into interface{}
func ParseJSONString(input string) (interface{}, error) {
	parser := NewParser()
	return parser.ParseString(context.Background(), input)
}

// ParseJSONBytes parses JSON bytes into interface{}
func ParseJSONBytes(data []byte) (interface{}, error) {
	parser := NewParser()
	return parser.ParseBytes(context.Background(), data)
}

// ValidateJSONData validates if the given data is valid JSON
func ValidateJSONData(data interface{}) error {
	parser := NewParser()
	return parser.ValidateJSON(context.Background(), data)
}

// ValidateJSONString validates if the given string is valid JSON
func ValidateJSONString(input string) error {
	if !jsonInstance.Valid([]byte(input)) {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "invalid JSON string",
		}
	}
	return nil
}
