# Domain Errors Library

Uma biblioteca robusta e idiomática para tratamento de erros de domínio em Go, seguindo os princípios de Domain-Driven Design (DDD).

## 🎯 Características

- **Categorização de erros** por tipos específicos (validação, negócios, infraestrutura, etc.)
- **Empilhamento de erros** com informações contextuais e código personalizado
- **Captura automática de stack trace** para facilitar a depuração
- **Mapeamento automático para códigos HTTP** para uso em APIs REST
- **Suporte para metadados** específicos por tipo de erro
- **Interfaces bem definidas** para máxima flexibilidade
- **Compatibilidade total** com o pacote `errors` padrão do Go
- **Suporte a context.Context** em todas as operações

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/domainerrors
```

## 🚀 Uso Básico

### Criação de Erros

```go
import "github.com/fsvxavier/nexs-lib/domainerrors"

// Erro simples
err := domainerrors.New("USR_001", "User validation failed")

// Erro com causa
originalErr := errors.New("database timeout")
err := domainerrors.NewWithCause("DB_001", "Failed to save user", originalErr)

// Erro com tipo específico
err := domainerrors.NewWithType("VAL_001", "Invalid input", domainerrors.ErrorTypeValidation)
```

### Tipos Específicos de Erro

```go
// Erro de validação
validationErr := domainerrors.NewValidationError("Validation failed", nil)
validationErr.WithField("email", "invalid format")
validationErr.WithField("age", "must be positive")

// Erro de negócio
businessErr := domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Account balance too low")
businessErr.WithRule("minimum balance required")

// Erro de banco de dados
dbErr := domainerrors.NewDatabaseError("Query failed", originalErr)
dbErr.WithOperation("SELECT", "users")
dbErr.WithQuery("SELECT * FROM users WHERE id = ?")

// Erro de serviço externo
extErr := domainerrors.NewExternalServiceError("payment-api", "Payment failed", originalErr)
extErr.WithEndpoint("/api/v1/charge")
extErr.WithResponse(503, "Service unavailable")
```

### Empilhamento de Erros

```go
// Erro base
baseErr := errors.New("connection refused")

// Camada de infraestrutura
infraErr := domainerrors.NewinfraestructureError("database", "Connection failed", baseErr)

// Camada de serviço
serviceErr := domainerrors.New("SERVICE_ERROR", "User service failed")
serviceErr.Wrap("processing user request", infraErr)

// Contexto adicional
ctx := context.Background()
serviceErr.WithContext(ctx, "handling user registration")
```

### Metadados e Contexto

```go
err := domainerrors.New("API_001", "Request processing failed")
err.WithMetadata("request_id", "req-12345")
err.WithMetadata("user_id", "user-789")
err.WithMetadata("endpoint", "/api/v1/users")
```

## 🔧 Tipos de Erro Disponíveis

### Erros Básicos
- `ValidationError` - Erros de validação de entrada
- `NotFoundError` - Recursos não encontrados
- `BusinessError` - Violações de regras de negócio

### Erros de Infraestrutura
- `DatabaseError` - Erros de banco de dados
- `ExternalServiceError` - Falhas de integração com serviços externos
- `infraestructureError` - Erros de infraestrutura geral
- `DependencyError` - Falhas de dependências externas

### Erros de Segurança
- `AuthenticationError` - Falhas de autenticação
- `AuthorizationError` - Problemas de autorização
- `SecurityError` - Violações de segurança e ameaças

### Erros de Performance
- `TimeoutError` - Operações que excedem tempo limite
- `RateLimitError` - Violações de limite de taxa
- `ResourceExhaustedError` - Recursos esgotados
- `CircuitBreakerError` - Circuit breakers abertos

### Erros de Sistema
- `ConfigurationError` - Problemas de configuração
- `UnsupportedOperationError` - Operações não suportadas
- `BadRequestError` - Requisições mal formadas
- `ConflictError` - Conflitos de recursos
- `InvalidSchemaError` - Erros de validação de schema
- `UnsupportedMediaTypeError` - Tipos de mídia não suportados
- `ServerError` - Erros internos do servidor
- `UnprocessableEntityError` - Entidades não processáveis
- `ServiceUnavailableError` - Serviços indisponíveis

## 🌐 Mapeamento HTTP

Todos os erros são automaticamente mapeados para códigos HTTP apropriados:

```go
// Diferentes tipos mapeiam para diferentes códigos HTTP
validationErr := domainerrors.NewValidationError("Invalid data", nil)
fmt.Println(validationErr.HTTPStatus()) // 400

notFoundErr := domainerrors.NewNotFoundError("User not found")
fmt.Println(notFoundErr.HTTPStatus()) // 404

// Função utilitária para qualquer erro
status := domainerrors.MapHTTPStatus(err)
```

## 🔍 Verificação de Tipos

```go
// Verificar se um erro é de tipo específico
if domainerrors.IsType(err, domainerrors.ErrorTypeValidation) {
    // Tratar erro de validação
}

// Compatibilidade com errors.Is e errors.As
if errors.Is(err, originalErr) {
    // Erro contém originalErr
}

var domainErr *domainerrors.DomainError
if errors.As(err, &domainErr) {
    // Erro é um DomainError
    fmt.Println(domainErr.Code)
}
```

## 📊 Stack Trace

```go
err := domainerrors.New("ERROR_001", "Something went wrong")
err.WithContext(ctx, "processing user request")

// Visualizar stack trace formatado
fmt.Println(err.StackTrace())
```

## 🏗️ Arquitetura

### Interfaces

O módulo define interfaces claras para máxima flexibilidade:

```go
type ErrorDomainInterface interface {
    Error() string
    Unwrap() error
    Type() ErrorType
    HTTPStatus() int
    StackTrace() string
    WithMetadata(key string, value interface{}) ErrorDomainInterface
}
```

### Separação de Domínio

- **domainerrors/**: Implementações concretas
- **interfaces/**: Interfaces e contratos
- **mocks/**: Mocks gerados com gomock
- **internal/**: Utilitários internos (stack trace)

## 🧪 Testes

O módulo possui cobertura completa de testes (>98%):

```bash
# Executar todos os testes
go test -race -timeout 30s -v -coverprofile=coverage.out ./...

# Executar testes unitários
go test -tags=unit -race -timeout 30s ./...

# Executar benchmarks
go test -bench=. -benchmem ./...
```

## 📚 Exemplos

### Básico
```bash
cd examples/basic
go run main.go
```

### Avançado
```bash
cd examples/advanced
go run main.go
```

## 🎯 Casos de Uso

### API REST
```go
func handleError(c *fiber.Ctx, err error) error {
    statusCode := domainerrors.MapHTTPStatus(err)
    
    return c.Status(statusCode).JSON(fiber.Map{
        "error": err.Error(),
        "code":  statusCode,
    })
}
```

### Logging Estruturado
```go
logger.Error("Operation failed",
    zap.String("error_type", reflect.TypeOf(err).String()),
    zap.String("error_code", domainErr.Code),
    zap.Any("metadata", domainErr.Metadata),
)
```

### Retry Pattern
```go
func retryOperation(operation func() error) error {
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        // Verificar se é retryável
        if domainerrors.IsType(err, domainerrors.ErrorTypeTimeout) ||
           domainerrors.IsType(err, domainerrors.ErrorTypeExternalService) {
            time.Sleep(backoffDelay(i))
            continue
        }
        
        return err // Erro não retryável
    }
    
    return domainerrors.NewTimeoutError("retry", "Max retries exceeded", nil)
}
```

## 🔧 Configuração

### Timeouts
```go
// Todos os testes incluem timeout de 30 segundos
go test -timeout 30s ./...
```

### Linting
```go
golangci-lint run
```

### Formatação
```go
gofmt -w .
```

## 🚀 Próximos Passos

Ver `NEXT_STEPS.md` para melhorias futuras e roadmap.

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Adicione testes para nova funcionalidade
4. Execute `go test` e `golangci-lint`
5. Submeta um pull request

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 🏷️ Versão

**v2.0.0** - Versão completa com suporte a todos os tipos de erro, interfaces bem definidas, mocks, stack trace avançado e cobertura completa de testes.
