// Package examples demonstrates how to use the tracer providers
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mocktracerDD "github.com/DataDog/dd-trace-go/v2/ddtrace/mocktracer"
	tracer "github.com/fsvxavier/nexs-lib/observability/tracer"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/datadog"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/newrelic"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/prometheus"
)

func main() {

	isMock := true

	// Example 1: Using Datadog Provider
	fmt.Println("=== Datadog Provider Example ===")
	datadogExample(isMock)
	fmt.Println()
	fmt.Println()
	fmt.Println()

	// Example 2: Using New Relic Provider
	fmt.Println("\n=== New Relic Provider Example ===")
	newRelicExample()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	// Example 3: Using Prometheus Provider
	fmt.Println("\n=== Prometheus Provider Example ===")
	prometheusExample()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	// Example 4: Multiple Providers
	fmt.Println("\n=== Multiple Providers Example ===")
	multipleProvidersExample()
	fmt.Println()
	fmt.Println()
	fmt.Println()
}

func datadogExample(isMock bool) {

	if isMock {
		mocktracerDD.Start()
	}

	// Create Datadog provider with configuration
	config := &datadog.Config{
		ServiceName:     "my-service",
		ServiceVersion:  "1.0.0",
		Environment:     "production",
		AgentHost:       "localhost",
		AgentPort:       8126,
		EnableProfiling: true,
		SampleRate:      1.0,
		Tags: map[string]string{
			"team":    "backend",
			"version": "v1.0.0",
		},
		Debug: false,
	}

	provider := datadog.NewProvider(config)
	defer provider.Shutdown(context.Background())

	// Create tracer
	tr := provider.CreateTracer("datadog-tracer",
		tracer.WithServiceName("my-service"),
		tracer.WithServiceVersion("1.0.0"),
		tracer.WithEnvironment("production"),
	)

	// Create a span
	ctx, span := tr.StartSpan(context.Background(), "example-operation",
		tracer.WithSpanKind(tracer.SpanKindServer),
		tracer.WithAttributes(map[string]interface{}{
			"user_id":    12345,
			"request_id": "req-123",
		}),
	)
	defer span.End()

	// Add attributes and events
	span.SetAttribute("processing_time", 100)
	span.AddEvent("validation_completed", map[string]interface{}{
		"validation_result": "success",
	})

	// Simulate some work
	simulateWork(ctx, tr)

	span.SetStatus(tracer.StatusCodeOk, "Operation completed successfully")
	fmt.Println("Datadog example completed")
}

func newRelicExample() {
	// Create New Relic provider with configuration
	config := &newrelic.Config{
		AppName:           "my-go-app",
		LicenseKey:        "your-license-key-here", // Replace with actual key
		Environment:       "production",
		ServiceVersion:    "1.0.0",
		DistributedTracer: true,
		Enabled:           true,
		LogLevel:          "info",
		Attributes: map[string]interface{}{
			"team":       "backend",
			"datacenter": "us-east-1",
		},
		Labels: map[string]string{
			"environment": "production",
			"service":     "my-service",
		},
	}

	provider, err := newrelic.NewProvider(config)
	if err != nil {
		log.Printf("Failed to create New Relic provider: %v", err)
		return
	}
	defer provider.Shutdown(context.Background())

	// Create tracer
	tr := provider.CreateTracer("newrelic-tracer",
		tracer.WithServiceName("my-service"),
		tracer.WithServiceVersion("1.0.0"),
		tracer.WithEnvironment("production"),
	)

	// Create a span
	ctx, span := tr.StartSpan(context.Background(), "newrelic-operation",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithAttributes(map[string]interface{}{
			"operation_type": "database_query",
			"table_name":     "users",
		}),
	)
	defer span.End()

	// Add more context
	span.SetAttribute("query_duration_ms", 25)
	span.AddEvent("query_executed", map[string]interface{}{
		"rows_affected": 1,
		"query_type":    "SELECT",
	})

	// Simulate database work
	simulateWork(ctx, tr)

	span.SetStatus(tracer.StatusCodeOk, "Database operation completed")
	fmt.Println("New Relic example completed")
}

func prometheusExample() {
	// Create Prometheus provider with configuration
	config := &prometheus.Config{
		ServiceName:    "my-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		Namespace:      "myapp",
		Subsystem:      "traces",
		Labels: map[string]string{
			"datacenter": "us-west-2",
			"team":       "platform",
		},
		EnableDuration:  true,
		EnableErrors:    true,
		EnableActive:    true,
		DurationBuckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
	}

	provider := prometheus.NewProvider(config)
	defer provider.Shutdown(context.Background())

	// Create tracer
	tr := provider.CreateTracer("prometheus-tracer",
		tracer.WithServiceName("my-service"),
		tracer.WithServiceVersion("1.0.0"),
		tracer.WithEnvironment("production"),
	)

	// Create multiple spans to generate metrics
	for i := 0; i < 5; i++ {
		_, span := tr.StartSpan(context.Background(), "prometheus-operation",
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithAttributes(map[string]interface{}{
				"iteration": i,
				"batch_id":  "batch-001",
			}),
		)

		// Simulate work with varying duration
		time.Sleep(time.Duration(i*10) * time.Millisecond)

		span.SetAttribute("work_duration_ms", i*10)

		// Simulate an error on the last iteration
		if i == 4 {
			err := fmt.Errorf("simulated error on iteration %d", i)
			span.RecordError(err, map[string]interface{}{
				"error_context": "final_iteration",
			})
			span.SetStatus(tracer.StatusCodeError, "Operation failed")
		} else {
			span.SetStatus(tracer.StatusCodeOk, "Operation completed")
		}

		span.End()
	}

	fmt.Println("Prometheus example completed - metrics have been recorded")

	// In a real application, you would expose metrics via HTTP endpoint:
	// http.Handle("/metrics", promhttp.HandlerFor(provider.GetRegistry(), promhttp.HandlerOpts{}))
}

func multipleProvidersExample() {
	// You can use multiple providers simultaneously

	// Datadog for APM
	ddProvider := datadog.NewProvider(&datadog.Config{
		ServiceName: "multi-provider-service",
		Environment: "staging",
	})

	// Prometheus for custom metrics
	promProvider := prometheus.NewProvider(&prometheus.Config{
		ServiceName: "multi-provider-service",
		Environment: "staging",
		Namespace:   "multiapp",
	})

	defer func() {
		ddProvider.Shutdown(context.Background())
		promProvider.Shutdown(context.Background())
	}()

	// Create tracers from both providers
	ddTracer := ddProvider.CreateTracer("datadog-tracer")
	promTracer := promProvider.CreateTracer("prometheus-tracer")
	// Use Datadog for detailed APM
	ctx, ddSpan := ddTracer.StartSpan(context.Background(), "multi-provider-operation")
	ddSpan.SetAttribute("provider", "datadog")

	// Use Prometheus for metrics collection
	_, promSpan := promTracer.StartSpan(ctx, "multi-provider-operation")
	promSpan.SetAttribute("provider", "prometheus")

	// Simulate work
	time.Sleep(50 * time.Millisecond)

	// End spans
	promSpan.SetStatus(tracer.StatusCodeOk, "Metrics recorded")
	promSpan.End()

	ddSpan.SetStatus(tracer.StatusCodeOk, "APM trace completed")
	ddSpan.End()

	fmt.Println("Multiple providers example completed")
}

func simulateWork(ctx context.Context, tr tracer.Tracer) {
	// Create a child span
	_, childSpan := tr.StartSpan(ctx, "child-operation",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer childSpan.End()

	// Simulate some processing time
	time.Sleep(50 * time.Millisecond)

	childSpan.SetAttribute("simulated_work", true)
	childSpan.SetStatus(tracer.StatusCodeOk, "Child operation completed")
}
