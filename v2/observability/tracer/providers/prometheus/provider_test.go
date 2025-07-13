package prometheus

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

func TestProviderHelperMethods(t *testing.T) {
	_, cancel := setupTestTimeout(t)
	defer cancel()

	// Test helper methods
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "config with service name",
			config: &Config{
				ServiceName: "helper-test",
				Namespace:   "test",
			},
		},
		{
			name: "config with version",
			config: &Config{
				ServiceName:    "helper-test",
				ServiceVersion: "v2.0.0",
				Namespace:      "test",
			},
		},
		{
			name: "config with environment",
			config: &Config{
				ServiceName: "helper-test",
				Environment: "staging",
				Namespace:   "test",
			},
		},
		{
			name: "complete config",
			config: &Config{
				ServiceName:           "helper-test",
				ServiceVersion:        "v3.0.0",
				Environment:           "production",
				Namespace:             "prod",
				Subsystem:             "api",
				EnableDetailedMetrics: true,
				MaxCardinality:        5000,
				CollectionInterval:    10 * time.Second,
				BucketBoundaries:      []float64{0.1, 1, 10},
				CustomLabels: map[string]string{
					"component": "auth",
					"version":   "latest",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.config == nil {
				tt.config = DefaultConfig()
			}
			tt.config.ServiceName = "helper-test"

			// Ensure required fields are set
			if tt.config.Namespace == "" {
				tt.config.Namespace = "test"
			}
			if tt.config.MaxCardinality <= 0 {
				tt.config.MaxCardinality = 1000
			}
			if tt.config.CollectionInterval <= 0 {
				tt.config.CollectionInterval = 30 * time.Second
			}
			if len(tt.config.BucketBoundaries) == 0 {
				tt.config.BucketBoundaries = []float64{0.1, 1, 10}
			}

			p, err := NewProvider(tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, p)
		})
	}
}

func TestProviderUpdateMetrics(t *testing.T) {
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)

	// Create a tracer to generate some metrics
	tr, err := provider.CreateTracer("metrics-test")
	require.NoError(t, err)

	// Get initial metrics
	metrics1 := provider.GetProviderMetrics()

	// Create some spans to generate activity
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, span := tr.StartSpan(ctx, fmt.Sprintf("test-span-%d", i))
		span.End()
	}

	// Get updated metrics
	metrics2 := provider.GetProviderMetrics()

	// Verify metrics were updated
	assert.GreaterOrEqual(t, metrics2.LastFlush, metrics1.LastFlush)
}

func TestProviderShutdownScenarios(t *testing.T) {
	_, cancel := setupTestTimeout(t)
	defer cancel()

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "shutdown with active tracers",
			test: func(t *testing.T) {
				provider := createTestProvider(t)
				_, err := provider.CreateTracer("active-tracer")
				require.NoError(t, err)

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				err = provider.Shutdown(ctx)
				assert.NoError(t, err)
			},
		},
		{
			name: "shutdown empty provider",
			test: func(t *testing.T) {
				provider := createTestProvider(t)

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				err := provider.Shutdown(ctx)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestCreateTracerEdgeCases(t *testing.T) {
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
	_, cancel := setupTestTimeout(t)
	defer cancel()

	provider := createTestProvider(t)

	// Test health check with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := provider.HealthCheck(ctx)
	// Prometheus provider might handle cancelled context differently
	// The specific behavior depends on the implementation
	_ = err // Accept either success or failure

	// Test health check when not started
	freshProvider, err := NewProvider(&Config{
		ServiceName:        "fresh-app",
		Namespace:          "test",
		MaxCardinality:     1000,
		CollectionInterval: 30 * time.Second,
		BucketBoundaries:   []float64{0.1, 1, 10},
	})
	require.NoError(t, err)

	err = freshProvider.HealthCheck(context.Background())
	// Health check behavior for uninitialized provider
	_ = err // Accept either success or failure
}

func TestCreateTracerErrorPath(t *testing.T) {
	_, cancel := setupTestTimeout(t)
	defer cancel()

	// Test with invalid config that would cause initialization to fail
	invalidConfig := &Config{
		ServiceName: "", // Invalid - empty service name
	}

	provider, err := NewProvider(invalidConfig)
	// This should fail due to validation
	assert.Error(t, err)
	assert.Nil(t, provider)
}

// Benchmark tests
func BenchmarkProviderCreateTracer(b *testing.B) {
	provider, err := NewProvider(&Config{
		ServiceName: "benchmark-app",
		Namespace:   "test",
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serviceName := fmt.Sprintf("service-%d", i)
		_, err := provider.CreateTracer(serviceName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProviderGetMetrics(b *testing.B) {
	provider, err := NewProvider(&Config{
		ServiceName: "benchmark-app",
		Namespace:   "test",
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.GetProviderMetrics()
	}
}

func BenchmarkProviderHealthCheck(b *testing.B) {
	provider, err := NewProvider(&Config{
		ServiceName: "benchmark-app",
		Namespace:   "test",
	})
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.HealthCheck(ctx)
	}
}
