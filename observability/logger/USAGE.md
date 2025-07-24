# Logger Usage Guide

## Overview

Este guia fornece uma visão completa do sistema de logging multi-provider, incluindo configuração, uso e exemplos práticos.

## Características Principais

- **Multi-Provider**: Suporte para slog, zap e zerolog
- **Provider Padrão**: Zap configurado automaticamente como padrão
- **Context-Aware**: Extração automática de trace_id, span_id, user_id, request_id
- **Structured Logging**: Campos tipados e estruturados
- **Performance**: Benchmarks indicam zap como o mais rápido (~240k logs/sec)
- **Auto-Registration**: Providers são registrados automaticamente via imports

## Instalação

```bash
go mod tidy
```

## Uso Básico

### Usando o Provider Padrão (Zap)

```go
package main

import (
    "context"
    
    "github.com/fsvxavier/nexs-lib/observability/logger"
    
    // Auto-registration dos providers
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
    ctx := context.Background()
    
    // Zap é configurado automaticamente como padrão
    logger.Info(ctx, "Aplicação iniciada")
    logger.Error(ctx, "Erro de exemplo", 
        logger.String("error", "não encontrado"))
    
    // Logging com contexto
    ctx = context.WithValue(ctx, "user_id", "user123")
    logger.Info(ctx, "Usuário logado")
}
```

### Configurando Providers Específicos

```go
package main

import (
    "context"
    
    "github.com/fsvxavier/nexs-lib/observability/logger"
    
    // Importa os providers
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
    ctx := context.Background()
    
    // Usar slog
    logger.ConfigureProvider("slog", nil)
    logger.Info(ctx, "Usando slog")
    
    // Usar zerolog
    logger.ConfigureProvider("zerolog", nil)
    logger.Info(ctx, "Usando zerolog")
    
    // Voltar para zap (padrão)
    logger.ConfigureProvider("zap", nil)
    logger.Info(ctx, "Usando zap")
}
```

## Exemplos Disponíveis

### 1. Default Provider (`examples/default-provider/`)
Demonstra o uso do provider padrão sem configuração explícita.

```bash
cd examples/default-provider
go run main.go
```

### 2. Basic Usage (`examples/basic/`)
Demonstra funcionalidades básicas com todos os providers.

```bash
cd examples/basic
go run main.go
```

### 3. Advanced Usage (`examples/advanced/`)
Demonstra cenários avançados com serviços e middleware.

```bash
cd examples/advanced
go run main.go
```

### 4. Multi-Provider (`examples/multi-provider/`)
Demonstra comparação e uso de todos os providers.

```bash
cd examples/multi-provider
go run main.go
```

### 5. Benchmark (`examples/benchmark/`)
Testa performance de todos os providers.

```bash
cd examples/benchmark
go run main.go
```

## Tipos de Campos Suportados

```go
// Campos básicos
logger.Info(ctx, "Mensagem",
    logger.String("nome", "João"),
    logger.Int("idade", 30),
    logger.Int64("timestamp", 1642248645),
    logger.Float64("altura", 1.75),
    logger.Bool("ativo", true),
)

// Campos de duração
logger.Info(ctx, "Operação completada",
    logger.Duration("tempo", time.Second*2),
)

// Campos de tempo
logger.Info(ctx, "Evento",
    logger.Time("timestamp", time.Now()),
)

// Campos de grupo
logger.Info(ctx, "Dados do usuário",
    logger.Group("usuario",
        logger.String("nome", "João"),
        logger.String("email", "joao@email.com"),
    ),
)

// Campos de erro
logger.Error(ctx, "Erro na operação",
    logger.String("error", err.Error()),
    logger.String("error_type", reflect.TypeOf(err).String()),
)
```

## Context-Aware Logging

O sistema extrai automaticamente informações do contexto:

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "trace_id", "abc123")
ctx = context.WithValue(ctx, "span_id", "xyz789")
ctx = context.WithValue(ctx, "user_id", "user456")
ctx = context.WithValue(ctx, "request_id", "req789")

logger.Info(ctx, "Operação executada")
// Output inclui automaticamente: trace_id, span_id, user_id, request_id
```

### Exemplo de Output com Contexto

#### Zap (Padrão)
```json
{
  "level": "info",
  "time": "2025-07-18T10:30:45Z",
  "trace_id": "abc123",
  "span_id": "xyz789",
  "user_id": "user456",
  "request_id": "req789",
  "msg": "Operação executada"
}
```

#### Slog
```json
{
  "time": "2025-07-18T10:30:45Z",
  "level": "INFO",
  "trace_id": "abc123",
  "span_id": "xyz789",
  "user_id": "user456",
  "request_id": "req789",
  "msg": "Operação executada"
}
```

#### Zerolog
```json
{
  "level": "info",
  "time": "2025-07-18T10:30:45Z",
  "trace_id": "abc123",
  "span_id": "xyz789",
  "user_id": "user456",
  "request_id": "req789",
  "message": "Operação executada"
}
```

## Níveis de Log

```go
logger.Debug(ctx, "Informação de debug")
logger.Info(ctx, "Informação geral")
logger.Warn(ctx, "Aviso importante")
logger.Error(ctx, "Erro ocorrido")
```

## Performance

Baseado nos benchmarks:

| Provider | Logs/segundo | Uso de Memória | CPU Usage | Recomendação |
|----------|-------------|----------------|-----------|--------------|
| Zap      | ~240k       | 145 MB         | 12%       | Padrão (alta performance) |
| Zerolog  | ~174k       | 98 MB          | 8%        | Aplicações com restrições de memória |
| Slog     | ~132k       | 167 MB         | 15%       | Compatibilidade com stdlib |

## Testando

Execute todos os exemplos:

```bash
# Usando Makefile
make test

# Ou script direto
bash test_examples.sh
```

## Configuração Avançada

### Personalizando o Provider Zap

```go
config := map[string]interface{}{
    "level":        "info",
    "format":       "json",
    "development":  false,
    "sampling":     true,
    "outputPaths":  []string{"stdout"},
    "errorOutputPaths": []string{"stderr"},
    "encoderConfig": map[string]interface{}{
        "timeKey":     "time",
        "levelKey":    "level",
        "nameKey":     "logger",
        "callerKey":   "caller",
        "messageKey":  "msg",
        "stacktraceKey": "stacktrace",
        "levelEncoder": "lowercase",
        "timeEncoder":  "iso8601",
        "callerEncoder": "short",
    },
}

logger.ConfigureProvider("zap", config)
```

### Personalizando o Provider Zerolog

```go
config := map[string]interface{}{
    "level":       "info",
    "pretty":      false,
    "timestamp":   true,
    "caller":      false,
    "sampling":    true,
    "timeFormat":  time.RFC3339,
}

logger.ConfigureProvider("zerolog", config)
```

### Personalizando o Provider Slog

```go
config := map[string]interface{}{
    "level":      "info",
    "format":     "json",
    "addSource":  false,
    "replaceAttr": nil,
}

logger.ConfigureProvider("slog", config)
```

### Verificando o Provider Atual

```go
providerName := logger.GetCurrentProviderName()
fmt.Printf("Provider atual: %s\n", providerName)

// Lista todos os providers disponíveis
providers := logger.ListProviders()
fmt.Printf("Providers disponíveis: %v\n", providers)
```

## Integração com Middleware

### HTTP Middleware

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Enriquece contexto
        ctx := r.Context()
        ctx = context.WithValue(ctx, "request_id", generateRequestID())
        ctx = context.WithValue(ctx, "trace_id", generateTraceID())
        
        // Log de início
        logger.Info(ctx, "Request iniciada",
            logger.String("method", r.Method),
            logger.String("path", r.URL.Path),
            logger.String("remote_addr", r.RemoteAddr),
            logger.String("user_agent", r.UserAgent()),
        )
        
        // Executa próximo handler
        next.ServeHTTP(w, r.WithContext(ctx))
        
        // Log de fim
        duration := time.Since(start)
        logger.Info(ctx, "Request finalizada",
            logger.Duration("duration", duration),
            logger.Int("status", 200), // Ou capturar status real
        )
    })
}
```

### Middleware de Serviço

```go
type LoggingService struct {
    next   UserService
    logger logger.Logger
}

func NewLoggingService(next UserService, logger logger.Logger) *LoggingService {
    return &LoggingService{
        next:   next,
        logger: logger,
    }
}

func (s *LoggingService) GetUser(ctx context.Context, userID string) (*User, error) {
    start := time.Now()
    
    s.logger.Info(ctx, "Iniciando operação",
        logger.String("operation", "get_user"),
        logger.String("user_id", userID),
    )
    
    user, err := s.next.GetUser(ctx, userID)
    duration := time.Since(start)
    
    if err != nil {
        s.logger.Error(ctx, "Erro na operação",
            logger.String("operation", "get_user"),
            logger.String("user_id", userID),
            logger.String("error", err.Error()),
            logger.Duration("duration", duration),
        )
        return nil, err
    }
    
    s.logger.Info(ctx, "Operação concluída",
        logger.String("operation", "get_user"),
        logger.String("user_id", userID),
        logger.Duration("duration", duration),
    )
    
    return user, nil
}
```

## Cenários de Uso

### 1. Aplicação Web de Alta Performance

```go
// Use Zap para máxima performance
logger.ConfigureProvider("zap", map[string]interface{}{
    "level":      "info",
    "sampling":   true,
    "development": false,
})
```

### 2. Aplicação com Restrições de Memória

```go
// Use Zerolog para menor consumo de memória
logger.ConfigureProvider("zerolog", map[string]interface{}{
    "level":   "info",
    "pretty":  false,
    "caller":  false,
})
```

### 3. Aplicação Simples ou Migração

```go
// Use Slog para compatibilidade com stdlib
logger.ConfigureProvider("slog", map[string]interface{}{
    "level":     "info",
    "addSource": false,
})
```

## Switching Dinâmico de Providers

```go
func switchProvider(env string) {
    switch env {
    case "development":
        logger.ConfigureProvider("slog", map[string]interface{}{
            "level":     "debug",
            "addSource": true,
        })
    case "production":
        logger.ConfigureProvider("zap", map[string]interface{}{
            "level":     "info",
            "sampling":  true,
        })
    case "embedded":
        logger.ConfigureProvider("zerolog", map[string]interface{}{
            "level":  "warn",
            "pretty": false,
        })
    }
}
```

## Melhores Práticas

### 1. Configuração de Ambiente

```go
// Development
logger.ConfigureProvider("slog", map[string]interface{}{
    "level":     "debug",
    "addSource": true,
    "format":    "text",
})

// Production
logger.ConfigureProvider("zap", map[string]interface{}{
    "level":     "info",
    "sampling":  true,
    "format":    "json",
})
```

### 2. Structured Logging

```go
// ❌ Não faça isso
logger.Info(ctx, fmt.Sprintf("Usuário %s criado com ID %d", name, id))

// ✅ Faça isso
logger.Info(ctx, "Usuário criado",
    logger.String("name", name),
    logger.Int("id", id),
)
```

### 3. Context Propagation

```go
// ❌ Não faça isso
logger.Info(context.Background(), "Operação executada")

// ✅ Faça isso
logger.Info(ctx, "Operação executada")
```

### 4. Error Handling

```go
// ❌ Não faça isso
logger.Error(ctx, err.Error())

// ✅ Faça isso
logger.Error(ctx, "Erro na operação",
    logger.String("error", err.Error()),
    logger.String("operation", "create_user"),
    logger.String("user_id", userID),
)
```

### 5. Performance Monitoring

```go
func monitorPerformance(ctx context.Context, operation string, fn func() error) error {
    start := time.Now()
    
    logger.Debug(ctx, "Iniciando operação",
        logger.String("operation", operation),
    )
    
    err := fn()
    duration := time.Since(start)
    
    if err != nil {
        logger.Error(ctx, "Erro na operação",
            logger.String("operation", operation),
            logger.String("error", err.Error()),
            logger.Duration("duration", duration),
        )
        return err
    }
    
    logger.Info(ctx, "Operação concluída",
        logger.String("operation", operation),
        logger.Duration("duration", duration),
    )
    
    return nil
}
```

## Troubleshooting

### Provider não encontrado
```
Error: provider "xyz" not found
```

**Solução:**
```go
// Verifique se o provider está registrado
providers := logger.ListProviders()
fmt.Printf("Providers disponíveis: %v\n", providers)

// Certifique-se de importar o provider
import _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
```

### Erro de configuração
```
Error: failed to configure provider
```

**Solução:**
```go
// Verifique se a configuração está correta
config := map[string]interface{}{
    "level": "info", // string, não int
    "format": "json", // formato válido
}
logger.ConfigureProvider("zap", config)
```

### Logs não aparecem

**Solução:**
```go
// Verifique o nível de log
logger.ConfigureProvider("zap", map[string]interface{}{
    "level": "debug", // Diminua o nível
})

// Teste com log simples
logger.Info(context.Background(), "Teste")
```

### Performance degradada

**Solução:**
```bash
# Execute benchmark
go run examples/benchmark/main.go

# Ou use Makefile
make benchmark
```

## Automação com Makefile

```bash
# Testa todos os exemplos
make test

# Executa exemplo específico
make default        # Provider padrão
make basic          # Exemplo básico
make advanced       # Exemplo avançado
make multi-provider # Comparação de providers
make benchmark      # Análise de performance

# Verificação completa
make check          # fmt + vet + test
```

## Monitoramento e Métricas

### Coleta de Métricas

```go
type LoggingMetrics struct {
    LogsPerSecond    float64
    AvgLatency       time.Duration
    ErrorRate        float64
    MemoryUsage      uint64
}

func collectMetrics(ctx context.Context, provider string) *LoggingMetrics {
    start := time.Now()
    
    // Simula logging
    for i := 0; i < 1000; i++ {
        logger.Info(ctx, "Metric test", logger.Int("iteration", i))
    }
    
    duration := time.Since(start)
    
    return &LoggingMetrics{
        LogsPerSecond: float64(1000) / duration.Seconds(),
        AvgLatency:    duration / 1000,
        MemoryUsage:   getMemoryUsage(),
    }
}
```

### Alertas

```go
func checkPerformance(ctx context.Context, metrics *LoggingMetrics) {
    if metrics.LogsPerSecond < 10000 {
        logger.Warn(ctx, "Performance degradada",
            logger.Float64("logs_per_second", metrics.LogsPerSecond),
            logger.String("provider", logger.GetCurrentProviderName()),
        )
    }
    
    if metrics.MemoryUsage > 500*1024*1024 { // 500MB
        logger.Error(ctx, "Uso excessivo de memória",
            logger.Uint64("memory_mb", metrics.MemoryUsage/1024/1024),
        )
    }
}
```

## Integração com Observabilidade

### Trace ID Propagation

```go
func generateTraceID() string {
    return fmt.Sprintf("trace-%d", time.Now().UnixNano())
}

func enrichContext(ctx context.Context, r *http.Request) context.Context {
    traceID := r.Header.Get("X-Trace-ID")
    if traceID == "" {
        traceID = generateTraceID()
    }
    
    ctx = context.WithValue(ctx, "trace_id", traceID)
    ctx = context.WithValue(ctx, "request_id", generateRequestID())
    
    return ctx
}
```

### Structured Error Logging

```go
func logError(ctx context.Context, err error, operation string) {
    errorData := map[string]interface{}{
        "error":     err.Error(),
        "operation": operation,
        "timestamp": time.Now(),
    }
    
    // Adiciona stack trace se disponível
    if stackErr, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
        errorData["stack_trace"] = fmt.Sprintf("%+v", stackErr.StackTrace())
    }
    
    logger.Error(ctx, "Erro na aplicação",
        logger.String("error", err.Error()),
        logger.String("operation", operation),
        logger.String("error_type", reflect.TypeOf(err).String()),
    )
}
```

## Contribuindo

### Adicionando um Novo Provider

1. **Implemente a interface Provider**:
```go
type MyProvider struct{}

func (p *MyProvider) Log(ctx context.Context, level Level, msg string, fields ...Field) {
    // Implementação específica
}

func (p *MyProvider) Configure(config map[string]interface{}) error {
    // Configuração específica
    return nil
}
```

2. **Registre o provider**:
```go
func init() {
    manager.RegisterProvider("myprovider", &MyProvider{})
}
```

3. **Adicione testes**:
```go
func TestMyProvider(t *testing.T) {
    // Testes específicos
}
```

4. **Atualize documentação**:
- Adicione exemplo em `examples/`
- Atualize README.md e USAGE.md
- Adicione benchmark

### Contribuindo com Melhorias

1. **Fork o repositório**
2. **Crie uma branch** para sua feature
3. **Adicione testes** para suas mudanças
4. **Execute os testes**: `make test`
5. **Submeta um Pull Request**

## Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.
