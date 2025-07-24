package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// RetryManager implementa IRetryManager com padrões robustos
type RetryManager struct {
	config interfaces.RetryConfig
	stats  interfaces.RetryStats
	mu     sync.RWMutex
}

// NewRetryManager cria um novo retry manager
func NewRetryManager(config interfaces.RetryConfig) interfaces.IRetryManager {
	return &RetryManager{
		config: config,
		stats:  interfaces.RetryStats{},
	}
}

// Execute executa uma operação com retry
func (rm *RetryManager) Execute(ctx context.Context, operation func() error) error {
	rm.mu.Lock()
	rm.stats.TotalAttempts++
	rm.mu.Unlock()

	var lastErr error
	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		if attempt > 0 {
			rm.mu.Lock()
			rm.stats.TotalRetries++
			rm.mu.Unlock()

			// Calcular backoff
			backoff := rm.calculateBackoff(attempt)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				rm.mu.Lock()
				rm.stats.FailedOps++
				rm.mu.Unlock()
				return ctx.Err()
			}
		}

		err := operation()
		if err == nil {
			rm.mu.Lock()
			rm.stats.SuccessfulOps++
			rm.updateAverageRetries()
			rm.mu.Unlock()
			return nil
		}

		lastErr = err

		// Verificar se é erro recuperável
		if !rm.isRetryableError(err) {
			break
		}
	}

	rm.mu.Lock()
	rm.stats.FailedOps++
	rm.stats.LastRetryTime = time.Now()
	rm.mu.Unlock()

	return lastErr
}

// ExecuteWithConn executa uma operação com conexão e retry
func (rm *RetryManager) ExecuteWithConn(ctx context.Context, pool interfaces.IPool, operation func(conn interfaces.IConn) error) error {
	return rm.Execute(ctx, func() error {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		return operation(conn)
	})
}

// UpdateConfig atualiza configuração de retry
func (rm *RetryManager) UpdateConfig(config interfaces.RetryConfig) error {
	if err := rm.validateConfig(config); err != nil {
		return err
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.config = config
	return nil
}

// GetStats retorna estatísticas de retry
func (rm *RetryManager) GetStats() interfaces.RetryStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.stats
}

// calculateBackoff calcula o tempo de backoff
func (rm *RetryManager) calculateBackoff(attempt int) time.Duration {
	duration := rm.config.InitialInterval
	for i := 1; i < attempt; i++ {
		duration = time.Duration(float64(duration) * rm.config.Multiplier)
		if duration > rm.config.MaxInterval {
			duration = rm.config.MaxInterval
			break
		}
	}

	if rm.config.RandomizeWait {
		jitterFactor := 0.5 + (float64(time.Now().UnixNano()%1000) / 1000.0)
		duration = time.Duration(float64(duration) * jitterFactor)
	}

	return duration
}

// isRetryableError verifica se erro é recuperável
func (rm *RetryManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Erros de contexto não são recuperáveis
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Verificar interface Temporary
	if tempErr, ok := err.(interface{ Temporary() bool }); ok {
		return tempErr.Temporary()
	}

	// Padrões de erro recuperáveis
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"temporary failure",
		"network is unreachable",
		"no route to host",
		"too many connections",
		"connection pool exhausted",
	}

	for _, pattern := range retryablePatterns {
		if containsPattern(errStr, pattern) {
			return true
		}
	}

	return false
}

// validateConfig valida configuração
func (rm *RetryManager) validateConfig(config interfaces.RetryConfig) error {
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	if config.InitialInterval <= 0 {
		return fmt.Errorf("initial interval must be positive")
	}
	if config.MaxInterval <= 0 {
		return fmt.Errorf("max interval must be positive")
	}
	if config.Multiplier <= 1.0 {
		return fmt.Errorf("multiplier must be greater than 1.0")
	}
	return nil
}

// updateAverageRetries atualiza média de retries - deve ser chamado com lock
func (rm *RetryManager) updateAverageRetries() {
	if rm.stats.TotalAttempts > 0 {
		rm.stats.AverageRetries = float64(rm.stats.TotalRetries) / float64(rm.stats.TotalAttempts)
	}
}

// FailoverManager implementa IFailoverManager
type FailoverManager struct {
	config       interfaces.FailoverConfig
	stats        interfaces.FailoverStats
	healthyNodes map[string]bool
	currentNode  string
	mu           sync.RWMutex
}

// NewFailoverManager cria um novo failover manager
func NewFailoverManager(config interfaces.FailoverConfig) interfaces.IFailoverManager {
	healthyNodes := make(map[string]bool)
	for _, node := range config.FallbackNodes {
		healthyNodes[node] = true
	}

	return &FailoverManager{
		config:       config,
		stats:        interfaces.FailoverStats{},
		healthyNodes: healthyNodes,
		currentNode:  "",
	}
}

// Execute executa operação com failover
func (fm *FailoverManager) Execute(ctx context.Context, operation func(conn interfaces.IConn) error) error {
	fm.mu.Lock()
	fm.stats.TotalFailovers++
	fm.mu.Unlock()

	// Implementação simplificada - em produção seria mais complexa
	// Precisaria gerenciar múltiplos pools de conexão

	fm.mu.Lock()
	fm.stats.FailedFailovers++
	fm.mu.Unlock()

	return fmt.Errorf("failover manager requires connection pool management - not fully implemented")
}

// MarkNodeDown marca um nó como inativo
func (fm *FailoverManager) MarkNodeDown(nodeID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.healthyNodes == nil {
		fm.healthyNodes = make(map[string]bool)
	}

	fm.healthyNodes[nodeID] = false
	fm.updateDownNodes()
	return nil
}

// MarkNodeUp marca um nó como ativo
func (fm *FailoverManager) MarkNodeUp(nodeID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.healthyNodes == nil {
		fm.healthyNodes = make(map[string]bool)
	}

	fm.healthyNodes[nodeID] = true
	fm.updateDownNodes()
	return nil
}

// GetHealthyNodes retorna nós saudáveis
func (fm *FailoverManager) GetHealthyNodes() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var healthy []string
	for node, isHealthy := range fm.healthyNodes {
		if isHealthy {
			healthy = append(healthy, node)
		}
	}
	return healthy
}

// GetUnhealthyNodes retorna nós não saudáveis
func (fm *FailoverManager) GetUnhealthyNodes() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var unhealthy []string
	for node, isHealthy := range fm.healthyNodes {
		if !isHealthy {
			unhealthy = append(unhealthy, node)
		}
	}
	return unhealthy
}

// GetStats retorna estatísticas de failover
func (fm *FailoverManager) GetStats() interfaces.FailoverStats {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.stats
}

// updateDownNodes atualiza lista de nós inativos - deve ser chamado com lock
func (fm *FailoverManager) updateDownNodes() {
	fm.stats.DownNodes = []string{}
	for node, isHealthy := range fm.healthyNodes {
		if !isHealthy {
			fm.stats.DownNodes = append(fm.stats.DownNodes, node)
		}
	}
}

// containsPattern verifica se string contém padrão
func containsPattern(str, pattern string) bool {
	return len(str) >= len(pattern) && indexOfPattern(str, pattern) >= 0
}

// indexOfPattern procura padrão na string
func indexOfPattern(str, pattern string) int {
	if len(pattern) == 0 {
		return 0
	}
	if len(str) < len(pattern) {
		return -1
	}

	for i := 0; i <= len(str)-len(pattern); i++ {
		if str[i:i+len(pattern)] == pattern {
			return i
		}
	}
	return -1
}
