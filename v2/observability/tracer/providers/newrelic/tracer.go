package newrelic

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// Tracer implements tracer.Tracer for New Relic with enhanced features
type Tracer struct {
	name        string
	application *newrelic.Application
	provider    *Provider
	config      *tracer.TracerConfig
	metrics     tracer.TracerMetrics
	started     time.Time
}

// StartSpan creates a new New Relic span with comprehensive configuration
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...tracer.SpanOption) (context.Context, tracer.Span) {
	// Parse span options
	config := tracer.ApplySpanOptions(opts)

	// Get or create New Relic transaction
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		// Create new transaction if none exists
		txn = t.application.StartTransaction(name)
		ctx = newrelic.NewContext(ctx, txn)
	}

	// Create segment based on span kind
	var segment *newrelic.Segment
	switch config.Kind {
	case tracer.SpanKindServer:
		segment = txn.StartSegment(name)
	case tracer.SpanKindClient:
		externalSegment := newrelic.ExternalSegment{
			StartTime: newrelic.StartSegmentNow(txn),
			URL:       name,
		}
		segment = &newrelic.Segment{
			StartTime: externalSegment.StartTime,
			Name:      name,
		}
	case tracer.SpanKindProducer, tracer.SpanKindConsumer:
		segment = txn.StartSegment(name + "-message")
	default:
		segment = txn.StartSegment(name)
	}

	// Determine start time
	startTime := time.Now()
	if !config.StartTime.IsZero() {
		startTime = config.StartTime
	}

	// Create wrapped span
	span := &Span{
		segment:     segment,
		transaction: txn,
		tracer:      t,
		startTime:   startTime,
		name:        name,
		kind:        config.Kind,
	}

	// Add attributes
	for key, value := range config.Attributes {
		span.SetAttribute(key, value)
	}

	// Add custom attributes from provider config
	for key, value := range t.provider.config.CustomAttributes {
		span.SetAttribute(key, value)
	}

	// Update metrics
	atomic.AddInt64(&t.metrics.SpansCreated, 1)
	t.metrics.LastActivity = time.Now()

	return ctx, span
}

// SpanFromContext extracts a span from context
func (t *Tracer) SpanFromContext(ctx context.Context) tracer.Span {
	txn := newrelic.FromContext(ctx)
	if txn != nil {
		return &Span{
			transaction: txn,
			tracer:      t,
			startTime:   time.Now(), // Approximation
			name:        "context-span",
		}
	}
	return &NoopSpan{tracer: t}
}

// ContextWithSpan returns a context with the span
func (t *Tracer) ContextWithSpan(ctx context.Context, span tracer.Span) context.Context {
	if nrSpan, ok := span.(*Span); ok && nrSpan.transaction != nil {
		return newrelic.NewContext(ctx, nrSpan.transaction)
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

// Span implements tracer.Span for New Relic with enhanced functionality
type Span struct {
	mu          sync.RWMutex // Protects attributes map and other mutable fields
	segment     *newrelic.Segment
	transaction *newrelic.Transaction
	tracer      *Tracer
	startTime   time.Time
	endTime     time.Time
	name        string
	kind        tracer.SpanKind
	attributes  map[string]interface{}
}

// Context returns the span context
func (s *Span) Context() tracer.SpanContext {
	// New Relic doesn't expose trace/span IDs directly, so we generate placeholders
	// In a real implementation, you might extract these from headers or use distributed tracing
	return tracer.SpanContext{
		TraceID: fmt.Sprintf("nr-trace-%d", time.Now().UnixNano()),
		SpanID:  fmt.Sprintf("nr-span-%d", time.Now().UnixNano()),
		Flags:   tracer.TraceFlagsSampled,
	}
}

// SetName sets the span operation name
func (s *Span) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.name = name
	// New Relic segments don't support name changes after creation
	// This would typically require recreating the segment
}

// SetAttributes sets multiple attributes
func (s *Span) SetAttributes(attributes map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.attributes == nil {
		s.attributes = make(map[string]interface{})
	}

	for key, value := range attributes {
		s.attributes[key] = value
		if s.transaction != nil {
			s.transaction.AddAttribute(key, value)
		}
	}
}

// SetAttribute sets a single attribute
func (s *Span) SetAttribute(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.attributes == nil {
		s.attributes = make(map[string]interface{})
	}

	s.attributes[key] = value
	if s.transaction != nil {
		s.transaction.AddAttribute(key, value)
	}
}

// AddEvent adds a structured event
func (s *Span) AddEvent(name string, attributes map[string]interface{}) {
	if s.transaction != nil {
		// New Relic uses custom events
		eventData := map[string]interface{}{
			"eventType": "CustomSpanEvent",
			"eventName": name,
			"timestamp": time.Now().Unix(),
			"spanName":  s.name,
		}

		// Add event attributes
		for key, value := range attributes {
			eventData[key] = value
		}

		s.transaction.Application().RecordCustomEvent("SpanEvent", eventData)
	}
}

// SetStatus sets the span status
func (s *Span) SetStatus(code tracer.StatusCode, message string) {
	if s.transaction != nil {
		switch code {
		case tracer.StatusCodeError:
			s.transaction.NoticeError(fmt.Errorf("span error: %s", message))
			s.transaction.AddAttribute("error", true)
			s.transaction.AddAttribute("error.message", message)
		case tracer.StatusCodeOk:
			s.transaction.AddAttribute("error", false)
			s.transaction.AddAttribute("status", "ok")
		case tracer.StatusCodeUnset:
			s.transaction.AddAttribute("status", "unset")
		}
	}
}

// RecordError records an error with comprehensive attributes
func (s *Span) RecordError(err error, attributes map[string]interface{}) {
	if s.transaction != nil {
		// Record the error
		s.transaction.NoticeError(err)

		// Add error context
		s.transaction.AddAttribute("error", true)
		s.transaction.AddAttribute("error.message", err.Error())
		s.transaction.AddAttribute("error.type", fmt.Sprintf("%T", err))
		s.transaction.AddAttribute("error.timestamp", time.Now().Unix())

		// Add error attributes
		for key, value := range attributes {
			s.transaction.AddAttribute(fmt.Sprintf("error.%s", key), value)
		}
	}

	// Set status to error
	s.SetStatus(tracer.StatusCodeError, err.Error())
}

// End finishes the span
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.endTime = time.Now()

	if s.segment != nil {
		s.segment.End()
	}

	// For transactions, we don't end them here as they might contain multiple spans
	// In a real implementation, you'd need span hierarchy management

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
	return s.endTime.IsZero() && s.transaction != nil
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

// NoopSpan is a no-op span implementation for New Relic
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
