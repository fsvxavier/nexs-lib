package config

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"
)

// Config represents the main configuration for PostgreSQL connections
type Config struct {
	// Connection settings
	Host     string
	Port     int
	Database string
	Username string
	Password string

	// Connection string (alternative to individual fields)
	ConnString string

	// Pool settings
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration

	// Connection timeouts
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration

	// TLS configuration
	TLSConfig *tls.Config
	TLSMode   TLSMode

	// Application settings
	ApplicationName string
	SearchPath      []string
	Timezone        string

	// Advanced settings
	QueryExecMode QueryExecMode
	RuntimeParams map[string]string

	// Multi-tenancy
	MultiTenantEnabled bool
	DefaultSchema      string

	// Retry and failover
	RetryConfig    *RetryConfig
	FailoverConfig *FailoverConfig

	// Read replicas
	ReadReplicas []ReplicaConfig

	// Monitoring and observability
	TracingEnabled     bool
	MetricsEnabled     bool
	LogLevel           LogLevel
	SlowQueryThreshold time.Duration

	// Health check
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration

	// Hooks and middleware
	Hooks *HooksConfig

	// Provider specific settings
	ProviderSpecific map[string]interface{}
}

// TLSMode represents TLS connection modes
type TLSMode string

const (
	TLSModeDisable    TLSMode = "disable"
	TLSModeAllow      TLSMode = "allow"
	TLSModePrefer     TLSMode = "prefer"
	TLSModeRequire    TLSMode = "require"
	TLSModeVerifyCA   TLSMode = "verify-ca"
	TLSModeVerifyFull TLSMode = "verify-full"
)

// QueryExecMode represents query execution modes
type QueryExecMode string

const (
	QueryExecModeDefault        QueryExecMode = "default"
	QueryExecModeCacheStatement QueryExecMode = "cache_statement"
	QueryExecModeCacheDescribe  QueryExecMode = "cache_describe"
	QueryExecModeDescribeExec   QueryExecMode = "describe_exec"
	QueryExecModeExec           QueryExecMode = "exec"
	QueryExecModeSimpleProtocol QueryExecMode = "simple_protocol"
)

// LogLevel represents logging levels
type LogLevel string

const (
	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelNone  LogLevel = "none"
)

// RetryConfig represents retry configuration
type RetryConfig struct {
	Enabled     bool
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
	Jitter      bool
}

// FailoverConfig represents failover configuration
type FailoverConfig struct {
	Enabled               bool
	MaxFailoverTime       time.Duration
	FailoverCheckInterval time.Duration
	AutoFailback          bool
	FailbackDelay         time.Duration
}

// ReplicaConfig represents read replica configuration
type ReplicaConfig struct {
	Host            string
	Port            int
	Weight          int
	MaxConns        int32
	HealthCheckURL  string
	PreferredMaster bool
}

// HooksConfig represents hooks configuration
type HooksConfig struct {
	BeforeConnect     func(ctx context.Context, cfg *Config) error
	AfterConnect      func(ctx context.Context, conn interface{}) error
	BeforeQuery       func(ctx context.Context, query string, args []interface{}) error
	AfterQuery        func(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) error
	BeforeTransaction func(ctx context.Context) error
	AfterTransaction  func(ctx context.Context, committed bool, duration time.Duration, err error) error
	BeforeRelease     func(ctx context.Context, conn interface{}) error
	AfterAcquire      func(ctx context.Context, conn interface{}) error
}

// ConfigOption represents a configuration option function
type ConfigOption func(*Config)

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		Username: "postgres",
		Password: "",

		MaxConns:        40,
		MinConns:        2,
		MaxConnLifetime: 9 * time.Second,
		MaxConnIdleTime: 3 * time.Second,

		ConnectTimeout: 30 * time.Second,
		QueryTimeout:   30 * time.Second,

		TLSMode:         TLSModePrefer,
		ApplicationName: "nexs-lib",
		Timezone:        "UTC",
		QueryExecMode:   QueryExecModeDefault,

		MultiTenantEnabled: false,
		DefaultSchema:      "public",

		RetryConfig: &RetryConfig{
			Enabled:     true,
			MaxRetries:  3,
			InitialWait: 100 * time.Millisecond,
			MaxWait:     2 * time.Second,
			Multiplier:  2.0,
			Jitter:      true,
		},

		FailoverConfig: &FailoverConfig{
			Enabled:               false,
			MaxFailoverTime:       30 * time.Second,
			FailoverCheckInterval: 5 * time.Second,
			AutoFailback:          true,
			FailbackDelay:         10 * time.Second,
		},

		TracingEnabled:     false,
		MetricsEnabled:     false,
		LogLevel:           LogLevelInfo,
		SlowQueryThreshold: 100 * time.Millisecond,

		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,

		RuntimeParams:    make(map[string]string),
		ProviderSpecific: make(map[string]interface{}),
	}
}

// WithHost sets the database host
func WithHost(host string) ConfigOption {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort sets the database port
func WithPort(port int) ConfigOption {
	return func(c *Config) {
		c.Port = port
	}
}

// WithDatabase sets the database name
func WithDatabase(database string) ConfigOption {
	return func(c *Config) {
		c.Database = database
	}
}

// WithUsername sets the database username
func WithUsername(username string) ConfigOption {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword sets the database password
func WithPassword(password string) ConfigOption {
	return func(c *Config) {
		c.Password = password
	}
}

// WithConnectionString sets the connection string
func WithConnectionString(connString string) ConfigOption {
	return func(c *Config) {
		c.ConnString = connString
	}
}

// WithMaxConns sets the maximum number of connections
func WithMaxConns(maxConns int32) ConfigOption {
	return func(c *Config) {
		c.MaxConns = maxConns
	}
}

// WithMinConns sets the minimum number of connections
func WithMinConns(minConns int32) ConfigOption {
	return func(c *Config) {
		c.MinConns = minConns
	}
}

// WithMaxConnLifetime sets the maximum connection lifetime
func WithMaxConnLifetime(lifetime time.Duration) ConfigOption {
	return func(c *Config) {
		c.MaxConnLifetime = lifetime
	}
}

// WithMaxConnIdleTime sets the maximum connection idle time
func WithMaxConnIdleTime(idleTime time.Duration) ConfigOption {
	return func(c *Config) {
		c.MaxConnIdleTime = idleTime
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.ConnectTimeout = timeout
	}
}

// WithQueryTimeout sets the query timeout
func WithQueryTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.QueryTimeout = timeout
	}
}

// WithTLSConfig sets the TLS configuration
func WithTLSConfig(tlsConfig *tls.Config) ConfigOption {
	return func(c *Config) {
		c.TLSConfig = tlsConfig
	}
}

// WithTLSMode sets the TLS mode
func WithTLSMode(mode TLSMode) ConfigOption {
	return func(c *Config) {
		c.TLSMode = mode
	}
}

// WithApplicationName sets the application name
func WithApplicationName(name string) ConfigOption {
	return func(c *Config) {
		c.ApplicationName = name
	}
}

// WithSearchPath sets the search path
func WithSearchPath(searchPath []string) ConfigOption {
	return func(c *Config) {
		c.SearchPath = searchPath
	}
}

// WithTimezone sets the timezone
func WithTimezone(timezone string) ConfigOption {
	return func(c *Config) {
		c.Timezone = timezone
	}
}

// WithQueryExecMode sets the query execution mode
func WithQueryExecMode(mode QueryExecMode) ConfigOption {
	return func(c *Config) {
		c.QueryExecMode = mode
	}
}

// WithMultiTenant enables multi-tenant mode
func WithMultiTenant(enabled bool) ConfigOption {
	return func(c *Config) {
		c.MultiTenantEnabled = enabled
	}
}

// WithDefaultSchema sets the default schema
func WithDefaultSchema(schema string) ConfigOption {
	return func(c *Config) {
		c.DefaultSchema = schema
	}
}

// WithRetryConfig sets the retry configuration
func WithRetryConfig(retryConfig *RetryConfig) ConfigOption {
	return func(c *Config) {
		c.RetryConfig = retryConfig
	}
}

// WithFailoverConfig sets the failover configuration
func WithFailoverConfig(failoverConfig *FailoverConfig) ConfigOption {
	return func(c *Config) {
		c.FailoverConfig = failoverConfig
	}
}

// WithReadReplicas sets the read replicas configuration
func WithReadReplicas(replicas []ReplicaConfig) ConfigOption {
	return func(c *Config) {
		c.ReadReplicas = replicas
	}
}

// WithTracing enables tracing
func WithTracing(enabled bool) ConfigOption {
	return func(c *Config) {
		c.TracingEnabled = enabled
	}
}

// WithMetrics enables metrics
func WithMetrics(enabled bool) ConfigOption {
	return func(c *Config) {
		c.MetricsEnabled = enabled
	}
}

// WithLogLevel sets the log level
func WithLogLevel(level LogLevel) ConfigOption {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// WithSlowQueryThreshold sets the slow query threshold
func WithSlowQueryThreshold(threshold time.Duration) ConfigOption {
	return func(c *Config) {
		c.SlowQueryThreshold = threshold
	}
}

// WithHealthCheckInterval sets the health check interval
func WithHealthCheckInterval(interval time.Duration) ConfigOption {
	return func(c *Config) {
		c.HealthCheckInterval = interval
	}
}

// WithHealthCheckTimeout sets the health check timeout
func WithHealthCheckTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.HealthCheckTimeout = timeout
	}
}

// WithHooks sets the hooks configuration
func WithHooks(hooks *HooksConfig) ConfigOption {
	return func(c *Config) {
		c.Hooks = hooks
	}
}

// WithRuntimeParam sets a runtime parameter
func WithRuntimeParam(key, value string) ConfigOption {
	return func(c *Config) {
		if c.RuntimeParams == nil {
			c.RuntimeParams = make(map[string]string)
		}
		c.RuntimeParams[key] = value
	}
}

// WithProviderSpecific sets provider-specific configuration
func WithProviderSpecific(key string, value interface{}) ConfigOption {
	return func(c *Config) {
		if c.ProviderSpecific == nil {
			c.ProviderSpecific = make(map[string]interface{})
		}
		c.ProviderSpecific[key] = value
	}
}

// NewConfig creates a new configuration with the provided options
func NewConfig(options ...ConfigOption) *Config {
	config := DefaultConfig()
	for _, option := range options {
		option(config)
	}
	return config
}

// ConnectionString builds a connection string from the configuration
func (c *Config) ConnectionString() string {
	if c.ConnString != "" {
		return c.ConnString
	}

	host := c.Host
	if c.Port != 0 && c.Port != 5432 {
		host = net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	}

	connStr := "postgres://"
	if c.Username != "" {
		connStr += c.Username
		if c.Password != "" {
			connStr += ":" + c.Password
		}
		connStr += "@"
	}

	connStr += host
	if c.Database != "" {
		connStr += "/" + c.Database
	}

	params := make(map[string]string)
	if c.TLSMode != "" {
		params["sslmode"] = string(c.TLSMode)
	}
	if c.ApplicationName != "" {
		params["application_name"] = c.ApplicationName
	}
	if c.Timezone != "" {
		params["timezone"] = c.Timezone
	}
	if c.ConnectTimeout > 0 {
		params["connect_timeout"] = strconv.Itoa(int(c.ConnectTimeout.Seconds()))
	}

	// Add runtime parameters
	for key, value := range c.RuntimeParams {
		params[key] = value
	}

	if len(params) > 0 {
		connStr += "?"
		first := true
		for key, value := range params {
			if !first {
				connStr += "&"
			}
			connStr += key + "=" + value
			first = false
		}
	}

	return connStr
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ConnString == "" && c.Host == "" {
		return ErrInvalidConfig{Field: "host", Message: "host is required when connection string is not provided"}
	}

	if c.MaxConns <= 0 {
		return ErrInvalidConfig{Field: "max_conns", Message: "max_conns must be greater than 0"}
	}

	if c.MinConns < 0 {
		return ErrInvalidConfig{Field: "min_conns", Message: "min_conns must be greater than or equal to 0"}
	}

	if c.MinConns > c.MaxConns {
		return ErrInvalidConfig{Field: "min_conns", Message: "min_conns cannot be greater than max_conns"}
	}

	if c.MaxConnLifetime < 0 {
		return ErrInvalidConfig{Field: "max_conn_lifetime", Message: "max_conn_lifetime must be greater than or equal to 0"}
	}

	if c.MaxConnIdleTime < 0 {
		return ErrInvalidConfig{Field: "max_conn_idle_time", Message: "max_conn_idle_time must be greater than or equal to 0"}
	}

	if c.ConnectTimeout < 0 {
		return ErrInvalidConfig{Field: "connect_timeout", Message: "connect_timeout must be greater than or equal to 0"}
	}

	if c.QueryTimeout < 0 {
		return ErrInvalidConfig{Field: "query_timeout", Message: "query_timeout must be greater than or equal to 0"}
	}

	return nil
}

// ErrInvalidConfig represents a configuration validation error
type ErrInvalidConfig struct {
	Field   string
	Message string
}

func (e ErrInvalidConfig) Error() string {
	return "invalid config field '" + e.Field + "': " + e.Message
}
