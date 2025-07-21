// Package hooks provides observer implementations for HTTP server lifecycle events.
package hooks

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// MetricsData holds metrics for a server.
type MetricsData struct {
	RequestCount      int64
	ErrorCount        int64
	TotalDuration     int64 // in nanoseconds
	ActiveRequests    int64
	StartTime         time.Time
	LastRequestTimeNs int64 // Using int64 for atomic operations
}

// GetLastRequestTime returns the last request time as time.Time.
func (m *MetricsData) GetLastRequestTime() time.Time {
	ns := atomic.LoadInt64(&m.LastRequestTimeNs)
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// MetricsObserver implements ServerObserver for collecting metrics.
type MetricsObserver struct {
	metrics map[string]*MetricsData
}

// NewMetricsObserver creates a new metrics observer.
func NewMetricsObserver() *MetricsObserver {
	return &MetricsObserver{
		metrics: make(map[string]*MetricsData),
	}
}

// OnStart records when the server starts.
func (m *MetricsObserver) OnStart(name string) {
	m.metrics[name] = &MetricsData{
		StartTime: time.Now(),
	}
}

// OnStop cleans up the metrics for the server.
func (m *MetricsObserver) OnStop(name string) {
	delete(m.metrics, name)
}

// OnRequest records metrics for each request.
func (m *MetricsObserver) OnRequest(name string, req *http.Request, status int, duration time.Duration) {
	data := m.metrics[name]
	if data == nil {
		return
	}

	atomic.AddInt64(&data.RequestCount, 1)
	if status >= 400 {
		atomic.AddInt64(&data.ErrorCount, 1)
	}
	atomic.AddInt64(&data.TotalDuration, duration.Nanoseconds())

	// Store current time as nanoseconds for atomic operations
	atomic.StoreInt64(&data.LastRequestTimeNs, time.Now().UnixNano())
}

// OnBeforeRequest increments the active request counter.
func (m *MetricsObserver) OnBeforeRequest(name string, req *http.Request) {
	data := m.metrics[name]
	if data != nil {
		atomic.AddInt64(&data.ActiveRequests, 1)
	}
}

// OnAfterRequest decrements the active request counter.
func (m *MetricsObserver) OnAfterRequest(name string, req *http.Request, status int, duration time.Duration) {
	data := m.metrics[name]
	if data != nil {
		atomic.AddInt64(&data.ActiveRequests, -1)
	}
}

// GetMetrics returns a copy of the metrics for a server.
func (m *MetricsObserver) GetMetrics(name string) *MetricsData {
	if m.metrics[name] == nil {
		return nil
	}

	data := m.metrics[name]
	return &MetricsData{
		RequestCount:      atomic.LoadInt64(&data.RequestCount),
		ErrorCount:        atomic.LoadInt64(&data.ErrorCount),
		TotalDuration:     atomic.LoadInt64(&data.TotalDuration),
		ActiveRequests:    atomic.LoadInt64(&data.ActiveRequests),
		StartTime:         data.StartTime,
		LastRequestTimeNs: atomic.LoadInt64(&data.LastRequestTimeNs),
	}
}

// Verify MetricsObserver implements ServerObserver.
var _ interfaces.ServerObserver = (*MetricsObserver)(nil)
