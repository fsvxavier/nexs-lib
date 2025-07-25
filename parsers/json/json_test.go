package json

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Error("Expected parser to be created")
	}
	if parser.config == nil {
		t.Error("Expected parser config to be initialized")
	}
}

func TestNewParserWithConfig(t *testing.T) {
	config := &interfaces.ParserConfig{
		StrictMode: false,
	}
	parser := NewParserWithConfig(config)
	if parser == nil {
		t.Error("Expected parser to be created")
	}
	if parser.config != config {
		t.Error("Expected parser to use provided config")
	}
}

func TestParseJSONToType(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantErr  bool
		expected interface{}
	}{
		{
			name:     "map to map",
			input:    map[string]interface{}{"key": "value"},
			wantErr:  false,
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "slice to slice",
			input:    []interface{}{1, 2, 3},
			wantErr:  false,
			expected: []interface{}{1, 2, 3},
		},
		{
			name:    "invalid type",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJSONToType[interface{}](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONToType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected result to not be nil")
			}
		})
	}
}

func TestParser_ParseJSON(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "valid map",
			input:   map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "valid slice",
			input:   []interface{}{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "valid JSON string",
			input:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON string",
			input:   `{"key": value}`,
			wantErr: true,
		},
		{
			name:    "valid JSON bytes",
			input:   []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON bytes",
			input:   []byte(`{"key": value}`),
			wantErr: true,
		},
		{
			name:    "struct input",
			input:   struct{ Key string }{Key: "value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseJSON(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected result to not be nil")
			}
		})
	}
}

func TestParser_ParseString(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "valid JSON object",
			input:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "valid JSON array",
			input:   `[1, 2, 3]`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"key": value}`,
			wantErr: true,
		},
		{
			name:    "valid JSON string",
			input:   `"hello"`,
			wantErr: false,
		},
		{
			name:    "valid JSON number",
			input:   `42`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseString(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil && tt.input != "null" {
				t.Error("Expected result to not be nil")
			}
		})
	}
}

func TestParser_ParseBytes(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "empty bytes",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "valid JSON object",
			input:   []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "valid JSON array",
			input:   []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{"key": value}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseBytes(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected result to not be nil")
			}
		})
	}
}

func TestParser_ValidateJSON(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "valid JSON string",
			input:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON string",
			input:   `{"key": value}`,
			wantErr: true,
		},
		{
			name:    "valid JSON bytes",
			input:   []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON bytes",
			input:   []byte(`{"key": value}`),
			wantErr: true,
		},
		{
			name:    "valid map",
			input:   map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "valid slice",
			input:   []interface{}{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "valid struct",
			input:   struct{ Key string }{Key: "value"},
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateJSON(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()
	if formatter == nil {
		t.Error("Expected formatter to be created")
	}
	if formatter.indent != "" {
		t.Error("Expected default formatter to have empty indent")
	}
}

func TestNewFormatterWithIndent(t *testing.T) {
	indent := "  "
	formatter := NewFormatterWithIndent(indent)
	if formatter == nil {
		t.Error("Expected formatter to be created")
	}
	if formatter.indent != indent {
		t.Error("Expected formatter to use provided indent")
	}
}

func TestFormatter_Format(t *testing.T) {
	formatter := NewFormatter()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "map input",
			input:   map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "slice input",
			input:   []interface{}{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "string input",
			input:   "hello",
			wantErr: false,
		},
		{
			name:    "invalid input",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.Format(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected result to not be nil")
			}
		})
	}
}

func TestFormatter_FormatString(t *testing.T) {
	formatter := NewFormatter()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "map input",
			input:   map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "invalid input",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.FormatString(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" && tt.input != nil {
				t.Error("Expected result to not be empty")
			}
		})
	}
}

func TestCompactJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid JSON with spaces",
			input:   `{ "key" : "value" }`,
			wantErr: false,
		},
		{
			name:    "already compact JSON",
			input:   `{"key":"value"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"key": value}`,
			wantErr: true,
		},
		{
			name:    "array JSON",
			input:   `[ 1 , 2 , 3 ]`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompactJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompactJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("Expected result to not be empty")
			}
		})
	}
}

func TestPrettyJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		indent  string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			input:   `{"key":"value"}`,
			indent:  "  ",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"key": value}`,
			indent:  "  ",
			wantErr: true,
		},
		{
			name:    "array JSON",
			input:   `[1,2,3]`,
			indent:  "\t",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrettyJSON(tt.input, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("Expected result to not be empty")
			}
		})
	}
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("ParseJSON", func(t *testing.T) {
		result, err := ParseJSON(map[string]interface{}{"key": "value"})
		if err != nil {
			t.Errorf("ParseJSON() error = %v", err)
		}
		if result == nil {
			t.Error("Expected result to not be nil")
		}
	})

	t.Run("ParseJSONString", func(t *testing.T) {
		result, err := ParseJSONString(`{"key": "value"}`)
		if err != nil {
			t.Errorf("ParseJSONString() error = %v", err)
		}
		if result == nil {
			t.Error("Expected result to not be nil")
		}
	})

	t.Run("ParseJSONBytes", func(t *testing.T) {
		result, err := ParseJSONBytes([]byte(`{"key": "value"}`))
		if err != nil {
			t.Errorf("ParseJSONBytes() error = %v", err)
		}
		if result == nil {
			t.Error("Expected result to not be nil")
		}
	})

	t.Run("ValidateJSONData", func(t *testing.T) {
		err := ValidateJSONData(map[string]interface{}{"key": "value"})
		if err != nil {
			t.Errorf("ValidateJSONData() error = %v", err)
		}
	})

	t.Run("ValidateJSONString", func(t *testing.T) {
		err := ValidateJSONString(`{"key": "value"}`)
		if err != nil {
			t.Errorf("ValidateJSONString() error = %v", err)
		}

		err = ValidateJSONString(`{"key": value}`)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestParseJSON_EdgeCases(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	// Test case: struct that needs marshal/unmarshal conversion
	type CustomStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:  "Custom struct conversion",
			input: CustomStruct{Name: "test", Value: 42},
			expected: map[string]interface{}{
				"name":  "test",
				"value": float64(42), // JSON numbers are float64
			},
			wantErr: false,
		},
		{
			name:     "Channel type (unmarshalable)",
			input:    make(chan int),
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Function type (unmarshalable)",
			input:    func() {},
			expected: nil,
			wantErr:  true,
		},
		{
			name:  "Pointer to struct",
			input: &CustomStruct{Name: "pointer", Value: 123},
			expected: map[string]interface{}{
				"name":  "pointer",
				"value": float64(123),
			},
			wantErr: false,
		},
		{
			name:     "Complex number",
			input:    complex(1, 2),
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseJSON(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// For complex comparisons, check type and key-value pairs
			if resultMap, ok := result.(map[string]interface{}); ok {
				expectedMap := tt.expected.(map[string]interface{})
				for key, expectedValue := range expectedMap {
					if actualValue, exists := resultMap[key]; !exists || actualValue != expectedValue {
						t.Errorf("Expected %v, got %v for key %s", expectedValue, actualValue, key)
					}
				}
			} else if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFormatter_Format_EdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		formatter *Formatter
		input     interface{}
		wantErr   bool
		checkFunc func([]byte) bool
	}{
		{
			name:      "Format nil with regular formatter",
			formatter: NewFormatter(),
			input:     nil,
			wantErr:   false,
			checkFunc: func(b []byte) bool { return string(b) == "null" },
		},
		{
			name:      "Format nil with indent formatter",
			formatter: NewFormatterWithIndent("  "),
			input:     nil,
			wantErr:   false,
			checkFunc: func(b []byte) bool { return string(b) == "null" },
		},
		{
			name:      "Format complex data with indent",
			formatter: NewFormatterWithIndent("  "),
			input:     map[string]interface{}{"nested": map[string]interface{}{"key": "value"}},
			wantErr:   false,
			checkFunc: func(b []byte) bool { return len(b) > 10 }, // Just check it's not empty
		},
		{
			name:      "Format unmarshalable data",
			formatter: NewFormatter(),
			input:     make(chan int),
			wantErr:   true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.formatter.Format(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil && !tt.checkFunc(result) {
				t.Errorf("Check function failed for result: %s", string(result))
			}
		})
	}
}

func TestCompactJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Invalid JSON",
			input:    `{"key": value}`, // missing quotes around value
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Empty object",
			input:    `{}`,
			expected: `{}`,
			wantErr:  false,
		},
		{
			name:     "Empty array",
			input:    `[]`,
			expected: `[]`,
			wantErr:  false,
		},
		{
			name:     "Complex JSON with whitespace",
			input:    "{\n  \"key\": [\n    1,\n    2\n  ]\n}",
			expected: `{"key":[1,2]}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompactJSON(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPrettyJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		indent    string
		wantErr   bool
		checkFunc func(string) bool
	}{
		{
			name:      "Invalid JSON",
			input:     `{"key": value}`,
			indent:    "  ",
			wantErr:   true,
			checkFunc: nil,
		},
		{
			name:      "Tab indent conversion",
			input:     `{"key":"value"}`,
			indent:    "\t",
			wantErr:   false,
			checkFunc: func(s string) bool { return len(s) > 10 }, // Check it's expanded
		},
		{
			name:      "Empty object with indent",
			input:     `{}`,
			indent:    "    ",
			wantErr:   false,
			checkFunc: func(s string) bool { return len(s) >= 2 }, // At least "{}"
		},
		{
			name:      "Array with indent",
			input:     `[1,2,3]`,
			indent:    "  ",
			wantErr:   false,
			checkFunc: func(s string) bool { return len(s) > 7 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrettyJSON(tt.input, tt.indent)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil && !tt.checkFunc(result) {
				t.Errorf("Check function failed for result: %s", result)
			}
		})
	}
}
func TestParser_ParseJSON_ErrorTypes(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	tests := []struct {
		name          string
		input         interface{}
		wantErr       bool
		expectedError string
		errorType     interfaces.ErrorType
	}{
		{
			name:          "nil input returns validation error",
			input:         nil,
			wantErr:       true,
			expectedError: "input data is nil",
			errorType:     interfaces.ErrorTypeValidation,
		},
		{
			name:          "invalid JSON string returns syntax error",
			input:         `{"key": value}`,
			wantErr:       true,
			expectedError: "failed to parse JSON string",
			errorType:     interfaces.ErrorTypeSyntax,
		},
		{
			name:          "invalid JSON bytes returns syntax error",
			input:         []byte(`{"key": value}`),
			wantErr:       true,
			expectedError: "failed to parse JSON bytes",
			errorType:     interfaces.ErrorTypeSyntax,
		},
		{
			name:          "unmarshalable type returns syntax error",
			input:         make(chan int),
			wantErr:       true,
			expectedError: "failed to marshal data",
			errorType:     interfaces.ErrorTypeSyntax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseJSON(ctx, tt.input)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Error("Expected error but got none")
				return
			}

			// Check if error is of type ParseError
			parseErr, ok := err.(*interfaces.ParseError)
			if !ok {
				t.Errorf("Expected ParseError, got %T", err)
				return
			}

			// Check error type
			if parseErr.Type != tt.errorType {
				t.Errorf("Expected error type %v, got %v", tt.errorType, parseErr.Type)
			}

			// Check error message contains expected text
			if !contains(parseErr.Message, tt.expectedError) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedError, parseErr.Message)
			}

			// Result should be nil on error
			if result != nil {
				t.Errorf("Expected nil result on error, got %v", result)
			}
		})
	}
}

func TestParser_ParseJSON_ComplexTypes(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	type CustomStruct struct {
		ID   int      `json:"id"`
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		check   func(interface{}) bool
	}{
		{
			name:    "nested struct",
			input:   CustomStruct{ID: 1, Name: "test", Tags: []string{"a", "b"}},
			wantErr: false,
			check: func(result interface{}) bool {
				m, ok := result.(map[string]interface{})
				return ok && m["id"].(float64) == 1 && m["name"].(string) == "test"
			},
		},
		{
			name:    "slice of structs",
			input:   []CustomStruct{{ID: 1, Name: "first"}, {ID: 2, Name: "second"}},
			wantErr: false,
			check: func(result interface{}) bool {
				arr, ok := result.([]interface{})
				return ok && len(arr) == 2
			},
		},
		{
			name:    "map with complex values",
			input:   map[string]CustomStruct{"item": {ID: 1, Name: "test"}},
			wantErr: false,
			check: func(result interface{}) bool {
				m, ok := result.(map[string]interface{})
				return ok && m["item"] != nil
			},
		},
		{
			name:    "nil pointer",
			input:   (*CustomStruct)(nil),
			wantErr: false,
			check: func(result interface{}) bool {
				return result == nil
			},
		},
		{
			name:    "pointer to struct",
			input:   &CustomStruct{ID: 1, Name: "pointer"},
			wantErr: false,
			check: func(result interface{}) bool {
				m, ok := result.(map[string]interface{})
				return ok && m["id"].(float64) == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseJSON(ctx, tt.input)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !tt.wantErr && !tt.check(result) {
				t.Errorf("Check failed for result: %v", result)
			}
		})
	}
}

// Helper functions for tests
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		(len(substr) < len(s) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch va := a.(type) {
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i, v := range va {
			if !deepEqual(v, vb[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
