// Package metrics provides a modern, extensible metrics abstraction for Go applications.
// This package supports multiple metrics backends including Prometheus, DataDog, and NewRelic.
package metrics

import (
	"context"
	"time"
)

// Provider represents a metrics provider that can create and manage metric collectors
type Provider interface {
	// Name returns the provider name
	Name() string

	// CreateCounter creates a new counter metric
	CreateCounter(opts CounterOptions) (Counter, error)

	// CreateHistogram creates a new histogram metric
	CreateHistogram(opts HistogramOptions) (Histogram, error)

	// CreateGauge creates a new gauge metric
	CreateGauge(opts GaugeOptions) (Gauge, error)

	// CreateSummary creates a new summary metric
	CreateSummary(opts SummaryOptions) (Summary, error)

	// Shutdown gracefully shuts down the provider
	Shutdown(ctx context.Context) error

	// GetRegistry returns the underlying registry (provider-specific)
	GetRegistry() interface{}
}

// Counter represents a monotonically increasing metric
type Counter interface {
	// Inc increments the counter by 1
	Inc(labels ...string)

	// Add adds the given value to the counter
	Add(value float64, labels ...string)

	// Get returns the current value of the counter
	Get(labels ...string) float64

	// Reset resets the counter to zero
	Reset(labels ...string)
}

// Histogram represents a metric that samples observations and counts them in buckets
type Histogram interface {
	// Observe records an observation
	Observe(value float64, labels ...string)

	// ObserveWithTimestamp records an observation with a specific timestamp
	ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string)

	// Time measures the time taken by a function
	Time(fn func(), labels ...string)

	// StartTimer returns a timer function
	StartTimer(labels ...string) func()

	// GetCount returns the total number of observations
	GetCount(labels ...string) uint64

	// GetSum returns the sum of all observations
	GetSum(labels ...string) float64
}

// Gauge represents a metric that can go up and down
type Gauge interface {
	// Set sets the gauge to a specific value
	Set(value float64, labels ...string)

	// Inc increments the gauge by 1
	Inc(labels ...string)

	// Dec decrements the gauge by 1
	Dec(labels ...string)

	// Add adds the given value to the gauge
	Add(value float64, labels ...string)

	// Sub subtracts the given value from the gauge
	Sub(value float64, labels ...string)

	// Get returns the current value of the gauge
	Get(labels ...string) float64

	// SetToCurrentTime sets the gauge to the current Unix time
	SetToCurrentTime(labels ...string)
}

// Summary represents a metric that samples observations and provides quantiles
type Summary interface {
	// Observe records an observation
	Observe(value float64, labels ...string)

	// ObserveWithTimestamp records an observation with a specific timestamp
	ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string)

	// Time measures the time taken by a function
	Time(fn func(), labels ...string)

	// StartTimer returns a timer function
	StartTimer(labels ...string) func()

	// GetQuantile returns the value at a specific quantile
	GetQuantile(quantile float64, labels ...string) float64

	// GetCount returns the total number of observations
	GetCount(labels ...string) uint64

	// GetSum returns the sum of all observations
	GetSum(labels ...string) float64
}

// MetricOptions contains common options for all metric types
type MetricOptions struct {
	Name      string            // Required: metric name
	Help      string            // Required: metric description
	Labels    []string          // Optional: label names
	Namespace string            // Optional: metric namespace
	Subsystem string            // Optional: metric subsystem
	Tags      map[string]string // Optional: additional tags
}

// CounterOptions represents options for creating a counter
type CounterOptions struct {
	MetricOptions
}

// HistogramOptions represents options for creating a histogram
type HistogramOptions struct {
	MetricOptions
	Buckets []float64 // Optional: custom buckets (provider-specific defaults if not set)
}

// GaugeOptions represents options for creating a gauge
type GaugeOptions struct {
	MetricOptions
}

// SummaryOptions represents options for creating a summary
type SummaryOptions struct {
	MetricOptions
	Objectives map[float64]float64 // Optional: quantile objectives (provider-specific defaults if not set)
	MaxAge     time.Duration       // Optional: max age of observations
	AgeBuckets uint32              // Optional: number of age buckets
	BufCap     uint32              // Optional: sample buffer capacity
}

// Config represents the configuration for a metrics provider
type Config struct {
	// Provider type ("prometheus", "datadog", "newrelic")
	Provider string `json:"provider" yaml:"provider"`

	// Common configuration
	Namespace string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Tags      map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`

	// Prometheus-specific configuration
	Prometheus PrometheusConfig `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`

	// DataDog-specific configuration
	DataDog DataDogConfig `json:"datadog,omitempty" yaml:"datadog,omitempty"`

	// NewRelic-specific configuration
	NewRelic NewRelicConfig `json:"newrelic,omitempty" yaml:"newrelic,omitempty"`
}

// PrometheusConfig contains Prometheus-specific configuration
type PrometheusConfig struct {
	Registry interface{} `json:"-" yaml:"-"` // Custom registry
	Prefix   string      `json:"prefix,omitempty" yaml:"prefix,omitempty"`
}

// DataDogConfig contains DataDog-specific configuration
type DataDogConfig struct {
	APIKey    string   `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	AppKey    string   `json:"app_key,omitempty" yaml:"app_key,omitempty"`
	Host      string   `json:"host,omitempty" yaml:"host,omitempty"`
	Service   string   `json:"service,omitempty" yaml:"service,omitempty"`
	Version   string   `json:"version,omitempty" yaml:"version,omitempty"`
	Env       string   `json:"env,omitempty" yaml:"env,omitempty"`
	Tags      []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	StatsdURL string   `json:"statsd_url,omitempty" yaml:"statsd_url,omitempty"`
}

// NewRelicConfig contains NewRelic-specific configuration
type NewRelicConfig struct {
	APIKey      string `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	AppName     string `json:"app_name,omitempty" yaml:"app_name,omitempty"`
	License     string `json:"license,omitempty" yaml:"license,omitempty"`
	Host        string `json:"host,omitempty" yaml:"host,omitempty"`
	MetricsURL  string `json:"metrics_url,omitempty" yaml:"metrics_url,omitempty"`
	EventsURL   string `json:"events_url,omitempty" yaml:"events_url,omitempty"`
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// ErrorType represents different types of metrics errors
type ErrorType string

const (
	// ErrTypeInvalidConfig indicates invalid configuration
	ErrTypeInvalidConfig ErrorType = "invalid_config"

	// ErrTypeProviderNotFound indicates the provider was not found
	ErrTypeProviderNotFound ErrorType = "provider_not_found"

	// ErrTypeMetricCreation indicates an error during metric creation
	ErrTypeMetricCreation ErrorType = "metric_creation"

	// ErrTypeMetricOperation indicates an error during metric operation
	ErrTypeMetricOperation ErrorType = "metric_operation"

	// ErrTypeProviderShutdown indicates an error during provider shutdown
	ErrTypeProviderShutdown ErrorType = "provider_shutdown"
)

// MetricsError represents a metrics-specific error
type MetricsError struct {
	Type     ErrorType `json:"type"`
	Message  string    `json:"message"`
	Provider string    `json:"provider,omitempty"`
	Metric   string    `json:"metric,omitempty"`
	Cause    error     `json:"-"`
}

// Error implements the error interface
func (e *MetricsError) Error() string {
	if e.Provider != "" {
		return e.Provider + ": " + e.Message
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *MetricsError) Unwrap() error {
	return e.Cause
}

// NewMetricsError creates a new metrics error
func NewMetricsError(errType ErrorType, message string, options ...func(*MetricsError)) *MetricsError {
	err := &MetricsError{
		Type:    errType,
		Message: message,
	}

	for _, opt := range options {
		opt(err)
	}

	return err
}

// WithProvider sets the provider for the error
func WithProvider(provider string) func(*MetricsError) {
	return func(e *MetricsError) {
		e.Provider = provider
	}
}

// WithMetric sets the metric name for the error
func WithMetric(metric string) func(*MetricsError) {
	return func(e *MetricsError) {
		e.Metric = metric
	}
}

// WithCause sets the underlying cause for the error
func WithCause(cause error) func(*MetricsError) {
	return func(e *MetricsError) {
		e.Cause = cause
	}
}
