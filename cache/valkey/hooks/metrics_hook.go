package hooks

import (
	"context"
	"sync"
	"time"
)

// MetricsHook implementa hooks para coleta de métricas.
type MetricsHook struct {
	mu sync.RWMutex

	// Contadores
	commandsExecuted  int64
	connectionsOpened int64
	connectionsClosed int64
	pipelinesExecuted int64
	retriesAttempted  int64
	errorsOccurred    int64

	// Tempos de execução
	executionTimes  []time.Duration
	connectionTimes []time.Duration
	pipelineTimes   []time.Duration

	// Configurações
	collectExecutionMetrics  bool
	collectConnectionMetrics bool
	collectPipelineMetrics   bool
	collectRetryMetrics      bool

	// Limites de histórico
	maxHistorySize int
}

// NewMetricsHook cria um novo MetricsHook.
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{
		collectExecutionMetrics:  true,
		collectConnectionMetrics: true,
		collectPipelineMetrics:   true,
		collectRetryMetrics:      true,
		maxHistorySize:           1000,
		executionTimes:           make([]time.Duration, 0),
		connectionTimes:          make([]time.Duration, 0),
		pipelineTimes:            make([]time.Duration, 0),
	}
}

// WithExecutionMetrics habilita/desabilita coleta de métricas de execução.
func (m *MetricsHook) WithExecutionMetrics(enabled bool) *MetricsHook {
	m.collectExecutionMetrics = enabled
	return m
}

// WithConnectionMetrics habilita/desabilita coleta de métricas de conexão.
func (m *MetricsHook) WithConnectionMetrics(enabled bool) *MetricsHook {
	m.collectConnectionMetrics = enabled
	return m
}

// WithPipelineMetrics habilita/desabilita coleta de métricas de pipeline.
func (m *MetricsHook) WithPipelineMetrics(enabled bool) *MetricsHook {
	m.collectPipelineMetrics = enabled
	return m
}

// WithRetryMetrics habilita/desabilita coleta de métricas de retry.
func (m *MetricsHook) WithRetryMetrics(enabled bool) *MetricsHook {
	m.collectRetryMetrics = enabled
	return m
}

// WithMaxHistorySize define o tamanho máximo do histórico de métricas.
func (m *MetricsHook) WithMaxHistorySize(size int) *MetricsHook {
	m.maxHistorySize = size
	return m
}

// BeforeExecution implementa ExecutionHook.
func (m *MetricsHook) BeforeExecution(ctx context.Context, cmd string, args []interface{}) context.Context {
	return ctx
}

// AfterExecution implementa ExecutionHook.
func (m *MetricsHook) AfterExecution(ctx context.Context, cmd string, args []interface{}, result interface{}, err error, duration time.Duration) {
	if !m.collectExecutionMetrics {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.commandsExecuted++
	if err != nil {
		m.errorsOccurred++
	}

	// Adicionar tempo de execução ao histórico
	m.executionTimes = append(m.executionTimes, duration)
	if len(m.executionTimes) > m.maxHistorySize {
		m.executionTimes = m.executionTimes[1:]
	}
}

// BeforeConnect implementa ConnectionHook.
func (m *MetricsHook) BeforeConnect(ctx context.Context, network, addr string) context.Context {
	return ctx
}

// AfterConnect implementa ConnectionHook.
func (m *MetricsHook) AfterConnect(ctx context.Context, network, addr string, err error, duration time.Duration) {
	if !m.collectConnectionMetrics {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.connectionsOpened++
	if err != nil {
		m.errorsOccurred++
	}

	// Adicionar tempo de conexão ao histórico
	m.connectionTimes = append(m.connectionTimes, duration)
	if len(m.connectionTimes) > m.maxHistorySize {
		m.connectionTimes = m.connectionTimes[1:]
	}
}

// BeforeDisconnect implementa ConnectionHook.
func (m *MetricsHook) BeforeDisconnect(ctx context.Context, network, addr string) context.Context {
	return ctx
}

// AfterDisconnect implementa ConnectionHook.
func (m *MetricsHook) AfterDisconnect(ctx context.Context, network, addr string, err error, duration time.Duration) {
	if !m.collectConnectionMetrics {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.connectionsClosed++
	if err != nil {
		m.errorsOccurred++
	}
}

// BeforePipelineExecution implementa PipelineHook.
func (m *MetricsHook) BeforePipelineExecution(ctx context.Context, commands []string) context.Context {
	return ctx
}

// AfterPipelineExecution implementa PipelineHook.
func (m *MetricsHook) AfterPipelineExecution(ctx context.Context, commands []string, results []interface{}, err error, duration time.Duration) {
	if !m.collectPipelineMetrics {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.pipelinesExecuted++
	if err != nil {
		m.errorsOccurred++
	}

	// Adicionar tempo de pipeline ao histórico
	m.pipelineTimes = append(m.pipelineTimes, duration)
	if len(m.pipelineTimes) > m.maxHistorySize {
		m.pipelineTimes = m.pipelineTimes[1:]
	}
}

// BeforeRetry implementa RetryHook.
func (m *MetricsHook) BeforeRetry(ctx context.Context, attempt int, err error) context.Context {
	return ctx
}

// AfterRetry implementa RetryHook.
func (m *MetricsHook) AfterRetry(ctx context.Context, attempt int, success bool, err error) {
	if !m.collectRetryMetrics {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.retriesAttempted++
	if err != nil {
		m.errorsOccurred++
	}
}

// Metrics representa as métricas coletadas.
type Metrics struct {
	CommandsExecuted  int64         `json:"commands_executed"`
	ConnectionsOpened int64         `json:"connections_opened"`
	ConnectionsClosed int64         `json:"connections_closed"`
	PipelinesExecuted int64         `json:"pipelines_executed"`
	RetriesAttempted  int64         `json:"retries_attempted"`
	ErrorsOccurred    int64         `json:"errors_occurred"`
	AvgExecutionTime  time.Duration `json:"avg_execution_time"`
	AvgConnectionTime time.Duration `json:"avg_connection_time"`
	AvgPipelineTime   time.Duration `json:"avg_pipeline_time"`
	MaxExecutionTime  time.Duration `json:"max_execution_time"`
	MinExecutionTime  time.Duration `json:"min_execution_time"`
	MaxConnectionTime time.Duration `json:"max_connection_time"`
	MinConnectionTime time.Duration `json:"min_connection_time"`
}

// GetMetrics retorna um snapshot das métricas atuais.
func (m *MetricsHook) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := Metrics{
		CommandsExecuted:  m.commandsExecuted,
		ConnectionsOpened: m.connectionsOpened,
		ConnectionsClosed: m.connectionsClosed,
		PipelinesExecuted: m.pipelinesExecuted,
		RetriesAttempted:  m.retriesAttempted,
		ErrorsOccurred:    m.errorsOccurred,
	}

	// Calcular estatísticas de tempo de execução
	if len(m.executionTimes) > 0 {
		var total time.Duration
		min := m.executionTimes[0]
		max := m.executionTimes[0]

		for _, d := range m.executionTimes {
			total += d
			if d < min {
				min = d
			}
			if d > max {
				max = d
			}
		}

		metrics.AvgExecutionTime = total / time.Duration(len(m.executionTimes))
		metrics.MinExecutionTime = min
		metrics.MaxExecutionTime = max
	}

	// Calcular estatísticas de tempo de conexão
	if len(m.connectionTimes) > 0 {
		var total time.Duration
		min := m.connectionTimes[0]
		max := m.connectionTimes[0]

		for _, d := range m.connectionTimes {
			total += d
			if d < min {
				min = d
			}
			if d > max {
				max = d
			}
		}

		metrics.AvgConnectionTime = total / time.Duration(len(m.connectionTimes))
		metrics.MinConnectionTime = min
		metrics.MaxConnectionTime = max
	}

	// Calcular estatísticas de tempo de pipeline
	if len(m.pipelineTimes) > 0 {
		var total time.Duration
		for _, d := range m.pipelineTimes {
			total += d
		}
		metrics.AvgPipelineTime = total / time.Duration(len(m.pipelineTimes))
	}

	return metrics
}

// Reset limpa todas as métricas.
func (m *MetricsHook) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commandsExecuted = 0
	m.connectionsOpened = 0
	m.connectionsClosed = 0
	m.pipelinesExecuted = 0
	m.retriesAttempted = 0
	m.errorsOccurred = 0

	m.executionTimes = m.executionTimes[:0]
	m.connectionTimes = m.connectionTimes[:0]
	m.pipelineTimes = m.pipelineTimes[:0]
}
