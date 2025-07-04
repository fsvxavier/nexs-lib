package pgx

import (
	"context"
	"crypto/tls"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool implementa a interface common.IPool usando pgx
type Pool struct {
	pool               *pgxpool.Pool
	config             *common.Config
	multiTenantEnabled bool
}

// NewPool cria um novo pool de conexões PostgreSQL usando pgx
func NewPool(ctx context.Context, config *common.Config) (common.IPool, error) {
	pgxConfig, err := pgxpool.ParseConfig(config.ConnectionString())
	if err != nil {
		return nil, err
	}

	// Configura TLS se necessário
	if config.SSLMode == "require" || config.SSLMode == "verify-ca" || config.SSLMode == "verify-full" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: config.SSLMode == "require",
		}
		pgxConfig.ConnConfig.TLSConfig = tlsConfig
	}

	// Configura outros parâmetros
	pgxConfig.ConnConfig.RuntimeParams["timezone"] = "UTC"

	// Configura o pool
	pgxConfig.MaxConns = config.MaxConns
	pgxConfig.MinConns = config.MinConns
	pgxConfig.MaxConnLifetime = config.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = config.MaxConnIdleTime

	// Configura tracer se habilitado
	if config.TraceEnabled {
		pgxConfig.ConnConfig.Tracer = NewTracer(config)
	}

	// Cria o pool
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	return &Pool{
		pool:               pool,
		config:             config,
		multiTenantEnabled: config.MultiTenantEnabled,
	}, nil
}

// Acquire adquire uma conexão do pool
func (p *Pool) Acquire(ctx context.Context) (common.IConn, error) {
	poolConn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		poolConn:           poolConn,
		multiTenantEnabled: p.multiTenantEnabled,
		config:             p.config,
	}

	// Se multi-tenant está habilitado, configura o contexto do tenant
	if p.multiTenantEnabled {
		if err := conn.setupTenantContext(ctx); err != nil {
			poolConn.Release()
			return nil, err
		}
	}

	return conn, nil
}

// Close fecha o pool de conexões
func (p *Pool) Close() error {
	p.pool.Close()
	return nil
}

// Ping verifica a conectividade com o banco de dados
func (p *Pool) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Stats retorna estatísticas do pool de conexões
func (p *Pool) Stats() *common.PoolStats {
	stats := p.pool.Stat()

	return &common.PoolStats{
		AcquiredConns:    int(stats.AcquiredConns()),
		TotalConns:       int(stats.TotalConns()),
		IdleConns:        int(stats.IdleConns()),
		MaxConns:         int(p.config.MaxConns),
		MinConns:         int(p.config.MinConns),
		ConstructedConns: int(stats.NewConnsCount()),
	}
}

// NewTracer cria um novo tracer para o pgx
func NewTracer(config *common.Config) pgx.QueryTracer {
	return &TracerConfig{
		QueryLogEnabled: config.QueryLogEnabled,
		TraceEnabled:    config.TraceEnabled,
	}
}

// TracerConfig implementa pgx.QueryTracer
type TracerConfig struct {
	QueryLogEnabled bool
	TraceEnabled    bool
}

// TraceQueryStart é chamado quando uma query começa a ser executada
func (tc *TracerConfig) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	// Implementação do tracer
	return ctx
}

// TraceQueryEnd é chamado quando uma query termina de ser executada
func (tc *TracerConfig) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	// Implementação do tracer
}
