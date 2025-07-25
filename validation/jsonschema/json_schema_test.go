package jsonschema

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name:    "should create validator with nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "should create validator with default config",
			config:  config.NewConfig(),
			wantErr: false,
		},
		{
			name:    "should create validator with gojsonschema provider",
			config:  config.NewConfig().WithProvider(config.GoJSONSchemaProvider),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewValidator(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, validator)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, validator)
				assert.NotNil(t, validator.provider)
				assert.NotNil(t, validator.config)
			}
		})
	}
}

func TestJSONSchemaValidator_ValidateFromBytes(t *testing.T) {
	tests := []struct {
		name       string
		schema     []byte
		data       interface{}
		wantErrors int
		wantErr    bool
	}{
		{
			name: "valid data should pass validation",
			schema: []byte(`{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "number"}
				},
				"required": ["name"]
			}`),
			data: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			wantErrors: 0,
			wantErr:    false,
		},
		{
			name: "invalid data should fail validation",
			schema: []byte(`{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "number"}
				},
				"required": ["name"]
			}`),
			data: map[string]interface{}{
				"age": 30,
			},
			wantErrors: 1, // Expecting at least 1 error
			wantErr:    false,
		},
		{
			name: "invalid schema should be handled gracefully",
			schema: []byte(`{
				"type": "object",
				"properties": {
					"name": {"type": "string"}
				}
			}`),
			data: map[string]interface{}{
				"name": "John",
			},
			wantErrors: 0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewValidator(config.NewConfig())
			require.NoError(t, err)

			errors, err := validator.ValidateFromBytes(tt.schema, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantErrors > 0 {
					// For validation errors, we expect at least the specified number
					assert.GreaterOrEqual(t, len(errors), tt.wantErrors)
				} else {
					assert.Len(t, errors, tt.wantErrors)
				}
			}
		})
	}
}

func TestJSONSchemaValidator_ValidateFromStruct(t *testing.T) {
	tests := []struct {
		name       string
		schemaName string
		data       interface{}
		setupFunc  func(*config.Config)
		wantErrors int
		wantErr    bool
	}{
		{
			name:       "should validate with registered schema",
			schemaName: "user-schema",
			data: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			setupFunc: func(cfg *config.Config) {
				schema := map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
						"age":  map[string]interface{}{"type": "number"},
					},
					"required": []string{"name"},
				}
				cfg.RegisterSchema("user-schema", schema)
			},
			wantErrors: 0,
			wantErr:    false,
		},
		{
			name:       "should fail with unregistered schema",
			schemaName: "unknown-schema",
			data: map[string]interface{}{
				"name": "John",
			},
			setupFunc:  func(cfg *config.Config) {},
			wantErrors: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig()
			tt.setupFunc(cfg)

			validator, err := NewValidator(cfg)
			require.NoError(t, err)

			errors, err := validator.ValidateFromStruct(tt.schemaName, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, errors, tt.wantErrors)
			}
		})
	}
}

func TestValidate_LegacyCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		loader       interface{}
		schemaLoader string
		wantErr      bool
	}{
		{
			name: "valid data should not return error",
			loader: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			schemaLoader: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "number"}
				},
				"required": ["name"]
			}`,
			wantErr: false,
		},
		{
			name: "invalid data should return error",
			loader: map[string]interface{}{
				"age": 30,
			},
			schemaLoader: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "number"}
				},
				"required": ["name"]
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.loader, tt.schemaLoader)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddCustomFormat_LegacyCompatibility(t *testing.T) {
	// Test that the function doesn't panic and can be called
	// Detailed testing would require more complex setup
	assert.NotPanics(t, func() {
		AddCustomFormat("test-format", "^[A-Z]+$")
	})
}

func TestCreateProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerType config.ProviderType
		wantName     string
	}{
		{
			name:         "should create gojsonschema provider",
			providerType: config.GoJSONSchemaProvider,
			wantName:     "xeipuuv/gojsonschema",
		},
		{
			name:         "should create kaptinlin provider",
			providerType: config.JSONSchemaProvider,
			wantName:     "kaptinlin/jsonschema",
		},
		{
			name:         "should create kaptinlin provider",
			providerType: config.SchemaJSONProvider,
			wantName:     "santhosh-tekuri/jsonschema-v6",
		},
		{
			name:         "should default to kaptinlin provider",
			providerType: "unknown",
			wantName:     "kaptinlin/jsonschema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := createProvider(tt.providerType)

			assert.NoError(t, err)
			assert.NotNil(t, provider)
			assert.Equal(t, tt.wantName, provider.GetName())
		})
	}
}
