# Exemplo Avançado: Integração Traces + Logs + Métricas

Este exemplo demonstra como integrar **traces**, **logs estruturados** e **métricas** OpenTelemetry em uma aplicação real de e-commerce.

## Conceito

O exemplo mostra uma aplicação completa que:

1. **Traces**: Rastreamento end-to-end de operações de pedidos
2. **Logs Estruturados**: Logging correlacionado com trace context 
3. **Métricas**: Métricas de negócio e performance
4. **Observabilidade**: Correlação entre todos os sinais

## Arquitetura da Aplicação

### HTTP Endpoints
- `GET /health` - Health check
- `POST /orders` - Criar pedido
- `GET /orders/{id}` - Buscar pedido

### Serviços Internos
- **Order Processor** - Processamento principal
- **Payment Service** - Processamento de pagamentos
- **Inventory Service** - Verificação de estoque
- **Shipping Service** - Cálculo de frete
- **Notification Service** - Envio de notificações

## Integração dos Três Pilares

### 1. Traces Distribuídos

```go
tracer := otel.Tracer("order-service")
ctx, span := tracer.Start(r.Context(), "create-order")
defer span.End()

// Atributos semânticos
span.SetAttributes(
    attribute.String("order.id", order.ID),
    attribute.Float64("order.amount", order.Amount),
    attribute.String("order.payment_method", order.PaymentMethod),
)
```

### 2. Logs Correlacionados

```go
logger.Info("Order created successfully",
    zap.String("order_id", order.ID),
    zap.Float64("amount", order.Amount),
    zap.String("trace_id", span.SpanContext().TraceID().String()),
    zap.String("span_id", span.SpanContext().SpanID().String()))
```

### 3. Métricas de Negócio

```go
// Contador de pedidos processados
orderProcessed.Add(ctx, 1, 
    metric.WithAttributes(
        attribute.String("status", "success"), 
        attribute.String("payment_method", order.PaymentMethod)))

// Histograma de valor dos pedidos
orderValue.Record(ctx, order.Amount, 
    metric.WithAttributes(
        attribute.String("payment_method", order.PaymentMethod)))
```

## Configuração

### 1. Variáveis de Ambiente

```bash
# Tracing
export TRACER_SERVICE_NAME="ecommerce-api"
export TRACER_ENVIRONMENT="production"
export TRACER_EXPORTER_TYPE="opentelemetry"
export TRACER_ENDPOINT="http://otel-collector:4318/v1/traces"

# Sampling para produção
export TRACER_SAMPLING_RATIO="0.1"  # 10%
```

### 2. Executar o Exemplo

```bash
# Configurar backend
export TRACER_EXPORTER_TYPE="opentelemetry"
export TRACER_ENDPOINT="http://localhost:4318/v1/traces"

# Executar
go run main.go
```

### 3. Testar a API

```bash
# Health check
curl http://localhost:8080/health

# Criar pedido
curl -X POST http://localhost:8080/orders

# Buscar pedido
curl http://localhost:8080/orders/ORD-123456
```

## Estrutura de Traces

### Trace Hierarquia

```
create-order (root span)
├── validate-order
├── process-payment
├── check-inventory
│   └── checking_item (events para cada produto)
├── calculate-shipping
└── send-notifications
    ├── send-email
    ├── send-sms
    └── send-push-notification
```

### Atributos Semânticos

#### HTTP
- `http.method`: Método HTTP
- `http.route`: Rota da API
- `http.status_code`: Status da resposta

#### Order
- `order.id`: Identificador do pedido
- `order.customer_id`: ID do cliente
- `order.amount`: Valor do pedido
- `order.payment_method`: Método de pagamento
- `order.items_count`: Número de itens

#### Payment
- `payment.method`: Método de pagamento
- `payment.amount`: Valor processado

#### Inventory
- `items.count`: Número de itens verificados
- `item.product_id`: ID do produto
- `item.quantity`: Quantidade

#### Shipping
- `shipping.cost`: Custo do frete
- `customer.id`: ID do cliente

#### Notifications
- Spans separados para email, SMS e push

## Métricas Coletadas

### 1. HTTP Metrics

```go
// Contador total de requests
http_requests_total{method="POST", endpoint="/orders", status="201"}

// Histograma de latência
http_request_duration_seconds{method="POST", endpoint="/orders"}

// Operações ativas
active_operations{operation_type="order_creation"}
```

### 2. Business Metrics

```go
// Pedidos processados
orders_processed_total{status="success", payment_method="credit_card"}

// Valor dos pedidos
order_value_dollars{payment_method="credit_card"}

// Traces gerados
traces_generated_total{operation="order_processing"}

// Duração das operações
span_duration_seconds{operation="process_payment"}
```

## Logs Estruturados

### Formato JSON

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:45Z",
  "message": "Order created successfully",
  "order_id": "ORD-123456",
  "customer_id": "CUST-7890",
  "amount": 149.99,
  "payment_method": "credit_card",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7"
}
```

### Correlação com Traces

Cada log inclui:
- `trace_id`: Para correlacionar com traces
- `span_id`: Para identificar o span específico
- Campos de contexto específicos da operação

## Observabilidade em Produção

### 1. Monitoramento

#### SLIs (Service Level Indicators)
- **Latência**: P95 < 500ms para criação de pedidos
- **Disponibilidade**: 99.9% uptime
- **Taxa de Erro**: < 0.1% de falhas de pagamento

#### SLOs (Service Level Objectives)
- 95% dos pedidos processados em < 2 segundos
- 99.5% de disponibilidade mensal
- < 0.5% de taxa de erro

#### Alertas
```yaml
# Exemplo de alerta Prometheus
- alert: HighOrderProcessingLatency
  expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{endpoint="/orders"}[5m])) > 0.5
  labels:
    severity: warning
  annotations:
    summary: "High order processing latency"
```

### 2. Dashboards

#### Grafana Dashboard
- **Golden Signals**: Latência, Tráfego, Erros, Saturação
- **Business Metrics**: Pedidos/min, Valor médio, Métodos de pagamento
- **Trace Analytics**: Distribuição de latência por serviço

#### Trace Views
- **Service Map**: Visualização das dependências
- **Flame Graph**: Análise de performance por span
- **Error Analysis**: Traces com erros

### 3. Análise de Performance

#### Detecção de Bottlenecks
```go
// Medir duração de operações críticas
start := time.Now()
defer func() {
    duration := time.Since(start).Seconds()
    spanDuration.Record(ctx, duration, metric.WithAttributes(
        attribute.String("operation", "process_payment"),
    ))
}()
```

#### Sampling Inteligente
- **Head-based**: 10% em produção
- **Tail-based**: 100% para traces com erros
- **Adaptive**: Aumento dinâmico em eventos

### 4. Debugging

#### Trace-based Debugging
1. **Identificar** - Encontrar trace com erro
2. **Analisar** - Examinar spans e atributos
3. **Correlacionar** - Verificar logs relacionados
4. **Resolver** - Aplicar correções

#### Log Correlation
```bash
# Buscar logs por trace ID
kubectl logs -f deployment/order-service | grep "4bf92f3577b34da6a3ce929d0e0e4736"
```

## Cenários de Uso

### 1. Análise de Performance
- Identificar operações lentas
- Analisar gargalos por serviço
- Otimizar queries de banco de dados

### 2. Debugging de Erros
- Rastrear origem de falhas
- Correlacionar logs com traces
- Analisar impacto de erros

### 3. Monitoramento de Negócio
- Acompanhar métricas de receita
- Analisar padrões de uso
- Detectar anomalias

### 4. Capacity Planning
- Prever crescimento de tráfego
- Identificar recursos limitantes
- Planejar scaling

## Integração com Ferramentas

### Jaeger
```yaml
jaeger:
  query:
    base-path: /jaeger
  collector:
    grpc-port: 14250
```

### Prometheus + Grafana
```yaml
prometheus:
  scrape_configs:
    - job_name: 'otel-collector'
      static_configs:
        - targets: ['otel-collector:8888']
```

### ELK Stack
```yaml
filebeat:
  inputs:
    - type: log
      paths:
        - "/var/log/app/*.log"
      json.keys_under_root: true
```

## Melhores Práticas

### 1. Atributos Semânticos
- Use convenções OpenTelemetry
- Inclua contexto de negócio
- Evite alta cardinalidade

### 2. Sampling
- Configure sampling adequado
- Use tail-based sampling
- Monitore overhead

### 3. Logs Correlacionados
- Sempre inclua trace/span IDs
- Use logging estruturado
- Evite logs verbosos em produção

### 4. Métricas
- Foque em métricas de negócio
- Use histogramas para latência
- Monitore cardinalidade

### 5. Performance
- Configure batching
- Use compression
- Monitore resource usage

Este exemplo fornece uma base sólida para implementar observabilidade completa em aplicações Go!
