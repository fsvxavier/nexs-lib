// Package mocks provides mock implementations for metrics providers
package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/metrics"
)

// MockProvider is a mock implementation of metrics.Provider
type MockProvider struct {
	name          string
	counters      map[string]*MockCounter
	histograms    map[string]*MockHistogram
	gauges        map[string]*MockGauge
	summaries     map[string]*MockSummary
	shutdownCalls int
	mutex         sync.RWMutex
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:       name,
		counters:   make(map[string]*MockCounter),
		histograms: make(map[string]*MockHistogram),
		gauges:     make(map[string]*MockGauge),
		summaries:  make(map[string]*MockSummary),
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) CreateCounter(opts metrics.CounterOptions) (metrics.Counter, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.metricKey("counter", opts.Name, opts.Namespace)
	if counter, exists := m.counters[key]; exists {
		return counter, nil
	}

	counter := &MockCounter{
		name:   opts.Name,
		help:   opts.Help,
		labels: opts.Labels,
		values: make(map[string]float64),
	}

	m.counters[key] = counter
	return counter, nil
}

func (m *MockProvider) CreateHistogram(opts metrics.HistogramOptions) (metrics.Histogram, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.metricKey("histogram", opts.Name, opts.Namespace)
	if histogram, exists := m.histograms[key]; exists {
		return histogram, nil
	}

	histogram := &MockHistogram{
		name:         opts.Name,
		help:         opts.Help,
		labels:       opts.Labels,
		buckets:      opts.Buckets,
		observations: make(map[string][]float64),
		counts:       make(map[string]uint64),
		sums:         make(map[string]float64),
	}

	m.histograms[key] = histogram
	return histogram, nil
}

func (m *MockProvider) CreateGauge(opts metrics.GaugeOptions) (metrics.Gauge, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.metricKey("gauge", opts.Name, opts.Namespace)
	if gauge, exists := m.gauges[key]; exists {
		return gauge, nil
	}

	gauge := &MockGauge{
		name:   opts.Name,
		help:   opts.Help,
		labels: opts.Labels,
		values: make(map[string]float64),
	}

	m.gauges[key] = gauge
	return gauge, nil
}

func (m *MockProvider) CreateSummary(opts metrics.SummaryOptions) (metrics.Summary, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.metricKey("summary", opts.Name, opts.Namespace)
	if summary, exists := m.summaries[key]; exists {
		return summary, nil
	}

	summary := &MockSummary{
		name:         opts.Name,
		help:         opts.Help,
		labels:       opts.Labels,
		objectives:   opts.Objectives,
		observations: make(map[string][]float64),
		counts:       make(map[string]uint64),
		sums:         make(map[string]float64),
	}

	m.summaries[key] = summary
	return summary, nil
}

func (m *MockProvider) Shutdown(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.shutdownCalls++
	return nil
}

func (m *MockProvider) GetRegistry() interface{} {
	return m
}

func (m *MockProvider) GetShutdownCalls() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.shutdownCalls
}

func (m *MockProvider) GetCounters() map[string]*MockCounter {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*MockCounter)
	for k, v := range m.counters {
		result[k] = v
	}
	return result
}

func (m *MockProvider) GetHistograms() map[string]*MockHistogram {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*MockHistogram)
	for k, v := range m.histograms {
		result[k] = v
	}
	return result
}

func (m *MockProvider) GetGauges() map[string]*MockGauge {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*MockGauge)
	for k, v := range m.gauges {
		result[k] = v
	}
	return result
}

func (m *MockProvider) GetSummaries() map[string]*MockSummary {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*MockSummary)
	for k, v := range m.summaries {
		result[k] = v
	}
	return result
}

func (m *MockProvider) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.counters = make(map[string]*MockCounter)
	m.histograms = make(map[string]*MockHistogram)
	m.gauges = make(map[string]*MockGauge)
	m.summaries = make(map[string]*MockSummary)
	m.shutdownCalls = 0
}

func (m *MockProvider) metricKey(metricType, name, namespace string) string {
	if namespace != "" {
		return metricType + "_" + namespace + "_" + name
	}
	return metricType + "_" + name
}

// MockCounter is a mock implementation of metrics.Counter
type MockCounter struct {
	name       string
	help       string
	labels     []string
	values     map[string]float64
	incCalls   map[string]int
	addCalls   map[string]int
	getCalls   map[string]int
	resetCalls map[string]int
	mutex      sync.RWMutex
}

func (c *MockCounter) Inc(labels ...string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.buildKey(labels...)
	c.values[key]++

	if c.incCalls == nil {
		c.incCalls = make(map[string]int)
	}
	c.incCalls[key]++
}

func (c *MockCounter) Add(value float64, labels ...string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.buildKey(labels...)
	c.values[key] += value

	if c.addCalls == nil {
		c.addCalls = make(map[string]int)
	}
	c.addCalls[key]++
}

func (c *MockCounter) Get(labels ...string) float64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.buildKey(labels...)

	if c.getCalls == nil {
		c.getCalls = make(map[string]int)
	}
	c.getCalls[key]++

	return c.values[key]
}

func (c *MockCounter) Reset(labels ...string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.buildKey(labels...)
	delete(c.values, key)

	if c.resetCalls == nil {
		c.resetCalls = make(map[string]int)
	}
	c.resetCalls[key]++
}

func (c *MockCounter) GetIncCalls(labels ...string) int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.incCalls == nil {
		return 0
	}

	key := c.buildKey(labels...)
	return c.incCalls[key]
}

func (c *MockCounter) GetAddCalls(labels ...string) int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.addCalls == nil {
		return 0
	}

	key := c.buildKey(labels...)
	return c.addCalls[key]
}

func (c *MockCounter) buildKey(labels ...string) string {
	if len(labels) == 0 {
		return "default"
	}
	key := ""
	for i, label := range labels {
		if i > 0 {
			key += ","
		}
		key += label
	}
	return key
}

// MockHistogram is a mock implementation of metrics.Histogram
type MockHistogram struct {
	name         string
	help         string
	labels       []string
	buckets      []float64
	observations map[string][]float64
	counts       map[string]uint64
	sums         map[string]float64
	observeCalls map[string]int
	timeCalls    map[string]int
	mutex        sync.RWMutex
}

func (h *MockHistogram) Observe(value float64, labels ...string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	key := h.buildKey(labels...)

	if h.observations == nil {
		h.observations = make(map[string][]float64)
	}
	if h.counts == nil {
		h.counts = make(map[string]uint64)
	}
	if h.sums == nil {
		h.sums = make(map[string]float64)
	}
	if h.observeCalls == nil {
		h.observeCalls = make(map[string]int)
	}

	h.observations[key] = append(h.observations[key], value)
	h.counts[key]++
	h.sums[key] += value
	h.observeCalls[key]++
}

func (h *MockHistogram) ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string) {
	h.Observe(value, labels...)
}

func (h *MockHistogram) Time(fn func(), labels ...string) {
	h.mutex.Lock()
	key := h.buildKey(labels...)
	if h.timeCalls == nil {
		h.timeCalls = make(map[string]int)
	}
	h.timeCalls[key]++
	h.mutex.Unlock()

	start := time.Now()
	fn()
	duration := time.Since(start).Seconds()
	h.Observe(duration, labels...)
}

func (h *MockHistogram) StartTimer(labels ...string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		h.Observe(duration, labels...)
	}
}

func (h *MockHistogram) GetCount(labels ...string) uint64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.counts == nil {
		return 0
	}

	key := h.buildKey(labels...)
	return h.counts[key]
}

func (h *MockHistogram) GetSum(labels ...string) float64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.sums == nil {
		return 0
	}

	key := h.buildKey(labels...)
	return h.sums[key]
}

func (h *MockHistogram) GetObservations(labels ...string) []float64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.observations == nil {
		return nil
	}

	key := h.buildKey(labels...)
	result := make([]float64, len(h.observations[key]))
	copy(result, h.observations[key])
	return result
}

func (h *MockHistogram) GetObserveCalls(labels ...string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.observeCalls == nil {
		return 0
	}

	key := h.buildKey(labels...)
	return h.observeCalls[key]
}

func (h *MockHistogram) buildKey(labels ...string) string {
	if len(labels) == 0 {
		return "default"
	}
	key := ""
	for i, label := range labels {
		if i > 0 {
			key += ","
		}
		key += label
	}
	return key
}

// MockGauge is a mock implementation of metrics.Gauge
type MockGauge struct {
	name     string
	help     string
	labels   []string
	values   map[string]float64
	setCalls map[string]int
	incCalls map[string]int
	decCalls map[string]int
	addCalls map[string]int
	subCalls map[string]int
	mutex    sync.RWMutex
}

func (g *MockGauge) Set(value float64, labels ...string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	key := g.buildKey(labels...)
	g.values[key] = value

	if g.setCalls == nil {
		g.setCalls = make(map[string]int)
	}
	g.setCalls[key]++
}

func (g *MockGauge) Inc(labels ...string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	key := g.buildKey(labels...)
	g.values[key]++

	if g.incCalls == nil {
		g.incCalls = make(map[string]int)
	}
	g.incCalls[key]++
}

func (g *MockGauge) Dec(labels ...string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	key := g.buildKey(labels...)
	g.values[key]--

	if g.decCalls == nil {
		g.decCalls = make(map[string]int)
	}
	g.decCalls[key]++
}

func (g *MockGauge) Add(value float64, labels ...string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	key := g.buildKey(labels...)
	g.values[key] += value

	if g.addCalls == nil {
		g.addCalls = make(map[string]int)
	}
	g.addCalls[key]++
}

func (g *MockGauge) Sub(value float64, labels ...string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	key := g.buildKey(labels...)
	g.values[key] -= value

	if g.subCalls == nil {
		g.subCalls = make(map[string]int)
	}
	g.subCalls[key]++
}

func (g *MockGauge) Get(labels ...string) float64 {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	key := g.buildKey(labels...)
	return g.values[key]
}

func (g *MockGauge) SetToCurrentTime(labels ...string) {
	g.Set(float64(time.Now().Unix()), labels...)
}

func (g *MockGauge) GetSetCalls(labels ...string) int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if g.setCalls == nil {
		return 0
	}

	key := g.buildKey(labels...)
	return g.setCalls[key]
}

func (g *MockGauge) buildKey(labels ...string) string {
	if len(labels) == 0 {
		return "default"
	}
	key := ""
	for i, label := range labels {
		if i > 0 {
			key += ","
		}
		key += label
	}
	return key
}

// MockSummary is a mock implementation of metrics.Summary
type MockSummary struct {
	name         string
	help         string
	labels       []string
	objectives   map[float64]float64
	observations map[string][]float64
	counts       map[string]uint64
	sums         map[string]float64
	observeCalls map[string]int
	timeCalls    map[string]int
	mutex        sync.RWMutex
}

func (s *MockSummary) Observe(value float64, labels ...string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.buildKey(labels...)

	if s.observations == nil {
		s.observations = make(map[string][]float64)
	}
	if s.counts == nil {
		s.counts = make(map[string]uint64)
	}
	if s.sums == nil {
		s.sums = make(map[string]float64)
	}
	if s.observeCalls == nil {
		s.observeCalls = make(map[string]int)
	}

	s.observations[key] = append(s.observations[key], value)
	s.counts[key]++
	s.sums[key] += value
	s.observeCalls[key]++
}

func (s *MockSummary) ObserveWithTimestamp(value float64, timestamp time.Time, labels ...string) {
	s.Observe(value, labels...)
}

func (s *MockSummary) Time(fn func(), labels ...string) {
	s.mutex.Lock()
	key := s.buildKey(labels...)
	if s.timeCalls == nil {
		s.timeCalls = make(map[string]int)
	}
	s.timeCalls[key]++
	s.mutex.Unlock()

	start := time.Now()
	fn()
	duration := time.Since(start).Seconds()
	s.Observe(duration, labels...)
}

func (s *MockSummary) StartTimer(labels ...string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		s.Observe(duration, labels...)
	}
}

func (s *MockSummary) GetCount(labels ...string) uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.counts == nil {
		return 0
	}

	key := s.buildKey(labels...)
	return s.counts[key]
}

func (s *MockSummary) GetSum(labels ...string) float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.sums == nil {
		return 0
	}

	key := s.buildKey(labels...)
	return s.sums[key]
}

func (s *MockSummary) GetQuantile(quantile float64, labels ...string) float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.observations == nil {
		return 0
	}

	key := s.buildKey(labels...)
	obs := s.observations[key]

	if len(obs) == 0 {
		return 0
	}

	// Simple quantile calculation for testing
	// This is not production-ready, just for mocking
	sortedObs := make([]float64, len(obs))
	copy(sortedObs, obs)

	// Simple bubble sort for testing
	for i := 0; i < len(sortedObs); i++ {
		for j := i + 1; j < len(sortedObs); j++ {
			if sortedObs[i] > sortedObs[j] {
				sortedObs[i], sortedObs[j] = sortedObs[j], sortedObs[i]
			}
		}
	}

	index := int(quantile * float64(len(sortedObs)-1))
	if index >= len(sortedObs) {
		index = len(sortedObs) - 1
	}

	return sortedObs[index]
}

func (s *MockSummary) GetObservations(labels ...string) []float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.observations == nil {
		return nil
	}

	key := s.buildKey(labels...)
	result := make([]float64, len(s.observations[key]))
	copy(result, s.observations[key])
	return result
}

func (s *MockSummary) buildKey(labels ...string) string {
	if len(labels) == 0 {
		return "default"
	}
	key := ""
	for i, label := range labels {
		if i > 0 {
			key += ","
		}
		key += label
	}
	return key
}
