package hooks

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExecutionHook implementa ExecutionHook para testes.
type MockExecutionHook struct {
	beforeCalls []BeforeExecutionCall
	afterCalls  []AfterExecutionCall
}

type BeforeExecutionCall struct {
	Ctx  context.Context
	Cmd  string
	Args []interface{}
}

type AfterExecutionCall struct {
	Ctx      context.Context
	Cmd      string
	Args     []interface{}
	Result   interface{}
	Err      error
	Duration time.Duration
}

func (m *MockExecutionHook) BeforeExecution(ctx context.Context, cmd string, args []interface{}) context.Context {
	m.beforeCalls = append(m.beforeCalls, BeforeExecutionCall{
		Ctx:  ctx,
		Cmd:  cmd,
		Args: args,
	})
	return ctx
}

func (m *MockExecutionHook) AfterExecution(ctx context.Context, cmd string, args []interface{}, result interface{}, err error, duration time.Duration) {
	m.afterCalls = append(m.afterCalls, AfterExecutionCall{
		Ctx:      ctx,
		Cmd:      cmd,
		Args:     args,
		Result:   result,
		Err:      err,
		Duration: duration,
	})
}

// MockConnectionHook implementa ConnectionHook para testes.
type MockConnectionHook struct {
	beforeConnectCalls    []BeforeConnectCall
	afterConnectCalls     []AfterConnectCall
	beforeDisconnectCalls []BeforeDisconnectCall
	afterDisconnectCalls  []AfterDisconnectCall
}

type BeforeConnectCall struct {
	Ctx     context.Context
	Network string
	Addr    string
}

type AfterConnectCall struct {
	Ctx      context.Context
	Network  string
	Addr     string
	Err      error
	Duration time.Duration
}

type BeforeDisconnectCall struct {
	Ctx     context.Context
	Network string
	Addr    string
}

type AfterDisconnectCall struct {
	Ctx      context.Context
	Network  string
	Addr     string
	Err      error
	Duration time.Duration
}

func (m *MockConnectionHook) BeforeConnect(ctx context.Context, network, addr string) context.Context {
	m.beforeConnectCalls = append(m.beforeConnectCalls, BeforeConnectCall{
		Ctx:     ctx,
		Network: network,
		Addr:    addr,
	})
	return ctx
}

func (m *MockConnectionHook) AfterConnect(ctx context.Context, network, addr string, err error, duration time.Duration) {
	m.afterConnectCalls = append(m.afterConnectCalls, AfterConnectCall{
		Ctx:      ctx,
		Network:  network,
		Addr:     addr,
		Err:      err,
		Duration: duration,
	})
}

func (m *MockConnectionHook) BeforeDisconnect(ctx context.Context, network, addr string) context.Context {
	m.beforeDisconnectCalls = append(m.beforeDisconnectCalls, BeforeDisconnectCall{
		Ctx:     ctx,
		Network: network,
		Addr:    addr,
	})
	return ctx
}

func (m *MockConnectionHook) AfterDisconnect(ctx context.Context, network, addr string, err error, duration time.Duration) {
	m.afterDisconnectCalls = append(m.afterDisconnectCalls, AfterDisconnectCall{
		Ctx:      ctx,
		Network:  network,
		Addr:     addr,
		Err:      err,
		Duration: duration,
	})
}

// MockPipelineHook implementa PipelineHook para testes.
type MockPipelineHook struct {
	beforeCalls []BeforePipelineCall
	afterCalls  []AfterPipelineCall
}

type BeforePipelineCall struct {
	Ctx      context.Context
	Commands []string
}

type AfterPipelineCall struct {
	Ctx      context.Context
	Commands []string
	Results  []interface{}
	Err      error
	Duration time.Duration
}

func (m *MockPipelineHook) BeforePipelineExecution(ctx context.Context, commands []string) context.Context {
	m.beforeCalls = append(m.beforeCalls, BeforePipelineCall{
		Ctx:      ctx,
		Commands: commands,
	})
	return ctx
}

func (m *MockPipelineHook) AfterPipelineExecution(ctx context.Context, commands []string, results []interface{}, err error, duration time.Duration) {
	m.afterCalls = append(m.afterCalls, AfterPipelineCall{
		Ctx:      ctx,
		Commands: commands,
		Results:  results,
		Err:      err,
		Duration: duration,
	})
}

// MockRetryHook implementa RetryHook para testes.
type MockRetryHook struct {
	beforeCalls []BeforeRetryCall
	afterCalls  []AfterRetryCall
}

type BeforeRetryCall struct {
	Ctx     context.Context
	Attempt int
	Err     error
}

type AfterRetryCall struct {
	Ctx     context.Context
	Attempt int
	Success bool
	Err     error
}

func (m *MockRetryHook) BeforeRetry(ctx context.Context, attempt int, err error) context.Context {
	m.beforeCalls = append(m.beforeCalls, BeforeRetryCall{
		Ctx:     ctx,
		Attempt: attempt,
		Err:     err,
	})
	return ctx
}

func (m *MockRetryHook) AfterRetry(ctx context.Context, attempt int, success bool, err error) {
	m.afterCalls = append(m.afterCalls, AfterRetryCall{
		Ctx:     ctx,
		Attempt: attempt,
		Success: success,
		Err:     err,
	})
}

func TestNewCompositeHook(t *testing.T) {
	hook := NewCompositeHook()
	require.NotNil(t, hook)
	assert.Empty(t, hook.executionHooks)
	assert.Empty(t, hook.connectionHooks)
	assert.Empty(t, hook.pipelineHooks)
	assert.Empty(t, hook.retryHooks)
}

func TestCompositeHook_AddExecutionHook(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockExecutionHook{}

	hook.AddExecutionHook(mockHook)

	assert.Len(t, hook.executionHooks, 1)
	assert.Equal(t, mockHook, hook.executionHooks[0])
}

func TestCompositeHook_AddConnectionHook(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockConnectionHook{}

	hook.AddConnectionHook(mockHook)

	assert.Len(t, hook.connectionHooks, 1)
	assert.Equal(t, mockHook, hook.connectionHooks[0])
}

func TestCompositeHook_AddPipelineHook(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockPipelineHook{}

	hook.AddPipelineHook(mockHook)

	assert.Len(t, hook.pipelineHooks, 1)
	assert.Equal(t, mockHook, hook.pipelineHooks[0])
}

func TestCompositeHook_AddRetryHook(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockRetryHook{}

	hook.AddRetryHook(mockHook)

	assert.Len(t, hook.retryHooks, 1)
	assert.Equal(t, mockHook, hook.retryHooks[0])
}

func TestCompositeHook_BeforeExecution(t *testing.T) {
	hook := NewCompositeHook()
	mockHook1 := &MockExecutionHook{}
	mockHook2 := &MockExecutionHook{}

	hook.AddExecutionHook(mockHook1)
	hook.AddExecutionHook(mockHook2)

	ctx := context.Background()
	cmd := "GET"
	args := []interface{}{"key1"}

	resultCtx := hook.BeforeExecution(ctx, cmd, args)

	assert.Equal(t, ctx, resultCtx)
	assert.Len(t, mockHook1.beforeCalls, 1)
	assert.Len(t, mockHook2.beforeCalls, 1)

	assert.Equal(t, cmd, mockHook1.beforeCalls[0].Cmd)
	assert.Equal(t, args, mockHook1.beforeCalls[0].Args)
	assert.Equal(t, cmd, mockHook2.beforeCalls[0].Cmd)
	assert.Equal(t, args, mockHook2.beforeCalls[0].Args)
}

func TestCompositeHook_AfterExecution(t *testing.T) {
	hook := NewCompositeHook()
	mockHook1 := &MockExecutionHook{}
	mockHook2 := &MockExecutionHook{}

	hook.AddExecutionHook(mockHook1)
	hook.AddExecutionHook(mockHook2)

	ctx := context.Background()
	cmd := "SET"
	args := []interface{}{"key1", "value1"}
	result := "OK"
	err := assert.AnError
	duration := 100 * time.Millisecond

	hook.AfterExecution(ctx, cmd, args, result, err, duration)

	assert.Len(t, mockHook1.afterCalls, 1)
	assert.Len(t, mockHook2.afterCalls, 1)

	assert.Equal(t, cmd, mockHook1.afterCalls[0].Cmd)
	assert.Equal(t, args, mockHook1.afterCalls[0].Args)
	assert.Equal(t, result, mockHook1.afterCalls[0].Result)
	assert.Equal(t, err, mockHook1.afterCalls[0].Err)
	assert.Equal(t, duration, mockHook1.afterCalls[0].Duration)
}

func TestCompositeHook_BeforeConnect(t *testing.T) {
	hook := NewCompositeHook()
	mockHook1 := &MockConnectionHook{}
	mockHook2 := &MockConnectionHook{}

	hook.AddConnectionHook(mockHook1)
	hook.AddConnectionHook(mockHook2)

	ctx := context.Background()
	network := "tcp"
	addr := "localhost:6379"

	resultCtx := hook.BeforeConnect(ctx, network, addr)

	assert.Equal(t, ctx, resultCtx)
	assert.Len(t, mockHook1.beforeConnectCalls, 1)
	assert.Len(t, mockHook2.beforeConnectCalls, 1)

	assert.Equal(t, network, mockHook1.beforeConnectCalls[0].Network)
	assert.Equal(t, addr, mockHook1.beforeConnectCalls[0].Addr)
}

func TestCompositeHook_AfterConnect(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockConnectionHook{}

	hook.AddConnectionHook(mockHook)

	ctx := context.Background()
	network := "tcp"
	addr := "localhost:6379"
	err := assert.AnError
	duration := 50 * time.Millisecond

	hook.AfterConnect(ctx, network, addr, err, duration)

	assert.Len(t, mockHook.afterConnectCalls, 1)
	assert.Equal(t, network, mockHook.afterConnectCalls[0].Network)
	assert.Equal(t, addr, mockHook.afterConnectCalls[0].Addr)
	assert.Equal(t, err, mockHook.afterConnectCalls[0].Err)
	assert.Equal(t, duration, mockHook.afterConnectCalls[0].Duration)
}

func TestCompositeHook_BeforeDisconnect(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockConnectionHook{}

	hook.AddConnectionHook(mockHook)

	ctx := context.Background()
	network := "tcp"
	addr := "localhost:6379"

	resultCtx := hook.BeforeDisconnect(ctx, network, addr)

	assert.Equal(t, ctx, resultCtx)
	assert.Len(t, mockHook.beforeDisconnectCalls, 1)
	assert.Equal(t, network, mockHook.beforeDisconnectCalls[0].Network)
	assert.Equal(t, addr, mockHook.beforeDisconnectCalls[0].Addr)
}

func TestCompositeHook_AfterDisconnect(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockConnectionHook{}

	hook.AddConnectionHook(mockHook)

	ctx := context.Background()
	network := "tcp"
	addr := "localhost:6379"
	err := assert.AnError
	duration := 25 * time.Millisecond

	hook.AfterDisconnect(ctx, network, addr, err, duration)

	assert.Len(t, mockHook.afterDisconnectCalls, 1)
	assert.Equal(t, network, mockHook.afterDisconnectCalls[0].Network)
	assert.Equal(t, addr, mockHook.afterDisconnectCalls[0].Addr)
	assert.Equal(t, err, mockHook.afterDisconnectCalls[0].Err)
	assert.Equal(t, duration, mockHook.afterDisconnectCalls[0].Duration)
}

func TestCompositeHook_BeforePipelineExecution(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockPipelineHook{}

	hook.AddPipelineHook(mockHook)

	ctx := context.Background()
	commands := []string{"GET key1", "SET key2 value2"}

	resultCtx := hook.BeforePipelineExecution(ctx, commands)

	assert.Equal(t, ctx, resultCtx)
	assert.Len(t, mockHook.beforeCalls, 1)
	assert.Equal(t, commands, mockHook.beforeCalls[0].Commands)
}

func TestCompositeHook_AfterPipelineExecution(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockPipelineHook{}

	hook.AddPipelineHook(mockHook)

	ctx := context.Background()
	commands := []string{"GET key1", "SET key2 value2"}
	results := []interface{}{"value1", "OK"}
	err := assert.AnError
	duration := 150 * time.Millisecond

	hook.AfterPipelineExecution(ctx, commands, results, err, duration)

	assert.Len(t, mockHook.afterCalls, 1)
	assert.Equal(t, commands, mockHook.afterCalls[0].Commands)
	assert.Equal(t, results, mockHook.afterCalls[0].Results)
	assert.Equal(t, err, mockHook.afterCalls[0].Err)
	assert.Equal(t, duration, mockHook.afterCalls[0].Duration)
}

func TestCompositeHook_BeforeRetry(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockRetryHook{}

	hook.AddRetryHook(mockHook)

	ctx := context.Background()
	attempt := 2
	err := assert.AnError

	resultCtx := hook.BeforeRetry(ctx, attempt, err)

	assert.Equal(t, ctx, resultCtx)
	assert.Len(t, mockHook.beforeCalls, 1)
	assert.Equal(t, attempt, mockHook.beforeCalls[0].Attempt)
	assert.Equal(t, err, mockHook.beforeCalls[0].Err)
}

func TestCompositeHook_AfterRetry(t *testing.T) {
	hook := NewCompositeHook()
	mockHook := &MockRetryHook{}

	hook.AddRetryHook(mockHook)

	ctx := context.Background()
	attempt := 3
	success := true
	err := assert.AnError

	hook.AfterRetry(ctx, attempt, success, err)

	assert.Len(t, mockHook.afterCalls, 1)
	assert.Equal(t, attempt, mockHook.afterCalls[0].Attempt)
	assert.Equal(t, success, mockHook.afterCalls[0].Success)
	assert.Equal(t, err, mockHook.afterCalls[0].Err)
}

func TestCompositeHook_MultipleHooks(t *testing.T) {
	hook := NewCompositeHook()

	execHook1 := &MockExecutionHook{}
	execHook2 := &MockExecutionHook{}
	connHook := &MockConnectionHook{}
	pipelineHook := &MockPipelineHook{}
	retryHook := &MockRetryHook{}

	hook.AddExecutionHook(execHook1)
	hook.AddExecutionHook(execHook2)
	hook.AddConnectionHook(connHook)
	hook.AddPipelineHook(pipelineHook)
	hook.AddRetryHook(retryHook)

	ctx := context.Background()

	// Test execution hooks
	hook.BeforeExecution(ctx, "GET", []interface{}{"key"})
	hook.AfterExecution(ctx, "GET", []interface{}{"key"}, "value", nil, time.Millisecond)

	assert.Len(t, execHook1.beforeCalls, 1)
	assert.Len(t, execHook1.afterCalls, 1)
	assert.Len(t, execHook2.beforeCalls, 1)
	assert.Len(t, execHook2.afterCalls, 1)

	// Test connection hooks
	hook.BeforeConnect(ctx, "tcp", "localhost:6379")
	hook.AfterConnect(ctx, "tcp", "localhost:6379", nil, time.Millisecond)
	hook.BeforeDisconnect(ctx, "tcp", "localhost:6379")
	hook.AfterDisconnect(ctx, "tcp", "localhost:6379", nil, time.Millisecond)

	assert.Len(t, connHook.beforeConnectCalls, 1)
	assert.Len(t, connHook.afterConnectCalls, 1)
	assert.Len(t, connHook.beforeDisconnectCalls, 1)
	assert.Len(t, connHook.afterDisconnectCalls, 1)

	// Test pipeline hooks
	hook.BeforePipelineExecution(ctx, []string{"GET key"})
	hook.AfterPipelineExecution(ctx, []string{"GET key"}, []interface{}{"value"}, nil, time.Millisecond)

	assert.Len(t, pipelineHook.beforeCalls, 1)
	assert.Len(t, pipelineHook.afterCalls, 1)

	// Test retry hooks
	hook.BeforeRetry(ctx, 1, assert.AnError)
	hook.AfterRetry(ctx, 1, false, assert.AnError)

	assert.Len(t, retryHook.beforeCalls, 1)
	assert.Len(t, retryHook.afterCalls, 1)
}

func TestCompositeHook_EmptyHooks(t *testing.T) {
	hook := NewCompositeHook()
	ctx := context.Background()

	// Should not panic with empty hooks
	resultCtx := hook.BeforeExecution(ctx, "GET", []interface{}{"key"})
	assert.Equal(t, ctx, resultCtx)

	hook.AfterExecution(ctx, "GET", []interface{}{"key"}, "value", nil, time.Millisecond)

	resultCtx = hook.BeforeConnect(ctx, "tcp", "localhost:6379")
	assert.Equal(t, ctx, resultCtx)

	hook.AfterConnect(ctx, "tcp", "localhost:6379", nil, time.Millisecond)

	resultCtx = hook.BeforeDisconnect(ctx, "tcp", "localhost:6379")
	assert.Equal(t, ctx, resultCtx)

	hook.AfterDisconnect(ctx, "tcp", "localhost:6379", nil, time.Millisecond)

	resultCtx = hook.BeforePipelineExecution(ctx, []string{"GET key"})
	assert.Equal(t, ctx, resultCtx)

	hook.AfterPipelineExecution(ctx, []string{"GET key"}, []interface{}{"value"}, nil, time.Millisecond)

	resultCtx = hook.BeforeRetry(ctx, 1, assert.AnError)
	assert.Equal(t, ctx, resultCtx)

	hook.AfterRetry(ctx, 1, false, assert.AnError)
}

func TestCompositeHook_ContextPropagation(t *testing.T) {
	hook := NewCompositeHook()

	// Hook que modifica o context
	modifyingHook := &ContextModifyingHook{}
	hook.AddExecutionHook(modifyingHook)

	ctx := context.Background()
	resultCtx := hook.BeforeExecution(ctx, "GET", []interface{}{"key"})

	// Verificar que o context foi modificado
	assert.NotEqual(t, ctx, resultCtx)
	assert.Equal(t, "modified", resultCtx.Value("key"))
}

// ContextModifyingHook é um hook que modifica o context para testar propagação
type ContextModifyingHook struct{}

func (h *ContextModifyingHook) BeforeExecution(ctx context.Context, cmd string, args []interface{}) context.Context {
	return context.WithValue(ctx, "key", "modified")
}

func (h *ContextModifyingHook) AfterExecution(ctx context.Context, cmd string, args []interface{}, result interface{}, err error, duration time.Duration) {
	// No-op
}

func BenchmarkCompositeHook_BeforeExecution(b *testing.B) {
	hook := NewCompositeHook()
	for i := 0; i < 10; i++ {
		hook.AddExecutionHook(&MockExecutionHook{})
	}

	ctx := context.Background()
	cmd := "GET"
	args := []interface{}{"key"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hook.BeforeExecution(ctx, cmd, args)
	}
}

func BenchmarkCompositeHook_AfterExecution(b *testing.B) {
	hook := NewCompositeHook()
	for i := 0; i < 10; i++ {
		hook.AddExecutionHook(&MockExecutionHook{})
	}

	ctx := context.Background()
	cmd := "GET"
	args := []interface{}{"key"}
	result := "value"
	duration := time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hook.AfterExecution(ctx, cmd, args, result, nil, duration)
	}
}
