package slog_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	"github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
)

func TestSlogProvider(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         &buf,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		AddSource:      false,
		AddStacktrace:  false,
		Fields: map[string]any{
			"component": "slog-test",
		},
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	// Testa logs básicos
	provider.Info(ctx, "test info message",
		logger.String("key1", "value1"),
		logger.Int("key2", 42),
	)

	output := buf.String()

	expectedFields := []string{
		"test info message",
		"service", "test-service",
		"version", "1.0.0",
		"environment", "test",
		"component", "slog-test",
		"key1", "value1",
		"key2", "42",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestSlogProviderLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	// Testa todos os níveis
	provider.Debug(ctx, "debug message")
	provider.Info(ctx, "info message")
	provider.Warn(ctx, "warn message")
	provider.Error(ctx, "error message")

	output := buf.String()
	levels := []string{"debug message", "info message", "warn message", "error message"}
	for _, level := range levels {
		if !strings.Contains(output, level) {
			t.Errorf("Expected '%s' in output: %s", level, output)
		}
	}
}

func TestSlogProviderContextFields(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	// Cria contexto com dados
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-123")
	ctx = context.WithValue(ctx, logger.SpanIDKey, "span-456")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-789")
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req-abc")

	provider.Info(ctx, "context test message")

	output := buf.String()

	expectedFields := []string{"trace-123", "span-456", "user-789", "req-abc"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestSlogProviderWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	// Cria logger com campos permanentes
	loggerWithFields := provider.WithFields(
		logger.String("component", "test"),
		logger.String("module", "slog"),
	)

	loggerWithFields.Info(ctx, "test message", logger.String("extra", "field"))

	output := buf.String()

	expectedFields := []string{"component", "test", "module", "slog", "extra", "field"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestSlogProviderFormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	// Testa logging formatado
	provider.Infof(ctx, "formatted message: %s, number: %d", "test", 42)

	output := buf.String()

	if !strings.Contains(output, "formatted message: test, number: 42") {
		t.Errorf("Expected formatted message in output: %s", output)
	}
}

func TestSlogProviderWithCode(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	// Testa logging com códigos
	provider.InfoWithCode(ctx, "USER_CREATED", "User created successfully",
		logger.String("user_id", "123"),
		logger.String("email", "test@example.com"),
	)

	output := buf.String()

	expectedFields := []string{"USER_CREATED", "User created successfully", "user_id", "123", "email", "test@example.com"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestSlogProviderLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.WarnLevel, // Só logs de WARN para cima
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	// Estes não devem aparecer
	provider.Debug(ctx, "debug message")
	provider.Info(ctx, "info message")

	// Estes devem aparecer
	provider.Warn(ctx, "warn message")
	provider.Error(ctx, "error message")

	output := buf.String()

	// Verifica que debug e info não aparecem
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear with WARN level")
	}

	if strings.Contains(output, "info message") {
		t.Error("Info message should not appear with WARN level")
	}

	// Verifica que warn e error aparecem
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear with WARN level")
	}

	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear with WARN level")
	}
}

func TestSlogProviderFormats(t *testing.T) {
	tests := []struct {
		name   string
		format logger.Format
	}{
		{"JSON", logger.JSONFormat},
		{"Text", logger.TextFormat},
		{"Console", logger.ConsoleFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &logger.Config{
				Level:  logger.InfoLevel,
				Format: tt.format,
				Output: &buf,
			}

			provider := slog.NewProvider()
			err := provider.Configure(config)
			if err != nil {
				t.Fatalf("Failed to configure provider: %v", err)
			}

			ctx := context.Background()
			provider.Info(ctx, "test message", logger.String("key", "value"))

			output := buf.String()
			if !strings.Contains(output, "test message") {
				t.Errorf("Expected 'test message' in output: %s", output)
			}
		})
	}
}

func TestSlogProviderTimeFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:      logger.InfoLevel,
		Format:     logger.JSONFormat,
		Output:     &buf,
		TimeFormat: time.RFC3339,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	provider.Info(ctx, "test message")

	output := buf.String()

	// Verifica se o timestamp está no formato correto
	if !strings.Contains(output, "time") {
		t.Errorf("Expected timestamp in output: %s", output)
	}
}

func TestSlogProviderLevelMethods(t *testing.T) {
	provider := slog.NewProvider()
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &bytes.Buffer{},
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	// Testa métodos de nível
	provider.SetLevel(logger.DebugLevel)
	if provider.GetLevel() != logger.DebugLevel {
		t.Errorf("Expected level to be Debug, got %v", provider.GetLevel())
	}

	provider.SetLevel(logger.ErrorLevel)
	if provider.GetLevel() != logger.ErrorLevel {
		t.Errorf("Expected level to be Error, got %v", provider.GetLevel())
	}
}

func TestSlogProviderCloneAndClose(t *testing.T) {
	provider := slog.NewProvider()
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &bytes.Buffer{},
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	// Testa clone
	clone := provider.Clone()
	if clone == nil {
		t.Error("Expected clone to not be nil")
	}

	// Testa close
	err = provider.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

func BenchmarkSlogProviderInfo(b *testing.B) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			provider.Info(ctx, "benchmark message",
				logger.String("key1", "value1"),
				logger.String("key2", "value2"),
				logger.Int("number", 42),
			)
		}
	})
}

func BenchmarkSlogProviderWithFields(b *testing.B) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	loggerWithFields := provider.WithFields(
		logger.String("service", "test"),
		logger.String("version", "1.0.0"),
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			loggerWithFields.Info(ctx, "benchmark message",
				logger.String("key", "value"),
				logger.Int("number", 42),
			)
		}
	})
}

func BenchmarkSlogProviderContextFields(b *testing.B) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := slog.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		b.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-123")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-456")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			provider.Info(ctx, "benchmark message",
				logger.String("key", "value"),
				logger.Int("number", 42),
			)
		}
	})
}
