package postgres

import (
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
	pgxprovider "github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx"
)

// ProviderFactory implementa IProviderFactory
type ProviderFactory struct {
	providers map[interfaces.ProviderType]interfaces.IPostgreSQLProvider
	mu        sync.RWMutex
}

// NewProviderFactory cria uma nova factory de providers
func NewProviderFactory() interfaces.IProviderFactory {
	return &ProviderFactory{
		providers: make(map[interfaces.ProviderType]interfaces.IPostgreSQLProvider),
	}
}

// CreateProvider cria um provider do tipo especificado
func (pf *ProviderFactory) CreateProvider(providerType interfaces.ProviderType) (interfaces.IPostgreSQLProvider, error) {
	pf.mu.RLock()
	if provider, exists := pf.providers[providerType]; exists {
		pf.mu.RUnlock()
		return provider, nil
	}
	pf.mu.RUnlock()

	// Criar novo provider
	var provider interfaces.IPostgreSQLProvider
	var err error

	switch providerType {
	case interfaces.ProviderTypePGX:
		provider = pgxprovider.NewProvider()
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Armazenar provider para reutilização
	pf.mu.Lock()
	pf.providers[providerType] = provider
	pf.mu.Unlock()

	return provider, nil
}

// RegisterProvider registra um provider customizado
func (pf *ProviderFactory) RegisterProvider(providerType interfaces.ProviderType, provider interfaces.IPostgreSQLProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	pf.mu.Lock()
	defer pf.mu.Unlock()

	pf.providers[providerType] = provider
	return nil
}

// ListProviders retorna todos os tipos de providers disponíveis
func (pf *ProviderFactory) ListProviders() []interfaces.ProviderType {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	types := make([]interfaces.ProviderType, 0, len(pf.providers))
	for providerType := range pf.providers {
		types = append(types, providerType)
	}

	// Adicionar tipos conhecidos que não foram criados ainda
	knownTypes := []interfaces.ProviderType{
		interfaces.ProviderTypePGX,
	}

	for _, knownType := range knownTypes {
		found := false
		for _, existingType := range types {
			if existingType == knownType {
				found = true
				break
			}
		}
		if !found {
			types = append(types, knownType)
		}
	}

	return types
}

// GetProvider retorna um provider por tipo
func (pf *ProviderFactory) GetProvider(providerType interfaces.ProviderType) (interfaces.IPostgreSQLProvider, bool) {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	provider, exists := pf.providers[providerType]
	return provider, exists
}

// RemoveProvider remove um provider
func (pf *ProviderFactory) RemoveProvider(providerType interfaces.ProviderType) {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	delete(pf.providers, providerType)
}

// ClearProviders limpa todos os providers
func (pf *ProviderFactory) ClearProviders() {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	pf.providers = make(map[interfaces.ProviderType]interfaces.IPostgreSQLProvider)
}

// Instance global da factory
var defaultFactory = NewProviderFactory()

// GetDefaultFactory retorna a factory padrão
func GetDefaultFactory() interfaces.IProviderFactory {
	return defaultFactory
}

// SetDefaultFactory define uma nova factory padrão
func SetDefaultFactory(factory interfaces.IProviderFactory) {
	defaultFactory = factory
}

// Quick factory methods para uso comum

// NewPGXProvider cria um provider PGX usando a factory padrão
func NewPGXProvider() (interfaces.IPostgreSQLProvider, error) {
	return defaultFactory.CreateProvider(interfaces.ProviderTypePGX)
}

// MustNewPGXProvider cria um provider PGX e entra em pânico se falhar
func MustNewPGXProvider() interfaces.IPostgreSQLProvider {
	provider, err := NewPGXProvider()
	if err != nil {
		panic(fmt.Sprintf("failed to create PGX provider: %v", err))
	}
	return provider
}

// ListAvailableProviders lista todos os providers disponíveis
func ListAvailableProviders() []interfaces.ProviderType {
	return defaultFactory.ListProviders()
}

// IsProviderAvailable verifica se um provider está disponível
func IsProviderAvailable(providerType interfaces.ProviderType) bool {
	providers := ListAvailableProviders()
	for _, provider := range providers {
		if provider == providerType {
			return true
		}
	}
	return false
}

// GetProviderInfo retorna informações sobre um provider
func GetProviderInfo(providerType interfaces.ProviderType) (map[string]interface{}, error) {
	provider, err := defaultFactory.CreateProvider(providerType)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"name":               provider.Name(),
		"version":            provider.Version(),
		"driver":             provider.GetDriverName(),
		"supported_features": provider.GetSupportedFeatures(),
	}

	return info, nil
}

// ValidateProviderConfig valida uma configuração para um provider
func ValidateProviderConfig(providerType interfaces.ProviderType, config interfaces.IConfig) error {
	provider, err := defaultFactory.CreateProvider(providerType)
	if err != nil {
		return err
	}

	return provider.ValidateConfig(config)
}
