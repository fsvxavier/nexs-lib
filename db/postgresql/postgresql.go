package postgresql

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pq"
)

// NewProvider creates a new database provider based on the configuration
func NewProvider(cfg *config.Config) (interfaces.DatabaseProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	switch cfg.Driver {
	case interfaces.DriverPGX:
		return pgx.NewProvider(cfg)
	case interfaces.DriverGORM:
		return gorm.NewProvider(cfg)
	case interfaces.DriverPQ:
		return pq.NewProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}

// NewProviderWithConfig creates a new database provider with configuration options
func NewProviderWithConfig(options ...config.ConfigOption) (interfaces.DatabaseProvider, error) {
	cfg := config.NewConfig(options...)
	return NewProvider(cfg)
}
