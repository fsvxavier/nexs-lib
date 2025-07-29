package hooks

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsHookBasic(t *testing.T) {
	hook := NewMetricsHook()

	require.NotNil(t, hook)
	assert.True(t, hook.collectExecutionMetrics)
	assert.True(t, hook.collectConnectionMetrics)
	assert.True(t, hook.collectPipelineMetrics)
	assert.True(t, hook.collectRetryMetrics)
	assert.Equal(t, 1000, hook.maxHistorySize)
}

func TestMetricsHook_WithExecutionMetricsBasic(t *testing.T) {
	hook := NewMetricsHook()

	// Desabilitar coleta de métricas de execução
	hook = hook.WithExecutionMetrics(false)
	assert.False(t, hook.collectExecutionMetrics)

	// Reabilitar coleta de métricas de execução
	hook = hook.WithExecutionMetrics(true)
	assert.True(t, hook.collectExecutionMetrics)
}

func TestMetricsHook_WithConnectionMetricsBasic(t *testing.T) {
	hook := NewMetricsHook()

	// Desabilitar coleta de métricas de conexão
	hook = hook.WithConnectionMetrics(false)
	assert.False(t, hook.collectConnectionMetrics)

	// Reabilitar coleta de métricas de conexão
	hook = hook.WithConnectionMetrics(true)
	assert.True(t, hook.collectConnectionMetrics)
}

func TestMetricsHook_WithPipelineMetricsBasic(t *testing.T) {
	hook := NewMetricsHook()

	// Desabilitar coleta de métricas de pipeline
	hook = hook.WithPipelineMetrics(false)
	assert.False(t, hook.collectPipelineMetrics)

	// Reabilitar coleta de métricas de pipeline
	hook = hook.WithPipelineMetrics(true)
	assert.True(t, hook.collectPipelineMetrics)
}

func TestMetricsHook_WithRetryMetricsBasic(t *testing.T) {
	hook := NewMetricsHook()

	// Desabilitar coleta de métricas de retry
	hook = hook.WithRetryMetrics(false)
	assert.False(t, hook.collectRetryMetrics)

	// Reabilitar coleta de métricas de retry
	hook = hook.WithRetryMetrics(true)
	assert.True(t, hook.collectRetryMetrics)
}

func TestMetricsHook_WithMaxHistorySizeBasic(t *testing.T) {
	hook := NewMetricsHook()

	// Alterar tamanho máximo do histórico
	hook = hook.WithMaxHistorySize(500)
	assert.Equal(t, 500, hook.maxHistorySize)

	// Testar valor zero (pode manter como 0 ou usar valor padrão)
	hook = hook.WithMaxHistorySize(0)
	assert.Equal(t, 0, hook.maxHistorySize) // Aceitar como está implementado
}

func TestMetricsHook_ExecutionMetricsBasic(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()
	duration := 100 * time.Millisecond

	// Simular execuções
	hook.AfterExecution(ctx, "GET", []interface{}{"key1"}, "value1", nil, duration)
	hook.AfterExecution(ctx, "SET", []interface{}{"key2", "value2"}, "OK", nil, 50*time.Millisecond)
	hook.AfterExecution(ctx, "DEL", []interface{}{"key3"}, nil, assert.AnError, 25*time.Millisecond)

	metrics := hook.GetMetrics()
	assert.Equal(t, int64(3), metrics.CommandsExecuted)
	assert.Equal(t, int64(1), metrics.ErrorsOccurred)
	assert.True(t, metrics.AvgExecutionTime > 0)
}

func TestMetricsHook_ConnectionMetricsBasic(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()
	duration := 50 * time.Millisecond

	// Simular conexões
	hook.AfterConnect(ctx, "tcp", "localhost:6379", nil, duration)
	hook.AfterConnect(ctx, "tcp", "localhost:6380", assert.AnError, duration)
	hook.AfterDisconnect(ctx, "tcp", "localhost:6379", nil, 10*time.Millisecond)

	metrics := hook.GetMetrics()
	assert.Equal(t, int64(2), metrics.ConnectionsOpened)
	assert.Equal(t, int64(1), metrics.ConnectionsClosed)
	assert.True(t, metrics.AvgConnectionTime > 0)
}

func TestMetricsHook_PipelineMetricsBasic(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()
	commands := []string{"GET key1", "SET key2 value2"}
	duration := 200 * time.Millisecond

	// Simular pipelines
	hook.AfterPipelineExecution(ctx, commands, []interface{}{"value1", "OK"}, nil, duration)
	hook.AfterPipelineExecution(ctx, commands, nil, assert.AnError, 150*time.Millisecond)

	metrics := hook.GetMetrics()
	assert.Equal(t, int64(2), metrics.PipelinesExecuted)
	assert.True(t, metrics.AvgPipelineTime > 0)
}

func TestMetricsHook_RetryMetricsBasic(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()

	// Simular retries
	hook.AfterRetry(ctx, 1, true, nil)
	hook.AfterRetry(ctx, 2, false, assert.AnError)
	hook.AfterRetry(ctx, 3, true, nil)

	metrics := hook.GetMetrics()
	assert.Equal(t, int64(3), metrics.RetriesAttempted)
}

func TestMetricsHook_ConcurrentAccessBasic(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()
	duration := 10 * time.Millisecond

	// Test concurrent access to metrics collection
	done := make(chan bool, 3)

	// Concurrent execution metrics
	go func() {
		for i := 0; i < 100; i++ {
			hook.AfterExecution(ctx, "GET", []interface{}{"key"}, "value", nil, duration)
		}
		done <- true
	}()

	// Concurrent connection metrics
	go func() {
		for i := 0; i < 100; i++ {
			hook.AfterConnect(ctx, "tcp", "localhost:6379", nil, duration)
		}
		done <- true
	}()

	// Concurrent metrics reading
	go func() {
		for i := 0; i < 100; i++ {
			_ = hook.GetMetrics()
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify final metrics make sense
	metrics := hook.GetMetrics()
	assert.Equal(t, int64(100), metrics.CommandsExecuted)
	assert.Equal(t, int64(100), metrics.ConnectionsOpened)
}

func BenchmarkMetricsHook_ExecutionBasic(b *testing.B) {
	hook := NewMetricsHook()
	ctx := context.Background()
	duration := 1 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hook.AfterExecution(ctx, "GET", []interface{}{"key"}, "value", nil, duration)
	}
}

func BenchmarkMetricsHook_GetMetricsBasic(b *testing.B) {
	hook := NewMetricsHook()
	ctx := context.Background()
	duration := 1 * time.Millisecond

	// Pre-populate some metrics
	for i := 0; i < 100; i++ {
		hook.AfterExecution(ctx, "GET", []interface{}{"key"}, "value", nil, duration)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hook.GetMetrics()
	}
}
