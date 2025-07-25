package config

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		want func(*Config) bool
	}{
		{
			name: "default config should use JSONSchema provider",
			want: func(c *Config) bool {
				return c.Provider == JSONSchemaProvider
			},
		},
		{
			name: "default config should not be in strict mode",
			want: func(c *Config) bool {
				return !c.StrictMode
			},
		},
		{
			name: "default config should have empty hooks",
			want: func(c *Config) bool {
				return len(c.PreValidationHooks) == 0 &&
					len(c.PostValidationHooks) == 0 &&
					len(c.ErrorHooks) == 0 &&
					len(c.AdditionalChecks) == 0
			},
		},
		{
			name: "default config should have error mapping",
			want: func(c *Config) bool {
				return len(c.ErrorMapping) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			assert.True(t, tt.want(config))
		})
	}
}

func TestConfig_WithProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		want     ProviderType
	}{
		{
			name:     "set GoJSONSchema provider",
			provider: GoJSONSchemaProvider,
			want:     GoJSONSchemaProvider,
		},
		{
			name:     "set JSONSchema provider",
			provider: JSONSchemaProvider,
			want:     JSONSchemaProvider,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig().WithProvider(tt.provider)
			assert.Equal(t, tt.want, config.Provider)
		})
	}
}

func TestConfig_WithStrictMode(t *testing.T) {
	tests := []struct {
		name   string
		strict bool
		want   bool
	}{
		{
			name:   "enable strict mode",
			strict: true,
			want:   true,
		},
		{
			name:   "disable strict mode",
			strict: false,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig().WithStrictMode(tt.strict)
			assert.Equal(t, tt.want, config.StrictMode)
		})
	}
}

func TestConfig_RegisterSchema(t *testing.T) {
	config := NewConfig()
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}

	result := config.RegisterSchema("test-schema", schema)

	assert.Equal(t, config, result) // Should return self for chaining
	assert.Equal(t, schema, config.SchemaRegistry["test-schema"])
}

func TestConfig_SetErrorMapping(t *testing.T) {
	config := NewConfig()
	customMapping := map[string]string{
		"custom_error": "CUSTOM_ERROR_TYPE",
	}

	result := config.SetErrorMapping(customMapping)

	assert.Equal(t, config, result) // Should return self for chaining
	assert.Equal(t, customMapping, config.ErrorMapping)
}

func TestConfig_AddHooks(t *testing.T) {
	config := NewConfig()

	// Mock hooks for testing
	preHook := &mockPreValidationHook{}
	postHook := &mockPostValidationHook{}
	errorHook := &mockErrorHook{}
	check := &mockCheck{}

	config.AddPreValidationHook(preHook)
	config.AddPostValidationHook(postHook)
	config.AddErrorHook(errorHook)
	config.AddCheck(check)

	assert.Len(t, config.PreValidationHooks, 1)
	assert.Len(t, config.PostValidationHooks, 1)
	assert.Len(t, config.ErrorHooks, 1)
	assert.Len(t, config.AdditionalChecks, 1)
}

func TestConfig_AddCustomFormat(t *testing.T) {
	config := NewConfig()
	checker := &mockFormatChecker{}

	result := config.AddCustomFormat("test-format", checker)

	assert.Equal(t, config, result) // Should return self for chaining
	assert.Equal(t, checker, config.CustomFormats["test-format"])
}

func TestGetDefaultErrorMapping(t *testing.T) {
	mapping := getDefaultErrorMapping()

	// Test some expected mappings
	expectedMappings := map[string]string{
		"required":     "REQUIRED_ATTRIBUTE_MISSING",
		"invalid_type": "INVALID_DATA_TYPE",
		"enum":         "INVALID_VALUE",
		"format":       "INVALID_FORMAT",
		"string_gte":   "INVALID_LENGTH",
	}

	for key, expectedValue := range expectedMappings {
		actualValue, exists := mapping[key]
		require.True(t, exists, "Key %s should exist in default error mapping", key)
		assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
	}

	// Ensure mapping is not empty
	assert.Greater(t, len(mapping), 10, "Default error mapping should have multiple entries")
}

// Mock implementations for testing

type mockPreValidationHook struct{}

func (m *mockPreValidationHook) Execute(data interface{}) (interface{}, error) {
	return data, nil
}

type mockPostValidationHook struct{}

func (m *mockPostValidationHook) Execute(data interface{}, errors []interfaces.ValidationError) ([]interfaces.ValidationError, error) {
	return errors, nil
}

type mockErrorHook struct{}

func (m *mockErrorHook) Execute(errors []interfaces.ValidationError) []interfaces.ValidationError {
	return errors
}

type mockCheck struct{}

func (m *mockCheck) Validate(data interface{}) []interfaces.ValidationError {
	return nil
}

func (m *mockCheck) GetName() string {
	return "mock-check"
}

type mockFormatChecker struct{}

func (m *mockFormatChecker) IsFormat(input interface{}) bool {
	return true
}
