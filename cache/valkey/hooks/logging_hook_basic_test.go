package hooks

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoggingHookBasic(t *testing.T) {
	hook := NewLoggingHook()

	require.NotNil(t, hook)
	assert.True(t, hook.logExecution)
	assert.True(t, hook.logConnection)
	assert.True(t, hook.logPipeline)
	assert.True(t, hook.logRetry)
}

func TestLoggingHook_WithExecutionLoggingBasic(t *testing.T) {
	hook := NewLoggingHook()

	// Desabilitar logging de comandos
	hook = hook.WithExecutionLogging(false)
	assert.False(t, hook.logExecution)

	// Reabilitar logging de comandos
	hook = hook.WithExecutionLogging(true)
	assert.True(t, hook.logExecution)
}

func TestLoggingHook_ExecutionMethodsBasic(t *testing.T) {
	hook := NewLoggingHook()
	ctx := context.Background()
	duration := 100 * time.Millisecond

	// Test BeforeExecution
	resultCtx := hook.BeforeExecution(ctx, "GET", []interface{}{"key1"})
	assert.NotNil(t, resultCtx)

	// Test AfterExecution - successful
	hook.AfterExecution(ctx, "GET", []interface{}{"key1"}, "value1", nil, duration)

	// Test AfterExecution - with error
	hook.AfterExecution(ctx, "GET", []interface{}{"key1"}, nil, assert.AnError, duration)
}

func TestLoggingHook_ConnectionMethodsBasic(t *testing.T) {
	hook := NewLoggingHook()
	ctx := context.Background()
	duration := 50 * time.Millisecond

	// Test BeforeConnect
	resultCtx := hook.BeforeConnect(ctx, "tcp", "localhost:6379")
	assert.NotNil(t, resultCtx)

	// Test AfterConnect - successful
	hook.AfterConnect(ctx, "tcp", "localhost:6379", nil, duration)

	// Test AfterConnect - with error
	hook.AfterConnect(ctx, "tcp", "localhost:6379", assert.AnError, duration)
}

func TestLoggingHook_PipelineMethodsBasic(t *testing.T) {
	hook := NewLoggingHook()
	ctx := context.Background()
	commands := []string{"GET key1", "SET key2 value2"}
	duration := 200 * time.Millisecond

	// Test BeforePipelineExecution
	resultCtx := hook.BeforePipelineExecution(ctx, commands)
	assert.NotNil(t, resultCtx)

	// Test AfterPipelineExecution - successful
	hook.AfterPipelineExecution(ctx, commands, []interface{}{"value1", "OK"}, nil, duration)

	// Test AfterPipelineExecution - with error
	hook.AfterPipelineExecution(ctx, commands, nil, assert.AnError, duration)
}

func TestLoggingHook_RetryMethodsBasic(t *testing.T) {
	hook := NewLoggingHook()
	ctx := context.Background()

	// Test BeforeRetry
	resultCtx := hook.BeforeRetry(ctx, 1, assert.AnError)
	assert.NotNil(t, resultCtx)

	// Test AfterRetry - successful
	hook.AfterRetry(ctx, 1, true, nil)

	// Test AfterRetry - failed
	hook.AfterRetry(ctx, 1, false, assert.AnError)
}

func TestLoggingHook_ConfigurationBasic(t *testing.T) {
	hook := NewLoggingHook()

	// Test disabling all logging
	hook = hook.
		WithExecutionLogging(false).
		WithConnectionLogging(false).
		WithPipelineLogging(false).
		WithRetryLogging(false)

	assert.False(t, hook.logExecution)
	assert.False(t, hook.logConnection)
	assert.False(t, hook.logPipeline)
	assert.False(t, hook.logRetry)

	// Test re-enabling all logging
	hook = hook.
		WithExecutionLogging(true).
		WithConnectionLogging(true).
		WithPipelineLogging(true).
		WithRetryLogging(true)

	assert.True(t, hook.logExecution)
	assert.True(t, hook.logConnection)
	assert.True(t, hook.logPipeline)
	assert.True(t, hook.logRetry)
}

func BenchmarkLoggingHook_ExecutionBasic(b *testing.B) {
	hook := NewLoggingHook()
	ctx := context.Background()
	duration := 1 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hook.BeforeExecution(ctx, "GET", []interface{}{"key"})
		hook.AfterExecution(ctx, "GET", []interface{}{"key"}, "value", nil, duration)
	}
}
