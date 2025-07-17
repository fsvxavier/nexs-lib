//go:build unit
// +build unit

package pgx

import (
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestPGXProvider_ComprehensiveFunctionality(t *testing.T) {
	provider := NewPGXProvider()

	t.Run("Provider Information", func(t *testing.T) {
		if provider.Name() != "PGX" {
			t.Errorf("Expected name 'PGX', got %s", provider.Name())
		}

		if provider.Version() != "5.x" {
			t.Errorf("Expected version '5.x', got %s", provider.Version())
		}

		if provider.GetDriverName() != "pgx" {
			t.Errorf("Expected driver name 'pgx', got %s", provider.GetDriverName())
		}
	})

	t.Run("Feature Support", func(t *testing.T) {
		requiredFeatures := []string{
			"transactions",
			"prepared_statements",
			"batch_operations",
			"connection_pooling",
			"hooks",
		}

		for _, feature := range requiredFeatures {
			if !provider.SupportsFeature(feature) {
				t.Errorf("Provider should support feature: %s", feature)
			}
		}

		// Test unsupported feature
		if provider.SupportsFeature("unsupported_feature") {
			t.Error("Provider should not support 'unsupported_feature'")
		}
	})

	t.Run("Configuration Validation", func(t *testing.T) {
		tests := []struct {
			name      string
			config    interfaces.Config
			wantError bool
		}{
			{
				name:      "valid config",
				config:    config.NewDefaultConfig("postgres://user:pass@localhost/db"),
				wantError: false,
			},
			{
				name:      "empty connection string",
				config:    config.NewDefaultConfig(""),
				wantError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := provider.ValidateConfig(tt.config)
				if (err != nil) != tt.wantError {
					t.Errorf("ValidateConfig() error = %v, wantError %v", err, tt.wantError)
				}
			})
		}
	})

	t.Run("Supported Features List", func(t *testing.T) {
		features := provider.GetSupportedFeatures()
		if len(features) == 0 {
			t.Error("GetSupportedFeatures() should return non-empty list")
		}

		// Verify no duplicates
		featureMap := make(map[string]bool)
		for _, feature := range features {
			if featureMap[feature] {
				t.Errorf("Duplicate feature found: %s", feature)
			}
			featureMap[feature] = true
		}
	})
}

func TestPGXProvider_EdgeCases(t *testing.T) {
	provider := NewPGXProvider()

	t.Run("Empty Feature Check", func(t *testing.T) {
		if provider.SupportsFeature("") {
			t.Error("Provider should not support empty feature")
		}
	})

	t.Run("Nil Config Validation", func(t *testing.T) {
		err := provider.ValidateConfig(nil)
		if err == nil {
			t.Error("ValidateConfig() should return error for nil config")
		}
	})

	t.Run("Multiple Feature Checks", func(t *testing.T) {
		features := []string{
			"transactions",
			"batch_operations",
			"unsupported1",
			"connection_pooling",
			"unsupported2",
		}

		supportedCount := 0
		for _, feature := range features {
			if provider.SupportsFeature(feature) {
				supportedCount++
			}
		}

		if supportedCount == 0 {
			t.Error("At least some features should be supported")
		}

		if supportedCount == len(features) {
			t.Error("Not all features should be supported (some are intentionally unsupported)")
		}
	})
}

func TestPGXProvider_ConfigValidation(t *testing.T) {
	provider := NewPGXProvider()

	t.Run("Invalid Pool Configuration", func(t *testing.T) {
		cfg := config.NewDefaultConfig("postgres://localhost/test")

		// Create a config with invalid pool settings
		cfg = cfg.WithPoolConfig(interfaces.PoolConfig{
			MaxConns:        -1, // Invalid
			MinConns:        0,
			MaxConnLifetime: time.Minute,
			MaxConnIdleTime: time.Minute,
			ConnectTimeout:  time.Second * 30,
		})

		err := provider.ValidateConfig(cfg)
		if err == nil {
			t.Error("ValidateConfig() should return error for invalid pool config")
		}
	})

	t.Run("Invalid Retry Configuration", func(t *testing.T) {
		cfg := config.NewDefaultConfig("postgres://localhost/test")

		cfg = cfg.WithRetryConfig(interfaces.RetryConfig{
			MaxRetries:      -1, // Invalid
			InitialInterval: time.Millisecond * 100,
			MaxInterval:     time.Second * 5,
			Multiplier:      2.0,
		})

		err := provider.ValidateConfig(cfg)
		if err == nil {
			t.Error("ValidateConfig() should return error for invalid retry config")
		}
	})
}

func TestPGXProvider_ThreadSafety(t *testing.T) {
	provider := NewPGXProvider()

	t.Run("Concurrent Feature Checks", func(t *testing.T) {
		done := make(chan bool)

		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				for j := 0; j < 100; j++ {
					provider.SupportsFeature("transactions")
					provider.Name()
					provider.Version()
					provider.GetDriverName()
					provider.GetSupportedFeatures()
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			select {
			case <-done:
			case <-time.After(time.Second * 5):
				t.Error("Timeout waiting for concurrent operations")
				return
			}
		}
	})
}

func TestPGXProvider_Performance(t *testing.T) {
	provider := NewPGXProvider()

	t.Run("Feature Check Performance", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < 10000; i++ {
			provider.SupportsFeature("transactions")
		}

		duration := time.Since(start)

		// Should complete 10k feature checks in under 100ms
		if duration > time.Millisecond*100 {
			t.Errorf("Feature checks too slow: %v", duration)
		}
	})

	t.Run("Get Supported Features Performance", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			features := provider.GetSupportedFeatures()
			if len(features) == 0 {
				t.Error("No features returned")
			}
		}

		duration := time.Since(start)

		// Should complete 1k calls in under 50ms
		if duration > time.Millisecond*50 {
			t.Errorf("GetSupportedFeatures too slow: %v", duration)
		}
	})
}
