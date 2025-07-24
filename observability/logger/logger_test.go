package logger_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
)

func TestLoggerInterface(t *testing.T) {
	// Testa se todos os providers implementam a interface corretamente
	providers := []string{"slog"}

	for _, providerName := range providers {
		t.Run(providerName, func(t *testing.T) {
			var buf bytes.Buffer
			config := &logger.Config{
				Level:       logger.InfoLevel,
				Format:      logger.JSONFormat,
				Output:      &buf,
				ServiceName: "test-service",
			}

			err := logger.SetProvider(providerName, config)
			if err != nil {
				t.Fatalf("Failed to set provider %s: %v", providerName, err)
			}

			ctx := context.Background()

			// Testa logs básicos
			logger.Info(ctx, "test message", logger.String("key", "value"))
			logger.Infof(ctx, "formatted message: %s", "test")

			output := buf.String()
			if !strings.Contains(output, "test message") {
				t.Errorf("Expected 'test message' in output, got: %s", output)
			}

			if !strings.Contains(output, "formatted message") {
				t.Errorf("Expected 'formatted message' in output, got: %s", output)
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Testa todos os níveis
	logger.Debug(ctx, "debug message")
	logger.Info(ctx, "info message")
	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message")

	output := buf.String()
	levels := []string{"debug message", "info message", "warn message", "error message"}
	for _, level := range levels {
		if !strings.Contains(output, level) {
			t.Errorf("Expected '%s' in output: %s", level, output)
		}
	}
}

func TestContextAwareLogging(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	// Cria contexto com dados
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-123")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-456")

	logger.Info(ctx, "context test message")

	output := buf.String()

	expectedFields := []string{"trace-123", "user-456"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Cria logger com campos permanentes
	loggerWithFields := logger.WithFields(
		logger.String("component", "test"),
		logger.String("module", "logging"),
	)

	loggerWithFields.Info(ctx, "test message", logger.String("extra", "field"))

	output := buf.String()

	expectedFields := []string{"component", "test", "module", "logging", "extra", "field"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Testa diferentes tipos de campos
	logger.Info(ctx, "structured test",
		logger.String("string_field", "value"),
		logger.Int("int_field", 42),
		logger.Int64("int64_field", 123456789),
		logger.Float64("float_field", 3.14),
		logger.Bool("bool_field", true),
		logger.Duration("duration_field", time.Second),
		logger.Time("time_field", time.Now()),
	)

	output := buf.String()

	expectedFields := []string{
		"string_field", "value",
		"int_field", "42",
		"int64_field", "123456789",
		"float_field", "3.14",
		"bool_field", "true",
		"duration_field",
		"time_field",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestConfig(t *testing.T) {
	// Testa configuração padrão
	defaultConf := logger.DefaultConfig()
	if defaultConf.Level != logger.InfoLevel {
		t.Errorf("Expected default level to be Info, got %v", defaultConf.Level)
	}

	if defaultConf.Format != logger.JSONFormat {
		t.Errorf("Expected default format to be JSON, got %v", defaultConf.Format)
	}

	// Testa configuração personalizada
	customConf := &logger.Config{
		Level:          logger.DebugLevel,
		Format:         logger.ConsoleFormat,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		AddSource:      true,
		AddStacktrace:  true,
		Fields: map[string]any{
			"custom": "field",
		},
	}

	var buf bytes.Buffer
	customConf.Output = &buf

	err := logger.SetProvider("slog", customConf)
	if err != nil {
		t.Fatalf("Failed to set provider with custom config: %v", err)
	}

	ctx := context.Background()
	logger.Info(ctx, "test with custom config")

	output := buf.String()

	expectedFields := []string{"service", "test-service", "version", "1.0.0", "environment", "test", "custom", "field"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestEnvironmentConfig(t *testing.T) {
	// Salva valores originais
	originalLevel := os.Getenv("LOG_LEVEL")
	originalFormat := os.Getenv("LOG_FORMAT")
	originalService := os.Getenv("SERVICE_NAME")

	// Define variáveis de ambiente
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "console")
	os.Setenv("SERVICE_NAME", "test-env-service")

	// Testa configuração por ambiente
	config := logger.EnvironmentConfig()

	if config.Level != logger.DebugLevel {
		t.Errorf("Expected level to be Debug, got %v", config.Level)
	}

	if config.Format != logger.ConsoleFormat {
		t.Errorf("Expected format to be Console, got %v", config.Format)
	}

	if config.ServiceName != "test-env-service" {
		t.Errorf("Expected service name to be 'test-env-service', got %s", config.ServiceName)
	}

	// Restaura valores originais
	os.Setenv("LOG_LEVEL", originalLevel)
	os.Setenv("LOG_FORMAT", originalFormat)
	os.Setenv("SERVICE_NAME", originalService)
}

func TestProviderManagement(t *testing.T) {
	// Testa listagem de providers
	providers := logger.ListProviders()
	if len(providers) == 0 {
		t.Error("Expected at least one provider to be registered")
	}

	found := false
	for _, provider := range providers {
		if provider == "slog" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'slog' provider to be registered")
	}

	// Testa configuração de provider inexistente
	config := logger.DefaultConfig()
	err := logger.SetProvider("nonexistent", config)
	if err == nil {
		t.Error("Expected error when setting nonexistent provider")
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.WarnLevel, // Só logs de WARN para cima
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Estes não devem aparecer
	logger.Debug(ctx, "debug message")
	logger.Info(ctx, "info message")

	// Estes devem aparecer
	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message")

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

func TestCodeLogging(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Testa logging com códigos
	current := logger.GetCurrentProvider()
	current.InfoWithCode(ctx, "USER_CREATED", "User created successfully",
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

func BenchmarkLoggerInfo(b *testing.B) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		b.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(ctx, "benchmark message",
				logger.String("key1", "value1"),
				logger.String("key2", "value2"),
				logger.Int("number", 42),
			)
		}
	})
}

func BenchmarkLoggerWithFields(b *testing.B) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		b.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()
	loggerWithFields := logger.WithFields(
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
