package zerolog

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

func TestZerologProvider_NewProvider(t *testing.T) {
	provider := NewProvider()
	if provider == nil {
		t.Error("NewProvider() returned nil")
	}
}

func TestZerologProvider_Configure(t *testing.T) {
	tests := []struct {
		name   string
		config *logger.Config
		want   bool
	}{
		{
			name: "json format",
			config: &logger.Config{
				Level:  logger.InfoLevel,
				Format: logger.JSONFormat,
			},
			want: true,
		},
		{
			name: "console format",
			config: &logger.Config{
				Level:  logger.DebugLevel,
				Format: logger.ConsoleFormat,
			},
			want: true,
		},
		{
			name: "with custom fields",
			config: &logger.Config{
				Level:          logger.InfoLevel,
				Format:         logger.JSONFormat,
				ServiceName:    "test-service",
				ServiceVersion: "1.0.0",
				Environment:    "test",
				AddSource:      true,
				AddStacktrace:  true,
				Fields: map[string]any{
					"custom_field": "custom_value",
				},
			},
			want: true,
		},
		{
			name: "with buffer config",
			config: &logger.Config{
				Level:  logger.InfoLevel,
				Format: logger.JSONFormat,
				BufferConfig: &interfaces.BufferConfig{
					Enabled:   true,
					Size:      100,
					BatchSize: 10,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider()
			err := provider.Configure(tt.config)
			if (err == nil) != tt.want {
				t.Errorf("Configure() error = %v, want success = %v", err, tt.want)
			}
		})
	}
}

func TestZerologProvider_BasicLogging(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	provider.Info(ctx, "test message")

	if buf.Len() == 0 {
		t.Error("Expected output, got none")
	}

	// Verifica se o output é um JSON válido
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Verifica campos obrigatórios do zerolog
	if _, ok := logEntry["time"]; !ok {
		t.Error("Missing time field")
	}
	if _, ok := logEntry["level"]; !ok {
		t.Error("Missing level field")
	}
	if _, ok := logEntry["message"]; !ok {
		t.Error("Missing message field")
	}
}

func TestZerologProvider_AllLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := NewProvider()
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
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Remove linhas vazias
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	if len(nonEmptyLines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(nonEmptyLines))
	}

	// Verifica se cada linha contém o nível correto
	levels := []string{"debug", "info", "warn", "error"}
	for i, line := range nonEmptyLines {
		if !strings.Contains(strings.ToLower(line), levels[i]) {
			t.Errorf("Line %d should contain level %s: %s", i, levels[i], line)
		}
	}
}

func TestZerologProvider_WithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	fields := []logger.Field{
		logger.String("key1", "value1"),
		logger.Int("key2", 42),
		logger.Bool("key3", true),
	}

	provider.Info(ctx, "test message", fields...)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verifica os campos
	if logEntry["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", logEntry["key1"])
	}
	if logEntry["key2"] != float64(42) { // JSON unmarshals numbers as float64
		t.Errorf("Expected key2=42, got %v", logEntry["key2"])
	}
	if logEntry["key3"] != true {
		t.Errorf("Expected key3=true, got %v", logEntry["key3"])
	}
}

func TestZerologProvider_WithCode(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	provider.InfoWithCode(ctx, "USER_001", "User created successfully")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if logEntry["code"] != "USER_001" {
		t.Errorf("Expected code=USER_001, got %v", logEntry["code"])
	}
}

func TestZerologProvider_FormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	provider.Infof(ctx, "User %s has %d points", "john", 100)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	expectedMsg := "User john has 100 points"
	if logEntry["message"] != expectedMsg {
		t.Errorf("Expected message=%s, got %v", expectedMsg, logEntry["message"])
	}
}

func TestZerologProvider_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.WarnLevel, // Só deve permitir Warn, Error, Fatal, Panic
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	// Estas não devem aparecer
	provider.Debug(ctx, "debug message")
	provider.Info(ctx, "info message")

	// Estas devem aparecer
	provider.Warn(ctx, "warn message")
	provider.Error(ctx, "error message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Remove linhas vazias
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	if len(nonEmptyLines) != 2 {
		t.Errorf("Expected 2 log lines (warn+error), got %d", len(nonEmptyLines))
	}
}

func TestZerologProvider_GetSetLevel(t *testing.T) {
	provider := NewProvider()
	config := &logger.Config{
		Level:  logger.WarnLevel,
		Format: logger.JSONFormat,
	}

	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	level := provider.GetLevel()
	if level != logger.WarnLevel {
		t.Errorf("Expected level %v, got %v", logger.WarnLevel, level)
	}

	provider.SetLevel(logger.ErrorLevel)
	level = provider.GetLevel()
	if level != logger.ErrorLevel {
		t.Errorf("Expected level %v after SetLevel, got %v", logger.ErrorLevel, level)
	}
}

func TestZerologProvider_Clone(t *testing.T) {
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	cloned := provider.Clone()
	if cloned == nil {
		t.Error("Clone returned nil")
	}

	if provider == cloned {
		t.Error("Clone returned same instance")
	}

	// Verifica se o clone tem as mesmas configurações
	if provider.GetLevel() != cloned.GetLevel() {
		t.Error("Clone has different level")
	}
}

func TestZerologProvider_Close(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	err = provider.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestZerologProvider_WithBuffer(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
		BufferConfig: &interfaces.BufferConfig{
			Enabled:   true,
			Size:      10,
			BatchSize: 2,
		},
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()

	// Escreve algumas mensagens
	provider.Info(ctx, "message 1")
	provider.Info(ctx, "message 2")
	provider.Info(ctx, "message 3")

	// Verifica se o buffer existe
	buffer := provider.GetBuffer()
	if buffer == nil {
		t.Error("Buffer not configured")
	}

	// Força o flush
	err = provider.FlushBuffer()
	if err != nil {
		t.Errorf("FlushBuffer failed: %v", err)
	}

	// Verifica se algo foi escrito
	if buf.Len() == 0 {
		t.Error("Expected output after flush, got none")
	}
}

func TestZerologProvider_BufferStats(t *testing.T) {
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		BufferConfig: &interfaces.BufferConfig{
			Enabled:   true,
			Size:      10,
			BatchSize: 5,
		},
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	provider.Info(ctx, "test message")

	stats := provider.GetBufferStats()
	if stats.BufferSize == 0 {
		t.Error("Expected buffer stats to show configured size")
	}
}

func TestZerologProvider_ContextFields(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	// Cria contexto com valores
	ctx := context.WithValue(context.Background(), logger.TraceIDKey, "trace-123")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-456")

	provider.Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verifica se os campos do contexto foram incluídos
	if logEntry["trace_id"] != "trace-123" {
		t.Errorf("Expected trace_id=trace-123, got %v", logEntry["trace_id"])
	}
	if logEntry["user_id"] != "user-456" {
		t.Errorf("Expected user_id=user-456, got %v", logEntry["user_id"])
	}
}

func TestZerologProvider_ServiceFields(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         &buf,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	provider.Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verifica campos de serviço
	if logEntry["service"] != "test-service" {
		t.Errorf("Expected service=test-service, got %v", logEntry["service"])
	}
	if logEntry["version"] != "1.0.0" {
		t.Errorf("Expected version=1.0.0, got %v", logEntry["version"])
	}
	if logEntry["environment"] != "test" {
		t.Errorf("Expected environment=test, got %v", logEntry["environment"])
	}
}

func TestZerologProvider_TimeFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:      logger.InfoLevel,
		Format:     logger.JSONFormat,
		Output:     &buf,
		TimeFormat: "2006-01-02T15:04:05Z07:00",
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	provider.Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verifica se o timestamp está presente (zerolog usa "time" como campo)
	timestampStr, ok := logEntry["time"].(string)
	if !ok {
		t.Error("Time field is not a string")
	}

	if _, err := time.Parse(time.RFC3339, timestampStr); err != nil {
		t.Errorf("Time format is incorrect: %v", err)
	}
}

func TestZerologProvider_ConsoleFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.ConsoleFormat,
		Output: &buf,
	}

	provider := NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	ctx := context.Background()
	provider.Info(ctx, "test message")

	if buf.Len() == 0 {
		t.Error("Expected output, got none")
	}

	// Para formato console, verificamos apenas que há saída
	// já que não é JSON
	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Output should contain the log message")
	}
}
