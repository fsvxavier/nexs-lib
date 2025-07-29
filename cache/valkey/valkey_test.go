package valkey

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/cache/valkey/config"
	"github.com/fsvxavier/nexs-lib/cache/valkey/hooks"
	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// MockProvider implementa interfaces.IProvider para testes.
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) NewClient(config interface{}) (interfaces.IClient, error) {
	args := m.Called(config)
	return args.Get(0).(interfaces.IClient), args.Error(1)
}

func (m *MockProvider) ValidateConfig(config interface{}) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockProvider) DefaultConfig() interface{} {
	args := m.Called()
	return args.Get(0)
}

// MockClient implementa interfaces.IClient para testes.
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockClient) Del(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockClient) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) HExists(ctx context.Context, key, field string) (bool, error) {
	args := m.Called(ctx, key, field)
	return args.Bool(0), args.Error(1)
}

func (m *MockClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	args := m.Called(ctx, key, values)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockClient) RPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockClient) LLen(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) SMembers(ctx context.Context, key string) ([]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	args := m.Called(ctx, key, member)
	return args.Bool(0), args.Error(1)
}

func (m *MockClient) ZAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	args := m.Called(ctx, key, members)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	args := m.Called(ctx, key, member)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockClient) Pipeline() interfaces.IPipeline {
	args := m.Called()
	return args.Get(0).(interfaces.IPipeline)
}

func (m *MockClient) TxPipeline() interfaces.ITransaction {
	args := m.Called()
	return args.Get(0).(interfaces.ITransaction)
}

func (m *MockClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	callArgs := m.Called(ctx, script, keys, args)
	return callArgs.Get(0), callArgs.Error(1)
}

func (m *MockClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	callArgs := m.Called(ctx, sha1, keys, args)
	return callArgs.Get(0), callArgs.Error(1)
}

func (m *MockClient) ScriptLoad(ctx context.Context, script string) (string, error) {
	args := m.Called(ctx, script)
	return args.String(0), args.Error(1)
}

func (m *MockClient) Subscribe(ctx context.Context, channels ...string) (interfaces.IPubSub, error) {
	args := m.Called(ctx, channels)
	return args.Get(0).(interfaces.IPubSub), args.Error(1)
}

func (m *MockClient) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClient) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	args := m.Called(ctx, stream, values)
	return args.String(0), args.Error(1)
}

func (m *MockClient) XRead(ctx context.Context, streams map[string]string) ([]interfaces.XMessage, error) {
	args := m.Called(ctx, streams)
	return args.Get(0).([]interfaces.XMessage), args.Error(1)
}

func (m *MockClient) XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]interfaces.XMessage, error) {
	args := m.Called(ctx, group, consumer, streams)
	return args.Get(0).([]interfaces.XMessage), args.Error(1)
}

func (m *MockClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	args := m.Called(ctx, cursor, match, count)
	return args.Get(0).([]string), args.Get(1).(uint64), args.Error(2)
}

func (m *MockClient) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	args := m.Called(ctx, key, cursor, match, count)
	return args.Get(0).([]string), args.Get(1).(uint64), args.Error(2)
}

func (m *MockClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClient) IsHealthy(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.providers)
	assert.NotNil(t, manager.clients)
	assert.Empty(t, manager.providers)
	assert.Empty(t, manager.clients)
}

func TestManager_RegisterProvider(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockProvider.On("Name").Return("test-provider")

	t.Run("successful registration", func(t *testing.T) {
		err := manager.RegisterProvider("test", mockProvider)
		assert.NoError(t, err)

		manager.mu.RLock()
		provider, exists := manager.providers["test"]
		manager.mu.RUnlock()

		assert.True(t, exists)
		assert.Equal(t, mockProvider, provider)
	})

	t.Run("empty name", func(t *testing.T) {
		err := manager.RegisterProvider("", mockProvider)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nome do provider não pode ser vazio")
	})

	t.Run("nil provider", func(t *testing.T) {
		err := manager.RegisterProvider("test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider não pode ser nil")
	})
}

func TestManager_NewClient(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockClient := &MockClient{}

	mockProvider.On("Name").Return("test-provider")
	mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	t.Run("successful client creation", func(t *testing.T) {
		cfg := config.DefaultConfig()
		client, err := manager.NewClient(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, mockClient, client.client)
		mockProvider.AssertExpectations(t)
	})

	t.Run("nil config uses default", func(t *testing.T) {
		client, err := manager.NewClient(nil)

		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("unregistered provider", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Provider = "unknown"

		client, err := manager.NewClient(cfg)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "provider 'unknown' não registrado")
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := &config.Config{
			Provider: "valkey-go",
			Host:     "",
			Port:     0,
		}

		client, err := manager.NewClient(cfg)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "configuração inválida")
	})

	t.Run("provider error", func(t *testing.T) {
		mockProvider2 := &MockProvider{}
		mockProvider2.On("Name").Return("error-provider")
		mockProvider2.On("NewClient", mock.Anything).Return((*MockClient)(nil), errors.New("provider error"))

		manager.RegisterProvider("error-provider", mockProvider2)

		cfg := config.DefaultConfig()
		cfg.Provider = "error-provider"

		client, err := manager.NewClient(cfg)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "erro ao criar cliente do provider")
	})
}

func TestClient_Get(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockClient := &MockClient{}

	mockProvider.On("Name").Return("test-provider")
	mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.KeyPrefix = "test:"
	client, err := manager.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		mockClient.On("Get", ctx, "test:key1").Return("value1", nil).Once()

		result, err := client.Get(ctx, "key1")

		assert.NoError(t, err)
		assert.Equal(t, "value1", result)
		mockClient.AssertExpectations(t)
	})

	t.Run("get with error", func(t *testing.T) {
		mockClient.On("Get", ctx, "test:key2").Return("", errors.New("get error")).Once()

		result, err := client.Get(ctx, "key2")

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "get error")
	})
}

func TestClient_Set(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockClient := &MockClient{}

	mockProvider.On("Name").Return("test-provider")
	mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.KeyPrefix = "test:"
	client, err := manager.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("successful set", func(t *testing.T) {
		mockClient.On("Set", ctx, "test:key1", "value1", time.Minute).Return(nil).Once()

		err := client.Set(ctx, "key1", "value1", time.Minute)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("set with error", func(t *testing.T) {
		mockClient.On("Set", ctx, "test:key2", "value2", time.Hour).Return(errors.New("set error")).Once()

		err := client.Set(ctx, "key2", "value2", time.Hour)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "set error")
	})
}

func TestClient_KeyPrefix(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockClient := &MockClient{}

	mockProvider.On("Name").Return("test-provider")
	mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	t.Run("with key prefix", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.KeyPrefix = "app:prod:"
		client, err := manager.NewClient(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		mockClient.On("Get", ctx, "app:prod:user:123").Return("john", nil).Once()

		result, err := client.Get(ctx, "user:123")

		assert.NoError(t, err)
		assert.Equal(t, "john", result)
	})

	t.Run("without key prefix", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.KeyPrefix = ""
		client, err := manager.NewClient(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		mockClient.On("Get", ctx, "user:123").Return("john", nil).Once()

		result, err := client.Get(ctx, "user:123")

		assert.NoError(t, err)
		assert.Equal(t, "john", result)
	})
}

func TestClient_Scan(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockClient := &MockClient{}

	mockProvider.On("Name").Return("test-provider")
	mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.KeyPrefix = "test:"
	client, err := manager.NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("scan with prefix removal", func(t *testing.T) {
		returnedKeys := []string{"test:key1", "test:key2", "test:key3"}
		mockClient.On("Scan", ctx, uint64(0), "test:*", int64(10)).Return(returnedKeys, uint64(5), nil).Once()

		keys, cursor, err := client.Scan(ctx, 0, "*", 10)

		assert.NoError(t, err)
		assert.Equal(t, uint64(5), cursor)
		assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
	})

	t.Run("scan without prefix", func(t *testing.T) {
		cfg2 := config.DefaultConfig()
		cfg2.KeyPrefix = ""
		client2, err := manager.NewClient(cfg2)
		require.NoError(t, err)

		returnedKeys := []string{"key1", "key2", "key3"}
		mockClient.On("Scan", ctx, uint64(0), "*", int64(10)).Return(returnedKeys, uint64(5), nil).Once()

		keys, cursor, err := client2.Scan(ctx, 0, "*", 10)

		assert.NoError(t, err)
		assert.Equal(t, uint64(5), cursor)
		assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
	})
}

func TestClient_AddHook(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockClient := &MockClient{}

	mockProvider.On("Name").Return("test-provider")
	mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	client, err := manager.NewClient(config.DefaultConfig())
	require.NoError(t, err)

	t.Run("add execution hook", func(t *testing.T) {
		hook := hooks.NewLoggingHook()
		err := client.AddHook(hook)
		assert.NoError(t, err)
	})

	t.Run("add metrics hook", func(t *testing.T) {
		hook := hooks.NewMetricsHook()
		err := client.AddHook(hook)
		assert.NoError(t, err)
	})

	t.Run("add unsupported hook", func(t *testing.T) {
		hook := "invalid hook"
		err := client.AddHook(hook)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tipo de hook não suportado")
	})

	t.Run("add hook to closed client", func(t *testing.T) {
		mockClient.On("Close").Return(nil).Once()
		err := client.Close()
		require.NoError(t, err)

		hook := hooks.NewLoggingHook()
		err = client.AddHook(hook)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cliente está fechado")
	})
}

func TestClient_Close(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockClient := &MockClient{}

	mockProvider.On("Name").Return("test-provider")
	mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	client, err := manager.NewClient(config.DefaultConfig())
	require.NoError(t, err)

	t.Run("close client", func(t *testing.T) {
		mockClient.On("Close").Return(nil).Once()

		err := client.Close()
		assert.NoError(t, err)
		assert.True(t, client.IsClosed())
	})

	t.Run("close already closed client", func(t *testing.T) {
		err := client.Close()
		assert.NoError(t, err)
	})
}

func TestClient_IsHealthy(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}
	mockClient := &MockClient{}

	mockProvider.On("Name").Return("test-provider")
	mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	client, err := manager.NewClient(config.DefaultConfig())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("healthy client", func(t *testing.T) {
		mockClient.On("IsHealthy", ctx).Return(true).Once()

		healthy := client.IsHealthy(ctx)
		assert.True(t, healthy)
	})

	t.Run("unhealthy client", func(t *testing.T) {
		mockClient.On("IsHealthy", ctx).Return(false).Once()

		healthy := client.IsHealthy(ctx)
		assert.False(t, healthy)
	})

	t.Run("closed client is unhealthy", func(t *testing.T) {
		mockClient.On("Close").Return(nil).Once()
		err := client.Close()
		require.NoError(t, err)

		healthy := client.IsHealthy(ctx)
		assert.False(t, healthy)
	})
}

func TestManager_CloseAll(t *testing.T) {
	manager := NewManager()
	mockProvider := &MockProvider{}

	mockProvider.On("Name").Return("test-provider")

	err := manager.RegisterProvider("valkey-go", mockProvider)
	require.NoError(t, err)

	// Criar alguns clientes
	mockClient1 := &MockClient{}
	mockClient2 := &MockClient{}
	mockProvider.On("NewClient", mock.Anything).Return(mockClient1, nil).Once()
	mockProvider.On("NewClient", mock.Anything).Return(mockClient2, nil).Once()

	_, err = manager.GetClient("client1", config.DefaultConfig())
	require.NoError(t, err)

	_, err = manager.GetClient("client2", config.DefaultConfig())
	require.NoError(t, err)

	t.Run("close all clients", func(t *testing.T) {
		mockClient1.On("Close").Return(nil).Once()
		mockClient2.On("Close").Return(nil).Once()

		err := manager.CloseAll()
		assert.NoError(t, err)

		// Verificar que não há clientes no manager
		manager.mu.RLock()
		assert.Empty(t, manager.clients)
		manager.mu.RUnlock()
	})
}

func TestDefaultManager(t *testing.T) {
	t.Run("default manager exists", func(t *testing.T) {
		assert.NotNil(t, DefaultManager)
		assert.IsType(t, &Manager{}, DefaultManager)
	})

	t.Run("NewClient uses default manager", func(t *testing.T) {
		mockProvider := &MockProvider{}
		mockClient := &MockClient{}

		mockProvider.On("Name").Return("test-provider")
		mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

		err := DefaultManager.RegisterProvider("valkey-go", mockProvider)
		require.NoError(t, err)

		client, err := NewClient(config.DefaultConfig())
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("NewClientFromEnv uses default manager", func(t *testing.T) {
		mockProvider := &MockProvider{}
		mockClient := &MockClient{}

		mockProvider.On("Name").Return("test-provider")
		mockProvider.On("NewClient", mock.Anything).Return(mockClient, nil)

		// Já registrado no teste anterior
		client, err := NewClientFromEnv()
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}
