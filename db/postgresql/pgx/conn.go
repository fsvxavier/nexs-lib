package pgx

import (
	"context"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Conn implementa a interface common.IConn usando pgx
type Conn struct {
	poolConn           *pgxpool.Conn
	conn               *pgx.Conn
	mockConn           interface{} // Para mock em testes
	multiTenantEnabled bool
	config             *common.Config
}

// NewConn cria uma nova conexão direta (sem pool) com o PostgreSQL usando pgx
func NewConn(ctx context.Context, config *common.Config) (common.IConn, error) {
	pgxConfig, err := pgx.ParseConfig(config.ConnectionString())
	if err != nil {
		return nil, err
	}

	// Configura outros parâmetros
	pgxConfig.RuntimeParams["timezone"] = "UTC"

	// Configura tracer se habilitado
	if config.TraceEnabled {
		pgxConfig.Tracer = NewTracer(config)
	}

	conn, err := pgx.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	pgxConn := &Conn{
		conn:               conn,
		multiTenantEnabled: config.MultiTenantEnabled,
		config:             config,
	}

	// Se multi-tenant está habilitado, configura o contexto do tenant
	if config.MultiTenantEnabled {
		if err := pgxConn.setupTenantContext(ctx); err != nil {
			conn.Close(ctx)
			return nil, err
		}
	}

	return pgxConn, nil
}

// QueryOne executa uma consulta e digitaliza uma única linha no destino
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.poolConn != nil {
		rows, err := c.poolConn.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		return pgxscan.ScanOne(dst, rows)
	}

	if c.conn != nil {
		rows, err := c.conn.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		return pgxscan.ScanOne(dst, rows)
	}

	return common.ErrNoConnection
}

// QueryAll executa uma consulta e digitaliza todas as linhas no destino
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.poolConn != nil {
		rows, err := c.poolConn.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		return pgxscan.ScanAll(dst, rows)
	}

	if c.conn != nil {
		rows, err := c.conn.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		return pgxscan.ScanAll(dst, rows)
	}

	return common.ErrNoConnection
}

// QueryCount executa uma consulta e retorna uma contagem
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int

	row, err := c.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	err = row.Scan(&count)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// Query executa uma consulta e retorna as linhas resultantes
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (common.IRows, error) {
	if c.poolConn != nil {
		rows, err := c.poolConn.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		return &Rows{rows: rows}, nil
	}

	if c.conn != nil {
		rows, err := c.conn.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		return &Rows{rows: rows}, nil
	}

	return nil, common.ErrNoConnection
}

// QueryRow executa uma consulta e retorna uma única linha
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) (common.IRow, error) {
	if c.poolConn != nil {
		row := c.poolConn.QueryRow(ctx, query, args...)
		return &Row{row: row}, nil
	}

	if c.conn != nil {
		row := c.conn.QueryRow(ctx, query, args...)
		return &Row{row: row}, nil
	}

	return nil, common.ErrNoConnection
}

// Exec executa um comando (sem retorno de linhas)
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if c.poolConn != nil {
		_, err := c.poolConn.Exec(ctx, query, args...)
		return err
	}

	if c.conn != nil {
		_, err := c.conn.Exec(ctx, query, args...)
		return err
	}

	return common.ErrNoConnection
}

// SendBatch envia um lote de comandos para execução
func (c *Conn) SendBatch(ctx context.Context, batch common.IBatch) (common.IBatchResults, error) {
	pgxBatch, ok := batch.GetBatch().(*pgx.Batch)
	if !ok {
		return nil, common.ErrInvalidOperation
	}

	if c.poolConn != nil {
		br := c.poolConn.SendBatch(ctx, pgxBatch)
		return &BatchResults{br: br}, nil
	}

	if c.conn != nil {
		br := c.conn.SendBatch(ctx, pgxBatch)
		return &BatchResults{br: br}, nil
	}

	return nil, common.ErrNoConnection
}

// Ping verifica a conectividade com o banco de dados
func (c *Conn) Ping(ctx context.Context) error {
	if c.poolConn != nil {
		return c.poolConn.Ping(ctx)
	}

	if c.conn != nil {
		return c.conn.Ping(ctx)
	}

	return common.ErrNoConnection
}

// Close fecha a conexão
func (c *Conn) Close(ctx context.Context) error {
	if c.poolConn != nil {
		c.poolConn.Release()
		return nil
	}

	if c.conn != nil {
		return c.conn.Close(ctx)
	}

	return common.ErrNoConnection
}

// BeginTransaction inicia uma transação
func (c *Conn) BeginTransaction(ctx context.Context) (common.ITransaction, error) {
	return c.BeginTransactionWithOptions(ctx, nil)
}

// BeginTransactionWithOptions inicia uma transação com opções específicas
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, opts *common.TxOptions) (common.ITransaction, error) {
	var pgxOpts pgx.TxOptions

	if opts != nil {
		// Mapeia os níveis de isolamento
		switch opts.IsoLevel {
		case common.ReadCommitted:
			pgxOpts.IsoLevel = pgx.ReadCommitted
		case common.RepeatableRead:
			pgxOpts.IsoLevel = pgx.RepeatableRead
		case common.Serializable:
			pgxOpts.IsoLevel = pgx.Serializable
		default:
			pgxOpts.IsoLevel = pgx.ReadCommitted
		}

		// Configura outras opções
		if opts.ReadOnly {
			pgxOpts.AccessMode = pgx.ReadOnly
		} else {
			pgxOpts.AccessMode = pgx.ReadWrite
		}

		if opts.Deferrable {
			pgxOpts.DeferrableMode = pgx.Deferrable
		} else {
			pgxOpts.DeferrableMode = pgx.NotDeferrable
		}
	}

	if c.poolConn != nil {
		tx, err := c.poolConn.BeginTx(ctx, pgxOpts)
		if err != nil {
			return nil, err
		}
		return &Transaction{tx: tx, config: c.config, multiTenantEnabled: c.multiTenantEnabled}, nil
	}

	if c.conn != nil {
		tx, err := c.conn.BeginTx(ctx, pgxOpts)
		if err != nil {
			return nil, err
		}
		return &Transaction{tx: tx, config: c.config, multiTenantEnabled: c.multiTenantEnabled}, nil
	}

	return nil, common.ErrNoConnection
}

// setupTenantContext configura o contexto do tenant para multi-tenancy
func (c *Conn) setupTenantContext(ctx context.Context) error {
	// Implementação da configuração do tenant
	// Este é um esboço - a implementação real dependeria da sua abordagem de multi-tenancy
	tenantID := getTenantIDFromContext(ctx)
	if tenantID == "" {
		return nil // Nenhum tenant especificado, não é necessário configurar
	}

	// Exemplo: definir um parâmetro de sessão para o tenant atual
	query := "SET app.tenant_id = $1"
	return c.Exec(ctx, query, tenantID)
}

// getTenantIDFromContext extrai o ID do tenant do contexto
func getTenantIDFromContext(ctx context.Context) string {
	// Implementação real dependeria da sua estrutura de contexto
	// Exemplo simples:
	type tenantKey struct{}
	if tenant, ok := ctx.Value(tenantKey{}).(string); ok {
		return tenant
	}
	return ""
}
