// Package newrelic provides New Relic APM tracing implementation
package newrelic

import (
	"context"
	"fmt"
	"time"

	tracer "github.com/fsvxavier/nexs-lib/observability/tracer"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// Provider implements tracer.Provider for New Relic APM
type Provider struct {
	config      *Config
	application *newrelic.Application
	tracers     map[string]*Tracer
}

// Config holds New Relic-specific configuration
type Config struct {
	AppName           string
	LicenseKey        string
	Environment       string
	ServiceVersion    string
	DistributedTracer bool
	Enabled           bool
	LogLevel          string
	Attributes        map[string]interface{}
	Labels            map[string]string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		AppName:           "unknown-service",
		LicenseKey:        "",
		Environment:       "development",
		ServiceVersion:    "1.0.0",
		DistributedTracer: true,
		Enabled:           true,
		LogLevel:          "info",
		Attributes:        make(map[string]interface{}),
		Labels:            make(map[string]string),
	}
}

// NewProvider creates a new New Relic provider
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.LicenseKey == "" {
		return nil, fmt.Errorf("New Relic license key is required")
	}
	// Create New Relic application
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(config.AppName),
		newrelic.ConfigLicense(config.LicenseKey),
		newrelic.ConfigDistributedTracerEnabled(config.DistributedTracer),
		newrelic.ConfigEnabled(config.Enabled),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create New Relic application: %w", err)
	}

	return &Provider{
		config:      config,
		application: app,
		tracers:     make(map[string]*Tracer),
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "newrelic"
}

// CreateTracer creates a new New Relic tracer
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

	// Create and cache tracer instance
	if t, exists := p.tracers[name]; exists {
		return t
	}

	nrTracer := &Tracer{
		name:        name,
		application: p.application,
		provider:    p,
		config:      config,
	}

	p.tracers[name] = nrTracer
	return nrTracer
}

// Shutdown gracefully shuts down the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.application != nil {
		p.application.Shutdown(10 * time.Second)
	}
	return nil
}

// Tracer implements tracer.Tracer for New Relic
type Tracer struct {
	name        string
	application *newrelic.Application
	provider    *Provider
	config      *tracer.TracerConfig
}

// StartSpan creates a new span with the given name and options
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

	// Check if there's already a transaction in context
	var txn *newrelic.Transaction
	if existingTxn := newrelic.FromContext(ctx); existingTxn != nil {
		// Start a segment within existing transaction
		segment := existingTxn.StartSegment(name)
		span := &Span{
			name:       name,
			segment:    segment,
			txn:        existingTxn,
			startTime:  config.StartTime,
			attributes: config.Attributes,
			tracer:     t,
		}

		// Set initial attributes
		if len(config.Attributes) > 0 {
			span.SetAttributes(config.Attributes)
		}

		return ctx, span
	}

	// Start new transaction
	txn = t.application.StartTransaction(name)

	// Set transaction attributes
	if t.config.ServiceName != "" {
		txn.AddAttribute("service.name", t.config.ServiceName)
	}
	if t.config.ServiceVersion != "" {
		txn.AddAttribute("service.version", t.config.ServiceVersion)
	}
	if t.config.Environment != "" {
		txn.AddAttribute("environment", t.config.Environment)
	}

	// Add span kind as attribute
	txn.AddAttribute("span.kind", spanKindToString(config.Kind))

	span := &Span{
		name:       name,
		txn:        txn,
		startTime:  config.StartTime,
		attributes: config.Attributes,
		tracer:     t,
	}

	// Set initial attributes
	if len(config.Attributes) > 0 {
		span.SetAttributes(config.Attributes)
	}

	// Add transaction to context
	ctx = newrelic.NewContext(ctx, txn)

	return ctx, span
}

// SpanFromContext extracts a span from the context
func (t *Tracer) SpanFromContext(ctx context.Context) tracer.Span {
	if txn := newrelic.FromContext(ctx); txn != nil {
		return &Span{
			name:       "current-span",
			txn:        txn,
			startTime:  time.Now(),
			attributes: make(map[string]interface{}),
			tracer:     t,
		}
	}
	return &tracer.NoopSpan{}
}

// ContextWithSpan returns a new context with the span attached
func (t *Tracer) ContextWithSpan(ctx context.Context, span tracer.Span) context.Context {
	if nrSpan, ok := span.(*Span); ok && nrSpan.txn != nil {
		return newrelic.NewContext(ctx, nrSpan.txn)
	}
	return ctx
}

// Close shuts down the tracer
func (t *Tracer) Close() error {
	return nil
}

// Span implements tracer.Span for New Relic
type Span struct {
	name       string
	txn        *newrelic.Transaction
	segment    *newrelic.Segment
	startTime  time.Time
	attributes map[string]interface{}
	tracer     *Tracer
	ended      bool
}

// Context returns the span context
func (s *Span) Context() tracer.SpanContext {
	if s.txn == nil {
		return tracer.SpanContext{}
	}

	// Get trace metadata from New Relic transaction
	metadata := s.txn.GetTraceMetadata()

	return tracer.SpanContext{
		TraceID: metadata.TraceID,
		SpanID:  metadata.SpanID,
		Flags:   1, // Sampled
	}
}

// SetName sets the span name
func (s *Span) SetName(name string) {
	s.name = name
	if s.txn != nil {
		s.txn.SetName(name)
	}
}

// SetAttributes sets key-value attributes on the span
func (s *Span) SetAttributes(attributes map[string]interface{}) {
	if s.attributes == nil {
		s.attributes = make(map[string]interface{})
	}

	for k, v := range attributes {
		s.attributes[k] = v
		s.SetAttribute(k, v)
	}
}

// SetAttribute sets a single attribute
func (s *Span) SetAttribute(key string, value interface{}) {
	if s.attributes == nil {
		s.attributes = make(map[string]interface{})
	}
	s.attributes[key] = value

	if s.txn != nil {
		s.txn.AddAttribute(key, value)
	}
}

// AddEvent adds a structured event to the span
func (s *Span) AddEvent(name string, attributes map[string]interface{}) {
	if s.txn != nil {
		// New Relic doesn't have direct event support, so we add as attributes
		eventKey := fmt.Sprintf("event.%s", name)
		s.txn.AddAttribute(eventKey, name)

		for k, v := range attributes {
			attrKey := fmt.Sprintf("event.%s.%s", name, k)
			s.txn.AddAttribute(attrKey, v)
		}
	}
}

// SetStatus sets the span status
func (s *Span) SetStatus(code tracer.StatusCode, message string) {
	if s.txn != nil {
		switch code {
		case tracer.StatusCodeError:
			s.txn.NoticeError(fmt.Errorf(message))
		case tracer.StatusCodeOk:
			s.txn.AddAttribute("status", "ok")
		}

		if message != "" {
			s.txn.AddAttribute("status.message", message)
		}
	}
}

// RecordError records an error on the span
func (s *Span) RecordError(err error, attributes map[string]interface{}) {
	if s.txn != nil {
		s.txn.NoticeError(err)

		// Add error attributes
		for k, v := range attributes {
			errorKey := fmt.Sprintf("error.%s", k)
			s.txn.AddAttribute(errorKey, v)
		}
	}
}

// End finishes the span
func (s *Span) End() {
	if s.ended {
		return
	}
	s.ended = true

	if s.segment != nil {
		s.segment.End()
	} else if s.txn != nil {
		s.txn.End()
	}
}

// IsRecording returns true if the span is recording
func (s *Span) IsRecording() bool {
	return !s.ended && s.txn != nil
}

// Helper function to convert span kind to string
func spanKindToString(kind tracer.SpanKind) string {
	switch kind {
	case tracer.SpanKindInternal:
		return "internal"
	case tracer.SpanKindServer:
		return "server"
	case tracer.SpanKindClient:
		return "client"
	case tracer.SpanKindProducer:
		return "producer"
	case tracer.SpanKindConsumer:
		return "consumer"
	default:
		return "unspecified"
	}
}
