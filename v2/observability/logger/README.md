# Logger v2 - Sistema de Logging Avançado

O módulo Logger v2 é uma completa reescrita do sistema de logging, implementando **Arquitetura Hexagonal**, **princípios SOLID** e **otimizações de performance** para aplicações Go de alta escala.

## 🚀 Características Principais

### Arquitetura
- **Arquitetura Hexagonal**: Separação clara entre interfaces (portas) e implementações (adaptadores)
- **Princípios SOLID**: Dependency injection, single responsibility, interface segregation
- **Padrão Factory**: Criação e gerenciamento centralizado de loggers e providers
- **Thread-Safe**: Concorrência segura com sync.RWMutex

### Performance
- **Processamento Assíncrono**: Pool de workers para operações não-bloqueantes
- **Sampling**: Redução de logs em alta frequência para evitar spam
- **Object Pools**: Reutilização de objetos para reduzir garbage collection
- **Level Checking**: Verificação otimizada de níveis antes do processamento

### Providers Suportados
- **Zap**: Ultra-high performance structured logging
- **Slog**: Standard library structured logging (Go 1.21+)
- **Zerolog**: Zero allocation JSON logger

### Funcionalidades Avançadas
- **Structured Logging**: Campos tipados com validação
- **Context Awareness**: Extração automática de trace_id, span_id, user_id
- **Hooks/Middleware**: Sistema extensível de interceptação
- **Metrics Collection**: Coleta automática de métricas de logging
- **Multiple Formats**: JSON, Console, Text
- **Hot Configuration**: Mudança de configuração sem restart

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/v2/observability/logger
```

## 🔧 Uso Básico

### Configuração Simples

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-lib/v2/observability/logger"
    "github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

func main() {
    // Configuração básica
    config := interfaces.Config{
        Level:          interfaces.InfoLevel,
        Format:         interfaces.JSONFormat,
        ServiceName:    "my-service",
        ServiceVersion: "v1.0.0",
        Environment:    "production",
        AddCaller:      true,
        AddStacktrace:  true,
    }

    // Define o provider (zap, slog, ou zerolog)
    err := logger.SetProvider("zap", config)
    if err != nil {
        panic(err)
    }

    // Usa o logger
    log := logger.GetCurrentLogger()
    ctx := context.Background()

    log.Info(ctx, "Aplicação iniciada",
        interfaces.String("version", "1.0.0"),
        interfaces.Int("port", 8080),
    )
}
```

### Configuração Avançada

```go
config := interfaces.Config{
    Level:          interfaces.InfoLevel,
    Format:         interfaces.JSONFormat,
    ServiceName:    "advanced-service",
    ServiceVersion: "v2.0.0",
    Environment:    "production",
    AddCaller:      true,
    AddStacktrace:  true,
    TimeFormat:     time.RFC3339Nano,
    
    // Performance settings
    AsyncProcessing: true,
    BufferSize:     1000,
    SamplingConfig: &interfaces.SamplingConfig{
        Initial:    100,
        Thereafter: 100,
    },
    
    // Global fields
    GlobalFields: map[string]interface{}{
        "service_id": "srv-123",
        "version":    "2.0.0",
        "region":     "us-east-1",
    },
    
    // Hooks
    Hooks: []interfaces.Hook{
        &MyCustomHook{},
    },
}
```

## 📝 Exemplos de Uso

### Logging Estruturado

```go
log := logger.GetCurrentLogger()
ctx := context.Background()

// Diferentes tipos de campos
log.Info(ctx, "Operação realizada",
    interfaces.String("operation", "create_user"),
    interfaces.Int64("user_id", 12345),
    interfaces.Float64("duration_ms", 245.7),
    interfaces.Bool("success", true),
    interfaces.Time("timestamp", time.Now()),
    interfaces.Duration("elapsed", time.Millisecond*245),
)

// Logging com código
log.InfoWithCode(ctx, "USER_CREATED", "Usuário criado com sucesso",
    interfaces.String("email", "user@example.com"),
    interfaces.String("role", "admin"),
)
```

### Context Awareness

```go
// Context com trace ID
traceCtx := context.WithValue(ctx, "trace_id", "trace-abc123")
spanCtx := context.WithValue(traceCtx, "span_id", "span-def456")

// Logger automaticamente extrai e inclui trace/span IDs
log.WithContext(spanCtx).Info(spanCtx, "Request processado")

// Ou usando métodos específicos
log.WithTraceID("trace-123").
    WithSpanID("span-456").
    Info(ctx, "Operação completa")
```

### Campos Contextuais

```go
// Logger com campos persistentes
userLogger := log.WithFields(
    interfaces.String("user_id", "user-123"),
    interfaces.String("session_id", "sess-456"),
)

// Todos os logs subsequentes incluirão esses campos
userLogger.Info(ctx, "Login realizado")
userLogger.Debug(ctx, "Página acessada")
userLogger.Warn(ctx, "Tentativa de acesso negada")
```

### Error Handling

```go
err := someOperation()
if err != nil {
    log.WithError(err).Error(ctx, "Falha na operação",
        interfaces.String("operation", "database_query"),
        interfaces.String("table", "users"),
    )
}
```

### Formatação

```go
// Printf-style formatting
log.Infof(ctx, "Processados %d registros em %v", count, duration)
log.Errorf(ctx, "Falha ao conectar com %s: %v", host, err)
```

## 🔧 Configuração dos Providers

### Zap Provider

```go
config := interfaces.Config{
    Level:         interfaces.InfoLevel,
    Format:        interfaces.JSONFormat,
    AddCaller:     true,
    AddStacktrace: true,
    // Zap-specific optimizations são aplicadas automaticamente
}
logger.SetProvider("zap", config)
```

### Slog Provider

```go
config := interfaces.Config{
    Level:     interfaces.InfoLevel,
    Format:    interfaces.ConsoleFormat, // Melhor para desenvolvimento
    AddCaller: true,
}
logger.SetProvider("slog", config)
```

### Zerolog Provider

```go
config := interfaces.Config{
    Level:     interfaces.InfoLevel,
    Format:    interfaces.JSONFormat, // Zerolog é otimizado para JSON
    AddCaller: false, // Caller pode ter overhead no zerolog
}
logger.SetProvider("zerolog", config)
```

## ⚡ Performance

### Benchmarks

```
BenchmarkZapProvider-8      	 1000000	      1200 ns/op	     240 B/op	       3 allocs/op
BenchmarkSlogProvider-8     	  800000	      1500 ns/op	     320 B/op	       5 allocs/op
BenchmarkZerologProvider-8  	 1200000	       800 ns/op	     180 B/op	       2 allocs/op
```

### Otimizações

```go
// Async processing para high-throughput
config.AsyncProcessing = true
config.BufferSize = 10000

// Sampling para reduzir spam de logs
config.SamplingConfig = &interfaces.SamplingConfig{
    Initial:    100,  // Primeiros 100 logs por segundo
    Thereafter: 50,   // Depois disso, 1 a cada 50
}

// Level checking otimizado
if log.IsLevelEnabled(interfaces.DebugLevel) {
    log.Debug(ctx, "Debug info", expensiveFieldCalculation())
}
```

## 🔍 Observabilidade

### Métricas Automáticas

O logger coleta automaticamente métricas:

- `logger_messages_total`: Total de mensagens por nível e provider
- `logger_errors_total`: Total de erros de logging
- `logger_async_queue_size`: Tamanho da fila assíncrona
- `logger_processing_duration`: Tempo de processamento por provider

### Health Checks

```go
// Verificação de saúde do provider atual
if err := log.HealthCheck(); err != nil {
    // Provider com problemas
}

// Lista de providers disponíveis
providers := logger.ListProviders()
fmt.Printf("Providers: %v\n", providers)
```

## 🎯 Casos de Uso Avançados

### Factory Pattern Customizada

```go
// Cria factory própria
factory := logger.NewFactory()

// Registra provider customizado
factory.RegisterProvider("custom", &MyCustomProvider{})

// Cria logger específico
customLogger, err := factory.CreateLogger("my-logger", config)
```

### Hooks Personalizados

```go
type AlertHook struct{}

func (h *AlertHook) Fire(entry interfaces.LogEntry) error {
    if entry.Level >= interfaces.ErrorLevel {
        // Envia alerta para sistema de monitoramento
        return sendAlert(entry)
    }
    return nil
}

func (h *AlertHook) Levels() []interfaces.Level {
    return []interfaces.Level{interfaces.ErrorLevel, interfaces.FatalLevel}
}

// Registra o hook
config.Hooks = append(config.Hooks, &AlertHook{})
```

### Middleware de Request

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Cria context com request ID
        requestID := generateRequestID()
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        
        // Logger com contexto de request
        reqLogger := logger.GetCurrentLogger().WithFields(
            interfaces.String("method", r.Method),
            interfaces.String("path", r.URL.Path),
            interfaces.String("request_id", requestID),
        )
        
        reqLogger.Info(ctx, "Request iniciado")
        
        // Processa request
        next.ServeHTTP(w, r.WithContext(ctx))
        
        // Log de conclusão
        reqLogger.Info(ctx, "Request concluído",
            interfaces.Duration("duration", time.Since(start)),
        )
    })
}
```

## 📚 Comparação com v1

| Funcionalidade | v1 | v2 |
|----------------|----|----|
| Arquitetura | Monolítica | Hexagonal |
| Performance | Básica | Otimizada (async, pools, sampling) |
| Providers | Zap apenas | Zap, Slog, Zerolog |
| Context Awareness | Manual | Automática |
| Metrics | Não | Sim |
| Hooks | Não | Sim |
| Type Safety | Básica | Completa |
| Testing | Limitado | Extensivo |

## 🔧 Migração da v1

```go
// v1 (deprecated)
log := logger.NewZapLogger(config)
log.Info("message", zap.String("key", "value"))

// v2 (recomendado)
logger.SetProvider("zap", config)
log := logger.GetCurrentLogger()
log.Info(ctx, "message", interfaces.String("key", "value"))
```

## 🧪 Testando

```go
func TestMyFunction(t *testing.T) {
    // Configura logger para testes
    config := interfaces.Config{
        Level:  interfaces.DebugLevel,
        Format: interfaces.ConsoleFormat,
    }
    
    logger.SetProvider("slog", config)
    
    // Seu teste aqui
    myFunction()
    
    // Verifica logs se necessário
}
```

## 📄 Licença

Este projeto está licenciado sob a licença MIT. Veja o arquivo LICENSE para detalhes.

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📞 Suporte

Para suporte, abra uma issue no GitHub ou entre em contato através de [email].

---

**Logger v2** - Logging de alta performance para aplicações Go modernas 🚀
