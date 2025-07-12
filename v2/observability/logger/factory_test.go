package logger

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// MockProvider implementa interfaces.Provider para testes
type MockProvider struct {
	name          string
	version       string
	configured    bool
	config        interfaces.Config
	healthError   error
	logMessages   []string
	level         interfaces.Level
	fields        []interfaces.Field
	contextFields map[string]interface{}
	configureFunc func(interfaces.Config) error
	mu            sync.RWMutex
}

func NewMockProvider(name, version string) *MockProvider {
	return &MockProvider{
		name:          name,
		version:       version,
		logMessages:   make([]string, 0),
		contextFields: make(map[string]interface{}),
	}
}

func (m *MockProvider) Configure(config interfaces.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.configureFunc != nil {
		return m.configureFunc(config)
	}

	m.configured = true
	m.config = config
	m.level = config.Level
	return nil
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Version() string {
	return m.version
}

func (m *MockProvider) HealthCheck() error {
	return m.healthError
}

func (m *MockProvider) SetHealthError(err error) {
	m.healthError = err
}

func (m *MockProvider) Trace(ctx context.Context, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.TraceLevel {
		m.addLogMessage("TRACE", msg, fields...)
	}
}

func (m *MockProvider) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.DebugLevel {
		m.addLogMessage("DEBUG", msg, fields...)
	}
}

func (m *MockProvider) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.InfoLevel {
		m.addLogMessage("INFO", msg, fields...)
	}
}

func (m *MockProvider) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.WarnLevel {
		m.addLogMessage("WARN", msg, fields...)
	}
}

func (m *MockProvider) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.ErrorLevel {
		m.addLogMessage("ERROR", msg, fields...)
	}
}

func (m *MockProvider) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.FatalLevel {
		m.addLogMessage("FATAL", msg, fields...)
	}
}

func (m *MockProvider) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.PanicLevel {
		m.addLogMessage("PANIC", msg, fields...)
	}
}

func (m *MockProvider) Tracef(ctx context.Context, format string, args ...interface{}) {
	if m.level <= interfaces.TraceLevel {
		m.addLogMessage("TRACE", fmt.Sprintf(format, args...))
	}
}

func (m *MockProvider) Debugf(ctx context.Context, format string, args ...interface{}) {
	if m.level <= interfaces.DebugLevel {
		m.addLogMessage("DEBUG", fmt.Sprintf(format, args...))
	}
}

func (m *MockProvider) Infof(ctx context.Context, format string, args ...interface{}) {
	if m.level <= interfaces.InfoLevel {
		m.addLogMessage("INFO", fmt.Sprintf(format, args...))
	}
}

func (m *MockProvider) Warnf(ctx context.Context, format string, args ...interface{}) {
	if m.level <= interfaces.WarnLevel {
		m.addLogMessage("WARN", fmt.Sprintf(format, args...))
	}
}

func (m *MockProvider) Errorf(ctx context.Context, format string, args ...interface{}) {
	if m.level <= interfaces.ErrorLevel {
		m.addLogMessage("ERROR", fmt.Sprintf(format, args...))
	}
}

func (m *MockProvider) Fatalf(ctx context.Context, format string, args ...interface{}) {
	if m.level <= interfaces.FatalLevel {
		m.addLogMessage("FATAL", fmt.Sprintf(format, args...))
	}
}

func (m *MockProvider) Panicf(ctx context.Context, format string, args ...interface{}) {
	if m.level <= interfaces.PanicLevel {
		m.addLogMessage("PANIC", fmt.Sprintf(format, args...))
	}
}

func (m *MockProvider) TraceWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.TraceLevel {
		m.addLogMessage("TRACE", fmt.Sprintf("[%s] %s", code, msg), fields...)
	}
}

func (m *MockProvider) DebugWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.DebugLevel {
		m.addLogMessage("DEBUG", fmt.Sprintf("[%s] %s", code, msg), fields...)
	}
}

func (m *MockProvider) InfoWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.InfoLevel {
		m.addLogMessage("INFO", fmt.Sprintf("[%s] %s", code, msg), fields...)
	}
}

func (m *MockProvider) WarnWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.WarnLevel {
		m.addLogMessage("WARN", fmt.Sprintf("[%s] %s", code, msg), fields...)
	}
}

func (m *MockProvider) ErrorWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	if m.level <= interfaces.ErrorLevel {
		m.addLogMessage("ERROR", fmt.Sprintf("[%s] %s", code, msg), fields...)
	}
}

func (m *MockProvider) WithFields(fields ...interfaces.Field) interfaces.Logger {
	clone := m.clone()
	clone.fields = append(clone.fields, fields...)
	return clone
}

func (m *MockProvider) WithContext(ctx context.Context) interfaces.Logger {
	clone := m.clone()
	// Extrai campos do contexto se necessário
	return clone
}

func (m *MockProvider) WithError(err error) interfaces.Logger {
	clone := m.clone()
	clone.fields = append(clone.fields, interfaces.Error(err))
	return clone
}

func (m *MockProvider) WithTraceID(traceID string) interfaces.Logger {
	clone := m.clone()
	clone.contextFields["trace_id"] = traceID
	return clone
}

func (m *MockProvider) WithSpanID(spanID string) interfaces.Logger {
	clone := m.clone()
	clone.contextFields["span_id"] = spanID
	return clone
}

func (m *MockProvider) SetLevel(level interfaces.Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.level = level
}

func (m *MockProvider) GetLevel() interfaces.Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.level
}

func (m *MockProvider) IsLevelEnabled(level interfaces.Level) bool {
	return m.level <= level
}

func (m *MockProvider) Clone() interfaces.Logger {
	return m.clone()
}

func (m *MockProvider) Flush() error {
	return nil
}

func (m *MockProvider) Close() error {
	return nil
}

func (m *MockProvider) clone() *MockProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := &MockProvider{
		name:          m.name,
		version:       m.version,
		configured:    m.configured,
		config:        m.config,
		healthError:   m.healthError,
		logMessages:   make([]string, len(m.logMessages)),
		level:         m.level,
		fields:        make([]interfaces.Field, len(m.fields)),
		contextFields: make(map[string]interface{}),
	}

	copy(clone.logMessages, m.logMessages)
	copy(clone.fields, m.fields)

	for k, v := range m.contextFields {
		clone.contextFields[k] = v
	}

	return clone
}

func (m *MockProvider) addLogMessage(level, msg string, fields ...interfaces.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()

	allFields := append(m.fields, fields...)
	message := fmt.Sprintf("[%s] %s", level, msg)

	if len(allFields) > 0 {
		message += " |"
		for _, field := range allFields {
			message += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}

	for k, v := range m.contextFields {
		message += fmt.Sprintf(" %s=%v", k, v)
	}

	m.logMessages = append(m.logMessages, message)
}

func (m *MockProvider) GetLogMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := make([]string, len(m.logMessages))
	copy(messages, m.logMessages)
	return messages
}

func (m *MockProvider) ClearLogMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logMessages = m.logMessages[:0]
}

func TestNewFactory(t *testing.T) {
	factory := NewFactory()

	if factory == nil {
		t.Fatal("Expected factory to be created")
	}

	if factory.providers == nil {
		t.Error("Expected providers map to be initialized")
	}

	if len(factory.providers) != 0 {
		t.Errorf("Expected empty providers map, got %d providers", len(factory.providers))
	}

	// Verifica se tem uma configuração padrão
	if factory.defaultConfig.Level != interfaces.InfoLevel {
		t.Errorf("Expected default level to be InfoLevel, got %v", factory.defaultConfig.Level)
	}
}

func TestFactoryRegisterProvider(t *testing.T) {
	factory := NewFactory()
	mockProvider := NewMockProvider("mock", "1.0.0")

	factory.RegisterProvider("mock", mockProvider)

	// Verifica se o provider foi registrado
	if len(factory.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(factory.providers))
	}

	provider, exists := factory.providers["mock"]
	if !exists {
		t.Error("Expected mock provider to be registered")
	}

	if provider != mockProvider {
		t.Error("Expected registered provider to be the same instance")
	}
}

func TestFactoryGetProvider(t *testing.T) {
	factory := NewFactory()
	mockProvider := NewMockProvider("mock", "1.0.0")

	// Testa provider inexistente
	provider, exists := factory.GetProvider("nonexistent")
	if exists {
		t.Error("Expected provider to not exist")
	}
	if provider != nil {
		t.Error("Expected nil provider for nonexistent provider")
	}

	// Registra e testa provider existente
	factory.RegisterProvider("mock", mockProvider)
	provider, exists = factory.GetProvider("mock")
	if !exists {
		t.Error("Expected provider to exist")
	}
	if provider != mockProvider {
		t.Error("Expected to get the same provider instance")
	}
}

func TestFactoryListProviders(t *testing.T) {
	factory := NewFactory()

	// Lista vazia inicialmente
	providers := factory.ListProviders()
	if len(providers) != 0 {
		t.Errorf("Expected empty list, got %v", providers)
	}

	// Adiciona providers
	factory.RegisterProvider("mock1", NewMockProvider("mock1", "1.0.0"))
	factory.RegisterProvider("mock2", NewMockProvider("mock2", "2.0.0"))
	factory.RegisterProvider("mock3", NewMockProvider("mock3", "3.0.0"))

	providers = factory.ListProviders()
	if len(providers) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(providers))
	}

	// Verifica se todos os nomes estão presentes
	expectedNames := map[string]bool{"mock1": true, "mock2": true, "mock3": true}
	for _, name := range providers {
		if !expectedNames[name] {
			t.Errorf("Unexpected provider name: %s", name)
		}
		delete(expectedNames, name)
	}

	if len(expectedNames) > 0 {
		t.Errorf("Missing provider names: %v", expectedNames)
	}
}

func TestFactoryCreateProvider(t *testing.T) {
	factory := NewFactory()
	config := TestConfig()

	// Testa provider inexistente
	provider, err := factory.CreateProvider("nonexistent", config)
	if err == nil {
		t.Error("Expected error for nonexistent provider")
	}
	if provider != nil {
		t.Error("Expected nil provider for nonexistent provider")
	}

	// Registra e testa provider existente
	mockProvider := NewMockProvider("mock", "1.0.0")
	factory.RegisterProvider("mock", mockProvider)

	provider, err = factory.CreateProvider("mock", config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if provider == nil {
		t.Error("Expected provider to be created")
	}

	// Verifica se foi configurado
	if !mockProvider.configured {
		t.Error("Expected provider to be configured")
	}
}

func TestFactoryCreateLogger(t *testing.T) {
	factory := NewFactory()
	config := TestConfig()

	// Testa sem providers registrados
	logger, err := factory.CreateLogger("test", config)
	if err == nil {
		t.Error("Expected error when no providers are registered")
	}
	if logger != nil {
		t.Error("Expected nil logger when no providers are registered")
	}

	// Registra um provider
	mockProvider := NewMockProvider("mock", "1.0.0")
	factory.RegisterProvider("mock", mockProvider)

	// Cria logger
	logger, err = factory.CreateLogger("test", config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if logger == nil {
		t.Error("Expected logger to be created")
	}

	// Verifica se o provider foi configurado
	if !mockProvider.configured {
		t.Error("Expected provider to be configured")
	}
}

func TestFactoryCreateLoggerMultipleProviders(t *testing.T) {
	factory := NewFactory()
	config := TestConfig()

	// Registra múltiplos providers
	mockProvider1 := NewMockProvider("mock1", "1.0.0")
	mockProvider2 := NewMockProvider("mock2", "2.0.0")
	factory.RegisterProvider("mock1", mockProvider1)
	factory.RegisterProvider("mock2", mockProvider2)

	// Cria logger - deve usar o primeiro disponível
	logger, err := factory.CreateLogger("test", config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if logger == nil {
		t.Error("Expected logger to be created")
	}
}

func TestFactoryRegisterDefaultProviders(t *testing.T) {
	factory := NewFactory()

	// Antes de registrar
	if len(factory.providers) != 0 {
		t.Errorf("Expected 0 providers initially, got %d", len(factory.providers))
	}

	// Registra providers padrão
	factory.RegisterDefaultProviders()

	// Verifica se os providers padrão foram registrados
	expectedProviders := []string{"zap", "slog", "zerolog"}
	for _, name := range expectedProviders {
		if _, exists := factory.providers[name]; !exists {
			t.Errorf("Expected provider %s to be registered", name)
		}
	}

	if len(factory.providers) != len(expectedProviders) {
		t.Errorf("Expected %d providers, got %d", len(expectedProviders), len(factory.providers))
	}
}

func TestFactoryCreateLoggerWithInvalidConfig(t *testing.T) {
	factory := NewFactory()
	mockProvider := NewMockProvider("mock", "1.0.0")
	factory.RegisterProvider("mock", mockProvider)

	// Config inválido
	invalidConfig := DefaultConfig()
	invalidConfig.ServiceName = "" // Inválido

	logger, err := factory.CreateLogger("test", invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
	if logger != nil {
		t.Error("Expected nil logger for invalid config")
	}
}

func TestFactoryCreateProviderConfigurationError(t *testing.T) {
	factory := NewFactory()

	// Mock provider que falha na configuração
	mockProvider := NewMockProvider("failing", "1.0.0")
	mockProvider.configureFunc = func(config interfaces.Config) error {
		return fmt.Errorf("configuration failed")
	}
	factory.RegisterProvider("failing", mockProvider)

	config := TestConfig()
	provider, err := factory.CreateProvider("failing", config)
	if err == nil {
		t.Error("Expected configuration error")
	}
	if provider != nil {
		t.Error("Expected nil provider on configuration error")
	}
}

func TestFactoryThreadSafety(t *testing.T) {
	factory := NewFactory()

	// Testa operações concorrentes
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup

	// Registra providers concorrentemente
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				providerName := fmt.Sprintf("provider_%d_%d", id, j)
				provider := NewMockProvider(providerName, "1.0.0")
				factory.RegisterProvider(providerName, provider)
			}
		}(i)
	}

	// Lista providers concorrentemente
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				factory.ListProviders()
			}
		}()
	}

	// Busca providers concorrentemente
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				providerName := fmt.Sprintf("provider_%d_%d", id, j)
				factory.GetProvider(providerName)
			}
		}(i)
	}

	wg.Wait()

	// Verifica o resultado final
	providers := factory.ListProviders()
	expectedCount := numGoroutines * numOperations
	if len(providers) != expectedCount {
		t.Errorf("Expected %d providers, got %d", expectedCount, len(providers))
	}
}

func TestFactoryDefaultConfig(t *testing.T) {
	factory := NewFactory()

	// Verifica se a configuração padrão é válida
	config := factory.defaultConfig
	if err := interfaces.ValidateConfig(config); err != nil {
		t.Errorf("Default config should be valid, got error: %v", err)
	}

	// Verifica valores esperados
	if config.Level != interfaces.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}

	if config.Format != interfaces.JSONFormat {
		t.Errorf("Expected JSONFormat, got %v", config.Format)
	}

	if config.ServiceName != "unknown" {
		t.Errorf("Expected 'unknown', got %s", config.ServiceName)
	}
}

func TestFactorySetDefaultConfig(t *testing.T) {
	factory := NewFactory()

	// Cria uma nova configuração
	newConfig := ProductionConfig()
	factory.SetDefaultConfig(newConfig)

	// Verifica se foi definida
	if factory.defaultConfig.Environment != "production" {
		t.Errorf("Expected production environment, got %s", factory.defaultConfig.Environment)
	}

	if factory.defaultConfig.Level != interfaces.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", factory.defaultConfig.Level)
	}
}

func TestFactoryGetDefaultConfig(t *testing.T) {
	factory := NewFactory()

	config := factory.GetDefaultConfig()

	// Verifica se retorna a configuração padrão
	if config.ServiceName != "unknown" {
		t.Errorf("Expected 'unknown', got %s", config.ServiceName)
	}

	if config.Level != interfaces.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}
}

func TestFactoryProviderHealthCheck(t *testing.T) {
	factory := NewFactory()

	// Provider saudável
	healthyProvider := NewMockProvider("healthy", "1.0.0")
	factory.RegisterProvider("healthy", healthyProvider)

	// Provider com erro
	unhealthyProvider := NewMockProvider("unhealthy", "1.0.0")
	unhealthyProvider.SetHealthError(fmt.Errorf("health check failed"))
	factory.RegisterProvider("unhealthy", unhealthyProvider)

	// Testa health check de provider saudável
	provider, exists := factory.GetProvider("healthy")
	if !exists {
		t.Fatal("Expected healthy provider to exist")
	}

	if err := provider.HealthCheck(); err != nil {
		t.Errorf("Expected healthy provider to pass health check, got error: %v", err)
	}

	// Testa health check de provider não saudável
	provider, exists = factory.GetProvider("unhealthy")
	if !exists {
		t.Fatal("Expected unhealthy provider to exist")
	}

	if err := provider.HealthCheck(); err == nil {
		t.Error("Expected unhealthy provider to fail health check")
	}
}

func TestFactoryProviderInfo(t *testing.T) {
	factory := NewFactory()

	expectedName := "test-provider"
	expectedVersion := "v2.1.0"

	provider := NewMockProvider(expectedName, expectedVersion)
	factory.RegisterProvider("test", provider)

	retrievedProvider, exists := factory.GetProvider("test")
	if !exists {
		t.Fatal("Expected provider to exist")
	}

	if retrievedProvider.Name() != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, retrievedProvider.Name())
	}

	if retrievedProvider.Version() != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, retrievedProvider.Version())
	}
}

func TestGlobalLoggerMethods(t *testing.T) {
	// Setup global logger for testing
	originalLogger := GetCurrentLogger()
	defer func() {
		SetCurrentLogger(originalLogger)
	}()

	// Create test logger
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.TraceLevel // Set to TraceLevel to include all log levels

	provider.Configure(config)
	testLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(testLogger)

	ctx := context.Background()

	tests := []struct {
		name     string
		testFunc func()
		expected string
	}{
		{
			name: "Global Trace",
			testFunc: func() {
				Trace(ctx, "global trace message")
			},
			expected: "global trace message",
		},
		{
			name: "Global Debug",
			testFunc: func() {
				Debug(ctx, "global debug message")
			},
			expected: "global debug message",
		},
		{
			name: "Global Info",
			testFunc: func() {
				Info(ctx, "global info message")
			},
			expected: "global info message",
		},
		{
			name: "Global Warn",
			testFunc: func() {
				Warn(ctx, "global warn message")
			},
			expected: "global warn message",
		},
		{
			name: "Global Error",
			testFunc: func() {
				Error(ctx, "global error message")
			},
			expected: "global error message",
		},
		{
			name: "Global Tracef",
			testFunc: func() {
				Tracef(ctx, "global tracef %s", "formatted")
			},
			expected: "global tracef formatted",
		},
		{
			name: "Global Debugf",
			testFunc: func() {
				Debugf(ctx, "global debugf %s", "formatted")
			},
			expected: "global debugf formatted",
		},
		{
			name: "Global Infof",
			testFunc: func() {
				Infof(ctx, "global infof %s", "formatted")
			},
			expected: "global infof formatted",
		},
		{
			name: "Global Warnf",
			testFunc: func() {
				Warnf(ctx, "global warnf %s", "formatted")
			},
			expected: "global warnf formatted",
		},
		{
			name: "Global Errorf",
			testFunc: func() {
				Errorf(ctx, "global errorf %s", "formatted")
			},
			expected: "global errorf formatted",
		},
		{
			name: "Global TraceWithCode",
			testFunc: func() {
				TraceWithCode(ctx, "TRACE_CODE", "global trace with code")
			},
			expected: "global trace with code",
		},
		{
			name: "Global DebugWithCode",
			testFunc: func() {
				DebugWithCode(ctx, "DEBUG_CODE", "global debug with code")
			},
			expected: "global debug with code",
		},
		{
			name: "Global InfoWithCode",
			testFunc: func() {
				InfoWithCode(ctx, "INFO_CODE", "global info with code")
			},
			expected: "global info with code",
		},
		{
			name: "Global WarnWithCode",
			testFunc: func() {
				WarnWithCode(ctx, "WARN_CODE", "global warn with code")
			},
			expected: "global warn with code",
		},
		{
			name: "Global ErrorWithCode",
			testFunc: func() {
				ErrorWithCode(ctx, "ERROR_CODE", "global error with code")
			},
			expected: "global error with code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous messages
			provider.mu.Lock()
			provider.logMessages = provider.logMessages[:0]
			provider.mu.Unlock()

			// Execute test function
			tt.testFunc()

			// Verify message was logged
			provider.mu.RLock()
			found := false
			for _, msg := range provider.logMessages {
				if contains(msg, tt.expected) {
					found = true
					break
				}
			}
			provider.mu.RUnlock()

			if !found {
				t.Errorf("Expected message containing '%s' to be logged", tt.expected)
			}
		})
	}
}

func TestGlobalLoggerPanicMethods(t *testing.T) {
	// Setup global logger for testing
	originalLogger := GetCurrentLogger()
	defer func() {
		SetCurrentLogger(originalLogger)
	}()

	// Create test logger
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.TraceLevel // Set to TraceLevel to include all log levels

	provider.Configure(config)
	testLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(testLogger)

	ctx := context.Background()

	tests := []struct {
		name     string
		testFunc func()
		expected string
	}{
		{
			name: "Global Panic",
			testFunc: func() {
				defer func() {
					recover() // Capture panic
				}()
				Panic(ctx, "global panic message")
			},
			expected: "global panic message",
		},
		{
			name: "Global Panicf",
			testFunc: func() {
				defer func() {
					recover() // Capture panic
				}()
				Panicf(ctx, "global panicf %s", "formatted")
			},
			expected: "global panicf formatted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous messages
			provider.mu.Lock()
			provider.logMessages = provider.logMessages[:0]
			provider.mu.Unlock()

			// Execute test function (with panic handling)
			tt.testFunc()

			// Verify message was logged
			provider.mu.RLock()
			found := false
			for _, msg := range provider.logMessages {
				if contains(msg, tt.expected) {
					found = true
					break
				}
			}
			provider.mu.RUnlock()

			if !found {
				t.Errorf("Expected message containing '%s' to be logged", tt.expected)
			}
		})
	}
}

func TestGlobalLoggerFatalMethods(t *testing.T) {
	// Setup global logger for testing
	originalLogger := GetCurrentLogger()
	defer func() {
		SetCurrentLogger(originalLogger)
	}()

	// Create test logger
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.TraceLevel // Set to TraceLevel to include all log levels

	provider.Configure(config)
	testLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(testLogger)

	// Test Fatal and Fatalf methods exist (but don't call them to avoid exit)
	t.Run("Global Fatal method exists", func(t *testing.T) {
		// Just verify the method exists and can be referenced
		_ = Fatal
		t.Log("Global Fatal method is available")
	})

	t.Run("Global Fatalf method exists", func(t *testing.T) {
		// Just verify the method exists and can be referenced
		_ = Fatalf
		t.Log("Global Fatalf method is available")
	})
}

func TestGlobalLoggerUtilityMethods(t *testing.T) {
	// Setup global logger for testing
	originalLogger := GetCurrentLogger()
	defer func() {
		SetCurrentLogger(originalLogger)
	}()

	// Create test logger
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.InfoLevel

	provider.Configure(config)
	testLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(testLogger)

	t.Run("Global WithFields", func(t *testing.T) {
		newLogger := WithFields(String("key", "value"))
		if newLogger == nil {
			t.Error("Expected WithFields to return a logger")
		}
	})

	t.Run("Global WithContext", func(t *testing.T) {
		ctx := context.Background()
		newLogger := WithContext(ctx)
		if newLogger == nil {
			t.Error("Expected WithContext to return a logger")
		}
	})

	t.Run("Global WithError", func(t *testing.T) {
		err := fmt.Errorf("test error")
		newLogger := WithError(err)
		if newLogger == nil {
			t.Error("Expected WithError to return a logger")
		}
	})

	t.Run("Global WithTraceID", func(t *testing.T) {
		traceID := "test-trace-123"
		newLogger := WithTraceID(traceID)
		if newLogger == nil {
			t.Error("Expected WithTraceID to return a logger")
		}
	})

	t.Run("Global WithSpanID", func(t *testing.T) {
		spanID := "test-span-456"
		newLogger := WithSpanID(spanID)
		if newLogger == nil {
			t.Error("Expected WithSpanID to return a logger")
		}
	})
}
