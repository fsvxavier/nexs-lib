package zerolog

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
	"github.com/rs/zerolog"
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
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
	}
	provider.Configure(config)

	// Create a zerolog event
	event := provider.logger.Info()

	tests := []struct {
		name  string
		field interfaces.Field
	}{
		{"string", interfaces.String("key", "value")},
		{"int", interfaces.Int("key", 42)},
		{"int64", interfaces.Int64("key", int64(42))},
		{"float64", interfaces.Float64("key", 42.5)},
		{"bool", interfaces.Bool("key", true)},
		{"time", interfaces.Time("timestamp", time.Now())},
		{"duration", interfaces.Duration("elapsed", time.Second)},
		{"error", interfaces.Error(errors.New("test error"))},
		{"nil_error", interfaces.Error(nil)},
		{"object", interfaces.Object("obj", map[string]interface{}{"nested": "value"})},
		{"array", interfaces.Array("arr", []interface{}{1, 2, 3})},
		{"unknown", interfaces.Field{Key: "unknown", Type: interfaces.FieldType(99), Value: "unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test addFieldToEvent with various field types
			addFieldToEvent(event, tt.field)
			t.Logf("Field %s added to event", tt.name)
		})
	}
}

func TestProviderAddFieldToContextEdgeCases(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
	}
	provider.Configure(config)

	// Create a zerolog context
	ctx := provider.logger.With()

	tests := []struct {
		name  string
		field interfaces.Field
	}{
		{"string", interfaces.String("key", "value")},
		{"int", interfaces.Int("key", 42)},
		{"int64", interfaces.Int64("key", int64(42))},
		{"float64", interfaces.Float64("key", 42.5)},
		{"bool", interfaces.Bool("key", true)},
		{"time", interfaces.Time("timestamp", time.Now())},
		{"duration", interfaces.Duration("elapsed", time.Second)},
		{"error", interfaces.Error(errors.New("test error"))},
		{"object", interfaces.Object("obj", map[string]interface{}{"nested": "value"})},
		{"array", interfaces.Array("arr", []interface{}{1, 2, 3})},
		{"unknown", interfaces.Field{Key: "unknown", Type: interfaces.FieldType(99), Value: "unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test addFieldToContext with various field types
			addFieldToContext(ctx, tt.field)
			t.Logf("Field %s added to context", tt.name)
		})
	}
}

func TestProviderAddContextToEventComplex(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
	}
	provider.Configure(config)

	// Create a zerolog event
	event := provider.logger.Info()

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{"empty", context.Background()},
		{"with_trace", context.WithValue(context.Background(), "trace_id", "trace-123")},
		{"with_span", context.WithValue(context.Background(), "span_id", "span-456")},
		{"with_user", context.WithValue(context.Background(), "user_id", "user-789")},
		{"with_request", context.WithValue(context.Background(), "request_id", "req-101")},
		{"with_multiple", func() context.Context {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "trace_id", "trace-123")
			ctx = context.WithValue(ctx, "span_id", "span-456")
			ctx = context.WithValue(ctx, "user_id", "user-789")
			ctx = context.WithValue(ctx, "request_id", "req-101")
			return ctx
		}()},
		{"with_non_string_values", func() context.Context {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "trace_id", 123)           // Non-string value
			ctx = context.WithValue(ctx, "span_id", []byte("span")) // Byte slice
			return ctx
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test addContextToEvent with various context types
			addContextToEvent(event, tt.ctx)
			t.Logf("Context %s added to event", tt.name)
		})
	}
}

func TestProviderExtractContextFunctions(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{"empty", context.Background()},
		{"with_trace", context.WithValue(context.Background(), "trace_id", "trace-123")},
		{"with_span", context.WithValue(context.Background(), "span_id", "span-456")},
		{"with_user", context.WithValue(context.Background(), "user_id", "user-789")},
		{"with_request", context.WithValue(context.Background(), "request_id", "req-101")},
		{"with_all", func() context.Context {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "trace_id", "trace-123")
			ctx = context.WithValue(ctx, "span_id", "span-456")
			ctx = context.WithValue(ctx, "user_id", "user-789")
			ctx = context.WithValue(ctx, "request_id", "req-101")
			return ctx
		}()},
		{"with_non_string_values", func() context.Context {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "trace_id", 123)
			ctx = context.WithValue(ctx, "span_id", []byte("span"))
			ctx = context.WithValue(ctx, "user_id", 456)
			ctx = context.WithValue(ctx, "request_id", 789)
			return ctx
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test all extraction functions
			traceID := extractTraceID(tt.ctx)
			spanID := extractSpanID(tt.ctx)
			userID := extractUserID(tt.ctx)
			requestID := extractRequestID(tt.ctx)

			t.Logf("Extracted - Trace: %s, Span: %s, User: %s, Request: %s",
				traceID, spanID, userID, requestID)
		})
	}
}

func TestProviderConfigureComplexOptions(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:         interfaces.TraceLevel,
		Format:        interfaces.ConsoleFormat,
		ServiceName:   "test-service",
		AddCaller:     true,
		AddStacktrace: true,
		AddSource:     true,
		TimeFormat:    "2006-01-02T15:04:05.000Z07:00",
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that all options were applied
	ctx := context.Background()
	provider.Trace(ctx, "trace message with all options")
	provider.Debug(ctx, "debug message with all options")
	provider.Error(ctx, "error message with all options")
}
