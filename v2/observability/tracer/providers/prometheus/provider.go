// Package prometheus provides Prometheus metrics-based tracing implementation
// with comprehensive observability and monitoring capabilities for production systems.
package prometheus

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Provider implements tracer.Provider for Prometheus metrics with enhanced reliability
type Provider struct {
	config       *Config
	tracers      map[string]*Tracer
	mu           sync.RWMutex
	metrics      tracer.ProviderMetrics
	registry     *prometheus.Registry
	healthStatus string
	lastError    error
	shutdownCh   chan struct{}

	// Prometheus metrics
	spanCounter  *prometheus.CounterVec
	spanDuration *prometheus.HistogramVec
	errorCounter *prometheus.CounterVec
	activeSpans  *prometheus.GaugeVec
}

// Config holds comprehensive Prometheus-specific configuration
type Config struct {
	// Core settings
	ServiceName    string `json:"service_name" validate:"required"`
	ServiceVersion string `json:"service_version"`
	Environment    string `json:"environment"`
	Namespace      string `json:"namespace"`
	Subsystem      string `json:"subsystem"`

	// Metrics configuration
	EnableDetailedMetrics bool              `json:"enable_detailed_metrics"`
	CustomLabels          map[string]string `json:"custom_labels"`
	BucketBoundaries      []float64         `json:"bucket_boundaries"`
	MaxCardinality        int               `json:"max_cardinality"`

	// Performance settings
	CollectionInterval time.Duration `json:"collection_interval"`
	RetentionPeriod    time.Duration `json:"retention_period"`
	BatchSize          int           `json:"batch_size"`

	// Registry settings
	UseGlobalRegistry bool                 `json:"use_global_registry"`
	Registry          *prometheus.Registry `json:"-"`

	// Custom settings
	CustomAttributes map[string]interface{} `json:"custom_attributes"`
}

// DefaultConfig returns a production-ready default configuration
func DefaultConfig() *Config {
	return &Config{
		ServiceName:           "unknown-service",
		ServiceVersion:        "1.0.0",
		Environment:           "production",
		Namespace:             "tracer",
		Subsystem:             "spans",
		EnableDetailedMetrics: true,
		CustomLabels:          make(map[string]string),
		BucketBoundaries:      prometheus.DefBuckets,
		MaxCardinality:        1000,
		CollectionInterval:    30 * time.Second,
		RetentionPeriod:       24 * time.Hour,
		BatchSize:             100,
		UseGlobalRegistry:     false,
		CustomAttributes:      make(map[string]interface{}),
	}
}

// NewProvider creates a new Prometheus provider with comprehensive metrics
func NewProvider(config *Config) (*Provider, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Setup registry
	var registry *prometheus.Registry
	if config.UseGlobalRegistry {
		registry = prometheus.DefaultRegisterer.(*prometheus.Registry)
	} else if config.Registry != nil {
		registry = config.Registry
	} else {
		registry = prometheus.NewRegistry()
	}

	provider := &Provider{
		config:     config,
		tracers:    make(map[string]*Tracer),
		registry:   registry,
		shutdownCh: make(chan struct{}),
		metrics: tracer.ProviderMetrics{
			ConnectionState: "connected", // Prometheus is always "connected"
			LastFlush:       time.Now(),
		},
		healthStatus: "healthy",
	}

	// Initialize Prometheus metrics
	if err := provider.initializeMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	return provider, nil
}

// initializeMetrics sets up all Prometheus metrics
func (p *Provider) initializeMetrics() error {
	labels := []string{"service_name", "service_version", "environment", "tracer_name", "span_name", "span_kind"}

	// Add custom labels
	for key := range p.config.CustomLabels {
		labels = append(labels, key)
	}

	// Span counter
	p.spanCounter = promauto.With(p.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.config.Namespace,
			Subsystem: p.config.Subsystem,
			Name:      "total",
			Help:      "Total number of spans created by tracer",
		},
		labels,
	)

	// Span duration histogram
	p.spanDuration = promauto.With(p.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: p.config.Namespace,
			Subsystem: p.config.Subsystem,
			Name:      "duration_seconds",
			Help:      "Duration of spans in seconds",
			Buckets:   p.config.BucketBoundaries,
		},
		labels,
	)

	// Error counter
	p.errorCounter = promauto.With(p.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.config.Namespace,
			Subsystem: p.config.Subsystem,
			Name:      "errors_total",
			Help:      "Total number of span errors",
		},
		append(labels, "error_type"),
	)

	// Active spans gauge
	p.activeSpans = promauto.With(p.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: p.config.Namespace,
			Subsystem: p.config.Subsystem,
			Name:      "active",
			Help:      "Number of currently active spans",
		},
		[]string{"service_name", "tracer_name"},
	)

	return nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "prometheus"
}

// CreateTracer creates a new Prometheus tracer with comprehensive metrics
func (p *Provider) CreateTracer(name string, options ...tracer.TracerOption) (tracer.Tracer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Parse tracer options
	config := tracer.ApplyTracerOptions(options)

	// Create and cache tracer instance
	if t, exists := p.tracers[name]; exists {
		return t, nil
	}

	promTracer := &Tracer{
		name:     name,
		provider: p,
		config:   config,
		metrics: tracer.TracerMetrics{
			LastActivity: time.Now(),
		},
		started:     time.Now(),
		activeSpans: make(map[string]*Span),
	}

	p.tracers[name] = promTracer
	p.metrics.TracersActive = len(p.tracers)

	return promTracer, nil
}

// Shutdown gracefully shuts down the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close all tracers
	for _, t := range p.tracers {
		if err := t.Close(); err != nil {
			p.lastError = err
		}
	}

	p.healthStatus = "shutdown"
	p.metrics.ConnectionState = "disconnected"
	p.tracers = make(map[string]*Tracer)
	p.metrics.TracersActive = 0

	// Signal shutdown
	close(p.shutdownCh)

	return p.lastError
}

// HealthCheck performs health check (always healthy for Prometheus)
func (p *Provider) HealthCheck(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.lastError != nil {
		return fmt.Errorf("last error: %w", p.lastError)
	}

	return nil // Prometheus provider is always healthy
}

// GetProviderMetrics returns comprehensive provider metrics
func (p *Provider) GetProviderMetrics() tracer.ProviderMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Update bytes sent (simulated metric)
	atomic.AddInt64(&p.metrics.BytesSent, 512)
	p.metrics.LastFlush = time.Now()

	return p.metrics
}

// GetRegistry returns the Prometheus registry
func (p *Provider) GetRegistry() *prometheus.Registry {
	return p.registry
}

// GetMetrics returns all Prometheus metrics
func (p *Provider) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"span_counter":  p.spanCounter,
		"span_duration": p.spanDuration,
		"error_counter": p.errorCounter,
		"active_spans":  p.activeSpans,
	}
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

// buildLabels builds Prometheus labels for metrics
func (p *Provider) buildLabels(tracerName, spanName string, kind tracer.SpanKind, config *tracer.TracerConfig) prometheus.Labels {
	labels := prometheus.Labels{
		"service_name":    p.getServiceName(config),
		"service_version": p.getServiceVersion(config),
		"environment":     p.getEnvironment(config),
		"tracer_name":     tracerName,
		"span_name":       spanName,
		"span_kind":       kind.String(),
	}

	// Add custom labels
	for key, value := range p.config.CustomLabels {
		labels[key] = value
	}

	return labels
}

// validateConfig validates the provider configuration
func validateConfig(config *Config) error {
	if config.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if config.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	if config.MaxCardinality <= 0 {
		return fmt.Errorf("max cardinality must be positive")
	}

	if config.CollectionInterval <= 0 {
		return fmt.Errorf("collection interval must be positive")
	}

	if len(config.BucketBoundaries) == 0 {
		return fmt.Errorf("bucket boundaries cannot be empty")
	}

	return nil
}
