package env

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Error("Expected parser to be created")
	}
	if parser.config == nil {
		t.Error("Expected config to be initialized")
	}
}

func TestParser_ParseString(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		input     string
		expectKey string
		expectVal string
		hasError  bool
	}{
		{"KEY=value", "KEY", "value", false},
		{"DB_HOST=localhost", "DB_HOST", "localhost", false},
		{"PORT=8080", "PORT", "8080", false},
		{"DEBUG=true", "DEBUG", "true", false},
		{"EMPTY=", "EMPTY", "", false},
		{"SPACES = value with spaces ", "SPACES", " value with spaces ", false},
		{"", "", "", true},
		{"invalid", "", "", true},
		{"KEY", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parser.ParseString(ctx, tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result to be non-nil")
				return
			}

			if result.Key != tt.expectKey {
				t.Errorf("Expected key %s, got %s", tt.expectKey, result.Key)
			}

			if result.RawValue != tt.expectVal {
				t.Errorf("Expected value %s, got %s", tt.expectVal, result.RawValue)
			}
		})
	}
}

func TestParser_ParseEnvVar(t *testing.T) {
	parser := NewParser()

	// Set test environment variable
	testKey := "TEST_ENV_VAR"
	testValue := "test_value"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	result := parser.ParseEnvVar(testKey)
	if result == nil {
		t.Error("Expected result to be non-nil")
		return
	}

	if !result.Found {
		t.Error("Expected env var to be found")
	}

	if result.Key != testKey {
		t.Errorf("Expected key %s, got %s", testKey, result.Key)
	}

	if result.RawValue != testValue {
		t.Errorf("Expected value %s, got %s", testValue, result.RawValue)
	}

	// Test non-existent variable
	result = parser.ParseEnvVar("NON_EXISTENT_VAR")
	if result.Found {
		t.Error("Expected env var to not be found")
	}
}

func TestParser_ParseInt(t *testing.T) {
	parser := NewParser()

	testKey := "TEST_INT"
	os.Setenv(testKey, "42")
	defer os.Unsetenv(testKey)

	result := parser.ParseInt(testKey)
	if result == nil || !result.Found {
		t.Error("Expected result to be found")
		return
	}

	if result.Type != "int" {
		t.Errorf("Expected type int, got %s", result.Type)
	}

	if result.Value == nil {
		t.Error("Expected value to be non-nil")
		return
	}

	value := result.Value.(*int)
	if *value != 42 {
		t.Errorf("Expected 42, got %d", *value)
	}
}

func TestParser_ParseBool(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"TRUE", true},
		{"FALSE", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			testKey := "TEST_BOOL"
			os.Setenv(testKey, tt.value)
			defer os.Unsetenv(testKey)

			result := parser.ParseBool(testKey)
			if result == nil || !result.Found {
				t.Error("Expected result to be found")
				return
			}

			if result.Type != "bool" {
				t.Errorf("Expected type bool, got %s", result.Type)
			}

			if result.Value == nil {
				t.Error("Expected value to be non-nil")
				return
			}

			value := result.Value.(*bool)
			if *value != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, *value)
			}
		})
	}
}

func TestParser_ParseDuration(t *testing.T) {
	parser := NewParser()

	testKey := "TEST_DURATION"
	os.Setenv(testKey, "5m30s")
	defer os.Unsetenv(testKey)

	result := parser.ParseDuration(testKey)
	if result == nil || !result.Found {
		t.Error("Expected result to be found")
		return
	}

	if result.Type != "duration" {
		t.Errorf("Expected type duration, got %s", result.Type)
	}

	if result.Value == nil {
		t.Error("Expected value to be non-nil")
		return
	}

	value := result.Value.(*time.Duration)
	expected := 5*time.Minute + 30*time.Second
	if *value != expected {
		t.Errorf("Expected %v, got %v", expected, *value)
	}
}

func TestParser_ParseSliceString(t *testing.T) {
	parser := NewParser()

	testKey := "TEST_SLICE"
	os.Setenv(testKey, "a,b,c,d")
	defer os.Unsetenv(testKey)

	result := parser.ParseSliceString(testKey, ",")
	if result == nil || !result.Found {
		t.Error("Expected result to be found")
		return
	}

	if result.Type != "[]string" {
		t.Errorf("Expected type []string, got %s", result.Type)
	}

	if result.Value == nil {
		t.Error("Expected value to be non-nil")
		return
	}

	value := result.Value.(*[]string)
	expected := []string{"a", "b", "c", "d"}
	if len(*value) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(*value))
		return
	}

	for i, v := range *value {
		if v != expected[i] {
			t.Errorf("Expected %s at index %d, got %s", expected[i], i, v)
		}
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test ParseEnvInt32
	os.Setenv("TEST_INT32", "32")
	defer os.Unsetenv("TEST_INT32")

	result32 := ParseEnvInt32("TEST_INT32")
	if result32 == nil {
		t.Error("Expected ParseEnvInt32 to return non-nil")
	} else if *result32 != 32 {
		t.Errorf("Expected 32, got %d", *result32)
	}

	// Test GetEnv
	os.Setenv("TEST_GET", "test_value")
	defer os.Unsetenv("TEST_GET")

	value := GetEnv("TEST_GET", "default")
	if value != "test_value" {
		t.Errorf("Expected test_value, got %s", value)
	}

	value = GetEnv("NON_EXISTENT", "default")
	if value != "default" {
		t.Errorf("Expected default, got %s", value)
	}

	// Test GetEnvInt
	os.Setenv("TEST_GET_INT", "123")
	defer os.Unsetenv("TEST_GET_INT")

	intValue := GetEnvInt("TEST_GET_INT", 0)
	if intValue != 123 {
		t.Errorf("Expected 123, got %d", intValue)
	}

	intValue = GetEnvInt("NON_EXISTENT_INT", 999)
	if intValue != 999 {
		t.Errorf("Expected 999, got %d", intValue)
	}

	// Test GetEnvBool
	os.Setenv("TEST_GET_BOOL", "true")
	defer os.Unsetenv("TEST_GET_BOOL")

	boolValue := GetEnvBool("TEST_GET_BOOL", false)
	if !boolValue {
		t.Error("Expected true, got false")
	}

	boolValue = GetEnvBool("NON_EXISTENT_BOOL", true)
	if !boolValue {
		t.Error("Expected true (default), got false")
	}
}

func TestSetEnv(t *testing.T) {
	testKey := "TEST_SET_ENV"
	defer os.Unsetenv(testKey)

	tests := []struct {
		value    interface{}
		expected string
	}{
		{"string_value", "string_value"},
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
		{time.Second * 5, "5s"},
	}

	for _, tt := range tests {
		err := SetEnv(testKey, tt.value)
		if err != nil {
			t.Errorf("Unexpected error setting env: %v", err)
			continue
		}

		result := os.Getenv(testKey)
		if result != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, result)
		}
	}
}
func TestParseEnvInt(t *testing.T) {
	testKey := "TEST_PARSE_ENV_INT"

	// Test valid integer
	os.Setenv(testKey, "123")
	defer os.Unsetenv(testKey)

	result := ParseEnvInt(testKey)
	if result == nil {
		t.Error("Expected ParseEnvInt to return non-nil")
	} else if *result != 123 {
		t.Errorf("Expected 123, got %d", *result)
	}

	// Test non-existent env var
	result = ParseEnvInt("NON_EXISTENT_INT")
	if result != nil {
		t.Error("Expected ParseEnvInt to return nil for non-existent env var")
	}

	// Test invalid integer
	os.Setenv(testKey, "invalid")
	result = ParseEnvInt(testKey)
	if result != nil {
		t.Error("Expected ParseEnvInt to return nil for invalid integer")
	}
}

func TestParseEnvInt64(t *testing.T) {
	testKey := "TEST_PARSE_ENV_INT64"

	// Test valid int64
	os.Setenv(testKey, "9223372036854775807")
	defer os.Unsetenv(testKey)

	result := ParseEnvInt64(testKey)
	if result == nil {
		t.Error("Expected ParseEnvInt64 to return non-nil")
	} else if *result != 9223372036854775807 {
		t.Errorf("Expected 9223372036854775807, got %d", *result)
	}

	// Test non-existent env var
	result = ParseEnvInt64("NON_EXISTENT_INT64")
	if result != nil {
		t.Error("Expected ParseEnvInt64 to return nil for non-existent env var")
	}

	// Test invalid int64
	os.Setenv(testKey, "invalid")
	result = ParseEnvInt64(testKey)
	if result != nil {
		t.Error("Expected ParseEnvInt64 to return nil for invalid int64")
	}
}

func TestParseEnvFloat64(t *testing.T) {
	testKey := "TEST_PARSE_ENV_FLOAT64"

	// Test valid float64
	os.Setenv(testKey, "3.14159")
	defer os.Unsetenv(testKey)

	result := ParseEnvFloat64(testKey)
	if result == nil {
		t.Error("Expected ParseEnvFloat64 to return non-nil")
	} else if *result != 3.14159 {
		t.Errorf("Expected 3.14159, got %f", *result)
	}

	// Test non-existent env var
	result = ParseEnvFloat64("NON_EXISTENT_FLOAT64")
	if result != nil {
		t.Error("Expected ParseEnvFloat64 to return nil for non-existent env var")
	}

	// Test invalid float64
	os.Setenv(testKey, "invalid")
	result = ParseEnvFloat64(testKey)
	if result != nil {
		t.Error("Expected ParseEnvFloat64 to return nil for invalid float64")
	}
}

func TestParseEnvDuration(t *testing.T) {
	testKey := "TEST_PARSE_ENV_DURATION"

	// Test valid duration
	os.Setenv(testKey, "10m30s")
	defer os.Unsetenv(testKey)

	result := ParseEnvDuration(testKey)
	expected := 10*time.Minute + 30*time.Second
	if result == nil {
		t.Error("Expected ParseEnvDuration to return non-nil")
	} else if *result != expected {
		t.Errorf("Expected %v, got %v", expected, *result)
	}

	// Test non-existent env var
	result = ParseEnvDuration("NON_EXISTENT_DURATION")
	if result != nil {
		t.Error("Expected ParseEnvDuration to return nil for non-existent env var")
	}

	// Test invalid duration
	os.Setenv(testKey, "invalid")
	result = ParseEnvDuration(testKey)
	if result != nil {
		t.Error("Expected ParseEnvDuration to return nil for invalid duration")
	}
}

func TestParseEnvBool(t *testing.T) {
	testKey := "TEST_PARSE_ENV_BOOL"

	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"TRUE", true},
		{"FALSE", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			os.Setenv(testKey, tt.value)
			defer os.Unsetenv(testKey)

			result := ParseEnvBool(testKey)
			if result == nil {
				t.Error("Expected ParseEnvBool to return non-nil")
			} else if *result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, *result)
			}
		})
	}

	// Test non-existent env var
	result := ParseEnvBool("NON_EXISTENT_BOOL")
	if result != nil {
		t.Error("Expected ParseEnvBool to return nil for non-existent env var")
	}

	// Test invalid bool
	os.Setenv(testKey, "invalid")
	defer os.Unsetenv(testKey)
	result = ParseEnvBool(testKey)
	if result != nil {
		t.Error("Expected ParseEnvBool to return nil for invalid bool")
	}
}

func TestParseEnvSliceString(t *testing.T) {
	testKey := "TEST_PARSE_ENV_SLICE"

	// Test valid slice with comma separator
	os.Setenv(testKey, "apple,banana,cherry")
	defer os.Unsetenv(testKey)

	result := ParseEnvSliceString(testKey, ",")
	expected := []string{"apple", "banana", "cherry"}
	if result == nil {
		t.Error("Expected ParseEnvSliceString to return non-nil")
	} else {
		if len(*result) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(*result))
		} else {
			for i, v := range *result {
				if v != expected[i] {
					t.Errorf("Expected %s at index %d, got %s", expected[i], i, v)
				}
			}
		}
	}

	// Test with custom separator
	os.Setenv(testKey, "one;two;three")
	result = ParseEnvSliceString(testKey, ";")
	expected = []string{"one", "two", "three"}
	if result == nil {
		t.Error("Expected ParseEnvSliceString to return non-nil")
	} else {
		if len(*result) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(*result))
		} else {
			for i, v := range *result {
				if v != expected[i] {
					t.Errorf("Expected %s at index %d, got %s", expected[i], i, v)
				}
			}
		}
	}

	// Test non-existent env var
	result = ParseEnvSliceString("NON_EXISTENT_SLICE", ",")
	if result != nil {
		t.Error("Expected ParseEnvSliceString to return nil for non-existent env var")
	}

	// Test empty separator (should default to comma)
	os.Setenv(testKey, "a,b,c")
	result = ParseEnvSliceString(testKey, "")
	expected = []string{"a", "b", "c"}
	if result == nil {
		t.Error("Expected ParseEnvSliceString to return non-nil with empty separator")
	} else {
		if len(*result) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(*result))
		}
	}
}
func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()
	if formatter == nil {
		t.Error("Expected formatter to be created")
	}
}

func TestFormatter_Format(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	tests := []struct {
		name     string
		data     *ParsedEnv
		expected string
		hasError bool
	}{
		{
			name: "valid data",
			data: &ParsedEnv{
				Key:      "DATABASE_URL",
				RawValue: "postgres://localhost:5432/db",
			},
			expected: "DATABASE_URL=postgres://localhost:5432/db",
			hasError: false,
		},
		{
			name: "empty value",
			data: &ParsedEnv{
				Key:      "EMPTY_VAR",
				RawValue: "",
			},
			expected: "EMPTY_VAR=",
			hasError: false,
		},
		{
			name: "value with spaces",
			data: &ParsedEnv{
				Key:      "MESSAGE",
				RawValue: "hello world",
			},
			expected: "MESSAGE=hello world",
			hasError: false,
		},
		{
			name:     "nil data",
			data:     nil,
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.Format(ctx, tt.data)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if string(result) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(result))
			}
		})
	}
}

func TestFormatter_FormatString(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	data := &ParsedEnv{
		Key:      "PORT",
		RawValue: "8080",
	}

	result, err := formatter.FormatString(ctx, data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	expected := "PORT=8080"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test with nil data
	result, err = formatter.FormatString(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil data")
	}
	if result != "" {
		t.Errorf("Expected empty string for error case, got %s", result)
	}
}

func TestFormatter_FormatWriter(t *testing.T) {
	ctx := context.Background()
	formatter := NewFormatter()

	data := &ParsedEnv{
		Key:      "TEST",
		RawValue: "value",
	}

	err := formatter.FormatWriter(ctx, data, nil)
	if err == nil {
		t.Error("Expected error for FormatWriter")
	}
}
func TestNewParserWithConfig(t *testing.T) {
	// Test with custom config
	customConfig := &interfaces.ParserConfig{
		MaxSize: 1024,
	}
	parser := NewParserWithConfig(customConfig)

	if parser == nil {
		t.Error("Expected parser to be created")
	}

	if parser.config == nil {
		t.Error("Expected config to be initialized")
	}

	if parser.config.MaxSize != 1024 {
		t.Errorf("Expected MaxSize to be 1024, got %d", parser.config.MaxSize)
	}

	// Test with nil config
	parser = NewParserWithConfig(nil)
	if parser == nil {
		t.Error("Expected parser to be created even with nil config")
	}

	if parser.config != nil {
		t.Error("Expected config to be nil when passed nil")
	}
}

func TestParser_Parse(t *testing.T) {
	ctx := context.Background()
	parser := NewParser()

	tests := []struct {
		name      string
		data      []byte
		expectKey string
		expectVal string
		hasError  bool
	}{
		{
			name:      "valid key=value",
			data:      []byte("API_KEY=secret123"),
			expectKey: "API_KEY",
			expectVal: "secret123",
			hasError:  false,
		},
		{
			name:      "value with special characters",
			data:      []byte("DATABASE_URL=postgres://user:pass@host:5432/db"),
			expectKey: "DATABASE_URL",
			expectVal: "postgres://user:pass@host:5432/db",
			hasError:  false,
		},
		{
			name:      "empty value",
			data:      []byte("EMPTY_VAR="),
			expectKey: "EMPTY_VAR",
			expectVal: "",
			hasError:  false,
		},
		{
			name:      "value with spaces",
			data:      []byte("MESSAGE=hello world with spaces"),
			expectKey: "MESSAGE",
			expectVal: "hello world with spaces",
			hasError:  false,
		},
		{
			name:     "empty data",
			data:     []byte(""),
			hasError: true,
		},
		{
			name:     "invalid format - no equals",
			data:     []byte("INVALID_FORMAT"),
			hasError: true,
		},
		{
			name:      "invalid format - key only",
			data:      []byte("KEY_ONLY="),
			expectKey: "KEY_ONLY",
			expectVal: "",
			hasError:  false,
		},
		{
			name:      "multiple equals signs",
			data:      []byte("KEY=value=with=equals"),
			expectKey: "KEY",
			expectVal: "value=with=equals",
			hasError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tt.data)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result to be non-nil")
				return
			}

			if result.Key != tt.expectKey {
				t.Errorf("Expected key %s, got %s", tt.expectKey, result.Key)
			}

			if result.RawValue != tt.expectVal {
				t.Errorf("Expected value %s, got %s", tt.expectVal, result.RawValue)
			}

			if result.Type != "string" {
				t.Errorf("Expected type string, got %s", result.Type)
			}

			if !result.Found {
				t.Error("Expected Found to be true")
			}
		})
	}
}

func TestParser_ParseInt32_EdgeCases(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		envKey   string
		envValue string
		setValue bool
		expected interface{}
		errType  string
	}{
		{
			name:     "Valid int32",
			envKey:   "TEST_INT32_VALID",
			envValue: "123456",
			setValue: true,
			expected: int32(123456),
			errType:  "int32",
		},
		{
			name:     "Invalid int32 - too large",
			envKey:   "TEST_INT32_LARGE",
			envValue: "9999999999999",
			setValue: true,
			expected: nil,
			errType:  "error",
		},
		{
			name:     "Invalid int32 - non-numeric",
			envKey:   "TEST_INT32_STRING",
			envValue: "not_a_number",
			setValue: true,
			expected: nil,
			errType:  "error",
		},
		{
			name:     "Missing env var",
			envKey:   "TEST_INT32_MISSING",
			envValue: "",
			setValue: false,
			expected: nil,
			errType:  "undefined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := parser.ParseInt32(tt.envKey)

			if result.Type != tt.errType {
				t.Errorf("Expected type %s, got %s", tt.errType, result.Type)
			}

			if tt.setValue && !result.Found {
				t.Error("Expected Found to be true when env var is set")
			}

			if !tt.setValue && result.Found {
				t.Error("Expected Found to be false when env var is not set")
			}

			if tt.expected != nil {
				if result.Value == nil {
					t.Error("Expected value to be non-nil")
				} else if val, ok := result.Value.(*int32); ok {
					if *val != tt.expected.(int32) {
						t.Errorf("Expected %v, got %v", tt.expected, *val)
					}
				}
			} else if result.Value != nil && tt.errType == "error" {
				t.Error("Expected value to be nil for error type")
			}
		})
	}
}

func TestParseEnvInt32_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		setValue bool
		expected *int32
	}{
		{
			name:     "Valid int32",
			envKey:   "TEST_UTILITY_INT32_VALID",
			envValue: "42",
			setValue: true,
			expected: func() *int32 { v := int32(42); return &v }(),
		},
		{
			name:     "Invalid int32",
			envKey:   "TEST_UTILITY_INT32_INVALID",
			envValue: "invalid",
			setValue: true,
			expected: nil,
		},
		{
			name:     "Missing env var",
			envKey:   "TEST_UTILITY_INT32_MISSING",
			envValue: "",
			setValue: false,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := ParseEnvInt32(tt.envKey)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("Expected non-nil result")
				} else if *result != *tt.expected {
					t.Errorf("Expected %v, got %v", *tt.expected, *result)
				}
			}
		})
	}
}

func TestSetEnv_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "String value",
			key:      "TEST_SET_STRING",
			value:    "test_value",
			expected: "test_value",
			wantErr:  false,
		},
		{
			name:     "Int value",
			key:      "TEST_SET_INT",
			value:    123,
			expected: "123",
			wantErr:  false,
		},
		{
			name:     "Bool value",
			key:      "TEST_SET_BOOL",
			value:    true,
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "Float value",
			key:      "TEST_SET_FLOAT",
			value:    3.14,
			expected: "3.14",
			wantErr:  false,
		},
		{
			name:     "Nil value",
			key:      "TEST_SET_NIL",
			value:    nil,
			expected: "<nil>",
			wantErr:  false,
		},
		{
			name:     "Complex type",
			key:      "TEST_SET_COMPLEX",
			value:    map[string]interface{}{"key": "value"},
			expected: "map[key:value]",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.Unsetenv(tt.key) // Clean up

			err := SetEnv(tt.key, tt.value)

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

			actualValue := os.Getenv(tt.key)
			if actualValue != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, actualValue)
			}
		})
	}
}

func TestValidateInput_EdgeCases(t *testing.T) {
	parser := NewParserWithConfig(&interfaces.ParserConfig{
		MaxSize: 10,
	})

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errType interfaces.ErrorType
	}{
		{
			name:    "Valid input within size limit",
			input:   "KEY=value",
			wantErr: false,
			errType: interfaces.ErrorTypeUnknown,
		},
		{
			name:    "Input exceeds size limit",
			input:   "VERY_LONG_KEY=very_long_value_that_exceeds_limit",
			wantErr: true,
			errType: interfaces.ErrorTypeSize,
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
			errType: interfaces.ErrorTypeValidation,
		},
		{
			name:    "Input at size limit",
			input:   "KEY=value2",
			wantErr: false,
			errType: interfaces.ErrorTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateInput(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if parseErr, ok := err.(*interfaces.ParseError); ok {
					if parseErr.Type != tt.errType {
						t.Errorf("Expected error type %v, got %v", tt.errType, parseErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
