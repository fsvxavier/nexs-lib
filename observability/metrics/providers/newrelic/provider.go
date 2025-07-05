// Package newrelic provides NewRelic metrics implementation
package newrelic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/metrics"
)

// Provider implements metrics.Provider for NewRelic
type Provider struct {
	config  *Config
	client  NewRelicClient
	metrics map[string]interface{}
	mutex   sync.RWMutex
}

// Config holds NewRelic-specific configuration
type Config struct {
	LicenseKey  string
	AppName     string
	Environment string
	Host        string
	Namespace   string
	Attributes  map[string]interface{}
}

// NewRelicClient interface for NewRelic operations (for mocking)
type NewRelicClient interface {
	RecordMetric(name string, value float64, attributes map[string]interface{}) error
	RecordCustomEvent(eventType string, attributes map[string]interface{}) error
	Shutdown(timeout time.Duration) error
}

// MockNewRelicClient is a mock implementation for testing
type MockNewRelicClient struct {
	metrics []MetricData
	events  []EventData
	mutex   sync.RWMutex
}

type MetricData struct {
	Name       string
	Value      float64
	Attributes map[string]interface{}
	Timestamp  time.Time
}

type EventData struct {
	EventType  string
	Attributes map[string]interface{}
	Timestamp  time.Time
}

func NewMockNewRelicClient() *MockNewRelicClient {
	return &MockNewRelicClient{
		metrics: make([]MetricData, 0),
		events:  make([]EventData, 0),
	}
}

func (m *MockNewRelicClient) RecordMetric(name string, value float64, attributes map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics = append(m.metrics, MetricData{
		Name:       name,
		Value:      value,
		Attributes: attributes,
		Timestamp:  time.Now(),
	})
	return nil
}

func (m *MockNewRelicClient) RecordCustomEvent(eventType string, attributes map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.events = append(m.events, EventData{
		EventType:  eventType,
		Attributes: attributes,
		Timestamp:  time.Now(),
	})
	return nil
}

func (m *MockNewRelicClient) Shutdown(timeout time.Duration) error {
	return nil
}

func (m *MockNewRelicClient) GetMetrics() []MetricData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]MetricData, len(m.metrics))
	copy(result, m.metrics)
	return result
}

func (m *MockNewRelicClient) GetEvents() []EventData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]EventData, len(m.events))
	copy(result, m.events)
	return result
}

func (m *MockNewRelicClient) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.metrics = m.metrics[:0]
	m.events = m.events[:0]
}

// NewProvider creates a new NewRelic metrics provider
func NewProvider(config *Config, client NewRelicClient) *Provider {
	if config == nil {
		config = &Config{
			AppName:     "unknown-app",
			Environment: "development",
			Attributes:  make(map[string]interface{}),
		}
	}

	if client == nil {
		client = NewMockNewRelicClient()
	}

	return &Provider{
		config:  config,
		client:  client,
		metrics: make(map[string]interface{}),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "newrelic"
}

// CreateCounter creates a new NewRelic counter
func (p *Provider) CreateCounter(opts metrics.CounterOptions) (metrics.Counter, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	key := p.metricKey("counter", opts.Name, opts.Namespace)
	if existing, exists := p.metrics[key]; exists {
		if counter, ok := existing.(*Counter); ok {
			return counter, nil
		}
		return nil, fmt.Errorf("metric %s already exists with different type", key)
	}

	counter := &Counter{
		name:            p.metricName(opts.Name, opts.Namespace),
		help:            opts.Help,
		labels:          opts.Labels,
		constAttributes: p.mergeAttributes(opts.Tags),
		client:          p.client,
		values:          make(map[string]float64),
	}

	p.metrics[key] = counter
	return counter, nil
}

// CreateHistogram creates a new NewRelic histogram
func (p *Provider) CreateHistogram(opts metrics.HistogramOptions) (metrics.Histogram, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	key := p.metricKey("histogram", opts.Name, opts.Namespace)
	if existing, exists := p.metrics[key]; exists {
		if histogram, ok := existing.(*Histogram); ok {
			return histogram, nil
		}
		return nil, fmt.Errorf("metric %s already exists with different type", key)
	}

	histogram := &Histogram{
		name:            p.metricName(opts.Name, opts.Namespace),
		help:            opts.Help,
		labels:          opts.Labels,
		constAttributes: p.mergeAttributes(opts.Tags),
		client:          p.client,
		buckets:         opts.Buckets,
	}

	p.metrics[key] = histogram
	return histogram, nil
}

// CreateGauge creates a new NewRelic gauge
func (p *Provider) CreateGauge(opts metrics.GaugeOptions) (metrics.Gauge, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	key := p.metricKey("gauge", opts.Name, opts.Namespace)
	if existing, exists := p.metrics[key]; exists {
		if gauge, ok := existing.(*Gauge); ok {
			return gauge, nil
		}
		return nil, fmt.Errorf("metric %s already exists with different type", key)
	}

	gauge := &Gauge{
		name:            p.metricName(opts.Name, opts.Namespace),
		help:            opts.Help,
		labels:          opts.Labels,
		constAttributes: p.mergeAttributes(opts.Tags),
		client:          p.client,
		values:          make(map[string]float64),
	}

	p.metrics[key] = gauge
	return gauge, nil
}

// CreateSummary creates a new NewRelic summary
func (p *Provider) CreateSummary(opts metrics.SummaryOptions) (metrics.Summary, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	key := p.metricKey("summary", opts.Name, opts.Namespace)
	if existing, exists := p.metrics[key]; exists {
		if summary, ok := existing.(*Summary); ok {
			return summary, nil
		}
		return nil, fmt.Errorf("metric %s already exists with different type", key)
	}

	summary := &Summary{
		name:            p.metricName(opts.Name, opts.Namespace),
		help:            opts.Help,
		labels:          opts.Labels,
		constAttributes: p.mergeAttributes(opts.Tags),
		client:          p.client,
		objectives:      opts.Objectives,
	}

	p.metrics[key] = summary
	return summary, nil
}

// Shutdown gracefully shuts down the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.client.Shutdown(5 * time.Second); err != nil {
		return err
	}

	p.metrics = make(map[string]interface{})
	return nil
}

// GetRegistry returns the NewRelic client
func (p *Provider) GetRegistry() interface{} {
	return p.client
}

// Helper methods

func (p *Provider) metricName(name, namespace string) string {
	if namespace != "" {
		return fmt.Sprintf("%s/%s/%s", p.config.Namespace, namespace, name)
	}
	if p.config.Namespace != "" {
		return fmt.Sprintf("%s/%s", p.config.Namespace, name)
	}
	return name
}

func (p *Provider) mergeAttributes(labels map[string]string) map[string]interface{} {
	attrs := make(map[string]interface{})

	// Add config attributes first
	for k, v := range p.config.Attributes {
		attrs[k] = v
	}

	// Add metric-specific labels
	for k, v := range labels {
		attrs[k] = v
	}

	return attrs
}

func (p *Provider) metricKey(metricType, name, namespace string) string {
	return fmt.Sprintf("%s_%s_%s", metricType, namespace, name)
}

// Counter wraps a NewRelic counter
type Counter struct {
	name            string
	help            string
	labels          []string
	constAttributes map[string]interface{}
	client          NewRelicClient
	values          map[string]float64
	mutex           sync.RWMutex
}

func (c *Counter) Inc(labels ...string) {
	c.Add(1, labels...)
}

func (c *Counter) Add(value float64, labels ...string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	attrs := c.buildAttributes(labels...)
	key := c.buildKey(labels...)

	c.values[key] += value
	attrs["value"] = c.values[key]
	attrs["metric_type"] = "counter"

	c.client.RecordMetric(c.name, c.values[key], attrs)
}

func (c *Counter) Get(labels ...string) float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := c.buildKey(labels...)
	return c.values[key]
}

func (c *Counter) Reset(labels ...string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.buildKey(labels...)
	delete(c.values, key)
}

func (c *Counter) buildAttributes(labels ...string) map[string]interface{} {
	attrs := make(map[string]interface{})

	// Add const attributes
	for k, v := range c.constAttributes {
		attrs[k] = v
	}

	// Add dynamic labels
	for i := 0; i < len(labels) && i < len(c.labels); i++ {
		attrs[c.labels[i]] = labels[i]
	}

	return attrs
}

func (c *Counter) buildKey(labels ...string) string {
	if len(labels) == 0 {
		return "default"
	}
	return fmt.Sprintf("%v", labels)
}

// Histogram wraps a NewRelic histogram
type Histogram struct {
	name            string
	help            string
	labels          []string
	constAttributes map[string]interface{}
	client          NewRelicClient
	buckets         []float64
}

func (h *Histogram) Observe(value float64, labels ...string) {
	attrs := h.buildAttributes(labels...)
	attrs["value"] = value
	attrs["metric_type"] = "histogram"

	h.client.RecordMetric(h.name, value, attrs)
}

func (h *Histogram) ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string) {
	attrs := h.buildAttributes(labels...)
	attrs["value"] = value
	attrs["metric_type"] = "histogram"
	attrs["timestamp"] = timestamp.Unix()

	h.client.RecordMetric(h.name, value, attrs)
}

func (h *Histogram) Time(fn func(), labels ...string) {
	start := time.Now()
	fn()
	duration := time.Since(start).Seconds()
	h.Observe(duration, labels...)
}

func (h *Histogram) StartTimer(labels ...string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		h.Observe(duration, labels...)
	}
}

func (h *Histogram) GetCount(labels ...string) uint64 {
	// NewRelic doesn't provide direct access to histogram counts
	return 0
}

func (h *Histogram) GetSum(labels ...string) float64 {
	// NewRelic doesn't provide direct access to histogram sums
	return 0
}

func (h *Histogram) buildAttributes(labels ...string) map[string]interface{} {
	attrs := make(map[string]interface{})

	// Add const attributes
	for k, v := range h.constAttributes {
		attrs[k] = v
	}

	// Add dynamic labels
	for i := 0; i < len(labels) && i < len(h.labels); i++ {
		attrs[h.labels[i]] = labels[i]
	}

	return attrs
}

// Gauge wraps a NewRelic gauge
type Gauge struct {
	name            string
	help            string
	labels          []string
	constAttributes map[string]interface{}
	client          NewRelicClient
	values          map[string]float64
	mutex           sync.RWMutex
}

func (g *Gauge) Set(value float64, labels ...string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	attrs := g.buildAttributes(labels...)
	key := g.buildKey(labels...)

	g.values[key] = value
	attrs["value"] = value
	attrs["metric_type"] = "gauge"

	g.client.RecordMetric(g.name, value, attrs)
}

func (g *Gauge) Inc(labels ...string) {
	g.Add(1, labels...)
}

func (g *Gauge) Dec(labels ...string) {
	g.Sub(1, labels...)
}

func (g *Gauge) Add(value float64, labels ...string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	key := g.buildKey(labels...)
	g.values[key] += value

	attrs := g.buildAttributes(labels...)
	attrs["value"] = g.values[key]
	attrs["metric_type"] = "gauge"

	g.client.RecordMetric(g.name, g.values[key], attrs)
}

func (g *Gauge) Sub(value float64, labels ...string) {
	g.Add(-value, labels...)
}

func (g *Gauge) Get(labels ...string) float64 {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	key := g.buildKey(labels...)
	return g.values[key]
}

func (g *Gauge) SetToCurrentTime(labels ...string) {
	g.Set(float64(time.Now().Unix()), labels...)
}

func (g *Gauge) buildAttributes(labels ...string) map[string]interface{} {
	attrs := make(map[string]interface{})

	// Add const attributes
	for k, v := range g.constAttributes {
		attrs[k] = v
	}

	// Add dynamic labels
	for i := 0; i < len(labels) && i < len(g.labels); i++ {
		attrs[g.labels[i]] = labels[i]
	}

	return attrs
}

func (g *Gauge) buildKey(labels ...string) string {
	if len(labels) == 0 {
		return "default"
	}
	return fmt.Sprintf("%v", labels)
}

// Summary wraps a NewRelic summary
type Summary struct {
	name            string
	help            string
	labels          []string
	constAttributes map[string]interface{}
	client          NewRelicClient
	objectives      map[float64]float64
}

func (s *Summary) Observe(value float64, labels ...string) {
	attrs := s.buildAttributes(labels...)
	attrs["value"] = value
	attrs["metric_type"] = "summary"

	s.client.RecordMetric(s.name, value, attrs)
}

func (s *Summary) ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string) {
	attrs := s.buildAttributes(labels...)
	attrs["value"] = value
	attrs["metric_type"] = "summary"
	attrs["timestamp"] = timestamp.Unix()

	s.client.RecordMetric(s.name, value, attrs)
}

func (s *Summary) Time(fn func(), labels ...string) {
	start := time.Now()
	fn()
	duration := time.Since(start).Seconds()
	s.Observe(duration, labels...)
}

func (s *Summary) StartTimer(labels ...string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		s.Observe(duration, labels...)
	}
}

func (s *Summary) GetCount(labels ...string) uint64 {
	// NewRelic doesn't provide direct access to summary counts
	return 0
}

func (s *Summary) GetSum(labels ...string) float64 {
	// NewRelic doesn't provide direct access to summary sums
	return 0
}

func (s *Summary) GetQuantile(quantile float64, labels ...string) float64 {
	// NewRelic doesn't provide direct access to quantiles
	return 0
}

func (s *Summary) buildAttributes(labels ...string) map[string]interface{} {
	attrs := make(map[string]interface{})

	// Add const attributes
	for k, v := range s.constAttributes {
		attrs[k] = v
	}

	// Add dynamic labels
	for i := 0; i < len(labels) && i < len(s.labels); i++ {
		attrs[s.labels[i]] = labels[i]
	}

	return attrs
}
