package activemq

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/config"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/go-stomp/stomp/v3"
)

// ActiveMQProvider implementa o provider para Apache ActiveMQ
type ActiveMQProvider struct {
	config      *config.ProviderConfig
	conn        *stomp.Conn
	connected   bool
	metrics     *interfaces.ProviderMetrics
	mutex       sync.RWMutex
	closeOnce   sync.Once
	connectedAt time.Time
}

// NewActiveMQProvider cria uma nova instância do provider ActiveMQ
func NewActiveMQProvider(cfg *config.ProviderConfig) (interfaces.MessageQueueProvider, error) {
	if cfg == nil {
		return nil, domainerrors.New(
			"INVALID_CONFIG",
			"activemq provider config cannot be nil",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	if !cfg.Enabled {
		return nil, domainerrors.New(
			"PROVIDER_DISABLED",
			"activemq provider is disabled",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	provider := &ActiveMQProvider{
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
			},
		},
	}

	return provider, nil
}

// Connect estabelece conexão com ActiveMQ
func (p *ActiveMQProvider) Connect(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.connected && p.conn != nil {
		return nil // Já conectado
	}

	// Configurações de conexão
	host := "localhost:61613" // Porta padrão STOMP do ActiveMQ
	if p.config.Connection != nil && len(p.config.Connection.Brokers) > 0 {
		host = p.config.Connection.Brokers[0]
	}

	// Conecta via STOMP
	netConn, err := net.Dial("tcp", host)
	if err != nil {
		p.metrics.ConnectionStats.FailedConnections++
		return domainerrors.New(
			"ACTIVEMQ_CONNECTION_ERROR",
			"failed to connect to ActiveMQ",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	// Opções de conexão STOMP
	var options []func(*stomp.Conn) error = []func(*stomp.Conn) error{
		stomp.ConnOpt.HeartBeat(10*time.Second, 10*time.Second),
	}

	// Adiciona autenticação se configurada
	if p.config.Connection != nil && p.config.Connection.Auth != nil {
		options = append(options,
			stomp.ConnOpt.Login(p.config.Connection.Auth.Username, p.config.Connection.Auth.Password),
		)
	}

	conn, err := stomp.Connect(netConn, options...)
	if err != nil {
		netConn.Close()
		p.metrics.ConnectionStats.FailedConnections++
		return domainerrors.New(
			"ACTIVEMQ_STOMP_ERROR",
			"failed to establish STOMP connection",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	p.conn = conn
	p.connected = true
	p.connectedAt = time.Now()
	p.metrics.ActiveConnections = 1
	p.metrics.ConnectionStats.TotalConnections++

	return nil
}

// GetType retorna o tipo do provider
func (p *ActiveMQProvider) GetType() interfaces.ProviderType {
	return interfaces.ProviderActiveMQ
}

// Disconnect fecha a conexão com ActiveMQ
func (p *ActiveMQProvider) Disconnect() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected || p.conn == nil {
		return nil
	}

	err := p.conn.Disconnect()
	p.conn = nil
	p.connected = false
	p.metrics.ActiveConnections = 0

	return err
}

// IsConnected verifica se está conectado
func (p *ActiveMQProvider) IsConnected() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.connected && p.conn != nil
}

// Ping verifica a conectividade
func (p *ActiveMQProvider) Ping(ctx context.Context) error {
	if !p.IsConnected() {
		return domainerrors.New(
			"ACTIVEMQ_NOT_CONNECTED",
			"not connected to ActiveMQ",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	return nil
}

// HealthCheck verifica a saúde da conexão
func (p *ActiveMQProvider) HealthCheck(ctx context.Context) error {
	p.metrics.LastHealthCheck = time.Now()

	if err := p.Ping(ctx); err != nil {
		p.metrics.HealthCheckStatus = false
		return err
	}

	p.metrics.HealthCheckStatus = true
	return nil
}

// GetMetrics retorna métricas do provider
func (p *ActiveMQProvider) GetMetrics() *interfaces.ProviderMetrics {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Cria uma cópia das métricas para evitar problemas de concorrência
	metricsCopy := *p.metrics
	return &metricsCopy
}

// Close fecha o provider
func (p *ActiveMQProvider) Close() error {
	var err error
	p.closeOnce.Do(func() {
		err = p.Disconnect()
	})
	return err
}

// CreateProducer cria um novo producer
func (p *ActiveMQProvider) CreateProducer(config *interfaces.ProducerConfig) (interfaces.MessageProducer, error) {
	if !p.IsConnected() {
		return nil, domainerrors.New(
			"ACTIVEMQ_NOT_CONNECTED",
			"provider not connected",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	destination, ok := config.ProviderConfig["destination"].(string)
	if !ok || destination == "" {
		return nil, domainerrors.New(
			"DESTINATION_REQUIRED",
			"destination is required in ProviderConfig",
		).WithType(domainerrors.ErrorTypeValidation)
	}

	producer, err := NewProducer(p, destination)
	if err != nil {
		return nil, err
	}

	p.mutex.Lock()
	p.metrics.ActiveProducers++
	p.mutex.Unlock()

	return producer, nil
}

// CreateConsumer cria um novo consumer
func (p *ActiveMQProvider) CreateConsumer(config *interfaces.ConsumerConfig) (interfaces.MessageConsumer, error) {
	if !p.IsConnected() {
		return nil, domainerrors.New(
			"ACTIVEMQ_NOT_CONNECTED",
			"provider not connected",
		).WithType(domainerrors.ErrorTypeInternal)
	}

	// Implementação stub - seria necessário criar um consumer que implementa MessageConsumer
	// Por enquanto retorna erro
	return nil, domainerrors.New(
		"NOT_IMPLEMENTED",
		"ActiveMQ consumer not yet implemented for this interface",
	).WithType(domainerrors.ErrorTypeInternal)
}
