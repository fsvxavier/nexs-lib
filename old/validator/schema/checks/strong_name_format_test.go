package checks

import (
	"testing"
)

func TestStrongNameFormat_IsFormat(t *testing.T) {
	tests := []struct {
		input    any
		expected bool
	}{
		{"VALID_NAME", true},
		{"INVALID-name", false},
		{"ANOTHER_VALID_NAME", true},
		{"invalid_name", false},
		{"123INVALID", false},
		{"VALID_NAME_123", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(test.input.(string), func(t *testing.T) {
			snf := StrongNameFormat{}
			result := snf.IsFormat(test.input)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}
