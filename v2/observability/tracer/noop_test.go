package tracer

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNoopTracer(t *testing.T) {
	tracer := NewNoopTracer()

	t.Run("StartSpan", func(t *testing.T) {
		ctx := context.Background()
		newCtx, span := tracer.StartSpan(ctx, "test-span")

		if newCtx != ctx {
			t.Error("Expected context to be unchanged")
		}
		if span == nil {
			t.Fatal("Expected span to not be nil")
		}
		if !span.IsRecording() {
			t.Error("Expected noop span to report as recording initially")
		}
	})

	t.Run("SpanFromContext", func(t *testing.T) {
		ctx := context.Background()
		span := tracer.SpanFromContext(ctx)

		if span == nil {
			t.Fatal("Expected span to not be nil")
		}
		if span.IsRecording() {
			t.Error("Expected span from context to not be recording")
		}
	})

	t.Run("ContextWithSpan", func(t *testing.T) {
		ctx := context.Background()
		_, span := tracer.StartSpan(ctx, "test-span")

		newCtx := tracer.ContextWithSpan(ctx, span)
		if newCtx != ctx {
			t.Error("Expected context to be unchanged")
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := tracer.Close()
		if err != nil {
			t.Errorf("Expected no error from Close(), got %v", err)
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		metrics := tracer.GetMetrics()
		if metrics.SpansCreated < 0 {
			t.Error("Expected SpansCreated to be non-negative")
		}
		if metrics.LastActivity.IsZero() {
			t.Error("Expected LastActivity to be set")
		}
	})
}

func TestNoopSpan(t *testing.T) {
	tracer := NewNoopTracer()
	_, span := tracer.StartSpan(context.Background(), "test-span")

	t.Run("Context", func(t *testing.T) {
		ctx := span.Context()
		expectedTraceID := "00000000000000000000000000000000"
		expectedSpanID := "0000000000000000"

		if ctx.TraceID != expectedTraceID {
			t.Errorf("Expected TraceID to be %s, got %s", expectedTraceID, ctx.TraceID)
		}
		if ctx.SpanID != expectedSpanID {
			t.Errorf("Expected SpanID to be %s, got %s", expectedSpanID, ctx.SpanID)
		}
	})

	t.Run("SetName", func(t *testing.T) {
		// Should not panic
		span.SetName("new-name")
	})

	t.Run("SetAttributes", func(t *testing.T) {
		// Should not panic
		span.SetAttributes(map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		})
	})

	t.Run("SetAttribute", func(t *testing.T) {
		// Should not panic
		span.SetAttribute("test-key", "test-value")
	})

	t.Run("AddEvent", func(t *testing.T) {
		// Should not panic
		span.AddEvent("test-event", map[string]interface{}{
			"event-key": "event-value",
		})
	})

	t.Run("SetStatus", func(t *testing.T) {
		// Should not panic
		span.SetStatus(StatusCodeOk, "all good")
		span.SetStatus(StatusCodeError, "something went wrong")
	})

	t.Run("RecordError", func(t *testing.T) {
		// Should not panic
		err := &testError{message: "test error"}
		span.RecordError(err, map[string]interface{}{
			"error-context": "test-context",
		})
	})

	t.Run("SetStatus", func(t *testing.T) {
		// Should not panic
		span.SetStatus(StatusCodeOk, "all good")
		span.SetStatus(StatusCodeError, "something went wrong")
	})

	t.Run("RecordError", func(t *testing.T) {
		// Should not panic
		err := &testError{message: "test error"}
		span.RecordError(err, map[string]interface{}{
			"error-context": "test-context",
		})
	})

	t.Run("End", func(t *testing.T) {
		initialDuration := span.GetDuration()

		// Should not panic
		span.End()

		// After ending, span should not be recording
		if span.IsRecording() {
			t.Error("Expected span to not be recording after End()")
		}

		// Duration should be available
		finalDuration := span.GetDuration()
		if finalDuration < initialDuration {
			t.Error("Expected duration to not decrease after ending span")
		}
	})

	t.Run("GetDuration", func(t *testing.T) {
		duration := span.GetDuration()
		if duration < 0 {
			t.Error("Expected duration to be non-negative")
		}
	})
}

func TestNoopProvider(t *testing.T) {
	provider := NewNoopProvider()

	t.Run("Name", func(t *testing.T) {
		name := provider.Name()
		expected := "noop"
		if name != expected {
			t.Errorf("Expected name to be %s, got %s", expected, name)
		}
	})

	t.Run("CreateTracer", func(t *testing.T) {
		tracer, err := provider.CreateTracer("test-tracer")
		if err != nil {
			t.Errorf("Expected no error creating tracer, got %v", err)
		}
		if tracer == nil {
			t.Fatal("Expected tracer to not be nil")
		}

		// Creating tracer with same name should return same instance
		tracer2, err := provider.CreateTracer("test-tracer")
		if err != nil {
			t.Errorf("Expected no error creating tracer, got %v", err)
		}
		if tracer != tracer2 {
			t.Error("Expected same tracer instance for same name")
		}
	})

	t.Run("CreateTracerWithOptions", func(t *testing.T) {
		options := []TracerOption{
			WithServiceName("test-service"),
			WithServiceVersion("v1.0.0"),
			WithEnvironment("test"),
		}

		tracer, err := provider.CreateTracer("test-tracer-opts", options...)
		if err != nil {
			t.Errorf("Expected no error creating tracer with options, got %v", err)
		}
		if tracer == nil {
			t.Fatal("Expected tracer to not be nil")
		}
	})

	t.Run("Shutdown", func(t *testing.T) {
		ctx := context.Background()
		err := provider.Shutdown(ctx)
		if err != nil {
			t.Errorf("Expected no error from Shutdown(), got %v", err)
		}

		// After shutdown, tracers should be cleared
		metrics := provider.GetProviderMetrics()
		if metrics.TracersActive != 0 {
			t.Errorf("Expected 0 active tracers after shutdown, got %d", metrics.TracersActive)
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		ctx := context.Background()
		err := provider.HealthCheck(ctx)
		if err != nil {
			t.Errorf("Expected no error from HealthCheck(), got %v", err)
		}
	})

	t.Run("GetProviderMetrics", func(t *testing.T) {
		metrics := provider.GetProviderMetrics()
		if metrics.ConnectionState != "connected" {
			t.Errorf("Expected ConnectionState to be 'connected', got %s", metrics.ConnectionState)
		}
		if metrics.LastFlush.IsZero() {
			t.Error("Expected LastFlush to be set")
		}
	})
}

func TestErrorTracer(t *testing.T) {
	testErr := &testError{message: "test error"}
	tracer := NewErrorTracer(testErr)

	t.Run("StartSpan", func(t *testing.T) {
		ctx := context.Background()
		newCtx, span := tracer.StartSpan(ctx, "test-span")

		if newCtx != ctx {
			t.Error("Expected context to be unchanged")
		}
		if span == nil {
			t.Fatal("Expected span to not be nil")
		}

		// ErrorSpan should not be recording
		if span.IsRecording() {
			t.Error("Expected error span to not be recording")
		}
	})

	t.Run("SpanFromContext", func(t *testing.T) {
		ctx := context.Background()
		span := tracer.SpanFromContext(ctx)

		if span == nil {
			t.Fatal("Expected span to not be nil")
		}
		if span.IsRecording() {
			t.Error("Expected error span to not be recording")
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := tracer.Close()
		if err != testErr {
			t.Errorf("Expected Close() to return test error, got %v", err)
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		metrics := tracer.GetMetrics()
		// Error tracer should return empty metrics
		if metrics.SpansCreated != 0 {
			t.Errorf("Expected SpansCreated to be 0, got %d", metrics.SpansCreated)
		}
	})
}

func TestErrorSpan(t *testing.T) {
	testErr := &testError{message: "test error"}
	tracer := NewErrorTracer(testErr)
	_, span := tracer.StartSpan(context.Background(), "test-span")

	t.Run("AllOperationsAreNoOp", func(t *testing.T) {
		// All these should not panic
		span.SetName("new-name")
		span.SetAttributes(map[string]interface{}{"key": "value"})
		span.SetAttribute("key", "value")
		span.AddEvent("event", map[string]interface{}{"key": "value"})
		span.SetStatus(StatusCodeError, "error")
		span.RecordError(testErr, map[string]interface{}{"key": "value"})
		span.End()

		// Should return empty context
		ctx := span.Context()
		if ctx.TraceID != "" || ctx.SpanID != "" {
			t.Error("Expected empty span context from error span")
		}

		// Should not be recording
		if span.IsRecording() {
			t.Error("Expected error span to not be recording")
		}

		// Should return zero duration
		duration := span.GetDuration()
		if duration != 0 {
			t.Errorf("Expected zero duration from error span, got %v", duration)
		}
	})

	t.Run("ContextWithSpan", func(t *testing.T) {
		ctx := context.Background()
		_, span := tracer.StartSpan(ctx, "test-span")

		// Should not panic and return context unchanged
		newCtx := tracer.ContextWithSpan(ctx, span)
		if newCtx != ctx {
			t.Error("Expected context to be unchanged for error tracer")
		}
	})

	t.Run("String", func(t *testing.T) {
		if errorSpan, ok := span.(*ErrorSpan); ok {
			str := errorSpan.String()
			expected := "ErrorSpan: test error"
			if str != expected {
				t.Errorf("Expected String() to return %s, got %s", expected, str)
			}
		} else {
			t.Error("Expected span to be *ErrorSpan")
		}
	})
}

// Test helper
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// Benchmark tests for noop implementations
func BenchmarkNoopTracerStartSpan(b *testing.B) {
	tracer := NewNoopTracer()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := tracer.StartSpan(ctx, "benchmark-span")
		span.End()
	}
}

func BenchmarkNoopSpanOperations(b *testing.B) {
	tracer := NewNoopTracer()
	ctx := context.Background()
	_, span := tracer.StartSpan(ctx, "benchmark-span")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span.SetAttribute("key", "value")
		span.AddEvent("event", map[string]interface{}{"key": "value"})
		span.SetStatus(StatusCodeOk, "ok")
	}
}

// Concurrency tests
func TestNoopTracerConcurrency(t *testing.T) {
	tracer := NewNoopTracer()
	ctx := context.Background()

	done := make(chan bool)
	numGoroutines := 100
	spansPerGoroutine := 10

	// Concurrent span creation and operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < spansPerGoroutine; j++ {
				spanName := fmt.Sprintf("span-%d-%d", id, j)
				_, span := tracer.StartSpan(ctx, spanName)

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
			t.Fatal("Test timed out waiting for goroutines to complete")
		}
	}

	// Verify metrics
	metrics := tracer.GetMetrics()
	expectedSpans := int64(numGoroutines * spansPerGoroutine)

	if metrics.SpansCreated < expectedSpans {
		t.Errorf("Expected at least %d spans created, got %d", expectedSpans, metrics.SpansCreated)
	}
	if metrics.SpansFinished < expectedSpans {
		t.Errorf("Expected at least %d spans finished, got %d", expectedSpans, metrics.SpansFinished)
	}
}
