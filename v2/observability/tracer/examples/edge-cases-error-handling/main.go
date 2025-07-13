package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/datadog"
)

// EdgeCaseSimulator simulates various edge cases and error scenarios
type EdgeCaseSimulator struct {
	tracer             tracer.Tracer
	errorRate          float64
	networkFailures    int64
	timeouts           int64
	circuitBreakerOpen bool
	mu                 sync.RWMutex
	activeConnections  int64
	maxConnections     int64
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failures        int64
	lastFailureTime time.Time
	state           CircuitBreakerState
	mu              sync.RWMutex
}

type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// RetryConfig defines retry behavior with exponential backoff
type RetryConfig struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	Jitter        bool
}

// ResourceManager manages system resources and limits
type ResourceManager struct {
	maxMemoryMB       int64
	maxGoroutines     int
	maxOpenFiles      int
	currentMemoryMB   int64
	currentGoroutines int64
	currentOpenFiles  int64
	mu                sync.RWMutex
}

// NewEdgeCaseSimulator creates a new edge case simulator
func NewEdgeCaseSimulator(tr tracer.Tracer) *EdgeCaseSimulator {
	return &EdgeCaseSimulator{
		tracer:         tr,
		errorRate:      0.05, // 5% error rate
		maxConnections: 100,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitBreakerClosed,
	}
}

// Execute wraps function execution with circuit breaker logic
func (cb *CircuitBreaker) Execute(ctx context.Context, tr tracer.Tracer, operation func() error) error {
	ctx, span := tr.StartSpan(ctx, "circuit_breaker.execute",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"circuit_breaker.state":    cb.getStateString(),
			"circuit_breaker.failures": atomic.LoadInt64(&cb.failures),
		}),
	)
	defer span.End()

	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()

	switch state {
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.mu.Lock()
			cb.state = CircuitBreakerHalfOpen
			cb.mu.Unlock()
			span.SetAttribute("circuit_breaker.transition", "open_to_half_open")
		} else {
			err := errors.New("circuit breaker is open")
			span.SetStatus(tracer.StatusCodeError, err.Error())
			span.RecordError(err, map[string]interface{}{
				"error.type": "circuit_breaker_open",
			})
			return err
		}
	}

	// Execute the operation
	err := operation()
	if err != nil {
		cb.recordFailure()
		span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Operation failed: %v", err))
		span.RecordError(err, map[string]interface{}{
			"error.type": "operation_failure",
		})
		return err
	}

	cb.recordSuccess()
	span.SetStatus(tracer.StatusCodeOk, "Operation successful")
	return nil
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.AddInt64(&cb.failures, 1)
	cb.lastFailureTime = time.Now()

	if int(atomic.LoadInt64(&cb.failures)) >= cb.maxFailures {
		cb.state = CircuitBreakerOpen
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	atomic.StoreInt64(&cb.failures, 0)
	cb.state = CircuitBreakerClosed
}

func (cb *CircuitBreaker) getStateString() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return "closed"
	case CircuitBreakerOpen:
		return "open"
	case CircuitBreakerHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// SimulateNetworkFailures simulates various network failure scenarios
func (ecs *EdgeCaseSimulator) SimulateNetworkFailures(ctx context.Context) error {
	ctx, span := ecs.tracer.StartSpan(ctx, "edge_case.network_failures",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"simulation.type": "network_failures",
		}),
	)
	defer span.End()

	scenarios := []struct {
		name        string
		probability float64
		simulate    func(context.Context) error
	}{
		{"connection_timeout", 0.3, ecs.simulateConnectionTimeout},
		{"dns_failure", 0.2, ecs.simulateDNSFailure},
		{"connection_refused", 0.25, ecs.simulateConnectionRefused},
		{"intermittent_failure", 0.15, ecs.simulateIntermittentFailure},
		{"ssl_handshake_failure", 0.1, ecs.simulateSSLHandshakeFailure},
	}

	for _, scenario := range scenarios {
		if rand.Float64() < scenario.probability {
			scenarioCtx, scenarioSpan := ecs.tracer.StartSpan(ctx, fmt.Sprintf("network.%s", scenario.name),
				tracer.WithSpanKind(tracer.SpanKindInternal),
				tracer.WithSpanAttributes(map[string]interface{}{
					"scenario.name":        scenario.name,
					"scenario.probability": scenario.probability,
				}),
			)

			if err := scenario.simulate(scenarioCtx); err != nil {
				atomic.AddInt64(&ecs.networkFailures, 1)
				scenarioSpan.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Network failure: %v", err))
				scenarioSpan.RecordError(err, map[string]interface{}{
					"error.type": "network_failure",
					"scenario":   scenario.name,
				})
				scenarioSpan.End()

				span.SetAttribute("failed_scenario", scenario.name)
				span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Network simulation failed: %v", err))
				return err
			}

			scenarioSpan.SetStatus(tracer.StatusCodeOk, "Network scenario completed")
			scenarioSpan.End()
		}
	}

	span.SetAttribute("network.failures_count", atomic.LoadInt64(&ecs.networkFailures))
	span.SetStatus(tracer.StatusCodeOk, "Network failure simulation completed")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateConnectionTimeout(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "network.connection_timeout",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	// Simulate long connection attempt
	select {
	case <-time.After(5 * time.Second):
		err := &net.OpError{
			Op:   "dial",
			Net:  "tcp",
			Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.100"), Port: 8080},
			Err:  errors.New("i/o timeout"),
		}
		span.SetStatus(tracer.StatusCodeError, err.Error())
		return err
	case <-ctx.Done():
		span.SetStatus(tracer.StatusCodeError, "Context cancelled")
		return ctx.Err()
	case <-time.After(100 * time.Millisecond): // Fast simulation
		span.SetStatus(tracer.StatusCodeOk, "Connection timeout simulated")
		return nil
	}
}

func (ecs *EdgeCaseSimulator) simulateDNSFailure(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "network.dns_failure",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	// Simulate DNS resolution failure
	time.Sleep(50 * time.Millisecond)

	err := &net.DNSError{
		Err:         "no such host",
		Name:        "nonexistent.example.com",
		Server:      "8.8.8.8",
		IsTimeout:   false,
		IsTemporary: false,
	}

	span.SetStatus(tracer.StatusCodeError, err.Error())
	return err
}

func (ecs *EdgeCaseSimulator) simulateConnectionRefused(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "network.connection_refused",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(25 * time.Millisecond)

	err := &net.OpError{
		Op:   "dial",
		Net:  "tcp",
		Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9999},
		Err:  errors.New("connection refused"),
	}

	span.SetStatus(tracer.StatusCodeError, err.Error())
	return err
}

func (ecs *EdgeCaseSimulator) simulateIntermittentFailure(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "network.intermittent_failure",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	// 50% chance of failure
	if rand.Float64() < 0.5 {
		time.Sleep(30 * time.Millisecond)
		err := errors.New("temporary network glitch")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		return err
	}

	time.Sleep(10 * time.Millisecond)
	span.SetStatus(tracer.StatusCodeOk, "Intermittent failure passed")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateSSLHandshakeFailure(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "network.ssl_handshake_failure",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(75 * time.Millisecond)

	err := errors.New("tls: handshake failure")
	span.SetStatus(tracer.StatusCodeError, err.Error())
	return err
}

// SimulateResourceExhaustion simulates resource exhaustion scenarios
func (ecs *EdgeCaseSimulator) SimulateResourceExhaustion(ctx context.Context) error {
	ctx, span := ecs.tracer.StartSpan(ctx, "edge_case.resource_exhaustion",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	rm := &ResourceManager{
		maxMemoryMB:   100,  // 100MB limit
		maxGoroutines: 1000, // 1000 goroutines limit
		maxOpenFiles:  500,  // 500 files limit
	}

	scenarios := []struct {
		name     string
		simulate func(context.Context, *ResourceManager) error
	}{
		{"memory_exhaustion", ecs.simulateMemoryExhaustion},
		{"goroutine_leak", ecs.simulateGoroutineLeak},
		{"file_descriptor_exhaustion", ecs.simulateFileDescriptorExhaustion},
		{"connection_pool_exhaustion", ecs.simulateConnectionPoolExhaustion},
	}

	for _, scenario := range scenarios {
		scenarioCtx, scenarioSpan := ecs.tracer.StartSpan(ctx, fmt.Sprintf("resource.%s", scenario.name),
			tracer.WithSpanKind(tracer.SpanKindInternal),
		)

		err := scenario.simulate(scenarioCtx, rm)
		if err != nil {
			scenarioSpan.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Resource exhaustion: %v", err))
			scenarioSpan.RecordError(err, map[string]interface{}{
				"error.type": "resource_exhaustion",
				"scenario":   scenario.name,
			})
		} else {
			scenarioSpan.SetStatus(tracer.StatusCodeOk, "Resource exhaustion simulated")
		}
		scenarioSpan.End()
	}

	span.SetAttribute("memory.current_mb", rm.currentMemoryMB)
	span.SetAttribute("goroutines.current", rm.currentGoroutines)
	span.SetAttribute("files.current", rm.currentOpenFiles)
	span.SetStatus(tracer.StatusCodeOk, "Resource exhaustion simulation completed")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateMemoryExhaustion(ctx context.Context, rm *ResourceManager) error {
	_, span := ecs.tracer.StartSpan(ctx, "resource.memory_exhaustion",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate gradual memory increase
	for i := 0; i < 10; i++ {
		rm.mu.Lock()
		rm.currentMemoryMB += 15 // Add 15MB per iteration
		currentMem := rm.currentMemoryMB
		rm.mu.Unlock()

		span.SetAttribute(fmt.Sprintf("memory.iteration_%d", i), currentMem)

		if currentMem > rm.maxMemoryMB {
			err := fmt.Errorf("memory exhausted: %dMB > %dMB limit", currentMem, rm.maxMemoryMB)
			span.SetStatus(tracer.StatusCodeError, err.Error())
			return err
		}

		time.Sleep(10 * time.Millisecond)
	}

	span.SetStatus(tracer.StatusCodeOk, "Memory exhaustion simulated without breach")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateGoroutineLeak(ctx context.Context, rm *ResourceManager) error {
	_, span := ecs.tracer.StartSpan(ctx, "resource.goroutine_leak",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Start goroutines that simulate leak
	leakedGoroutines := make(chan struct{}, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			defer func() { leakedGoroutines <- struct{}{} }()

			// Simulate work that might cause leak
			select {
			case <-time.After(1 * time.Second): // Simulate hanging operation
				// Goroutine would leak here in real scenario
			case <-time.After(50 * time.Millisecond): // Fast simulation
				// Clean exit for simulation
			}
		}(i)

		atomic.AddInt64(&rm.currentGoroutines, 1)
	}

	// Wait for goroutines to complete (or timeout)
	completed := 0
	timeout := time.After(100 * time.Millisecond)

	for completed < 100 {
		select {
		case <-leakedGoroutines:
			completed++
			atomic.AddInt64(&rm.currentGoroutines, -1)
		case <-timeout:
			leaked := 100 - completed
			span.SetAttribute("goroutines.leaked", leaked)
			if leaked > 10 { // More than 10 leaked is considered failure
				err := fmt.Errorf("goroutine leak detected: %d goroutines leaked", leaked)
				span.SetStatus(tracer.StatusCodeError, err.Error())
				return err
			}
			span.SetStatus(tracer.StatusCodeOk, "Minor goroutine leak simulated")
			return nil
		}
	}

	span.SetAttribute("goroutines.leaked", 0)
	span.SetStatus(tracer.StatusCodeOk, "No goroutine leak detected")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateFileDescriptorExhaustion(ctx context.Context, rm *ResourceManager) error {
	_, span := ecs.tracer.StartSpan(ctx, "resource.file_descriptor_exhaustion",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate opening many files
	for i := 0; i < 600; i++ {
		rm.mu.Lock()
		rm.currentOpenFiles++
		currentFiles := rm.currentOpenFiles
		rm.mu.Unlock()

		if currentFiles > int64(rm.maxOpenFiles) {
			err := fmt.Errorf("file descriptor limit exceeded: %d > %d", currentFiles, rm.maxOpenFiles)
			span.SetStatus(tracer.StatusCodeError, err.Error())
			return err
		}

		// Simulate some files being closed
		if i%10 == 0 && rm.currentOpenFiles > 0 {
			rm.mu.Lock()
			rm.currentOpenFiles -= 5
			rm.mu.Unlock()
		}

		time.Sleep(1 * time.Millisecond)
	}

	span.SetAttribute("files.final_count", rm.currentOpenFiles)
	span.SetStatus(tracer.StatusCodeOk, "File descriptor exhaustion simulated")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateConnectionPoolExhaustion(ctx context.Context, rm *ResourceManager) error {
	_, span := ecs.tracer.StartSpan(ctx, "resource.connection_pool_exhaustion",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate rapid connection requests
	for i := 0; i < 150; i++ {
		current := atomic.AddInt64(&ecs.activeConnections, 1)

		if current > ecs.maxConnections {
			atomic.AddInt64(&ecs.activeConnections, -1)
			err := fmt.Errorf("connection pool exhausted: %d > %d", current, ecs.maxConnections)
			span.SetStatus(tracer.StatusCodeError, err.Error())
			return err
		}

		// Simulate connection being released after some time
		go func() {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			atomic.AddInt64(&ecs.activeConnections, -1)
		}()

		time.Sleep(5 * time.Millisecond)
	}

	span.SetAttribute("connections.active", atomic.LoadInt64(&ecs.activeConnections))
	span.SetStatus(tracer.StatusCodeOk, "Connection pool exhaustion simulated")
	return nil
}

// RetryWithBackoff implements retry logic with exponential backoff and jitter
func RetryWithBackoff(ctx context.Context, tr tracer.Tracer, config RetryConfig, operation func() error) error {
	ctx, span := tr.StartSpan(ctx, "retry.with_backoff",
		tracer.WithSpanKind(tracer.SpanKindInternal),
		tracer.WithSpanAttributes(map[string]interface{}{
			"retry.max_attempts":   config.MaxAttempts,
			"retry.base_delay_ms":  config.BaseDelay.Milliseconds(),
			"retry.backoff_factor": config.BackoffFactor,
			"retry.jitter_enabled": config.Jitter,
		}),
	)
	defer span.End()

	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		_, attemptSpan := tr.StartSpan(ctx, fmt.Sprintf("retry.attempt_%d", attempt),
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithSpanAttributes(map[string]interface{}{
				"retry.attempt": attempt,
			}),
		)

		err := operation()
		if err == nil {
			attemptSpan.SetStatus(tracer.StatusCodeOk, "Operation successful")
			attemptSpan.End()
			span.SetAttribute("retry.successful_attempt", attempt)
			span.SetStatus(tracer.StatusCodeOk, fmt.Sprintf("Operation succeeded on attempt %d", attempt))
			return nil
		}

		lastErr = err
		attemptSpan.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Attempt failed: %v", err))
		attemptSpan.RecordError(err, map[string]interface{}{
			"retry.attempt": attempt,
		})
		attemptSpan.End()

		if attempt < config.MaxAttempts {
			delay := calculateBackoffDelay(config, attempt)
			span.AddEvent("retry.backoff", map[string]interface{}{
				"attempt":  attempt,
				"delay_ms": delay.Milliseconds(),
			})

			select {
			case <-time.After(delay):
				// Continue to next attempt
			case <-ctx.Done():
				span.SetStatus(tracer.StatusCodeError, "Retry cancelled by context")
				return ctx.Err()
			}
		}
	}

	span.SetAttribute("retry.final_attempt", config.MaxAttempts)
	span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("All retry attempts failed: %v", lastErr))
	span.RecordError(lastErr, map[string]interface{}{
		"retry.exhausted": true,
	})

	return fmt.Errorf("retry exhausted after %d attempts: %w", config.MaxAttempts, lastErr)
}

func calculateBackoffDelay(config RetryConfig, attempt int) time.Duration {
	// Calculate exponential backoff
	delay := time.Duration(float64(config.BaseDelay) *
		pow(config.BackoffFactor, float64(attempt-1)))

	// Cap at max delay
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	// Add jitter if enabled
	if config.Jitter {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1) // 10% jitter
		delay += jitter
	}

	return delay
}

// Simple power function for integer exponents
func pow(base float64, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// SimulateDataCorruption simulates data corruption scenarios
func (ecs *EdgeCaseSimulator) SimulateDataCorruption(ctx context.Context) error {
	ctx, span := ecs.tracer.StartSpan(ctx, "edge_case.data_corruption",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	corruptionScenarios := []struct {
		name        string
		description string
		simulate    func(context.Context) error
	}{
		{"invalid_trace_context", "Malformed trace context headers", ecs.simulateInvalidTraceContext},
		{"corrupted_span_data", "Corrupted span attribute data", ecs.simulateCorruptedSpanData},
		{"encoding_errors", "Character encoding issues", ecs.simulateEncodingErrors},
		{"oversized_payloads", "Payloads exceeding size limits", ecs.simulateOversizedPayloads},
	}

	for _, scenario := range corruptionScenarios {
		scenarioCtx, scenarioSpan := ecs.tracer.StartSpan(ctx, fmt.Sprintf("corruption.%s", scenario.name),
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithSpanAttributes(map[string]interface{}{
				"corruption.type":        scenario.name,
				"corruption.description": scenario.description,
			}),
		)

		err := scenario.simulate(scenarioCtx)
		if err != nil {
			scenarioSpan.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Data corruption: %v", err))
			scenarioSpan.RecordError(err, map[string]interface{}{
				"error.type": "data_corruption",
			})
		} else {
			scenarioSpan.SetStatus(tracer.StatusCodeOk, "Data corruption handled gracefully")
		}
		scenarioSpan.End()
	}

	span.SetStatus(tracer.StatusCodeOk, "Data corruption simulation completed")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateInvalidTraceContext(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "corruption.invalid_trace_context",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate various invalid trace context scenarios
	invalidContexts := []string{
		"",               // Empty context
		"invalid-format", // Wrong format
		"00-invalid-trace-id-00000000000000000000000000000000-0000000000000000-01", // Invalid trace ID
		"00-4bf92f3577b34da6a3ce929d0e0e4736-invalid-span-id-01",                   // Invalid span ID
		"99-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",                  // Invalid version
	}

	for i, invalidContext := range invalidContexts {
		span.AddEvent("invalid_context_processed", map[string]interface{}{
			"context_index":  i,
			"context_value":  invalidContext,
			"context_length": len(invalidContext),
		})

		// In real implementation, this would attempt to parse the context
		// Here we just simulate the validation
		if len(invalidContext) == 0 {
			span.SetAttribute("validation.empty_context", true)
		} else if len(invalidContext) < 55 { // W3C trace context minimum length
			span.SetAttribute("validation.context_too_short", true)
		}
	}

	span.SetStatus(tracer.StatusCodeOk, "Invalid trace contexts processed")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateCorruptedSpanData(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "corruption.corrupted_span_data",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate corrupted attribute data
	corruptedAttributes := map[string]interface{}{
		"null_value":         nil,
		"circular_reference": "self-referential data structure",
		"invalid_utf8":       "\xff\xfe\xfd",
		"extremely_long_key": string(make([]byte, 10000)),
		"binary_data":        []byte{0x00, 0x01, 0x02, 0xFF},
		"nested_complexity": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "deep nesting",
				},
			},
		},
	}

	for key, value := range corruptedAttributes {
		// In real implementation, this would attempt to serialize the attribute
		// Here we simulate the validation and handling
		span.AddEvent("corrupted_attribute_processed", map[string]interface{}{
			"attribute_key":  key,
			"attribute_type": fmt.Sprintf("%T", value),
		})

		// Simulate attribute validation
		if value == nil {
			span.SetAttribute(fmt.Sprintf("validation.%s.is_null", key), true)
		}
		if len(key) > 1000 {
			span.SetAttribute(fmt.Sprintf("validation.%s.key_too_long", key), true)
		}
	}

	span.SetStatus(tracer.StatusCodeOk, "Corrupted span data processed")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateEncodingErrors(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "corruption.encoding_errors",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate various encoding issues
	encodingIssues := []struct {
		name string
		data string
	}{
		{"invalid_utf8", "\xff\xfe\xfd"},
		{"mixed_encoding", "Hello \xc4\x85\xc5\x82\xc4\x99 World"},
		{"null_bytes", "String with \x00 null bytes"},
		{"control_chars", "String with \x01\x02\x03 control characters"},
		{"emoji_complex", "🏴󠁧󠁢󠁳󠁣󠁴󠁿 Complex emoji"},
	}

	for _, issue := range encodingIssues {
		span.AddEvent("encoding_issue_processed", map[string]interface{}{
			"issue_type":  issue.name,
			"data_length": len(issue.data),
			"has_non_utf8": func() bool {
				for _, b := range []byte(issue.data) {
					if b > 127 {
						return true
					}
				}
				return false
			}(),
		})
	}

	span.SetStatus(tracer.StatusCodeOk, "Encoding errors processed")
	return nil
}

func (ecs *EdgeCaseSimulator) simulateOversizedPayloads(ctx context.Context) error {
	_, span := ecs.tracer.StartSpan(ctx, "corruption.oversized_payloads",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate payloads of various sizes
	payloadSizes := []int{
		1024,     // 1KB - normal
		65536,    // 64KB - large
		1048576,  // 1MB - very large
		10485760, // 10MB - excessive
	}

	for _, size := range payloadSizes {
		payload := make([]byte, size)
		for i := range payload {
			payload[i] = byte(i % 256)
		}

		span.AddEvent("oversized_payload_processed", map[string]interface{}{
			"payload_size_bytes": size,
			"payload_size_mb":    float64(size) / 1024 / 1024,
		})

		// Simulate size validation
		if size > 1048576 { // 1MB limit
			span.SetAttribute("validation.payload_too_large", true)
			span.AddEvent("payload_rejected", map[string]interface{}{
				"reason":      "size_limit_exceeded",
				"actual_size": size,
				"max_size":    1048576,
			})
		}
	}

	span.SetStatus(tracer.StatusCodeOk, "Oversized payloads processed")
	return nil
}

// SimulateConcurrencyIssues simulates race conditions and concurrency problems
func (ecs *EdgeCaseSimulator) SimulateConcurrencyIssues(ctx context.Context) error {
	ctx, span := ecs.tracer.StartSpan(ctx, "edge_case.concurrency_issues",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Test concurrent span creation
	var wg sync.WaitGroup
	numGoroutines := 100
	sharedCounter := int64(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			goroutineCtx, goroutineSpan := ecs.tracer.StartSpan(ctx, "concurrent.worker",
				tracer.WithSpanKind(tracer.SpanKindInternal),
				tracer.WithSpanAttributes(map[string]interface{}{
					"goroutine.id": goroutineID,
				}),
			)
			defer goroutineSpan.End()

			// Simulate concurrent operations that might cause race conditions
			for j := 0; j < 10; j++ {
				_, childSpan := ecs.tracer.StartSpan(goroutineCtx, "concurrent.operation",
					tracer.WithSpanKind(tracer.SpanKindInternal),
					tracer.WithSpanAttributes(map[string]interface{}{
						"operation.id": j,
						"goroutine.id": goroutineID,
					}),
				)

				// Simulate work
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

				// Safe concurrent access
				currentCount := atomic.AddInt64(&sharedCounter, 1)
				childSpan.SetAttribute("shared_counter", currentCount)

				childSpan.SetStatus(tracer.StatusCodeOk, "Concurrent operation completed")
				childSpan.End()
			}

			goroutineSpan.SetAttribute("operations_completed", 10)
			goroutineSpan.SetStatus(tracer.StatusCodeOk, "Goroutine completed")
		}(i)
	}

	wg.Wait()

	span.SetAttribute("goroutines_executed", numGoroutines)
	span.SetAttribute("total_operations", atomic.LoadInt64(&sharedCounter))
	span.SetAttribute("expected_operations", numGoroutines*10)

	// Verify no race conditions occurred
	expectedTotal := int64(numGoroutines * 10)
	actualTotal := atomic.LoadInt64(&sharedCounter)

	if actualTotal != expectedTotal {
		err := fmt.Errorf("race condition detected: expected %d operations, got %d", expectedTotal, actualTotal)
		span.SetStatus(tracer.StatusCodeError, err.Error())
		return err
	}

	span.SetStatus(tracer.StatusCodeOk, "Concurrency test completed without race conditions")
	return nil
}

func main() {
	// Configure Datadog provider for edge case testing
	config := &datadog.Config{
		ServiceName:        "edge-cases-example",
		ServiceVersion:     "1.0.0",
		Environment:        "testing",
		AgentHost:          "localhost",
		AgentPort:          8126,
		SampleRate:         1.0,   // 100% sampling for testing
		EnableProfiling:    false, // Disable for edge case testing
		RuntimeMetrics:     false,
		Debug:              true,
		MaxTracesPerSecond: 10000,
		Tags: map[string]string{
			"example":   "edge-cases",
			"test_type": "resilience",
		},
	}

	// Create provider
	provider, err := datadog.NewProvider(config)
	if err != nil {
		log.Fatalf("Failed to create Datadog provider: %v", err)
	}

	// Create tracer
	tr, err := provider.CreateTracer("edge-cases",
		tracer.WithServiceName("edge-cases-example"),
		tracer.WithEnvironment("testing"),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// Create edge case simulator
	simulator := NewEdgeCaseSimulator(tr)

	// Create circuit breaker
	circuitBreaker := NewCircuitBreaker(5, 30*time.Second)

	fmt.Println("Starting Edge Cases & Error Handling Example")
	fmt.Println("============================================")

	ctx := context.Background()

	// Test scenarios
	scenarios := []struct {
		name        string
		description string
		execute     func(context.Context) error
	}{
		{
			"network_failures",
			"Testing various network failure scenarios",
			simulator.SimulateNetworkFailures,
		},
		{
			"resource_exhaustion",
			"Testing resource exhaustion scenarios",
			simulator.SimulateResourceExhaustion,
		},
		{
			"data_corruption",
			"Testing data corruption handling",
			simulator.SimulateDataCorruption,
		},
		{
			"concurrency_issues",
			"Testing concurrent access and race conditions",
			simulator.SimulateConcurrencyIssues,
		},
	}

	// Execute each scenario
	for i, scenario := range scenarios {
		fmt.Printf("\n%d. %s\n", i+1, scenario.description)
		fmt.Printf("   Running scenario: %s\n", scenario.name)

		// Test with circuit breaker
		err := circuitBreaker.Execute(ctx, tr, func() error {
			return scenario.execute(ctx)
		})

		if err != nil {
			fmt.Printf("   ❌ Scenario failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Scenario completed successfully\n")
		}

		// Test retry mechanism
		fmt.Printf("   Testing retry mechanism...\n")
		retryConfig := RetryConfig{
			MaxAttempts:   3,
			BaseDelay:     100 * time.Millisecond,
			MaxDelay:      2 * time.Second,
			BackoffFactor: 2.0,
			Jitter:        true,
		}

		retryErr := RetryWithBackoff(ctx, tr, retryConfig, func() error {
			// 30% chance of failure for retry testing
			if rand.Float64() < 0.3 {
				return fmt.Errorf("simulated transient failure")
			}
			return nil
		})

		if retryErr != nil {
			fmt.Printf("   ⚠️  Retry mechanism exhausted: %v\n", retryErr)
		} else {
			fmt.Printf("   ✅ Retry mechanism succeeded\n")
		}

		// Short delay between scenarios
		time.Sleep(500 * time.Millisecond)
	}

	// Final statistics
	fmt.Printf("\n============================================\n")
	fmt.Printf("Edge Cases Testing Summary:\n")
	fmt.Printf("Network failures encountered: %d\n", atomic.LoadInt64(&simulator.networkFailures))
	fmt.Printf("Circuit breaker state: %s\n", circuitBreaker.getStateString())
	fmt.Printf("Active connections: %d\n", atomic.LoadInt64(&simulator.activeConnections))
	fmt.Printf("Current goroutines: %d\n", runtime.NumGoroutine())

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("Memory usage: %.2f MB\n", float64(memStats.Alloc)/1024/1024)
	fmt.Printf("GC cycles: %d\n", memStats.NumGC)

	fmt.Println("\nEdge Cases & Error Handling Example completed!")
	fmt.Println("Check your tracing backend for detailed traces and metrics.")

	// Cleanup
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := provider.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down provider: %v", err)
	}
}
