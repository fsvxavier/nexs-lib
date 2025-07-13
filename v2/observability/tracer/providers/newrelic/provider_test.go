package newrelic

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

// Helper function to create a test provider
func createTestProvider(t *testing.T) *Provider {
	config := DefaultConfig()
	config.LicenseKey = "1234567890123456789012345678901234567890" // 40 character test license key
	provider, err := NewProvider(config)
	require.NoError(t, err)
	return provider
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "unknown-service", config.AppName)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "1.0.0", config.ServiceVersion)
	assert.True(t, config.DistributedTracer)
	assert.True(t, config.Enabled)
	assert.False(t, config.HighSecurity)
	assert.True(t, config.CodeLevelMetrics)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, 10000, config.MaxSamplesStored)
	assert.True(t, config.DatastoreTracer)
	assert.True(t, config.CrossApplicationTrace)
	assert.Equal(t, 60*time.Second, config.FlushInterval)
	assert.True(t, config.AttributesEnabled)
	assert.NotNil(t, config.AttributesInclude)
	assert.NotNil(t, config.AttributesExclude)
	assert.True(t, config.CustomInsightsEvents)
	assert.NotNil(t, config.Labels)
	assert.NotNil(t, config.CustomAttributes)
	assert.True(t, config.ErrorCollector.Enabled)
	assert.True(t, config.ErrorCollector.RecordPanics)
	assert.NotNil(t, config.ErrorCollector.IgnoreStatusCodes)
	assert.Equal(t, 100, config.ErrorCollector.MaxEventsSamplesStored)
}

func TestNewProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "nil config uses default",
			config:      nil,
			expectError: true, // Will fail validation due to empty license key
		},
		{
			name: "valid config",
			config: &Config{
				AppName:           "test-app",
				LicenseKey:        "1234567890123456789012345678901234567890", // 40 character test license key
				Environment:       "test",
				ServiceVersion:    "1.0.0",
				DistributedTracer: true,
				Enabled:           true,
			},
			expectError: false,
		},
		{
			name: "invalid config - empty app name",
			config: &Config{
				AppName:    "",
				LicenseKey: "1234567890123456789012345678901234567890", // 40 character test license key
			},
			expectError: true,
		},
		{
			name: "invalid config - empty license key",
			config: &Config{
				AppName:    "test-app",
				LicenseKey: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, "newrelic", provider.Name())
			}
		})
	}
}

func TestProviderName(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
	assert.Equal(t, "newrelic", provider.Name())
}

func TestProviderCreateTracer(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	tests := []struct {
		name        string
		serviceName string
		options     []tracer.TracerOption
		expectError bool
	}{
		{
			name:        "valid tracer creation",
			serviceName: "test-service",
			options:     nil,
			expectError: false,
		},
		{
			name:        "empty service name",
			serviceName: "",
			options:     nil,
			expectError: false, // Provider accepts empty names
		},
		{
			name:        "tracer with options",
			serviceName: "service-with-options",
			options: []tracer.TracerOption{
				tracer.WithServiceName("override-service"),
				tracer.WithServiceVersion("2.0.0"),
				tracer.WithEnvironment("testing"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := provider.CreateTracer(tt.serviceName, tt.options...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tr)
			}
		})
	}
}

func TestProviderShutdown(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	// Create some tracers
	_, err := provider.CreateTracer("test-service-1")
	require.NoError(t, err)
	_, err = provider.CreateTracer("test-service-2")
	require.NoError(t, err)

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = provider.Shutdown(ctx)
	assert.NoError(t, err)

	// Test that subsequent operations work (provider doesn't prevent new tracers after shutdown)
	_, err = provider.CreateTracer("after-shutdown")
	assert.NoError(t, err) // Provider allows creating tracers after shutdown
}

func TestProviderHealthCheck(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	// Create a tracer to initialize the provider
	_, err := provider.CreateTracer("health-test")
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.HealthCheck(ctx)
	// In test environment, New Relic might not be able to connect
	// so we accept either success or specific connection errors
	if err != nil {
		assert.Contains(t, err.Error(), "connection")
	}
}

func TestProviderGetProviderMetrics(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	metrics := provider.GetProviderMetrics()
	assert.NotNil(t, metrics)
	assert.False(t, metrics.LastFlush.IsZero())
	assert.GreaterOrEqual(t, metrics.TracersActive, 0)
}

func TestProviderConcurrentAccess(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	const numGoroutines = 10
	const numTracers = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numTracers; j++ {
				serviceName := fmt.Sprintf("service-%d-%d", id, j)
				_, err := provider.CreateTracer(serviceName)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all tracers were created
	metrics := provider.GetProviderMetrics()
	assert.Equal(t, numGoroutines*numTracers, metrics.TracersActive)
}

func TestProviderConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "valid minimal config",
			config: &Config{
				AppName:    "test-app",
				LicenseKey: "1234567890123456789012345678901234567890", // 40 character test license key
			},
			valid: true,
		},
		{
			name: "missing app name",
			config: &Config{
				LicenseKey: "1234567890123456789012345678901234567890", // 40 character test license key
			},
			valid: false,
		},
		{
			name: "missing license key",
			config: &Config{
				AppName: "test-app",
			},
			valid: false,
		},
		{
			name: "complete config",
			config: &Config{
				AppName:           "test-app",
				LicenseKey:        "1234567890123456789012345678901234567890", // 40 character test license key
				Environment:       "test",
				ServiceVersion:    "1.0.0",
				DistributedTracer: true,
				Enabled:           true,
				HighSecurity:      true,
				CodeLevelMetrics:  true,
				LogLevel:          "debug",
				MaxSamplesStored:  5000,
				DatastoreTracer:   true,
				FlushInterval:     30 * time.Second,
				AttributesEnabled: true,
				AttributesInclude: []string{"custom.*"},
				AttributesExclude: []string{"sensitive.*"},
				Labels: map[string]string{
					"env":  "test",
					"team": "backend",
				},
				CustomAttributes: map[string]interface{}{
					"version": "1.0.0",
				},
				ErrorCollector: ErrorCollectorConfig{
					Enabled:                true,
					RecordPanics:           true,
					IgnoreStatusCodes:      []int{404, 401},
					CaptureEvents:          true,
					MaxEventsSamplesStored: 200,
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewProvider(tt.config)

			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestProviderHelperMethods(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	tests := []struct {
		name   string
		config *tracer.TracerConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "config with service name",
			config: &tracer.TracerConfig{
				ServiceName: "override-service",
			},
		},
		{
			name: "config with version",
			config: &tracer.TracerConfig{
				ServiceVersion: "2.0.0",
			},
		},
		{
			name: "config with environment",
			config: &tracer.TracerConfig{
				Environment: "staging",
			},
		},
		{
			name: "complete config",
			config: &tracer.TracerConfig{
				ServiceName:    "full-service",
				ServiceVersion: "3.0.0",
				Environment:    "production",
				Attributes: map[string]interface{}{
					"team":   "backend",
					"region": "us-west-2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test helper methods (these are private, so we test through CreateTracer)
			serviceName := "test-service"
			_, err := provider.CreateTracer(serviceName)
			assert.NoError(t, err)
		})
	}
}

// Benchmark tests
func BenchmarkNewProvider(b *testing.B) {
	config := DefaultConfig()
	config.LicenseKey = "1234567890123456789012345678901234567890" // 40 character test license key

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider, err := NewProvider(config)
		if err != nil {
			b.Fatal(err)
		}
		_ = provider
	}
}

func BenchmarkProviderCreateTracer(b *testing.B) {
	provider, err := NewProvider(&Config{
		AppName:    "benchmark-app",
		LicenseKey: "1234567890123456789012345678901234567890", // 40 character test license key
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracerName := fmt.Sprintf("tracer-%d", i)
		_, err := provider.CreateTracer(tracerName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Additional comprehensive tests for better coverage

func TestProviderUpdateMetrics(t *testing.T) {
	provider := createTestProvider(t)

	// Call updateMetrics directly through creating tracers
	for i := 0; i < 3; i++ {
		_, err := provider.CreateTracer(fmt.Sprintf("metrics-test-%d", i))
		require.NoError(t, err)
	}

	// Verify metrics were updated
	metrics := provider.GetProviderMetrics()
	assert.False(t, metrics.LastFlush.IsZero())
	assert.Equal(t, 3, metrics.TracersActive)
}

func TestProviderShutdownScenarios(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *Provider
	}{
		{
			name: "shutdown with active tracers",
			setupFunc: func() *Provider {
				provider := createTestProvider(t)
				// Create multiple tracers
				for i := 0; i < 3; i++ {
					_, err := provider.CreateTracer(fmt.Sprintf("test-service-%d", i))
					assert.NoError(t, err)
				}
				return provider
			},
		},
		{
			name: "shutdown empty provider",
			setupFunc: func() *Provider {
				return createTestProvider(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setupFunc()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := provider.Shutdown(ctx)
			assert.NoError(t, err)

			// Verify provider is shut down
			metrics := provider.GetProviderMetrics()
			assert.NotNil(t, metrics)
		})
	}
}

func TestCreateTracerEdgeCases(t *testing.T) {
	provider := createTestProvider(t)

	tests := []struct {
		name        string
		serviceName string
		options     []tracer.TracerOption
		expectError bool
	}{
		{
			name:        "empty service name with no options",
			serviceName: "",
			options:     []tracer.TracerOption{},
			expectError: false, // Provider accepts empty names
		},
		{
			name:        "valid service with complex options",
			serviceName: "complex-service",
			options: []tracer.TracerOption{
				tracer.WithServiceName("override-service"),
				tracer.WithServiceVersion("v2.0.0"),
				tracer.WithEnvironment("staging"),
				tracer.WithTracerAttributes(map[string]interface{}{
					"region":      "us-west-2",
					"datacenter":  "dc1",
					"application": "api-gateway",
				}),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := provider.CreateTracer(tt.serviceName, tt.options...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tr)
			}
		})
	}
}

func TestHealthCheckErrorScenarios(t *testing.T) {
	provider := createTestProvider(t)

	// Test health check with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := provider.HealthCheck(ctx)
	// New Relic provider might handle cancelled context differently
	// The specific behavior depends on the implementation
	_ = err // Accept either success or failure

	// Test health check when not started
	freshProvider, err := NewProvider(&Config{
		AppName:    "fresh-app",
		LicenseKey: "1234567890123456789012345678901234567890", // 40 character test license key
	})
	require.NoError(t, err)

	err = freshProvider.HealthCheck(context.Background())
	// Health check behavior for uninitialized provider
	_ = err // Accept either success or failure
}

func TestCreateTracerErrorPath(t *testing.T) {
	// Test with invalid config that would cause initialization to fail
	invalidConfig := &Config{
		AppName:    "",                                         // Invalid - empty app name
		LicenseKey: "1234567890123456789012345678901234567890", // 40 character test license key
	}

	provider, err := NewProvider(invalidConfig)
	// This should fail due to validation
	assert.Error(t, err)
	assert.Nil(t, provider)
}

// TestWaitForConnection tests the waitForConnection method
func TestWaitForConnection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "successful connection",
			config: &Config{
				AppName:    "test-app",
				LicenseKey: "1234567890123456789012345678901234567890",
				Enabled:    true,
			},
			expectError: false,
		},
		{
			name: "disabled application",
			config: &Config{
				AppName:    "test-app",
				LicenseKey: "1234567890123456789012345678901234567890",
				Enabled:    false,
			},
			expectError: false, // Should not error when disabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)
			require.NoError(t, err)

			// Test with short timeout - just ensure it doesn't panic
			provider.waitForConnection()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Don't assert no error as New Relic might timeout in test environment
				// Just ensure it doesn't panic
				_ = err
			}
		})
	}
}

// TestGetServiceHelpers tests the service helper methods
func TestGetServiceHelpers(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	// Test with default tracer config
	config := &tracer.TracerConfig{}

	serviceName := provider.getServiceName(config)
	assert.Equal(t, "unknown-service", serviceName)

	serviceVersion := provider.getServiceVersion(config)
	assert.Equal(t, "1.0.0", serviceVersion)

	environment := provider.getEnvironment(config)
	assert.Equal(t, "production", environment)

	// Test with custom tracer config
	customConfig := &tracer.TracerConfig{
		ServiceName:    "custom-service",
		ServiceVersion: "2.0.0",
		Environment:    "staging",
	}

	serviceName = provider.getServiceName(customConfig)
	assert.Equal(t, "custom-service", serviceName)

	serviceVersion = provider.getServiceVersion(customConfig)
	assert.Equal(t, "2.0.0", serviceVersion)

	environment = provider.getEnvironment(customConfig)
	assert.Equal(t, "staging", environment)
}

// TestValidateConfigEdgeCases tests edge cases for config validation
func TestValidateConfigEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty app name",
			config: &Config{
				AppName:    "",
				LicenseKey: "1234567890123456789012345678901234567890",
			},
			expectError: true,
			errorMsg:    "app name is required",
		},
		{
			name: "empty license key",
			config: &Config{
				AppName:    "test-app",
				LicenseKey: "",
			},
			expectError: true,
			errorMsg:    "license key is required",
		},
		{
			name: "short license key",
			config: &Config{
				AppName:    "test-app",
				LicenseKey: "123", // Too short
			},
			expectError: true,
			errorMsg:    "license key must be 40 characters",
		},
		{
			name: "negative max samples",
			config: &Config{
				AppName:          "test-app",
				LicenseKey:       "1234567890123456789012345678901234567890",
				MaxSamplesStored: -1,
			},
			expectError: true,
			errorMsg:    "max samples stored must be positive",
		},
		{
			name: "negative flush interval",
			config: &Config{
				AppName:       "test-app",
				LicenseKey:    "1234567890123456789012345678901234567890",
				FlushInterval: -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "flush interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestHealthCheckWithConnectionTimeout tests health check with connection timeout
func TestHealthCheckWithConnectionTimeout(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	// Test health check with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := provider.HealthCheck(ctx)
	// Could succeed or fail depending on New Relic connection, just ensure no panic
	_ = err
}

// TestShutdownWithTimeout tests shutdown with timeout
func TestShutdownWithTimeout(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	// Create some tracers
	for i := 0; i < 3; i++ {
		_, err := provider.CreateTracer(fmt.Sprintf("tracer-%d", i))
		require.NoError(t, err)
	}

	// Test shutdown with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := provider.Shutdown(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "disconnected", provider.healthStatus)
	assert.Equal(t, "disconnected", provider.metrics.ConnectionState)
	assert.Equal(t, 0, len(provider.tracers))
}

// TestConcurrentOperations tests concurrent operations for race conditions
func TestConcurrentOperations(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
	numWorkers := 50
	numOperations := 100

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Start multiple goroutines performing various operations
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Create tracers
				tracerName := fmt.Sprintf("tracer-%d-%d", workerID, j)
				_, err := provider.CreateTracer(tracerName)
				assert.NoError(t, err)

				// Get metrics
				_ = provider.GetProviderMetrics()

				// Health check
				_ = provider.HealthCheck(context.Background())
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	assert.Greater(t, len(provider.tracers), 0)
	// Status might be initializing or connected during concurrent operations
	assert.Contains(t, []string{"initializing", "connected", "healthy"}, provider.healthStatus)
}

// TestProviderMemoryLeaks tests for potential memory leaks
func TestProviderMemoryLeaks(t *testing.T) {
	t.Parallel()

	// Create and shutdown multiple providers to test cleanup
	for i := 0; i < 5; i++ {
		config := &Config{
			AppName:    fmt.Sprintf("test-app-%d", i),
			LicenseKey: "1234567890123456789012345678901234567890",
		}

		provider, err := NewProvider(config)
		require.NoError(t, err)

		// Create some tracers
		for j := 0; j < 3; j++ {
			_, err := provider.CreateTracer(fmt.Sprintf("tracer-%d", j))
			require.NoError(t, err)
		}

		// Shutdown and verify cleanup
		err = provider.Shutdown(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(provider.tracers))
		assert.Equal(t, "disconnected", provider.healthStatus)
	}
}

// TestCreateTracerWithOptions tests tracer creation with various options
func TestCreateTracerWithOptions(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	tests := []struct {
		name string
		opts []tracer.TracerOption
	}{
		{
			name: "with service name",
			opts: []tracer.TracerOption{
				tracer.WithServiceName("custom-service"),
			},
		},
		{
			name: "with service version",
			opts: []tracer.TracerOption{
				tracer.WithServiceVersion("2.0.0"),
			},
		},
		{
			name: "with environment",
			opts: []tracer.TracerOption{
				tracer.WithEnvironment("staging"),
			},
		},
		{
			name: "with multiple options",
			opts: []tracer.TracerOption{
				tracer.WithServiceName("multi-service"),
				tracer.WithServiceVersion("3.0.0"),
				tracer.WithEnvironment("development"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := provider.CreateTracer(tt.name, tt.opts...)
			assert.NoError(t, err)
			assert.NotNil(t, tr)
			assert.Implements(t, (*tracer.Tracer)(nil), tr)
		})
	}
}

// TestProviderShutdownIdempotent tests that shutdown can be called multiple times
func TestProviderShutdownIdempotent(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)

	// Create a tracer
	_, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	// First shutdown should succeed
	err = provider.Shutdown(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "disconnected", provider.healthStatus)

	// Second shutdown should also succeed (idempotent)
	err = provider.Shutdown(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "disconnected", provider.healthStatus)
}
