package pgxprovider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/hooks"
	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/memory"
	"github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/monitoring"
	"github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/replicas"
	"github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/resilience"
)

// Provider implementa IPostgreSQLProvider para PGX
type Provider struct {
	name              string
	version           string
	supportedFeatures []string
	retryManager      interfaces.IRetryManager
	failoverManager   interfaces.IFailoverManager
	replicaManager    interfaces.IReplicaManager
	hookManager       interfaces.IHookManager
	bufferPool        interfaces.IBufferPool
	safetyMonitor     interfaces.ISafetyMonitor
	mu                sync.RWMutex
}

// NewProvider cria um novo provider PGX
func NewProvider() interfaces.IPostgreSQLProvider {
	// Configurações padrão
	retryConfig := interfaces.RetryConfig{
		MaxRetries:      3,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		Multiplier:      2.0,
		RandomizeWait:   true,
	}

	failoverConfig := interfaces.FailoverConfig{
		Enabled:             false,
		FallbackNodes:       []string{},
		HealthCheckInterval: 30 * time.Second,
		RetryInterval:       5 * time.Second,
		MaxFailoverAttempts: 3,
	}

	return &Provider{
		name:    "PGX",
		version: "5.x",
		supportedFeatures: []string{
			"transactions",
			"prepared_statements",
			"batch_operations",
			"copy_operations",
			"listen_notify",
			"connection_pooling",
			"hooks",
			"multi_tenancy",
			"health_checks",
			"statistics",
			"ssl_connections",
			"async_operations",
			"failover",
			"retry",
			"buffer_pooling",
			"safety_monitoring",
			"read_replicas",
		},
		retryManager:    resilience.NewRetryManager(retryConfig),
		failoverManager: resilience.NewFailoverManager(failoverConfig),
		replicaManager: replicas.NewReplicaManager(replicas.ReplicaManagerConfig{
			LoadBalancingStrategy: interfaces.LoadBalancingRoundRobin,
			ReadPreference:        interfaces.ReadPreferenceSecondaryPreferred,
			HealthCheckInterval:   30 * time.Second,
			HealthCheckTimeout:    5 * time.Second,
		}),
		hookManager:   hooks.NewDefaultHookManager(),
		bufferPool:    memory.NewBufferPool(1024, 1024*1024), // 1KB min, 1MB max
		safetyMonitor: monitoring.NewSafetyMonitor(),
	}
}

// Name retorna o nome do provider
func (p *Provider) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.name
}

// Version retorna a versão do provider
func (p *Provider) Version() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.version
}

// SupportsFeature verifica se uma feature é suportada
func (p *Provider) SupportsFeature(feature string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, supportedFeature := range p.supportedFeatures {
		if supportedFeature == feature {
			return true
		}
	}
	return false
}

// GetDriverName retorna o nome do driver
func (p *Provider) GetDriverName() string {
	return "pgx"
}

// GetSupportedFeatures retorna todas as features suportadas
func (p *Provider) GetSupportedFeatures() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	features := make([]string, len(p.supportedFeatures))
	copy(features, p.supportedFeatures)
	return features
}

// ValidateConfig valida a configuração
func (p *Provider) ValidateConfig(config interfaces.IConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	return config.Validate()
}

// NewPool cria um novo pool de conexões
func (p *Provider) NewPool(ctx context.Context, config interfaces.IConfig) (interfaces.IPool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return NewPool(ctx, config, p.bufferPool, p.safetyMonitor, p.hookManager)
}

// NewConn cria uma nova conexão
func (p *Provider) NewConn(ctx context.Context, config interfaces.IConfig) (interfaces.IConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return NewConn(ctx, config, p.bufferPool, p.hookManager)
}

// NewListenConn cria uma conexão para LISTEN/NOTIFY
func (p *Provider) NewListenConn(ctx context.Context, config interfaces.IConfig) (interfaces.IConn, error) {
	if !p.SupportsFeature("listen_notify") {
		return nil, fmt.Errorf("LISTEN/NOTIFY not supported")
	}

	return p.NewConn(ctx, config)
}

// HealthCheck verifica saúde do provider
func (p *Provider) HealthCheck(ctx context.Context, config interfaces.IConfig) error {
	// Verificar se safety monitor está saudável
	if !p.safetyMonitor.IsHealthy() {
		return fmt.Errorf("safety monitor reports unhealthy state")
	}

	// Teste de conexão básico
	conn, err := p.NewConn(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}
	defer conn.Close(ctx)

	return conn.Ping(ctx)
}

// CreateSchema cria um schema
func (p *Provider) CreateSchema(ctx context.Context, conn interfaces.IConn, schemaName string) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if schemaName == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	_, err := conn.Exec(ctx, query)
	return err
}

// DropSchema remove um schema
func (p *Provider) DropSchema(ctx context.Context, conn interfaces.IConn, schemaName string) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if schemaName == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	_, err := conn.Exec(ctx, query)
	return err
}

// ListSchemas lista todos os schemas
func (p *Provider) ListSchemas(ctx context.Context, conn interfaces.IConn) ([]string, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}

	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
		ORDER BY schema_name
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, err
		}
		schemas = append(schemas, schemaName)
	}

	return schemas, rows.Err()
}

// CreateDatabase cria um banco de dados
func (p *Provider) CreateDatabase(ctx context.Context, conn interfaces.IConn, dbName string) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := conn.Exec(ctx, query)
	return err
}

// DropDatabase remove um banco de dados
func (p *Provider) DropDatabase(ctx context.Context, conn interfaces.IConn, dbName string) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := conn.Exec(ctx, query)
	return err
}

// ListDatabases lista todos os bancos de dados
func (p *Provider) ListDatabases(ctx context.Context, conn interfaces.IConn) ([]string, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}

	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false
		ORDER BY datname
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}

	return databases, rows.Err()
}

// WithRetry executa operação com retry
func (p *Provider) WithRetry(ctx context.Context, operation func() error) error {
	return p.retryManager.Execute(ctx, operation)
}

// WithFailover executa operação com failover
func (p *Provider) WithFailover(ctx context.Context, operation func(conn interfaces.IConn) error) error {
	return p.failoverManager.Execute(ctx, operation)
}

// GetRetryManager retorna o retry manager
func (p *Provider) GetRetryManager() interfaces.IRetryManager {
	return p.retryManager
}

// GetFailoverManager retorna o failover manager
func (p *Provider) GetFailoverManager() interfaces.IFailoverManager {
	return p.failoverManager
}

// GetReplicaManager retorna o replica manager
func (p *Provider) GetReplicaManager() interfaces.IReplicaManager {
	return p.replicaManager
}

// GetHookManager retorna o hook manager
func (p *Provider) GetHookManager() interfaces.IHookManager {
	return p.hookManager
}

// GetBufferPool retorna o buffer pool
func (p *Provider) GetBufferPool() interfaces.IBufferPool {
	return p.bufferPool
}

// GetSafetyMonitor retorna o safety monitor
func (p *Provider) GetSafetyMonitor() interfaces.ISafetyMonitor {
	return p.safetyMonitor
}

// Close fecha o provider e libera recursos
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Fechar componentes
	if bufferPool, ok := p.bufferPool.(*memory.BufferPool); ok {
		bufferPool.Close()
	}

	if safetyMonitor, ok := p.safetyMonitor.(*monitoring.SafetyMonitor); ok {
		safetyMonitor.Close()
	}

	return nil
}

// Stats retorna estatísticas do provider
func (p *Provider) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := map[string]interface{}{
		"name":               p.name,
		"version":            p.version,
		"supported_features": p.supportedFeatures,
		"buffer_pool_stats":  p.bufferPool.Stats(),
		"retry_stats":        p.retryManager.GetStats(),
		"failover_stats":     p.failoverManager.GetStats(),
		"safety_healthy":     p.safetyMonitor.IsHealthy(),
		"deadlocks":          len(p.safetyMonitor.CheckDeadlocks()),
		"race_conditions":    len(p.safetyMonitor.CheckRaceConditions()),
		"resource_leaks":     len(p.safetyMonitor.CheckLeaks()),
	}

	return stats
}
