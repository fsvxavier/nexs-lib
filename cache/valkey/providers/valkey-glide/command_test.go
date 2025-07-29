package valkeyglide

import (
	"errors"
	"testing"
)

func TestCommand_NewCommand(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      error
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "successful result",
			result:   "test_value",
			err:      nil,
			expected: "test_value",
			wantErr:  false,
		},
		{
			name:     "error result",
			result:   nil,
			err:      errors.New("test error"),
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "nil result no error",
			result:   nil,
			err:      nil,
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(tt.result, tt.err)

			result, err := cmd.Result()
			if (err != nil) != tt.wantErr {
				t.Errorf("Result() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("Result() = %v, want %v", result, tt.expected)
			}

			if cmd.Err() != tt.err {
				t.Errorf("Err() = %v, want %v", cmd.Err(), tt.err)
			}
		})
	}
}

func TestCommand_String(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      error
		expected string
		wantErr  bool
	}{
		{
			name:     "string value",
			result:   "test",
			err:      nil,
			expected: "test",
			wantErr:  false,
		},
		{
			name:     "byte slice",
			result:   []byte("test"),
			err:      nil,
			expected: "test",
			wantErr:  false,
		},
		{
			name:     "int value",
			result:   42,
			err:      nil,
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "int64 value",
			result:   int64(42),
			err:      nil,
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "float64 value",
			result:   3.14,
			err:      nil,
			expected: "3.14",
			wantErr:  false,
		},
		{
			name:     "bool true",
			result:   true,
			err:      nil,
			expected: "1",
			wantErr:  false,
		},
		{
			name:     "bool false",
			result:   false,
			err:      nil,
			expected: "0",
			wantErr:  false,
		},
		{
			name:     "nil result",
			result:   nil,
			err:      nil,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "error case",
			result:   "test",
			err:      errors.New("test error"),
			expected: "",
			wantErr:  true,
		},
		{
			name:     "complex type",
			result:   map[string]string{"key": "value"},
			err:      nil,
			expected: "map[key:value]",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(tt.result, tt.err)

			result, err := cmd.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommand_Int64(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      error
		expected int64
		wantErr  bool
	}{
		{
			name:     "int64 value",
			result:   int64(42),
			err:      nil,
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "int value",
			result:   42,
			err:      nil,
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "int32 value",
			result:   int32(42),
			err:      nil,
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "string number",
			result:   "42",
			err:      nil,
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "byte slice number",
			result:   []byte("42"),
			err:      nil,
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "invalid string",
			result:   "not_a_number",
			err:      nil,
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "nil result",
			result:   nil,
			err:      nil,
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "error case",
			result:   42,
			err:      errors.New("test error"),
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "unsupported type",
			result:   3.14,
			err:      nil,
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(tt.result, tt.err)

			result, err := cmd.Int64()
			if (err != nil) != tt.wantErr {
				t.Errorf("Int64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("Int64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommand_Bool(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      error
		expected bool
		wantErr  bool
	}{
		{
			name:     "bool true",
			result:   true,
			err:      nil,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "bool false",
			result:   false,
			err:      nil,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "int non-zero",
			result:   42,
			err:      nil,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "int zero",
			result:   0,
			err:      nil,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "string non-empty",
			result:   "test",
			err:      nil,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "string empty",
			result:   "",
			err:      nil,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "string zero",
			result:   "0",
			err:      nil,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "byte slice non-empty",
			result:   []byte("test"),
			err:      nil,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "byte slice empty",
			result:   []byte{},
			err:      nil,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "nil result",
			result:   nil,
			err:      nil,
			expected: false,
			wantErr:  true,
		},
		{
			name:     "error case",
			result:   true,
			err:      errors.New("test error"),
			expected: false,
			wantErr:  true,
		},
		{
			name:     "unsupported type",
			result:   map[string]string{},
			err:      nil,
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(tt.result, tt.err)

			result, err := cmd.Bool()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("Bool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommand_Float64(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      error
		expected float64
		wantErr  bool
	}{
		{
			name:     "float64 value",
			result:   3.14,
			err:      nil,
			expected: 3.14,
			wantErr:  false,
		},
		{
			name:     "float32 value",
			result:   float32(3.14),
			err:      nil,
			expected: 3.1400001049041748, // float32 precision
			wantErr:  false,
		},
		{
			name:     "string number",
			result:   "3.14",
			err:      nil,
			expected: 3.14,
			wantErr:  false,
		},
		{
			name:     "byte slice number",
			result:   []byte("3.14"),
			err:      nil,
			expected: 3.14,
			wantErr:  false,
		},
		{
			name:     "invalid string",
			result:   "not_a_number",
			err:      nil,
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "nil result",
			result:   nil,
			err:      nil,
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "error case",
			result:   3.14,
			err:      errors.New("test error"),
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(tt.result, tt.err)

			result, err := cmd.Float64()
			if (err != nil) != tt.wantErr {
				t.Errorf("Float64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("Float64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommand_Slice(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      error
		expected []interface{}
		wantErr  bool
	}{
		{
			name:     "interface slice",
			result:   []interface{}{"a", "b", "c"},
			err:      nil,
			expected: []interface{}{"a", "b", "c"},
			wantErr:  false,
		},
		{
			name:     "string slice",
			result:   []string{"a", "b", "c"},
			err:      nil,
			expected: []interface{}{"a", "b", "c"},
			wantErr:  false,
		},
		{
			name:     "int slice",
			result:   []int{1, 2, 3},
			err:      nil,
			expected: []interface{}{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "int64 slice",
			result:   []int64{1, 2, 3},
			err:      nil,
			expected: []interface{}{int64(1), int64(2), int64(3)},
			wantErr:  false,
		},
		{
			name:     "single value",
			result:   "test",
			err:      nil,
			expected: []interface{}{"test"},
			wantErr:  false,
		},
		{
			name:     "nil result",
			result:   nil,
			err:      nil,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "error case",
			result:   []interface{}{"a", "b"},
			err:      errors.New("test error"),
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(tt.result, tt.err)

			result, err := cmd.Slice()
			if (err != nil) != tt.wantErr {
				t.Errorf("Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !compareSlices(result, tt.expected) {
				t.Errorf("Slice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommand_StringSlice(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      error
		expected []string
		wantErr  bool
	}{
		{
			name:     "string slice",
			result:   []string{"a", "b", "c"},
			err:      nil,
			expected: []string{"a", "b", "c"},
			wantErr:  false,
		},
		{
			name:     "interface slice",
			result:   []interface{}{"a", 1, true},
			err:      nil,
			expected: []string{"a", "1", "true"},
			wantErr:  false,
		},
		{
			name:     "single string",
			result:   "test",
			err:      nil,
			expected: []string{"test"},
			wantErr:  false,
		},
		{
			name:     "nil result",
			result:   nil,
			err:      nil,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "error case",
			result:   []string{"a", "b"},
			err:      errors.New("test error"),
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "unsupported type",
			result:   42,
			err:      nil,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(tt.result, tt.err)

			result, err := cmd.StringSlice()
			if (err != nil) != tt.wantErr {
				t.Errorf("StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !compareStringSlices(result, tt.expected) {
				t.Errorf("StringSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommand_StringMap(t *testing.T) {
	tests := []struct {
		name     string
		result   interface{}
		err      error
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "string map",
			result:   map[string]string{"key1": "value1", "key2": "value2"},
			err:      nil,
			expected: map[string]string{"key1": "value1", "key2": "value2"},
			wantErr:  false,
		},
		{
			name:     "interface map",
			result:   map[string]interface{}{"key1": "value1", "key2": 42},
			err:      nil,
			expected: map[string]string{"key1": "value1", "key2": "42"},
			wantErr:  false,
		},
		{
			name:     "generic interface map",
			result:   map[interface{}]interface{}{"key1": "value1", 123: "value2"},
			err:      nil,
			expected: map[string]string{"key1": "value1", "123": "value2"},
			wantErr:  false,
		},
		{
			name:     "nil result",
			result:   nil,
			err:      nil,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "error case",
			result:   map[string]string{"key": "value"},
			err:      errors.New("test error"),
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "unsupported type",
			result:   "not a map",
			err:      nil,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCommand(tt.result, tt.err)

			result, err := cmd.StringMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("StringMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !compareStringMaps(result, tt.expected) {
				t.Errorf("StringMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper functions for comparisons

func compareSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareStringMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
