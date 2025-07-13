package goredis

import (
	"errors"
	"time"
)

// Error definitions
var (
	ErrNilGoRedisConfig = errors.New("redis config cannot be nil")
	ErrEmptyAddresses   = errors.New("no redis addresses provided")
)
var (
	ErrPoolClosed     = errors.New("redis connection pool is closed")
	ErrInvalidClient  = errors.New("invalid client type")
	ErrClientNotInUse = errors.New("client not in use or already released")
	ErrContextTimeout = errors.New("timeout acquiring redis connection")
)

// GoRedisConfig is a functional option type for configuring Redis connections
type GoRedisConfig func(*goRedisConfig)

// goRedisConfig holds Redis connection goRedisConfiguration
type goRedisConfig struct {
	// Addresses of Redis servers
	Addresses []string
	// Password for authentication
	Password string
	// Username for Redis 6+ ACL authentication
	Username string
	// Database index
	DB int
	// Connection pool size
	PoolSize int
	// Max retry attempts
	MaxRetries int
	// Minimum idle connections to maintain
	MinIdleConns int
	// Dial timeout
	DialTimeout time.Duration
	// Read timeout
	ReadTimeout time.Duration
	// Write timeout
	WriteTimeout time.Duration
	// Use TLS for secure connection
	UseTLS bool
	// Pool timeout
	PoolTimeout time.Duration
	// MaxIdleConns defines the maximum number of idle connections in the pool
	MaxIdleConns int
	// MaxActiveConns defines the maximum number of active connections in the pool
	MaxActiveConns int
	// TraceEnabled indicates if tracing is enabled for the client
	TraceEnabled bool
	// TraceServiceName is the name of the service for tracing
	TraceService string
}

// DefaultgoRedisConfig returns a default Redis goRedisConfiguration
func DefaultsGoRedisConfig(cfg *goRedisConfig) {
	cfg.Addresses = []string{"localhost:6379"}
	cfg.Password = ""
	cfg.Username = ""
	cfg.DB = 0
	cfg.PoolSize = 10
	cfg.MinIdleConns = 2
	cfg.MaxRetries = 3
	cfg.DialTimeout = 5 * time.Second
	cfg.ReadTimeout = 3 * time.Second
	cfg.WriteTimeout = 3 * time.Second
	cfg.PoolTimeout = 4 * time.Second
	cfg.UseTLS = false
	cfg.MaxIdleConns = 1000
	cfg.MaxActiveConns = 1000
	cfg.TraceEnabled = false
	cfg.TraceService = "redis-client"
}

// GetGoRedisConfig returns a new instance of goRedisConfig
func GetGoRedisConfig() goRedisConfig {
	return goRedisConfig{}
}

// Clone creates a deep copy of the goRedisConfiguration
func (c *goRedisConfig) Clone() *goRedisConfig {
	if c == nil {
		return nil
	}

	clone := *c

	// Deep copy of slices
	if c.Addresses != nil {
		clone.Addresses = make([]string, len(c.Addresses))
		copy(clone.Addresses, c.Addresses)
	}

	return &clone
}

// Validate checks if the goRedisConfiguration is valid
func (c *goRedisConfig) Validate() error {
	if c == nil {
		return ErrNilGoRedisConfig
	}

	if len(c.Addresses) == 0 {
		return ErrEmptyAddresses
	}

	return nil
}

// WithAddresses sets the Redis server addresses
func WithAddresses(addresses ...string) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.Addresses = addresses
	}
}

// WithPassword sets the Redis password
func WithPassword(password string) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.Password = password
	}
}

// WithUsername sets the Redis username (for Redis 6+ ACL)
func WithUsername(username string) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.Username = username
	}
}

// WithDB sets the Redis database index
func WithDB(db int) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.DB = db
	}
}

// WithPoolSize sets the connection pool size
func WithPoolSize(size int) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.PoolSize = size
	}
}

// WithMaxRetries sets the maximum retry attempts
func WithMaxRetries(retries int) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.MaxRetries = retries
	}
}

// WithMinIdleConns sets the minimum idle connections
func WithMinIdleConns(conns int) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.MinIdleConns = conns
	}
}

// WithDialTimeout sets the dial timeout
func WithDialTimeout(timeout time.Duration) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.DialTimeout = timeout
	}
}

// WithReadTimeout sets the read timeout
func WithReadTimeout(timeout time.Duration) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout
func WithWriteTimeout(timeout time.Duration) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.WriteTimeout = timeout
	}
}

// WithTLS enables or disables TLS
func WithTLS(useTLS bool) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.UseTLS = useTLS
	}
}

// WithPoolTimeout sets the pool timeout
func WithPoolTimeout(timeout time.Duration) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.PoolTimeout = timeout
	}
}

// WithMaxIdleConns sets the maximum number of idle connections in the pool
func WithMaxIdleConns(maxIdle int) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.MaxIdleConns = maxIdle
	}
}

// WithMaxActiveConns sets the maximum number of active connections in the pool
func WithMaxActiveConns(maxActive int) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.MaxActiveConns = maxActive
	}
}

// WithTraceEnabled enables or disables tracing for the client
func WithTraceEnabled(enabled bool) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.TraceEnabled = enabled
	}
}

// WithTraceService sets the service name for tracing
func WithTraceService(service string) GoRedisConfig {
	return func(c *goRedisConfig) {
		c.TraceService = service
	}
}
