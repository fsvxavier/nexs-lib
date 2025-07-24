package checks

import "testing"

func TestIsString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "string value returns true",
			input:    "test",
			expected: true,
		},
		{
			name:     "empty string returns true",
			input:    "",
			expected: true,
		},
		{
			name:     "integer returns false",
			input:    123,
			expected: false,
		},
		{
			name:     "float returns false",
			input:    123.45,
			expected: false,
		},
		{
			name:     "boolean returns false",
			input:    true,
			expected: false,
		},
		{
			name:     "nil returns false",
			input:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsString(tt.input); got != tt.expected {
				t.Errorf("IsString() = %v, want %v", got, tt.expected)
			}
		})
	}
}
