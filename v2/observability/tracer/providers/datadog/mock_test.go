// Package datadog provides a no-op implementation for testing purposes
// This is a simplified version that doesn't depend on the problematic Datadog library
//go:build test
// +build test

package datadog

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

// MockProvider is a test-only implementation
type MockProvider struct {
	config       *Config
	tracers      map[string]*MockTracer
	mu           sync.RWMutex
	started      bool
	metrics      tracer.ProviderMetrics
	healthStatus string
	lastError    error
}

// MockTracer is a test-only implementation
type MockTracer struct {
	name     string
	provider *MockProvider
	config   *tracer.TracerConfig
	metrics  tracer.TracerMetrics
	started  time.Time
}

// MockSpan is a test-only implementation
type MockSpan struct {
	tracer    *MockTracer
	startTime time.Time
	endTime   time.Time
	name      string
	attrs     map[string]interface{}
	status    tracer.StatusCode
	recording bool
}

// NewProvider creates a mock provider for testing
func NewProvider(config *Config) (*MockProvider, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	provider := &MockProvider{
		config:  config,
		tracers: make(map[string]*MockTracer),
		metrics: tracer.ProviderMetrics{
			ConnectionState: "disconnected",
			LastFlush:       time.Now(),
		},
		healthStatus: "initializing",
	}

	return provider, nil
}

// Name returns the provider name
func (p *MockProvider) Name() string {
	return "datadog"
}

// CreateTracer creates a mock tracer
func (p *MockProvider) CreateTracer(name string, options ...tracer.TracerOption) (tracer.Tracer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	config := tracer.ApplyTracerOptions(options)

	if !p.started {
		p.started = true
		p.healthStatus = "connected"
		p.metrics.ConnectionState = "connected"
	}

	if t, exists := p.tracers[name]; exists {
		return t, nil
	}

	mockTracer := &MockTracer{
		name:     name,
		provider: p,
		config:   config,
		metrics: tracer.TracerMetrics{
			LastActivity: time.Now(),
		},
		started: time.Now(),
	}

	p.tracers[name] = mockTracer
	p.metrics.TracersActive = len(p.tracers)

	return mockTracer, nil
}

// Shutdown shuts down the mock provider
func (p *MockProvider) Shutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout")
	default:
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, t := range p.tracers {
		_ = t.Close()
	}

	p.started = false
	p.healthStatus = "disconnected"
	p.metrics.ConnectionState = "disconnected"
	p.tracers = make(map[string]*MockTracer)
	p.metrics.TracersActive = 0

	return nil
}

// HealthCheck performs a mock health check
func (p *MockProvider) HealthCheck(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return fmt.Errorf("provider not started")
	}

	if p.lastError != nil {
		return fmt.Errorf("last error: %w", p.lastError)
	}

	if p.metrics.ConnectionState != "connected" {
		return fmt.Errorf("agent connection state: %s", p.metrics.ConnectionState)
	}

	return nil
}

// GetProviderMetrics returns mock provider metrics
func (p *MockProvider) GetProviderMetrics() tracer.ProviderMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.metrics
}

// MockTracer methods
func (t *MockTracer) StartSpan(ctx context.Context, name string, opts ...tracer.SpanOption) (context.Context, tracer.Span) {
	config := tracer.ApplySpanOptions(opts)

	span := &MockSpan{
		tracer:    t,
		startTime: time.Now(),
		name:      name,
		attrs:     make(map[string]interface{}),
		recording: true,
	}

	if !config.StartTime.IsZero() {
		span.startTime = config.StartTime
	}

	for k, v := range config.Attributes {
		span.attrs[k] = v
	}

	// Update metrics
	t.metrics.LastActivity = time.Now()

	return ctx, span
}

func (t *MockTracer) SpanFromContext(ctx context.Context) tracer.Span {
	return &NoopSpan{tracer: t}
}

func (t *MockTracer) ContextWithSpan(ctx context.Context, span tracer.Span) context.Context {
	return ctx
}

func (t *MockTracer) Close() error {
	t.metrics.LastActivity = time.Now()
	return nil
}

func (t *MockTracer) GetMetrics() tracer.TracerMetrics {
	return t.metrics
}

// MockSpan methods
func (s *MockSpan) Context() tracer.SpanContext {
	return tracer.SpanContext{
		TraceID: "mock-trace-id",
		SpanID:  "mock-span-id",
		Flags:   tracer.TraceFlagsSampled,
	}
}

func (s *MockSpan) SetName(name string) {
	s.name = name
}

func (s *MockSpan) SetAttributes(attributes map[string]interface{}) {
	for k, v := range attributes {
		s.attrs[k] = v
	}
}

func (s *MockSpan) SetAttribute(key string, value interface{}) {
	s.attrs[key] = value
}

func (s *MockSpan) AddEvent(name string, attributes map[string]interface{}) {
	s.attrs["event."+name] = attributes
}

func (s *MockSpan) SetStatus(code tracer.StatusCode, message string) {
	s.status = code
	if message != "" {
		s.attrs["status.message"] = message
	}
}

func (s *MockSpan) RecordError(err error, attributes map[string]interface{}) {
	s.status = tracer.StatusCodeError
	s.attrs["error"] = err.Error()
	for k, v := range attributes {
		s.attrs["error."+k] = v
	}
}

func (s *MockSpan) End() {
	s.endTime = time.Now()
	s.recording = false
	if s.tracer != nil {
		s.tracer.metrics.LastActivity = time.Now()
	}
}

func (s *MockSpan) IsRecording() bool {
	return s.recording
}

func (s *MockSpan) GetDuration() time.Duration {
	if s.endTime.IsZero() {
		return time.Since(s.startTime)
	}
	return s.endTime.Sub(s.startTime)
}

// NoopSpan implementation
type NoopSpan struct {
	tracer *MockTracer
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
