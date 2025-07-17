package config

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	// Test with connection string
	config := NewDefaultConfig("postgres://localhost:5432/testdb")
	assert.NotNil(t, config)
	assert.Equal(t, "postgres://localhost:5432/testdb", config.GetConnectionString())

	// Test default values
	poolConfig := config.GetPoolConfig()
	assert.True(t, poolConfig.MaxConns > 0)
	assert.True(t, poolConfig.MinConns >= 0)
	assert.True(t, poolConfig.MaxConnLifetime > 0)
	assert.True(t, poolConfig.MaxConnIdleTime > 0)

	tlsConfig := config.GetTLSConfig()
	assert.False(t, tlsConfig.Enabled)

	retryConfig := config.GetRetryConfig()
	assert.True(t, retryConfig.MaxRetries >= 0)
	assert.True(t, retryConfig.InitialInterval > 0)
	assert.True(t, retryConfig.MaxInterval > 0)
	assert.True(t, retryConfig.Multiplier > 1)

	hookConfig := config.GetHookConfig()
	assert.True(t, hookConfig.HookTimeout > 0)
	assert.NotNil(t, hookConfig.EnabledHooks)
	assert.NotNil(t, hookConfig.CustomHooks)

	assert.False(t, config.IsMultiTenantEnabled())

	replicaConfig := config.GetReadReplicaConfig()
	assert.False(t, replicaConfig.Enabled)

	failoverConfig := config.GetFailoverConfig()
	assert.False(t, failoverConfig.Enabled)
}

func TestNewConfig_WithOptions(t *testing.T) {
	// Test NewConfig with various options
	config := NewConfig(
		WithConnectionString("postgres://user:password@localhost:5432/db"),
		WithMaxConns(100),
		WithMinConns(5),
		WithMaxConnLifetime(1*time.Hour),
		WithConnectTimeout(30*time.Second),
		WithLazyConnect(true),
		WithMaxRetries(3),
		WithMultiTenant(true),
	)

	assert.NotNil(t, config)
	assert.Equal(t, "postgres://user:password@localhost:5432/db", config.GetConnectionString())
	assert.True(t, config.IsMultiTenantEnabled())

	poolConfig := config.GetPoolConfig()
	assert.NotNil(t, poolConfig)

	retryConfig := config.GetRetryConfig()
	assert.NotNil(t, retryConfig)
	assert.Equal(t, 3, retryConfig.MaxRetries)
}

func TestConfigOption_WithConnectionString(t *testing.T) {
	config := NewConfig(WithConnectionString("postgres://test:5432/db"))
	assert.Equal(t, "postgres://test:5432/db", config.GetConnectionString())
}

func TestConfigOption_WithMaxConns(t *testing.T) {
	config := NewConfig(WithMaxConns(50))
	poolConfig := config.GetPoolConfig()
	assert.NotNil(t, poolConfig)
	// MaxConns should be applied through the option
}

func TestConfigOption_WithMinConns(t *testing.T) {
	config := NewConfig(WithMinConns(10))
	poolConfig := config.GetPoolConfig()
	assert.NotNil(t, poolConfig)
	// MinConns should be applied through the option
}

func TestConfigOption_WithMaxConnLifetime(t *testing.T) {
	config := NewConfig(WithMaxConnLifetime(2 * time.Hour))
	poolConfig := config.GetPoolConfig()
	assert.NotNil(t, poolConfig)
	// MaxConnLifetime should be applied through the option
}

func TestConfigOption_WithMaxConnIdleTime(t *testing.T) {
	config := NewConfig(WithMaxConnIdleTime(15 * time.Minute))
	poolConfig := config.GetPoolConfig()
	assert.NotNil(t, poolConfig)
	// MaxConnIdleTime should be applied through the option
}

func TestConfigOption_WithHealthCheckPeriod(t *testing.T) {
	config := NewConfig(WithHealthCheckPeriod(1 * time.Minute))
	poolConfig := config.GetPoolConfig()
	assert.NotNil(t, poolConfig)
	// HealthCheckPeriod should be applied through the option
}

func TestConfigOption_WithConnectTimeout(t *testing.T) {
	config := NewConfig(WithConnectTimeout(45 * time.Second))
	poolConfig := config.GetPoolConfig()
	assert.NotNil(t, poolConfig)
	// ConnectTimeout should be applied through the option
}

func TestConfigOption_WithLazyConnect(t *testing.T) {
	config := NewConfig(WithLazyConnect(false))
	poolConfig := config.GetPoolConfig()
	assert.NotNil(t, poolConfig)
	// LazyConnect should be applied through the option
}

func TestConfigOption_WithTLS(t *testing.T) {
	tlsConf := &tls.Config{InsecureSkipVerify: true}
	config := NewConfig(WithTLS(true, tlsConf))
	tlsConfig := config.GetTLSConfig()
	assert.NotNil(t, tlsConfig)
	assert.True(t, tlsConfig.Enabled)
}

func TestConfigOption_WithTLSFiles(t *testing.T) {
	config := NewConfig(WithTLSFiles("cert.pem", "key.pem", "ca.pem"))
	tlsConfig := config.GetTLSConfig()
	assert.NotNil(t, tlsConfig)
	assert.Equal(t, "cert.pem", tlsConfig.CertFile)
	assert.Equal(t, "key.pem", tlsConfig.KeyFile)
	assert.Equal(t, "ca.pem", tlsConfig.CAFile)
}

func TestConfigOption_WithRetryConfig(t *testing.T) {
	// Create a simple retry config
	retryConfig := interfaces.RetryConfig{
		MaxRetries:      5,
		InitialInterval: time.Millisecond * 200,
		MaxInterval:     time.Second * 10,
		Multiplier:      3.0,
		RandomizeWait:   false,
	}
	config := NewConfig(WithRetryConfig(retryConfig))
	resultConfig := config.GetRetryConfig()
	assert.Equal(t, 5, resultConfig.MaxRetries)
	assert.Equal(t, time.Millisecond*200, resultConfig.InitialInterval)
}

func TestConfigOption_WithMaxRetries(t *testing.T) {
	config := NewConfig(WithMaxRetries(5))
	retryConfig := config.GetRetryConfig()
	assert.NotNil(t, retryConfig)
	assert.Equal(t, 5, retryConfig.MaxRetries)
}

func TestConfigOption_WithHookConfig(t *testing.T) {
	// Create a simple hook config
	hookConfig := interfaces.HookConfig{
		EnabledHooks: []interfaces.HookType{interfaces.BeforeQueryHook},
		CustomHooks:  make(map[string]interfaces.HookType),
		HookTimeout:  time.Second * 15,
	}
	config := NewConfig(WithHookConfig(hookConfig))
	resultConfig := config.GetHookConfig()
	assert.Equal(t, time.Second*15, resultConfig.HookTimeout)
	assert.Len(t, resultConfig.EnabledHooks, 1)
}

func TestConfigOption_WithEnabledHooks(t *testing.T) {
	config := NewConfig(WithEnabledHooks(interfaces.BeforeQueryHook, interfaces.AfterQueryHook))
	hookConfig := config.GetHookConfig()
	assert.NotNil(t, hookConfig)
	assert.Len(t, hookConfig.EnabledHooks, 2)
	assert.Contains(t, hookConfig.EnabledHooks, interfaces.BeforeQueryHook)
	assert.Contains(t, hookConfig.EnabledHooks, interfaces.AfterQueryHook)
}

func TestConfigOption_WithCustomHook(t *testing.T) {
	config := NewConfig(WithCustomHook("custom", interfaces.BeforeQueryHook))
	hookConfig := config.GetHookConfig()
	assert.NotNil(t, hookConfig)
	assert.Contains(t, hookConfig.CustomHooks, "custom")
	assert.Equal(t, interfaces.BeforeQueryHook, hookConfig.CustomHooks["custom"])
}

func TestConfigOption_WithMultiTenant(t *testing.T) {
	config := NewConfig(WithMultiTenant(true))
	assert.True(t, config.IsMultiTenantEnabled())

	config2 := NewConfig(WithMultiTenant(false))
	assert.False(t, config2.IsMultiTenantEnabled())
}

func TestConfigOption_WithReadReplicas(t *testing.T) {
	replicas := []string{"postgres://replica1:5432/db", "postgres://replica2:5432/db"}
	config := NewConfig(WithReadReplicas(replicas, interfaces.LoadBalanceModeRoundRobin))
	replicaConfig := config.GetReadReplicaConfig()
	assert.NotNil(t, replicaConfig)
	assert.True(t, replicaConfig.Enabled)
	assert.Len(t, replicaConfig.ConnectionStrings, 2)
}

func TestConfigOption_WithFailover(t *testing.T) {
	fallbackNodes := []string{"postgres://fallback1:5432/db", "postgres://fallback2:5432/db"}
	config := NewConfig(WithFailover(fallbackNodes, 3))
	failoverConfig := config.GetFailoverConfig()
	assert.NotNil(t, failoverConfig)
	assert.True(t, failoverConfig.Enabled)
	assert.Len(t, failoverConfig.FallbackNodes, 2)
	assert.Equal(t, 3, failoverConfig.MaxFailoverAttempts)
}

func TestDefaultConfig_WithPoolConfig(t *testing.T) {
	originalConfig := NewDefaultConfig("postgres://localhost:5432/testdb")

	// Create a simple pool config
	poolConfig := interfaces.PoolConfig{
		MaxConns:          100,
		MinConns:          10,
		MaxConnLifetime:   time.Hour * 2,
		MaxConnIdleTime:   time.Minute * 30,
		HealthCheckPeriod: time.Minute * 2,
		ConnectTimeout:    time.Second * 45,
		LazyConnect:       true,
	}

	newConfig := originalConfig.WithPoolConfig(poolConfig)
	assert.NotNil(t, newConfig)
	assert.Equal(t, int32(100), newConfig.GetPoolConfig().MaxConns)
	// Original config should remain unchanged
	assert.NotEqual(t, int32(100), originalConfig.GetPoolConfig().MaxConns)
}

func TestDefaultConfig_WithTLSConfig(t *testing.T) {
	originalConfig := NewDefaultConfig("postgres://localhost:5432/testdb")

	// Create a simple TLS config
	tlsConfig := interfaces.TLSConfig{
		Enabled:            true,
		InsecureSkipVerify: true,
		ServerName:         "test-server",
	}

	newConfig := originalConfig.WithTLSConfig(tlsConfig)
	assert.NotNil(t, newConfig)
	assert.True(t, newConfig.GetTLSConfig().Enabled)
	// Original config should remain unchanged
	assert.False(t, originalConfig.GetTLSConfig().Enabled)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      interfaces.Config
		expectError bool
	}{
		{
			name:        "Valid config",
			config:      NewDefaultConfig("postgres://localhost:5432/testdb"),
			expectError: false,
		},
		{
			name:        "Empty connection string",
			config:      NewDefaultConfig(""),
			expectError: true,
		},
		{
			name:        "Invalid connection string",
			config:      NewDefaultConfig("://invalid-url-with-bad-scheme"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_Clone(t *testing.T) {
	original := NewDefaultConfig("postgres://localhost:5432/testdb")
	cloned := original.Clone()

	assert.NotNil(t, cloned)
	assert.Equal(t, original.GetConnectionString(), cloned.GetConnectionString())

	// Ensure they are different instances
	assert.NotSame(t, original, cloned)
}

func TestConfigOption_WithPoolConfig(t *testing.T) {
	poolConfig := interfaces.PoolConfig{
		MaxConns:          50,
		MinConns:          5,
		MaxConnLifetime:   time.Hour * 3,
		MaxConnIdleTime:   time.Minute * 45,
		HealthCheckPeriod: time.Minute * 3,
		ConnectTimeout:    time.Second * 60,
		LazyConnect:       true,
	}
	config := NewConfig(WithPoolConfig(poolConfig))
	resultConfig := config.GetPoolConfig()
	assert.Equal(t, int32(50), resultConfig.MaxConns)
	assert.Equal(t, int32(5), resultConfig.MinConns)
	assert.True(t, resultConfig.LazyConnect)
}

func TestDefaultConfig_ApplyOptions(t *testing.T) {
	config := NewDefaultConfig("postgres://localhost:5432/testdb")

	err := config.ApplyOptions(
		WithMaxConns(200),
		WithMultiTenant(true),
	)
	assert.NoError(t, err)
	assert.True(t, config.IsMultiTenantEnabled())
}

func TestDefaultConfig_WithRetryConfig(t *testing.T) {
	originalConfig := NewDefaultConfig("postgres://localhost:5432/testdb")

	retryConfig := interfaces.RetryConfig{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 500,
		MaxInterval:     time.Second * 30,
		Multiplier:      2.5,
		RandomizeWait:   true,
	}

	newConfig := originalConfig.WithRetryConfig(retryConfig)
	assert.NotNil(t, newConfig)
	assert.Equal(t, 10, newConfig.GetRetryConfig().MaxRetries)
	assert.Equal(t, 2.5, newConfig.GetRetryConfig().Multiplier)
	// Original config should remain unchanged
	assert.NotEqual(t, 10, originalConfig.GetRetryConfig().MaxRetries)
}

func TestDefaultConfig_WithHookConfig(t *testing.T) {
	originalConfig := NewDefaultConfig("postgres://localhost:5432/testdb")

	hookConfig := interfaces.HookConfig{
		EnabledHooks: []interfaces.HookType{interfaces.BeforeQueryHook, interfaces.AfterQueryHook},
		CustomHooks:  map[string]interfaces.HookType{"test": interfaces.OnErrorHook},
		HookTimeout:  time.Second * 30,
	}

	newConfig := originalConfig.WithHookConfig(hookConfig)
	assert.NotNil(t, newConfig)
	assert.Equal(t, time.Second*30, newConfig.GetHookConfig().HookTimeout)
	assert.Len(t, newConfig.GetHookConfig().EnabledHooks, 2)
	// Original config should remain unchanged
	assert.NotEqual(t, time.Second*30, originalConfig.GetHookConfig().HookTimeout)
}

func TestDefaultConfig_WithReadReplicaConfig(t *testing.T) {
	originalConfig := NewDefaultConfig("postgres://localhost:5432/testdb")

	replicaConfig := interfaces.ReadReplicaConfig{
		Enabled:             true,
		ConnectionStrings:   []string{"postgres://replica1:5432/db", "postgres://replica2:5432/db"},
		LoadBalanceMode:     interfaces.LoadBalanceModeRandom,
		HealthCheckInterval: time.Second * 45,
	}

	newConfig := originalConfig.WithReadReplicaConfig(replicaConfig)
	assert.NotNil(t, newConfig)
	assert.True(t, newConfig.GetReadReplicaConfig().Enabled)
	assert.Len(t, newConfig.GetReadReplicaConfig().ConnectionStrings, 2)
	// Original config should remain unchanged
	assert.False(t, originalConfig.GetReadReplicaConfig().Enabled)
}

func TestDefaultConfig_WithFailoverConfig(t *testing.T) {
	originalConfig := NewDefaultConfig("postgres://localhost:5432/testdb")

	failoverConfig := interfaces.FailoverConfig{
		Enabled:             true,
		FallbackNodes:       []string{"postgres://fallback1:5432/db"},
		HealthCheckInterval: time.Second * 60,
		RetryInterval:       time.Second * 10,
		MaxFailoverAttempts: 5,
	}

	newConfig := originalConfig.WithFailoverConfig(failoverConfig)
	assert.NotNil(t, newConfig)
	assert.True(t, newConfig.GetFailoverConfig().Enabled)
	assert.Equal(t, 5, newConfig.GetFailoverConfig().MaxFailoverAttempts)
	// Original config should remain unchanged
	assert.False(t, originalConfig.GetFailoverConfig().Enabled)
}

func TestConfig_ValidateErrors(t *testing.T) {
	tests := []struct {
		name     string
		config   *DefaultConfig
		expected string
	}{
		{
			name: "MaxConns <= 0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns: 0,
					MinConns: 1,
				},
			},
			expected: "max connections must be positive",
		},
		{
			name: "MinConns < 0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns: 10,
					MinConns: -1,
				},
			},
			expected: "min connections cannot be negative",
		},
		{
			name: "MinConns > MaxConns",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns: 5,
					MinConns: 10,
				},
			},
			expected: "min connections cannot be greater than max connections",
		},
		{
			name: "MaxConnLifetime <= 0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns:        10,
					MinConns:        1,
					MaxConnLifetime: 0,
				},
			},
			expected: "max connection lifetime must be positive",
		},
		{
			name: "MaxConnIdleTime <= 0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns:        10,
					MinConns:        1,
					MaxConnLifetime: time.Minute,
					MaxConnIdleTime: 0,
				},
			},
			expected: "max connection idle time must be positive",
		},
		{
			name: "ConnectTimeout <= 0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns:        10,
					MinConns:        1,
					MaxConnLifetime: time.Minute,
					MaxConnIdleTime: time.Minute,
					ConnectTimeout:  0,
				},
			},
			expected: "connect timeout must be positive",
		},
		{
			name: "MaxRetries < 0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns:        10,
					MinConns:        1,
					MaxConnLifetime: time.Minute,
					MaxConnIdleTime: time.Minute,
					ConnectTimeout:  time.Second * 30,
				},
				retryConfig: interfaces.RetryConfig{
					MaxRetries: -1,
				},
			},
			expected: "max retries cannot be negative",
		},
		{
			name: "InitialInterval <= 0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns:        10,
					MinConns:        1,
					MaxConnLifetime: time.Minute,
					MaxConnIdleTime: time.Minute,
					ConnectTimeout:  time.Second * 30,
				},
				retryConfig: interfaces.RetryConfig{
					MaxRetries:      3,
					InitialInterval: 0,
				},
			},
			expected: "initial interval must be positive",
		},
		{
			name: "MaxInterval <= 0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns:        10,
					MinConns:        1,
					MaxConnLifetime: time.Minute,
					MaxConnIdleTime: time.Minute,
					ConnectTimeout:  time.Second * 30,
				},
				retryConfig: interfaces.RetryConfig{
					MaxRetries:      3,
					InitialInterval: time.Millisecond * 100,
					MaxInterval:     0,
				},
			},
			expected: "max interval must be positive",
		},
		{
			name: "Multiplier <= 1.0",
			config: &DefaultConfig{
				connectionString: "postgres://localhost:5432/db",
				poolConfig: interfaces.PoolConfig{
					MaxConns:        10,
					MinConns:        1,
					MaxConnLifetime: time.Minute,
					MaxConnIdleTime: time.Minute,
					ConnectTimeout:  time.Second * 30,
				},
				retryConfig: interfaces.RetryConfig{
					MaxRetries:      3,
					InitialInterval: time.Millisecond * 100,
					MaxInterval:     time.Second * 5,
					Multiplier:      1.0,
				},
			},
			expected: "multiplier must be greater than 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

func TestNewConfig_WithErrorHandling(t *testing.T) {
	// Test that errors in options are handled gracefully
	// (though current implementation continues on error)
	config := NewConfig(
		WithConnectionString("postgres://localhost:5432/db"),
		// Add a function that would cause error if properly validated
		func(c *DefaultConfig) error {
			// This should be ignored in current implementation
			return fmt.Errorf("test error")
		},
	)
	assert.NotNil(t, config)
	assert.Equal(t, "postgres://localhost:5432/db", config.GetConnectionString())
}

func TestConfigOptions_ErrorCases(t *testing.T) {
	// Test various error cases for config options
	config := NewDefaultConfig("postgres://localhost:5432/db")

	// Test WithMaxConns with negative value - should return error
	err := config.ApplyOptions(WithMaxConns(-1))
	assert.Error(t, err) // ApplyOptions should return error
	assert.Contains(t, err.Error(), "max connections must be positive")

	// Config should remain valid after error
	validateErr := config.Validate()
	assert.NoError(t, validateErr) // Original config should still be valid
}
