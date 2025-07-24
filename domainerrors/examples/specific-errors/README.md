# Specific Error Types Examples

Este exemplo demonstra como usar todos os tipos específicos de erro disponíveis no módulo `domainerrors`.

## Tipos de Erro Demonstrados

### 1. Validation Errors
- Erros de validação de dados
- Múltiplos campos com erros
- Mensagens específicas por campo

### 2. Business Errors
- Violações de regras de negócio
- Códigos de negócio customizados
- Múltiplas regras violadas

### 3. Database Errors
- Erros de conexão com banco de dados
- Operações e queries com falha
- Violações de restrições

### 4. External Service Errors
- Erros de serviços externos
- Informações de endpoint e resposta
- Códigos de status HTTP

### 5. Authentication/Authorization Errors
- Erros de autenticação
- Erros de autorização
- Erros de segurança

### 6. Timeout and Circuit Breaker Errors
- Erros de timeout
- Erros de circuit breaker
- Informações de duração e estado

### 7. Rate Limiting Errors
- Erros de limite de taxa
- Erros de recursos esgotados
- Informações de limites e quotas

### 8. Infrastructure Errors
- Erros de infraestrutura
- Erros de dependências
- Informações de componentes

### 9. Workflow Errors
- Erros de workflow
- Erros de serviço indisponível
- Informações de estado e transições

### 10. Cache and Configuration Errors
- Erros de cache
- Erros de configuração
- Erros de migração
- Erros de serialização

## Como executar

```bash
cd domainerrors/examples/specific-errors
go run main.go
```

## Características dos Erros

Cada tipo de erro possui:
- **Código personalizado**: Definido pelo usuário
- **Mensagem descritiva**: Explicação do erro
- **Contexto específico**: Informações relevantes para cada tipo
- **Fluent interface**: Métodos para adicionar informações extras
- **Mapeamento HTTP**: Status code apropriado para cada tipo

## Exemplo de Uso

```go
// Erro de validação com múltiplos campos
err := domainerrors.NewValidationError("USER_VALIDATION_FAILED", "User data validation failed", nil)
err.WithField("email", "invalid email format")
err.WithField("age", "must be 18 or older")

// Erro de negócio com regras específicas
businessErr := domainerrors.NewBusinessError("INSUFFICIENT_BALANCE", "Account balance insufficient")
businessErr.WithRule("Minimum balance of $10 required")
businessErr.WithRule("Account must be active")

// Erro de banco de dados com contexto
dbErr := domainerrors.NewDatabaseError("DB_CONNECTION_FAILED", "Database connection failed", connErr)
dbErr.WithOperation("SELECT", "users")
dbErr.WithQuery("SELECT * FROM users WHERE email = ?")
```

## Benefícios

- **Códigos consistentes**: Padronização de códigos de erro
- **Contexto rico**: Informações detalhadas para debugging
- **Tipo-específico**: Cada tipo tem suas próprias funcionalidades
- **HTTP-friendly**: Mapeamento automático para códigos HTTP
- **Flexibilidade**: Códigos customizados para diferentes contextos
