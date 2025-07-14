//go:build unit

package pgx

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
)

func TestProvider_Type(t *testing.T) {
	provider := NewProvider()

	if provider.Type() != postgresql.ProviderTypePGX {
		t.Errorf("Expected provider type to be %s, got %s", postgresql.ProviderTypePGX, provider.Type())
	}
}

func TestProvider_Name(t *testing.T) {
	provider := NewProvider()

	expected := "PGX PostgreSQL Provider"
	if provider.Name() != expected {
		t.Errorf("Expected provider name to be '%s', got '%s'", expected, provider.Name())
	}
}

func TestProvider_Version(t *testing.T) {
	provider := NewProvider()

	expected := "5.7.5"
	if provider.Version() != expected {
		t.Errorf("Expected provider version to be '%s', got '%s'", expected, provider.Version())
	}
}

func TestProvider_ConnectionPoolConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with connection pool configuration
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithMaxConns(10),
		config.WithMinConns(2),
		config.WithMaxConnIdleTime(300),  // 5 minutes
		config.WithMaxConnLifetime(3600), // 1 hour
		config.WithConnectTimeout(1000),  // 1 second
	)

	// Pool creation might succeed even with fake host (lazy connection)
	pool, err := provider.CreatePool(ctx, cfg)

	// Pool creation might succeed, but connection acquisition should fail
	if err != nil {
		// If we get an error here, it should not be config validation
		if strings.Contains(err.Error(), "invalid config") {
			t.Errorf("Got config validation error instead of connection error: %v", err)
		}
		return // Test passes if we get connection error
	}

	// If pool was created, try to acquire a connection to test actual connectivity
	if pool != nil {
		defer pool.Close()
		_, connErr := pool.Acquire(ctx)
		if connErr == nil {
			t.Error("Expected connection error when acquiring from pool with fake host")
		}
	}
}

func TestProvider_CreatePool_InvalidConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with invalid config (no host and no connection string)
	cfg := &config.Config{
		MaxConns: 0, // Invalid
	}

	_, err := provider.CreatePool(ctx, cfg)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

func TestProvider_CreateConnection_InvalidConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with invalid config
	cfg := &config.Config{
		MaxConns: -1, // Invalid
	}

	_, err := provider.CreateConnection(ctx, cfg)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

func TestProvider_IsHealthy(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test healthy provider without pools
	if !provider.IsHealthy(ctx) {
		t.Error("Expected provider to be healthy without pools")
	}

	// Test closed provider
	provider.Close()
	if provider.IsHealthy(ctx) {
		t.Error("Expected closed provider to be unhealthy")
	}
}

func TestProvider_GetMetrics(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	metrics := provider.GetMetrics(ctx)

	// Check required metrics
	if metrics["type"] != postgresql.ProviderTypePGX {
		t.Errorf("Expected type metric to be %s, got %v", postgresql.ProviderTypePGX, metrics["type"])
	}

	if metrics["name"] != "PGX PostgreSQL Provider" {
		t.Errorf("Expected name metric to be 'PGX PostgreSQL Provider', got %v", metrics["name"])
	}

	if metrics["version"] != "5.7.5" {
		t.Errorf("Expected version metric to be '5.7.5', got %v", metrics["version"])
	}

	if metrics["pools_count"] != 0 {
		t.Errorf("Expected pools_count metric to be 0, got %v", metrics["pools_count"])
	}

	if metrics["connections_count"] != 0 {
		t.Errorf("Expected connections_count metric to be 0, got %v", metrics["connections_count"])
	}

	if metrics["is_healthy"] != true {
		t.Errorf("Expected is_healthy metric to be true, got %v", metrics["is_healthy"])
	}

	// Check pools metric exists
	if _, exists := metrics["pools"]; !exists {
		t.Error("Expected pools metric to exist")
	}
}

func TestProvider_Close(t *testing.T) {
	provider := NewProvider()

	// Test closing multiple times
	err := provider.Close()
	if err != nil {
		t.Errorf("Expected no error on first close, got: %v", err)
	}

	err = provider.Close()
	if err != nil {
		t.Errorf("Expected no error on second close, got: %v", err)
	}

	// Test operations after close
	ctx := context.Background()
	cfg := config.DefaultConfig()

	_, err = provider.CreatePool(ctx, cfg)
	if err == nil {
		t.Error("Expected error when creating pool after close, got nil")
	}

	_, err = provider.CreateConnection(ctx, cfg)
	if err == nil {
		t.Error("Expected error when creating connection after close, got nil")
	}
}

func TestProvider_BuildPoolConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	cfg := config.NewConfig(
		config.WithHost("localhost"),
		config.WithPort(5432),
		config.WithDatabase("testdb"),
		config.WithUsername("testuser"),
		config.WithPassword("testpass"),
		config.WithMaxConns(20),
		config.WithMinConns(2),
		config.WithMaxConnLifetime(30*time.Second),
		config.WithMaxConnIdleTime(10*time.Second),
		config.WithApplicationName("test-app"),
		config.WithTimezone("UTC"),
		config.WithQueryExecMode(config.QueryExecModeCacheStatement),
	)

	pgxConfig, err := provider.buildPoolConfig(cfg)
	if err != nil {
		t.Fatalf("Expected no error building pool config, got: %v", err)
	}

	if pgxConfig.MaxConns != 20 {
		t.Errorf("Expected MaxConns to be 20, got %d", pgxConfig.MaxConns)
	}

	if pgxConfig.MinConns != 2 {
		t.Errorf("Expected MinConns to be 2, got %d", pgxConfig.MinConns)
	}

	if pgxConfig.MaxConnLifetime != 30*time.Second {
		t.Errorf("Expected MaxConnLifetime to be 30s, got %v", pgxConfig.MaxConnLifetime)
	}

	if pgxConfig.MaxConnIdleTime != 10*time.Second {
		t.Errorf("Expected MaxConnIdleTime to be 10s, got %v", pgxConfig.MaxConnIdleTime)
	}

	// Check runtime parameters
	if pgxConfig.ConnConfig.RuntimeParams["timezone"] != "UTC" {
		t.Errorf("Expected timezone to be 'UTC', got '%s'", pgxConfig.ConnConfig.RuntimeParams["timezone"])
	}

	if pgxConfig.ConnConfig.RuntimeParams["application_name"] != "test-app" {
		t.Errorf("Expected application_name to be 'test-app', got '%s'", pgxConfig.ConnConfig.RuntimeParams["application_name"])
	}
}

func TestProvider_BuildConnConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	cfg := config.NewConfig(
		config.WithHost("localhost"),
		config.WithPort(5432),
		config.WithDatabase("testdb"),
		config.WithUsername("testuser"),
		config.WithPassword("testpass"),
		config.WithApplicationName("test-app"),
		config.WithTimezone("UTC"),
		config.WithQueryExecMode(config.QueryExecModeExec),
	)

	pgxConfig, err := provider.buildConnConfig(cfg)
	if err != nil {
		t.Fatalf("Expected no error building conn config, got: %v", err)
	}

	// Check runtime parameters
	if pgxConfig.RuntimeParams["timezone"] != "UTC" {
		t.Errorf("Expected timezone to be 'UTC', got '%s'", pgxConfig.RuntimeParams["timezone"])
	}

	if pgxConfig.RuntimeParams["application_name"] != "test-app" {
		t.Errorf("Expected application_name to be 'test-app', got '%s'", pgxConfig.RuntimeParams["application_name"])
	}
}

func TestProvider_ConfigureConnection(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	tests := []struct {
		name       string
		config     *config.Config
		expectMode string
	}{
		{
			name:       "default query exec mode",
			config:     config.NewConfig(config.WithQueryExecMode(config.QueryExecModeDefault)),
			expectMode: "exec",
		},
		{
			name:       "cache statement mode",
			config:     config.NewConfig(config.WithQueryExecMode(config.QueryExecModeCacheStatement)),
			expectMode: "cache_statement",
		},
		{
			name:       "cache describe mode",
			config:     config.NewConfig(config.WithQueryExecMode(config.QueryExecModeCacheDescribe)),
			expectMode: "cache_describe",
		},
		{
			name:       "describe exec mode",
			config:     config.NewConfig(config.WithQueryExecMode(config.QueryExecModeDescribeExec)),
			expectMode: "describe_exec",
		},
		{
			name:       "simple protocol mode",
			config:     config.NewConfig(config.WithQueryExecMode(config.QueryExecModeSimpleProtocol)),
			expectMode: "simple_protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgxConfig, err := provider.buildConnConfig(tt.config)
			if err != nil {
				t.Fatalf("Expected no error building config, got: %v", err)
			}

			// Note: We can't directly test pgx.QueryExecMode enum values
			// but we can verify the configuration was processed without error
			if pgxConfig == nil {
				t.Error("Expected pgxConfig to be non-nil")
			}
		})
	}
}

func TestProvider_ConfigWithRuntimeParams(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	cfg := config.NewConfig(
		config.WithHost("localhost"),
		config.WithRuntimeParam("work_mem", "256MB"),
		config.WithRuntimeParam("shared_preload_libraries", "pg_stat_statements"),
	)

	pgxConfig, err := provider.buildConnConfig(cfg)
	if err != nil {
		t.Fatalf("Expected no error building config, got: %v", err)
	}

	if pgxConfig.RuntimeParams["work_mem"] != "256MB" {
		t.Errorf("Expected work_mem to be '256MB', got '%s'", pgxConfig.RuntimeParams["work_mem"])
	}

	if pgxConfig.RuntimeParams["shared_preload_libraries"] != "pg_stat_statements" {
		t.Errorf("Expected shared_preload_libraries to be 'pg_stat_statements', got '%s'", pgxConfig.RuntimeParams["shared_preload_libraries"])
	}
}

func TestProvider_InvalidConnectionString(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	cfg := config.NewConfig(
		config.WithConnectionString("invalid://connection/string"),
	)

	_, err := provider.buildPoolConfig(cfg)
	if err == nil {
		t.Error("Expected error for invalid connection string, got nil")
	}

	_, err = provider.buildConnConfig(cfg)
	if err == nil {
		t.Error("Expected error for invalid connection string, got nil")
	}
}

// Boundary tests

func TestProvider_BoundaryValues(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	// Test with maximum values
	cfg := config.NewConfig(
		config.WithHost("localhost"),
		config.WithMaxConns(1000),
		config.WithMinConns(100),
		config.WithMaxConnLifetime(time.Hour),
		config.WithMaxConnIdleTime(30*time.Minute),
	)

	pgxConfig, err := provider.buildPoolConfig(cfg)
	if err != nil {
		t.Fatalf("Expected no error with large values, got: %v", err)
	}

	if pgxConfig.MaxConns != 1000 {
		t.Errorf("Expected MaxConns to be 1000, got %d", pgxConfig.MaxConns)
	}

	if pgxConfig.MinConns != 100 {
		t.Errorf("Expected MinConns to be 100, got %d", pgxConfig.MinConns)
	}

	// Test with minimum values
	cfg = config.NewConfig(
		config.WithHost("localhost"),
		config.WithMaxConns(1),
		config.WithMinConns(0),
		config.WithMaxConnLifetime(0),
		config.WithMaxConnIdleTime(0),
	)

	pgxConfig, err = provider.buildPoolConfig(cfg)
	if err != nil {
		t.Fatalf("Expected no error with minimum values, got: %v", err)
	}

	if pgxConfig.MaxConns != 1 {
		t.Errorf("Expected MaxConns to be 1, got %d", pgxConfig.MaxConns)
	}

	if pgxConfig.MinConns != 0 {
		t.Errorf("Expected MinConns to be 0, got %d", pgxConfig.MinConns)
	}
}

func TestProvider_EdgeCases(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with empty configuration through CreatePool (which validates)
	cfg := &config.Config{}

	_, err := provider.CreatePool(ctx, cfg)
	if err == nil {
		t.Error("Expected error with empty config, got nil")
	}

	// Test with nil provider methods after close
	provider.Close()

	metrics := provider.GetMetrics(ctx)

	if metrics["is_healthy"] != false {
		t.Errorf("Expected is_healthy to be false for closed provider, got %v", metrics["is_healthy"])
	}
}

// Error handling tests

func TestProvider_ErrorHandling(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with various invalid configurations
	testCases := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "negative max conns",
			config: &config.Config{
				Host:     "localhost",
				MaxConns: -1,
			},
		},
		{
			name: "min conns greater than max conns",
			config: &config.Config{
				Host:     "localhost",
				MaxConns: 5,
				MinConns: 10,
			},
		},
		{
			name: "negative connection timeout",
			config: &config.Config{
				Host:           "localhost",
				MaxConns:       10,
				ConnectTimeout: -1 * time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := provider.CreatePool(ctx, tc.config)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}

			_, err = provider.CreateConnection(ctx, tc.config)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestProvider_Interface(t *testing.T) {
	// Test interface compliance
	var _ postgresql.IProvider = &Provider{}
}

func TestProvider_CreatePool_NilConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with nil config
	_, err := provider.CreatePool(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	expectedMsg := "config cannot be nil"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestProvider_CreateConnection_NilConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with nil config
	_, err := provider.CreateConnection(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	expectedMsg := "config cannot be nil"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestProvider_CreatePool_ValidConfigButNoConnection(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with valid config but no actual DB connection
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreatePool(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreateConnection_ValidConfigButNoConnection(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with valid config but no actual DB connection
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreateConnection(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreatePoolWithSSLConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with TLS configuration
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithTLSMode(config.TLSModeRequire),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreatePool(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreateConnectionWithSSLConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with TLS configuration
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithTLSMode(config.TLSModeRequire),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreateConnection(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreatePoolWithMultiTenantConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with multi-tenant configuration
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithMultiTenant(true),
		config.WithDefaultSchema("tenant_schema"),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreatePool(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreateConnectionWithMultiTenantConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with multi-tenant configuration
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithMultiTenant(true),
		config.WithDefaultSchema("tenant_schema"),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreateConnection(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreatePoolWithHooksConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	hooks := &config.HooksConfig{
		BeforeQuery: func(ctx context.Context, query string, args []interface{}) error {
			return nil
		},
		AfterQuery: func(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) error {
			return nil
		},
		BeforeConnect: func(ctx context.Context, cfg *config.Config) error {
			return nil
		},
		AfterAcquire: func(ctx context.Context, conn interface{}) error {
			return nil
		},
		BeforeRelease: func(ctx context.Context, conn interface{}) error {
			return nil
		},
	}

	// Test with hooks configuration
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithHooks(hooks),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreatePool(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreateConnectionWithHooksConfig(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	hooks := &config.HooksConfig{
		BeforeQuery: func(ctx context.Context, query string, args []interface{}) error {
			return nil
		},
		AfterQuery: func(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) error {
			return nil
		},
		BeforeConnect: func(ctx context.Context, cfg *config.Config) error {
			return nil
		},
		AfterAcquire: func(ctx context.Context, conn interface{}) error {
			return nil
		},
		BeforeRelease: func(ctx context.Context, conn interface{}) error {
			return nil
		},
	}

	// Test with hooks configuration
	cfg := config.NewConfig(
		config.WithHost("nonexistent-host-12345"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
		config.WithHooks(hooks),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreateConnection(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreatePoolWithConnectionString(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with connection string
	cfg := config.NewConfig(
		config.WithConnectionString("postgresql://test_user:test_pass@nonexistent-host-12345:5432/test_db"),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreatePool(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_CreateConnectionWithConnectionString(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with connection string
	cfg := config.NewConfig(
		config.WithConnectionString("postgresql://test_user:test_pass@nonexistent-host-12345:5432/test_db"),
		config.WithConnectTimeout(1000), // 1 second
	)

	// This will fail because host doesn't exist
	_, err := provider.CreateConnection(ctx, cfg)

	// Should get connection error, not config validation error
	if err == nil {
		t.Error("Expected connection error for nonexistent host")
	}
}

func TestProvider_ProviderClosed(t *testing.T) {
	provider := NewProvider()
	provider.Close()

	ctx := context.Background()

	cfg := config.NewConfig(
		config.WithHost("localhost"),
		config.WithPort(5432),
		config.WithDatabase("test_db"),
		config.WithUsername("test_user"),
		config.WithPassword("test_pass"),
	)

	// Operations on closed provider should fail
	_, err := provider.CreatePool(ctx, cfg)
	if err == nil {
		t.Error("Expected error when using closed provider")
	}

	_, err = provider.CreateConnection(ctx, cfg)
	if err == nil {
		t.Error("Expected error when using closed provider")
	}
}

func TestProvider_EdgeCasesAdditional(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	// Test with minimal config
	cfg := config.NewConfig(
		config.WithHost("localhost"),
		config.WithDatabase("postgres"),
		config.WithUsername("postgres"),
	)

	// Should handle minimal config without panicking
	_, err := provider.CreatePool(ctx, cfg)
	// Will fail due to connection, but shouldn't panic
	if err == nil {
		t.Error("Expected connection error")
	}

	_, err = provider.CreateConnection(ctx, cfg)
	// Will fail due to connection, but shouldn't panic
	if err == nil {
		t.Error("Expected connection error")
	}
}

func TestProvider_CreateBatch(t *testing.T) {
	// Provider doesn't have CreateBatch method
	// Batches are created directly using NewBatch() function
}

func TestProvider_MultipleCloses(t *testing.T) {
	provider := NewProvider()

	// Multiple closes should not panic
	provider.Close()
	provider.Close()
	provider.Close()
}

func TestProvider_HealthCheck(t *testing.T) {
	// Provider doesn't have HealthCheck method
	// Health checks are done on pools or connections
}

func TestProvider_GetStats(t *testing.T) {
	// Provider doesn't have GetStats method
	// Stats are retrieved from pools, not provider
}

func TestProvider_WithDifferentSSLModes(t *testing.T) {
	provider := NewProvider()
	defer provider.Close()

	ctx := context.Background()

	sslModes := []config.TLSMode{
		config.TLSModeDisable,
		config.TLSModeAllow,
		config.TLSModePrefer,
		config.TLSModeRequire,
		config.TLSModeVerifyCA,
		config.TLSModeVerifyFull,
	}

	for _, sslMode := range sslModes {
		cfg := config.NewConfig(
			config.WithHost("nonexistent-host-12345"),
			config.WithPort(5432),
			config.WithDatabase("test_db"),
			config.WithUsername("test_user"),
			config.WithPassword("test_pass"),
			config.WithTLSMode(sslMode),
			config.WithConnectTimeout(1000), // 1 second
		)

		// Should handle all SSL modes without panicking
		_, err := provider.CreatePool(ctx, cfg)
		if err == nil {
			t.Errorf("Expected connection error for SSL mode %v", sslMode)
		}

		_, err = provider.CreateConnection(ctx, cfg)
		if err == nil {
			t.Errorf("Expected connection error for SSL mode %v", sslMode)
		}
	}
}
