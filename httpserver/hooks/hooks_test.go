package hooks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// TestLogger is a test logger that captures log messages for testing.
type TestLogger struct {
	InfoMessages  []string
	ErrorMessages []string
	DebugMessages []string
	WarnMessages  []string
}

func (l *TestLogger) Info(msg string, args ...interface{}) {
	l.InfoMessages = append(l.InfoMessages, msg)
}

func (l *TestLogger) Error(msg string, args ...interface{}) {
	l.ErrorMessages = append(l.ErrorMessages, msg)
}

func (l *TestLogger) Debug(msg string, args ...interface{}) {
	l.DebugMessages = append(l.DebugMessages, msg)
}

func (l *TestLogger) Warn(msg string, args ...interface{}) {
	l.WarnMessages = append(l.WarnMessages, msg)
}

func TestNewBaseHook(t *testing.T) {
	hook := NewBaseHook("test-hook")

	if hook.GetName() != "test-hook" {
		t.Errorf("Expected hook name to be 'test-hook', got %s", hook.GetName())
	}

	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled by default")
	}
}

func TestBaseHook_SetEnabled(t *testing.T) {
	hook := NewBaseHook("test-hook")

	hook.SetEnabled(false)
	if hook.IsEnabled() {
		t.Error("Expected hook to be disabled")
	}

	hook.SetEnabled(true)
	if !hook.IsEnabled() {
		t.Error("Expected hook to be enabled")
	}
}

func TestBaseHook_SetLogger(t *testing.T) {
	hook := NewBaseHook("test-hook")
	testLogger := &TestLogger{}

	hook.SetLogger(testLogger)

	// Test that nil logger is ignored
	hook.SetLogger(nil)
	if hook.logger == nil {
		t.Error("Expected logger to not be nil after setting nil logger")
	}
}

func TestBaseHook_OnStart(t *testing.T) {
	hook := NewBaseHook("test-hook")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	addr := "localhost:8080"

	err := hook.OnStart(ctx, addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}

	// Test disabled hook
	hook.SetEnabled(false)
	testLogger.InfoMessages = nil
	err = hook.OnStart(ctx, addr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 0 {
		t.Errorf("Expected 0 info messages for disabled hook, got %d", len(testLogger.InfoMessages))
	}
}

func TestBaseHook_OnStop(t *testing.T) {
	hook := NewBaseHook("test-hook")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()

	err := hook.OnStop(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}
}

func TestBaseHook_OnError(t *testing.T) {
	hook := NewBaseHook("test-hook")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	testErr := errors.New("test error")

	err := hook.OnError(ctx, testErr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.ErrorMessages) != 1 {
		t.Errorf("Expected 1 error message, got %d", len(testLogger.ErrorMessages))
	}
}

func TestBaseHook_OnRequest(t *testing.T) {
	hook := NewBaseHook("test-hook")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	req := "test request"

	err := hook.OnRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}
}

func TestBaseHook_OnResponse(t *testing.T) {
	hook := NewBaseHook("test-hook")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	req := "test request"
	resp := "test response"
	duration := time.Millisecond * 100

	err := hook.OnResponse(ctx, req, resp, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}
}

func TestBaseHook_OnRouteEnter(t *testing.T) {
	hook := NewBaseHook("test-hook")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	method := "GET"
	path := "/test"
	req := "test request"

	err := hook.OnRouteEnter(ctx, method, path, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}
}

func TestBaseHook_OnRouteExit(t *testing.T) {
	hook := NewBaseHook("test-hook")
	testLogger := &TestLogger{}
	hook.SetLogger(testLogger)

	ctx := context.Background()
	method := "GET"
	path := "/test"
	req := "test request"
	duration := time.Millisecond * 50

	err := hook.OnRouteExit(ctx, method, path, req, duration)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}
}

func TestNewHookManager(t *testing.T) {
	manager := NewHookManager()

	if manager == nil {
		t.Fatal("Expected hook manager to be created")
	}

	if manager.observerManager == nil {
		t.Error("Expected observer manager to be initialized")
	}

	if manager.hooks == nil {
		t.Error("Expected hooks map to be initialized")
	}
}

func TestHookManager_RegisterHook(t *testing.T) {
	manager := NewHookManager()
	hook := NewBaseHook("test-hook")

	// Test successful registration
	err := manager.RegisterHook("test-hook", hook)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate registration
	err = manager.RegisterHook("test-hook", hook)
	if err == nil {
		t.Error("Expected error for duplicate hook registration")
	}

	// Test nil hook registration
	err = manager.RegisterHook("nil-hook", nil)
	if err == nil {
		t.Error("Expected error for nil hook registration")
	}
}

func TestHookManager_UnregisterHook(t *testing.T) {
	manager := NewHookManager()
	hook := NewBaseHook("test-hook")

	// Test unregistering non-existent hook
	err := manager.UnregisterHook("non-existent")
	if err == nil {
		t.Error("Expected error for unregistering non-existent hook")
	}

	// Register and then unregister hook
	err = manager.RegisterHook("test-hook", hook)
	if err != nil {
		t.Errorf("Expected no error during registration, got %v", err)
	}

	err = manager.UnregisterHook("test-hook")
	if err != nil {
		t.Errorf("Expected no error during unregistration, got %v", err)
	}

	// Verify hook is removed
	_, err = manager.GetHook("test-hook")
	if err == nil {
		t.Error("Expected error when getting unregistered hook")
	}
}

func TestHookManager_GetHook(t *testing.T) {
	manager := NewHookManager()
	hook := NewBaseHook("test-hook")

	// Test getting non-existent hook
	_, err := manager.GetHook("non-existent")
	if err == nil {
		t.Error("Expected error for getting non-existent hook")
	}

	// Register and get hook
	err = manager.RegisterHook("test-hook", hook)
	if err != nil {
		t.Errorf("Expected no error during registration, got %v", err)
	}

	retrievedHook, err := manager.GetHook("test-hook")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrievedHook != hook {
		t.Error("Expected to get the same hook instance")
	}
}

func TestHookManager_ListHooks(t *testing.T) {
	manager := NewHookManager()

	// Test empty list
	hooks := manager.ListHooks()
	if len(hooks) != 0 {
		t.Errorf("Expected empty hook list, got %d hooks", len(hooks))
	}

	// Register hooks and test list
	hook1 := NewBaseHook("hook1")
	hook2 := NewBaseHook("hook2")

	manager.RegisterHook("hook1", hook1)
	manager.RegisterHook("hook2", hook2)

	hooks = manager.ListHooks()
	if len(hooks) != 2 {
		t.Errorf("Expected 2 hooks, got %d", len(hooks))
	}

	// Check if both hooks are in the list
	hookMap := make(map[string]bool)
	for _, name := range hooks {
		hookMap[name] = true
	}

	if !hookMap["hook1"] || !hookMap["hook2"] {
		t.Error("Expected both 'hook1' and 'hook2' to be in the list")
	}
}

func TestHookManager_AttachHookFunc(t *testing.T) {
	manager := NewHookManager()

	hookFunc := func(ctx context.Context, data interface{}) error {
		return nil
	}

	err := manager.AttachHookFunc(interfaces.EventStart, hookFunc)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test nil hook function
	err = manager.AttachHookFunc(interfaces.EventStart, nil)
	if err == nil {
		t.Error("Expected error for nil hook function")
	}
}

func TestHookManager_DetachHookFunc(t *testing.T) {
	manager := NewHookManager()

	hookFunc := func(ctx context.Context, data interface{}) error {
		return nil
	}

	// Attach and then detach
	err := manager.AttachHookFunc(interfaces.EventStart, hookFunc)
	if err != nil {
		t.Errorf("Expected no error during attachment, got %v", err)
	}

	err = manager.DetachHookFunc(interfaces.EventStart)
	if err != nil {
		t.Errorf("Expected no error during detachment, got %v", err)
	}
}

func TestHookManager_Clear(t *testing.T) {
	manager := NewHookManager()
	hook := NewBaseHook("test-hook")

	// Register hook and attach hook function
	manager.RegisterHook("test-hook", hook)
	manager.AttachHookFunc(interfaces.EventStart, func(ctx context.Context, data interface{}) error {
		return nil
	})

	// Clear all
	manager.Clear()

	// Verify everything is cleared
	hooks := manager.ListHooks()
	if len(hooks) != 0 {
		t.Errorf("Expected empty hook list after clear, got %d hooks", len(hooks))
	}
}

func TestHookManager_NotifyHooks(t *testing.T) {
	manager := NewHookManager()
	testLogger := &TestLogger{}
	hook := NewBaseHook("test-hook")
	hook.SetLogger(testLogger)

	err := manager.RegisterHook("test-hook", hook)
	if err != nil {
		t.Errorf("Expected no error during registration, got %v", err)
	}

	ctx := context.Background()
	addr := "localhost:8080"

	err = manager.NotifyHooks(interfaces.EventStart, ctx, addr)
	if err != nil {
		t.Errorf("Expected no error during notification, got %v", err)
	}

	if len(testLogger.InfoMessages) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(testLogger.InfoMessages))
	}
}
