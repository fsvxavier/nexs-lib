package gorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Conn implementa a interface common.IConn usando GORM
type Conn struct {
	db                 *gorm.DB
	multiTenantEnabled bool
	config             *common.Config
}

// NewConn cria uma nova conexão com o PostgreSQL usando GORM
func NewConn(ctx context.Context, config *common.Config) (common.IConn, error) {
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

	conn := &Conn{
		db:                 db,
		multiTenantEnabled: config.MultiTenantEnabled,
		config:             config,
	}

	// Se multi-tenant está habilitado, configura o contexto do tenant
	if config.MultiTenantEnabled {
		if err := conn.setupTenantContext(ctx); err != nil {
			sqlDB.Close()
			return nil, err
		}
	}

	return conn, nil
}

// setupTenantContext configura o contexto do tenant para a conexão
func (c *Conn) setupTenantContext(ctx context.Context) error {
	// Implementação específica para multi-tenant, se necessário
	return nil
}

// QueryOne executa uma consulta e escaneia o primeiro resultado no destino
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := c.db.WithContext(ctx).Raw(query, args...).Scan(dst)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar QueryOne")
	}

	if result.RowsAffected == 0 {
		return common.ErrNoRows
	}

	return nil
}

// QueryAll executa uma consulta e escaneia todos os resultados no destino
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := c.db.WithContext(ctx).Raw(query, args...).Scan(dst)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar QueryAll")
	}

	return nil
}

// QueryCount executa uma consulta que retorna uma contagem
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int
	result := c.db.WithContext(ctx).Raw(query, args...).Scan(&count)
	if result.Error != nil {
		return nil, WrapError(result.Error, "falha ao executar QueryCount")
	}

	return &count, nil
}

// Query executa uma consulta e retorna as linhas
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (common.IRows, error) {
	rows, err := c.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, WrapError(err, "falha ao executar Query")
	}

	return &Rows{rows: rows}, nil
}

// QueryRow executa uma consulta e retorna uma única linha
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) (common.IRow, error) {
	row := c.db.WithContext(ctx).Raw(query, args...).Row()
	if row.Err() != nil {
		return nil, WrapError(row.Err(), "falha ao executar QueryRow")
	}

	return &Row{row: row}, nil
}

// Exec executa um comando SQL
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	result := c.db.WithContext(ctx).Exec(query, args...)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar Exec")
	}

	return nil
}

// SendBatch envia um lote de consultas para execução
func (c *Conn) SendBatch(ctx context.Context, batch common.IBatch) (common.IBatchResults, error) {
	// GORM não possui uma API direta para lotes como pgx, então vamos implementar
	// uma versão básica usando transações
	gormBatch, ok := batch.GetBatch().(*Batch)
	if !ok {
		return nil, errors.New("batch inválido: esperado um batch do tipo GORM")
	}

	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, WrapError(tx.Error, "falha ao iniciar transação para batch")
	}

	results := make([]*gorm.DB, len(gormBatch.queries))
	for i, query := range gormBatch.queries {
		results[i] = tx.Exec(query.sql, query.args...)
		if results[i].Error != nil {
			tx.Rollback()
			return nil, WrapError(results[i].Error, fmt.Sprintf("falha ao executar query %d no batch", i))
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, WrapError(err, "falha ao fazer commit do batch")
	}

	return &BatchResults{
		results: results,
		index:   0,
	}, nil
}

// Ping verifica a conexão com o banco de dados
func (c *Conn) Ping(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return WrapError(err, "falha ao obter DB para ping")
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return WrapError(err, "falha no ping ao banco de dados")
	}

	return nil
}

// Close fecha a conexão com o banco de dados
func (c *Conn) Close(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return WrapError(err, "falha ao obter DB para fechamento")
	}

	if err := sqlDB.Close(); err != nil {
		return WrapError(err, "falha ao fechar conexão")
	}

	return nil
}

// BeginTransaction inicia uma nova transação
func (c *Conn) BeginTransaction(ctx context.Context) (common.ITransaction, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, WrapError(tx.Error, "falha ao iniciar transação")
	}

	return &Transaction{
		db:                 tx,
		multiTenantEnabled: c.multiTenantEnabled,
		config:             c.config,
	}, nil
}

// BeginTransactionWithOptions inicia uma nova transação com opções
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, opts *common.TxOptions) (common.ITransaction, error) {
	// Configura as opções de transação do GORM
	txOpts := &sql.TxOptions{}

	if opts != nil {
		// Configura nível de isolamento
		switch opts.IsoLevel {
		case common.ReadCommitted:
			txOpts.Isolation = sql.LevelReadCommitted
		case common.RepeatableRead:
			txOpts.Isolation = sql.LevelRepeatableRead
		case common.Serializable:
			txOpts.Isolation = sql.LevelSerializable
		}

		// Configura modo de leitura
		txOpts.ReadOnly = opts.ReadOnly
	}

	// Inicia a transação com as opções
	tx := c.db.WithContext(ctx).Session(&gorm.Session{
		SkipDefaultTransaction: true,
	}).Begin(txOpts)

	if tx.Error != nil {
		return nil, WrapError(tx.Error, "falha ao iniciar transação com opções")
	}

	return &Transaction{
		db:                 tx,
		multiTenantEnabled: c.multiTenantEnabled,
		config:             c.config,
	}, nil
}

// Model retorna o objeto DB do GORM para operações com modelos
func (c *Conn) Model(model interface{}) *gorm.DB {
	return c.db.Model(model)
}

// Create cria um novo registro usando GORM
func (c *Conn) Create(ctx context.Context, model interface{}) error {
	result := c.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao criar registro")
	}
	return nil
}

// Find busca registros usando GORM
func (c *Conn) Find(ctx context.Context, dest interface{}, conds ...interface{}) error {
	result := c.db.WithContext(ctx).Find(dest, conds...)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao buscar registros")
	}
	return nil
}

// First busca o primeiro registro usando GORM
func (c *Conn) First(ctx context.Context, dest interface{}, conds ...interface{}) error {
	result := c.db.WithContext(ctx).First(dest, conds...)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return common.ErrNoRows
		}
		return WrapError(result.Error, "falha ao buscar primeiro registro")
	}
	return nil
}

// Update atualiza registros usando GORM
func (c *Conn) Update(ctx context.Context, model interface{}, columns map[string]interface{}) error {
	result := c.db.WithContext(ctx).Model(model).Updates(columns)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao atualizar registro")
	}
	return nil
}

// Delete exclui registros usando GORM
func (c *Conn) Delete(ctx context.Context, model interface{}, conds ...interface{}) error {
	result := c.db.WithContext(ctx).Delete(model, conds...)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao excluir registro")
	}
	return nil
}
