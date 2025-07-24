package checks

import (
	"testing"
)

func TestDateTimeChecker_IsFormat(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected bool
	}{
		{"2006-01-02", true},
		{"2006-01-02T15:04:05Z", true},
		{"2006-01-02T15:04:05.999Z", true},
		{"2006-01-02T15:04:05Z07:00", false},
		{"2006-01-02T15:04:05.999Z07:00", false},
		{"2021-12-31T23:59:59Z", true},
		{"2021-12-31T23:59:59.999Z", true},
		{"2021-12-31T23:59:59+00:00", true},
		{"2021-12-31T23:59:59.999+00:00", true},
		{"invalid-date", false},
		{12345, false},
		{nil, false},
	}

	checker := DateTimeChecker{}

	for _, test := range tests {
		result := checker.IsFormat(test.input)
		if result != test.expected {
			t.Errorf("IsFormat(%v) = %v; expected %v", test.input, result, test.expected)
		}
	}
}
