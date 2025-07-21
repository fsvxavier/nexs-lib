package hooks

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestCustomHook(t *testing.T) {
	executed := false
	hook := NewCustomHook(
		"test-hook",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		func(ctx *interfaces.HookContext) error {
			executed = true
			return nil
		},
	)

	// Test basic properties
	if hook.Name() != "test-hook" {
		t.Errorf("Expected name 'test-hook', got '%s'", hook.Name())
	}

	if len(hook.Events()) != 1 || hook.Events()[0] != interfaces.HookEventRequestStart {
		t.Errorf("Expected events [HookEventRequestStart], got %v", hook.Events())
	}

	if hook.Priority() != 100 {
		t.Errorf("Expected priority 100, got %d", hook.Priority())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}

	// Test execution
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &interfaces.HookContext{
		Event:   interfaces.HookEventRequestStart,
		Request: req,
	}

	err := hook.Execute(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !executed {
		t.Error("Expected hook to be executed")
	}
}

func TestCustomHookWithCondition(t *testing.T) {
	hook := NewCustomHook(
		"conditional-hook",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		func(ctx *interfaces.HookContext) error {
			return nil
		},
	)

	// Set condition
	hook.SetCondition(func(ctx *interfaces.HookContext) bool {
		return ctx.Request.URL.Path == "/allowed"
	})

	// Test with allowed path
	req := httptest.NewRequest("GET", "/allowed", nil)
	ctx := &interfaces.HookContext{
		Event:   interfaces.HookEventRequestStart,
		Request: req,
	}

	if !hook.ShouldExecute(ctx) {
		t.Error("Expected hook to execute for allowed path")
	}

	// Test with disallowed path
	req = httptest.NewRequest("GET", "/denied", nil)
	ctx.Request = req

	if hook.ShouldExecute(ctx) {
		t.Error("Expected hook not to execute for denied path")
	}
}

func TestCustomHookWithFilters(t *testing.T) {
	hook := NewCustomHook(
		"filtered-hook",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		func(ctx *interfaces.HookContext) error {
			return nil
		},
	)

	// Set filters
	hook.SetPathFilter(func(path string) bool {
		return path == "/api/test"
	})

	hook.SetMethodFilter(func(method string) bool {
		return method == "POST"
	})

	// Test with matching path and method
	req := httptest.NewRequest("POST", "/api/test", nil)
	ctx := &interfaces.HookContext{
		Event:   interfaces.HookEventRequestStart,
		Request: req,
	}

	if !hook.ShouldExecute(ctx) {
		t.Error("Expected hook to execute for matching path and method")
	}

	// Test with non-matching path
	req = httptest.NewRequest("POST", "/api/other", nil)
	ctx.Request = req

	if hook.ShouldExecute(ctx) {
		t.Error("Expected hook not to execute for non-matching path")
	}

	// Test with non-matching method
	req = httptest.NewRequest("GET", "/api/test", nil)
	ctx.Request = req

	if hook.ShouldExecute(ctx) {
		t.Error("Expected hook not to execute for non-matching method")
	}
}

func TestCustomHookAsyncExecution(t *testing.T) {
	hook := NewCustomHook(
		"async-hook",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		func(ctx *interfaces.HookContext) error {
			time.Sleep(10 * time.Millisecond) // Simulate some work
			return nil
		},
	)

	hook.SetAsyncExecution(5, 1*time.Second)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &interfaces.HookContext{
		Event:   interfaces.HookEventRequestStart,
		Request: req,
	}

	// Test async execution
	errChan := hook.ExecuteAsync(ctx)

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Expected no error from async execution, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Async execution timed out")
	}

	// Test buffer size and timeout
	if hook.BufferSize() != 5 {
		t.Errorf("Expected buffer size 5, got %d", hook.BufferSize())
	}

	if hook.Timeout() != 1*time.Second {
		t.Errorf("Expected timeout 1s, got %v", hook.Timeout())
	}
}

func TestCustomHookBuilder(t *testing.T) {
	builder := NewCustomHookBuilder()

	hook, err := builder.
		WithName("builder-hook").
		WithEvents(interfaces.HookEventRequestStart, interfaces.HookEventRequestEnd).
		WithPriority(200).
		WithCondition(func(ctx *interfaces.HookContext) bool {
			return true
		}).
		WithPathFilter(func(path string) bool {
			return path != "/excluded"
		}).
		WithAsyncExecution(10, 5*time.Second).
		WithExecuteFunc(func(ctx *interfaces.HookContext) error {
			return nil
		}).
		Build()

	if err != nil {
		t.Errorf("Expected no error building hook, got %v", err)
	}

	if hook.Name() != "builder-hook" {
		t.Errorf("Expected name 'builder-hook', got '%s'", hook.Name())
	}

	if len(hook.Events()) != 2 {
		t.Errorf("Expected 2 events, got %d", len(hook.Events()))
	}

	if hook.Priority() != 200 {
		t.Errorf("Expected priority 200, got %d", hook.Priority())
	}
}

func TestCustomHookBuilderValidation(t *testing.T) {
	// Test missing name
	builder1 := NewCustomHookBuilder()
	_, err := builder1.
		WithEvents(interfaces.HookEventRequestStart).
		WithExecuteFunc(func(ctx *interfaces.HookContext) error { return nil }).
		Build()

	if err == nil {
		t.Error("Expected error for missing name")
	}

	// Test missing events
	builder2 := NewCustomHookBuilder()
	_, err = builder2.
		WithName("test").
		WithExecuteFunc(func(ctx *interfaces.HookContext) error { return nil }).
		Build()

	if err == nil {
		t.Error("Expected error for missing events")
	}

	// Test missing execute function
	builder3 := NewCustomHookBuilder()
	_, err = builder3.
		WithName("test").
		WithEvents(interfaces.HookEventRequestStart).
		Build()

	if err == nil {
		t.Error("Expected error for missing execute function")
	}
}

func TestCustomHookFactory(t *testing.T) {
	factory := NewCustomHookFactory()

	// Test simple hook creation
	hook := factory.NewSimpleHook(
		"simple",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		func(ctx *interfaces.HookContext) error { return nil },
	)

	if hook.Name() != "simple" {
		t.Errorf("Expected name 'simple', got '%s'", hook.Name())
	}

	// Test conditional hook creation
	conditionalHook := factory.NewConditionalHook(
		"conditional",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		func(ctx *interfaces.HookContext) bool { return true },
		func(ctx *interfaces.HookContext) error { return nil },
	)

	if conditionalHook.Name() != "conditional" {
		t.Errorf("Expected name 'conditional', got '%s'", conditionalHook.Name())
	}

	// Test async hook creation
	asyncHook := factory.NewAsyncHook(
		"async",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		5,
		1*time.Second,
		func(ctx *interfaces.HookContext) error { return nil },
	)

	if asyncHook.Name() != "async" {
		t.Errorf("Expected name 'async', got '%s'", asyncHook.Name())
	}

	// Test filtered hook creation
	filteredHook := factory.NewFilteredHook(
		"filtered",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		func(path string) bool { return true },
		func(method string) bool { return true },
		func(ctx *interfaces.HookContext) error { return nil },
	)

	if filteredHook.Name() != "filtered" {
		t.Errorf("Expected name 'filtered', got '%s'", filteredHook.Name())
	}
}

func TestCustomHookDisabling(t *testing.T) {
	hook := NewCustomHook(
		"test-hook",
		[]interfaces.HookEvent{interfaces.HookEventRequestStart},
		100,
		func(ctx *interfaces.HookContext) error {
			return nil
		},
	)

	// Disable the hook
	hook.SetEnabled(false)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &interfaces.HookContext{
		Event:   interfaces.HookEventRequestStart,
		Request: req,
	}

	// Hook should not execute when disabled
	if hook.ShouldExecute(ctx) {
		t.Error("Expected disabled hook not to execute")
	}

	// Re-enable the hook
	hook.SetEnabled(true)

	if !hook.ShouldExecute(ctx) {
		t.Error("Expected enabled hook to execute")
	}
}
