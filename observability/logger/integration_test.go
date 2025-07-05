package logger_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func TestProviderIntegration(t *testing.T) {
	// Testa se todos os providers foram registrados corretamente
	providers := logger.ListProviders()

	expectedProviders := []string{"slog", "zap", "zerolog"}
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
}

func TestSlogProvider(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         &buf,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider slog: %v", err)
	}

	ctx := context.Background()
	logger.Info(ctx, "test message", logger.String("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Expected test message in output")
	}
	if !strings.Contains(output, "test-service") {
		t.Error("Expected service name in output")
	}
}

func TestZapProvider(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.ConsoleFormat,
		Output:         &buf,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
	}

	err := logger.SetProvider("zap", config)
	if err != nil {
		t.Fatalf("Failed to set provider zap: %v", err)
	}

	ctx := context.Background()
	logger.Info(ctx, "test message", logger.String("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Expected test message in output")
	}
}
