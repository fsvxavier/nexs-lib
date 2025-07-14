package pq

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	_ "github.com/lib/pq"
)

// Provider implements the IProvider interface for lib/pq
type Provider struct {
	name string
}

// NewProvider creates a new lib/pq provider instance
func NewProvider() *Provider {
	return &Provider{
		name: "pq",
	}
}

// Type returns the provider type
func (p *Provider) Type() postgresql.ProviderType {
	return postgresql.ProviderTypePQ
}

// Name returns the provider name
func (p *Provider) Name() string {
	return p.name
}

// Version returns the provider version
func (p *Provider) Version() string {
	return "pq-v1.10.9"
}

// CreatePool creates a new connection pool using lib/pq
func (p *Provider) CreatePool(ctx context.Context, cfg *config.Config) (postgresql.IPool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Build connection string
	connStr := p.buildConnectionString(cfg)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	if cfg.MaxConns > 0 {
		db.SetMaxOpenConns(int(cfg.MaxConns))
	}
	if cfg.MinConns > 0 {
		db.SetMaxIdleConns(int(cfg.MinConns))
	}
	if cfg.MaxConnLifetime > 0 {
		db.SetConnMaxLifetime(cfg.MaxConnLifetime)
	}
	if cfg.MaxConnIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.MaxConnIdleTime)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{
		db:     db,
		config: cfg,
		closed: false,
	}, nil
}

// CreateConnection creates a single database connection using lib/pq
func (p *Provider) CreateConnection(ctx context.Context, cfg *config.Config) (postgresql.IConn, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Build connection string
	connStr := p.buildConnectionString(cfg)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection limits for single connection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Conn{
		db:       db,
		config:   cfg,
		released: false,
	}, nil
}

// buildConnectionString builds a PostgreSQL connection string
func (p *Provider) buildConnectionString(cfg *config.Config) string {
	if cfg.ConnString != "" {
		return cfg.ConnString
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, string(cfg.TLSMode))

	// Add connect timeout
	if cfg.ConnectTimeout > 0 {
		connStr += fmt.Sprintf(" connect_timeout=%d", int(cfg.ConnectTimeout.Seconds()))
	}

	// Add application name
	if cfg.ApplicationName != "" {
		connStr += fmt.Sprintf(" application_name=%s", cfg.ApplicationName)
	}

	// Add search path
	if len(cfg.SearchPath) > 0 {
		searchPath := ""
		for i, path := range cfg.SearchPath {
			if i > 0 {
				searchPath += ","
			}
			searchPath += path
		}
		connStr += fmt.Sprintf(" search_path=%s", searchPath)
	}

	// Add timezone
	if cfg.Timezone != "" {
		connStr += fmt.Sprintf(" timezone=%s", cfg.Timezone)
	}

	// Add runtime parameters
	if len(cfg.RuntimeParams) > 0 {
		for key, value := range cfg.RuntimeParams {
			connStr += fmt.Sprintf(" %s=%s", key, value)
		}
	}

	return connStr
}

// IsHealthy performs a health check
func (p *Provider) IsHealthy(ctx context.Context) bool {
	return true
}

// GetMetrics returns provider metrics
func (p *Provider) GetMetrics(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"provider":      "pq",
		"version":       "v1.10.9",
		"driver":        "postgres",
		"batch_support": true,
		"listen_notify": true,
	}
}

// Close performs any cleanup operations
func (p *Provider) Close() error {
	// lib/pq provider doesn't need explicit cleanup
	return nil
}
