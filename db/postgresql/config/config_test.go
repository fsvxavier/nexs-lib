package config

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "localhost" {
		t.Errorf("Expected host to be 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 5432 {
		t.Errorf("Expected port to be 5432, got %d", cfg.Port)
	}

	if cfg.Database != "postgres" {
		t.Errorf("Expected database to be 'postgres', got '%s'", cfg.Database)
	}

	if cfg.Username != "postgres" {
		t.Errorf("Expected username to be 'postgres', got '%s'", cfg.Username)
	}

	if cfg.MaxConns != 40 {
		t.Errorf("Expected MaxConns to be 40, got %d", cfg.MaxConns)
	}

	if cfg.MinConns != 2 {
		t.Errorf("Expected MinConns to be 2, got %d", cfg.MinConns)
	}

	if cfg.MaxConnLifetime != 9*time.Second {
		t.Errorf("Expected MaxConnLifetime to be 9s, got %v", cfg.MaxConnLifetime)
	}

	if cfg.MaxConnIdleTime != 3*time.Second {
		t.Errorf("Expected MaxConnIdleTime to be 3s, got %v", cfg.MaxConnIdleTime)
	}

	if cfg.TLSMode != TLSModePrefer {
		t.Errorf("Expected TLSMode to be 'prefer', got '%s'", cfg.TLSMode)
	}

	if cfg.ApplicationName != "nexs-lib" {
		t.Errorf("Expected ApplicationName to be 'nexs-lib', got '%s'", cfg.ApplicationName)
	}

	if cfg.Timezone != "UTC" {
		t.Errorf("Expected Timezone to be 'UTC', got '%s'", cfg.Timezone)
	}

	if cfg.RetryConfig == nil {
		t.Error("Expected RetryConfig to be set")
	} else {
		if !cfg.RetryConfig.Enabled {
			t.Error("Expected RetryConfig.Enabled to be true")
		}
		if cfg.RetryConfig.MaxRetries != 3 {
			t.Errorf("Expected RetryConfig.MaxRetries to be 3, got %d", cfg.RetryConfig.MaxRetries)
		}
	}

	if cfg.FailoverConfig == nil {
		t.Error("Expected FailoverConfig to be set")
	} else {
		if cfg.FailoverConfig.Enabled {
			t.Error("Expected FailoverConfig.Enabled to be false")
		}
	}

	if cfg.RuntimeParams == nil {
		t.Error("Expected RuntimeParams to be initialized")
	}

	if cfg.ProviderSpecific == nil {
		t.Error("Expected ProviderSpecific to be initialized")
	}
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig(
		WithHost("test-host"),
		WithPort(5433),
		WithDatabase("test-db"),
		WithUsername("test-user"),
		WithPassword("test-pass"),
		WithMaxConns(20),
		WithMinConns(5),
		WithApplicationName("test-app"),
		WithTLSMode(TLSModeRequire),
	)

	if cfg.Host != "test-host" {
		t.Errorf("Expected host to be 'test-host', got '%s'", cfg.Host)
	}

	if cfg.Port != 5433 {
		t.Errorf("Expected port to be 5433, got %d", cfg.Port)
	}

	if cfg.Database != "test-db" {
		t.Errorf("Expected database to be 'test-db', got '%s'", cfg.Database)
	}

	if cfg.Username != "test-user" {
		t.Errorf("Expected username to be 'test-user', got '%s'", cfg.Username)
	}

	if cfg.Password != "test-pass" {
		t.Errorf("Expected password to be 'test-pass', got '%s'", cfg.Password)
	}

	if cfg.MaxConns != 20 {
		t.Errorf("Expected MaxConns to be 20, got %d", cfg.MaxConns)
	}

	if cfg.MinConns != 5 {
		t.Errorf("Expected MinConns to be 5, got %d", cfg.MinConns)
	}

	if cfg.ApplicationName != "test-app" {
		t.Errorf("Expected ApplicationName to be 'test-app', got '%s'", cfg.ApplicationName)
	}

	if cfg.TLSMode != TLSModeRequire {
		t.Errorf("Expected TLSMode to be 'require', got '%s'", cfg.TLSMode)
	}
}

func TestConfigWithConnectionString(t *testing.T) {
	connStr := "postgres://user:pass@localhost:5432/testdb"
	cfg := NewConfig(WithConnectionString(connStr))

	if cfg.ConnString != connStr {
		t.Errorf("Expected ConnString to be '%s', got '%s'", connStr, cfg.ConnString)
	}

	result := cfg.ConnectionString()
	if result != connStr {
		t.Errorf("Expected ConnectionString() to return '%s', got '%s'", connStr, result)
	}
}

func TestConfigWithTLSConfig(t *testing.T) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	cfg := NewConfig(WithTLSConfig(tlsConfig))

	if cfg.TLSConfig == nil {
		t.Error("Expected TLSConfig to be set")
	}

	if !cfg.TLSConfig.InsecureSkipVerify {
		t.Error("Expected TLSConfig.InsecureSkipVerify to be true")
	}
}

func TestConfigWithRetryConfig(t *testing.T) {
	retryConfig := &RetryConfig{
		Enabled:     false,
		MaxRetries:  5,
		InitialWait: 200 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  1.5,
		Jitter:      false,
	}

	cfg := NewConfig(WithRetryConfig(retryConfig))

	if cfg.RetryConfig.Enabled {
		t.Error("Expected RetryConfig.Enabled to be false")
	}

	if cfg.RetryConfig.MaxRetries != 5 {
		t.Errorf("Expected RetryConfig.MaxRetries to be 5, got %d", cfg.RetryConfig.MaxRetries)
	}

	if cfg.RetryConfig.InitialWait != 200*time.Millisecond {
		t.Errorf("Expected RetryConfig.InitialWait to be 200ms, got %v", cfg.RetryConfig.InitialWait)
	}

	if cfg.RetryConfig.MaxWait != 5*time.Second {
		t.Errorf("Expected RetryConfig.MaxWait to be 5s, got %v", cfg.RetryConfig.MaxWait)
	}

	if cfg.RetryConfig.Multiplier != 1.5 {
		t.Errorf("Expected RetryConfig.Multiplier to be 1.5, got %f", cfg.RetryConfig.Multiplier)
	}

	if cfg.RetryConfig.Jitter {
		t.Error("Expected RetryConfig.Jitter to be false")
	}
}

func TestConfigWithRuntimeParam(t *testing.T) {
	cfg := NewConfig(
		WithRuntimeParam("work_mem", "256MB"),
		WithRuntimeParam("shared_preload_libraries", "pg_stat_statements"),
	)

	if cfg.RuntimeParams["work_mem"] != "256MB" {
		t.Errorf("Expected RuntimeParams['work_mem'] to be '256MB', got '%s'", cfg.RuntimeParams["work_mem"])
	}

	if cfg.RuntimeParams["shared_preload_libraries"] != "pg_stat_statements" {
		t.Errorf("Expected RuntimeParams['shared_preload_libraries'] to be 'pg_stat_statements', got '%s'", cfg.RuntimeParams["shared_preload_libraries"])
	}
}

func TestConfigWithProviderSpecific(t *testing.T) {
	cfg := NewConfig(
		WithProviderSpecific("pgx_pool_max_conn_lifetime", 30*time.Second),
		WithProviderSpecific("gorm_disable_foreign_key_constraint_when_migrating", true),
	)

	if cfg.ProviderSpecific["pgx_pool_max_conn_lifetime"] != 30*time.Second {
		t.Errorf("Expected ProviderSpecific['pgx_pool_max_conn_lifetime'] to be 30s, got %v", cfg.ProviderSpecific["pgx_pool_max_conn_lifetime"])
	}

	if cfg.ProviderSpecific["gorm_disable_foreign_key_constraint_when_migrating"] != true {
		t.Errorf("Expected ProviderSpecific['gorm_disable_foreign_key_constraint_when_migrating'] to be true, got %v", cfg.ProviderSpecific["gorm_disable_foreign_key_constraint_when_migrating"])
	}
}

func TestConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "basic connection string",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
			},
			expected: "postgres://user:pass@localhost/testdb",
		},
		{
			name: "connection string with port",
			config: &Config{
				Host:     "localhost",
				Port:     5433,
				Database: "testdb",
				Username: "user",
				Password: "pass",
			},
			expected: "postgres://user:pass@localhost:5433/testdb",
		},
		{
			name: "connection string without password",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
			},
			expected: "postgres://user@localhost/testdb",
		},
		{
			name: "connection string with TLS mode",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				TLSMode:  TLSModeRequire,
			},
			expected: "postgres://user:pass@localhost/testdb?sslmode=require",
		},
		{
			name: "connection string with application name",
			config: &Config{
				Host:            "localhost",
				Port:            5432,
				Database:        "testdb",
				Username:        "user",
				Password:        "pass",
				ApplicationName: "test-app",
			},
			expected: "postgres://user:pass@localhost/testdb?application_name=test-app",
		},
		{
			name: "existing connection string takes precedence",
			config: &Config{
				ConnString: "postgres://override:override@override:1234/override",
				Host:       "localhost",
				Port:       5432,
				Database:   "testdb",
				Username:   "user",
				Password:   "pass",
			},
			expected: "postgres://override:override@override:1234/override",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ConnectionString()
			if result != tt.expected {
				t.Errorf("Expected connection string '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
		errField  string
	}{
		{
			name:      "valid config",
			config:    DefaultConfig(),
			expectErr: false,
		},
		{
			name: "valid config with connection string",
			config: &Config{
				ConnString: "postgres://user:pass@localhost/db",
				MaxConns:   10,
				MinConns:   1,
			},
			expectErr: false,
		},
		{
			name: "missing host and connection string",
			config: &Config{
				MaxConns: 10,
				MinConns: 1,
			},
			expectErr: true,
			errField:  "host",
		},
		{
			name: "invalid max conns",
			config: &Config{
				Host:     "localhost",
				MaxConns: 0,
				MinConns: 1,
			},
			expectErr: true,
			errField:  "max_conns",
		},
		{
			name: "invalid min conns",
			config: &Config{
				Host:     "localhost",
				MaxConns: 10,
				MinConns: -1,
			},
			expectErr: true,
			errField:  "min_conns",
		},
		{
			name: "min conns greater than max conns",
			config: &Config{
				Host:     "localhost",
				MaxConns: 5,
				MinConns: 10,
			},
			expectErr: true,
			errField:  "min_conns",
		},
		{
			name: "negative max conn lifetime",
			config: &Config{
				Host:            "localhost",
				MaxConns:        10,
				MinConns:        1,
				MaxConnLifetime: -1 * time.Second,
			},
			expectErr: true,
			errField:  "max_conn_lifetime",
		},
		{
			name: "negative max conn idle time",
			config: &Config{
				Host:            "localhost",
				MaxConns:        10,
				MinConns:        1,
				MaxConnIdleTime: -1 * time.Second,
			},
			expectErr: true,
			errField:  "max_conn_idle_time",
		},
		{
			name: "negative connect timeout",
			config: &Config{
				Host:           "localhost",
				MaxConns:       10,
				MinConns:       1,
				ConnectTimeout: -1 * time.Second,
			},
			expectErr: true,
			errField:  "connect_timeout",
		},
		{
			name: "negative query timeout",
			config: &Config{
				Host:         "localhost",
				MaxConns:     10,
				MinConns:     1,
				QueryTimeout: -1 * time.Second,
			},
			expectErr: true,
			errField:  "query_timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectErr {
				if err == nil {
					t.Error("Expected validation error, got nil")
					return
				}

				configErr, ok := err.(ErrInvalidConfig)
				if !ok {
					t.Errorf("Expected ErrInvalidConfig, got %T", err)
					return
				}

				if configErr.Field != tt.errField {
					t.Errorf("Expected error field '%s', got '%s'", tt.errField, configErr.Field)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %v", err)
				}
			}
		})
	}
}

func TestErrInvalidConfig(t *testing.T) {
	err := ErrInvalidConfig{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "invalid config field 'test_field': test message"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

// Edge cases and boundary tests

func TestConfigWithTimeouts(t *testing.T) {
	cfg := NewConfig(
		WithConnectTimeout(5*time.Second),
		WithQueryTimeout(30*time.Second),
		WithMaxConnLifetime(60*time.Minute),
		WithMaxConnIdleTime(10*time.Minute),
	)

	if cfg.ConnectTimeout != 5*time.Second {
		t.Errorf("Expected ConnectTimeout to be 5s, got %v", cfg.ConnectTimeout)
	}

	if cfg.QueryTimeout != 30*time.Second {
		t.Errorf("Expected QueryTimeout to be 30s, got %v", cfg.QueryTimeout)
	}

	if cfg.MaxConnLifetime != 60*time.Minute {
		t.Errorf("Expected MaxConnLifetime to be 60m, got %v", cfg.MaxConnLifetime)
	}

	if cfg.MaxConnIdleTime != 10*time.Minute {
		t.Errorf("Expected MaxConnIdleTime to be 10m, got %v", cfg.MaxConnIdleTime)
	}
}

func TestConfigWithSearchPath(t *testing.T) {
	searchPath := []string{"public", "auth", "app"}
	cfg := NewConfig(WithSearchPath(searchPath))

	if len(cfg.SearchPath) != len(searchPath) {
		t.Errorf("Expected SearchPath length to be %d, got %d", len(searchPath), len(cfg.SearchPath))
	}

	for i, path := range searchPath {
		if cfg.SearchPath[i] != path {
			t.Errorf("Expected SearchPath[%d] to be '%s', got '%s'", i, path, cfg.SearchPath[i])
		}
	}
}

func TestConfigWithMultipleTLSModes(t *testing.T) {
	modes := []TLSMode{
		TLSModeDisable,
		TLSModeAllow,
		TLSModePrefer,
		TLSModeRequire,
		TLSModeVerifyCA,
		TLSModeVerifyFull,
	}

	for _, mode := range modes {
		cfg := NewConfig(WithTLSMode(mode))
		if cfg.TLSMode != mode {
			t.Errorf("Expected TLSMode to be '%s', got '%s'", mode, cfg.TLSMode)
		}
	}
}

func TestConfigWithQueryExecModes(t *testing.T) {
	modes := []QueryExecMode{
		QueryExecModeDefault,
		QueryExecModeCacheStatement,
		QueryExecModeCacheDescribe,
		QueryExecModeDescribeExec,
		QueryExecModeExec,
		QueryExecModeSimpleProtocol,
	}

	for _, mode := range modes {
		cfg := NewConfig(WithQueryExecMode(mode))
		if cfg.QueryExecMode != mode {
			t.Errorf("Expected QueryExecMode to be '%s', got '%s'", mode, cfg.QueryExecMode)
		}
	}
}

func TestConfigZeroValues(t *testing.T) {
	cfg := &Config{}

	// Test with zero values
	if cfg.MaxConns != 0 {
		t.Errorf("Expected MaxConns to be 0, got %d", cfg.MaxConns)
	}

	// Test validation fails for zero MaxConns
	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation to fail for zero MaxConns")
	}
}

func TestConfigBoundaryValues(t *testing.T) {
	// Test maximum values
	cfg := NewConfig(
		WithMaxConns(1000),
		WithMinConns(100),
		WithConnectTimeout(time.Hour),
		WithQueryTimeout(time.Hour),
	)

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected validation to pass for large values, got: %v", err)
	}

	// Test minimum valid values
	cfg = NewConfig(
		WithMaxConns(1),
		WithMinConns(0),
		WithConnectTimeout(0),
		WithQueryTimeout(0),
	)

	err = cfg.Validate()
	if err != nil {
		t.Errorf("Expected validation to pass for minimum values, got: %v", err)
	}
}
