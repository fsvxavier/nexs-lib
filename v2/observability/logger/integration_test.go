//go:build integration
// +build integration

package logger

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/providers/zap"
)

// TestIntegration_ConcurrentLogging testa logging concorrente em alta escala
func TestIntegration_ConcurrentLogging(t *testing.T) {
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		ServiceName: "integration-test",
		Environment: "test",
		Async: &interfaces.AsyncConfig{
			Enabled:       true,
			BufferSize:    1000,
			Workers:       4,
			FlushInterval: 100 * time.Millisecond,
		},
	}

	provider := zap.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	// Test concurrent logging
	const (
		numGoroutines    = 10
		logsPerGoroutine = 50
	)

	var wg sync.WaitGroup
	ctx := context.Background()

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Info(ctx, "Concurrent log message")
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Force flush all pending messages
	err = logger.Flush()
	if err != nil {
		t.Errorf("Flush failed: %v", err)
	}

	t.Logf("Logged %d messages in %v", numGoroutines*logsPerGoroutine, duration)
}

// TestIntegration_ProviderSwitching testa mudança de provider em runtime
func TestIntegration_ProviderSwitching(t *testing.T) {
	zapProvider := zap.NewProvider()
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		ServiceName: "integration-test",
		Environment: "test",
	}

	err := zapProvider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure zap provider: %v", err)
	}

	logger := NewCoreLogger(zapProvider, config)
	defer logger.Close()

	ctx := context.Background()
	logger.Info(ctx, "Message with zap provider")

	// Create mock provider
	mockProvider := NewMockProvider("test", "1.0.0")
	mockProvider.Configure(config)

	mockLogger := NewCoreLogger(mockProvider, config)
	defer mockLogger.Close()

	mockLogger.Info(ctx, "Message with mock provider")
}

// TestIntegration_SamplingUnderLoad testa sampling sob alta carga
func TestIntegration_SamplingUnderLoad(t *testing.T) {
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		ServiceName: "integration-test",
		Environment: "test",
		Sampling: &interfaces.SamplingConfig{
			Enabled:    true,
			Initial:    10,
			Thereafter: 10,
			Levels: []interfaces.Level{
				interfaces.InfoLevel,
				interfaces.DebugLevel,
			},
		},
	}

	provider := zap.NewProvider()
	err := provider.Configure(config)
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	ctx := context.Background()
	const totalMessages = 100 // Reduzido para teste mais rápido

	for i := 0; i < totalMessages; i++ {
		logger.Info(ctx, "High volume message")
	}

	err = logger.Flush()
	if err != nil {
		t.Errorf("Flush failed: %v", err)
	}
}
