package interfaces

import (
	"context"
	"time"
)

// LoadBalancingStrategy representa as estratégias de balanceamento de carga
type LoadBalancingStrategy string

const (
	// LoadBalancingRoundRobin distribui as requisições de forma circular
	LoadBalancingRoundRobin LoadBalancingStrategy = "round_robin"

	// LoadBalancingRandom distribui as requisições aleatoriamente
	LoadBalancingRandom LoadBalancingStrategy = "random"

	// LoadBalancingWeighted distribui baseado em pesos configurados
	LoadBalancingWeighted LoadBalancingStrategy = "weighted"

	// LoadBalancingLatency distribui baseado na latência medida
	LoadBalancingLatency LoadBalancingStrategy = "latency"
)

// ReadPreference define a preferência de leitura
type ReadPreference string

const (
	// ReadPreferencePrimary força leitura apenas no primary
	ReadPreferencePrimary ReadPreference = "primary"

	// ReadPreferenceSecondary força leitura apenas nas réplicas
	ReadPreferenceSecondary ReadPreference = "secondary"

	// ReadPreferenceSecondaryPreferred prefere réplicas, mas pode usar primary
	ReadPreferenceSecondaryPreferred ReadPreference = "secondary_preferred"

	// ReadPreferenceNearest usa a conexão com menor latência
	ReadPreferenceNearest ReadPreference = "nearest"
)

// ReplicaStatus representa o status de uma réplica
type ReplicaStatus string

const (
	// ReplicaStatusHealthy indica que a réplica está saudável
	ReplicaStatusHealthy ReplicaStatus = "healthy"

	// ReplicaStatusUnhealthy indica que a réplica está com problemas
	ReplicaStatusUnhealthy ReplicaStatus = "unhealthy"

	// ReplicaStatusRecovering indica que a réplica está se recuperando
	ReplicaStatusRecovering ReplicaStatus = "recovering"

	// ReplicaStatusMaintenance indica que a réplica está em manutenção
	ReplicaStatusMaintenance ReplicaStatus = "maintenance"
)

// IReplicaInfo contém informações sobre uma réplica
type IReplicaInfo interface {
	GetID() string
	GetDSN() string
	GetWeight() int
	GetStatus() ReplicaStatus
	GetLatency() time.Duration
	GetLastHealthCheck() time.Time
	GetConnectionCount() int
	GetMaxConnections() int
	GetRegion() string
	GetTags() map[string]string

	// Estatísticas
	GetSuccessRate() float64
	GetErrorRate() float64
	GetAvgLatency() time.Duration
	GetTotalQueries() int64
	GetFailedQueries() int64

	// Controle
	IsAvailable() bool
	MarkHealthy()
	MarkUnhealthy()
	MarkRecovering()
	MarkMaintenance()

	// Configuração
	SetWeight(weight int)
	SetMaxConnections(max int)
	SetTags(tags map[string]string)

	// Pool de conexões
	GetPool() IPool
	SetPool(pool IPool)
}

// IReplicaManager gerencia as réplicas de leitura
type IReplicaManager interface {
	// Configuração
	AddReplica(ctx context.Context, id string, dsn string, weight int) error
	RemoveReplica(ctx context.Context, id string) error
	GetReplica(id string) (IReplicaInfo, error)
	ListReplicas() []IReplicaInfo

	// Balanceamento
	SelectReplica(ctx context.Context, preference ReadPreference) (IReplicaInfo, error)
	SelectReplicaWithStrategy(ctx context.Context, strategy LoadBalancingStrategy) (IReplicaInfo, error)

	// Health checking
	HealthCheck(ctx context.Context, replicaID string) error
	HealthCheckAll(ctx context.Context) error
	GetHealthyReplicas() []IReplicaInfo
	GetUnhealthyReplicas() []IReplicaInfo

	// Configuração de estratégias
	SetLoadBalancingStrategy(strategy LoadBalancingStrategy)
	GetLoadBalancingStrategy() LoadBalancingStrategy
	SetReadPreference(preference ReadPreference)
	GetReadPreference() ReadPreference

	// Configuração de health checks
	SetHealthCheckInterval(interval time.Duration)
	GetHealthCheckInterval() time.Duration
	SetHealthCheckTimeout(timeout time.Duration)
	GetHealthCheckTimeout() time.Duration

	// Estatísticas
	GetStats() IReplicaStats
	GetReplicaStats(id string) (IReplicaStats, error)

	// Eventos
	OnReplicaHealthChange(callback func(replica IReplicaInfo, oldStatus, newStatus ReplicaStatus))
	OnReplicaFailover(callback func(from, to IReplicaInfo))

	// Controle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Manutenção
	SetReplicaMaintenance(id string, maintenance bool) error
	DrainReplica(ctx context.Context, id string, timeout time.Duration) error
}

// IReplicaStats contém estatísticas das réplicas
type IReplicaStats interface {
	GetTotalReplicas() int
	GetHealthyReplicas() int
	GetUnhealthyReplicas() int
	GetRecoveringReplicas() int
	GetMaintenanceReplicas() int

	GetTotalQueries() int64
	GetSuccessfulQueries() int64
	GetFailedQueries() int64
	GetAvgLatency() time.Duration
	GetMaxLatency() time.Duration
	GetMinLatency() time.Duration

	GetQueryDistribution() map[string]int64           // replica_id -> query_count
	GetLatencyDistribution() map[string]time.Duration // replica_id -> avg_latency
	GetErrorDistribution() map[string]int64           // replica_id -> error_count

	GetFailoverCount() int64
	GetLastFailoverTime() time.Time
	GetUptime() time.Duration

	// Métricas por período
	GetQueriesPerSecond() float64
	GetErrorsPerSecond() float64
	GetLatencyPercentile(percentile float64) time.Duration

	// Export para monitoramento
	ToMap() map[string]interface{}
	ToJSON() ([]byte, error)
}

// IReplicaPool representa um pool de conexões para réplicas
type IReplicaPool interface {
	// Herda de IPool
	IPool

	// Funcionalidades específicas de réplicas
	GetReplicaManager() IReplicaManager
	SetReadPreference(preference ReadPreference)
	GetReadPreference() ReadPreference

	// Conexões com preferência
	AcquireRead(ctx context.Context, preference ReadPreference) (IConn, error)
	AcquireWrite(ctx context.Context) (IConn, error)

	// Estatísticas
	GetReadStats() IReplicaStats
	GetWriteStats() IReplicaStats
}
