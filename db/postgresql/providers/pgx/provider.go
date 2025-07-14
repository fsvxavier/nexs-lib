package pgx

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Provider implements postgresql.IProvider for PGX driver
type Provider struct {
	mu      sync.RWMutex
	pools   map[string]*Pool
	conns   map[string]*Conn
	metrics map[string]interface{}
	logger  *log.Logger
	closed  bool
}

// NewProvider creates a new PGX provider
func NewProvider() *Provider {
	return &Provider{
		pools:   make(map[string]*Pool),
		conns:   make(map[string]*Conn),
		metrics: make(map[string]interface{}),
		logger:  log.New(os.Stdout, "[PGX] ", log.LstdFlags|log.Lshortfile),
	}
}

// Type returns the provider type
func (p *Provider) Type() postgresql.ProviderType {
	return postgresql.ProviderTypePGX
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "PGX PostgreSQL Provider"
}

// Version returns the provider version
func (p *Provider) Version() string {
	return "5.7.5" // pgx version
}

// CreatePool creates a new connection pool
func (p *Provider) CreatePool(ctx context.Context, cfg *config.Config) (postgresql.IPool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, fmt.Errorf("provider is closed")
	}

	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	pgxConfig, err := p.buildPoolConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build pool config: %w", err)
	}

	pgxPool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	pool := &Pool{
		pool:               pgxPool,
		config:             cfg,
		multiTenantEnabled: cfg.MultiTenantEnabled,
		logger:             p.logger,
		metrics:            make(map[string]interface{}),
	}

	// Store pool for management
	poolKey := fmt.Sprintf("%s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	p.pools[poolKey] = pool

	p.logger.Printf("Created pool for %s", poolKey)

	return pool, nil
}

// CreateConnection creates a new single connection
func (p *Provider) CreateConnection(ctx context.Context, cfg *config.Config) (postgresql.IConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, fmt.Errorf("provider is closed")
	}

	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	pgxConfig, err := p.buildConnConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build connection config: %w", err)
	}

	pgxConn, err := pgx.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	conn := &Conn{
		conn:               pgxConn,
		config:             cfg,
		multiTenantEnabled: cfg.MultiTenantEnabled,
		logger:             p.logger,
		isPooled:           false,
	}

	// Store connection for management
	connKey := fmt.Sprintf("conn_%d", time.Now().UnixNano())
	p.conns[connKey] = conn

	p.logger.Printf("Created connection %s", connKey)

	return conn, nil
}

// IsHealthy checks if the provider is healthy
func (p *Provider) IsHealthy(ctx context.Context) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return false
	}

	// Check all pools health
	for _, pool := range p.pools {
		if err := pool.Ping(ctx); err != nil {
			p.logger.Printf("Pool health check failed: %v", err)
			return false
		}
	}

	return true
}

// GetMetrics returns provider metrics
func (p *Provider) GetMetrics(ctx context.Context) map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	metrics := make(map[string]interface{})

	// Provider level metrics
	metrics["type"] = p.Type()
	metrics["name"] = p.Name()
	metrics["version"] = p.Version()
	metrics["pools_count"] = len(p.pools)
	metrics["connections_count"] = len(p.conns)
	metrics["is_healthy"] = p.IsHealthy(ctx)

	// Pool metrics
	poolMetrics := make(map[string]interface{})
	for key, pool := range p.pools {
		poolMetrics[key] = pool.Stats()
	}
	metrics["pools"] = poolMetrics

	return metrics
}

// Close closes the provider and all resources
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// Close all pools
	for key, pool := range p.pools {
		pool.Close()
		delete(p.pools, key)
	}

	// Close all connections
	for key, conn := range p.conns {
		conn.Release(context.Background())
		delete(p.conns, key)
	}

	p.logger.Printf("Provider closed")

	return nil
}

// buildPoolConfig builds pgxpool.Config from our config
func (p *Provider) buildPoolConfig(cfg *config.Config) (*pgxpool.Config, error) {
	connString := cfg.ConnectionString()

	pgxConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Pool settings
	pgxConfig.MaxConns = cfg.MaxConns
	pgxConfig.MinConns = cfg.MinConns
	pgxConfig.MaxConnLifetime = cfg.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Connection settings
	p.configureConnection(pgxConfig.ConnConfig, cfg)

	return pgxConfig, nil
}

// buildConnConfig builds pgx.ConnConfig from our config
func (p *Provider) buildConnConfig(cfg *config.Config) (*pgx.ConnConfig, error) {
	connString := cfg.ConnectionString()

	pgxConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Connection settings
	p.configureConnection(pgxConfig, cfg)

	return pgxConfig, nil
}

// configureConnection configures common connection settings
func (p *Provider) configureConnection(pgxConfig *pgx.ConnConfig, cfg *config.Config) {
	// TLS configuration
	if cfg.TLSConfig != nil {
		pgxConfig.TLSConfig = cfg.TLSConfig
	} else if enabled, _ := strconv.ParseBool(os.Getenv("DB_TLS_ENABLED")); enabled {
		pgxConfig.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// Runtime parameters
	if pgxConfig.RuntimeParams == nil {
		pgxConfig.RuntimeParams = make(map[string]string)
	}

	// Set timezone
	pgxConfig.RuntimeParams["timezone"] = cfg.Timezone

	// Add custom runtime parameters
	for key, value := range cfg.RuntimeParams {
		pgxConfig.RuntimeParams[key] = value
	}

	// Query execution mode
	switch cfg.QueryExecMode {
	case config.QueryExecModeCacheStatement:
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	case config.QueryExecModeCacheDescribe:
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	case config.QueryExecModeDescribeExec:
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec
	case config.QueryExecModeExec:
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	case config.QueryExecModeSimpleProtocol:
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	default:
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	}

	// Set application name
	if cfg.ApplicationName != "" {
		pgxConfig.RuntimeParams["application_name"] = cfg.ApplicationName
	}

	// Timeouts (handled at higher level)
	// Connection timeout is handled during connection establishment
	// Query timeout should be handled in context
}
