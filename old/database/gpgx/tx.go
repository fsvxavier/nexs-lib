package gpgx

import (
	"context"

	"github.com/dock-tech/isis-golang-lib/observability/logger"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type TXConn struct {
	tx                 pgx.Tx
	multiTenantEnabled bool
}

func (pgxt *TXConn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if pgxt.tx == nil {
		return ErrNoTransaction
	}
	rows, err := pgxt.tx.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return pgxscan.ScanOne(dst, rows)
}

func (pgxt *TXConn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if pgxt.tx == nil {
		return ErrNoTransaction
	}
	rows, err := pgxt.tx.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return pgxscan.ScanAll(dst, rows)
}

func (pgxt *TXConn) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return pgxt.tx.Query(ctx, query, args...)
}

func (pgxt *TXConn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {

	var counter int

	if pgxt.tx == nil {
		return nil, ErrNoTransaction
	}
	rows, err := pgxt.tx.Query(ctx, query, args...)
	if err != nil && !NewPgError(err.Error()).IsEmptyResult() {
		return nil, err
	}

	if err != nil && NewPgError(err.Error()).IsEmptyResult() {
		return &counter, nil
	}

	// Only defer rows.Close() after we know rows is not nil
	defer rows.Close()

	for rows.Next() {
		counter++
	}

	return &counter, nil
}

func (pgxt *TXConn) QueryRow(ctx context.Context, query string, args ...interface{}) (IRow, error) {
	if pgxt.tx == nil {
		return nil, ErrNoTransaction
	}
	row := pgxt.tx.QueryRow(ctx, query, args...)
	return NewPgxRow(row), nil
}

func (pgxt *TXConn) QueryRows(ctx context.Context, query string, args ...interface{}) (IRows, error) {
	if pgxt.tx == nil {
		return nil, ErrNoTransaction
	}

	rows, err := pgxt.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return NewPgxRows(rows), nil
}

func (pgxt *TXConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if pgxt.tx == nil {
		return ErrNoTransaction
	}
	_, err := pgxt.tx.Exec(ctx, query, args...)
	return err
}

func (pgxt *TXConn) SendBatch(ctx context.Context, batch IBatch) (IBatchResults, error) {
	if pgxt.tx == nil {
		return nil, ErrNoTransaction
	}

	batchResults := pgxt.tx.SendBatch(ctx, batch.getBatch())

	return NewPgxBatchResults(batchResults), nil
}

func (pgx *TXConn) BeginTransaction(ctx context.Context) (ITransaction, error) {
	return nil, ErrInvalidNestedTransaction
}

func (pgxt *TXConn) Release(ctx context.Context) {}

func (pgxt *TXConn) Commit(ctx context.Context) error {
	if pgxt.tx == nil {
		return ErrNoTransaction
	}
	err := pgxt.tx.Commit(ctx)
	pgxt.tx = nil
	return err
}

func (pgxt *TXConn) Rollback(ctx context.Context) error {
	if pgxt.tx == nil {
		return ErrNoTransaction
	}
	err := pgxt.tx.Rollback(ctx)
	pgxt.tx = nil
	return err
}

func (pgxt *TXConn) Ping(ctx context.Context) error {
	if pgxt.tx == nil {
		return ErrNoConnection
	}
	rows, errPing := pgxt.tx.Query(ctx, `SELECT 1`)
	if errPing != nil {
		return errPing
	}
	defer rows.Close()
	return nil
}

func (pgxt *TXConn) BeforeReleaseHook(ctx context.Context) (err error) {
	if tid := ctx.Value("tenant_id"); tid != nil && tid != "" {
		_, err = pgxt.tx.Exec(ctx, "select set_config($1,$2,$3);", "app.current_tenant", "", false)
		if err != nil {
			logger.Errorf(ctx, "Failed to unset tenant ID for session: %s", err.Error())
			return err
		}
	}
	return nil
}

func (pgxt *TXConn) AfterAcquireHook(ctx context.Context) (err error) {
	if tid := ctx.Value("tenant_id"); tid != nil && tid != "" {
		if tenantID, ok := tid.(string); ok {
			_, err = pgxt.tx.Exec(ctx, "select set_config($1,$2,$3);", "app.current_tenant", tenantID, false)
			if err != nil {
				logger.Errorf(ctx, "Failed to set tenant ID for session: %s\n", err.Error())
				return err
			}
		}
	}

	return nil
}
