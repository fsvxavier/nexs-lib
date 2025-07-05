package logger

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoggerInterface(t *testing.T) {
	// Testa se todos os providers implementam a interface corretamente
	providers := []string{"slog", "zap"}

	for _, providerName := range providers {
		t.Run(providerName, func(t *testing.T) {
			var buf bytes.Buffer
			config := &Config{
				Level:       InfoLevel,
				Format:      JSONFormat,
				Output:      &buf,
				ServiceName: "test-service",
			}

			err := SetProvider(providerName, config)
			if err != nil {
				t.Fatalf("Failed to set provider %s: %v", providerName, err)
			}

			ctx := context.Background()

			// Testa logs básicos
			Info(ctx, "test message", String("key", "value"))
			Infof(ctx, "formatted message: %s", "test")

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

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  WarnLevel,
		Format: JSONFormat,
		Output: &buf,
	}

	err := SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Debug e Info não devem aparecer com nível Warn
	Debug(ctx, "debug message")
	Info(ctx, "info message")
	Warn(ctx, "warn message")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear with Warn level")
	}

	if strings.Contains(output, "info message") {
		t.Error("Info message should not appear with Warn level")
	}

	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear with Warn level")
	}
}

func TestFields(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: &buf,
	}

	err := SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Testa diferentes tipos de campos
	Info(ctx, "test fields",
		String("string_field", "test"),
		Int("int_field", 42),
		Int64("int64_field", 123456789),
		Float64("float_field", 3.14),
		Bool("bool_field", true),
		Duration("duration_field", 5*time.Second),
	)

	output := buf.String()

	expectedFields := []string{
		"string_field", "test",
		"int_field", "42",
		"int64_field", "123456789",
		"float_field", "3.14",
		"bool_field", "true",
		"duration_field",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: &buf,
	}

	err := SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Cria logger com campos permanentes
	logger := WithFields(
		String("component", "test"),
		String("module", "logging"),
	)

	logger.Info(ctx, "test message", String("extra", "field"))

	output := buf.String()

	expectedFields := []string{"component", "test", "module", "logging", "extra", "field"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestContextFields(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: &buf,
	}

	err := SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	// Cria contexto com valores
	ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	logger := WithContext(ctx)
	logger.Info(ctx, "test message")

	output := buf.String()

	expectedFields := []string{"trace_id", "trace-123", "user_id", "user-456"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestConfig(t *testing.T) {
	// Testa configuração padrão
	defaultConf := DefaultConfig()
	if defaultConf.Level != InfoLevel {
		t.Errorf("Expected default level to be Info, got %v", defaultConf.Level)
	}

	if defaultConf.Format != JSONFormat {
		t.Errorf("Expected default format to be JSON, got %v", defaultConf.Format)
	}

	// Testa configuração personalizada
	customConf := &Config{
		Level:          DebugLevel,
		Format:         ConsoleFormat,
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

	err := SetProvider("slog", customConf)
	if err != nil {
		t.Fatalf("Failed to set provider with custom config: %v", err)
	}

	ctx := context.Background()
	Info(ctx, "test with custom config")

	output := buf.String()

	expectedFields := []string{"service", "test-service", "version", "1.0.0", "environment", "test", "custom", "field"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' in output: %s", field, output)
		}
	}
}

func TestProviderManagement(t *testing.T) {
	// Testa listagem de providers
	providers := ListProviders()
	if len(providers) == 0 {
		t.Error("Expected at least one provider to be registered")
	}

	expectedProviders := []string{"slog", "zap"}
	for _, expected := range expectedProviders {
		found := false
		for _, provider := range providers {
			if provider == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider '%s' to be registered", expected)
		}
	}

	// Testa provider inexistente
	err := SetProvider("nonexistent", DefaultConfig())
	if err == nil {
		t.Error("Expected error when setting nonexistent provider")
	}
}

func TestErrorHandling(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: &buf,
	}

	err := SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Testa diferentes níveis de erro
	logger := WithFields(String("test", "error"))

	logger.Error(ctx, "error message", ErrorField(nil))
	logger.Warn(ctx, "warning message")

	output := buf.String()

	if !strings.Contains(output, "error message") {
		t.Error("Expected error message in output")
	}

	if !strings.Contains(output, "warning message") {
		t.Error("Expected warning message in output")
	}
}

func BenchmarkLogging(b *testing.B) {
	// Registrar provider mock para evitar problemas de concorrência
	RegisterProvider("mock", &mockProvider{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Buffer e config por goroutine para evitar race condition
		var buf bytes.Buffer
		config := &Config{
			Level:  InfoLevel,
			Format: JSONFormat,
			Output: &buf,
		}

		// Usar provider mock por goroutine
		SetProvider("mock", config)
		ctx := context.Background()

		for pb.Next() {
			Info(ctx, "benchmark message",
				String("key1", "value1"),
				Int("key2", 42),
				Float64("key3", 3.14),
			)
		}
	})
}

func BenchmarkWithFields(b *testing.B) {
	// Registrar provider mock para evitar problemas de concorrência
	RegisterProvider("mock", &mockProvider{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Buffer e config por goroutine para evitar race condition
		var buf bytes.Buffer
		config := &Config{
			Level:  InfoLevel,
			Format: JSONFormat,
			Output: &buf,
		}

		// Usar provider mock por goroutine
		SetProvider("mock", config)
		ctx := context.Background()

		logger := WithFields(
			String("component", "benchmark"),
			String("module", "testing"),
		)

		for pb.Next() {
			logger.Info(ctx, "benchmark message",
				String("dynamic", "value"),
				Int("counter", 1),
			)
		}
	})
}

// mockProvider implementação thread-safe para benchmarks
type mockProvider struct {
	mu sync.Mutex
}

func (m *mockProvider) Configure(config *Config) error {
	return nil
}

func (m *mockProvider) Debug(ctx context.Context, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Info(ctx context.Context, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Warn(ctx context.Context, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Error(ctx context.Context, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Fatal(ctx context.Context, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Panic(ctx context.Context, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Debugf(ctx context.Context, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Infof(ctx context.Context, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Warnf(ctx context.Context, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Errorf(ctx context.Context, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Fatalf(ctx context.Context, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) Panicf(ctx context.Context, format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) DebugWithCode(ctx context.Context, code, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) InfoWithCode(ctx context.Context, code, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) WarnWithCode(ctx context.Context, code, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) ErrorWithCode(ctx context.Context, code, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) WithFields(fields ...Field) Logger {
	return m
}

func (m *mockProvider) WithContext(ctx context.Context) Logger {
	return m
}

func (m *mockProvider) SetLevel(level Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// noop
}

func (m *mockProvider) GetLevel() Level {
	return InfoLevel
}

func (m *mockProvider) Clone() Logger {
	return &mockProvider{}
}

func (m *mockProvider) Close() error {
	return nil
}
