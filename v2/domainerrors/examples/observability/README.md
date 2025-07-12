# Observability Examples

Este exemplo demonstra a integração completa do Domain Errors v2 com ferramentas de observabilidade, incluindo logging estruturado, coleta de métricas, distributed tracing, health checks, alerting, monitoramento de performance e análise de erros.

## 🎯 Funcionalidades Demonstradas

### 1. Structured Logging
- Logging estruturado com contexto completo de erro
- Múltiplos outputs (console, arquivo, Elasticsearch)
- Correlação de logs com trace IDs
- Campos padronizados para análise

### 2. Metrics Collection
- Contadores de erro por código, tipo e severidade
- Histogramas de duração de processamento
- Gauges para erros ativos por serviço
- Estatísticas agregadas por serviço

### 3. Distributed Tracing
- Integração com Jaeger e Zipkin
- Propagação de contexto entre serviços
- Spans com metadata de erro completa
- Análise de critical path

### 4. Health Checks
- Verificação de saúde de componentes
- Timeouts configuráveis
- Metadata detalhada de status
- Agregação de status geral

### 5. Alerting System
- Regras de alerta baseadas em métricas
- Interpolação de templates
- Múltiplos níveis de severidade
- Estado de alertas (firing, pending, silenced)

### 6. Performance Monitoring
- Benchmarks de criação de erros
- Monitoramento de serialização
- Thresholds configuráveis
- Análise de percentis (P95, P99)

### 7. Error Aggregation
- Janelas temporais configuráveis
- Distribuição por código e severidade
- Análise de tendências
- Identificação de padrões

### 8. Dashboard Integration
- Métricas consolidadas para dashboards
- SLA tracking e error budgets
- Trends históricos
- Alertas ativos

## 🏗️ Arquitetura

### Structured Logger
```go
type StructuredLogger struct {
    level   LogLevel
    outputs map[string]LogOutput
}

// Múltiplos outputs para flexibilidade
logger.AddOutput("console", &ConsoleOutput{})
logger.AddOutput("elastic", &ElasticOutput{URL: "http://elasticsearch:9200"})
```

### Metrics Collector
```go
type MetricsCollector struct {
    counters   map[string]*Counter
    histograms map[string]*Histogram
    gauges     map[string]*Gauge
}

// Métricas especializadas para erros
metrics.RegisterCounter("errors_total", "Total number of errors", []string{"code", "type", "severity"})
```

### Distributed Tracer
```go
type DistributedTracer struct {
    samplingRate float64
    exporters    map[string]TraceExporter
}

// Múltiplos exporters para máxima compatibilidade
tracer.AddExporter("jaeger", &JaegerExporter{})
tracer.AddExporter("zipkin", &ZipkinExporter{})
```

### Health Checker
```go
type HealthChecker struct {
    checks map[string]HealthCheck
}

// Diferentes tipos de health checks
healthChecker.RegisterCheck("database", &DatabaseHealthCheck{})
healthChecker.RegisterCheck("redis", &RedisHealthCheck{})
```

## 📊 Métricas Coletadas

### Error Metrics
- **errors_total**: Total de erros (counter)
  - Labels: code, type, severity, service
- **error_duration**: Duração do processamento (histogram)
  - Labels: operation, service
- **active_errors**: Erros atualmente ativos (gauge)
  - Labels: service, type

### Performance Metrics
- **error_creation_duration**: Tempo de criação de erros
- **error_serialization_duration**: Tempo de serialização
- **error_transmission_duration**: Tempo de transmissão

### Health Metrics
- **component_health**: Status de saúde dos componentes
- **health_check_duration**: Duração dos health checks
- **health_check_success_rate**: Taxa de sucesso

## 🎮 Como Executar

```bash
cd /mnt/e/go/src/github.com/fsvxavier/nexs-lib/v2/domainerrors/examples/observability
go run main.go
```

## 📈 Integração com Ferramentas

### Prometheus Integration
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'domain-errors'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "Domain Errors Monitoring",
    "panels": [
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(errors_total[5m])",
            "legendFormat": "{{service}} - {{code}}"
          }
        ]
      }
    ]
  }
}
```

### Jaeger Tracing
```yaml
# jaeger-config.yml
jaeger:
  endpoint: "http://jaeger-collector:14268/api/traces"
  sampling:
    type: "probabilistic"
    param: 1.0
```

### Elasticsearch Logging
```json
{
  "index_patterns": ["domain-errors-*"],
  "template": {
    "mappings": {
      "properties": {
        "@timestamp": {"type": "date"},
        "level": {"type": "keyword"},
        "error_code": {"type": "keyword"},
        "error_type": {"type": "keyword"},
        "service": {"type": "keyword"},
        "trace_id": {"type": "keyword"}
      }
    }
  }
}
```

## 🚨 Configuração de Alertas

### Prometheus Alert Rules
```yaml
groups:
  - name: domain-errors
    rules:
      - alert: HighErrorRate
        expr: rate(errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors/sec"

      - alert: CriticalErrors
        expr: increase(errors_total{severity="critical"}[1m]) > 0
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "Critical errors detected"
          description: "{{ $value }} critical errors in the last minute"
```

### PagerDuty Integration
```yaml
pagerduty_configs:
  - service_key: "your-service-key"
    description: "{{ .CommonAnnotations.summary }}"
    severity: "{{ .CommonLabels.severity }}"
```

## 📊 Dashboard Widgets

### Overview Panel
- Total errors (24h)
- Error rate trend
- Critical errors count
- Services affected
- MTTR (Mean Time to Recovery)

### Service Health Panel
- Service status indicators
- Error count per service
- Response time metrics
- Availability percentage

### Error Analysis Panel
- Top error codes
- Error distribution by type
- Severity breakdown
- Trend analysis

### SLA Monitoring Panel
- Availability SLA tracking
- Error budget consumption
- Monthly error budget status
- SLA violations history

## 🔧 Configuração Avançada

### Custom Metrics
```go
// Registrar métricas customizadas
metrics.RegisterHistogram("business_operation_duration", "Duration of business operations", []string{"operation", "result"})
metrics.RegisterCounter("business_rule_violations", "Business rule violations", []string{"rule", "entity"})
```

### Sampling Configuration
```go
// Configurar sampling para tracing
tracer.SetSamplingRate(0.1) // 10% para produção
tracer.SetSamplingRate(1.0) // 100% para desenvolvimento
```

### Log Level Management
```go
// Configurar níveis de log por ambiente
if env == "production" {
    logger.SetLevel(LogLevelWarn)
} else {
    logger.SetLevel(LogLevelDebug)
}
```

## 🎯 Casos de Uso Empresariais

### 1. Financial Services
- Fraud detection error monitoring
- Transaction failure analysis
- Compliance audit logging
- Real-time alerting for critical failures

### 2. E-commerce Platform
- Payment processing error tracking
- Inventory error monitoring
- User experience impact analysis
- Peak traffic error handling

### 3. Healthcare Systems
- Patient data access error monitoring
- Medical device integration failures
- Compliance error tracking
- Critical system health monitoring

### 4. SaaS Applications
- Multi-tenant error isolation
- Feature usage error tracking
- Performance impact analysis
- Customer-specific SLA monitoring

## 🔍 Troubleshooting

### High Memory Usage
```go
// Configurar retention de métricas
aggregator.SetRetention("1m", 1000)  // Máximo 1000 erros por janela
aggregator.SetRetention("1h", 10000) // Máximo 10000 erros por hora
```

### Trace Sampling Issues
```go
// Debugging de sampling
tracer.EnableDebugLogging(true)
tracer.LogSamplingDecisions(true)
```

### Dashboard Performance
```go
// Otimizar queries do dashboard
dashboard.SetRefreshInterval(30 * time.Second)
dashboard.EnableCaching(true)
dashboard.SetMaxDataPoints(1000)
```

Este exemplo fornece uma base sólida para implementação de observabilidade enterprise com Domain Errors v2, seguindo as melhores práticas da indústria para monitoring, alerting e analysis.
