package tracer

import (
	"time"
)

// spanOptionFunc implements SpanOption interface
type spanOptionFunc func(*SpanConfig)

func (f spanOptionFunc) apply(config *SpanConfig) {
	f(config)
}

// WithSpanKind sets the span kind for the operation
func WithSpanKind(kind SpanKind) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		config.Kind = kind
	})
}

// WithStartTime sets a custom start time for the span
func WithStartTime(t time.Time) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		config.StartTime = t
	})
}

// WithSpanAttributes sets initial attributes for the span
func WithSpanAttributes(attrs map[string]interface{}) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		if config.Attributes == nil {
			config.Attributes = make(map[string]interface{})
		}
		for k, v := range attrs {
			config.Attributes[k] = v
		}
	})
}

// WithParentSpan sets the parent span context
func WithParentSpan(parent SpanContext) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		config.Parent = &parent
	})
}

// WithSpanLinks adds links to other spans
func WithSpanLinks(links ...SpanLink) SpanOption {
	return spanOptionFunc(func(config *SpanConfig) {
		config.Links = append(config.Links, links...)
	})
}

// tracerOptionFunc implements TracerOption interface
type tracerOptionFunc func(*TracerConfig)

func (f tracerOptionFunc) apply(config *TracerConfig) {
	f(config)
}

// WithServiceName sets the service name for the tracer
func WithServiceName(name string) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		config.ServiceName = name
	})
}

// WithServiceVersion sets the service version for the tracer
func WithServiceVersion(version string) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		config.ServiceVersion = version
	})
}

// WithEnvironment sets the environment for the tracer
func WithEnvironment(env string) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		config.Environment = env
	})
}

// WithTracerAttributes sets tracer-level attributes
func WithTracerAttributes(attrs map[string]interface{}) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		if config.Attributes == nil {
			config.Attributes = make(map[string]interface{})
		}
		for k, v := range attrs {
			config.Attributes[k] = v
		}
	})
}

// WithSamplingRate sets the sampling rate (0.0 to 1.0)
func WithSamplingRate(rate float64) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		if rate >= 0.0 && rate <= 1.0 {
			config.SamplingRate = rate
		}
	})
}

// WithProfiling enables or disables profiling
func WithProfiling(enabled bool) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		config.EnableProfiling = enabled
	})
}

// WithBatchSize sets the batch size for span processing
func WithBatchSize(size int) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		if size > 0 {
			config.BatchSize = size
		}
	})
}

// WithFlushInterval sets the flush interval for pending spans
func WithFlushInterval(interval time.Duration) TracerOption {
	return tracerOptionFunc(func(config *TracerConfig) {
		if interval > 0 {
			config.FlushInterval = interval
		}
	})
}

// defaultSpanConfig returns a default span configuration
func defaultSpanConfig() *SpanConfig {
	return &SpanConfig{
		Kind:       SpanKindInternal,
		StartTime:  time.Now(),
		Attributes: make(map[string]interface{}),
		Links:      make([]SpanLink, 0),
	}
}

// defaultTracerConfig returns a default tracer configuration
func defaultTracerConfig() *TracerConfig {
	return &TracerConfig{
		ServiceName:     "unknown-service",
		ServiceVersion:  "1.0.0",
		Environment:     "development",
		Attributes:      make(map[string]interface{}),
		SamplingRate:    1.0,
		EnableProfiling: false,
		BatchSize:       100,
		FlushInterval:   5 * time.Second,
	}
}

// ApplySpanOptions applies span options to configuration
func ApplySpanOptions(opts []SpanOption) *SpanConfig {
	config := defaultSpanConfig()
	for _, opt := range opts {
		opt.apply(config)
	}
	return config
}

// applySpanOptions applies span options to configuration (internal)
func applySpanOptions(opts []SpanOption) *SpanConfig {
	return ApplySpanOptions(opts)
}

// ApplyTracerOptions applies tracer options to configuration
func ApplyTracerOptions(opts []TracerOption) *TracerConfig {
	config := defaultTracerConfig()
	for _, opt := range opts {
		opt.apply(config)
	}
	return config
}

// applyTracerOptions applies tracer options to configuration (internal)
func applyTracerOptions(opts []TracerOption) *TracerConfig {
	return ApplyTracerOptions(opts)
}
