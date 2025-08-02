// Package config provides extensible configuration management for HTTP servers.
// It implements the builder pattern with functional options for flexible configuration.
package config

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// BaseConfig provides the default configuration for HTTP servers.
// It implements the Config interface and can be extended by provider-specific configurations.
type BaseConfig struct {
	addr             string
	port             int
	readTimeout      time.Duration
	writeTimeout     time.Duration
	idleTimeout      time.Duration
	gracefulShutdown bool
	shutdownTimeout  time.Duration
	provider         string
	middlewares      []interface{}
	observers        []interfaces.ServerObserver
	hooks            map[interfaces.EventType][]interfaces.HookFunc
	custom           map[string]interface{}
}

// NewBaseConfig creates a new BaseConfig with sensible defaults.
func NewBaseConfig() *BaseConfig {
	return &BaseConfig{
		addr:             "0.0.0.0",
		port:             8080,
		readTimeout:      30 * time.Second,
		writeTimeout:     30 * time.Second,
		idleTimeout:      60 * time.Second,
		gracefulShutdown: true,
		shutdownTimeout:  30 * time.Second,
		middlewares:      make([]interface{}, 0),
		observers:        make([]interfaces.ServerObserver, 0),
		hooks:            make(map[interfaces.EventType][]interfaces.HookFunc),
		custom:           make(map[string]interface{}),
	}
}

// GetAddr returns the address to bind the server to.
func (c *BaseConfig) GetAddr() string {
	return c.addr
}

// GetPort returns the port to bind the server to.
func (c *BaseConfig) GetPort() int {
	return c.port
}

// GetReadTimeout returns the read timeout for requests.
func (c *BaseConfig) GetReadTimeout() time.Duration {
	return c.readTimeout
}

// GetWriteTimeout returns the write timeout for responses.
func (c *BaseConfig) GetWriteTimeout() time.Duration {
	return c.writeTimeout
}

// GetIdleTimeout returns the idle timeout for connections.
func (c *BaseConfig) GetIdleTimeout() time.Duration {
	return c.idleTimeout
}

// IsGracefulShutdown returns true if graceful shutdown is enabled.
func (c *BaseConfig) IsGracefulShutdown() bool {
	return c.gracefulShutdown
}

// GetShutdownTimeout returns the timeout for graceful shutdown.
func (c *BaseConfig) GetShutdownTimeout() time.Duration {
	return c.shutdownTimeout
}

// GetProvider returns the provider name.
func (c *BaseConfig) GetProvider() string {
	return c.provider
}

// GetMiddlewares returns the registered middlewares.
func (c *BaseConfig) GetMiddlewares() []interface{} {
	return c.middlewares
}

// GetObservers returns the registered observers.
func (c *BaseConfig) GetObservers() []interfaces.ServerObserver {
	return c.observers
}

// GetHooks returns the registered hooks for the specified event type.
func (c *BaseConfig) GetHooks(eventType interfaces.EventType) []interfaces.HookFunc {
	return c.hooks[eventType]
}

// GetCustom returns a custom configuration value.
func (c *BaseConfig) GetCustom(key string) (interface{}, bool) {
	value, exists := c.custom[key]
	return value, exists
}

// SetCustom sets a custom configuration value.
func (c *BaseConfig) SetCustom(key string, value interface{}) *BaseConfig {
	c.custom[key] = value
	return c
}

// GetFullAddr returns the full address string (addr:port).
func (c *BaseConfig) GetFullAddr() string {
	return fmt.Sprintf("%s:%d", c.addr, c.port)
}

// Validate validates the configuration.
func (c *BaseConfig) Validate() error {
	if c.port <= 0 || c.port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 1-65535)", c.port)
	}

	if c.readTimeout <= 0 {
		return fmt.Errorf("invalid read timeout: %v (must be positive)", c.readTimeout)
	}

	if c.writeTimeout <= 0 {
		return fmt.Errorf("invalid write timeout: %v (must be positive)", c.writeTimeout)
	}

	if c.idleTimeout <= 0 {
		return fmt.Errorf("invalid idle timeout: %v (must be positive)", c.idleTimeout)
	}

	if c.shutdownTimeout <= 0 {
		return fmt.Errorf("invalid shutdown timeout: %v (must be positive)", c.shutdownTimeout)
	}

	return nil
}

// Clone creates a deep copy of the configuration.
func (c *BaseConfig) Clone() *BaseConfig {
	clone := &BaseConfig{
		addr:             c.addr,
		port:             c.port,
		readTimeout:      c.readTimeout,
		writeTimeout:     c.writeTimeout,
		idleTimeout:      c.idleTimeout,
		gracefulShutdown: c.gracefulShutdown,
		shutdownTimeout:  c.shutdownTimeout,
		provider:         c.provider,
		middlewares:      make([]interface{}, len(c.middlewares)),
		observers:        make([]interfaces.ServerObserver, len(c.observers)),
		hooks:            make(map[interfaces.EventType][]interfaces.HookFunc),
		custom:           make(map[string]interface{}),
	}

	copy(clone.middlewares, c.middlewares)
	copy(clone.observers, c.observers)

	for eventType, hooks := range c.hooks {
		clone.hooks[eventType] = make([]interfaces.HookFunc, len(hooks))
		copy(clone.hooks[eventType], hooks)
	}

	for key, value := range c.custom {
		clone.custom[key] = value
	}

	return clone
}

// Option defines a functional option for configuring BaseConfig.
type Option func(*BaseConfig) error

// WithAddr sets the server address.
func WithAddr(addr string) Option {
	return func(c *BaseConfig) error {
		if addr == "" {
			return fmt.Errorf("address cannot be empty")
		}
		c.addr = addr
		return nil
	}
}

// WithPort sets the server port.
func WithPort(port int) Option {
	return func(c *BaseConfig) error {
		if port <= 0 || port > 65535 {
			return fmt.Errorf("invalid port: %d (must be between 1-65535)", port)
		}
		c.port = port
		return nil
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(timeout time.Duration) Option {
	return func(c *BaseConfig) error {
		if timeout <= 0 {
			return fmt.Errorf("read timeout must be positive")
		}
		c.readTimeout = timeout
		return nil
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *BaseConfig) error {
		if timeout <= 0 {
			return fmt.Errorf("write timeout must be positive")
		}
		c.writeTimeout = timeout
		return nil
	}
}

// WithIdleTimeout sets the idle timeout.
func WithIdleTimeout(timeout time.Duration) Option {
	return func(c *BaseConfig) error {
		if timeout <= 0 {
			return fmt.Errorf("idle timeout must be positive")
		}
		c.idleTimeout = timeout
		return nil
	}
}

// WithGracefulShutdown enables or disables graceful shutdown.
func WithGracefulShutdown(enabled bool) Option {
	return func(c *BaseConfig) error {
		c.gracefulShutdown = enabled
		return nil
	}
}

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(c *BaseConfig) error {
		if timeout <= 0 {
			return fmt.Errorf("shutdown timeout must be positive")
		}
		c.shutdownTimeout = timeout
		return nil
	}
}

// WithProvider sets the provider name.
func WithProvider(provider string) Option {
	return func(c *BaseConfig) error {
		if provider == "" {
			return fmt.Errorf("provider cannot be empty")
		}
		c.provider = provider
		return nil
	}
}

// WithMiddleware adds a middleware to the configuration.
func WithMiddleware(middleware interface{}) Option {
	return func(c *BaseConfig) error {
		if middleware == nil {
			return fmt.Errorf("middleware cannot be nil")
		}
		c.middlewares = append(c.middlewares, middleware)
		return nil
	}
}

// WithObserver adds an observer to the configuration.
func WithObserver(observer interfaces.ServerObserver) Option {
	return func(c *BaseConfig) error {
		if observer == nil {
			return fmt.Errorf("observer cannot be nil")
		}
		c.observers = append(c.observers, observer)
		return nil
	}
}

// WithHook adds a hook for the specified event type.
func WithHook(eventType interfaces.EventType, hook interfaces.HookFunc) Option {
	return func(c *BaseConfig) error {
		if hook == nil {
			return fmt.Errorf("hook cannot be nil")
		}
		if c.hooks[eventType] == nil {
			c.hooks[eventType] = make([]interfaces.HookFunc, 0)
		}
		c.hooks[eventType] = append(c.hooks[eventType], hook)
		return nil
	}
}

// WithCustom adds a custom configuration value.
func WithCustom(key string, value interface{}) Option {
	return func(c *BaseConfig) error {
		if key == "" {
			return fmt.Errorf("custom key cannot be empty")
		}
		c.custom[key] = value
		return nil
	}
}

// Builder provides a fluent interface for building configurations.
type Builder struct {
	config *BaseConfig
	errors []error
}

// NewBuilder creates a new configuration builder.
func NewBuilder() *Builder {
	return &Builder{
		config: NewBaseConfig(),
		errors: make([]error, 0),
	}
}

// Apply applies the given options to the configuration.
func (b *Builder) Apply(options ...Option) *Builder {
	for _, option := range options {
		if err := option(b.config); err != nil {
			b.errors = append(b.errors, err)
		}
	}
	return b
}

// Build builds the final configuration.
func (b *Builder) Build() (*BaseConfig, error) {
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("configuration errors: %v", b.errors)
	}

	if err := b.config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return b.config.Clone(), nil
}

// MustBuild builds the configuration and panics on error.
func (b *Builder) MustBuild() *BaseConfig {
	config, err := b.Build()
	if err != nil {
		panic(err)
	}
	return config
}
