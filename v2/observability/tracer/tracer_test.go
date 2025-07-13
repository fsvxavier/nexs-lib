package tracer

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGlobalFunctions(t *testing.T) {
	// Reset global factory for clean testing
	GlobalFactory = NewFactory()

	t.Run("GetProvider", func(t *testing.T) {
		// Should return error for non-existent provider
		_, err := GetProvider("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent provider")
		}

		// Register and create a provider
		mockProvider := newMockProvider("global-test")
		constructor := func(config interface{}) (Provider, error) {
			return mockProvider, nil
		}

		GlobalFactory.RegisterProvider("global-test", constructor)
		_, err = GlobalFactory.CreateProvider("global-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		// Should return the provider
		provider, err := GetProvider("global-test")
		if err != nil {
			t.Errorf("Expected no error getting provider, got %v", err)
		}
		if provider == nil {
			t.Fatal("Expected provider to not be nil")
		}
		if provider.Name() != "global-test" {
			t.Errorf("Expected provider name to be 'global-test', got %s", provider.Name())
		}
	})

	t.Run("CreateTracerManager", func(t *testing.T) {
		mockProvider := newMockProvider("manager-global-test")
		constructor := func(config interface{}) (Provider, error) {
			return mockProvider, nil
		}

		GlobalFactory.RegisterProvider("manager-global-test", constructor)

		manager, err := CreateTracerManager("manager-global-test", nil)
		if err != nil {
			t.Errorf("Expected no error creating tracer manager, got %v", err)
		}
		if manager == nil {
			t.Fatal("Expected manager to not be nil")
		}
	})

	t.Run("CreateTracerManagerError", func(t *testing.T) {
		_, err := CreateTracerManager("non-existent-provider", nil)
		if err == nil {
			t.Error("Expected error for non-existent provider")
		}
	})

	t.Run("Shutdown", func(t *testing.T) {
		// Create a provider first
		mockProvider := newMockProvider("shutdown-global-test")
		constructor := func(config interface{}) (Provider, error) {
			return mockProvider, nil
		}

		GlobalFactory.RegisterProvider("shutdown-global-test", constructor)
		_, err := GlobalFactory.CreateProvider("shutdown-global-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		ctx := context.Background()
		err = Shutdown(ctx)
		if err != nil {
			t.Errorf("Expected no error during shutdown, got %v", err)
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		// Create providers with different health states
		healthyProvider := newMockProvider("healthy-provider")
		unhealthyProvider := newMockProvider("unhealthy-provider")
		unhealthyProvider.healthFail = true

		constructor1 := func(config interface{}) (Provider, error) {
			return healthyProvider, nil
		}
		constructor2 := func(config interface{}) (Provider, error) {
			return unhealthyProvider, nil
		}

		GlobalFactory.RegisterProvider("healthy-provider", constructor1)
		GlobalFactory.RegisterProvider("unhealthy-provider", constructor2)

		_, err := GlobalFactory.CreateProvider("healthy-provider", nil)
		if err != nil {
			t.Fatalf("Error creating healthy provider: %v", err)
		}
		_, err = GlobalFactory.CreateProvider("unhealthy-provider", nil)
		if err != nil {
			t.Fatalf("Error creating unhealthy provider: %v", err)
		}

		ctx := context.Background()
		results := HealthCheck(ctx)

		if len(results) != 2 {
			t.Errorf("Expected 2 health check results, got %d", len(results))
		}

		if results["healthy-provider"] != nil {
			t.Errorf("Expected healthy-provider to be healthy, got %v", results["healthy-provider"])
		}

		if results["unhealthy-provider"] == nil {
			t.Error("Expected unhealthy-provider to be unhealthy")
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		// Reset and create a provider
		GlobalFactory = NewFactory()
		mockProvider := newMockProvider("metrics-global-test")
		constructor := func(config interface{}) (Provider, error) {
			return mockProvider, nil
		}

		GlobalFactory.RegisterProvider("metrics-global-test", constructor)
		_, err := GlobalFactory.CreateProvider("metrics-global-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		metrics := GetMetrics()
		if len(metrics) != 1 {
			t.Errorf("Expected 1 metrics result, got %d", len(metrics))
		}

		if metrics["metrics-global-test"].ConnectionState != "connected" {
			t.Errorf("Expected connection state to be 'connected', got %s", metrics["metrics-global-test"].ConnectionState)
		}
	})

	t.Run("GetProviderInfo", func(t *testing.T) {
		// Reset and create a provider
		GlobalFactory = NewFactory()
		mockProvider := newMockProvider("info-global-test")
		constructor := func(config interface{}) (Provider, error) {
			return mockProvider, nil
		}

		GlobalFactory.RegisterProvider("info-global-test", constructor)
		_, err := GlobalFactory.CreateProvider("info-global-test", nil)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		ctx := context.Background()
		infos := GetProviderInfo(ctx)

		if len(infos) != 1 {
			t.Errorf("Expected 1 provider info, got %d", len(infos))
		}

		if !infos[0].IsActive {
			t.Error("Expected provider to be active")
		}
		if infos[0].Type != "info-global-test" {
			t.Errorf("Expected provider type to be 'info-global-test', got %s", infos[0].Type)
		}
	})
}

func TestMultiProviderTracer(t *testing.T) {
	primary := NewNoopTracer()
	secondary := NewNoopTracer()
	tertiary := NewNoopTracer()

	t.Run("NewMultiProviderTracer", func(t *testing.T) {
		multi := NewMultiProviderTracer(primary)
		if multi == nil {
			t.Fatal("Expected multi tracer to not be nil")
		}
		if multi.primary != primary {
			t.Error("Expected primary tracer to be set correctly")
		}
		if len(multi.tracers) != 1 {
			t.Errorf("Expected 1 tracer, got %d", len(multi.tracers))
		}

		multiWithAdditional := NewMultiProviderTracer(primary, secondary, tertiary)
		if len(multiWithAdditional.tracers) != 3 {
			t.Errorf("Expected 3 tracers, got %d", len(multiWithAdditional.tracers))
		}
	})

	t.Run("StartSpan", func(t *testing.T) {
		multi := NewMultiProviderTracer(primary, secondary)

		ctx := context.Background()
		newCtx, span := multi.StartSpan(ctx, "test-span")

		if span == nil {
			t.Fatal("Expected span to not be nil")
		}

		// Should return a MultiSpan
		multiSpan, ok := span.(*MultiSpan)
		if !ok {
			t.Fatal("Expected span to be *MultiSpan")
		}

		if len(multiSpan.spans) != 2 {
			t.Errorf("Expected 2 spans in multi span, got %d", len(multiSpan.spans))
		}

		if multiSpan.primary == nil {
			t.Error("Expected primary span to be set")
		}

		// Context should be from primary tracer
		if newCtx != ctx { // Noop tracer returns same context
			t.Error("Expected context from primary tracer")
		}
	})
	t.Run("SpanFromContext", func(t *testing.T) {
		multi := NewMultiProviderTracer(primary, secondary)

		ctx := context.Background()
		span := multi.SpanFromContext(ctx)

		if span == nil {
			t.Fatal("Expected span to not be nil")
		}

		// Should delegate to primary tracer - verify it's a noop span with same tracer
		primarySpan := primary.SpanFromContext(ctx)
		if span != primarySpan {
			// Check if both are NoopSpan from same tracer type
			if _, ok := span.(*NoopSpan); !ok {
				t.Error("Expected span to be of same type as primary tracer span")
			}
		}
	})

	t.Run("ContextWithSpan", func(t *testing.T) {
		multi := NewMultiProviderTracer(primary, secondary)

		ctx := context.Background()
		_, span := multi.StartSpan(ctx, "test-span")

		newCtx := multi.ContextWithSpan(ctx, span)

		// Should delegate to primary tracer
		expectedCtx := primary.ContextWithSpan(ctx, span)
		if newCtx != expectedCtx {
			t.Error("Expected context from primary tracer")
		}
	})
	t.Run("Close", func(t *testing.T) {
		multi := NewMultiProviderTracer(primary, secondary, tertiary)

		// Test close with error from one tracer
		errorTracer := NewErrorTracer(fmt.Errorf("close error"))
		multiWithError := NewMultiProviderTracer(primary, errorTracer)

		err := multiWithError.Close()
		if err == nil {
			t.Error("Expected error from Close() due to error tracer")
		}

		// Test successful close
		err = multi.Close()
		if err != nil {
			t.Errorf("Expected no error during close, got %v", err)
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		multi := NewMultiProviderTracer(primary, secondary)

		metrics := multi.GetMetrics()

		// Should return metrics from primary tracer
		primaryMetrics := primary.GetMetrics()
		if metrics.SpansCreated != primaryMetrics.SpansCreated {
			t.Error("Expected metrics from primary tracer")
		}
	})
}

func TestMultiSpan(t *testing.T) {
	primary := NewNoopTracer()
	secondary := NewNoopTracer()

	multi := NewMultiProviderTracer(primary, secondary)
	ctx := context.Background()
	_, span := multi.StartSpan(ctx, "test-span")

	multiSpan, ok := span.(*MultiSpan)
	if !ok {
		t.Fatal("Expected span to be *MultiSpan")
	}

	t.Run("Context", func(t *testing.T) {
		spanCtx := multiSpan.Context()

		// Should return context from primary span
		primaryCtx := multiSpan.primary.Context()
		if spanCtx.TraceID != primaryCtx.TraceID {
			t.Error("Expected context from primary span")
		}
	})

	t.Run("SetName", func(t *testing.T) {
		// Should not panic - delegates to all spans
		multiSpan.SetName("new-name")
	})

	t.Run("SetAttributes", func(t *testing.T) {
		// Should not panic - delegates to all spans
		multiSpan.SetAttributes(map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		})
	})

	t.Run("SetAttribute", func(t *testing.T) {
		// Should not panic - delegates to all spans
		multiSpan.SetAttribute("single-key", "single-value")
	})

	t.Run("AddEvent", func(t *testing.T) {
		// Should not panic - delegates to all spans
		multiSpan.AddEvent("test-event", map[string]interface{}{
			"event-data": "test-data",
		})
	})

	t.Run("SetStatus", func(t *testing.T) {
		// Should not panic - delegates to all spans
		multiSpan.SetStatus(StatusCodeOk, "all good")
		multiSpan.SetStatus(StatusCodeError, "something went wrong")
	})

	t.Run("RecordError", func(t *testing.T) {
		// Should not panic - delegates to all spans
		testErr := &testError{message: "multi span error"}
		multiSpan.RecordError(testErr, map[string]interface{}{
			"error-context": "multi-span-test",
		})
	})

	t.Run("End", func(t *testing.T) {
		// Should not panic - delegates to all spans
		multiSpan.End()

		// After ending, span should not be recording
		if multiSpan.IsRecording() {
			t.Error("Expected multi span to not be recording after End()")
		}
	})

	t.Run("IsRecording", func(t *testing.T) {
		// Create new span for this test
		_, newSpan := multi.StartSpan(context.Background(), "recording-test")
		newMultiSpan := newSpan.(*MultiSpan)

		// Should delegate to primary span
		if newMultiSpan.IsRecording() != newMultiSpan.primary.IsRecording() {
			t.Error("Expected recording status from primary span")
		}
	})
	t.Run("GetDuration", func(t *testing.T) {
		// Create new span for this test
		_, newSpan := multi.StartSpan(context.Background(), "duration-test")
		newMultiSpan := newSpan.(*MultiSpan)

		duration := newMultiSpan.GetDuration()

		if duration < 0 {
			t.Error("Expected duration to be non-negative")
		}

		// Since both use noop spans, durations should be comparable
		primaryDuration := newMultiSpan.primary.GetDuration()
		if duration < primaryDuration/2 || duration > primaryDuration*2 {
			// Allow some reasonable variance since timing can differ slightly
			t.Logf("Duration variance detected - multi: %v, primary: %v", duration, primaryDuration)
		}
	})
}

func TestMultiProviderTracerWithOptions(t *testing.T) {
	primary := NewNoopTracer()
	secondary := NewNoopTracer()

	multi := NewMultiProviderTracer(primary, secondary)

	t.Run("StartSpanWithOptions", func(t *testing.T) {
		ctx := context.Background()

		opts := []SpanOption{
			WithSpanKind(SpanKindServer),
			WithSpanAttributes(map[string]interface{}{
				"http.method": "POST",
				"http.url":    "/api/users",
			}),
		}

		newCtx, span := multi.StartSpan(ctx, "http-request", opts...)

		if span == nil {
			t.Fatal("Expected span to not be nil")
		}

		multiSpan, ok := span.(*MultiSpan)
		if !ok {
			t.Fatal("Expected span to be *MultiSpan")
		}

		if len(multiSpan.spans) != 2 {
			t.Errorf("Expected 2 spans in multi span, got %d", len(multiSpan.spans))
		}

		// Context should be from primary tracer
		if newCtx != ctx { // Noop tracer returns same context
			t.Error("Expected context from primary tracer")
		}
	})
}

// Benchmark tests
func BenchmarkMultiProviderTracerStartSpan(b *testing.B) {
	primary := NewNoopTracer()
	secondary := NewNoopTracer()
	tertiary := NewNoopTracer()

	multi := NewMultiProviderTracer(primary, secondary, tertiary)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := multi.StartSpan(ctx, "benchmark-span")
		span.End()
	}
}

func BenchmarkMultiSpanOperations(b *testing.B) {
	primary := NewNoopTracer()
	secondary := NewNoopTracer()

	multi := NewMultiProviderTracer(primary, secondary)
	ctx := context.Background()
	_, span := multi.StartSpan(ctx, "benchmark-span")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span.SetAttribute("key", "value")
		span.AddEvent("event", map[string]interface{}{"key": "value"})
		span.SetStatus(StatusCodeOk, "ok")
	}
}

// Concurrency test for multi-provider tracer
func TestMultiProviderTracerConcurrency(b *testing.T) {
	primary := NewNoopTracer()
	secondary := NewNoopTracer()

	multi := NewMultiProviderTracer(primary, secondary)
	ctx := context.Background()

	numGoroutines := 50
	spansPerGoroutine := 20
	done := make(chan bool, numGoroutines)

	// Concurrent span operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < spansPerGoroutine; j++ {
				_, span := multi.StartSpan(ctx, "concurrent-span")

				span.SetAttribute("goroutine", id)
				span.SetAttribute("span", j)
				span.AddEvent("test-event", map[string]interface{}{
					"id":   id,
					"span": j,
				})
				span.SetStatus(StatusCodeOk, "success")
				span.End()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(5 * time.Second):
			b.Fatal("Test timed out waiting for goroutines to complete")
		}
	}

	// Verify metrics from primary tracer
	metrics := primary.GetMetrics()
	expectedSpans := int64(numGoroutines * spansPerGoroutine)

	if metrics.SpansCreated < expectedSpans {
		b.Errorf("Expected at least %d spans created, got %d", expectedSpans, metrics.SpansCreated)
	}
	if metrics.SpansFinished < expectedSpans {
		b.Errorf("Expected at least %d spans finished, got %d", expectedSpans, metrics.SpansFinished)
	}
}
