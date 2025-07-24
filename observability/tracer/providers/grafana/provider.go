// Package grafana fornece implementação de tracer provider usando Grafana Tempo via OpenTelemetry
package grafana

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.28.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

// Provider implementa TracerProvider para Grafana Tempo
type Provider struct {
	tracerProvider *trace.TracerProvider
	exporter       trace.SpanExporter
}

// NewProvider cria uma nova instância do provider Grafana
func NewProvider() *Provider {
	return &Provider{}
}

// Init inicializa o tracer provider Grafana Tempo
func (p *Provider) Init(ctx context.Context, config interfaces.Config) (oteltrace.TracerProvider, error) {
	// Criar resource com metadados do serviço
	res, err := p.createResource(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Criar exporter para Grafana Tempo
	exporter, err := p.createExporter(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}
	p.exporter = exporter

	// Configurar sampler
	sampler := trace.TraceIDRatioBased(config.SamplingRatio)

	// Criar tracer provider
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(sampler),
	)

	p.tracerProvider = tracerProvider

	// Configurar propagadores
	if err := p.configurePropagators(config.Propagators); err != nil {
		return nil, fmt.Errorf("failed to configure propagators: %w", err)
	}

	// Definir como global
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}

// Shutdown finaliza o tracer provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown tracer provider: %w", err)
		}
	}
	return nil
}

// createResource cria um resource com metadados do serviço específicos para Grafana
func (p *Provider) createResource(config interfaces.Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(config.ServiceName),
		semconv.ServiceVersion(config.Version),
		semconv.DeploymentEnvironmentName(config.Environment),
		// Adicionar atributos específicos para Grafana
		attribute.String("grafana.tempo", "true"),
	}

	// Adicionar atributos customizados
	for key, value := range config.Attributes {
		attrs = append(attrs, attribute.String(key, value))
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			attrs...,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return res, nil
}

// createExporter cria um exporter OTLP para Grafana Tempo
func (p *Provider) createExporter(ctx context.Context, config interfaces.Config) (trace.SpanExporter, error) {
	// Grafana Tempo geralmente usa gRPC na porta 9095 ou HTTP na porta 3200
	if isGRPCEndpoint(config.Endpoint) {
		return p.createGRPCExporter(ctx, config)
	}
	return p.createHTTPExporter(ctx, config)
}

// createGRPCExporter cria um exporter gRPC para Grafana Tempo
func (p *Provider) createGRPCExporter(ctx context.Context, config interfaces.Config) (trace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithTimeout(30 * time.Second),
	}

	// Configurar headers para Grafana (pode incluir autenticação)
	headers := make(map[string]string)
	for k, v := range config.Headers {
		headers[k] = v
	}

	// Adicionar headers específicos para Grafana se necessário
	if _, exists := headers["tenant-id"]; !exists && config.Attributes["tenant-id"] != "" {
		headers["tenant-id"] = config.Attributes["tenant-id"]
	}

	if len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}

	// Configurar TLS
	if config.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	} else {
		opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	}

	return otlptracegrpc.New(ctx, opts...)
}

// createHTTPExporter cria um exporter HTTP para Grafana Tempo
func (p *Provider) createHTTPExporter(ctx context.Context, config interfaces.Config) (trace.SpanExporter, error) {
	// Para Grafana Tempo via HTTP, geralmente usa /api/traces
	endpoint := config.Endpoint
	if !strings.HasSuffix(endpoint, "/api/traces") && !strings.Contains(endpoint, "/v1/traces") {
		if strings.HasSuffix(endpoint, "/") {
			endpoint += "api/traces"
		} else {
			endpoint += "/api/traces"
		}
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithTimeout(30 * time.Second),
	}

	// Configurar headers para Grafana
	headers := make(map[string]string)
	for k, v := range config.Headers {
		headers[k] = v
	}

	// Adicionar headers específicos para Grafana se necessário
	headers["Content-Type"] = "application/x-protobuf"

	if _, exists := headers["X-Scope-OrgID"]; !exists && config.Attributes["org-id"] != "" {
		headers["X-Scope-OrgID"] = config.Attributes["org-id"]
	}

	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}

	// Configurar TLS
	if config.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	return otlptracehttp.New(ctx, opts...)
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

// isGRPCEndpoint determina se o endpoint é gRPC baseado na porta ou esquema
func isGRPCEndpoint(endpoint string) bool {
	// Grafana Tempo padrão: gRPC na porta 9095, HTTP na porta 3200
	return strings.Contains(endpoint, ":9095") ||
		(!hasHTTPScheme(endpoint) && !strings.Contains(endpoint, ":3200"))
}

// hasHTTPScheme verifica se o endpoint tem esquema HTTP
func hasHTTPScheme(endpoint string) bool {
	return strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://")
}
