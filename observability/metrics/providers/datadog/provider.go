// Package datadog provides DataDog metrics implementation
package datadog

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/metrics"
)

// Provider implements metrics.Provider for DataDog
type Provider struct {
	config  *Config
	client  DataDogClient
	metrics map[string]interface{}
	mutex   sync.RWMutex
}

// Config holds DataDog-specific configuration
type Config struct {
	APIKey      string
	AppKey      string
	Environment string
	Service     string
	Version     string
	Tags        []string
	Host        string
	Namespace   string
	FlushPeriod time.Duration
}

// DataDogClient interface for DataDog operations (for mocking)
type DataDogClient interface {
	SendMetric(name string, value float64, metricType string, tags []string, timestamp time.Time) error
	Flush() error
	Close() error
}

// MockDataDogClient is a mock implementation for testing
type MockDataDogClient struct {
	metrics []MetricData
	mutex   sync.RWMutex
}

type MetricData struct {
	Name      string
	Value     float64
	Type      string
	Tags      []string
	Timestamp time.Time
}

func NewMockDataDogClient() *MockDataDogClient {
	return &MockDataDogClient{
		metrics: make([]MetricData, 0),
	}
}

func (m *MockDataDogClient) SendMetric(name string, value float64, metricType string, tags []string, timestamp time.Time) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics = append(m.metrics, MetricData{
		Name:      name,
		Value:     value,
		Type:      metricType,
		Tags:      tags,
		Timestamp: timestamp,
	})
	return nil
}

func (m *MockDataDogClient) Flush() error {
	return nil
}

func (m *MockDataDogClient) Close() error {
	return nil
}

func (m *MockDataDogClient) GetMetrics() []MetricData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]MetricData, len(m.metrics))
	copy(result, m.metrics)
	return result
}

func (m *MockDataDogClient) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.metrics = m.metrics[:0]
}

// NewProvider creates a new DataDog metrics provider
func NewProvider(config *Config, client DataDogClient) *Provider {
	if config == nil {
		config = &Config{
			FlushPeriod: 10 * time.Second,
			Environment: "development",
		}
	}

	if client == nil {
		client = NewMockDataDogClient()
	}

	return &Provider{
		config:  config,
		client:  client,
		metrics: make(map[string]interface{}),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "datadog"
}

// CreateCounter creates a new DataDog counter
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
		name:      p.metricName(opts.Name, opts.Namespace),
		help:      opts.Help,
		labels:    opts.Labels,
		constTags: p.mergeTags(opts.Tags),
		client:    p.client,
		values:    make(map[string]float64),
	}

	p.metrics[key] = counter
	return counter, nil
}

// CreateHistogram creates a new DataDog histogram
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
		name:      p.metricName(opts.Name, opts.Namespace),
		help:      opts.Help,
		labels:    opts.Labels,
		constTags: p.mergeTags(opts.Tags),
		client:    p.client,
		buckets:   opts.Buckets,
	}

	p.metrics[key] = histogram
	return histogram, nil
}

// CreateGauge creates a new DataDog gauge
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
		name:      p.metricName(opts.Name, opts.Namespace),
		help:      opts.Help,
		labels:    opts.Labels,
		constTags: p.mergeTags(opts.Tags),
		client:    p.client,
		values:    make(map[string]float64),
	}

	p.metrics[key] = gauge
	return gauge, nil
}

// CreateSummary creates a new DataDog summary
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
		name:       p.metricName(opts.Name, opts.Namespace),
		help:       opts.Help,
		labels:     opts.Labels,
		constTags:  p.mergeTags(opts.Tags),
		client:     p.client,
		objectives: opts.Objectives,
	}

	p.metrics[key] = summary
	return summary, nil
}

// Shutdown gracefully shuts down the provider
func (p *Provider) Shutdown(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.client.Flush(); err != nil {
		return err
	}

	if err := p.client.Close(); err != nil {
		return err
	}

	p.metrics = make(map[string]interface{})
	return nil
}

// GetRegistry returns the DataDog client
func (p *Provider) GetRegistry() interface{} {
	return p.client
}

// Helper methods

func (p *Provider) metricName(name, namespace string) string {
	if namespace != "" {
		return fmt.Sprintf("%s.%s.%s", p.config.Namespace, namespace, name)
	}
	if p.config.Namespace != "" {
		return fmt.Sprintf("%s.%s", p.config.Namespace, name)
	}
	return name
}

func (p *Provider) mergeTags(labels map[string]string) []string {
	tags := make([]string, 0, len(p.config.Tags)+len(labels))

	// Add config tags first
	tags = append(tags, p.config.Tags...)

	// Add metric-specific labels
	for k, v := range labels {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	return tags
}

func (p *Provider) metricKey(metricType, name, namespace string) string {
	return fmt.Sprintf("%s_%s_%s", metricType, namespace, name)
}

// Counter wraps a DataDog counter
type Counter struct {
	name      string
	help      string
	labels    []string
	constTags []string
	client    DataDogClient
	values    map[string]float64
	mutex     sync.RWMutex
}

func (c *Counter) Inc(labels ...string) {
	c.Add(1, labels...)
}

func (c *Counter) Add(value float64, labels ...string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tags := c.buildTags(labels...)
	key := c.buildKey(labels...)

	c.values[key] += value
	c.client.SendMetric(c.name, value, "count", tags, time.Now())
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

func (c *Counter) buildTags(labels ...string) []string {
	tags := make([]string, len(c.constTags))
	copy(tags, c.constTags)

	for i := 0; i < len(labels) && i < len(c.labels); i++ {
		tags = append(tags, fmt.Sprintf("%s:%s", c.labels[i], labels[i]))
	}

	return tags
}

func (c *Counter) buildKey(labels ...string) string {
	if len(labels) == 0 {
		return "default"
	}
	return fmt.Sprintf("%v", labels)
}

// Histogram wraps a DataDog histogram
type Histogram struct {
	name      string
	help      string
	labels    []string
	constTags []string
	client    DataDogClient
	buckets   []float64
}

func (h *Histogram) Observe(value float64, labels ...string) {
	tags := h.buildTags(labels...)
	h.client.SendMetric(h.name, value, "histogram", tags, time.Now())
}

func (h *Histogram) ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string) {
	tags := h.buildTags(labels...)
	h.client.SendMetric(h.name, value, "histogram", tags, timestamp)
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
	// DataDog doesn't provide direct access to histogram counts
	return 0
}

func (h *Histogram) GetSum(labels ...string) float64 {
	// DataDog doesn't provide direct access to histogram sums
	return 0
}

func (h *Histogram) buildTags(labels ...string) []string {
	tags := make([]string, len(h.constTags))
	copy(tags, h.constTags)

	for i := 0; i < len(labels) && i < len(h.labels); i++ {
		tags = append(tags, fmt.Sprintf("%s:%s", h.labels[i], labels[i]))
	}

	return tags
}

// Gauge wraps a DataDog gauge
type Gauge struct {
	name      string
	help      string
	labels    []string
	constTags []string
	client    DataDogClient
	values    map[string]float64
	mutex     sync.RWMutex
}

func (g *Gauge) Set(value float64, labels ...string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	tags := g.buildTags(labels...)
	key := g.buildKey(labels...)

	g.values[key] = value
	g.client.SendMetric(g.name, value, "gauge", tags, time.Now())
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

	tags := g.buildTags(labels...)
	g.client.SendMetric(g.name, g.values[key], "gauge", tags, time.Now())
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

func (g *Gauge) buildTags(labels ...string) []string {
	tags := make([]string, len(g.constTags))
	copy(tags, g.constTags)

	for i := 0; i < len(labels) && i < len(g.labels); i++ {
		tags = append(tags, fmt.Sprintf("%s:%s", g.labels[i], labels[i]))
	}

	return tags
}

func (g *Gauge) buildKey(labels ...string) string {
	if len(labels) == 0 {
		return "default"
	}
	return fmt.Sprintf("%v", labels)
}

// Summary wraps a DataDog summary
type Summary struct {
	name       string
	help       string
	labels     []string
	constTags  []string
	client     DataDogClient
	objectives map[float64]float64
}

func (s *Summary) Observe(value float64, labels ...string) {
	tags := s.buildTags(labels...)
	s.client.SendMetric(s.name, value, "histogram", tags, time.Now())
}

func (s *Summary) ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string) {
	tags := s.buildTags(labels...)
	s.client.SendMetric(s.name, value, "histogram", tags, timestamp)
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
	// DataDog doesn't provide direct access to summary counts
	return 0
}

func (s *Summary) GetSum(labels ...string) float64 {
	// DataDog doesn't provide direct access to summary sums
	return 0
}

func (s *Summary) GetQuantile(quantile float64, labels ...string) float64 {
	// DataDog doesn't provide direct access to quantiles
	return 0
}

func (s *Summary) buildTags(labels ...string) []string {
	tags := make([]string, len(s.constTags))
	copy(tags, s.constTags)

	for i := 0; i < len(labels) && i < len(s.labels); i++ {
		tags = append(tags, fmt.Sprintf("%s:%s", s.labels[i], labels[i]))
	}

	return tags
}
