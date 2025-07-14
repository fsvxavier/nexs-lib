package gorm

import (
	"context"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Provider implements the IProvider interface for GORM
type Provider struct {
	name string
}

// NewProvider creates a new GORM provider instance
func NewProvider() *Provider {
	return &Provider{
		name: "gorm",
	}
}

// Type returns the provider type
func (p *Provider) Type() postgresql.ProviderType {
	return postgresql.ProviderTypeGORM
}

// Name returns the provider name
func (p *Provider) Name() string {
	return p.name
}

// Version returns the provider version
func (p *Provider) Version() string {
	return "gorm-v1.30"
}

// CreatePool creates a new GORM database instance (GORM doesn't use traditional pools)
func (p *Provider) CreatePool(ctx context.Context, cfg *config.Config) (postgresql.IPool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Build DSN
	dsn := cfg.ConnString
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, string(cfg.TLSMode))
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// Apply connection timeouts
	if cfg.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", int(cfg.ConnectTimeout.Seconds()))
	}

	// Apply additional parameters
	if len(cfg.RuntimeParams) > 0 {
		for key, value := range cfg.RuntimeParams {
			dsn += fmt.Sprintf(" %s=%s", key, value)
		}
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	if cfg.MaxConns > 0 {
		sqlDB.SetMaxOpenConns(int(cfg.MaxConns))
	}
	if cfg.MinConns > 0 {
		sqlDB.SetMaxIdleConns(int(cfg.MinConns))
	}
	if cfg.MaxConnLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.MaxConnLifetime)
	}
	if cfg.MaxConnIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.MaxConnIdleTime)
	}

	return &Pool{
		db:     db,
		sqlDB:  sqlDB,
		config: cfg,
	}, nil
}

// CreateConnection creates a single GORM database connection
func (p *Provider) CreateConnection(ctx context.Context, cfg *config.Config) (postgresql.IConn, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Build DSN
	dsn := cfg.ConnString
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, string(cfg.TLSMode))
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// Apply connection timeout
	if cfg.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", int(cfg.ConnectTimeout.Seconds()))
	}

	// Apply additional parameters
	if len(cfg.RuntimeParams) > 0 {
		for key, value := range cfg.RuntimeParams {
			dsn += fmt.Sprintf(" %s=%s", key, value)
		}
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Conn{
		db:       db,
		config:   cfg,
		released: false,
	}, nil
}

// IsHealthy performs a health check
func (p *Provider) IsHealthy(ctx context.Context) bool {
	// This would need a connection to check, so it's implemented per connection/pool
	return true
}

// GetMetrics returns provider metrics
func (p *Provider) GetMetrics(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"provider":      "gorm",
		"version":       "v1.30",
		"driver":        "postgres",
		"batch_support": false,
		"listen_notify": false,
	}
}

// Close performs any cleanup operations
func (p *Provider) Close() error {
	// GORM provider doesn't need explicit cleanup
	return nil
}
