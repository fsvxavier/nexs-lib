package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
)

// KafkaProvider implementa o provider para Apache Kafka
type KafkaProvider struct {
	config      *config.ProviderConfig
	client      sarama.Client
	adminClient sarama.ClusterAdmin
	connected   bool
	metrics     *interfaces.ProviderMetrics
	mutex       sync.RWMutex
	closeOnce   sync.Once
	connectedAt time.Time
}

// NewKafkaProvider cria uma nova instância do provider Kafka
func NewKafkaProvider(cfg *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
	if cfg == nil {
		return nil, domainerrors.New(
			"INVALID_CONFIG",
			"kafka provider config cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if !cfg.Enabled {
		return nil, domainerrors.New(
			"PROVIDER_DISABLED",
			"kafka provider is disabled",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	provider := &KafkaProvider{
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
func (p *KafkaProvider) GetType() interfaces.ProviderType {
	return interfaces.ProviderKafka
}

// Connect estabelece conexão com o cluster Kafka
func (p *KafkaProvider) Connect(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.connected && p.client != nil {
		return nil
	}

	startTime := time.Now()

	// Configuração do Sarama
	saramaConfig := sarama.NewConfig()

	// Configurações básicas
	saramaConfig.Version = sarama.V2_8_0_0 // Versão padrão, pode ser configurada
	saramaConfig.ClientID = fmt.Sprintf("nexs-kafka-client-%d", time.Now().UnixNano())

	// Configurações de rede
	if p.config.Connection.ConnectTimeout > 0 {
		saramaConfig.Net.DialTimeout = p.config.Connection.ConnectTimeout
	}
	if p.config.Connection.OperationTimeout > 0 {
		saramaConfig.Net.ReadTimeout = p.config.Connection.OperationTimeout
		saramaConfig.Net.WriteTimeout = p.config.Connection.OperationTimeout
	}

	// Configurações TLS
	if p.config.Connection.TLS != nil && p.config.Connection.TLS.Enabled {
		saramaConfig.Net.TLS.Enable = true
		if p.config.Connection.TLS.Config != nil {
			saramaConfig.Net.TLS.Config = p.config.Connection.TLS.Config
		}
	}

	// Configurações de autenticação SASL
	if p.config.Connection.Auth != nil {
		if err := p.configureSASL(saramaConfig); err != nil {
			p.metrics.ConnectionStats.FailedConnections++
			return domainerrors.New(
				"SASL_CONFIG_ERROR",
				"failed to configure SASL authentication",
			).WithType(domainerrors.ErrorTypeRepository).
				Wrap("sasl configuration error", err)
		}
	}

	// Configurações específicas do provider se existirem
	if p.config.ProviderSpecific != nil {
		if err := p.applyProviderSpecificConfig(saramaConfig); err != nil {
			return err
		}
	}

	// Cria o cliente Kafka
	client, err := sarama.NewClient(p.config.Connection.Brokers, saramaConfig)
	if err != nil {
		p.metrics.ConnectionStats.FailedConnections++
		return domainerrors.New(
			"KAFKA_CONNECTION_FAILED",
			"failed to connect to Kafka cluster",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("brokers", p.config.Connection.Brokers).
			Wrap("kafka connection error", err)
	}

	// Cria o admin client
	adminClient, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		client.Close()
		p.metrics.ConnectionStats.FailedConnections++
		return domainerrors.New(
			"KAFKA_ADMIN_CLIENT_FAILED",
			"failed to create Kafka admin client",
		).WithType(domainerrors.ErrorTypeRepository).
			Wrap("admin client creation error", err)
	}

	p.client = client
	p.adminClient = adminClient
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

// Disconnect fecha a conexão com o cluster Kafka
func (p *KafkaProvider) Disconnect() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected {
		return nil
	}

	var errs []error

	if p.adminClient != nil {
		if err := p.adminClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close admin client: %w", err))
		}
		p.adminClient = nil
	}

	if p.client != nil {
		if err := p.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close client: %w", err))
		}
		p.client = nil
	}

	p.connected = false
	p.metrics.ActiveConnections = 0

	if len(errs) > 0 {
		return domainerrors.New(
			"KAFKA_DISCONNECT_ERROR",
			"errors occurred during disconnect",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("errors", errs)
	}

	return nil
}

// CreateProducer cria um novo producer Kafka
func (p *KafkaProvider) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if !p.connected {
		if err := p.Connect(context.Background()); err != nil {
			return nil, err
		}
	}

	producer, err := NewKafkaProducer(p, config)
	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	p.metrics.ActiveProducers++
	p.mutex.Unlock()

	return producer, nil
}

// CreateConsumer cria um novo consumer Kafka
func (p *KafkaProvider) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if !p.connected {
		if err := p.Connect(context.Background()); err != nil {
			return nil, err
		}
	}

	consumer, err := NewKafkaConsumer(p, config)
	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	p.metrics.ActiveConsumers++
	p.mutex.Unlock()

	return consumer, nil
}

// IsConnected verifica se está conectado ao cluster Kafka
func (p *KafkaProvider) IsConnected() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.connected || p.client == nil {
		return false
	}

	// Verifica se o cliente ainda está válido
	brokers := p.client.Brokers()
	for _, broker := range brokers {
		connected, _ := broker.Connected()
		if connected {
			return true
		}
	}

	return false
}

// HealthCheck verifica a saúde da conexão
func (p *KafkaProvider) HealthCheck(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.metrics.LastHealthCheck = time.Now()

	if !p.connected || p.client == nil {
		p.metrics.HealthCheckStatus = false
		return domainerrors.New(
			"KAFKA_NOT_CONNECTED",
			"kafka client is not connected",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	// Tenta obter configuração como teste de conectividade
	config := p.client.Config()
	if config == nil {
		p.metrics.HealthCheckStatus = false
		return domainerrors.New(
			"KAFKA_HEALTH_CHECK_FAILED",
			"failed to get config from Kafka client",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	p.metrics.HealthCheckStatus = true
	return nil
}

// GetMetrics retorna métricas do provider
func (p *KafkaProvider) GetMetrics() *interfaces.ProviderMetrics {
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
func (p *KafkaProvider) Close() error {
	var err error
	p.closeOnce.Do(func() {
		err = p.Disconnect()
	})
	return err
}

// GetClient retorna o cliente Sarama (para uso interno)
func (p *KafkaProvider) GetClient() sarama.Client {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.client
}

// GetAdminClient retorna o admin client (para uso interno)
func (p *KafkaProvider) GetAdminClient() sarama.ClusterAdmin {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.adminClient
}

// configureSASL configura autenticação SASL
func (p *KafkaProvider) configureSASL(config *sarama.Config) error {
	auth := p.config.Connection.Auth

	switch auth.Type {
	case "SASL_PLAINTEXT", "PLAIN":
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.User = auth.Username
		config.Net.SASL.Password = auth.Password

	case "SASL_SCRAM_SHA256":
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		config.Net.SASL.User = auth.Username
		config.Net.SASL.Password = auth.Password

	case "SASL_SCRAM_SHA512":
		config.Net.SASL.Enable = true
		config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		config.Net.SASL.User = auth.Username
		config.Net.SASL.Password = auth.Password

	default:
		return domainerrors.New(
			"UNSUPPORTED_SASL_TYPE",
			fmt.Sprintf("unsupported SASL type: %s", auth.Type),
		).WithType(domainerrors.ErrorTypeValidation).
			WithDetail("sasl_type", auth.Type)
	}

	return nil
}

// applyProviderSpecificConfig aplica configurações específicas do Kafka
func (p *KafkaProvider) applyProviderSpecificConfig(config *sarama.Config) error {
	specific := p.config.ProviderSpecific

	// Configura compression
	if compression, exists := specific["compression"]; exists {
		if compressionStr, ok := compression.(string); ok {
			switch compressionStr {
			case "gzip":
				config.Producer.Compression = sarama.CompressionGZIP
			case "snappy":
				config.Producer.Compression = sarama.CompressionSnappy
			case "lz4":
				config.Producer.Compression = sarama.CompressionLZ4
			case "zstd":
				config.Producer.Compression = sarama.CompressionZSTD
			default:
				config.Producer.Compression = sarama.CompressionNone
			}
		}
	}

	// Configura required acks
	if requiredAcks, exists := specific["requiredAcks"]; exists {
		if acks, ok := requiredAcks.(int); ok {
			config.Producer.RequiredAcks = sarama.RequiredAcks(acks)
		}
	}

	// Configura idempotence
	if enableIdempotence, exists := specific["enableIdempotence"]; exists {
		if enabled, ok := enableIdempotence.(bool); ok {
			config.Producer.Idempotent = enabled
			if enabled {
				config.Producer.RequiredAcks = sarama.WaitForAll
				config.Net.MaxOpenRequests = 1
			}
		}
	}

	// Configura max message bytes
	if maxMessageBytes, exists := specific["maxMessageBytes"]; exists {
		if maxBytes, ok := maxMessageBytes.(int); ok {
			config.Producer.MaxMessageBytes = maxBytes
		}
	}

	return nil
}
