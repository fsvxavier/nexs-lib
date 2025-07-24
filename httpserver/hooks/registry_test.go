package hooks

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestDefaultHookRegistry_Register(t *testing.T) {
	registry := NewDefaultHookRegistry()
	hook := NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1)

	err := registry.Register(hook, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	hooks := registry.GetHooks(interfaces.HookEventRequestStart)
	if len(hooks) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(hooks))
	}

	if hooks[0].Name() != "test" {
		t.Fatalf("Expected hook name 'test', got '%s'", hooks[0].Name())
	}
}

func TestDefaultHookRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewDefaultHookRegistry()
	hook1 := NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1)
	hook2 := NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 2)

	err := registry.Register(hook1, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = registry.Register(hook2, interfaces.HookEventRequestStart)
	if err == nil {
		t.Fatal("Expected error for duplicate hook registration")
	}
}

func TestDefaultHookRegistry_Unregister(t *testing.T) {
	registry := NewDefaultHookRegistry()
	hook := NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1)

	err := registry.Register(hook, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = registry.Unregister("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	hooks := registry.GetHooks(interfaces.HookEventRequestStart)
	if len(hooks) != 0 {
		t.Fatalf("Expected 0 hooks after unregister, got %d", len(hooks))
	}
}

func TestDefaultHookRegistry_Execute(t *testing.T) {
	registry := NewDefaultHookRegistry()

	// Create a test hook that sets a flag
	executed := false
	hook := &testHook{
		BaseHook: NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			executed = true
			return nil
		},
	}

	err := registry.Register(hook, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err = registry.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !executed {
		t.Fatal("Expected hook to be executed")
	}
}

func TestDefaultHookRegistry_ExecuteWithError(t *testing.T) {
	registry := NewDefaultHookRegistry()

	expectedError := errors.New("test error")
	hook := &testHook{
		BaseHook: NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			return expectedError
		},
	}

	err := registry.Register(hook, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err = registry.Execute(ctx)
	if err != expectedError {
		t.Fatalf("Expected error %v, got %v", expectedError, err)
	}
}

func TestDefaultHookRegistry_ExecuteParallel(t *testing.T) {
	registry := NewDefaultHookRegistry()

	executed1 := false
	executed2 := false

	hook1 := &testHook{
		BaseHook: NewBaseHook("test1", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			time.Sleep(10 * time.Millisecond)
			executed1 = true
			return nil
		},
	}

	hook2 := &testHook{
		BaseHook: NewBaseHook("test2", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			time.Sleep(10 * time.Millisecond)
			executed2 = true
			return nil
		},
	}

	hooks := []interfaces.Hook{hook1, hook2}

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	start := time.Now()
	err := registry.ExecuteParallel(hooks, ctx)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !executed1 || !executed2 {
		t.Fatal("Expected both hooks to be executed")
	}

	// Parallel execution should be faster than sequential
	if duration > 15*time.Millisecond {
		t.Fatalf("Expected parallel execution to be faster, took %v", duration)
	}
}

func TestDefaultHookRegistry_ExecuteWithTimeout(t *testing.T) {
	registry := NewDefaultHookRegistry()

	hook := &testHook{
		BaseHook: NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			time.Sleep(50 * time.Millisecond) // Sleep longer than timeout
			return nil
		},
	}

	hooks := []interfaces.Hook{hook}

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err := registry.ExecuteWithTimeout(hooks, ctx, 10*time.Millisecond)
	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if err.Error() != "hook execution timed out after 10ms" {
		t.Fatalf("Expected timeout error message, got %v", err)
	}
}

func TestDefaultHookRegistry_EnableDisableHook(t *testing.T) {
	registry := NewDefaultHookRegistry()
	hook := NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1)

	err := registry.Register(hook, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Hook should be enabled by default
	if !hook.IsEnabled() {
		t.Fatal("Expected hook to be enabled by default")
	}

	// Disable hook
	err = registry.DisableHook("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hook.IsEnabled() {
		t.Fatal("Expected hook to be disabled")
	}

	// Enable hook
	err = registry.EnableHook("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !hook.IsEnabled() {
		t.Fatal("Expected hook to be enabled")
	}
}

func TestDefaultHookRegistry_Clear(t *testing.T) {
	registry := NewDefaultHookRegistry()
	hook := NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1)

	err := registry.Register(hook, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	registry.Clear()

	hooks := registry.GetHooks(interfaces.HookEventRequestStart)
	if len(hooks) != 0 {
		t.Fatalf("Expected 0 hooks after clear, got %d", len(hooks))
	}

	listedHooks := registry.ListHooks()
	if len(listedHooks) != 0 {
		t.Fatalf("Expected 0 hooks in list after clear, got %d", len(listedHooks))
	}
}

func TestDefaultHookRegistry_Shutdown(t *testing.T) {
	registry := NewDefaultHookRegistry()
	hook := NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1)

	err := registry.Register(hook, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	err = registry.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// After shutdown, hooks should be cleared
	hooks := registry.GetHooks(interfaces.HookEventRequestStart)
	if len(hooks) != 0 {
		t.Fatalf("Expected 0 hooks after shutdown, got %d", len(hooks))
	}
}

func TestHookMetricsCollector(t *testing.T) {
	collector := NewHookMetricsCollector()

	// Record some executions
	collector.RecordExecution("hook1", interfaces.HookEventRequestStart, 10*time.Millisecond, nil)
	collector.RecordExecution("hook1", interfaces.HookEventRequestStart, 20*time.Millisecond, nil)
	collector.RecordExecution("hook2", interfaces.HookEventRequestEnd, 15*time.Millisecond, errors.New("test error"))

	metrics := collector.GetMetrics()

	if metrics.TotalExecutions != 3 {
		t.Fatalf("Expected 3 total executions, got %d", metrics.TotalExecutions)
	}

	if metrics.SuccessfulExecutions != 2 {
		t.Fatalf("Expected 2 successful executions, got %d", metrics.SuccessfulExecutions)
	}

	if metrics.FailedExecutions != 1 {
		t.Fatalf("Expected 1 failed execution, got %d", metrics.FailedExecutions)
	}

	if metrics.AverageLatency != 15*time.Millisecond {
		t.Fatalf("Expected average latency 15ms, got %v", metrics.AverageLatency)
	}

	if metrics.ExecutionsByEvent[interfaces.HookEventRequestStart] != 2 {
		t.Fatalf("Expected 2 executions for RequestStart event, got %d",
			metrics.ExecutionsByEvent[interfaces.HookEventRequestStart])
	}

	if metrics.ExecutionsByHook["hook1"] != 2 {
		t.Fatalf("Expected 2 executions for hook1, got %d", metrics.ExecutionsByHook["hook1"])
	}

	if metrics.ErrorsByHook["hook2"] != 1 {
		t.Fatalf("Expected 1 error for hook2, got %d", metrics.ErrorsByHook["hook2"])
	}
}

func TestDefaultHookRegistry_HookPriority(t *testing.T) {
	registry := NewDefaultHookRegistry()

	executed := []string{}

	// Create hooks with different priorities
	hook1 := &testHook{
		BaseHook: NewBaseHook("high", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			executed = append(executed, "high")
			return nil
		},
	}

	hook2 := &testHook{
		BaseHook: NewBaseHook("low", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 10),
		execute: func(ctx *interfaces.HookContext) error {
			executed = append(executed, "low")
			return nil
		},
	}

	hook3 := &testHook{
		BaseHook: NewBaseHook("medium", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 5),
		execute: func(ctx *interfaces.HookContext) error {
			executed = append(executed, "medium")
			return nil
		},
	}

	// Register in random order
	err := registry.Register(hook2, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = registry.Register(hook1, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = registry.Register(hook3, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err = registry.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check execution order (high priority first)
	expectedOrder := []string{"high", "medium", "low"}
	if len(executed) != len(expectedOrder) {
		t.Fatalf("Expected %d executions, got %d", len(expectedOrder), len(executed))
	}

	for i, expected := range expectedOrder {
		if executed[i] != expected {
			t.Fatalf("Expected execution order %v, got %v", expectedOrder, executed)
		}
	}
}

// testHook is a helper hook for testing
type testHook struct {
	*BaseHook
	execute func(ctx *interfaces.HookContext) error
}

func (h *testHook) Execute(ctx *interfaces.HookContext) error {
	if h.execute != nil {
		return h.execute(ctx)
	}
	return nil
}

// Mock HTTP request for testing
func createTestRequest() *http.Request {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("Origin", "http://localhost:3000")
	return req
}

func TestDefaultHookRegistry_ExecuteAsync(t *testing.T) {
	registry := NewDefaultHookRegistry()

	executed := false
	hook := &testHook{
		BaseHook: NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			time.Sleep(10 * time.Millisecond)
			executed = true
			return nil
		},
	}

	err := registry.Register(hook, interfaces.HookEventRequestStart)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	errChan := registry.ExecuteAsync(ctx)

	// Should return immediately
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	case <-time.After(50 * time.Millisecond):
		// Give some time for execution
	}

	// Hook should have been executed
	if !executed {
		t.Fatal("Expected hook to be executed asynchronously")
	}
}

func TestDefaultHookRegistry_ExecuteWithRetry(t *testing.T) {
	registry := NewDefaultHookRegistry()

	attempts := 0
	hook := &testHook{
		BaseHook: NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary failure")
			}
			return nil
		},
	}

	hooks := []interfaces.Hook{hook}

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err := registry.ExecuteWithRetry(hooks, ctx, 3)
	if err != nil {
		t.Fatalf("Expected no error after retries, got %v", err)
	}

	if attempts != 3 {
		t.Fatalf("Expected 3 attempts, got %d", attempts)
	}
}

func TestDefaultHookRegistry_ExecuteWithRetryExhausted(t *testing.T) {
	registry := NewDefaultHookRegistry()

	hook := &testHook{
		BaseHook: NewBaseHook("test", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			return errors.New("persistent failure")
		},
	}

	hooks := []interfaces.Hook{hook}

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test-server",
		Timestamp:  time.Now(),
	}

	err := registry.ExecuteWithRetry(hooks, ctx, 2)
	if err == nil {
		t.Fatal("Expected error after retry exhaustion")
	}

	if !strings.Contains(err.Error(), "hook execution failed after 2 retries") {
		t.Fatalf("Expected retry exhaustion error, got %v", err)
	}
}
