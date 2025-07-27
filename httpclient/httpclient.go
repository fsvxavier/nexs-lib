// Package httpclient provides a unified HTTP client with multiple provider backends.
// This version is optimized for dependency injection and connection reuse.
package httpclient

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/batch"
	"github.com/fsvxavier/nexs-lib/httpclient/config"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"github.com/fsvxavier/nexs-lib/httpclient/providers/fasthttp"
	"github.com/fsvxavier/nexs-lib/httpclient/providers/fiber"
	"github.com/fsvxavier/nexs-lib/httpclient/providers/nethttp"
	"github.com/fsvxavier/nexs-lib/httpclient/streaming"
	"github.com/fsvxavier/nexs-lib/httpclient/unmarshaling"
)

// ClientManager manages HTTP clients with connection reuse and dependency injection support.
type ClientManager struct {
	clients map[string]interfaces.Client
	factory interfaces.Factory
	mu      sync.RWMutex
}

// Client implements the HTTP client interface with provider abstraction.
type Client struct {
	provider        interfaces.Provider
	config          *interfaces.Config
	errorHandler    interfaces.ErrorHandler
	retryConfig     *interfaces.RetryConfig
	unmarshalTarget interface{}
	// Add client ID for tracking and reuse
	id string
	// Advanced features
	middlewares []interfaces.Middleware
	hooks       []interfaces.Hook
	mu          sync.RWMutex // Protects middlewares and hooks
}

// Factory implements the factory pattern for creating HTTP clients.
type Factory struct {
	providers map[interfaces.ProviderType]interfaces.ProviderConstructor
}

var (
	// Global client manager instance for dependency injection
	globalManager *ClientManager
	managerOnce   sync.Once
)

// GetManager returns the global client manager instance (singleton pattern).
// This is the recommended way to get clients for dependency injection.
func GetManager() *ClientManager {
	managerOnce.Do(func() {
		globalManager = NewClientManager()
	})
	return globalManager
}

// NewClientManager creates a new client manager for managing HTTP clients.
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]interfaces.Client),
		factory: NewFactory(),
	}
}

// GetOrCreateClient returns an existing client or creates a new one if it doesn't exist.
// This method ensures connection reuse and is ideal for dependency injection scenarios.
func (cm *ClientManager) GetOrCreateClient(name string, providerType interfaces.ProviderType, cfg *interfaces.Config) (interfaces.Client, error) {
	cm.mu.RLock()
	if client, exists := cm.clients[name]; exists {
		cm.mu.RUnlock()
		return client, nil
	}
	cm.mu.RUnlock()

	// Create client if it doesn't exist
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := cm.clients[name]; exists {
		return client, nil
	}

	client, err := cm.factory.CreateClient(providerType, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client %s: %w", name, err)
	}

	// Cast to our internal client type to set ID
	if internalClient, ok := client.(*Client); ok {
		internalClient.id = name
	}

	cm.clients[name] = client
	return client, nil
}

// GetClient returns an existing client by name.
func (cm *ClientManager) GetClient(name string) (interfaces.Client, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, exists := cm.clients[name]
	return client, exists
}

// RemoveClient removes a client from the manager.
func (cm *ClientManager) RemoveClient(name string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clients, name)
}

// ListClients returns a list of all client names.
func (cm *ClientManager) ListClients() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	names := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		names = append(names, name)
	}
	return names
}

// Shutdown gracefully shuts down all clients and cleans up resources.
func (cm *ClientManager) Shutdown() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for name := range cm.clients {
		delete(cm.clients, name)
	}
	return nil
}

// NewFactory creates a new HTTP client factory with registered providers.
func NewFactory() interfaces.Factory {
	factory := &Factory{
		providers: make(map[interfaces.ProviderType]interfaces.ProviderConstructor),
	}

	// Register default providers with optimized configurations for connection reuse
	factory.RegisterProvider(interfaces.ProviderNetHTTP, func(config *interfaces.Config) (interfaces.Provider, error) {
		// Ensure connection reuse is enabled
		if config.MaxIdleConns == 0 {
			config.MaxIdleConns = 100
		}
		if config.IdleConnTimeout == 0 {
			config.IdleConnTimeout = 90 * time.Second
		}
		if config.DisableKeepAlives {
			config.DisableKeepAlives = false // Force keep-alives for reuse
		}
		return nethttp.NewProvider(config)
	})

	factory.RegisterProvider(interfaces.ProviderFiber, func(config *interfaces.Config) (interfaces.Provider, error) {
		return fiber.NewProvider(config)
	})

	factory.RegisterProvider(interfaces.ProviderFastHTTP, func(config *interfaces.Config) (interfaces.Provider, error) {
		return fasthttp.NewProvider(config)
	})

	return factory
}

// CreateClient creates a new HTTP client with the specified provider and configuration.
func (f *Factory) CreateClient(providerType interfaces.ProviderType, cfg *interfaces.Config) (interfaces.Client, error) {
	constructor, exists := f.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Optimize configuration for connection reuse
	cfg = optimizeConfigForReuse(cfg)

	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	provider, err := constructor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	client := &Client{
		provider:    provider,
		config:      cfg,
		retryConfig: cfg.RetryConfig,
	}

	return client, nil
}

// optimizeConfigForReuse optimizes configuration for connection reuse scenarios.
func optimizeConfigForReuse(cfg *interfaces.Config) *interfaces.Config {
	// Clone the config to avoid modifying the original
	optimized := config.CloneConfig(cfg)

	// Set optimal values for connection reuse
	if optimized.MaxIdleConns == 0 {
		optimized.MaxIdleConns = 100 // Allow more idle connections
	}

	if optimized.IdleConnTimeout == 0 {
		optimized.IdleConnTimeout = 90 * time.Second // Keep connections alive longer
	}

	// Force keep-alives for better connection reuse
	optimized.DisableKeepAlives = false

	// Set reasonable defaults for long-running services
	if optimized.TLSHandshakeTimeout == 0 {
		optimized.TLSHandshakeTimeout = 10 * time.Second
	}

	return optimized
}

// RegisterProvider registers a new provider constructor.
func (f *Factory) RegisterProvider(providerType interfaces.ProviderType, constructor interfaces.ProviderConstructor) error {
	if constructor == nil {
		return errors.New("constructor cannot be nil")
	}
	f.providers[providerType] = constructor
	return nil
}

// GetAvailableProviders returns a list of available provider types.
func (f *Factory) GetAvailableProviders() []interfaces.ProviderType {
	providers := make([]interfaces.ProviderType, 0, len(f.providers))
	for providerType := range f.providers {
		providers = append(providers, providerType)
	}
	return providers
}

// Convenience functions for dependency injection

// New creates a new HTTP client using the global manager (recommended for DI).
func New(providerType interfaces.ProviderType, baseURL string) (interfaces.Client, error) {
	cfg := config.DefaultConfig()
	cfg.BaseURL = baseURL
	return GetManager().GetOrCreateClient(fmt.Sprintf("%s_%s", providerType, baseURL), providerType, cfg)
}

// NewWithConfig creates a new HTTP client with custom configuration using the global manager.
func NewWithConfig(providerType interfaces.ProviderType, cfg *interfaces.Config) (interfaces.Client, error) {
	clientID := fmt.Sprintf("%s_%s_%d", providerType, cfg.BaseURL, time.Now().UnixNano())
	return GetManager().GetOrCreateClient(clientID, providerType, cfg)
}

// NewNamed creates a named HTTP client for explicit dependency injection.
func NewNamed(name string, providerType interfaces.ProviderType, baseURL string) (interfaces.Client, error) {
	cfg := config.DefaultConfig()
	cfg.BaseURL = baseURL
	return GetManager().GetOrCreateClient(name, providerType, cfg)
}

// NewNamedWithConfig creates a named HTTP client with custom configuration.
func NewNamedWithConfig(name string, providerType interfaces.ProviderType, cfg *interfaces.Config) (interfaces.Client, error) {
	return GetManager().GetOrCreateClient(name, providerType, cfg)
}

// GetNamedClient retrieves a named client from the global manager.
func GetNamedClient(name string) (interfaces.Client, bool) {
	return GetManager().GetClient(name)
}

// HTTP method implementations remain the same...

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return c.Execute(ctx, "GET", endpoint, nil)
}

// Post performs a POST request with a body.
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return c.Execute(ctx, "POST", endpoint, body)
}

// Put performs a PUT request with a body.
func (c *Client) Put(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return c.Execute(ctx, "PUT", endpoint, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return c.Execute(ctx, "DELETE", endpoint, nil)
}

// Patch performs a PATCH request with a body.
func (c *Client) Patch(ctx context.Context, endpoint string, body interface{}) (*interfaces.Response, error) {
	return c.Execute(ctx, "PATCH", endpoint, body)
}

// Head performs a HEAD request.
func (c *Client) Head(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return c.Execute(ctx, "HEAD", endpoint, nil)
}

// Options performs an OPTIONS request.
func (c *Client) Options(ctx context.Context, endpoint string) (*interfaces.Response, error) {
	return c.Execute(ctx, "OPTIONS", endpoint, nil)
}

// Execute performs an HTTP request with the specified method, endpoint, and body.
func (c *Client) Execute(ctx context.Context, method, endpoint string, body interface{}) (*interfaces.Response, error) {
	req := &interfaces.Request{
		Method:  method,
		URL:     endpoint,
		Headers: make(map[string]string),
		Body:    body,
		Timeout: c.config.Timeout,
		Context: ctx,
	}

	// Execute pre-request hooks
	c.mu.RLock()
	for _, hook := range c.hooks {
		if err := hook.BeforeRequest(ctx, req); err != nil {
			c.mu.RUnlock()
			return nil, err
		}
	}
	c.mu.RUnlock()

	// Execute through middleware chain
	var resp *interfaces.Response
	var err error

	c.mu.RLock()
	middlewares := make([]interfaces.Middleware, len(c.middlewares))
	copy(middlewares, c.middlewares)
	c.mu.RUnlock()

	if len(middlewares) > 0 {
		// Create middleware chain
		handler := func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
			if c.retryConfig != nil && c.retryConfig.MaxRetries > 0 {
				return c.executeWithRetry(ctx, req)
			}
			return c.provider.DoRequest(ctx, req)
		}

		// Execute middleware chain in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			currentHandler := handler
			middleware := middlewares[i]
			handler = func(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
				return middleware.Process(ctx, req, currentHandler)
			}
		}

		resp, err = handler(ctx, req)
	} else {
		// Execute with retry logic
		if c.retryConfig != nil && c.retryConfig.MaxRetries > 0 {
			resp, err = c.executeWithRetry(ctx, req)
		} else {
			resp, err = c.provider.DoRequest(ctx, req)
		}
	}

	// Execute post-request hooks
	c.mu.RLock()
	for _, hook := range c.hooks {
		hook.AfterResponse(ctx, req, resp)
	}
	c.mu.RUnlock()

	if err != nil {
		return resp, err
	}

	// Apply custom error handler if set (should be called for all responses)
	if c.errorHandler != nil && resp != nil {
		if handlerErr := c.errorHandler(resp); handlerErr != nil {
			return resp, handlerErr
		}
	}

	return resp, nil
}

// executeWithRetry executes a request with retry logic.
func (c *Client) executeWithRetry(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
	var lastResp *interfaces.Response
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := c.calculateRetryDelay(attempt)
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := c.provider.DoRequest(ctx, req)
		lastResp = resp
		lastErr = err

		// Check if we should retry
		if c.retryConfig.RetryCondition != nil {
			if !c.retryConfig.RetryCondition(resp, err) {
				break // Don't retry
			}
		} else {
			// Default retry condition: retry on error or 5xx status codes
			if err == nil && (resp == nil || (resp.StatusCode < 500)) {
				break // Don't retry on success or client errors
			}
		}
	}

	return lastResp, lastErr
}

// calculateRetryDelay calculates the delay for the next retry attempt.
func (c *Client) calculateRetryDelay(attempt int) time.Duration {
	if c.retryConfig == nil {
		return time.Second
	}

	delay := c.retryConfig.InitialInterval
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * c.retryConfig.Multiplier)
		if delay > c.retryConfig.MaxInterval {
			delay = c.retryConfig.MaxInterval
			break
		}
	}

	return delay
}

// SetHeaders sets default headers for all requests.
func (c *Client) SetHeaders(headers map[string]string) interfaces.Client {
	if c.config.Headers == nil {
		c.config.Headers = make(map[string]string)
	}
	for key, value := range headers {
		c.config.Headers[key] = value
	}
	return c
}

// SetTimeout sets the request timeout.
func (c *Client) SetTimeout(timeout time.Duration) interfaces.Client {
	c.config.Timeout = timeout
	return c
}

// SetErrorHandler sets a custom error handler.
func (c *Client) SetErrorHandler(handler interfaces.ErrorHandler) interfaces.Client {
	c.errorHandler = handler
	return c
}

// SetRetryConfig sets the retry configuration.
func (c *Client) SetRetryConfig(retryConfig *interfaces.RetryConfig) interfaces.Client {
	c.retryConfig = retryConfig
	return c
}

// Unmarshal sets the target for response unmarshaling.
func (c *Client) Unmarshal(v interface{}) interfaces.Client {
	c.unmarshalTarget = v
	return c
}

// GetProvider returns the underlying provider.
func (c *Client) GetProvider() interfaces.Provider {
	return c.provider
}

// GetConfig returns the client configuration.
func (c *Client) GetConfig() *interfaces.Config {
	return c.config
}

// IsHealthy checks if the client is healthy and ready to handle requests.
func (c *Client) IsHealthy() bool {
	return c.provider.IsHealthy()
}

// GetMetrics returns the provider metrics.
func (c *Client) GetMetrics() *interfaces.ProviderMetrics {
	return c.provider.GetMetrics()
}

// GetID returns the client ID (useful for tracking and debugging).
func (c *Client) GetID() string {
	return c.id
}

// AddMiddleware adds a middleware to the client.
func (c *Client) AddMiddleware(middleware interfaces.Middleware) interfaces.Client {
	if middleware == nil {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if middleware already exists
	for _, m := range c.middlewares {
		if m == middleware {
			return c // Already exists, don't add duplicate
		}
	}

	c.middlewares = append(c.middlewares, middleware)
	return c
}

// RemoveMiddleware removes a middleware from the client.
func (c *Client) RemoveMiddleware(middleware interfaces.Middleware) interfaces.Client {
	if middleware == nil {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for i, m := range c.middlewares {
		if m == middleware {
			// Remove middleware by slicing
			c.middlewares = append(c.middlewares[:i], c.middlewares[i+1:]...)
			break
		}
	}

	return c
}

// AddHook adds a hook to the client.
func (c *Client) AddHook(hook interfaces.Hook) interfaces.Client {
	if hook == nil {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if hook already exists
	for _, h := range c.hooks {
		if h == hook {
			return c // Already exists, don't add duplicate
		}
	}

	c.hooks = append(c.hooks, hook)
	return c
}

// RemoveHook removes a hook from the client.
func (c *Client) RemoveHook(hook interfaces.Hook) interfaces.Client {
	if hook == nil {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for i, h := range c.hooks {
		if h == hook {
			// Remove hook by slicing
			c.hooks = append(c.hooks[:i], c.hooks[i+1:]...)
			break
		}
	}

	return c
}

// Batch creates a new batch request builder.
func (c *Client) Batch() interfaces.BatchRequestBuilder {
	return batch.NewBuilder(c)
}

// Stream performs a streaming request.
func (c *Client) Stream(ctx context.Context, method, endpoint string, handler interfaces.StreamHandler) error {
	if handler == nil {
		return fmt.Errorf("stream handler cannot be nil")
	}

	processor := streaming.NewStreamProcessor(c)

	// Use the endpoint directly for streaming
	return processor.StreamDownload(ctx, method, endpoint, handler)
}

// UnmarshalResponse unmarshals response body to a target.
func (c *Client) UnmarshalResponse(response *interfaces.Response, target interface{}) error {
	if response == nil {
		return fmt.Errorf("response cannot be nil")
	}
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	// Use the unmarshaling package with auto strategy
	unmarshaler := unmarshaling.NewUnmarshaler(interfaces.UnmarshalAuto)

	return unmarshaler.UnmarshalResponse(response, target)
}
