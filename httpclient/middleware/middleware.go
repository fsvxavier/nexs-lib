// Package middleware provides middleware chain support for request/response processing.
package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Chain represents a middleware chain for processing HTTP requests and responses.
type Chain struct {
	middlewares []interfaces.Middleware
	mu          sync.RWMutex
}

// NewChain creates a new middleware chain.
func NewChain() *Chain {
	return &Chain{
		middlewares: make([]interfaces.Middleware, 0),
	}
}

// Add adds a middleware to the end of the chain.
func (c *Chain) Add(middleware interfaces.Middleware) *Chain {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Remove removes a middleware from the chain.
func (c *Chain) Remove(middleware interfaces.Middleware) *Chain {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, m := range c.middlewares {
		if m == middleware {
			c.middlewares = append(c.middlewares[:i], c.middlewares[i+1:]...)
			break
		}
	}
	return c
}

// Execute executes the middleware chain.
func (c *Chain) Execute(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	c.mu.RLock()
	middlewares := make([]interfaces.Middleware, len(c.middlewares))
	copy(middlewares, c.middlewares)
	c.mu.RUnlock()

	if len(middlewares) == 0 {
		return next(ctx, req)
	}

	return c.executeMiddleware(ctx, req, middlewares, 0, next)
}

// executeMiddleware recursively executes middleware in the chain.
func (c *Chain) executeMiddleware(ctx context.Context, req *interfaces.Request, middlewares []interfaces.Middleware, index int, finalNext func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	if index >= len(middlewares) {
		return finalNext(ctx, req)
	}

	middleware := middlewares[index]
	next := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
		return c.executeMiddleware(ctx, req, middlewares, index+1, finalNext)
	}

	return middleware.Process(ctx, req, next)
}

// Count returns the number of middlewares in the chain.
func (c *Chain) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.middlewares)
}

// Clear removes all middlewares from the chain.
func (c *Chain) Clear() *Chain {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.middlewares = c.middlewares[:0]
	return c
}

// Built-in middleware implementations

// LoggingMiddleware logs request and response information.
type LoggingMiddleware struct {
	logger func(format string, args ...interface{})
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(logger func(format string, args ...interface{})) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// Process implements the Middleware interface.
func (m *LoggingMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	start := time.Now()
	m.logger("HTTP Request: %s %s", req.Method, req.URL)

	resp, err := next(ctx, req)

	duration := time.Since(start)
	if err != nil {
		m.logger("HTTP Request failed: %s %s - Error: %v - Duration: %v", req.Method, req.URL, err, duration)
	} else {
		m.logger("HTTP Response: %s %s - Status: %d - Duration: %v", req.Method, req.URL, resp.StatusCode, duration)
	}

	return resp, err
}

// RetryMiddleware implements automatic retry logic.
type RetryMiddleware struct {
	maxRetries int
	retryFunc  func(*interfaces.Response, error) bool
}

// NewRetryMiddleware creates a new retry middleware.
func NewRetryMiddleware(maxRetries int, retryFunc func(*interfaces.Response, error) bool) *RetryMiddleware {
	if retryFunc == nil {
		retryFunc = DefaultRetryCondition
	}
	return &RetryMiddleware{
		maxRetries: maxRetries,
		retryFunc:  retryFunc,
	}
}

// Process implements the Middleware interface.
func (m *RetryMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	var lastResp *interfaces.Response
	var lastErr error

	for attempt := 0; attempt <= m.maxRetries; attempt++ {
		resp, err := next(ctx, req)

		if err == nil && !m.retryFunc(resp, err) {
			return resp, nil
		}

		lastResp = resp
		lastErr = err

		if attempt < m.maxRetries {
			// Wait before retry (exponential backoff could be added here)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 100 * time.Millisecond):
			}
		}
	}

	return lastResp, lastErr
}

// DefaultRetryCondition is the default retry condition.
func DefaultRetryCondition(resp *interfaces.Response, err error) bool {
	if err != nil {
		return true
	}

	if resp == nil {
		return true
	}

	// Retry on server errors (5xx) and some client errors
	return resp.StatusCode >= 500 || resp.StatusCode == 429 || resp.StatusCode == 408
}

// AuthMiddleware adds authentication headers to requests.
type AuthMiddleware struct {
	headerName  string
	headerValue string
	provider    func() string
}

// NewAuthMiddleware creates a new authentication middleware with static token.
func NewAuthMiddleware(headerName, headerValue string) *AuthMiddleware {
	return &AuthMiddleware{
		headerName:  headerName,
		headerValue: headerValue,
	}
}

// NewDynamicAuthMiddleware creates a new authentication middleware with dynamic token provider.
func NewDynamicAuthMiddleware(headerName string, provider func() string) *AuthMiddleware {
	return &AuthMiddleware{
		headerName: headerName,
		provider:   provider,
	}
}

// Process implements the Middleware interface.
func (m *AuthMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}

	value := m.headerValue
	if m.provider != nil {
		value = m.provider()
	}

	req.Headers[m.headerName] = value
	return next(ctx, req)
}

// CompressionMiddleware handles request compression.
type CompressionMiddleware struct {
	compressionTypes []interfaces.CompressionType
}

// NewCompressionMiddleware creates a new compression middleware.
func NewCompressionMiddleware(types ...interfaces.CompressionType) *CompressionMiddleware {
	if len(types) == 0 {
		types = []interfaces.CompressionType{interfaces.CompressionGzip}
	}
	return &CompressionMiddleware{
		compressionTypes: types,
	}
}

// Process implements the Middleware interface.
func (m *CompressionMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}

	// Add Accept-Encoding header
	encodings := make([]string, len(m.compressionTypes))
	for i, ct := range m.compressionTypes {
		encodings[i] = string(ct)
	}
	req.Headers["Accept-Encoding"] = fmt.Sprintf("%s", encodings[0])
	if len(encodings) > 1 {
		for i := 1; i < len(encodings); i++ {
			req.Headers["Accept-Encoding"] += ", " + encodings[i]
		}
	}

	return next(ctx, req)
}

// MetricsMiddleware collects request metrics.
type MetricsMiddleware struct {
	collector func(method, url string, statusCode int, duration time.Duration, err error)
}

// NewMetricsMiddleware creates a new metrics middleware.
func NewMetricsMiddleware(collector func(method, url string, statusCode int, duration time.Duration, err error)) *MetricsMiddleware {
	return &MetricsMiddleware{collector: collector}
}

// Process implements the Middleware interface.
func (m *MetricsMiddleware) Process(ctx context.Context, req *interfaces.Request, next func(context.Context, *interfaces.Request) (*interfaces.Response, error)) (*interfaces.Response, error) {
	start := time.Now()
	resp, err := next(ctx, req)
	duration := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	if m.collector != nil {
		m.collector(req.Method, req.URL, statusCode, duration, err)
	}

	return resp, err
}
