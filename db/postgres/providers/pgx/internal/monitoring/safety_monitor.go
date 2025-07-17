package monitoring

import (
	"runtime"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// SafetyMonitor implementa ISafetyMonitor para monitoramento de thread-safety
type SafetyMonitor struct {
	mu                sync.RWMutex
	deadlocks         []interfaces.DeadlockInfo
	raceConditions    []interfaces.RaceConditionInfo
	leaks             []interfaces.LeakInfo
	startTime         time.Time
	goroutineCount    int
	lastHealthCheck   time.Time
	healthCheckTicker *time.Ticker
	stopChan          chan struct{}
}

// NewSafetyMonitor cria um novo monitor de segurança
func NewSafetyMonitor() interfaces.ISafetyMonitor {
	monitor := &SafetyMonitor{
		deadlocks:       []interfaces.DeadlockInfo{},
		raceConditions:  []interfaces.RaceConditionInfo{},
		leaks:           []interfaces.LeakInfo{},
		startTime:       time.Now(),
		goroutineCount:  runtime.NumGoroutine(),
		lastHealthCheck: time.Now(),
		stopChan:        make(chan struct{}),
	}

	// Iniciar monitoramento periódico
	monitor.healthCheckTicker = time.NewTicker(30 * time.Second)
	go monitor.monitoringLoop()

	return monitor
}

// CheckDeadlocks verifica deadlocks
func (sm *SafetyMonitor) CheckDeadlocks() []interfaces.DeadlockInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Fazer cópia para evitar race conditions
	deadlocks := make([]interfaces.DeadlockInfo, len(sm.deadlocks))
	copy(deadlocks, sm.deadlocks)
	return deadlocks
}

// CheckRaceConditions verifica race conditions
func (sm *SafetyMonitor) CheckRaceConditions() []interfaces.RaceConditionInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Fazer cópia para evitar race conditions
	raceConditions := make([]interfaces.RaceConditionInfo, len(sm.raceConditions))
	copy(raceConditions, sm.raceConditions)
	return raceConditions
}

// CheckLeaks verifica vazamentos de recursos
func (sm *SafetyMonitor) CheckLeaks() []interfaces.LeakInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Fazer cópia para evitar race conditions
	leaks := make([]interfaces.LeakInfo, len(sm.leaks))
	copy(leaks, sm.leaks)
	return leaks
}

// IsHealthy verifica se o sistema está saudável
func (sm *SafetyMonitor) IsHealthy() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Verificar se há problemas críticos
	if len(sm.deadlocks) > 0 {
		return false
	}

	// Verificar se há muitas race conditions
	if len(sm.raceConditions) > 10 {
		return false
	}

	// Verificar se há vazamentos críticos
	for _, leak := range sm.leaks {
		if leak.Count > 1000 { // Limite arbitrário
			return false
		}
	}

	// Verificar se o número de goroutines está crescendo descontroladamente
	currentGoroutines := runtime.NumGoroutine()
	if currentGoroutines > sm.goroutineCount*3 { // 3x o valor inicial
		return false
	}

	return true
}

// Close para o monitoramento
func (sm *SafetyMonitor) Close() {
	if sm.healthCheckTicker != nil {
		sm.healthCheckTicker.Stop()
	}
	close(sm.stopChan)
}

// monitoringLoop executa monitoramento periódico
func (sm *SafetyMonitor) monitoringLoop() {
	for {
		select {
		case <-sm.healthCheckTicker.C:
			sm.performHealthCheck()
		case <-sm.stopChan:
			return
		}
	}
}

// performHealthCheck executa verificação de saúde
func (sm *SafetyMonitor) performHealthCheck() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.lastHealthCheck = time.Now()

	// Verificar crescimento de goroutines
	currentGoroutines := runtime.NumGoroutine()
	if currentGoroutines > sm.goroutineCount*2 {
		sm.leaks = append(sm.leaks, interfaces.LeakInfo{
			Timestamp: time.Now(),
			Resource:  "goroutines",
			Count:     int64(currentGoroutines - sm.goroutineCount),
			Details:   "Potential goroutine leak detected",
		})
	}

	// Atualizar contador de goroutines
	sm.goroutineCount = currentGoroutines

	// Limpeza de registros antigos (manter apenas últimos 100 de cada tipo)
	if len(sm.deadlocks) > 100 {
		sm.deadlocks = sm.deadlocks[len(sm.deadlocks)-100:]
	}
	if len(sm.raceConditions) > 100 {
		sm.raceConditions = sm.raceConditions[len(sm.raceConditions)-100:]
	}
	if len(sm.leaks) > 100 {
		sm.leaks = sm.leaks[len(sm.leaks)-100:]
	}
}

// reportDeadlock registra um deadlock
func (sm *SafetyMonitor) reportDeadlock(goroutines []string, stackTraces map[string]string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.deadlocks = append(sm.deadlocks, interfaces.DeadlockInfo{
		Timestamp:   time.Now(),
		Goroutines:  goroutines,
		StackTraces: stackTraces,
	})
}

// reportRaceCondition registra uma race condition
func (sm *SafetyMonitor) reportRaceCondition(location, details string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.raceConditions = append(sm.raceConditions, interfaces.RaceConditionInfo{
		Timestamp: time.Now(),
		Location:  location,
		Details:   details,
	})
}

// reportLeak registra um vazamento de recurso
func (sm *SafetyMonitor) reportLeak(resource string, count int64, details string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.leaks = append(sm.leaks, interfaces.LeakInfo{
		Timestamp: time.Now(),
		Resource:  resource,
		Count:     count,
		Details:   details,
	})
}

// ConnectionMonitor monitora estatísticas de conexões
type ConnectionMonitor struct {
	mu           sync.RWMutex
	stats        interfaces.ConnectionStats
	startTime    time.Time
	lastActivity time.Time
}

// NewConnectionMonitor cria um novo monitor de conexões
func NewConnectionMonitor() *ConnectionMonitor {
	now := time.Now()
	return &ConnectionMonitor{
		stats: interfaces.ConnectionStats{
			CreatedAt:    now,
			LastActivity: now,
		},
		startTime:    now,
		lastActivity: now,
	}
}

// UpdateStats atualiza estatísticas de conexão
func (cm *ConnectionMonitor) UpdateStats(operation string, duration time.Duration, err error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.lastActivity = time.Now()
	cm.stats.LastActivity = cm.lastActivity

	switch operation {
	case "query":
		cm.stats.TotalQueries++
		if err != nil {
			cm.stats.FailedQueries++
		}
		cm.updateAverageQueryTime(duration)
	case "exec":
		cm.stats.TotalExecs++
		if err != nil {
			cm.stats.FailedExecs++
		}
		cm.updateAverageExecTime(duration)
	case "transaction":
		cm.stats.TotalTransactions++
		if err != nil {
			cm.stats.FailedTransactions++
		}
	case "batch":
		cm.stats.TotalBatches++
	}
}

// GetStats retorna estatísticas atuais
func (cm *ConnectionMonitor) GetStats() interfaces.ConnectionStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.stats
}

// updateAverageQueryTime atualiza tempo médio de query - deve ser chamado com lock
func (cm *ConnectionMonitor) updateAverageQueryTime(duration time.Duration) {
	if cm.stats.TotalQueries == 1 {
		cm.stats.AverageQueryTime = duration
	} else {
		// Média móvel simples
		cm.stats.AverageQueryTime = (cm.stats.AverageQueryTime + duration) / 2
	}
}

// updateAverageExecTime atualiza tempo médio de exec - deve ser chamado com lock
func (cm *ConnectionMonitor) updateAverageExecTime(duration time.Duration) {
	if cm.stats.TotalExecs == 1 {
		cm.stats.AverageExecTime = duration
	} else {
		// Média móvel simples
		cm.stats.AverageExecTime = (cm.stats.AverageExecTime + duration) / 2
	}
}
