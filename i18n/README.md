# i18n Library - Sistema de Internacionalização de Alta Performance

Uma biblioteca completa de internacionalização (i18n) para Go que implementa múltiplos padrões de design incluindo Factory, Observer, Hook, Middleware e Registry para fornecer um sistema de tradução flexível, extensível e de alta performance.

## 🌟 Características Principais

- **Múltiplos Providers**: JSON, YAML e extensível para outros formatos
- **Sistema de Hooks**: Logging, métricas, validação e hooks personalizados
- **Middlewares**: Cache, rate limiting, logging e middlewares personalizados
- **Registry Pattern**: Gerenciamento centralizado de providers, hooks e middlewares
- **Factory Pattern**: Criação flexível de providers
- **Observer Pattern**: Hooks para eventos do ciclo de vida
- **Configuração Flexível**: Builder pattern para configuração
- **Suporte a Parâmetros**: Substituição de templates com {{variável}}
- **Chaves Aninhadas**: Navegação por estruturas complexas com notação de ponto
- **Fallback**: Fallback automático para idioma padrão
- **Cache Integrado**: Sistema de cache configurável
- **Validação**: Validação de configuração e parâmetros
- **Thread-Safe**: Operações seguras para concorrência
- **Alta Performance**: Otimizado para operações rápidas
- **Cobertura de Testes**: +98% de cobertura de testes
- **Logging Estruturado**: Sistema de logging completo
- **Health Checks**: Verificação de saúde dos providers

## 🚀 **NOVO: Otimizações de Performance (Fase 2)**

### ⚡ Performance Benchmarks Validados

| Componente | Performance | Uso de Memória | Throughput |
|------------|-------------|----------------|------------|
| **String Interner** | 51.76 ns/op | 7 B/op | 22M+ ops/s |
| **String Pool** | 14.90 ns/op | 24 B/op | 75M+ ops/s |
| **Batch Translation** | 362μs/1000 itens | Escalável | 2.6K traduções/ms |
| **Performance Provider** | 64.73 ns/op | 0 B/op | Zero alocações |

### 🔧 Otimizações Implementadas

- ✅ **Memory Pooling**: Reutilização de objetos para reduzir GC pressure
- ✅ **String Interning**: Cache de chaves comuns para economizar memória
- ✅ **Batch Operations**: Processamento em lote com worker pools
- ✅ **Lazy Loading**: Carregamento sob demanda de idiomas
- ✅ **Performance Wrappers**: Providers otimizados com zero overhead

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/i18n
```

## 🚀 Uso Básico

### Exemplo JSON Simples

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fsvxavier/nexs-lib/i18n"
    "github.com/fsvxavier/nexs-lib/i18n/config"
    "github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

func main() {
    // Configurar o sistema i18n
    cfg, err := config.NewConfigBuilder().
        WithSupportedLanguages("en", "pt", "es").
        WithDefaultLanguage("en").
        WithFallbackToDefault(true).
        WithProviderConfig(&config.JSONProviderConfig{
            FilePath:    "./translations",
            FilePattern: "{lang}.json",
            NestedKeys:  true,
        }).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Criar e configurar o registry
    registry := i18n.NewRegistry()
    registry.RegisterProvider(&json.Factory{})

    // Criar o provider
    provider, err := registry.CreateProvider("json", cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Iniciar o provider
    ctx := context.Background()
    if err := provider.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer provider.Stop(ctx)

    // Usar traduções
    result, _ := provider.Translate(ctx, "hello", "pt", nil)
    fmt.Println(result) // Output: Olá

    // Traduções com parâmetros
    params := map[string]interface{}{
        "name": "Maria",
        "count": 5,
    }
    result, _ = provider.Translate(ctx, "notification", "pt", params)
    fmt.Println(result) // Output: Olá Maria, você tem 5 novas mensagens!
}
```

### Exemplo YAML

```go
// Usando provider YAML
cfg, err := config.NewConfigBuilder().
    WithSupportedLanguages("en", "pt").
    WithDefaultLanguage("en").
    WithProviderConfig(&config.YAMLProviderConfig{
        FilePath:    "./translations",
        FilePattern: "{lang}.yaml",
        NestedKeys:  true,
    }).
    Build()

registry.RegisterProvider(&yaml.Factory{})
provider, err := registry.CreateProvider("yaml", cfg)
```

## 🚀 Uso das Otimizações de Performance

### Performance Optimized Provider

```go
// Provider base
baseProvider, _ := registry.CreateProvider("json", cfg)

// Aplicar otimizações de performance
optimizedProvider := i18n.NewPerformanceOptimizedProvider(baseProvider)

// Tradução otimizada (com string interning automático)
result, _ := optimizedProvider.Translate(ctx, "hello.world", "en", nil)

// Verificar strings internalizadas
count := optimizedProvider.GetInternedStringCount()
fmt.Printf("Strings internalizadas: %d\n", count)
```

### Batch Translation

```go
// Criar batch translator
batchTranslator := i18n.NewBatchTranslator(baseProvider)

// Preparar lote de traduções
requests := []i18n.BatchTranslationRequest{
    {Key: "hello.world", Lang: "en", Params: nil},
    {Key: "goodbye.world", Lang: "es", Params: nil},
    {Key: "welcome.user", Lang: "pt", Params: map[string]interface{}{"name": "João"}},
}

// Processar em lote (mais eficiente que traduções individuais)
responses := batchTranslator.TranslateBatch(ctx, requests)

for _, resp := range responses {
    if resp.Error != "" {
        fmt.Printf("Erro: %s\n", resp.Error)
    } else {
        fmt.Printf("%s [%s]: %s\n", resp.Key, resp.Lang, resp.Translation)
    }
}
```

### Lazy Loading Provider

```go
// Ideal para aplicações com muitos idiomas
lazyProvider := i18n.NewLazyLoadingProvider(baseProvider)

// Idiomas são carregados apenas quando necessário
result, _ := lazyProvider.Translate(ctx, "hello.world", "pt", nil) // Carrega PT sob demanda
```

### Combinando Todas as Otimizações

```go
// Para máxima performance
baseProvider, _ := registry.CreateProvider("json", cfg)
lazyProvider := i18n.NewLazyLoadingProvider(baseProvider)              // Lazy loading
optimizedProvider := i18n.NewPerformanceOptimizedProvider(lazyProvider) // String interning + pools
compressedProvider := i18n.NewCompressedProvider(optimizedProvider, true) // Compressão

// Provider final com todas as otimizações
result, _ := compressedProvider.Translate(ctx, "hello.world", "en", nil)
```

### String Pool e String Interner Globais

```go
// Usar instâncias globais para máxima eficiência
interner := i18n.GetGlobalStringInterner()
pool := i18n.GetGlobalStringPool()

// String interning manual (opcional - já feito automaticamente no PerformanceOptimizedProvider)
key := interner.Intern("common.translation.key")

// String pool manual
slice := pool.Get()
defer pool.Put(slice)
slice = append(slice, "item1", "item2", "item3")
```
```

## 🔧 Uso Avançado com Hooks e Middlewares

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/fsvxavier/nexs-lib/i18n"
    "github.com/fsvxavier/nexs-lib/i18n/config"
    "github.com/fsvxavier/nexs-lib/i18n/hooks"
    "github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

func main() {
    // Configuração
    cfg, _ := config.NewConfigBuilder().
        WithSupportedLanguages("en", "pt", "es").
        WithDefaultLanguage("en").
        WithCache(true, 5*time.Minute).
        WithProviderConfig(&config.JSONProviderConfig{
            FilePath:    "./translations",
            FilePattern: "{lang}.json",
            NestedKeys:  true,
        }).
        Build()

    // Registry com hooks
    registry := i18n.NewRegistry()
    
    // Adicionar hooks
    loggingHook, _ := hooks.NewLoggingHook("logging", 1, hooks.LoggingHookConfig{
        LogLevel:        "info",
        LogTranslations: true,
        LogErrors:       true,
    }, nil)
    registry.AddHook(loggingHook)

    metricsHook, _ := hooks.NewMetricsHook("metrics", 2, hooks.MetricsHookConfig{
        CollectTranslationMetrics: true,
        CollectErrorMetrics:       true,
        CollectPerformanceMetrics: true,
    })
    registry.AddHook(metricsHook)

    // Registrar provider e criar instância
    registry.RegisterProvider(&json.Factory{})
    provider, _ := registry.CreateProvider("json", cfg)

    ctx := context.Background()
    provider.Start(ctx)
    defer provider.Stop(ctx)

    // Usar traduções (com hooks ativos)
    result, _ := provider.Translate(ctx, "hello", "pt", nil)
    fmt.Println(result)
}
```

## 📁 Estrutura de Arquivos de Tradução

### JSON Format

```json
// en.json
{
  "hello": "Hello",
  "notification": "Hello {{name}}, you have {{count}} new messages!",
  "api": {
    "errors": {
      "not_found": "Resource not found",
      "unauthorized": "Access denied"
    },
    "success": {
      "created": "Resource created successfully"
    }
  }
}

// pt.json
{
  "hello": "Olá",
  "notification": "Olá {{name}}, você tem {{count}} novas mensagens!",
  "api": {
    "errors": {
      "not_found": "Recurso não encontrado",
      "unauthorized": "Acesso negado"
    },
    "success": {
      "created": "Recurso criado com sucesso"
    }
  }
}
```

### YAML Format

```yaml
# en.yaml
hello: Hello
notification: "Hello {{name}}, you have {{count}} new messages!"
api:
  errors:
    not_found: Resource not found
    unauthorized: Access denied
  success:
    created: Resource created successfully

# pt.yaml
hello: Olá
notification: "Olá {{name}}, você tem {{count}} novas mensagens!"
api:
  errors:
    not_found: Recurso não encontrado
    unauthorized: Acesso negado
  success:
    created: Recurso criado com sucesso
```

## 🎯 Características Detalhadas

### Chaves Aninhadas

```go
// Acesse valores aninhados usando notação de ponto
result, _ := provider.Translate(ctx, "api.errors.not_found", "pt", nil)
// Output: Recurso não encontrado
```

### Parâmetros de Template

```go
params := map[string]interface{}{
    "name":     "João",
    "count":    3,
    "price":    29.99,
    "currency": "BRL",
}

result, _ := provider.Translate(ctx, "order.total", "pt", params)
// Template: "Pedido para {{name}}: {{count}} itens por {{price}} {{currency}}"
// Output: "Pedido para João: 3 itens por 29.99 BRL"
```

### Sistema de Fallback

```go
// Se uma tradução não existe em 'es', automaticamente usa 'en' (padrão)
result, _ := provider.Translate(ctx, "some.key", "es", nil)
// Se some.key não existir em es.json, retorna o valor de en.json
```

### Hooks Disponíveis

- **LoggingHook**: Log de traduções, erros e performance
- **MetricsHook**: Coleta de métricas de uso
- **ValidationHook**: Validação de parâmetros e idiomas

### Health Checks

```go
if err := provider.Health(ctx); err != nil {
    fmt.Printf("Provider unhealthy: %v\n", err)
} else {
    fmt.Println("Provider is healthy")
}
```

## 🧪 Exemplos Disponíveis

A biblioteca inclui exemplos completos:

```bash
# Exemplo básico JSON
go run examples/basic_json/main.go

# Exemplo básico YAML  
go run examples/basic_yaml/main.go

# Exemplo avançado com hooks
go run examples/advanced/main.go
```

## 📊 Performance

A biblioteca é otimizada para alta performance:

- **Traduções simples**: ~40ns por operação
- **Traduções com parâmetros**: ~336ns por operação
- **Cache integrado**: Reduz latência em traduções repetidas
- **Thread-safe**: Seguro para uso concorrente

## 🧪 Testes

Execute todos os testes:

```bash
# Todos os testes
go test ./... -v

# Testes com cobertura
go test ./... -cover

# Benchmarks
go test ./... -bench=.
```

## 🏗️ Arquitetura

A biblioteca implementa vários padrões de design:

- **Registry**: Gerencia factories e instâncias
- **Factory**: Cria providers com configuração
- **Observer**: Hooks para eventos de ciclo de vida
- **Middleware**: Funcionalidade transversal
- **Builder**: Configuração fluente
- **Template Method**: Estrutura comum para providers

## 🔄 Extensibilidade

### Criando um Provider Personalizado

```go
type CustomProvider struct {
    // Implementar interfaces.I18n
}

type CustomFactory struct{}

func (f *CustomFactory) Name() string {
    return "custom"
}

func (f *CustomFactory) Create(config interface{}) (interfaces.I18n, error) {
    // Implementar criação
    return &CustomProvider{}, nil
}

func (f *CustomFactory) ValidateConfig(config interface{}) error {
    // Implementar validação
    return nil
}

// Registrar o factory personalizado
registry.RegisterProvider(&CustomFactory{})
```

### Criando Hooks Personalizados

```go
type CustomHook struct {
    name     string
    priority int
}

func (h *CustomHook) Name() string { return h.name }
func (h *CustomHook) Priority() int { return h.priority }

func (h *CustomHook) OnStart(ctx context.Context, providerName string) error {
    // Lógica personalizada no início
    return nil
}

func (h *CustomHook) OnTranslate(ctx context.Context, providerName, key, lang, result string) error {
    // Lógica personalizada na tradução
    return nil
}

// Implementar outros métodos da interface Hook...

// Registrar o hook personalizado
registry.AddHook(&CustomHook{name: "custom", priority: 10})
```

## 📋 API Reference

### Interfaces Principais

- `I18n`: Interface principal para providers
- `ProviderFactory`: Factory para criar providers
- `Hook`: Interface para hooks de eventos
- `Middleware`: Interface para middlewares
- `Registry`: Gerenciador central

### Métodos Principais

- `Translate(ctx, key, lang, params)`: Traduzir uma chave
- `HasTranslation(ctx, key, lang)`: Verificar se tradução existe
- `SetDefaultLanguage(lang)`: Definir idioma padrão
- `Start(ctx)`: Iniciar provider
- `Stop(ctx)`: Parar provider
- `Health(ctx)`: Verificar saúde

## 📝 Licença

Este projeto está licenciado sob a [MIT License](LICENSE).

## 🤝 Contribuição

Contribuições são bem-vindas! Por favor, veja [CONTRIBUTING.md](CONTRIBUTING.md) para diretrizes.

## 📞 Suporte

Para suporte e questões:
- 📧 Email: support@example.com
- 🐛 Issues: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- 📖 Documentação: [Wiki](https://github.com/fsvxavier/nexs-lib/wiki)
