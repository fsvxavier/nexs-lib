package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

// Config represents the unified configuration for all PostgreSQL drivers
type Config struct {
	// Connection settings
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string

	// Driver selection
	Driver interfaces.DriverType

	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Driver-specific settings
	MinConns int // for pgx

	// Advanced settings
	QueryTimeout      time.Duration
	ConnectTimeout    time.Duration
	TLSEnabled        bool
	QueryMode         string
	Timezone          string
	ApplicationName   string
	SearchPath        string
	StatementTimeout  time.Duration
	LockTimeout       time.Duration
	IdleInTransaction time.Duration

	// Observability
	TracingEnabled bool
	LoggingEnabled bool
	MetricsEnabled bool

	// Multi-tenant support
	MultiTenantEnabled bool
	RLSEnabled         bool
}

// ConfigOption represents a configuration option function
type ConfigOption func(*Config)

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefaultInt("POSTGRES_PORT", 5432),
		User:     getEnvOrDefault("POSTGRES_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", ""),
		Database: getEnvOrDefault("POSTGRES_DATABASE", "postgres"),
		SSLMode:  getEnvOrDefault("POSTGRES_SSL_MODE", "disable"),

		Driver: interfaces.DriverType(getEnvOrDefault("POSTGRES_DRIVER", string(interfaces.DriverPGX))),

		MaxOpenConns:    getEnvOrDefaultInt("POSTGRES_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvOrDefaultInt("POSTGRES_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvOrDefaultDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getEnvOrDefaultDuration("POSTGRES_CONN_MAX_IDLE_TIME", 5*time.Minute),
		MinConns:        getEnvOrDefaultInt("POSTGRES_MIN_CONNS", 2),

		QueryTimeout:      getEnvOrDefaultDuration("POSTGRES_QUERY_TIMEOUT", 30*time.Second),
		ConnectTimeout:    getEnvOrDefaultDuration("POSTGRES_CONNECT_TIMEOUT", 10*time.Second),
		TLSEnabled:        getEnvOrDefaultBool("POSTGRES_TLS_ENABLED", false),
		QueryMode:         getEnvOrDefault("POSTGRES_QUERY_MODE", "EXEC"),
		Timezone:          getEnvOrDefault("POSTGRES_TIMEZONE", "UTC"),
		ApplicationName:   getEnvOrDefault("POSTGRES_APPLICATION_NAME", "nexs-lib"),
		SearchPath:        getEnvOrDefault("POSTGRES_SEARCH_PATH", "public"),
		StatementTimeout:  getEnvOrDefaultDuration("POSTGRES_STATEMENT_TIMEOUT", 0),
		LockTimeout:       getEnvOrDefaultDuration("POSTGRES_LOCK_TIMEOUT", 0),
		IdleInTransaction: getEnvOrDefaultDuration("POSTGRES_IDLE_IN_TRANSACTION", 0),

		TracingEnabled: getEnvOrDefaultBool("POSTGRES_TRACING_ENABLED", false),
		LoggingEnabled: getEnvOrDefaultBool("POSTGRES_LOGGING_ENABLED", true),
		MetricsEnabled: getEnvOrDefaultBool("POSTGRES_METRICS_ENABLED", false),

		MultiTenantEnabled: getEnvOrDefaultBool("POSTGRES_MULTI_TENANT_ENABLED", false),
		RLSEnabled:         getEnvOrDefaultBool("POSTGRES_RLS_ENABLED", false),
	}
}

// NewConfig creates a new configuration with optional modifications
func NewConfig(options ...ConfigOption) *Config {
	cfg := DefaultConfig()
	for _, option := range options {
		option(cfg)
	}
	return cfg
}

// ConnectionString returns a formatted connection string for the given driver
func (c *Config) ConnectionString() string {
	switch c.Driver {
	case interfaces.DriverPGX, interfaces.DriverPQ:
		return c.buildPostgresConnectionString()
	case interfaces.DriverGORM:
		return c.buildGORMConnectionString()
	default:
		return c.buildPostgresConnectionString()
	}
}

// DSN returns the data source name
func (c *Config) DSN() string {
	return c.ConnectionString()
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.User == "" {
		return fmt.Errorf("user is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database is required")
	}
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be greater than 0")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns must be greater than or equal to 0")
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("max_idle_conns cannot be greater than max_open_conns")
	}
	return nil
}

// buildPostgresConnectionString builds a standard PostgreSQL connection string
func (c *Config) buildPostgresConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s application_name=%s search_path=%s timezone=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode, c.ApplicationName, c.SearchPath, c.Timezone,
	)
}

// buildGORMConnectionString builds a GORM-compatible connection string
func (c *Config) buildGORMConnectionString() string {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s application_name=%s search_path=%s",
		c.Host, c.User, c.Password, c.Database, c.Port, c.SSLMode, c.Timezone, c.ApplicationName, c.SearchPath,
	)

	if c.StatementTimeout > 0 {
		dsn += fmt.Sprintf(" statement_timeout=%dms", c.StatementTimeout.Milliseconds())
	}
	if c.LockTimeout > 0 {
		dsn += fmt.Sprintf(" lock_timeout=%dms", c.LockTimeout.Milliseconds())
	}
	if c.IdleInTransaction > 0 {
		dsn += fmt.Sprintf(" idle_in_transaction_session_timeout=%dms", c.IdleInTransaction.Milliseconds())
	}

	return dsn
}

// Configuration option functions

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

// WithUser sets the database user
func WithUser(user string) ConfigOption {
	return func(c *Config) {
		c.User = user
	}
}

// WithPassword sets the database password
func WithPassword(password string) ConfigOption {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDatabase sets the database name
func WithDatabase(database string) ConfigOption {
	return func(c *Config) {
		c.Database = database
	}
}

// WithSSLMode sets the SSL mode
func WithSSLMode(sslMode string) ConfigOption {
	return func(c *Config) {
		c.SSLMode = sslMode
	}
}

// WithDriver sets the database driver
func WithDriver(driver interfaces.DriverType) ConfigOption {
	return func(c *Config) {
		c.Driver = driver
	}
}

// WithMaxOpenConns sets the maximum number of open connections
func WithMaxOpenConns(maxConns int) ConfigOption {
	return func(c *Config) {
		c.MaxOpenConns = maxConns
	}
}

// WithMaxIdleConns sets the maximum number of idle connections
func WithMaxIdleConns(maxIdle int) ConfigOption {
	return func(c *Config) {
		c.MaxIdleConns = maxIdle
	}
}

// WithConnMaxLifetime sets the maximum connection lifetime
func WithConnMaxLifetime(lifetime time.Duration) ConfigOption {
	return func(c *Config) {
		c.ConnMaxLifetime = lifetime
	}
}

// WithConnMaxIdleTime sets the maximum connection idle time
func WithConnMaxIdleTime(idleTime time.Duration) ConfigOption {
	return func(c *Config) {
		c.ConnMaxIdleTime = idleTime
	}
}

// WithMinConns sets the minimum number of connections (pgx only)
func WithMinConns(minConns int) ConfigOption {
	return func(c *Config) {
		c.MinConns = minConns
	}
}

// WithQueryTimeout sets the query timeout
func WithQueryTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.QueryTimeout = timeout
	}
}

// WithConnectTimeout sets the connection timeout
func WithConnectTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.ConnectTimeout = timeout
	}
}

// WithTLSEnabled enables or disables TLS
func WithTLSEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.TLSEnabled = enabled
	}
}

// WithQueryMode sets the query execution mode
func WithQueryMode(mode string) ConfigOption {
	return func(c *Config) {
		c.QueryMode = mode
	}
}

// WithTimezone sets the timezone
func WithTimezone(timezone string) ConfigOption {
	return func(c *Config) {
		c.Timezone = timezone
	}
}

// WithApplicationName sets the application name
func WithApplicationName(name string) ConfigOption {
	return func(c *Config) {
		c.ApplicationName = name
	}
}

// WithTracingEnabled enables or disables tracing
func WithTracingEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.TracingEnabled = enabled
	}
}

// WithLoggingEnabled enables or disables logging
func WithLoggingEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.LoggingEnabled = enabled
	}
}

// WithMetricsEnabled enables or disables metrics
func WithMetricsEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.MetricsEnabled = enabled
	}
}

// WithMultiTenantEnabled enables or disables multi-tenant support
func WithMultiTenantEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.MultiTenantEnabled = enabled
	}
}

// WithRLSEnabled enables or disables Row Level Security
func WithRLSEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.RLSEnabled = enabled
	}
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvOrDefaultDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
