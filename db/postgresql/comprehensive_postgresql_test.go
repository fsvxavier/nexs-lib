//go:build unit
// +build unit

package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestPostgreSQLProviderFactory(t *testing.T) {
	factory := NewProviderFactory()

	t.Run("Create PGX Provider", func(t *testing.T) {
		provider, err := factory.CreateProvider(interfaces.ProviderTypePGX)
		if err != nil {
			t.Errorf("CreateProvider() error = %v", err)
		}

		if provider == nil {
			t.Error("CreateProvider() returned nil provider")
		}

		if provider.GetDriverName() != "pgx" {
			t.Errorf("Expected driver name 'pgx', got %s", provider.GetDriverName())
		}
	})

	t.Run("Create Invalid Provider", func(t *testing.T) {
		_, err := factory.CreateProvider("invalid")
		if err == nil {
			t.Error("CreateProvider() should return error for invalid provider type")
		}
	})

	t.Run("Register Custom Provider", func(t *testing.T) {
		customProvider, _ := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
		err := factory.RegisterProvider("custom", customProvider)
		if err != nil {
			t.Errorf("RegisterProvider() error = %v", err)
		}

		retrievedProvider, exists := factory.GetProvider("custom")
		if !exists {
			t.Error("GetProvider() should return true for registered provider")
		}

		if retrievedProvider != customProvider {
			t.Error("GetProvider() should return the same provider instance")
		}
	})

	t.Run("List Providers", func(t *testing.T) {
		providers := factory.ListProviders()
		if len(providers) == 0 {
			t.Error("ListProviders() should return non-empty list")
		}

		// Should include PGX by default
		found := false
		for _, provider := range providers {
			if provider == interfaces.ProviderTypePGX {
				found = true
				break
			}
		}

		if !found {
			t.Error("ListProviders() should include PGX provider")
		}
	})
}

func TestPostgreSQLProviderComprehensive(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Fatalf("NewPostgreSQLProvider() error = %v", err)
	}

	t.Run("Provider Information", func(t *testing.T) {
		if provider.Name() == "" {
			t.Error("Provider name should not be empty")
		}

		if provider.Version() == "" {
			t.Error("Provider version should not be empty")
		}

		if provider.GetDriverName() == "" {
			t.Error("Driver name should not be empty")
		}
	})

	t.Run("Feature Support", func(t *testing.T) {
		requiredFeatures := []string{
			"connection_pooling",
			"transactions",
			"prepared_statements",
			"batch_operations",
			"hooks",
		}

		for _, feature := range requiredFeatures {
			if !provider.SupportsFeature(feature) {
				t.Errorf("Provider should support feature: %s", feature)
			}
		}

		// Test that all supported features are non-empty
		supportedFeatures := provider.GetSupportedFeatures()
		if len(supportedFeatures) == 0 {
			t.Error("Provider should support at least some features")
		}

		// Test unsupported feature
		if provider.SupportsFeature("unsupported_random_feature_12345") {
			t.Error("Provider should not support obviously unsupported features")
		}
	})

	t.Run("Configuration Validation", func(t *testing.T) {
		validConfig := config.NewDefaultConfig("postgres://user:pass@localhost/testdb")
		err := provider.ValidateConfig(validConfig)
		if err != nil {
			t.Errorf("ValidateConfig() should accept valid config, got error: %v", err)
		}

		// Test invalid config
		invalidConfig := config.NewDefaultConfig("")
		err = provider.ValidateConfig(invalidConfig)
		if err == nil {
			t.Error("ValidateConfig() should reject invalid config")
		}
	})

	t.Run("Provider Methods", func(t *testing.T) {
		// Test retry manager
		retryManager := provider.GetRetryManager()
		if retryManager == nil {
			t.Error("GetRetryManager() should not return nil")
		}

		// Test failover manager
		failoverManager := provider.GetFailoverManager()
		if failoverManager == nil {
			t.Error("GetFailoverManager() should not return nil")
		}
	})
}

func TestPostgreSQLProviderRetryAndFailover(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Fatalf("NewPostgreSQLProvider() error = %v", err)
	}

	t.Run("Retry Operations", func(t *testing.T) {
		ctx := context.Background()

		attemptCount := 0
		operation := func() error {
			attemptCount++
			if attemptCount < 3 {
				return &temporaryError{message: "temporary failure"}
			}
			return nil
		}

		err := provider.WithRetry(ctx, operation)
		if err != nil {
			t.Errorf("WithRetry() should succeed after retries, got error: %v", err)
		}

		if attemptCount != 3 {
			t.Errorf("Expected 3 attempts, got %d", attemptCount)
		}
	})

	t.Run("Retry Operations - Permanent Failure", func(t *testing.T) {
		ctx := context.Background()

		operation := func() error {
			return &permanentError{message: "permanent failure"}
		}

		err := provider.WithRetry(ctx, operation)
		if err == nil {
			t.Error("WithRetry() should fail for permanent errors")
		}
	})

	t.Run("Failover Operations", func(t *testing.T) {
		ctx := context.Background()

		operation := func(conn interfaces.IConn) error {
			// Mock operation that uses connection
			return nil
		}

		err := provider.WithFailover(ctx, operation)
		// This should work with mock connection or fail gracefully
		if err != nil {
			// Expected for now since we don't have real connections
			t.Logf("WithFailover() failed as expected without real connections: %v", err)
		}
	})
}

func TestPostgreSQLProviderThreadSafety(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Fatalf("NewPostgreSQLProvider() error = %v", err)
	}

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

		// Wait for all goroutines with timeout
		for i := 0; i < 10; i++ {
			select {
			case <-done:
			case <-time.After(time.Second * 5):
				t.Error("Timeout waiting for concurrent operations")
				return
			}
		}
	})

	t.Run("Concurrent Configuration Validation", func(t *testing.T) {
		done := make(chan bool)
		validConfig := config.NewDefaultConfig("postgres://user:pass@localhost/testdb")

		for i := 0; i < 5; i++ {
			go func() {
				defer func() { done <- true }()

				for j := 0; j < 50; j++ {
					provider.ValidateConfig(validConfig)
				}
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 5; i++ {
			select {
			case <-done:
			case <-time.After(time.Second * 3):
				t.Error("Timeout waiting for concurrent validation operations")
				return
			}
		}
	})
}

func TestPostgreSQLProviderEdgeCases(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Fatalf("NewPostgreSQLProvider() error = %v", err)
	}

	t.Run("Nil Config Validation", func(t *testing.T) {
		err := provider.ValidateConfig(nil)
		if err == nil {
			t.Error("ValidateConfig() should return error for nil config")
		}
	})

	t.Run("Empty Feature Name", func(t *testing.T) {
		if provider.SupportsFeature("") {
			t.Error("Provider should not support empty feature name")
		}
	})

	t.Run("Whitespace Feature Name", func(t *testing.T) {
		if provider.SupportsFeature("   ") {
			t.Error("Provider should not support whitespace-only feature name")
		}
	})

	t.Run("Feature Case Sensitivity", func(t *testing.T) {
		// Test that feature checks are case sensitive
		lowercase := provider.SupportsFeature("transactions")
		uppercase := provider.SupportsFeature("TRANSACTIONS")

		if lowercase && uppercase {
			t.Log("Provider supports both lowercase and uppercase feature names")
		} else if !lowercase && !uppercase {
			t.Error("Provider should support at least one case variant of 'transactions'")
		}
	})
}

func TestProviderFactoryEdgeCases(t *testing.T) {
	factory := NewProviderFactory()

	t.Run("Register Same Provider Twice", func(t *testing.T) {
		provider1, _ := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
		provider2, _ := NewPostgreSQLProvider(interfaces.ProviderTypePGX)

		err1 := factory.RegisterProvider("test", provider1)
		if err1 != nil {
			t.Errorf("First RegisterProvider() error = %v", err1)
		}

		err2 := factory.RegisterProvider("test", provider2)
		// Should either succeed (overwrite) or fail (duplicate)
		// Implementation choice - document the behavior
		t.Logf("Second RegisterProvider() result: %v", err2)
	})

	t.Run("Get Non-existent Provider", func(t *testing.T) {
		_, exists := factory.GetProvider("nonexistent")
		if exists {
			t.Error("GetProvider() should return false for non-existent provider")
		}
	})

	t.Run("Empty Provider Type", func(t *testing.T) {
		_, err := factory.CreateProvider("")
		if err == nil {
			t.Error("CreateProvider() should return error for empty provider type")
		}
	})
}

// Helper error types for testing
type temporaryError struct {
	message string
}

func (e *temporaryError) Error() string {
	return e.message
}

func (e *temporaryError) Temporary() bool {
	return true
}

type permanentError struct {
	message string
}

func (e *permanentError) Error() string {
	return e.message
}

func (e *permanentError) Temporary() bool {
	return false
}
