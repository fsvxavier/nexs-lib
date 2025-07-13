package datadog

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test provider
func createTestProvider(t *testing.T) *Provider {
	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)
	return provider
}

func TestNewProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name:        "nil config uses default",
			config:      nil,
			expectError: false,
		},
		{
			name: "valid custom config",
			config: &Config{
				ServiceName:      "test-service",
				ServiceVersion:   "v1.0.0",
				Environment:      "test",
				AgentHost:        "localhost",
				AgentPort:        8126,
				SampleRate:       0.8,
				EnableProfiling:  true,
				Tags:             map[string]string{"env": "test"},
				Debug:            true,
				RuntimeMetrics:   true,
				AnalyticsEnabled: true,
				PrioritySampling: true,
				CustomAttributes: map[string]interface{}{"custom": "value"},
			},
			expectError: false,
		},
		{
			name: "invalid service name",
			config: &Config{
				ServiceName: "",
				AgentPort:   8126,
			},
			expectError: true,
			errorMsg:    "invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider, err := NewProvider(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.NotNil(t, provider.config)
				assert.NotNil(t, provider.tracers)
				assert.Equal(t, "disconnected", provider.metrics.ConnectionState)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	assert.Equal(t, "unknown-service", config.ServiceName)
	assert.Equal(t, "1.0.0", config.ServiceVersion)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "localhost", config.AgentHost)
	assert.Equal(t, 8126, config.AgentPort)
	assert.Equal(t, 1.0, config.SampleRate)
	assert.False(t, config.EnableProfiling)
	assert.NotNil(t, config.Tags)
	assert.True(t, config.RuntimeMetrics)
	assert.True(t, config.AnalyticsEnabled)
	assert.True(t, config.PrioritySampling)
	assert.Equal(t, 1000, config.MaxTracesPerSecond)
	assert.Equal(t, 5*time.Second, config.FlushInterval)
	assert.True(t, config.ObfuscationEnabled)
	assert.Contains(t, config.ObfuscatedTags, "password")
	assert.Contains(t, config.ObfuscatedTags, "token")
	assert.NotNil(t, config.CustomAttributes)
}

func TestProviderCreateTracer(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, provider)

	tests := []struct {
		name        string
		tracerName  string
		options     []tracer.TracerOption
		expectError bool
		errorMsg    string
	}{
		{
			name:       "valid tracer creation",
			tracerName: "test-tracer",
			options: []tracer.TracerOption{
				tracer.WithServiceName("test-service"),
				tracer.WithServiceVersion("v1.0.0"),
			},
			expectError: false,
		},
		{
			name:        "nil options uses defaults",
			tracerName:  "test-tracer-2",
			options:     nil,
			expectError: false,
		},
		{
			name:        "duplicate tracer name returns existing",
			tracerName:  "test-tracer",
			options:     nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := provider.CreateTracer(tt.tracerName, tt.options...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, tr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tr)
				assert.IsType(t, &Tracer{}, tr)

				// Verify tracer is stored
				provider.mu.RLock()
				storedTracer, exists := provider.tracers[tt.tracerName]
				provider.mu.RUnlock()

				assert.True(t, exists)
				assert.Equal(t, tr, storedTracer)
			}
		})
	}
}

func TestProviderName(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, provider)

	assert.Equal(t, "datadog", provider.Name())
}

func TestProviderShutdown(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create some tracers
	_, err = provider.CreateTracer("tracer1")
	require.NoError(t, err)
	_, err = provider.CreateTracer("tracer2")
	require.NoError(t, err)

	// Test shutdown
	err = provider.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestProviderGetProviderMetrics(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, provider)

	metrics := provider.GetProviderMetrics()

	assert.Equal(t, "disconnected", metrics.ConnectionState)
	assert.Equal(t, 0, metrics.TracersActive)
	assert.False(t, metrics.LastFlush.IsZero())
}

func TestProviderHealthCheck(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test health check on unstarted provider
	err = provider.HealthCheck(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not started")
}

// TestProviderUpdateMetrics tests the updateMetrics method
func TestProviderUpdateMetrics(t *testing.T) {
	provider := createTestProvider(t)

	// Call updateMetrics directly
	provider.updateMetrics()

	// Verify metrics were updated
	metrics := provider.GetProviderMetrics()
	assert.False(t, metrics.LastFlush.IsZero(), "LastFlush should be set")
	assert.GreaterOrEqual(t, metrics.BytesSent, int64(1024), "BytesSent should be updated")
}

// TestProviderCollectMetricsShutdown tests collectMetrics with shutdown scenario
func TestProviderCollectMetricsShutdown(t *testing.T) {
	provider := createTestProvider(t)

	// Start metrics collection
	ctx, cancel := context.WithCancel(context.Background())

	// Run collectMetrics in background
	go provider.collectMetrics()

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Cancel context to simulate shutdown
	cancel()

	// Shutdown provider to stop metrics collection
	err := provider.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestProviderHealthCheckDetailed tests all health check scenarios
func TestProviderHealthCheckDetailed(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() *Provider
		expectError bool
	}{
		{
			name: "healthy provider",
			setupFunc: func() *Provider {
				provider := createTestProvider(t)
				// Start the provider to make it healthy
				_, err := provider.CreateTracer("health-test")
				require.NoError(t, err)
				return provider
			},
			expectError: false,
		},
		{
			name: "provider with closed context",
			setupFunc: func() *Provider {
				provider := createTestProvider(t)
				// Simulate closed context by shutting down
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				provider.Shutdown(ctx)
				return provider
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.setupFunc()

			ctx := context.Background()
			err := provider.HealthCheck(ctx)

			if tt.expectError {
				assert.Error(t, err, "Expected health check to fail")
			} else {
				assert.NoError(t, err, "Expected health check to pass")
			}
		})
	}
}

// TestProviderShutdownScenarios tests various shutdown scenarios
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
			assert.NoError(t, err, "Shutdown should complete successfully")

			// Verify provider is shut down
			metrics := provider.GetProviderMetrics()
			assert.NotNil(t, metrics, "Metrics should still be accessible after shutdown")
		})
	}
}

// TestCreateTracerEdgeCases tests edge cases in CreateTracer
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
			expectError: true,
		},
		{
			name:        "empty service name with empty options",
			serviceName: "",
			options:     []tracer.TracerOption{},
			expectError: true,
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
			tracer, err := provider.CreateTracer(tt.serviceName, tt.options...)

			if tt.expectError {
				assert.Error(t, err, "Expected CreateTracer to fail")
				assert.Nil(t, tracer, "Tracer should be nil on error")
			} else {
				assert.NoError(t, err, "Expected CreateTracer to succeed")
				assert.NotNil(t, tracer, "Tracer should not be nil")
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewProvider(b *testing.B) {
	config := DefaultConfig()

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
	provider, err := NewProvider(DefaultConfig())
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

func TestProviderHelperMethods(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	// Test helper methods with different configurations
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
				ServiceName: "custom-service",
			},
		},
		{
			name: "config with version",
			config: &tracer.TracerConfig{
				ServiceVersion: "v2.0.0",
			},
		},
		{
			name: "config with environment",
			config: &tracer.TracerConfig{
				Environment: "production",
			},
		},
		{
			name: "complete config",
			config: &tracer.TracerConfig{
				ServiceName:    "test-service",
				ServiceVersion: "v1.2.3",
				Environment:    "staging",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test getServiceName
			serviceName := provider.getServiceName(tt.config)
			assert.NotEmpty(t, serviceName)

			// Test getServiceVersion
			version := provider.getServiceVersion(tt.config)
			assert.NotEmpty(t, version)

			// Test getEnvironment
			env := provider.getEnvironment(tt.config)
			assert.NotEmpty(t, env)
		})
	}
}

func TestProviderConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "empty service name",
			config: &Config{
				ServiceName: "",
				AgentPort:   8126,
			},
			expectError: true,
			errorMsg:    "service name is required",
		},
		{
			name: "invalid agent port - zero",
			config: &Config{
				ServiceName: "test",
				AgentPort:   0,
			},
			expectError: true,
			errorMsg:    "agent port must be between 1 and 65535",
		},
		{
			name: "invalid agent port - negative",
			config: &Config{
				ServiceName: "test",
				AgentPort:   -1,
			},
			expectError: true,
			errorMsg:    "agent port must be between 1 and 65535",
		},
		{
			name: "invalid agent port - too high",
			config: &Config{
				ServiceName: "test",
				AgentPort:   65536,
			},
			expectError: true,
			errorMsg:    "agent port must be between 1 and 65535",
		},
		{
			name: "invalid sample rate - negative",
			config: &Config{
				ServiceName: "test",
				AgentPort:   8126,
				SampleRate:  -0.1,
			},
			expectError: true,
			errorMsg:    "sample rate must be between 0 and 1",
		},
		{
			name: "invalid sample rate - too high",
			config: &Config{
				ServiceName: "test",
				AgentPort:   8126,
				SampleRate:  1.5,
			},
			expectError: true,
			errorMsg:    "sample rate must be between 0 and 1",
		},
		{
			name: "edge case - sample rate 0",
			config: &Config{
				ServiceName: "test",
				AgentPort:   8126,
				SampleRate:  0.0,
			},
			expectError: false,
		},
		{
			name: "edge case - sample rate 1",
			config: &Config{
				ServiceName: "test",
				AgentPort:   8126,
				SampleRate:  1.0,
			},
			expectError: false,
		},
		{
			name: "negative max traces",
			config: &Config{
				ServiceName:        "test",
				AgentPort:          8126,
				MaxTracesPerSecond: -1,
			},
			expectError: true,
			errorMsg:    "max traces per second cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

func TestProviderWithComplexConfig(t *testing.T) {
	t.Parallel()

	config := &Config{
		ServiceName:        "complex-test-service",
		ServiceVersion:     "v1.2.3",
		Environment:        "testing",
		AgentHost:          "custom-host",
		AgentPort:          9999,
		SampleRate:         0.75,
		EnableProfiling:    true,
		Tags:               map[string]string{"team": "engineering", "project": "test"},
		Debug:              true,
		RuntimeMetrics:     false,
		AnalyticsEnabled:   false,
		PrioritySampling:   false,
		MaxTracesPerSecond: 500,
		FlushInterval:      10 * time.Second,
		ObfuscationEnabled: false,
		ObfuscatedTags:     []string{"custom-secret"},
		CustomAttributes:   map[string]interface{}{"build": "123", "commit": "abc"},
	}

	provider, err := NewProvider(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	assert.Equal(t, config, provider.config)
	assert.Equal(t, "datadog", provider.Name())

	// Test creating tracer with complex config
	tr, err := provider.CreateTracer("complex-tracer",
		tracer.WithServiceName("override-service"),
		tracer.WithTracerAttributes(map[string]interface{}{
			"custom": "attribute",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, tr)

	// Test span creation with this tracer
	ctx := context.Background()
	newCtx, span := tr.StartSpan(ctx, "complex-span")
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	assert.True(t, span.IsRecording())

	span.End()
	assert.False(t, span.IsRecording())
}

// TestHealthCheckErrorScenarios tests HealthCheck error paths
func TestHealthCheckErrorScenarios(t *testing.T) {
	provider := createTestProvider(t)

	// Test health check with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := provider.HealthCheck(ctx)
	assert.Error(t, err, "Expected health check to fail with cancelled context")

	// Test health check when not started
	freshProvider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	err = freshProvider.HealthCheck(context.Background())
	assert.Error(t, err, "Expected health check to fail when provider not started")
}

// TestShutdownErrorPaths tests Shutdown scenarios
func TestShutdownErrorPaths(t *testing.T) {
	provider := createTestProvider(t)

	// Test shutdown with normal context (should succeed)
	ctx := context.Background()

	err := provider.Shutdown(ctx)
	assert.NoError(t, err, "Expected shutdown to succeed")

	// Test shutdown when already shutdown (should be idempotent)
	err = provider.Shutdown(ctx)
	assert.NoError(t, err, "Expected second shutdown to succeed (idempotent)")
}

// TestCollectMetricsStopCondition tests collectMetrics stop condition
func TestCollectMetricsStopCondition(t *testing.T) {
	provider := createTestProvider(t)

	// Create a tracer to start the provider
	_, err := provider.CreateTracer("metrics-test")
	require.NoError(t, err)

	// Test that collectMetrics respects shutdown
	done := make(chan struct{})
	go func() {
		defer close(done)
		provider.collectMetrics()
	}()

	// Shutdown provider
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = provider.Shutdown(ctx)
	assert.NoError(t, err)

	// Wait for goroutine to finish
	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Error("collectMetrics did not stop after shutdown")
	}
}

// TestCreateTracerErrorPath tests CreateTracer error scenarios
func TestCreateTracerErrorPath(t *testing.T) {
	// Test with invalid config that would cause startDatadogTracer to fail
	invalidConfig := &Config{
		ServiceName: "", // Invalid - empty service name
		Environment: "test",
	}

	provider, err := NewProvider(invalidConfig)
	// This should fail due to validation
	assert.Error(t, err, "Expected NewProvider to fail with invalid config")
	assert.Nil(t, provider, "Provider should be nil on error")
}
