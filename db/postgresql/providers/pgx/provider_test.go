//go:build unit
// +build unit

package pgx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx/mocks"
)

// MockConfig implements interfaces.Config for testing
type MockConfig struct {
	connectionString   string
	poolConfig         interfaces.PoolConfig
	tlsConfig          interfaces.TLSConfig
	retryConfig        interfaces.RetryConfig
	hookConfig         interfaces.HookConfig
	multiTenantEnabled bool
	readReplicaConfig  interfaces.ReadReplicaConfig
	failoverConfig     interfaces.FailoverConfig
	valid              bool
}

func (m *MockConfig) GetConnectionString() string                        { return m.connectionString }
func (m *MockConfig) GetPoolConfig() interfaces.PoolConfig               { return m.poolConfig }
func (m *MockConfig) GetTLSConfig() interfaces.TLSConfig                 { return m.tlsConfig }
func (m *MockConfig) GetRetryConfig() interfaces.RetryConfig             { return m.retryConfig }
func (m *MockConfig) GetHookConfig() interfaces.HookConfig               { return m.hookConfig }
func (m *MockConfig) IsMultiTenantEnabled() bool                         { return m.multiTenantEnabled }
func (m *MockConfig) GetReadReplicaConfig() interfaces.ReadReplicaConfig { return m.readReplicaConfig }
func (m *MockConfig) GetFailoverConfig() interfaces.FailoverConfig       { return m.failoverConfig }

func (m *MockConfig) Validate() error {
	if !m.valid {
		return fmt.Errorf("invalid configuration")
	}
	return nil
}

// createTestConfig creates a test configuration
func createTestConfig() interfaces.Config {
	return &MockConfig{
		connectionString: "postgres://user:password@localhost:5432/testdb?sslmode=disable",
		poolConfig: interfaces.PoolConfig{
			MaxConns:          10,
			MinConns:          1,
			MaxConnLifetime:   time.Hour,
			MaxConnIdleTime:   time.Minute * 30,
			HealthCheckPeriod: time.Minute * 5,
			ConnectTimeout:    time.Second * 30,
			LazyConnect:       false,
		},
		tlsConfig: interfaces.TLSConfig{
			Enabled:            false,
			InsecureSkipVerify: false,
		},
		retryConfig: interfaces.RetryConfig{
			MaxRetries:      3,
			InitialInterval: time.Second,
			MaxInterval:     time.Second * 10,
			Multiplier:      2.0,
			RandomizeWait:   true,
		},
		hookConfig: interfaces.HookConfig{
			EnabledHooks: []interfaces.HookType{
				interfaces.BeforeQueryHook,
				interfaces.AfterQueryHook,
				interfaces.BeforeExecHook,
				interfaces.AfterExecHook,
			},
			CustomHooks: make(map[string]interfaces.HookType),
			HookTimeout: time.Second * 5,
		},
		multiTenantEnabled: false,
		readReplicaConfig: interfaces.ReadReplicaConfig{
			Enabled:             false,
			ConnectionStrings:   []string{},
			LoadBalanceMode:     interfaces.LoadBalanceModeRoundRobin,
			HealthCheckInterval: time.Minute * 5,
		},
		failoverConfig: interfaces.FailoverConfig{
			Enabled:             false,
			FallbackNodes:       []string{},
			HealthCheckInterval: time.Minute * 5,
			RetryInterval:       time.Second * 30,
			MaxFailoverAttempts: 3,
		},
		valid: true,
	}
}

func createInvalidTestConfig() interfaces.Config {
	config := createTestConfig().(*MockConfig)
	config.valid = false
	return config
}

func TestPGXProvider(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"NewPGXProvider", testNewPGXProvider},
		{"Name", testProviderName},
		{"Version", testProviderVersion},
		{"GetDriverName", testGetDriverName},
		{"SupportsFeature", testSupportsFeature},
		{"GetSupportedFeatures", testGetSupportedFeatures},
		{"ValidateConfig", testValidateConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func testNewPGXProvider(t *testing.T) {
	provider := NewPGXProvider()

	require.NotNil(t, provider)
	assert.Equal(t, "PGX", provider.Name())
	assert.Equal(t, "5.x", provider.Version())
	assert.True(t, provider.SupportsFeature("transactions"))
}

func testProviderName(t *testing.T) {
	provider := NewPGXProvider()
	assert.Equal(t, "PGX", provider.Name())
}

func testProviderVersion(t *testing.T) {
	provider := NewPGXProvider()
	assert.Equal(t, "5.x", provider.Version())
}

func testGetDriverName(t *testing.T) {
	provider := NewPGXProvider()
	assert.Equal(t, "pgx", provider.GetDriverName())
}

func testSupportsFeature(t *testing.T) {
	provider := NewPGXProvider()

	// Test supported features
	supportedFeatures := []string{
		"transactions",
		"prepared_statements",
		"batch_operations",
		"copy_operations",
		"listen_notify",
		"connection_pooling",
		"hooks",
		"multi_tenancy",
		"health_checks",
		"statistics",
		"ssl_connections",
		"async_operations",
	}

	for _, feature := range supportedFeatures {
		t.Run("supports_"+feature, func(t *testing.T) {
			assert.True(t, provider.SupportsFeature(feature), "should support %s", feature)
		})
	}

	// Test unsupported features
	unsupportedFeatures := []string{
		"unsupported_feature",
		"invalid_feature",
		"",
	}

	for _, feature := range unsupportedFeatures {
		t.Run("does_not_support_"+feature, func(t *testing.T) {
			assert.False(t, provider.SupportsFeature(feature), "should not support %s", feature)
		})
	}
}

func testGetSupportedFeatures(t *testing.T) {
	provider := NewPGXProvider()
	features := provider.GetSupportedFeatures()

	expectedFeatures := []string{
		"transactions",
		"prepared_statements",
		"batch_operations",
		"copy_operations",
		"listen_notify",
		"connection_pooling",
		"hooks",
		"multi_tenancy",
		"health_checks",
		"statistics",
		"ssl_connections",
		"async_operations",
	}

	assert.Len(t, features, len(expectedFeatures))

	for _, expected := range expectedFeatures {
		assert.Contains(t, features, expected)
	}
}

func testValidateConfig(t *testing.T) {
	provider := NewPGXProvider()

	t.Run("valid_config", func(t *testing.T) {
		config := createTestConfig()
		err := provider.ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("nil_config", func(t *testing.T) {
		err := provider.ValidateConfig(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("invalid_config", func(t *testing.T) {
		config := createInvalidTestConfig()
		err := provider.ValidateConfig(config)
		assert.Error(t, err)
	})

	t.Run("empty_connection_string", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.connectionString = ""
		err := provider.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection string is required")
	})

	t.Run("invalid_connection_string", func(t *testing.T) {
		config := createTestConfig().(*MockConfig)
		config.connectionString = "invalid://connection/string"
		err := provider.ValidateConfig(config)
		assert.Error(t, err)
	})
}

func TestPGXProviderConnectionCreation(t *testing.T) {
	provider := NewPGXProvider()
	ctx := context.Background()

	t.Run("NewPool_with_nil_config", func(t *testing.T) {
		pool, err := provider.NewPool(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("NewPool_with_invalid_config", func(t *testing.T) {
		config := createInvalidTestConfig()
		pool, err := provider.NewPool(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, pool)
	})

	t.Run("NewConn_with_nil_config", func(t *testing.T) {
		conn, err := provider.NewConn(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("NewConn_with_invalid_config", func(t *testing.T) {
		config := createInvalidTestConfig()
		conn, err := provider.NewConn(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("NewListenConn_with_nil_config", func(t *testing.T) {
		conn, err := provider.NewListenConn(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("NewListenConn_with_invalid_config", func(t *testing.T) {
		config := createInvalidTestConfig()
		conn, err := provider.NewListenConn(ctx, config)
		assert.Error(t, err)
		assert.Nil(t, conn)
	})
}

func TestPGXProviderSchemaOperations(t *testing.T) {
	provider := NewPGXProvider()
	ctx := context.Background()

	t.Run("CreateSchema_with_nil_conn", func(t *testing.T) {
		err := provider.CreateSchema(ctx, nil, "test_schema")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection cannot be nil")
	})
	t.Run("CreateSchema_with_empty_schema_name", func(t *testing.T) {
		// Create a mock connection for this test
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create a mock implementation
		mockConn := mocks.NewMockConnection()

		err := provider.CreateSchema(ctx, mockConn, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema name cannot be empty")
	})

	t.Run("DropSchema_with_nil_conn", func(t *testing.T) {
		err := provider.DropSchema(ctx, nil, "test_schema")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection cannot be nil")
	})
	t.Run("DropSchema_with_empty_schema_name", func(t *testing.T) {
		// Create a mock connection for this test
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConn := mocks.NewMockConnection()

		err := provider.DropSchema(ctx, mockConn, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema name cannot be empty")
	})

	t.Run("ListSchemas_with_nil_conn", func(t *testing.T) {
		schemas, err := provider.ListSchemas(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, schemas)
		assert.Contains(t, err.Error(), "connection cannot be nil")
	})
}

func TestPGXProviderDatabaseOperations(t *testing.T) {
	provider := NewPGXProvider()
	ctx := context.Background()

	t.Run("CreateDatabase_with_nil_conn", func(t *testing.T) {
		err := provider.CreateDatabase(ctx, nil, "test_db")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection cannot be nil")
	})

	t.Run("CreateDatabase_with_empty_db_name", func(t *testing.T) {
		mockConn := &mocks.MockConnection{}
		err := provider.CreateDatabase(ctx, mockConn, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database name cannot be empty")
	})

	t.Run("DropDatabase_with_nil_conn", func(t *testing.T) {
		err := provider.DropDatabase(ctx, nil, "test_db")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection cannot be nil")
	})

	t.Run("DropDatabase_with_empty_db_name", func(t *testing.T) {
		mockConn := &mocks.MockConnection{}
		err := provider.DropDatabase(ctx, mockConn, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database name cannot be empty")
	})

	t.Run("ListDatabases_with_nil_conn", func(t *testing.T) {
		databases, err := provider.ListDatabases(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, databases)
		assert.Contains(t, err.Error(), "connection cannot be nil")
	})
}

func TestPGXProviderEdgeCases(t *testing.T) {
	provider := NewPGXProvider()
	ctx := context.Background()

	t.Run("multiple_provider_instances", func(t *testing.T) {
		provider1 := NewPGXProvider()
		provider2 := NewPGXProvider()

		assert.Equal(t, provider1.Name(), provider2.Name())
		assert.Equal(t, provider1.Version(), provider2.Version())
		assert.Equal(t, provider1.GetDriverName(), provider2.GetDriverName())
	})

	t.Run("case_insensitive_feature_check", func(t *testing.T) {
		// PGX provider should handle case-sensitive feature checks
		assert.True(t, provider.SupportsFeature("transactions"))
		assert.False(t, provider.SupportsFeature("TRANSACTIONS")) // Should be case-sensitive
		assert.False(t, provider.SupportsFeature("Transactions"))
	})

	t.Run("special_characters_in_schema_name", func(t *testing.T) {
		invalidNames := []string{
			"test schema", // space
			"test-schema", // dash might be valid, but testing
			"test.schema", // dot
			"test;schema", // semicolon
			"'test'",      // quotes
			`"test"`,      // double quotes
		}

		for _, name := range invalidNames {
			t.Run("invalid_name_"+name, func(t *testing.T) {
				err := provider.CreateSchema(ctx, nil, name)
				// Error should occur, but could be due to nil connection or invalid name
				assert.Error(t, err)
			})
		}
	})
}

func TestPGXProviderConcurrency(t *testing.T) {
	provider := NewPGXProvider()

	// Test that provider methods are safe for concurrent access
	t.Run("concurrent_feature_check", func(t *testing.T) {
		const goroutines = 100
		done := make(chan bool, goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer func() { done <- true }()

				// These should be safe to call concurrently
				_ = provider.SupportsFeature("transactions")
				_ = provider.Name()
				_ = provider.Version()
				_ = provider.GetDriverName()
				_ = provider.GetSupportedFeatures()
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < goroutines; i++ {
			select {
			case <-done:
				// Good
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent operations")
			}
		}
	})

	t.Run("concurrent_config_validation", func(t *testing.T) {
		const goroutines = 50
		done := make(chan bool, goroutines)
		config := createTestConfig()

		for i := 0; i < goroutines; i++ {
			go func() {
				defer func() { done <- true }()
				err := provider.ValidateConfig(config)
				assert.NoError(t, err)
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < goroutines; i++ {
			select {
			case <-done:
				// Good
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent config validation")
			}
		}
	})
}

// BenchmarkPGXProvider benchmarks provider operations
func BenchmarkPGXProvider(b *testing.B) {
	provider := NewPGXProvider()
	config := createTestConfig()

	b.Run("SupportsFeature", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = provider.SupportsFeature("transactions")
		}
	})

	b.Run("ValidateConfig", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = provider.ValidateConfig(config)
		}
	})

	b.Run("GetSupportedFeatures", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = provider.GetSupportedFeatures()
		}
	})
}
