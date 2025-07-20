// Package mocks fornece implementações mock para testes dos tracer providers
package mocks

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

// MockProvider é um mock genérico que implementa interfaces.TracerProvider
// Pode ser usado para testar qualquer backend (Datadog, Grafana, New Relic, OpenTelemetry)
type MockProvider struct {
	mu sync.RWMutex

	// Estado do mock
	initCalled     bool
	shutdownCalled bool
	initError      error
	shutdownError  error
	initialized    bool

	// Configuração recebida
	lastConfig interfaces.Config

	// TracerProvider mock
	tracerProvider trace.TracerProvider

	// Contadores para verificações
	initCallCount     int
	shutdownCallCount int

	// Nome do provider para mensagens de erro específicas
	providerName string
}

// NewMockProvider cria uma nova instância do mock provider genérico
func NewMockProvider() *MockProvider {
	return &MockProvider{
		tracerProvider: noop.NewTracerProvider(),
		providerName:   "generic",
	}
}

// NewMockProviderForBackend cria um mock provider para um backend específico
func NewMockProviderForBackend(backendName string) *MockProvider {
	return &MockProvider{
		tracerProvider: noop.NewTracerProvider(),
		providerName:   backendName,
	}
}

// Init simula a inicialização do provider
func (m *MockProvider) Init(ctx context.Context, config interfaces.Config) (trace.TracerProvider, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initCalled = true
	m.initCallCount++
	m.lastConfig = config

	if m.initError != nil {
		return nil, m.initError
	}

	m.initialized = true
	return m.tracerProvider, nil
}

// Shutdown simula o shutdown do provider
func (m *MockProvider) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shutdownCalled = true
	m.shutdownCallCount++

	if m.shutdownError != nil {
		return m.shutdownError
	}

	m.initialized = false
	return nil
}

// Métodos para configurar o comportamento do mock

// SetInitError configura o erro a ser retornado por Init
func (m *MockProvider) SetInitError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initError = err
}

// SetShutdownError configura o erro a ser retornado por Shutdown
func (m *MockProvider) SetShutdownError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownError = err
}

// SetTracerProvider configura o TracerProvider a ser retornado
func (m *MockProvider) SetTracerProvider(tp trace.TracerProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tracerProvider = tp
}

// SetProviderName configura o nome do provider para mensagens de erro
func (m *MockProvider) SetProviderName(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providerName = name
}

// Métodos para verificações em testes

// WasInitCalled retorna se Init foi chamado
func (m *MockProvider) WasInitCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initCalled
}

// WasShutdownCalled retorna se Shutdown foi chamado
func (m *MockProvider) WasShutdownCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shutdownCalled
}

// IsInitialized retorna se o provider está inicializado
func (m *MockProvider) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// GetLastConfig retorna a última configuração recebida
func (m *MockProvider) GetLastConfig() interfaces.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastConfig
}

// GetInitCallCount retorna quantas vezes Init foi chamado
func (m *MockProvider) GetInitCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initCallCount
}

// GetShutdownCallCount retorna quantas vezes Shutdown foi chamado
func (m *MockProvider) GetShutdownCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shutdownCallCount
}

// GetProviderName retorna o nome do provider
func (m *MockProvider) GetProviderName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providerName
}

// Reset reseta o estado do mock
func (m *MockProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initCalled = false
	m.shutdownCalled = false
	m.initError = nil
	m.shutdownError = nil
	m.initialized = false
	m.lastConfig = interfaces.Config{}
	m.initCallCount = 0
	m.shutdownCallCount = 0
	m.tracerProvider = noop.NewTracerProvider()
}

// Providers pré-configurados para backends específicos

// NewDatadogMockProvider cria um mock provider configurado para Datadog
func NewDatadogMockProvider() *MockProvider {
	return NewMockProviderForBackend("datadog")
}

// NewGrafanaMockProvider cria um mock provider configurado para Grafana
func NewGrafanaMockProvider() *MockProvider {
	return NewMockProviderForBackend("grafana")
}

// NewNewRelicMockProvider cria um mock provider configurado para New Relic
func NewNewRelicMockProvider() *MockProvider {
	return NewMockProviderForBackend("newrelic")
}

// NewOpenTelemetryMockProvider cria um mock provider configurado para OpenTelemetry
func NewOpenTelemetryMockProvider() *MockProvider {
	return NewMockProviderForBackend("opentelemetry")
}

// Providers especializados para casos de erro

// ErrorProvider é um mock que sempre retorna erro
type ErrorProvider struct {
	*MockProvider
}

// NewErrorProvider cria um provider que sempre falha
func NewErrorProvider(backendName string) *ErrorProvider {
	p := &ErrorProvider{
		MockProvider: NewMockProviderForBackend(backendName),
	}
	p.SetInitError(errors.New("mock error: failed to initialize " + backendName + " provider"))
	return p
}

// FailingShutdownProvider é um mock que falha no shutdown
type FailingShutdownProvider struct {
	*MockProvider
}

// NewFailingShutdownProvider cria um provider que falha no shutdown
func NewFailingShutdownProvider(backendName string) *FailingShutdownProvider {
	p := &FailingShutdownProvider{
		MockProvider: NewMockProviderForBackend(backendName),
	}
	p.SetShutdownError(errors.New("mock error: failed to shutdown " + backendName + " provider"))
	return p
}

// MockTracerProviderFactory é um factory mock para testes do factory principal
type MockTracerProviderFactory struct {
	mu sync.RWMutex

	// Providers mock para cada backend
	providers map[string]interfaces.TracerProvider

	// Controle de criação
	createCallCount map[string]int
	createError     error
}

// NewMockTracerProviderFactory cria um factory mock
func NewMockTracerProviderFactory() *MockTracerProviderFactory {
	return &MockTracerProviderFactory{
		providers:       make(map[string]interfaces.TracerProvider),
		createCallCount: make(map[string]int),
	}
}

// CreateProvider simula a criação de um provider
func (f *MockTracerProviderFactory) CreateProvider(exporterType string) (interfaces.TracerProvider, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.createCallCount[exporterType]++

	if f.createError != nil {
		return nil, f.createError
	}

	if provider, exists := f.providers[exporterType]; exists {
		return provider, nil
	}

	// Retorna um mock provider padrão
	return NewMockProviderForBackend(exporterType), nil
}

// SetProvider configura um provider mock para um tipo específico
func (f *MockTracerProviderFactory) SetProvider(exporterType string, provider interfaces.TracerProvider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[exporterType] = provider
}

// SetCreateError configura um erro para ser retornado na criação
func (f *MockTracerProviderFactory) SetCreateError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createError = err
}

// GetCreateCallCount retorna quantas vezes CreateProvider foi chamado para um tipo
func (f *MockTracerProviderFactory) GetCreateCallCount(exporterType string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.createCallCount[exporterType]
}

// Reset reseta o estado do factory mock
func (f *MockTracerProviderFactory) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers = make(map[string]interfaces.TracerProvider)
	f.createCallCount = make(map[string]int)
	f.createError = nil
}
