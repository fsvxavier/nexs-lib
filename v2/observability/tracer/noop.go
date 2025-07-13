package tracer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// NoopTracer implements Tracer interface with no-operation behavior
type NoopTracer struct {
	metrics     TracerMetrics
	spansCount  int64 // atomic counter for spans created
	finishCount int64 // atomic counter for spans finished
	mu          sync.RWMutex
}

// NewNoopTracer creates a new no-operation tracer
func NewNoopTracer() *NoopTracer {
	return &NoopTracer{
		metrics: TracerMetrics{
			LastActivity: time.Now(),
		},
		spansCount:  0,
		finishCount: 0,
	}
}

// StartSpan creates a no-operation span
func (n *NoopTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	atomic.AddInt64(&n.spansCount, 1)
	n.updateLastActivity()
	return ctx, &NoopSpan{
		name:      name,
		startTime: time.Now(),
		tracer:    n,
		recording: true,
	}
}

// SpanFromContext returns a no-operation span
func (n *NoopTracer) SpanFromContext(ctx context.Context) Span {
	return &NoopSpan{
		tracer:    n,
		recording: false,
	}
}

// ContextWithSpan returns the context unchanged
func (n *NoopTracer) ContextWithSpan(ctx context.Context, span Span) context.Context {
	return ctx
}

// Close performs no operation
func (n *NoopTracer) Close() error {
	return nil
}

// GetMetrics returns tracer metrics
func (n *NoopTracer) GetMetrics() TracerMetrics {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Safely copy metrics with atomic counters
	metrics := n.metrics
	metrics.SpansCreated = atomic.LoadInt64(&n.spansCount)
	metrics.SpansFinished = atomic.LoadInt64(&n.finishCount)
	return metrics
}

func (n *NoopTracer) updateLastActivity() {
	n.mu.Lock()
	n.metrics.LastActivity = time.Now()
	n.mu.Unlock()
}

// NoopSpan implements Span interface with no-operation behavior
type NoopSpan struct {
	name      string
	startTime time.Time
	endTime   time.Time
	tracer    *NoopTracer
	recording bool
}

// Context returns an empty span context
func (n *NoopSpan) Context() SpanContext {
	return SpanContext{
		TraceID: "00000000000000000000000000000000",
		SpanID:  "0000000000000000",
	}
}

// SetName sets the span name (no-op)
func (n *NoopSpan) SetName(name string) {
	n.name = name
}

// SetAttributes sets attributes (no-op)
func (n *NoopSpan) SetAttributes(attributes map[string]interface{}) {
	// No-op
}

// SetAttribute sets a single attribute (no-op)
func (n *NoopSpan) SetAttribute(key string, value interface{}) {
	// No-op
}

// AddEvent adds an event (no-op)
func (n *NoopSpan) AddEvent(name string, attributes map[string]interface{}) {
	// No-op
}

// SetStatus sets the span status (no-op)
func (n *NoopSpan) SetStatus(code StatusCode, message string) {
	// No-op
}

// RecordError records an error (no-op)
func (n *NoopSpan) RecordError(err error, attributes map[string]interface{}) {
	// No-op
}

// End finishes the span
func (n *NoopSpan) End() {
	n.endTime = time.Now()
	n.recording = false
	if n.tracer != nil {
		atomic.AddInt64(&n.tracer.finishCount, 1)
		n.tracer.updateLastActivity()
	}
}

// IsRecording returns recording status
func (n *NoopSpan) IsRecording() bool {
	return n.recording && n.endTime.IsZero()
}

// GetDuration returns the span duration
func (n *NoopSpan) GetDuration() time.Duration {
	if n.endTime.IsZero() {
		return time.Since(n.startTime)
	}
	return n.endTime.Sub(n.startTime)
}

// NoopProvider implements Provider interface with no-operation behavior
type NoopProvider struct {
	tracers map[string]*NoopTracer
	mu      sync.RWMutex
	metrics ProviderMetrics
}

// NewNoopProvider creates a new no-operation provider
func NewNoopProvider() *NoopProvider {
	return &NoopProvider{
		tracers: make(map[string]*NoopTracer),
		metrics: ProviderMetrics{
			ConnectionState: "connected",
			LastFlush:       time.Now(),
		},
	}
}

// Name returns the provider name
func (n *NoopProvider) Name() string {
	return "noop"
}

// CreateTracer creates a new no-operation tracer
func (n *NoopProvider) CreateTracer(name string, options ...TracerOption) (Tracer, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if tracer, exists := n.tracers[name]; exists {
		return tracer, nil
	}

	tracer := NewNoopTracer()
	n.tracers[name] = tracer
	n.metrics.TracersActive = len(n.tracers)

	return tracer, nil
}

// Shutdown performs graceful shutdown (no-op)
func (n *NoopProvider) Shutdown(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.tracers = make(map[string]*NoopTracer)
	n.metrics.TracersActive = 0
	return nil
}

// HealthCheck performs health check
func (n *NoopProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// GetProviderMetrics returns provider metrics
func (n *NoopProvider) GetProviderMetrics() ProviderMetrics {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.metrics
}

// ErrorTracer implements Tracer interface but always returns errors
type ErrorTracer struct {
	err error
}

// NewErrorTracer creates a tracer that always returns the given error
func NewErrorTracer(err error) *ErrorTracer {
	return &ErrorTracer{err: err}
}

// StartSpan always returns an error span
func (e *ErrorTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &ErrorSpan{err: e.err}
}

// SpanFromContext returns an error span
func (e *ErrorTracer) SpanFromContext(ctx context.Context) Span {
	return &ErrorSpan{err: e.err}
}

// ContextWithSpan returns the context unchanged
func (e *ErrorTracer) ContextWithSpan(ctx context.Context, span Span) context.Context {
	return ctx
}

// Close returns the error
func (e *ErrorTracer) Close() error {
	return e.err
}

// GetMetrics returns empty metrics
func (e *ErrorTracer) GetMetrics() TracerMetrics {
	return TracerMetrics{}
}

// ErrorSpan implements Span interface but represents an error state
type ErrorSpan struct {
	err error
}

// Context returns an empty span context
func (e *ErrorSpan) Context() SpanContext {
	return SpanContext{}
}

// SetName is a no-op for error spans
func (e *ErrorSpan) SetName(name string) {}

// SetAttributes is a no-op for error spans
func (e *ErrorSpan) SetAttributes(attributes map[string]interface{}) {}

// SetAttribute is a no-op for error spans
func (e *ErrorSpan) SetAttribute(key string, value interface{}) {}

// AddEvent is a no-op for error spans
func (e *ErrorSpan) AddEvent(name string, attributes map[string]interface{}) {}

// SetStatus is a no-op for error spans
func (e *ErrorSpan) SetStatus(code StatusCode, message string) {}

// RecordError is a no-op for error spans
func (e *ErrorSpan) RecordError(err error, attributes map[string]interface{}) {}

// End is a no-op for error spans
func (e *ErrorSpan) End() {}

// IsRecording always returns false for error spans
func (e *ErrorSpan) IsRecording() bool {
	return false
}

// GetDuration returns zero duration
func (e *ErrorSpan) GetDuration() time.Duration {
	return 0
}

// String returns the error message
func (e *ErrorSpan) String() string {
	return fmt.Sprintf("ErrorSpan: %v", e.err)
}
