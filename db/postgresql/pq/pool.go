package pq

import (
	"context"
	"database/sql"
	"sync"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	_ "github.com/lib/pq" // Driver PostgreSQL
)

// Pool implementa a interface common.IPool usando database/sql e lib/pq
type Pool struct {
	db                 *sql.DB
	config             *common.Config
	mu                 sync.RWMutex
	multiTenantEnabled bool
}

// NewPool cria um novo pool de conexões PostgreSQL usando lib/pq
func NewPool(ctx context.Context, config *common.Config) (common.IPool, error) {
	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return nil, err
	}

	// Configura o pool
	db.SetMaxOpenConns(int(config.MaxConns))
	db.SetMaxIdleConns(int(config.MinConns))
	db.SetConnMaxLifetime(config.MaxConnLifetime)
	db.SetConnMaxIdleTime(config.MaxConnIdleTime)

	// Verifica a conexão
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &Pool{
		db:                 db,
		config:             config,
		multiTenantEnabled: config.MultiTenantEnabled,
	}, nil
}

// Acquire adquire uma conexão do pool
func (p *Pool) Acquire(ctx context.Context) (common.IConn, error) {
	conn, err := p.db.Conn(ctx)
	if err != nil {
		return nil, err
	}

	// Configura o fuso horário para UTC
	_, err = conn.ExecContext(ctx, "SET timezone TO 'UTC'")
	if err != nil {
		conn.Close()
		return nil, err
	}

	pqConn := &Conn{
		conn:               conn,
		multiTenantEnabled: p.multiTenantEnabled,
		config:             p.config,
	}

	// Se multi-tenant está habilitado, configura o contexto do tenant
	if p.multiTenantEnabled {
		if err := pqConn.setupTenantContext(ctx); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return pqConn, nil
}

// Close fecha o pool de conexões
func (p *Pool) Close() error {
	return p.db.Close()
}

// Ping verifica a conectividade com o banco de dados
func (p *Pool) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Stats retorna estatísticas do pool de conexões
func (p *Pool) Stats() *common.PoolStats {
	stats := p.db.Stats()

	return &common.PoolStats{
		AcquiredConns:    stats.InUse,
		TotalConns:       stats.OpenConnections,
		IdleConns:        stats.Idle,
		MaxConns:         int(p.config.MaxConns),
		MinConns:         int(p.config.MinConns),
		ConstructedConns: 0, // O pacote database/sql não fornece esta informação
	}
}
