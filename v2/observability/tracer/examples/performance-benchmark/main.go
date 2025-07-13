package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/prometheus"
)

// PerformanceMonitor tracks performance metrics and SLAs
type PerformanceMonitor struct {
	tracer         tracer.Tracer
	latencies      []time.Duration
	requestCount   int64
	errorCount     int64
	successCount   int64
	startTime      time.Time
	mu             sync.RWMutex
	slaTargets     SLATargets
	currentMetrics Metrics
}

// SLATargets defines performance targets
type SLATargets struct {
	MaxP95Latency    time.Duration
	MaxP99Latency    time.Duration
	MinSuccessRate   float64
	MaxErrorRate     float64
	MinThroughputRPS float64
}

// Metrics holds current performance metrics
type Metrics struct {
	P50Latency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
	AvgLatency     time.Duration
	ThroughputRPS  float64
	ErrorRate      float64
	SuccessRate    float64
	MemoryUsageMB  float64
	GoroutineCount int
	GCCount        uint32
}

// LoadTestScenario represents a load testing scenario
type LoadTestScenario struct {
	Name            string
	ConcurrentUsers int
	Duration        time.Duration
	RampUpTime      time.Duration
	RequestRate     int // requests per second per user
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(tr tracer.Tracer) *PerformanceMonitor {
	return &PerformanceMonitor{
		tracer:    tr,
		latencies: make([]time.Duration, 0, 10000),
		startTime: time.Now(),
		slaTargets: SLATargets{
			MaxP95Latency:    500 * time.Millisecond,
			MaxP99Latency:    1 * time.Second,
			MinSuccessRate:   99.5,
			MaxErrorRate:     0.5,
			MinThroughputRPS: 100,
		},
	}
}

// RecordRequest records metrics for a request
func (pm *PerformanceMonitor) RecordRequest(ctx context.Context, latency time.Duration, success bool) {
	_, span := pm.tracer.StartSpan(ctx, "performance.record_request",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"performance.latency_ms": latency.Milliseconds(),
			"performance.success":    success,
		}),
	)
	defer span.End()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Record latency
	pm.latencies = append(pm.latencies, latency)

	// Keep only last 10000 requests for memory efficiency
	if len(pm.latencies) > 10000 {
		pm.latencies = pm.latencies[1000:]
	}

	// Update counters
	atomic.AddInt64(&pm.requestCount, 1)
	if success {
		atomic.AddInt64(&pm.successCount, 1)
	} else {
		atomic.AddInt64(&pm.errorCount, 1)
	}

	// Update span with current metrics
	span.SetAttribute("performance.total_requests", atomic.LoadInt64(&pm.requestCount))
	span.SetAttribute("performance.success_rate", pm.calculateSuccessRate())
	span.SetAttribute("performance.error_rate", pm.calculateErrorRate())

	span.SetStatus(tracer.StatusCodeOk, "Request metrics recorded")
}

// SimulateBusinessOperation simulates a realistic business operation
func (pm *PerformanceMonitor) SimulateBusinessOperation(ctx context.Context, operationType string) error {
	ctx, span := pm.tracer.StartSpan(ctx, fmt.Sprintf("business.%s", operationType),
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"business.operation": operationType,
			"business.category":  "core",
		}),
	)
	defer span.End()

	startTime := time.Now()

	// Simulate different types of operations with varying complexity
	switch operationType {
	case "quick_lookup":
		if err := pm.simulateQuickLookup(ctx); err != nil {
			pm.RecordRequest(ctx, time.Since(startTime), false)
			return err
		}
	case "standard_processing":
		if err := pm.simulateStandardProcessing(ctx); err != nil {
			pm.RecordRequest(ctx, time.Since(startTime), false)
			return err
		}
	case "complex_analysis":
		if err := pm.simulateComplexAnalysis(ctx); err != nil {
			pm.RecordRequest(ctx, time.Since(startTime), false)
			return err
		}
	case "data_aggregation":
		if err := pm.simulateDataAggregation(ctx); err != nil {
			pm.RecordRequest(ctx, time.Since(startTime), false)
			return err
		}
	}

	pm.RecordRequest(ctx, time.Since(startTime), true)
	span.SetAttribute("business.latency_ms", time.Since(startTime).Milliseconds())
	span.SetStatus(tracer.StatusCodeOk, "Business operation completed")
	return nil
}

func (pm *PerformanceMonitor) simulateQuickLookup(ctx context.Context) error {
	_, span := pm.tracer.StartSpan(ctx, "cache.lookup",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"cache.type": "redis",
			"cache.key":  "user_profile",
		}),
	)
	defer span.End()

	// Simulate cache lookup (fast operation)
	latency := time.Duration(1+rand.Intn(10)) * time.Millisecond
	time.Sleep(latency)

	// 1% chance of cache miss
	if rand.Float32() < 0.01 {
		err := fmt.Errorf("cache miss")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		return err
	}

	span.SetAttribute("cache.hit", true)
	span.SetStatus(tracer.StatusCodeOk, "Cache hit")
	return nil
}

func (pm *PerformanceMonitor) simulateStandardProcessing(ctx context.Context) error {
	_, span := pm.tracer.StartSpan(ctx, "processing.standard",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate multiple processing steps
	steps := []string{"validate", "transform", "enrich", "persist"}

	for i, step := range steps {
		stepCtx, stepSpan := pm.tracer.StartSpan(ctx, fmt.Sprintf("processing.%s", step),
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithSpanAttributes(map[string]interface{}{
				"step.index": i,
				"step.name":  step,
			}),
		)

		// Simulate processing time
		stepLatency := time.Duration(20+rand.Intn(50)) * time.Millisecond
		time.Sleep(stepLatency)

		// 2% chance of step failure
		if rand.Float32() < 0.02 {
			err := fmt.Errorf("processing step %s failed", step)
			stepSpan.SetStatus(tracer.StatusCodeError, err.Error())
			stepSpan.End()
			span.SetStatus(tracer.StatusCodeError, err.Error())
			return err
		}

		stepSpan.SetStatus(tracer.StatusCodeOk, fmt.Sprintf("Step %s completed", step))
		stepSpan.End()
		_ = stepCtx // suppress unused warning
	}

	span.SetStatus(tracer.StatusCodeOk, "Standard processing completed")
	return nil
}

func (pm *PerformanceMonitor) simulateComplexAnalysis(ctx context.Context) error {
	_, span := pm.tracer.StartSpan(ctx, "analysis.complex",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"analysis.type":       "ml_inference",
			"analysis.complexity": "high",
		}),
	)
	defer span.End()

	// Simulate complex computation
	latency := time.Duration(200+rand.Intn(800)) * time.Millisecond
	time.Sleep(latency)

	// 5% chance of analysis failure
	if rand.Float32() < 0.05 {
		err := fmt.Errorf("analysis computation failed")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		return err
	}

	span.SetAttribute("analysis.confidence_score", 0.85+rand.Float64()*0.15)
	span.SetStatus(tracer.StatusCodeOk, "Complex analysis completed")
	return nil
}

func (pm *PerformanceMonitor) simulateDataAggregation(ctx context.Context) error {
	_, span := pm.tracer.StartSpan(ctx, "aggregation.data",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"aggregation.type":   "time_series",
			"aggregation.window": "1h",
		}),
	)
	defer span.End()

	// Simulate database queries
	queries := []string{"user_events", "purchase_history", "engagement_metrics"}

	for _, query := range queries {
		queryCtx, querySpan := pm.tracer.StartSpan(ctx, "db.query",
			tracer.WithSpanKind(tracer.SpanKindClient),
			tracer.WithSpanAttributes(map[string]interface{}{
				"db.system":    "postgresql",
				"db.operation": "SELECT",
				"db.table":     query,
			}),
		)

		// Simulate query execution time
		queryLatency := time.Duration(50+rand.Intn(200)) * time.Millisecond
		time.Sleep(queryLatency)

		querySpan.SetAttribute("db.rows_returned", 1000+rand.Intn(9000))
		querySpan.SetStatus(tracer.StatusCodeOk, "Query executed")
		querySpan.End()
		_ = queryCtx // suppress unused warning
	}

	span.SetAttribute("aggregation.records_processed", 25000+rand.Intn(75000))
	span.SetStatus(tracer.StatusCodeOk, "Data aggregation completed")
	return nil
}

// RunLoadTest executes a load testing scenario
func (pm *PerformanceMonitor) RunLoadTest(ctx context.Context, scenario LoadTestScenario) error {
	ctx, span := pm.tracer.StartSpan(ctx, "loadtest.scenario",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"loadtest.name":             scenario.Name,
			"loadtest.concurrent_users": scenario.ConcurrentUsers,
			"loadtest.duration_seconds": scenario.Duration.Seconds(),
			"loadtest.request_rate":     scenario.RequestRate,
		}),
	)
	defer span.End()

	log.Printf("Starting load test: %s", scenario.Name)
	log.Printf("Concurrent users: %d, Duration: %v, Request rate: %d rps/user",
		scenario.ConcurrentUsers, scenario.Duration, scenario.RequestRate)

	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	// Start users gradually (ramp-up)
	userStartInterval := scenario.RampUpTime / time.Duration(scenario.ConcurrentUsers)

	for i := 0; i < scenario.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			pm.simulateUser(ctx, userID, scenario, stopCh)
		}(i)

		time.Sleep(userStartInterval)
	}

	// Run for specified duration
	time.Sleep(scenario.Duration)
	close(stopCh)

	wg.Wait()

	// Calculate and log final metrics
	metrics := pm.GetCurrentMetrics()
	pm.logScenarioResults(span, scenario, metrics)

	span.SetStatus(tracer.StatusCodeOk, "Load test scenario completed")
	return nil
}

func (pm *PerformanceMonitor) simulateUser(ctx context.Context, userID int, scenario LoadTestScenario, stopCh <-chan struct{}) {
	requestInterval := time.Second / time.Duration(scenario.RequestRate)
	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()

	userCtx, userSpan := pm.tracer.StartSpan(ctx, "loadtest.user",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"user.id":       userID,
			"user.scenario": scenario.Name,
		}),
	)
	defer userSpan.End()

	requestCount := 0
	for {
		select {
		case <-stopCh:
			userSpan.SetAttribute("user.requests_completed", requestCount)
			userSpan.SetStatus(tracer.StatusCodeOk, "User simulation completed")
			return
		case <-ticker.C:
			// Select random operation based on realistic distribution
			var operation string
			switch rand.Intn(100) {
			case 0, 1, 2, 3, 4: // 5% complex analysis
				operation = "complex_analysis"
			case 5, 6, 7, 8, 9, 10, 11, 12, 13, 14: // 10% data aggregation
				operation = "data_aggregation"
			case 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34: // 20% standard processing
				operation = "standard_processing"
			default: // 65% quick lookup
				operation = "quick_lookup"
			}

			if err := pm.SimulateBusinessOperation(userCtx, operation); err != nil {
				// Log error but continue simulation
				log.Printf("User %d operation %s failed: %v", userID, operation, err)
			}
			requestCount++
		}
	}
}

func (pm *PerformanceMonitor) logScenarioResults(span tracer.Span, scenario LoadTestScenario, metrics Metrics) {
	log.Printf("\n=== Load Test Results: %s ===", scenario.Name)
	log.Printf("Duration: %v", scenario.Duration)
	log.Printf("Concurrent Users: %d", scenario.ConcurrentUsers)
	log.Printf("Total Requests: %d", atomic.LoadInt64(&pm.requestCount))
	log.Printf("Throughput: %.2f RPS", metrics.ThroughputRPS)
	log.Printf("Success Rate: %.2f%%", metrics.SuccessRate)
	log.Printf("Error Rate: %.2f%%", metrics.ErrorRate)
	log.Printf("P50 Latency: %v", metrics.P50Latency)
	log.Printf("P95 Latency: %v", metrics.P95Latency)
	log.Printf("P99 Latency: %v", metrics.P99Latency)
	log.Printf("Memory Usage: %.2f MB", metrics.MemoryUsageMB)
	log.Printf("Goroutines: %d", metrics.GoroutineCount)

	// Check SLA compliance
	slaCompliant := pm.checkSLACompliance(metrics)
	log.Printf("SLA Compliant: %t", slaCompliant)

	// Add metrics to span
	span.SetAttribute("loadtest.throughput_rps", metrics.ThroughputRPS)
	span.SetAttribute("loadtest.success_rate", metrics.SuccessRate)
	span.SetAttribute("loadtest.p95_latency_ms", metrics.P95Latency.Milliseconds())
	span.SetAttribute("loadtest.p99_latency_ms", metrics.P99Latency.Milliseconds())
	span.SetAttribute("loadtest.sla_compliant", slaCompliant)
}

// GetCurrentMetrics calculates current performance metrics
func (pm *PerformanceMonitor) GetCurrentMetrics() Metrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.latencies) == 0 {
		return Metrics{}
	}

	// Sort latencies for percentile calculation
	sortedLatencies := make([]time.Duration, len(pm.latencies))
	copy(sortedLatencies, pm.latencies)

	// Simple sort implementation for demo
	for i := 0; i < len(sortedLatencies); i++ {
		for j := i + 1; j < len(sortedLatencies); j++ {
			if sortedLatencies[i] > sortedLatencies[j] {
				sortedLatencies[i], sortedLatencies[j] = sortedLatencies[j], sortedLatencies[i]
			}
		}
	}

	totalRequests := atomic.LoadInt64(&pm.requestCount)
	successCount := atomic.LoadInt64(&pm.successCount)
	errorCount := atomic.LoadInt64(&pm.errorCount)

	duration := time.Since(pm.startTime)
	throughput := float64(totalRequests) / duration.Seconds()

	var successRate, errorRate float64
	if totalRequests > 0 {
		successRate = float64(successCount) / float64(totalRequests) * 100
		errorRate = float64(errorCount) / float64(totalRequests) * 100
	}

	// Calculate percentiles
	p50Index := len(sortedLatencies) * 50 / 100
	p95Index := len(sortedLatencies) * 95 / 100
	p99Index := len(sortedLatencies) * 99 / 100

	if p50Index >= len(sortedLatencies) {
		p50Index = len(sortedLatencies) - 1
	}
	if p95Index >= len(sortedLatencies) {
		p95Index = len(sortedLatencies) - 1
	}
	if p99Index >= len(sortedLatencies) {
		p99Index = len(sortedLatencies) - 1
	}

	// Calculate average latency
	var totalLatency time.Duration
	for _, lat := range sortedLatencies {
		totalLatency += lat
	}
	avgLatency := totalLatency / time.Duration(len(sortedLatencies))

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return Metrics{
		P50Latency:     sortedLatencies[p50Index],
		P95Latency:     sortedLatencies[p95Index],
		P99Latency:     sortedLatencies[p99Index],
		AvgLatency:     avgLatency,
		ThroughputRPS:  throughput,
		ErrorRate:      errorRate,
		SuccessRate:    successRate,
		MemoryUsageMB:  float64(memStats.Alloc) / 1024 / 1024,
		GoroutineCount: runtime.NumGoroutine(),
		GCCount:        memStats.NumGC,
	}
}

func (pm *PerformanceMonitor) calculateSuccessRate() float64 {
	total := atomic.LoadInt64(&pm.requestCount)
	success := atomic.LoadInt64(&pm.successCount)
	if total == 0 {
		return 0
	}
	return float64(success) / float64(total) * 100
}

func (pm *PerformanceMonitor) calculateErrorRate() float64 {
	total := atomic.LoadInt64(&pm.requestCount)
	errors := atomic.LoadInt64(&pm.errorCount)
	if total == 0 {
		return 0
	}
	return float64(errors) / float64(total) * 100
}

func (pm *PerformanceMonitor) checkSLACompliance(metrics Metrics) bool {
	return metrics.P95Latency <= pm.slaTargets.MaxP95Latency &&
		metrics.P99Latency <= pm.slaTargets.MaxP99Latency &&
		metrics.SuccessRate >= pm.slaTargets.MinSuccessRate &&
		metrics.ErrorRate <= pm.slaTargets.MaxErrorRate &&
		metrics.ThroughputRPS >= pm.slaTargets.MinThroughputRPS
}

func main() {
	// Configure Prometheus provider for metrics collection
	config := &prometheus.Config{
		ServiceName:        "performance-benchmark",
		Namespace:          "perftest",
		MaxCardinality:     10000,
		CollectionInterval: 5 * time.Second,
		BucketBoundaries:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}

	// Create provider
	provider, err := prometheus.NewProvider(config)
	if err != nil {
		log.Fatalf("Failed to create Prometheus provider: %v", err)
	}

	// Create tracer
	tr, err := provider.CreateTracer("performance",
		tracer.WithServiceName("performance-benchmark"),
		tracer.WithEnvironment("testing"),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// Create performance monitor
	monitor := NewPerformanceMonitor(tr)

	// Define load test scenarios
	scenarios := []LoadTestScenario{
		{
			Name:            "baseline_load",
			ConcurrentUsers: 10,
			Duration:        30 * time.Second,
			RampUpTime:      5 * time.Second,
			RequestRate:     2, // 2 requests per second per user
		},
		{
			Name:            "moderate_load",
			ConcurrentUsers: 50,
			Duration:        60 * time.Second,
			RampUpTime:      10 * time.Second,
			RequestRate:     3,
		},
		{
			Name:            "stress_test",
			ConcurrentUsers: 100,
			Duration:        90 * time.Second,
			RampUpTime:      15 * time.Second,
			RequestRate:     5,
		},
	}

	fmt.Println("Starting Performance Benchmark Example")
	fmt.Println("======================================")

	ctx := context.Background()

	// Run all scenarios
	for i, scenario := range scenarios {
		if i > 0 {
			// Cool down period between scenarios
			fmt.Printf("\nCooldown period (30 seconds)...\n")
			time.Sleep(30 * time.Second)

			// Reset counters for next scenario
			atomic.StoreInt64(&monitor.requestCount, 0)
			atomic.StoreInt64(&monitor.successCount, 0)
			atomic.StoreInt64(&monitor.errorCount, 0)
			monitor.mu.Lock()
			monitor.latencies = monitor.latencies[:0]
			monitor.startTime = time.Now()
			monitor.mu.Unlock()
		}

		if err := monitor.RunLoadTest(ctx, scenario); err != nil {
			log.Printf("Load test scenario %s failed: %v", scenario.Name, err)
		}
	}

	fmt.Println("\n======================================")
	fmt.Println("Performance Benchmark Completed!")
	fmt.Println("Check your Prometheus metrics at http://localhost:9090/metrics")

	// Keep the program running to expose metrics
	fmt.Println("Press Ctrl+C to stop...")
	select {} // Block forever
}
