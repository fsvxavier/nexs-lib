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

func TestRFC3339TimeOnlyFormat_Validation(t *testing.T) {
	checker := NewTimeOnlyChecker()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// RFC3339 time-only with UTC (Z)
		{"RFC3339 UTC time", "15:04:05Z", true},
		{"RFC3339 UTC time noon", "12:00:00Z", true},
		{"RFC3339 UTC time midnight", "00:00:00Z", true},

		// RFC3339 time-only with timezone offset
		{"RFC3339 positive offset", "15:04:05+07:00", true},
		{"RFC3339 negative offset", "15:04:05-07:00", true},
		{"RFC3339 half hour offset", "15:04:05+05:30", true},

		// Standard time formats (should still work)
		{"time only no timezone", "15:04:05", true},
		{"ISO8601 time", "15:04:05", true},

		// Invalid formats
		{"invalid time", "25:61:61Z", false},
		{"missing seconds", "15:04Z", false},
		{"invalid timezone", "15:04:05+25:00", false},
		{"empty string", "", true}, // AllowEmpty is true by default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		})
	}
}

func TestRFC3339TimeOnlyFormat_ConstantValidation(t *testing.T) {
	// Test that our RFC3339TimeOnlyFormat constant works correctly
	assert.Equal(t, "15:04:05Z07:00", RFC3339TimeOnlyFormat, "RFC3339TimeOnlyFormat constant should match RFC3339 specification")

	// Test parsing with the constant
	checker := NewTimeOnlyChecker()

	// These should work with our RFC3339TimeOnlyFormat
	validTimes := []string{
		"15:04:05Z",      // UTC
		"15:04:05+07:00", // Positive offset
		"15:04:05-07:00", // Negative offset
		"00:00:00Z",      // Midnight UTC
		"23:59:59-12:00", // End of day with max negative offset
	}

	for _, timeStr := range validTimes {
		assert.True(t, checker.IsFormat(timeStr), "Should accept RFC3339 time: %s", timeStr)
	}
}

func TestRFC3339_FullDateTime_Formats(t *testing.T) {
	checker := NewDateTimeFormatChecker()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Full RFC3339 datetime formats
		{"RFC3339 UTC datetime", "2006-01-02T15:04:05Z", true},
		{"RFC3339 with positive timezone", "2006-01-02T15:04:05+07:00", true},
		{"RFC3339 with negative timezone", "2006-01-02T15:04:05-07:00", true},
		{"RFC3339 with nanoseconds UTC", "2006-01-02T15:04:05.999999999Z", true},
		{"RFC3339 with nanoseconds and timezone", "2006-01-02T15:04:05.123456789-05:00", true},

		// Real-world examples
		{"Real datetime UTC", "2023-12-25T10:30:45Z", true},
		{"Real datetime with timezone", "2023-12-25T10:30:45-03:00", true},
		{"Real datetime with milliseconds", "2023-12-25T10:30:45.123Z", true},

		// Edge cases
		{"Leap year date", "2024-02-29T12:00:00Z", true},
		{"New Year UTC", "2024-01-01T00:00:00Z", true},
		{"End of year", "2023-12-31T23:59:59Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		})
	}
}
