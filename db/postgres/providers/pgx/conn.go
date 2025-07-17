package pgxprovider

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/monitoring"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Conn implementa IConn usando pgx.Conn ou pgxpool.Conn
type Conn struct {
	conn        interface{} // *pgx.Conn ou *pgxpool.Conn
	config      interfaces.IConfig
	bufferPool  interfaces.IBufferPool
	hookManager interfaces.IHookManager
	monitor     *monitoring.ConnectionMonitor
	mu          sync.RWMutex
	acquired    bool
	fromPool    bool
	closed      bool
	tenantID    string
}

// getConn retorna a conexão com a interface adequada
func (c *Conn) getConn() pgxConnInterface {
	if pgxConn, ok := c.conn.(*pgx.Conn); ok {
		return pgxConn
	}
	if poolConn, ok := c.conn.(*pgxpool.Conn); ok {
		return poolConn
	}
	return nil
}

// NewConn cria uma nova conexão PGX
func NewConn(ctx context.Context, config interfaces.IConfig,
	bufferPool interfaces.IBufferPool,
	hookManager interfaces.IHookManager) (interfaces.IConn, error) {

	// Parse da string de conexão
	pgxConfig, err := pgx.ParseConfig(config.GetConnectionString())
	if err != nil {
		return nil, err
	}

	// Criar conexão PGX
	pgxConn, err := pgx.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}

	// Criar wrapper para conexão direta (não do pool)
	conn := &Conn{
		conn:        pgxConn,
		config:      config,
		bufferPool:  bufferPool,
		hookManager: hookManager,
		monitor:     monitoring.NewConnectionMonitor(),
		acquired:    false,
		fromPool:    false,
		closed:      false,
	}

	return conn, nil
}

// QueryRow executa uma query que retorna uma linha
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) interfaces.IRow {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return &Row{err: ErrConnClosed}
	}

	// Executar hook de query
	if c.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "query_row",
			Query:     query,
			Args:      args,
			StartTime: time.Now(),
		}
		if err := c.hookManager.ExecuteHooks(interfaces.BeforeQueryHook, execCtx); err != nil {
			return &Row{err: err}
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			c.hookManager.ExecuteHooks(interfaces.AfterQueryHook, execCtx)
		}()
	}

	row := c.getConn().QueryRow(ctx, query, args...)
	return &Row{row: row}
}

// Query executa uma query que retorna múltiplas linhas
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnClosed
	}

	// Executar hook de query
	if c.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "query",
			Query:     query,
			Args:      args,
			StartTime: time.Now(),
		}
		if err := c.hookManager.ExecuteHooks(interfaces.BeforeQueryHook, execCtx); err != nil {
			return nil, err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			c.hookManager.ExecuteHooks(interfaces.AfterQueryHook, execCtx)
		}()
	}

	rows, err := c.getConn().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{rows: rows}, nil
}

// QueryOne executa uma query que retorna um único resultado
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	row := c.QueryRow(ctx, query, args...)
	return row.Scan(dst)
}

// QueryAll executa uma query que retorna todos os resultados usando reflection
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrConnClosed
	}
	c.mu.RUnlock()

	// Executar hook de query
	if c.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "query_all",
			Query:     query,
			Args:      args,
			StartTime: time.Now(),
		}
		if err := c.hookManager.ExecuteHooks(interfaces.BeforeQueryHook, execCtx); err != nil {
			return err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			c.hookManager.ExecuteHooks(interfaces.AfterQueryHook, execCtx)
		}()
	}

	// Usar implementação com reflection do arquivo reflection.go
	return c.queryAllWithReflection(ctx, dst, query, args...)
}

// QueryCount executa uma query que retorna contagem
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var count int64
	err := c.QueryOne(ctx, &count, query, args...)
	return count, err
}

// Exec executa uma query que não retorna resultados
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) (interfaces.ICommandTag, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnClosed
	}

	// Executar hook de exec
	if c.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "exec",
			Query:     query,
			Args:      args,
			StartTime: time.Now(),
		}
		if err := c.hookManager.ExecuteHooks(interfaces.BeforeExecHook, execCtx); err != nil {
			return nil, err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			c.hookManager.ExecuteHooks(interfaces.AfterExecHook, execCtx)
		}()
	}

	cmdTag, err := c.getConn().Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &CommandTag{tag: cmdTag}, nil
}

// SendBatch envia um batch de comandos
func (c *Conn) SendBatch(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return &BatchResults{err: ErrConnClosed}
	}

	// Converter o batch para pgx.Batch
	pgxBatch, ok := batch.(*Batch)
	if !ok {
		return &BatchResults{err: errors.New("invalid batch type")}
	}

	batchResults := c.getConn().SendBatch(ctx, pgxBatch.batch)
	return &BatchResults{results: batchResults}
}

// Begin inicia uma transação
func (c *Conn) Begin(ctx context.Context) (interfaces.ITransaction, error) {
	return c.BeginTx(ctx, interfaces.TxOptions{})
}

// BeginTx inicia uma transação com opções
func (c *Conn) BeginTx(ctx context.Context, txOptions interfaces.TxOptions) (interfaces.ITransaction, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnClosed
	}

	// Converter opções para pgx.TxOptions
	pgxTxOptions := pgx.TxOptions{
		IsoLevel:       pgx.TxIsoLevel(txOptions.IsoLevel),
		AccessMode:     pgx.TxAccessMode(txOptions.AccessMode),
		DeferrableMode: pgx.TxDeferrableMode(txOptions.DeferrableMode),
	}

	tx, err := c.getConn().BeginTx(ctx, pgxTxOptions)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		tx:          tx,
		config:      c.config,
		bufferPool:  c.bufferPool,
		hookManager: c.hookManager,
		monitor:     c.monitor,
	}, nil
}

// Release libera a conexão (se for do pool)
func (c *Conn) Release() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Para conexões diretas, não há nada a fazer
	// Este método é mantido para compatibilidade com a interface
	if c.fromPool && c.acquired {
		c.acquired = false
	}
}

// Close fecha a conexão
func (c *Conn) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	var err error
	if !c.fromPool {
		if directConn, ok := c.conn.(*pgx.Conn); ok {
			err = directConn.Close(ctx)
		}
	} else {
		// Para conexões do pool, chamamos Release
		if poolConn, ok := c.conn.(*pgxpool.Conn); ok {
			poolConn.Release()
		}
	}

	c.closed = true
	return err
}

// Ping verifica conectividade
func (c *Conn) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnClosed
	}

	return c.getConn().Ping(ctx)
}

// IsClosed verifica se a conexão está fechada
func (c *Conn) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// Prepare prepara uma declaração
func (c *Conn) Prepare(ctx context.Context, name, query string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnClosed
	}

	// Para pgxpool.Conn, precisamos usar a conexão direta
	if poolConn, ok := c.conn.(*pgxpool.Conn); ok {
		_, err := poolConn.Conn().Prepare(ctx, name, query)
		return err
	}

	// Para pgx.Conn direta
	if directConn, ok := c.conn.(*pgx.Conn); ok {
		_, err := directConn.Prepare(ctx, name, query)
		return err
	}

	return errors.New("unsupported connection type")
}

// Deallocate remove uma declaração preparada
func (c *Conn) Deallocate(ctx context.Context, name string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnClosed
	}

	_, err := c.getConn().Exec(ctx, "DEALLOCATE "+name)
	return err
}

// CopyFrom copia dados de uma fonte para uma tabela
func (c *Conn) CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.ICopyFromSource) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrConnClosed
	}

	// Converter rowSrc para pgx.CopyFromSource
	pgxRowSrc, ok := rowSrc.(*CopyFromSource)
	if !ok {
		return 0, errors.New("invalid copy from source type")
	}

	return c.getConn().CopyFrom(ctx, pgx.Identifier{tableName}, columnNames, pgxRowSrc.source)
}

// CopyTo copia dados de uma query para um writer
func (c *Conn) CopyTo(ctx context.Context, w interfaces.ICopyToWriter, query string, args ...interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnClosed
	}

	// Para pgx.Conn, não há CopyTo direto. Vamos implementar uma versão básica
	rows, err := c.getConn().Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return err
		}
		if err := w.Write(values); err != nil {
			return err
		}
	}

	return rows.Err()
}

// Listen escuta notificações em um canal
func (c *Conn) Listen(ctx context.Context, channel string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnClosed
	}

	_, err := c.getConn().Exec(ctx, "LISTEN "+channel)
	return err
}

// Unlisten para de escutar notificações em um canal
func (c *Conn) Unlisten(ctx context.Context, channel string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnClosed
	}

	_, err := c.getConn().Exec(ctx, "UNLISTEN "+channel)
	return err
}

// WaitForNotification espera por uma notificação
func (c *Conn) WaitForNotification(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnClosed
	}

	// Para pgxpool.Conn, precisamos usar a conexão direta
	if poolConn, ok := c.conn.(*pgxpool.Conn); ok {
		notification, err := poolConn.Conn().WaitForNotification(ctx)
		if err != nil {
			return nil, err
		}
		return &interfaces.Notification{
			PID:     notification.PID,
			Channel: notification.Channel,
			Payload: notification.Payload,
		}, nil
	}

	// Para pgx.Conn direta
	if directConn, ok := c.conn.(*pgx.Conn); ok {
		notification, err := directConn.WaitForNotification(ctx)
		if err != nil {
			return nil, err
		}
		return &interfaces.Notification{
			PID:     notification.PID,
			Channel: notification.Channel,
			Payload: notification.Payload,
		}, nil
	}

	return nil, errors.New("unsupported connection type")
}

// SetTenant define o tenant atual
func (c *Conn) SetTenant(ctx context.Context, tenantID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tenantID = tenantID
	// Aqui você pode implementar lógica específica de multi-tenancy
	// Por exemplo, definir search_path ou variáveis de sessão
	return nil
}

// GetTenant obtém o tenant atual
func (c *Conn) GetTenant(ctx context.Context) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.tenantID, nil
}

// GetHookManager retorna o hook manager
func (c *Conn) GetHookManager() interfaces.IHookManager {
	return c.hookManager
}

// HealthCheck verifica saúde da conexão
func (c *Conn) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx)
}

// Stats retorna estatísticas da conexão
func (c *Conn) Stats() interfaces.ConnectionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return interfaces.ConnectionStats{
		TotalQueries:       0, // Seria necessário implementar contadores
		TotalExecs:         0,
		TotalTransactions:  0,
		TotalBatches:       0,
		FailedQueries:      0,
		FailedExecs:        0,
		FailedTransactions: 0,
		AverageQueryTime:   0,
		AverageExecTime:    0,
		LastActivity:       time.Now(),
		CreatedAt:          time.Now(),
		MemoryUsage:        interfaces.MemoryStats{},
	}
}
