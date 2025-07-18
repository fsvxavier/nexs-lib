package replicas

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// ReplicaPool implementa IReplicaPool
type ReplicaPool struct {
	// Pool principal (master/primary)
	primaryPool interfaces.IPool

	// Gerenciador de réplicas
	replicaManager interfaces.IReplicaManager

	// Preferência de leitura
	readPreference interfaces.ReadPreference

	// Estatísticas
	readStats  interfaces.IReplicaStats
	writeStats interfaces.IReplicaStats

	// Controle de acesso
	mu sync.RWMutex

	// Status
	closed bool
}

// NewReplicaPool cria um novo pool de réplicas
func NewReplicaPool(primaryPool interfaces.IPool, replicaManager interfaces.IReplicaManager) *ReplicaPool {
	return &ReplicaPool{
		primaryPool:    primaryPool,
		replicaManager: replicaManager,
		readPreference: interfaces.ReadPreferenceSecondaryPreferred,
		readStats:      NewReplicaStats(),
		writeStats:     NewReplicaStats(),
	}
}

// GetReplicaManager retorna o gerenciador de réplicas
func (rp *ReplicaPool) GetReplicaManager() interfaces.IReplicaManager {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.replicaManager
}

// SetReadPreference define a preferência de leitura
func (rp *ReplicaPool) SetReadPreference(preference interfaces.ReadPreference) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.readPreference = preference
}

// GetReadPreference retorna a preferência de leitura
func (rp *ReplicaPool) GetReadPreference() interfaces.ReadPreference {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.readPreference
}

// AcquireRead adquire uma conexão de leitura
func (rp *ReplicaPool) AcquireRead(ctx context.Context, preference interfaces.ReadPreference) (interfaces.IConn, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if rp.closed {
		return nil, fmt.Errorf("replica pool is closed")
	}

	start := time.Now()

	// Tentar selecionar uma réplica
	replica, err := rp.replicaManager.SelectReplica(ctx, preference)
	if err != nil {
		// Fallback para o primary se configurado
		if preference == interfaces.ReadPreferenceSecondaryPreferred {
			return rp.acquireFromPrimary(ctx, start)
		}
		return nil, fmt.Errorf("failed to select replica: %w", err)
	}

	// Adquirir conexão da réplica
	conn, err := rp.acquireFromReplica(ctx, replica, start)
	if err != nil {
		// Fallback para o primary se a réplica falhar
		if preference == interfaces.ReadPreferenceSecondaryPreferred {
			return rp.acquireFromPrimary(ctx, start)
		}
		return nil, err
	}

	return conn, nil
}

// AcquireWrite adquire uma conexão de escrita
func (rp *ReplicaPool) AcquireWrite(ctx context.Context) (interfaces.IConn, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	if rp.closed {
		return nil, fmt.Errorf("replica pool is closed")
	}

	start := time.Now()
	return rp.acquireFromPrimary(ctx, start)
}

// acquireFromReplica adquire conexão de uma réplica específica
func (rp *ReplicaPool) acquireFromReplica(ctx context.Context, replica interfaces.IReplicaInfo, start time.Time) (interfaces.IConn, error) {
	// Incrementar contador de conexões
	replicaInfo, ok := replica.(*ReplicaInfo)
	if ok {
		replicaInfo.IncrementConnections()
		defer func() {
			if !ok {
				replicaInfo.DecrementConnections()
			}
		}()
	}

	// Obter pool da réplica
	pool := replica.GetPool()
	if pool == nil {
		return nil, fmt.Errorf("replica %s has no pool", replica.GetID())
	}

	// Adquirir conexão
	conn, err := pool.Acquire(ctx)
	if err != nil {
		// Registrar erro na réplica
		if replicaInfo != nil {
			replicaInfo.RecordQuery(false, time.Since(start))
		}
		return nil, fmt.Errorf("failed to acquire connection from replica %s: %w", replica.GetID(), err)
	}

	// Registrar sucesso na réplica
	if replicaInfo != nil {
		replicaInfo.RecordQuery(true, time.Since(start))
	}

	// Registrar nas estatísticas de leitura
	if readStats, ok := rp.readStats.(*ReplicaStats); ok {
		readStats.RecordQuery(replica.GetID(), true, time.Since(start))
	}

	// Envolver conexão para rastrear liberação
	return &ReplicaConn{
		IConn:       conn,
		replicaInfo: replicaInfo,
	}, nil
}

// acquireFromPrimary adquire conexão do primary
func (rp *ReplicaPool) acquireFromPrimary(ctx context.Context, start time.Time) (interfaces.IConn, error) {
	if rp.primaryPool == nil {
		return nil, fmt.Errorf("no primary pool configured")
	}

	conn, err := rp.primaryPool.Acquire(ctx)
	if err != nil {
		if writeStats, ok := rp.writeStats.(*ReplicaStats); ok {
			writeStats.RecordQuery("primary", false, time.Since(start))
		}
		return nil, fmt.Errorf("failed to acquire connection from primary: %w", err)
	}

	if writeStats, ok := rp.writeStats.(*ReplicaStats); ok {
		writeStats.RecordQuery("primary", true, time.Since(start))
	}
	return conn, nil
}

// GetReadStats retorna estatísticas de leitura
func (rp *ReplicaPool) GetReadStats() interfaces.IReplicaStats {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.readStats
}

// GetWriteStats retorna estatísticas de escrita
func (rp *ReplicaPool) GetWriteStats() interfaces.IReplicaStats {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.writeStats
}

// Implementação da interface IPool

// Acquire adquire uma conexão (usa preferência padrão)
func (rp *ReplicaPool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	return rp.AcquireRead(ctx, rp.readPreference)
}

// Release libera uma conexão (não usado diretamente)
func (rp *ReplicaPool) Release(conn interfaces.IConn) {
	conn.Release()
}

// Close fecha o pool
func (rp *ReplicaPool) Close() {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if rp.closed {
		return
	}

	rp.closed = true

	// Fechar primary pool
	if rp.primaryPool != nil {
		rp.primaryPool.Close()
	}

	// Parar replica manager
	if rp.replicaManager != nil {
		rp.replicaManager.Stop(context.Background())
	}
}

// Stats retorna estatísticas do pool (combina read e write)
func (rp *ReplicaPool) Stats() map[string]interface{} {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	stats := make(map[string]interface{})

	// Estatísticas do primary
	if rp.primaryPool != nil {
		stats["primary"] = rp.primaryPool.Stats()
	}

	// Estatísticas de leitura
	stats["read"] = rp.readStats.ToMap()

	// Estatísticas de escrita
	stats["write"] = rp.writeStats.ToMap()

	// Estatísticas do replica manager
	if rp.replicaManager != nil {
		stats["replicas"] = rp.replicaManager.GetStats().ToMap()
	}

	return stats
}

// ReplicaConn wrapper para conexão de réplica
type ReplicaConn struct {
	interfaces.IConn
	replicaInfo *ReplicaInfo
}

// Release libera a conexão e decrementa contador
func (rc *ReplicaConn) Release() {
	if rc.replicaInfo != nil {
		rc.replicaInfo.DecrementConnections()
	}
	rc.IConn.Release()
}

// ReplicaPoolBuilder para construir pools de réplicas
type ReplicaPoolBuilder struct {
	primaryPool          interfaces.IPool
	replicaManagerConfig ReplicaManagerConfig
	replicas             []ReplicaConfig
}

// ReplicaConfig configuração de uma réplica
type ReplicaConfig struct {
	ID     string
	DSN    string
	Weight int
	Tags   map[string]string
}

// NewReplicaPoolBuilder cria um novo builder
func NewReplicaPoolBuilder(primaryPool interfaces.IPool) *ReplicaPoolBuilder {
	return &ReplicaPoolBuilder{
		primaryPool: primaryPool,
		replicaManagerConfig: ReplicaManagerConfig{
			LoadBalancingStrategy: interfaces.LoadBalancingRoundRobin,
			ReadPreference:        interfaces.ReadPreferenceSecondaryPreferred,
			HealthCheckInterval:   30 * time.Second,
			HealthCheckTimeout:    5 * time.Second,
		},
	}
}

// SetLoadBalancingStrategy define estratégia de balanceamento
func (b *ReplicaPoolBuilder) SetLoadBalancingStrategy(strategy interfaces.LoadBalancingStrategy) *ReplicaPoolBuilder {
	b.replicaManagerConfig.LoadBalancingStrategy = strategy
	return b
}

// SetReadPreference define preferência de leitura
func (b *ReplicaPoolBuilder) SetReadPreference(preference interfaces.ReadPreference) *ReplicaPoolBuilder {
	b.replicaManagerConfig.ReadPreference = preference
	return b
}

// SetHealthCheckInterval define intervalo de health check
func (b *ReplicaPoolBuilder) SetHealthCheckInterval(interval time.Duration) *ReplicaPoolBuilder {
	b.replicaManagerConfig.HealthCheckInterval = interval
	return b
}

// SetProviderFactory define factory para criar pools
func (b *ReplicaPoolBuilder) SetProviderFactory(factory func(dsn string) (interfaces.IPool, error)) *ReplicaPoolBuilder {
	b.replicaManagerConfig.ProviderFactory = factory
	return b
}

// AddReplica adiciona uma réplica
func (b *ReplicaPoolBuilder) AddReplica(id, dsn string, weight int) *ReplicaPoolBuilder {
	b.replicas = append(b.replicas, ReplicaConfig{
		ID:     id,
		DSN:    dsn,
		Weight: weight,
		Tags:   make(map[string]string),
	})
	return b
}

// Build constrói o pool de réplicas
func (b *ReplicaPoolBuilder) Build(ctx context.Context) (*ReplicaPool, error) {
	// Criar replica manager
	replicaManager := NewReplicaManager(b.replicaManagerConfig)

	// Adicionar réplicas
	for _, replica := range b.replicas {
		err := replicaManager.AddReplica(ctx, replica.ID, replica.DSN, replica.Weight)
		if err != nil {
			return nil, fmt.Errorf("failed to add replica %s: %w", replica.ID, err)
		}
	}

	// Iniciar replica manager
	if err := replicaManager.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start replica manager: %w", err)
	}

	// Criar pool
	pool := NewReplicaPool(b.primaryPool, replicaManager)
	pool.SetReadPreference(b.replicaManagerConfig.ReadPreference)

	return pool, nil
}
