package datadog

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

// Tracer implements tracer.Tracer for Datadog with enhanced features
type Tracer struct {
	name     string
	provider *Provider
	config   *tracer.TracerConfig
	metrics  tracer.TracerMetrics
	started  time.Time
}

// StartSpan creates a new simulated Datadog span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...tracer.SpanOption) (context.Context, tracer.Span) {
	// Parse span options
	config := tracer.ApplySpanOptions(opts)

	// Create wrapped span (simplified implementation)
	span := &Span{
		tracer:    t,
		startTime: time.Now(),
		name:      name,
		attrs:     make(map[string]interface{}),
		recording: true,
	}

	// Apply start time if specified
	if !config.StartTime.IsZero() {
		span.startTime = config.StartTime
	}

	// Apply attributes
	for key, value := range config.Attributes {
		span.attrs[key] = value
	}

	// Apply custom attributes from provider config
	for key, value := range t.provider.config.CustomAttributes {
		span.attrs[key] = value
	}

	// Update metrics
	atomic.AddInt64(&t.metrics.SpansCreated, 1)
	t.metrics.LastActivity = time.Now()

	// Return context with span
	newCtx := context.WithValue(ctx, "datadog-span", span)
	return newCtx, span
}

// SpanFromContext extracts a span from context (simplified)
func (t *Tracer) SpanFromContext(ctx context.Context) tracer.Span {
	// Check if there's a span value in the context
	if span, ok := ctx.Value("datadog-span").(*Span); ok && span != nil {
		return span
	}
	// Return a noop span if no span is found
	return &NoopSpan{tracer: t}
}

// ContextWithSpan returns a context with the span (simplified)
func (t *Tracer) ContextWithSpan(ctx context.Context, span tracer.Span) context.Context {
	if ddSpan, ok := span.(*Span); ok {
		// Store the span in the context
		return context.WithValue(ctx, "datadog-span", ddSpan)
	}
	return ctx
}

// Close shuts down the tracer
func (t *Tracer) Close() error {
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

// Span implements tracer.Span for Datadog with enhanced functionality
type Span struct {
	mu        sync.RWMutex // Mutex for thread-safety
	tracer    *Tracer
	startTime time.Time
	endTime   time.Time
	name      string
	attrs     map[string]interface{}
	status    tracer.StatusCode
	recording bool
}

// Context returns the span context
func (s *Span) Context() tracer.SpanContext {
	// Generate a mock trace and span ID for testing
	return tracer.SpanContext{
		TraceID:  fmt.Sprintf("datadog-trace-%d", time.Now().UnixNano()),
		SpanID:   fmt.Sprintf("datadog-span-%d", time.Now().UnixNano()),
		ParentID: "",
		Flags:    tracer.TraceFlagsSampled,
	}
}

// SetName sets the span operation name
func (s *Span) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.name = name
}

// SetAttributes sets multiple attributes
func (s *Span) SetAttributes(attributes map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range attributes {
		s.attrs[key] = value
	}
}

// SetAttribute sets a single attribute
func (s *Span) SetAttribute(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attrs[key] = value
}

// AddEvent adds a structured event (simulated with attributes)
func (s *Span) AddEvent(name string, attributes map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	eventPrefix := fmt.Sprintf("event.%s", name)
	s.attrs[eventPrefix+".timestamp"] = time.Now().Unix()
	s.attrs[eventPrefix+".name"] = name

	for key, value := range attributes {
		s.attrs[fmt.Sprintf("%s.%s", eventPrefix, key)] = value
	}
}

// SetStatus sets the span status
func (s *Span) SetStatus(code tracer.StatusCode, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = code
	switch code {
	case tracer.StatusCodeError:
		s.attrs["error"] = true
		s.attrs["error.status"] = "error"
		if message != "" {
			s.attrs["error.message"] = message
		}
	case tracer.StatusCodeOk:
		s.attrs["error"] = false
		s.attrs["error.status"] = "ok"
	case tracer.StatusCodeUnset:
		s.attrs["error.status"] = "unset"
	}
}

// RecordError records an error with comprehensive attributes
func (s *Span) RecordError(err error, attributes map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attrs["error"] = true
	s.attrs["error.message"] = err.Error()
	s.attrs["error.type"] = fmt.Sprintf("%T", err)
	s.attrs["error.timestamp"] = time.Now().Unix()

	// Add error attributes with prefix
	for key, value := range attributes {
		s.attrs[fmt.Sprintf("error.%s", key)] = value
	}

	// Set status to error directly (avoiding recursive lock)
	s.status = tracer.StatusCodeError
	s.attrs["error"] = true
	s.attrs["error.status"] = "error"
	if err.Error() != "" {
		s.attrs["error.message"] = err.Error()
	}
}

// End finishes the span
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endTime = time.Now()
	s.recording = false

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
	return s.recording
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

// NoopSpan is a no-op span implementation for Datadog
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
	case tracer.SpanKindInternal:
		return "custom"
	default:
		return "custom"
	}
}
