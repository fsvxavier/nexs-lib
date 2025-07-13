// Package tracer provides OpenTelemetry integration support
package tracer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// OpenTelemetryConfig represents configuration for OpenTelemetry tracer
type OpenTelemetryConfig struct {
	ServiceName      string            `json:"service_name"`
	ServiceVersion   string            `json:"service_version"`
	ServiceNamespace string            `json:"service_namespace,omitempty"`
	Endpoint         string            `json:"endpoint"`
	Headers          map[string]string `json:"headers,omitempty"`
	Insecure         bool              `json:"insecure"`
	Timeout          time.Duration     `json:"timeout"`
	BatchTimeout     time.Duration     `json:"batch_timeout"`
	MaxExportBatch   int               `json:"max_export_batch"`
	MaxQueueSize     int               `json:"max_queue_size"`

	// Sampling configuration
	SamplingRatio float64 `json:"sampling_ratio"`

	// Resource attributes
	ResourceAttrs map[string]string `json:"resource_attrs,omitempty"`

	// Propagation configuration
	Propagators []string `json:"propagators,omitempty"`
}

// DefaultOpenTelemetryConfig returns default configuration for OpenTelemetry
func DefaultOpenTelemetryConfig() *OpenTelemetryConfig {
	return &OpenTelemetryConfig{
		ServiceName:    "unknown-service",
		ServiceVersion: "1.0.0",
		Endpoint:       "localhost:4317",
		Insecure:       true,
		Timeout:        30 * time.Second,
		BatchTimeout:   5 * time.Second,
		MaxExportBatch: 512,
		MaxQueueSize:   2048,
		SamplingRatio:  1.0,
		Propagators:    []string{"tracecontext", "baggage"},
		ResourceAttrs:  make(map[string]string),
	}
}

// OpenTelemetryTracer implements Tracer interface with OpenTelemetry
type OpenTelemetryTracer struct {
	provider   *sdktrace.TracerProvider
	tracer     trace.Tracer
	config     *OpenTelemetryConfig
	exporter   sdktrace.SpanExporter
	propagator propagation.TextMapPropagator
	resource   *resource.Resource
	metrics    *TracerMetrics
}

// NewOpenTelemetryTracer creates a new OpenTelemetry tracer
func NewOpenTelemetryTracer(config *OpenTelemetryConfig) (*OpenTelemetryTracer, error) {
	if config == nil {
		config = DefaultOpenTelemetryConfig()
	}

	if err := validateOpenTelemetryConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create OTLP exporter
	exporter, err := createOTLPExporter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := createResource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(config.BatchTimeout),
			sdktrace.WithMaxExportBatchSize(config.MaxExportBatch),
			sdktrace.WithMaxQueueSize(config.MaxQueueSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SamplingRatio)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(provider)

	// Create propagator
	propagator := createPropagator(config.Propagators)
	otel.SetTextMapPropagator(propagator)

	tracer := provider.Tracer(
		config.ServiceName,
		trace.WithInstrumentationVersion(config.ServiceVersion),
	)

	return &OpenTelemetryTracer{
		provider:   provider,
		tracer:     tracer,
		config:     config,
		exporter:   exporter,
		propagator: propagator,
		resource:   res,
		metrics:    &TracerMetrics{},
	}, nil
}

// StartSpan creates a new span with OpenTelemetry
func (ot *OpenTelemetryTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	config := &SpanConfig{}
	for _, opt := range opts {
		opt.apply(config)
	}

	// Convert to OpenTelemetry options
	otelOpts := []trace.SpanStartOption{
		trace.WithSpanKind(convertSpanKind(config.Kind)),
	}

	if !config.StartTime.IsZero() {
		otelOpts = append(otelOpts, trace.WithTimestamp(config.StartTime))
	}

	if len(config.Attributes) > 0 {
		attrs := make([]attribute.KeyValue, 0, len(config.Attributes))
		for k, v := range config.Attributes {
			attrs = append(attrs, convertAttribute(k, v))
		}
		otelOpts = append(otelOpts, trace.WithAttributes(attrs...))
	}

	ctx, otelSpan := ot.tracer.Start(ctx, name, otelOpts...)

	span := &OpenTelemetrySpan{
		span:    otelSpan,
		tracer:  ot,
		context: ctx,
	}

	ot.metrics.SpansCreated++
	return ctx, span
}

// SpanFromContext extracts a span from context
func (ot *OpenTelemetryTracer) SpanFromContext(ctx context.Context) Span {
	otelSpan := trace.SpanFromContext(ctx)
	if otelSpan == nil || !otelSpan.IsRecording() {
		return nil
	}

	return &OpenTelemetrySpan{
		span:    otelSpan,
		tracer:  ot,
		context: ctx,
	}
}

// ContextWithSpan returns a new context with the span attached
func (ot *OpenTelemetryTracer) ContextWithSpan(ctx context.Context, span Span) context.Context {
	if otelSpan, ok := span.(*OpenTelemetrySpan); ok {
		return trace.ContextWithSpan(ctx, otelSpan.span)
	}
	return ctx
}

// Close shuts down the tracer
func (ot *OpenTelemetryTracer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ot.config.Timeout)
	defer cancel()

	if err := ot.provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown tracer provider: %w", err)
	}

	return nil
}

// GetMetrics returns tracer metrics
func (ot *OpenTelemetryTracer) GetMetrics() TracerMetrics {
	return *ot.metrics
}

// OpenTelemetrySpan implements Span interface with OpenTelemetry
type OpenTelemetrySpan struct {
	span    trace.Span
	tracer  *OpenTelemetryTracer
	context context.Context
}

// Context returns the span context for propagation
func (os *OpenTelemetrySpan) Context() SpanContext {
	spanCtx := os.span.SpanContext()
	return SpanContext{
		TraceID:    spanCtx.TraceID().String(),
		SpanID:     spanCtx.SpanID().String(),
		Flags:      TraceFlags(spanCtx.TraceFlags()),
		TraceState: convertTraceState(spanCtx.TraceState()),
	}
}

// SetName updates the span operation name
func (os *OpenTelemetrySpan) SetName(name string) {
	os.span.SetName(name)
}

// SetAttributes sets multiple key-value attributes on the span
func (os *OpenTelemetrySpan) SetAttributes(attributes map[string]interface{}) {
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		attrs = append(attrs, convertAttribute(k, v))
	}
	os.span.SetAttributes(attrs...)
}

// SetAttribute sets a single attribute on the span
func (os *OpenTelemetrySpan) SetAttribute(key string, value interface{}) {
	os.span.SetAttributes(convertAttribute(key, value))
}

// AddEvent adds a structured event to the span
func (os *OpenTelemetrySpan) AddEvent(name string, attributes map[string]interface{}) {
	opts := []trace.EventOption{}
	if len(attributes) > 0 {
		attrs := make([]attribute.KeyValue, 0, len(attributes))
		for k, v := range attributes {
			attrs = append(attrs, convertAttribute(k, v))
		}
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	os.span.AddEvent(name, opts...)
}

// SetStatus sets the span status
func (os *OpenTelemetrySpan) SetStatus(code StatusCode, message string) {
	switch code {
	case StatusCodeOk:
		os.span.SetStatus(codes.Ok, message)
	case StatusCodeError:
		os.span.SetStatus(codes.Error, message)
	default:
		os.span.SetStatus(codes.Unset, message)
	}
}

// RecordError records an error on the span
func (os *OpenTelemetrySpan) RecordError(err error, attributes map[string]interface{}) {
	opts := []trace.EventOption{}
	if len(attributes) > 0 {
		attrs := make([]attribute.KeyValue, 0, len(attributes))
		for k, v := range attributes {
			attrs = append(attrs, convertAttribute(k, v))
		}
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	os.span.RecordError(err, opts...)
	os.tracer.metrics.SpansDropped++
}

// End finishes the span
func (os *OpenTelemetrySpan) End() {
	os.span.End()
	os.tracer.metrics.SpansFinished++
}

// IsRecording returns true if the span is actively recording
func (os *OpenTelemetrySpan) IsRecording() bool {
	return os.span.IsRecording()
}

// GetDuration returns the span duration (only valid after End())
func (os *OpenTelemetrySpan) GetDuration() time.Duration {
	// OpenTelemetry doesn't provide direct access to duration
	// This would need to be tracked manually if needed
	return 0
}

// Helper functions

func validateOpenTelemetryConfig(config *OpenTelemetryConfig) error {
	if config.ServiceName == "" {
		return errors.New("service name is required")
	}
	if config.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if config.SamplingRatio < 0 || config.SamplingRatio > 1 {
		return errors.New("sampling ratio must be between 0 and 1")
	}
	if config.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	return nil
}

func createOTLPExporter(config *OpenTelemetryConfig) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithTimeout(config.Timeout),
	}

	if config.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if len(config.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(config.Headers))
	}

	client := otlptracegrpc.NewClient(opts...)
	return otlptrace.New(context.Background(), client)
}

func createResource(config *OpenTelemetryConfig) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(config.ServiceName),
		semconv.ServiceVersion(config.ServiceVersion),
	}

	if config.ServiceNamespace != "" {
		attrs = append(attrs, semconv.ServiceNamespace(config.ServiceNamespace))
	}

	for k, v := range config.ResourceAttrs {
		attrs = append(attrs, attribute.String(k, v))
	}

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	), nil
}

func createPropagator(propagators []string) propagation.TextMapPropagator {
	var props []propagation.TextMapPropagator

	for _, p := range propagators {
		switch p {
		case "tracecontext":
			props = append(props, propagation.TraceContext{})
		case "baggage":
			props = append(props, propagation.Baggage{})
		case "b3":
			// B3 propagator would need to be imported separately
			// props = append(props, b3.New())
		}
	}

	if len(props) == 0 {
		props = []propagation.TextMapPropagator{
			propagation.TraceContext{},
			propagation.Baggage{},
		}
	}

	return propagation.NewCompositeTextMapPropagator(props...)
}

func convertSpanKind(kind SpanKind) trace.SpanKind {
	switch kind {
	case SpanKindClient:
		return trace.SpanKindClient
	case SpanKindServer:
		return trace.SpanKindServer
	case SpanKindProducer:
		return trace.SpanKindProducer
	case SpanKindConsumer:
		return trace.SpanKindConsumer
	default:
		return trace.SpanKindInternal
	}
}

func convertAttribute(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	case []string:
		return attribute.StringSlice(key, v)
	case []int:
		return attribute.IntSlice(key, v)
	case []int64:
		return attribute.Int64Slice(key, v)
	case []float64:
		return attribute.Float64Slice(key, v)
	case []bool:
		return attribute.BoolSlice(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}

func convertTraceState(ts trace.TraceState) map[string]string {
	result := make(map[string]string)
	ts.Walk(func(key, value string) bool {
		result[key] = value
		return true
	})
	return result
}
