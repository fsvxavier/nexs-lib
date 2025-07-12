package interfaces

import (
	"errors"
	"testing"
	"time"
)

// TestErrorNamedCoverage testa completamente a função ErrorNamed
func TestErrorNamedCoverage(t *testing.T) {
	t.Run("ErrorNamed with nil error", func(t *testing.T) {
		field := ErrorNamed("test_error", nil)

		if field.Key != "test_error" {
			t.Errorf("Expected key 'test_error', got %s", field.Key)
		}

		if field.Value != nil {
			t.Errorf("Expected nil value, got %v", field.Value)
		}

		if field.Type != ErrorType {
			t.Errorf("Expected ErrorType, got %v", field.Type)
		}
	})

	t.Run("ErrorNamed with actual error", func(t *testing.T) {
		testErr := errors.New("test error message")
		field := ErrorNamed("actual_error", testErr)

		if field.Key != "actual_error" {
			t.Errorf("Expected key 'actual_error', got %s", field.Key)
		}

		if field.Value != testErr.Error() {
			t.Errorf("Expected error value, got %v", field.Value)
		}

		if field.Type != ErrorType {
			t.Errorf("Expected ErrorType, got %v", field.Type)
		}
	})
}

// TestAllFieldTypesCreation testa criação de campos de todos os tipos para cobertura completa
func TestAllFieldTypesCreation(t *testing.T) {
	t.Run("All primitive field types", func(t *testing.T) {
		fields := []Field{
			String("str", "value"),
			Int("int", 42),
			Int64("int64", int64(42)),
			Float64("float64", 3.14),
			Bool("bool", true),
			Time("time", time.Now()),
			Duration("duration", time.Second),
			Object("object", map[string]interface{}{"key": "value"}),
			Array("array", []interface{}{1, 2, 3}),
		}

		expectedTypes := []FieldType{
			StringType, IntType, Int64Type, Float64Type, BoolType,
			TimeType, DurationType, ObjectType, ArrayType,
		}

		for i, field := range fields {
			if field.Type != expectedTypes[i] {
				t.Errorf("Field %d: expected type %v, got %v", i, expectedTypes[i], field.Type)
			}
		}
	})

	t.Run("Error field with non-nil error", func(t *testing.T) {
		testErr := errors.New("test error")
		field := Error(testErr)

		if field.Key != "error" {
			t.Errorf("Expected key 'error', got %s", field.Key)
		}

		if field.Value != testErr.Error() {
			t.Errorf("Expected error value, got %v", field.Value)
		}

		if field.Type != ErrorType {
			t.Errorf("Expected ErrorType, got %v", field.Type)
		}
	})

	t.Run("Error field with nil error", func(t *testing.T) {
		field := Error(nil)

		if field.Key != "error" {
			t.Errorf("Expected key 'error', got %s", field.Key)
		}

		if field.Value != nil {
			t.Errorf("Expected nil value, got %v", field.Value)
		}

		if field.Type != ErrorType {
			t.Errorf("Expected ErrorType, got %v", field.Type)
		}
	})
}

// TestPredefinedFieldsComplete testa todos os campos predefinidos
func TestPredefinedFieldsComplete(t *testing.T) {
	testCases := []struct {
		name     string
		field    Field
		value    interface{}
		expected string
	}{
		{"TraceID", TraceID("trace-123"), "trace-123", "trace_id"},
		{"SpanID", SpanID("span-456"), "span-456", "span_id"},
		{"UserID", UserID("user-789"), "user-789", "user_id"},
		{"RequestID", RequestID("req-abc"), "req-abc", "request_id"},
		{"CorrelationID", CorrelationID("corr-def"), "corr-def", "correlation_id"},
		{"Method", Method("POST"), "POST", "method"},
		{"Path", Path("/api/users"), "/api/users", "path"},
		{"StatusCode", StatusCode(200), 200, "status_code"},
		{"Latency", Latency(time.Duration(100500)), time.Duration(100500), "latency"},
		{"Operation", Operation("create_user"), "create_user", "operation"},
		{"Component", Component("auth-service"), "auth-service", "component"},
		{"Version", Version("v1.2.3"), "v1.2.3", "version"},
		{"Environment", Environment("production"), "production", "environment"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.field.Key != tc.expected {
				t.Errorf("Expected key '%s', got '%s'", tc.expected, tc.field.Key)
			}

			if tc.field.Value != tc.value {
				t.Errorf("Expected value %v, got %v", tc.value, tc.field.Value)
			}
		})
	}
}

// TestValidationFunctions testa as funções de validação
func TestValidationFunctions(t *testing.T) {
	t.Run("ValidateLevel edge cases", func(t *testing.T) {
		validLevels := []Level{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, PanicLevel}
		for _, level := range validLevels {
			if err := ValidateLevel(level); err != nil {
				t.Errorf("Level %v should be valid, got error: %v", level, err)
			}
		}

		// Teste com nível inválido
		invalidLevel := Level(100)
		if err := ValidateLevel(invalidLevel); err == nil {
			t.Error("Invalid level should return error")
		}
	})

	t.Run("ValidateFormat edge cases", func(t *testing.T) {
		validFormats := []Format{JSONFormat, TextFormat, ConsoleFormat}
		for _, format := range validFormats {
			if err := ValidateFormat(format); err != nil {
				t.Errorf("Format %v should be valid, got error: %v", format, err)
			}
		}

		// Teste com formato inválido
		invalidFormat := Format(100)
		if err := ValidateFormat(invalidFormat); err == nil {
			t.Error("Invalid format should return error")
		}
	})

	t.Run("ValidateConfig comprehensive", func(t *testing.T) {
		// Configuração válida
		validConfig := Config{
			Level:       InfoLevel,
			Format:      JSONFormat,
			ServiceName: "test-service",
		}

		if err := ValidateConfig(validConfig); err != nil {
			t.Errorf("Valid config should not return error, got: %v", err)
		}

		// Configuração com nível inválido
		invalidLevelConfig := validConfig
		invalidLevelConfig.Level = Level(100)
		if err := ValidateConfig(invalidLevelConfig); err == nil {
			t.Error("Config with invalid level should return error")
		}

		// Configuração com formato inválido
		invalidFormatConfig := validConfig
		invalidFormatConfig.Format = Format(100)
		if err := ValidateConfig(invalidFormatConfig); err == nil {
			t.Error("Config with invalid format should return error")
		}

		// Configuração com service name vazio
		emptyServiceConfig := validConfig
		emptyServiceConfig.ServiceName = ""
		if err := ValidateConfig(emptyServiceConfig); err == nil {
			t.Error("Config with empty service name should return error")
		}

		// Configuração com async inválida
		invalidAsyncConfig := validConfig
		invalidAsyncConfig.Async = &AsyncConfig{
			Enabled:    true,
			BufferSize: -1, // Inválido
		}
		if err := ValidateConfig(invalidAsyncConfig); err == nil {
			t.Error("Config with invalid async config should return error")
		}

		// Configuração com workers inválidos
		invalidWorkersConfig := validConfig
		invalidWorkersConfig.Async = &AsyncConfig{
			Enabled:    true,
			BufferSize: 1024,
			Workers:    0, // Inválido
		}
		if err := ValidateConfig(invalidWorkersConfig); err == nil {
			t.Error("Config with invalid workers should return error")
		}

		// Configuração com sampling inválida - initial
		invalidSamplingInitial := validConfig
		invalidSamplingInitial.Sampling = &SamplingConfig{
			Enabled: true,
			Initial: -1, // Inválido
		}
		if err := ValidateConfig(invalidSamplingInitial); err == nil {
			t.Error("Config with invalid sampling initial should return error")
		}

		// Configuração com sampling inválida - thereafter
		invalidSamplingThereafter := validConfig
		invalidSamplingThereafter.Sampling = &SamplingConfig{
			Enabled:    true,
			Initial:    10,
			Thereafter: -1, // Inválido
		}
		if err := ValidateConfig(invalidSamplingThereafter); err == nil {
			t.Error("Config with invalid sampling thereafter should return error")
		}
	})
}

// TestLevelStringConversion testa conversão de Level para string de forma mais abrangente
func TestLevelStringConversion(t *testing.T) {
	testCases := []struct {
		level    Level
		expected string
	}{
		{TraceLevel, "TRACE"},
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{PanicLevel, "PANIC"},
		{Level(100), "UNKNOWN"}, // Nível inválido
		{Level(-10), "UNKNOWN"}, // Nível inválido
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.level.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestFormatStringConversion testa conversão de Format para string de forma mais abrangente
func TestFormatStringConversion(t *testing.T) {
	testCases := []struct {
		format   Format
		expected string
	}{
		{JSONFormat, "json"},
		{TextFormat, "text"},
		{ConsoleFormat, "console"},
		{Format(100), "unknown"}, // Formato inválido
		{Format(-1), "unknown"},  // Formato inválido
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.format.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}
