package replicas

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// ReplicaInfo implementa IReplicaInfo
type ReplicaInfo struct {
	id              string
	dsn             string
	weight          int
	status          interfaces.ReplicaStatus
	lastHealthCheck time.Time
	connectionCount int32
	maxConnections  int
	region          string
	tags            map[string]string

	// Métricas de performance
	latency           time.Duration
	totalQueries      int64
	failedQueries     int64
	successfulQueries int64

	// Controle de acesso
	mu sync.RWMutex

	// Pool de conexões para esta réplica
	pool interfaces.IPool

	// Contexto para cancelamento
	ctx    context.Context
	cancel context.CancelFunc
}

// NewReplicaInfo cria uma nova instância de ReplicaInfo
func NewReplicaInfo(id, dsn string, weight int) *ReplicaInfo {
	ctx, cancel := context.WithCancel(context.Background())

	return &ReplicaInfo{
		id:              id,
		dsn:             dsn,
		weight:          weight,
		status:          interfaces.ReplicaStatusHealthy,
		lastHealthCheck: time.Now(),
		maxConnections:  10,
		region:          "default",
		tags:            make(map[string]string),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// GetID retorna o ID da réplica
func (r *ReplicaInfo) GetID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

// GetDSN retorna o DSN da réplica
func (r *ReplicaInfo) GetDSN() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.dsn
}

// GetWeight retorna o peso da réplica
func (r *ReplicaInfo) GetWeight() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.weight
}

// GetStatus retorna o status da réplica
func (r *ReplicaInfo) GetStatus() interfaces.ReplicaStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status
}

// GetLatency retorna a latência da réplica
func (r *ReplicaInfo) GetLatency() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.latency
}

// GetLastHealthCheck retorna o tempo do último health check
func (r *ReplicaInfo) GetLastHealthCheck() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastHealthCheck
}

// GetConnectionCount retorna o número atual de conexões
func (r *ReplicaInfo) GetConnectionCount() int {
	return int(atomic.LoadInt32(&r.connectionCount))
}

// GetMaxConnections retorna o número máximo de conexões
func (r *ReplicaInfo) GetMaxConnections() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.maxConnections
}

// GetRegion retorna a região da réplica
func (r *ReplicaInfo) GetRegion() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.region
}

// GetTags retorna as tags da réplica
func (r *ReplicaInfo) GetTags() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make(map[string]string)
	for k, v := range r.tags {
		tags[k] = v
	}
	return tags
}

// GetSuccessRate retorna a taxa de sucesso
func (r *ReplicaInfo) GetSuccessRate() float64 {
	total := atomic.LoadInt64(&r.totalQueries)
	if total == 0 {
		return 100.0
	}

	successful := atomic.LoadInt64(&r.successfulQueries)
	return float64(successful) / float64(total) * 100.0
}

// GetErrorRate retorna a taxa de erro
func (r *ReplicaInfo) GetErrorRate() float64 {
	total := atomic.LoadInt64(&r.totalQueries)
	if total == 0 {
		return 0.0
	}

	failed := atomic.LoadInt64(&r.failedQueries)
	return float64(failed) / float64(total) * 100.0
}

// GetAvgLatency retorna a latência média
func (r *ReplicaInfo) GetAvgLatency() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.latency
}

// GetTotalQueries retorna o total de queries
func (r *ReplicaInfo) GetTotalQueries() int64 {
	return atomic.LoadInt64(&r.totalQueries)
}

// GetFailedQueries retorna o total de queries falhadas
func (r *ReplicaInfo) GetFailedQueries() int64 {
	return atomic.LoadInt64(&r.failedQueries)
}

// IsAvailable verifica se a réplica está disponível
func (r *ReplicaInfo) IsAvailable() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.status == interfaces.ReplicaStatusHealthy &&
		r.GetConnectionCount() < r.maxConnections
}

// MarkHealthy marca a réplica como saudável
func (r *ReplicaInfo) MarkHealthy() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.status = interfaces.ReplicaStatusHealthy
	r.lastHealthCheck = time.Now()
}

// MarkUnhealthy marca a réplica como não saudável
func (r *ReplicaInfo) MarkUnhealthy() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.status = interfaces.ReplicaStatusUnhealthy
	r.lastHealthCheck = time.Now()
}

// MarkRecovering marca a réplica como recuperando
func (r *ReplicaInfo) MarkRecovering() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.status = interfaces.ReplicaStatusRecovering
	r.lastHealthCheck = time.Now()
}

// MarkMaintenance marca a réplica como em manutenção
func (r *ReplicaInfo) MarkMaintenance() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.status = interfaces.ReplicaStatusMaintenance
	r.lastHealthCheck = time.Now()
}

// SetWeight define o peso da réplica
func (r *ReplicaInfo) SetWeight(weight int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.weight = weight
}

// SetMaxConnections define o número máximo de conexões
func (r *ReplicaInfo) SetMaxConnections(max int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.maxConnections = max
}

// SetTags define as tags da réplica
func (r *ReplicaInfo) SetTags(tags map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tags = make(map[string]string)
	for k, v := range tags {
		r.tags[k] = v
	}
}

// SetPool define o pool de conexões para esta réplica
func (r *ReplicaInfo) SetPool(pool interfaces.IPool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pool = pool
}

// GetPool retorna o pool de conexões
func (r *ReplicaInfo) GetPool() interfaces.IPool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.pool
}

// UpdateLatency atualiza a latência medida
func (r *ReplicaInfo) UpdateLatency(latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Média móvel simples para suavizar oscilações
	if r.latency == 0 {
		r.latency = latency
	} else {
		r.latency = time.Duration((int64(r.latency) + int64(latency)) / 2)
	}
}

// IncrementConnections incrementa o contador de conexões
func (r *ReplicaInfo) IncrementConnections() {
	atomic.AddInt32(&r.connectionCount, 1)
}

// DecrementConnections decrementa o contador de conexões
func (r *ReplicaInfo) DecrementConnections() {
	atomic.AddInt32(&r.connectionCount, -1)
}

// RecordQuery registra uma query executada
func (r *ReplicaInfo) RecordQuery(success bool, latency time.Duration) {
	atomic.AddInt64(&r.totalQueries, 1)

	if success {
		atomic.AddInt64(&r.successfulQueries, 1)
	} else {
		atomic.AddInt64(&r.failedQueries, 1)
	}

	r.UpdateLatency(latency)
}

// Close fecha a réplica e libera recursos
func (r *ReplicaInfo) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cancel()

	if r.pool != nil {
		r.pool.Close()
	}

	return nil
}
