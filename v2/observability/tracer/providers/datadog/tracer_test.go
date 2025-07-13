package datadog

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracerStartSpan(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)
	require.NotNil(t, tr)

	tests := []struct {
		name     string
		spanName string
		opts     []tracer.SpanOption
	}{
		{
			name:     "basic span",
			spanName: "test-span",
			opts:     nil,
		},
		{
			name:     "span with attributes",
			spanName: "test-span-with-attrs",
			opts: []tracer.SpanOption{
				tracer.WithSpanAttributes(map[string]interface{}{
					"key1": "value1",
					"key2": 42,
				}),
			},
		},
		{
			name:     "span with kind",
			spanName: "test-span-with-kind",
			opts: []tracer.SpanOption{
				tracer.WithSpanKind(tracer.SpanKindServer),
			},
		},
		{
			name:     "span with start time",
			spanName: "test-span-with-start-time",
			opts: []tracer.SpanOption{
				tracer.WithStartTime(time.Now().Add(-1 * time.Minute)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			newCtx, span := tr.StartSpan(ctx, tt.spanName, tt.opts...)

			assert.NotNil(t, newCtx)
			assert.NotNil(t, span)
			assert.True(t, span.IsRecording())

			// Test span context
			spanCtx := span.Context()
			assert.NotEmpty(t, spanCtx.TraceID)
			assert.NotEmpty(t, spanCtx.SpanID)

			// End the span
			span.End()
			assert.False(t, span.IsRecording())
			assert.Greater(t, span.GetDuration(), time.Duration(0))
		})
	}
}

func TestTracerSpanFromContext(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	tests := []struct {
		name       string
		setupCtx   func() context.Context
		expectNoop bool
	}{
		{
			name: "context with span",
			setupCtx: func() context.Context {
				ctx, _ := tr.StartSpan(context.Background(), "test-span")
				return ctx
			},
			expectNoop: false,
		},
		{
			name: "context without span",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectNoop: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := tt.setupCtx()
			span := tr.SpanFromContext(ctx)

			assert.NotNil(t, span)
			if tt.expectNoop {
				assert.IsType(t, &NoopSpan{}, span)
				assert.False(t, span.IsRecording())
			} else {
				assert.IsType(t, &Span{}, span)
			}
		})
	}
}

func TestTracerContextWithSpan(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "test-span")

	newCtx := tr.ContextWithSpan(ctx, span)
	assert.NotNil(t, newCtx)
	assert.NotEqual(t, ctx, newCtx)

	// Test with noop span
	noopSpan := &NoopSpan{}
	noopCtx := tr.ContextWithSpan(ctx, noopSpan)
	assert.Equal(t, ctx, noopCtx)

	span.End()
}

func TestTracerClose(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	err = tr.Close()
	assert.NoError(t, err)
}

func TestTracerGetMetrics(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	metrics := tr.GetMetrics()
	assert.NotNil(t, metrics)
	assert.False(t, metrics.LastActivity.IsZero())
}

func TestSpanSetName(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "original-name")

	span.SetName("new-name")
	// No easy way to verify the name was changed in Datadog tracer
	// but we can ensure the call doesn't panic
	assert.NotNil(t, span)

	span.End()
}

func TestSpanSetAttributes(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "test-span")

	// Test SetAttributes
	attributes := map[string]interface{}{
		"string_attr": "value",
		"int_attr":    42,
		"bool_attr":   true,
		"float_attr":  3.14,
	}
	span.SetAttributes(attributes)

	// Test SetAttribute
	span.SetAttribute("single_attr", "single_value")

	assert.NotNil(t, span)
	span.End()
}

func TestSpanAddEvent(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "test-span")

	span.AddEvent("test-event", map[string]interface{}{
		"event_attr": "event_value",
		"count":      1,
	})

	assert.NotNil(t, span)
	span.End()
}

func TestSpanSetStatus(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		statusCode tracer.StatusCode
		message    string
	}{
		{
			name:       "ok status",
			statusCode: tracer.StatusCodeOk,
			message:    "operation successful",
		},
		{
			name:       "error status",
			statusCode: tracer.StatusCodeError,
			message:    "operation failed",
		},
		{
			name:       "unset status",
			statusCode: tracer.StatusCodeUnset,
			message:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, span := tr.StartSpan(ctx, "test-span")

			span.SetStatus(tt.statusCode, tt.message)

			assert.NotNil(t, span)
			span.End()
		})
	}
}

func TestSpanRecordError(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "test-span")

	testError := errors.New("test error")
	attributes := map[string]interface{}{
		"error_code":    "TEST_ERROR",
		"error_context": "testing",
	}

	span.RecordError(testError, attributes)

	assert.NotNil(t, span)
	span.End()
}

func TestSpanDuration(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "test-span")

	// Duration while recording
	duration1 := span.GetDuration()
	assert.Greater(t, duration1, time.Duration(0))

	// Small delay
	time.Sleep(1 * time.Millisecond)

	// Duration should be larger
	duration2 := span.GetDuration()
	assert.Greater(t, duration2, duration1)

	// End span
	span.End()

	// Duration should be fixed after end
	duration3 := span.GetDuration()
	assert.Greater(t, duration3, duration2)

	// Wait and check duration is still the same
	time.Sleep(1 * time.Millisecond)
	duration4 := span.GetDuration()
	assert.Equal(t, duration3, duration4)
}

func TestNoopSpan(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	noop := &NoopSpan{tracer: tr.(*Tracer)}

	// Test all methods don't panic and return appropriate values
	ctx := noop.Context()
	assert.Empty(t, ctx.TraceID)
	assert.Empty(t, ctx.SpanID)

	noop.SetName("test")
	noop.SetAttributes(map[string]interface{}{"key": "value"})
	noop.SetAttribute("key", "value")
	noop.AddEvent("event", map[string]interface{}{"attr": "value"})
	noop.SetStatus(tracer.StatusCodeOk, "message")
	noop.RecordError(errors.New("error"), map[string]interface{}{"attr": "value"})
	noop.End()

	assert.False(t, noop.IsRecording())
	assert.Equal(t, time.Duration(0), noop.GetDuration())
}

// TestNoopSpan tests all NoopSpan methods for coverage
func TestNoopSpanCoverage(t *testing.T) {
	provider := createTestProvider(t)
	tracer, err := provider.CreateTracer("test-service")
	require.NoError(t, err)

	// Create a context without a span to get NoopSpan
	ctx := context.Background()
	noopSpan := tracer.SpanFromContext(ctx)

	// Test all NoopSpan methods
	t.Run("Context", func(t *testing.T) {
		ctx := noopSpan.Context()
		assert.NotNil(t, ctx, "Context should not be nil")
	})

	t.Run("SetName", func(t *testing.T) {
		noopSpan.SetName("test-name") // Should not panic
	})

	t.Run("SetAttributes", func(t *testing.T) {
		attrs := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		noopSpan.SetAttributes(attrs) // Should not panic
	})

	t.Run("SetAttribute", func(t *testing.T) {
		noopSpan.SetAttribute("test-key", "test-value") // Should not panic
	})

	t.Run("AddEvent", func(t *testing.T) {
		attrs := map[string]interface{}{
			"event-key": "event-value",
		}
		noopSpan.AddEvent("test-event", attrs) // Should not panic
	})

	t.Run("SetStatus", func(t *testing.T) {
		noopSpan.SetStatus(1, "success") // Should not panic, using numeric value
	})

	t.Run("RecordError", func(t *testing.T) {
		err := errors.New("test error")
		attrs := map[string]interface{}{
			"error-context": "test",
		}
		noopSpan.RecordError(err, attrs) // Should not panic
	})

	t.Run("End", func(t *testing.T) {
		noopSpan.End() // Should not panic
	})

	t.Run("IsRecording", func(t *testing.T) {
		recording := noopSpan.IsRecording()
		assert.False(t, recording, "NoopSpan should not be recording")
	})

	t.Run("GetDuration", func(t *testing.T) {
		duration := noopSpan.GetDuration()
		assert.Equal(t, time.Duration(0), duration, "NoopSpan duration should be 0")
	})
}

// TestNoopSpanCoverageComplete tests all uncovered NoopSpan methods
func TestNoopSpanCoverageComplete(t *testing.T) {
	provider := createTestProvider(t)
	tracer, err := provider.CreateTracer("noop-test-service")
	require.NoError(t, err)

	// Get NoopSpan by calling SpanFromContext without a span
	ctx := context.Background()
	noopSpan := tracer.SpanFromContext(ctx)

	// Test that we actually got a NoopSpan
	assert.IsType(t, &NoopSpan{}, noopSpan, "Should return NoopSpan from empty context")

	// Call all NoopSpan methods through the interface to ensure they're covered
	// These calls will hit the NoopSpan implementation
	noopSpan.SetName("test-name")

	noopSpan.SetAttributes(map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})

	noopSpan.SetAttribute("test-key", "test-value")

	noopSpan.AddEvent("test-event", map[string]interface{}{
		"event-key": "event-value",
	})

	noopSpan.SetStatus(1, "test status") // Using numeric value to avoid import issues

	noopSpan.RecordError(errors.New("test error"), map[string]interface{}{
		"error-context": "test",
	})

	noopSpan.End()

	// Test methods that return values
	assert.False(t, noopSpan.IsRecording())
	assert.Equal(t, time.Duration(0), noopSpan.GetDuration())
}

// TestNoopSpanDirectCoverage tests NoopSpan methods directly
func TestNoopSpanDirectCoverage(t *testing.T) {
	provider := createTestProvider(t)
	tracer, err := provider.CreateTracer("noop-direct-test")
	require.NoError(t, err)

	// Create NoopSpan directly to ensure we call its methods
	noop := &NoopSpan{tracer: tracer.(*Tracer)}

	// Test all methods to ensure coverage
	noop.SetName("direct-test")
	noop.SetAttributes(map[string]interface{}{"key": "value"})
	noop.SetAttribute("single", "value")
	noop.AddEvent("event", map[string]interface{}{"event": "data"})
	noop.SetStatus(1, "status") // Using numeric value
	noop.RecordError(errors.New("test"), map[string]interface{}{"error": "data"})
	noop.End()

	// Test return methods
	ctx := noop.Context()
	assert.NotNil(t, ctx)
	assert.False(t, noop.IsRecording())
	assert.Equal(t, time.Duration(0), noop.GetDuration())
}

func TestConvertSpanKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    tracer.SpanKind
		expected string
	}{
		{tracer.SpanKindServer, "web"},
		{tracer.SpanKindClient, "http"},
		{tracer.SpanKindProducer, "queue"},
		{tracer.SpanKindConsumer, "queue"},
		{tracer.SpanKindInternal, "custom"},
		{tracer.SpanKind(999), "custom"}, // Unknown kind
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("kind_%d", tt.input), func(t *testing.T) {
			result := convertSpanKind(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkSpanStartEnd(b *testing.B) {
	provider, err := NewProvider(DefaultConfig())
	if err != nil {
		b.Fatal(err)
	}

	tr, err := provider.CreateTracer("benchmark-tracer")
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := tr.StartSpan(ctx, "benchmark-span")
		span.End()
	}
}

func BenchmarkSpanSetAttributes(b *testing.B) {
	provider, err := NewProvider(DefaultConfig())
	if err != nil {
		b.Fatal(err)
	}

	tr, err := provider.CreateTracer("benchmark-tracer")
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "benchmark-span")
	defer span.End()

	attributes := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span.SetAttributes(attributes)
	}
}

func BenchmarkSpanAddEvent(b *testing.B) {
	provider, err := NewProvider(DefaultConfig())
	if err != nil {
		b.Fatal(err)
	}

	tr, err := provider.CreateTracer("benchmark-tracer")
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "benchmark-span")
	defer span.End()

	eventAttrs := map[string]interface{}{
		"event_key": "event_value",
		"count":     1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span.AddEvent("benchmark-event", eventAttrs)
	}
}

// Additional comprehensive tests for better coverage

func TestSpanComplexScenarios(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("complex-tracer")
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("span with all features", func(t *testing.T) {
		_, span := tr.StartSpan(ctx, "complex-span",
			tracer.WithSpanAttributes(map[string]interface{}{
				"user_id":   12345,
				"operation": "test",
			}),
			tracer.WithSpanKind(tracer.SpanKindServer),
		)

		// Test all methods
		span.SetName("updated-name")
		span.SetAttribute("single", "value")
		span.SetAttributes(map[string]interface{}{
			"multiple": "attributes",
			"count":    42,
		})
		span.AddEvent("test-event", map[string]interface{}{
			"event_data": "important",
		})
		span.SetStatus(tracer.StatusCodeOk, "success")

		// Record error
		testErr := fmt.Errorf("test error")
		span.RecordError(testErr, map[string]interface{}{
			"error_code": "TEST_001",
		})

		// Test context
		spanCtx := span.Context()
		assert.NotEmpty(t, spanCtx.TraceID)
		assert.NotEmpty(t, spanCtx.SpanID)
		assert.True(t, spanCtx.Flags.IsSampled())

		// Test duration while running
		duration1 := span.GetDuration()
		assert.Greater(t, duration1, time.Duration(0))

		// End span
		span.End()
		assert.False(t, span.IsRecording())

		// Test duration after end
		duration2 := span.GetDuration()
		assert.GreaterOrEqual(t, duration2, duration1)
	})

	t.Run("span context propagation", func(t *testing.T) {
		// Start initial span
		newCtx, span1 := tr.StartSpan(ctx, "parent-span")

		// Extract span from context
		extractedSpan := tr.SpanFromContext(newCtx)
		assert.NotNil(t, extractedSpan)
		assert.IsType(t, &Span{}, extractedSpan)

		// Create new context with span
		ctxWithSpan := tr.ContextWithSpan(ctx, span1)
		assert.NotNil(t, ctxWithSpan)

		// Extract from new context
		extractedAgain := tr.SpanFromContext(ctxWithSpan)
		assert.NotNil(t, extractedAgain)
		assert.Equal(t, span1, extractedAgain)

		span1.End()
	})

	t.Run("noop span behavior", func(t *testing.T) {
		noopSpan := &NoopSpan{tracer: tr.(*Tracer)}

		// Test all noop methods
		spanCtx := noopSpan.Context()
		assert.Empty(t, spanCtx.TraceID)
		assert.Empty(t, spanCtx.SpanID)

		noopSpan.SetName("test")
		noopSpan.SetAttributes(map[string]interface{}{"key": "value"})
		noopSpan.SetAttribute("key", "value")
		noopSpan.AddEvent("event", map[string]interface{}{"attr": "value"})
		noopSpan.SetStatus(tracer.StatusCodeOk, "message")
		noopSpan.RecordError(fmt.Errorf("error"), map[string]interface{}{"attr": "value"})
		noopSpan.End()

		assert.False(t, noopSpan.IsRecording())
		assert.Equal(t, time.Duration(0), noopSpan.GetDuration())
	})
}

func TestSpanStatusTransitions(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("status-tracer")
	require.NoError(t, err)

	tests := []struct {
		name       string
		status     tracer.StatusCode
		message    string
		expectAttr map[string]interface{}
	}{
		{
			name:    "unset status",
			status:  tracer.StatusCodeUnset,
			message: "",
			expectAttr: map[string]interface{}{
				"error.status": "unset",
			},
		},
		{
			name:    "ok status",
			status:  tracer.StatusCodeOk,
			message: "operation successful",
			expectAttr: map[string]interface{}{
				"error":        false,
				"error.status": "ok",
			},
		},
		{
			name:    "error status",
			status:  tracer.StatusCodeError,
			message: "operation failed",
			expectAttr: map[string]interface{}{
				"error":         true,
				"error.status":  "error",
				"error.message": "operation failed",
			},
		},
		{
			name:    "error status without message",
			status:  tracer.StatusCodeError,
			message: "",
			expectAttr: map[string]interface{}{
				"error":        true,
				"error.status": "error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, span := tr.StartSpan(ctx, "status-test-span")

			span.SetStatus(tt.status, tt.message)

			// Access the underlying span to check attributes
			if ddSpan, ok := span.(*Span); ok {
				for key, expectedValue := range tt.expectAttr {
					actualValue, exists := ddSpan.attrs[key]
					assert.True(t, exists, "Attribute %s should exist", key)
					assert.Equal(t, expectedValue, actualValue, "Attribute %s should have correct value", key)
				}
			}

			span.End()
		})
	}
}

func TestTracerMetrics(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("metrics-tracer")
	require.NoError(t, err)

	// Initial metrics
	metrics := tr.GetMetrics()
	assert.Equal(t, int64(0), metrics.SpansCreated)
	assert.Equal(t, int64(0), metrics.SpansFinished)

	ctx := context.Background()

	// Create and end several spans
	for i := 0; i < 5; i++ {
		_, span := tr.StartSpan(ctx, fmt.Sprintf("span-%d", i))
		span.End()
	}

	// Check updated metrics
	updatedMetrics := tr.GetMetrics()
	assert.Equal(t, int64(5), updatedMetrics.SpansCreated)
	assert.Equal(t, int64(5), updatedMetrics.SpansFinished)
	assert.False(t, updatedMetrics.LastActivity.IsZero())
}

func TestSpanWithCustomStartTime(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("time-tracer")
	require.NoError(t, err)

	customTime := time.Now().Add(-1 * time.Hour)
	ctx := context.Background()

	_, span := tr.StartSpan(ctx, "custom-time-span",
		tracer.WithStartTime(customTime),
	)

	duration := span.GetDuration()
	assert.Greater(t, duration, 50*time.Minute) // Should be close to 1 hour

	span.End()
}

func TestConvertSpanKindEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    tracer.SpanKind
		expected string
	}{
		{tracer.SpanKindServer, "web"},
		{tracer.SpanKindClient, "http"},
		{tracer.SpanKindProducer, "queue"},
		{tracer.SpanKindConsumer, "queue"},
		{tracer.SpanKindInternal, "custom"},
		{tracer.SpanKind(999), "custom"}, // Unknown kind
		{tracer.SpanKind(-1), "custom"},  // Negative kind
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("kind_%d", tt.input), func(t *testing.T) {
			result := convertSpanKind(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSpanConcurrentAccess(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(DefaultConfig())
	require.NoError(t, err)

	tr, err := provider.CreateTracer("concurrent-tracer")
	require.NoError(t, err)

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "concurrent-span")

	// Test concurrent access to span methods
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				span.SetAttribute(fmt.Sprintf("key-%d-%d", id, j), fmt.Sprintf("value-%d-%d", id, j))
				span.AddEvent(fmt.Sprintf("event-%d-%d", id, j), map[string]interface{}{
					"id": id,
					"op": j,
				})
			}
		}(i)
	}

	wg.Wait()
	span.End()

	assert.False(t, span.IsRecording())
}
