# Sistema de Logging ISIS - Go

Um sistema de logging flexível e extensível para Go, com suporte a múltiplos providers (Zap, Zerolog, Slog) e logging estruturado moderno.

## Características

- 🔧 **Múltiplos Providers**: Suporte para Zap, Zerolog e Slog
- 📊 **Logging Estruturado**: Campos tipados e estruturados
- 🎯 **Context-Aware**: Extração automática de dados do contexto
- ⚡ **Alto Performance**: Otimizado para aplicações de alta performance
- 🔄 **Troca Dinâmica**: Mudança de provider em runtime
- 📱 **Múltiplos Formatos**: JSON, Console e Text
- 🏷️ **Sampling**: Controle de volume de logs para alta escala
- 🔍 **Tracing Integration**: Suporte nativo para trace_id, span_id
- 🛡️ **Type Safe**: Interface tipada para máxima segurança

## Instalação

```bash
go get github.com/fsvxavier/nexs-lib/logging
```

### Dependências Opcionais

Para usar providers específicos, adicione as dependências:

```bash
# Para Zap (já incluído)
go get go.uber.org/zap

# Para Zerolog
go get github.com/rs/zerolog

# Slog já está incluído no Go 1.21+
```

## Uso Básico

### Configuração Simples

```go
package main

import (
    "context"
    "os"
    
    "github.com/fsvxavier/nexs-lib/logging"
    _ "github.com/fsvxavier/nexs-lib/logging/providers/slog"
)

func main() {
    ctx := context.Background()
    
    // Configuração básica
    config := &logger.Config{
        Level:          logger.InfoLevel,
        Format:         logger.JSONFormat,
        Output:         os.Stdout,
        ServiceName:    "minha-aplicacao",
        ServiceVersion: "1.0.0",
        Environment:    "production",
    }
    
    // Define o provider
    err := logger.SetProvider("slog", config)
    if err != nil {
        panic(err)
    }
    
    // Logging básico
    logger.Info(ctx, "Aplicação iniciada",
        logger.String("status", "starting"),
        logger.Int("port", 8080),
    )
}
```

### Logging Estruturado

```go
// Diferentes tipos de campos
logger.Info(ctx, "Processando requisição",
    logger.String("method", "POST"),
    logger.String("path", "/api/users"),
    logger.Int("user_id", 12345),
    logger.Duration("duration", 150*time.Millisecond),
    logger.Bool("success", true),
    logger.Float64("score", 95.5),
)

// Logging com formatação
logger.Infof(ctx, "Usuário %s logou com sucesso em %v", username, time.Now())

// Logging com código de erro/evento
logger.ErrorWithCode(ctx, "USER_NOT_FOUND", "Usuário não encontrado",
    logger.String("user_id", "12345"),
    logger.String("operation", "login"),
)
```

### Context-Aware Logging

```go
// Adiciona informações ao contexto
ctx = context.WithValue(ctx, "trace_id", "trace-123")
ctx = context.WithValue(ctx, "user_id", "user-456")
ctx = context.WithValue(ctx, "request_id", "req-789")

// O logger extrai automaticamente esses valores
logger.WithContext(ctx).Info(ctx, "Operação executada")
// Output inclui automaticamente: trace_id, user_id, request_id
```

### Logger com Campos Persistentes

```go
// Cria logger com campos fixos
logger := logger.WithFields(
    logger.String("component", "database"),
    logger.String("module", "user-service"),
)

// Todos os logs deste logger incluirão os campos acima
logger.Info(ctx, "Conectando ao banco")
logger.Error(ctx, "Falha na conexão", logger.Error(err))
```

## Configuração Avançada

### Configuração Completa

```go
config := &logger.Config{
    Level:          logger.DebugLevel,
    Format:         logger.JSONFormat,
    Output:         os.Stdout,
    TimeFormat:     time.RFC3339Nano,
    ServiceName:    "minha-aplicacao",
    ServiceVersion: "2.1.0",
    Environment:    "production",
    AddSource:      true,  // Adiciona arquivo:linha nos logs
    AddStacktrace:  true,  // Adiciona stack trace em erros
    Fields: map[string]any{
        "datacenter": "us-east-1",
        "instance":   "web-01",
    },
    SamplingConfig: &logger.SamplingConfig{
        Initial:    100,  // Primeiros 100 logs passam
        Thereafter: 10,   // Depois, 1 a cada 10
        Tick:       time.Second,
    },
}
```

### Múltiplos Providers

```go
// Configuração para desenvolvimento (console colorido)
devConfig := &logger.Config{
    Level:  logger.DebugLevel,
    Format: logger.ConsoleFormat,
    Output: os.Stdout,
}

// Configuração para produção (JSON estruturado)
prodConfig := &logger.Config{
    Level:         logger.InfoLevel,
    Format:        logger.JSONFormat,
    Output:        logFile,
    AddSource:     false,
    AddStacktrace: true,
}

// Muda provider baseado no ambiente
if os.Getenv("ENV") == "development" {
    logger.SetProvider("zerolog", devConfig)
} else {
    logger.SetProvider("zap", prodConfig)
}
```

## Providers Disponíveis

### Slog (Padrão Go 1.21+)

```go
import _ "github.com/fsvxavier/nexs-lib/logging/providers/slog"

logger.SetProvider("slog", config)
```

**Características:**
- Parte do Go standard library
- Performance excelente
- Suporte nativo a structured logging
- Handlers customizáveis

### Zap (Uber)

```go
import _ "github.com/fsvxavier/nexs-lib/logging/providers/zap"

logger.SetProvider("zap", config)
```

**Características:**
- Performance superior
- Zero-allocation em hot paths
- Sampling avançado
- Suporte completo a structured logging

### Zerolog

```go
import _ "github.com/fsvxavier/nexs-lib/logging/providers/zerolog"

logger.SetProvider("zerolog", config)
```

**Características:**
- JSON-first design
- Zero-allocation
- Performance excelente
- API fluent

## Níveis de Log

```go
const (
    DebugLevel Level = iota - 1
    InfoLevel
    WarnLevel
    ErrorLevel
    FatalLevel  // Termina a aplicação
    PanicLevel  // Causa panic
)
```

## Formatos de Saída

### JSON (Produção)

```go
config.Format = logger.JSONFormat
```

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "info",
  "message": "Usuário logou",
  "service": "auth-service",
  "trace_id": "trace-123",
  "user_id": "user-456",
  "method": "POST"
}
```

### Console (Desenvolvimento)

```go
config.Format = logger.ConsoleFormat
```

```
2024-01-15T10:30:45Z INF Usuário logou service=auth-service trace_id=trace-123 user_id=user-456 method=POST
```

### Text (Simples)

```go
config.Format = logger.TextFormat
```

```
timestamp=2024-01-15T10:30:45Z level=info msg="Usuário logou" service=auth-service trace_id=trace-123
```

## Performance e Sampling

Para aplicações de alto volume, use sampling:

```go
config.SamplingConfig = &logger.SamplingConfig{
    Initial:    1000,           // Primeiros 1000 logs passam
    Thereafter: 100,            // Depois, 1 a cada 100
    Tick:       time.Second,    // Por segundo
}
```

## Integração com Tracing

O sistema extrai automaticamente do contexto:

- `trace_id`: ID do trace distribuído
- `span_id`: ID do span atual  
- `user_id`: ID do usuário
- `request_id`: ID da requisição

```go
// Em middleware HTTP
ctx = context.WithValue(ctx, "trace_id", traceID)
ctx = context.WithValue(ctx, "request_id", requestID)

// Em handlers
logger.Info(ctx, "Processando requisição") // Inclui automaticamente trace_id e request_id
```

## Melhores Práticas

### 1. Use Campos Estruturados

```go
// ✅ Bom
logger.Info(ctx, "Usuário criado",
    logger.String("user_id", userID),
    logger.String("email", email),
    logger.Duration("duration", elapsed),
)

// ❌ Evite
logger.Infof(ctx, "Usuário %s criado com email %s em %v", userID, email, elapsed)
```

### 2. Use Níveis Apropriados

```go
// Debug: Informações detalhadas para debugging
logger.Debug(ctx, "Executando query SQL", logger.String("query", sql))

// Info: Eventos importantes do negócio
logger.Info(ctx, "Usuário criado", logger.String("user_id", id))

// Warn: Situações que precisam atenção mas não são erros
logger.Warn(ctx, "Rate limit próximo", logger.Int("requests", count))

// Error: Erros que precisam investigação
logger.Error(ctx, "Falha ao conectar", logger.Error(err))
```

### 3. Use Context Adequadamente

```go
// ✅ Passe contexto com informações relevantes
ctx = context.WithValue(ctx, "user_id", userID)
logger := logger.WithContext(ctx)
logger.Info(ctx, "Operação realizada")

// ✅ Use campos persistentes para componentes
dbLogger := logger.WithFields(
    logger.String("component", "database"),
)
```

### 4. Gerencie Performance

```go
// ✅ Use sampling em logs de alto volume
if logger.GetCurrentProvider().GetLevel() <= logger.DebugLevel {
    logger.Debug(ctx, "Debug info", expensiveField())
}

// ✅ Use defer para logs de duração
start := time.Now()
defer func() {
    logger.Info(ctx, "Operação concluída",
        logger.Duration("duration", time.Since(start)),
    )
}()
```

## Testes

```bash
# Executa todos os testes
go test ./logging/...

# Testes com coverage
go test -cover ./logging/...

# Benchmarks
go test -bench=. ./logging/...
```

## Exemplos Completos

Veja a pasta `examples/` para exemplos completos de uso com diferentes providers e configurações.

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para detalhes.
