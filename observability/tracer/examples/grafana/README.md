# Exemplo Grafana Tempo

Este exemplo demonstra como usar a biblioteca de tracer com Grafana Tempo para observabilidade distribuída.

## Pré-requisitos

1. Grafana com Tempo datasource configurado
2. Tempo rodando (local ou remoto)
3. (Opcional) Prometheus para métricas

## Configuração

### 1. Variáveis de Ambiente

```bash
export TRACER_SERVICE_NAME="grafana-example-service"
export TRACER_ENVIRONMENT="development"
export TRACER_EXPORTER_TYPE="grafana"
export TRACER_ENDPOINT="http://tempo:3200"  # HTTP
# ou
export TRACER_ENDPOINT="tempo:9095"         # gRPC
export TRACER_SAMPLING_RATIO="1.0"
```

### 2. Docker Compose para Grafana Stack

```yaml
version: '3.8'
services:
  tempo:
    image: grafana/tempo:latest
    command: [ "-config.file=/etc/tempo.yaml" ]
    volumes:
      - ./tempo.yaml:/etc/tempo.yaml
    ports:
      - "3200:3200"   # HTTP
      - "9095:9095"   # gRPC

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - ./grafana-datasources.yml:/etc/grafana/provisioning/datasources/datasources.yml
```

### 3. Configuração do Tempo (tempo.yaml)

```yaml
server:
  http_listen_port: 3200
  grpc_listen_port: 9095

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318

ingester:
  trace_idle_period: 10s
  max_block_bytes: 1_000_000
  max_block_duration: 5m

compactor:
  compaction:
    compaction_window: 1h
    max_compaction_objects: 1000000
    block_retention: 1h

storage:
  trace:
    backend: local
    local:
      path: /tmp/tempo/traces
```

### 4. Datasource do Grafana (grafana-datasources.yml)

```yaml
apiVersion: 1

datasources:
  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200
    isDefault: true
```

## Executar o Exemplo

```bash
# Subir a stack do Grafana
docker-compose up -d

# Na pasta do exemplo
go run main.go
```

## O que o Exemplo Faz

Este exemplo simula um **workflow de e-commerce completo**:

1. **Validação do Pedido**: Verifica regras de negócio
2. **Verificação de Inventário**: Consulta múltiplos warehouses
3. **Processamento de Pagamento**: Integração com gateway de pagamento
4. **Reserva de Inventário**: Confirma disponibilidade
5. **Criação de Shipping**: Gera etiqueta e tracking
6. **Notificação ao Cliente**: Envia confirmação por email

## Estrutura dos Traces

```
ecommerce-order-workflow (span raiz)
├── validate-order
├── check-inventory
│   ├── check-warehouse (warehouse-1)
│   ├── check-warehouse (warehouse-2)
│   └── check-warehouse (warehouse-3)
├── process-payment
├── reserve-inventory
├── create-shipping
└── notify-customer
```

## Visualizar Traces

1. **Grafana**: http://localhost:3000
   - Login: admin/admin
   - Vá para Explore
   - Selecione Tempo datasource
   - Busque por traces recentes ou por service name

2. **Busca por Atributos**:
   - `service.name = "grafana-example-service"`
   - `order.id = "order-123456"`
   - `workflow.type = "order-processing"`

## Atributos Incluídos

### Workflow Principal
- `workflow.type`: Tipo do workflow
- `order.id`: ID do pedido
- `customer.id`: ID do cliente
- `order.items_count`: Quantidade de itens
- `order.total`: Valor total

### Inventário
- `warehouse.id`: ID do warehouse
- `items.available`: Itens disponíveis
- `inventory.system`: Sistema de inventário

### Pagamento
- `payment.amount`: Valor do pagamento
- `payment.method`: Método de pagamento
- `payment.gateway`: Gateway utilizado

### Shipping
- `shipping.carrier`: Transportadora
- `shipping.service`: Tipo de serviço
- `tracking.number`: Número de rastreamento

### Notificação
- `notification.type`: Tipo de notificação
- `notification.channel`: Canal utilizado

## Queries Úteis no Grafana

### Buscar traces por duração
```
{duration > 100ms}
```

### Buscar por serviço e status
```
{service.name="grafana-example-service" && status=error}
```

### Buscar por atributos específicos
```
{order.id="order-123456"}
```

## Troubleshooting

### Tempo não está recebendo traces
- Verifique se o endpoint está correto
- Confirme se Tempo está rodando na porta 3200 (HTTP) ou 9095 (gRPC)
- Verifique logs do container: `docker logs tempo`

### Traces não aparecem no Grafana
- Confirme se o datasource Tempo está configurado
- Verifique se há dados no período selecionado
- Teste a conexão do datasource

### Performance
- Ajuste o sampling ratio para produção (ex: 0.1 = 10%)
- Configure retenção adequada no Tempo
- Use filtros específicos para buscas grandes
