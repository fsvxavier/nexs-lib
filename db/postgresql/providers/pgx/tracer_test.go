//go:build unit

package pgx

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/hooks"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHookManager for testing
type MockHookManager struct {
	mock.Mock
}

func (m *MockHookManager) RegisterHook(hookType interfaces.HookType, hook interfaces.Hook) error {
	args := m.Called(hookType, hook)
	return args.Error(0)
}

func (m *MockHookManager) RegisterCustomHook(hookType interfaces.HookType, name string, hook interfaces.Hook) error {
	args := m.Called(hookType, name, hook)
	return args.Error(0)
}

func (m *MockHookManager) ExecuteHooks(hookType interfaces.HookType, ctx *interfaces.ExecutionContext) error {
	args := m.Called(hookType, ctx)
	return args.Error(0)
}

func (m *MockHookManager) UnregisterHook(hookType interfaces.HookType) error {
	args := m.Called(hookType)
	return args.Error(0)
}

func (m *MockHookManager) UnregisterCustomHook(hookType interfaces.HookType, name string) error {
	args := m.Called(hookType, name)
	return args.Error(0)
}

func (m *MockHookManager) ListHooks() map[interfaces.HookType][]interfaces.Hook {
	args := m.Called()
	return args.Get(0).(map[interfaces.HookType][]interfaces.Hook)
}

func TestNewPGXTracer(t *testing.T) {
	hookManager := &MockHookManager{}
	tracer := NewPGXTracer(hookManager)

	assert.NotNil(t, tracer)
	assert.Equal(t, hookManager, tracer.hookManager)
	assert.True(t, tracer.enabled)
}

func TestPGXTracer_SetEnabled(t *testing.T) {
	hookManager := &MockHookManager{}
	tracer := NewPGXTracer(hookManager)

	tracer.SetEnabled(false)
	assert.False(t, tracer.IsEnabled())

	tracer.SetEnabled(true)
	assert.True(t, tracer.IsEnabled())
}

func TestPGXTracer_TraceQueryStart(t *testing.T) {
	hookManager := &MockHookManager{}
	tracer := NewPGXTracer(hookManager)

	ctx := context.Background()

	// Mock the ExecuteHooks call
	hookManager.On("ExecuteHooks", interfaces.BeforeQueryHook, mock.AnythingOfType("*interfaces.ExecutionContext")).Return(nil)

	data := pgx.TraceQueryStartData{
		SQL:  "SELECT * FROM users",
		Args: []interface{}{1, "test"},
	}

	resultCtx := tracer.TraceQueryStart(ctx, nil, data)

	assert.NotNil(t, resultCtx)

	// Verify execution context was stored
	execCtxVal := resultCtx.Value("execution_context")
	assert.NotNil(t, execCtxVal)

	execCtx, ok := execCtxVal.(*interfaces.ExecutionContext)
	assert.True(t, ok)
	assert.Equal(t, "query", execCtx.Operation)
	assert.Equal(t, "SELECT * FROM users", execCtx.Query)
	assert.Equal(t, []interface{}{1, "test"}, execCtx.Args)
	assert.NotNil(t, execCtx.Metadata)
	assert.True(t, execCtx.Metadata["trace_query_start"].(bool))

	hookManager.AssertExpectations(t)
}

func TestPGXTracer_TraceQueryStart_Disabled(t *testing.T) {
	hookManager := &MockHookManager{}
	tracer := NewPGXTracer(hookManager)
	tracer.SetEnabled(false)

	ctx := context.Background()
	data := pgx.TraceQueryStartData{
		SQL:  "SELECT * FROM users",
		Args: []interface{}{1, "test"},
	}

	resultCtx := tracer.TraceQueryStart(ctx, nil, data)

	// Should return original context when disabled
	assert.Equal(t, ctx, resultCtx)

	// Verify no execution context was stored
	execCtxVal := resultCtx.Value("execution_context")
	assert.Nil(t, execCtxVal)

	// Hook manager should not be called when disabled
	hookManager.AssertNotCalled(t, "ExecuteHooks")
}

func TestPGXTracer_TraceQueryEnd_Success(t *testing.T) {
	hookManager := &MockHookManager{}
	tracer := NewPGXTracer(hookManager)

	// Create execution context
	execCtx := &interfaces.ExecutionContext{
		Context:   context.Background(),
		Operation: "query",
		Query:     "SELECT * FROM users",
		StartTime: time.Now().Add(-100 * time.Millisecond),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.WithValue(context.Background(), "execution_context", execCtx)

	// Mock the ExecuteHooks call for success
	hookManager.On("ExecuteHooks", interfaces.AfterQueryHook, mock.AnythingOfType("*interfaces.ExecutionContext")).Return(nil)

	data := pgx.TraceQueryEndData{
		Err: nil,
	}

	tracer.TraceQueryEnd(ctx, nil, data)

	// Verify execution context was updated
	assert.Greater(t, execCtx.Duration, time.Duration(0))
	assert.Nil(t, execCtx.Error)
	assert.True(t, execCtx.Metadata["trace_query_success"].(bool))

	hookManager.AssertExpectations(t)
}

func TestPGXTracer_TraceQueryEnd_Error(t *testing.T) {
	hookManager := &MockHookManager{}
	tracer := NewPGXTracer(hookManager)

	// Create execution context
	execCtx := &interfaces.ExecutionContext{
		Context:   context.Background(),
		Operation: "query",
		Query:     "SELECT * FROM users",
		StartTime: time.Now().Add(-100 * time.Millisecond),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.WithValue(context.Background(), "execution_context", execCtx)

	// Mock the ExecuteHooks call for error
	hookManager.On("ExecuteHooks", interfaces.OnErrorHook, mock.AnythingOfType("*interfaces.ExecutionContext")).Return(nil)

	testErr := assert.AnError
	data := pgx.TraceQueryEndData{
		Err: testErr,
	}

	tracer.TraceQueryEnd(ctx, nil, data)

	// Verify execution context was updated
	assert.Greater(t, execCtx.Duration, time.Duration(0))
	assert.Equal(t, testErr, execCtx.Error)
	assert.True(t, execCtx.Metadata["trace_query_error"].(bool))

	hookManager.AssertExpectations(t)
}

func TestPGXTracer_TraceQueryEnd_NoExecutionContext(t *testing.T) {
	hookManager := &MockHookManager{}
	tracer := NewPGXTracer(hookManager)

	ctx := context.Background() // No execution context stored

	data := pgx.TraceQueryEndData{
		Err: nil,
	}

	// Should not panic and not call hooks
	tracer.TraceQueryEnd(ctx, nil, data)

	hookManager.AssertNotCalled(t, "ExecuteHooks")
}

func TestPGXTracer_TraceBatchStart(t *testing.T) {
	hookManager := &MockHookManager{}
	tracer := NewPGXTracer(hookManager)

	ctx := context.Background()

	// Mock the ExecuteHooks call
	hookManager.On("ExecuteHooks", interfaces.BeforeBatchHook, mock.AnythingOfType("*interfaces.ExecutionContext")).Return(nil)

	// Create a mock batch
	batch := &pgx.Batch{}
	batch.Queue("SELECT 1")
	batch.Queue("SELECT 2")

	data := pgx.TraceBatchStartData{
		Batch: batch,
	}

	resultCtx := tracer.TraceBatchStart(ctx, nil, data)

	assert.NotNil(t, resultCtx)

	// Verify execution context was stored
	execCtxVal := resultCtx.Value("batch_execution_context")
	assert.NotNil(t, execCtxVal)

	execCtx, ok := execCtxVal.(*interfaces.ExecutionContext)
	assert.True(t, ok)
	assert.Equal(t, "batch", execCtx.Operation)
	assert.NotNil(t, execCtx.Metadata)
	assert.True(t, execCtx.Metadata["trace_batch_start"].(bool))
	assert.Equal(t, 2, execCtx.Metadata["batch_size"].(int))

	hookManager.AssertExpectations(t)
}

func TestPGXTracer_Integration(t *testing.T) {
	// Test integration with real hook manager
	hookManager := hooks.NewDefaultHookManager()
	tracer := NewPGXTracer(hookManager)

	// Register a test hook
	var capturedOperations []string
	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		capturedOperations = append(capturedOperations, ctx.Operation)
		return &interfaces.HookResult{Continue: true}
	}

	err := hookManager.RegisterHook(interfaces.BeforeQueryHook, testHook)
	assert.NoError(t, err)

	// Trace a query
	ctx := context.Background()
	data := pgx.TraceQueryStartData{
		SQL:  "SELECT * FROM users",
		Args: []interface{}{},
	}

	resultCtx := tracer.TraceQueryStart(ctx, nil, data)
	assert.NotNil(t, resultCtx)

	// Verify hook was called
	assert.Len(t, capturedOperations, 1)
	assert.Equal(t, "query", capturedOperations[0])
}
