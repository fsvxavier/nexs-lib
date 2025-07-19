package logger

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

// metricsCollector implementação thread-safe do coletor de métricas
type metricsCollector struct {
	// Contadores atômicos por nível
	debugCount int64
	infoCount  int64
	warnCount  int64
	errorCount int64
	fatalCount int64
	panicCount int64

	// Contadores de erro e sampling
	errorTotal  int64
	sampleCount int64
	sampleTotal int64

	// Métricas de tempo (usando mutex para proteção)
	mu               sync.RWMutex
	processingTimes  map[interfaces.Level][]time.Duration
	providerStats    map[string]*interfaces.ProviderStats
	bufferOperations map[string][]time.Duration
	startTime        time.Time
	lastResetTime    time.Time
}

// NewMetricsCollector cria um novo coletor de métricas
func NewMetricsCollector() interfaces.MetricsCollector {
	return &metricsCollector{
		processingTimes:  make(map[interfaces.Level][]time.Duration),
		providerStats:    make(map[string]*interfaces.ProviderStats),
		bufferOperations: make(map[string][]time.Duration),
		startTime:        time.Now(),
		lastResetTime:    time.Now(),
	}
}

// RecordLog registra uma operação de log
func (m *metricsCollector) RecordLog(level interfaces.Level, duration time.Duration) {
	// Incrementa contador atômico
	switch level {
	case interfaces.DebugLevel:
		atomic.AddInt64(&m.debugCount, 1)
	case interfaces.InfoLevel:
		atomic.AddInt64(&m.infoCount, 1)
	case interfaces.WarnLevel:
		atomic.AddInt64(&m.warnCount, 1)
	case interfaces.ErrorLevel:
		atomic.AddInt64(&m.errorCount, 1)
	case interfaces.FatalLevel:
		atomic.AddInt64(&m.fatalCount, 1)
	case interfaces.PanicLevel:
		atomic.AddInt64(&m.panicCount, 1)
	}

	// Registra tempo de processamento (limitado para evitar consumo excessivo de memória)
	m.mu.Lock()
	times := m.processingTimes[level]

	// Mantém apenas os últimos 1000 tempos para cada nível
	const maxTimes = 1000
	if len(times) >= maxTimes {
		times = times[1:] // Remove o mais antigo
	}
	times = append(times, duration)
	m.processingTimes[level] = times
	m.mu.Unlock()
}

// RecordError registra um erro
func (m *metricsCollector) RecordError(err error) {
	if err != nil {
		atomic.AddInt64(&m.errorTotal, 1)
	}
}

// RecordSample registra operação de sampling
func (m *metricsCollector) RecordSample(sampled bool) {
	atomic.AddInt64(&m.sampleTotal, 1)
	if sampled {
		atomic.AddInt64(&m.sampleCount, 1)
	}
}

// RecordProviderOperation registra operação de provider
func (m *metricsCollector) RecordProviderOperation(provider string, operation string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.providerStats[provider]
	if !exists {
		stats = &interfaces.ProviderStats{
			ProviderName:      provider,
			LogCount:          make(map[interfaces.Level]int64),
			ConfigurationTime: time.Now(),
		}
		m.providerStats[provider] = stats
	}

	stats.TotalLogs++
	stats.LastLogTime = time.Now()

	// Calcula latência média (usando média móvel simples)
	if stats.AverageLatency == 0 {
		stats.AverageLatency = duration
	} else {
		stats.AverageLatency = (stats.AverageLatency + duration) / 2
	}
}

// RecordProviderError registra erro de provider
func (m *metricsCollector) RecordProviderError(provider string, err error) {
	if err == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.providerStats[provider]
	if !exists {
		stats = &interfaces.ProviderStats{
			ProviderName:      provider,
			LogCount:          make(map[interfaces.Level]int64),
			ConfigurationTime: time.Now(),
		}
		m.providerStats[provider] = stats
	}

	stats.ErrorCount++
}

// RecordBufferOperation registra operação de buffer
func (m *metricsCollector) RecordBufferOperation(operation string, size int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	times := m.bufferOperations[operation]

	// Mantém apenas os últimos 100 tempos para cada operação
	const maxTimes = 100
	if len(times) >= maxTimes {
		times = times[1:]
	}
	times = append(times, duration)
	m.bufferOperations[operation] = times
}

// GetMetrics retorna as métricas atuais
func (m *metricsCollector) GetMetrics() interfaces.Metrics {
	return &metrics{collector: m}
}

// metrics implementação da interface Metrics
type metrics struct {
	collector *metricsCollector
}

// GetLogCount retorna contagem de logs por nível
func (m *metrics) GetLogCount(level interfaces.Level) int64 {
	switch level {
	case interfaces.DebugLevel:
		return atomic.LoadInt64(&m.collector.debugCount)
	case interfaces.InfoLevel:
		return atomic.LoadInt64(&m.collector.infoCount)
	case interfaces.WarnLevel:
		return atomic.LoadInt64(&m.collector.warnCount)
	case interfaces.ErrorLevel:
		return atomic.LoadInt64(&m.collector.errorCount)
	case interfaces.FatalLevel:
		return atomic.LoadInt64(&m.collector.fatalCount)
	case interfaces.PanicLevel:
		return atomic.LoadInt64(&m.collector.panicCount)
	default:
		return 0
	}
}

// GetTotalLogCount retorna contagem total de logs
func (m *metrics) GetTotalLogCount() int64 {
	return m.GetLogCount(interfaces.DebugLevel) +
		m.GetLogCount(interfaces.InfoLevel) +
		m.GetLogCount(interfaces.WarnLevel) +
		m.GetLogCount(interfaces.ErrorLevel) +
		m.GetLogCount(interfaces.FatalLevel) +
		m.GetLogCount(interfaces.PanicLevel)
}

// GetAverageProcessingTime retorna tempo médio de processamento
func (m *metrics) GetAverageProcessingTime() time.Duration {
	m.collector.mu.RLock()
	defer m.collector.mu.RUnlock()

	var totalDuration time.Duration
	var totalCount int

	for _, times := range m.collector.processingTimes {
		for _, duration := range times {
			totalDuration += duration
			totalCount++
		}
	}

	if totalCount == 0 {
		return 0
	}

	return totalDuration / time.Duration(totalCount)
}

// GetProcessingTimeByLevel retorna tempo médio por nível
func (m *metrics) GetProcessingTimeByLevel(level interfaces.Level) time.Duration {
	m.collector.mu.RLock()
	defer m.collector.mu.RUnlock()

	times, exists := m.collector.processingTimes[level]
	if !exists || len(times) == 0 {
		return 0
	}

	var totalDuration time.Duration
	for _, duration := range times {
		totalDuration += duration
	}

	return totalDuration / time.Duration(len(times))
}

// GetErrorRate retorna taxa de erro
func (m *metrics) GetErrorRate() float64 {
	totalErrors := atomic.LoadInt64(&m.collector.errorTotal)
	totalLogs := m.GetTotalLogCount()

	if totalLogs == 0 {
		return 0.0
	}

	return float64(totalErrors) / float64(totalLogs)
}

// GetSamplingRate retorna taxa de sampling
func (m *metrics) GetSamplingRate() float64 {
	sampledCount := atomic.LoadInt64(&m.collector.sampleCount)
	totalSamples := atomic.LoadInt64(&m.collector.sampleTotal)

	if totalSamples == 0 {
		return 0.0
	}

	return float64(sampledCount) / float64(totalSamples)
}

// GetProviderStats retorna estatísticas de um provider
func (m *metrics) GetProviderStats(provider string) *interfaces.ProviderStats {
	m.collector.mu.RLock()
	defer m.collector.mu.RUnlock()

	stats, exists := m.collector.providerStats[provider]
	if !exists {
		return nil
	}

	// Retorna uma cópia para evitar modificações concorrentes
	result := *stats
	result.LogCount = make(map[interfaces.Level]int64)
	for level, count := range stats.LogCount {
		result.LogCount[level] = count
	}

	return &result
}

// Reset reseta todas as métricas
func (m *metrics) Reset() {
	// Reset contadores atômicos
	atomic.StoreInt64(&m.collector.debugCount, 0)
	atomic.StoreInt64(&m.collector.infoCount, 0)
	atomic.StoreInt64(&m.collector.warnCount, 0)
	atomic.StoreInt64(&m.collector.errorCount, 0)
	atomic.StoreInt64(&m.collector.fatalCount, 0)
	atomic.StoreInt64(&m.collector.panicCount, 0)
	atomic.StoreInt64(&m.collector.errorTotal, 0)
	atomic.StoreInt64(&m.collector.sampleCount, 0)
	atomic.StoreInt64(&m.collector.sampleTotal, 0)

	// Reset mapas protegidos por mutex
	m.collector.mu.Lock()
	m.collector.processingTimes = make(map[interfaces.Level][]time.Duration)
	m.collector.providerStats = make(map[string]*interfaces.ProviderStats)
	m.collector.bufferOperations = make(map[string][]time.Duration)
	m.collector.lastResetTime = time.Now()
	m.collector.mu.Unlock()
}

// Export exporta métricas para sistemas externos
func (m *metrics) Export() map[string]interface{} {
	m.collector.mu.RLock()
	defer m.collector.mu.RUnlock()

	export := map[string]interface{}{
		"start_time":      m.collector.startTime,
		"last_reset_time": m.collector.lastResetTime,
		"log_counts": map[string]int64{
			"debug": m.GetLogCount(interfaces.DebugLevel),
			"info":  m.GetLogCount(interfaces.InfoLevel),
			"warn":  m.GetLogCount(interfaces.WarnLevel),
			"error": m.GetLogCount(interfaces.ErrorLevel),
			"fatal": m.GetLogCount(interfaces.FatalLevel),
			"panic": m.GetLogCount(interfaces.PanicLevel),
			"total": m.GetTotalLogCount(),
		},
		"performance": map[string]interface{}{
			"average_processing_time": m.GetAverageProcessingTime().String(),
			"error_rate":              m.GetErrorRate(),
			"sampling_rate":           m.GetSamplingRate(),
		},
		"providers": m.collector.providerStats,
	}

	// Adiciona tempos de processamento por nível
	processingByLevel := make(map[string]string)
	for level := interfaces.DebugLevel; level <= interfaces.PanicLevel; level++ {
		processingByLevel[level.String()] = m.GetProcessingTimeByLevel(level).String()
	}
	export["processing_time_by_level"] = processingByLevel

	return export
}
