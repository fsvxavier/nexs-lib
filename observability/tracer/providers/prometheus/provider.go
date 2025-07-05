// Package prometheus provides Prometheus/Grafana tracing and metrics implementation
package prometheus

import (
	"context"
	"fmt"
	"sync"
	"time"

	tracer "github.com/fsvxavier/nexs-lib/observability/tracer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Provider implements tracer.Provider for Prometheus/Grafana
type Provider struct {
	config   *Config
	registry *prometheus.Registry
	tracers  map[string]*Tracer
	mutex    sync.RWMutex

	// Metrics collectors
	spanCounter     *prometheus.CounterVec
	spanDuration    *prometheus.HistogramVec
	spanErrors      *prometheus.CounterVec
	activeSpans     *prometheus.GaugeVec
	operationErrors *prometheus.CounterVec
}

// Config holds Prometheus-specific configuration
type Config struct {
	ServiceName     string
	ServiceVersion  string
	Environment     string
	Namespace       string
	Subsystem       string
	Registry        *prometheus.Registry
	Labels          map[string]string
	EnableDuration  bool
	EnableErrors    bool
	EnableActive    bool
	DurationBuckets []float64
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		ServiceName:     "unknown-service",
		ServiceVersion:  "1.0.0",
		Environment:     "development",
		Namespace:       "tracer",
		Subsystem:       "spans",
		Registry:        prometheus.NewRegistry(),
		Labels:          make(map[string]string),
		EnableDuration:  true,
		EnableErrors:    true,
		EnableActive:    true,
		DurationBuckets: prometheus.DefBuckets,
	}
}

// NewProvider creates a new Prometheus provider
func NewProvider(config *Config) *Provider {
	if config == nil {
		config = DefaultConfig()
	}

	// Use provided registry or create a new one
	registry := config.Registry
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	// Common labels
	constLabels := prometheus.Labels{
		"service_name":    config.ServiceName,
		"service_version": config.ServiceVersion,
		"environment":     config.Environment,
	}

	// Add custom labels
	for k, v := range config.Labels {
		constLabels[k] = v
	}

	provider := &Provider{
		config:   config,
		registry: registry,
		tracers:  make(map[string]*Tracer),
	}

	// Initialize metrics
	provider.initMetrics(constLabels)

	return provider
}

// initMetrics initializes Prometheus metrics
func (p *Provider) initMetrics(constLabels prometheus.Labels) {
	// Span counter - counts total spans
	p.spanCounter = promauto.With(p.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   p.config.Namespace,
			Subsystem:   p.config.Subsystem,
			Name:        "total",
			Help:        "Total number of spans created",
			ConstLabels: constLabels,
		},
		[]string{"operation", "span_kind", "status"},
	)

	// Span duration histogram
	if p.config.EnableDuration {
		p.spanDuration = promauto.With(p.registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   p.config.Namespace,
				Subsystem:   p.config.Subsystem,
				Name:        "duration_seconds",
				Help:        "Duration of spans in seconds",
				Buckets:     p.config.DurationBuckets,
				ConstLabels: constLabels,
			},
			[]string{"operation", "span_kind", "status"},
		)
	}

	// Span errors counter
	if p.config.EnableErrors {
		p.spanErrors = promauto.With(p.registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   p.config.Namespace,
				Subsystem:   p.config.Subsystem,
				Name:        "errors_total",
				Help:        "Total number of span errors",
				ConstLabels: constLabels,
			},
			[]string{"operation", "span_kind", "error_type"},
		)

		p.operationErrors = promauto.With(p.registry).NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   p.config.Namespace,
				Subsystem:   "operations",
				Name:        "errors_total",
				Help:        "Total number of operation errors",
				ConstLabels: constLabels,
			},
			[]string{"operation", "error_type"},
		)
	}

	// Active spans gauge
	if p.config.EnableActive {
		p.activeSpans = promauto.With(p.registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   p.config.Namespace,
				Subsystem:   p.config.Subsystem,
				Name:        "active",
				Help:        "Number of active spans",
				ConstLabels: constLabels,
			},
			[]string{"operation", "span_kind"},
		)
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "prometheus"
}

// CreateTracer creates a new Prometheus tracer
func (p *Provider) CreateTracer(name string, options ...tracer.TracerOption) tracer.Tracer {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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

	promTracer := &Tracer{
		name:     name,
		provider: p,
		config:   config,
	}

	p.tracers[name] = promTracer
	return promTracer
}

// Shutdown gracefully shuts down the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	// Prometheus doesn't need explicit shutdown
	return nil
}

// GetRegistry returns the Prometheus registry
func (p *Provider) GetRegistry() *prometheus.Registry {
	return p.registry
}

// Tracer implements tracer.Tracer for Prometheus
type Tracer struct {
	name     string
	provider *Provider
	config   *tracer.TracerConfig
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

	// Generate span and trace IDs
	spanID := generateID()
	traceID := spanID // Simple implementation - in real scenarios, you'd extract from parent

	// Create span context
	spanCtx := tracer.SpanContext{
		TraceID: traceID,
		SpanID:  spanID,
		Flags:   1, // Sampled
	}

	span := &Span{
		name:       name,
		spanCtx:    spanCtx,
		startTime:  config.StartTime,
		attributes: config.Attributes,
		kind:       config.Kind,
		tracer:     t,
		provider:   t.provider,
	}

	// Set initial attributes
	if len(config.Attributes) > 0 {
		span.SetAttributes(config.Attributes)
	}

	// Update metrics
	spanKindStr := spanKindToString(config.Kind)

	// Increment span counter
	t.provider.spanCounter.WithLabelValues(name, spanKindStr, "started").Inc()

	// Increment active spans
	if t.provider.activeSpans != nil {
		t.provider.activeSpans.WithLabelValues(name, spanKindStr).Inc()
	}

	// Store span in context
	ctx = context.WithValue(ctx, spanContextKey, span)

	return ctx, span
}

// SpanFromContext extracts a span from the context
func (t *Tracer) SpanFromContext(ctx context.Context) tracer.Span {
	if span, ok := ctx.Value(spanContextKey).(*Span); ok {
		return span
	}
	return &tracer.NoopSpan{}
}

// ContextWithSpan returns a new context with the span attached
func (t *Tracer) ContextWithSpan(ctx context.Context, span tracer.Span) context.Context {
	return context.WithValue(ctx, spanContextKey, span)
}

// Close shuts down the tracer
func (t *Tracer) Close() error {
	return nil
}

// Span implements tracer.Span for Prometheus
type Span struct {
	name       string
	spanCtx    tracer.SpanContext
	startTime  time.Time
	endTime    time.Time
	attributes map[string]interface{}
	events     []SpanEvent
	kind       tracer.SpanKind
	status     tracer.StatusCode
	statusMsg  string
	tracer     *Tracer
	provider   *Provider
	ended      bool
	mutex      sync.RWMutex
}

// SpanEvent represents an event within a span
type SpanEvent struct {
	Name       string
	Time       time.Time
	Attributes map[string]interface{}
}

// Context returns the span context
func (s *Span) Context() tracer.SpanContext {
	return s.spanCtx
}

// SetName sets the span name
func (s *Span) SetName(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.name = name
}

// SetAttributes sets key-value attributes on the span
func (s *Span) SetAttributes(attributes map[string]interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.attributes == nil {
		s.attributes = make(map[string]interface{})
	}

	for k, v := range attributes {
		s.attributes[k] = v
	}
}

// SetAttribute sets a single attribute
func (s *Span) SetAttribute(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.attributes == nil {
		s.attributes = make(map[string]interface{})
	}
	s.attributes[key] = value
}

// AddEvent adds a structured event to the span
func (s *Span) AddEvent(name string, attributes map[string]interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := SpanEvent{
		Name:       name,
		Time:       time.Now(),
		Attributes: attributes,
	}
	s.events = append(s.events, event)
}

// SetStatus sets the span status
func (s *Span) SetStatus(code tracer.StatusCode, message string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.status = code
	s.statusMsg = message
}

// RecordError records an error on the span
func (s *Span) RecordError(err error, attributes map[string]interface{}) {
	s.SetStatus(tracer.StatusCodeError, err.Error())

	// Add error attributes
	s.SetAttribute("error", true)
	s.SetAttribute("error.message", err.Error())
	s.SetAttribute("error.type", fmt.Sprintf("%T", err))

	for k, v := range attributes {
		s.SetAttribute(fmt.Sprintf("error.%s", k), v)
	}

	// Increment error metrics
	if s.provider.spanErrors != nil {
		spanKindStr := spanKindToString(s.kind)
		errorType := fmt.Sprintf("%T", err)
		s.provider.spanErrors.WithLabelValues(s.name, spanKindStr, errorType).Inc()
	}

	if s.provider.operationErrors != nil {
		errorType := fmt.Sprintf("%T", err)
		s.provider.operationErrors.WithLabelValues(s.name, errorType).Inc()
	}
}

// End finishes the span
func (s *Span) End() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.ended {
		return
	}
	s.ended = true
	s.endTime = time.Now()

	// Calculate duration
	duration := s.endTime.Sub(s.startTime)

	// Update metrics
	spanKindStr := spanKindToString(s.kind)
	statusStr := statusCodeToString(s.status)

	// Update span counter with final status
	s.provider.spanCounter.WithLabelValues(s.name, spanKindStr, statusStr).Inc()

	// Record duration
	if s.provider.spanDuration != nil {
		s.provider.spanDuration.WithLabelValues(s.name, spanKindStr, statusStr).Observe(duration.Seconds())
	}

	// Decrement active spans
	if s.provider.activeSpans != nil {
		s.provider.activeSpans.WithLabelValues(s.name, spanKindStr).Dec()
	}
}

// IsRecording returns true if the span is recording
func (s *Span) IsRecording() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return !s.ended
}

// Context key for span storage
type contextKey string

const spanContextKey contextKey = "prometheus-span"

// Helper functions

// generateID generates a simple ID for spans/traces
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// spanKindToString converts span kind to string
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

// statusCodeToString converts status code to string
func statusCodeToString(code tracer.StatusCode) string {
	switch code {
	case tracer.StatusCodeOk:
		return "ok"
	case tracer.StatusCodeError:
		return "error"
	default:
		return "unset"
	}
}
