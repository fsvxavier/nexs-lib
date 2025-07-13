// Package tracer provides a modern, extensible distributed tracing abstraction for Go applications.
// This package follows OpenTelemetry standards and supports multiple tracing backends including
// Datadog, New Relic, and Prometheus with full observability capabilities.
package tracer

import (
	"context"
	"time"
)

// Tracer represents a distributed tracing implementation with full lifecycle management
type Tracer interface {
	// StartSpan creates a new span with the given name and options
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// SpanFromContext extracts a span from the context if available
	SpanFromContext(ctx context.Context) Span

	// ContextWithSpan returns a new context with the span attached
	ContextWithSpan(ctx context.Context, span Span) context.Context

	// Close shuts down the tracer and flushes any pending data
	Close() error

	// GetMetrics returns tracer-specific metrics
	GetMetrics() TracerMetrics
}

// Span represents a single operation within a distributed trace
type Span interface {
	// Context returns the span context for propagation
	Context() SpanContext

	// SetName updates the span operation name
	SetName(name string)

	// SetAttributes sets multiple key-value attributes on the span
	SetAttributes(attributes map[string]interface{})

	// SetAttribute sets a single attribute on the span
	SetAttribute(key string, value interface{})

	// AddEvent adds a structured event to the span with optional attributes
	AddEvent(name string, attributes map[string]interface{})

	// SetStatus sets the span status with code and optional message
	SetStatus(code StatusCode, message string)

	// RecordError records an error on the span with optional attributes
	RecordError(err error, attributes map[string]interface{})

	// End finishes the span and sends it for processing
	End()

	// IsRecording returns true if the span is actively recording
	IsRecording() bool

	// GetDuration returns the span duration (only valid after End())
	GetDuration() time.Duration
}

// SpanContext contains trace identification information for propagation
type SpanContext struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Flags      TraceFlags        `json:"flags"`
	TraceState map[string]string `json:"trace_state,omitempty"`
	Baggage    map[string]string `json:"baggage,omitempty"`
}

// TraceFlags represents trace context flags
type TraceFlags byte

const (
	// TraceFlagsSampled indicates the trace is sampled
	TraceFlagsSampled TraceFlags = 1 << iota
	// TraceFlagsDebug indicates debug mode is enabled
	TraceFlagsDebug
)

// IsSampled returns true if the trace is sampled
func (f TraceFlags) IsSampled() bool {
	return f&TraceFlagsSampled != 0
}

// IsDebug returns true if debug mode is enabled
func (f TraceFlags) IsDebug() bool {
	return f&TraceFlagsDebug != 0
}

// StatusCode represents the status of a span operation
type StatusCode int32

const (
	// StatusCodeUnset indicates no status has been set
	StatusCodeUnset StatusCode = 0
	// StatusCodeOk indicates successful operation
	StatusCodeOk StatusCode = 1
	// StatusCodeError indicates operation failed
	StatusCodeError StatusCode = 2
)

// String returns the string representation of the status code
func (s StatusCode) String() string {
	switch s {
	case StatusCodeUnset:
		return "UNSET"
	case StatusCodeOk:
		return "OK"
	case StatusCodeError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// SpanKind represents the kind of span operation
type SpanKind int32

const (
	// SpanKindUnspecified indicates unspecified span kind
	SpanKindUnspecified SpanKind = 0
	// SpanKindInternal indicates internal application operation
	SpanKindInternal SpanKind = 1
	// SpanKindServer indicates server-side operation
	SpanKindServer SpanKind = 2
	// SpanKindClient indicates client-side operation
	SpanKindClient SpanKind = 3
	// SpanKindProducer indicates message producer operation
	SpanKindProducer SpanKind = 4
	// SpanKindConsumer indicates message consumer operation
	SpanKindConsumer SpanKind = 5
)

// String returns the string representation of the span kind
func (k SpanKind) String() string {
	switch k {
	case SpanKindUnspecified:
		return "UNSPECIFIED"
	case SpanKindInternal:
		return "INTERNAL"
	case SpanKindServer:
		return "SERVER"
	case SpanKindClient:
		return "CLIENT"
	case SpanKindProducer:
		return "PRODUCER"
	case SpanKindConsumer:
		return "CONSUMER"
	default:
		return "UNKNOWN"
	}
}

// Provider interface for tracer backend implementations
type Provider interface {
	// Name returns the unique provider name
	Name() string

	// CreateTracer creates a new tracer instance with the given name and options
	CreateTracer(name string, options ...TracerOption) (Tracer, error)

	// Shutdown gracefully shuts down the provider and flushes pending data
	Shutdown(ctx context.Context) error

	// HealthCheck performs a health check on the provider
	HealthCheck(ctx context.Context) error

	// GetProviderMetrics returns provider-specific metrics
	GetProviderMetrics() ProviderMetrics
}

// TracerMetrics contains metrics about tracer performance
type TracerMetrics struct {
	SpansCreated    int64         `json:"spans_created"`
	SpansFinished   int64         `json:"spans_finished"`
	SpansDropped    int64         `json:"spans_dropped"`
	AvgSpanDuration time.Duration `json:"avg_span_duration"`
	LastActivity    time.Time     `json:"last_activity"`
}

// ProviderMetrics contains metrics about provider performance
type ProviderMetrics struct {
	TracersActive   int       `json:"tracers_active"`
	ConnectionState string    `json:"connection_state"`
	LastFlush       time.Time `json:"last_flush"`
	ErrorCount      int64     `json:"error_count"`
	BytesSent       int64     `json:"bytes_sent"`
}

// SpanOption configures span creation behavior
type SpanOption interface {
	apply(*SpanConfig)
}

// TracerOption configures tracer creation behavior
type TracerOption interface {
	apply(*TracerConfig)
}

// SpanConfig holds configuration for creating spans
type SpanConfig struct {
	Kind       SpanKind               `json:"kind"`
	StartTime  time.Time              `json:"start_time"`
	Attributes map[string]interface{} `json:"attributes"`
	Parent     *SpanContext           `json:"parent"`
	Links      []SpanLink             `json:"links"`
}

// TracerConfig holds configuration for creating tracers
type TracerConfig struct {
	ServiceName     string                 `json:"service_name"`
	ServiceVersion  string                 `json:"service_version"`
	Environment     string                 `json:"environment"`
	Attributes      map[string]interface{} `json:"attributes"`
	SamplingRate    float64                `json:"sampling_rate"`
	EnableProfiling bool                   `json:"enable_profiling"`
	BatchSize       int                    `json:"batch_size"`
	FlushInterval   time.Duration          `json:"flush_interval"`
}

// SpanLink represents a link to another span
type SpanLink struct {
	SpanContext SpanContext            `json:"span_context"`
	Attributes  map[string]interface{} `json:"attributes"`
}
