# DomainErrors

Um módulo robusto e idiomático para tratamento de erros em Go, seguindo os princípios de Domain-Driven Design (DDD).

## 🚀 Características

- **Genérico e Reutilizável**: Funciona em qualquer aplicação Go
- **Categorização de Erros**: Tipos específicos para diferentes cenários
- **Stack Trace Automático**: Captura contexto de execução
- **Metadados Ricos**: Informações adicionais para debugging
- **Serialização JSON**: Conversão automática para APIs
- **Mapeamento HTTP**: Códigos de status apropriados
- **Empilhamento de Erros**: Preserva cadeia de causas
- **Contexto Integrado**: Suporte nativo ao context.Context
- **Observabilidade**: Métricas e tracing integrados
- **Testabilidade**: Mocks e utilitários para testes

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/domainerrors
```

## 🔧 Uso Básico

### Criando Erros

```go
// Erro básico
err := domainerrors.New("USER_001", "Usuário não encontrado")

// Erro com tipo específico
err := domainerrors.NewWithType("VAL_001", "Dados inválidos", domainerrors.ErrorTypeValidation)

// Erro encapsulando outro erro
err := domainerrors.NewWithError("DB_001", "Falha na consulta", originalErr)
```

### Tipos de Erro Específicos

```go
// Erro de validação
validationErr := domainerrors.NewValidationError("Dados inválidos", nil)
validationErr.WithField("email", "Email é obrigatório")

// Erro de recurso não encontrado
notFoundErr := domainerrors.NewNotFoundError("Usuário não encontrado")

// Erro de regra de negócio
businessErr := domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Saldo insuficiente")
```

### Adicionando Contexto

```go
err := domainerrors.New("API_001", "Falha na API")
err.WithMetadata("user_id", "12345")
err.WithMetadata("operation", "create_user")
err.WithContext(ctx)
```

### Tratamento de Erros

```go
if err != nil {
    // Verificar tipo
    if domainerrors.IsType(err, domainerrors.ErrorTypeValidation) {
        // Tratar erro de validação
    }
    
    // Obter código HTTP
    statusCode := domainerrors.GetHTTPStatus(err)
    
    // Serializar para JSON
    if domainErr, ok := err.(*domainerrors.DomainError); ok {
        jsonData, _ := domainErr.JSON()
        // Enviar como resposta da API
    }
}
```

## 🏗️ Tipos de Erro Disponíveis

| Tipo | Descrição | HTTP Status | Uso |
|------|-----------|-------------|-----|
| `ErrorTypeValidation` | Dados inválidos | 400 | Validação de entrada |
| `ErrorTypeNotFound` | Recurso não encontrado | 404 | Entidades não localizadas |
| `ErrorTypeBusiness` | Regra de negócio violada | 422 | Lógica de domínio |
| `ErrorTypeDatabase` | Falha de banco de dados | 500 | Persistência |
| `ErrorTypeExternalService` | Falha em API externa | 502 | Integração |
| `ErrorTypeInfrastructure` | Problema de infraestrutura | 503 | Recursos do sistema |
| `ErrorTypeTimeout` | Operação expirou | 408 | Limites de tempo |
| `ErrorTypeAuthentication` | Falha de autenticação | 401 | Identidade |
| `ErrorTypeAuthorization` | Acesso negado | 403 | Permissões |
| `ErrorTypeSecurity` | Problema de segurança | 403 | Segurança |

## 🎯 Exemplos

### Exemplo Básico

```go
package main

import (
    "fmt"
    "github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
    // Criar erro
    err := domainerrors.NewValidationError("Dados inválidos", nil)
    err.WithField("email", "Email é obrigatório")
    
    // Verificar propriedades
    fmt.Printf("Código: %s\n", err.Code)
    fmt.Printf("Tipo: %s\n", err.ErrorType)
    fmt.Printf("Status HTTP: %d\n", err.HTTPStatus())
    
    // Serializar
    if jsonData, err := err.JSON(); err == nil {
        fmt.Printf("JSON: %s\n", string(jsonData))
    }
}
```

### Exemplo com Contexto

```go
func CreateUser(ctx context.Context, userData UserData) error {
    // Validar dados
    if err := validateUserData(userData); err != nil {
        validationErr := domainerrors.NewValidationError("Dados de usuário inválidos", err)
        validationErr.WithContext(ctx)
        validationErr.WithMetadata("user_id", userData.ID)
        return validationErr
    }
    
    // Verificar se usuário já existe
    exists, err := userRepo.Exists(ctx, userData.Email)
    if err != nil {
        return domainerrors.NewDatabaseError("Falha ao verificar usuário", err)
    }
    
    if exists {
        conflictErr := domainerrors.NewWithType("USER_EXISTS", "Usuário já existe", domainerrors.ErrorTypeConflict)
        conflictErr.WithMetadata("email", userData.Email)
        return conflictErr
    }
    
    // Criar usuário
    if err := userRepo.Create(ctx, userData); err != nil {
        return domainerrors.NewDatabaseError("Falha ao criar usuário", err)
    }
    
    return nil
}
```

## 📁 Estrutura do Projeto

```
domainerrors/
├── domainerrors.go          # Implementação principal
├── domainerrors_test.go     # Testes unitários
├── interfaces/
│   └── interfaces.go        # Interfaces do módulo
├── internal/
│   └── stack.go            # Captura de stack trace
├── mocks/
│   └── mocks.go            # Mocks para testes
└── examples/
    ├── basic/              # Exemplo básico
    ├── advanced/           # Padrões avançados
    └── global/             # Configuração global
```

## 🧪 Testes

```bash
# Executar testes
go test -v

# Executar testes com coverage
go test -v -cover

# Executar testes com tags
go test -tags=unit -v
```

## 📊 Cobertura de Testes

- **Meta**: 98% de cobertura mínima
- **Atual**: 91.4% de cobertura
- **Testes**: 60+ casos de teste
- **Cenários**: Casos normais, edge cases, falhas

## 🔧 Configuração

### Configuração Global

```go
// Configurar stack trace globalmente
domainerrors.GlobalStackTraceEnabled = true
domainerrors.GlobalMaxStackDepth = 10
domainerrors.GlobalSkipFrames = 2
```

### Handler Centralizado

```go
type ErrorHandler struct {
    logger *log.Logger
    config *Config
}

func (h *ErrorHandler) HandleError(err error) Response {
    // Processar erro
    // Fazer log
    // Enviar métricas
    // Retornar response
}
```

## 🌐 Integração com Frameworks

### Gin

```go
func ErrorMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                domainErr := domainerrors.RecoverWithStackTrace()
                response := convertToHTTPResponse(domainErr)
                c.JSON(response.Status, response)
            }
        }()
        c.Next()
    }
}
```

### Echo

```go
func CustomErrorHandler(err error, c echo.Context) {
    response := convertToHTTPResponse(err)
    c.JSON(response.Status, response)
}
```

### gRPC

```go
func ErrorInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        resp, err := handler(ctx, req)
        if err != nil {
            return resp, enrichErrorWithContext(ctx, err)
        }
        return resp, nil
    }
}
```

## 📈 Observabilidade

### Métricas

```go
// Prometheus
errorCounter := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "errors_total",
        Help: "Total errors by type",
    },
    []string{"type", "code"},
)

// Registrar métrica
if domainErr, ok := err.(*domainerrors.DomainError); ok {
    errorCounter.WithLabelValues(domainErr.ErrorType, domainErr.Code).Inc()
}
```

### Tracing

```go
// OpenTelemetry
span := trace.SpanFromContext(ctx)
span.SetStatus(codes.Error, err.Error())
span.RecordError(err)

// Adicionar ao erro
domainErr.WithMetadata("trace_id", span.SpanContext().TraceID().String())
```

### Logging

```go
// Structured logging
log.Error("Operation failed",
    zap.String("error_code", domainErr.Code),
    zap.String("error_type", domainErr.ErrorType),
    zap.Any("metadata", domainErr.Metadata()),
    zap.String("stack_trace", domainErr.StackTrace()),
)
```

## 🔍 Debugging

### Stack Trace

```go
err := domainerrors.New("DEBUG_001", "Erro para debug")
fmt.Printf("Stack trace:\n%s\n", err.StackTrace())
```

### Cadeia de Erros

```go
// Encadear erros
err1 := errors.New("erro original")
err2 := domainerrors.Wrap("contexto adicional", err1)
err3 := domainerrors.Wrap("mais contexto", err2)

// Analisar cadeia
fmt.Printf("Cadeia: %s\n", domainerrors.FormatErrorChain(err3))
fmt.Printf("Causa raiz: %s\n", domainerrors.GetRootCause(err3))
```

## 📚 Documentação Adicional

- [Exemplos Práticos](examples/README.md)
- [API Reference](docs/api.md)
- [Guia de Migração](docs/migration.md)
- [Melhores Práticas](docs/best-practices.md)

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.

## 🙏 Agradecimentos

- Inspirado nas melhores práticas da comunidade Go
- Baseado nos princípios de Domain-Driven Design
- Influenciado por bibliotecas como `pkg/errors` e `go-kit`

---

**Desenvolvido com ❤️ em Go**
