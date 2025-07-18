package interfaces

import (
	"context"
	"time"
)

// ProviderType representa o tipo de provider de banco de dados
type ProviderType string

const (
	ProviderTypePGX ProviderType = "pgx"
)

// IProvider interface base para todos os providers
type IProvider interface {
	Name() string
	Version() string
	SupportsFeature(feature string) bool
	GetDriverName() string
	GetSupportedFeatures() []string
	ValidateConfig(config IConfig) error

	// Connection management
	NewPool(ctx context.Context, config IConfig) (IPool, error)
	NewConn(ctx context.Context, config IConfig) (IConn, error)

	// Health check
	HealthCheck(ctx context.Context, config IConfig) error
}

// IPostgreSQLProvider interface específica para PostgreSQL
type IPostgreSQLProvider interface {
	IProvider

	// PostgreSQL specific features
	NewListenConn(ctx context.Context, config IConfig) (IConn, error)
	CreateSchema(ctx context.Context, conn IConn, schemaName string) error
	DropSchema(ctx context.Context, conn IConn, schemaName string) error
	ListSchemas(ctx context.Context, conn IConn) ([]string, error)

	// Multi-database support
	CreateDatabase(ctx context.Context, conn IConn, dbName string) error
	DropDatabase(ctx context.Context, conn IConn, dbName string) error
	ListDatabases(ctx context.Context, conn IConn) ([]string, error)

	// Resilience features
	WithRetry(ctx context.Context, operation func() error) error
	WithFailover(ctx context.Context, operation func(conn IConn) error) error
	GetRetryManager() IRetryManager
	GetReplicaManager() IReplicaManager
	GetFailoverManager() IFailoverManager
}

// IProviderFactory cria providers
type IProviderFactory interface {
	CreateProvider(providerType ProviderType) (IPostgreSQLProvider, error)
	RegisterProvider(providerType ProviderType, provider IPostgreSQLProvider) error
	ListProviders() []ProviderType
	GetProvider(providerType ProviderType) (IPostgreSQLProvider, bool)
} // IConfig interface para configuração do banco
type IConfig interface {
	GetConnectionString() string
	GetPoolConfig() PoolConfig
	GetTLSConfig() TLSConfig
	GetRetryConfig() RetryConfig
	GetHookConfig() HookConfig
	GetFailoverConfig() FailoverConfig
	GetReadReplicaConfig() ReadReplicaConfig
	IsMultiTenantEnabled() bool
	Validate() error
}

// PoolConfig configuração do pool de conexões
type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
	LazyConnect       bool
}

// TLSConfig configuração TLS
type TLSConfig struct {
	Enabled            bool
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	CAFile             string
	ServerName         string
}

// RetryConfig configuração de retry
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomizeWait   bool
}

// HookConfig configuração de hooks
type HookConfig struct {
	EnabledHooks []HookType
	CustomHooks  map[string]HookType
	HookTimeout  time.Duration
}

// FailoverConfig configuração de failover
type FailoverConfig struct {
	Enabled             bool
	FallbackNodes       []string
	HealthCheckInterval time.Duration
	RetryInterval       time.Duration
	MaxFailoverAttempts int
}

// ReadReplicaConfig configuração de read replicas
type ReadReplicaConfig struct {
	Enabled             bool
	ConnectionStrings   []string
	LoadBalanceMode     LoadBalanceMode
	HealthCheckInterval time.Duration
}

// LoadBalanceMode modo de balanceamento de carga
type LoadBalanceMode int

const (
	LoadBalanceModeRoundRobin LoadBalanceMode = iota
	LoadBalanceModeRandom
	LoadBalanceModeWeighted
)
