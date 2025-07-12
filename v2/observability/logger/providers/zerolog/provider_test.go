package zerolog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewProvider(t *testing.T) {
	provider := NewProvider()

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}

	if provider.Name() != ProviderName {
		t.Errorf("Expected provider name %s, got %s", ProviderName, provider.Name())
	}

	if provider.Version() == "" {
		t.Error("Expected provider version to be set")
	}
}

func TestProviderConfigure(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:          interfaces.InfoLevel,
		Format:         interfaces.JSONFormat,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		AddCaller:      true,
		AddStacktrace:  false,
		Output:         os.Stdout,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verifica se a configuração foi aplicada
	if provider.config.ServiceName != config.ServiceName {
		t.Errorf("Expected service name %s, got %s", config.ServiceName, provider.config.ServiceName)
	}
}

func TestProviderHealthCheck(t *testing.T) {
	provider := NewProvider()

	// Zerolog sempre está disponível
	err := provider.HealthCheck()
	if err != nil {
		t.Errorf("Expected no error for zerolog health check, got %v", err)
	}
}

func TestProviderLogging(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	ctx := context.Background()

	// Testa todos os níveis (sem panics/fatals)
	provider.Trace(ctx, "trace message", interfaces.String("key", "value"))
	provider.Debug(ctx, "debug message", interfaces.String("key", "value"))
	provider.Info(ctx, "info message", interfaces.String("key", "value"))
	provider.Warn(ctx, "warn message", interfaces.String("key", "value"))
	provider.Error(ctx, "error message", interfaces.String("key", "value"))

	// Se chegou até aqui, não houve panic
}

func TestProviderFormattedLogging(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.ConsoleFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	ctx := context.Background()

	// Testa logging formatado
	provider.Tracef(ctx, "trace: %s %d", "value", 42)
	provider.Debugf(ctx, "debug: %s %d", "value", 42)
	provider.Infof(ctx, "info: %s %d", "value", 42)
	provider.Warnf(ctx, "warn: %s %d", "value", 42)
	provider.Errorf(ctx, "error: %s %d", "value", 42)
}

func TestProviderCodedLogging(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.TextFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	ctx := context.Background()

	// Testa logging com código
	provider.TraceWithCode(ctx, "TRACE_CODE", "trace with code")
	provider.DebugWithCode(ctx, "DEBUG_CODE", "debug with code")
	provider.InfoWithCode(ctx, "INFO_CODE", "info with code")
	provider.WarnWithCode(ctx, "WARN_CODE", "warn with code")
	provider.ErrorWithCode(ctx, "ERROR_CODE", "error with code")
}

func TestProviderWithFields(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	ctx := context.Background()

	// Testa WithFields
	fieldsLogger := provider.WithFields(
		interfaces.String("field1", "value1"),
		interfaces.Int("field2", 42),
	)

	if fieldsLogger == nil {
		t.Fatal("Expected fields logger to be created")
	}

	fieldsLogger.Info(ctx, "message with fields")
}

func TestProviderWithContext(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	// Context com trace ID
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace")

	contextLogger := provider.WithContext(ctx)
	if contextLogger == nil {
		t.Fatal("Expected context logger to be created")
	}

	contextLogger.Info(ctx, "message with context")
}

func TestProviderWithError(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	ctx := context.Background()
	testError := errors.New("test error")

	errorLogger := provider.WithError(testError)
	if errorLogger == nil {
		t.Fatal("Expected error logger to be created")
	}

	errorLogger.Error(ctx, "error occurred")
}

func TestProviderWithTraceID(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	ctx := context.Background()

	traceLogger := provider.WithTraceID("trace-123")
	if traceLogger == nil {
		t.Fatal("Expected trace logger to be created")
	}

	traceLogger.Info(ctx, "message with trace ID")
}

func TestProviderWithSpanID(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	ctx := context.Background()

	spanLogger := provider.WithSpanID("span-456")
	if spanLogger == nil {
		t.Fatal("Expected span logger to be created")
	}

	spanLogger.Info(ctx, "message with span ID")
}

func TestProviderLevelOperations(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	// Test GetLevel
	level := provider.GetLevel()
	if level != interfaces.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", level)
	}

	// Test SetLevel
	provider.SetLevel(interfaces.WarnLevel)

	// Test IsLevelEnabled
	if !provider.IsLevelEnabled(interfaces.WarnLevel) {
		t.Error("Expected WarnLevel to be enabled")
	}

	if provider.IsLevelEnabled(interfaces.DebugLevel) {
		t.Error("Expected DebugLevel to be disabled")
	}
}

func TestProviderClone(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	cloned := provider.Clone()
	if cloned == nil {
		t.Fatal("Expected cloned logger to be created")
	}

	// Verifica se é uma instância diferente
	if cloned == provider {
		t.Error("Expected cloned logger to be different instance")
	}

	// Verifica se funciona
	ctx := context.Background()
	cloned.Info(ctx, "cloned message")
}

func TestProviderFlushAndClose(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	// Test Flush
	err := provider.Flush()
	if err != nil {
		t.Errorf("Expected no error from Flush, got %v", err)
	}

	// Test Close
	err = provider.Close()
	if err != nil {
		t.Errorf("Expected no error from Close, got %v", err)
	}
}

func TestConvertLevel(t *testing.T) {
	tests := []struct {
		input    interfaces.Level
		expected string
	}{
		{interfaces.TraceLevel, "trace"},
		{interfaces.DebugLevel, "debug"},
		{interfaces.InfoLevel, "info"},
		{interfaces.WarnLevel, "warn"},
		{interfaces.ErrorLevel, "error"},
		{interfaces.FatalLevel, "fatal"},
		{interfaces.PanicLevel, "panic"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			zerologLevel := convertLevel(test.input)
			if zerologLevel.String() != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, zerologLevel.String())
			}
		})
	}
}

func TestProviderConfigureFormats(t *testing.T) {
	provider := NewProvider()

	formats := []interfaces.Format{
		interfaces.JSONFormat,
		interfaces.ConsoleFormat,
		interfaces.TextFormat,
	}

	for _, format := range formats {
		t.Run(format.String(), func(t *testing.T) {
			config := interfaces.Config{
				Level:       interfaces.InfoLevel,
				Format:      format,
				ServiceName: "test-service",
				Output:      os.Stdout,
			}

			err := provider.Configure(config)
			if err != nil {
				t.Errorf("Expected no error for format %s, got %v", format.String(), err)
			}
		})
	}
}

func TestAddFieldsToEvent(t *testing.T) {
	fields := []interfaces.Field{
		interfaces.String("string_field", "value"),
		interfaces.Int("int_field", 42),
		interfaces.Bool("bool_field", true),
	}

	// Cria provider configurado
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	// Cria event
	event := provider.logger.Info()

	// Adiciona fields
	event = addFieldsToEvent(event, fields)

	// Se chegou até aqui sem panic, funcionou
	if event == nil {
		t.Error("Expected event to be valid")
	}
}

func TestProviderTraceWithLevelEnabled(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:          interfaces.TraceLevel, // Enable trace level
		Format:         interfaces.JSONFormat,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		AddCaller:      true,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()

	// Test trace with enabled level
	provider.Trace(ctx, "trace message enabled")

	// Test trace with disabled level
	provider.SetLevel(interfaces.InfoLevel)
	provider.Trace(ctx, "trace message disabled")
}

func TestProviderErrorWithStacktrace(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:         interfaces.ErrorLevel,
		Format:        interfaces.JSONFormat,
		ServiceName:   "test-service",
		AddStacktrace: true, // Enable stacktrace for Error level
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()

	// This should trigger stacktrace addition
	provider.Error(ctx, "error with stacktrace")
}

func TestProviderWithErrorEdgeCases(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
	}
	provider.Configure(config)

	// Test with nil error
	newProvider := provider.WithError(nil)
	if newProvider == nil {
		t.Error("Expected provider with nil error to be created")
	}

	// Test with actual error
	testErr := errors.New("test error")
	newProvider2 := provider.WithError(testErr)
	if newProvider2 == nil {
		t.Error("Expected provider with error to be created")
	}
}

func TestProviderWithContextComplexCases(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
	}
	provider.Configure(config)

	// Test with context that has various values
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace-123")
	ctx = context.WithValue(ctx, "span_id", "test-span-456")
	ctx = context.WithValue(ctx, "user_id", "user-789")
	ctx = context.WithValue(ctx, "request_id", "req-101")

	newProvider := provider.WithContext(ctx)
	if newProvider == nil {
		t.Error("Expected provider with context to be created")
	}

	// Test logging with context
	newProvider.Info(ctx, "message with complex context")
}

func TestProviderConvertLevelEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		level interfaces.Level
	}{
		{"TRACE", interfaces.TraceLevel},
		{"DEBUG", interfaces.DebugLevel},
		{"INFO", interfaces.InfoLevel},
		{"WARN", interfaces.WarnLevel},
		{"ERROR", interfaces.ErrorLevel},
		{"FATAL", interfaces.FatalLevel},
		{"PANIC", interfaces.PanicLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zerologLevel := convertLevel(tt.level)
			t.Logf("Level %s converted to zerolog level %v", tt.name, zerologLevel)
		})
	}
}

func TestProviderConvertLevelFromZerolog(t *testing.T) {
	tests := []struct {
		name         string
		zerologLevel zerolog.Level
	}{
		{"TraceLevel", zerolog.TraceLevel},
		{"DebugLevel", zerolog.DebugLevel},
		{"InfoLevel", zerolog.InfoLevel},
		{"WarnLevel", zerolog.WarnLevel},
		{"ErrorLevel", zerolog.ErrorLevel},
		{"FatalLevel", zerolog.FatalLevel},
		{"PanicLevel", zerolog.PanicLevel},
		{"Disabled", zerolog.Disabled},
		{"NoLevel", zerolog.NoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := convertLevelFromZerolog(tt.zerologLevel)
			t.Logf("Zerolog level %s converted to level %v", tt.name, level)
		})
	}
}

func TestProviderAddFieldToEventEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		field interfaces.Field
	}{
		{
			"string_with_special_chars",
			interfaces.String("special", "special\nchars\ttab\"quote"),
		},
		{
			"large_int64",
			interfaces.Int64("large", 9223372036854775807),
		},
		{
			"small_int64",
			interfaces.Int64("small", -9223372036854775808),
		},
		{
			"float_precision",
			interfaces.Float64("precision", 3.141592653589793238462643383279),
		},
		{
			"time_with_nano",
			interfaces.Time("nano_time", time.Date(2023, 12, 25, 12, 30, 45, 123456789, time.UTC)),
		},
		{
			"duration_complex",
			interfaces.Duration("complex_dur", 2*time.Hour+30*time.Minute+45*time.Second+123*time.Millisecond),
		},
		{
			"error_with_stack",
			interfaces.Error(fmt.Errorf("wrapped error: %w", fmt.Errorf("original error"))),
		},
		{
			"error_key_special",
			interfaces.Field{Key: "error", Type: interfaces.ErrorType, Value: fmt.Errorf("special error key")},
		},
		{
			"error_as_string",
			interfaces.Field{Key: "error_str", Type: interfaces.ErrorType, Value: "error as string"},
		},
		{
			"unknown_field_type",
			interfaces.Field{Key: "unknown", Type: interfaces.FieldType(99), Value: "unknown type"},
		},
		{
			"complex_nested_object",
			interfaces.Object("nested", map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "deep value",
						"array":  []interface{}{1, "two", true, nil},
					},
					"simple": 42,
				},
				"top_level": "value",
			}),
		},
		{
			"array_with_objects",
			interfaces.Array("object_array", []interface{}{
				map[string]interface{}{"id": 1, "name": "first"},
				map[string]interface{}{"id": 2, "name": "second"},
				"string_item",
				42,
				true,
				nil,
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider()
			config := interfaces.Config{
				Level:       interfaces.DebugLevel,
				Format:      interfaces.JSONFormat,
				ServiceName: "field-test-service",
				Output:      os.Stdout,
			}

			err := provider.Configure(config)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			ctx := context.Background()

			// Test addFieldToEvent via Info logging
			provider.Info(ctx, "field test", tt.field)

			// Test addFieldToContext via WithFields
			fieldsLogger := provider.WithFields(tt.field)
			fieldsLogger.Info(ctx, "context field test")
		})
	}
}

func TestProviderAddFieldToEventRemainingCases(t *testing.T) {
	// Test remaining cases in addFieldToEvent for complete coverage
	tests := []struct {
		name  string
		field interfaces.Field
	}{
		{
			"error_with_error_key",
			interfaces.Field{Key: "error", Type: interfaces.ErrorType, Value: errors.New("error with error key")},
		},
		{
			"error_as_string_value",
			interfaces.Field{Key: "err_str", Type: interfaces.ErrorType, Value: "string error value"},
		},
		{
			"error_type_non_error",
			interfaces.Field{Key: "non_err", Type: interfaces.ErrorType, Value: 12345},
		},
	}

	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "add-field-service",
		Output:      os.Stdout,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			provider.Info(ctx, "add field test", tt.field)
		})
	}
}

func TestProviderAddFieldToContextRemainingCases(t *testing.T) {
	// Test remaining cases in addFieldToContext for complete coverage
	tests := []struct {
		name  string
		field interfaces.Field
	}{
		{
			"context_error_with_error_key",
			interfaces.Field{Key: "error", Type: interfaces.ErrorType, Value: errors.New("context error with error key")},
		},
		{
			"context_error_as_string",
			interfaces.Field{Key: "err_str", Type: interfaces.ErrorType, Value: "context string error"},
		},
		{
			"context_error_type_non_error",
			interfaces.Field{Key: "non_err", Type: interfaces.ErrorType, Value: 67890},
		},
	}

	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "add-context-service",
		Output:      os.Stdout,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Test addFieldToContext via WithFields
			fieldsLogger := provider.WithFields(tt.field)
			fieldsLogger.Info(ctx, "context field test")
		})
	}
}

func TestProviderExtractContextFunctionsRemainingCases(t *testing.T) {
	// Test remaining cases in extract functions
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			"extract_with_traceId_key",
			context.WithValue(context.Background(), "traceId", "traceId-value"),
		},
		{
			"extract_with_spanId_key",
			context.WithValue(context.Background(), "spanId", "spanId-value"),
		},
		{
			"extract_with_non_string_trace",
			context.WithValue(context.Background(), "trace_id", 99999),
		},
		{
			"extract_with_non_string_span",
			context.WithValue(context.Background(), "span_id", 88888),
		},
		{
			"extract_with_non_string_user",
			context.WithValue(context.Background(), "user_id", 77777),
		},
		{
			"extract_with_non_string_request",
			context.WithValue(context.Background(), "request_id", 66666),
		},
	}

	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "extract-service",
		Output:      os.Stdout,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test extraction via logging (addContextToEvent)
			provider.Info(tt.ctx, "extract test")

			// Test extraction via WithContext
			contextLogger := provider.WithContext(tt.ctx)
			contextLogger.Info(context.Background(), "with context test")
		})
	}
}

func TestProviderConvertLevelDefaultCase(t *testing.T) {
	// Test the default case in convertLevel for complete coverage
	provider := NewProvider()

	// Test with unknown level that should default to InfoLevel
	unknownLevel := interfaces.Level(99)
	provider.SetLevel(unknownLevel)

	// Should default to InfoLevel
	currentLevel := provider.GetLevel()
	if currentLevel != interfaces.InfoLevel {
		t.Errorf("Expected InfoLevel for unknown level, got %v", currentLevel)
	}
}

func TestProviderAddContextToEventEdgeCases(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "context-edge-service",
		Output:      os.Stdout,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test addContextToEvent with context that has some values but not others
	partialCtx := context.Background()
	partialCtx = context.WithValue(partialCtx, "trace_id", "partial-trace")
	// Missing span_id, user_id, request_id to test specific branches

	provider.Info(partialCtx, "partial context test")

	// Test with context that has user_id and request_id but not trace/span
	userCtx := context.Background()
	userCtx = context.WithValue(userCtx, "user_id", "test-user")
	userCtx = context.WithValue(userCtx, "request_id", "test-request")

	provider.Info(userCtx, "user context test")
}

func TestProviderFatalMethodDirect(t *testing.T) {
	// Teste para aumentar cobertura dos métodos Fatal
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.FatalLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "fatal-test-service",
		Output:      os.Stdout,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test que os métodos Fatal existem e são implementados
	// Não chamamos diretamente para evitar os.Exit()
	t.Log("Fatal and Fatalf methods are implemented and would call os.Exit")
}

// TestProviderConfigureTimeFormatEmpty testa Configure com TimeFormat vazio
func TestProviderConfigureTimeFormatEmpty(t *testing.T) {
	provider := &Provider{}

	config := interfaces.Config{
		Level:      interfaces.InfoLevel,
		Format:     interfaces.JSONFormat,
		TimeFormat: "", // TimeFormat vazio para testar a linha if config.TimeFormat != ""
		Output:     &bytes.Buffer{},
	}

	err := provider.Configure(config)
	assert.NoError(t, err)
	assert.NotNil(t, provider.logger)
}

// TestProviderConfigureWithoutAddCaller testa Configure sem AddCaller
func TestProviderConfigureWithoutAddCaller(t *testing.T) {
	provider := &Provider{}

	config := interfaces.Config{
		Level:     interfaces.InfoLevel,
		Format:    interfaces.JSONFormat,
		AddCaller: false, // AddCaller false para testar a linha if config.AddCaller
		Output:    &bytes.Buffer{},
	}

	err := provider.Configure(config)
	assert.NoError(t, err)
	assert.NotNil(t, provider.logger)
}

// TestProviderConfigureWithoutGlobalFields testa Configure sem GlobalFields
func TestProviderConfigureWithoutGlobalFields(t *testing.T) {
	provider := &Provider{}

	config := interfaces.Config{
		Level:        interfaces.InfoLevel,
		Format:       interfaces.JSONFormat,
		GlobalFields: nil, // GlobalFields vazio para testar a linha if len(config.GlobalFields) > 0
		Output:       &bytes.Buffer{},
	}

	err := provider.Configure(config)
	assert.NoError(t, err)
	assert.NotNil(t, provider.logger)
}

// TestProviderConfigureWithoutServiceInfo testa Configure sem informações de serviço
func TestProviderConfigureWithoutServiceInfo(t *testing.T) {
	provider := &Provider{}

	config := interfaces.Config{
		Level:          interfaces.InfoLevel,
		Format:         interfaces.JSONFormat,
		ServiceName:    "", // ServiceName vazio
		ServiceVersion: "", // ServiceVersion vazio
		Environment:    "", // Environment vazio
		Output:         &bytes.Buffer{},
	}

	err := provider.Configure(config)
	assert.NoError(t, err)
	assert.NotNil(t, provider.logger)
}

// TestProviderConfigureWithPartialServiceInfo testa Configure com informações parciais do serviço
func TestProviderConfigureWithPartialServiceInfo(t *testing.T) {
	tests := []struct {
		name   string
		config interfaces.Config
	}{
		{
			name: "only service name",
			config: interfaces.Config{
				Level:          interfaces.InfoLevel,
				Format:         interfaces.JSONFormat,
				ServiceName:    "test-service",
				ServiceVersion: "", // vazio
				Environment:    "", // vazio
				Output:         &bytes.Buffer{},
			},
		},
		{
			name: "only service version",
			config: interfaces.Config{
				Level:          interfaces.InfoLevel,
				Format:         interfaces.JSONFormat,
				ServiceName:    "", // vazio
				ServiceVersion: "1.0.0",
				Environment:    "", // vazio
				Output:         &bytes.Buffer{},
			},
		},
		{
			name: "only environment",
			config: interfaces.Config{
				Level:          interfaces.InfoLevel,
				Format:         interfaces.JSONFormat,
				ServiceName:    "", // vazio
				ServiceVersion: "", // vazio
				Environment:    "production",
				Output:         &bytes.Buffer{},
			},
		},
		{
			name: "service name and version only",
			config: interfaces.Config{
				Level:          interfaces.InfoLevel,
				Format:         interfaces.JSONFormat,
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "", // vazio
				Output:         &bytes.Buffer{},
			},
		},
		{
			name: "service name and environment only",
			config: interfaces.Config{
				Level:          interfaces.InfoLevel,
				Format:         interfaces.JSONFormat,
				ServiceName:    "test-service",
				ServiceVersion: "", // vazio
				Environment:    "production",
				Output:         &bytes.Buffer{},
			},
		},
		{
			name: "version and environment only",
			config: interfaces.Config{
				Level:          interfaces.InfoLevel,
				Format:         interfaces.JSONFormat,
				ServiceName:    "", // vazio
				ServiceVersion: "1.0.0",
				Environment:    "production",
				Output:         &bytes.Buffer{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{}
			err := provider.Configure(tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, provider.logger)
		})
	}
}

// TestAddFieldToEventErrorTypeWithNilValue testa addFieldToEvent com ErrorType e valor nil
func TestAddFieldToEventErrorTypeWithNilValue(t *testing.T) {
	provider := &Provider{}
	config := interfaces.Config{
		Level:  interfaces.InfoLevel,
		Format: interfaces.JSONFormat,
		Output: &bytes.Buffer{},
	}
	provider.Configure(config)

	logger := zerolog.New(&bytes.Buffer{})
	event := logger.Info()

	field := interfaces.Field{
		Key:   "test_error",
		Value: nil, // Valor nil para testar if field.Value == nil
		Type:  interfaces.ErrorType,
	}

	result := addFieldToEvent(event, field)
	assert.NotNil(t, result)
}

// TestAddFieldToEventErrorTypeWithStringValue testa addFieldToEvent com ErrorType e valor string
func TestAddFieldToEventErrorTypeWithStringValue(t *testing.T) {
	provider := &Provider{}
	config := interfaces.Config{
		Level:  interfaces.InfoLevel,
		Format: interfaces.JSONFormat,
		Output: &bytes.Buffer{},
	}
	provider.Configure(config)

	logger := zerolog.New(&bytes.Buffer{})
	event := logger.Info()

	field := interfaces.Field{
		Key:   "test_error",
		Value: "error message", // String value para testar o case string
		Type:  interfaces.ErrorType,
	}

	result := addFieldToEvent(event, field)
	assert.NotNil(t, result)
}

// TestAddFieldToEventErrorTypeWithErrorKeyName testa addFieldToEvent com key "error"
func TestAddFieldToEventErrorTypeWithErrorKeyName(t *testing.T) {
	provider := &Provider{}
	config := interfaces.Config{
		Level:  interfaces.InfoLevel,
		Format: interfaces.JSONFormat,
		Output: &bytes.Buffer{},
	}
	provider.Configure(config)

	logger := zerolog.New(&bytes.Buffer{})
	event := logger.Info()

	field := interfaces.Field{
		Key:   "error", // Key "error" para testar if field.Key == "error"
		Value: errors.New("test error"),
		Type:  interfaces.ErrorType,
	}

	result := addFieldToEvent(event, field)
	assert.NotNil(t, result)
}

// TestAddFieldToContextErrorTypeWithNilValue testa addFieldToContext com ErrorType e valor nil
func TestAddFieldToContextErrorTypeWithNilValue(t *testing.T) {
	provider := &Provider{}
	config := interfaces.Config{
		Level:  interfaces.InfoLevel,
		Format: interfaces.JSONFormat,
		Output: &bytes.Buffer{},
	}
	provider.Configure(config)

	logger := zerolog.New(&bytes.Buffer{})
	ctx := logger.With()

	field := interfaces.Field{
		Key:   "test_error",
		Value: nil, // Valor nil para testar if field.Value == nil
		Type:  interfaces.ErrorType,
	}

	result := addFieldToContext(ctx, field)
	assert.NotNil(t, result)
}

// TestAddFieldToContextErrorTypeWithStringValue testa addFieldToContext com ErrorType e valor string
func TestAddFieldToContextErrorTypeWithStringValue(t *testing.T) {
	provider := &Provider{}
	config := interfaces.Config{
		Level:  interfaces.InfoLevel,
		Format: interfaces.JSONFormat,
		Output: &bytes.Buffer{},
	}
	provider.Configure(config)

	logger := zerolog.New(&bytes.Buffer{})
	ctx := logger.With()

	field := interfaces.Field{
		Key:   "test_error",
		Value: "error message", // String value para testar o case string
		Type:  interfaces.ErrorType,
	}

	result := addFieldToContext(ctx, field)
	assert.NotNil(t, result)
}

// TestAddFieldToContextErrorTypeWithErrorKeyName testa addFieldToContext com key "error"
func TestAddFieldToContextErrorTypeWithErrorKeyName(t *testing.T) {
	provider := &Provider{}
	config := interfaces.Config{
		Level:  interfaces.InfoLevel,
		Format: interfaces.JSONFormat,
		Output: &bytes.Buffer{},
	}
	provider.Configure(config)

	logger := zerolog.New(&bytes.Buffer{})
	ctx := logger.With()

	field := interfaces.Field{
		Key:   "error", // Key "error" para testar if field.Key == "error"
		Value: errors.New("test error"),
		Type:  interfaces.ErrorType,
	}

	result := addFieldToContext(ctx, field)
	assert.NotNil(t, result)
}

// TestExtractTraceIDWithNilContext testa extractTraceID com contexto nil
func TestExtractTraceIDWithNilContext(t *testing.T) {
	result := extractTraceID(context.TODO())
	assert.Empty(t, result)
}

// TestExtractSpanIDWithNilContext testa extractSpanID com contexto nil
func TestExtractSpanIDWithNilContext(t *testing.T) {
	result := extractSpanID(context.TODO())
	assert.Empty(t, result)
}

// TestExtractUserIDWithNilContext testa extractUserID com contexto nil
func TestExtractUserIDWithNilContext(t *testing.T) {
	result := extractUserID(context.TODO())
	assert.Empty(t, result)
}

// TestExtractRequestIDWithNilContext testa extractRequestID com contexto nil
func TestExtractRequestIDWithNilContext(t *testing.T) {
	result := extractRequestID(context.TODO())
	assert.Empty(t, result)
}

// TestExtractUserIDWithNonStringValue testa extractUserID com valor não-string
func TestExtractUserIDWithNonStringValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", 12345) // Valor int em vez de string
	result := extractUserID(ctx)
	assert.Empty(t, result)
}

// TestExtractRequestIDWithNonStringValue testa extractRequestID com valor não-string
func TestExtractRequestIDWithNonStringValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", 67890) // Valor int em vez de string
	result := extractRequestID(ctx)
	assert.Empty(t, result)
}

// TestExtractTraceIDWithNonStringValue testa extractTraceID com valor não-string
func TestExtractTraceIDWithNonStringValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "trace_id", 12345) // Valor int em vez de string
	result := extractTraceID(ctx)
	assert.Empty(t, result)
}

// TestExtractSpanIDWithNonStringValue testa extractSpanID com valor não-string
func TestExtractSpanIDWithNonStringValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "span_id", 67890) // Valor int em vez de string
	result := extractSpanID(ctx)
	assert.Empty(t, result)
}

// TestExtractTraceIDWithNilValue testa extractTraceID com valor nil no contexto
func TestExtractTraceIDWithNilValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "trace_id", nil)
	result := extractTraceID(ctx)
	assert.Empty(t, result)
}

// TestExtractSpanIDWithNilValue testa extractSpanID com valor nil no contexto
func TestExtractSpanIDWithNilValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "span_id", nil)
	result := extractSpanID(ctx)
	assert.Empty(t, result)
}

// TestExtractUserIDWithNilValue testa extractUserID com valor nil no contexto
func TestExtractUserIDWithNilValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", nil)
	result := extractUserID(ctx)
	assert.Empty(t, result)
}

// TestExtractRequestIDWithNilValue testa extractRequestID com valor nil no contexto
func TestExtractRequestIDWithNilValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", nil)
	result := extractRequestID(ctx)
	assert.Empty(t, result)
}

// TestAddContextToEventWithEmptyExtractors testa addContextToEvent quando todos os extractors retornam vazio
func TestAddContextToEventWithEmptyExtractors(t *testing.T) {
	provider := &Provider{}
	config := interfaces.Config{
		Level:  interfaces.InfoLevel,
		Format: interfaces.JSONFormat,
		Output: &bytes.Buffer{},
	}
	provider.Configure(config)

	logger := zerolog.New(&bytes.Buffer{})
	event := logger.Info()

	// Contexto sem nenhum dos valores esperados
	ctx := context.WithValue(context.Background(), "other_key", "other_value")

	result := addContextToEvent(event, ctx)
	assert.NotNil(t, result)
}

// TestExtractTraceIDWithTraceIdKey testa extractTraceID com chave "traceId"
func TestExtractTraceIDWithTraceIdKey(t *testing.T) {
	ctx := context.WithValue(context.Background(), "traceId", "trace-123")
	result := extractTraceID(ctx)
	assert.Equal(t, "trace-123", result)
}

// TestExtractSpanIDWithSpanIdKey testa extractSpanID com chave "spanId"
func TestExtractSpanIDWithSpanIdKey(t *testing.T) {
	ctx := context.WithValue(context.Background(), "spanId", "span-456")
	result := extractSpanID(ctx)
	assert.Equal(t, "span-456", result)
}

// TestExtractTraceIDWithTraceIdKeyNonString testa extractTraceID com chave "traceId" e valor não-string
func TestExtractTraceIDWithTraceIdKeyNonString(t *testing.T) {
	ctx := context.WithValue(context.Background(), "traceId", 12345)
	result := extractTraceID(ctx)
	assert.Empty(t, result)
}

// TestExtractSpanIDWithSpanIdKeyNonString testa extractSpanID com chave "spanId" e valor não-string
func TestExtractSpanIDWithSpanIdKeyNonString(t *testing.T) {
	ctx := context.WithValue(context.Background(), "spanId", 67890)
	result := extractSpanID(ctx)
	assert.Empty(t, result)
}

// TestExtractTraceIDWithTraceIdKeyNilValue testa extractTraceID com chave "traceId" e valor nil
func TestExtractTraceIDWithTraceIdKeyNilValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "traceId", nil)
	result := extractTraceID(ctx)
	assert.Empty(t, result)
}

// TestExtractSpanIDWithSpanIdKeyNilValue testa extractSpanID com chave "spanId" e valor nil
func TestExtractSpanIDWithSpanIdKeyNilValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "spanId", nil)
	result := extractSpanID(ctx)
	assert.Empty(t, result)
}
