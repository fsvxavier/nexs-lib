// Package json provides JSON parsing functionality with backward compatibility.
// This file contains compatibility functions that maintain the original signatures
// from the _old/parse package.
package json

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Compatibility layer for _old/parse package

// ParseJSONToTypeCompat is a compatibility function that maintains the original signature
// from the _old/parse package while using the new implementation
func ParseJSONToTypeCompat[T any](data interface{}) (T, error) {
	return ParseJSONToType[T](data)
}

// Advanced parsing functionality

// AdvancedParser provides enhanced JSON parsing capabilities
type AdvancedParser struct {
	*Parser
	allowComments       bool
	allowTrailingCommas bool
	strictNumbers       bool
	customUnmarshalers  map[reflect.Type]func([]byte, interface{}) error
}

// NewAdvancedParser creates a new advanced JSON parser
func NewAdvancedParser() *AdvancedParser {
	return &AdvancedParser{
		Parser:              NewParser(),
		allowComments:       false,
		allowTrailingCommas: false,
		strictNumbers:       false,
		customUnmarshalers:  make(map[reflect.Type]func([]byte, interface{}) error),
	}
}

// WithComments enables/disables comment parsing
func (ap *AdvancedParser) WithComments(allow bool) *AdvancedParser {
	ap.allowComments = allow
	return ap
}

// WithTrailingCommas enables/disables trailing comma parsing
func (ap *AdvancedParser) WithTrailingCommas(allow bool) *AdvancedParser {
	ap.allowTrailingCommas = allow
	return ap
}

// WithStrictNumbers enables/disables strict number parsing
func (ap *AdvancedParser) WithStrictNumbers(strict bool) *AdvancedParser {
	ap.strictNumbers = strict
	return ap
}

// RegisterCustomUnmarshaler registers a custom unmarshaler for a specific type
func (ap *AdvancedParser) RegisterCustomUnmarshaler(t reflect.Type, unmarshaler func([]byte, interface{}) error) {
	ap.customUnmarshalers[t] = unmarshaler
}

// ParseAdvanced parses JSON with advanced features
func (ap *AdvancedParser) ParseAdvanced(ctx context.Context, data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case string:
		return ap.parseStringAdvanced(ctx, v)
	case []byte:
		return ap.parseBytesAdvanced(ctx, v)
	case io.Reader:
		return ap.parseReaderAdvanced(ctx, v)
	default:
		// Fallback to standard parsing
		return ap.ParseJSON(ctx, data)
	}
}

// parseStringAdvanced handles advanced string parsing
func (ap *AdvancedParser) parseStringAdvanced(ctx context.Context, input string) (interface{}, error) {
	if ap.allowComments {
		input = ap.removeComments(input)
	}

	if ap.allowTrailingCommas {
		input = ap.removeTrailingCommas(input)
	}

	return ap.ParseString(ctx, input)
}

// parseBytesAdvanced handles advanced bytes parsing
func (ap *AdvancedParser) parseBytesAdvanced(ctx context.Context, data []byte) (interface{}, error) {
	input := string(data)
	if ap.allowComments {
		input = ap.removeComments(input)
	}

	if ap.allowTrailingCommas {
		input = ap.removeTrailingCommas(input)
	}

	return ap.ParseString(ctx, input)
}

// parseReaderAdvanced handles advanced reader parsing
func (ap *AdvancedParser) parseReaderAdvanced(ctx context.Context, reader io.Reader) (interface{}, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %v", err)
	}

	return ap.parseBytesAdvanced(ctx, data)
}

// removeComments removes JSON comments (// and /* */)
func (ap *AdvancedParser) removeComments(input string) string {
	var result strings.Builder
	runes := []rune(input)
	length := len(runes)

	inString := false
	inLineComment := false
	inBlockComment := false
	escaped := false

	for i := 0; i < length; i++ {
		char := runes[i]

		if escaped {
			result.WriteRune(char)
			escaped = false
			continue
		}

		if inString {
			if char == '\\' {
				escaped = true
				result.WriteRune(char)
				continue
			}
			if char == '"' {
				inString = false
			}
			result.WriteRune(char)
			continue
		}

		if inLineComment {
			if char == '\n' {
				inLineComment = false
				result.WriteRune(char)
			}
			continue
		}

		if inBlockComment {
			if char == '*' && i+1 < length && runes[i+1] == '/' {
				inBlockComment = false
				i++ // Skip the '/'
			}
			continue
		}

		// Check for start of comments
		if char == '/' && i+1 < length {
			nextChar := runes[i+1]
			if nextChar == '/' {
				inLineComment = true
				i++ // Skip the second '/'
				continue
			}
			if nextChar == '*' {
				inBlockComment = true
				i++ // Skip the '*'
				continue
			}
		}

		if char == '"' {
			inString = true
		}

		result.WriteRune(char)
	}

	return result.String()
}

// removeTrailingCommas removes trailing commas from JSON
func (ap *AdvancedParser) removeTrailingCommas(input string) string {
	// More robust implementation
	var result strings.Builder
	runes := []rune(input)
	length := len(runes)

	inString := false
	escaped := false

	for i := 0; i < length; i++ {
		char := runes[i]

		if escaped {
			result.WriteRune(char)
			escaped = false
			continue
		}

		if inString {
			if char == '\\' {
				escaped = true
			} else if char == '"' {
				inString = false
			}
			result.WriteRune(char)
			continue
		}

		if char == '"' {
			inString = true
			result.WriteRune(char)
			continue
		}

		// Check for trailing comma
		if char == ',' {
			// Look ahead to see if we have a closing bracket/brace
			j := i + 1
			for j < length && (runes[j] == ' ' || runes[j] == '\t' || runes[j] == '\n' || runes[j] == '\r') {
				j++
			}

			if j < length && (runes[j] == '}' || runes[j] == ']') {
				// Skip the trailing comma
				continue
			}
		}

		result.WriteRune(char)
	}

	return result.String()
} // Additional format support

// ParseJSONL parses JSON Lines format (one JSON object per line)
func ParseJSONL(input string) ([]interface{}, error) {
	lines := strings.Split(input, "\n")
	var results []interface{}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var result interface{}
		if err := jsonInstance.UnmarshalFromString(line, &result); err != nil {
			return nil, fmt.Errorf("error parsing line %d: %v", i+1, err)
		}

		results = append(results, result)
	}

	return results, nil
}

// ParseNDJSON parses Newline Delimited JSON format
func ParseNDJSON(input string) ([]interface{}, error) {
	return ParseJSONL(input) // Same as JSONL
}

// ParseJSON5 provides basic JSON5 parsing support
func ParseJSON5(input string) (interface{}, error) {
	parser := NewAdvancedParser().WithComments(true).WithTrailingCommas(true)

	// Basic JSON5 preprocessing - handle unquoted keys
	processed := preprocessJSON5(input)

	return parser.parseStringAdvanced(context.Background(), processed)
}

// preprocessJSON5 does basic JSON5 to JSON conversion
func preprocessJSON5(input string) string {
	var result strings.Builder
	runes := []rune(input)
	length := len(runes)

	inString := false
	escaped := false

	for i := 0; i < length; i++ {
		char := runes[i]

		if escaped {
			result.WriteRune(char)
			escaped = false
			continue
		}

		if inString {
			if char == '\\' {
				escaped = true
			} else if char == '"' {
				inString = false
			}
			result.WriteRune(char)
			continue
		}

		if char == '"' {
			inString = true
			result.WriteRune(char)
			continue
		}

		// Handle unquoted keys (basic implementation)
		if char == ':' {
			// Look backwards to find the key
			j := i - 1
			for j >= 0 && (runes[j] == ' ' || runes[j] == '\t') {
				j--
			}

			if j >= 0 && runes[j] != '"' && runes[j] != '}' && runes[j] != ']' {
				// Found unquoted key, need to quote it
				keyEnd := j + 1

				// Find key start
				keyStart := j
				for keyStart >= 0 && isValidKeyChar(runes[keyStart]) {
					keyStart--
				}
				keyStart++

				if keyStart <= keyEnd-1 {
					// Extract and quote the key
					key := string(runes[keyStart:keyEnd])

					// Remove the unquoted key from result
					resultStr := result.String()
					prefixEnd := len([]rune(resultStr)) - (keyEnd - keyStart)
					if prefixEnd >= 0 {
						result.Reset()
						result.WriteString(string([]rune(resultStr)[:prefixEnd]))
						result.WriteString(`"` + key + `"`)
					}
				}
			}
		}

		result.WriteRune(char)
	}

	return result.String()
}

// isValidKeyChar checks if a character is valid in an unquoted JSON5 key
func isValidKeyChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$'
}

// StreamParser provides streaming JSON parsing
type StreamParser struct {
	decoder *json.Decoder
}

// NewStreamParser creates a new streaming JSON parser
func NewStreamParser(reader io.Reader) *StreamParser {
	return &StreamParser{
		decoder: json.NewDecoder(reader),
	}
}

// ParseNext parses the next JSON object from the stream
func (sp *StreamParser) ParseNext(result interface{}) error {
	return sp.decoder.Decode(result)
}

// HasMore checks if there are more objects to parse
func (sp *StreamParser) HasMore() bool {
	return sp.decoder.More()
}

// Utility functions for different JSON formats

// IsValidJSONL validates JSON Lines format
func IsValidJSONL(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !jsonInstance.Valid([]byte(line)) {
			return false
		}
	}
	return true
}

// ConvertToJSONL converts a slice of objects to JSON Lines format
func ConvertToJSONL(objects []interface{}) (string, error) {
	var lines []string
	for _, obj := range objects {
		jsonBytes, err := jsonInstance.Marshal(obj)
		if err != nil {
			return "", fmt.Errorf("failed to marshal object: %v", err)
		}
		lines = append(lines, string(jsonBytes))
	}
	return strings.Join(lines, "\n"), nil
}

// MergeJSON merges multiple JSON objects into one
func MergeJSON(objects ...interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, obj := range objects {
		var objMap map[string]interface{}

		// Convert to map if not already
		switch v := obj.(type) {
		case map[string]interface{}:
			objMap = v
		case string:
			if err := jsonInstance.UnmarshalFromString(v, &objMap); err != nil {
				return nil, fmt.Errorf("failed to parse JSON string: %v", err)
			}
		case []byte:
			if err := jsonInstance.Unmarshal(v, &objMap); err != nil {
				return nil, fmt.Errorf("failed to parse JSON bytes: %v", err)
			}
		default:
			// Try to convert via marshal/unmarshal
			jsonBytes, err := jsonInstance.Marshal(obj)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal object: %v", err)
			}
			if err := jsonInstance.Unmarshal(jsonBytes, &objMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal object: %v", err)
			}
		}

		// Merge into result
		for key, value := range objMap {
			result[key] = value
		}
	}

	return result, nil
}

// ExtractPath extracts a value from JSON using a path (e.g., "user.profile.name")
func ExtractPath(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("path not found: %s", path)
			}
		default:
			return nil, fmt.Errorf("cannot traverse path %s: not an object", path)
		}
	}

	return current, nil
}
