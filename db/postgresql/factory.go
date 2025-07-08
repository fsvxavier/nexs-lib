package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
)

var (
	// ErrInvalidProviderType indica que o tipo de provider fornecido é inválido
	ErrInvalidProviderType = errors.New("tipo de provider inválido")
	// ErrNilConfig indica que a configuração fornecida é nil
	ErrNilConfig = errors.New("configuração não pode ser nil")
	// ErrInvalidContext indica que o contexto fornecido é inválido
	ErrInvalidContext = errors.New("contexto não pode ser nil")
)

// ProviderStrategy define a estratégia para criação de diferentes tipos de providers
type ProviderStrategy interface {
	CreateConnection(ctx context.Context, config *common.Config) (common.IConn, error)
	CreatePool(ctx context.Context, config *common.Config) (common.IPool, error)
	CreateBatch() (common.IBatch, error)
	ValidateConfig(config *common.Config) error
}

// DatabaseFactory implementa o padrão Factory para criação de conexões PostgreSQL
type DatabaseFactory struct {
	strategies map[ProviderType]ProviderStrategy
}

// NewDatabaseFactory cria uma nova instância da factory
func NewDatabaseFactory() *DatabaseFactory {
	factory := &DatabaseFactory{
		strategies: make(map[ProviderType]ProviderStrategy),
	}

	// Registra as estratégias padrão
	factory.RegisterStrategy(PGX, NewPGXStrategy())
	factory.RegisterStrategy(PQ, NewPQStrategy())
	factory.RegisterStrategy(GORM, NewGORMStrategy())

	return factory
}

// RegisterStrategy registra uma nova estratégia de provider
func (f *DatabaseFactory) RegisterStrategy(providerType ProviderType, strategy ProviderStrategy) {
	if f.strategies == nil {
		f.strategies = make(map[ProviderType]ProviderStrategy)
	}
	f.strategies[providerType] = strategy
}

// CreateConnection cria uma nova conexão usando a estratégia apropriada
func (f *DatabaseFactory) CreateConnection(ctx context.Context, providerType ProviderType, config *common.Config) (common.IConn, error) {
	if err := f.validateInputs(ctx, config); err != nil {
		return nil, err
	}

	strategy, exists := f.strategies[providerType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrInvalidProviderType, providerType)
	}

	if err := strategy.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("validação de configuração falhou: %w", err)
	}

	return strategy.CreateConnection(ctx, config)
}

// CreatePool cria um novo pool de conexões usando a estratégia apropriada
func (f *DatabaseFactory) CreatePool(ctx context.Context, providerType ProviderType, config *common.Config) (common.IPool, error) {
	if err := f.validateInputs(ctx, config); err != nil {
		return nil, err
	}

	strategy, exists := f.strategies[providerType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrInvalidProviderType, providerType)
	}

	if err := strategy.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("validação de configuração falhou: %w", err)
	}

	return strategy.CreatePool(ctx, config)
}

// CreateBatch cria um novo batch usando a estratégia apropriada
func (f *DatabaseFactory) CreateBatch(providerType ProviderType) (common.IBatch, error) {
	strategy, exists := f.strategies[providerType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrInvalidProviderType, providerType)
	}

	return strategy.CreateBatch()
}

// GetSupportedProviders retorna lista de providers suportados
func (f *DatabaseFactory) GetSupportedProviders() []ProviderType {
	providers := make([]ProviderType, 0, len(f.strategies))
	for providerType := range f.strategies {
		providers = append(providers, providerType)
	}
	return providers
}

// validateInputs valida entradas comuns
func (f *DatabaseFactory) validateInputs(ctx context.Context, config *common.Config) error {
	if ctx == nil {
		return ErrInvalidContext
	}
	if config == nil {
		return ErrNilConfig
	}
	return nil
}

// Instância global da factory
var defaultFactory = NewDatabaseFactory()

// GetFactory retorna a instância global da factory
func GetFactory() *DatabaseFactory {
	return defaultFactory
}

// SetFactory define uma nova instância global da factory (útil para testes)
func SetFactory(factory *DatabaseFactory) {
	defaultFactory = factory
}
