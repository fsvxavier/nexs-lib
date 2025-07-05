// Package prometheus provides a Prometheus implementation of the metrics interface
package prometheus

import (
	"context"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Provider implements the metrics.Provider interface for Prometheus
type Provider struct {
	registry  prometheus.Registerer
	gatherer  prometheus.Gatherer
	namespace string
	prefix    string
	mu        sync.RWMutex
	counters  map[string]*promCounter
	histos    map[string]*promHistogram
	gauges    map[string]*promGauge
	summaries map[string]*promSummary
	shutdown  bool
}

// NewProvider creates a new Prometheus metrics provider
func NewProvider(cfg metrics.PrometheusConfig) (*Provider, error) {
	var registry prometheus.Registerer
	var gatherer prometheus.Gatherer

	if cfg.Registry != nil {
		if reg, ok := cfg.Registry.(prometheus.Registerer); ok {
			registry = reg
		} else if reg, ok := cfg.Registry.(*prometheus.Registry); ok {
			registry = reg
			gatherer = reg
		} else {
			return nil, metrics.NewMetricsError(
				metrics.ErrTypeInvalidConfig,
				"invalid registry type, must be prometheus.Registerer or *prometheus.Registry",
				metrics.WithProvider("prometheus"),
			)
		}
	} else {
		registry = prometheus.DefaultRegisterer
		gatherer = prometheus.DefaultGatherer
	}

	return &Provider{
		registry:  registry,
		gatherer:  gatherer,
		prefix:    cfg.Prefix,
		counters:  make(map[string]*promCounter),
		histos:    make(map[string]*promHistogram),
		gauges:    make(map[string]*promGauge),
		summaries: make(map[string]*promSummary),
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "prometheus"
}

// CreateCounter creates a new counter metric
func (p *Provider) CreateCounter(opts metrics.CounterOptions) (metrics.Counter, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shutdown {
		return nil, metrics.NewMetricsError(
			metrics.ErrTypeMetricCreation,
			"provider is shutdown",
			metrics.WithProvider("prometheus"),
		)
	}

	name := p.buildMetricName(opts.Name, opts.Namespace, opts.Subsystem)
	if existing, found := p.counters[name]; found {
		return existing, nil
	}

	promOpts := prometheus.CounterOpts{
		Name:      name,
		Help:      opts.Help,
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
	}

	var counter prometheus.Counter
	var counterVec *prometheus.CounterVec

	if len(opts.Labels) > 0 {
		counterVec = promauto.With(p.registry).NewCounterVec(
			prometheus.CounterOpts(promOpts),
			opts.Labels,
		)
	} else {
		counter = promauto.With(p.registry).NewCounter(promOpts)
	}

	promCounter := &promCounter{
		counter:    counter,
		counterVec: counterVec,
		labels:     opts.Labels,
		name:       name,
	}

	p.counters[name] = promCounter
	return promCounter, nil
}

// CreateHistogram creates a new histogram metric
func (p *Provider) CreateHistogram(opts metrics.HistogramOptions) (metrics.Histogram, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shutdown {
		return nil, metrics.NewMetricsError(
			metrics.ErrTypeMetricCreation,
			"provider is shutdown",
			metrics.WithProvider("prometheus"),
		)
	}

	name := p.buildMetricName(opts.Name, opts.Namespace, opts.Subsystem)
	if existing, found := p.histos[name]; found {
		return existing, nil
	}

	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	promOpts := prometheus.HistogramOpts{
		Name:      name,
		Help:      opts.Help,
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Buckets:   buckets,
	}

	var histogram prometheus.Histogram
	var histogramVec *prometheus.HistogramVec

	if len(opts.Labels) > 0 {
		histogramVec = promauto.With(p.registry).NewHistogramVec(
			prometheus.HistogramOpts(promOpts),
			opts.Labels,
		)
	} else {
		histogram = promauto.With(p.registry).NewHistogram(promOpts)
	}

	promHistogram := &promHistogram{
		histogram:    histogram,
		histogramVec: histogramVec,
		labels:       opts.Labels,
		name:         name,
	}

	p.histos[name] = promHistogram
	return promHistogram, nil
}

// CreateGauge creates a new gauge metric
func (p *Provider) CreateGauge(opts metrics.GaugeOptions) (metrics.Gauge, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shutdown {
		return nil, metrics.NewMetricsError(
			metrics.ErrTypeMetricCreation,
			"provider is shutdown",
			metrics.WithProvider("prometheus"),
		)
	}

	name := p.buildMetricName(opts.Name, opts.Namespace, opts.Subsystem)
	if existing, found := p.gauges[name]; found {
		return existing, nil
	}

	promOpts := prometheus.GaugeOpts{
		Name:      name,
		Help:      opts.Help,
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
	}

	var gauge prometheus.Gauge
	var gaugeVec *prometheus.GaugeVec

	if len(opts.Labels) > 0 {
		gaugeVec = promauto.With(p.registry).NewGaugeVec(
			prometheus.GaugeOpts(promOpts),
			opts.Labels,
		)
	} else {
		gauge = promauto.With(p.registry).NewGauge(promOpts)
	}

	promGauge := &promGauge{
		gauge:    gauge,
		gaugeVec: gaugeVec,
		labels:   opts.Labels,
		name:     name,
	}

	p.gauges[name] = promGauge
	return promGauge, nil
}

// CreateSummary creates a new summary metric
func (p *Provider) CreateSummary(opts metrics.SummaryOptions) (metrics.Summary, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shutdown {
		return nil, metrics.NewMetricsError(
			metrics.ErrTypeMetricCreation,
			"provider is shutdown",
			metrics.WithProvider("prometheus"),
		)
	}

	name := p.buildMetricName(opts.Name, opts.Namespace, opts.Subsystem)
	if existing, found := p.summaries[name]; found {
		return existing, nil
	}

	objectives := opts.Objectives
	if len(objectives) == 0 {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	promOpts := prometheus.SummaryOpts{
		Name:       name,
		Help:       opts.Help,
		Namespace:  opts.Namespace,
		Subsystem:  opts.Subsystem,
		Objectives: objectives,
		MaxAge:     opts.MaxAge,
		AgeBuckets: opts.AgeBuckets,
		BufCap:     opts.BufCap,
	}

	var summary prometheus.Summary
	var summaryVec *prometheus.SummaryVec

	if len(opts.Labels) > 0 {
		summaryVec = promauto.With(p.registry).NewSummaryVec(
			prometheus.SummaryOpts(promOpts),
			opts.Labels,
		)
	} else {
		summary = promauto.With(p.registry).NewSummary(promOpts)
	}

	promSummary := &promSummary{
		summary:    summary,
		summaryVec: summaryVec,
		labels:     opts.Labels,
		name:       name,
	}

	p.summaries[name] = promSummary
	return promSummary, nil
}

// Shutdown gracefully shuts down the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.shutdown = true
	return nil
}

// GetRegistry returns the underlying Prometheus registry
func (p *Provider) GetRegistry() interface{} {
	return p.registry
}

// buildMetricName builds a metric name with optional prefix
func (p *Provider) buildMetricName(name, namespace, subsystem string) string {
	if p.prefix != "" {
		if namespace != "" {
			return p.prefix + "_" + namespace + "_" + name
		}
		return p.prefix + "_" + name
	}
	return name
}

// promCounter implements metrics.Counter using Prometheus
type promCounter struct {
	counter    prometheus.Counter
	counterVec *prometheus.CounterVec
	labels     []string
	name       string
}

func (c *promCounter) Inc(labels ...string) {
	if c.counterVec != nil {
		c.counterVec.WithLabelValues(labels...).Inc()
	} else {
		c.counter.Inc()
	}
}

func (c *promCounter) Add(value float64, labels ...string) {
	if c.counterVec != nil {
		c.counterVec.WithLabelValues(labels...).Add(value)
	} else {
		c.counter.Add(value)
	}
}

func (c *promCounter) Get(labels ...string) float64 {
	// Note: Prometheus doesn't provide a direct way to get counter values
	// This would typically require gathering metrics from the registry
	// For simplicity, we'll return 0 here
	// In a real implementation, you'd use the Gather() method from the registry
	return 0
}

func (c *promCounter) Reset(labels ...string) {
	// Prometheus counters cannot be reset, this is a no-op
	// In production, you might want to log this or handle differently
}

// promHistogram implements metrics.Histogram using Prometheus
type promHistogram struct {
	histogram    prometheus.Histogram
	histogramVec *prometheus.HistogramVec
	labels       []string
	name         string
}

func (h *promHistogram) Observe(value float64, labels ...string) {
	if h.histogramVec != nil {
		h.histogramVec.WithLabelValues(labels...).Observe(value)
	} else {
		h.histogram.Observe(value)
	}
}

func (h *promHistogram) ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string) {
	// Prometheus doesn't support timestamped observations in the client library
	// We'll just observe without timestamp
	h.Observe(value, labels...)
}

func (h *promHistogram) Time(fn func(), labels ...string) {
	start := time.Now()
	fn()
	duration := time.Since(start).Seconds()
	h.Observe(duration, labels...)
}

func (h *promHistogram) StartTimer(labels ...string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		h.Observe(duration, labels...)
	}
}

func (h *promHistogram) GetCount(labels ...string) uint64 {
	// Note: Prometheus doesn't provide a direct way to get histogram counts
	// This would typically require gathering metrics from the registry
	// For simplicity, we'll return 0 here
	return 0
}

func (h *promHistogram) GetSum(labels ...string) float64 {
	// Note: Prometheus doesn't provide a direct way to get histogram sums
	// This would typically require gathering metrics from the registry
	// For simplicity, we'll return 0 here
	return 0
}

// promGauge implements metrics.Gauge using Prometheus
type promGauge struct {
	gauge    prometheus.Gauge
	gaugeVec *prometheus.GaugeVec
	labels   []string
	name     string
}

func (g *promGauge) Set(value float64, labels ...string) {
	if g.gaugeVec != nil {
		g.gaugeVec.WithLabelValues(labels...).Set(value)
	} else {
		g.gauge.Set(value)
	}
}

func (g *promGauge) Inc(labels ...string) {
	if g.gaugeVec != nil {
		g.gaugeVec.WithLabelValues(labels...).Inc()
	} else {
		g.gauge.Inc()
	}
}

func (g *promGauge) Dec(labels ...string) {
	if g.gaugeVec != nil {
		g.gaugeVec.WithLabelValues(labels...).Dec()
	} else {
		g.gauge.Dec()
	}
}

func (g *promGauge) Add(value float64, labels ...string) {
	if g.gaugeVec != nil {
		g.gaugeVec.WithLabelValues(labels...).Add(value)
	} else {
		g.gauge.Add(value)
	}
}

func (g *promGauge) Sub(value float64, labels ...string) {
	if g.gaugeVec != nil {
		g.gaugeVec.WithLabelValues(labels...).Sub(value)
	} else {
		g.gauge.Sub(value)
	}
}

func (g *promGauge) Get(labels ...string) float64 {
	// Note: Prometheus doesn't provide a direct way to get gauge values
	// This would typically require gathering metrics from the registry
	// For simplicity, we'll return 0 here
	return 0
}

func (g *promGauge) SetToCurrentTime(labels ...string) {
	g.Set(float64(time.Now().Unix()), labels...)
}

// promSummary implements metrics.Summary using Prometheus
type promSummary struct {
	summary    prometheus.Summary
	summaryVec *prometheus.SummaryVec
	labels     []string
	name       string
}

func (s *promSummary) Observe(value float64, labels ...string) {
	if s.summaryVec != nil {
		s.summaryVec.WithLabelValues(labels...).Observe(value)
	} else {
		s.summary.Observe(value)
	}
}

func (s *promSummary) ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string) {
	// Prometheus doesn't support timestamped observations in the client library
	// We'll just observe without timestamp
	s.Observe(value, labels...)
}

func (s *promSummary) Time(fn func(), labels ...string) {
	start := time.Now()
	fn()
	duration := time.Since(start).Seconds()
	s.Observe(duration, labels...)
}

func (s *promSummary) StartTimer(labels ...string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		s.Observe(duration, labels...)
	}
}

func (s *promSummary) GetQuantile(quantile float64, labels ...string) float64 {
	// Note: Prometheus doesn't provide a direct way to get quantile values
	// This would typically require gathering metrics from the registry
	// For simplicity, we'll return 0 here
	return 0
}

func (s *promSummary) GetCount(labels ...string) uint64 {
	// Note: Prometheus doesn't provide a direct way to get summary counts
	// This would typically require gathering metrics from the registry
	// For simplicity, we'll return 0 here
	return 0
}

func (s *promSummary) GetSum(labels ...string) float64 {
	// Note: Prometheus doesn't provide a direct way to get summary sums
	// This would typically require gathering metrics from the registry
	// For simplicity, we'll return 0 here
	return 0
}
