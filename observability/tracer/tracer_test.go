package tracer_test

import (
	"context"
	"testing"

	tracer "github.com/fsvxavier/nexs-lib/observability/tracer"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/datadog"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/newrelic"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/prometheus"
)

func TestDatadogProvider(t *testing.T) {
	config := &datadog.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
	}

	provider := datadog.NewProvider(config)
	defer provider.Shutdown(context.Background())

	tr := provider.CreateTracer("test-tracer")

	ctx, span := tr.StartSpan(context.Background(), "test-operation",
		tracer.WithSpanKind(tracer.SpanKindServer),
		tracer.WithAttributes(map[string]interface{}{
			"test": "value",
		}),
	)
	defer span.End()

	// Test span operations
	span.SetAttribute("key", "value")
	span.AddEvent("test-event", map[string]interface{}{
		"event_data": "test",
	})

	// Test context operations
	extractedSpan := tr.SpanFromContext(ctx)
	if extractedSpan == nil {
		t.Error("Failed to extract span from context")
	}

	span.SetStatus(tracer.StatusCodeOk, "Test completed")
}

func TestNewRelicProvider(t *testing.T) {
	config := &newrelic.Config{
		AppName:        "test-app",
		LicenseKey:     "test-key",
		Environment:    "test",
		ServiceVersion: "1.0.0",
		Enabled:        false, // Disable to avoid requiring real license key
	}

	provider, err := newrelic.NewProvider(config)
	if err != nil {
		t.Skipf("Skipping New Relic test due to configuration error: %v", err)
	}
	defer provider.Shutdown(context.Background())

	tr := provider.CreateTracer("test-tracer")

	_, span := tr.StartSpan(context.Background(), "test-operation",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	span.SetAttribute("test", "value")
	span.SetStatus(tracer.StatusCodeOk, "Test completed")

	// Verify span context
	spanCtx := span.Context()
	if spanCtx.TraceID == "" {
		t.Error("Trace ID should not be empty")
	}
}

func TestPrometheusProvider(t *testing.T) {
	config := &prometheus.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Namespace:      "test",
		Subsystem:      "tracer",
	}

	provider := prometheus.NewProvider(config)
	defer provider.Shutdown(context.Background())

	tr := provider.CreateTracer("test-tracer")

	_, span := tr.StartSpan(context.Background(), "test-operation",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	span.SetAttribute("test", "value")
	span.SetStatus(tracer.StatusCodeOk, "Test completed")

	// Test metrics recording
	registry := provider.GetRegistry()
	if registry == nil {
		t.Error("Registry should not be nil")
	}

	// Verify span context
	spanCtx := span.Context()
	if spanCtx.TraceID == "" {
		t.Error("Trace ID should not be empty")
	}
}

func TestMultipleProviders(t *testing.T) {
	// Test using multiple providers simultaneously
	ddProvider := datadog.NewProvider(&datadog.Config{
		ServiceName: "test-service",
		Environment: "test",
	})
	defer ddProvider.Shutdown(context.Background())

	promProvider := prometheus.NewProvider(&prometheus.Config{
		ServiceName: "test-service",
		Environment: "test",
		Namespace:   "test",
	})
	defer promProvider.Shutdown(context.Background())

	ddTracer := ddProvider.CreateTracer("datadog-tracer")
	promTracer := promProvider.CreateTracer("prometheus-tracer")

	// Create spans with both tracers
	ctx, ddSpan := ddTracer.StartSpan(context.Background(), "multi-provider-test")
	_, promSpan := promTracer.StartSpan(ctx, "multi-provider-test")

	ddSpan.SetAttribute("provider", "datadog")
	promSpan.SetAttribute("provider", "prometheus")

	promSpan.End()
	ddSpan.End()

	t.Log("Multiple providers test completed successfully")
}

func TestSpanOptions(t *testing.T) {
	provider := prometheus.NewProvider(&prometheus.Config{
		ServiceName: "test-service",
		Namespace:   "test",
	})

	tr := provider.CreateTracer("test-tracer")

	// Test all span options
	ctx, span := tr.StartSpan(context.Background(), "test-span",
		tracer.WithSpanKind(tracer.SpanKindServer),
		tracer.WithAttributes(map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}),
	)
	defer span.End()

	// Test span operations
	span.SetName("updated-name")
	span.SetAttribute("dynamic", "attribute")
	span.AddEvent("test-event", map[string]interface{}{
		"event_key": "event_value",
	})

	// Test error recording
	err := &testError{message: "test error"}
	span.RecordError(err, map[string]interface{}{
		"error_context": "test",
	})

	span.SetStatus(tracer.StatusCodeError, "Test error occurred")

	// Verify span is recording
	if !span.IsRecording() {
		t.Error("Span should be recording")
	}

	// Get span context
	spanCtx := span.Context()
	if spanCtx.TraceID == "" || spanCtx.SpanID == "" {
		t.Error("Span context should have valid IDs")
	}

	// Test context with span
	newCtx := tr.ContextWithSpan(ctx, span)
	if newCtx == nil {
		t.Error("Context with span should not be nil")
	}
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func BenchmarkSpanCreation(b *testing.B) {
	provider := prometheus.NewProvider(&prometheus.Config{
		ServiceName: "benchmark-service",
		Namespace:   "bench",
	})

	tr := provider.CreateTracer("benchmark-tracer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, span := tr.StartSpan(context.Background(), "benchmark-operation")
		span.SetAttribute("iteration", i)
		span.End()
	}
}
