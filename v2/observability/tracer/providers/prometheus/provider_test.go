package prometheus

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

// Helper function to create a test provider
func createTestProvider(t *testing.T) *Provider {
	config := DefaultConfig()
	config.ServiceName = "test-service"
	provider, err := NewProvider(config)
	require.NoError(t, err)
	return provider
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	config := DefaultConfig()

	// Test default values
	assert.Equal(t, "tracer", config.Namespace)
	assert.Equal(t, "spans", config.Subsystem)
	assert.Equal(t, "production", config.Environment)
	assert.True(t, config.EnableDetailedMetrics)
	assert.NotNil(t, config.CustomLabels)
	assert.NotNil(t, config.BucketBoundaries)
	assert.Equal(t, 1000, config.MaxCardinality)
	assert.Equal(t, 30*time.Second, config.CollectionInterval)
	assert.Equal(t, 100, config.BatchSize)
	assert.Equal(t, 24*time.Hour, config.RetentionPeriod)
	assert.False(t, config.UseGlobalRegistry)
}

func TestNewProvider(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "nil config uses default",
			config:      nil,
			expectError: false,
		},
		{
			name: "valid config",
			config: &Config{
				ServiceName:           "test-app",
				ServiceVersion:        "1.0.0",
				Environment:           "test",
				Namespace:             "test",
				Subsystem:             "tracer",
				EnableDetailedMetrics: true,
				CustomLabels: map[string]string{
					"component": "api",
					"version":   "v1",
				},
				BucketBoundaries:   []float64{0.1, 1, 10},
				MaxCardinality:     5000,
				CollectionInterval: 10 * time.Second,
				BatchSize:          500,
				UseGlobalRegistry:  false,
			},
			expectError: false,
		},
		{
			name: "invalid config - empty service name",
			config: &Config{
				ServiceName: "",
				Environment: "test",
			},
			expectError: true,
		},
		{
			name: "invalid config - negative batch size",
			config: &Config{
				ServiceName: "test-app",
				BatchSize:   -1,
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
				assert.Equal(t, "prometheus", provider.Name())
			}
		})
	}
}

func TestProviderName(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)
	assert.Equal(t, "prometheus", provider.Name())
}

func TestProviderCreateTracer(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

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
	_, cancel := setupTestTimeout(t)
	defer cancel()

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
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)

	// Create a tracer to initialize the provider
	_, err := provider.CreateTracer("health-test")
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestProviderGetProviderMetrics(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)

	metrics := provider.GetProviderMetrics()
	assert.NotNil(t, metrics)
	assert.False(t, metrics.LastFlush.IsZero())
	assert.GreaterOrEqual(t, metrics.TracersActive, 0)
}

func TestProviderConcurrentAccess(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)

	const numGoroutines = 10
	const tracersPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < tracersPerGoroutine; j++ {
				serviceName := fmt.Sprintf("service-%d-%d", id, j)
				tracer, err := provider.CreateTracer(serviceName)
				assert.NoError(t, err)
				assert.NotNil(t, tracer)
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	metrics := provider.GetProviderMetrics()
	assert.GreaterOrEqual(t, metrics.TracersActive, 1)
}

func TestProviderConfigValidation(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "valid minimal config",
			config: &Config{
				ServiceName:        "test-app",
				Namespace:          "test",
				MaxCardinality:     1000,
				CollectionInterval: 30 * time.Second,
				BucketBoundaries:   []float64{0.1, 1, 10},
			},
			valid: true,
		},
		{
			name: "missing service name",
			config: &Config{
				Environment: "test",
				Namespace:   "test",
			},
			valid: false,
		},
		{
			name: "complete config",
			config: &Config{
				ServiceName:           "test-app",
				ServiceVersion:        "1.0.0",
				Environment:           "test",
				Namespace:             "custom",
				Subsystem:             "metrics",
				EnableDetailedMetrics: true,
				CustomLabels: map[string]string{
					"team":    "platform",
					"region":  "us-west-2",
					"cluster": "prod",
				},
				BucketBoundaries:   []float64{0.1, 1, 10, 100},
				MaxCardinality:     15000,
				CollectionInterval: 15 * time.Second,
				BatchSize:          2000,
				RetentionPeriod:    10 * time.Minute,
				UseGlobalRegistry:  true,
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

// TestGetRegistry tests the GetRegistry method
func TestGetRegistry(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)
	registry := provider.GetRegistry()
	assert.NotNil(t, registry)
	assert.IsType(t, &prometheus.Registry{}, registry)
}

// TestGetMetrics tests the GetMetrics method
func TestGetMetrics(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)
	metrics := provider.GetMetrics()

	assert.NotNil(t, metrics)
	assert.Contains(t, metrics, "span_counter")
	assert.Contains(t, metrics, "span_duration")
	assert.Contains(t, metrics, "error_counter")
	assert.Contains(t, metrics, "active_spans")
}

// TestHelperMethods tests the helper methods with different configurations
func TestHelperMethods(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)

	// Test with default tracer config
	config := &tracer.TracerConfig{}

	serviceName := provider.getServiceName(config)
	assert.Equal(t, "test-service", serviceName)

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

// TestBuildLabels tests the buildLabels method
func TestBuildLabels(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)
	provider.config.CustomLabels["region"] = "us-east-1"
	provider.config.CustomLabels["team"] = "backend"

	config := &tracer.TracerConfig{
		ServiceName:    "test-app",
		ServiceVersion: "1.5.0",
		Environment:    "development",
	}

	labels := provider.buildLabels("test-tracer", "test-span", tracer.SpanKindClient, config)

	expected := prometheus.Labels{
		"service_name":    "test-app",
		"service_version": "1.5.0",
		"environment":     "development",
		"tracer_name":     "test-tracer",
		"span_name":       "test-span",
		"span_kind":       "CLIENT", // SpanKindClient returns "CLIENT" in uppercase
		"region":          "us-east-1",
		"team":            "backend",
	}

	assert.Equal(t, expected, labels)
}

// TestValidateConfigEdgeCases tests edge cases for config validation
func TestValidateConfigEdgeCases(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty namespace",
			config: &Config{
				ServiceName:        "test",
				Namespace:          "",
				MaxCardinality:     1000,
				CollectionInterval: 30 * time.Second,
				BucketBoundaries:   []float64{0.1, 1, 10},
			},
			expectError: true,
			errorMsg:    "namespace is required",
		},
		{
			name: "zero max cardinality",
			config: &Config{
				ServiceName:        "test",
				Namespace:          "test",
				MaxCardinality:     0,
				CollectionInterval: 30 * time.Second,
				BucketBoundaries:   []float64{0.1, 1, 10},
			},
			expectError: true,
			errorMsg:    "max cardinality must be positive",
		},
		{
			name: "negative collection interval",
			config: &Config{
				ServiceName:        "test",
				Namespace:          "test",
				MaxCardinality:     1000,
				CollectionInterval: -1 * time.Second,
				BucketBoundaries:   []float64{0.1, 1, 10},
			},
			expectError: true,
			errorMsg:    "collection interval must be positive",
		},
		{
			name: "empty bucket boundaries",
			config: &Config{
				ServiceName:        "test",
				Namespace:          "test",
				MaxCardinality:     1000,
				CollectionInterval: 30 * time.Second,
				BucketBoundaries:   []float64{},
			},
			expectError: true,
			errorMsg:    "bucket boundaries cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProviderWithCustomRegistry tests provider with custom registry
func TestProviderWithCustomRegistry(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	customRegistry := prometheus.NewRegistry()
	config := &Config{
		ServiceName:        "test-service",
		Namespace:          "test",
		UseGlobalRegistry:  false,
		Registry:           customRegistry,
		MaxCardinality:     1000,
		CollectionInterval: 30 * time.Second,
		BucketBoundaries:   []float64{0.1, 1, 10},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	assert.Equal(t, customRegistry, provider.GetRegistry())
}

// TestProviderWithGlobalRegistry tests provider with global registry
func TestProviderWithGlobalRegistry(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	config := &Config{
		ServiceName:        "test-service",
		Namespace:          "test",
		UseGlobalRegistry:  true,
		MaxCardinality:     1000,
		CollectionInterval: 30 * time.Second,
		BucketBoundaries:   []float64{0.1, 1, 10},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	assert.NotNil(t, provider.GetRegistry())
}

// TestHealthCheckWithError tests health check when there's a last error
func TestHealthCheckWithError(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)
	provider.lastError = fmt.Errorf("test error")

	err := provider.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "last error: test error")
}

// TestShutdownWithTracerError tests shutdown when tracer close returns error
func TestShutdownWithTracerError(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)

	// Create a tracer
	_, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	// Shutdown should handle any tracer close errors gracefully
	err = provider.Shutdown(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "shutdown", provider.healthStatus)
	assert.Equal(t, "disconnected", provider.metrics.ConnectionState)
	assert.Equal(t, 0, provider.metrics.TracersActive)
}

// TestConcurrentOperations tests concurrent operations for race conditions
func TestConcurrentOperations(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

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

				// Get registry and metrics
				_ = provider.GetRegistry()
				_ = provider.GetMetrics()
			}
		}(i)
	}

	wg.Wait()

	// Verify final state
	assert.Greater(t, len(provider.tracers), 0)
	assert.Equal(t, "healthy", provider.healthStatus)
}

// TestProviderMemoryLeaks tests for potential memory leaks
func TestProviderMemoryLeaks(t *testing.T) {
	t.Parallel()
	_, cancel := setupTestTimeout(t)
	defer cancel()

	// Create and shutdown multiple providers to test cleanup
	for i := 0; i < 10; i++ {
		config := &Config{
			ServiceName:        fmt.Sprintf("test-service-%d", i),
			Namespace:          "test",
			MaxCardinality:     1000,
			CollectionInterval: 30 * time.Second,
			BucketBoundaries:   []float64{0.1, 1, 10},
		}

		provider, err := NewProvider(config)
		require.NoError(t, err)

		// Create some tracers
		for j := 0; j < 5; j++ {
			_, err := provider.CreateTracer(fmt.Sprintf("tracer-%d", j))
			require.NoError(t, err)
		}

		// Shutdown and verify cleanup
		err = provider.Shutdown(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(provider.tracers))
		assert.Equal(t, "shutdown", provider.healthStatus)
	}
}
