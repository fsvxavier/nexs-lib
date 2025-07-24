# Provider Logrus

Provider completo para integração do [Logrus](https://github.com/sirupsen/logrus) com o sistema de logging unificado da nexs-lib.

## Características

- **Compatibilidade Total**: Integração completa com loggers Logrus existentes
- **Hooks Nativos**: Suporte a todos os hooks do Logrus
- **Migração Facilitada**: Permite migração gradual de sistemas legados
- **Performance Nativa**: Mantém a performance original do Logrus
- **Interface Unificada**: Beneficia-se da interface padronizada da nexs-lib

## Instalação

```bash
go get github.com/sirupsen/logrus
```

## Uso Básico

### Provider Simples

```go
import (
    "context"
    "github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
    "github.com/fsvxavier/nexs-lib/observability/logger/providers/logrus"
)

func main() {
    provider := logrus.NewProvider()
    ctx := context.Background()
    
    provider.Info(ctx, "Aplicação iniciada",
        interfaces.Field{Key: "version", Value: "1.0.0"},
    )
}
```

### Provider Configurado

```go
config := &interfaces.Config{
    Level:          interfaces.InfoLevel,
    Format:         interfaces.JSONFormat,
    ServiceName:    "meu-servico",
    ServiceVersion: "1.0.0",
    Environment:    "production",
}

provider, err := logrus.NewWithConfig(config)
if err != nil {
    log.Fatal(err)
}
```

### Integração com Logger Existente

```go
import "github.com/sirupsen/logrus"

// Logger Logrus existente
existingLogger := logrus.New()
existingLogger.SetLevel(logrus.WarnLevel)

// Wrap com o provider da nexs-lib
provider := logrus.NewProviderWithLogger(existingLogger)
```

## Funcionalidades Avançadas

### Hooks do Logrus

```go
// Hook personalizado
type CustomHook struct{}

func (h *CustomHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

func (h *CustomHook) Fire(entry *logrus.Entry) error {
    entry.Data["custom_field"] = "valor"
    return nil
}

// Adiciona o hook
provider := logrus.NewProvider()
provider.AddHook(&CustomHook{})
```

### Acesso ao Logger Subjacente

```go
provider := logrus.NewProvider()

// Acessa o logger Logrus para configurações avançadas
logrusLogger := provider.GetLogrusLogger()
logrusLogger.AddHook(someHook)
logrusLogger.SetReportCaller(true)
```

### Diferentes Formatos

```go
// JSON (padrão)
jsonProvider := logrus.NewJSONProvider()

// Texto
textProvider := logrus.NewTextProvider()

// Console
consoleProvider := logrus.NewConsoleProvider()
```

### Provider com Buffer

```go
// Buffer com 1000 entradas e flush a cada 5 segundos
bufferedProvider := logrus.NewBufferedProvider(1000, 5*time.Second)
```

## API Completa

### Construtores

- `NewProvider()` - Provider básico com configuração padrão
- `NewProviderWithLogger(logger)` - Wrap de logger Logrus existente
- `NewWithConfig(config)` - Provider com configuração personalizada
- `NewWithWriter(writer)` - Provider com writer específico
- `NewJSONProvider()` - Provider configurado para JSON
- `NewTextProvider()` - Provider configurado para texto
- `NewConsoleProvider()` - Provider otimizado para console
- `NewBufferedProvider(size, timeout)` - Provider com buffer

### Métodos Específicos do Logrus

- `GetLogrusLogger()` - Retorna o logger Logrus subjacente
- `AddHook(hook)` - Adiciona hook do Logrus
- `ReplaceHooks(hooks)` - Substitui todos os hooks

### Métodos de Logging

```go
// Logs estruturados
provider.Info(ctx, "mensagem", fields...)
provider.Debug(ctx, "mensagem", fields...)
provider.Warn(ctx, "mensagem", fields...)
provider.Error(ctx, "mensagem", fields...)

// Logs formatados
provider.Infof(ctx, "template %s", args...)

// Logs com código
provider.InfoWithCode(ctx, "CODE001", "mensagem", fields...)
```

## Configuração

### Estrutura de Configuração

```go
config := &interfaces.Config{
    Level:          interfaces.InfoLevel,        // Nível de log
    Format:         interfaces.JSONFormat,       // Formato: JSON/Text/Console
    Output:         os.Stdout,                   // Writer de saída
    TimeFormat:     time.RFC3339,                // Formato de timestamp
    ServiceName:    "meu-servico",               // Nome do serviço
    ServiceVersion: "1.0.0",                     // Versão do serviço
    Environment:    "production",                // Ambiente
    AddSource:      false,                       // Adiciona info do código fonte
    AddStacktrace:  false,                       // Adiciona stacktrace
    Fields: map[string]any{                      // Campos globais
        "component": "auth",
    },
    BufferConfig: &interfaces.BufferConfig{      // Configuração de buffer
        Size:         1000,
        FlushTimeout: 5 * time.Second,
    },
}
```

## Vantagens

### Para Migração
- **Zero Breaking Changes**: Mantém compatibilidade total
- **Migração Gradual**: Permite migração incremental
- **Hooks Preservados**: Mantém todos os hooks existentes
- **Configuração Existente**: Reutiliza configurações atuais

### Para Performance
- **Performance Nativa**: Sem overhead significativo
- **Buffer Opcional**: Sistema de buffer para alta throughput
- **Thread-Safe**: Suporte completo à concorrência
- **Memory Efficient**: Controle de uso de memória

### Para Desenvolvimento
- **Interface Unificada**: API consistente com outros providers
- **Type Safety**: Campos estruturados type-safe
- **Context Aware**: Suporte completo a contexto
- **Extensível**: Sistema de hooks flexível

## Benchmarks

```
BenchmarkProviderLogInfo-12        322710    3607 ns/op    1787 B/op    29 allocs/op
BenchmarkProviderWithFields-12    176920    5964 ns/op    4522 B/op    38 allocs/op
BenchmarkProviderClone-12        2230276     521 ns/op     760 B/op     7 allocs/op
```

## Exemplos

Veja a pasta [examples/logrus-provider](../examples/logrus-provider) para exemplos completos incluindo:

- Provider básico e configurado
- Integração com logger existente
- Uso de hooks personalizados
- Diferentes formatos de saída
- Sistema de buffer
- Clonagem de loggers

## Compatibilidade

- **Logrus**: v1.9.0+
- **Go**: 1.19+
- **nexs-lib**: v2.0.0+

## Casos de Uso Ideais

1. **Migração de sistemas legados** que já usam Logrus
2. **Aplicações que dependem de hooks específicos** do Logrus
3. **Sistemas que requerem compatibilidade** com bibliotecas existentes
4. **Desenvolvimento incremental** mantendo funcionalidades atuais
5. **Projetos que precisam de logging JSON** com hooks customizados
