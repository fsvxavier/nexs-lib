package postgresql

import (
	"context"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/fsvxavier/nexs-lib/db/postgresql/gorm"
	"github.com/fsvxavier/nexs-lib/db/postgresql/pgx"
	"github.com/fsvxavier/nexs-lib/db/postgresql/pq"
)

// ProviderType define o tipo de provider PostgreSQL a ser utilizado
type ProviderType string

const (
	// PGX utiliza o driver pgx
	PGX ProviderType = "pgx"
	// PQ utiliza o driver pq
	PQ ProviderType = "pq"
	// GORM utiliza o ORM GORM
	GORM ProviderType = "gorm"
)

// NewConnection cria uma nova conexão PostgreSQL usando o provider especificado
func NewConnection(ctx context.Context, providerType ProviderType, config *common.Config) (common.IConn, error) {
	switch providerType {
	case PGX:
		return pgx.NewConn(ctx, config)
	case PQ:
		return pq.NewConn(ctx, config)
	case GORM:
		return gorm.NewConn(ctx, config)
	default:
		return nil, fmt.Errorf("provider não suportado: %s", providerType)
	}
}

// NewPool cria um novo pool de conexões PostgreSQL usando o provider especificado
func NewPool(ctx context.Context, providerType ProviderType, config *common.Config) (common.IPool, error) {
	switch providerType {
	case PGX:
		return pgx.NewPool(ctx, config)
	case PQ:
		return pq.NewPool(ctx, config)
	case GORM:
		return gorm.NewPool(ctx, config)
	default:
		return nil, fmt.Errorf("provider não suportado: %s", providerType)
	}
}

// NewBatch cria um novo lote de consultas usando o provider especificado
func NewBatch(providerType ProviderType) (common.IBatch, error) {
	switch providerType {
	case PGX:
		return pgx.NewBatch(), nil
	case PQ:
		return pq.NewBatch(), nil
	case GORM:
		return gorm.NewBatch(), nil
	default:
		return nil, fmt.Errorf("provider não suportado: %s", providerType)
	}
}

// IsEmptyResultError verifica se um erro indica ausência de resultados
func IsEmptyResultError(err error) bool {
	return common.IsEmptyResultError(err)
}

// IsDuplicateKeyError verifica se um erro indica violação de chave única
func IsDuplicateKeyError(err error) bool {
	return common.IsDuplicateKeyError(err)
}

// WithConfig configura o provider PostgreSQL com opções fornecidas
func WithConfig(opts ...common.Option) *common.Config {
	config := common.DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// Opções de configuração
var (
	WithHost               = common.WithHost
	WithPort               = common.WithPort
	WithDatabase           = common.WithDatabase
	WithUser               = common.WithUser
	WithPassword           = common.WithPassword
	WithSSLMode            = common.WithSSLMode
	WithMaxConns           = common.WithMaxConns
	WithMinConns           = common.WithMinConns
	WithMaxConnLifetime    = common.WithMaxConnLifetime
	WithMaxConnIdleTime    = common.WithMaxConnIdleTime
	WithTraceEnabled       = common.WithTraceEnabled
	WithQueryLogEnabled    = common.WithQueryLogEnabled
	WithMultiTenantEnabled = common.WithMultiTenantEnabled
)

// Erros comuns
var (
	ErrNoRows                   = common.ErrNoRows
	ErrNoTransaction            = common.ErrNoTransaction
	ErrNoConnection             = common.ErrNoConnection
	ErrInvalidNestedTransaction = common.ErrInvalidNestedTransaction
	ErrInvalidOperation         = common.ErrInvalidOperation
)
