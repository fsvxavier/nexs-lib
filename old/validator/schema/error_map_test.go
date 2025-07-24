package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultErrors(t *testing.T) {
	tests := []struct {
		name     string
		errorKey string
		expected string
	}{
		{
			name:     "required error message",
			errorKey: "required",
			expected: "REQUIRED_ATTRIBUTE_MISSING",
		},
		{
			name:     "invalid type error message",
			errorKey: "invalid_type",
			expected: "INVALID_DATA_TYPE",
		},
		{
			name:     "enum error message",
			errorKey: "enum",
			expected: "INVALID_VALUE",
		},
		{
			name:     "format error message",
			errorKey: "format",
			expected: "INVALID_FORMAT",
		},
		{
			name:     "string length error message",
			errorKey: "string_gte",
			expected: "INVALID_LENGTH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, exists := defaultErrors[tt.errorKey]
			assert.True(t, exists, "Error key should exist in defaultErrors map")
			assert.Equal(t, tt.expected, result)
		})
	}
}
