// Package datadog provides Datadog APM tracing implementation
package datadog

import (
	"context"
	"fmt"
	"time"

	ddtracer "github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	tracer "github.com/fsvxavier/nexs-lib/observability/tracer"
)

// Provider implements tracer.Provider for Datadog APM
type Provider struct {
	config  *Config
	tracers map[string]*Tracer
	started bool
}

// Config holds Datadog-specific configuration
type Config struct {
	ServiceName     string
	ServiceVersion  string
	Environment     string
	AgentHost       string
	AgentPort       int
	EnableProfiling bool
	SampleRate      float64
	Tags            map[string]string
	Debug           bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		ServiceName:     "unknown-service",
		ServiceVersion:  "1.0.0",
		Environment:     "development",
		AgentHost:       "localhost",
		AgentPort:       8126,
		EnableProfiling: false,
		SampleRate:      1.0,
		Tags:            make(map[string]string),
		Debug:           false,
	}
}

// NewProvider creates a new Datadog provider
func NewProvider(config *Config) *Provider {
	if config == nil {
		config = DefaultConfig()
	}

	return &Provider{
		config:  config,
		tracers: make(map[string]*Tracer),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "datadog"
}

// CreateTracer creates a new Datadog tracer
func (p *Provider) CreateTracer(name string, options ...tracer.TracerOption) tracer.Tracer {
	// Parse tracer options
	config := &tracer.TracerConfig{}
	if len(options) > 0 {
		for _, opt := range options {
			if optFunc, ok := opt.(interface{ apply(*tracer.TracerConfig) }); ok {
				optFunc.apply(config)
			}
		}
	}

	// Start Datadog tracer if not already started
	if !p.started {
		p.startDatadogTracer(config)
		p.started = true
	}

	// Create and cache tracer instance
	if t, exists := p.tracers[name]; exists {
		return t
	}

	ddTracerInstance := &Tracer{
		name:     name,
		provider: p,
		config:   config,
	}

	p.tracers[name] = ddTracerInstance
	return ddTracerInstance
}

// startDatadogTracer initializes the Datadog tracer
func (p *Provider) startDatadogTracer(config *tracer.TracerConfig) {
	opts := []ddtracer.StartOption{
		ddtracer.WithService(p.getServiceName(config)),
		ddtracer.WithServiceVersion(p.getServiceVersion(config)),
		ddtracer.WithEnv(p.getEnvironment(config)),
	}

	// Agent configuration
	if p.config.AgentHost != "" && p.config.AgentPort > 0 {
		agentAddr := fmt.Sprintf("%s:%d", p.config.AgentHost, p.config.AgentPort)
		opts = append(opts, ddtracer.WithAgentAddr(agentAddr))
	}

	// Sampling rate
	if p.config.SampleRate > 0 && p.config.SampleRate <= 1.0 {
		opts = append(opts, ddtracer.WithSamplingRules([]ddtracer.SamplingRule{
			{Rate: p.config.SampleRate},
		}))
	}

	// Global tags
	for key, value := range p.config.Tags {
		opts = append(opts, ddtracer.WithGlobalTag(key, value))
	}

	// Debug mode
	if p.config.Debug {
		opts = append(opts, ddtracer.WithDebugMode(true))
	}

	// Enable profiling if configured
	if p.config.EnableProfiling {
		opts = append(opts, ddtracer.WithProfilerCodeHotspots(true))
		opts = append(opts, ddtracer.WithProfilerEndpoints(true))
	}

	//mocktracer.Start()
	ddtracer.Start(opts...)
}

// Shutdown gracefully shuts down the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.started {
		ddtracer.Stop()
		p.started = false
	}
	return nil
}

// Helper methods
func (p *Provider) getServiceName(config *tracer.TracerConfig) string {
	if config.ServiceName != "" {
		return config.ServiceName
	}
	return p.config.ServiceName
}

func (p *Provider) getServiceVersion(config *tracer.TracerConfig) string {
	if config.ServiceVersion != "" {
		return config.ServiceVersion
	}
	return p.config.ServiceVersion
}

func (p *Provider) getEnvironment(config *tracer.TracerConfig) string {
	if config.Environment != "" {
		return config.Environment
	}
	return p.config.Environment
}

// Tracer implements tracer.Tracer for Datadog
type Tracer struct {
	name     string
	provider *Provider
	config   *tracer.TracerConfig
}

// StartSpan creates a new span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...tracer.SpanOption) (context.Context, tracer.Span) {
	// Parse span options
	config := &tracer.SpanConfig{
		Kind:       tracer.SpanKindInternal,
		StartTime:  time.Now(),
		Attributes: make(map[string]interface{}),
	}

	for _, opt := range opts {
		if optFunc, ok := opt.(interface{ apply(*tracer.SpanConfig) }); ok {
			optFunc.apply(config)
		}
	}

	// Build Datadog span options
	ddOpts := []ddtracer.StartSpanOption{
		ddtracer.ServiceName(t.provider.getServiceName(t.config)),
	}

	// Set start time if specified
	if !config.StartTime.IsZero() {
		ddOpts = append(ddOpts, ddtracer.StartTime(config.StartTime))
	}

	// Convert span kind to Datadog span type
	if spanType := convertSpanKind(config.Kind); spanType != "" {
		ddOpts = append(ddOpts, ddtracer.SpanType(spanType))
	}

	// Add attributes as tags
	for key, value := range config.Attributes {
		ddOpts = append(ddOpts, ddtracer.Tag(key, fmt.Sprintf("%v", value)))
	}

	// Create Datadog span
	ddSpan, newCtx := ddtracer.StartSpanFromContext(ctx, name, ddOpts...)

	// Wrap in our span implementation
	span := &Span{
		ddSpan: ddSpan,
		tracer: t,
	}

	return newCtx, span
}

// SpanFromContext extracts a span from context
func (t *Tracer) SpanFromContext(ctx context.Context) tracer.Span {
	if ddSpan, ok := ddtracer.SpanFromContext(ctx); ok && ddSpan != nil {
		return &Span{
			ddSpan: ddSpan,
			tracer: t,
		}
	}
	return &NoopSpan{}
}

// ContextWithSpan returns a context with the span
func (t *Tracer) ContextWithSpan(ctx context.Context, span tracer.Span) context.Context {
	if ddSpan, ok := span.(*Span); ok {
		return ddtracer.ContextWithSpan(ctx, ddSpan.ddSpan)
	}
	return ctx
}

// Close shuts down the tracer
func (t *Tracer) Close() error {
	// Individual tracers don't need cleanup in Datadog
	return nil
}

// Span implements tracer.Span for Datadog
type Span struct {
	ddSpan *ddtracer.Span
	tracer *Tracer
}

// Context returns the span context
func (s *Span) Context() tracer.SpanContext {
	ctx := s.ddSpan.Context()
	return tracer.SpanContext{
		TraceID: ctx.TraceID(),
		SpanID:  fmt.Sprintf("%d", ctx.SpanID()),
		Flags:   1, // Sampled
	}
}

// SetName sets the span name
func (s *Span) SetName(name string) {
	s.ddSpan.SetOperationName(name)
}

// SetAttributes sets multiple attributes
func (s *Span) SetAttributes(attributes map[string]interface{}) {
	for key, value := range attributes {
		s.ddSpan.SetTag(key, value)
	}
}

// SetAttribute sets a single attribute
func (s *Span) SetAttribute(key string, value interface{}) {
	s.ddSpan.SetTag(key, value)
}

// AddEvent adds a structured event
func (s *Span) AddEvent(name string, attributes map[string]interface{}) {
	// Datadog doesn't have native events, so we simulate with tags
	eventPrefix := fmt.Sprintf("event.%s", name)
	s.ddSpan.SetTag(eventPrefix+".timestamp", time.Now().Unix())

	for key, value := range attributes {
		s.ddSpan.SetTag(fmt.Sprintf("%s.%s", eventPrefix, key), value)
	}
}

// SetStatus sets the span status
func (s *Span) SetStatus(code tracer.StatusCode, message string) {
	switch code {
	case tracer.StatusCodeError:
		s.ddSpan.SetTag("error", true)
		if message != "" {
			s.ddSpan.SetTag("error.message", message)
		}
	case tracer.StatusCodeOk:
		s.ddSpan.SetTag("error", false)
	}
}

// RecordError records an error
func (s *Span) RecordError(err error, attributes map[string]interface{}) {
	s.ddSpan.SetTag("error", true)
	s.ddSpan.SetTag("error.message", err.Error())
	s.ddSpan.SetTag("error.type", fmt.Sprintf("%T", err))

	// Add error attributes
	for key, value := range attributes {
		s.ddSpan.SetTag(fmt.Sprintf("error.%s", key), value)
	}
}

// End finishes the span
func (s *Span) End() {
	s.ddSpan.Finish()
}

// IsRecording returns true if the span is recording
func (s *Span) IsRecording() bool {
	// Datadog spans are always recording until finished
	return true
}

// NoopSpan is a no-op span implementation
type NoopSpan struct{}

func (n *NoopSpan) Context() tracer.SpanContext                              { return tracer.SpanContext{} }
func (n *NoopSpan) SetName(name string)                                      {}
func (n *NoopSpan) SetAttributes(attributes map[string]interface{})          {}
func (n *NoopSpan) SetAttribute(key string, value interface{})               {}
func (n *NoopSpan) AddEvent(name string, attributes map[string]interface{})  {}
func (n *NoopSpan) SetStatus(code tracer.StatusCode, message string)         {}
func (n *NoopSpan) RecordError(err error, attributes map[string]interface{}) {}
func (n *NoopSpan) End()                                                     {}
func (n *NoopSpan) IsRecording() bool                                        { return false }

// convertSpanKind converts our span kind to Datadog span type
func convertSpanKind(kind tracer.SpanKind) string {
	switch kind {
	case tracer.SpanKindServer:
		return "web"
	case tracer.SpanKindClient:
		return "http"
	case tracer.SpanKindProducer:
		return "queue"
	case tracer.SpanKindConsumer:
		return "queue"
	default:
		return ""
	}
}
