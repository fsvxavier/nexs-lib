// Package hooks provides concrete hook implementations for HTTP server lifecycle events.
// It contains specific hook examples that can be used with any HTTP server provider.
package hooks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// BaseHook provides a base implementation for common hook functionality.
// Other hooks can embed this to inherit basic observer behavior.
type BaseHook struct {
	name    string
	enabled bool
	logger  Logger
}

// Logger defines the interface for logging in hooks.
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// DefaultLogger provides a simple default logger implementation.
type DefaultLogger struct{}

// Info logs an info message.
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

// Error logs an error message.
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

// Debug logs a debug message.
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	log.Printf("[DEBUG] "+msg, args...)
}

// Warn logs a warning message.
func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

// NewBaseHook creates a new base hook with the specified name.
func NewBaseHook(name string) *BaseHook {
	return &BaseHook{
		name:    name,
		enabled: true,
		logger:  &DefaultLogger{},
	}
}

// GetName returns the hook name.
func (h *BaseHook) GetName() string {
	return h.name
}

// SetEnabled sets the enabled state of the hook.
func (h *BaseHook) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// IsEnabled returns true if the hook is enabled.
func (h *BaseHook) IsEnabled() bool {
	return h.enabled
}

// SetLogger sets a custom logger for the hook.
func (h *BaseHook) SetLogger(logger Logger) {
	if logger != nil {
		h.logger = logger
	}
}

// OnStart implements the ServerObserver interface.
func (h *BaseHook) OnStart(ctx context.Context, addr string) error {
	if !h.enabled {
		return nil
	}
	h.logger.Info("Hook %s: Server started on %s", h.name, addr)
	return nil
}

// OnStop implements the ServerObserver interface.
func (h *BaseHook) OnStop(ctx context.Context) error {
	if !h.enabled {
		return nil
	}
	h.logger.Info("Hook %s: Server stopped", h.name)
	return nil
}

// OnError implements the ServerObserver interface.
func (h *BaseHook) OnError(ctx context.Context, err error) error {
	if !h.enabled {
		return nil
	}
	h.logger.Error("Hook %s: Server error: %v", h.name, err)
	return nil
}

// OnRequest implements the ServerObserver interface.
func (h *BaseHook) OnRequest(ctx context.Context, req interface{}) error {
	if !h.enabled {
		return nil
	}
	h.logger.Debug("Hook %s: Request received", h.name)
	return nil
}

// OnResponse implements the ServerObserver interface.
func (h *BaseHook) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	if !h.enabled {
		return nil
	}
	h.logger.Debug("Hook %s: Response sent in %v", h.name, duration)
	return nil
}

// OnRouteEnter implements the ServerObserver interface.
func (h *BaseHook) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	if !h.enabled {
		return nil
	}
	h.logger.Debug("Hook %s: Entering route %s %s", h.name, method, path)
	return nil
}

// OnRouteExit implements the ServerObserver interface.
func (h *BaseHook) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	if !h.enabled {
		return nil
	}
	h.logger.Debug("Hook %s: Exiting route %s %s (duration: %v)", h.name, method, path, duration)
	return nil
}

// HookManager manages the lifecycle and execution of hooks.
type HookManager struct {
	observerManager *ObserverManager
	hooks           map[string]interfaces.ServerObserver
	logger          Logger
}

// NewHookManager creates a new hook manager.
func NewHookManager() *HookManager {
	return &HookManager{
		observerManager: NewObserverManager(),
		hooks:           make(map[string]interfaces.ServerObserver),
		logger:          &DefaultLogger{},
	}
}

// RegisterHook registers a hook with the manager.
func (hm *HookManager) RegisterHook(name string, hook interfaces.ServerObserver) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if _, exists := hm.hooks[name]; exists {
		return fmt.Errorf("hook with name %s already exists", name)
	}

	hm.hooks[name] = hook
	return hm.observerManager.AttachObserver(hook)
}

// UnregisterHook removes a hook from the manager.
func (hm *HookManager) UnregisterHook(name string) error {
	hook, exists := hm.hooks[name]
	if !exists {
		return fmt.Errorf("hook with name %s not found", name)
	}

	delete(hm.hooks, name)
	return hm.observerManager.DetachObserver(hook)
}

// GetHook returns a hook by name.
func (hm *HookManager) GetHook(name string) (interfaces.ServerObserver, error) {
	hook, exists := hm.hooks[name]
	if !exists {
		return nil, fmt.Errorf("hook with name %s not found", name)
	}
	return hook, nil
}

// ListHooks returns a list of all registered hook names.
func (hm *HookManager) ListHooks() []string {
	names := make([]string, 0, len(hm.hooks))
	for name := range hm.hooks {
		names = append(names, name)
	}
	return names
}

// NotifyHooks notifies all hooks about an event.
func (hm *HookManager) NotifyHooks(eventType interfaces.EventType, ctx context.Context, data interface{}) error {
	return hm.observerManager.NotifyObservers(eventType, ctx, data)
}

// AttachHookFunc attaches a hook function for a specific event type.
func (hm *HookManager) AttachHookFunc(eventType interfaces.EventType, hook interfaces.HookFunc) error {
	return hm.observerManager.AttachHook(eventType, hook)
}

// DetachHookFunc removes all hook functions for a specific event type.
func (hm *HookManager) DetachHookFunc(eventType interfaces.EventType) error {
	return hm.observerManager.DetachHook(eventType)
}

// Clear removes all hooks and observers.
func (hm *HookManager) Clear() {
	hm.hooks = make(map[string]interfaces.ServerObserver)
	hm.observerManager.Clear()
}

// SetLogger sets a custom logger for the hook manager.
func (hm *HookManager) SetLogger(logger Logger) {
	if logger != nil {
		hm.logger = logger
	}
}
