# Logger System - nexs-lib

Sistema de logging robusto e extensível para Go com suporte a múltiplos providers, logging estruturado e consciente de contexto.

## ✨ Características

- **Multi-Provider**: Suporte para slog, zap e zerolog
- **Provider Padrão**: Zap configurado automaticamente como padrão
- **Logging Estruturado**: Campos tipados e validados
- **Context-Aware**: Extração automática de trace_id, span_id, user_id, request_id
- **Flexível**: Troca de providers em runtime
- **Configurável**: Níveis, formatos, sampling e stacktraces
- **Performático**: Otimizado para alta throughput
- **Testado**: Cobertura de testes > 98%

## � Instalação

```bash
go get github.com/fsvxavier/nexs-lib/observability/logger
```

## 📋 Providers Suportados

### 1. **zap** (Uber) - PADRÃO
- **Provider padrão** configurado automaticamente
- Alta performance com zero allocation
- Stacktraces detalhados
- Configuração avançada de encoders

### 2. **slog** (Go Standard Library)
- Provider padrão do Go 1.21+
- Balanceamento ideal entre performance e funcionalidades
- Suporte completo a structured logging

### 3. **zerolog** (RS)
- Zero allocation JSON logger
- Extremamente eficiente
- Formato JSON nativo

Resultados de performance obtidos com o exemplo `examples/benchmark/`:

### Ranking Geral (Média de todos os cenários)
🥇 **zap**: ~240,000 logs/seg (Alta performance, ideal para aplicações críticas)
🥈 **zerolog**: ~174,000 logs/seg (JSON nativo, boa eficiência de memória)
🥉 **slog**: ~132,000 logs/seg (Padrão Go, melhor compatibilidade)

### Performance por Cenário

| Cenário | slog | zap | zerolog | Melhor |
|---------|------|-----|---------|--------|
| **Logs Simples** | 189k logs/seg | **401k logs/seg** | 257k logs/seg | zap |
| **Logs Estruturados** | 117k logs/seg | **223k logs/seg** | 156k logs/seg | zap |
| **Logs com Contexto** | 148k logs/seg | **306k logs/seg** | 183k logs/seg | zap |
| **Logs de Erro** | 120k logs/seg | **213k logs/seg** | 173k logs/seg | zap |
| **Logs Formatados** | 125k logs/seg | 174k logs/seg | **198k logs/seg** | zerolog |
| **Logs Complexos** | 92k logs/seg | **125k logs/seg** | 80k logs/seg | zap |

### Eficiência de Memória
- **zerolog**: Mais eficiente na maioria dos cenários
- **slog**: Bom equilíbrio entre performance e uso de memória
- **zap**: Foca em performance máxima, pode usar mais memória

### Uso de benchmark detalhado
```bash
cd examples/benchmark
go run main.go  # Executa benchmark completo com métricas detalhadas
```s tipados e validados
- **Context-Aware**: Extração automática de trace_id, span_id, user_id, request_id
- **Flexível**: Troca de providers em runtime
- **Configurável**: Níveis, formatos, sampling e stacktraces
- **Performático**: Otimizado para alta throughput
- **Testado**: Cobertura de testes > 98%

## 🚀 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/observability/logger
```

## 📋 Providers Suportados

### 1. **slog** (Go Standard Library)
- Provider padrão do Go 1.21+
- Balanceamento ideal entre performance e funcionalidades
- Suporte completo a structured logging

### 2. **zap** (Uber)
- Alta performance com zero allocation
- Stacktraces detalhados
- Configuração avançada de encoders

### 3. **zerolog** (RS)
- Zero allocation JSON logger
- Extremamente eficiente
- Formato JSON nativo

## 🔧 Uso Básico

### Uso Rápido (com Provider Padrão)

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-lib/observability/logger"
    
    // Importa providers para auto-registração
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
    ctx := context.Background()
    
    // Zap é configurado automaticamente como provider padrão
    logger.Info(ctx, "Aplicação iniciada")
    logger.Debug(ctx, "Debug info", logger.String("key", "value"))
    
    // Verifica qual provider está sendo usado
    currentProvider := logger.GetCurrentProviderName()
    fmt.Printf("Provider atual: %s\n", currentProvider) // Output: zap
}
```

### Configuração Personalizada

```go
package main

import (
    "context"
    "github.com/fsvxavier/nexs-lib/observability/logger"
    
    // Importa providers para auto-registração
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
    // Configuração personalizada (substitui a configuração padrão)
    config := logger.DefaultConfig()
    config.Level = logger.InfoLevel
    config.Format = logger.JSONFormat
    config.ServiceName = "meu-app"
    config.ServiceVersion = "1.0.0"
    config.Environment = "production"
    
    // Configura o provider zap com configuração personalizada
    err := logger.ConfigureProvider("zap", config)
    if err != nil {
        panic(err)
    }
    
    // Define como provider ativo (opcional, pois zap já é padrão)
    err = logger.SetActiveProvider("zap")
    if err != nil {
        panic(err)
    }
    
    // Usa o logger
    ctx := context.Background()
    logger.Info(ctx, "Aplicação iniciada")
    logger.Debug(ctx, "Debug info", logger.String("key", "value"))
}
```
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
    // Configuração básica
    config := logger.DefaultConfig()
    config.Level = logger.InfoLevel
    config.Format = logger.JSONFormat
    config.ServiceName = "meu-app"
    config.ServiceVersion = "1.0.0"
    config.Environment = "production"
    
    // Configura o provider slog
    err := logger.ConfigureProvider("slog", config)
    if err != nil {
        panic(err)
    }
    
    // Define como provider ativo
    err = logger.SetActiveProvider("slog")
    if err != nil {
        panic(err)
    }
    
    // Usa o logger
    ctx := context.Background()
    logger.Info(ctx, "Aplicação iniciada")
    logger.Debug(ctx, "Debug info", logger.String("key", "value"))
}
```

### Context-Aware Logging

```go
// Cria contexto com informações de rastreamento
ctx := context.Background()
ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-123")
ctx = context.WithValue(ctx, logger.SpanIDKey, "span-456")
ctx = context.WithValue(ctx, logger.UserIDKey, "user-789")
ctx = context.WithValue(ctx, logger.RequestIDKey, "req-101")

// Automaticamente inclui os campos do contexto
logger.Info(ctx, "Processando requisição")
logger.Error(ctx, "Erro no processamento", logger.String("error", "details"))
```

### Logging Estruturado

```go
// Diferentes tipos de campos
logger.Info(ctx, "Operação completada",
    logger.String("operation", "create_user"),
    logger.Int("user_id", 123),
    logger.Bool("success", true),
    logger.Duration("elapsed", time.Millisecond*150),
    logger.Float64("score", 95.5),
    logger.Time("timestamp", time.Now()),
)

// Logs formatados
logger.Infof(ctx, "Processados %d items em %v", 100, time.Second)

// Logs com código de erro
logger.ErrorWithCode(ctx, "E001", "Falha na validação",
    logger.String("field", "email"),
    logger.String("value", "invalid-email"),
)
```

### Configuração Avançada

```go
config := &logger.Config{
    Level:          logger.DebugLevel,
    Format:         logger.JSONFormat,
    Output:         os.Stdout,
    AddSource:      true,
    AddStacktrace:  true,
    TimeFormat:     time.RFC3339,
    ServiceName:    "api-service",
    ServiceVersion: "2.1.0",
    Environment:    "production",
    Fields: map[string]any{
        "region":    "us-east-1",
        "component": "auth",
    },
    SamplingConfig: &logger.SamplingConfig{
        Initial:    100,
        Thereafter: 10,
    },
}
```

## 🔄 Troca de Providers

```go
// Configura múltiplos providers
providers := []string{"slog", "zap", "zerolog"}
for _, provider := range providers {
    err := logger.ConfigureProvider(provider, config)
    if err != nil {
        log.Fatal(err)
    }
}

// Troca providers em runtime
logger.SetActiveProvider("zap")    // Usa zap
logger.Info(ctx, "Usando zap")

logger.SetActiveProvider("zerolog") // Usa zerolog
logger.Info(ctx, "Usando zerolog")
```

## 🎯 Contexto Pré-definido

```go
// Logger com campos fixos
contextLogger := logger.WithFields(
    logger.String("module", "auth"),
    logger.String("operation", "login"),
)

// Logger com contexto extraído
ctxLogger := logger.WithContext(ctx)

// Todos os logs subsequentes incluem os campos
contextLogger.Info(ctx, "Tentativa de login")
contextLogger.Error(ctx, "Falha na autenticação")
```

## � Exemplos Práticos

O sistema de logging inclui quatro exemplos completos que demonstram diferentes cenários de uso:

### 1. Exemplo Básico (`examples/basic/`)
Demonstra o uso básico do sistema com todos os três providers:

```bash
cd examples/basic
go run main.go
```

**Funcionalidades demonstradas:**
- Configuração básica de cada provider
- Logging estruturado com campos tipados
- Context-aware logging com trace_id, user_id, etc.
- Comparação de formatos (Console vs JSON)
- Configuração por variáveis de ambiente

### 2. Exemplo Avançado (`examples/advanced/`)
Cenários mais complexos com serviços e middleware:

```bash
cd examples/advanced
go run main.go
```

**Funcionalidades demonstradas:**
- Integração com serviços (UserService)
- Middleware HTTP com logging automático
- Diferentes formatos de saída (Console, Text, JSON)
- Configuração específica por ambiente
- Tratamento de erros com domain errors
- Logging comparativo entre providers
- Benchmarks de performance

### 3. Exemplo Multi-Provider (`examples/multi-provider/`)
Demonstração completa da arquitetura multi-provider:

```bash
cd examples/multi-provider
go run main.go
```

**Funcionalidades demonstradas:**
- Configuração de múltiplos providers
- Troca de providers em runtime
- Benchmarks comparativos de performance
- Context-aware logging
- Logging estruturado com diferentes tipos de campos
- Extração automática de contexto

### 4. Benchmark Completo (`examples/benchmark/`)
Análise detalhada de performance de todos os providers:

```bash
cd examples/benchmark
go run main.go
```

**Funcionalidades demonstradas:**
- Benchmarks detalhados com métricas de memória e GC
- Testes de diferentes cenários (simples, estruturados, contexto, erros)
- Análise comparativa de performance
- Recomendações de uso baseadas em dados
- Informações do sistema e estatísticas detalhadas

### Executar Todos os Exemplos

```bash
# Script para testar todos os exemplos
chmod +x test_examples.sh
./test_examples.sh
```

## �📊 Benchmarks

Resultados de performance (1000 logs com campos estruturados):

| Provider | Tempo Total | Tempo/Log | Logs/Segundo |
|----------|-------------|-----------|--------------|
| slog     | ~45ms       | ~45µs     | ~22,000      |
| zap      | ~25ms       | ~25µs     | ~40,000      |
| zerolog  | ~58ms       | ~58µs     | ~17,000      |

> **Nota**: Resultados podem variar dependendo do hardware e configuração

## 🧪 Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com coverage
go test -cover ./...

# Executar benchmarks
go test -bench=. ./...
```

## 📁 Estrutura do Projeto

```
observability/logger/
├── interfaces/
│   └── interfaces.go       # Interfaces e tipos
├── providers/
│   ├── slog/              # Provider slog
│   ├── zap/               # Provider zap
│   └── zerolog/           # Provider zerolog
├── examples/
│   ├── basic/             # Exemplo básico
│   ├── advanced/          # Exemplo avançado
│   ├── multi-provider/    # Exemplo multi-provider
│   └── benchmark/         # Benchmark completo
├── mocks/
│   └── mocks.go           # Mocks para testes
├── logger.go              # API principal
├── manager.go             # Gerenciamento de providers
└── README.md              # Esta documentação
```

## 🎨 Formatos de Saída

### JSON Format
```json
{
  "time": "2025-07-18T15:59:50-03:00",
  "level": "INFO",
  "msg": "Processando requisição",
  "service": "api-service",
  "version": "1.0.0",
  "environment": "production",
  "trace_id": "trace-123",
  "span_id": "span-456",
  "user_id": "user-789",
  "request_id": "req-101"
}
```

### Console Format
```
2025-07-18T15:59:50-03:00 INFO Processando requisição service=api-service version=1.0.0 trace_id=trace-123
```

## 🔒 Configurações de Segurança

```go
// Configuração para produção
config := logger.ProductionConfig()
config.Level = logger.InfoLevel
config.AddSource = false      // Remove info do código fonte
config.AddStacktrace = false  // Remove stacktraces exceto para Fatal/Panic
config.SamplingConfig = &logger.SamplingConfig{
    Initial:    100,
    Thereafter: 10,           // Sampling para controle de volume
}
```

## 🚀 Próximos Passos

- [x] Implementar providers slog, zap e zerolog
- [x] Sistema de configuração flexível
- [x] Context-aware logging
- [x] Testes unitários e de integração
- [x] Benchmarks de performance
- [x] Documentação completa
- [ ] Métricas de logging
- [ ] Hooks customizados
- [ ] Rotação de logs
- [ ] Integração com sistemas de monitoramento

## 📝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature
3. Faça commit das mudanças
4. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.

## 🆘 Suporte

Para dúvidas ou problemas, abra uma issue no repositório ou consulte a documentação técnica em `/docs`.
