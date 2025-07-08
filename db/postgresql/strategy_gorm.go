package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/fsvxavier/nexs-lib/db/postgresql/gorm"
)

// GORMStrategy implementa a estratégia para o provider GORM
type GORMStrategy struct{}

// NewGORMStrategy cria uma nova instância da estratégia GORM
func NewGORMStrategy() ProviderStrategy {
	return &GORMStrategy{}
}

// CreateConnection cria uma nova conexão GORM
func (s *GORMStrategy) CreateConnection(ctx context.Context, config *common.Config) (common.IConn, error) {
	return gorm.NewConn(ctx, config)
}

// CreatePool cria um novo pool de conexões GORM
func (s *GORMStrategy) CreatePool(ctx context.Context, config *common.Config) (common.IPool, error) {
	return gorm.NewPool(ctx, config)
}

// CreateBatch cria um novo batch GORM
func (s *GORMStrategy) CreateBatch() (common.IBatch, error) {
	return gorm.NewBatch(), nil
}

// ValidateConfig valida a configuração para GORM
func (s *GORMStrategy) ValidateConfig(config *common.Config) error {
	if config.Host == "" {
		return errors.New("host é obrigatório para GORM")
	}
	if config.Database == "" {
		return errors.New("database é obrigatório para GORM")
	}
	if config.User == "" {
		return errors.New("user é obrigatório para GORM")
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
