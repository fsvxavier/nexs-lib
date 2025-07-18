package replicas

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// ReplicaManager implementa IReplicaManager
type ReplicaManager struct {
	replicas            map[string]*ReplicaInfo
	loadBalancer        *LoadBalancer
	readPreference      interfaces.ReadPreference
	healthCheckInterval time.Duration
	healthCheckTimeout  time.Duration
	stats               *ReplicaStats

	// Callbacks
	healthChangeCallbacks []func(replica interfaces.IReplicaInfo, oldStatus, newStatus interfaces.ReplicaStatus)
	failoverCallbacks     []func(from, to interfaces.IReplicaInfo)

	// Controle de lifecycle
	running           bool
	ctx               context.Context
	cancel            context.CancelFunc
	healthCheckTicker *time.Ticker

	// Controle de acesso
	mu sync.RWMutex

	// Provider factory para criar pools de conexão
	providerFactory func(dsn string) (interfaces.IPool, error)
}

// ReplicaManagerConfig configuração do gerenciador de réplicas
type ReplicaManagerConfig struct {
	LoadBalancingStrategy interfaces.LoadBalancingStrategy
	ReadPreference        interfaces.ReadPreference
	HealthCheckInterval   time.Duration
	HealthCheckTimeout    time.Duration
	ProviderFactory       func(dsn string) (interfaces.IPool, error)
}

// NewReplicaManager cria um novo gerenciador de réplicas
func NewReplicaManager(config ReplicaManagerConfig) *ReplicaManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ReplicaManager{
		replicas:            make(map[string]*ReplicaInfo),
		loadBalancer:        NewLoadBalancer(config.LoadBalancingStrategy),
		readPreference:      config.ReadPreference,
		healthCheckInterval: config.HealthCheckInterval,
		healthCheckTimeout:  config.HealthCheckTimeout,
		stats:               NewReplicaStats(),
		running:             false,
		ctx:                 ctx,
		cancel:              cancel,
		providerFactory:     config.ProviderFactory,
	}
}

// AddReplica adiciona uma nova réplica
func (rm *ReplicaManager) AddReplica(ctx context.Context, id string, dsn string, weight int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.replicas[id]; exists {
		return fmt.Errorf("replica %s already exists", id)
	}

	replica := NewReplicaInfo(id, dsn, weight)

	// Criar pool de conexões se factory estiver disponível
	if rm.providerFactory != nil {
		pool, err := rm.providerFactory(dsn)
		if err != nil {
			return fmt.Errorf("failed to create pool for replica %s: %w", id, err)
		}
		replica.SetPool(pool)
	}

	rm.replicas[id] = replica
	rm.updateStats()

	return nil
}

// RemoveReplica remove uma réplica
func (rm *ReplicaManager) RemoveReplica(ctx context.Context, id string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	replica, exists := rm.replicas[id]
	if !exists {
		return fmt.Errorf("replica %s not found", id)
	}

	// Fechar conexões da réplica
	if err := replica.Close(); err != nil {
		return fmt.Errorf("failed to close replica %s: %w", id, err)
	}

	delete(rm.replicas, id)
	rm.updateStats()

	return nil
}

// GetReplica retorna uma réplica específica
func (rm *ReplicaManager) GetReplica(id string) (interfaces.IReplicaInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	replica, exists := rm.replicas[id]
	if !exists {
		return nil, fmt.Errorf("replica %s not found", id)
	}

	return replica, nil
}

// ListReplicas retorna todas as réplicas
func (rm *ReplicaManager) ListReplicas() []interfaces.IReplicaInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	replicas := make([]interfaces.IReplicaInfo, 0, len(rm.replicas))
	for _, replica := range rm.replicas {
		replicas = append(replicas, replica)
	}

	return replicas
}

// SelectReplica seleciona uma réplica baseada na preferência
func (rm *ReplicaManager) SelectReplica(ctx context.Context, preference interfaces.ReadPreference) (interfaces.IReplicaInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var candidates []interfaces.IReplicaInfo

	switch preference {
	case interfaces.ReadPreferencePrimary:
		// Para primary, retornar erro se não tiver primary configurado
		return nil, fmt.Errorf("primary replica not configured")

	case interfaces.ReadPreferenceSecondary:
		candidates = rm.getSecondaryReplicas()

	case interfaces.ReadPreferenceSecondaryPreferred:
		candidates = rm.getSecondaryReplicas()
		if len(candidates) == 0 {
			// Fallback para todas as réplicas disponíveis
			candidates = rm.getHealthyReplicas()
		}

	case interfaces.ReadPreferenceNearest:
		candidates = rm.getHealthyReplicas()

	default:
		candidates = rm.getHealthyReplicas()
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no suitable replicas available for preference: %s", preference)
	}

	return rm.loadBalancer.SelectReplica(ctx, candidates)
}

// SelectReplicaWithStrategy seleciona uma réplica com estratégia específica
func (rm *ReplicaManager) SelectReplicaWithStrategy(ctx context.Context, strategy interfaces.LoadBalancingStrategy) (interfaces.IReplicaInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Criar um load balancer temporário com a estratégia desejada
	tempLB := NewLoadBalancer(strategy)
	candidates := rm.getHealthyReplicas()

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no healthy replicas available")
	}

	return tempLB.SelectReplica(ctx, candidates)
}

// HealthCheck executa health check em uma réplica específica
func (rm *ReplicaManager) HealthCheck(ctx context.Context, replicaID string) error {
	rm.mu.RLock()
	replica, exists := rm.replicas[replicaID]
	rm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("replica %s not found", replicaID)
	}

	return rm.performHealthCheck(ctx, replica)
}

// HealthCheckAll executa health check em todas as réplicas
func (rm *ReplicaManager) HealthCheckAll(ctx context.Context) error {
	rm.mu.RLock()
	replicas := make([]*ReplicaInfo, 0, len(rm.replicas))
	for _, replica := range rm.replicas {
		replicas = append(replicas, replica)
	}
	rm.mu.RUnlock()

	var lastErr error
	for _, replica := range replicas {
		if err := rm.performHealthCheck(ctx, replica); err != nil {
			lastErr = err
		}
	}

	rm.updateStats()
	return lastErr
}

// performHealthCheck executa health check em uma réplica
func (rm *ReplicaManager) performHealthCheck(ctx context.Context, replica *ReplicaInfo) error {
	start := time.Now()

	// Criar contexto com timeout
	healthCtx, cancel := context.WithTimeout(ctx, rm.healthCheckTimeout)
	defer cancel()

	pool := replica.GetPool()
	if pool == nil {
		replica.MarkUnhealthy()
		return fmt.Errorf("no pool available for replica %s", replica.GetID())
	}

	// Tentar adquirir conexão
	conn, err := pool.Acquire(healthCtx)
	if err != nil {
		oldStatus := replica.GetStatus()
		replica.MarkUnhealthy()
		rm.notifyHealthChange(replica, oldStatus, interfaces.ReplicaStatusUnhealthy)
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Executar query simples de health check
	err = conn.Ping(healthCtx)
	latency := time.Since(start)

	oldStatus := replica.GetStatus()
	if err != nil {
		replica.MarkUnhealthy()
		replica.RecordQuery(false, latency)
		rm.notifyHealthChange(replica, oldStatus, interfaces.ReplicaStatusUnhealthy)
		return err
	}

	replica.MarkHealthy()
	replica.RecordQuery(true, latency)
	replica.UpdateLatency(latency)

	if oldStatus != interfaces.ReplicaStatusHealthy {
		rm.notifyHealthChange(replica, oldStatus, interfaces.ReplicaStatusHealthy)
	}

	return nil
}

// GetHealthyReplicas retorna réplicas saudáveis
func (rm *ReplicaManager) GetHealthyReplicas() []interfaces.IReplicaInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.getHealthyReplicas()
}

// getHealthyReplicas retorna réplicas saudáveis (thread unsafe)
func (rm *ReplicaManager) getHealthyReplicas() []interfaces.IReplicaInfo {
	var healthy []interfaces.IReplicaInfo
	for _, replica := range rm.replicas {
		if replica.GetStatus() == interfaces.ReplicaStatusHealthy {
			healthy = append(healthy, replica)
		}
	}
	return healthy
}

// GetUnhealthyReplicas retorna réplicas não saudáveis
func (rm *ReplicaManager) GetUnhealthyReplicas() []interfaces.IReplicaInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var unhealthy []interfaces.IReplicaInfo
	for _, replica := range rm.replicas {
		if replica.GetStatus() == interfaces.ReplicaStatusUnhealthy {
			unhealthy = append(unhealthy, replica)
		}
	}
	return unhealthy
}

// getSecondaryReplicas retorna réplicas secundárias (todas por enquanto)
func (rm *ReplicaManager) getSecondaryReplicas() []interfaces.IReplicaInfo {
	return rm.getHealthyReplicas()
}

// SetLoadBalancingStrategy define a estratégia de balanceamento
func (rm *ReplicaManager) SetLoadBalancingStrategy(strategy interfaces.LoadBalancingStrategy) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.loadBalancer.SetStrategy(strategy)
}

// GetLoadBalancingStrategy retorna a estratégia atual
func (rm *ReplicaManager) GetLoadBalancingStrategy() interfaces.LoadBalancingStrategy {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.loadBalancer.GetStrategy()
}

// SetReadPreference define a preferência de leitura
func (rm *ReplicaManager) SetReadPreference(preference interfaces.ReadPreference) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.readPreference = preference
}

// GetReadPreference retorna a preferência atual
func (rm *ReplicaManager) GetReadPreference() interfaces.ReadPreference {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.readPreference
}

// SetHealthCheckInterval define o intervalo de health check
func (rm *ReplicaManager) SetHealthCheckInterval(interval time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.healthCheckInterval = interval

	// Reiniciar ticker se estiver rodando
	if rm.running && rm.healthCheckTicker != nil {
		rm.healthCheckTicker.Stop()
		rm.healthCheckTicker = time.NewTicker(interval)
	}
}

// GetHealthCheckInterval retorna o intervalo atual
func (rm *ReplicaManager) GetHealthCheckInterval() time.Duration {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.healthCheckInterval
}

// SetHealthCheckTimeout define o timeout de health check
func (rm *ReplicaManager) SetHealthCheckTimeout(timeout time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.healthCheckTimeout = timeout
}

// GetHealthCheckTimeout retorna o timeout atual
func (rm *ReplicaManager) GetHealthCheckTimeout() time.Duration {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.healthCheckTimeout
}

// GetStats retorna estatísticas
func (rm *ReplicaManager) GetStats() interfaces.IReplicaStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.stats
}

// GetReplicaStats retorna estatísticas de uma réplica específica
func (rm *ReplicaManager) GetReplicaStats(id string) (interfaces.IReplicaStats, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	replica, exists := rm.replicas[id]
	if !exists {
		return nil, fmt.Errorf("replica %s not found", id)
	}

	// Criar estatísticas específicas da réplica
	stats := NewReplicaStats()
	stats.RecordQuery(id, true, replica.GetAvgLatency())

	return stats, nil
}

// OnReplicaHealthChange registra callback para mudanças de health
func (rm *ReplicaManager) OnReplicaHealthChange(callback func(replica interfaces.IReplicaInfo, oldStatus, newStatus interfaces.ReplicaStatus)) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.healthChangeCallbacks = append(rm.healthChangeCallbacks, callback)
}

// OnReplicaFailover registra callback para failover
func (rm *ReplicaManager) OnReplicaFailover(callback func(from, to interfaces.IReplicaInfo)) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.failoverCallbacks = append(rm.failoverCallbacks, callback)
}

// notifyHealthChange notifica callbacks de mudança de health
func (rm *ReplicaManager) notifyHealthChange(replica interfaces.IReplicaInfo, oldStatus, newStatus interfaces.ReplicaStatus) {
	for _, callback := range rm.healthChangeCallbacks {
		go callback(replica, oldStatus, newStatus)
	}
}

// notifyFailover notifica callbacks de failover
func (rm *ReplicaManager) notifyFailover(from, to interfaces.IReplicaInfo) {
	for _, callback := range rm.failoverCallbacks {
		go callback(from, to)
	}
	rm.stats.RecordFailover()
}

// Start inicia o gerenciador
func (rm *ReplicaManager) Start(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.running {
		return fmt.Errorf("replica manager already running")
	}

	rm.running = true

	// Iniciar health check periódico
	if rm.healthCheckInterval > 0 {
		rm.healthCheckTicker = time.NewTicker(rm.healthCheckInterval)
		go rm.healthCheckLoop()
	}

	return nil
}

// Stop para o gerenciador
func (rm *ReplicaManager) Stop(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.running {
		return fmt.Errorf("replica manager not running")
	}

	rm.running = false
	rm.cancel()

	// Parar health check
	if rm.healthCheckTicker != nil {
		rm.healthCheckTicker.Stop()
		rm.healthCheckTicker = nil
	}

	// Fechar todas as réplicas
	for _, replica := range rm.replicas {
		replica.Close()
	}

	return nil
}

// IsRunning verifica se o gerenciador está rodando
func (rm *ReplicaManager) IsRunning() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.running
}

// SetReplicaMaintenance define modo de manutenção
func (rm *ReplicaManager) SetReplicaMaintenance(id string, maintenance bool) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	replica, exists := rm.replicas[id]
	if !exists {
		return fmt.Errorf("replica %s not found", id)
	}

	oldStatus := replica.GetStatus()
	if maintenance {
		replica.MarkMaintenance()
	} else {
		replica.MarkHealthy()
	}

	newStatus := replica.GetStatus()
	if oldStatus != newStatus {
		rm.notifyHealthChange(replica, oldStatus, newStatus)
	}

	rm.updateStats()
	return nil
}

// DrainReplica drena uma réplica gradualmente
func (rm *ReplicaManager) DrainReplica(ctx context.Context, id string, timeout time.Duration) error {
	rm.mu.Lock()
	replica, exists := rm.replicas[id]
	rm.mu.Unlock()

	if !exists {
		return fmt.Errorf("replica %s not found", id)
	}

	// Marcar como em manutenção
	replica.MarkMaintenance()

	// Aguardar conexões terminarem ou timeout
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if replica.GetConnectionCount() == 0 {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("replica %s still has active connections after timeout", id)
}

// healthCheckLoop executa health check periódico
func (rm *ReplicaManager) healthCheckLoop() {
	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-rm.healthCheckTicker.C:
			rm.HealthCheckAll(rm.ctx)
		}
	}
}

// updateStats atualiza estatísticas gerais
func (rm *ReplicaManager) updateStats() {
	var healthy, unhealthy, recovering, maintenance int

	for _, replica := range rm.replicas {
		switch replica.GetStatus() {
		case interfaces.ReplicaStatusHealthy:
			healthy++
		case interfaces.ReplicaStatusUnhealthy:
			unhealthy++
		case interfaces.ReplicaStatusRecovering:
			recovering++
		case interfaces.ReplicaStatusMaintenance:
			maintenance++
		}
	}

	rm.stats.UpdateReplicaCount(len(rm.replicas), healthy, unhealthy, recovering, maintenance)
}
