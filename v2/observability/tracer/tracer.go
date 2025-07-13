// Package tracer provides a comprehensive distributed tracing abstraction with
// support for multiple providers including Datadog, New Relic, and Prometheus.
package tracer

import (
	"context"
	"time"
)

// GlobalFactory is the global tracer factory instance
var GlobalFactory = NewFactory()

// GetProvider returns an existing provider from the global factory
func GetProvider(providerType ProviderType) (Provider, error) {
	return GlobalFactory.GetProvider(providerType)
}

// CreateTracerManager creates a tracer manager with the specified provider
func CreateTracerManager(providerType ProviderType, config interface{}) (*TracerManager, error) {
	return NewTracerManager(GlobalFactory, providerType, config)
}

// Shutdown shuts down all providers in the global factory
func Shutdown(ctx context.Context) error {
	return GlobalFactory.Shutdown(ctx)
}

// HealthCheck performs health check on all active providers
func HealthCheck(ctx context.Context) map[ProviderType]error {
	return GlobalFactory.HealthCheck(ctx)
}

// GetMetrics returns metrics from all active providers
func GetMetrics() map[ProviderType]ProviderMetrics {
	return GlobalFactory.GetMetrics()
}

// GetProviderInfo returns detailed information about all providers
func GetProviderInfo(ctx context.Context) []ProviderInfo {
	return GlobalFactory.GetProviderInfo(ctx)
}

// Multi-provider tracer for sending traces to multiple backends
type MultiProviderTracer struct {
	tracers []Tracer
	primary Tracer
}

// NewMultiProviderTracer creates a tracer that sends to multiple providers
func NewMultiProviderTracer(primary Tracer, additional ...Tracer) *MultiProviderTracer {
	tracers := append([]Tracer{primary}, additional...)
	return &MultiProviderTracer{
		tracers: tracers,
		primary: primary,
	}
}

// StartSpan creates spans in all configured tracers
func (m *MultiProviderTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	// Start span in primary tracer
	newCtx, primarySpan := m.primary.StartSpan(ctx, name, opts...)

	// Start spans in additional tracers
	var spans []Span
	spans = append(spans, primarySpan)

	for _, tracer := range m.tracers[1:] {
		_, span := tracer.StartSpan(ctx, name, opts...)
		spans = append(spans, span)
	}

	// Return multi-span wrapper
	multiSpan := &MultiSpan{spans: spans, primary: primarySpan}
	return newCtx, multiSpan
}

// SpanFromContext extracts span from primary tracer
func (m *MultiProviderTracer) SpanFromContext(ctx context.Context) Span {
	return m.primary.SpanFromContext(ctx)
}

// ContextWithSpan uses primary tracer for context propagation
func (m *MultiProviderTracer) ContextWithSpan(ctx context.Context, span Span) context.Context {
	return m.primary.ContextWithSpan(ctx, span)
}

// Close closes all tracers
func (m *MultiProviderTracer) Close() error {
	var lastError error
	for _, tracer := range m.tracers {
		if err := tracer.Close(); err != nil {
			lastError = err
		}
	}
	return lastError
}

// GetMetrics returns metrics from primary tracer
func (m *MultiProviderTracer) GetMetrics() TracerMetrics {
	return m.primary.GetMetrics()
}

// MultiSpan wraps multiple spans and delegates operations to all of them
type MultiSpan struct {
	spans   []Span
	primary Span
}

// Context returns context from primary span
func (m *MultiSpan) Context() SpanContext {
	return m.primary.Context()
}

// SetName sets name on all spans
func (m *MultiSpan) SetName(name string) {
	for _, span := range m.spans {
		span.SetName(name)
	}
}

// SetAttributes sets attributes on all spans
func (m *MultiSpan) SetAttributes(attributes map[string]interface{}) {
	for _, span := range m.spans {
		span.SetAttributes(attributes)
	}
}

// SetAttribute sets attribute on all spans
func (m *MultiSpan) SetAttribute(key string, value interface{}) {
	for _, span := range m.spans {
		span.SetAttribute(key, value)
	}
}

// AddEvent adds event to all spans
func (m *MultiSpan) AddEvent(name string, attributes map[string]interface{}) {
	for _, span := range m.spans {
		span.AddEvent(name, attributes)
	}
}

// SetStatus sets status on all spans
func (m *MultiSpan) SetStatus(code StatusCode, message string) {
	for _, span := range m.spans {
		span.SetStatus(code, message)
	}
}

// RecordError records error on all spans
func (m *MultiSpan) RecordError(err error, attributes map[string]interface{}) {
	for _, span := range m.spans {
		span.RecordError(err, attributes)
	}
}

// End ends all spans
func (m *MultiSpan) End() {
	for _, span := range m.spans {
		span.End()
	}
}

// IsRecording returns recording status from primary span
func (m *MultiSpan) IsRecording() bool {
	return m.primary.IsRecording()
}

// GetDuration returns duration from primary span
func (m *MultiSpan) GetDuration() time.Duration {
	return m.primary.GetDuration()
}
