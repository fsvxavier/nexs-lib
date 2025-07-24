# Exemplo OpenTelemetry OTLP

Este exemplo demonstra como usar a biblioteca de tracer com OpenTelemetry genérico para diferentes backends.

## Pré-requisitos

1. OpenTelemetry Collector
2. Backend de sua escolha (Jaeger, Zipkin, Prometheus, etc.)
3. (Opcional) Kubernetes cluster para exemplos avançados

## Configuração

### 1. OpenTelemetry Collector

#### Docker Compose

```yaml
version: '3.8'
services:
  # OpenTelemetry Collector
  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    command: ["--config=/etc/otelcol-contrib/otel-collector.yaml"]
    volumes:
      - ./otel-collector.yaml:/etc/otelcol-contrib/otel-collector.yaml
    ports:
      - "4317:4317"   # OTLP gRPC receiver
      - "4318:4318"   # OTLP HTTP receiver
      - "8888:8888"   # Prometheus metrics
      - "8889:8889"   # Prometheus exporter metrics
    depends_on:
      - jaeger
      - zipkin

  # Jaeger
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686" # Jaeger UI
      - "14250:14250" # gRPC
    environment:
      - COLLECTOR_OTLP_ENABLED=true

  # Zipkin  
  zipkin:
    image: openzipkin/zipkin:latest
    ports:
      - "9411:9411"   # Zipkin UI

  # Prometheus
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"   # Prometheus UI
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

#### Configuração do Collector (otel-collector.yaml)

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  memory_limiter:
    limit_mib: 512

exporters:
  # Jaeger
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

  # Zipkin
  zipkin:
    endpoint: "http://zipkin:9411/api/v2/spans"

  # Prometheus (for metrics)
  prometheus:
    endpoint: "0.0.0.0:8889"

  # Logging (debug)
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [jaeger, zipkin, logging]
    
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [prometheus, logging]
```

### 2. Variáveis de Ambiente

```bash
export TRACER_SERVICE_NAME="otel-example-service"
export TRACER_ENVIRONMENT="development"
export TRACER_EXPORTER_TYPE="opentelemetry"
export TRACER_ENDPOINT="http://otel-collector:4318/v1/traces"  # HTTP
# ou
export TRACER_ENDPOINT="otel-collector:4317"                   # gRPC
export TRACER_SAMPLING_RATIO="1.0"
export TRACER_INSECURE="true"
```

### 3. Configurações por Backend

#### Jaeger Direto
```bash
export TRACER_ENDPOINT="http://jaeger:14268/api/traces"  # HTTP
export TRACER_ENDPOINT="jaeger:14250"                    # gRPC
```

#### Zipkin Direto  
```bash
export TRACER_ENDPOINT="http://zipkin:9411/api/v2/spans"
```

#### OTLP Genérico
```bash
export TRACER_ENDPOINT="http://otel-collector:4318/v1/traces"
export TRACER_HEADERS='{"Authorization":"Bearer token","X-Tenant-ID":"tenant-1"}'
```

## Executar o Exemplo

```bash
# Subir infraestrutura
docker-compose up -d

# Aguardar serviços ficarem prontos
sleep 10

# Na pasta do exemplo
go run main.go
```

## O que o Exemplo Faz

Este exemplo simula um **workflow de deployment Kubernetes**:

### 1. Validação de Manifesto
- **kubeval**: Validação de sintaxe
- **Labels**: Verificação de labels obrigatórios
- **Resources**: Validação de requests/limits
- **Security**: Verificação de security context

### 2. Criação de Recursos
- **Deployment**: Aplicação principal
- **Service**: Exposição de rede
- **ConfigMap**: Configurações
- **Secret**: Credenciais

### 3. Verificação de Pods
- **Status**: Verificação se pods estão running
- **Readiness**: Aguarda readiness probes
- **Replicas**: Confirma número desejado

### 4. Service Mesh (Istio)
- **VirtualService**: Roteamento de tráfego
- **DestinationRule**: Políticas de load balancing
- **Sidecar**: Injeção automática

### 5. Testes de Smoke
- **Health Checks**: Endpoints de health
- **API**: Testes de endpoints principais
- **Database**: Conectividade com banco

## Estrutura dos Traces

```
k8s-deployment-workflow (span raiz)
├── k8s.validate-manifest
│   ├── validation.labels
│   ├── validation.resources
│   └── validation.security-context
├── k8s.create-resources
│   ├── k8s.create-deployment
│   ├── k8s.create-service
│   ├── k8s.create-configmap
│   └── k8s.create-secret
├── k8s.wait-for-pods
│   ├── k8s.check-pod-status (pod-1)
│   ├── k8s.check-pod-status (pod-2)
│   └── k8s.check-pod-status (pod-3)
├── k8s.configure-health-checks
├── service-mesh.configure
│   ├── istio.configure-virtual-service
│   └── istio.configure-destination-rule
└── testing.smoke-tests
    ├── test.health-check
    ├── test.api-endpoint
    └── test.database-connection
```

## Visualizar Traces

### 1. Jaeger UI
- **URL**: http://localhost:16686
- **Search**: Por service name ou operation
- **Dependencies**: Visualização de dependências

### 2. Zipkin UI
- **URL**: http://localhost:9411
- **Search**: Por service name ou span name
- **Timeline**: Visualização temporal

### 3. Prometheus Metrics
- **URL**: http://localhost:9090
- **Metrics**: Duração, contadores, histogramas

## Atributos Incluídos

### Kubernetes
- `k8s.deployment.name`: Nome do deployment
- `k8s.namespace`: Namespace
- `k8s.replicas.desired/ready`: Número de réplicas
- `k8s.cluster`: Nome do cluster
- `k8s.resource.type/name`: Tipo e nome do recurso

### Validação
- `validation.tool`: Ferramenta usada
- `validation.type`: Tipo de validação
- `cpu/memory.request/limit`: Recursos definidos

### Service Mesh
- `service_mesh.type`: Tipo (istio, linkerd)
- `service_mesh.sidecar_injection`: Injeção automática
- `istio.resource.type/name`: Recursos do Istio

### Testes
- `test.type`: Tipo de teste
- `test.name`: Nome específico
- `test.result`: Resultado (passed/failed)
- `test.count`: Número total de testes

## Queries e Alertas

### Jaeger Queries
- Service: `otel-example-service`
- Operation: `k8s-deployment-workflow`
- Tags: `k8s.namespace=production`

### Prometheus Metrics
```promql
# Latência média por operação
rate(traces_spans_duration_sum[5m]) / rate(traces_spans_duration_count[5m])

# Taxa de erro por serviço
rate(traces_spans_total{status_code="ERROR"}[5m])

# Throughput por operação
rate(traces_spans_total[5m])
```

## Backends Suportados

### 1. Jaeger
- **Pros**: UI rica, search avançado, service map
- **Uso**: Desenvolvimento e debugging
- **Config**: Endpoint direto ou via collector

### 2. Zipkin
- **Pros**: Simples, leve, compatibilidade
- **Uso**: Ambientes simples
- **Config**: HTTP endpoint direto

### 3. Grafana Tempo
- **Pros**: Integração com Grafana, cost-effective
- **Uso**: Produção com Grafana stack
- **Config**: Via OpenTelemetry collector

### 4. Vendor-specific (Datadog, New Relic)
- **Pros**: Recursos enterprise, alertas, dashboards
- **Uso**: Produção enterprise
- **Config**: Via collector com vendor exporters

## Troubleshooting

### Collector não recebe traces
```bash
# Verificar logs do collector
docker logs otel-collector

# Testar conectividade
curl -v http://localhost:4318/v1/traces
```

### Traces não aparecem no backend
- Verificar configuração do exporter no collector
- Confirmar se backend está rodando
- Verificar logs de ambos os serviços

### Performance
```yaml
# Otimizações para produção
processors:
  batch:
    timeout: 200ms
    send_batch_size: 512
  sampling:
    sampling_percentage: 10  # 10% sampling
```

### Debugging
```bash
# Ativar logs detalhados
export OTEL_LOG_LEVEL=debug

# Verificar métricas do collector
curl http://localhost:8888/metrics
```
