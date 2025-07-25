package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringFormatChecker_IsFormat(t *testing.T) {
	checker := NewStringFormatChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid string", "hello", true},
		{"empty string", "", true}, // AllowEmpty is true by default
		{"non-string int", 123, false},
		{"non-string float", 123.45, false},
		{"non-string nil", nil, false},
		{"non-string bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringFormatChecker_NoEmptyString(t *testing.T) {
	checker := NewStringFormatChecker()
	checker.AllowEmpty = false

	assert.True(t, checker.IsFormat("hello"))
	assert.False(t, checker.IsFormat(""))
	assert.False(t, checker.IsFormat(123))
}

func TestNonEmptyStringChecker_IsFormat(t *testing.T) {
	checker := NewNonEmptyStringChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid non-empty string", "hello", true},
		{"empty string", "", false},
		{"whitespace string", "   ", true}, // whitespace is not considered empty
		{"non-string", 123, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNonEmptyStringChecker_Check(t *testing.T) {
	checker := NewNonEmptyStringChecker()

	// Valid data
	errors := checker.Check("hello")
	assert.Empty(t, errors)

	// Invalid data
	errors = checker.Check("")
	require.Len(t, errors, 1)
	assert.Equal(t, "non_empty_string", errors[0].Field)
	assert.Equal(t, "EMPTY_STRING", errors[0].ErrorType)
}

func TestTextMatchChecker_IsFormat(t *testing.T) {
	checker := NewTextMatchChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid text with letters", "hello", true},
		{"valid text with spaces", "hello world", true},
		{"valid text with underscores", "hello_world", true},
		{"valid text mixed", "hello world_test", true},
		{"empty string", "", false}, // AllowEmpty is false by default
		{"text with numbers", "hello123", false},
		{"text with special chars", "hello!", false},
		{"non-string", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTextMatchWithNumberChecker_IsFormat(t *testing.T) {
	checker := NewTextMatchWithNumberChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid text with letters", "hello", true},
		{"valid text with numbers", "hello123", true},
		{"valid text with spaces", "hello world", true},
		{"valid text with underscores", "hello_world", true},
		{"valid text mixed", "hello world_test 123", true},
		{"empty string", "", false},
		{"text with zero", "hello0", false}, // regex excludes 0
		{"text with special chars", "hello!", false},
		{"non-string", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomRegexChecker_IsFormat(t *testing.T) {
	// Test email pattern
	emailChecker, err := NewCustomRegexChecker(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid email", "test@example.com", true},
		{"invalid email", "test@", false},
		{"non-string", 123, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailChecker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomRegexChecker_InvalidPattern(t *testing.T) {
	_, err := NewCustomRegexChecker(`[invalid regex`)
	assert.Error(t, err)
}

func TestStrongNameFormatChecker_IsFormat(t *testing.T) {
	checker := NewStrongNameFormatChecker()

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid uppercase letters", "HELLO", true},
		{"valid with underscores", "HELLO_WORLD", true},
		{"valid underscores only", "___", true},
		{"empty string", "", false},
		{"lowercase letters", "hello", false},
		{"mixed case", "Hello", false},
		{"with numbers", "HELLO123", false},
		{"with spaces", "HELLO WORLD", false},
		{"non-string", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStrongNameFormatChecker_AllowEmpty(t *testing.T) {
	checker := NewStrongNameFormatChecker()
	checker.AllowEmpty = true

	assert.True(t, checker.IsFormat(""))
	assert.True(t, checker.IsFormat("HELLO"))
	assert.False(t, checker.IsFormat("hello"))
}
