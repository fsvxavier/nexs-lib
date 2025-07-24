# Exemplo New Relic Distributed Tracing

Este exemplo demonstra como usar a biblioteca de tracer com New Relic para observabilidade de microserviços.

## Pré-requisitos

1. Conta no New Relic
2. License Key do New Relic (40 caracteres)
3. Aplicação configurada no New Relic

## Configuração

### 1. Obter License Key

1. Login no New Relic: https://one.newrelic.com
2. Vá para **Account settings > API keys**
3. Copie a **License Key** (40 caracteres)

### 2. Variáveis de Ambiente

```bash
export TRACER_SERVICE_NAME="newrelic-example-service"
export TRACER_ENVIRONMENT="development"
export TRACER_EXPORTER_TYPE="newrelic"
export NEW_RELIC_LICENSE_KEY="your-40-char-license-key"
export TRACER_SAMPLING_RATIO="1.0"

# Para região EU (opcional)
export TRACER_ENDPOINT="https://trace-api.eu.newrelic.com/trace/v1"
```

### 3. Arquivo .env

```env
TRACER_SERVICE_NAME=newrelic-example-service
TRACER_ENVIRONMENT=development
TRACER_EXPORTER_TYPE=newrelic
NEW_RELIC_LICENSE_KEY=your-40-char-license-key
TRACER_SAMPLING_RATIO=1.0
```

### 4. Configuração de Região

```go
// US Datacenter (padrão)
cfg.Endpoint = "https://trace-api.newrelic.com/trace/v1"

// EU Datacenter
cfg.Endpoint = "https://trace-api.eu.newrelic.com/trace/v1"
```

## Executar o Exemplo

```bash
# Na pasta do exemplo
go run main.go
```

## O que o Exemplo Faz

Este exemplo simula uma **arquitetura de microserviços** completa:

### 1. API Gateway / BFF
- Recebe request HTTP
- Autentica usuário
- Orquestra chamadas para microserviços

### 2. Microserviços Incluídos
- **Auth Service**: Autenticação JWT
- **User Service**: Dados do usuário + PostgreSQL
- **Product Service**: Catálogo + MongoDB + Redis Cache
- **Inventory Service**: Controle de estoque
- **Payment Service**: Pagamento + Stripe API
- **Order Service**: Criação de pedidos
- **Notification Service**: Notificações por email

### 3. Integrações Externas
- **Stripe API**: Gateway de pagamento
- **Cache Redis**: Cache de produtos
- **Bancos de Dados**: PostgreSQL, MongoDB

## Estrutura dos Traces

```
microservices-api-workflow (span raiz)
├── auth-service.authenticate
├── user-service.get-user
│   └── database.query-user
├── product-service.get-product
│   ├── cache.lookup-product
│   └── database.query-product
├── inventory-service.check-stock
├── payment-service.process-payment
│   └── external.stripe-api
├── order-service.create-order
└── notification-service.send-confirmation
```

## Visualizar Traces

1. **New Relic One**: https://one.newrelic.com
2. **Distributed Tracing**: https://one.newrelic.com/distributed-tracing
3. **Service Map**: Visualização de dependências
4. **Trace Search**: Busca por traces específicos

## Atributos Incluídos

### Request Principal
- `workflow.type`: Tipo do workflow
- `http.method`: Método HTTP
- `http.route`: Rota da API
- `request.id`: ID da requisição
- `user.id`: ID do usuário
- `http.status_code`: Status de resposta

### Serviços Individuais
- `service.name`: Nome do microserviço
- `auth.method`: Método de autenticação
- `cache.hit`: Hit/miss do cache
- `payment.provider`: Provedor de pagamento
- `order.status`: Status do pedido

### Base de Dados
- `db.system`: Sistema de banco (PostgreSQL, MongoDB)
- `db.name`: Nome do banco
- `db.operation`: Operação (SELECT, findOne)
- `db.table`/`db.collection`: Tabela/coleção

### Integrações Externas
- `external.service`: Nome do serviço externo
- `payment.currency`: Moeda do pagamento
- `notification.channel`: Canal de notificação

## Queries Úteis no New Relic

### Service Performance
```sql
SELECT average(duration.ms) 
FROM Span 
WHERE service.name = 'newrelic-example-service' 
FACET name 
SINCE 1 hour ago
```

### Error Rate
```sql
SELECT percentage(count(*), WHERE error IS true) 
FROM Span 
WHERE service.name = 'newrelic-example-service' 
SINCE 1 hour ago
```

### Latency P95
```sql
SELECT percentile(duration.ms, 95) 
FROM Span 
WHERE service.name = 'newrelic-example-service' 
FACET name 
SINCE 1 hour ago
```

### Database Performance
```sql
SELECT average(duration.ms) 
FROM Span 
WHERE db.system IS NOT NULL 
FACET db.system, db.operation 
SINCE 1 hour ago
```

## Alertas Recomendados

### 1. High Error Rate
```sql
SELECT percentage(count(*), WHERE error IS true) 
FROM Span 
WHERE service.name = 'newrelic-example-service'
```
- **Threshold**: > 5%
- **Duration**: 5 minutes

### 2. High Latency
```sql
SELECT average(duration.ms) 
FROM Span 
WHERE service.name = 'newrelic-example-service' 
AND name = 'microservices-api-workflow'
```
- **Threshold**: > 2000ms
- **Duration**: 10 minutes

### 3. External Service Failures
```sql
SELECT count(*) 
FROM Span 
WHERE external.service IS NOT NULL 
AND error IS true
```
- **Threshold**: > 10 errors
- **Duration**: 5 minutes

## Dashboards Sugeridos

### 1. Service Overview
- Request rate (RPM)
- Average response time
- Error rate percentage
- Apdex score

### 2. Service Dependencies
- Service map
- External services performance
- Database query performance
- Cache hit ratio

### 3. Business Metrics
- Orders created per minute
- Payment success rate
- User authentication rate
- Notification delivery rate

## Troubleshooting

### License Key inválida
```bash
# Verificar se a license key tem 40 caracteres
echo $NEW_RELIC_LICENSE_KEY | wc -c
```

### Traces não aparecem
- Aguarde 1-2 minutos para ingestão
- Verifique se a license key está correta
- Confirme se está usando o endpoint correto (US/EU)

### Performance em Produção
- Reduza sampling ratio: `TRACER_SAMPLING_RATIO=0.1` (10%)
- Configure tail-based sampling no New Relic
- Use filtros por importância de transação

### Região/Datacenter
- **US**: `https://trace-api.newrelic.com/trace/v1`
- **EU**: `https://trace-api.eu.newrelic.com/trace/v1`
- Verifique qual região sua conta está configurada
