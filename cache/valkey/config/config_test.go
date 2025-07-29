package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "valkey-go", config.Provider)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 6379, config.Port)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 5*time.Second, config.DialTimeout)
	assert.Equal(t, 3*time.Second, config.ReadTimeout)
	assert.Equal(t, 3*time.Second, config.WriteTimeout)
	assert.Equal(t, 5*time.Minute, config.IdleTimeout)
	assert.Equal(t, 30*time.Second, config.PoolTimeout)
	assert.Equal(t, 5, config.MinIdleConns)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, "", config.URI)
	assert.Equal(t, "", config.KeyPrefix)
	assert.False(t, config.ClusterMode)
	assert.False(t, config.SentinelMode)
	assert.Empty(t, config.Addrs)
	assert.Empty(t, config.SentinelAddrs)
	assert.Equal(t, "", config.SentinelMasterName)
	assert.False(t, config.TLSEnabled)
	assert.Equal(t, "", config.TLSCertFile)
	assert.Equal(t, "", config.TLSKeyFile)
	assert.Equal(t, "", config.TLSCACertFile)
	assert.False(t, config.TLSInsecureSkipVerify)
	assert.True(t, config.CircuitBreakerEnabled)
	assert.Equal(t, 5, config.CircuitBreakerThreshold)
	assert.Equal(t, 30*time.Second, config.CircuitBreakerTimeout)
	assert.Equal(t, 10, config.CircuitBreakerMaxRequests)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.MinRetryBackoff)
	assert.Equal(t, 3*time.Second, config.MaxRetryBackoff)
	assert.True(t, config.HealthCheckEnabled)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
	assert.Equal(t, 5*time.Second, config.HealthCheckTimeout)
	assert.Equal(t, "info", config.LogLevel)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty provider",
			config: &Config{
				Host: "localhost",
				Port: 6379,
			},
			wantErr: true,
			errMsg:  "provider não pode ser vazio",
		},
		{
			name: "invalid provider",
			config: &Config{
				Provider: "invalid",
				Host:     "localhost",
				Port:     6379,
			},
			wantErr: true,
			errMsg:  "provider deve ser 'valkey-go' ou 'valkey-glide'",
		},
		{
			name: "valid valkey-glide provider",
			config: &Config{
				Provider:    "valkey-glide",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty host without URI",
			config: &Config{
				Provider: "valkey-go",
				Port:     6379,
			},
			wantErr: true,
			errMsg:  "host não pode ser vazio quando URI não está especificado",
		},
		{
			name: "valid config with URI",
			config: &Config{
				Provider:    "valkey-go",
				URI:         "valkey://localhost:6379",
				PoolSize:    10,
				DialTimeout: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid port negative",
			config: &Config{
				Provider: "valkey-go",
				Host:     "localhost",
				Port:     -1,
			},
			wantErr: true,
			errMsg:  "port deve estar entre 1 e 65535",
		},
		{
			name: "invalid port zero",
			config: &Config{
				Provider: "valkey-go",
				Host:     "localhost",
				Port:     0,
			},
			wantErr: true,
			errMsg:  "port deve estar entre 1 e 65535",
		},
		{
			name: "invalid port too high",
			config: &Config{
				Provider: "valkey-go",
				Host:     "localhost",
				Port:     65536,
			},
			wantErr: true,
			errMsg:  "port deve estar entre 1 e 65535",
		},
		{
			name: "cluster mode without addrs",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				ClusterMode: true,
			},
			wantErr: true,
			errMsg:  "cluster_addrs não pode ser vazio no modo cluster",
		},
		{
			name: "valid cluster mode",
			config: &Config{
				Provider:     "valkey-go",
				ClusterMode:  true,
				Addrs: []string{"localhost:6379", "localhost:6380"},
				PoolSize:     10,
				DialTimeout:  5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "sentinel mode without addrs",
			config: &Config{
				Provider:     "valkey-go",
				Host:         "localhost",
				Port:         6379,
				SentinelMode: true,
			},
			wantErr: true,
			errMsg:  "sentinel_addrs não pode ser vazio no modo sentinel",
		},
		{
			name: "sentinel mode without master name",
			config: &Config{
				Provider:      "valkey-go",
				Host:          "localhost",
				Port:          6379,
				SentinelMode:  true,
				SentinelAddrs: []string{"localhost:26379"},
			},
			wantErr: true,
			errMsg:  "sentinel_master_name não pode ser vazio no modo sentinel",
		},
		{
			name: "valid sentinel mode",
			config: &Config{
				Provider:           "valkey-go",
				SentinelMode:       true,
				SentinelAddrs:      []string{"localhost:26379", "localhost:26380"},
				SentinelMasterName: "mymaster",
				PoolSize:           10,
				DialTimeout:        5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid pool size zero",
			config: &Config{
				Provider: "valkey-go",
				Host:     "localhost",
				Port:     6379,
				PoolSize: 0,
			},
			wantErr: true,
			errMsg:  "pool_size deve ser maior que 0",
		},
		{
			name: "invalid pool size negative",
			config: &Config{
				Provider: "valkey-go",
				Host:     "localhost",
				Port:     6379,
				PoolSize: -1,
			},
			wantErr: true,
			errMsg:  "pool_size deve ser maior que 0",
		},
		{
			name: "negative min idle conns",
			config: &Config{
				Provider:     "valkey-go",
				Host:         "localhost",
				Port:         6379,
				PoolSize:     10,
				MinIdleConns: -1,
			},
			wantErr: true,
			errMsg:  "min_idle_conns não pode ser negativo",
		},
		{
			name: "min idle conns greater than pool size",
			config: &Config{
				Provider:     "valkey-go",
				Host:         "localhost",
				Port:         6379,
				PoolSize:     10,
				MinIdleConns: 15,
				DialTimeout:  5 * time.Second,
			},
			wantErr: true,
			errMsg:  "min_idle_conns não pode ser maior que pool_size",
		},
		{
			name: "zero dial timeout",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: 0,
			},
			wantErr: true,
			errMsg:  "dial_timeout deve ser maior que 0",
		},
		{
			name: "negative dial timeout",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "dial_timeout deve ser maior que 0",
		},
		{
			name: "TLS cert without key",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: 5 * time.Second,
				TLSEnabled:  true,
				TLSCertFile: "/path/to/cert.pem",
			},
			wantErr: true,
			errMsg:  "tls_key_file deve ser especificado quando tls_cert_file está definido",
		},
		{
			name: "TLS key without cert",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: 5 * time.Second,
				TLSEnabled:  true,
				TLSKeyFile:  "/path/to/key.pem",
			},
			wantErr: true,
			errMsg:  "tls_cert_file deve ser especificado quando tls_key_file está definido",
		},
		{
			name: "valid TLS config",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: 5 * time.Second,
				TLSEnabled:  true,
				TLSCertFile: "/path/to/cert.pem",
				TLSKeyFile:  "/path/to/key.pem",
				TLSCACertFile:   "/path/to/ca.pem",
			},
			wantErr: false,
		},
		{
			name: "circuit breaker zero threshold",
			config: &Config{
				Provider:                "valkey-go",
				Host:                    "localhost",
				Port:                    6379,
				PoolSize:                10,
				DialTimeout:             5 * time.Second,
				CircuitBreakerEnabled:   true,
				CircuitBreakerThreshold: 0,
			},
			wantErr: true,
			errMsg:  "circuit_breaker_threshold deve ser maior que 0",
		},
		{
			name: "circuit breaker negative threshold",
			config: &Config{
				Provider:                "valkey-go",
				Host:                    "localhost",
				Port:                    6379,
				PoolSize:                10,
				DialTimeout:             5 * time.Second,
				CircuitBreakerEnabled:   true,
				CircuitBreakerThreshold: -1,
			},
			wantErr: true,
			errMsg:  "circuit_breaker_threshold deve ser maior que 0",
		},
		{
			name: "circuit breaker zero timeout",
			config: &Config{
				Provider:                "valkey-go",
				Host:                    "localhost",
				Port:                    6379,
				PoolSize:                10,
				DialTimeout:             5 * time.Second,
				CircuitBreakerEnabled:   true,
				CircuitBreakerThreshold: 5,
				CircuitBreakerTimeout:   0,
			},
			wantErr: true,
			errMsg:  "circuit_breaker_timeout deve ser maior que 0",
		},
		{
			name: "circuit breaker zero max requests",
			config: &Config{
				Provider:                  "valkey-go",
				Host:                      "localhost",
				Port:                      6379,
				PoolSize:                  10,
				DialTimeout:               5 * time.Second,
				CircuitBreakerEnabled:     true,
				CircuitBreakerThreshold:   5,
				CircuitBreakerTimeout:     30 * time.Second,
				CircuitBreakerMaxRequests: 0,
			},
			wantErr: true,
			errMsg:  "circuit_breaker_max_requests deve ser maior que 0",
		},
		{
			name: "negative max retries",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: 5 * time.Second,
				MaxRetries:  -1,
			},
			wantErr: true,
			errMsg:  "max_retries não pode ser negativo",
		},
		{
			name: "zero min retry backoff",
			config: &Config{
				Provider:        "valkey-go",
				Host:            "localhost",
				Port:            6379,
				PoolSize:        10,
				DialTimeout:     5 * time.Second,
				MaxRetries:      3,
				MinRetryBackoff: 0,
			},
			wantErr: true,
			errMsg:  "min_retry_backoff deve ser maior que 0",
		},
		{
			name: "zero max retry backoff",
			config: &Config{
				Provider:        "valkey-go",
				Host:            "localhost",
				Port:            6379,
				PoolSize:        10,
				DialTimeout:     5 * time.Second,
				MaxRetries:      3,
				MinRetryBackoff: 100 * time.Millisecond,
				MaxRetryBackoff: 0,
			},
			wantErr: true,
			errMsg:  "max_retry_backoff deve ser maior que 0",
		},
		{
			name: "min retry backoff greater than max",
			config: &Config{
				Provider:        "valkey-go",
				Host:            "localhost",
				Port:            6379,
				PoolSize:        10,
				DialTimeout:     5 * time.Second,
				MaxRetries:      3,
				MinRetryBackoff: 5 * time.Second,
				MaxRetryBackoff: 1 * time.Second,
			},
			wantErr: true,
			errMsg:  "min_retry_backoff não pode ser maior que max_retry_backoff",
		},
		{
			name: "zero health check interval",
			config: &Config{
				Provider:            "valkey-go",
				Host:                "localhost",
				Port:                6379,
				PoolSize:            10,
				DialTimeout:         5 * time.Second,
				HealthCheckEnabled:  true,
				HealthCheckInterval: 0,
			},
			wantErr: true,
			errMsg:  "health_check_interval deve ser maior que 0",
		},
		{
			name: "zero health check timeout",
			config: &Config{
				Provider:            "valkey-go",
				Host:                "localhost",
				Port:                6379,
				PoolSize:            10,
				DialTimeout:         5 * time.Second,
				HealthCheckEnabled:  true,
				HealthCheckInterval: 30 * time.Second,
				HealthCheckTimeout:  0,
			},
			wantErr: true,
			errMsg:  "health_check_timeout deve ser maior que 0",
		},
		{
			name: "negative DB index",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: 5 * time.Second,
				DB:          -1,
			},
			wantErr: true,
			errMsg:  "db não pode ser negativo",
		},
		{
			name: "DB index too high",
			config: &Config{
				Provider:    "valkey-go",
				Host:        "localhost",
				Port:        6379,
				PoolSize:    10,
				DialTimeout: 5 * time.Second,
				DB:          16,
			},
			wantErr: true,
			errMsg:  "db deve estar entre 0 e 15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Backup original env vars
	origVars := map[string]string{
		"VALKEY_PROVIDER":     os.Getenv("VALKEY_PROVIDER"),
		"VALKEY_HOST":         os.Getenv("VALKEY_HOST"),
		"VALKEY_PORT":         os.Getenv("VALKEY_PORT"),
		"VALKEY_PASSWORD":     os.Getenv("VALKEY_PASSWORD"),
		"VALKEY_DB":           os.Getenv("VALKEY_DB"),
		"VALKEY_URI":          os.Getenv("VALKEY_URI"),
		"VALKEY_POOL_SIZE":    os.Getenv("VALKEY_POOL_SIZE"),
		"VALKEY_CLUSTER_MODE": os.Getenv("VALKEY_CLUSTER_MODE"),
		"VALKEY_TLS_ENABLED":  os.Getenv("VALKEY_TLS_ENABLED"),
		"VALKEY_LOG_LEVEL":    os.Getenv("VALKEY_LOG_LEVEL"),
	}

	// Clean env vars
	for key := range origVars {
		os.Unsetenv(key)
	}

	// Restore env vars after test
	defer func() {
		for key, value := range origVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("default values when no env vars", func(t *testing.T) {
		config := LoadFromEnv()
		defaultConfig := DefaultConfig()

		assert.Equal(t, defaultConfig.Provider, config.Provider)
		assert.Equal(t, defaultConfig.Host, config.Host)
		assert.Equal(t, defaultConfig.Port, config.Port)
		assert.Equal(t, defaultConfig.PoolSize, config.PoolSize)
	})

	t.Run("load from env vars", func(t *testing.T) {
		os.Setenv("VALKEY_PROVIDER", "valkey-glide")
		os.Setenv("VALKEY_HOST", "redis.example.com")
		os.Setenv("VALKEY_PORT", "6380")
		os.Setenv("VALKEY_PASSWORD", "secret")
		os.Setenv("VALKEY_DB", "2")
		os.Setenv("VALKEY_URI", "redis://user:pass@localhost:6379/1")
		os.Setenv("VALKEY_POOL_SIZE", "20")
		os.Setenv("VALKEY_CLUSTER_MODE", "true")
		os.Setenv("VALKEY_TLS_ENABLED", "true")
		os.Setenv("VALKEY_LOG_LEVEL", "debug")

		config := LoadFromEnv()

		assert.Equal(t, "valkey-glide", config.Provider)
		assert.Equal(t, "redis.example.com", config.Host)
		assert.Equal(t, 6380, config.Port)
		assert.Equal(t, "secret", config.Password)
		assert.Equal(t, 2, config.DB)
		assert.Equal(t, "redis://user:pass@localhost:6379/1", config.URI)
		assert.Equal(t, 20, config.PoolSize)
		assert.True(t, config.ClusterMode)
		assert.True(t, config.TLSEnabled)
		assert.Equal(t, "debug", config.LogLevel)
	})

	t.Run("invalid env values use defaults", func(t *testing.T) {
		os.Setenv("VALKEY_PORT", "invalid")
		os.Setenv("VALKEY_DB", "invalid")
		os.Setenv("VALKEY_POOL_SIZE", "invalid")

		config := LoadFromEnv()
		defaultConfig := DefaultConfig()

		assert.Equal(t, defaultConfig.Port, config.Port)
		assert.Equal(t, defaultConfig.DB, config.DB)
		assert.Equal(t, defaultConfig.PoolSize, config.PoolSize)
	})
}

func TestConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "URI takes precedence",
			config: &Config{
				URI:      "redis://user:pass@localhost:6379/1",
				Host:     "other.host.com",
				Port:     6380,
				Password: "other",
				DB:       2,
			},
			expected: "redis://user:pass@localhost:6379/1",
		},
		{
			name: "with password",
			config: &Config{
				Host:     "localhost",
				Port:     6379,
				Password: "secret",
				DB:       1,
			},
			expected: "valkey://:secret@localhost:6379/1",
		},
		{
			name: "without password",
			config: &Config{
				Host: "localhost",
				Port: 6379,
				DB:   0,
			},
			expected: "valkey://localhost:6379/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ConnectionString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_Copy(t *testing.T) {
	original := &Config{
		Provider:      "valkey-go",
		Host:          "localhost",
		Port:          6379,
		Addrs:         []string{"addr1", "addr2"},
		SentinelAddrs: []string{"sentinel1", "sentinel2"},
	}

	copy := original.Copy()

	// Test that it's a different object
	assert.NotSame(t, original, copy)

	// Test that values are equal
	assert.Equal(t, original.Provider, copy.Provider)
	assert.Equal(t, original.Host, copy.Host)
	assert.Equal(t, original.Port, copy.Port)

	// Test that slices are different objects but equal values
	assert.NotSame(t, original.Addrs, copy.Addrs)
	assert.Equal(t, original.Addrs, copy.Addrs)
	assert.NotSame(t, original.SentinelAddrs, copy.SentinelAddrs)
	assert.Equal(t, original.SentinelAddrs, copy.SentinelAddrs)

	// Test that modifying copy doesn't affect original
	copy.Provider = "valkey-glide"
	copy.Addrs[0] = "modified"

	assert.Equal(t, "valkey-go", original.Provider)
	assert.Equal(t, "addr1", original.Addrs[0])
}

func TestConfig_WithMethods(t *testing.T) {
	config := DefaultConfig()

	result := config.
		WithProvider("valkey-glide").
		WithHost("example.com").
		WithPort(6380).
		WithPassword("secret").
		WithDB(1).
		WithURI("redis://localhost:6379").
		WithPoolSize(20).
		WithKeyPrefix("test:").
		WithClusterMode(true).
		WithClusterAddrs([]string{"addr1", "addr2"}).
		WithSentinelMode(true).
		WithSentinelAddrs([]string{"sentinel1"}).
		WithSentinelMasterName("master").
		WithTLS(true).
		WithCircuitBreaker(false).
		WithHealthCheck(false)

	assert.Equal(t, "valkey-glide", result.Provider)
	assert.Equal(t, "example.com", result.Host)
	assert.Equal(t, 6380, result.Port)
	assert.Equal(t, "secret", result.Password)
	assert.Equal(t, 1, result.DB)
	assert.Equal(t, "redis://localhost:6379", result.URI)
	assert.Equal(t, 20, result.PoolSize)
	assert.Equal(t, "test:", result.KeyPrefix)
	assert.True(t, result.ClusterMode)
	assert.Equal(t, []string{"addr1", "addr2"}, result.Addrs)
	assert.True(t, result.SentinelMode)
	assert.Equal(t, []string{"sentinel1"}, result.SentinelAddrs)
	assert.Equal(t, "master", result.SentinelMasterName)
	assert.True(t, result.TLSEnabled)
	assert.False(t, result.CircuitBreakerEnabled)
	assert.False(t, result.HealthCheckEnabled)

	// Test that methods return the same instance for chaining
	assert.Same(t, config, result)
}

func TestConfig_Validate_EdgeCases(t *testing.T) {
	t.Run("port edge cases", func(t *testing.T) {
		config := DefaultConfig()

		// Test port 1 (valid)
		config.Port = 1
		assert.NoError(t, config.Validate())

		// Test port 65535 (valid)
		config.Port = 65535
		assert.NoError(t, config.Validate())

		// Test port 65536 (invalid)
		config.Port = 65536
		assert.Error(t, config.Validate())
	})

	t.Run("valid cluster config", func(t *testing.T) {
		config := DefaultConfig()
		config.ClusterMode = true
		config.Addrs = []string{"localhost:6379", "localhost:6380"}

		assert.NoError(t, config.Validate())
	})

	t.Run("valid sentinel config", func(t *testing.T) {
		config := DefaultConfig()
		config.SentinelMode = true
		config.SentinelAddrs = []string{"localhost:26379"}
		config.SentinelMasterName = "mymaster"

		assert.NoError(t, config.Validate())
	})

	t.Run("valid TLS config", func(t *testing.T) {
		config := DefaultConfig()
		config.TLSEnabled = true
		config.TLSCertFile = "/path/to/cert.pem"
		config.TLSKeyFile = "/path/to/key.pem"

		assert.NoError(t, config.Validate())
	})

	t.Run("URI overrides host/port validation", func(t *testing.T) {
		config := &Config{
			Provider:     "valkey-go",
			URI:          "redis://localhost:6379",
			Host:         "", // Empty host should be OK when URI is set
			Port:         0,  // Invalid port should be OK when URI is set
			PoolSize:     10,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}

		assert.NoError(t, config.Validate())
	})
}
