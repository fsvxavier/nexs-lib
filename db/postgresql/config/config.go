package config

import (
	"crypto/tls"
	"fmt"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// DefaultConfig provides default configuration values
type DefaultConfig struct {
	connectionString   string
	poolConfig         interfaces.PoolConfig
	tlsConfig          interfaces.TLSConfig
	retryConfig        interfaces.RetryConfig
	hookConfig         interfaces.HookConfig
	multiTenantEnabled bool
	readReplicaConfig  interfaces.ReadReplicaConfig
	failoverConfig     interfaces.FailoverConfig
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig(connectionString string) *DefaultConfig {
	return &DefaultConfig{
		connectionString: connectionString,
		poolConfig: interfaces.PoolConfig{
			MaxConns:          40,
			MinConns:          2,
			MaxConnLifetime:   time.Minute * 30,
			MaxConnIdleTime:   time.Minute * 5,
			HealthCheckPeriod: time.Minute * 1,
			ConnectTimeout:    time.Second * 30,
			LazyConnect:       false,
		},
		tlsConfig: interfaces.TLSConfig{
			Enabled:            false,
			InsecureSkipVerify: false,
		},
		retryConfig: interfaces.RetryConfig{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			MaxInterval:     time.Second * 5,
			Multiplier:      2.0,
			RandomizeWait:   true,
		},
		hookConfig: interfaces.HookConfig{
			EnabledHooks: []interfaces.HookType{},
			CustomHooks:  make(map[string]interfaces.HookType),
			HookTimeout:  time.Second * 10,
		},
		multiTenantEnabled: false,
		readReplicaConfig: interfaces.ReadReplicaConfig{
			Enabled:             false,
			ConnectionStrings:   []string{},
			LoadBalanceMode:     interfaces.LoadBalanceModeRoundRobin,
			HealthCheckInterval: time.Minute * 1,
		},
		failoverConfig: interfaces.FailoverConfig{
			Enabled:             false,
			FallbackNodes:       []string{},
			HealthCheckInterval: time.Second * 30,
			RetryInterval:       time.Second * 5,
			MaxFailoverAttempts: 3,
		},
	}
}

// GetConnectionString returns the connection string
func (c *DefaultConfig) GetConnectionString() string {
	return c.connectionString
}

// GetPoolConfig returns the pool configuration
func (c *DefaultConfig) GetPoolConfig() interfaces.PoolConfig {
	return c.poolConfig
}

// GetTLSConfig returns the TLS configuration
func (c *DefaultConfig) GetTLSConfig() interfaces.TLSConfig {
	return c.tlsConfig
}

// GetRetryConfig returns the retry configuration
func (c *DefaultConfig) GetRetryConfig() interfaces.RetryConfig {
	return c.retryConfig
}

// GetHookConfig returns the hook configuration
func (c *DefaultConfig) GetHookConfig() interfaces.HookConfig {
	return c.hookConfig
}

// IsMultiTenantEnabled returns whether multi-tenancy is enabled
func (c *DefaultConfig) IsMultiTenantEnabled() bool {
	return c.multiTenantEnabled
}

// GetReadReplicaConfig returns the read replica configuration
func (c *DefaultConfig) GetReadReplicaConfig() interfaces.ReadReplicaConfig {
	return c.readReplicaConfig
}

// GetFailoverConfig returns the failover configuration
func (c *DefaultConfig) GetFailoverConfig() interfaces.FailoverConfig {
	return c.failoverConfig
}

// Validate validates the configuration
func (c *DefaultConfig) Validate() error {
	if c.connectionString == "" {
		return fmt.Errorf("connection string cannot be empty")
	}

	if c.poolConfig.MaxConns <= 0 {
		return fmt.Errorf("max connections must be positive")
	}

	if c.poolConfig.MinConns < 0 {
		return fmt.Errorf("min connections cannot be negative")
	}

	if c.poolConfig.MinConns > c.poolConfig.MaxConns {
		return fmt.Errorf("min connections cannot be greater than max connections")
	}

	if c.poolConfig.MaxConnLifetime <= 0 {
		return fmt.Errorf("max connection lifetime must be positive")
	}

	if c.poolConfig.MaxConnIdleTime <= 0 {
		return fmt.Errorf("max connection idle time must be positive")
	}

	if c.poolConfig.ConnectTimeout <= 0 {
		return fmt.Errorf("connect timeout must be positive")
	}

	if c.retryConfig.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if c.retryConfig.InitialInterval <= 0 {
		return fmt.Errorf("initial interval must be positive")
	}

	if c.retryConfig.MaxInterval <= 0 {
		return fmt.Errorf("max interval must be positive")
	}

	if c.retryConfig.Multiplier <= 1.0 {
		return fmt.Errorf("multiplier must be greater than 1.0")
	}

	return nil
}

// ConfigOption represents a configuration option function
type ConfigOption func(*DefaultConfig) error

// WithConnectionString sets the connection string
func WithConnectionString(connectionString string) ConfigOption {
	return func(c *DefaultConfig) error {
		if connectionString == "" {
			return fmt.Errorf("connection string cannot be empty")
		}
		c.connectionString = connectionString
		return nil
	}
}

// WithPoolConfig sets the pool configuration
func WithPoolConfig(poolConfig interfaces.PoolConfig) ConfigOption {
	return func(c *DefaultConfig) error {
		c.poolConfig = poolConfig
		return nil
	}
}

// WithMaxConns sets the maximum number of connections
func WithMaxConns(maxConns int32) ConfigOption {
	return func(c *DefaultConfig) error {
		if maxConns <= 0 {
			return fmt.Errorf("max connections must be positive")
		}
		c.poolConfig.MaxConns = maxConns
		return nil
	}
}

// WithMinConns sets the minimum number of connections
func WithMinConns(minConns int32) ConfigOption {
	return func(c *DefaultConfig) error {
		if minConns < 0 {
			return fmt.Errorf("min connections cannot be negative")
		}
		c.poolConfig.MinConns = minConns
		return nil
	}
}

// WithMaxConnLifetime sets the maximum connection lifetime
func WithMaxConnLifetime(maxConnLifetime time.Duration) ConfigOption {
	return func(c *DefaultConfig) error {
		if maxConnLifetime <= 0 {
			return fmt.Errorf("max connection lifetime must be positive")
		}
		c.poolConfig.MaxConnLifetime = maxConnLifetime
		return nil
	}
}

// WithMaxConnIdleTime sets the maximum connection idle time
func WithMaxConnIdleTime(maxConnIdleTime time.Duration) ConfigOption {
	return func(c *DefaultConfig) error {
		if maxConnIdleTime <= 0 {
			return fmt.Errorf("max connection idle time must be positive")
		}
		c.poolConfig.MaxConnIdleTime = maxConnIdleTime
		return nil
	}
}

// WithHealthCheckPeriod sets the health check period
func WithHealthCheckPeriod(healthCheckPeriod time.Duration) ConfigOption {
	return func(c *DefaultConfig) error {
		if healthCheckPeriod <= 0 {
			return fmt.Errorf("health check period must be positive")
		}
		c.poolConfig.HealthCheckPeriod = healthCheckPeriod
		return nil
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(connectTimeout time.Duration) ConfigOption {
	return func(c *DefaultConfig) error {
		if connectTimeout <= 0 {
			return fmt.Errorf("connect timeout must be positive")
		}
		c.poolConfig.ConnectTimeout = connectTimeout
		return nil
	}
}

// WithLazyConnect sets lazy connection mode
func WithLazyConnect(lazyConnect bool) ConfigOption {
	return func(c *DefaultConfig) error {
		c.poolConfig.LazyConnect = lazyConnect
		return nil
	}
}

// WithTLS enables TLS configuration
func WithTLS(enabled bool, config *tls.Config) ConfigOption {
	return func(c *DefaultConfig) error {
		c.tlsConfig.Enabled = enabled
		if config != nil {
			c.tlsConfig.InsecureSkipVerify = config.InsecureSkipVerify
			c.tlsConfig.ServerName = config.ServerName
		}
		return nil
	}
}

// WithTLSFiles sets TLS configuration using certificate files
func WithTLSFiles(certFile, keyFile, caFile string) ConfigOption {
	return func(c *DefaultConfig) error {
		c.tlsConfig.Enabled = true
		c.tlsConfig.CertFile = certFile
		c.tlsConfig.KeyFile = keyFile
		c.tlsConfig.CAFile = caFile
		return nil
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(retryConfig interfaces.RetryConfig) ConfigOption {
	return func(c *DefaultConfig) error {
		c.retryConfig = retryConfig
		return nil
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(maxRetries int) ConfigOption {
	return func(c *DefaultConfig) error {
		if maxRetries < 0 {
			return fmt.Errorf("max retries cannot be negative")
		}
		c.retryConfig.MaxRetries = maxRetries
		return nil
	}
}

// WithHookConfig sets the hook configuration
func WithHookConfig(hookConfig interfaces.HookConfig) ConfigOption {
	return func(c *DefaultConfig) error {
		c.hookConfig = hookConfig
		return nil
	}
}

// WithEnabledHooks sets the enabled hooks
func WithEnabledHooks(hooks ...interfaces.HookType) ConfigOption {
	return func(c *DefaultConfig) error {
		c.hookConfig.EnabledHooks = hooks
		return nil
	}
}

// WithCustomHook adds a custom hook
func WithCustomHook(name string, hookType interfaces.HookType) ConfigOption {
	return func(c *DefaultConfig) error {
		if c.hookConfig.CustomHooks == nil {
			c.hookConfig.CustomHooks = make(map[string]interfaces.HookType)
		}
		c.hookConfig.CustomHooks[name] = hookType
		return nil
	}
}

// WithMultiTenant enables multi-tenancy
func WithMultiTenant(enabled bool) ConfigOption {
	return func(c *DefaultConfig) error {
		c.multiTenantEnabled = enabled
		return nil
	}
}

// WithReadReplicas sets read replica configuration
func WithReadReplicas(connectionStrings []string, loadBalanceMode interfaces.LoadBalanceMode) ConfigOption {
	return func(c *DefaultConfig) error {
		c.readReplicaConfig.Enabled = len(connectionStrings) > 0
		c.readReplicaConfig.ConnectionStrings = connectionStrings
		c.readReplicaConfig.LoadBalanceMode = loadBalanceMode
		return nil
	}
}

// WithFailover sets failover configuration
func WithFailover(fallbackNodes []string, maxAttempts int) ConfigOption {
	return func(c *DefaultConfig) error {
		c.failoverConfig.Enabled = len(fallbackNodes) > 0
		c.failoverConfig.FallbackNodes = fallbackNodes
		c.failoverConfig.MaxFailoverAttempts = maxAttempts
		return nil
	}
}

// ApplyOptions applies configuration options to the config
func (c *DefaultConfig) ApplyOptions(options ...ConfigOption) error {
	for _, option := range options {
		if err := option(c); err != nil {
			return fmt.Errorf("failed to apply config option: %w", err)
		}
	}
	return nil
}

// Clone creates a deep copy of the configuration
func (c *DefaultConfig) Clone() *DefaultConfig {
	clone := &DefaultConfig{
		connectionString:   c.connectionString,
		poolConfig:         c.poolConfig,
		tlsConfig:          c.tlsConfig,
		retryConfig:        c.retryConfig,
		hookConfig:         c.hookConfig,
		multiTenantEnabled: c.multiTenantEnabled,
		readReplicaConfig:  c.readReplicaConfig,
		failoverConfig:     c.failoverConfig,
	}

	// Deep copy maps and slices
	if c.hookConfig.CustomHooks != nil {
		clone.hookConfig.CustomHooks = make(map[string]interfaces.HookType)
		for k, v := range c.hookConfig.CustomHooks {
			clone.hookConfig.CustomHooks[k] = v
		}
	}

	if c.hookConfig.EnabledHooks != nil {
		clone.hookConfig.EnabledHooks = make([]interfaces.HookType, len(c.hookConfig.EnabledHooks))
		copy(clone.hookConfig.EnabledHooks, c.hookConfig.EnabledHooks)
	}

	if c.readReplicaConfig.ConnectionStrings != nil {
		clone.readReplicaConfig.ConnectionStrings = make([]string, len(c.readReplicaConfig.ConnectionStrings))
		copy(clone.readReplicaConfig.ConnectionStrings, c.readReplicaConfig.ConnectionStrings)
	}

	if c.failoverConfig.FallbackNodes != nil {
		clone.failoverConfig.FallbackNodes = make([]string, len(c.failoverConfig.FallbackNodes))
		copy(clone.failoverConfig.FallbackNodes, c.failoverConfig.FallbackNodes)
	}

	return clone
}
