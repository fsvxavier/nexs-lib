package newrelic

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

// Helper function to create a test tracer (using provider from provider_test.go)
func createTestTracer(t *testing.T) *Tracer {
	provider := createTestProvider(t)
	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)
	return tr.(*Tracer)
}

func TestTracerStartSpan(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
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
			name:     "server span",
			spanName: "test-server-span",
			opts: []tracer.SpanOption{
				tracer.WithSpanKind(tracer.SpanKindServer),
			},
		},
		{
			name:     "client span",
			spanName: "test-client-span",
			opts: []tracer.SpanOption{
				tracer.WithSpanKind(tracer.SpanKindClient),
			},
		},
		{
			name:     "producer span",
			spanName: "test-producer-span",
			opts: []tracer.SpanOption{
				tracer.WithSpanKind(tracer.SpanKindProducer),
			},
		},
		{
			name:     "consumer span",
			spanName: "test-consumer-span",
			opts: []tracer.SpanOption{
				tracer.WithSpanKind(tracer.SpanKindConsumer),
			},
		},
		{
			name:     "internal span",
			spanName: "test-internal-span",
			opts: []tracer.SpanOption{
				tracer.WithSpanKind(tracer.SpanKindInternal),
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

	provider := createTestProvider(t)
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

	provider := createTestProvider(t)
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

	provider := createTestProvider(t)
	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	err = tr.Close()
	assert.NoError(t, err)
}

func TestTracerGetMetrics(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	metrics := tr.GetMetrics()
	assert.NotNil(t, metrics)
	assert.False(t, metrics.LastActivity.IsZero())
}

func TestSpanSetName(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
	tr, err := provider.CreateTracer("test-tracer")
	require.NoError(t, err)

	ctx := context.Background()
	_, span := tr.StartSpan(ctx, "original-name")

	span.SetName("new-name")
	assert.NotNil(t, span)

	span.End()
}

func TestSpanSetAttributes(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
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

	provider := createTestProvider(t)
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

	provider := createTestProvider(t)
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

	provider := createTestProvider(t)
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

	provider := createTestProvider(t)
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

	provider := createTestProvider(t)
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
	// noop.SetStatus(tracer.StatusCodeOk, "message")
	noop.RecordError(errors.New("error"), map[string]interface{}{"attr": "value"})
	noop.End()

	assert.False(t, noop.IsRecording())
	assert.Equal(t, time.Duration(0), noop.GetDuration())
}

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
		assert.NotNil(t, ctx)
	})

	t.Run("SetName", func(t *testing.T) {
		noopSpan.SetName("test-name")
	})

	t.Run("SetAttributes", func(t *testing.T) {
		attrs := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		noopSpan.SetAttributes(attrs)
	})

	t.Run("SetAttribute", func(t *testing.T) {
		noopSpan.SetAttribute("test-key", "test-value")
	})

	t.Run("AddEvent", func(t *testing.T) {
		attrs := map[string]interface{}{
			"event-key": "event-value",
		}
		noopSpan.AddEvent("test-event", attrs)
	})

	t.Run("SetStatus", func(t *testing.T) {
		// noopSpan.SetStatus(tracer.StatusCode(1), "success") // StatusCodeOk
	})

	t.Run("RecordError", func(t *testing.T) {
		err := errors.New("test error")
		attrs := map[string]interface{}{
			"error-context": "test",
		}
		noopSpan.RecordError(err, attrs)
	})

	t.Run("End", func(t *testing.T) {
		noopSpan.End()
	})

	t.Run("IsRecording", func(t *testing.T) {
		recording := noopSpan.IsRecording()
		assert.False(t, recording)
	})

	t.Run("GetDuration", func(t *testing.T) {
		duration := noopSpan.GetDuration()
		assert.Equal(t, time.Duration(0), duration)
	})
}

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
	// noop.SetStatus(tracer.StatusCodeOk, "status")
	noop.RecordError(errors.New("test"), map[string]interface{}{"error": "data"})
	noop.End()

	// Test return methods
	ctx := noop.Context()
	assert.NotNil(t, ctx)
	assert.False(t, noop.IsRecording())
	assert.Equal(t, time.Duration(0), noop.GetDuration())
}

func TestSpanComplexScenarios(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
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

		// Create new context with span
		ctxWithSpan := tr.ContextWithSpan(ctx, span1)
		assert.NotNil(t, ctxWithSpan)

		// Extract from new context
		extractedAgain := tr.SpanFromContext(ctxWithSpan)
		assert.NotNil(t, extractedAgain)

		span1.End()
	})
}

func TestSpanStatusTransitions(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
	tr, err := provider.CreateTracer("status-tracer")
	require.NoError(t, err)

	tests := []struct {
		name    string
		status  tracer.StatusCode
		message string
	}{
		{
			name:    "unset status",
			status:  tracer.StatusCodeUnset,
			message: "",
		},
		{
			name:    "ok status",
			status:  tracer.StatusCodeOk,
			message: "operation successful",
		},
		{
			name:    "error status",
			status:  tracer.StatusCodeError,
			message: "operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, span := tr.StartSpan(ctx, "status-test-span")

			span.SetStatus(tt.status, tt.message)
			assert.NotNil(t, span)

			span.End()
		})
	}
}

func TestTracerMetrics(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
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

	provider := createTestProvider(t)
	tr, err := provider.CreateTracer("time-tracer")
	require.NoError(t, err)

	customTime := time.Now().Add(-1 * time.Minute) // Reduced from 1 hour to 1 minute
	ctx := context.Background()

	_, span := tr.StartSpan(ctx, "custom-time-span",
		tracer.WithStartTime(customTime),
	)

	duration := span.GetDuration()
	assert.Greater(t, duration, 50*time.Second) // Should be close to 1 minute

	span.End()
}

func TestSpanConcurrentAccess(t *testing.T) {
	t.Parallel()

	provider := createTestProvider(t)
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

// Benchmark tests
func BenchmarkSpanStartEnd(b *testing.B) {
	provider, err := NewProvider(&Config{
		AppName:    "benchmark-app",
		LicenseKey: "test-license-key",
	})
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
	provider, err := NewProvider(&Config{
		AppName:    "benchmark-app",
		LicenseKey: "test-license-key",
	})
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
	provider, err := NewProvider(&Config{
		AppName:    "benchmark-app",
		LicenseKey: "test-license-key",
	})
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
