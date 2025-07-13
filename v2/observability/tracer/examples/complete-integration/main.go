// Package main demonstrates complete integration of all Critical Improvements
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

func main() {
	fmt.Println("=== Complete Integration Example ===")

	// 1. Setup OpenTelemetry
	otelConfig := &tracer.OpenTelemetryConfig{
		ServiceName:   "integration-service",
		Endpoint:      "localhost:4317",
		Insecure:      true,
		Timeout:       30 * time.Second,
		SamplingRatio: 1.0,
	}

	otelTracer, err := tracer.NewOpenTelemetryTracer(otelConfig)
	if err != nil {
		log.Printf("Failed to create tracer: %v", err)
		return
	}
	defer otelTracer.Close()

	// 2. Setup performance optimizations
	perfConfig := tracer.PerformanceConfig{
		EnableSpanPooling: true,
		SpanPoolSize:      100,
		EnableFastPaths:   true,
		EnableZeroAlloc:   true,
	}

	spanPool := tracer.NewSpanPool(perfConfig)

	// 3. Setup error handling
	retryConfig := tracer.DefaultRetryConfig()
	circuitBreakerConfig := tracer.DefaultCircuitBreakerConfig()
	errorHandler := tracer.NewDefaultErrorHandler(retryConfig, circuitBreakerConfig)

	fmt.Println("✅ All components initialized")

	// Demonstrate integration
	ctx := context.Background()

	// Start span with OpenTelemetry
	mainCtx, mainSpan := otelTracer.StartSpan(ctx, "integration_test")
	defer mainSpan.End()

	// Use performance pool
	fastSpan := spanPool.Get()
	fastSpan.SetOperationName("fast_operation")
	fastSpan.SetAttributeFast("optimized", true)
	fastSpan.Start()

	// Simulate operation with error handling
	operation := func() error {
		time.Sleep(10 * time.Millisecond)
		return errors.New("test error")
	}

	err = tracer.RetryWithBackoff(mainCtx, operation, errorHandler, "test")
	if err != nil {
		classification := errorHandler.ClassifyError(err)
		fmt.Printf("Operation failed with classification: %s\n", classification)
	}

	fastSpan.End()
	spanPool.Put(fastSpan)

	// Show metrics
	poolMetrics := spanPool.GetMetrics()
	tracerMetrics := otelTracer.GetMetrics()

	fmt.Printf("Pool: Created=%d, Reused=%d\n", poolMetrics.SpansCreated, poolMetrics.SpansReused)
	fmt.Printf("Tracer: Created=%d, Finished=%d\n", tracerMetrics.SpansCreated, tracerMetrics.SpansFinished)

	fmt.Println("🎉 Integration example completed!")
}
