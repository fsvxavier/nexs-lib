package config

import (
	"time"

	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// Config representa a configuração completa do sistema de message queue
type Config struct {
	// Configurações globais
	Global *GlobalConfig `json:"global" yaml:"global"`

	// Configurações por provider
	Providers map[interfaces.ProviderType]*ProviderConfig `json:"providers" yaml:"providers"`

	// Configurações de observabilidade
	Observability *ObservabilityConfig `json:"observability" yaml:"observability"`

	// Configurações de idempotência
	Idempotency *IdempotencyConfig `json:"idempotency" yaml:"idempotency"`
}

// GlobalConfig representa configurações globais
type GlobalConfig struct {
	// Provider padrão a ser usado
	DefaultProvider interfaces.ProviderType `json:"default_provider" yaml:"default_provider"`

	// Timeout padrão para operações
	DefaultTimeout time.Duration `json:"default_timeout" yaml:"default_timeout"`

	// Número padrão de workers
	DefaultWorkers int `json:"default_workers" yaml:"default_workers"`

	// Tamanho padrão do buffer
	DefaultBufferSize int `json:"default_buffer_size" yaml:"default_buffer_size"`

	// Se deve habilitar métricas por padrão
	MetricsEnabled bool `json:"metrics_enabled" yaml:"metrics_enabled"`

	// Se deve habilitar tracing por padrão
	TracingEnabled bool `json:"tracing_enabled" yaml:"tracing_enabled"`

	// Configurações de health check
	HealthCheck *HealthCheckConfig `json:"health_check" yaml:"health_check"`
}

// ProviderConfig representa configurações específicas de um provider
type ProviderConfig struct {
	// Se o provider está habilitado
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Configurações de conexão
	Connection *interfaces.ConnectionConfig `json:"connection" yaml:"connection"`

	// Configurações padrão para producers
	DefaultProducer *interfaces.ProducerConfig `json:"default_producer" yaml:"default_producer"`

	// Configurações padrão para consumers
	DefaultConsumer *interfaces.ConsumerConfig `json:"default_consumer" yaml:"default_consumer"`

	// Configurações específicas do provider
	ProviderSpecific map[string]interface{} `json:"provider_specific" yaml:"provider_specific"`
}

// ObservabilityConfig representa configurações de observabilidade
type ObservabilityConfig struct {
	// Se logging está habilitado
	LoggingEnabled bool `json:"logging_enabled" yaml:"logging_enabled"`

	// Nível de log
	LogLevel string `json:"log_level" yaml:"log_level"`

	// Se tracing está habilitado
	TracingEnabled bool `json:"tracing_enabled" yaml:"tracing_enabled"`

	// Service name para tracing
	ServiceName string `json:"service_name" yaml:"service_name"`

	// Se métricas estão habilitadas
	MetricsEnabled bool `json:"metrics_enabled" yaml:"metrics_enabled"`

	// Configurações de exportação de métricas
	MetricsExporter *MetricsExporterConfig `json:"metrics_exporter" yaml:"metrics_exporter"`

	// Configurações de sampling para traces
	TracingSamplingRate float64 `json:"tracing_sampling_rate" yaml:"tracing_sampling_rate"`
}

// MetricsExporterConfig representa configurações de exportação de métricas
type MetricsExporterConfig struct {
	// Tipo de exportador (prometheus, datadog, etc.)
	Type string `json:"type" yaml:"type"`

	// Endpoint para exportação
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// Intervalo de exportação
	Interval time.Duration `json:"interval" yaml:"interval"`

	// Labels adicionais
	Labels map[string]string `json:"labels" yaml:"labels"`
}

// IdempotencyConfig representa configurações de idempotência
type IdempotencyConfig struct {
	// Se idempotência está habilitada
	Enabled bool `json:"enabled" yaml:"enabled"`

	// TTL padrão para cache de idempotência
	CacheTTL time.Duration `json:"cache_ttl" yaml:"cache_ttl"`

	// Tamanho máximo do cache
	CacheSize int `json:"cache_size" yaml:"cache_size"`

	// Tipo de storage para idempotência (memory, redis, etc.)
	StorageType string `json:"storage_type" yaml:"storage_type"`

	// Configurações específicas do storage
	StorageConfig map[string]interface{} `json:"storage_config" yaml:"storage_config"`
}

// HealthCheckConfig representa configurações de health check
type HealthCheckConfig struct {
	// Se health check está habilitado
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Intervalo entre verificações
	Interval time.Duration `json:"interval" yaml:"interval"`

	// Timeout para verificações
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// Número de falhas consecutivas antes de considerar unhealthy
	FailureThreshold int `json:"failure_threshold" yaml:"failure_threshold"`

	// Número de sucessos consecutivos antes de considerar healthy novamente
	SuccessThreshold int `json:"success_threshold" yaml:"success_threshold"`
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() *Config {
	return &Config{
		Global: &GlobalConfig{
			DefaultProvider:   interfaces.ProviderKafka,
			DefaultTimeout:    30 * time.Second,
			DefaultWorkers:    10,
			DefaultBufferSize: 1000,
			MetricsEnabled:    true,
			TracingEnabled:    true,
			HealthCheck: &HealthCheckConfig{
				Enabled:          true,
				Interval:         30 * time.Second,
				Timeout:          5 * time.Second,
				FailureThreshold: 3,
				SuccessThreshold: 2,
			},
		},
		Providers: make(map[interfaces.ProviderType]*ProviderConfig),
		Observability: &ObservabilityConfig{
			LoggingEnabled:      true,
			LogLevel:            "info",
			TracingEnabled:      true,
			ServiceName:         "message-queue",
			MetricsEnabled:      true,
			TracingSamplingRate: 0.1,
			MetricsExporter: &MetricsExporterConfig{
				Type:     "prometheus",
				Interval: 15 * time.Second,
				Labels:   make(map[string]string),
			},
		},
		Idempotency: &IdempotencyConfig{
			Enabled:       true,
			CacheTTL:      1 * time.Hour,
			CacheSize:     10000,
			StorageType:   "memory",
			StorageConfig: make(map[string]interface{}),
		},
	}
}

// KafkaConfig retorna uma configuração padrão para Kafka
func KafkaConfig() *ProviderConfig {
	return &ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers:          []string{"localhost:9092"},
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 30 * time.Second,
			Pool: &interfaces.PoolConfig{
				MaxConnections:        10,
				MinIdleConnections:    2,
				MaxConnectionLifetime: 1 * time.Hour,
				MaxIdleTime:           30 * time.Minute,
				AcquireTimeout:        5 * time.Second,
			},
			RetryConfig: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      1 * time.Second,
				BackoffMultiplier: 2.0,
				MaxDelay:          30 * time.Second,
				Jitter:            true,
			},
		},
		DefaultProducer: &interfaces.ProducerConfig{
			SendTimeout:   10 * time.Second,
			BufferSize:    1000,
			Transactional: false,
			Compression:   "snappy",
			RetryPolicy: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      100 * time.Millisecond,
				BackoffMultiplier: 2.0,
				MaxDelay:          5 * time.Second,
				Jitter:            true,
			},
		},
		DefaultConsumer: &interfaces.ConsumerConfig{
			ConsumerGroup:  "default-group",
			InitialOffset:  "latest",
			CommitInterval: 1 * time.Second,
			AutoCommit:     true,
		},
		ProviderSpecific: map[string]interface{}{
			"version":             "2.8.0",
			"requiredAcks":        1,
			"maxMessageBytes":     1000000,
			"compressionLevel":    6,
			"enableIdempotence":   true,
			"maxInFlightRequests": 5,
		},
	}
}

// RabbitMQConfig retorna uma configuração padrão para RabbitMQ
func RabbitMQConfig() *ProviderConfig {
	return &ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers:          []string{"amqp://localhost:5672"},
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 30 * time.Second,
			Pool: &interfaces.PoolConfig{
				MaxConnections:        5,
				MinIdleConnections:    1,
				MaxConnectionLifetime: 1 * time.Hour,
				MaxIdleTime:           30 * time.Minute,
				AcquireTimeout:        5 * time.Second,
			},
			RetryConfig: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      1 * time.Second,
				BackoffMultiplier: 2.0,
				MaxDelay:          30 * time.Second,
				Jitter:            true,
			},
		},
		DefaultProducer: &interfaces.ProducerConfig{
			SendTimeout:   10 * time.Second,
			BufferSize:    500,
			Transactional: false,
			RetryPolicy: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      100 * time.Millisecond,
				BackoffMultiplier: 2.0,
				MaxDelay:          5 * time.Second,
				Jitter:            true,
			},
		},
		DefaultConsumer: &interfaces.ConsumerConfig{
			CommitInterval: 1 * time.Second,
			AutoCommit:     false,
		},
		ProviderSpecific: map[string]interface{}{
			"exchangeType":     "topic",
			"durable":          true,
			"autoDelete":       false,
			"exclusive":        false,
			"noWait":           false,
			"prefetchCount":    10,
			"prefetchSize":     0,
			"confirmMode":      true,
			"publisherConfirm": true,
		},
	}
}

// SQSConfig retorna uma configuração padrão para Amazon SQS
func SQSConfig() *ProviderConfig {
	return &ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 30 * time.Second,
			RetryConfig: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      1 * time.Second,
				BackoffMultiplier: 2.0,
				MaxDelay:          30 * time.Second,
				Jitter:            true,
			},
		},
		DefaultProducer: &interfaces.ProducerConfig{
			SendTimeout: 10 * time.Second,
			BufferSize:  100,
			RetryPolicy: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      100 * time.Millisecond,
				BackoffMultiplier: 2.0,
				MaxDelay:          5 * time.Second,
				Jitter:            true,
			},
		},
		DefaultConsumer: &interfaces.ConsumerConfig{
			CommitInterval: 5 * time.Second,
			AutoCommit:     true,
		},
		ProviderSpecific: map[string]interface{}{
			"region":                 "us-east-1",
			"maxReceiveCount":        3,
			"visibilityTimeout":      30,
			"waitTimeSeconds":        20,
			"maxMessages":            10,
			"messageRetentionPeriod": 1209600, // 14 dias
			"delaySeconds":           0,
		},
	}
}

// ActiveMQConfig retorna uma configuração padrão para Apache ActiveMQ
func ActiveMQConfig() *ProviderConfig {
	return &ProviderConfig{
		Enabled: true,
		Connection: &interfaces.ConnectionConfig{
			Brokers:          []string{"tcp://localhost:61616"},
			ConnectTimeout:   10 * time.Second,
			OperationTimeout: 30 * time.Second,
			Pool: &interfaces.PoolConfig{
				MaxConnections:        5,
				MinIdleConnections:    1,
				MaxConnectionLifetime: 1 * time.Hour,
				MaxIdleTime:           30 * time.Minute,
				AcquireTimeout:        5 * time.Second,
			},
			RetryConfig: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      1 * time.Second,
				BackoffMultiplier: 2.0,
				MaxDelay:          30 * time.Second,
				Jitter:            true,
			},
		},
		DefaultProducer: &interfaces.ProducerConfig{
			SendTimeout:   10 * time.Second,
			BufferSize:    500,
			Transactional: false,
			RetryPolicy: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      100 * time.Millisecond,
				BackoffMultiplier: 2.0,
				MaxDelay:          5 * time.Second,
				Jitter:            true,
			},
		},
		DefaultConsumer: &interfaces.ConsumerConfig{
			CommitInterval: 1 * time.Second,
			AutoCommit:     false,
		},
		ProviderSpecific: map[string]interface{}{
			"protocol":         "openwire",
			"persistent":       true,
			"ackMode":          "auto",
			"prefetchSize":     1000,
			"sessionCacheSize": 1,
			"useCompression":   true,
		},
	}
}
