package sqs

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
)

// SQSProvider implementa o provider para Amazon SQS
type SQSProvider struct {
	config      *config.ProviderConfig
	client      *sqs.Client
	connected   bool
	metrics     *interfaces.ProviderMetrics
	mutex       sync.RWMutex
	closeOnce   sync.Once
	connectedAt time.Time
}

// NewSQSProvider cria uma nova instância do provider SQS
func NewSQSProvider(cfg *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
	if cfg == nil {
		return nil, domainerrors.New(
			"INVALID_CONFIG",
			"sqs provider config cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if !cfg.Enabled {
		return nil, domainerrors.New(
			"PROVIDER_DISABLED",
			"sqs provider is disabled",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	provider := &SQSProvider{
		config:    cfg,
		connected: false,
		metrics: &interfaces.ProviderMetrics{
			ActiveConnections: 0,
			ActiveProducers:   0,
			ActiveConsumers:   0,
			LastHealthCheck:   time.Now(),
			HealthCheckStatus: false,
			ConnectionStats: &interfaces.ConnectionStats{
				TotalConnections:  0,
				FailedConnections: 0,
				AvgConnectionTime: 0,
				Reconnections:     0,
				LastConnectedAt:   time.Time{},
			},
		},
	}

	return provider, nil
}

// GetType retorna o tipo do provider
func (p *SQSProvider) GetType() interfaces.ProviderType {
	return interfaces.ProviderSQS
}

// Connect estabelece conexão com SQS
func (p *SQSProvider) Connect(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.connected && p.client != nil {
		return nil
	}

	startTime := time.Now()

	// Cria cliente SQS com configuração padrão
	// Em produção, seria necessário configurar credenciais AWS
	cfg := aws.Config{
		Region: "us-east-1", // Padrão
	}

	// Aplica configurações específicas se disponíveis
	if p.config.ProviderSpecific != nil {
		if region, exists := p.config.ProviderSpecific["region"]; exists {
			if regionStr, ok := region.(string); ok {
				cfg.Region = regionStr
			}
		}
	}

	p.client = sqs.NewFromConfig(cfg)
	p.connected = true
	p.connectedAt = time.Now()

	// Atualiza métricas
	connectionTime := time.Since(startTime)
	p.metrics.ConnectionStats.TotalConnections++
	p.metrics.ConnectionStats.LastConnectedAt = p.connectedAt
	p.metrics.ConnectionStats.AvgConnectionTime = connectionTime
	p.metrics.ActiveConnections = 1

	return nil
}

// Disconnect fecha a conexão com SQS
func (p *SQSProvider) Disconnect() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.client = nil
	p.connected = false
	p.metrics.ActiveConnections = 0

	return nil
}

// CreateProducer cria um novo producer SQS
func (p *SQSProvider) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if !p.connected {
		if err := p.Connect(context.Background()); err != nil {
			return nil, err
		}
	}

	producer, err := NewSQSProducer(p, config)
	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	p.metrics.ActiveProducers++
	p.mutex.Unlock()

	return producer, nil
}

// CreateConsumer cria um novo consumer SQS
func (p *SQSProvider) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if !p.connected {
		if err := p.Connect(context.Background()); err != nil {
			return nil, err
		}
	}

	consumer, err := NewSQSConsumer(p, config)
	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	p.metrics.ActiveConsumers++
	p.mutex.Unlock()

	return consumer, nil
}

// IsConnected verifica se está conectado ao SQS
func (p *SQSProvider) IsConnected() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.connected && p.client != nil
}

// HealthCheck verifica a saúde da conexão
func (p *SQSProvider) HealthCheck(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.metrics.LastHealthCheck = time.Now()

	if !p.connected || p.client == nil {
		p.metrics.HealthCheckStatus = false
		return domainerrors.New(
			"SQS_NOT_CONNECTED",
			"sqs client is not available",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	// Tenta listar filas como teste de conectividade
	_, err := p.client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		p.metrics.HealthCheckStatus = false
		return domainerrors.New(
			"SQS_HEALTH_CHECK_FAILED",
			"failed to list queues for health check",
		).WithType(domainerrors.ErrorTypeRepository).
			Wrap("health check error", err)
	}

	p.metrics.HealthCheckStatus = true
	return nil
}

// GetMetrics retorna métricas do provider
func (p *SQSProvider) GetMetrics() *interfaces.ProviderMetrics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Calcula o uptime
	if p.connected {
		p.metrics.Uptime = time.Since(p.connectedAt)
	}

	// Cria uma cópia das métricas para evitar race conditions
	return &interfaces.ProviderMetrics{
		Uptime:            p.metrics.Uptime,
		ActiveConnections: p.metrics.ActiveConnections,
		ActiveProducers:   p.metrics.ActiveProducers,
		ActiveConsumers:   p.metrics.ActiveConsumers,
		LastHealthCheck:   p.metrics.LastHealthCheck,
		HealthCheckStatus: p.metrics.HealthCheckStatus,
		ConnectionStats: &interfaces.ConnectionStats{
			TotalConnections:  p.metrics.ConnectionStats.TotalConnections,
			FailedConnections: p.metrics.ConnectionStats.FailedConnections,
			AvgConnectionTime: p.metrics.ConnectionStats.AvgConnectionTime,
			Reconnections:     p.metrics.ConnectionStats.Reconnections,
			LastConnectedAt:   p.metrics.ConnectionStats.LastConnectedAt,
		},
	}
}

// Close fecha todas as conexões e libera recursos
func (p *SQSProvider) Close() error {
	var err error
	p.closeOnce.Do(func() {
		err = p.Disconnect()
	})
	return err
}

// GetClient retorna o cliente SQS (para uso interno)
func (p *SQSProvider) GetClient() *sqs.Client {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.client
}

// Função stub para producers/consumers que serão implementados
func NewSQSProducer(provider *SQSProvider, config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	return nil, domainerrors.New(
		"NOT_IMPLEMENTED",
		"SQS producer not yet implemented",
	).WithType(domainerrors.ErrorTypeRepository)
}

func NewSQSConsumer(provider *SQSProvider, config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	return nil, domainerrors.New(
		"NOT_IMPLEMENTED",
		"SQS consumer not yet implemented",
	).WithType(domainerrors.ErrorTypeRepository)
}
