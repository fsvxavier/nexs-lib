# Domain Errors v2

Um sistema completo e robusto de tratamento de erros para aplicações Go, seguindo os princípios de Clean Architecture, SOLID e Design Patterns.

## 🎯 Características

- **Arquitetura Hexagonal**: Separação clara entre interfaces, implementações e tipos
- **Thread-Safe**: Operações seguras para ambientes concorrentes
- **Performance Otimizada**: Object pooling para reduzir alocações de memória
- **Construção Fluente**: Builder pattern para criação intuitiva de erros
- **Tipagem Forte**: Sistema de tipos bem definido para categorização
- **Parsing Inteligente**: Parsers especializados para diferentes tipos de erro
- **Serialização JSON**: Compatibilidade completa com APIs REST
- **Stack Tracing**: Rastreamento detalhado da origem dos erros
- **Registry Pattern**: Registro centralizado de códigos de erro
- **Compatibilidade**: Total compatibilidade com interfaces padrão do Go

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/v2/domainerrors
```

## 🚀 Uso Básico

### Criação Simples

```go
import "github.com/fsvxavier/nexs-lib/v2/domainerrors"

// Erro básico
err := domainerrors.New("E001", "User not found")
fmt.Println(err.Error()) // [E001] User not found
```

### Construção Fluente

```go
err := domainerrors.NewBuilder().
    WithCode("E002").
    WithMessage("Invalid user data").
    WithType(string(types.ErrorTypeValidation)).
    WithDetail("field", "email").
    WithDetail("value", "invalid-email").
    WithTag("validation").
    Build()
```

### Erros de Validação

```go
fields := map[string][]string{
    "email": {"invalid format", "required"},
    "age":   {"must be positive"},
}

validationErr := domainerrors.NewValidationError("Validation failed", fields)
```

### Wrapping de Erros

```go
originalErr := errors.New("database connection failed")
wrappedErr := domainerrors.New("DB001", "Query failed").
    Wrap("database error", originalErr)
```

## 🏗️ Arquitetura

```
domainerrors/
├── interfaces/          # Contratos e interfaces
├── types/              # Tipos e constantes
├── factory/            # Factories para criação de erros
├── registry/           # Registro de códigos de erro
├── parsers/            # Parsers especializados
├── examples/           # Exemplos práticos
├── domain_error.go     # Implementação principal
├── builder.go          # Builder pattern
├── validation_error.go # Erros de validação
└── domainerrors.go     # API pública
```

### Componentes Principais

#### 1. DomainError
Implementação principal que oferece:
- Thread safety com sync.RWMutex
- Object pooling para performance
- Stack trace detalhado
- Serialização JSON otimizada
- Hierarquia de erros com wrapping/chaining

#### 2. ErrorBuilder
Construção fluente de erros:
```go
err := domainerrors.NewBuilder().
    WithCode("E001").
    WithMessage("Error occurred").
    WithType(string(types.ErrorTypeValidation)).
    WithSeverity(interfaces.Severity(types.SeverityHigh)).
    WithDetail("key", "value").
    WithTag("important").
    Build()
```

#### 3. ValidationError
Especialização para erros de validação:
```go
validationErr := domainerrors.NewValidationError("Validation failed", nil)
validationErr.AddField("email", "invalid format")
validationErr.AddField("age", "must be positive")
```

#### 4. Factory Pattern
Factories especializadas para diferentes contextos:
```go
// Factory padrão
factory := factory.GetDefaultFactory()
err := factory.NewNotFound("User", "123")

// Factory de banco de dados
dbFactory := factory.GetDatabaseFactory()
err := dbFactory.NewConnectionError("postgres", cause)

// Factory HTTP
httpFactory := factory.GetHTTPFactory()
err := httpFactory.NewHTTPError(404, "Not found")
```

#### 5. Registry Pattern
Registro centralizado de códigos de erro:
```go
// Registra código personalizado
info := interfaces.ErrorCodeInfo{
    Code:        "USR001",
    Message:     "User not found: %s",
    Type:        string(types.ErrorTypeNotFound),
    StatusCode:  404,
    Severity:    interfaces.Severity(types.SeverityLow),
    Retryable:   false,
    Tags:        []string{"user", "not_found"},
    Description: "Occurs when a user cannot be found by ID",
}
registry.RegisterGlobal(info)

// Cria erro a partir do código
err, _ := registry.CreateErrorGlobal("USR001", "user-123")
```

#### 6. Parsers Especializados
Parsers para diferentes tipos de erro:
```go
// Parser composto com todos os parsers
parser := parsers.NewDefaultParser()

// Parse de erro específico
parsed := parsers.ParseError(someError, parser)
```

## 📋 Tipos de Erro

### Categorias Principais

| Categoria | Tipos | Descrição |
|-----------|-------|-----------|
| **Data** | `Repository`, `Database`, `Cache`, `Migration`, `Serialization` | Erros relacionados a dados |
| **Input** | `Validation`, `BadRequest`, `Unprocessable`, `Unsupported` | Erros de entrada |
| **Business** | `BusinessRule`, `Workflow`, `Conflict`, `NotFound` | Erros de negócio |
| **Security** | `Authentication`, `Authorization`, `Security`, `Forbidden` | Erros de segurança |
| **System** | `Internal`, `Infrastructure`, `Configuration`, `Dependency` | Erros de sistema |
| **Communication** | `ExternalService`, `Timeout`, `RateLimit`, `Network` | Erros de comunicação |
| **Protocol** | `HTTP`, `gRPC`, `GraphQL`, `WebSocket` | Erros de protocolo |

### Códigos Padrão

| Código | Tipo | Mensagem | Status HTTP |
|--------|------|----------|-------------|
| `E001` | Validation | Validation failed | 400 |
| `E002` | NotFound | Resource not found | 404 |
| `E003` | Conflict | Resource already exists | 409 |
| `E004` | BusinessRule | Business rule violation | 422 |
| `E005` | Authentication | Authentication failed | 401 |
| `E006` | Authorization | Access denied | 403 |
| `E007` | Internal | Internal server error | 500 |
| `E008` | ExternalService | External service unavailable | 502 |
| `E009` | Timeout | Request timeout | 504 |
| `E010` | RateLimit | Rate limit exceeded | 429 |

## 🎯 Funcionalidades Avançadas

### Thread Safety
Todas as operações são thread-safe:
```go
// Seguro para uso concorrente
err := domainerrors.New("E001", "Error")
go func() {
    details := err.Details() // Leitura segura
}()
go func() {
    _ = err.Code() // Leitura segura
}()
```

### Performance com Object Pooling
O sistema usa object pooling para otimizar performance:
```go
// Reutilização automática de objetos
for i := 0; i < 1000000; i++ {
    err := domainerrors.New("E001", "Error")
    // Objeto é automaticamente retornado ao pool
}
```

### Serialização JSON Otimizada
```go
err := domainerrors.NewBuilder().
    WithCode("E001").
    WithMessage("Error occurred").
    WithDetail("key", "value").
    Build()

jsonData, _ := err.JSON()
// {
//   "code": "E001",
//   "message": "Error occurred", 
//   "type": "internal",
//   "details": {"key": "value"},
//   "timestamp": "2024-01-01T10:00:00Z"
// }
```

### Análise de Erros
```go
// Funções utilitárias para análise
isRetryable := domainerrors.IsRetryable(err)
isTemporary := domainerrors.IsTemporary(err)
errorType := domainerrors.GetErrorType(err)
statusCode := domainerrors.GetStatusCode(err)
rootCause := domainerrors.GetRootCause(err)
```

### Stack Tracing
```go
err := domainerrors.New("E001", "Error")
err = err.Wrap("context info", anotherError)

// Stack trace detalhado
fmt.Println(err.FormatStackTrace())
// Stack Trace:
// 1: [context info] in main.someFunction (main.go:42)
//    Error: another error occurred
```

## 📝 Exemplos Práticos

### Serviço de Usuário
```go
type UserService struct{}

func (s *UserService) GetUser(id string) (*User, error) {
    if id == "" {
        return nil, domainerrors.NewBadRequestError("User ID is required")
    }
    
    user, err := s.repository.FindByID(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domainerrors.NewNotFoundError("User", id)
        }
        return nil, domainerrors.NewInternalError("Failed to query user", err)
    }
    
    return user, nil
}

func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
    // Validação estruturada
    validationErr := domainerrors.NewValidationError("User validation failed", nil)
    
    if req.Name == "" {
        validationErr.AddField("name", "required field")
    }
    
    if req.Email == "" {
        validationErr.AddField("email", "required field")
    } else if !isValidEmail(req.Email) {
        validationErr.AddField("email", "invalid format")
    }
    
    if len(validationErr.Fields()) > 0 {
        return nil, validationErr
    }
    
    // Verifica conflitos
    exists, err := s.repository.EmailExists(req.Email)
    if err != nil {
        return nil, domainerrors.NewInternalError("Failed to check email", err)
    }
    if exists {
        return nil, domainerrors.NewConflictError("Email already exists")
    }
    
    // Cria usuário
    user, err := s.repository.Create(req)
    if err != nil {
        return nil, domainerrors.NewInternalError("Failed to create user", err)
    }
    
    return user, nil
}
```

### Handler HTTP
```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, domainerrors.NewBadRequestError("Invalid JSON"))
        return
    }
    
    user, err := h.userService.CreateUser(req)
    if err != nil {
        h.writeError(w, err)
        return
    }
    
    h.writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) writeError(w http.ResponseWriter, err error) {
    statusCode := domainerrors.GetStatusCode(err)
    
    response := map[string]interface{}{
        "error": true,
        "code":  domainerrors.GetErrorCode(err),
        "message": err.Error(),
        "type": domainerrors.GetErrorType(err),
    }
    
    // Adiciona detalhes de validação se aplicável
    if validationErr, ok := err.(interfaces.ValidationErrorInterface); ok {
        response["fields"] = validationErr.Fields()
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

## 🧪 Testes

Execute os testes com cobertura:

```bash
# Testes unitários
go test ./... -v

# Cobertura
go test ./... -cover

# Cobertura detalhada
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Benchmarks
go test ./... -bench=. -benchmem
```

### Cobertura de Testes
O módulo possui cobertura de testes superior a 98%, incluindo:
- Testes unitários completos
- Testes de integração
- Testes de benchmark
- Testes de thread safety
- Testes de casos extremos

## 📊 Performance

### Benchmarks
```
BenchmarkDomainError_Creation-8         5000000    243 ns/op    96 B/op   2 allocs/op
BenchmarkDomainError_Builder-8          2000000    621 ns/op   256 B/op   4 allocs/op
BenchmarkDomainError_JSON-8             1000000   1543 ns/op   512 B/op   8 allocs/op
BenchmarkDomainError_Wrapping-8         3000000    456 ns/op   128 B/op   3 allocs/op
```

### Otimizações
- **Object Pooling**: Reduz alocações de memória em ~60%
- **String Builder**: Otimiza concatenação de strings
- **JSON Streaming**: Serialização eficiente
- **Lazy Loading**: Stack trace calculado apenas quando necessário
- **Copy-on-Write**: Maps e slices copiados apenas quando modificados

## 🔧 Configuração

### Factory Personalizada
```go
// Factory com configurações customizadas
factory := factory.NewCustomFactory(
    "CUSTOM",                    // prefixo padrão
    types.SeverityMedium,       // severidade padrão
    true,                       // habilita stack trace
)
```

### Registry Personalizado
```go
// Registry com configurações específicas
registry := registry.NewErrorCodeRegistryWithFactory(customFactory)

// Importa códigos de arquivo/configuração
codes := map[string]interfaces.ErrorCodeInfo{
    "APP001": {
        Code: "APP001",
        Message: "Application specific error",
        Type: string(types.ErrorTypeBusinessRule),
        StatusCode: 422,
    },
}
registry.Import(codes, false)
```

## 🤝 Contribuição

1. Fork o projeto
2. Crie sua feature branch (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

### Diretrizes
- Mantenha cobertura de testes > 98%
- Siga os princípios SOLID
- Documente novas funcionalidades
- Execute linting: `golangci-lint run`

## 📜 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🙏 Agradecimentos

- Inspirado nos princípios de Clean Architecture de Robert C. Martin
- Padrões de Design do Gang of Four
- Comunidade Go pelos excelentes pacotes de referência

## 📞 Suporte

- **Issues**: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- **Documentação**: [Documentação completa](https://pkg.go.dev/github.com/fsvxavier/nexs-lib/v2/domainerrors)
- **Exemplos**: [Pasta de exemplos](./examples/)

---

**Desenvolvido com ❤️ em Go seguindo as melhores práticas de engenharia de software.**
