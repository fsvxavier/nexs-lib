package pgx

import (
	"context"
	"fmt"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGXProvider implements the PostgreSQLProvider interface
type PGXProvider struct {
	name              string
	version           string
	supportedFeatures []string
}

// NewPGXProvider creates a new PGX provider instance
func NewPGXProvider() interfaces.PostgreSQLProvider {
	return &PGXProvider{
		name:    "PGX",
		version: "5.x",
		supportedFeatures: []string{
			"transactions",
			"prepared_statements",
			"batch_operations",
			"copy_operations",
			"listen_notify",
			"connection_pooling",
			"hooks",
			"multi_tenancy",
			"health_checks",
			"statistics",
			"ssl_connections",
			"async_operations",
		},
	}
}

// Name implements Provider.Name
func (p *PGXProvider) Name() string {
	return p.name
}

// Version implements Provider.Version
func (p *PGXProvider) Version() string {
	return p.version
}

// SupportsFeature implements Provider.SupportsFeature
func (p *PGXProvider) SupportsFeature(feature string) bool {
	for _, f := range p.supportedFeatures {
		if f == feature {
			return true
		}
	}
	return false
}

// NewPool implements Provider.NewPool
func (p *PGXProvider) NewPool(ctx context.Context, config interfaces.Config) (interfaces.IPool, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	connStr := config.GetConnectionString()
	if connStr == "" {
		return nil, fmt.Errorf("connection string is required")
	}

	// Parse connection string to get pgxpool config
	pgxConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Apply pool configuration
	poolConfig := config.GetPoolConfig()
	pgxConfig.MaxConns = poolConfig.MaxConns
	pgxConfig.MinConns = poolConfig.MinConns
	pgxConfig.MaxConnLifetime = poolConfig.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = poolConfig.MaxConnIdleTime
	pgxConfig.HealthCheckPeriod = poolConfig.HealthCheckPeriod

	// Apply TLS configuration
	tlsConfig := config.GetTLSConfig()
	if tlsConfig.Enabled {
		// TLS configuration would be applied here
		// This is typically done via the connection string or pgx.ConnConfig
	}

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Create PGX pool wrapper
	pgxPool := &PGXPool{
		pool:   pool,
		config: config,
		closed: false,
	}

	return pgxPool, nil
}

// NewConn implements Provider.NewConn
func (p *PGXProvider) NewConn(ctx context.Context, config interfaces.Config) (interfaces.IConn, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	connStr := config.GetConnectionString()
	if connStr == "" {
		return nil, fmt.Errorf("connection string is required")
	}

	// Parse connection string
	pgxConfig, err := pgx.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Apply TLS configuration
	tlsConfig := config.GetTLSConfig()
	if tlsConfig.Enabled {
		// TLS configuration would be applied here
	}

	// Create direct connection
	conn, err := pgx.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Create PGX connection wrapper
	pgxConn := &PGXConn{
		rawConn:  conn,
		config:   config,
		stats:    NewConnectionStats(),
		acquired: false,
		closed:   false,
	}

	return pgxConn, nil
}

// NewListenConn implements PostgreSQLProvider.NewListenConn
func (p *PGXProvider) NewListenConn(ctx context.Context, config interfaces.Config) (interfaces.IConn, error) {
	// For listen connections, we need a direct connection, not a pooled one
	return p.NewConn(ctx, config)
}

// ValidateConfig implements Provider.ValidateConfig
func (p *PGXProvider) ValidateConfig(config interfaces.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// First validate using the config's own validation
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	connStr := config.GetConnectionString()
	if connStr == "" {
		return fmt.Errorf("connection string is required")
	}

	// Validate that the connection string can be parsed
	_, err := pgx.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	// Validate pool configuration
	poolConfig := config.GetPoolConfig()
	if poolConfig.MaxConns <= 0 {
		return fmt.Errorf("max connections must be greater than 0")
	}
	if poolConfig.MinConns < 0 {
		return fmt.Errorf("min connections cannot be negative")
	}
	if poolConfig.MinConns > poolConfig.MaxConns {
		return fmt.Errorf("min connections cannot be greater than max connections")
	}

	return nil
}

// GetDriverName implements Provider.GetDriverName
func (p *PGXProvider) GetDriverName() string {
	return "pgx"
}

// GetSupportedFeatures implements Provider.GetSupportedFeatures
func (p *PGXProvider) GetSupportedFeatures() []string {
	features := make([]string, len(p.supportedFeatures))
	copy(features, p.supportedFeatures)
	return features
}

// CreateSchema implements PostgreSQLProvider.CreateSchema
func (p *PGXProvider) CreateSchema(ctx context.Context, conn interfaces.IConn, schemaName string) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if schemaName == "" {
		return fmt.Errorf("schema name cannot be empty")
	}
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	_, err := conn.Exec(ctx, query)
	return err
}

// DropSchema implements PostgreSQLProvider.DropSchema
func (p *PGXProvider) DropSchema(ctx context.Context, conn interfaces.IConn, schemaName string) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if schemaName == "" {
		return fmt.Errorf("schema name cannot be empty")
	}
	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	_, err := conn.Exec(ctx, query)
	return err
}

// ListSchemas implements PostgreSQLProvider.ListSchemas
func (p *PGXProvider) ListSchemas(ctx context.Context, conn interfaces.IConn) ([]string, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}

	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
		ORDER BY schema_name
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, err
		}
		schemas = append(schemas, schemaName)
	}

	return schemas, rows.Err()
}

// CreateDatabase implements PostgreSQLProvider.CreateDatabase
func (p *PGXProvider) CreateDatabase(ctx context.Context, conn interfaces.IConn, dbName string) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := conn.Exec(ctx, query)
	return err
}

// DropDatabase implements PostgreSQLProvider.DropDatabase
func (p *PGXProvider) DropDatabase(ctx context.Context, conn interfaces.IConn, dbName string) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := conn.Exec(ctx, query)
	return err
}

// ListDatabases implements PostgreSQLProvider.ListDatabases
func (p *PGXProvider) ListDatabases(ctx context.Context, conn interfaces.IConn) ([]string, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}

	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false
		ORDER BY datname
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}

	return databases, rows.Err()
}

// WithRetry executes an operation with retry logic
func (p *PGXProvider) WithRetry(ctx context.Context, operation func() error) error {
	// For now, execute directly without retry (retry logic can be added later)
	return operation()
}

// WithFailover executes an operation with failover logic
func (p *PGXProvider) WithFailover(ctx context.Context, operation func(conn interfaces.IConn) error) error {
	// For now, get a single connection and execute (failover logic can be added later)
	// This is a simplified implementation
	return fmt.Errorf("failover not implemented in PGXProvider")
}

// GetRetryManager returns the retry manager
func (p *PGXProvider) GetRetryManager() interfaces.RetryManager {
	return nil // Not implemented yet
}

// GetFailoverManager returns the failover manager
func (p *PGXProvider) GetFailoverManager() interfaces.FailoverManager {
	return nil // Not implemented yet
}
