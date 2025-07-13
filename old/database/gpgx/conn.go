package gpgx

import (
	"context"
	"crypto/tls"
	"os"
	"strconv"

	"github.com/dock-tech/isis-golang-lib/observability/logger"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGXConn struct {
	poolConn           *pgxpool.Conn
	conn               *pgx.Conn
	multiTenantEnabled bool
}

// NewConn creates a new connection and immediately establishes it.
func NewConn(ctx context.Context, connString string, options ...PgxConfig) (IConn, error) {
	cfg := GetConfig()

	for _, opt := range options {
		opt(cfg)
	}

	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	if enabled, _ := strconv.ParseBool(os.Getenv("DB_TLS_ENABLED")); enabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		config.TLSConfig = tlsConfig
	}

	config.RuntimeParams["timezone"] = "UTC"
	switch os.Getenv("DB_QUERY_MODE_EXEC") {
	case "CACHE_STATEMENT":
		config.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	case "CACHE_DESCRIBE":
		config.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	case "DESCRIBE_EXEC":
		config.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec
	case "SIMPLE_PROTOCOL":
		config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	default:
		config.DefaultQueryExecMode = pgx.QueryExecModeExec
	}

	config.Tracer = &TracerConfig{
		QueryTracerEnabled:   isQueryTracerEnabled(cfg),
		ObservabilityEnabled: isObservabilityEnabled(cfg),
	}

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	pgxConn := &PGXConn{
		conn: conn,
	}

	return pgxConn, nil
}

func (pgc *PGXConn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if pgc.conn == nil {
		return ErrNoConnection
	}
	rows, err := pgc.conn.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return pgxscan.ScanOne(dst, rows)
}

func (pgc *PGXConn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if pgc.conn == nil {
		return ErrNoConnection
	}
	rows, err := pgc.conn.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return pgxscan.ScanAll(dst, rows)
}

func (pgc *PGXConn) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return pgc.conn.Query(ctx, query, args...)
}

func (pgc *PGXConn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {

	var counter int

	if pgc.conn == nil {
		return nil, ErrNoConnection
	}
	rows, err := pgc.conn.Query(ctx, query, args...)
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

func (pgc *PGXConn) QueryRow(ctx context.Context, query string, args ...interface{}) (IRow, error) {
	if pgc.conn == nil {
		return nil, ErrNoConnection
	}
	row := pgc.conn.QueryRow(ctx, query, args...)
	return NewPgxRow(row), nil
}

func (pgc *PGXConn) QueryRows(ctx context.Context, query string, args ...interface{}) (IRows, error) {
	if pgc.conn == nil {
		return nil, ErrNoConnection
	}
	rows, err := pgc.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return NewPgxRows(rows), nil
}

func (pgc *PGXConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if pgc.conn == nil {
		return ErrNoConnection
	}
	_, err := pgc.conn.Exec(ctx, query, args...)
	return err
}

func (pgc *PGXConn) SendBatch(ctx context.Context, batch IBatch) (IBatchResults, error) {
	if pgc.conn == nil {
		return nil, ErrNoConnection
	}

	batchResults := pgc.conn.SendBatch(ctx, batch.getBatch())

	return NewPgxBatchResults(batchResults), nil
}

func (pgc *PGXConn) BeginTransaction(ctx context.Context) (ITransaction, error) {
	if pgc.conn == nil {
		return nil, ErrNoConnection
	}
	tx, err := pgc.conn.Begin(ctx)
	return &TXConn{tx, pgc.multiTenantEnabled}, err
}

func (pgc *PGXConn) Release(ctx context.Context) {
	if pgc.poolConn != nil {
		if pgc.multiTenantEnabled {
			err := pgc.BeforeReleaseHook(ctx)
			if err != nil {
				pgc.poolConn.Release()
				logger.Errorf(ctx, "Failed to unset tenant ID for session into release conn: %s", err.Error())
			}
		}
		pgc.poolConn.Release()
		pgc.conn = nil
	}
}

func (pgc *PGXConn) Ping(ctx context.Context) error {
	if pgc.conn == nil {
		return ErrNoConnection
	}
	rows, errPing := pgc.conn.Query(ctx, `SELECT 1`)
	if errPing != nil {
		return errPing
	}
	defer rows.Close()
	return nil
}

func (pgc *PGXConn) BeforeReleaseHook(ctx context.Context) (err error) {
	if tid := ctx.Value("tenant_id"); tid != nil && tid != "" {
		_, err = pgc.poolConn.Exec(ctx, "SELECT set_config($1,$2,$3);", "app.current_tenant", "", false)
		if err != nil {
			logger.Errorf(ctx, "Failed to unset tenant ID for session: %s", err.Error())
			return err
		} else {
			logger.Debugf(ctx, "Unset tenant ID for session")
		}
	}
	return nil
}

func (pgc *PGXConn) AfterAcquireHook(ctx context.Context) (err error) {
	if tid := ctx.Value("tenant_id"); tid != nil && tid != "" {
		if tenantID, ok := tid.(string); ok {
			_, err = pgc.poolConn.Exec(ctx, "SELECT set_config($1,$2,$3);", "app.current_tenant", tenantID, false)
			if err != nil {
				logger.Errorf(ctx, "Failed to set tenant ID for session: %s\n", err.Error())
				return err
			} else {
				logger.Debugf(ctx, "Set tenant ID for session: %s", tenantID)
			}
		}
	}

	return nil
}
