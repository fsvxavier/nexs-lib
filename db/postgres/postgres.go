// Package postgres provides a refactored, high-performance PostgreSQL database library
// with memory optimization, resilience patterns, and robust architecture.
//
// This package follows Domain-Driven Design (DDD) principles and implements
// the Hexagonal Architecture pattern for clean separation of concerns.
//
// Key Features:
// - Memory-optimized buffer pooling
// - Thread-safety monitoring
// - Retry and failover mechanisms
// - Hook system for extensibility
// - Comprehensive error handling
// - Performance monitoring
// - Multi-tenancy support
// - Connection pooling with advanced configuration
//
// Example usage:
//
//	// Create a provider
//	provider, err := postgres.NewPGXProvider()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create configuration
//	config := postgres.NewDefaultConfig("postgres://user:pass@localhost/db")
//
//	// Create connection pool
//	pool, err := provider.NewPool(ctx, config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer pool.Close()
//
//	// Use the pool
//	conn, err := pool.Acquire(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer conn.Release()
//
//	// Execute queries
//	rows, err := conn.Query(ctx, "SELECT * FROM users")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer rows.Close()
package postgres

import (
	"context"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/config"
	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// Version da biblioteca
const Version = "2.0.0"

// Configuração padrão
const (
	DefaultMaxConns          = 30
	DefaultMinConns          = 5
	DefaultMaxConnLifetime   = time.Hour
	DefaultMaxConnIdleTime   = time.Minute * 30
	DefaultConnectTimeout    = time.Second * 30
	DefaultHealthCheckPeriod = time.Minute * 5
)

// NewDefaultConfig cria uma configuração padrão
func NewDefaultConfig(connectionString string) interfaces.IConfig {
	return config.NewDefaultConfig(connectionString)
}

// NewConfigWithOptions cria uma configuração com opções personalizadas
func NewConfigWithOptions(connectionString string, options ...config.ConfigOption) interfaces.IConfig {
	cfg := config.NewDefaultConfig(connectionString).(*config.DefaultConfig)
	cfg.Apply(options...)
	return cfg
}

// ConfigOption re-exporta para facilitar o uso
type ConfigOption = config.ConfigOption

// Opções de configuração re-exportadas
var (
	WithConnectionString = config.WithConnectionString
	WithMaxConns         = config.WithMaxConns
	WithMinConns         = config.WithMinConns
	WithMaxConnLifetime  = config.WithMaxConnLifetime
	WithMaxConnIdleTime  = config.WithMaxConnIdleTime
	WithMultiTenant      = config.WithMultiTenant
	WithTLS              = config.WithTLS
	WithRetry            = config.WithRetry
	WithFailover         = config.WithFailover
	WithReadReplicas     = config.WithReadReplicas
	WithEnabledHooks     = config.WithEnabledHooks
	WithCustomHook       = config.WithCustomHook
)

// ProviderType re-exporta para facilitar o uso
type ProviderType = interfaces.ProviderType

// Tipos de provider re-exportados
const (
	ProviderTypePGX = interfaces.ProviderTypePGX
)

// Interfaces re-exportadas para facilitar o uso
type (
	IProvider           = interfaces.IProvider
	IPostgreSQLProvider = interfaces.IPostgreSQLProvider
	IProviderFactory    = interfaces.IProviderFactory
	IConfig             = interfaces.IConfig
	IPool               = interfaces.IPool
	IConn               = interfaces.IConn
	ITransaction        = interfaces.ITransaction
	IRows               = interfaces.IRows
	IRow                = interfaces.IRow
	IBatch              = interfaces.IBatch
	IBatchResults       = interfaces.IBatchResults
	IHookManager        = interfaces.IHookManager
	IRetryManager       = interfaces.IRetryManager
	IFailoverManager    = interfaces.IFailoverManager
	IBufferPool         = interfaces.IBufferPool
	ISafetyMonitor      = interfaces.ISafetyMonitor
)

// Structs re-exportadas
type (
	PoolConfig        = interfaces.PoolConfig
	TLSConfig         = interfaces.TLSConfig
	RetryConfig       = interfaces.RetryConfig
	HookConfig        = interfaces.HookConfig
	FailoverConfig    = interfaces.FailoverConfig
	ReadReplicaConfig = interfaces.ReadReplicaConfig
	ConnectionStats   = interfaces.ConnectionStats
	PoolStats         = interfaces.PoolStats
	MemoryStats       = interfaces.MemoryStats
	RetryStats        = interfaces.RetryStats
	FailoverStats     = interfaces.FailoverStats
	TxOptions         = interfaces.TxOptions
	ExecutionContext  = interfaces.ExecutionContext
	HookResult        = interfaces.HookResult
	Notification      = interfaces.Notification
	DeadlockInfo      = interfaces.DeadlockInfo
	RaceConditionInfo = interfaces.RaceConditionInfo
	LeakInfo          = interfaces.LeakInfo
)

// Enums re-exportados
type (
	LoadBalanceMode  = interfaces.LoadBalanceMode
	TxIsoLevel       = interfaces.TxIsoLevel
	TxAccessMode     = interfaces.TxAccessMode
	TxDeferrableMode = interfaces.TxDeferrableMode
	HookType         = interfaces.HookType
	Hook             = interfaces.Hook
)

// Constantes re-exportadas
const (
	// Load balance modes
	LoadBalanceModeRoundRobin = interfaces.LoadBalanceModeRoundRobin
	LoadBalanceModeRandom     = interfaces.LoadBalanceModeRandom
	LoadBalanceModeWeighted   = interfaces.LoadBalanceModeWeighted

	// Transaction isolation levels
	TxIsoLevelDefault         = interfaces.TxIsoLevelDefault
	TxIsoLevelReadUncommitted = interfaces.TxIsoLevelReadUncommitted
	TxIsoLevelReadCommitted   = interfaces.TxIsoLevelReadCommitted
	TxIsoLevelRepeatableRead  = interfaces.TxIsoLevelRepeatableRead
	TxIsoLevelSerializable    = interfaces.TxIsoLevelSerializable

	// Transaction access modes
	TxAccessModeReadWrite = interfaces.TxAccessModeReadWrite
	TxAccessModeReadOnly  = interfaces.TxAccessModeReadOnly

	// Transaction deferrable modes
	TxDeferrableModeNotDeferrable = interfaces.TxDeferrableModeNotDeferrable
	TxDeferrableModeDeferrable    = interfaces.TxDeferrableModeDeferrable

	// Hook types
	BeforeConnectionHook  = interfaces.BeforeConnectionHook
	AfterConnectionHook   = interfaces.AfterConnectionHook
	BeforeReleaseHook     = interfaces.BeforeReleaseHook
	AfterReleaseHook      = interfaces.AfterReleaseHook
	BeforeQueryHook       = interfaces.BeforeQueryHook
	AfterQueryHook        = interfaces.AfterQueryHook
	BeforeExecHook        = interfaces.BeforeExecHook
	AfterExecHook         = interfaces.AfterExecHook
	BeforeTransactionHook = interfaces.BeforeTransactionHook
	AfterTransactionHook  = interfaces.AfterTransactionHook
	BeforeBatchHook       = interfaces.BeforeBatchHook
	AfterBatchHook        = interfaces.AfterBatchHook
	BeforeAcquireHook     = interfaces.BeforeAcquireHook
	AfterAcquireHook      = interfaces.AfterAcquireHook
	OnErrorHook           = interfaces.OnErrorHook
	CustomHookBase        = interfaces.CustomHookBase
)

// Funções de conveniência

// Connect cria uma conexão simples usando configuração padrão
func Connect(ctx context.Context, connectionString string) (IConn, error) {
	provider, err := NewPGXProvider()
	if err != nil {
		return nil, err
	}

	config := NewDefaultConfig(connectionString)
	return provider.NewConn(ctx, config)
}

// ConnectPool cria um pool de conexões usando configuração padrão
func ConnectPool(ctx context.Context, connectionString string) (IPool, error) {
	provider, err := NewPGXProvider()
	if err != nil {
		return nil, err
	}

	config := NewDefaultConfig(connectionString)
	return provider.NewPool(ctx, config)
}

// ConnectWithConfig cria uma conexão com configuração personalizada
func ConnectWithConfig(ctx context.Context, config IConfig) (IConn, error) {
	provider, err := NewPGXProvider()
	if err != nil {
		return nil, err
	}

	return provider.NewConn(ctx, config)
}

// ConnectPoolWithConfig cria um pool com configuração personalizada
func ConnectPoolWithConfig(ctx context.Context, config IConfig) (IPool, error) {
	provider, err := NewPGXProvider()
	if err != nil {
		return nil, err
	}

	return provider.NewPool(ctx, config)
}

// GetVersion retorna a versão da biblioteca
func GetVersion() string {
	return Version
}

// GetSupportedProviders retorna lista de providers suportados
func GetSupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderTypePGX,
	}
}

// IsProviderSupported verifica se um provider é suportado
func IsProviderSupported(providerType ProviderType) bool {
	supported := GetSupportedProviders()
	for _, p := range supported {
		if p == providerType {
			return true
		}
	}
	return false
}
