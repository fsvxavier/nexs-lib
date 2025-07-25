package json

import (
	"context"
	"strings"
	"testing"
)

func TestParseJSONToTypeCompat(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantErr  bool
		expected interface{}
	}{
		{
			name:     "map compatibility",
			input:    map[string]interface{}{"key": "value"},
			wantErr:  false,
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "slice compatibility",
			input:    []interface{}{1, 2, 3},
			wantErr:  false,
			expected: []interface{}{1, 2, 3},
		},
		{
			name:    "invalid type compatibility",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJSONToTypeCompat[interface{}](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONToTypeCompat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected result to not be nil")
			}
		})
	}
}

func TestAdvancedParser(t *testing.T) {
	parser := NewAdvancedParser()
	if parser == nil {
		t.Error("Expected advanced parser to be created")
	}
}

func TestAdvancedParser_WithOptions(t *testing.T) {
	parser := NewAdvancedParser().
		WithComments(true).
		WithTrailingCommas(true).
		WithStrictNumbers(true)

	if !parser.allowComments {
		t.Error("Expected comments to be enabled")
	}
	if !parser.allowTrailingCommas {
		t.Error("Expected trailing commas to be enabled")
	}
	if !parser.strictNumbers {
		t.Error("Expected strict numbers to be enabled")
	}
}

func TestAdvancedParser_ParseAdvanced_WithComments(t *testing.T) {
	parser := NewAdvancedParser().WithComments(true)
	ctx := context.Background()

	jsonWithComments := `{
		// This is a comment
		"name": "test",
		/* This is a block comment */
		"value": 42
	}`

	result, err := parser.ParseAdvanced(ctx, jsonWithComments)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected result to be a map")
	}

	if resultMap["name"] != "test" {
		t.Errorf("Expected name to be 'test', got: %v", resultMap["name"])
	}

	if resultMap["value"] != float64(42) {
		t.Errorf("Expected value to be 42, got: %v", resultMap["value"])
	}
}

func TestAdvancedParser_ParseAdvanced_WithTrailingCommas(t *testing.T) {
	parser := NewAdvancedParser().WithTrailingCommas(true)
	ctx := context.Background()

	jsonWithTrailingCommas := `{
		"name": "test",
		"value": 42,
	}`

	result, err := parser.ParseAdvanced(ctx, jsonWithTrailingCommas)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected result to be a map")
	}

	if resultMap["name"] != "test" {
		t.Errorf("Expected name to be 'test', got: %v", resultMap["name"])
	}
}

func TestParseJSONL(t *testing.T) {
	jsonl := `{"name": "Alice", "age": 30}
{"name": "Bob", "age": 25}
{"name": "Charlie", "age": 35}`

	results, err := ParseJSONL(jsonl)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got: %d", len(results))
	}

	// Check first object
	firstObj, ok := results[0].(map[string]interface{})
	if !ok {
		t.Error("Expected first result to be a map")
	}

	if firstObj["name"] != "Alice" {
		t.Errorf("Expected first name to be 'Alice', got: %v", firstObj["name"])
	}
}

func TestParseNDJSON(t *testing.T) {
	ndjson := `{"id": 1, "data": "first"}
{"id": 2, "data": "second"}`

	results, err := ParseNDJSON(ndjson)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got: %d", len(results))
	}
}

func TestParseJSON5(t *testing.T) {
	// Test with simpler JSON5 - just comments and trailing commas
	json5 := `{
		// JSON5 allows comments
		"name": "test",
		"value": 42,
	}`

	result, err := ParseJSON5(json5)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Error("Expected result to not be nil")
	}

	// Verify the parsed content
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Expected result to be a map")
	} else {
		if resultMap["name"] != "test" {
			t.Errorf("Expected name to be 'test', got: %v", resultMap["name"])
		}
		if resultMap["value"] != float64(42) {
			t.Errorf("Expected value to be 42, got: %v", resultMap["value"])
		}
	}
}

func TestStreamParser(t *testing.T) {
	jsonStream := `{"name": "Alice"}
{"name": "Bob"}`

	parser := NewStreamParser(strings.NewReader(jsonStream))

	var first map[string]interface{}
	err := parser.ParseNext(&first)
	if err != nil {
		t.Errorf("Expected no error parsing first object, got: %v", err)
	}

	if first["name"] != "Alice" {
		t.Errorf("Expected first name to be 'Alice', got: %v", first["name"])
	}
}

func TestIsValidJSONL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name: "valid JSONL",
			input: `{"name": "Alice"}
{"name": "Bob"}`,
			expected: true,
		},
		{
			name: "invalid JSONL",
			input: `{"name": "Alice"}
{"name": "Bob"`, // missing closing brace
			expected: false,
		},
		{
			name:     "empty lines should be valid",
			input:    "{\"name\": \"Alice\"}\n\n{\"name\": \"Bob\"}",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidJSONL(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidJSONL() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToJSONL(t *testing.T) {
	objects := []interface{}{
		map[string]interface{}{"name": "Alice", "age": 30},
		map[string]interface{}{"name": "Bob", "age": 25},
	}

	result, err := ConvertToJSONL(objects)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	lines := strings.Split(result, "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got: %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		if err := ValidateJSONString(line); err != nil {
			t.Errorf("Line %d is not valid JSON: %s", i+1, line)
		}
	}
}

func TestMergeJSON(t *testing.T) {
	obj1 := map[string]interface{}{"name": "Alice", "age": 30}
	obj2 := `{"city": "New York", "age": 31}` // Override age

	result, err := MergeJSON(obj1, obj2)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result["name"] != "Alice" {
		t.Errorf("Expected name to be 'Alice', got: %v", result["name"])
	}

	if result["city"] != "New York" {
		t.Errorf("Expected city to be 'New York', got: %v", result["city"])
	}

	// Should be overridden by obj2
	if result["age"] != float64(31) {
		t.Errorf("Expected age to be 31, got: %v", result["age"])
	}
}

func TestExtractPath(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "extract nested name",
			path:     "user.profile.name",
			expected: "Alice",
			wantErr:  false,
		},
		{
			name:     "extract nested age",
			path:     "user.profile.age",
			expected: 30,
			wantErr:  false,
		},
		{
			name:     "extract user object",
			path:     "user",
			expected: nil, // We'll check this differently
			wantErr:  false,
		},
		{
			name:    "non-existent path",
			path:    "user.profile.email",
			wantErr: true,
		},
		{
			name:    "invalid path",
			path:    "user.profile.name.invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractPath(data, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.path == "user" {
					// Special handling for user object
					if userMap, ok := result.(map[string]interface{}); !ok {
						t.Errorf("ExtractPath() for user should return a map, got %T", result)
					} else if userMap["profile"] == nil {
						t.Errorf("ExtractPath() for user should contain profile")
					}
				} else if result != tt.expected {
					t.Errorf("ExtractPath() = %v, expected %v", result, tt.expected)
				}
			}
		})
	}
}

func TestAdvancedParser_removeComments(t *testing.T) {
	parser := NewAdvancedParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "line comment",
			input:    `{"name": "test"} // comment`,
			expected: `{"name": "test"} `,
		},
		{
			name:     "block comment",
			input:    `{"name": /* comment */ "test"}`,
			expected: `{"name":  "test"}`,
		},
		{
			name:     "comment in string should be preserved",
			input:    `{"comment": "this // is not a comment"}`,
			expected: `{"comment": "this // is not a comment"}`,
		},
		{
			name: "multiline block comment",
			input: `{
				"name": "test",
				/* this is a
				   multiline comment */
				"value": 42
			}`,
			expected: `{
				"name": "test",
				
				"value": 42
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.removeComments(tt.input)
			if result != tt.expected {
				t.Errorf("removeComments() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestAdvancedParser_removeTrailingCommas(t *testing.T) {
	parser := NewAdvancedParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trailing comma in object",
			input:    `{"name": "test", "value": 42,}`,
			expected: `{"name": "test", "value": 42}`,
		},
		{
			name:     "trailing comma in array",
			input:    `[1, 2, 3,]`,
			expected: `[1, 2, 3]`,
		},
		{
			name:     "no trailing comma",
			input:    `{"name": "test", "value": 42}`,
			expected: `{"name": "test", "value": 42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.removeTrailingCommas(tt.input)
			if result != tt.expected {
				t.Errorf("removeTrailingCommas() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
