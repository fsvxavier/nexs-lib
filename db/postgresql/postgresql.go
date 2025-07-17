package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/hooks"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
)

// PostgreSQLProviderImpl implements the PostgreSQLProvider interface
type PostgreSQLProviderImpl struct {
	name              string
	version           string
	supportedFeatures []string
	driverName        string
	hookManager       interfaces.HookManager
	retryManager      interfaces.RetryManager
	failoverManager   interfaces.FailoverManager
}

// NewPostgreSQLProvider creates a new PostgreSQL provider
func NewPostgreSQLProvider(providerType interfaces.ProviderType) (interfaces.PostgreSQLProvider, error) {
	var driverName string
	var supportedFeatures []string

	switch providerType {
	case interfaces.ProviderTypePGX:
		driverName = "pgx"
		supportedFeatures = []string{
			"connection_pooling",
			"transactions",
			"prepared_statements",
			"batch_operations",
			"listen_notify",
			"copy_operations",
			"multi_tenancy",
			"read_replicas",
			"failover",
			"ssl_tls",
			"context_support",
			"hooks",
			"health_check",
			"statistics",
		}
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	// Create default retry and failover configurations
	defaultRetryConfig := interfaces.RetryConfig{
		MaxRetries:      3,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second * 5,
		Multiplier:      2.0,
		RandomizeWait:   true,
	}

	defaultFailoverConfig := interfaces.FailoverConfig{
		Enabled:             false,
		FallbackNodes:       []string{},
		HealthCheckInterval: time.Second * 30,
		RetryInterval:       time.Second * 5,
		MaxFailoverAttempts: 3,
	}

	provider := &PostgreSQLProviderImpl{
		name:              fmt.Sprintf("postgresql-%s", providerType),
		version:           "1.0.0",
		supportedFeatures: supportedFeatures,
		driverName:        driverName,
		hookManager:       hooks.NewDefaultHookManager(),
		retryManager:      NewDefaultRetryManager(defaultRetryConfig),
		failoverManager:   NewDefaultFailoverManager(defaultFailoverConfig),
	}

	return provider, nil
}

// Name returns the provider name
func (p *PostgreSQLProviderImpl) Name() string {
	return p.name
}

// Version returns the provider version
func (p *PostgreSQLProviderImpl) Version() string {
	return p.version
}

// SupportsFeature checks if a feature is supported
func (p *PostgreSQLProviderImpl) SupportsFeature(feature string) bool {
	for _, supportedFeature := range p.supportedFeatures {
		if supportedFeature == feature {
			return true
		}
	}
	return false
}

// GetDriverName returns the underlying driver name
func (p *PostgreSQLProviderImpl) GetDriverName() string {
	return p.driverName
}

// GetSupportedFeatures returns all supported features
func (p *PostgreSQLProviderImpl) GetSupportedFeatures() []string {
	features := make([]string, len(p.supportedFeatures))
	copy(features, p.supportedFeatures)
	return features
}

// NewPool creates a new connection pool
func (p *PostgreSQLProviderImpl) NewPool(ctx context.Context, config interfaces.Config) (interfaces.IPool, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	switch p.driverName {
	case "pgx":
		return pgx.NewPool(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", p.driverName)
	}
}

// NewConn creates a new single connection
func (p *PostgreSQLProviderImpl) NewConn(ctx context.Context, config interfaces.Config) (interfaces.IConn, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	switch p.driverName {
	case "pgx":
		return pgx.NewConn(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", p.driverName)
	}
}

// ValidateConfig validates the provided configuration
func (p *PostgreSQLProviderImpl) ValidateConfig(config interfaces.Config) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	return config.Validate()
}

// NewListenConn creates a new connection specifically for LISTEN/NOTIFY operations
func (p *PostgreSQLProviderImpl) NewListenConn(ctx context.Context, config interfaces.Config) (interfaces.IConn, error) {
	if !p.SupportsFeature("listen_notify") {
		return nil, fmt.Errorf("LISTEN/NOTIFY not supported by provider %s", p.name)
	}

	switch p.driverName {
	case "pgx":
		return pgx.NewListenConn(ctx, config)
	default:
		return nil, fmt.Errorf("LISTEN/NOTIFY not implemented for driver: %s", p.driverName)
	}
}

// CreateSchema creates a new schema
func (p *PostgreSQLProviderImpl) CreateSchema(ctx context.Context, conn interfaces.IConn, schemaName string) error {
	if schemaName == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	_, err := conn.Exec(ctx, query)
	return err
}

// DropSchema drops a schema
func (p *PostgreSQLProviderImpl) DropSchema(ctx context.Context, conn interfaces.IConn, schemaName string) error {
	if schemaName == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	_, err := conn.Exec(ctx, query)
	return err
}

// ListSchemas lists all schemas
func (p *PostgreSQLProviderImpl) ListSchemas(ctx context.Context, conn interfaces.IConn) ([]string, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
		ORDER BY schema_name
	`

	var schemas []string
	err := conn.QueryAll(ctx, &schemas, query)
	return schemas, err
}

// CreateDatabase creates a new database
func (p *PostgreSQLProviderImpl) CreateDatabase(ctx context.Context, conn interfaces.IConn, dbName string) error {
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := conn.Exec(ctx, query)
	return err
}

// DropDatabase drops a database
func (p *PostgreSQLProviderImpl) DropDatabase(ctx context.Context, conn interfaces.IConn, dbName string) error {
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := conn.Exec(ctx, query)
	return err
}

// ListDatabases lists all databases
func (p *PostgreSQLProviderImpl) ListDatabases(ctx context.Context, conn interfaces.IConn) ([]string, error) {
	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false 
		ORDER BY datname
	`

	var databases []string
	err := conn.QueryAll(ctx, &databases, query)
	return databases, err
}

// WithRetry executes an operation with retry logic
func (p *PostgreSQLProviderImpl) WithRetry(ctx context.Context, operation func() error) error {
	if p.retryManager == nil {
		return operation() // No retry manager, execute directly
	}
	return p.retryManager.Execute(ctx, operation)
}

// WithFailover executes an operation with failover logic
func (p *PostgreSQLProviderImpl) WithFailover(ctx context.Context, operation func(conn interfaces.IConn) error) error {
	if p.failoverManager == nil {
		// No failover manager, try to get a connection and execute
		pool, err := p.NewPool(ctx, nil) // This would need a config parameter in real implementation
		if err != nil {
			return err
		}
		defer pool.Close()

		conn, err := pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		return operation(conn)
	}
	return p.failoverManager.Execute(ctx, operation)
}

// GetRetryManager returns the retry manager
func (p *PostgreSQLProviderImpl) GetRetryManager() interfaces.RetryManager {
	return p.retryManager
}

// GetFailoverManager returns the failover manager
func (p *PostgreSQLProviderImpl) GetFailoverManager() interfaces.FailoverManager {
	return p.failoverManager
}

// ProviderFactoryImpl implements the ProviderFactory interface
type ProviderFactoryImpl struct {
	providers map[interfaces.ProviderType]interfaces.PostgreSQLProvider
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactoryImpl {
	return &ProviderFactoryImpl{
		providers: make(map[interfaces.ProviderType]interfaces.PostgreSQLProvider),
	}
}

// CreateProvider creates a provider of the specified type
func (pf *ProviderFactoryImpl) CreateProvider(providerType interfaces.ProviderType) (interfaces.PostgreSQLProvider, error) {
	// Check if provider already exists
	if provider, exists := pf.providers[providerType]; exists {
		return provider, nil
	}

	// Create new provider
	provider, err := NewPostgreSQLProvider(providerType)
	if err != nil {
		return nil, err
	}

	// Store provider
	pf.providers[providerType] = provider
	return provider, nil
}

// RegisterProvider registers a custom provider
func (pf *ProviderFactoryImpl) RegisterProvider(providerType interfaces.ProviderType, provider interfaces.PostgreSQLProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	pf.providers[providerType] = provider
	return nil
}

// ListProviders returns all available provider types
func (pf *ProviderFactoryImpl) ListProviders() []interfaces.ProviderType {
	types := make([]interfaces.ProviderType, 0, len(pf.providers))
	for providerType := range pf.providers {
		types = append(types, providerType)
	}
	return types
}

// GetProvider returns a provider by type
func (pf *ProviderFactoryImpl) GetProvider(providerType interfaces.ProviderType) (interfaces.PostgreSQLProvider, bool) {
	provider, exists := pf.providers[providerType]
	return provider, exists
}

// Global factory instance
var defaultFactory = NewProviderFactory()

// GetDefaultFactory returns the default provider factory
func GetDefaultFactory() interfaces.ProviderFactory {
	return defaultFactory
}

// Quick factory methods for common operations

// NewPGXProvider creates a PGX provider
func NewPGXProvider() (interfaces.PostgreSQLProvider, error) {
	return defaultFactory.CreateProvider(interfaces.ProviderTypePGX)
}

// NewDefaultConfig creates a default configuration
func NewDefaultConfig(connectionString string) interfaces.Config {
	return config.NewDefaultConfig(connectionString)
}

// ConfigOption represents a configuration option
type ConfigOption = config.ConfigOption

// Configuration option functions
var (
	WithConnectionString = config.WithConnectionString
	WithMaxConns         = config.WithMaxConns
	WithMinConns         = config.WithMinConns
	WithMaxConnLifetime  = config.WithMaxConnLifetime
	WithMaxConnIdleTime  = config.WithMaxConnIdleTime
	WithMultiTenant      = config.WithMultiTenant
	WithReadReplicas     = config.WithReadReplicas
	WithFailover         = config.WithFailover
	WithTLS              = config.WithTLS
	WithMaxRetries       = config.WithMaxRetries
	WithEnabledHooks     = config.WithEnabledHooks
	WithCustomHook       = config.WithCustomHook
)
