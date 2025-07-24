package checks

import "testing"

func TestEmptyStringChecker_IsFormat(t *testing.T) {
	checker := EmptyStringChecker{}

	tests := []struct {
		input    interface{}
		expected bool
	}{
		{"", false},
		{"non-empty", true},
		{123, false},
		{nil, false},
		{true, false},
	}

	for _, test := range tests {
		result := checker.IsFormat(test.input)
		if result != test.expected {
			t.Errorf("IsFormat(%v) = %v; expected %v", test.input, result, test.expected)
		}
	}
}
