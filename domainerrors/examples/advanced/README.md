# Advanced Domain Errors Examples

Este exemplo demonstra o uso avançado da biblioteca de erros de domínio, incluindo tipos específicos de erro, encadeamento complexo e padrões de recuperação.

## Executar o Exemplo

```bash
cd examples/advanced
go run main.go
```

## Exemplos Incluídos

### 1. Tipos Específicos de Erro
- **ValidationError**: Erro de validação com campos específicos
- **BusinessError**: Erro de regra de negócio com regras violadas
- **DatabaseError**: Erro de banco de dados com operação e query
- **ExternalServiceError**: Erro de serviço externo com endpoint e resposta

### 2. Encadeamento com Contexto
- Como usar `WithContext()` para adicionar informações contextuais
- Propagação de informações de request através da cadeia de erros
- Stack trace detalhado mostrando toda a cadeia

### 3. Cadeias de Erro Complexas
- Simulação de erro em múltiplas camadas (database → repository → service → controller)
- Como cada camada adiciona seu próprio contexto
- Navegação através da cadeia de erros para análise

### 4. Tratamento de Erro em Camadas
- Padrão de erro em arquitetura em camadas
- Como diferentes camadas lidam com erros
- Mapeamento automático de tipos de erro para códigos HTTP

### 5. Metadados Personalizados
- Uso de `UnprocessableEntityError` com informações detalhadas
- Metadados customizados para contexto adicional
- Análise de erro com verificação de tipos

## Padrões Demonstrados

### Encadeamento de Erro
```go
// Erro base
originalErr := errors.New("network connection failed")

// Camada de infraestrutura
infraErr := domainerrors.NewinfraestructureError("redis", "Cache operation failed", originalErr)

// Camada de serviço
serviceErr := domainerrors.NewServerError("User service failed", infraErr)
serviceErr.WithContext(ctx, "processing user registration")
```

### Metadados Ricos
```go
err := domainerrors.NewUnprocessableEntityError("User entity validation failed")
err.WithEntityInfo("User", "user-12345")
err.WithValidationErrors(map[string][]string{
    "email": {"invalid format", "already exists"},
    "age":   {"must be 18 or older"},
})
err.WithBusinessRuleViolation("User must be verified")
```

### Análise de Erro
```go
// Verificar tipo de erro
if domainerrors.IsType(err, domainerrors.ErrorTypeTimeout) {
    // Implementar retry
}

// Mapear para HTTP status
statusCode := domainerrors.MapHTTPStatus(err)
```

## Padrões de Recuperação

### Retry Pattern
- Verificação se o erro é retryável
- Implementação de backoff exponencial
- Limite de tentativas

### Circuit Breaker
- Detecção de falhas consecutivas
- Abertura do circuito para proteção
- Recuperação gradual

### Fallback
- Alternativas quando serviços falham
- Degradação graceful
- Resiliência da aplicação

## Integração com Observabilidade

### Logging Estruturado
- Metadados para correlação
- Stack traces para debugging
- Contexto de request

### Métricas
- Contadores por tipo de erro
- Latência e timeout
- Taxa de erro por serviço

### Tracing
- Propagação de erro através de spans
- Correlação de traces
- Análise de performance

## Próximos Passos

- Veja exemplos específicos por tipo em `../types/`
- Explore integração com APIs em `../api/`
- Aprenda sobre observabilidade em `../observability/`
