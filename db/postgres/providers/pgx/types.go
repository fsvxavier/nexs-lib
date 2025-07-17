package pgxprovider

import (
	"context"
	"errors"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx/internal/monitoring"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Row implementa IRow usando pgx.Row
type Row struct {
	row pgx.Row
	err error
}

// Scan implementa IRow.Scan
func (r *Row) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return r.row.Scan(dest...)
}

// Rows implementa IRows usando pgx.Rows
type Rows struct {
	rows pgx.Rows
	err  error
}

// Next implementa IRows.Next
func (r *Rows) Next() bool {
	if r.err != nil {
		return false
	}
	return r.rows.Next()
}

// Scan implementa IRows.Scan
func (r *Rows) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return r.rows.Scan(dest...)
}

// Close implementa IRows.Close
func (r *Rows) Close() error {
	if r.rows != nil {
		r.rows.Close()
	}
	return nil
}

// Err implementa IRows.Err
func (r *Rows) Err() error {
	if r.err != nil {
		return r.err
	}
	if r.rows != nil {
		return r.rows.Err()
	}
	return nil
}

// Values implementa IRows.Values
func (r *Rows) Values() ([]any, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.rows.Values()
}

// CommandTag implementa IRows.CommandTag
func (r *Rows) CommandTag() interfaces.ICommandTag {
	if r.rows != nil {
		return &CommandTag{tag: r.rows.CommandTag()}
	}
	return &CommandTag{}
}

// FieldDescriptions implementa IRows.FieldDescriptions
func (r *Rows) FieldDescriptions() []interfaces.IFieldDescription {
	if r.rows != nil {
		descs := r.rows.FieldDescriptions()
		result := make([]interfaces.IFieldDescription, len(descs))
		for i, desc := range descs {
			result[i] = &FieldDescription{desc: desc}
		}
		return result
	}
	return nil
}

// RawValues implementa IRows.RawValues
func (r *Rows) RawValues() [][]byte {
	if r.rows != nil {
		return r.rows.RawValues()
	}
	return nil
}

// CommandTag implementa ICommandTag usando pgconn.CommandTag
type CommandTag struct {
	tag pgconn.CommandTag
}

// RowsAffected implementa ICommandTag.RowsAffected
func (ct *CommandTag) RowsAffected() int64 {
	return ct.tag.RowsAffected()
}

// String implementa ICommandTag.String
func (ct *CommandTag) String() string {
	return ct.tag.String()
}

// Insert implementa ICommandTag.Insert
func (ct *CommandTag) Insert() bool {
	return ct.tag.String() == "INSERT"
}

// Update implementa ICommandTag.Update
func (ct *CommandTag) Update() bool {
	return ct.tag.String() == "UPDATE"
}

// Delete implementa ICommandTag.Delete
func (ct *CommandTag) Delete() bool {
	return ct.tag.String() == "DELETE"
}

// Select implementa ICommandTag.Select
func (ct *CommandTag) Select() bool {
	return ct.tag.String() == "SELECT"
}

// FieldDescription implementa IFieldDescription
type FieldDescription struct {
	desc pgconn.FieldDescription
}

// Name implementa IFieldDescription.Name
func (fd *FieldDescription) Name() string {
	return fd.desc.Name
}

// TableOID implementa IFieldDescription.TableOID
func (fd *FieldDescription) TableOID() uint32 {
	return fd.desc.TableOID
}

// TableAttributeNumber implementa IFieldDescription.TableAttributeNumber
func (fd *FieldDescription) TableAttributeNumber() uint16 {
	return fd.desc.TableAttributeNumber
}

// DataTypeOID implementa IFieldDescription.DataTypeOID
func (fd *FieldDescription) DataTypeOID() uint32 {
	return fd.desc.DataTypeOID
}

// DataTypeSize implementa IFieldDescription.DataTypeSize
func (fd *FieldDescription) DataTypeSize() int16 {
	return fd.desc.DataTypeSize
}

// TypeModifier implementa IFieldDescription.TypeModifier
func (fd *FieldDescription) TypeModifier() int32 {
	return fd.desc.TypeModifier
}

// Format implementa IFieldDescription.Format
func (fd *FieldDescription) Format() int16 {
	return fd.desc.Format
}

// Transaction implementa ITransaction usando pgx.Tx
type Transaction struct {
	tx          pgx.Tx
	config      interfaces.IConfig
	bufferPool  interfaces.IBufferPool
	hookManager interfaces.IHookManager
	monitor     *monitoring.ConnectionMonitor
}

// QueryRow implementa ITransaction.QueryRow
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) interfaces.IRow {
	// Executar hook de query
	if t.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "tx_query_row",
			Query:     query,
			Args:      args,
			StartTime: time.Now(),
		}
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeQueryHook, execCtx); err != nil {
			return &Row{err: err}
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			t.hookManager.ExecuteHooks(interfaces.AfterQueryHook, execCtx)
		}()
	}

	row := t.tx.QueryRow(ctx, query, args...)
	return &Row{row: row}
}

// Query implementa ITransaction.Query
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	// Executar hook de query
	if t.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "tx_query",
			Query:     query,
			Args:      args,
			StartTime: time.Now(),
		}
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeQueryHook, execCtx); err != nil {
			return nil, err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			t.hookManager.ExecuteHooks(interfaces.AfterQueryHook, execCtx)
		}()
	}

	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{rows: rows}, nil
}

// QueryOne implementa ITransaction.QueryOne
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	row := t.QueryRow(ctx, query, args...)
	return row.Scan(dst)
}

// QueryAll implementa ITransaction.QueryAll
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Implementar lógica para popular dst com todos os resultados
	// Por agora, retornar erro não implementado
	return errors.New("QueryAll not implemented yet")
}

// QueryCount implementa ITransaction.QueryCount
func (t *Transaction) QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var count int64
	err := t.QueryOne(ctx, &count, query, args...)
	return count, err
}

// Exec implementa ITransaction.Exec
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (interfaces.ICommandTag, error) {
	// Executar hook de exec
	if t.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "tx_exec",
			Query:     query,
			Args:      args,
			StartTime: time.Now(),
		}
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeExecHook, execCtx); err != nil {
			return nil, err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			t.hookManager.ExecuteHooks(interfaces.AfterExecHook, execCtx)
		}()
	}

	cmdTag, err := t.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &CommandTag{tag: cmdTag}, nil
}

// SendBatch implementa ITransaction.SendBatch
func (t *Transaction) SendBatch(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults {
	// Converter o batch para pgx.Batch
	pgxBatch, ok := batch.(*Batch)
	if !ok {
		return &BatchResults{err: errors.New("invalid batch type")}
	}

	batchResults := t.tx.SendBatch(ctx, pgxBatch.batch)
	return &BatchResults{results: batchResults}
}

// Begin implementa ITransaction.Begin (não aplicável para transações)
func (t *Transaction) Begin(ctx context.Context) (interfaces.ITransaction, error) {
	return nil, errors.New("cannot begin transaction within transaction")
}

// BeginTx implementa ITransaction.BeginTx (não aplicável para transações)
func (t *Transaction) BeginTx(ctx context.Context, txOptions interfaces.TxOptions) (interfaces.ITransaction, error) {
	return nil, errors.New("cannot begin transaction within transaction")
}

// Commit implementa ITransaction.Commit
func (t *Transaction) Commit(ctx context.Context) error {
	// Executar hook de commit
	if t.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "tx_commit",
			StartTime: time.Now(),
		}
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeCommitHook, execCtx); err != nil {
			return err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			t.hookManager.ExecuteHooks(interfaces.AfterCommitHook, execCtx)
		}()
	}

	return t.tx.Commit(ctx)
}

// Rollback implementa ITransaction.Rollback
func (t *Transaction) Rollback(ctx context.Context) error {
	// Executar hook de rollback
	if t.hookManager != nil {
		execCtx := &interfaces.ExecutionContext{
			Context:   ctx,
			Operation: "tx_rollback",
			StartTime: time.Now(),
		}
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeRollbackHook, execCtx); err != nil {
			return err
		}
		defer func() {
			execCtx.Duration = time.Since(execCtx.StartTime)
			t.hookManager.ExecuteHooks(interfaces.AfterRollbackHook, execCtx)
		}()
	}

	return t.tx.Rollback(ctx)
}

// Release implementa ITransaction.Release (não aplicável para transações)
func (t *Transaction) Release() {
	// Não faz nada para transações
}

// Close implementa ITransaction.Close
func (t *Transaction) Close(ctx context.Context) error {
	return t.Rollback(ctx)
}

// Ping implementa ITransaction.Ping (não aplicável para transações)
func (t *Transaction) Ping(ctx context.Context) error {
	return errors.New("ping not available in transaction")
}

// IsClosed implementa ITransaction.IsClosed
func (t *Transaction) IsClosed() bool {
	return false // Transações não têm estado "closed" per se
}

// Prepare implementa ITransaction.Prepare
func (t *Transaction) Prepare(ctx context.Context, name, query string) error {
	_, err := t.tx.Prepare(ctx, name, query)
	return err
}

// Deallocate implementa ITransaction.Deallocate
func (t *Transaction) Deallocate(ctx context.Context, name string) error {
	_, err := t.tx.Exec(ctx, "DEALLOCATE "+name)
	return err
}

// CopyFrom implementa ITransaction.CopyFrom
func (t *Transaction) CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.ICopyFromSource) (int64, error) {
	// Converter rowSrc para pgx.CopyFromSource
	pgxRowSrc, ok := rowSrc.(*CopyFromSource)
	if !ok {
		return 0, errors.New("invalid copy from source type")
	}

	return t.tx.CopyFrom(ctx, pgx.Identifier{tableName}, columnNames, pgxRowSrc.source)
}

// CopyTo implementa ITransaction.CopyTo
func (t *Transaction) CopyTo(ctx context.Context, w interfaces.ICopyToWriter, query string, args ...interface{}) error {
	// Para pgx.Tx, não há CopyTo direto. Vamos implementar uma versão básica
	rows, err := t.tx.Query(ctx, query, args...)
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

// Listen implementa ITransaction.Listen
func (t *Transaction) Listen(ctx context.Context, channel string) error {
	_, err := t.tx.Exec(ctx, "LISTEN "+channel)
	return err
}

// Unlisten implementa ITransaction.Unlisten
func (t *Transaction) Unlisten(ctx context.Context, channel string) error {
	_, err := t.tx.Exec(ctx, "UNLISTEN "+channel)
	return err
}

// WaitForNotification implementa ITransaction.WaitForNotification
func (t *Transaction) WaitForNotification(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error) {
	return nil, errors.New("wait for notification not available in transaction")
}

// SetTenant implementa ITransaction.SetTenant
func (t *Transaction) SetTenant(ctx context.Context, tenantID string) error {
	// Implementar lógica específica de multi-tenancy
	return nil
}

// GetTenant implementa ITransaction.GetTenant
func (t *Transaction) GetTenant(ctx context.Context) (string, error) {
	// Implementar lógica específica de multi-tenancy
	return "", nil
}

// GetHookManager implementa ITransaction.GetHookManager
func (t *Transaction) GetHookManager() interfaces.IHookManager {
	return t.hookManager
}

// HealthCheck implementa ITransaction.HealthCheck
func (t *Transaction) HealthCheck(ctx context.Context) error {
	return errors.New("health check not available in transaction")
}

// Stats implementa ITransaction.Stats
func (t *Transaction) Stats() interfaces.ConnectionStats {
	return interfaces.ConnectionStats{
		TotalQueries:       0,
		TotalExecs:         0,
		TotalTransactions:  1,
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
