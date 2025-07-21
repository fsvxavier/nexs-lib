// Package hooks provides base implementations for HTTP server hooks.
package hooks

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// BaseHook provides a base implementation for the Hook interface.
type BaseHook struct {
	name     string
	events   []interfaces.HookEvent
	priority int
	enabled  bool
	mu       sync.RWMutex
}

// NewBaseHook creates a new base hook.
func NewBaseHook(name string, events []interfaces.HookEvent, priority int) *BaseHook {
	return &BaseHook{
		name:     name,
		events:   events,
		priority: priority,
		enabled:  true,
	}
}

// Execute executes the hook with the given context.
func (h *BaseHook) Execute(ctx *interfaces.HookContext) error {
	// Base implementation does nothing
	return nil
}

// Name returns the hook name for identification.
func (h *BaseHook) Name() string {
	return h.name
}

// Events returns the events this hook handles.
func (h *BaseHook) Events() []interfaces.HookEvent {
	return h.events
}

// Priority returns the hook priority (lower numbers execute first).
func (h *BaseHook) Priority() int {
	return h.priority
}

// IsEnabled returns whether the hook is enabled.
func (h *BaseHook) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.enabled
}

// SetEnabled sets the enabled state of the hook.
func (h *BaseHook) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = enabled
}

// ShouldExecute determines if the hook should execute for the given context.
func (h *BaseHook) ShouldExecute(ctx *interfaces.HookContext) bool {
	return h.IsEnabled()
}

// ConditionalBaseHook extends BaseHook with conditional execution.
type ConditionalBaseHook struct {
	*BaseHook
	condition func(ctx *interfaces.HookContext) bool
}

// NewConditionalBaseHook creates a new conditional base hook.
func NewConditionalBaseHook(name string, events []interfaces.HookEvent, priority int, condition func(ctx *interfaces.HookContext) bool) *ConditionalBaseHook {
	return &ConditionalBaseHook{
		BaseHook:  NewBaseHook(name, events, priority),
		condition: condition,
	}
}

// ShouldExecute determines if the hook should execute for the given context.
func (h *ConditionalBaseHook) ShouldExecute(ctx *interfaces.HookContext) bool {
	if !h.BaseHook.ShouldExecute(ctx) {
		return false
	}
	if h.condition != nil {
		return h.condition(ctx)
	}
	return true
}

// Condition returns the condition function.
func (h *ConditionalBaseHook) Condition() func(ctx *interfaces.HookContext) bool {
	return h.condition
}

// FilteredBaseHook extends BaseHook with filtering capabilities.
type FilteredBaseHook struct {
	*BaseHook
	pathFilter   func(path string) bool
	methodFilter func(method string) bool
	headerFilter func(headers http.Header) bool
}

// NewFilteredBaseHook creates a new filtered base hook.
func NewFilteredBaseHook(name string, events []interfaces.HookEvent, priority int) *FilteredBaseHook {
	return &FilteredBaseHook{
		BaseHook: NewBaseHook(name, events, priority),
	}
}

// SetPathFilter sets the path filter function.
func (h *FilteredBaseHook) SetPathFilter(filter func(path string) bool) {
	h.pathFilter = filter
}

// SetMethodFilter sets the method filter function.
func (h *FilteredBaseHook) SetMethodFilter(filter func(method string) bool) {
	h.methodFilter = filter
}

// SetHeaderFilter sets the header filter function.
func (h *FilteredBaseHook) SetHeaderFilter(filter func(headers http.Header) bool) {
	h.headerFilter = filter
}

// ShouldExecute determines if the hook should execute for the given context.
func (h *FilteredBaseHook) ShouldExecute(ctx *interfaces.HookContext) bool {
	if !h.BaseHook.ShouldExecute(ctx) {
		return false
	}

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

// PathFilter returns the path filter function.
func (h *FilteredBaseHook) PathFilter() func(path string) bool {
	return h.pathFilter
}

// MethodFilter returns the method filter function.
func (h *FilteredBaseHook) MethodFilter() func(method string) bool {
	return h.methodFilter
}

// HeaderFilter returns the header filter function.
func (h *FilteredBaseHook) HeaderFilter() func(headers http.Header) bool {
	return h.headerFilter
}

// AsyncBaseHook extends BaseHook with asynchronous execution.
type AsyncBaseHook struct {
	*BaseHook
	bufferSize int
	timeout    time.Duration
}

// NewAsyncBaseHook creates a new async base hook.
func NewAsyncBaseHook(name string, events []interfaces.HookEvent, priority int, bufferSize int, timeout time.Duration) *AsyncBaseHook {
	if bufferSize <= 0 {
		bufferSize = 10
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &AsyncBaseHook{
		BaseHook:   NewBaseHook(name, events, priority),
		bufferSize: bufferSize,
		timeout:    timeout,
	}
}

// ExecuteAsync executes the hook asynchronously.
func (h *AsyncBaseHook) ExecuteAsync(ctx *interfaces.HookContext) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)
		if err := h.Execute(ctx); err != nil {
			errChan <- err
		}
	}()

	return errChan
}

// BufferSize returns the buffer size for async execution.
func (h *AsyncBaseHook) BufferSize() int {
	return h.bufferSize
}

// Timeout returns the timeout for async execution.
func (h *AsyncBaseHook) Timeout() time.Duration {
	return h.timeout
}

// HookChain provides a chain implementation for hooks.
type HookChain struct {
	hooks []interfaces.Hook
}

// NewHookChain creates a new hook chain.
func NewHookChain() *HookChain {
	return &HookChain{
		hooks: make([]interfaces.Hook, 0),
	}
}

// Add adds a hook to the chain.
func (c *HookChain) Add(hook interfaces.Hook) interfaces.HookChain {
	c.hooks = append(c.hooks, hook)
	return c
}

// Execute executes all hooks in the chain.
func (c *HookChain) Execute(ctx *interfaces.HookContext) error {
	var lastError error
	for _, hook := range c.hooks {
		if !hook.IsEnabled() || !hook.ShouldExecute(ctx) {
			continue
		}
		if err := hook.Execute(ctx); err != nil {
			lastError = err
		}
	}
	return lastError
}

// ExecuteUntil executes hooks until a condition is met.
func (c *HookChain) ExecuteUntil(ctx *interfaces.HookContext, condition func(*interfaces.HookContext) bool) error {
	for _, hook := range c.hooks {
		if !hook.IsEnabled() || !hook.ShouldExecute(ctx) {
			continue
		}
		if err := hook.Execute(ctx); err != nil {
			return err
		}
		if condition(ctx) {
			break
		}
	}
	return nil
}

// ExecuteIf executes hooks if a condition is met.
func (c *HookChain) ExecuteIf(ctx *interfaces.HookContext, condition func(*interfaces.HookContext) bool) error {
	if !condition(ctx) {
		return nil
	}
	return c.Execute(ctx)
}

// PathFilterBuilder helps build path filter functions.
type PathFilterBuilder struct {
	includePaths []string
	excludePaths []string
	prefixes     []string
	suffixes     []string
}

// NewPathFilterBuilder creates a new path filter builder.
func NewPathFilterBuilder() *PathFilterBuilder {
	return &PathFilterBuilder{}
}

// Include adds paths to include.
func (b *PathFilterBuilder) Include(paths ...string) *PathFilterBuilder {
	b.includePaths = append(b.includePaths, paths...)
	return b
}

// Exclude adds paths to exclude.
func (b *PathFilterBuilder) Exclude(paths ...string) *PathFilterBuilder {
	b.excludePaths = append(b.excludePaths, paths...)
	return b
}

// WithPrefix adds path prefixes to match.
func (b *PathFilterBuilder) WithPrefix(prefixes ...string) *PathFilterBuilder {
	b.prefixes = append(b.prefixes, prefixes...)
	return b
}

// WithSuffix adds path suffixes to match.
func (b *PathFilterBuilder) WithSuffix(suffixes ...string) *PathFilterBuilder {
	b.suffixes = append(b.suffixes, suffixes...)
	return b
}

// Build builds the path filter function.
func (b *PathFilterBuilder) Build() func(path string) bool {
	return func(path string) bool {
		// Check excludes first
		for _, exclude := range b.excludePaths {
			if path == exclude {
				return false
			}
		}

		// Check prefixes
		for _, prefix := range b.prefixes {
			if strings.HasPrefix(path, prefix) {
				return false
			}
		}

		// Check suffixes
		for _, suffix := range b.suffixes {
			if strings.HasSuffix(path, suffix) {
				return false
			}
		}

		// If includes are specified, path must be in includes
		if len(b.includePaths) > 0 {
			for _, include := range b.includePaths {
				if path == include {
					return true
				}
			}
			return false
		}

		return true
	}
}

// MethodFilterBuilder helps build method filter functions.
type MethodFilterBuilder struct {
	allowedMethods []string
	deniedMethods  []string
}

// NewMethodFilterBuilder creates a new method filter builder.
func NewMethodFilterBuilder() *MethodFilterBuilder {
	return &MethodFilterBuilder{}
}

// Allow adds allowed HTTP methods.
func (b *MethodFilterBuilder) Allow(methods ...string) *MethodFilterBuilder {
	for _, method := range methods {
		b.allowedMethods = append(b.allowedMethods, strings.ToUpper(method))
	}
	return b
}

// Deny adds denied HTTP methods.
func (b *MethodFilterBuilder) Deny(methods ...string) *MethodFilterBuilder {
	for _, method := range methods {
		b.deniedMethods = append(b.deniedMethods, strings.ToUpper(method))
	}
	return b
}

// Build builds the method filter function.
func (b *MethodFilterBuilder) Build() func(method string) bool {
	return func(method string) bool {
		method = strings.ToUpper(method)

		// Check denied methods first
		for _, denied := range b.deniedMethods {
			if method == denied {
				return false
			}
		}

		// If allowed methods are specified, method must be in allowed list
		if len(b.allowedMethods) > 0 {
			for _, allowed := range b.allowedMethods {
				if method == allowed {
					return true
				}
			}
			return false
		}

		return true
	}
}

// Verify implementations
var _ interfaces.Hook = (*BaseHook)(nil)
var _ interfaces.Hook = (*ConditionalBaseHook)(nil)
var _ interfaces.ConditionalHook = (*ConditionalBaseHook)(nil)
var _ interfaces.Hook = (*FilteredBaseHook)(nil)
var _ interfaces.FilteredHook = (*FilteredBaseHook)(nil)
var _ interfaces.Hook = (*AsyncBaseHook)(nil)
var _ interfaces.AsyncHook = (*AsyncBaseHook)(nil)
var _ interfaces.HookChain = (*HookChain)(nil)
