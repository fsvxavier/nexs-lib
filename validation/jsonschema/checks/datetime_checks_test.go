package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateTimeFormatChecker_IsFormat(t *testing.T) {
	checker := NewDateTimeFormatChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid date only", "2006-01-02", true},
		{"valid datetime RFC3339", "2006-01-02T15:04:05Z", true},
		{"valid datetime RFC3339 with timezone", "2006-01-02T15:04:05-07:00", true},
		{"valid datetime RFC3339 nano", "2006-01-02T15:04:05.999999999Z", true},
		{"valid datetime ISO8601", "2006-01-02T15:04:05-07:00", true},
		{"valid time only", "15:04:05", true},
		{"valid time with timezone", "15:04:05-07:00", true},
		{"recent datetime", "2023-12-25T10:30:45Z", true},
		{"recent datetime with ms", "2023-12-25T10:30:45.999Z", true},
		{"recent datetime with timezone", "2023-12-25T10:30:45+00:00", true},
		{"invalid date string", "invalid-date", false},
		{"non-string input", 12345, false},
		{"nil input", nil, false},
		{"empty string", "", true}, // AllowEmpty is true by default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %v", tt.input)
		})
	}
}

func TestDateTimeFormatChecker_AllowEmpty(t *testing.T) {
	checker := NewDateTimeFormatChecker()
	checker.AllowEmpty = false

	assert.False(t, checker.IsFormat(""))
	assert.True(t, checker.IsFormat("2006-01-02T15:04:05Z"))
}

func TestDateTimeFormatChecker_CustomFormats(t *testing.T) {
	customFormats := []string{"2006-01-02", "15:04:05"}
	checker := NewDateTimeFormatChecker().WithFormats(customFormats)

	assert.True(t, checker.IsFormat("2006-01-02"))
	assert.True(t, checker.IsFormat("15:04:05"))
	assert.False(t, checker.IsFormat("2006-01-02T15:04:05Z"))
}

func TestDateTimeFormatChecker_Check(t *testing.T) {
	checker := NewDateTimeFormatChecker()

	// Valid data
	errors := checker.Check("2006-01-02T15:04:05Z")
	assert.Empty(t, errors)

	// Invalid data
	errors = checker.Check("invalid-date")
	require.Len(t, errors, 1)
	assert.Equal(t, "datetime", errors[0].Field)
	assert.Equal(t, "INVALID_DATETIME_FORMAT", errors[0].ErrorType)
}

func TestISO8601DateChecker_IsFormat(t *testing.T) {
	checker := NewISO8601DateChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid ISO8601 datetime", "2006-01-02T15:04:05-07:00", true},
		{"invalid format", "invalid-date", false},
		{"non-string", 123, false},
		{"empty string", "", true}, // AllowEmpty is true by default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeOnlyChecker_IsFormat(t *testing.T) {
	checker := NewTimeOnlyChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid time", "15:04:05", true},
		{"valid time with timezone", "15:04:05-07:00", true},
		{"invalid time", "25:99:99", false},
		{"date string", "2006-01-02", false},
		{"non-string", 123, false},
		{"empty string", "", true}, // AllowEmpty is true by default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDateOnlyChecker_IsFormat(t *testing.T) {
	checker := NewDateOnlyChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid date", "2006-01-02", true},
		{"invalid date", "2006-13-32", false},
		{"datetime string", "2006-01-02T15:04:05Z", false},
		{"non-string", 123, false},
		{"empty string", "", true}, // AllowEmpty is true by default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
