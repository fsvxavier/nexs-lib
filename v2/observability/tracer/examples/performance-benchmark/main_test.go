package main

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/prometheus"
)

func BenchmarkBusinessOperations(b *testing.B) {
	// Setup
	config := &prometheus.Config{
		ServiceName:        "benchmark-test",
		Namespace:          "test",
		MaxCardinality:     1000,
		CollectionInterval: 10 * time.Second,
		BucketBoundaries:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}

	provider, err := prometheus.NewProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	tr, err := provider.CreateTracer("benchmark")
	if err != nil {
		b.Fatalf("Failed to create tracer: %v", err)
	}

	monitor := NewPerformanceMonitor(tr)

	b.ResetTimer()

	b.Run("QuickLookup", func(b *testing.B) {
		ctx := context.Background()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				monitor.SimulateBusinessOperation(ctx, "quick_lookup")
			}
		})
	})

	b.Run("StandardProcessing", func(b *testing.B) {
		ctx := context.Background()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				monitor.SimulateBusinessOperation(ctx, "standard_processing")
			}
		})
	})

	b.Run("ComplexAnalysis", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			monitor.SimulateBusinessOperation(ctx, "complex_analysis")
		}
	})

	b.Run("DataAggregation", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			monitor.SimulateBusinessOperation(ctx, "data_aggregation")
		}
	})
}

func BenchmarkPerformanceMonitor(b *testing.B) {
	config := &prometheus.Config{
		ServiceName:        "monitor-benchmark",
		Namespace:          "test",
		MaxCardinality:     1000,
		CollectionInterval: 10 * time.Second,
		BucketBoundaries:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}

	provider, err := prometheus.NewProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	tr, err := provider.CreateTracer("monitor")
	if err != nil {
		b.Fatalf("Failed to create tracer: %v", err)
	}

	monitor := NewPerformanceMonitor(tr)

	b.Run("RecordRequest", func(b *testing.B) {
		ctx := context.Background()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				monitor.RecordRequest(ctx, 100*time.Millisecond, true)
			}
		})
	})

	b.Run("GetCurrentMetrics", func(b *testing.B) {
		ctx := context.Background()
		// Add some data first
		for i := 0; i < 1000; i++ {
			monitor.RecordRequest(ctx, time.Duration(i)*time.Millisecond, true)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			monitor.GetCurrentMetrics()
		}
	})
}

func TestLoadTestScenario(t *testing.T) {
	config := &prometheus.Config{
		ServiceName:        "test-scenario",
		Namespace:          "test",
		MaxCardinality:     1000,
		CollectionInterval: 10 * time.Second,
		BucketBoundaries:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}

	provider, err := prometheus.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	tr, err := provider.CreateTracer("test")
	if err != nil {
		t.Fatalf("Failed to create tracer: %v", err)
	}

	monitor := NewPerformanceMonitor(tr)

	scenario := LoadTestScenario{
		Name:            "test_scenario",
		ConcurrentUsers: 5,
		Duration:        2 * time.Second,
		RampUpTime:      500 * time.Millisecond,
		RequestRate:     10,
	}

	ctx := context.Background()
	err = monitor.RunLoadTest(ctx, scenario)
	if err != nil {
		t.Errorf("Load test failed: %v", err)
	}

	metrics := monitor.GetCurrentMetrics()
	if metrics.ThroughputRPS <= 0 {
		t.Errorf("Expected positive throughput, got %f", metrics.ThroughputRPS)
	}

	if metrics.SuccessRate < 80 {
		t.Errorf("Expected success rate >= 80%%, got %f%%", metrics.SuccessRate)
	}
}

func TestPerformanceMonitorMetrics(t *testing.T) {
	config := &prometheus.Config{
		ServiceName:        "test-metrics",
		Namespace:          "test",
		MaxCardinality:     1000,
		CollectionInterval: 10 * time.Second,
		BucketBoundaries:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}

	provider, err := prometheus.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	tr, err := provider.CreateTracer("metrics")
	if err != nil {
		t.Fatalf("Failed to create tracer: %v", err)
	}

	monitor := NewPerformanceMonitor(tr)
	ctx := context.Background()

	// Record some test data
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
	}

	for _, latency := range latencies {
		monitor.RecordRequest(ctx, latency, true)
	}

	// Record one error
	monitor.RecordRequest(ctx, 50*time.Millisecond, false)

	metrics := monitor.GetCurrentMetrics()

	// Verify metrics calculation
	if metrics.SuccessRate < 80 || metrics.SuccessRate > 90 {
		t.Errorf("Expected success rate around 83.33%%, got %f%%", metrics.SuccessRate)
	}

	if metrics.ErrorRate < 15 || metrics.ErrorRate > 20 {
		t.Errorf("Expected error rate around 16.67%%, got %f%%", metrics.ErrorRate)
	}

	if metrics.P50Latency <= 0 {
		t.Errorf("Expected positive P50 latency, got %v", metrics.P50Latency)
	}

	if metrics.P95Latency <= metrics.P50Latency {
		t.Errorf("Expected P95 latency > P50 latency, got P95=%v P50=%v",
			metrics.P95Latency, metrics.P50Latency)
	}
}

func TestSLACompliance(t *testing.T) {
	config := &prometheus.Config{
		ServiceName:        "test-sla",
		Namespace:          "test",
		MaxCardinality:     1000,
		CollectionInterval: 10 * time.Second,
		BucketBoundaries:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}

	provider, err := prometheus.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(context.Background())

	tr, err := provider.CreateTracer("sla")
	if err != nil {
		t.Fatalf("Failed to create tracer: %v", err)
	}

	monitor := NewPerformanceMonitor(tr)
	ctx := context.Background()

	// Record data that meets SLA
	for i := 0; i < 100; i++ {
		monitor.RecordRequest(ctx, 100*time.Millisecond, true)
	}

	metrics := monitor.GetCurrentMetrics()
	slaCompliant := monitor.checkSLACompliance(metrics)

	if !slaCompliant {
		t.Errorf("Expected SLA compliance with good metrics, got non-compliant")
	}

	// Record data that violates SLA
	for i := 0; i < 10; i++ {
		monitor.RecordRequest(ctx, 2*time.Second, false) // High latency + errors
	}

	metrics = monitor.GetCurrentMetrics()
	slaCompliant = monitor.checkSLACompliance(metrics)

	if slaCompliant {
		t.Errorf("Expected SLA violation with bad metrics, got compliant")
	}
}
