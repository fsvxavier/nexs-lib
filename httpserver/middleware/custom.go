// Package middleware provides custom middleware implementations.
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// CustomMiddleware implements a fully customizable middleware.
type CustomMiddleware struct {
	name         string
	priority     int
	enabled      bool
	condition    func(r *http.Request) bool
	pathFilter   func(path string) bool
	methodFilter func(method string) bool
	headerFilter func(headers http.Header) bool
	middleware   func(next http.Handler) http.Handler
	timeout      time.Duration
	skipOnError  bool
	skipPaths    []string
	skipFunc     func(path string) bool
	config       interface{}
	beforeFunc   func(w http.ResponseWriter, r *http.Request)
	afterFunc    func(w http.ResponseWriter, r *http.Request, statusCode int, duration time.Duration)
}

// NewCustomMiddleware creates a new custom middleware with the provided configuration.
func NewCustomMiddleware(name string, priority int, middleware func(next http.Handler) http.Handler) *CustomMiddleware {
	return &CustomMiddleware{
		name:       name,
		priority:   priority,
		enabled:    true,
		middleware: middleware,
		timeout:    30 * time.Second,
	}
}

// Name returns the middleware name.
func (m *CustomMiddleware) Name() string {
	return m.name
}

// Priority returns the middleware priority.
func (m *CustomMiddleware) Priority() int {
	return m.priority
}

// IsEnabled returns whether the middleware is enabled.
func (m *CustomMiddleware) IsEnabled() bool {
	return m.enabled
}

// SetEnabled sets the enabled state of the middleware.
func (m *CustomMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// Wrap wraps an http.Handler with the middleware functionality.
func (m *CustomMiddleware) Wrap(next http.Handler) http.Handler {
	if !m.enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check filters
		if !m.shouldApply(r) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Execute before function if set
		if m.beforeFunc != nil {
			m.beforeFunc(w, r)
		}

		// Create a response writer to capture status code if afterFunc is set
		var rw *responseWriter
		if m.afterFunc != nil {
			rw = &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			w = rw
		}

		// Apply the custom middleware
		if m.middleware != nil {
			m.middleware(next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}

		// Execute after function if set
		if m.afterFunc != nil && rw != nil {
			duration := time.Since(start)
			m.afterFunc(w, r, rw.statusCode, duration)
		}
	})
}

// shouldApply determines if the middleware should be applied to the request.
func (m *CustomMiddleware) shouldApply(r *http.Request) bool {
	// Check skip paths
	for _, skipPath := range m.skipPaths {
		if r.URL.Path == skipPath {
			return false
		}
	}

	// Check skip function
	if m.skipFunc != nil && m.skipFunc(r.URL.Path) {
		return false
	}

	// Check condition
	if m.condition != nil && !m.condition(r) {
		return false
	}

	// Check filters
	if m.pathFilter != nil && !m.pathFilter(r.URL.Path) {
		return false
	}
	if m.methodFilter != nil && !m.methodFilter(r.Method) {
		return false
	}
	if m.headerFilter != nil && !m.headerFilter(r.Header) {
		return false
	}

	return true
}

// SetCondition sets a condition function for conditional execution.
func (m *CustomMiddleware) SetCondition(condition func(r *http.Request) bool) {
	m.condition = condition
}

// SetPathFilter sets a path filter for the middleware.
func (m *CustomMiddleware) SetPathFilter(filter func(path string) bool) {
	m.pathFilter = filter
}

// SetMethodFilter sets a method filter for the middleware.
func (m *CustomMiddleware) SetMethodFilter(filter func(method string) bool) {
	m.methodFilter = filter
}

// SetHeaderFilter sets a header filter for the middleware.
func (m *CustomMiddleware) SetHeaderFilter(filter func(headers http.Header) bool) {
	m.headerFilter = filter
}

// SetSkipFunc sets the skip function for the middleware.
func (m *CustomMiddleware) SetSkipFunc(skipFunc func(path string) bool) {
	m.skipFunc = skipFunc
}

// SetSkipPaths sets the skip paths for the middleware.
func (m *CustomMiddleware) SetSkipPaths(paths []string) {
	m.skipPaths = paths
}

// SetConfig sets the configuration for the middleware.
func (m *CustomMiddleware) SetConfig(config interface{}) {
	m.config = config
}

// SetBeforeFunc sets the before function for the middleware.
func (m *CustomMiddleware) SetBeforeFunc(beforeFunc func(w http.ResponseWriter, r *http.Request)) {
	m.beforeFunc = beforeFunc
}

// SetAfterFunc sets the after function for the middleware.
func (m *CustomMiddleware) SetAfterFunc(afterFunc func(w http.ResponseWriter, r *http.Request, statusCode int, duration time.Duration)) {
	m.afterFunc = afterFunc
}

// SetTimeout sets the timeout for the middleware.
func (m *CustomMiddleware) SetTimeout(timeout time.Duration) {
	m.timeout = timeout
}

// SetSkipOnError sets whether to skip the middleware on error.
func (m *CustomMiddleware) SetSkipOnError(skip bool) {
	m.skipOnError = skip
}

// Timeout returns the timeout for the middleware.
func (m *CustomMiddleware) Timeout() time.Duration {
	return m.timeout
}

// SkipOnError returns whether to skip the middleware on error.
func (m *CustomMiddleware) SkipOnError() bool {
	return m.skipOnError
}

// CustomMiddlewareBuilder implements the CustomMiddlewareBuilder interface.
type CustomMiddlewareBuilder struct {
	middleware *CustomMiddleware
}

// NewCustomMiddlewareBuilder creates a new custom middleware builder.
func NewCustomMiddlewareBuilder() *CustomMiddlewareBuilder {
	return &CustomMiddlewareBuilder{
		middleware: &CustomMiddleware{
			enabled: true,
			timeout: 30 * time.Second,
		},
	}
}

// WithName sets the name of the custom middleware.
func (b *CustomMiddlewareBuilder) WithName(name string) interfaces.CustomMiddlewareBuilder {
	b.middleware.name = name
	return b
}

// WithPriority sets the priority of the middleware.
func (b *CustomMiddlewareBuilder) WithPriority(priority int) interfaces.CustomMiddlewareBuilder {
	b.middleware.priority = priority
	return b
}

// WithCondition sets a condition function for conditional execution.
func (b *CustomMiddlewareBuilder) WithCondition(condition func(r *http.Request) bool) interfaces.CustomMiddlewareBuilder {
	b.middleware.condition = condition
	return b
}

// WithPathFilter sets a path filter for the middleware.
func (b *CustomMiddlewareBuilder) WithPathFilter(filter func(path string) bool) interfaces.CustomMiddlewareBuilder {
	b.middleware.pathFilter = filter
	return b
}

// WithMethodFilter sets a method filter for the middleware.
func (b *CustomMiddlewareBuilder) WithMethodFilter(filter func(method string) bool) interfaces.CustomMiddlewareBuilder {
	b.middleware.methodFilter = filter
	return b
}

// WithHeaderFilter sets a header filter for the middleware.
func (b *CustomMiddlewareBuilder) WithHeaderFilter(filter func(headers http.Header) bool) interfaces.CustomMiddlewareBuilder {
	b.middleware.headerFilter = filter
	return b
}

// WithTimeout sets the timeout for the middleware.
func (b *CustomMiddlewareBuilder) WithTimeout(timeout time.Duration) interfaces.CustomMiddlewareBuilder {
	b.middleware.timeout = timeout
	return b
}

// WithSkipOnError sets whether to skip the middleware on error.
func (b *CustomMiddlewareBuilder) WithSkipOnError(skip bool) interfaces.CustomMiddlewareBuilder {
	b.middleware.skipOnError = skip
	return b
}

// WithHandler sets the middleware handler function.
func (b *CustomMiddlewareBuilder) WithHandler(handler func(next http.Handler) http.Handler) interfaces.CustomMiddlewareBuilder {
	b.middleware.middleware = handler
	return b
}

// WithSkipPaths sets paths to skip for this middleware.
func (b *CustomMiddlewareBuilder) WithSkipPaths(paths ...string) interfaces.CustomMiddlewareBuilder {
	b.middleware.skipPaths = paths
	return b
}

// WithSkipFunc sets a custom skip function.
func (b *CustomMiddlewareBuilder) WithSkipFunc(fn func(path string) bool) interfaces.CustomMiddlewareBuilder {
	b.middleware.skipFunc = fn
	return b
}

// WithConfig sets configuration for the middleware.
func (b *CustomMiddlewareBuilder) WithConfig(config interface{}) interfaces.CustomMiddlewareBuilder {
	b.middleware.config = config
	return b
}

// WithWrapFunc sets the main wrapper function.
func (b *CustomMiddlewareBuilder) WithWrapFunc(fn func(http.Handler) http.Handler) interfaces.CustomMiddlewareBuilder {
	b.middleware.middleware = fn
	return b
}

// WithBeforeFunc sets a function to execute before the request.
func (b *CustomMiddlewareBuilder) WithBeforeFunc(fn func(w http.ResponseWriter, r *http.Request)) interfaces.CustomMiddlewareBuilder {
	b.middleware.beforeFunc = fn
	return b
}

// WithAfterFunc sets a function to execute after the request.
func (b *CustomMiddlewareBuilder) WithAfterFunc(fn func(w http.ResponseWriter, r *http.Request, statusCode int, duration time.Duration)) interfaces.CustomMiddlewareBuilder {
	b.middleware.afterFunc = fn
	return b
}

// Build creates the custom middleware.
func (b *CustomMiddlewareBuilder) Build() (interfaces.Middleware, error) {
	if b.middleware.name == "" {
		return nil, fmt.Errorf("middleware name is required")
	}
	if b.middleware.middleware == nil {
		return nil, fmt.Errorf("middleware handler is required")
	}

	return b.middleware, nil
}

// CustomMiddlewareFactory implements the CustomMiddlewareFactory interface.
type CustomMiddlewareFactory struct{}

// NewCustomMiddlewareFactory creates a new custom middleware factory.
func NewCustomMiddlewareFactory() *CustomMiddlewareFactory {
	return &CustomMiddlewareFactory{}
}

// NewBuilder creates a new custom middleware builder.
func (f *CustomMiddlewareFactory) NewBuilder() interfaces.CustomMiddlewareBuilder {
	return NewCustomMiddlewareBuilder()
}

// NewSimpleMiddleware creates a simple custom middleware with basic configuration.
func (f *CustomMiddlewareFactory) NewSimpleMiddleware(name string, priority int, handler func(next http.Handler) http.Handler) interfaces.Middleware {
	return NewCustomMiddleware(name, priority, handler)
}

// NewConditionalMiddleware creates a conditional custom middleware.
func (f *CustomMiddlewareFactory) NewConditionalMiddleware(name string, priority int, skipFunc func(string) bool, handler func(next http.Handler) http.Handler) interfaces.Middleware {
	middleware := NewCustomMiddleware(name, priority, handler)
	middleware.SetSkipFunc(skipFunc)
	return middleware
}

// NewPathFilteredMiddleware creates a path-filtered custom middleware.
func (f *CustomMiddlewareFactory) NewPathFilteredMiddleware(name string, priority int, pathFilter func(string) bool, handler func(next http.Handler) http.Handler) interfaces.Middleware {
	middleware := NewCustomMiddleware(name, priority, handler)
	middleware.SetPathFilter(pathFilter)
	return middleware
}

// NewMethodFilteredMiddleware creates a method-filtered custom middleware.
func (f *CustomMiddlewareFactory) NewMethodFilteredMiddleware(name string, priority int, methodFilter func(string) bool, handler func(next http.Handler) http.Handler) interfaces.Middleware {
	middleware := NewCustomMiddleware(name, priority, handler)
	middleware.SetMethodFilter(methodFilter)
	return middleware
}

// NewTimeoutMiddleware creates a middleware with timeout configuration.
func (f *CustomMiddlewareFactory) NewTimeoutMiddleware(name string, priority int, timeout time.Duration, handler func(next http.Handler) http.Handler) interfaces.Middleware {
	middleware := NewCustomMiddleware(name, priority, handler)
	middleware.SetTimeout(timeout)
	return middleware
}

// NewTimingMiddleware creates a timing middleware with custom handler.
func (f *CustomMiddlewareFactory) NewTimingMiddleware(name string, priority int, handler func(duration time.Duration, path string)) interfaces.Middleware {
	return f.NewSimpleMiddleware(name, priority, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			if handler != nil {
				handler(duration, r.URL.Path)
			}
		})
	})
}

// NewLoggingMiddleware creates a logging middleware with custom logger.
func (f *CustomMiddlewareFactory) NewLoggingMiddleware(name string, priority int, logger func(method, path string, statusCode int, duration time.Duration)) interfaces.Middleware {
	return f.NewSimpleMiddleware(name, priority, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			if logger != nil {
				logger(r.Method, r.URL.Path, rw.statusCode, duration)
			}
		})
	})
}

// Common Custom Middleware Helpers

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Verify that CustomMiddleware implements the Middleware interface
var _ interfaces.Middleware = (*CustomMiddleware)(nil)
var _ interfaces.CustomMiddlewareBuilder = (*CustomMiddlewareBuilder)(nil)
var _ interfaces.CustomMiddlewareFactory = (*CustomMiddlewareFactory)(nil)
