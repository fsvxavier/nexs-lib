package replicas

import (
	"encoding/json"
	"sync"
	"time"
)

// ReplicaStats implementa IReplicaStats
type ReplicaStats struct {
	// Contadores básicos
	totalReplicas       int
	healthyReplicas     int
	unhealthyReplicas   int
	recoveringReplicas  int
	maintenanceReplicas int

	// Métricas de queries
	totalQueries      int64
	successfulQueries int64
	failedQueries     int64

	// Métricas de latência
	avgLatency time.Duration
	maxLatency time.Duration
	minLatency time.Duration

	// Distribuições
	queryDistribution   map[string]int64
	latencyDistribution map[string]time.Duration
	errorDistribution   map[string]int64

	// Failover
	failoverCount    int64
	lastFailoverTime time.Time
	uptime           time.Duration
	startTime        time.Time

	// Controle de acesso
	mu sync.RWMutex
}

// NewReplicaStats cria uma nova instância de estatísticas
func NewReplicaStats() *ReplicaStats {
	return &ReplicaStats{
		queryDistribution:   make(map[string]int64),
		latencyDistribution: make(map[string]time.Duration),
		errorDistribution:   make(map[string]int64),
		startTime:           time.Now(),
		minLatency:          time.Duration(^uint64(0) >> 1), // Max duration
	}
}

// GetTotalReplicas retorna o total de réplicas
func (rs *ReplicaStats) GetTotalReplicas() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.totalReplicas
}

// GetHealthyReplicas retorna o número de réplicas saudáveis
func (rs *ReplicaStats) GetHealthyReplicas() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.healthyReplicas
}

// GetUnhealthyReplicas retorna o número de réplicas não saudáveis
func (rs *ReplicaStats) GetUnhealthyReplicas() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.unhealthyReplicas
}

// GetRecoveringReplicas retorna o número de réplicas em recuperação
func (rs *ReplicaStats) GetRecoveringReplicas() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.recoveringReplicas
}

// GetMaintenanceReplicas retorna o número de réplicas em manutenção
func (rs *ReplicaStats) GetMaintenanceReplicas() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.maintenanceReplicas
}

// GetTotalQueries retorna o total de queries
func (rs *ReplicaStats) GetTotalQueries() int64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.totalQueries
}

// GetSuccessfulQueries retorna o número de queries bem-sucedidas
func (rs *ReplicaStats) GetSuccessfulQueries() int64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.successfulQueries
}

// GetFailedQueries retorna o número de queries falhadas
func (rs *ReplicaStats) GetFailedQueries() int64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.failedQueries
}

// GetAvgLatency retorna a latência média
func (rs *ReplicaStats) GetAvgLatency() time.Duration {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.avgLatency
}

// GetMaxLatency retorna a latência máxima
func (rs *ReplicaStats) GetMaxLatency() time.Duration {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.maxLatency
}

// GetMinLatency retorna a latência mínima
func (rs *ReplicaStats) GetMinLatency() time.Duration {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.minLatency
}

// GetQueryDistribution retorna a distribuição de queries por réplica
func (rs *ReplicaStats) GetQueryDistribution() map[string]int64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	dist := make(map[string]int64)
	for k, v := range rs.queryDistribution {
		dist[k] = v
	}
	return dist
}

// GetLatencyDistribution retorna a distribuição de latência por réplica
func (rs *ReplicaStats) GetLatencyDistribution() map[string]time.Duration {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	dist := make(map[string]time.Duration)
	for k, v := range rs.latencyDistribution {
		dist[k] = v
	}
	return dist
}

// GetErrorDistribution retorna a distribuição de erros por réplica
func (rs *ReplicaStats) GetErrorDistribution() map[string]int64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	dist := make(map[string]int64)
	for k, v := range rs.errorDistribution {
		dist[k] = v
	}
	return dist
}

// GetFailoverCount retorna o número de failovers
func (rs *ReplicaStats) GetFailoverCount() int64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.failoverCount
}

// GetLastFailoverTime retorna o tempo do último failover
func (rs *ReplicaStats) GetLastFailoverTime() time.Time {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.lastFailoverTime
}

// GetUptime retorna o tempo de atividade
func (rs *ReplicaStats) GetUptime() time.Duration {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return time.Since(rs.startTime)
}

// GetQueriesPerSecond retorna queries por segundo
func (rs *ReplicaStats) GetQueriesPerSecond() float64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	uptime := time.Since(rs.startTime)
	if uptime == 0 {
		return 0
	}

	return float64(rs.totalQueries) / uptime.Seconds()
}

// GetErrorsPerSecond retorna erros por segundo
func (rs *ReplicaStats) GetErrorsPerSecond() float64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	uptime := time.Since(rs.startTime)
	if uptime == 0 {
		return 0
	}

	return float64(rs.failedQueries) / uptime.Seconds()
}

// GetLatencyPercentile retorna o percentil de latência
func (rs *ReplicaStats) GetLatencyPercentile(percentile float64) time.Duration {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	// Implementação simplificada - em produção usaria uma estrutura mais eficiente
	if percentile <= 0 {
		return rs.minLatency
	}
	if percentile >= 1.0 {
		return rs.maxLatency
	}

	// Aproximação linear entre min e max
	diff := rs.maxLatency - rs.minLatency
	return rs.minLatency + time.Duration(float64(diff)*percentile)
}

// UpdateReplicaCount atualiza contadores de réplicas
func (rs *ReplicaStats) UpdateReplicaCount(total, healthy, unhealthy, recovering, maintenance int) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.totalReplicas = total
	rs.healthyReplicas = healthy
	rs.unhealthyReplicas = unhealthy
	rs.recoveringReplicas = recovering
	rs.maintenanceReplicas = maintenance
}

// RecordQuery registra uma query executada
func (rs *ReplicaStats) RecordQuery(replicaID string, success bool, latency time.Duration) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.totalQueries++
	rs.queryDistribution[replicaID]++

	if success {
		rs.successfulQueries++
	} else {
		rs.failedQueries++
		rs.errorDistribution[replicaID]++
	}

	// Atualizar latência
	rs.updateLatency(replicaID, latency)
}

// updateLatency atualiza métricas de latência
func (rs *ReplicaStats) updateLatency(replicaID string, latency time.Duration) {
	// Atualizar latência por réplica
	rs.latencyDistribution[replicaID] = latency

	// Atualizar min/max/avg globais
	if latency < rs.minLatency {
		rs.minLatency = latency
	}
	if latency > rs.maxLatency {
		rs.maxLatency = latency
	}

	// Calcular média móvel simples
	if rs.avgLatency == 0 {
		rs.avgLatency = latency
	} else {
		rs.avgLatency = time.Duration((int64(rs.avgLatency) + int64(latency)) / 2)
	}
}

// RecordFailover registra um failover
func (rs *ReplicaStats) RecordFailover() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.failoverCount++
	rs.lastFailoverTime = time.Now()
}

// ToMap converte estatísticas para map
func (rs *ReplicaStats) ToMap() map[string]interface{} {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	return map[string]interface{}{
		"total_replicas":       rs.totalReplicas,
		"healthy_replicas":     rs.healthyReplicas,
		"unhealthy_replicas":   rs.unhealthyReplicas,
		"recovering_replicas":  rs.recoveringReplicas,
		"maintenance_replicas": rs.maintenanceReplicas,
		"total_queries":        rs.totalQueries,
		"successful_queries":   rs.successfulQueries,
		"failed_queries":       rs.failedQueries,
		"avg_latency_ms":       rs.avgLatency.Milliseconds(),
		"max_latency_ms":       rs.maxLatency.Milliseconds(),
		"min_latency_ms":       rs.minLatency.Milliseconds(),
		"failover_count":       rs.failoverCount,
		"uptime_seconds":       time.Since(rs.startTime).Seconds(),
		"queries_per_second":   rs.GetQueriesPerSecond(),
		"errors_per_second":    rs.GetErrorsPerSecond(),
		"query_distribution":   rs.queryDistribution,
		"error_distribution":   rs.errorDistribution,
	}
}

// ToJSON converte estatísticas para JSON
func (rs *ReplicaStats) ToJSON() ([]byte, error) {
	data := rs.ToMap()
	return json.Marshal(data)
}

// Reset reseta todas as estatísticas
func (rs *ReplicaStats) Reset() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.totalReplicas = 0
	rs.healthyReplicas = 0
	rs.unhealthyReplicas = 0
	rs.recoveringReplicas = 0
	rs.maintenanceReplicas = 0
	rs.totalQueries = 0
	rs.successfulQueries = 0
	rs.failedQueries = 0
	rs.avgLatency = 0
	rs.maxLatency = 0
	rs.minLatency = time.Duration(^uint64(0) >> 1)
	rs.failoverCount = 0
	rs.lastFailoverTime = time.Time{}
	rs.startTime = time.Now()

	// Limpar maps
	rs.queryDistribution = make(map[string]int64)
	rs.latencyDistribution = make(map[string]time.Duration)
	rs.errorDistribution = make(map[string]int64)
}
