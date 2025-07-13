package postgresql

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pq"
)

// CreateProvider creates a new database provider based on the configuration
func CreateProvider(cfg *config.Config) (interfaces.DatabaseProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
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
