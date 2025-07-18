# Logger Usage Guide

## Overview

Este guia fornece uma visão completa do sistema de logging multi-provider, incluindo configuração, uso e exemplos práticos.

## Características Principais

- **Multi-Provider**: Suporte para slog, zap e zerolog
- **Provider Padrão**: Zap configurado automaticamente como padrão
- **Context-Aware**: Extração automática de trace_id, span_id, user_id, request_id
- **Structured Logging**: Campos tipados e estruturados
- **Performance**: Benchmarks indicam zap como o mais rápido (~240k logs/sec)

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
    "log/slog"
    
    "github.com/fsvxavier/nexs-lib/observability/logger"
)

func main() {
    // Zap é configurado automaticamente como padrão
    logger := logger.NewLogger()
    
    // Logging básico
    logger.Info("Aplicação iniciada")
    logger.Error("Erro de exemplo", slog.String("error", "não encontrado"))
    
    // Logging com contexto
    ctx := context.WithValue(context.Background(), "user_id", "user123")
    logger.InfoContext(ctx, "Usuário logado")
}
```

### Configurando Providers Específicos

```go
package main

import (
    "github.com/fsvxavier/nexs-lib/observability/logger"
)

func main() {
    // Configurar provider específico
    logger := logger.NewLogger()
    
    // Usar slog
    logger.ConfigureProvider("slog", nil)
    logger.Info("Usando slog")
    
    // Usar zerolog
    logger.ConfigureProvider("zerolog", nil)
    logger.Info("Usando zerolog")
    
    // Voltar para zap (padrão)
    logger.ConfigureProvider("zap", nil)
    logger.Info("Usando zap")
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
logger.Info("Mensagem",
    slog.String("nome", "João"),
    slog.Int("idade", 30),
    slog.Float64("altura", 1.75),
    slog.Bool("ativo", true),
)

// Campos de duração
logger.Info("Operação completada",
    slog.Duration("tempo", time.Second*2),
)

// Campos de tempo
logger.Info("Evento",
    slog.Time("timestamp", time.Now()),
)

// Campos de grupo
logger.Info("Dados do usuário",
    slog.Group("usuario",
        slog.String("nome", "João"),
        slog.String("email", "joao@email.com"),
    ),
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

logger.InfoContext(ctx, "Operação executada")
// Output inclui automaticamente: trace_id, span_id, user_id, request_id
```

## Níveis de Log

```go
logger.Debug("Informação de debug")
logger.Info("Informação geral")
logger.Warn("Aviso importante")
logger.Error("Erro ocorrido")
```

## Performance

Baseado nos benchmarks:

| Provider | Logs/segundo | Uso de Memória | Recomendação |
|----------|-------------|----------------|--------------|
| Zap      | ~240k       | Baixo          | Padrão (alta performance) |
| Zerolog  | ~174k       | Muito baixo    | Aplicações com restrições de memória |
| Slog     | ~132k       | Médio          | Compatibilidade com stdlib |

## Testando

Execute todos os exemplos:

```bash
bash test_examples.sh
```

## Configuração Avançada

### Personalizando o Provider Zap

```go
config := map[string]interface{}{
    "level":      "info",
    "format":     "json",
    "outputPath": "stdout",
}

logger.ConfigureProvider("zap", config)
```

### Verificando o Provider Atual

```go
providerName := logger.GetCurrentProviderName()
fmt.Printf("Provider atual: %s\n", providerName)
```

## Integração com Middleware

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        ctx = context.WithValue(ctx, "request_id", generateRequestID())
        
        logger.InfoContext(ctx, "Request recebido",
            slog.String("method", r.Method),
            slog.String("path", r.URL.Path),
        )
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Melhores Práticas

1. **Use o provider padrão** para a maioria dos casos
2. **Inclua contexto** sempre que possível
3. **Use campos estruturados** em vez de concatenação de strings
4. **Mantenha mensagens concisas** e descritivas
5. **Use níveis apropriados** (Debug para desenvolvimento, Info para produção)

## Troubleshooting

### Provider não encontrado
```
Error: provider "xyz" not found
```
Verifique se o provider está registrado. Providers disponíveis: slog, zap, zerolog

### Erro de configuração
```
Error: failed to configure provider
```
Verifique se a configuração está no formato correto para o provider específico.

## Contribuindo

Para adicionar novos providers ou funcionalidades:

1. Implemente a interface `Provider`
2. Registre o provider no `manager.go`
3. Adicione testes e exemplos
4. Atualize a documentação
