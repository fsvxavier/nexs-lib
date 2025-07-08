package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/fsvxavier/nexs-lib/db/postgresql/pgx"
)

// PGXStrategy implementa a estratégia para o provider PGX
type PGXStrategy struct{}

// NewPGXStrategy cria uma nova instância da estratégia PGX
func NewPGXStrategy() ProviderStrategy {
	return &PGXStrategy{}
}

// CreateConnection cria uma nova conexão PGX
func (s *PGXStrategy) CreateConnection(ctx context.Context, config *common.Config) (common.IConn, error) {
	return pgx.NewConn(ctx, config)
}

// CreatePool cria um novo pool de conexões PGX
func (s *PGXStrategy) CreatePool(ctx context.Context, config *common.Config) (common.IPool, error) {
	return pgx.NewPool(ctx, config)
}

// CreateBatch cria um novo batch PGX
func (s *PGXStrategy) CreateBatch() (common.IBatch, error) {
	return pgx.NewBatch(), nil
}

// ValidateConfig valida a configuração para PGX
func (s *PGXStrategy) ValidateConfig(config *common.Config) error {
	if config.Host == "" {
		return errors.New("host é obrigatório para PGX")
	}
	if config.Database == "" {
		return errors.New("database é obrigatório para PGX")
	}
	if config.User == "" {
		return errors.New("user é obrigatório para PGX")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("port deve estar entre 1 e 65535, recebido: %d", config.Port)
	}
	if config.MaxConns < 0 {
		return fmt.Errorf("maxConns deve ser >= 0, recebido: %d", config.MaxConns)
	}
	if config.MinConns < 0 {
		return fmt.Errorf("minConns deve ser >= 0, recebido: %d", config.MinConns)
	}
	if config.MaxConns > 0 && config.MinConns > config.MaxConns {
		return fmt.Errorf("minConns (%d) não pode ser maior que maxConns (%d)", config.MinConns, config.MaxConns)
	}
	return nil
}
