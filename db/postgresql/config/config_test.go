package config

import (
	"crypto/tls"
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestNewDefaultConfig(t *testing.T) {
	tests := []struct {
		name             string
		connectionString string
		wantError        bool
	}{
		{
			name:             "valid connection string",
			connectionString: "postgres://user:password@localhost:5432/dbname",
			wantError:        false,
		},
		{
			name:             "empty connection string",
			connectionString: "",
			wantError:        false, // Config creation should succeed, validation should fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig(tt.connectionString)

			if config == nil {
				t.Error("NewDefaultConfig() returned nil")
				return
			}

			if config.GetConnectionString() != tt.connectionString {
				t.Errorf("GetConnectionString() = %v, want %v", config.GetConnectionString(), tt.connectionString)
			}

			// Verify default values
			poolConfig := config.GetPoolConfig()
			if poolConfig.MaxConns != 40 {
				t.Errorf("Default MaxConns = %v, want 40", poolConfig.MaxConns)
			}

			if poolConfig.MinConns != 2 {
				t.Errorf("Default MinConns = %v, want 2", poolConfig.MinConns)
			}

			if poolConfig.MaxConnLifetime != time.Minute*30 {
				t.Errorf("Default MaxConnLifetime = %v, want %v", poolConfig.MaxConnLifetime, time.Minute*30)
			}
		})
	}
}

func TestDefaultConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *DefaultConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			setupFunc: func() *DefaultConfig {
				return NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
			},
			wantError: false,
		},
		{
			name: "empty connection string",
			setupFunc: func() *DefaultConfig {
				return NewDefaultConfig("")
			},
			wantError: true,
			errorMsg:  "connection string cannot be empty",
		},
		{
			name: "invalid max connections",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.poolConfig.MaxConns = 0
				return config
			},
			wantError: true,
			errorMsg:  "max connections must be positive",
		},
		{
			name: "negative min connections",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.poolConfig.MinConns = -1
				return config
			},
			wantError: true,
			errorMsg:  "min connections cannot be negative",
		},
		{
			name: "min connections greater than max",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.poolConfig.MinConns = 10
				config.poolConfig.MaxConns = 5
				return config
			},
			wantError: true,
			errorMsg:  "min connections cannot be greater than max connections",
		},
		{
			name: "invalid max connection lifetime",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.poolConfig.MaxConnLifetime = 0
				return config
			},
			wantError: true,
			errorMsg:  "max connection lifetime must be positive",
		},
		{
			name: "invalid max connection idle time",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.poolConfig.MaxConnIdleTime = 0
				return config
			},
			wantError: true,
			errorMsg:  "max connection idle time must be positive",
		},
		{
			name: "invalid connect timeout",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.poolConfig.ConnectTimeout = 0
				return config
			},
			wantError: true,
			errorMsg:  "connect timeout must be positive",
		},
		{
			name: "negative max retries",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.retryConfig.MaxRetries = -1
				return config
			},
			wantError: true,
			errorMsg:  "max retries cannot be negative",
		},
		{
			name: "invalid initial interval",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.retryConfig.InitialInterval = 0
				return config
			},
			wantError: true,
			errorMsg:  "initial interval must be positive",
		},
		{
			name: "invalid max interval",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.retryConfig.MaxInterval = 0
				return config
			},
			wantError: true,
			errorMsg:  "max interval must be positive",
		},
		{
			name: "invalid multiplier",
			setupFunc: func() *DefaultConfig {
				config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")
				config.retryConfig.Multiplier = 1.0
				return config
			},
			wantError: true,
			errorMsg:  "multiplier must be greater than 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupFunc()
			err := config.Validate()

			if tt.wantError {
				if err == nil {
					t.Error("Validate() expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfigOptions(t *testing.T) {
	t.Run("WithConnectionString", func(t *testing.T) {
		config := NewDefaultConfig("initial")
		newConnStr := "postgres://user:password@localhost:5432/newdb"

		err := config.ApplyOptions(WithConnectionString(newConnStr))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		if config.GetConnectionString() != newConnStr {
			t.Errorf("GetConnectionString() = %v, want %v", config.GetConnectionString(), newConnStr)
		}
	})

	t.Run("WithConnectionString empty", func(t *testing.T) {
		config := NewDefaultConfig("initial")

		err := config.ApplyOptions(WithConnectionString(""))
		if err == nil {
			t.Error("ApplyOptions() expected error for empty connection string")
		}
	})

	t.Run("WithMaxConns", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")

		err := config.ApplyOptions(WithMaxConns(100))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		if config.GetPoolConfig().MaxConns != 100 {
			t.Errorf("MaxConns = %v, want 100", config.GetPoolConfig().MaxConns)
		}
	})

	t.Run("WithMaxConns invalid", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")

		err := config.ApplyOptions(WithMaxConns(0))
		if err == nil {
			t.Error("ApplyOptions() expected error for invalid max connections")
		}
	})

	t.Run("WithMinConns", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")

		err := config.ApplyOptions(WithMinConns(5))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		if config.GetPoolConfig().MinConns != 5 {
			t.Errorf("MinConns = %v, want 5", config.GetPoolConfig().MinConns)
		}
	})

	t.Run("WithMaxConnLifetime", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")
		lifetime := time.Hour

		err := config.ApplyOptions(WithMaxConnLifetime(lifetime))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		if config.GetPoolConfig().MaxConnLifetime != lifetime {
			t.Errorf("MaxConnLifetime = %v, want %v", config.GetPoolConfig().MaxConnLifetime, lifetime)
		}
	})

	t.Run("WithTLS", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")
		tlsConfig := &tls.Config{InsecureSkipVerify: true, ServerName: "test"}

		err := config.ApplyOptions(WithTLS(true, tlsConfig))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		tls := config.GetTLSConfig()
		if !tls.Enabled {
			t.Error("TLS should be enabled")
		}
		if !tls.InsecureSkipVerify {
			t.Error("InsecureSkipVerify should be true")
		}
		if tls.ServerName != "test" {
			t.Errorf("ServerName = %v, want test", tls.ServerName)
		}
	})

	t.Run("WithMultiTenant", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")

		err := config.ApplyOptions(WithMultiTenant(true))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		if !config.IsMultiTenantEnabled() {
			t.Error("MultiTenantEnabled should be true")
		}
	})

	t.Run("WithEnabledHooks", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")
		hooks := []interfaces.HookType{interfaces.BeforeQueryHook, interfaces.AfterQueryHook}

		err := config.ApplyOptions(WithEnabledHooks(hooks...))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		enabledHooks := config.GetHookConfig().EnabledHooks
		if len(enabledHooks) != 2 {
			t.Errorf("EnabledHooks length = %v, want 2", len(enabledHooks))
		}
	})

	t.Run("WithCustomHook", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")

		err := config.ApplyOptions(WithCustomHook("test_hook", interfaces.CustomHookBase+1))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		customHooks := config.GetHookConfig().CustomHooks
		if len(customHooks) != 1 {
			t.Errorf("CustomHooks length = %v, want 1", len(customHooks))
		}
		if customHooks["test_hook"] != interfaces.CustomHookBase+1 {
			t.Errorf("CustomHooks[test_hook] = %v, want %v", customHooks["test_hook"], interfaces.CustomHookBase+1)
		}
	})

	t.Run("WithReadReplicas", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")
		replicas := []string{"postgres://replica1/db", "postgres://replica2/db"}

		err := config.ApplyOptions(WithReadReplicas(replicas, interfaces.LoadBalanceModeRandom))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		replicaConfig := config.GetReadReplicaConfig()
		if !replicaConfig.Enabled {
			t.Error("ReadReplica should be enabled")
		}
		if len(replicaConfig.ConnectionStrings) != 2 {
			t.Errorf("ConnectionStrings length = %v, want 2", len(replicaConfig.ConnectionStrings))
		}
		if replicaConfig.LoadBalanceMode != interfaces.LoadBalanceModeRandom {
			t.Errorf("LoadBalanceMode = %v, want %v", replicaConfig.LoadBalanceMode, interfaces.LoadBalanceModeRandom)
		}
	})

	t.Run("WithFailover", func(t *testing.T) {
		config := NewDefaultConfig("postgres://localhost/db")
		fallbackNodes := []string{"postgres://fallback1/db", "postgres://fallback2/db"}

		err := config.ApplyOptions(WithFailover(fallbackNodes, 5))
		if err != nil {
			t.Errorf("ApplyOptions() error = %v", err)
			return
		}

		failoverConfig := config.GetFailoverConfig()
		if !failoverConfig.Enabled {
			t.Error("Failover should be enabled")
		}
		if len(failoverConfig.FallbackNodes) != 2 {
			t.Errorf("FallbackNodes length = %v, want 2", len(failoverConfig.FallbackNodes))
		}
		if failoverConfig.MaxFailoverAttempts != 5 {
			t.Errorf("MaxFailoverAttempts = %v, want 5", failoverConfig.MaxFailoverAttempts)
		}
	})
}

func TestConfigClone(t *testing.T) {
	originalConfig := NewDefaultConfig("postgres://localhost/db")

	// Apply some options to the original
	err := originalConfig.ApplyOptions(
		WithMaxConns(50),
		WithMinConns(5),
		WithMultiTenant(true),
		WithCustomHook("test_hook", interfaces.CustomHookBase+1),
	)
	if err != nil {
		t.Errorf("ApplyOptions() error = %v", err)
		return
	}

	// Clone the config
	clonedConfig := originalConfig.Clone()

	// Verify that clone has same values
	if clonedConfig.GetConnectionString() != originalConfig.GetConnectionString() {
		t.Error("Cloned config has different connection string")
	}

	if clonedConfig.GetPoolConfig().MaxConns != originalConfig.GetPoolConfig().MaxConns {
		t.Error("Cloned config has different MaxConns")
	}

	if clonedConfig.IsMultiTenantEnabled() != originalConfig.IsMultiTenantEnabled() {
		t.Error("Cloned config has different MultiTenantEnabled")
	}

	// Verify that custom hooks are properly cloned
	originalCustomHooks := originalConfig.GetHookConfig().CustomHooks
	clonedCustomHooks := clonedConfig.GetHookConfig().CustomHooks

	if len(clonedCustomHooks) != len(originalCustomHooks) {
		t.Error("Cloned config has different custom hooks length")
	}

	// Modify clone and verify original is not affected
	clonedConfig.poolConfig.MaxConns = 100

	if originalConfig.GetPoolConfig().MaxConns == 100 {
		t.Error("Modifying clone affected original config")
	}
}
