package hooks

import (
	"context"
	"fmt"
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestNewDefaultHookManager(t *testing.T) {
	hm := NewDefaultHookManager()

	if hm == nil {
		t.Error("NewDefaultHookManager() returned nil")
		return
	}

	if hm.hooks == nil {
		t.Error("hooks map not initialized")
	}

	if hm.customHooks == nil {
		t.Error("customHooks map not initialized")
	}

	// Verify empty state
	if len(hm.ListHooks()) != 0 {
		t.Error("New hook manager should have no hooks")
	}
}

func TestHookManager_RegisterHook(t *testing.T) {
	hm := NewDefaultHookManager()

	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: true}
	}

	tests := []struct {
		name      string
		hookType  interfaces.HookType
		hook      interfaces.Hook
		wantError bool
	}{
		{
			name:      "valid hook",
			hookType:  interfaces.BeforeQueryHook,
			hook:      testHook,
			wantError: false,
		},
		{
			name:      "nil hook",
			hookType:  interfaces.BeforeQueryHook,
			hook:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hm.RegisterHook(tt.hookType, tt.hook)

			if tt.wantError {
				if err == nil {
					t.Error("RegisterHook() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("RegisterHook() unexpected error = %v", err)
				}

				// Verify hook was registered
				if !hm.HasHooks(tt.hookType) {
					t.Error("Hook was not registered")
				}
			}
		})
	}
}

func TestHookManager_RegisterCustomHook(t *testing.T) {
	hm := NewDefaultHookManager()

	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: true}
	}

	tests := []struct {
		name      string
		hookType  interfaces.HookType
		hookName  string
		hook      interfaces.Hook
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid custom hook",
			hookType:  interfaces.CustomHookBase + 1,
			hookName:  "test_hook",
			hook:      testHook,
			wantError: false,
		},
		{
			name:      "nil hook",
			hookType:  interfaces.CustomHookBase + 1,
			hookName:  "test_hook",
			hook:      nil,
			wantError: true,
			errorMsg:  "hook cannot be nil",
		},
		{
			name:      "empty name",
			hookType:  interfaces.CustomHookBase + 1,
			hookName:  "",
			hook:      testHook,
			wantError: true,
			errorMsg:  "custom hook name cannot be empty",
		},
		{
			name:      "invalid hook type",
			hookType:  interfaces.BeforeQueryHook,
			hookName:  "test_hook",
			hook:      testHook,
			wantError: true,
			errorMsg:  "custom hook type must be >= CustomHookBase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hm.RegisterCustomHook(tt.hookType, tt.hookName, tt.hook)

			if tt.wantError {
				if err == nil {
					t.Error("RegisterCustomHook() expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("RegisterCustomHook() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("RegisterCustomHook() unexpected error = %v", err)
				}

				// Verify hook was registered
				if !hm.HasHooks(tt.hookType) {
					t.Error("Custom hook was not registered")
				}
			}
		})
	}
}

func TestHookManager_ExecuteHooks(t *testing.T) {
	hm := NewDefaultHookManager()

	// Test hook that succeeds
	successHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		ctx.Metadata = map[string]interface{}{"success": true}
		return &interfaces.HookResult{Continue: true}
	}

	// Test hook that fails
	failHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{
			Continue: false,
			Error:    fmt.Errorf("hook failed"),
		}
	}

	// Test hook that stops execution
	stopHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: false}
	}

	tests := []struct {
		name      string
		setupFunc func()
		hookType  interfaces.HookType
		wantError bool
		errorMsg  string
	}{
		{
			name: "successful hook execution",
			setupFunc: func() {
				_ = hm.RegisterHook(interfaces.BeforeQueryHook, successHook)
			},
			hookType:  interfaces.BeforeQueryHook,
			wantError: false,
		},
		{
			name: "hook with error",
			setupFunc: func() {
				hm.ClearHooks()
				_ = hm.RegisterHook(interfaces.BeforeQueryHook, failHook)
			},
			hookType:  interfaces.BeforeQueryHook,
			wantError: true,
			errorMsg:  "hook execution failed: hook failed",
		},
		{
			name: "hook stops execution",
			setupFunc: func() {
				hm.ClearHooks()
				_ = hm.RegisterHook(interfaces.BeforeQueryHook, stopHook)
			},
			hookType:  interfaces.BeforeQueryHook,
			wantError: true,
			errorMsg:  "hook requested execution stop",
		},
		{
			name: "no hooks registered",
			setupFunc: func() {
				hm.ClearHooks()
			},
			hookType:  interfaces.BeforeQueryHook,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			ctx := &interfaces.ExecutionContext{
				Context:   context.Background(),
				Operation: "test",
				Metadata:  make(map[string]interface{}),
			}

			err := hm.ExecuteHooks(tt.hookType, ctx)

			if tt.wantError {
				if err == nil {
					t.Error("ExecuteHooks() expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ExecuteHooks() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ExecuteHooks() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestHookManager_ExecuteCustomHooks(t *testing.T) {
	hm := NewDefaultHookManager()

	customHookType := interfaces.CustomHookBase + 1

	customHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		ctx.Metadata["custom"] = true
		return &interfaces.HookResult{Continue: true}
	}

	failCustomHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{
			Continue: false,
			Error:    fmt.Errorf("custom hook failed"),
		}
	}

	err := hm.RegisterCustomHook(customHookType, "success_hook", customHook)
	if err != nil {
		t.Errorf("RegisterCustomHook() error = %v", err)
		return
	}

	err = hm.RegisterCustomHook(customHookType, "fail_hook", failCustomHook)
	if err != nil {
		t.Errorf("RegisterCustomHook() error = %v", err)
		return
	}

	tests := []struct {
		name      string
		setupFunc func()
		wantError bool
		errorMsg  string
	}{
		{
			name: "successful custom hook",
			setupFunc: func() {
				_ = hm.UnregisterCustomHook(customHookType, "fail_hook")
			},
			wantError: false,
		},
		{
			name: "failing custom hook",
			setupFunc: func() {
				_ = hm.RegisterCustomHook(customHookType, "fail_hook", failCustomHook)
			},
			wantError: true,
			errorMsg:  "custom hook 'fail_hook' execution failed: custom hook failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			ctx := &interfaces.ExecutionContext{
				Context:   context.Background(),
				Operation: "test",
				Metadata:  make(map[string]interface{}),
			}

			err := hm.ExecuteHooks(customHookType, ctx)

			if tt.wantError {
				if err == nil {
					t.Error("ExecuteHooks() expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ExecuteHooks() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ExecuteHooks() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestHookManager_UnregisterHook(t *testing.T) {
	hm := NewDefaultHookManager()

	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: true}
	}

	// Register a hook
	err := hm.RegisterHook(interfaces.BeforeQueryHook, testHook)
	if err != nil {
		t.Errorf("RegisterHook() error = %v", err)
		return
	}

	// Verify hook exists
	if !hm.HasHooks(interfaces.BeforeQueryHook) {
		t.Error("Hook should exist before unregistering")
	}

	// Unregister hook
	err = hm.UnregisterHook(interfaces.BeforeQueryHook)
	if err != nil {
		t.Errorf("UnregisterHook() error = %v", err)
		return
	}

	// Verify hook no longer exists
	if hm.HasHooks(interfaces.BeforeQueryHook) {
		t.Error("Hook should not exist after unregistering")
	}
}

func TestHookManager_UnregisterCustomHook(t *testing.T) {
	hm := NewDefaultHookManager()

	customHookType := interfaces.CustomHookBase + 1
	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: true}
	}

	tests := []struct {
		name      string
		setupFunc func()
		hookName  string
		wantError bool
		errorMsg  string
	}{
		{
			name: "unregister existing custom hook",
			setupFunc: func() {
				_ = hm.RegisterCustomHook(customHookType, "test_hook", testHook)
			},
			hookName:  "test_hook",
			wantError: false,
		},
		{
			name: "unregister non-existing custom hook",
			setupFunc: func() {
				hm.ClearHooks()
			},
			hookName:  "non_existing",
			wantError: false, // Should not error for non-existing hooks
		},
		{
			name:      "empty hook name",
			setupFunc: func() {},
			hookName:  "",
			wantError: true,
			errorMsg:  "custom hook name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			err := hm.UnregisterCustomHook(customHookType, tt.hookName)

			if tt.wantError {
				if err == nil {
					t.Error("UnregisterCustomHook() expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("UnregisterCustomHook() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("UnregisterCustomHook() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestHookManager_ListHooks(t *testing.T) {
	hm := NewDefaultHookManager()

	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: true}
	}

	// Register multiple hooks
	_ = hm.RegisterHook(interfaces.BeforeQueryHook, testHook)
	_ = hm.RegisterHook(interfaces.AfterQueryHook, testHook)
	_ = hm.RegisterCustomHook(interfaces.CustomHookBase+1, "custom1", testHook)

	hooks := hm.ListHooks()

	// Verify we have the correct number of hook types
	expectedTypes := 3
	if len(hooks) != expectedTypes {
		t.Errorf("ListHooks() returned %d hook types, want %d", len(hooks), expectedTypes)
	}

	// Verify specific hooks exist
	if _, exists := hooks[interfaces.BeforeQueryHook]; !exists {
		t.Error("BeforeQueryHook should exist in list")
	}

	if _, exists := hooks[interfaces.AfterQueryHook]; !exists {
		t.Error("AfterQueryHook should exist in list")
	}

	if _, exists := hooks[interfaces.CustomHookBase+1]; !exists {
		t.Error("Custom hook should exist in list")
	}
}

func TestHookManager_ConcurrentAccess(t *testing.T) {
	hm := NewDefaultHookManager()

	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: true}
	}

	// Test concurrent registration and execution
	done := make(chan bool, 2)

	// Goroutine 1: Register hooks
	go func() {
		for i := 0; i < 100; i++ {
			_ = hm.RegisterHook(interfaces.BeforeQueryHook, testHook)
		}
		done <- true
	}()

	// Goroutine 2: Execute hooks
	go func() {
		for i := 0; i < 100; i++ {
			ctx := &interfaces.ExecutionContext{
				Context:   context.Background(),
				Operation: "test",
				StartTime: time.Now(),
			}
			_ = hm.ExecuteHooks(interfaces.BeforeQueryHook, ctx)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Should not panic and should have hooks registered
	if !hm.HasHooks(interfaces.BeforeQueryHook) {
		t.Error("Hooks should be registered after concurrent access")
	}
}

func TestHookManager_GetHookCount(t *testing.T) {
	hm := NewDefaultHookManager()

	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: true}
	}

	// Initially should have no hooks
	if hm.GetHookCount(interfaces.BeforeQueryHook) != 0 {
		t.Error("Should have 0 hooks initially")
	}

	// Register hooks
	_ = hm.RegisterHook(interfaces.BeforeQueryHook, testHook)
	_ = hm.RegisterHook(interfaces.BeforeQueryHook, testHook)
	_ = hm.RegisterCustomHook(interfaces.CustomHookBase+1, "custom1", testHook)

	// Check counts
	if hm.GetHookCount(interfaces.BeforeQueryHook) != 2 {
		t.Errorf("Should have 2 BeforeQueryHooks, got %d", hm.GetHookCount(interfaces.BeforeQueryHook))
	}

	if hm.GetHookCount(interfaces.CustomHookBase+1) != 1 {
		t.Errorf("Should have 1 custom hook, got %d", hm.GetHookCount(interfaces.CustomHookBase+1))
	}

	if hm.GetHookCount(interfaces.AfterQueryHook) != 0 {
		t.Errorf("Should have 0 AfterQueryHooks, got %d", hm.GetHookCount(interfaces.AfterQueryHook))
	}
}

func TestHookManager_ClearHooks(t *testing.T) {
	hm := NewDefaultHookManager()

	testHook := func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		return &interfaces.HookResult{Continue: true}
	}

	// Register some hooks
	_ = hm.RegisterHook(interfaces.BeforeQueryHook, testHook)
	_ = hm.RegisterCustomHook(interfaces.CustomHookBase+1, "custom1", testHook)

	// Verify hooks exist
	if !hm.HasHooks(interfaces.BeforeQueryHook) {
		t.Error("Hooks should exist before clearing")
	}

	// Clear all hooks
	hm.ClearHooks()

	// Verify no hooks exist
	if hm.HasHooks(interfaces.BeforeQueryHook) {
		t.Error("No hooks should exist after clearing")
	}

	if hm.HasHooks(interfaces.CustomHookBase + 1) {
		t.Error("No custom hooks should exist after clearing")
	}

	if len(hm.ListHooks()) != 0 {
		t.Error("ListHooks should return empty map after clearing")
	}
}
