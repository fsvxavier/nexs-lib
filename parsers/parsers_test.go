package parsers

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Error("Expected factory to be created")
	}
	if factory.config == nil {
		t.Error("Expected factory config to be initialized")
	}
}

func TestNewFactoryWithConfig(t *testing.T) {
	config := &interfaces.ParserConfig{
		StrictMode: false,
	}
	factory := NewFactoryWithConfig(config)
	if factory == nil {
		t.Error("Expected factory to be created")
	}
	if factory.config != config {
		t.Error("Expected factory to use provided config")
	}
}

func TestFactory_JSON(t *testing.T) {
	factory := NewFactory()
	parser := factory.JSON()
	if parser == nil {
		t.Error("Expected JSON parser to be created")
	}
}

func TestFactory_URL(t *testing.T) {
	factory := NewFactory()
	parser := factory.URL()
	if parser == nil {
		t.Error("Expected URL parser to be created")
	}
}

func TestFactory_Datetime(t *testing.T) {
	factory := NewFactory()
	parser := factory.Datetime()
	if parser == nil {
		t.Error("Expected Datetime parser to be created")
	}
}

func TestFactory_Duration(t *testing.T) {
	factory := NewFactory()
	parser := factory.Duration()
	if parser == nil {
		t.Error("Expected Duration parser to be created")
	}
}

func TestFactory_Env(t *testing.T) {
	factory := NewFactory()
	parser := factory.Env()
	if parser == nil {
		t.Error("Expected Env parser to be created")
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Error("Expected manager to be created")
	}
	if manager.factory == nil {
		t.Error("Expected manager factory to be initialized")
	}
}

func TestNewManagerWithConfig(t *testing.T) {
	config := &interfaces.ParserConfig{
		StrictMode: false,
	}
	manager := NewManagerWithConfig(config)
	if manager == nil {
		t.Error("Expected manager to be created")
	}
	if manager.factory == nil {
		t.Error("Expected manager factory to be initialized")
	}
}

func TestManager_ParseJSON(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid JSON",
			input:   []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{"key": value}`),
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.ParseJSON(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected result to not be nil")
			}
			if !tt.wantErr && result.Metadata == nil {
				t.Error("Expected metadata to not be nil")
			}
		})
	}
}

func TestManager_ParseURL(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid URL",
			input:   "https://example.com/path",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			input:   "not-a-url",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.ParseURL(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("Expected result to not be nil")
			}
			if !tt.wantErr && result.Metadata == nil {
				t.Error("Expected metadata to not be nil")
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Error("Expected validator to be created")
	}
}

func TestValidator_ValidateJSON(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid JSON",
			input:   []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{"key": value}`),
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateJSONString(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			input:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"key": value}`,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateJSONString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSONString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateURL(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid URL",
			input:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			input:   "not-a-url",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTransformer(t *testing.T) {
	transformer := NewTransformer()
	if transformer == nil {
		t.Error("Expected transformer to be created")
	}
}

func TestTransformer_CompactJSON(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid JSON with spaces",
			input:   `{ "key" : "value" }`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"key": value}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.CompactJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompactJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("Expected result to not be empty")
			}
		})
	}
}

func TestTransformer_PrettyJSON(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			input:   `{"key":"value"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"key": value}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.PrettyJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrettyJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("Expected result to not be empty")
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("ParseJSON", func(t *testing.T) {
		result, err := ParseJSON(map[string]interface{}{"key": "value"})
		if err != nil {
			t.Errorf("ParseJSON() error = %v", err)
		}
		if result == nil {
			t.Error("Expected result to not be nil")
		}
	})

	t.Run("ParseJSONString", func(t *testing.T) {
		result, err := ParseJSONString(`{"key": "value"}`)
		if err != nil {
			t.Errorf("ParseJSONString() error = %v", err)
		}
		if result == nil {
			t.Error("Expected result to not be nil")
		}
	})

	t.Run("ParseJSONBytes", func(t *testing.T) {
		result, err := ParseJSONBytes([]byte(`{"key": "value"}`))
		if err != nil {
			t.Errorf("ParseJSONBytes() error = %v", err)
		}
		if result == nil {
			t.Error("Expected result to not be nil")
		}
	})

	t.Run("ParseURLString", func(t *testing.T) {
		result, err := ParseURLString("https://example.com")
		if err != nil {
			t.Errorf("ParseURLString() error = %v", err)
		}
		if result == nil {
			t.Error("Expected result to not be nil")
		}
	})

	t.Run("ValidateJSONData", func(t *testing.T) {
		err := ValidateJSONData(map[string]interface{}{"key": "value"})
		if err != nil {
			t.Errorf("ValidateJSONData() error = %v", err)
		}
	})

	t.Run("ValidateJSONStr", func(t *testing.T) {
		err := ValidateJSONStr(`{"key": "value"}`)
		if err != nil {
			t.Errorf("ValidateJSONStr() error = %v", err)
		}
	})

	t.Run("ValidateURLStr", func(t *testing.T) {
		err := ValidateURLStr("https://example.com")
		if err != nil {
			t.Errorf("ValidateURLStr() error = %v", err)
		}
	})
}
