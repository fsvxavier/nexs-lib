// Package main demonstrates OpenTelemetry integration with nexs-lib tracer
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

func main() {
	fmt.Println("=== OpenTelemetry Integration Example ===")

	// Create OpenTelemetry configuration
	config := &tracer.OpenTelemetryConfig{
		ServiceName:      "example-service",
		ServiceVersion:   "1.0.0",
		ServiceNamespace: "examples",
		Endpoint:         "localhost:4317", // OTLP gRPC endpoint
		Insecure:         true,             // For local development
		Timeout:          30 * time.Second,
		BatchTimeout:     5 * time.Second,
		MaxExportBatch:   512,
		MaxQueueSize:     2048,
		SamplingRatio:    1.0, // Sample all traces for demo
		Propagators:      []string{"tracecontext", "baggage"},
		ResourceAttrs: map[string]string{
			"environment": "development",
			"team":        "platform",
			"version":     "v1.0.0",
		},
	}

	// Create OpenTelemetry tracer
	otelTracer, err := tracer.NewOpenTelemetryTracer(config)
	if err != nil {
		log.Fatalf("Failed to create OpenTelemetry tracer: %v", err)
	}
	defer otelTracer.Close()

	fmt.Println("✅ OpenTelemetry tracer created successfully")

	// Example 1: Basic span creation
	basicSpanExample(otelTracer)

	// Example 2: Nested spans with attributes
	nestedSpansExample(otelTracer)

	// Example 3: Error handling and recording
	errorHandlingExample(otelTracer)

	// Example 4: Context propagation
	contextPropagationExample(otelTracer)

	// Example 5: Events and structured logging
	eventsExample(otelTracer)

	fmt.Println("\n🎉 All OpenTelemetry examples completed successfully!")
	fmt.Println("📊 Check your OpenTelemetry collector/backend for traces")
}

func basicSpanExample(t tracer.Tracer) {
	fmt.Println("\n--- Basic Span Example ---")

	ctx := context.Background()

	// Start a root span
	ctx, span := t.StartSpan(ctx, "user-registration",
		tracer.WithSpanKind(tracer.SpanKindServer),
		tracer.WithSpanAttributes(map[string]interface{}{
			"user.id":    "12345",
			"user.email": "user@example.com",
			"operation":  "register",
		}),
	)
	defer span.End()

	// Simulate work
	time.Sleep(50 * time.Millisecond)

	// Set additional attributes
	span.SetAttribute("user.verified", true)
	span.SetAttribute("registration.duration_ms", 50)

	// Set successful status
	span.SetStatus(tracer.StatusCodeOk, "User registered successfully")

	fmt.Println("✅ Basic span created with attributes and status")
}

func nestedSpansExample(t tracer.Tracer) {
	fmt.Println("\n--- Nested Spans Example ---")

	ctx := context.Background()

	// Parent span: HTTP request handler
	ctx, parentSpan := t.StartSpan(ctx, "http-request",
		tracer.WithSpanKind(tracer.SpanKindServer),
		tracer.WithSpanAttributes(map[string]interface{}{
			"http.method":     "POST",
			"http.url":        "/api/orders",
			"http.user_agent": "MyApp/1.0",
		}),
	)
	defer parentSpan.End()

	// Child span: Database query
	ctx, dbSpan := t.StartSpan(ctx, "database-query",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"db.system":    "postgresql",
			"db.operation": "INSERT",
			"db.table":     "orders",
		}),
	)

	// Simulate database work
	time.Sleep(30 * time.Millisecond)
	dbSpan.SetAttribute("db.rows_affected", 1)
	dbSpan.SetStatus(tracer.StatusCodeOk, "Query executed successfully")
	dbSpan.End()

	// Child span: External API call
	ctx, apiSpan := t.StartSpan(ctx, "payment-service",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"http.method":      "POST",
			"http.url":         "https://api.payment.com/charge",
			"payment.amount":   99.99,
			"payment.currency": "USD",
		}),
	)

	// Simulate API call
	time.Sleep(100 * time.Millisecond)
	apiSpan.SetAttribute("http.status_code", 200)
	apiSpan.SetAttribute("payment.transaction_id", "txn_12345")
	apiSpan.SetStatus(tracer.StatusCodeOk, "Payment processed")
	apiSpan.End()

	// Complete parent span
	parentSpan.SetAttribute("http.status_code", 201)
	parentSpan.SetAttribute("order.id", "order_67890")
	parentSpan.SetStatus(tracer.StatusCodeOk, "Order created successfully")

	fmt.Println("✅ Nested spans created: HTTP -> DB + Payment API")
}

func errorHandlingExample(t tracer.Tracer) {
	fmt.Println("\n--- Error Handling Example ---")

	ctx := context.Background()

	// Start span that will encounter an error
	ctx, span := t.StartSpan(ctx, "user-authentication",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"auth.method": "password",
			"user.id":     "user123",
		}),
	)
	defer span.End()

	// Simulate error condition
	err := fmt.Errorf("invalid credentials: password does not match")

	// Record the error with additional context
	span.RecordError(err, map[string]interface{}{
		"error.type":        "authentication_failed",
		"error.retry_count": 3,
		"error.user_agent":  "MyApp/1.0",
	})

	// Set error status
	span.SetStatus(tracer.StatusCodeError, "Authentication failed")

	// Add event for debugging
	span.AddEvent("authentication_attempt", map[string]interface{}{
		"attempt_number": 3,
		"reason":         "invalid_password",
		"timestamp":      time.Now().Unix(),
	})

	fmt.Println("✅ Error recorded with context and debugging information")
}

func contextPropagationExample(t tracer.Tracer) {
	fmt.Println("\n--- Context Propagation Example ---")

	ctx := context.Background()

	// Start root span
	ctx, rootSpan := t.StartSpan(ctx, "order-processing",
		tracer.WithSpanKind(tracer.SpanKindServer),
		tracer.WithSpanAttributes(map[string]interface{}{
			"order.id":    "order_123",
			"customer.id": "cust_456",
		}),
	)
	defer rootSpan.End()

	// Extract span from context to verify propagation
	extractedSpan := t.SpanFromContext(ctx)
	if extractedSpan != nil {
		fmt.Println("✅ Span successfully extracted from context")

		// Get span context for propagation
		spanCtx := extractedSpan.Context()
		fmt.Printf("📍 Trace ID: %s\n", spanCtx.TraceID)
		fmt.Printf("📍 Span ID: %s\n", spanCtx.SpanID)
	}

	// Create new context with span for downstream services
	newCtx := t.ContextWithSpan(context.Background(), rootSpan)

	// Simulate calling downstream service
	processOrderStep(t, newCtx, "validate-inventory")
	processOrderStep(t, newCtx, "reserve-items")
	processOrderStep(t, newCtx, "calculate-shipping")

	rootSpan.SetStatus(tracer.StatusCodeOk, "Order processed successfully")
	fmt.Println("✅ Context propagation verified across multiple steps")
}

func processOrderStep(t tracer.Tracer, ctx context.Context, stepName string) {
	_, span := t.StartSpan(ctx, stepName,
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"step.name":  stepName,
			"step.order": stepName,
		}),
	)
	defer span.End()

	// Simulate work
	time.Sleep(20 * time.Millisecond)
	span.SetStatus(tracer.StatusCodeOk, fmt.Sprintf("%s completed", stepName))
}

func eventsExample(t tracer.Tracer) {
	fmt.Println("\n--- Events and Structured Logging Example ---")

	ctx := context.Background()

	ctx, span := t.StartSpan(ctx, "file-processing",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"file.name": "data.csv",
			"file.size": 1024000,
			"file.type": "csv",
		}),
	)
	defer span.End()

	// Add structured events during processing
	span.AddEvent("file_opened", map[string]interface{}{
		"timestamp":        time.Now().Unix(),
		"file.path":        "/uploads/data.csv",
		"file.permissions": "0644",
	})

	time.Sleep(30 * time.Millisecond)

	span.AddEvent("validation_started", map[string]interface{}{
		"validator.type":    "csv_schema",
		"validator.version": "v2.1.0",
		"expected_columns":  12,
	})

	time.Sleep(50 * time.Millisecond)

	span.AddEvent("processing_progress", map[string]interface{}{
		"rows_processed":   1000,
		"rows_total":       5000,
		"progress_percent": 20.0,
		"errors_found":     2,
	})

	time.Sleep(40 * time.Millisecond)

	span.AddEvent("file_processed", map[string]interface{}{
		"total_rows":             5000,
		"valid_rows":             4998,
		"invalid_rows":           2,
		"processing_duration_ms": 120,
		"output_file":            "/processed/data_clean.csv",
	})

	// Update span with final attributes
	span.SetAttribute("processing.total_rows", 5000)
	span.SetAttribute("processing.success_rate", 99.96)
	span.SetStatus(tracer.StatusCodeOk, "File processed successfully")

	fmt.Println("✅ Structured events added throughout processing lifecycle")
}
