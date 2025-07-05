package interfaces

import (
	"context"
	"crypto/tls"
	"time"
)

// ProviderType representa o tipo de provider de mensageria
type ProviderType string

const (
	// ProviderKafka representa o provider Apache Kafka
	ProviderKafka ProviderType = "kafka"

	// ProviderRabbitMQ representa o provider RabbitMQ
	ProviderRabbitMQ ProviderType = "rabbitmq"

	// ProviderSQS representa o provider Amazon SQS
	ProviderSQS ProviderType = "sqs"

	// ProviderActiveMQ representa o provider Apache ActiveMQ
	ProviderActiveMQ ProviderType = "activemq"
)

// MessageQueueProvider define a interface principal para providers de mensageria
type MessageQueueProvider interface {
	// GetType retorna o tipo do provider
	GetType() ProviderType

	// Connect estabelece conexão com o broker
	Connect(ctx context.Context) error

	// Disconnect fecha a conexão com o broker
	Disconnect() error

	// CreateProducer cria um novo producer
	CreateProducer(config *ProducerConfig) (MessageProducer, error)

	// CreateConsumer cria um novo consumer
	CreateConsumer(config *ConsumerConfig) (MessageConsumer, error)

	// IsConnected verifica se está conectado ao broker
	IsConnected() bool

	// HealthCheck verifica a saúde da conexão
	HealthCheck(ctx context.Context) error

	// GetMetrics retorna métricas do provider
	GetMetrics() *ProviderMetrics

	// Close fecha todas as conexões e libera recursos
	Close() error
}

// ConnectionConfig representa configurações de conexão
type ConnectionConfig struct {
	// URLs/Endereços dos brokers
	Brokers []string

	// Configurações de autenticação
	Auth *AuthConfig

	// Configurações TLS/SSL
	TLS *TLSConfig

	// Timeout para conexão
	ConnectTimeout time.Duration

	// Timeout para operações
	OperationTimeout time.Duration

	// Configurações de pool de conexões
	Pool *PoolConfig

	// Configurações específicas do provider
	ProviderConfig map[string]interface{}

	// Configurações de retry para conexão
	RetryConfig *RetryPolicy
}

// AuthConfig representa configurações de autenticação
type AuthConfig struct {
	// Tipo de autenticação (SASL_PLAIN, SASL_SCRAM, etc.)
	Type string

	// Username
	Username string

	// Password
	Password string

	// Token (para OAuth, JWT, etc.)
	Token string

	// Certificado (para autenticação mTLS)
	Certificate []byte

	// Chave privada (para autenticação mTLS)
	PrivateKey []byte

	// Configurações específicas do provider
	Extra map[string]interface{}
}

// TLSConfig representa configurações TLS/SSL
type TLSConfig struct {
	// Se TLS está habilitado
	Enabled bool

	// Configuração TLS do Go
	Config *tls.Config

	// Caminho para o certificado CA
	CAFile string

	// Caminho para o certificado do cliente
	CertFile string

	// Caminho para a chave privada do cliente
	KeyFile string

	// Se deve verificar o certificado do servidor
	InsecureSkipVerify bool

	// Nome do servidor para verificação
	ServerName string
}

// PoolConfig representa configurações de pool de conexões
type PoolConfig struct {
	// Número máximo de conexões
	MaxConnections int

	// Número mínimo de conexões ociosas
	MinIdleConnections int

	// Tempo máximo de vida de uma conexão
	MaxConnectionLifetime time.Duration

	// Tempo máximo que uma conexão pode ficar ociosa
	MaxIdleTime time.Duration

	// Timeout para obter uma conexão do pool
	AcquireTimeout time.Duration
}

// ProducerConfig representa configurações para producers
type ProducerConfig struct {
	// ID único do producer
	ID string

	// Configurações de conexão
	Connection *ConnectionConfig

	// Se deve usar transações
	Transactional bool

	// Timeout para operações de envio
	SendTimeout time.Duration

	// Tamanho do buffer de mensagens
	BufferSize int

	// Configurações de retry
	RetryPolicy *RetryPolicy

	// Configurações de compressão
	Compression string

	// Configurações específicas do provider
	ProviderConfig map[string]interface{}
}

// ConsumerConfig representa configurações para consumers
type ConsumerConfig struct {
	// ID único do consumer
	ID string

	// Configurações de conexão
	Connection *ConnectionConfig

	// Nome do grupo de consumidores (para Kafka)
	ConsumerGroup string

	// Offset inicial (para Kafka)
	InitialOffset string

	// Configurações de commit
	CommitInterval time.Duration

	// Se deve fazer auto commit
	AutoCommit bool

	// Configurações específicas do provider
	ProviderConfig map[string]interface{}
}

// ProviderMetrics representa métricas do provider
type ProviderMetrics struct {
	// Tempo de atividade
	Uptime time.Duration

	// Número de conexões ativas
	ActiveConnections int

	// Número de producers ativos
	ActiveProducers int

	// Número de consumers ativos
	ActiveConsumers int

	// Última verificação de saúde
	LastHealthCheck time.Time

	// Status da última verificação de saúde
	HealthCheckStatus bool

	// Estatísticas de conexão
	ConnectionStats *ConnectionStats
}

// ConnectionStats representa estatísticas de conexão
type ConnectionStats struct {
	// Número total de conexões estabelecidas
	TotalConnections int64

	// Número total de conexões com falha
	FailedConnections int64

	// Tempo médio de conexão
	AvgConnectionTime time.Duration

	// Número de reconexões
	Reconnections int64

	// Última conexão estabelecida
	LastConnectedAt time.Time
}
