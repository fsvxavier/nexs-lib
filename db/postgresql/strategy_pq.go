package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/fsvxavier/nexs-lib/db/postgresql/pq"
)

// PQStrategy implementa a estratégia para o provider PQ
type PQStrategy struct{}

// NewPQStrategy cria uma nova instância da estratégia PQ
func NewPQStrategy() ProviderStrategy {
	return &PQStrategy{}
}

// CreateConnection cria uma nova conexão PQ
func (s *PQStrategy) CreateConnection(ctx context.Context, config *common.Config) (common.IConn, error) {
	return pq.NewConn(ctx, config)
}

// CreatePool cria um novo pool de conexões PQ
func (s *PQStrategy) CreatePool(ctx context.Context, config *common.Config) (common.IPool, error) {
	return pq.NewPool(ctx, config)
}

// CreateBatch cria um novo batch PQ
func (s *PQStrategy) CreateBatch() (common.IBatch, error) {
	return pq.NewBatch(), nil
}

// ValidateConfig valida a configuração para PQ
func (s *PQStrategy) ValidateConfig(config *common.Config) error {
	if config.Host == "" {
		return errors.New("host é obrigatório para PQ")
	}
	if config.Database == "" {
		return errors.New("database é obrigatório para PQ")
	}
	if config.User == "" {
		return errors.New("user é obrigatório para PQ")
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
