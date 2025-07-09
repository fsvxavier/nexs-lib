# Domain Errors Module

Uma biblioteca robusta e moderna para tratamento de erros em aplicações Go, seguindo os princípios de Domain-Driven Design (DDD) e implementando padrões de design avançados para máxima flexibilidade, observabilidade e manutenibilidade.

## 🚀 Características Principais

- **🎯 Categorização Completa**: Tipos específicos de erro (validação, negócios, infraestrutura, etc.)
- **🏗️ Builder Pattern**: Construção fluente de erros complexos
- **🏭 Factory Pattern**: Criação padronizada com observers e context enrichers
- **👁️ Observer Pattern**: Logging automático e coleta de métricas
- **🔍 Context Enricher**: Enriquecimento automático com dados contextuais
- **📚 Stack Trace**: Captura automática para facilitar depuração
- **🌐 HTTP Mapping**: Mapeamento inteligente para códigos HTTP
- **📊 Metadata Rica**: Suporte robusto para detalhes específicos por tipo
- **🧪 Testabilidade**: Cobertura de testes 87%+
- **🔧 Utilitários**: Helpers avançados para manipulação e análise

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/domainerrors
```

## 🏗️ Arquitetura

```
domainerrors/
├── error.go                    # 🎯 Tipo base DomainError + interfaces
├── error_types.go              # 📋 Definições de tipos específicos
├── error_utils.go              # 🛠️ Utilitários e helpers
├── builder.go                  # 🔨 Builder Pattern para construção fluente
├── factory.go                  # 🏭 Factory Pattern + Observers + Enrichers
├── example/                    # 💡 Exemplos práticos
│   └── builder_factory_example.go # 🆕 Padrões modernos
└── tests/                      # 🧪 Testes (87%+ cobertura)
    ├── *_test.go              # Testes abrangentes
    └── benchmarks/            # Performance tests
```

## 🚀 Uso Rápido

### Criação Básica de Erros

```go
package main

import (
    "fmt"
    "github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
    // Erro básico
    err := domainerrors.New("USER_NOT_FOUND", "Usuário não encontrado")
    fmt.Printf("Erro: %v (HTTP: %d)\n", err, err.StatusCode())
    
    // Erro com causa
    originalErr := fmt.Errorf("database connection failed")
    err = domainerrors.NewWithError("DB_ERROR", "Erro no banco de dados", originalErr)
    
    // Erro tipado
    validationErr := domainerrors.NewValidationError("Email inválido", nil)
    fmt.Printf("É erro de validação: %v\n", domainerrors.IsValidationError(validationErr))
}
```

### Tipos de Erro Disponíveis

```go
// Erros de Validação
validationErr := domainerrors.NewValidationError("Campo obrigatório", nil)
invalidSchemaErr := domainerrors.NewInvalidSchemaError("Schema inválido")

// Erros de Negócio
businessErr := domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Saldo insuficiente")
notFoundErr := domainerrors.NewNotFoundError("Usuário não encontrado", nil)

// Erros de Infraestrutura
dbErr := domainerrors.NewDatabaseError("Falha na conexão", originalErr)
externalErr := domainerrors.NewExternalServiceError("payment-api", "Timeout", timeoutErr)

// Erros de Autenticação/Autorização
authErr := domainerrors.NewAuthenticationError("Token inválido", nil)
authzErr := domainerrors.NewAuthorizationError("Permissão negada", nil)

// Erros de Sistema
timeoutErr := domainerrors.NewTimeoutError("Operação expirou", nil)
conflictErr := domainerrors.NewConflictError("Recurso já existe", nil)
rateLimitErr := domainerrors.NewRateLimitError("Limite de taxa excedido")
```

## 🔨 Builder Pattern (Novo)

Construção fluente de erros complexos:

```go
// Construção básica
err := domainerrors.NewBuilder().
    Code("USER_VALIDATION_FAILED").
    Message("Dados do usuário inválidos").
    Type(domainerrors.ValidationError).
    Build()

// Construção com metadata rica
err = domainerrors.NewBuilder().
    Code("ORDER_PROCESSING_FAILED").
    Message("Falha ao processar pedido").
    Type(domainerrors.BusinessError).
    WithMetadata(map[string]interface{}{
        "order_id": "12345",
        "customer_id": "67890",
        "amount": 99.99,
    }).
    WithTimestamp(time.Now()).
    Build()

// Construção com validação específica
err = domainerrors.NewBuilder().
    Code("FORM_VALIDATION_ERROR").
    Message("Formulário contém erros").
    Type(domainerrors.ValidationError).
    WithValidationField("email", "Email é obrigatório").
    WithValidationField("password", "Senha deve ter ao menos 8 caracteres").
    Build()
```

## 🏭 Factory Pattern (Novo)

Criação de erros com observers e enrichers automáticos:

```go
// Observer para logging
type LoggingObserver struct{}

func (l *LoggingObserver) OnError(ctx context.Context, err *domainerrors.DomainError) {
    log.Printf("[ERROR] %s: %s (Code: %s)", err.Type(), err.Message(), err.Code())
}

// Observer para métricas
type MetricsObserver struct{}

func (m *MetricsObserver) OnError(ctx context.Context, err *domainerrors.DomainError) {
    // Incrementar contador de erros por tipo
    log.Printf("[METRICS] Error type: %s, Code: %s", err.Type(), err.Code())
}

// Context Enricher para informações de request
type RequestEnricher struct{}

func (r *RequestEnricher) EnrichContext(ctx context.Context, builder *domainerrors.ErrorBuilder) {
    if userID := ctx.Value("user_id"); userID != nil {
        builder.WithMetadata(map[string]interface{}{
            "user_id": userID,
        })
    }
}

func ExampleFactory() {
    // Criar factory com observers e enrichers
    factory := domainerrors.NewFactory().
        WithObserver(&LoggingObserver{}).
        WithObserver(&MetricsObserver{}).
        WithContextEnricher(&RequestEnricher{})
    
    ctx := context.WithValue(context.Background(), "user_id", "123")
    
    // Criar erro - observers e enrichers são chamados automaticamente
    err := factory.CreateValidationError(ctx, "Email inválido", nil)
    err = factory.CreateBusinessError(ctx, "INSUFFICIENT_FUNDS", "Saldo insuficiente")
}
```

## 🔍 Verificação de Tipos

```go
err := domainerrors.NewValidationError("Email inválido", nil)

// Verificações específicas
if domainerrors.IsValidationError(err) {
    fmt.Println("É um erro de validação")
}

if domainerrors.IsNotFoundError(err) {
    fmt.Println("É um erro de não encontrado")
}

// Verificação genérica
if domainerrors.IsErrorType(err, domainerrors.ValidationError) {
    fmt.Println("É um erro de validação (verificação genérica)")
}

// Type assertion
if valErr, ok := err.(*domainerrors.ValidationError); ok {
    fmt.Printf("Campos inválidos: %v\n", valErr.Fields)
}
```

## 📊 Mapeamento HTTP

Diferentes erros mapeiam automaticamente para códigos HTTP apropriados:

```go
errors := []error{
    domainerrors.NewValidationError("Email inválido", nil),           // 400
    domainerrors.NewNotFoundError("Usuário não encontrado", nil),     // 404
    domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Saldo insuficiente"), // 422
    domainerrors.NewAuthenticationError("Token inválido", nil),       // 401
    domainerrors.NewAuthorizationError("Permissão negada", nil),      // 403
    domainerrors.NewConflictError("Email já existe", nil),            // 409
    domainerrors.NewRateLimitError("Limite excedido"),                // 429
    domainerrors.NewTimeoutError("Operação expirou", nil),            // 408
    domainerrors.NewDatabaseError("Erro no banco", nil),              // 500
    domainerrors.NewExternalServiceError("api", "Erro", nil),         // 502
}

for _, err := range errors {
    code := domainerrors.GetStatusCode(err)
    fmt.Printf("Erro: %v -> HTTP %d\n", err, code)
}
```

## 🌐 Integração com Frameworks HTTP

### Gin

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/fsvxavier/nexs-lib/domainerrors"
)

func ErrorMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            statusCode := domainerrors.GetStatusCode(err)
            
            response := gin.H{
                "error": err.Error(),
                "code":  statusCode,
            }
            
            // Adicionar detalhes específicos do erro
            if domainErr, ok := err.(*domainerrors.DomainError); ok {
                response["error_code"] = domainErr.Code()
                response["error_type"] = domainErr.Type().String()
                
                if domainErr.Metadata() != nil {
                    response["metadata"] = domainErr.Metadata()
                }
            }
            
            c.JSON(statusCode, response)
        }
    }
}
```

### Fiber

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/fsvxavier/nexs-lib/domainerrors"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
    statusCode := domainerrors.GetStatusCode(err)
    
    response := fiber.Map{
        "error": err.Error(),
        "code":  statusCode,
    }
    
    if domainErr, ok := err.(*domainerrors.DomainError); ok {
        response["error_code"] = domainErr.Code()
        response["error_type"] = domainErr.Type().String()
        
        if metadata := domainErr.Metadata(); metadata != nil {
            response["metadata"] = metadata
        }
    }
    
    return c.Status(statusCode).JSON(response)
}
```

## 📊 Observabilidade

### Métricas com Prometheus

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/fsvxavier/nexs-lib/domainerrors"
)

var (
    errorCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "domain_errors_total",
            Help: "Total number of domain errors",
        },
        []string{"type", "code"},
    )
)

// Observer para métricas
type PrometheusObserver struct{}

func (p *PrometheusObserver) OnError(ctx context.Context, err *domainerrors.DomainError) {
    errorCounter.WithLabelValues(err.Type().String(), err.Code()).Inc()
}
```

### Logging Estruturado

```go
import "go.uber.org/zap"

type StructuredLoggingObserver struct {
    logger *zap.Logger
}

func (s *StructuredLoggingObserver) OnError(ctx context.Context, err *domainerrors.DomainError) {
    fields := []zap.Field{
        zap.String("error_code", err.Code()),
        zap.String("error_type", err.Type().String()),
        zap.String("message", err.Message()),
        zap.Time("timestamp", err.Timestamp()),
    }
    
    if metadata := err.Metadata(); metadata != nil {
        for k, v := range metadata {
            fields = append(fields, zap.Any(k, v))
        }
    }
    
    s.logger.Error("Domain error occurred", fields...)
}
```

## 🔧 Utilitários Avançados

### Registry de Códigos

```go
// Criar registry para códigos personalizados
registry := domainerrors.NewErrorCodeRegistry()

// Registrar códigos customizados
registry.Register("USER_001", "Usuário não encontrado", 404)
registry.Register("USER_002", "Email já existe", 409)

// Usar códigos registrados
err := registry.WrapWithCode("USER_001", fmt.Errorf("usuário 123 não existe"))
```

### Stack de Erros

```go
stack := domainerrors.NewErrorStack()

// Adicionar erros sequencialmente
stack.Push(domainerrors.NewDatabaseError("Conexão perdida", nil))
stack.Push(domainerrors.NewExternalServiceError("api", "Timeout", nil))
stack.Push(domainerrors.NewBusinessError("OPERATION_FAILED", "Operação falhou"))

// Obter informações do stack
fmt.Printf("Total de erros: %d\n", stack.Len())
fmt.Printf("Stack formatado:\n%s\n", stack.Format())
```

### Middleware de Recuperação

```go
// Middleware para capturar panics
err := domainerrors.RecoverMiddleware(func() error {
    panic("algo deu errado!")
    return nil
})

if err != nil {
    fmt.Printf("Panic capturado: %v\n", err)
}
```

## 🧪 Testes

### Testes Unitários

```go
func TestValidationError(t *testing.T) {
    err := domainerrors.NewValidationError("Email inválido", nil)
    
    assert.True(t, domainerrors.IsValidationError(err))
    assert.False(t, domainerrors.IsNotFoundError(err))
    assert.Equal(t, 400, domainerrors.GetStatusCode(err))
    assert.Equal(t, "Email inválido", err.Error())
}

func TestErrorBuilder(t *testing.T) {
    err := domainerrors.NewBuilder().
        Code("TEST_ERROR").
        Message("Erro de teste").
        Type(domainerrors.ValidationError).
        Build()
    
    assert.Equal(t, "TEST_ERROR", err.Code())
    assert.True(t, domainerrors.IsValidationError(err))
}
```

### Cobertura de Testes

```bash
# Executar testes com cobertura
go test -cover ./...

# Resultado atual: 87%+ de cobertura
# - Testes unitários: ✅ 100% das funções principais
# - Testes de integração: ✅ Frameworks HTTP
# - Testes de performance: ✅ Benchmarks
```

## 🔄 Migração

### De Erros Padrão Go

```go
// Antes (erros padrão)
import "errors"

func GetUser(id string) (*User, error) {
    if id == "" {
        return nil, errors.New("ID é obrigatório")
    }
    return nil, errors.New("usuário não encontrado")
}

// Depois (domain errors)
import "github.com/fsvxavier/nexs-lib/domainerrors"

func GetUser(id string) (*User, error) {
    if id == "" {
        return nil, domainerrors.NewValidationError("ID é obrigatório", nil)
    }
    return nil, domainerrors.NewNotFoundError("Usuário não encontrado", nil)
}
```

## 🎯 Casos de Uso Avançados

### Sistema de Auditoria

```go
type AuditObserver struct {
    auditRepo AuditRepository
}

func (a *AuditObserver) OnError(ctx context.Context, err *domainerrors.DomainError) {
    audit := &AuditRecord{
        ErrorCode:   err.Code(),
        ErrorType:   err.Type().String(),
        Message:     err.Message(),
        Timestamp:   err.Timestamp(),
        UserID:      ctx.Value("user_id").(string),
        Metadata:    err.Metadata(),
    }
    
    a.auditRepo.Save(audit)
}
```

### Retry Logic com Backoff

```go
func WithRetry(operation func() error, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Não retry para erros de validação ou business
        if domainerrors.IsValidationError(err) || domainerrors.IsBusinessError(err) {
            break
        }
        
        time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
    }
    
    return domainerrors.NewBuilder().
        Code("OPERATION_FAILED_AFTER_RETRIES").
        Message("Operação falhou após múltiplas tentativas").
        Type(domainerrors.InternalError).
        WithMetadata(map[string]interface{}{
            "max_retries": maxRetries,
            "last_error": lastErr.Error(),
        }).
        WithCause(lastErr).
        Build()
}
```

## 📈 Performance e Benchmarks

### Resultados de Performance

```bash
# Benchmarks típicos:
BenchmarkNewError-8                     5000000    300 ns/op    200 B/op    3 allocs/op
BenchmarkNewErrorWithBuilder-8          3000000    450 ns/op    350 B/op    5 allocs/op
BenchmarkNewErrorWithFactory-8          2000000    600 ns/op    400 B/op    6 allocs/op
BenchmarkErrorTypeCheck-8              50000000     25 ns/op      0 B/op    0 allocs/op
```

### Otimização com Pool

```go
var builderPool = sync.Pool{
    New: func() interface{} {
        return domainerrors.NewBuilder()
    },
}

func CreateOptimizedError(code, message string) error {
    builder := builderPool.Get().(*domainerrors.ErrorBuilder)
    defer func() {
        builder.Reset()
        builderPool.Put(builder)
    }()
    
    return builder.Code(code).Message(message).Build()
}
```

## 🔧 Troubleshooting

### Problemas Comuns

#### Stack Traces Muito Grandes

```go
// Configurar limite de profundidade
config := &domainerrors.Config{
    MaxStackTraceDepth: 10,
    FilterStackTrace:   true,
}
```

#### Performance com Muitos Observers

```go
// Usar observers assíncronos
type AsyncObserver struct {
    ch chan *domainerrors.DomainError
}

func (a *AsyncObserver) OnError(ctx context.Context, err *domainerrors.DomainError) {
    select {
    case a.ch <- err:
    default:
        // Canal cheio, descartar
    }
}
```

## 📚 Documentação Adicional

### Interfaces Principais

```go
// ErrorObserver - Observer para erros
type ErrorObserver interface {
    OnError(ctx context.Context, err *DomainError)
}

// ContextEnricher - Enriquecedor de contexto
type ContextEnricher interface {
    EnrichContext(ctx context.Context, builder *ErrorBuilder)
}

// HttpStatusProvider - Provedor de status HTTP
type HttpStatusProvider interface {
    StatusCode() int
}
```

### Tipos de Erro Disponíveis

```go
const (
    ValidationError ErrorType = iota
    NotFoundError
    BusinessError
    InfrastructureError
    ExternalServiceError
    AuthenticationError
    AuthorizationError
    TimeoutError
    ConflictError
    RateLimitError
    CircuitBreakerError
    ConfigurationError
    SecurityError
    ResourceExhaustedError
    DependencyError
    SerializationError
    CacheError
    WorkflowError
    MigrationError
    InvalidSchemaError
    UnsupportedMediaTypeError
    ServerError
    UnprocessableEntityError
    ServiceUnavailableError
    // ... e mais
)
```

## 🤝 Contribuição

Contribuições são bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### Desenvolvimento

```bash
# Clone o repositório
git clone https://github.com/fsvxavier/nexs-lib.git
cd nexs-lib/domainerrors

# Instalar dependências
go mod tidy

# Executar testes
go test ./...

# Executar testes com cobertura
go test -cover ./...

# Executar benchmarks
go test -bench=. -benchmem
```

## 📝 Changelog

### v2.0.0 (2025-07-07)
- ✨ **NOVO**: Builder Pattern para construção fluente de erros
- ✨ **NOVO**: Factory Pattern com observers e context enrichers
- ✨ **NOVO**: Observer Pattern para logging e métricas automáticas
- ✨ **NOVO**: Context Enricher para enriquecimento automático
- ✨ **NOVO**: 25+ tipos de erro específicos
- ✨ **NOVO**: Integração com frameworks HTTP (Gin, Fiber, Echo)
- ✨ **NOVO**: Observabilidade integrada (métricas, logging, tracing)
- ✨ **NOVO**: Utilitários avançados (registry, stack, recovery)
- 🧪 **MELHORADO**: Cobertura de testes para 87%+
- 📚 **MELHORADO**: Documentação completamente reescrita
- 🔧 **MELHORADO**: Performance otimizada com pools e lazy loading

### v1.x.x (Legado)
- 🔧 Implementação básica com tipos de erro simples
- 🔧 Mapeamento HTTP básico
- 🔧 Utilitários simples

## 📄 Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🆘 Suporte

- 📧 Email: suporte@nexs-lib.com
- 🐛 Issues: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- 📖 Documentação: [docs.nexs-lib.com](https://docs.nexs-lib.com)
- 💬 Discussões: [GitHub Discussions](https://github.com/fsvxavier/nexs-lib/discussions)

## 🏆 Estatísticas

- **Cobertura de testes**: 87%+
- **Tipos de erro**: 25+ tipos específicos
- **Padrões de design**: Builder, Factory, Observer, Context Enricher
- **Integrações**: 3+ frameworks HTTP suportados
- **Performance**: ~300ns/op para criação de erros básicos

---

⭐ **Se este projeto foi útil, considere dar uma estrela no GitHub!**
