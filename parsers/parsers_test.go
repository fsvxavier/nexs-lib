package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	assert.Equal(t, "1.0.0", Version)
	assert.Equal(t, "nexs-lib/parsers", Name)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.NotNil(t, config.DefaultLocation)
	assert.False(t, config.StrictMode)
	assert.True(t, config.IgnoreCase)
	assert.False(t, config.AllowPartial)
	assert.Equal(t, DateOrderMDY, config.DateOrder)
	assert.NotNil(t, config.CustomUnits)
}

func TestOptions(t *testing.T) {
	config := DefaultConfig()

	// Test WithStrictMode
	WithStrictMode(true).Apply(config)
	assert.True(t, config.StrictMode)

	// Test WithIgnoreCase
	WithIgnoreCase(false).Apply(config)
	assert.False(t, config.IgnoreCase)

	// Test WithDateOrder
	WithDateOrder(DateOrderDMY).Apply(config)
	assert.Equal(t, DateOrderDMY, config.DateOrder)

	// Test WithCustomFormats
	WithCustomFormats("format1", "format2").Apply(config)
	assert.Contains(t, config.CustomFormats, "format1")
	assert.Contains(t, config.CustomFormats, "format2")
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  string
	}{
		{"invalid format", ErrorTypeInvalidFormat, "invalid_format"},
		{"invalid value", ErrorTypeInvalidValue, "invalid_value"},
		{"unsupported type", ErrorTypeUnsupportedType, "unsupported_type"},
		{"not found", ErrorTypeNotFound, "not_found"},
		{"validation", ErrorTypeValidation, "validation"},
		{"timeout", ErrorTypeTimeout, "timeout"},
		{"internal", ErrorTypeInternal, "internal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.errorType))
		})
	}
}

func TestDateOrder(t *testing.T) {
	assert.Equal(t, DateOrder(0), DateOrderMDY)
	assert.Equal(t, DateOrder(1), DateOrderDMY)
	assert.Equal(t, DateOrder(2), DateOrderYMD)
}
