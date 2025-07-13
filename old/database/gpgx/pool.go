package gpgx

import (
	"context"
	"crypto/tls"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGXPool struct {
	pool               *pgxpool.Pool
	multiTenantEnabled bool
}

func (pgp *PGXPool) Pool() *pgxpool.Pool {
	return pgp.pool
}

// NewPool creates a new Pool and immediately establishes one connection.
// maxConns is the maximum size of the pool. The default is the max(4, runtime.NumCPU()).
func NewPool(ctx context.Context, connString string, options ...PgxConfig) (IPool, error) {
	cfg := GetConfig()

	for _, opt := range options {
		opt(cfg)
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	if enabled, _ := strconv.ParseBool(os.Getenv("DB_TLS_ENABLED")); enabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		config.ConnConfig.TLSConfig = tlsConfig
	}

	config.ConnConfig.RuntimeParams["timezone"] = "UTC"
	switch os.Getenv("DB_QUERY_MODE_EXEC") {
	case "CACHE_STATEMENT":
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	case "CACHE_DESCRIBE":
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	case "DESCRIBE_EXEC":
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec
	case "SIMPLE_PROTOCOL":
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	default:
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	}

	// Creates a new pool with the given configuration.
	// MaxConns is the maximum size of the pool. The default is the greater of 4 or runtime.NumCPU().
	config.MaxConns = cfg.maxConns
	config.MinConns = cfg.minConns
	config.MaxConnLifetime = cfg.maxConnLifetime
	config.MaxConnIdleTime = cfg.maxConnIdletime

	config.ConnConfig.Tracer = &TracerConfig{
		QueryTracerEnabled:   isQueryTracerEnabled(cfg),
		ObservabilityEnabled: isObservabilityEnabled(cfg),
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	pgxPool := &PGXPool{
		pool:               pool,
		multiTenantEnabled: cfg.multiTenantEnabled,
	}

	return pgxPool, nil
}

// Acquire acquires a connection from the pool.
// If the pool is nil, it returns an error.
// If the pool is closed, it returns an error.
// If the pool is not initialized, it returns an error.
// If multi-tenant mode is enabled, it calls the AfterAcquireHook on the connection.
// If the connection is successfully acquired, it returns the connection and nil error.
// If there is an error acquiring the connection, it returns nil and the error.
// It is safe to call Acquire multiple times, but it is recommended to release the connection after use.
// The returned connection should be released using the Release method when it is no longer needed.
// If the connection is nil, it returns an error.
// If the connection is not nil, it returns the existing connection and a no-op release function.
// The returned connection implements the IConn interface.
// It is recommended to use the GetConnWithNotPresent method to acquire a connection if it is not already present.
func (pgp *PGXPool) Acquire(ctx context.Context) (IConn, error) {
	poolConn, err := pgp.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	pgxConn := &PGXConn{
		poolConn:           poolConn,
		conn:               poolConn.Conn(),
		multiTenantEnabled: pgp.multiTenantEnabled,
	}

	if pgp.multiTenantEnabled {
		err = pgxConn.AfterAcquireHook(ctx)
		if err != nil {
			return nil, err
		}
	}

	return pgxConn, nil
}

// Close closes the connection pool and releases all resources.
// It is safe to call Close multiple times.
// After closing, the pool cannot be used anymore.
// It is recommended to call Close when the application is shutting down.
// If the pool is nil, it does nothing.
// If the pool is already closed, it does nothing.
func (pgp *PGXPool) Close() {
	if pgp.Pool() != nil {
		pgp.Pool().Close()
	}
}

// Ping checks if the connection pool is alive by executing a simple query.
// It returns nil if the pool is alive and reachable.
func (pgp *PGXPool) Ping(ctx context.Context) error {
	conn, errConn := pgp.pool.Acquire(ctx)
	defer conn.Release()
	if errConn != nil {
		return errConn
	}
	rows, errPing := conn.Query(ctx, `SELECT 1`)
	if errPing != nil {
		return errPing
	}
	defer rows.Close()
	return nil
}

// Stat returns the statistics of the connection pool.
// It returns nil if the pool is not initialized.
func (pgp *PGXPool) Stat() *pgxpool.Stat {
	if pgp.pool != nil {
		return pgp.pool.Stat()
	}
	return nil
}

// GetConnWithNotPresent checks if the provided connection is nil.
// If it is nil, it acquires a new connection from the pool.
// If the connection is not nil, it returns the existing connection and a no-op release function.
func (pgp *PGXPool) GetConnWithNotPresent(ctx context.Context, conn IConn) (IConn, func(), error) {

	// If the connection is nil, acquire a new connection from the pool
	if conn == nil {
		var err error
		conn, err = pgp.Acquire(ctx)
		if err != nil {
			return nil, func() {}, err
		}
		// Return the connection and a release function
		return conn, func() { conn.Release(ctx) }, nil
	}

	// Return the connection and a release function
	return conn, func() {}, nil
}

// isObservabilityEnabled checks if observability is enabled based on the pgxConfig and environment variables.
func isObservabilityEnabled(cfg *pgxConfig) bool {
	return cfg.datadogEnabled && os.Getenv("DD_AGENT_HOST") != ""
}

// isQueryTracerEnabled checks if the query tracer is enabled based on the pgxConfig.
func isQueryTracerEnabled(cfg *pgxConfig) bool {
	return cfg.queryTracerEnabled
}
