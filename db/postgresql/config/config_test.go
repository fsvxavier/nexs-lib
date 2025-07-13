//go:build unit
// +build unit

package config

import (
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			expected: &Config{
				Host:               "localhost",
				Port:               5432,
				User:               "postgres",
				Password:           "",
				Database:           "postgres",
				SSLMode:            "disable",
				Driver:             interfaces.DriverPGX,
				MaxOpenConns:       25,
				MaxIdleConns:       5,
				ConnMaxLifetime:    5 * time.Minute,
				ConnMaxIdleTime:    5 * time.Minute,
				MinConns:           2,
				QueryTimeout:       30 * time.Second,
				ConnectTimeout:     10 * time.Second,
				TLSEnabled:         false,
				QueryMode:          "EXEC",
				Timezone:           "UTC",
				ApplicationName:    "nexs-lib",
				SearchPath:         "public",
				StatementTimeout:   0,
				LockTimeout:        0,
				IdleInTransaction:  0,
				TracingEnabled:     false,
				LoggingEnabled:     true,
				MetricsEnabled:     false,
				MultiTenantEnabled: false,
				RLSEnabled:         false,
			},
		},
		{
			name: "configuration with environment variables",
			envVars: map[string]string{
				"POSTGRES_HOST":                 "testhost",
				"POSTGRES_PORT":                 "5433",
				"POSTGRES_USER":                 "testuser",
				"POSTGRES_PASSWORD":             "testpass",
				"POSTGRES_DATABASE":             "testdb",
				"POSTGRES_SSL_MODE":             "require",
				"POSTGRES_DRIVER":               "gorm",
				"POSTGRES_MAX_OPEN_CONNS":       "50",
				"POSTGRES_MAX_IDLE_CONNS":       "10",
				"POSTGRES_CONN_MAX_LIFETIME":    "10m",
				"POSTGRES_CONN_MAX_IDLE_TIME":   "2m",
				"POSTGRES_MIN_CONNS":            "5",
				"POSTGRES_QUERY_TIMEOUT":        "45s",
				"POSTGRES_CONNECT_TIMEOUT":      "15s",
				"POSTGRES_TLS_ENABLED":          "true",
				"POSTGRES_QUERY_MODE":           "CACHE_STATEMENT",
				"POSTGRES_TIMEZONE":             "America/New_York",
				"POSTGRES_APPLICATION_NAME":     "test-app",
				"POSTGRES_SEARCH_PATH":          "test,public",
				"POSTGRES_TRACING_ENABLED":      "true",
				"POSTGRES_LOGGING_ENABLED":      "false",
				"POSTGRES_METRICS_ENABLED":      "true",
				"POSTGRES_MULTI_TENANT_ENABLED": "true",
				"POSTGRES_RLS_ENABLED":          "true",
			},
			expected: &Config{
				Host:               "testhost",
				Port:               5433,
				User:               "testuser",
				Password:           "testpass",
				Database:           "testdb",
				SSLMode:            "require",
				Driver:             interfaces.DriverGORM,
				MaxOpenConns:       50,
				MaxIdleConns:       10,
				ConnMaxLifetime:    10 * time.Minute,
				ConnMaxIdleTime:    2 * time.Minute,
				MinConns:           5,
				QueryTimeout:       45 * time.Second,
				ConnectTimeout:     15 * time.Second,
				TLSEnabled:         true,
				QueryMode:          "CACHE_STATEMENT",
				Timezone:           "America/New_York",
				ApplicationName:    "test-app",
				SearchPath:         "test,public",
				StatementTimeout:   0,
				LockTimeout:        0,
				IdleInTransaction:  0,
				TracingEnabled:     true,
				LoggingEnabled:     false,
				MetricsEnabled:     true,
				MultiTenantEnabled: true,
				RLSEnabled:         true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			cfg := DefaultConfig()

			// Compare all fields
			if cfg.Host != tt.expected.Host {
				t.Errorf("Host = %v, want %v", cfg.Host, tt.expected.Host)
			}
			if cfg.Port != tt.expected.Port {
				t.Errorf("Port = %v, want %v", cfg.Port, tt.expected.Port)
			}
			if cfg.User != tt.expected.User {
				t.Errorf("User = %v, want %v", cfg.User, tt.expected.User)
			}
			if cfg.Password != tt.expected.Password {
				t.Errorf("Password = %v, want %v", cfg.Password, tt.expected.Password)
			}
			if cfg.Database != tt.expected.Database {
				t.Errorf("Database = %v, want %v", cfg.Database, tt.expected.Database)
			}
			if cfg.SSLMode != tt.expected.SSLMode {
				t.Errorf("SSLMode = %v, want %v", cfg.SSLMode, tt.expected.SSLMode)
			}
			if cfg.Driver != tt.expected.Driver {
				t.Errorf("Driver = %v, want %v", cfg.Driver, tt.expected.Driver)
			}
			if cfg.MaxOpenConns != tt.expected.MaxOpenConns {
				t.Errorf("MaxOpenConns = %v, want %v", cfg.MaxOpenConns, tt.expected.MaxOpenConns)
			}
			if cfg.MaxIdleConns != tt.expected.MaxIdleConns {
				t.Errorf("MaxIdleConns = %v, want %v", cfg.MaxIdleConns, tt.expected.MaxIdleConns)
			}
			if cfg.ConnMaxLifetime != tt.expected.ConnMaxLifetime {
				t.Errorf("ConnMaxLifetime = %v, want %v", cfg.ConnMaxLifetime, tt.expected.ConnMaxLifetime)
			}
			if cfg.ConnMaxIdleTime != tt.expected.ConnMaxIdleTime {
				t.Errorf("ConnMaxIdleTime = %v, want %v", cfg.ConnMaxIdleTime, tt.expected.ConnMaxIdleTime)
			}
			if cfg.MinConns != tt.expected.MinConns {
				t.Errorf("MinConns = %v, want %v", cfg.MinConns, tt.expected.MinConns)
			}
			if cfg.QueryTimeout != tt.expected.QueryTimeout {
				t.Errorf("QueryTimeout = %v, want %v", cfg.QueryTimeout, tt.expected.QueryTimeout)
			}
			if cfg.ConnectTimeout != tt.expected.ConnectTimeout {
				t.Errorf("ConnectTimeout = %v, want %v", cfg.ConnectTimeout, tt.expected.ConnectTimeout)
			}
			if cfg.TLSEnabled != tt.expected.TLSEnabled {
				t.Errorf("TLSEnabled = %v, want %v", cfg.TLSEnabled, tt.expected.TLSEnabled)
			}
			if cfg.QueryMode != tt.expected.QueryMode {
				t.Errorf("QueryMode = %v, want %v", cfg.QueryMode, tt.expected.QueryMode)
			}
			if cfg.Timezone != tt.expected.Timezone {
				t.Errorf("Timezone = %v, want %v", cfg.Timezone, tt.expected.Timezone)
			}
			if cfg.ApplicationName != tt.expected.ApplicationName {
				t.Errorf("ApplicationName = %v, want %v", cfg.ApplicationName, tt.expected.ApplicationName)
			}
			if cfg.SearchPath != tt.expected.SearchPath {
				t.Errorf("SearchPath = %v, want %v", cfg.SearchPath, tt.expected.SearchPath)
			}
			if cfg.TracingEnabled != tt.expected.TracingEnabled {
				t.Errorf("TracingEnabled = %v, want %v", cfg.TracingEnabled, tt.expected.TracingEnabled)
			}
			if cfg.LoggingEnabled != tt.expected.LoggingEnabled {
				t.Errorf("LoggingEnabled = %v, want %v", cfg.LoggingEnabled, tt.expected.LoggingEnabled)
			}
			if cfg.MetricsEnabled != tt.expected.MetricsEnabled {
				t.Errorf("MetricsEnabled = %v, want %v", cfg.MetricsEnabled, tt.expected.MetricsEnabled)
			}
			if cfg.MultiTenantEnabled != tt.expected.MultiTenantEnabled {
				t.Errorf("MultiTenantEnabled = %v, want %v", cfg.MultiTenantEnabled, tt.expected.MultiTenantEnabled)
			}
			if cfg.RLSEnabled != tt.expected.RLSEnabled {
				t.Errorf("RLSEnabled = %v, want %v", cfg.RLSEnabled, tt.expected.RLSEnabled)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		options  []ConfigOption
		expected *Config
	}{
		{
			name:    "default configuration with no options",
			options: []ConfigOption{},
			expected: &Config{
				Host:               "localhost",
				Port:               5432,
				User:               "postgres",
				Password:           "",
				Database:           "postgres",
				SSLMode:            "disable",
				Driver:             interfaces.DriverPGX,
				MaxOpenConns:       25,
				MaxIdleConns:       5,
				ConnMaxLifetime:    5 * time.Minute,
				ConnMaxIdleTime:    5 * time.Minute,
				MinConns:           2,
				QueryTimeout:       30 * time.Second,
				ConnectTimeout:     10 * time.Second,
				TLSEnabled:         false,
				QueryMode:          "EXEC",
				Timezone:           "UTC",
				ApplicationName:    "nexs-lib",
				SearchPath:         "public",
				StatementTimeout:   0,
				LockTimeout:        0,
				IdleInTransaction:  0,
				TracingEnabled:     false,
				LoggingEnabled:     true,
				MetricsEnabled:     false,
				MultiTenantEnabled: false,
				RLSEnabled:         false,
			},
		},
		{
			name: "configuration with custom options",
			options: []ConfigOption{
				WithHost("custom-host"),
				WithPort(5433),
				WithUser("custom-user"),
				WithPassword("custom-pass"),
				WithDatabase("custom-db"),
				WithSSLMode("require"),
				WithDriver(interfaces.DriverGORM),
				WithMaxOpenConns(50),
				WithMaxIdleConns(10),
				WithConnMaxLifetime(10 * time.Minute),
				WithConnMaxIdleTime(2 * time.Minute),
				WithMinConns(5),
				WithQueryTimeout(45 * time.Second),
				WithConnectTimeout(15 * time.Second),
				WithTLSEnabled(true),
				WithQueryMode("CACHE_STATEMENT"),
				WithTimezone("America/New_York"),
				WithApplicationName("custom-app"),
				WithTracingEnabled(true),
				WithLoggingEnabled(false),
				WithMetricsEnabled(true),
				WithMultiTenantEnabled(true),
				WithRLSEnabled(true),
			},
			expected: &Config{
				Host:               "custom-host",
				Port:               5433,
				User:               "custom-user",
				Password:           "custom-pass",
				Database:           "custom-db",
				SSLMode:            "require",
				Driver:             interfaces.DriverGORM,
				MaxOpenConns:       50,
				MaxIdleConns:       10,
				ConnMaxLifetime:    10 * time.Minute,
				ConnMaxIdleTime:    2 * time.Minute,
				MinConns:           5,
				QueryTimeout:       45 * time.Second,
				ConnectTimeout:     15 * time.Second,
				TLSEnabled:         true,
				QueryMode:          "CACHE_STATEMENT",
				Timezone:           "America/New_York",
				ApplicationName:    "custom-app",
				SearchPath:         "public",
				StatementTimeout:   0,
				LockTimeout:        0,
				IdleInTransaction:  0,
				TracingEnabled:     true,
				LoggingEnabled:     false,
				MetricsEnabled:     true,
				MultiTenantEnabled: true,
				RLSEnabled:         true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.options...)

			// Compare key fields
			if cfg.Host != tt.expected.Host {
				t.Errorf("Host = %v, want %v", cfg.Host, tt.expected.Host)
			}
			if cfg.Port != tt.expected.Port {
				t.Errorf("Port = %v, want %v", cfg.Port, tt.expected.Port)
			}
			if cfg.User != tt.expected.User {
				t.Errorf("User = %v, want %v", cfg.User, tt.expected.User)
			}
			if cfg.Driver != tt.expected.Driver {
				t.Errorf("Driver = %v, want %v", cfg.Driver, tt.expected.Driver)
			}
			if cfg.MaxOpenConns != tt.expected.MaxOpenConns {
				t.Errorf("MaxOpenConns = %v, want %v", cfg.MaxOpenConns, tt.expected.MaxOpenConns)
			}
			if cfg.TracingEnabled != tt.expected.TracingEnabled {
				t.Errorf("TracingEnabled = %v, want %v", cfg.TracingEnabled, tt.expected.TracingEnabled)
			}
		})
	}
}

func TestConfigConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "pgx driver connection string",
			config: &Config{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Database:        "testdb",
				SSLMode:         "disable",
				Driver:          interfaces.DriverPGX,
				ApplicationName: "test-app",
				SearchPath:      "public",
				Timezone:        "UTC",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable application_name=test-app search_path=public timezone=UTC",
		},
		{
			name: "gorm driver connection string",
			config: &Config{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Database:        "testdb",
				SSLMode:         "disable",
				Driver:          interfaces.DriverGORM,
				ApplicationName: "test-app",
				SearchPath:      "public",
				Timezone:        "UTC",
			},
			expected: "host=localhost user=postgres password=password dbname=testdb port=5432 sslmode=disable TimeZone=UTC application_name=test-app search_path=public",
		},
		{
			name: "pq driver connection string",
			config: &Config{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Database:        "testdb",
				SSLMode:         "disable",
				Driver:          interfaces.DriverPQ,
				ApplicationName: "test-app",
				SearchPath:      "public",
				Timezone:        "UTC",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable application_name=test-app search_path=public timezone=UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ConnectionString()
			if result != tt.expected {
				t.Errorf("ConnectionString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid configuration",
			config: &Config{
				Host:         "localhost",
				Port:         5432,
				User:         "postgres",
				Database:     "testdb",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantError: false,
		},
		{
			name: "missing host",
			config: &Config{
				Host:         "",
				Port:         5432,
				User:         "postgres",
				Database:     "testdb",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantError: true,
			errorMsg:  "host is required",
		},
		{
			name: "invalid port",
			config: &Config{
				Host:         "localhost",
				Port:         0,
				User:         "postgres",
				Database:     "testdb",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantError: true,
			errorMsg:  "port must be between 1 and 65535",
		},
		{
			name: "missing user",
			config: &Config{
				Host:         "localhost",
				Port:         5432,
				User:         "",
				Database:     "testdb",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantError: true,
			errorMsg:  "user is required",
		},
		{
			name: "missing database",
			config: &Config{
				Host:         "localhost",
				Port:         5432,
				User:         "postgres",
				Database:     "",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantError: true,
			errorMsg:  "database is required",
		},
		{
			name: "invalid max open conns",
			config: &Config{
				Host:         "localhost",
				Port:         5432,
				User:         "postgres",
				Database:     "testdb",
				MaxOpenConns: 0,
				MaxIdleConns: 5,
			},
			wantError: true,
			errorMsg:  "max_open_conns must be greater than 0",
		},
		{
			name: "invalid max idle conns",
			config: &Config{
				Host:         "localhost",
				Port:         5432,
				User:         "postgres",
				Database:     "testdb",
				MaxOpenConns: 25,
				MaxIdleConns: -1,
			},
			wantError: true,
			errorMsg:  "max_idle_conns must be greater than or equal to 0",
		},
		{
			name: "max idle greater than max open",
			config: &Config{
				Host:         "localhost",
				Port:         5432,
				User:         "postgres",
				Database:     "testdb",
				MaxOpenConns: 5,
				MaxIdleConns: 10,
			},
			wantError: true,
			errorMsg:  "max_idle_conns cannot be greater than max_open_conns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Validate() error = nil, wantError %v", tt.wantError)
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				}
			}
		})
	}
}

func TestConfigDSN(t *testing.T) {
	config := &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		Database:        "testdb",
		SSLMode:         "disable",
		Driver:          interfaces.DriverPGX,
		ApplicationName: "test-app",
		SearchPath:      "public",
		Timezone:        "UTC",
	}

	dsn := config.DSN()
	connectionString := config.ConnectionString()

	if dsn != connectionString {
		t.Errorf("DSN() = %v, want %v", dsn, connectionString)
	}
}

// Benchmark tests
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

func BenchmarkNewConfigWithOptions(b *testing.B) {
	options := []ConfigOption{
		WithHost("localhost"),
		WithPort(5432),
		WithUser("postgres"),
		WithDatabase("testdb"),
		WithMaxOpenConns(50),
	}

	for i := 0; i < b.N; i++ {
		_ = NewConfig(options...)
	}
}

func BenchmarkConnectionString(b *testing.B) {
	config := DefaultConfig()

	for i := 0; i < b.N; i++ {
		_ = config.ConnectionString()
	}
}

func BenchmarkValidate(b *testing.B) {
	config := DefaultConfig()

	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}
