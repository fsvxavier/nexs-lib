package factory

import (
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
)

// Factory é uma implementação de interfaces.Factory
type Factory struct {
	providers map[interfaces.IDType]interfaces.IDProvider
	mu        sync.RWMutex
}

// NewFactory cria uma nova instância de Factory
func NewFactory() *Factory {
	return &Factory{
		providers: make(map[interfaces.IDType]interfaces.IDProvider),
	}
}

// RegisterProvider registra um provedor de ID
func (f *Factory) RegisterProvider(provider interfaces.IDProvider) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	idType := provider.Type()
	if _, exists := f.providers[idType]; exists {
		return fmt.Errorf("provider do tipo %s já está registrado", idType)
	}

	f.providers[idType] = provider
	return nil
}

// GetProvider retorna um provedor de ID do tipo especificado
func (f *Factory) GetProvider(idType interfaces.IDType) (interfaces.IDProvider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, exists := f.providers[idType]
	if !exists {
		return nil, fmt.Errorf("provider do tipo %s não encontrado", idType)
	}

	return provider, nil
}
