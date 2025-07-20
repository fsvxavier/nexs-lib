// Package tracer fornece uma implementação unificada de tracing distribuído
// com suporte a múltiplos providers (Datadog, Grafana, New Relic, OpenTelemetry)
package tracer

import (
	"context"
	"fmt"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/fsvxavier/nexs-lib/observability/tracer/config"
	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/datadog"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/grafana"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/newrelic"
	"github.com/fsvxavier/nexs-lib/observability/tracer/providers/opentelemetry"
)

// TracerManager gerencia o tracer provider ativo
type TracerManager struct {
	provider interfaces.TracerProvider
	config   interfaces.Config
}

// NewTracerManager cria um novo gerenciador de tracer
func NewTracerManager() *TracerManager {
	return &TracerManager{}
}

// Init inicializa o tracer manager com a configuração fornecida
func (tm *TracerManager) Init(ctx context.Context, cfg interfaces.Config) (oteltrace.TracerProvider, error) {
	// Validar configuração
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Criar provider baseado no tipo
	provider, err := NewTracerProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	// Inicializar provider
	tracerProvider, err := provider.Init(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
	}

	tm.provider = provider
	tm.config = cfg

	return tracerProvider, nil
}

// InitFromEnv inicializa o tracer manager com configuração das variáveis de ambiente
func (tm *TracerManager) InitFromEnv(ctx context.Context) (oteltrace.TracerProvider, error) {
	cfg := config.LoadFromEnv()
	return tm.Init(ctx, cfg)
}

// Shutdown finaliza o tracer manager
func (tm *TracerManager) Shutdown(ctx context.Context) error {
	if tm.provider != nil {
		return tm.provider.Shutdown(ctx)
	}
	return nil
}

// GetConfig retorna a configuração atual
func (tm *TracerManager) GetConfig() interfaces.Config {
	return tm.config
}

// GetProvider retorna o provider atual
func (tm *TracerManager) GetProvider() interfaces.TracerProvider {
	return tm.provider
}

// NewTracerProvider cria um novo tracer provider baseado na configuração
func NewTracerProvider(config interfaces.Config) (interfaces.TracerProvider, error) {
	switch config.ExporterType {
	case "datadog":
		return datadog.NewProvider(), nil
	case "grafana":
		return grafana.NewProvider(), nil
	case "newrelic":
		return newrelic.NewProvider(), nil
	case "opentelemetry":
		return opentelemetry.NewProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", config.ExporterType)
	}
}

// Factory implementa TracerProviderFactory
type Factory struct{}

// NewFactory cria uma nova instância de Factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProvider cria um novo tracer provider baseado na configuração
func (f *Factory) CreateProvider(config interfaces.Config) (interfaces.TracerProvider, error) {
	return NewTracerProvider(config)
}

// SupportedTypes retorna os tipos de exporters suportados
func (f *Factory) SupportedTypes() []string {
	return []string{"datadog", "grafana", "newrelic", "opentelemetry"}
}

// QuickStart inicializa rapidamente um tracer com configuração mínima
func QuickStart(serviceName, exporterType string) (oteltrace.TracerProvider, *TracerManager, error) {
	cfg := config.DefaultConfig()
	cfg.ServiceName = serviceName
	cfg.ExporterType = exporterType

	tm := NewTracerManager()
	provider, err := tm.Init(context.Background(), cfg)
	if err != nil {
		return nil, nil, err
	}

	return provider, tm, nil
}

// QuickStartFromEnv inicializa rapidamente um tracer com configuração das variáveis de ambiente
func QuickStartFromEnv() (oteltrace.TracerProvider, *TracerManager, error) {
	tm := NewTracerManager()
	provider, err := tm.InitFromEnv(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return provider, tm, nil
}
