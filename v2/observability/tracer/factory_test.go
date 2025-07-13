package tracer

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// Mock provider for testing
type mockProvider struct {
	name         string
	shouldFail   bool
	shutdownFail bool
	healthFail   bool
	tracers      map[string]Tracer
}

func newMockProvider(name string) *mockProvider {
	return &mockProvider{
		name:    name,
		tracers: make(map[string]Tracer),
	}
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) CreateTracer(name string, options ...TracerOption) (Tracer, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock provider error")
	}

	tracer := NewNoopTracer()
	m.tracers[name] = tracer
	return tracer, nil
}

func (m *mockProvider) Shutdown(ctx context.Context) error {
	if m.shutdownFail {
		return fmt.Errorf("shutdown failed")
	}
	m.tracers = make(map[string]Tracer)
	return nil
}

func (m *mockProvider) HealthCheck(ctx context.Context) error {
	if m.healthFail {
		return fmt.Errorf("health check failed")
	}
	return nil
}

func (m *mockProvider) GetProviderMetrics() ProviderMetrics {
	return ProviderMetrics{
		TracersActive:   len(m.tracers),
		ConnectionState: "connected",
		LastFlush:       time.Now(),
		ErrorCount:      0,
		BytesSent:       1024,
	}
}

func TestFactory(t *testing.T) {
	factory := NewFactory()

	t.Run("NewFactory", func(t *testing.T) {
		if factory == nil {
			t.Fatal("Expected factory to not be nil")
		}
	})

	t.Run("RegisterProvider", func(t *testing.T) {
		provider := newMockProvider("test-provider")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("test", constructor)

		// Verify provider was registered by checking list
		providers := factory.ListProviders()
		found := false
		for _, p := range providers {
			if p == "test" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected provider to be registered")
		}
	})

	t.Run("CreateProvider", func(t *testing.T) {
		provider := newMockProvider("create-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("create-test", constructor)

		// Test creating provider
		createdProvider, err := factory.CreateProvider("create-test", nil)
		if err != nil {
			t.Errorf("Expected no error creating provider, got %v", err)
		}
		if createdProvider == nil {
			t.Fatal("Expected provider to not be nil")
		}
		if createdProvider.Name() != "create-test" {
			t.Errorf("Expected provider name to be 'create-test', got %s", createdProvider.Name())
		}

		// Test creating same provider again (should return cached)
		cachedProvider, err := factory.CreateProvider("create-test", nil)
		if err != nil {
			t.Errorf("Expected no error getting cached provider, got %v", err)
		}
		if cachedProvider != createdProvider {
			t.Error("Expected same provider instance from cache")
		}
	})

	t.Run("CreateProviderNotRegistered", func(t *testing.T) {
		_, err := factory.CreateProvider("non-existent", nil)
		if err == nil {
			t.Error("Expected error for non-existent provider")
		}
		expectedMsg := "provider type non-existent is not registered"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})
	t.Run("CreateProviderConstructorError", func(t *testing.T) {
		constructor := func(config interface{}) (Provider, error) {
			return nil, fmt.Errorf("constructor error")
		}

		factory.RegisterProvider("error-test", constructor)

		_, err := factory.CreateProvider("error-test", nil)
		if err == nil {
			t.Error("Expected error from constructor")
		}
		if !errors.Is(err, fmt.Errorf("constructor error")) && err.Error() != "failed to create provider error-test: constructor error" {
			t.Errorf("Expected wrapped constructor error, got: %v", err)
		}
	})

	t.Run("GetProvider", func(t *testing.T) {
		provider := newMockProvider("get-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("get-test", constructor)

		// Create provider first
		_, err := factory.CreateProvider("get-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		// Test getting existing provider
		gotProvider, err := factory.GetProvider("get-test")
		if err != nil {
			t.Errorf("Expected no error getting provider, got %v", err)
		}
		if gotProvider.Name() != "get-test" {
			t.Errorf("Expected provider name to be 'get-test', got %s", gotProvider.Name())
		}

		// Test getting non-existent provider
		_, err = factory.GetProvider("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent provider")
		}
	})

	t.Run("ListProviders", func(t *testing.T) {
		newFactory := NewFactory()

		// Should be empty initially
		providers := newFactory.ListProviders()
		if len(providers) != 0 {
			t.Errorf("Expected 0 providers, got %d", len(providers))
		}

		// Register some providers
		constructor := func(config interface{}) (Provider, error) {
			return newMockProvider("test"), nil
		}

		newFactory.RegisterProvider("provider1", constructor)
		newFactory.RegisterProvider("provider2", constructor)

		providers = newFactory.ListProviders()
		if len(providers) != 2 {
			t.Errorf("Expected 2 providers, got %d", len(providers))
		}
	})

	t.Run("GetActiveProviders", func(t *testing.T) {
		newFactory := NewFactory()
		provider := newMockProvider("active-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		newFactory.RegisterProvider("active-test", constructor)

		// Should be empty initially
		active := newFactory.GetActiveProviders()
		if len(active) != 0 {
			t.Errorf("Expected 0 active providers, got %d", len(active))
		}

		// Create provider
		_, err := newFactory.CreateProvider("active-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		// Should have one active provider
		active = newFactory.GetActiveProviders()
		if len(active) != 1 {
			t.Errorf("Expected 1 active provider, got %d", len(active))
		}
	})

	t.Run("Shutdown", func(t *testing.T) {
		newFactory := NewFactory()
		provider := newMockProvider("shutdown-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		newFactory.RegisterProvider("shutdown-test", constructor)

		// Create provider
		_, err := newFactory.CreateProvider("shutdown-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		// Test successful shutdown
		ctx := context.Background()
		err = newFactory.Shutdown(ctx)
		if err != nil {
			t.Errorf("Expected no error during shutdown, got %v", err)
		}

		// Providers should be cleared
		active := newFactory.GetActiveProviders()
		if len(active) != 0 {
			t.Errorf("Expected 0 active providers after shutdown, got %d", len(active))
		}
	})

	t.Run("ShutdownWithErrors", func(t *testing.T) {
		newFactory := NewFactory()
		provider := newMockProvider("shutdown-error-test")
		provider.shutdownFail = true
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		newFactory.RegisterProvider("shutdown-error-test", constructor)

		// Create provider
		_, err := newFactory.CreateProvider("shutdown-error-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		// Test shutdown with errors
		ctx := context.Background()
		err = newFactory.Shutdown(ctx)
		if err == nil {
			t.Error("Expected error during shutdown")
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		newFactory := NewFactory()
		provider1 := newMockProvider("health1")
		provider2 := newMockProvider("health2")
		provider2.healthFail = true

		constructor1 := func(config interface{}) (Provider, error) {
			return provider1, nil
		}
		constructor2 := func(config interface{}) (Provider, error) {
			return provider2, nil
		}

		newFactory.RegisterProvider("health1", constructor1)
		newFactory.RegisterProvider("health2", constructor2)

		// Create providers
		_, err := newFactory.CreateProvider("health1", nil)
		if err != nil {
			t.Fatalf("Error creating provider1: %v", err)
		}
		_, err = newFactory.CreateProvider("health2", nil)
		if err != nil {
			t.Fatalf("Error creating provider2: %v", err)
		}

		ctx := context.Background()
		results := newFactory.HealthCheck(ctx)

		if len(results) != 2 {
			t.Errorf("Expected 2 health check results, got %d", len(results))
		}

		if results["health1"] != nil {
			t.Errorf("Expected health1 to be healthy, got %v", results["health1"])
		}

		if results["health2"] == nil {
			t.Error("Expected health2 to be unhealthy")
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		newFactory := NewFactory()
		provider := newMockProvider("metrics-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		newFactory.RegisterProvider("metrics-test", constructor)

		// Create provider
		_, err := newFactory.CreateProvider("metrics-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		metrics := newFactory.GetMetrics()
		if len(metrics) != 1 {
			t.Errorf("Expected 1 metrics result, got %d", len(metrics))
		}

		if metrics["metrics-test"].ConnectionState != "connected" {
			t.Errorf("Expected connection state to be 'connected', got %s", metrics["metrics-test"].ConnectionState)
		}
	})

	t.Run("GetProviderInfo", func(t *testing.T) {
		newFactory := NewFactory()
		provider := newMockProvider("info-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		newFactory.RegisterProvider("info-test", constructor)

		// Test with no active providers
		ctx := context.Background()
		infos := newFactory.GetProviderInfo(ctx)
		if len(infos) != 1 {
			t.Errorf("Expected 1 provider info, got %d", len(infos))
		}
		if infos[0].IsActive {
			t.Error("Expected provider to not be active")
		}

		// Create provider
		_, err := newFactory.CreateProvider("info-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		// Test with active provider
		infos = newFactory.GetProviderInfo(ctx)
		if len(infos) != 1 {
			t.Errorf("Expected 1 provider info, got %d", len(infos))
		}
		if !infos[0].IsActive {
			t.Error("Expected provider to be active")
		}
		if infos[0].Type != "info-test" {
			t.Errorf("Expected provider type to be 'info-test', got %s", infos[0].Type)
		}
	})
}

func TestTracerManager(t *testing.T) {
	t.Run("NewTracerManager", func(t *testing.T) {
		factory := NewFactory()
		provider := newMockProvider("manager-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("manager-test", constructor)

		manager, err := NewTracerManager(factory, "manager-test", nil)
		if err != nil {
			t.Errorf("Expected no error creating manager, got %v", err)
		}
		if manager == nil {
			t.Fatal("Expected manager to not be nil")
		}
	})

	t.Run("NewTracerManagerProviderError", func(t *testing.T) {
		factory := NewFactory()

		_, err := NewTracerManager(factory, "non-existent", nil)
		if err == nil {
			t.Error("Expected error for non-existent provider")
		}
	})

	t.Run("GetTracer", func(t *testing.T) {
		factory := NewFactory()
		provider := newMockProvider("tracer-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("tracer-test", constructor)

		manager, err := NewTracerManager(factory, "tracer-test", nil)
		if err != nil {
			t.Fatalf("Error creating manager: %v", err)
		}

		// Test creating tracer
		tracer, err := manager.GetTracer("test-tracer")
		if err != nil {
			t.Errorf("Expected no error getting tracer, got %v", err)
		}
		if tracer == nil {
			t.Fatal("Expected tracer to not be nil")
		}

		// Test getting same tracer again (should return cached)
		tracer2, err := manager.GetTracer("test-tracer")
		if err != nil {
			t.Errorf("Expected no error getting cached tracer, got %v", err)
		}
		if tracer != tracer2 {
			t.Error("Expected same tracer instance from cache")
		}
	})

	t.Run("GetTracerWithError", func(t *testing.T) {
		factory := NewFactory()
		provider := newMockProvider("tracer-error-test")
		provider.shouldFail = true
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("tracer-error-test", constructor)

		manager, err := NewTracerManager(factory, "tracer-error-test", nil)
		if err != nil {
			t.Fatalf("Error creating manager: %v", err)
		}

		_, err = manager.GetTracer("test-tracer")
		if err == nil {
			t.Error("Expected error getting tracer")
		}
	})

	t.Run("ListTracers", func(t *testing.T) {
		factory := NewFactory()
		provider := newMockProvider("list-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("list-test", constructor)

		manager, err := NewTracerManager(factory, "list-test", nil)
		if err != nil {
			t.Fatalf("Error creating manager: %v", err)
		}

		// Should be empty initially
		tracers := manager.ListTracers()
		if len(tracers) != 0 {
			t.Errorf("Expected 0 tracers, got %d", len(tracers))
		}

		// Create some tracers
		_, err = manager.GetTracer("tracer1")
		if err != nil {
			t.Fatalf("Error creating tracer1: %v", err)
		}
		_, err = manager.GetTracer("tracer2")
		if err != nil {
			t.Fatalf("Error creating tracer2: %v", err)
		}

		tracers = manager.ListTracers()
		if len(tracers) != 2 {
			t.Errorf("Expected 2 tracers, got %d", len(tracers))
		}
	})

	t.Run("GetAllMetrics", func(t *testing.T) {
		factory := NewFactory()
		provider := newMockProvider("metrics-manager-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("metrics-manager-test", constructor)

		manager, err := NewTracerManager(factory, "metrics-manager-test", nil)
		if err != nil {
			t.Fatalf("Error creating manager: %v", err)
		}

		// Create tracer
		_, err = manager.GetTracer("test-tracer")
		if err != nil {
			t.Fatalf("Error creating tracer: %v", err)
		}

		metrics := manager.GetAllMetrics()
		if len(metrics) != 1 {
			t.Errorf("Expected 1 metrics result, got %d", len(metrics))
		}

		if _, exists := metrics["test-tracer"]; !exists {
			t.Error("Expected metrics for 'test-tracer'")
		}
	})
	t.Run("Shutdown", func(t *testing.T) {
		factory := NewFactory()
		provider := newMockProvider("shutdown-manager-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("shutdown-manager-test", constructor)

		manager, err := NewTracerManager(factory, "shutdown-manager-test", nil)
		if err != nil {
			t.Fatalf("Error creating manager: %v", err)
		}

		// Create tracer
		_, err = manager.GetTracer("test-tracer")
		if err != nil {
			t.Fatalf("Error creating tracer: %v", err)
		}

		// Test shutdown error from provider
		provider.shutdownFail = true

		ctx := context.Background()
		err = manager.Shutdown(ctx)
		if err == nil {
			t.Error("Expected error during shutdown with provider failure")
		}

		// Reset provider state and test successful shutdown
		provider.shutdownFail = false

		// Create new manager for clean shutdown test
		newManager, err := NewTracerManager(factory, "shutdown-manager-test", nil)
		if err != nil {
			t.Fatalf("Error creating new manager: %v", err)
		}

		_, err = newManager.GetTracer("test-tracer-2")
		if err != nil {
			t.Fatalf("Error creating tracer: %v", err)
		}

		err = newManager.Shutdown(ctx)
		if err != nil {
			t.Errorf("Expected no error during shutdown, got %v", err)
		}

		// Tracers should be cleared
		tracers := newManager.ListTracers()
		if len(tracers) != 0 {
			t.Errorf("Expected 0 tracers after shutdown, got %d", len(tracers))
		}
	})

	t.Run("GetProvider", func(t *testing.T) {
		factory := NewFactory()
		provider := newMockProvider("get-provider-test")
		constructor := func(config interface{}) (Provider, error) {
			return provider, nil
		}

		factory.RegisterProvider("get-provider-test", constructor)

		manager, err := NewTracerManager(factory, "get-provider-test", nil)
		if err != nil {
			t.Fatalf("Error creating manager: %v", err)
		}

		gotProvider := manager.GetProvider()
		if gotProvider == nil {
			t.Fatal("Expected provider to not be nil")
		}
		if gotProvider.Name() != "get-provider-test" {
			t.Errorf("Expected provider name to be 'get-provider-test', got %s", gotProvider.Name())
		}
	})
}

// Benchmark tests
func BenchmarkFactoryCreateProvider(b *testing.B) {
	factory := NewFactory()
	provider := newMockProvider("bench-test")
	constructor := func(config interface{}) (Provider, error) {
		return provider, nil
	}

	factory.RegisterProvider("bench-test", constructor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = factory.CreateProvider("bench-test", nil)
	}
}

func BenchmarkTracerManagerGetTracer(b *testing.B) {
	factory := NewFactory()
	provider := newMockProvider("bench-tracer-test")
	constructor := func(config interface{}) (Provider, error) {
		return provider, nil
	}

	factory.RegisterProvider("bench-tracer-test", constructor)

	manager, err := NewTracerManager(factory, "bench-tracer-test", nil)
	if err != nil {
		b.Fatalf("Error creating manager: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GetTracer("bench-tracer")
	}
}

// Concurrency test
func TestFactoryConcurrency(t *testing.T) {
	factory := NewFactory()
	provider := newMockProvider("concurrency-test")
	constructor := func(config interface{}) (Provider, error) {
		return provider, nil
	}

	factory.RegisterProvider("concurrency-test", constructor)

	numGoroutines := 100
	done := make(chan bool, numGoroutines)

	// Concurrent provider creation
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Try to create provider
			_, err := factory.CreateProvider("concurrency-test", nil)
			if err != nil {
				t.Errorf("Error creating provider in goroutine %d: %v", id, err)
				return
			}

			// Try to get provider
			_, err = factory.GetProvider("concurrency-test")
			if err != nil {
				t.Errorf("Error getting provider in goroutine %d: %v", id, err)
				return
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for goroutines to complete")
		}
	}
}
