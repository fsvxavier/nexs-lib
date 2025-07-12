package zap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
	"go.uber.org/zap/zapcore"
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

	// Sem configuração deve retornar erro
	err := provider.HealthCheck()
	if err == nil {
		t.Error("Expected error for unconfigured provider")
	}

	// Com configuração deve passar
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
	}
	provider.Configure(config)

	err = provider.HealthCheck()
	if err != nil {
		t.Errorf("Expected no error for configured provider, got %v", err)
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
		Format:      interfaces.JSONFormat,
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
		Format:      interfaces.JSONFormat,
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

	// Use um buffer em vez de stdout para evitar problemas de sync
	var buf bytes.Buffer
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      &buf,
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
		{interfaces.TraceLevel, "debug"}, // Zap não tem trace, mapeia para debug
		{interfaces.DebugLevel, "debug"},
		{interfaces.InfoLevel, "info"},
		{interfaces.WarnLevel, "warn"},
		{interfaces.ErrorLevel, "error"},
		{interfaces.FatalLevel, "fatal"},
		{interfaces.PanicLevel, "panic"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			zapLevel := convertLevel(test.input)
			if zapLevel.String() != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, zapLevel.String())
			}
		})
	}
}

func TestConvertFields(t *testing.T) {
	fields := []interfaces.Field{
		interfaces.String("string_field", "value"),
		interfaces.Int("int_field", 42),
		interfaces.Bool("bool_field", true),
	}

	zapFields := convertFields(fields)

	if len(zapFields) != len(fields) {
		t.Errorf("Expected %d zap fields, got %d", len(fields), len(zapFields))
	}

	// Verifica se os tipos estão corretos
	for i, zapField := range zapFields {
		if zapField.Key != fields[i].Key {
			t.Errorf("Expected key %s, got %s", fields[i].Key, zapField.Key)
		}
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

func TestProviderWithErrorEdgeCases(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
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

func TestProviderWithContextEdgeCases(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
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
		{"UNKNOWN", interfaces.Level(99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapLevel := convertLevel(tt.level)
			t.Logf("Level %s converted to zap level %v", tt.name, zapLevel)
		})
	}
}

func TestProviderConvertLevelFromZap(t *testing.T) {
	tests := []struct {
		name     string
		zapLevel zapcore.Level
	}{
		{"DebugLevel", zapcore.DebugLevel},
		{"InfoLevel", zapcore.InfoLevel},
		{"WarnLevel", zapcore.WarnLevel},
		{"ErrorLevel", zapcore.ErrorLevel},
		{"DPanicLevel", zapcore.DPanicLevel},
		{"PanicLevel", zapcore.PanicLevel},
		{"FatalLevel", zapcore.FatalLevel},
		{"UnknownLevel", zapcore.Level(99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := convertLevelFromZap(tt.zapLevel)
			t.Logf("Zap level %s converted to level %v", tt.name, level)
		})
	}
}

func TestProviderConvertFieldEdgeCases(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapField := convertField(tt.field)
			if zapField.Key == "" {
				t.Errorf("Expected field %s to be converted", tt.name)
			}
		})
	}
}

func TestProviderGetZapFieldValue(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First convert to zap field, then get value
			zapField := convertField(tt.field)
			value := getZapFieldValue(zapField)
			t.Logf("Field %s converted to value: %v", tt.name, value)
		})
	}
}

func TestProviderBuildEncoderConfigFormats(t *testing.T) {
	tests := []struct {
		name   string
		format interfaces.Format
	}{
		{"json", interfaces.JSONFormat},
		{"console", interfaces.ConsoleFormat},
		{"text", interfaces.TextFormat},
		{"unknown", interfaces.Format(99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := interfaces.Config{
				Format:        tt.format,
				AddCaller:     true,
				AddStacktrace: true,
				AddSource:     true,
				TimeFormat:    "2006-01-02T15:04:05Z07:00",
			}
			encoderConfig := buildEncoderConfig(config)
			if encoderConfig.MessageKey == "" {
				t.Errorf("Expected encoder config to be built for format %s", tt.name)
			}
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

func TestProviderExtractContextFields(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      os.Stdout,
	}
	provider.Configure(config)

	// Test with context that has fields
	ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
	ctx = context.WithValue(ctx, "span_id", "span-456")

	fields := extractContextFields(ctx)
	if len(fields) == 0 {
		t.Log("No fields extracted from context (which is expected)")
	} else {
		t.Logf("Extracted %d fields from context", len(fields))
	}
}

func TestProviderConfigureWithComplexOptions(t *testing.T) {
	provider := NewProvider()
	config := interfaces.Config{
		Level:         interfaces.DebugLevel,
		Format:        interfaces.JSONFormat,
		ServiceName:   "test-service",
		Output:        os.Stdout,
		AddCaller:     true,
		AddStacktrace: true,
		TimeFormat:    "2006-01-02T15:04:05.000Z07:00",
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that all options were applied
	ctx := context.Background()
	provider.Debug(ctx, "test message with all options")
}

func TestProviderConfigureComplexEncoderOptions(t *testing.T) {
	provider := NewProvider()

	tests := []struct {
		name   string
		config interfaces.Config
	}{
		{
			"json_with_all_options",
			interfaces.Config{
				Level:         interfaces.TraceLevel,
				Format:        interfaces.JSONFormat,
				ServiceName:   "zap-service",
				AddCaller:     true,
				AddStacktrace: true,
				AddSource:     true,
				TimeFormat:    time.RFC3339Nano,
			},
		},
		{
			"console_with_colors",
			interfaces.Config{
				Level:       interfaces.DebugLevel,
				Format:      interfaces.ConsoleFormat,
				ServiceName: "console-service",
				AddCaller:   true,
			},
		},
		{
			"text_format_simple",
			interfaces.Config{
				Level:       interfaces.InfoLevel,
				Format:      interfaces.TextFormat,
				ServiceName: "text-service",
			},
		},
		{
			"unknown_format_fallback",
			interfaces.Config{
				Level:       interfaces.InfoLevel,
				Format:      interfaces.Format(99), // Unknown format
				ServiceName: "fallback-service",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.Configure(tt.config)
			if err != nil {
				t.Fatalf("Expected no error for %s, got %v", tt.name, err)
			}

			// Test logging with configured provider
			ctx := context.Background()
			provider.Info(ctx, "test message with "+tt.name)
			provider.Error(ctx, "test error with "+tt.name)
		})
	}
}

func TestProviderBuildEncoderConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config interfaces.Config
	}{
		{
			"json_format_custom_time",
			interfaces.Config{
				Format:     interfaces.JSONFormat,
				TimeFormat: "2006-01-02T15:04:05.000Z",
				AddCaller:  true,
			},
		},
		{
			"console_format_no_caller",
			interfaces.Config{
				Format:    interfaces.ConsoleFormat,
				AddCaller: false,
			},
		},
		{
			"text_format_with_source",
			interfaces.Config{
				Format:    interfaces.TextFormat,
				AddSource: true,
			},
		},
		{
			"unknown_format",
			interfaces.Config{
				Format: interfaces.Format(99),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoderConfig := buildEncoderConfig(tt.config)

			// Verify encoder config is valid
			if encoderConfig.MessageKey == "" {
				t.Error("Expected MessageKey to be set")
			}
			if encoderConfig.LevelKey == "" {
				t.Error("Expected LevelKey to be set")
			}
			if encoderConfig.TimeKey == "" {
				t.Error("Expected TimeKey to be set")
			}
		})
	}
}

func TestProviderGetZapFieldValueEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		field interfaces.Field
	}{
		{"string_field", interfaces.String("key", "test_value")},
		{"int_field", interfaces.Int("key", 12345)},
		{"int64_field", interfaces.Int64("key", int64(9223372036854775807))},
		{"float64_field", interfaces.Float64("key", 3.141592653589793)},
		{"bool_true", interfaces.Bool("key", true)},
		{"bool_false", interfaces.Bool("key", false)},
		{"time_field", interfaces.Time("timestamp", time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))},
		{"duration_field", interfaces.Duration("elapsed", time.Hour+30*time.Minute+15*time.Second)},
		{"error_field", interfaces.Error(errors.New("detailed test error message"))},
		{"nil_error_field", interfaces.Error(nil)},
		{"object_simple", interfaces.Object("obj", map[string]interface{}{"nested": "value"})},
		{"array_strings", interfaces.Array("strings", []interface{}{"alpha", "beta", "gamma"})},
		{"unknown_type", interfaces.Field{Key: "unknown", Type: interfaces.FieldType(99), Value: "fallback_value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert interface field to zap field first
			zapField := convertField(tt.field)

			// Then test getZapFieldValue
			value := getZapFieldValue(zapField)

			t.Logf("Field %s converted to Zap field with value: %v", tt.name, value)
		})
	}
}

func TestProviderConvertZapLevelCompleteEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		level    interfaces.Level
		expected zapcore.Level
	}{
		{"TRACE", interfaces.TraceLevel, zapcore.DebugLevel},
		{"DEBUG", interfaces.DebugLevel, zapcore.DebugLevel},
		{"INFO", interfaces.InfoLevel, zapcore.InfoLevel},
		{"WARN", interfaces.WarnLevel, zapcore.WarnLevel},
		{"ERROR", interfaces.ErrorLevel, zapcore.ErrorLevel},
		{"FATAL", interfaces.FatalLevel, zapcore.FatalLevel},
		{"PANIC", interfaces.PanicLevel, zapcore.PanicLevel},
		{"UNKNOWN", interfaces.Level(99), zapcore.InfoLevel}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertLevel(tt.level)
			if result != tt.expected {
				t.Errorf("convertLevel(%v) = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}

func TestProviderConvertFromZapLevelEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		zapLevel zapcore.Level
		expected interfaces.Level
	}{
		{"DebugLevel", zapcore.DebugLevel, interfaces.DebugLevel},
		{"InfoLevel", zapcore.InfoLevel, interfaces.InfoLevel},
		{"WarnLevel", zapcore.WarnLevel, interfaces.WarnLevel},
		{"ErrorLevel", zapcore.ErrorLevel, interfaces.ErrorLevel},
		{"DPanicLevel", zapcore.DPanicLevel, interfaces.InfoLevel}, // Falls to default
		{"PanicLevel", zapcore.PanicLevel, interfaces.PanicLevel},
		{"FatalLevel", zapcore.FatalLevel, interfaces.FatalLevel},
		{"InvalidLevel", zapcore.Level(99), interfaces.InfoLevel}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertLevelFromZap(tt.zapLevel)
			if result != tt.expected {
				t.Errorf("convertLevelFromZap(%v) = %v, want %v", tt.zapLevel, result, tt.expected)
			}
		})
	}
}

func TestProviderComplexFieldScenarios(t *testing.T) {
	var buf bytes.Buffer
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "test-service",
		Output:      &buf,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()

	// Test with many complex fields
	complexFields := []interfaces.Field{
		interfaces.String("service", "user-service"),
		interfaces.String("version", "1.2.3"),
		interfaces.Int("user_id", 12345),
		interfaces.Int64("session_id", int64(9876543210)),
		interfaces.Float64("response_time", 0.025),
		interfaces.Bool("success", true),
		interfaces.Time("timestamp", time.Now()),
		interfaces.Duration("duration", 25*time.Millisecond),
		interfaces.Error(errors.New("validation failed")),
		interfaces.Object("request", map[string]interface{}{
			"method": "POST",
			"path":   "/api/v1/users",
			"headers": map[string]interface{}{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
			"body": map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
				"age":   30,
			},
		}),
		interfaces.Array("tags", []interface{}{"api", "user", "create"}),
		interfaces.Array("numbers", []interface{}{1, 2, 3, 4, 5}),
	}

	fieldsLogger := provider.WithFields(complexFields...)
	fieldsLogger.Info(ctx, "Complex logging scenario")

	// Verify output contains expected data
	output := buf.String()
	if output == "" {
		t.Error("Expected log output, got empty string")
	}
}

func TestProviderTraceWithComplexContext(t *testing.T) {
	var buf bytes.Buffer
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.TraceLevel, // Enable trace level
		Format:      interfaces.JSONFormat,
		ServiceName: "trace-service",
		AddCaller:   true,
		Output:      &buf,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Create complex context
	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "trace-abc-123")
	ctx = context.WithValue(ctx, "span_id", "span-def-456")
	ctx = context.WithValue(ctx, "user_id", "user-ghi-789")
	ctx = context.WithValue(ctx, "request_id", "req-jkl-012")

	provider.Trace(ctx, "Trace message with complex context")

	output := buf.String()
	if output == "" {
		t.Error("Expected trace output, got empty string")
	}
}

func TestProviderWithErrorComplexScenarios(t *testing.T) {
	var buf bytes.Buffer
	provider := NewProvider()
	config := interfaces.Config{
		Level:       interfaces.ErrorLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "error-service",
		Output:      &buf,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()

	// Test with different error types
	errors := []error{
		errors.New("simple error"),
		fmt.Errorf("wrapped error: %w", errors.New("original error")),
		nil, // nil error
	}

	for i, testErr := range errors {
		t.Run(fmt.Sprintf("error_%d", i), func(t *testing.T) {
			errorLogger := provider.WithError(testErr)
			if errorLogger == nil {
				t.Error("Expected error logger to be created")
			}

			errorLogger.Error(ctx, fmt.Sprintf("Error scenario %d", i))
		})
	}
}

func TestProviderStacktraceConfiguration(t *testing.T) {
	var buf bytes.Buffer
	provider := NewProvider()
	config := interfaces.Config{
		Level:         interfaces.ErrorLevel,
		Format:        interfaces.JSONFormat,
		ServiceName:   "stacktrace-service",
		AddStacktrace: true,
		Output:        &buf,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()

	// This should include stacktrace
	provider.Error(ctx, "Error with stacktrace enabled")

	output := buf.String()
	if output == "" {
		t.Error("Expected error output with stacktrace, got empty string")
	}
}
