// package tracer provides a modern, extensible tracing abstraction for Go applications.
// This package follows OpenTelemetry standards and supports multiple tracing backends.
package tracer

import (
	"context"
	"time"
)

// Tracer represents a distributed tracing implementation
type Tracer interface {
	// StartSpan creates a new span with the given name and options
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// SpanFromContext extracts a span from the context
	SpanFromContext(ctx context.Context) Span

	// ContextWithSpan returns a new context with the span attached
	ContextWithSpan(ctx context.Context, span Span) context.Context

	// Close shuts down the tracer
	Close() error
}

// Span represents a single operation within a trace
type Span interface {
	// Context returns the span context
	Context() SpanContext

	// SetName sets the span name
	SetName(name string)

	// SetAttributes sets key-value attributes on the span
	SetAttributes(attributes map[string]interface{})

	// SetAttribute sets a single attribute
	SetAttribute(key string, value interface{})

	// AddEvent adds a structured event to the span
	AddEvent(name string, attributes map[string]interface{})

	// SetStatus sets the span status
	SetStatus(code StatusCode, message string)

	// RecordError records an error on the span
	RecordError(err error, attributes map[string]interface{})

	// End finishes the span
	End()

	// IsRecording returns true if the span is recording
	IsRecording() bool
}

// SpanContext contains trace identification information
type SpanContext struct {
	TraceID string
	SpanID  string
	Flags   byte
}

// StatusCode represents the status of a span
type StatusCode int32

const (
	StatusCodeUnset StatusCode = 0
	StatusCodeOk    StatusCode = 1
	StatusCodeError StatusCode = 2
)

// SpanKind represents the kind of span
type SpanKind int32

const (
	SpanKindUnspecified SpanKind = 0
	SpanKindInternal    SpanKind = 1
	SpanKindServer      SpanKind = 2
	SpanKindClient      SpanKind = 3
	SpanKindProducer    SpanKind = 4
	SpanKindConsumer    SpanKind = 5
)

// SpanOption configures a span
type SpanOption interface {
	apply(*SpanConfig)
}

// SpanConfig holds configuration for creating spans
type SpanConfig struct {
	Kind       SpanKind
	StartTime  time.Time
	Attributes map[string]interface{}
	Parent     SpanContext
}

type spanOptionFunc func(*SpanConfig)

func (f spanOptionFunc) apply(config *SpanConfig) {
	f(config)
}

// WithSpanKind sets the span kind
func WithSpanKind(kind SpanKind) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		config.Kind = kind
	})
}

// WithStartTime sets the span start time
func WithStartTime(t time.Time) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		config.StartTime = t
	})
}

// WithAttributes sets initial attributes
func WithAttributes(attrs map[string]interface{}) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		if config.Attributes == nil {
			config.Attributes = make(map[string]interface{})
		}
		for k, v := range attrs {
			config.Attributes[k] = v
		}
	})
}

// WithParent sets the parent span context
func WithParent(parent SpanContext) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		config.Parent = parent
	})
}

// Provider interface for tracer implementations
type Provider interface {
	// Name returns the provider name
	Name() string

	// CreateTracer creates a new tracer instance
	CreateTracer(name string, options ...TracerOption) Tracer

	// Shutdown gracefully shuts down the provider
	Shutdown(ctx context.Context) error
}

// TracerOption configures a tracer
type TracerOption interface {
	apply(*TracerConfig)
}

// TracerConfig holds configuration for creating tracers
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Attributes     map[string]interface{}
}

type tracerOptionFunc func(*TracerConfig)

func (f tracerOptionFunc) apply(config *TracerConfig) {
	f(config)
}

// WithServiceName sets the service name
func WithServiceName(name string) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		config.ServiceName = name
	})
}

// WithServiceVersion sets the service version
func WithServiceVersion(version string) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		config.ServiceVersion = version
	})
}

// WithEnvironment sets the environment
func WithEnvironment(env string) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		config.Environment = env
	})
}

// WithTracerAttributes sets tracer-level attributes
func WithTracerAttributes(attrs map[string]interface{}) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		if config.Attributes == nil {
			config.Attributes = make(map[string]interface{})
		}
		for k, v := range attrs {
			config.Attributes[k] = v
		}
	})
}

// NoopTracer is a tracer that does nothing
type NoopTracer struct{}

func (n *NoopTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &NoopSpan{}
}

func (n *NoopTracer) SpanFromContext(ctx context.Context) Span {
	return &NoopSpan{}
}

func (n *NoopTracer) ContextWithSpan(ctx context.Context, span Span) context.Context {
	return ctx
}

func (n *NoopTracer) Close() error {
	return nil
}

// NoopSpan is a span that does nothing
type NoopSpan struct{}

func (n *NoopSpan) Context() SpanContext                                     { return SpanContext{} }
func (n *NoopSpan) SetName(name string)                                      {}
func (n *NoopSpan) SetAttributes(attributes map[string]interface{})          {}
func (n *NoopSpan) SetAttribute(key string, value interface{})               {}
func (n *NoopSpan) AddEvent(name string, attributes map[string]interface{})  {}
func (n *NoopSpan) SetStatus(code StatusCode, message string)                {}
func (n *NoopSpan) RecordError(err error, attributes map[string]interface{}) {}
func (n *NoopSpan) End()                                                     {}
func (n *NoopSpan) IsRecording() bool                                        { return false }
