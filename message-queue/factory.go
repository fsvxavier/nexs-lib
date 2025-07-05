package messagequeue

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/activemq"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/kafka"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/rabbitmq"
	"github.com/fsvxavier/nexs-lib/message-queue/providers/sqs"
)

// Factory define a interface para criação de providers
type Factory interface {
	// CreateProvider cria um provider baseado no tipo especificado
	CreateProvider(providerType interfaces.ProviderType, config *config.ProviderConfig) (interfaces.MessageQueueProvider, error)

	// GetProvider retorna um provider existente ou cria um novo
	GetProvider(providerType interfaces.ProviderType) (interfaces.MessageQueueProvider, error)

	// RegisterProvider registra um provider customizado
	RegisterProvider(providerType interfaces.ProviderType, provider interfaces.MessageQueueProvider) error

	// ListProviders lista todos os providers disponíveis
	ListProviders() []interfaces.ProviderType

	// IsProviderAvailable verifica se um provider está disponível e habilitado
	IsProviderAvailable(providerType interfaces.ProviderType) bool

	// GetDefaultProvider retorna o provider padrão configurado
	GetDefaultProvider() (interfaces.MessageQueueProvider, error)

	// Close fecha todos os providers
	Close() error
}

// MessageQueueFactory implementa a factory de message queue providers
type MessageQueueFactory struct {
	config    *config.Config
	providers map[interfaces.ProviderType]interfaces.MessageQueueProvider
	creators  map[interfaces.ProviderType]ProviderCreator
	mutex     sync.RWMutex
}

// ProviderCreator define a função para criação de providers
type ProviderCreator func(config *config.ProviderConfig) (interfaces.MessageQueueProvider, error)

// NewFactory cria uma nova instância da factory
func NewFactory(cfg *config.Config) Factory {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	factory := &MessageQueueFactory{
		config:    cfg,
		providers: make(map[interfaces.ProviderType]interfaces.MessageQueueProvider),
		creators:  make(map[interfaces.ProviderType]ProviderCreator),
	}

	// Registra os creators padrão
	factory.registerDefaultCreators()

	return factory
}

// CreateProvider cria um provider baseado no tipo especificado
func (f *MessageQueueFactory) CreateProvider(providerType interfaces.ProviderType, providerConfig *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	creator, exists := f.creators[providerType]
	if !exists {
		return nil, domainerrors.New(
			"UNSUPPORTED_PROVIDER",
			fmt.Sprintf("provider type '%s' is not supported", providerType),
		).WithType(domainerrors.ErrorTypeValidation).
			WithDetail("provider_type", string(providerType)).
			WithDetail("supported_providers", f.getSupportedProviders())
	}

	// Usa a configuração específica se fornecida, senão usa a configuração global
	config := providerConfig
	if config == nil {
		if globalConfig, exists := f.config.Providers[providerType]; exists {
			config = globalConfig
		} else {
			return nil, domainerrors.New(
				"MISSING_PROVIDER_CONFIG",
				fmt.Sprintf("no configuration found for provider '%s'", providerType),
			).WithType(domainerrors.ErrorTypeValidation).
				WithDetail("provider_type", string(providerType))
		}
	}

	// Verifica se o provider está habilitado
	if !config.Enabled {
		return nil, domainerrors.New(
			"PROVIDER_DISABLED",
			fmt.Sprintf("provider '%s' is disabled", providerType),
		).WithType(domainerrors.ErrorTypeValidation).
			WithDetail("provider_type", string(providerType))
	}

	provider, err := creator(config)
	if err != nil {
		return nil, domainerrors.New(
			"PROVIDER_CREATION_FAILED",
			fmt.Sprintf("failed to create provider '%s'", providerType),
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("provider_type", string(providerType)).
			Wrap("provider creation error", err)
	}

	// Armazena o provider criado
	f.providers[providerType] = provider

	return provider, nil
}

// GetProvider retorna um provider existente ou cria um novo
func (f *MessageQueueFactory) GetProvider(providerType interfaces.ProviderType) (interfaces.MessageQueueProvider, error) {
	f.mutex.RLock()
	provider, exists := f.providers[providerType]
	f.mutex.RUnlock()

	if exists && provider.IsConnected() {
		return provider, nil
	}

	// Provider não existe ou está desconectado, cria um novo
	return f.CreateProvider(providerType, nil)
}

// RegisterProvider registra um provider customizado
func (f *MessageQueueFactory) RegisterProvider(providerType interfaces.ProviderType, provider interfaces.MessageQueueProvider) error {
	if provider == nil {
		return domainerrors.New(
			"INVALID_PROVIDER",
			"provider cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.providers[providerType] = provider

	return nil
}

// ListProviders lista todos os providers disponíveis
func (f *MessageQueueFactory) ListProviders() []interfaces.ProviderType {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	providers := make([]interfaces.ProviderType, 0, len(f.creators))
	for providerType := range f.creators {
		providers = append(providers, providerType)
	}

	return providers
}

// IsProviderAvailable verifica se um provider está disponível e habilitado
func (f *MessageQueueFactory) IsProviderAvailable(providerType interfaces.ProviderType) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// Verifica se o creator existe
	_, creatorExists := f.creators[providerType]
	if !creatorExists {
		return false
	}

	// Verifica se o provider está habilitado na configuração
	if f.config.Providers != nil {
		if providerConfig, exists := f.config.Providers[providerType]; exists {
			return providerConfig.Enabled
		}
	}

	// Se não há configuração específica, assume que está disponível se o creator existe
	return true
}

// GetDefaultProvider retorna o provider padrão configurado
func (f *MessageQueueFactory) GetDefaultProvider() (interfaces.MessageQueueProvider, error) {
	return f.GetProvider(f.config.Global.DefaultProvider)
}

// Close fecha todos os providers
func (f *MessageQueueFactory) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var errors []error

	for providerType, provider := range f.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close provider '%s': %w", providerType, err))
		}
	}

	// Limpa o mapa de providers
	f.providers = make(map[interfaces.ProviderType]interfaces.MessageQueueProvider)

	if len(errors) > 0 {
		return domainerrors.New(
			"PROVIDERS_CLOSE_FAILED",
			"failed to close some providers",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("error_count", len(errors)).
			WithDetail("errors", errors)
	}

	return nil
}

// registerDefaultCreators registra os creators padrão para todos os providers
func (f *MessageQueueFactory) registerDefaultCreators() {
	f.creators[interfaces.ProviderKafka] = func(config *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
		return kafka.NewKafkaProvider(config)
	}

	f.creators[interfaces.ProviderRabbitMQ] = func(config *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
		return rabbitmq.NewRabbitMQProvider(config)
	}

	f.creators[interfaces.ProviderSQS] = func(config *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
		return sqs.NewSQSProvider(config)
	}

	f.creators[interfaces.ProviderActiveMQ] = func(config *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
		return activemq.NewActiveMQProvider(config)
	}
}

// getSupportedProviders retorna uma lista dos providers suportados
func (f *MessageQueueFactory) getSupportedProviders() []string {
	providers := make([]string, 0, len(f.creators))
	for providerType := range f.creators {
		providers = append(providers, string(providerType))
	}
	return providers
}

// Manager representa o gerenciador principal do sistema de message queue
type Manager struct {
	factory   Factory
	config    *config.Config
	providers map[interfaces.ProviderType]interfaces.MessageQueueProvider
	mutex     sync.RWMutex
}

// NewManager cria um novo gerenciador de message queue
func NewManager(cfg *config.Config) *Manager {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &Manager{
		factory:   NewFactory(cfg),
		config:    cfg,
		providers: make(map[interfaces.ProviderType]interfaces.MessageQueueProvider),
	}
}

// GetProvider retorna um provider específico
func (m *Manager) GetProvider(providerType interfaces.ProviderType) (interfaces.MessageQueueProvider, error) {
	m.mutex.RLock()
	provider, exists := m.providers[providerType]
	m.mutex.RUnlock()

	if exists && provider.IsConnected() {
		return provider, nil
	}

	// Provider não existe ou está desconectado, busca da factory
	provider, err := m.factory.GetProvider(providerType)
	if err != nil {
		return nil, err
	}

	m.mutex.Lock()
	m.providers[providerType] = provider
	m.mutex.Unlock()

	return provider, nil
}

// GetDefaultProvider retorna o provider padrão configurado
func (m *Manager) GetDefaultProvider() (interfaces.MessageQueueProvider, error) {
	return m.GetProvider(m.config.Global.DefaultProvider)
}

// CreateProducer cria um producer usando o provider padrão
func (m *Manager) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	provider, err := m.GetDefaultProvider()
	if err != nil {
		return nil, err
	}

	return provider.CreateProducer(config)
}

// CreateConsumer cria um consumer usando o provider padrão
func (m *Manager) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	provider, err := m.GetDefaultProvider()
	if err != nil {
		return nil, err
	}

	return provider.CreateConsumer(config)
}

// CreateProducerWithProvider cria um producer usando um provider específico
func (m *Manager) CreateProducerWithProvider(providerType interfaces.ProviderType, config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	provider, err := m.GetProvider(providerType)
	if err != nil {
		return nil, err
	}

	return provider.CreateProducer(config)
}

// CreateConsumerWithProvider cria um consumer usando um provider específico
func (m *Manager) CreateConsumerWithProvider(providerType interfaces.ProviderType, config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	provider, err := m.GetProvider(providerType)
	if err != nil {
		return nil, err
	}

	return provider.CreateConsumer(config)
}

// HealthCheck verifica a saúde de todos os providers
func (m *Manager) HealthCheck(ctx context.Context) error {
	m.mutex.RLock()
	providers := make(map[interfaces.ProviderType]interfaces.MessageQueueProvider)
	for k, v := range m.providers {
		providers[k] = v
	}
	m.mutex.RUnlock()

	var errors []error

	for providerType, provider := range providers {
		if err := provider.HealthCheck(ctx); err != nil {
			errors = append(errors, fmt.Errorf("provider '%s' health check failed: %w", providerType, err))
		}
	}

	if len(errors) > 0 {
		return domainerrors.New(
			"HEALTH_CHECK_FAILED",
			"some providers failed health check",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("error_count", len(errors)).
			WithDetail("errors", errors)
	}

	return nil
}

// Close fecha o gerenciador e todos os providers
func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var errors []error

	// Fecha todos os providers
	for providerType, provider := range m.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close provider '%s': %w", providerType, err))
		}
	}

	// Fecha a factory
	if err := m.factory.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close factory: %w", err))
	}

	// Limpa o mapa de providers
	m.providers = make(map[interfaces.ProviderType]interfaces.MessageQueueProvider)

	if len(errors) > 0 {
		return domainerrors.New(
			"MANAGER_CLOSE_FAILED",
			"failed to close manager properly",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("error_count", len(errors)).
			WithDetail("errors", errors)
	}

	return nil
}
