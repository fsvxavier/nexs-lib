package schema

import (
	"testing"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		schema      string
		expectedErr error
	}{
		{
			name: "valid complex object",
			data: map[string]interface{}{
				"name": "John Doe",
				"age":  30,
				"address": map[string]interface{}{
					"street": "Main St",
					"number": 123,
				},
				"phones": []string{"1234-5678", "8765-4321"},
			},
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string", "minLength": 1},
					"age": {"type": "integer", "minimum": 0},
					"address": {
						"type": "object",
						"properties": {
							"street": {"type": "string"},
							"number": {"type": "integer"}
						},
						"required": ["street", "number"]
					},
					"phones": {
						"type": "array",
						"items": {"type": "string"}
					}
				},
				"required": ["name", "age"]
			}`,
			expectedErr: nil,
		},
		{
			name: "invalid string format",
			data: map[string]interface{}{
				"date": "invalid-date",
			},
			schema: `{
				"type": "object",
				"properties": {
					"date": {"type": "string", "format": "date_time"}
				}
			}`,
			expectedErr: &domainerrors.InvalidSchemaError{
				Details: map[string][]string{
					"date": {"INVALID_FORMAT"},
				},
			},
		},
		{
			name: "array validation",
			data: map[string]interface{}{
				"items": []interface{}{1, "2", 3},
			},
			schema: `{
				"type": "object",
				"properties": {
					"items": {
						"type": "array",
						"items": {"type": "integer"}
					}
				}
			}`,
			expectedErr: &domainerrors.InvalidSchemaError{
				Details: map[string][]string{
					"items.1": {"INVALID_DATA_TYPE"},
				},
			},
		},
		{
			name: "number range validation",
			data: map[string]interface{}{
				"score": 101,
			},
			schema: `{
				"type": "object",
				"properties": {
					"score": {
						"type": "integer",
						"minimum": 0,
						"maximum": 100
					}
				}
			}`,
			expectedErr: &domainerrors.InvalidSchemaError{
				Details: map[string][]string{
					"score": {"INVALID_VALUE"},
				},
			},
		},
		{
			name: "pattern validation",
			data: map[string]interface{}{
				"code": "123-invalid",
			},
			schema: `{
				"type": "object",
				"properties": {
					"code": {
						"type": "string",
						"pattern": "^[A-Z]{3}-[0-9]{3}$"
					}
				}
			}`,
			expectedErr: &domainerrors.InvalidSchemaError{
				Details: map[string][]string{
					"code": {"INVALID_DATA_TYPE"},
				},
			},
		},
		{
			name: "enum validation",
			data: map[string]interface{}{
				"status": "PENDING",
			},
			schema: `{
				"type": "object",
				"properties": {
					"status": {
						"type": "string",
						"enum": ["ACTIVE", "INACTIVE", "SUSPENDED"]
					}
				}
			}`,
			expectedErr: &domainerrors.InvalidSchemaError{
				Details: map[string][]string{
					"status": {"INVALID_VALUE"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data, tt.schema)
			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestValidateWithInvalidSchema(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
	}
	schema := `{invalid json schema}`

	err := Validate(data, schema)
	assert.Error(t, err)
	assert.NotEqual(t, &domainerrors.InvalidSchemaError{}, err)
}
