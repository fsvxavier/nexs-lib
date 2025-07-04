package gorm

import (
	"context"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Pool implementa a interface common.IPool usando GORM
type Pool struct {
	db     *gorm.DB
	config *common.Config
}

// NewPool cria um novo pool de conexões com o PostgreSQL usando GORM
func NewPool(ctx context.Context, config *common.Config) (common.IPool, error) {
	// Configura o logger do GORM com base nas configurações
	logLevel := logger.Silent
	if config.QueryLogEnabled {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	// Conecta ao banco de dados
	db, err := gorm.Open(postgres.Open(config.ConnectionString()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar com PostgreSQL via GORM: %w", err)
	}

	// Configura o pool de conexões
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter o pool de conexões: %w", err)
	}

	// Configura os limites do pool
	sqlDB.SetMaxOpenConns(int(config.MaxConns))
	sqlDB.SetMaxIdleConns(int(config.MinConns))
	sqlDB.SetConnMaxLifetime(config.MaxConnLifetime)
	sqlDB.SetConnMaxIdleTime(config.MaxConnIdleTime)

	return &Pool{
		db:     db,
		config: config,
	}, nil
}

// Acquire obtém uma conexão do pool
func (p *Pool) Acquire(ctx context.Context) (common.IConn, error) {
	// GORM não tem um método direto para adquirir uma conexão do pool,
	// então vamos criar uma nova instância de Conn com o mesmo DB
	return &Conn{
		db:                 p.db.WithContext(ctx),
		multiTenantEnabled: p.config.MultiTenantEnabled,
		config:             p.config,
	}, nil
}

// Close fecha o pool de conexões
func (p *Pool) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return WrapError(err, "falha ao obter DB para fechamento")
	}

	if err := sqlDB.Close(); err != nil {
		return WrapError(err, "falha ao fechar pool")
	}

	return nil
}

// Ping verifica a conexão com o banco de dados
func (p *Pool) Ping(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return WrapError(err, "falha ao obter DB para ping")
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return WrapError(err, "falha no ping ao banco de dados")
	}

	return nil
}

// Stats retorna estatísticas do pool
func (p *Pool) Stats() *common.PoolStats {
	sqlDB, err := p.db.DB()
	if err != nil {
		// Se houver erro, retorna estatísticas zeradas
		return &common.PoolStats{}
	}

	stats := sqlDB.Stats()
	return &common.PoolStats{
		AcquiredConns: stats.InUse,
		TotalConns:    stats.OpenConnections,
		IdleConns:     stats.Idle,
	}
}
