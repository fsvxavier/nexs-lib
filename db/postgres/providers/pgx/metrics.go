package pgxprovider

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceMetrics coleta métricas detalhadas de performance
type PerformanceMetrics struct {
	// Query metrics
	totalQueries     int64
	failedQueries    int64
	avgQueryDuration int64 // em nanosegundos
	maxQueryDuration int64
	minQueryDuration int64

	// Connection metrics
	totalConnections  int64
	failedConnections int64
	avgConnectionTime int64
	maxConnectionTime int64
	minConnectionTime int64

	// Transaction metrics
	totalTransactions  int64
	failedTransactions int64
	avgTransactionTime int64
	maxTransactionTime int64
	minTransactionTime int64

	// Throughput metrics
	queriesPerSecond     int64
	connectionsPerSecond int64

	// Buffer pool metrics
	bufferHits        int64
	bufferMisses      int64
	bufferAllocations int64

	// Error metrics
	errorsByType map[string]int64
	errorsMutex  sync.RWMutex

	// Latency histograms
	queryLatencyBuckets []LatencyBucket
	connLatencyBuckets  []LatencyBucket

	// Timing
	lastReset time.Time
	startTime time.Time

	mu sync.RWMutex
}

// LatencyBucket representa um bucket de latência
type LatencyBucket struct {
	UpperBound time.Duration
	Count      int64
}

// MetricsCollector interface para coleta de métricas
type MetricsCollector interface {
	RecordQuery(ctx context.Context, query string, duration time.Duration, err error)
	RecordConnection(ctx context.Context, duration time.Duration, err error)
	RecordTransaction(ctx context.Context, duration time.Duration, err error)
	RecordBufferHit()
	RecordBufferMiss()
	RecordBufferAllocation()
	GetMetrics() *PerformanceMetrics
	ResetMetrics()
}

// NewPerformanceMetrics cria um novo coletor de métricas
func NewPerformanceMetrics() *PerformanceMetrics {
	now := time.Now()

	return &PerformanceMetrics{
		minQueryDuration:    int64(time.Hour), // Inicializar com valor alto
		minConnectionTime:   int64(time.Hour),
		minTransactionTime:  int64(time.Hour),
		errorsByType:        make(map[string]int64),
		lastReset:           now,
		startTime:           now,
		queryLatencyBuckets: createLatencyBuckets(),
		connLatencyBuckets:  createLatencyBuckets(),
	}
}

// createLatencyBuckets cria buckets de latência
func createLatencyBuckets() []LatencyBucket {
	return []LatencyBucket{
		{UpperBound: 1 * time.Millisecond, Count: 0},
		{UpperBound: 5 * time.Millisecond, Count: 0},
		{UpperBound: 10 * time.Millisecond, Count: 0},
		{UpperBound: 25 * time.Millisecond, Count: 0},
		{UpperBound: 50 * time.Millisecond, Count: 0},
		{UpperBound: 100 * time.Millisecond, Count: 0},
		{UpperBound: 250 * time.Millisecond, Count: 0},
		{UpperBound: 500 * time.Millisecond, Count: 0},
		{UpperBound: 1 * time.Second, Count: 0},
		{UpperBound: 5 * time.Second, Count: 0},
		{UpperBound: 10 * time.Second, Count: 0},
	}
}

// RecordQuery registra métricas de uma query
func (pm *PerformanceMetrics) RecordQuery(ctx context.Context, query string, duration time.Duration, err error) {
	durationNs := int64(duration)

	// Incrementar contadores
	atomic.AddInt64(&pm.totalQueries, 1)

	if err != nil {
		atomic.AddInt64(&pm.failedQueries, 1)
		pm.recordError(err)
	}

	// Atualizar latência
	pm.updateQueryLatency(durationNs)

	// Atualizar buckets de latência
	pm.updateLatencyBuckets(pm.queryLatencyBuckets, duration)
}

// RecordConnection registra métricas de conexão
func (pm *PerformanceMetrics) RecordConnection(ctx context.Context, duration time.Duration, err error) {
	durationNs := int64(duration)

	// Incrementar contadores
	atomic.AddInt64(&pm.totalConnections, 1)

	if err != nil {
		atomic.AddInt64(&pm.failedConnections, 1)
		pm.recordError(err)
	}

	// Atualizar latência
	pm.updateConnectionLatency(durationNs)

	// Atualizar buckets de latência
	pm.updateLatencyBuckets(pm.connLatencyBuckets, duration)
}

// RecordTransaction registra métricas de transação
func (pm *PerformanceMetrics) RecordTransaction(ctx context.Context, duration time.Duration, err error) {
	durationNs := int64(duration)

	// Incrementar contadores
	atomic.AddInt64(&pm.totalTransactions, 1)

	if err != nil {
		atomic.AddInt64(&pm.failedTransactions, 1)
		pm.recordError(err)
	}

	// Atualizar latência
	pm.updateTransactionLatency(durationNs)
}

// RecordBufferHit registra um buffer hit
func (pm *PerformanceMetrics) RecordBufferHit() {
	atomic.AddInt64(&pm.bufferHits, 1)
}

// RecordBufferMiss registra um buffer miss
func (pm *PerformanceMetrics) RecordBufferMiss() {
	atomic.AddInt64(&pm.bufferMisses, 1)
}

// RecordBufferAllocation registra uma alocação de buffer
func (pm *PerformanceMetrics) RecordBufferAllocation() {
	atomic.AddInt64(&pm.bufferAllocations, 1)
}

// recordError registra um erro por tipo
func (pm *PerformanceMetrics) recordError(err error) {
	errorType := "unknown"
	if err != nil {
		errorType = err.Error()
		// Simplificar tipo do erro para agrupamento
		if len(errorType) > 50 {
			errorType = errorType[:50] + "..."
		}
	}

	pm.errorsMutex.Lock()
	pm.errorsByType[errorType]++
	pm.errorsMutex.Unlock()
}

// updateQueryLatency atualiza métricas de latência de query
func (pm *PerformanceMetrics) updateQueryLatency(durationNs int64) {
	// Atualizar média usando atomic operations
	total := atomic.LoadInt64(&pm.totalQueries)
	if total > 0 {
		currentAvg := atomic.LoadInt64(&pm.avgQueryDuration)
		newAvg := (currentAvg*(total-1) + durationNs) / total
		atomic.StoreInt64(&pm.avgQueryDuration, newAvg)
	}

	// Atualizar máximo
	for {
		current := atomic.LoadInt64(&pm.maxQueryDuration)
		if durationNs <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.maxQueryDuration, current, durationNs) {
			break
		}
	}

	// Atualizar mínimo
	for {
		current := atomic.LoadInt64(&pm.minQueryDuration)
		if durationNs >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.minQueryDuration, current, durationNs) {
			break
		}
	}
}

// updateConnectionLatency atualiza métricas de latência de conexão
func (pm *PerformanceMetrics) updateConnectionLatency(durationNs int64) {
	// Atualizar média usando atomic operations
	total := atomic.LoadInt64(&pm.totalConnections)
	if total > 0 {
		currentAvg := atomic.LoadInt64(&pm.avgConnectionTime)
		newAvg := (currentAvg*(total-1) + durationNs) / total
		atomic.StoreInt64(&pm.avgConnectionTime, newAvg)
	}

	// Atualizar máximo
	for {
		current := atomic.LoadInt64(&pm.maxConnectionTime)
		if durationNs <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.maxConnectionTime, current, durationNs) {
			break
		}
	}

	// Atualizar mínimo
	for {
		current := atomic.LoadInt64(&pm.minConnectionTime)
		if durationNs >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.minConnectionTime, current, durationNs) {
			break
		}
	}
}

// updateTransactionLatency atualiza métricas de latência de transação
func (pm *PerformanceMetrics) updateTransactionLatency(durationNs int64) {
	// Atualizar média usando atomic operations
	total := atomic.LoadInt64(&pm.totalTransactions)
	if total > 0 {
		currentAvg := atomic.LoadInt64(&pm.avgTransactionTime)
		newAvg := (currentAvg*(total-1) + durationNs) / total
		atomic.StoreInt64(&pm.avgTransactionTime, newAvg)
	}

	// Atualizar máximo
	for {
		current := atomic.LoadInt64(&pm.maxTransactionTime)
		if durationNs <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.maxTransactionTime, current, durationNs) {
			break
		}
	}

	// Atualizar mínimo
	for {
		current := atomic.LoadInt64(&pm.minTransactionTime)
		if durationNs >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&pm.minTransactionTime, current, durationNs) {
			break
		}
	}
}

// updateLatencyBuckets atualiza buckets de latência
func (pm *PerformanceMetrics) updateLatencyBuckets(buckets []LatencyBucket, duration time.Duration) {
	for i := range buckets {
		if duration <= buckets[i].UpperBound {
			atomic.AddInt64(&buckets[i].Count, 1)
			break
		}
	}
}

// GetMetrics retorna uma cópia das métricas atuais
func (pm *PerformanceMetrics) GetMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Criar cópia das métricas
	metrics := &PerformanceMetrics{
		totalQueries:        atomic.LoadInt64(&pm.totalQueries),
		failedQueries:       atomic.LoadInt64(&pm.failedQueries),
		avgQueryDuration:    atomic.LoadInt64(&pm.avgQueryDuration),
		maxQueryDuration:    atomic.LoadInt64(&pm.maxQueryDuration),
		minQueryDuration:    atomic.LoadInt64(&pm.minQueryDuration),
		totalConnections:    atomic.LoadInt64(&pm.totalConnections),
		failedConnections:   atomic.LoadInt64(&pm.failedConnections),
		avgConnectionTime:   atomic.LoadInt64(&pm.avgConnectionTime),
		maxConnectionTime:   atomic.LoadInt64(&pm.maxConnectionTime),
		minConnectionTime:   atomic.LoadInt64(&pm.minConnectionTime),
		totalTransactions:   atomic.LoadInt64(&pm.totalTransactions),
		failedTransactions:  atomic.LoadInt64(&pm.failedTransactions),
		avgTransactionTime:  atomic.LoadInt64(&pm.avgTransactionTime),
		maxTransactionTime:  atomic.LoadInt64(&pm.maxTransactionTime),
		minTransactionTime:  atomic.LoadInt64(&pm.minTransactionTime),
		bufferHits:          atomic.LoadInt64(&pm.bufferHits),
		bufferMisses:        atomic.LoadInt64(&pm.bufferMisses),
		bufferAllocations:   atomic.LoadInt64(&pm.bufferAllocations),
		lastReset:           pm.lastReset,
		startTime:           pm.startTime,
		errorsByType:        make(map[string]int64),
		queryLatencyBuckets: make([]LatencyBucket, len(pm.queryLatencyBuckets)),
		connLatencyBuckets:  make([]LatencyBucket, len(pm.connLatencyBuckets)),
	}

	// Copiar errors
	pm.errorsMutex.RLock()
	for k, v := range pm.errorsByType {
		metrics.errorsByType[k] = v
	}
	pm.errorsMutex.RUnlock()

	// Copiar buckets
	copy(metrics.queryLatencyBuckets, pm.queryLatencyBuckets)
	copy(metrics.connLatencyBuckets, pm.connLatencyBuckets)

	return metrics
}

// ResetMetrics reseta todas as métricas
func (pm *PerformanceMetrics) ResetMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Reset contadores
	atomic.StoreInt64(&pm.totalQueries, 0)
	atomic.StoreInt64(&pm.failedQueries, 0)
	atomic.StoreInt64(&pm.avgQueryDuration, 0)
	atomic.StoreInt64(&pm.maxQueryDuration, 0)
	atomic.StoreInt64(&pm.minQueryDuration, int64(time.Hour))

	atomic.StoreInt64(&pm.totalConnections, 0)
	atomic.StoreInt64(&pm.failedConnections, 0)
	atomic.StoreInt64(&pm.avgConnectionTime, 0)
	atomic.StoreInt64(&pm.maxConnectionTime, 0)
	atomic.StoreInt64(&pm.minConnectionTime, int64(time.Hour))

	atomic.StoreInt64(&pm.totalTransactions, 0)
	atomic.StoreInt64(&pm.failedTransactions, 0)
	atomic.StoreInt64(&pm.avgTransactionTime, 0)
	atomic.StoreInt64(&pm.maxTransactionTime, 0)
	atomic.StoreInt64(&pm.minTransactionTime, int64(time.Hour))

	atomic.StoreInt64(&pm.bufferHits, 0)
	atomic.StoreInt64(&pm.bufferMisses, 0)
	atomic.StoreInt64(&pm.bufferAllocations, 0)

	// Reset errors
	pm.errorsMutex.Lock()
	pm.errorsByType = make(map[string]int64)
	pm.errorsMutex.Unlock()

	// Reset buckets
	pm.queryLatencyBuckets = createLatencyBuckets()
	pm.connLatencyBuckets = createLatencyBuckets()

	pm.lastReset = time.Now()
}

// GetSummary retorna um resumo das métricas
func (pm *PerformanceMetrics) GetSummary() map[string]interface{} {
	metrics := pm.GetMetrics()
	uptime := time.Since(metrics.startTime)

	// Calcular rates
	var queryRate, connectionRate, transactionRate float64
	if uptime.Seconds() > 0 {
		queryRate = float64(metrics.totalQueries) / uptime.Seconds()
		connectionRate = float64(metrics.totalConnections) / uptime.Seconds()
		transactionRate = float64(metrics.totalTransactions) / uptime.Seconds()
	}

	// Calcular success rates
	var querySuccessRate, connectionSuccessRate, transactionSuccessRate float64
	if metrics.totalQueries > 0 {
		querySuccessRate = float64(metrics.totalQueries-metrics.failedQueries) / float64(metrics.totalQueries) * 100
	}
	if metrics.totalConnections > 0 {
		connectionSuccessRate = float64(metrics.totalConnections-metrics.failedConnections) / float64(metrics.totalConnections) * 100
	}
	if metrics.totalTransactions > 0 {
		transactionSuccessRate = float64(metrics.totalTransactions-metrics.failedTransactions) / float64(metrics.totalTransactions) * 100
	}

	// Calcular buffer hit rate
	var bufferHitRate float64
	totalBufferOps := metrics.bufferHits + metrics.bufferMisses
	if totalBufferOps > 0 {
		bufferHitRate = float64(metrics.bufferHits) / float64(totalBufferOps) * 100
	}

	return map[string]interface{}{
		"uptime_seconds": uptime.Seconds(),
		"queries": map[string]interface{}{
			"total":           metrics.totalQueries,
			"failed":          metrics.failedQueries,
			"success_rate":    querySuccessRate,
			"rate_per_sec":    queryRate,
			"avg_duration_ms": float64(metrics.avgQueryDuration) / 1000000,
			"max_duration_ms": float64(metrics.maxQueryDuration) / 1000000,
			"min_duration_ms": float64(metrics.minQueryDuration) / 1000000,
		},
		"connections": map[string]interface{}{
			"total":           metrics.totalConnections,
			"failed":          metrics.failedConnections,
			"success_rate":    connectionSuccessRate,
			"rate_per_sec":    connectionRate,
			"avg_duration_ms": float64(metrics.avgConnectionTime) / 1000000,
			"max_duration_ms": float64(metrics.maxConnectionTime) / 1000000,
			"min_duration_ms": float64(metrics.minConnectionTime) / 1000000,
		},
		"transactions": map[string]interface{}{
			"total":           metrics.totalTransactions,
			"failed":          metrics.failedTransactions,
			"success_rate":    transactionSuccessRate,
			"rate_per_sec":    transactionRate,
			"avg_duration_ms": float64(metrics.avgTransactionTime) / 1000000,
			"max_duration_ms": float64(metrics.maxTransactionTime) / 1000000,
			"min_duration_ms": float64(metrics.minTransactionTime) / 1000000,
		},
		"buffer_pool": map[string]interface{}{
			"hits":        metrics.bufferHits,
			"misses":      metrics.bufferMisses,
			"allocations": metrics.bufferAllocations,
			"hit_rate":    bufferHitRate,
		},
		"errors": metrics.errorsByType,
	}
}
