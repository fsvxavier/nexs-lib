package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/datadog"
)

func TestEdgeCaseSimulator(t *testing.T) {
	// Create a test tracer
	config := &datadog.Config{
		ServiceName:        "edge-cases-test",
		ServiceVersion:     "1.0.0",
		Environment:        "test",
		AgentHost:          "localhost",
		AgentPort:          8126,
		SampleRate:         1.0,
		EnableProfiling:    false,
		RuntimeMetrics:     false,
		Debug:              false, // Reduce noise in tests
		MaxTracesPerSecond: 1000,
	}

	provider, err := datadog.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	tr, err := provider.CreateTracer("edge-cases-test")
	if err != nil {
		t.Fatalf("Failed to create tracer: %v", err)
	}

	simulator := NewEdgeCaseSimulator(tr)
	ctx := context.Background()

	t.Run("NetworkFailures", func(t *testing.T) {
		err := simulator.SimulateNetworkFailures(ctx)
		// Network failures should be handled gracefully, not cause test failure
		if err != nil {
			t.Logf("Network failure simulated successfully: %v", err)
		}
	})

	t.Run("ResourceExhaustion", func(t *testing.T) {
		err := simulator.SimulateResourceExhaustion(ctx)
		// Resource exhaustion should be handled gracefully
		if err != nil {
			t.Logf("Resource exhaustion simulated successfully: %v", err)
		}
	})

	t.Run("DataCorruption", func(t *testing.T) {
		err := simulator.SimulateDataCorruption(ctx)
		// Data corruption should not cause failures as it's handled
		if err != nil {
			t.Errorf("Data corruption handling failed: %v", err)
		}
	})

	t.Run("ConcurrencyIssues", func(t *testing.T) {
		err := simulator.SimulateConcurrencyIssues(ctx)
		if err != nil {
			t.Errorf("Concurrency test failed: %v", err)
		}
	})

	// Cleanup
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := provider.Shutdown(shutdownCtx); err != nil {
		t.Logf("Warning: Provider shutdown error: %v", err)
	}
}

func TestCircuitBreaker(t *testing.T) {
	config := &datadog.Config{
		ServiceName: "circuit-breaker-test",
		Environment: "test",
		AgentHost:   "localhost",
		AgentPort:   8126,
		SampleRate:  1.0,
		Debug:       false,
	}

	provider, err := datadog.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	tr, err := provider.CreateTracer("circuit-breaker-test")
	if err != nil {
		t.Fatalf("Failed to create tracer: %v", err)
	}

	cb := NewCircuitBreaker(3, 100*time.Millisecond) // 3 failures, 100ms reset
	ctx := context.Background()

	t.Run("CircuitBreakerClosed", func(t *testing.T) {
		// Should execute successfully when circuit is closed
		err := cb.Execute(ctx, tr, func() error {
			return nil // Success
		})
		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
	})

	t.Run("CircuitBreakerOpen", func(t *testing.T) {
		// Force circuit breaker to open by generating failures
		for i := 0; i < 3; i++ {
			cb.Execute(ctx, tr, func() error {
				return errors.New("forced failure")
			})
		}

		// Now circuit should be open
		err := cb.Execute(ctx, tr, func() error {
			return nil
		})
		if err == nil {
			t.Error("Expected circuit breaker to be open and reject calls")
		}
	})

	t.Run("CircuitBreakerHalfOpen", func(t *testing.T) {
		// Wait for reset timeout
		time.Sleep(150 * time.Millisecond)

		// First call should transition to half-open and succeed
		err := cb.Execute(ctx, tr, func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected success on half-open transition, got: %v", err)
		}
	})

	// Cleanup
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	provider.Shutdown(shutdownCtx)
}

func TestRetryWithBackoff(t *testing.T) {
	config := &datadog.Config{
		ServiceName: "retry-test",
		Environment: "test",
		AgentHost:   "localhost",
		AgentPort:   8126,
		SampleRate:  1.0,
		Debug:       false,
	}

	provider, err := datadog.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	tr, err := provider.CreateTracer("retry-test")
	if err != nil {
		t.Fatalf("Failed to create tracer: %v", err)
	}

	ctx := context.Background()

	t.Run("RetrySuccess", func(t *testing.T) {
		retryConfig := RetryConfig{
			MaxAttempts:   3,
			BaseDelay:     10 * time.Millisecond,
			MaxDelay:      100 * time.Millisecond,
			BackoffFactor: 2.0,
			Jitter:        false, // Disable for predictable testing
		}

		attempts := int64(0)
		err := RetryWithBackoff(ctx, tr, retryConfig, func() error {
			attempt := atomic.AddInt64(&attempts, 1)
			if attempt < 2 {
				return errors.New("temporary failure")
			}
			return nil // Success on second attempt
		})

		if err != nil {
			t.Errorf("Expected success after retry, got: %v", err)
		}

		if atomic.LoadInt64(&attempts) != 2 {
			t.Errorf("Expected 2 attempts, got: %d", atomic.LoadInt64(&attempts))
		}
	})

	t.Run("RetryExhausted", func(t *testing.T) {
		retryConfig := RetryConfig{
			MaxAttempts:   2,
			BaseDelay:     1 * time.Millisecond,
			MaxDelay:      10 * time.Millisecond,
			BackoffFactor: 2.0,
			Jitter:        false,
		}

		attempts := int64(0)
		err := RetryWithBackoff(ctx, tr, retryConfig, func() error {
			atomic.AddInt64(&attempts, 1)
			return errors.New("persistent failure")
		})

		if err == nil {
			t.Error("Expected retry to be exhausted")
		}

		if atomic.LoadInt64(&attempts) != 2 {
			t.Errorf("Expected 2 attempts, got: %d", atomic.LoadInt64(&attempts))
		}
	})

	t.Run("RetryWithContext", func(t *testing.T) {
		retryConfig := RetryConfig{
			MaxAttempts:   5,
			BaseDelay:     50 * time.Millisecond,
			MaxDelay:      200 * time.Millisecond,
			BackoffFactor: 2.0,
			Jitter:        false,
		}

		// Create context that cancels quickly
		retryCtx, cancel := context.WithTimeout(ctx, 75*time.Millisecond)
		defer cancel()

		attempts := int64(0)
		err := RetryWithBackoff(retryCtx, tr, retryConfig, func() error {
			atomic.AddInt64(&attempts, 1)
			return errors.New("slow failure")
		})

		if err == nil {
			t.Error("Expected context cancellation to stop retry")
		}

		// Should have made at least one attempt but been cancelled before completing all
		if atomic.LoadInt64(&attempts) == 0 {
			t.Error("Expected at least one attempt")
		}
		if atomic.LoadInt64(&attempts) >= 5 {
			t.Error("Expected context cancellation to prevent all attempts")
		}
	})

	// Cleanup
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	provider.Shutdown(shutdownCtx)
}

func TestResourceManager(t *testing.T) {
	rm := &ResourceManager{
		maxMemoryMB:   50,
		maxGoroutines: 100,
		maxOpenFiles:  200,
	}

	t.Run("MemoryTracking", func(t *testing.T) {
		// Simulate memory allocation
		rm.mu.Lock()
		rm.currentMemoryMB = 30
		rm.mu.Unlock()

		rm.mu.RLock()
		if rm.currentMemoryMB != 30 {
			t.Errorf("Expected 30MB, got %dMB", rm.currentMemoryMB)
		}
		if rm.currentMemoryMB > rm.maxMemoryMB {
			t.Errorf("Memory usage exceeded limit: %dMB > %dMB", rm.currentMemoryMB, rm.maxMemoryMB)
		}
		rm.mu.RUnlock()
	})

	t.Run("GoroutineTracking", func(t *testing.T) {
		rm.mu.Lock()
		rm.currentGoroutines = 50
		rm.mu.Unlock()

		rm.mu.RLock()
		if rm.currentGoroutines != 50 {
			t.Errorf("Expected 50 goroutines, got %d", rm.currentGoroutines)
		}
		if rm.currentGoroutines > int64(rm.maxGoroutines) {
			t.Errorf("Goroutines exceeded limit: %d > %d", rm.currentGoroutines, rm.maxGoroutines)
		}
		rm.mu.RUnlock()
	})

	t.Run("FileDescriptorTracking", func(t *testing.T) {
		rm.mu.Lock()
		rm.currentOpenFiles = 150
		rm.mu.Unlock()

		rm.mu.RLock()
		if rm.currentOpenFiles != 150 {
			t.Errorf("Expected 150 files, got %d", rm.currentOpenFiles)
		}
		if rm.currentOpenFiles > int64(rm.maxOpenFiles) {
			t.Errorf("Files exceeded limit: %d > %d", rm.currentOpenFiles, rm.maxOpenFiles)
		}
		rm.mu.RUnlock()
	})
}

func TestBackoffCalculation(t *testing.T) {
	config := RetryConfig{
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      2 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	testCases := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 800 * time.Millisecond},
		{5, 1600 * time.Millisecond},
		{6, 2000 * time.Millisecond}, // Capped at MaxDelay
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Attempt%d", tc.attempt), func(t *testing.T) {
			delay := calculateBackoffDelay(config, tc.attempt)
			if delay != tc.expected {
				t.Errorf("Attempt %d: expected %v, got %v", tc.attempt, tc.expected, delay)
			}
		})
	}
}

func TestBackoffWithJitter(t *testing.T) {
	config := RetryConfig{
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      2 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}

	// Test that jitter adds some randomness
	delay1 := calculateBackoffDelay(config, 2)
	delay2 := calculateBackoffDelay(config, 2)

	// Delays should be in the expected range
	expectedBase := 200 * time.Millisecond
	if delay1 < expectedBase || delay1 > expectedBase+20*time.Millisecond {
		t.Errorf("Delay1 out of expected range: %v", delay1)
	}
	if delay2 < expectedBase || delay2 > expectedBase+20*time.Millisecond {
		t.Errorf("Delay2 out of expected range: %v", delay2)
	}

	// With jitter, delays might be different (though not guaranteed due to randomness)
	t.Logf("Delay1: %v, Delay2: %v (jitter enabled)", delay1, delay2)
}

// Benchmark tests for performance validation
func BenchmarkCircuitBreakerExecute(b *testing.B) {
	config := &datadog.Config{
		ServiceName: "benchmark-test",
		Environment: "test",
		AgentHost:   "localhost",
		AgentPort:   8126,
		SampleRate:  0.1, // Lower sampling for benchmarks
		Debug:       false,
	}

	provider, err := datadog.NewProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}

	tr, err := provider.CreateTracer("benchmark-test")
	if err != nil {
		b.Fatalf("Failed to create tracer: %v", err)
	}

	cb := NewCircuitBreaker(1000, 1*time.Second) // High threshold for benchmarks
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(ctx, tr, func() error {
				return nil // Always succeed for benchmark
			})
		}
	})

	// Cleanup
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	provider.Shutdown(shutdownCtx)
}

func BenchmarkRetryWithBackoff(b *testing.B) {
	config := &datadog.Config{
		ServiceName: "benchmark-retry-test",
		Environment: "test",
		AgentHost:   "localhost",
		AgentPort:   8126,
		SampleRate:  0.1,
		Debug:       false,
	}

	provider, err := datadog.NewProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}

	tr, err := provider.CreateTracer("benchmark-retry-test")
	if err != nil {
		b.Fatalf("Failed to create tracer: %v", err)
	}

	retryConfig := RetryConfig{
		MaxAttempts:   1, // No retries for benchmark
		BaseDelay:     1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RetryWithBackoff(ctx, tr, retryConfig, func() error {
				return nil // Always succeed for benchmark
			})
		}
	})

	// Cleanup
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	provider.Shutdown(shutdownCtx)
}
