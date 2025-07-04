# Biblioteca de Erros de Domínio (errordomain)

Esta biblioteca fornece uma estrutura robusta para tratamento de erros em aplicações Go, seguindo os princípios de Domain-Driven Design (DDD). Ela foi projetada para facilitar a criação, manipulação e categorização de erros, proporcionando um empilhamento rico e detalhado para depuração.

## Características

- **Categorização de erros** por tipos específicos (validação, negócios, infraestrutura, etc.)
- **Empilhamento de erros** com informações contextuais
- **Captura automática de stack trace** para facilitar a depuração
- **Mapeamento para códigos HTTP** para uso em APIs REST
- **Suporte para detalhes e metadados** específicos por tipo de erro
- **Utilitários** para manipulação e análise de erros

## Tipos de Erros

A biblioteca inclui diversos tipos específicos de erros para cobrir os cenários mais comuns em aplicações modernas:

### Erros Básicos
- **DomainError**: tipo base para todos os erros de domínio
- **ValidationError**: para erros de validação de entrada
- **NotFoundError**: para recursos não encontrados
- **BusinessError**: para violações de regras de negócio

### Erros de Infraestrutura
- **DatabaseError**: para erros de banco de dados
- **ExternalServiceError**: para falhas de integração com serviços externos
- **InfrastructureError**: para erros de infraestrutura geral
- **DependencyError**: para falhas de dependências externas

### Erros de Segurança e Autenticação
- **AuthenticationError**: para falhas de autenticação
- **AuthorizationError**: para problemas de autorização
- **SecurityError**: para violações de segurança e ameaças

### Erros de Performance e Recursos
- **TimeoutError**: para operações que excedem o tempo limite
- **RateLimitError**: para violações de limite de taxa
- **ResourceExhaustedError**: para recursos esgotados (memória, conexões, etc.)
- **CircuitBreakerError**: para quando circuit breakers estão abertos

### Erros de Dados e Processamento
- **SerializationError**: para falhas de serialização/deserialização
- **CacheError**: para problemas relacionados a cache
- **MigrationError**: para falhas durante migrações de dados

### Erros de Sistema
- **ConfigurationError**: para problemas de configuração
- **UnsupportedOperationError**: para operações não suportadas
- **BadRequestError**: para requisições mal formadas
- **ConflictError**: para conflitos de recursos (duplicação)
- **InvalidSchemaError**: para erros de validação de schema
- **UnsupportedMediaTypeError**: para tipos de mídia não suportados
- **ServerError**: para erros internos do servidor
- **UnprocessableEntityError**: para entidades que não podem ser processadas
- **ServiceUnavailableError**: para serviços temporariamente indisponíveis

### Erros de Workflow e Negócio
- **WorkflowError**: para erros em processos de negócio e workflows

## Novos Tipos Adicionados (2025)

Esta versão inclui novos tipos de erro que estavam presentes no pacote `domainerrors` mas faltavam no `errordomain`. Todos os novos tipos seguem o padrão de design fluente com métodos chainable para configuração.

### InvalidSchemaError
Para erros de validação de schema de dados, útil em APIs que validam entrada contra schemas JSON, XML ou outros formatos.

```go
// Exemplo básico
err := errordomain.NewInvalidSchemaError("Schema validation failed").
    WithSchemaInfo("user-schema", "v1.0").
    WithSchemaDetails(map[string][]string{
        "name": {"required field missing"},
        "age":  {"must be a positive number"},
        "email": {"invalid format", "missing domain"},
    })

// Helper function para criação rápida
err := errordomain.NewInvalidSchemaErrorFromDetails("user-schema", map[string][]string{
    "field1": {"error1", "error2"},
})

// Status HTTP: 400 Bad Request
```

### UnsupportedMediaTypeError
Para requisições HTTP com tipos de mídia não suportados, essencial em APIs REST que aceitam apenas formatos específicos.

```go
// Exemplo completo
err := errordomain.NewUnsupportedMediaTypeError("Media type not supported").
    WithMediaTypeInfo("text/plain", []string{"application/json", "application/xml"})

// Helper function
err := errordomain.NewUnsupportedMediaTypeErrorFromTypes("text/csv", []string{
    "application/json",
    "application/xml",
    "application/yaml",
})

// Status HTTP: 415 Unsupported Media Type
```

### ServerError
Para erros internos do servidor com contexto rico, incluindo informações de requisição e metadados para debugging.

```go
// Exemplo com metadados completos
err := errordomain.NewServerError("Database connection failed", originalErr).
    WithErrorCode("DB_CONN_001").
    WithRequestInfo("req-123", "corr-456").
    WithMetadata(map[string]any{
        "db_host": "localhost",
        "db_port": 5432,
        "connection_pool": "main",
        "retry_count": 3,
    })

// Helper function para código específico
err := errordomain.NewServerErrorWithCode("DB_001", "Database error", originalErr)

// Status HTTP: 500 Internal Server Error
```

### UnprocessableEntityError
Para entidades que não podem ser processadas devido a violações de regras de negócio ou validações complexas.

```go
// Exemplo completo com validações e regras de negócio
err := errordomain.NewUnprocessableEntityError("Entity validation failed").
    WithEntityInfo("User", "user-123").
    WithValidationErrors(map[string][]string{
        "email": {"invalid format", "already exists"},
        "age": {"must be 18 or older"},
        "phone": {"invalid country code"},
    }).
    WithBusinessRuleViolation("User must be verified before activation").
    WithBusinessRuleViolation("Premium features require subscription")

// Helper function para validações
err := errordomain.NewUnprocessableEntityErrorFromValidation(
    "User", 
    "user-123", 
    map[string][]string{
        "email": {"invalid"},
        "age": {"required"},
    },
)

// Status HTTP: 422 Unprocessable Entity
```

### ServiceUnavailableError
Para serviços temporariamente indisponíveis, com informações de retry e health checks.

```go
// Exemplo com informações de retry
err := errordomain.NewServiceUnavailableError(
    "payment-service", 
    "Service temporarily unavailable", 
    originalErr,
).
    WithServiceInfo("payment", "/health").
    WithRetryInfo("30s", "5 minutes")

// Helper function com retry info
err := errordomain.NewServiceUnavailableErrorWithRetry(
    "payment-service", 
    "30s", 
    originalErr,
)

// Status HTTP: 503 Service Unavailable
```

## Exemplos de Uso

### Criação de erros básicos

```go
// Erro simples
err := errordomain.New("E001", "Erro simples")

// Erro com causa
baseErr := errors.New("causa original")
err := errordomain.NewWithError("E002", "Falha ao processar", baseErr)

// Erro com tipo específico
err := errordomain.New("E003", "Acesso negado").WithType(errordomain.ErrorTypeAuthorization)
```

### Erros específicos

```go
// Erro de validação
validationErr := errordomain.NewValidationError("Dados inválidos", nil)
validationErr.WithField("email", "Email inválido")
validationErr.WithField("senha", "Senha muito curta")

// Erro de recurso não encontrado
notFoundErr := errordomain.NewNotFoundError("Usuário não encontrado").WithResource("user", "123")

// Erro de regra de negócio
businessErr := errordomain.NewBusinessError("INSUF_FUNDS", "Saldo insuficiente")

// Erro de conflito
conflictErr := errordomain.NewConflictError("Email já está em uso")
conflictErr.WithConflictingResource("user", "email duplicado")

// Erro de limite de taxa
rateLimitErr := errordomain.NewRateLimitError("Muitas tentativas")
rateLimitErr.WithRateLimit(100, 0, "2025-01-01T15:00:00Z", "60s")

// Erro de circuit breaker
circuitErr := errordomain.NewCircuitBreakerError("payment-api", "Serviço indisponível")
circuitErr.WithCircuitState("OPEN", 5)

// Erro de segurança
securityErr := errordomain.NewSecurityError("Acesso suspeito detectado")
securityErr.WithSecurityContext("login_brute_force", "HIGH")
securityErr.WithClientInfo("curl/7.68.0", "192.168.1.100")

// Erro de recurso esgotado
resourceErr := errordomain.NewResourceExhaustedError("memory", "Memória insuficiente")
resourceErr.WithResourceLimits(2048, 2048, "MB")

// Erro de dependência
depErr := errors.New("connection timeout")
dependencyErr := errordomain.NewDependencyError("elasticsearch", "Falha na busca", depErr)
dependencyErr.WithDependencyInfo("search_engine", "7.10.0", "UNHEALTHY")

// Erro de serialização
serErr := errors.New("invalid JSON")
serializationErr := errordomain.NewSerializationError("JSON", "Falha ao serializar", serErr)
serializationErr.WithTypeInfo("user.age", "int", "string")

// Erro de cache
cacheErr := errordomain.NewCacheError("redis", "GET", "Cache indisponível", err)
cacheErr.WithCacheDetails("user:123", "300s")

// Erro de workflow
workflowErr := errordomain.NewWorkflowError("order-process", "payment", "Falha no pagamento")
workflowErr.WithStateInfo("pending_payment", "completed_payment")
```

### Empilhamento de erros

```go
// Cria um erro base
baseErr := errors.New("erro na consulta SQL")

// Adiciona contexto em camadas
dbErr := errordomain.NewDatabaseError("Falha ao buscar usuário", baseErr)
    .WithOperation("SELECT", "users")

serviceErr := errordomain.New("SERVICE_ERROR", "Falha no serviço de usuários")
    .Wrap("Ao buscar perfil de usuário", dbErr)

// O erro final contém todo o stack de informações
```

### Verificação de tipos de erro

```go
// Verificações básicas
if errordomain.IsNotFoundError(err) {
    // Trata erro de recurso não encontrado
}

if errordomain.IsValidationError(err) {
    // Trata erro de validação
}

if errordomain.IsBusinessError(err) {
    // Trata erro de regra de negócio
}

// Verificações de novos tipos
if errordomain.IsConflictError(err) {
    // Trata erro de conflito
}

if errordomain.IsRateLimitError(err) {
    // Trata erro de limite de taxa
}

if errordomain.IsSecurityError(err) {
    // Trata erro de segurança
}

if errordomain.IsResourceExhaustedError(err) {
    // Trata erro de recurso esgotado
}

if errordomain.IsCircuitBreakerError(err) {
    // Trata erro de circuit breaker
}

if errordomain.IsDependencyError(err) {
    // Trata erro de dependência
}

if errordomain.IsSerializationError(err) {
    // Trata erro de serialização
}

if errordomain.IsCacheError(err) {
    // Trata erro de cache
}

if errordomain.IsWorkflowError(err) {
    // Trata erro de workflow
}

if errordomain.IsMigrationError(err) {
    // Trata erro de migração
}

// Novos tipos
if errordomain.IsInvalidSchemaError(err) {
    // Trata erro de schema inválido
}

if errordomain.IsUnsupportedMediaTypeError(err) {
    // Trata erro de tipo de mídia não suportado
}

if errordomain.IsServerError(err) {
    // Trata erro interno do servidor
}

if errordomain.IsUnprocessableEntityError(err) {
    // Trata erro de entidade não processável
}

if errordomain.IsServiceUnavailableError(err) {
    // Trata erro de serviço indisponível
}
```

### Extraindo informações de status HTTP

```go
statusCode := err.StatusCode() // Retorna o código HTTP correspondente ao tipo de erro
```

### Registro de códigos de erro

```go
registry := errordomain.NewErrorCodeRegistry()
registry.Register("AUTH001", "Credenciais inválidas", http.StatusUnauthorized)
registry.Register("AUTH002", "Token expirado", http.StatusUnauthorized)

// Usar código registrado
err := registry.WrapWithCode("AUTH001", baseErr)
```

### Recuperação de pânicos

```go
err := errordomain.RecoverMiddleware(func() error {
    // Código que pode causar pânico
    return nil
})

// err conterá um DomainError se ocorrer um pânico
```

## Mapeamento de Códigos HTTP

Todos os tipos de erro implementam a interface `HttpStatusProvider` e retornam códigos HTTP apropriados:

```go
// Função GetStatusCode retorna o código HTTP correto para qualquer erro
statusCode := errordomain.GetStatusCode(err)

// Exemplos de mapeamento:
// ValidationError         → 400 Bad Request
// InvalidSchemaError      → 400 Bad Request  
// BadRequestError         → 400 Bad Request
// UnauthorizedError       → 401 Unauthorized
// AuthenticationError     → 401 Unauthorized
// AuthorizationError      → 403 Forbidden
// ForbiddenError          → 403 Forbidden
// SecurityError           → 403 Forbidden
// NotFoundError           → 404 Not Found
// ConflictError           → 409 Conflict
// UnsupportedMediaTypeError → 415 Unsupported Media Type
// UnprocessableEntityError → 422 Unprocessable Entity
// BusinessError           → 422 Unprocessable Entity
// SerializationError      → 422 Unprocessable Entity
// WorkflowError           → 422 Unprocessable Entity
// RateLimitError          → 429 Too Many Requests
// InfrastructureError     → 500 Internal Server Error
// DatabaseError           → 500 Internal Server Error
// ServerError             → 500 Internal Server Error
// ConfigurationError      → 500 Internal Server Error
// CacheError              → 500 Internal Server Error
// MigrationError          → 500 Internal Server Error
// ExternalServiceError    → 502 Bad Gateway (ou código específico)
// DependencyError         → 424 Failed Dependency
// ServiceUnavailableError → 503 Service Unavailable
// CircuitBreakerError     → 503 Service Unavailable
// UnsupportedOperationError → 501 Not Implemented
// TimeoutError            → 408 Request Timeout
// ResourceExhaustedError  → 507 Insufficient Storage
```

### Uso em Handlers HTTP

```go
func handleError(c *fiber.Ctx, err error) error {
    statusCode := errordomain.GetStatusCode(err)
    
    return c.Status(statusCode).JSON(fiber.Map{
        "error": err.Error(),
        "code": statusCode,
    })
}
```

## Integração com Frameworks HTTP

### Fiber
```go
func errorHandler(c *fiber.Ctx, err error) error {
    statusCode := errordomain.GetStatusCode(err)
    
    // Para erros de validação, inclua detalhes
    if validationErr, ok := err.(*errordomain.ValidationError); ok {
        return c.Status(statusCode).JSON(fiber.Map{
            "error": "Validation failed",
            "details": validationErr.ValidatedFields,
        })
    }
    
    // Para erros de schema inválido
    if schemaErr, ok := err.(*errordomain.InvalidSchemaError); ok {
        return c.Status(statusCode).JSON(fiber.Map{
            "error": "Schema validation failed",
            "schema": schemaErr.SchemaName,
            "details": schemaErr.Details,
        })
    }
    
    return c.Status(statusCode).JSON(fiber.Map{
        "error": err.Error(),
    })
}
```

### Echo
```go
func errorHandler(err error, c echo.Context) {
    statusCode := errordomain.GetStatusCode(err)
    
    response := map[string]interface{}{
        "error": err.Error(),
    }
    
    // Adicionar contexto específico para diferentes tipos
    switch e := err.(type) {
    case *errordomain.UnprocessableEntityError:
        response["entity_type"] = e.EntityType
        response["entity_id"] = e.EntityID
        response["validation_errors"] = e.ValidationErrors
        response["business_rules"] = e.BusinessRules
    case *errordomain.ServiceUnavailableError:
        response["service"] = e.ServiceName
        response["retry_after"] = e.RetryAfter
    case *errordomain.RateLimitError:
        response["limit"] = e.Limit
        response["remaining"] = e.Remaining
        response["reset_time"] = e.ResetTime
    }
    
    c.JSON(statusCode, response)
}
```

## Boas Práticas

### 1. Escolha o Tipo Correto de Erro
```go
// ❌ Usar erro genérico
err := errors.New("invalid email format")

// ✅ Usar tipo específico
err := errordomain.NewValidationError("Validation failed", nil).
    WithField("email", "invalid format")

// ✅ Para schemas específicos
err := errordomain.NewInvalidSchemaError("Schema validation failed").
    WithSchemaInfo("user-v1", "1.0")
```

### 2. Adicione Contexto em Cada Camada
```go
// Camada de dados
dbErr := errordomain.NewDatabaseError("Failed to insert user", sqlErr).
    WithOperation("INSERT", "users")

// Camada de serviço  
serviceErr := errordomain.NewServerError("User creation failed", dbErr).
    WithErrorCode("USR_001").
    WithRequestInfo(requestID, correlationID)

// Camada de API
if errordomain.IsUnprocessableEntityError(serviceErr) {
    return c.Status(422).JSON(fiber.Map{
        "error": "Cannot process user data",
        "request_id": requestID,
    })
}
```

### 3. Use Helper Functions para Casos Comuns
```go
// ❌ Criação manual complexa
err := errordomain.NewUnprocessableEntityError("Entity validation failed")
err.WithEntityInfo("User", userID)
err.WithValidationErrors(validationMap)

// ✅ Use helper function
err := errordomain.NewUnprocessableEntityErrorFromValidation("User", userID, validationMap)
```

### 4. Mantenha Códigos Consistentes
```go
// Defina constantes para códigos de erro
const (
    ErrCodeInsufficientFunds = "PAY_001"
    ErrCodeInvalidCard       = "PAY_002"
    ErrCodeServiceDown       = "PAY_003"
)

// Use em toda a aplicação
err := errordomain.NewBusinessError(ErrCodeInsufficientFunds, "Insufficient account balance")
```

### 5. Configure Timeout e Retry Adequadamente
```go
// Para serviços externos
err := errordomain.NewServiceUnavailableError("payment-gateway", "Service timeout", originalErr).
    WithRetryInfo("30s", "5 minutes").
    WithServiceInfo("payment", "/health")

// Para rate limiting
err := errordomain.NewRateLimitError("Too many requests").
    WithRateLimit(1000, 0, time.Now().Add(time.Hour).Format(time.RFC3339), "3600s")
```

### 6. Preserve Stack Traces
```go
// ❌ Perder contexto original
return errordomain.New("PROC_ERROR", "Processing failed")

// ✅ Preserve o erro original
return errordomain.NewServerError("Processing failed", originalErr).
    WithErrorCode("PROC_ERROR")
```

### 7. Exponha Informações Apropriadas
```go
// Handler de API
func handleUserError(err error) Response {
    switch e := err.(type) {
    case *errordomain.ValidationError:
        // Seguro expor detalhes de validação
        return Response{
            Status: "validation_error",
            Details: e.ValidatedFields,
        }
    case *errordomain.ServerError:
        // ❌ Não exponha detalhes internos
        // return Response{Error: e.Metadata}
        
        // ✅ Exponha apenas o necessário
        return Response{
            Status: "internal_error",
            RequestID: e.RequestID,
        }
    }
}
```

### 8. Use Verificações de Tipo para Tratamento Específico
```go
func handleDependencyFailure(err error) {
    if errordomain.IsServiceUnavailableError(err) {
        serviceErr := err.(*errordomain.ServiceUnavailableError)
        // Implementar retry baseado em RetryAfter
        retryAfter := serviceErr.RetryAfter
        scheduleRetry(retryAfter)
    } else if errordomain.IsCircuitBreakerError(err) {
        // Circuit breaker aberto, usar fallback
        useFallbackService()
    }
}
```

## Changelog (2025)

### ✨ Novos Recursos
- **InvalidSchemaError**: Para validação de schemas com detalhes específicos
- **UnsupportedMediaTypeError**: Para tipos de mídia não suportados
- **ServerError**: Para erros internos com contexto rico
- **UnprocessableEntityError**: Para entidades não processáveis
- **ServiceUnavailableError**: Para serviços temporariamente indisponíveis

### 🔧 Melhorias
- Implementação consistente de `StatusCode()` em todos os tipos
- Helper functions para criação rápida de erros comuns
- Funções de verificação de tipo (`IsInvalidSchemaError`, etc.)
- Mapeamento completo de códigos HTTP
- Documentação expandida com exemplos práticos

### 🎯 Compatibilidade
- ✅ Mantém 100% de compatibilidade com código existente
- ✅ Migração suave do pacote `domainerrors`
- ✅ Interfaces consistentes em todos os tipos

## Migração do domainerrors

Se você está migrando do pacote `domainerrors`, aqui está um guia de equivalências:

```go
// domainerrors → errordomain

// RepositoryError → DatabaseError ou InfrastructureError
old := domainerrors.RepositoryError{Description: "DB error"}
new := errordomain.NewDatabaseError("DB error", nil)

// ExternalIntegrationError → ExternalServiceError
old := domainerrors.ExternalIntegrationError{Code: 502}
new := errordomain.NewExternalServiceError("service", "error", nil).WithStatusCode(502)

// InvalidEntityError → ValidationError
old := domainerrors.InvalidEntityError{EntityName: "User"}
new := errordomain.NewValidationError("validation failed", nil)

// InvalidSchemaError → InvalidSchemaError (compatível)
old := domainerrors.InvalidSchemaError{}
new := errordomain.NewInvalidSchemaError("schema error")

// UsecaseError → BusinessError
old := domainerrors.UsecaseError{Code: "BIZ_001"}
new := errordomain.NewBusinessError("BIZ_001", "business rule violated")

// ServerError → ServerError (compatível)
old := domainerrors.ServerError{Description: "server error"}
new := errordomain.NewServerError("server error", nil)

// UnprocessableEntity → UnprocessableEntityError (compatível)
old := domainerrors.UnprocessableEntity{Description: "cannot process"}
new := errordomain.NewUnprocessableEntityError("cannot process")

// ErrTargetServiceUnavailable → ServiceUnavailableError
old := domainerrors.ErrTargetServiceUnavailable{}
new := errordomain.NewServiceUnavailableError("service", "unavailable", nil)
```

## Integrações Avançadas

### OpenTelemetry
```go
import "go.opentelemetry.io/otel/trace"

func traceError(ctx context.Context, err error) {
    span := trace.SpanFromContext(ctx)
    
    span.SetStatus(codes.Error, err.Error())
    span.SetAttributes(
        attribute.String("error.type", reflect.TypeOf(err).String()),
    )
    
    if serverErr, ok := err.(*errordomain.ServerError); ok {
        span.SetAttributes(
            attribute.String("error.code", serverErr.ErrorCode),
            attribute.String("request.id", serverErr.RequestID),
        )
    }
}
```

### Logging Estruturado
```go
import "go.uber.org/zap"

func logStructuredError(err error) {
    fields := []zap.Field{
        zap.String("error_type", reflect.TypeOf(err).String()),
        zap.String("error_message", err.Error()),
    }
    
    switch e := err.(type) {
    case *errordomain.UnprocessableEntityError:
        fields = append(fields,
            zap.String("entity_type", e.EntityType),
            zap.String("entity_id", e.EntityID),
            zap.Any("validation_errors", e.ValidationErrors),
        )
    case *errordomain.ServiceUnavailableError:
        fields = append(fields,
            zap.String("service_name", e.ServiceName),
            zap.String("retry_after", e.RetryAfter),
        )
    }
    
    logger.Error("Application error", fields...)
}
```

## Contribuição

Contribuições são bem-vindas! Para contribuir:

1. **Fork** o repositório
2. **Crie** uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. **Commit** suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. **Push** para a branch (`git push origin feature/nova-funcionalidade`)
5. **Abra** um Pull Request

### Diretrizes de Contribuição

- Mantenha compatibilidade com versões anteriores
- Adicione testes para novas funcionalidades
- Documente novos tipos de erro com exemplos
- Siga as convenções de nomenclatura existentes
- Implemente a interface `HttpStatusProvider` em novos tipos

### Reportando Issues

Ao reportar bugs ou solicitar features, inclua:

- Versão do Go utilizada
- Exemplo de código que reproduz o problema
- Comportamento esperado vs. atual
- Logs relevantes (sem informações sensíveis)

---

**Versão**: 2025.1  
**Licença**: MIT  
**Autor**: Fabrício Xavier  
**Repositório**: [isis-golang-lib](https://github.com/fsvxavier/nexs-lib)
