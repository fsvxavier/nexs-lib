// Package interfaces define as interfaces para os tracer providers
package interfaces

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// TracerProvider define a interface comum para todos os tracer providers
type TracerProvider interface {
	// Init inicializa o tracer provider com a configuração fornecida
	Init(ctx context.Context, config Config) (trace.TracerProvider, error)

	// Shutdown finaliza o tracer provider de forma segura
	Shutdown(ctx context.Context) error
}

// Config define a estrutura de configuração para tracer providers
type Config struct {
	// ServiceName é o nome do serviço
	ServiceName string `json:"service_name" yaml:"service_name"`

	// Environment é o ambiente de execução (dev, staging, prod)
	Environment string `json:"environment" yaml:"environment"`

	// ExporterType define qual provider usar (datadog, grafana, newrelic, opentelemetry)
	ExporterType string `json:"exporter_type" yaml:"exporter_type"`

	// Endpoint é o endpoint do trace collector
	Endpoint string `json:"endpoint" yaml:"endpoint"`

	// Headers são cabeçalhos adicionais para envio de traces
	Headers map[string]string `json:"headers" yaml:"headers"`

	// SamplingRatio define a proporção de traces que serão coletados (0.0 a 1.0)
	SamplingRatio float64 `json:"sampling_ratio" yaml:"sampling_ratio"`

	// Propagators define os propagadores a serem utilizados
	Propagators []string `json:"propagators" yaml:"propagators"`

	// APIKey para autenticação (usado por Datadog e New Relic)
	APIKey string `json:"api_key" yaml:"api_key"`

	// LicenseKey para New Relic
	LicenseKey string `json:"license_key" yaml:"license_key"`

	// Insecure define se deve usar conexão insegura
	Insecure bool `json:"insecure" yaml:"insecure"`

	// Version é a versão da aplicação
	Version string `json:"version" yaml:"version"`

	// Attributes são atributos adicionais para os traces
	Attributes map[string]string `json:"attributes" yaml:"attributes"`
}

// TracerProviderFactory define a interface para criação de tracer providers
type TracerProviderFactory interface {
	// CreateProvider cria um novo tracer provider baseado na configuração
	CreateProvider(config Config) (TracerProvider, error)

	// SupportedTypes retorna os tipos de exporters suportados
	SupportedTypes() []string
}

// Instrumenter define a interface para instrumentação
type Instrumenter interface {
	// InstrumentHTTP adiciona instrumentação HTTP
	InstrumentHTTP(provider trace.TracerProvider) error

	// InstrumentGRPC adiciona instrumentação gRPC
	InstrumentGRPC(provider trace.TracerProvider) error

	// InstrumentSQL adiciona instrumentação SQL
	InstrumentSQL(provider trace.TracerProvider) error
}
