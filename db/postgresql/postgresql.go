package postgresql

import (
	"context"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
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
// Utiliza o padrão Factory com Strategy para flexibilidade e extensibilidade
func NewConnection(ctx context.Context, providerType ProviderType, config *common.Config) (common.IConn, error) {
	return GetFactory().CreateConnection(ctx, providerType, config)
}

// NewPool cria um novo pool de conexões PostgreSQL usando o provider especificado
// Utiliza o padrão Factory com Strategy para flexibilidade e extensibilidade
func NewPool(ctx context.Context, providerType ProviderType, config *common.Config) (common.IPool, error) {
	return GetFactory().CreatePool(ctx, providerType, config)
}

// NewBatch cria um novo lote de consultas usando o provider especificado
// Utiliza o padrão Factory com Strategy para flexibilidade e extensibilidade
func NewBatch(providerType ProviderType) (common.IBatch, error) {
	return GetFactory().CreateBatch(providerType)
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
