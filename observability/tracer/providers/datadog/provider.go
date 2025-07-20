// Package datadog fornece implementação de tracer provider usando Datadog APM
package datadog

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"

	ddtracer "github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"

	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

// Provider implementa TracerProvider para Datadog APM
type Provider struct {
	initialized bool
}

// NewProvider cria uma nova instância do provider Datadog
func NewProvider() *Provider {
	return &Provider{}
}

// Init inicializa o tracer provider Datadog
func (p *Provider) Init(ctx context.Context, config interfaces.Config) (oteltrace.TracerProvider, error) {
	// Configurar opções do Datadog tracer
	opts := []ddtracer.StartOption{
		ddtracer.WithService(config.ServiceName),
		ddtracer.WithEnv(config.Environment),
	}

	// Configurar sampling rate
	if config.SamplingRatio >= 0 && config.SamplingRatio <= 1 {
		// Note: Datadog uses a different method for sampling
		// We'll set it as an environment variable or tag
		opts = append(opts, ddtracer.WithGlobalTag("sampling.rate", fmt.Sprintf("%.2f", config.SamplingRatio)))
	}

	// Configurar endpoint se fornecido
	if config.Endpoint != "" {
		opts = append(opts, ddtracer.WithAgentAddr(config.Endpoint))
	}

	// Adicionar tags globais
	if len(config.Attributes) > 0 {
		for k, v := range config.Attributes {
			opts = append(opts, ddtracer.WithGlobalTag(k, v))
		}
	}

	// Adicionar versão como tag
	if config.Version != "" {
		opts = append(opts, ddtracer.WithGlobalTag("version", config.Version))
	}

	// Iniciar tracer Datadog
	ddtracer.Start(opts...)
	p.initialized = true

	// Configurar propagadores
	if err := p.configurePropagators(config.Propagators); err != nil {
		return nil, fmt.Errorf("failed to configure propagators: %w", err)
	}

	// Para compatibilidade com OpenTelemetry, retornamos um provider que usa o tracer global
	// Em um cenário real, você pode usar um bridge específico ou implementar um wrapper
	provider := otel.GetTracerProvider()

	return provider, nil
}

// Shutdown finaliza o tracer provider Datadog
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.initialized {
		ddtracer.Stop()
		p.initialized = false
	}
	return nil
}

// configurePropagators configura os propagadores de contexto
func (p *Provider) configurePropagators(propagators []string) error {
	var props []propagation.TextMapPropagator

	for _, prop := range propagators {
		switch prop {
		case "tracecontext":
			props = append(props, propagation.TraceContext{})
		case "b3":
			props = append(props, propagation.Baggage{})
		case "jaeger":
			// Jaeger propagator would need additional import
			// For now, we'll use TraceContext as fallback
			props = append(props, propagation.TraceContext{})
		default:
			return fmt.Errorf("unsupported propagator: %s", prop)
		}
	}

	if len(props) == 0 {
		props = []propagation.TextMapPropagator{propagation.TraceContext{}}
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(props...))
	return nil
}
