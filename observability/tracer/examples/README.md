# Exemplos OpenTelemetry Tracer

Esta pasta contém exemplos completos e funcionais para usar o sistema de tracing com diferentes backends e cenários.

## 📁 Estrutura dos Exemplos

### 🔧 Providers Específicos
- **`datadog/`** - Integração com Datadog APM
- **`grafana/`** - Integração com Grafana Tempo 
- **`newrelic/`** - Integração com New Relic
- **`opentelemetry/`** - Backend OpenTelemetry genérico

### 🌍 Configuração Global
- **`global/`** - Usando `otel.SetTracerProvider()` globalmente

### 🚀 Exemplo Avançado
- **`advanced/`** - Integração completa de **traces + logs + métricas**

## 🏃‍♂️ Quick Start

### 1. Escolher um Exemplo

```bash
# Para Datadog
cd datadog/
export TRACER_EXPORTER_TYPE="datadog"
export DATADOG_API_KEY="your-api-key"

# Para Grafana Tempo
cd grafana/
export TRACER_EXPORTER_TYPE="grafana"
export TRACER_ENDPOINT="http://tempo:3200"

# Para New Relic
cd newrelic/
export TRACER_EXPORTER_TYPE="newrelic"
export NEW_RELIC_LICENSE_KEY="your-license-key"

# Para OpenTelemetry
cd opentelemetry/
export TRACER_EXPORTER_TYPE="opentelemetry"
export TRACER_ENDPOINT="http://otel-collector:4318/v1/traces"
```

### 2. Executar

```bash
go run main.go
```

### 3. Testar

```bash
# Health check
curl http://localhost:8080/health

# Criar usuário (gera traces)
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"João","email":"joao@example.com"}'

# Buscar usuário
curl http://localhost:8080/users/12345
```

## 🎯 Cenários de Uso

### 📊 Provider Específicos (datadog/, grafana/, newrelic/, opentelemetry/)
- **Objetivo**: Demonstrar integração com backend específico
- **Cenário**: API de gerenciamento de usuários
- **Features**:
  - CRUD de usuários
  - Operações de banco de dados simuladas
  - Cache Redis simulado
  - Validação e processamento
  - Error handling com traces

### 🌐 Configuração Global (global/)
- **Objetivo**: TracerProvider global com `otel.SetTracerProvider()`
- **Cenário**: Aplicação web complexa
- **Features**:
  - Múltiplos componentes usando tracer global
  - Middlewares HTTP automaticamente instrumentados
  - Integrações com bibliotecas de terceiros
  - Propagação automática de context

### 🔬 Exemplo Avançado (advanced/)
- **Objetivo**: Observabilidade completa (traces + logs + métricas)
- **Cenário**: Sistema de e-commerce
- **Features**:
  - **Traces**: Rastreamento end-to-end de pedidos
  - **Logs**: Correlacionados com trace context
  - **Métricas**: Business e performance metrics
  - Múltiplos serviços internos
  - Simulação de carga de trabalho

## 📋 Configuração por Backend

### Datadog
```bash
export TRACER_SERVICE_NAME="my-service"
export TRACER_ENVIRONMENT="production"
export TRACER_EXPORTER_TYPE="datadog"
export DATADOG_API_KEY="your-datadog-api-key"
export TRACER_SAMPLING_RATIO="0.1"
```

### Grafana Tempo
```bash
export TRACER_SERVICE_NAME="my-service"
export TRACER_ENVIRONMENT="production"
export TRACER_EXPORTER_TYPE="grafana"
export TRACER_ENDPOINT="http://tempo:3200"
export TRACER_SAMPLING_RATIO="1.0"
```

### New Relic
```bash
export TRACER_SERVICE_NAME="my-service"
export TRACER_ENVIRONMENT="production"
export TRACER_EXPORTER_TYPE="newrelic"
export NEW_RELIC_LICENSE_KEY="your-40-char-license-key"
export TRACER_SAMPLING_RATIO="0.1"
```

### OpenTelemetry Collector
```bash
export TRACER_SERVICE_NAME="my-service"
export TRACER_ENVIRONMENT="production"
export TRACER_EXPORTER_TYPE="opentelemetry"
export TRACER_ENDPOINT="http://otel-collector:4318/v1/traces"
export TRACER_HEADERS='{"Authorization":"Bearer token"}'
export TRACER_INSECURE="false"
export TRACER_SAMPLING_RATIO="0.1"
```

## 🔍 Estrutura de Traces

Todos os exemplos geram traces seguindo padrões similares:

### Exemplo de Provider Específico
```
http-request (root span)
├── create-user
│   ├── validate-user-data
│   ├── check-user-exists
│   │   └── query-user-by-email
│   ├── create-user-record
│   │   └── insert-user
│   └── cache-user-data
│       └── redis-set
└── http-response
```

### Exemplo Global
```
http-request (root span)
├── authentication
│   ├── validate-jwt
│   └── check-permissions
├── process-business-logic
│   ├── fetch-user-data
│   │   ├── query-user
│   │   └── cache-user-data
│   ├── enrich-user-profile
│   │   ├── fetch-preferences
│   │   ├── fetch-history
│   │   └── fetch-recommendations
│   └── audit-user-access
```

### Exemplo Avançado
```
create-order (root span)
├── validate-order
├── process-payment
├── check-inventory
│   └── checking_item (events)
├── calculate-shipping
└── send-notifications
    ├── send-email
    ├── send-sms
    └── send-push-notification
```

## 📊 Atributos Semânticos

### HTTP Attributes
- `http.method`: GET, POST, PUT, DELETE
- `http.route`: Route pattern (/users/{id})
- `http.status_code`: Response status
- `http.user_agent`: Client user agent

### Database Attributes
- `db.system`: Database system (postgres, mysql)
- `db.operation`: SQL operation (SELECT, INSERT)
- `db.name`: Database name
- `db.table`: Table name
- `db.rows_affected`: Number of affected rows

### Cache Attributes
- `cache.system`: Cache system (redis, memcached)
- `cache.key`: Cache key
- `cache.hit`: true/false
- `cache.ttl`: Time to live

### Business Attributes
- `user.id`: User identifier
- `user.email`: User email
- `order.id`: Order identifier
- `payment.method`: Payment method
- `operation.type`: Type of operation

## 🚀 Docker Compose

Para facilitar os testes, você pode usar este `docker-compose.yml`:

```yaml
version: '3.8'
services:
  # Jaeger
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "14250:14250"
    environment:
      - COLLECTOR_OTLP_ENABLED=true

  # OpenTelemetry Collector
  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
      - "8888:8888"   # Prometheus metrics
    depends_on:
      - jaeger

  # Grafana Tempo
  tempo:
    image: grafana/tempo:latest
    command: [ "-config.file=/etc/tempo.yaml" ]
    volumes:
      - ./tempo.yaml:/etc/tempo.yaml
    ports:
      - "3200:3200"   # Tempo
      - "9095:9095"   # Tempo gRPC

  # Grafana
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

## 🔧 Troubleshooting

### Traces não aparecem
1. Verifique as variáveis de ambiente
2. Confirme que o backend está rodando
3. Verifique se o sampling está adequado
4. Confirme conectividade de rede

### Performance Issues
1. Ajuste o sampling ratio
2. Configure batching adequado
3. Monitore uso de CPU/memória
4. Use compression se disponível

### Logs de Debug
```bash
export OTEL_LOG_LEVEL=debug
go run main.go
```

## 📚 Próximos Passos

1. **Estude** um exemplo específico do seu backend
2. **Execute** o exemplo localmente
3. **Adapte** para seu caso de uso
4. **Configure** em produção com sampling adequado
5. **Monitore** performance e ajuste conforme necessário

## 🤝 Contribuindo

Para adicionar novos exemplos:

1. Crie uma pasta com nome descritivo
2. Inclua `main.go` funcional
3. Adicione `README.md` detalhado
4. Garanta que compila sem erros
5. Documente configuração necessária

## 📖 Documentação Adicional

- [Configuração Principal](../config/)
- [Interfaces](../interfaces/)
- [Providers](../providers/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
