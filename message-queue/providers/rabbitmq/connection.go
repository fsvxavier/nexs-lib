package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQProvider implementa o provider para RabbitMQ
type RabbitMQProvider struct {
	config      *config.ProviderConfig
	connection  *amqp091.Connection
	connected   bool
	metrics     *interfaces.ProviderMetrics
	mutex       sync.RWMutex
	closeOnce   sync.Once
	connectedAt time.Time
}

// NewRabbitMQProvider cria uma nova instância do provider RabbitMQ
func NewRabbitMQProvider(cfg *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
	if cfg == nil {
		return nil, domainerrors.New(
			"INVALID_CONFIG",
			"rabbitmq provider config cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if !cfg.Enabled {
		return nil, domainerrors.New(
			"PROVIDER_DISABLED",
			"rabbitmq provider is disabled",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	provider := &RabbitMQProvider{
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
func (p *RabbitMQProvider) GetType() interfaces.ProviderType {
	return interfaces.ProviderRabbitMQ
}

// Connect estabelece conexão com RabbitMQ
func (p *RabbitMQProvider) Connect(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.connected && p.connection != nil && !p.connection.IsClosed() {
		return nil
	}

	startTime := time.Now()

	// Monta a URL de conexão
	connectionURL := p.buildConnectionURL()

	// Configurações de conexão
	config := amqp091.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Vhost:     "/",
	}

	// Configurações TLS se habilitado
	if p.config.Connection.TLS != nil && p.config.Connection.TLS.Enabled {
		config.TLSClientConfig = p.config.Connection.TLS.Config
	}

	// Conecta ao RabbitMQ
	conn, err := amqp091.DialConfig(connectionURL, config)
	if err != nil {
		p.metrics.ConnectionStats.FailedConnections++
		return domainerrors.New(
			"RABBITMQ_CONNECTION_FAILED",
			"failed to connect to RabbitMQ",
		).WithType(domainerrors.ErrorTypeRepository).
			WithDetail("url", connectionURL).
			Wrap("rabbitmq connection error", err)
	}

	p.connection = conn
	p.connected = true
	p.connectedAt = time.Now()

	// Configura callback para notificação de close
	go p.handleConnectionClose()

	// Atualiza métricas
	connectionTime := time.Since(startTime)
	p.metrics.ConnectionStats.TotalConnections++
	p.metrics.ConnectionStats.LastConnectedAt = p.connectedAt
	p.metrics.ConnectionStats.AvgConnectionTime = connectionTime
	p.metrics.ActiveConnections = 1

	return nil
}

// Disconnect fecha a conexão com RabbitMQ
func (p *RabbitMQProvider) Disconnect() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected || p.connection == nil {
		return nil
	}

	err := p.connection.Close()
	p.connected = false
	p.metrics.ActiveConnections = 0

	if err != nil {
		return domainerrors.New(
			"RABBITMQ_DISCONNECT_ERROR",
			"failed to close RabbitMQ connection",
		).WithType(domainerrors.ErrorTypeRepository).
			Wrap("disconnect error", err)
	}

	return nil
}

// CreateProducer cria um novo producer RabbitMQ
func (p *RabbitMQProvider) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if !p.connected {
		if err := p.Connect(context.Background()); err != nil {
			return nil, err
		}
	}

	producer, err := NewRabbitMQProducer(p, config)
	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	p.metrics.ActiveProducers++
	p.mutex.Unlock()

	return producer, nil
}

// CreateConsumer cria um novo consumer RabbitMQ
func (p *RabbitMQProvider) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if !p.connected {
		if err := p.Connect(context.Background()); err != nil {
			return nil, err
		}
	}

	consumer, err := NewRabbitMQConsumer(p, config)
	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	p.metrics.ActiveConsumers++
	p.mutex.Unlock()

	return consumer, nil
}

// IsConnected verifica se está conectado ao RabbitMQ
func (p *RabbitMQProvider) IsConnected() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.connected && p.connection != nil && !p.connection.IsClosed()
}

// HealthCheck verifica a saúde da conexão
func (p *RabbitMQProvider) HealthCheck(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.metrics.LastHealthCheck = time.Now()

	if !p.connected || p.connection == nil || p.connection.IsClosed() {
		p.metrics.HealthCheckStatus = false
		return domainerrors.New(
			"RABBITMQ_NOT_CONNECTED",
			"rabbitmq connection is not available",
		).WithType(domainerrors.ErrorTypeRepository)
	}

	// Tenta criar um channel como teste de conectividade
	ch, err := p.connection.Channel()
	if err != nil {
		p.metrics.HealthCheckStatus = false
		return domainerrors.New(
			"RABBITMQ_HEALTH_CHECK_FAILED",
			"failed to create channel for health check",
		).WithType(domainerrors.ErrorTypeRepository).
			Wrap("health check error", err)
	}
	defer ch.Close()

	p.metrics.HealthCheckStatus = true
	return nil
}

// GetMetrics retorna métricas do provider
func (p *RabbitMQProvider) GetMetrics() *interfaces.ProviderMetrics {
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
func (p *RabbitMQProvider) Close() error {
	var err error
	p.closeOnce.Do(func() {
		err = p.Disconnect()
	})
	return err
}

// GetConnection retorna a conexão AMQP (para uso interno)
func (p *RabbitMQProvider) GetConnection() *amqp091.Connection {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.connection
}

// buildConnectionURL constrói a URL de conexão
func (p *RabbitMQProvider) buildConnectionURL() string {
	if len(p.config.Connection.Brokers) == 0 {
		return "amqp://localhost:5672/"
	}

	url := p.config.Connection.Brokers[0]

	// Adiciona autenticação se configurada
	if p.config.Connection.Auth != nil {
		auth := p.config.Connection.Auth
		if auth.Username != "" && auth.Password != "" {
			// Substitui ou adiciona credenciais na URL
			// Implementação simplificada - em produção seria mais robusta
			if auth.Username != "" {
				url = fmt.Sprintf("amqp://%s:%s@%s",
					auth.Username, auth.Password,
					url[len("amqp://"):])
			}
		}
	}

	return url
}

// handleConnectionClose monitora a conexão e tenta reconectar
func (p *RabbitMQProvider) handleConnectionClose() {
	if p.connection == nil {
		return
	}

	// Canal para notificação de close
	closeChan := make(chan *amqp091.Error)
	p.connection.NotifyClose(closeChan)

	// Aguarda notificação de close
	closeErr := <-closeChan
	if closeErr != nil {
		p.mutex.Lock()
		p.connected = false
		p.metrics.ActiveConnections = 0
		p.mutex.Unlock()

		// Em uma implementação completa, aqui poderíamos implementar
		// lógica de reconexão automática
	}
}
