# Nexs Observability Library

Biblioteca completa de observabilidade para aplicações Go, fornecendo instrumentação padronizada para **logging**, **tracing** e **metrics** com suporte a múltiplos providers e infraestrutura Docker para desenvolvimento e testes.

## 🎯 Visão Geral

A Nexs Observability Library oferece uma solução unificada para instrumentação de aplicações, permitindo:

- **Logging estruturado** com múltiplos backends
- **Distributed tracing** com providers populares
- **Infraestrutura completa** para desenvolvimento e testes
- **Integração simples** e APIs consistentes
- **Performance otimizada** com baixo overhead

## 📦 Componentes

### 🔍 Logger
Sistema de logging estruturado com suporte a múltiplos providers.

**Providers Suportados:**
- **Zap**: High-performance structured logging
- **Logrus**: Feature-rich logging library
- **Slog**: Go's native structured logging (1.21+)

### 📊 Tracer  
Sistema de distributed tracing com integração a plataformas populares.

**Providers Suportados:**
- **Datadog**: APM integration completa
- **Grafana Tempo**: Backend de tracing open-source
- **New Relic**: Full observability platform
- **OpenTelemetry**: Vendor-neutral tracing

### 🐳 Infrastructure
Stack completa Docker para desenvolvimento e testes com:
- **Tracing**: Jaeger, Tempo, OpenTelemetry Collector
- **Logging**: Elasticsearch, Logstash, Fluentd, Kibana
- **Metrics**: Prometheus, Grafana
- **Databases**: PostgreSQL, MongoDB, Redis, RabbitMQ

## 🚀 Quick Start

### Instalação

```bash
go get github.com/fsvxavier/nexs-lib/observability
```

### Logger Básico

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-lib/observability/logger"
    "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
)

func main() {
    // Configurar logger
    config := zap.Config{
        Level:    "info",
        Format:   "json",
        AddStack: true,
    }
    
    provider := zap.NewZapProvider(config)
    log := logger.NewLogger(provider)
    
    // Usar logger
    ctx := context.Background()
    log.Info(ctx, "Aplicação iniciada", 
        logger.String("service", "api"),
        logger.Int("port", 8080),
    )
}
```

### Tracer Básico

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-lib/observability/tracer"
    "github.com/fsvxavier/nexs-lib/observability/tracer/providers/opentelemetry"
)

func main() {
    // Configurar tracer
    config := opentelemetry.Config{
        ServiceName: "my-service",
        Endpoint:    "http://localhost:4317",
        Environment: "development",
    }
    
    provider := opentelemetry.NewOpenTelemetryProvider(config)
    tracer := tracer.NewTracer(provider)
    
    // Criar span
    ctx := context.Background()
    span := tracer.StartSpan(ctx, "operation-name")
    defer span.End()
    
    span.SetTag("user.id", "12345")
    span.SetTag("operation.type", "database-query")
}
```

### Infraestrutura de Desenvolvimento

```bash
# Clonar repositório
git clone <repository-url>
cd nexs-lib/observability/infrastructure

# Iniciar stack completa
make infra-up

# Verificar status
make infra-status

# Acessar UIs
# Jaeger: http://localhost:16686
# Grafana: http://localhost:3000 (admin/nexs123)  
# Kibana: http://localhost:5601
```

## 📚 Documentação Detalhada

### Logger
- [📖 Logger README](./logger/README.md) - Documentação completa
- [🔧 Logger Providers](./logger/providers/) - Configuração de providers
- [💡 Logger Examples](./logger/examples/) - Exemplos práticos

### Tracer
- [📖 Tracer README](./tracer/README.md) - Documentação completa
- [🔧 Tracer Providers](./tracer/providers/) - Configuração de providers  
- [💡 Tracer Examples](./tracer/examples/) - 6 exemplos completos

### Infrastructure
- [📖 Infrastructure README](./infrastructure/README.md) - Setup completo
- [🔧 Infrastructure Config](./infrastructure/configs/) - Configurações
- [📋 Infrastructure Next Steps](./infrastructure/NEXT_STEPS.md) - Roadmap

## 🏗️ Arquitetura

### Padrão de Design

```
Application Code
       ↓
   Nexs Logger/Tracer (Interface)
       ↓
   Provider Abstraction
       ↓
   Backend Implementation
   (Zap, Datadog, etc.)
```

### Principais Vantagens

- **Vendor Independence**: Troque providers sem alterar código
- **Consistent API**: Interface unificada para todos os providers
- **Type Safety**: APIs type-safe com Go generics
- **Performance**: Overhead mínimo e otimizações por provider
- **Testing**: Mocks centralizados para todos os componentes

## 📊 Comparação de Providers

### Logger Providers

| Provider | Performance | Features | Memory | Use Case |
|----------|-------------|----------|---------|----------|
| **Zap** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | High-performance APIs |
| **Logrus** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Feature-rich applications |
| **Slog** | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | Go 1.21+ applications |

### Tracer Providers

| Provider | Setup | Features | Cost | Use Case |
|----------|-------|----------|------|----------|
| **Datadog** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 💰💰💰 | Enterprise APM |
| **Grafana** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 💰 | Open-source stack |
| **New Relic** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 💰💰 | Full observability |
| **OpenTelemetry** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Free | Vendor-neutral |

## 🔧 Configuração Avançada

### Logger com Context Propagation

```go
// Logger com correlation IDs
ctx := logger.WithFields(context.Background(), 
    logger.String("trace_id", "abc123"),
    logger.String("user_id", "user456"),
)

log.Info(ctx, "Processing request")
// Output: {"level":"info","trace_id":"abc123","user_id":"user456","msg":"Processing request"}
```

### Tracer com Custom Tags

```go
// Tracer com tags personalizados
span := tracer.StartSpan(ctx, "database-query")
span.SetTag("db.table", "users")
span.SetTag("db.operation", "SELECT")
span.SetTag("query.rows", 150)

// Correlação com logs
logger.Info(ctx, "Query executed", 
    logger.String("span_id", span.ID()),
    logger.Duration("duration", span.Duration()),
)
```

### Configuration Management

```go
// Configuração centralizada
type ObservabilityConfig struct {
    Logger LoggerConfig `yaml:"logger"`
    Tracer TracerConfig `yaml:"tracer"`
}

// Carregar de arquivo
config, err := LoadConfigFromFile("observability.yaml")
if err != nil {
    log.Fatal("Failed to load config", err)
}
```

## 🧪 Testing

### Unit Tests com Mocks

```go
func TestUserService(t *testing.T) {
    // Setup mocks
    mockLogger := logger_mocks.NewMockProvider()
    mockTracer := tracer_mocks.NewMockProvider()
    
    logger := logger.NewLogger(mockLogger)
    tracer := tracer.NewTracer(mockTracer)
    
    service := NewUserService(logger, tracer)
    
    // Test
    user, err := service.GetUser(ctx, "123")
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, "123", user.ID)
    
    // Verify observability calls
    mockLogger.AssertLogCalled(t, "info", "User retrieved")
    mockTracer.AssertSpanStarted(t, "user-service.get-user")
}
```

### Integration Tests

```bash
# Usar infraestrutura real para testes
make infra-up
make test-integration

# Ou grupos específicos
make dev-tracer
make test-integration-tracer
```

## 🚀 Performance

### Benchmarks

```bash
# Logger performance
go test -bench=BenchmarkLogger -benchmem ./logger/...

# Tracer performance  
go test -bench=BenchmarkTracer -benchmem ./tracer/...

# Com infraestrutura real
make perf-test
```

### Otimizações

- **Zero-allocation logging** com Zap provider
- **Sampling strategies** para high-throughput
- **Async processing** para operações I/O
- **Connection pooling** para backends remotos
- **Batch processing** para múltiplas operações

## 🔐 Segurança

### Dados Sensíveis

```go
// Sanitização automática
logger.Info(ctx, "User authenticated",
    logger.String("user_id", userID),
    logger.Redacted("password", password), // Será mostrado como [REDACTED]
    logger.PII("email", email),           // Será hasheado em prod
)
```

### Network Security

```go
// TLS para backends remotos
config := datadog.Config{
    APIKey:  os.Getenv("DD_API_KEY"),
    UseTLS:  true,
    Timeout: 10 * time.Second,
}
```

## 📈 Production Readiness

### Health Checks

```go
// Health check para observability
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Verificar logger
    if !s.logger.IsHealthy(ctx) {
        http.Error(w, "Logger unhealthy", http.StatusServiceUnavailable)
        return
    }
    
    // Verificar tracer
    if !s.tracer.IsHealthy(ctx) {
        http.Error(w, "Tracer unhealthy", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
```

### Graceful Shutdown

```go
// Shutdown com flush de buffers
func (s *Server) Shutdown(ctx context.Context) error {
    // Flush tracer buffers
    if err := s.tracer.Shutdown(ctx); err != nil {
        s.logger.Error(ctx, "Failed to shutdown tracer", logger.Error(err))
    }
    
    // Flush logger buffers  
    if err := s.logger.Shutdown(ctx); err != nil {
        fmt.Printf("Failed to shutdown logger: %v\n", err)
    }
    
    return nil
}
```

## 🤝 Contribuição

### Development Setup

```bash
# Setup ambiente de desenvolvimento
git clone <repository-url>
cd nexs-lib/observability

# Iniciar infraestrutura
make dev-setup

# Executar testes
make test-integration

# Validar código
go fmt ./...
go vet ./...
golangci-lint run
```

### Adicionando Novos Providers

1. **Implementar interface do provider**
2. **Adicionar testes unitários com mocks**
3. **Criar exemplo prático**
4. **Documentar configurações**
5. **Integrar com CI/CD**

### Code Style

- Seguir [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Usar `gofmt` e `goimports`
- Documentar APIs públicas
- Manter coverage > 90%

## 📋 Roadmap

### Próximas Releases

#### v1.1.0 - Metrics Integration
- [ ] Prometheus metrics provider
- [ ] Custom metrics collection
- [ ] Dashboard templates
- [ ] Performance monitoring

#### v1.2.0 - Advanced Features  
- [ ] Correlation automática trace/log
- [ ] Context propagation aprimorado
- [ ] Sampling strategies avançadas
- [ ] Error tracking integration

#### v1.3.0 - Cloud Native
- [ ] Kubernetes integration
- [ ] Service mesh support
- [ ] Cloud provider optimizations
- [ ] Auto-scaling integration

## 📞 Suporte

### Recursos
- [📖 Documentação](./NEXT_STEPS.md)
- [💡 Exemplos](./tracer/examples/)
- [🐛 Issues](https://github.com/fsvxavier/nexs-lib/issues)
- [💬 Discussions](https://github.com/fsvxavier/nexs-lib/discussions)

### FAQ

**Q: Posso usar múltiplos providers simultaneamente?**
A: Sim, você pode configurar diferentes providers para diferentes serviços ou ambientes.

**Q: Como migrar de outras bibliotecas?**
A: Consulte nossos [guias de migração](./docs/migration/) para bibliotecas populares.

**Q: Qual o overhead de performance?**
A: < 1ms de latência adicional em 99% dos casos. Veja nossos [benchmarks](./docs/performance/).

## 📝 Licença

Este projeto está licenciado sob a [MIT License](../LICENSE).

---

## 🎯 Status do Projeto

- **Logger**: ✅ Production ready
- **Tracer**: ✅ Production ready  
- **Infrastructure**: ✅ Development ready
- **Documentation**: ✅ Complete
- **Examples**: ✅ Complete
- **Tests**: ✅ Complete

**Current Status**: 🚀 **Ready for production use with comprehensive development infrastructure**
