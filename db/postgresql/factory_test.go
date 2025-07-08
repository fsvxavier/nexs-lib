package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStrategy implementa ProviderStrategy para testes
type MockStrategy struct {
	mock.Mock
}

func (m *MockStrategy) CreateConnection(ctx context.Context, config *common.Config) (common.IConn, error) {
	args := m.Called(ctx, config)
	return args.Get(0).(common.IConn), args.Error(1)
}

func (m *MockStrategy) CreatePool(ctx context.Context, config *common.Config) (common.IPool, error) {
	args := m.Called(ctx, config)
	return args.Get(0).(common.IPool), args.Error(1)
}

func (m *MockStrategy) CreateBatch() (common.IBatch, error) {
	args := m.Called()
	return args.Get(0).(common.IBatch), args.Error(1)
}

func (m *MockStrategy) ValidateConfig(config *common.Config) error {
	args := m.Called(config)
	return args.Error(0)
}

// MockConn implementa common.IConn para testes
type MockConn struct {
	mock.Mock
}

func (m *MockConn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	mockArgs := m.Called(ctx, dst, query, args)
	return mockArgs.Error(0)
}

func (m *MockConn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	mockArgs := m.Called(ctx, dst, query, args)
	return mockArgs.Error(0)
}

func (m *MockConn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(*int), mockArgs.Error(1)
}

func (m *MockConn) Query(ctx context.Context, query string, args ...interface{}) (common.IRows, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(common.IRows), mockArgs.Error(1)
}

func (m *MockConn) QueryRow(ctx context.Context, query string, args ...interface{}) (common.IRow, error) {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Get(0).(common.IRow), mockArgs.Error(1)
}

func (m *MockConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	mockArgs := m.Called(ctx, query, args)
	return mockArgs.Error(0)
}

func (m *MockConn) SendBatch(ctx context.Context, batch common.IBatch) (common.IBatchResults, error) {
	mockArgs := m.Called(ctx, batch)
	return mockArgs.Get(0).(common.IBatchResults), mockArgs.Error(1)
}

func (m *MockConn) Ping(ctx context.Context) error {
	mockArgs := m.Called(ctx)
	return mockArgs.Error(0)
}

func (m *MockConn) Close(ctx context.Context) error {
	mockArgs := m.Called(ctx)
	return mockArgs.Error(0)
}

func (m *MockConn) BeginTransaction(ctx context.Context) (common.ITransaction, error) {
	mockArgs := m.Called(ctx)
	return mockArgs.Get(0).(common.ITransaction), mockArgs.Error(1)
}

func (m *MockConn) BeginTransactionWithOptions(ctx context.Context, opts *common.TxOptions) (common.ITransaction, error) {
	mockArgs := m.Called(ctx, opts)
	return mockArgs.Get(0).(common.ITransaction), mockArgs.Error(1)
}

// MockPool implementa common.IPool para testes
type MockPool struct {
	mock.Mock
}

func (m *MockPool) Acquire(ctx context.Context) (common.IConn, error) {
	args := m.Called(ctx)
	return args.Get(0).(common.IConn), args.Error(1)
}

func (m *MockPool) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPool) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPool) Stats() *common.PoolStats {
	args := m.Called()
	return args.Get(0).(*common.PoolStats)
}

// MockBatch implementa common.IBatch para testes
type MockBatch struct {
	mock.Mock
}

func (m *MockBatch) Queue(query string, arguments ...any) {
	m.Called(query, arguments)
}

func (m *MockBatch) GetBatch() interface{} {
	args := m.Called()
	return args.Get(0)
}

func TestDatabaseFactory_NewDatabaseFactory(t *testing.T) {
	factory := NewDatabaseFactory()

	assert.NotNil(t, factory)
	assert.NotNil(t, factory.strategies)

	// Verifica se todas as estratégias padrão foram registradas
	supportedProviders := factory.GetSupportedProviders()
	assert.Contains(t, supportedProviders, PGX)
	assert.Contains(t, supportedProviders, PQ)
	assert.Contains(t, supportedProviders, GORM)
	assert.Len(t, supportedProviders, 3)
}

func TestDatabaseFactory_RegisterStrategy(t *testing.T) {
	factory := NewDatabaseFactory()
	mockStrategy := &MockStrategy{}
	customProvider := ProviderType("custom")

	// Registra nova estratégia
	factory.RegisterStrategy(customProvider, mockStrategy)

	// Verifica se foi registrada
	supportedProviders := factory.GetSupportedProviders()
	assert.Contains(t, supportedProviders, customProvider)
	assert.Len(t, supportedProviders, 4) // 3 padrão + 1 custom
}

func TestDatabaseFactory_CreateConnection_Success(t *testing.T) {
	factory := NewDatabaseFactory()
	mockStrategy := &MockStrategy{}
	mockConn := &MockConn{}

	testProvider := ProviderType("test")
	factory.RegisterStrategy(testProvider, mockStrategy)

	ctx := context.Background()
	config := &common.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		User:     "testuser",
	}

	// Configura expectativas do mock
	mockStrategy.On("ValidateConfig", config).Return(nil)
	mockStrategy.On("CreateConnection", ctx, config).Return(mockConn, nil)

	// Executa o teste
	conn, err := factory.CreateConnection(ctx, testProvider, config)

	// Verifica resultados
	assert.NoError(t, err)
	assert.Equal(t, mockConn, conn)
	mockStrategy.AssertExpectations(t)
}

func TestDatabaseFactory_CreateConnection_ValidationErrors(t *testing.T) {
	factory := NewDatabaseFactory()

	t.Run("Nil Context", func(t *testing.T) {
		config := &common.Config{}
		conn, err := factory.CreateConnection(nil, PGX, config)

		assert.Nil(t, conn)
		assert.ErrorIs(t, err, ErrInvalidContext)
	})

	t.Run("Nil Config", func(t *testing.T) {
		ctx := context.Background()
		conn, err := factory.CreateConnection(ctx, PGX, nil)

		assert.Nil(t, conn)
		assert.ErrorIs(t, err, ErrNilConfig)
	})

	t.Run("Invalid Provider", func(t *testing.T) {
		ctx := context.Background()
		config := &common.Config{}
		conn, err := factory.CreateConnection(ctx, ProviderType("invalid"), config)

		assert.Nil(t, conn)
		assert.ErrorIs(t, err, ErrInvalidProviderType)
	})
}

func TestDatabaseFactory_CreateConnection_ConfigValidationFails(t *testing.T) {
	factory := NewDatabaseFactory()
	mockStrategy := &MockStrategy{}

	testProvider := ProviderType("test")
	factory.RegisterStrategy(testProvider, mockStrategy)

	ctx := context.Background()
	config := &common.Config{}
	configError := errors.New("configuração inválida")

	// Configura expectativas do mock
	mockStrategy.On("ValidateConfig", config).Return(configError)

	// Executa o teste
	conn, err := factory.CreateConnection(ctx, testProvider, config)

	// Verifica resultados
	assert.Nil(t, conn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validação de configuração falhou")
	assert.Contains(t, err.Error(), configError.Error())
	mockStrategy.AssertExpectations(t)
}

func TestDatabaseFactory_CreatePool_Success(t *testing.T) {
	factory := NewDatabaseFactory()
	mockStrategy := &MockStrategy{}
	mockPool := &MockPool{}

	testProvider := ProviderType("test")
	factory.RegisterStrategy(testProvider, mockStrategy)

	ctx := context.Background()
	config := &common.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		User:     "testuser",
	}

	// Configura expectativas do mock
	mockStrategy.On("ValidateConfig", config).Return(nil)
	mockStrategy.On("CreatePool", ctx, config).Return(mockPool, nil)

	// Executa o teste
	pool, err := factory.CreatePool(ctx, testProvider, config)

	// Verifica resultados
	assert.NoError(t, err)
	assert.Equal(t, mockPool, pool)
	mockStrategy.AssertExpectations(t)
}

func TestDatabaseFactory_CreateBatch_Success(t *testing.T) {
	factory := NewDatabaseFactory()
	mockStrategy := &MockStrategy{}
	mockBatch := &MockBatch{}

	testProvider := ProviderType("test")
	factory.RegisterStrategy(testProvider, mockStrategy)

	// Configura expectativas do mock
	mockStrategy.On("CreateBatch").Return(mockBatch, nil)

	// Executa o teste
	batch, err := factory.CreateBatch(testProvider)

	// Verifica resultados
	assert.NoError(t, err)
	assert.Equal(t, mockBatch, batch)
	mockStrategy.AssertExpectations(t)
}

func TestDatabaseFactory_CreateBatch_InvalidProvider(t *testing.T) {
	factory := NewDatabaseFactory()

	// Executa o teste com provider inválido
	batch, err := factory.CreateBatch(ProviderType("invalid"))

	// Verifica resultados
	assert.Nil(t, batch)
	assert.ErrorIs(t, err, ErrInvalidProviderType)
}

func TestGetFactory(t *testing.T) {
	factory := GetFactory()
	assert.NotNil(t, factory)
	assert.Same(t, defaultFactory, factory)
}

func TestSetFactory(t *testing.T) {
	originalFactory := GetFactory()
	newFactory := NewDatabaseFactory()

	// Define nova factory
	SetFactory(newFactory)

	// Verifica se foi alterada
	assert.Same(t, newFactory, GetFactory())
	assert.NotSame(t, originalFactory, GetFactory())

	// Restaura factory original
	SetFactory(originalFactory)
	assert.Same(t, originalFactory, GetFactory())
}

func TestDatabaseFactory_GetSupportedProviders(t *testing.T) {
	factory := NewDatabaseFactory()

	providers := factory.GetSupportedProviders()

	// Verifica que todos os providers padrão estão presentes
	expectedProviders := map[ProviderType]bool{
		PGX:  false,
		PQ:   false,
		GORM: false,
	}

	for _, provider := range providers {
		if _, exists := expectedProviders[provider]; exists {
			expectedProviders[provider] = true
		}
	}

	// Verifica se todos foram encontrados
	for provider, found := range expectedProviders {
		assert.True(t, found, "Provider %s não foi encontrado", provider)
	}
}
