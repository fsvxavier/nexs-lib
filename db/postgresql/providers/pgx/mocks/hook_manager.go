package mocks

import (
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// MockHookManager implements interfaces.HookManager for testing
type MockHookManager struct {
	RegisterHookFunc         func(hookType interfaces.HookType, hook interfaces.Hook) error
	RegisterCustomHookFunc   func(hookType interfaces.HookType, name string, hook interfaces.Hook) error
	ExecuteHooksFunc         func(hookType interfaces.HookType, ctx *interfaces.ExecutionContext) error
	UnregisterHookFunc       func(hookType interfaces.HookType) error
	UnregisterCustomHookFunc func(hookType interfaces.HookType, name string) error
	ListHooksFunc            func() map[interfaces.HookType][]interfaces.Hook
}

// NewMockHookManager creates a new mock hook manager
func NewMockHookManager() *MockHookManager {
	return &MockHookManager{}
}

func (m *MockHookManager) RegisterHook(hookType interfaces.HookType, hook interfaces.Hook) error {
	if m.RegisterHookFunc != nil {
		return m.RegisterHookFunc(hookType, hook)
	}
	return nil
}

func (m *MockHookManager) RegisterCustomHook(hookType interfaces.HookType, name string, hook interfaces.Hook) error {
	if m.RegisterCustomHookFunc != nil {
		return m.RegisterCustomHookFunc(hookType, name, hook)
	}
	return nil
}

func (m *MockHookManager) ExecuteHooks(hookType interfaces.HookType, ctx *interfaces.ExecutionContext) error {
	if m.ExecuteHooksFunc != nil {
		return m.ExecuteHooksFunc(hookType, ctx)
	}
	return nil
}

func (m *MockHookManager) UnregisterHook(hookType interfaces.HookType) error {
	if m.UnregisterHookFunc != nil {
		return m.UnregisterHookFunc(hookType)
	}
	return nil
}

func (m *MockHookManager) UnregisterCustomHook(hookType interfaces.HookType, name string) error {
	if m.UnregisterCustomHookFunc != nil {
		return m.UnregisterCustomHookFunc(hookType, name)
	}
	return nil
}

func (m *MockHookManager) ListHooks() map[interfaces.HookType][]interfaces.Hook {
	if m.ListHooksFunc != nil {
		return m.ListHooksFunc()
	}
	return make(map[interfaces.HookType][]interfaces.Hook)
}
