package prometheus

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/prometheus/client_golang/prometheus"
)

// Tracer implements tracer.Tracer for Prometheus with enhanced metrics collection
type Tracer struct {
	name        string
	provider    *Provider
	config      *tracer.TracerConfig
	metrics     tracer.TracerMetrics
	started     time.Time
	activeSpans map[string]*Span
	mu          sync.RWMutex
}

// StartSpan creates a new Prometheus span with comprehensive metrics
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...tracer.SpanOption) (context.Context, tracer.Span) {
	// Parse span options
	config := tracer.ApplySpanOptions(opts)

	// Generate unique span ID
	spanID := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())

	// Create wrapped span
	span := &Span{
		id:         spanID,
		name:       name,
		kind:       config.Kind,
		tracer:     t,
		startTime:  time.Now(),
		attributes: make(map[string]interface{}),
		labels:     t.provider.buildLabels(t.name, name, config.Kind, t.config),
	}

	// Add initial attributes
	for key, value := range config.Attributes {
		span.attributes[key] = value
	}

	// Add custom attributes from provider config
	for key, value := range t.provider.config.CustomAttributes {
		span.attributes[key] = value
	}

	// Track active span
	t.mu.Lock()
	t.activeSpans[spanID] = span
	t.mu.Unlock()

	// Update Prometheus metrics
	t.provider.spanCounter.With(span.labels).Inc()
	t.provider.activeSpans.WithLabelValues(
		t.provider.getServiceName(t.config),
		t.name,
	).Inc()

	// Update tracer metrics
	atomic.AddInt64(&t.metrics.SpansCreated, 1)
	t.metrics.LastActivity = time.Now()

	return ctx, span
}

// SpanFromContext extracts a span from context (not directly supported in Prometheus)
func (t *Tracer) SpanFromContext(ctx context.Context) tracer.Span {
	// Prometheus doesn't have native span context propagation
	// Return a noop span
	return &NoopSpan{tracer: t}
}

// ContextWithSpan returns the context unchanged (Prometheus doesn't use context propagation)
func (t *Tracer) ContextWithSpan(ctx context.Context, span tracer.Span) context.Context {
	return ctx
}

// Close shuts down the tracer
func (t *Tracer) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// End all active spans
	for _, span := range t.activeSpans {
		span.End()
	}
	t.activeSpans = make(map[string]*Span)

	// Update final metrics
	t.metrics.LastActivity = time.Now()
	return nil
}

// GetMetrics returns tracer metrics
func (t *Tracer) GetMetrics() tracer.TracerMetrics {
	// Calculate average span duration
	if t.metrics.SpansFinished > 0 {
		t.metrics.AvgSpanDuration = time.Since(t.started) / time.Duration(t.metrics.SpansFinished)
	}
	return t.metrics
}

// Span implements tracer.Span for Prometheus with comprehensive metrics
type Span struct {
	id         string
	name       string
	kind       tracer.SpanKind
	tracer     *Tracer
	startTime  time.Time
	endTime    time.Time
	attributes map[string]interface{}
	labels     prometheus.Labels
	mu         sync.RWMutex
}

// Context returns the span context
func (s *Span) Context() tracer.SpanContext {
	return tracer.SpanContext{
		TraceID: fmt.Sprintf("prom-trace-%d", time.Now().UnixNano()),
		SpanID:  s.id,
		Flags:   tracer.TraceFlagsSampled,
	}
}

// SetName sets the span operation name
func (s *Span) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.name = name
	// Update labels if needed
	s.labels["span_name"] = name
}

// SetAttributes sets multiple attributes
func (s *Span) SetAttributes(attributes map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range attributes {
		s.attributes[key] = value
	}
}

// SetAttribute sets a single attribute
func (s *Span) SetAttribute(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attributes[key] = value
}

// AddEvent adds a structured event (stored as attributes in Prometheus)
func (s *Span) AddEvent(name string, attributes map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store event as attributes with timestamp
	eventKey := fmt.Sprintf("event.%s.timestamp", name)
	s.attributes[eventKey] = time.Now().Unix()

	// Add event attributes with prefix
	for key, value := range attributes {
		attrKey := fmt.Sprintf("event.%s.%s", name, key)
		s.attributes[attrKey] = value
	}
}

// SetStatus sets the span status
func (s *Span) SetStatus(code tracer.StatusCode, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.attributes["status_code"] = code.String()
	if message != "" {
		s.attributes["status_message"] = message
	}

	// Update labels for metrics
	s.labels["status"] = code.String()

	// Record error if status is error
	if code == tracer.StatusCodeError {
		errorLabels := make(prometheus.Labels)
		for k, v := range s.labels {
			// Skip the status label since errorCounter doesn't have it in its base labels
			if k != "status" {
				errorLabels[k] = v
			}
		}
		errorLabels["error_type"] = "status_error"

		s.tracer.provider.errorCounter.With(errorLabels).Inc()
	}
}

// RecordError records an error with comprehensive metrics
func (s *Span) RecordError(err error, attributes map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store error details
	s.attributes["error"] = true
	s.attributes["error_message"] = err.Error()
	s.attributes["error_type"] = fmt.Sprintf("%T", err)
	s.attributes["error_timestamp"] = time.Now().Unix()

	// Add error attributes
	for key, value := range attributes {
		s.attributes[fmt.Sprintf("error.%s", key)] = value
	}

	// Record error metrics
	errorLabels := make(prometheus.Labels)
	for k, v := range s.labels {
		// Skip the status label since errorCounter doesn't have it in its base labels
		if k != "status" {
			errorLabels[k] = v
		}
	}
	errorLabels["error_type"] = fmt.Sprintf("%T", err)

	s.tracer.provider.errorCounter.With(errorLabels).Inc()

	// Set status to error (but avoid recursion by setting directly)
	s.attributes["status_code"] = tracer.StatusCodeError.String()
	if err.Error() != "" {
		s.attributes["status_message"] = err.Error()
	}
	s.labels["status"] = tracer.StatusCodeError.String()
}

// End finishes the span and records duration metrics
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.endTime.IsZero() {
		return // Already ended
	}

	s.endTime = time.Now()
	duration := s.endTime.Sub(s.startTime)

	// Create base labels for duration metric (without dynamic labels like status)
	durationLabels := make(prometheus.Labels)
	for k, v := range s.labels {
		// Skip dynamic labels that aren't in the metric definition
		if k != "status" {
			durationLabels[k] = v
		}
	}

	// Record duration metrics
	s.tracer.provider.spanDuration.With(durationLabels).Observe(duration.Seconds())

	// Update active spans count
	s.tracer.provider.activeSpans.WithLabelValues(
		s.tracer.provider.getServiceName(s.tracer.config),
		s.tracer.name,
	).Dec()

	// Remove from active spans
	s.tracer.mu.Lock()
	delete(s.tracer.activeSpans, s.id)
	s.tracer.mu.Unlock()

	// Update tracer metrics
	if s.tracer != nil {
		atomic.AddInt64(&s.tracer.metrics.SpansFinished, 1)
		s.tracer.metrics.LastActivity = time.Now()
	}
}

// IsRecording returns true if the span is recording
func (s *Span) IsRecording() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.endTime.IsZero()
}

// GetDuration returns the span duration
func (s *Span) GetDuration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.endTime.IsZero() {
		return time.Since(s.startTime)
	}
	return s.endTime.Sub(s.startTime)
}

// NoopSpan is a no-op span implementation for Prometheus
type NoopSpan struct {
	tracer *Tracer
}

func (n *NoopSpan) Context() tracer.SpanContext                              { return tracer.SpanContext{} }
func (n *NoopSpan) SetName(name string)                                      {}
func (n *NoopSpan) SetAttributes(attributes map[string]interface{})          {}
func (n *NoopSpan) SetAttribute(key string, value interface{})               {}
func (n *NoopSpan) AddEvent(name string, attributes map[string]interface{})  {}
func (n *NoopSpan) SetStatus(code tracer.StatusCode, message string)         {}
func (n *NoopSpan) RecordError(err error, attributes map[string]interface{}) {}
func (n *NoopSpan) End()                                                     {}
func (n *NoopSpan) IsRecording() bool                                        { return false }
func (n *NoopSpan) GetDuration() time.Duration                               { return 0 }
