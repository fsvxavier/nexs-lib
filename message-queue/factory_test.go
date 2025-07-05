package messagequeue

import (
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFactory(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantNil bool
	}{
		{
			name:    "nil config uses default",
			config:  nil,
			wantNil: false,
		},
		{
			name: "valid config",
			config: &config.Config{
				Global: &config.GlobalConfig{
					DefaultProvider: interfaces.ProviderRabbitMQ,
				},
				Providers: map[interfaces.ProviderType]*config.ProviderConfig{
					interfaces.ProviderRabbitMQ: {
						Enabled: true,
						Connection: &interfaces.ConnectionConfig{
							Brokers: []string{"localhost:5672"},
						},
					},
				},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory(tt.config)

			if tt.wantNil {
				if factory != nil {
					t.Error("NewFactory() should return nil, but got non-nil")
				}
			} else {
				if factory == nil {
					t.Error("NewFactory() should return non-nil, but got nil")
				}
			}
		})
	}
}

func TestFactory_CreateProvider(t *testing.T) {
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider: interfaces.ProviderRabbitMQ,
		},
		Providers: map[interfaces.ProviderType]*config.ProviderConfig{
			interfaces.ProviderRabbitMQ: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"localhost:5672"},
				},
			},
		},
	}

	factory := NewFactory(cfg)
	if factory == nil {
		t.Fatal("Failed to create factory")
	}

	tests := []struct {
		name           string
		providerType   interfaces.ProviderType
		providerConfig *config.ProviderConfig
		wantErr        bool
	}{
		{
			name:         "unsupported provider",
			providerType: interfaces.ProviderType("unknown"),
			providerConfig: &config.ProviderConfig{
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name:         "supported provider",
			providerType: interfaces.ProviderRabbitMQ,
			providerConfig: &config.ProviderConfig{
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"localhost:5672"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(tt.providerType, tt.providerConfig)

			if tt.wantErr {
				if err == nil {
					t.Error("CreateProvider() should return error, but got nil")
				}
				if provider != nil {
					t.Error("CreateProvider() should return nil provider on error")
				}
			} else {
				if err != nil {
					t.Errorf("CreateProvider() should not return error, but got: %v", err)
				}
				if provider == nil {
					t.Error("CreateProvider() should return non-nil provider")
				}
			}
		})
	}
}

func TestFactory_GetProvider(t *testing.T) {
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider: interfaces.ProviderRabbitMQ,
		},
		Providers: map[interfaces.ProviderType]*config.ProviderConfig{
			interfaces.ProviderRabbitMQ: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"localhost:5672"},
				},
			},
		},
	}

	factory := NewFactory(cfg)
	if factory == nil {
		t.Fatal("Failed to create factory")
	}

	tests := []struct {
		name         string
		providerType interfaces.ProviderType
		wantErr      bool
	}{
		{
			name:         "get existing provider",
			providerType: interfaces.ProviderRabbitMQ,
			wantErr:      false,
		},
		{
			name:         "get non-existing provider",
			providerType: interfaces.ProviderType("unknown"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.GetProvider(tt.providerType)

			if tt.wantErr {
				if err == nil {
					t.Error("GetProvider() should return error, but got nil")
				}
				if provider != nil {
					t.Error("GetProvider() should return nil provider on error")
				}
			} else {
				if err != nil {
					t.Errorf("GetProvider() should not return error, but got: %v", err)
				}
				if provider == nil {
					t.Error("GetProvider() should return non-nil provider")
				}
			}
		})
	}
}

func TestFactory_ListProviders(t *testing.T) {
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider: interfaces.ProviderRabbitMQ,
		},
	}

	factory := NewFactory(cfg)
	if factory == nil {
		t.Fatal("Failed to create factory")
	}

	providers := factory.ListProviders()

	// Verificar se todos os providers padrão estão presentes
	expectedProviders := map[interfaces.ProviderType]bool{
		interfaces.ProviderKafka:    true,
		interfaces.ProviderRabbitMQ: true,
		interfaces.ProviderSQS:      true,
		interfaces.ProviderActiveMQ: true,
	}

	if len(providers) != len(expectedProviders) {
		t.Errorf("Expected %d providers, got %d", len(expectedProviders), len(providers))
	}

	for _, provider := range providers {
		if !expectedProviders[provider] {
			t.Errorf("Unexpected provider: %s", provider)
		}
	}
}

func TestFactory_IsProviderAvailable(t *testing.T) {
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider: interfaces.ProviderRabbitMQ,
		},
		Providers: map[interfaces.ProviderType]*config.ProviderConfig{
			interfaces.ProviderRabbitMQ: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"localhost:5672"},
				},
			},
			interfaces.ProviderKafka: {
				Enabled: false,
			},
		},
	}

	factory := NewFactory(cfg)
	if factory == nil {
		t.Fatal("Failed to create factory")
	}

	tests := []struct {
		name         string
		providerType interfaces.ProviderType
		want         bool
	}{
		{
			name:         "enabled provider",
			providerType: interfaces.ProviderRabbitMQ,
			want:         true,
		},
		{
			name:         "disabled provider",
			providerType: interfaces.ProviderKafka,
			want:         false,
		},
		{
			name:         "unconfigured provider",
			providerType: interfaces.ProviderSQS,
			want:         true, // Assume disponível se tem creator
		},
		{
			name:         "unknown provider",
			providerType: interfaces.ProviderType("unknown"),
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := factory.IsProviderAvailable(tt.providerType)
			if got != tt.want {
				t.Errorf("IsProviderAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFactory_Close(t *testing.T) {
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider: interfaces.ProviderRabbitMQ,
		},
	}

	factory := NewFactory(cfg)
	if factory == nil {
		t.Fatal("Failed to create factory")
	}

	// Teste de fechamento da factory
	err := factory.Close()
	if err != nil {
		t.Errorf("Close() should not return error for empty factory, but got: %v", err)
	}
}

// TestFactory_WithMocks testa a factory com os mocks específicos
func TestFactory_WithMocks(t *testing.T) {
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider: interfaces.ProviderKafka,
		},
		Providers: map[interfaces.ProviderType]*config.ProviderConfig{
			interfaces.ProviderKafka: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"localhost:9092"},
				},
			},
			interfaces.ProviderRabbitMQ: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"localhost:5672"},
				},
			},
			interfaces.ProviderSQS: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"https://sqs.us-east-1.amazonaws.com"},
				},
			},
			interfaces.ProviderActiveMQ: {
				Enabled: true,
				Connection: &interfaces.ConnectionConfig{
					Brokers: []string{"tcp://localhost:61616"},
				},
			},
		},
	}

	factory := NewFactory(cfg)
	require.NotNil(t, factory)

	// Test all providers can be created
	providers := []interfaces.ProviderType{
		interfaces.ProviderKafka,
		interfaces.ProviderRabbitMQ,
		interfaces.ProviderSQS,
		interfaces.ProviderActiveMQ,
	}

	for _, providerType := range providers {
		t.Run(fmt.Sprintf("test_%s_provider", providerType), func(t *testing.T) {
			provider, err := factory.GetProvider(providerType)
			require.NoError(t, err)
			require.NotNil(t, provider)

			assert.Equal(t, providerType, provider.GetType())

			// Test provider availability
			available := factory.IsProviderAvailable(providerType)
			assert.True(t, available)
		})
	}

	// Test factory close
	err := factory.Close()
	assert.NoError(t, err)
}

// TestFactory_DefaultProvider testa provider padrão
func TestFactory_DefaultProvider(t *testing.T) {
	cfg := &config.Config{
		Global: &config.GlobalConfig{
			DefaultProvider: interfaces.ProviderRabbitMQ,
		},
	}

	factory := NewFactory(cfg)
	require.NotNil(t, factory)

	// Verificar se todos os providers estão listados
	providers := factory.ListProviders()
	expectedCount := 4 // Kafka, RabbitMQ, SQS, ActiveMQ
	assert.Len(t, providers, expectedCount)

	// Verificar se RabbitMQ está disponível como padrão
	available := factory.IsProviderAvailable(interfaces.ProviderRabbitMQ)
	assert.True(t, available)
}
