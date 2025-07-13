package mocks

import (
	"context"
	"errors"

	"github.com/dock-tech/isis-golang-lib/database/gpgx"
	"github.com/dock-tech/isis-golang-lib/observability/logger"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

type PgxConnMock struct {
	pgxmock.PgxConnIface
}

func (mock *PgxConnMock) Release(ctx context.Context) {}

func (mock *PgxConnMock) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := mock.PgxConnIface.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	return pgxscan.ScanOne(dst, rows)
}

func (mock *PgxConnMock) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (mock *PgxConnMock) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := mock.PgxConnIface.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	return pgxscan.ScanAll(dst, rows)
}

func (mock *PgxConnMock) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := mock.PgxConnIface.Exec(ctx, query, args...)
	return err
}

func (mock *PgxConnMock) BeginTransaction(ctx context.Context) (gpgx.ITransaction, error) {
	_, err := mock.PgxConnIface.Begin(ctx)
	return &gpgx.TXConn{}, err
}

func (m *PgxConnMock) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	if m.PgxConnIface == nil {
		counter := 0
		return &counter, nil
	}

	rows, err := m.PgxConnIface.Query(ctx, query, args...)
	if err != nil {
		if err.Error() == "no rows in result set" {
			counter := 0
			return &counter, nil
		}
		return nil, err
	}

	counter := 0
	for rows.Next() {
		counter++
	}

	return &counter, nil
}

func (m *PgxConnMock) QueryRow(ctx context.Context, query string, args ...interface{}) (gpgx.IRow, error) {

	if m.PgxConnIface == nil {
		return nil, errors.New("no connection")
	}

	return m.PgxConnIface.QueryRow(ctx, query, args...), nil
}

func (m *PgxConnMock) QueryRows(ctx context.Context, query string, args ...interface{}) (gpgx.IRows, error) {

	if m.PgxConnIface == nil {
		return nil, errors.New("no connection")
	}

	rows, err := m.PgxConnIface.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (m *PgxConnMock) Ping(ctx context.Context) error {
	if m.PgxConnIface == nil {
		return errors.New("no connection")
	}
	rows, errPing := m.PgxConnIface.Query(ctx, `SELECT 1`)
	if errPing != nil {
		return errPing
	}
	defer rows.Close()
	return nil
}

func (m *PgxConnMock) SendBatch(ctx context.Context, batch gpgx.IBatch) (gpgx.IBatchResults, error) {

	if m.PgxConnIface == nil {
		return nil, errors.New("no connection")
	}

	return nil, nil
}

func (m *PgxConnMock) AfterAcquireHook(ctx context.Context) error {
	if tid := ctx.Value("tenant_id"); tid != nil && tid != "" {
		_, err := m.PgxConnIface.Exec(ctx, "SELECT set_config($1,$2,$3);", "app.current_tenant", "", false)
		if err != nil {
			logger.Errorf(ctx, "Failed to unset tenant ID for session: %s", err.Error())
			return err
		} else {
			logger.Debugf(ctx, "Unset tenant ID for session")
		}
	}
	return nil
}

func (m *PgxConnMock) BeforeReleaseHook(ctx context.Context) (err error) {
	if tid := ctx.Value("tenant_id"); tid != nil && tid != "" {
		_, err := m.PgxConnIface.Exec(ctx, "SELECT set_config($1,$2,$3);", "app.current_tenant", "", false)
		if err != nil {
			logger.Errorf(ctx, "Failed to unset tenant ID for session: %s", err.Error())
			return err
		} else {
			logger.Debugf(ctx, "Unset tenant ID for session")
		}
	}
	return nil
}
