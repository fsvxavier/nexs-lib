package postgresql

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestNewPostgreSQLProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerType interfaces.ProviderType
		wantError    bool
		expectedName string
		minFeatures  int
	}{
		{
			name:         "PGX provider",
			providerType: interfaces.ProviderTypePGX,
			wantError:    false,
			expectedName: "postgresql-pgx",
			minFeatures:  10,
		},
		{
			name:         "invalid provider",
			providerType: "invalid",
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewPostgreSQLProvider(tt.providerType)

			if tt.wantError {
				if err == nil {
					t.Error("NewPostgreSQLProvider() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewPostgreSQLProvider() unexpected error = %v", err)
				return
			}

			if provider == nil {
				t.Error("NewPostgreSQLProvider() returned nil provider")
				return
			}

			// Verify provider properties
			if provider.Name() != tt.expectedName {
				t.Errorf("Provider name = %v, want %v", provider.Name(), tt.expectedName)
			}

			if provider.Version() == "" {
				t.Error("Provider version should not be empty")
			}

			features := provider.GetSupportedFeatures()
			if len(features) < tt.minFeatures {
				t.Errorf("Provider should support at least %d features, got %d", tt.minFeatures, len(features))
			}
		})
	}
}

func TestPostgreSQLProvider_SupportsFeature(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Errorf("NewPostgreSQLProvider() error = %v", err)
		return
	}

	tests := []struct {
		name     string
		feature  string
		expected bool
	}{
		{
			name:     "supported feature - connection_pooling",
			feature:  "connection_pooling",
			expected: true,
		},
		{
			name:     "supported feature - transactions",
			feature:  "transactions",
			expected: true,
		},
		{
			name:     "supported feature - listen_notify",
			feature:  "listen_notify",
			expected: true,
		},
		{
			name:     "unsupported feature",
			feature:  "unsupported_feature",
			expected: false,
		},
		{
			name:     "empty feature",
			feature:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.SupportsFeature(tt.feature)
			if result != tt.expected {
				t.Errorf("SupportsFeature(%v) = %v, want %v", tt.feature, result, tt.expected)
			}
		})
	}
}

func TestPostgreSQLProvider_ValidateConfig(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Errorf("NewPostgreSQLProvider() error = %v", err)
		return
	}

	tests := []struct {
		name      string
		config    interfaces.Config
		wantError bool
	}{
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
		},
		{
			name:      "valid config",
			config:    config.NewDefaultConfig("postgres://user:password@localhost:5432/dbname"),
			wantError: false,
		},
		{
			name:      "invalid config - empty connection string",
			config:    config.NewDefaultConfig(""),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateConfig(tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("ValidateConfig() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestPostgreSQLProvider_NewPool(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Errorf("NewPostgreSQLProvider() error = %v", err)
		return
	}

	tests := []struct {
		name      string
		config    interfaces.Config
		wantError bool
	}{
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
		},
		{
			name:      "invalid config",
			config:    config.NewDefaultConfig(""),
			wantError: true,
		},
		{
			name:      "valid config - should create pool successfully",
			config:    config.NewDefaultConfig("postgres://user:password@localhost:5432/dbname"),
			wantError: false, // PGX pool creation is lazy, so this should succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			pool, err := provider.NewPool(ctx, tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("NewPool() expected error but got nil")
				}
				// For now, we expect connection errors to non-existent server
				return
			}

			if err != nil {
				t.Errorf("NewPool() unexpected error = %v", err)
				return
			}

			if pool == nil {
				t.Error("NewPool() returned nil pool")
			}
		})
	}
}

func TestPostgreSQLProvider_NewConn(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Errorf("NewPostgreSQLProvider() error = %v", err)
		return
	}

	tests := []struct {
		name      string
		config    interfaces.Config
		wantError bool
	}{
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
		},
		{
			name:      "invalid config",
			config:    config.NewDefaultConfig(""),
			wantError: true,
		},
		{
			name:      "valid config - should return not implemented error for now",
			config:    config.NewDefaultConfig("postgres://user:password@localhost:5432/dbname"),
			wantError: true, // Since PGX provider is not fully implemented yet
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			conn, err := provider.NewConn(ctx, tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("NewConn() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewConn() unexpected error = %v", err)
				return
			}

			if conn == nil {
				t.Error("NewConn() returned nil connection")
			}
		})
	}
}

func TestPostgreSQLProvider_NewListenConn(t *testing.T) {
	tests := []struct {
		name         string
		providerType interfaces.ProviderType
		config       interfaces.Config
		wantError    bool
		errorMsg     string
	}{
		{
			name:         "PGX provider - should support listen/notify",
			providerType: interfaces.ProviderTypePGX,
			config:       config.NewDefaultConfig("postgres://user:password@localhost:5432/dbname"),
			wantError:    true, // Not implemented yet
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewPostgreSQLProvider(tt.providerType)
			if err != nil {
				t.Errorf("NewPostgreSQLProvider() error = %v", err)
				return
			}

			ctx := context.Background()
			conn, err := provider.NewListenConn(ctx, tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("NewListenConn() expected error but got nil")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("NewListenConn() error = %v, want %v", err.Error(), tt.errorMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewListenConn() unexpected error = %v", err)
				return
			}

			if conn == nil {
				t.Error("NewListenConn() returned nil connection")
			}
		})
	}
}

func TestProviderFactory(t *testing.T) {
	factory := NewProviderFactory()

	if factory == nil {
		t.Error("NewProviderFactory() returned nil")
		return
	}

	t.Run("CreateProvider", func(t *testing.T) {
		provider, err := factory.CreateProvider(interfaces.ProviderTypePGX)
		if err != nil {
			t.Errorf("CreateProvider() error = %v", err)
			return
		}

		if provider == nil {
			t.Error("CreateProvider() returned nil provider")
			return
		}

		// Create same provider again - should return the same instance
		provider2, err := factory.CreateProvider(interfaces.ProviderTypePGX)
		if err != nil {
			t.Errorf("CreateProvider() second call error = %v", err)
			return
		}

		if provider != provider2 {
			t.Error("CreateProvider() should return same instance for same type")
		}
	})

	t.Run("RegisterProvider", func(t *testing.T) {
		customProvider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
		if err != nil {
			t.Errorf("NewPostgreSQLProvider() error = %v", err)
			return
		}

		customType := interfaces.ProviderType("custom")
		err = factory.RegisterProvider(customType, customProvider)
		if err != nil {
			t.Errorf("RegisterProvider() error = %v", err)
			return
		}

		// Verify provider was registered
		retrievedProvider, exists := factory.GetProvider(customType)
		if !exists {
			t.Error("RegisterProvider() provider was not registered")
			return
		}

		if retrievedProvider != customProvider {
			t.Error("RegisterProvider() retrieved provider is not the same instance")
		}
	})

	t.Run("RegisterProvider with nil", func(t *testing.T) {
		err := factory.RegisterProvider("test", nil)
		if err == nil {
			t.Error("RegisterProvider() expected error for nil provider")
		}
	})

	t.Run("ListProviders", func(t *testing.T) {
		providers := factory.ListProviders()
		if len(providers) == 0 {
			t.Error("ListProviders() should return at least one provider")
		}
	})

	t.Run("GetProvider non-existing", func(t *testing.T) {
		_, exists := factory.GetProvider("non-existing")
		if exists {
			t.Error("GetProvider() should return false for non-existing provider")
		}
	})
}

func TestDefaultFactory(t *testing.T) {
	factory := GetDefaultFactory()
	if factory == nil {
		t.Error("GetDefaultFactory() returned nil")
	}
}

func TestQuickFactoryMethods(t *testing.T) {
	t.Run("NewPGXProvider", func(t *testing.T) {
		provider, err := NewPGXProvider()
		if err != nil {
			t.Errorf("NewPGXProvider() error = %v", err)
			return
		}

		if provider == nil {
			t.Error("NewPGXProvider() returned nil")
			return
		}

		if provider.GetDriverName() != "pgx" {
			t.Errorf("NewPGXProvider() driver name = %v, want pgx", provider.GetDriverName())
		}
	})
}

func TestNewDefaultConfig_Integration(t *testing.T) {
	config := NewDefaultConfig("postgres://user:password@localhost:5432/dbname")

	if config == nil {
		t.Error("NewDefaultConfig() returned nil")
		return
	}

	// Test that it implements the interface correctly
	var _ interfaces.Config = config

	// Test validation
	err := config.Validate()
	if err != nil {
		t.Errorf("NewDefaultConfig() validation error = %v", err)
	}
}

func TestConfigOptionFunctions(t *testing.T) {
	config := NewDefaultConfig("postgres://localhost/db").(*config.DefaultConfig)

	// Test that config option functions are available
	err := config.ApplyOptions(
		WithMaxConns(50),
		WithMinConns(5),
		WithMultiTenant(true),
	)

	if err != nil {
		t.Errorf("ApplyOptions() error = %v", err)
	}

	// Verify options were applied
	poolConfig := config.GetPoolConfig()
	if poolConfig.MaxConns != 50 {
		t.Errorf("MaxConns = %v, want 50", poolConfig.MaxConns)
	}

	if poolConfig.MinConns != 5 {
		t.Errorf("MinConns = %v, want 5", poolConfig.MinConns)
	}

	if !config.IsMultiTenantEnabled() {
		t.Error("MultiTenantEnabled should be true")
	}
}
