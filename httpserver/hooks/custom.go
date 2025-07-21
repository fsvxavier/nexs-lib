// Package hooks provides custom hook implementations.
package hooks

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// CustomHook implements a fully customizable hook.
type CustomHook struct {
	name         string
	events       []interfaces.HookEvent
	priority     int
	enabled      bool
	condition    func(ctx *interfaces.HookContext) bool
	pathFilter   func(path string) bool
	methodFilter func(method string) bool
	headerFilter func(headers http.Header) bool
	executeFunc  func(ctx *interfaces.HookContext) error
	isAsync      bool
	bufferSize   int
	timeout      time.Duration
}

// NewCustomHook creates a new custom hook with the provided configuration.
func NewCustomHook(name string, events []interfaces.HookEvent, priority int, executeFunc func(ctx *interfaces.HookContext) error) *CustomHook {
	return &CustomHook{
		name:        name,
		events:      events,
		priority:    priority,
		enabled:     true,
		executeFunc: executeFunc,
	}
}

// Execute executes the custom hook.
func (h *CustomHook) Execute(ctx *interfaces.HookContext) error {
	if h.executeFunc != nil {
		return h.executeFunc(ctx)
	}
	return nil
}

// Name returns the hook name.
func (h *CustomHook) Name() string {
	return h.name
}

// Events returns the events this hook handles.
func (h *CustomHook) Events() []interfaces.HookEvent {
	return h.events
}

// Priority returns the hook priority.
func (h *CustomHook) Priority() int {
	return h.priority
}

// IsEnabled returns whether the hook is enabled.
func (h *CustomHook) IsEnabled() bool {
	return h.enabled
}

// SetEnabled sets the enabled state of the hook.
func (h *CustomHook) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// ShouldExecute determines if the hook should execute for the given context.
func (h *CustomHook) ShouldExecute(ctx *interfaces.HookContext) bool {
	if !h.IsEnabled() {
		return false
	}

	// Check condition
	if h.condition != nil && !h.condition(ctx) {
		return false
	}

	// Check filters
	if ctx.Request != nil {
		if h.pathFilter != nil && !h.pathFilter(ctx.Request.URL.Path) {
			return false
		}
		if h.methodFilter != nil && !h.methodFilter(ctx.Request.Method) {
			return false
		}
		if h.headerFilter != nil && !h.headerFilter(ctx.Request.Header) {
			return false
		}
	}

	return true
}

// SetCondition sets a condition function for conditional execution.
func (h *CustomHook) SetCondition(condition func(ctx *interfaces.HookContext) bool) {
	h.condition = condition
}

// SetPathFilter sets a path filter for the hook.
func (h *CustomHook) SetPathFilter(filter func(path string) bool) {
	h.pathFilter = filter
}

// SetMethodFilter sets a method filter for the hook.
func (h *CustomHook) SetMethodFilter(filter func(method string) bool) {
	h.methodFilter = filter
}

// SetHeaderFilter sets a header filter for the hook.
func (h *CustomHook) SetHeaderFilter(filter func(headers http.Header) bool) {
	h.headerFilter = filter
}

// SetAsyncExecution enables asynchronous execution.
func (h *CustomHook) SetAsyncExecution(bufferSize int, timeout time.Duration) {
	h.isAsync = true
	h.bufferSize = bufferSize
	h.timeout = timeout
}

// Condition returns the condition function (implements ConditionalHook).
func (h *CustomHook) Condition() func(ctx *interfaces.HookContext) bool {
	return h.condition
}

// PathFilter returns the path filter function (implements FilteredHook).
func (h *CustomHook) PathFilter() func(path string) bool {
	return h.pathFilter
}

// MethodFilter returns the method filter function (implements FilteredHook).
func (h *CustomHook) MethodFilter() func(method string) bool {
	return h.methodFilter
}

// HeaderFilter returns the header filter function (implements FilteredHook).
func (h *CustomHook) HeaderFilter() func(headers http.Header) bool {
	return h.headerFilter
}

// ExecuteAsync executes the hook asynchronously (implements AsyncHook).
func (h *CustomHook) ExecuteAsync(ctx *interfaces.HookContext) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)
		if err := h.Execute(ctx); err != nil {
			errChan <- err
		}
	}()

	return errChan
}

// BufferSize returns the buffer size for async execution (implements AsyncHook).
func (h *CustomHook) BufferSize() int {
	return h.bufferSize
}

// Timeout returns the timeout for async execution (implements AsyncHook).
func (h *CustomHook) Timeout() time.Duration {
	return h.timeout
}

// CustomHookBuilder implements the CustomHookBuilder interface.
type CustomHookBuilder struct {
	hook *CustomHook
}

// NewCustomHookBuilder creates a new custom hook builder.
func NewCustomHookBuilder() *CustomHookBuilder {
	return &CustomHookBuilder{
		hook: &CustomHook{
			enabled:    true,
			bufferSize: 10,
			timeout:    30 * time.Second,
		},
	}
}

// WithName sets the name of the custom hook.
func (b *CustomHookBuilder) WithName(name string) interfaces.CustomHookBuilder {
	b.hook.name = name
	return b
}

// WithEvents sets the events this hook should handle.
func (b *CustomHookBuilder) WithEvents(events ...interfaces.HookEvent) interfaces.CustomHookBuilder {
	b.hook.events = events
	return b
}

// WithPriority sets the priority of the hook.
func (b *CustomHookBuilder) WithPriority(priority int) interfaces.CustomHookBuilder {
	b.hook.priority = priority
	return b
}

// WithCondition sets a condition function for conditional execution.
func (b *CustomHookBuilder) WithCondition(condition func(ctx *interfaces.HookContext) bool) interfaces.CustomHookBuilder {
	b.hook.condition = condition
	return b
}

// WithPathFilter sets a path filter for the hook.
func (b *CustomHookBuilder) WithPathFilter(filter func(path string) bool) interfaces.CustomHookBuilder {
	b.hook.pathFilter = filter
	return b
}

// WithMethodFilter sets a method filter for the hook.
func (b *CustomHookBuilder) WithMethodFilter(filter func(method string) bool) interfaces.CustomHookBuilder {
	b.hook.methodFilter = filter
	return b
}

// WithHeaderFilter sets a header filter for the hook.
func (b *CustomHookBuilder) WithHeaderFilter(filter func(headers http.Header) bool) interfaces.CustomHookBuilder {
	b.hook.headerFilter = filter
	return b
}

// WithAsyncExecution enables asynchronous execution.
func (b *CustomHookBuilder) WithAsyncExecution(bufferSize int, timeout time.Duration) interfaces.CustomHookBuilder {
	b.hook.isAsync = true
	b.hook.bufferSize = bufferSize
	b.hook.timeout = timeout
	return b
}

// WithExecuteFunc sets the main execution function.
func (b *CustomHookBuilder) WithExecuteFunc(fn func(ctx *interfaces.HookContext) error) interfaces.CustomHookBuilder {
	b.hook.executeFunc = fn
	return b
}

// Build creates the custom hook.
func (b *CustomHookBuilder) Build() (interfaces.Hook, error) {
	if b.hook.name == "" {
		return nil, fmt.Errorf("hook name is required")
	}
	if len(b.hook.events) == 0 {
		return nil, fmt.Errorf("hook events are required")
	}
	if b.hook.executeFunc == nil {
		return nil, fmt.Errorf("execute function is required")
	}

	return b.hook, nil
}

// CustomHookFactory implements the CustomHookFactory interface.
type CustomHookFactory struct{}

// NewCustomHookFactory creates a new custom hook factory.
func NewCustomHookFactory() *CustomHookFactory {
	return &CustomHookFactory{}
}

// NewBuilder creates a new custom hook builder.
func (f *CustomHookFactory) NewBuilder() interfaces.CustomHookBuilder {
	return NewCustomHookBuilder()
}

// NewSimpleHook creates a simple custom hook with basic configuration.
func (f *CustomHookFactory) NewSimpleHook(name string, events []interfaces.HookEvent, priority int, fn func(ctx *interfaces.HookContext) error) interfaces.Hook {
	return NewCustomHook(name, events, priority, fn)
}

// NewConditionalHook creates a conditional custom hook.
func (f *CustomHookFactory) NewConditionalHook(name string, events []interfaces.HookEvent, priority int, condition func(ctx *interfaces.HookContext) bool, fn func(ctx *interfaces.HookContext) error) interfaces.ConditionalHook {
	hook := NewCustomHook(name, events, priority, fn)
	hook.SetCondition(condition)
	return hook
}

// NewAsyncHook creates an asynchronous custom hook.
func (f *CustomHookFactory) NewAsyncHook(name string, events []interfaces.HookEvent, priority int, bufferSize int, timeout time.Duration, fn func(ctx *interfaces.HookContext) error) interfaces.AsyncHook {
	hook := NewCustomHook(name, events, priority, fn)
	hook.SetAsyncExecution(bufferSize, timeout)
	return hook
}

// NewFilteredHook creates a filtered custom hook.
func (f *CustomHookFactory) NewFilteredHook(name string, events []interfaces.HookEvent, priority int, pathFilter func(string) bool, methodFilter func(string) bool, fn func(ctx *interfaces.HookContext) error) interfaces.FilteredHook {
	hook := NewCustomHook(name, events, priority, fn)
	hook.SetPathFilter(pathFilter)
	hook.SetMethodFilter(methodFilter)
	return hook
}

// Verify that CustomHook implements all necessary interfaces
var _ interfaces.Hook = (*CustomHook)(nil)
var _ interfaces.ConditionalHook = (*CustomHook)(nil)
var _ interfaces.FilteredHook = (*CustomHook)(nil)
var _ interfaces.AsyncHook = (*CustomHook)(nil)
var _ interfaces.CustomHookBuilder = (*CustomHookBuilder)(nil)
var _ interfaces.CustomHookFactory = (*CustomHookFactory)(nil)
