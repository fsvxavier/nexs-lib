# Exemplo Avançado - Domain Errors

Este exemplo demonstra padrões avançados e funcionalidades empresariais do sistema de domainerrors, incluindo métricas, audit, circuit breakers e processamento complexo de middlewares.

## Funcionalidades Demonstradas

### 1. Componentes Avançados

#### ErrorMetrics
- Sistema thread-safe de métricas de erro
- Contadores por tipo de erro
- Análise estatística em tempo real

#### AuditLogger
- Sistema de audit trail para compliance
- Registro detalhado com contexto
- Armazenamento thread-safe de logs

#### CircuitBreaker
- Implementação de circuit breaker pattern
- Proteção contra falhas em cascata
- Estados: closed, open, half-open

### 2. Hooks Avançados

#### Hook de Inicialização
```go
hooks.RegisterGlobalStartHook(func(ctx context.Context) error {
    // Verificação de dependências
    // Validação de configuração
    // Inicialização de componentes
})
```

#### Hook de Finalização
```go
hooks.RegisterGlobalStopHook(func(ctx context.Context) error {
    // Cleanup de recursos
    // Relatórios finais
    // Graceful shutdown
})
```

#### Hook de Erro com Métricas
```go
hooks.RegisterGlobalErrorHook(func(ctx context.Context, err interfaces.DomainErrorInterface) error {
    // Incrementar métricas
    // Atualizar circuit breaker
    // Log com severity
})
```

### 3. Middlewares Avançados

#### Context Enrichment Middleware
- Adiciona request_id, user_id, correlation_id
- Enriquece com metadados de ambiente
- Timestamp de processamento

#### Rate Limiting Middleware
- Aplica rate limiting por tipo de erro
- Transforma erros quando limite excedido
- Políticas configuráveis

#### Audit Middleware
- Registra todos os erros em audit trail
- Contextualização completa
- Compliance e rastreabilidade

#### I18n Avançado
- Tradução com fallback
- Confiança na tradução
- Detecção automática de locale

### 4. Classificação de Erros

#### Por Criticidade
- **LOW**: Validation, BadRequest
- **MEDIUM**: NotFound, Authentication  
- **HIGH**: Business, Authorization
- **CRITICAL**: Database, ExternalService, Security

#### Por Impacto no Circuit Breaker
Erros críticos contribuem para abertura do circuit breaker:
- DatabaseError
- ExternalServiceError
- InfrastructureError
- SecurityError

### 5. Funcionalidades Empresariais

#### Correlation IDs
Cada erro recebe um correlation ID único para rastreamento distribuído.

#### Multi-tenancy
Suporte para user_id e context segregation.

#### Observability
- Métricas detalhadas
- Logs estruturados
- Tracing distribuído

#### Compliance
- Audit trail completo
- Retenção de logs
- Contexto regulatório

## Como Executar

```bash
cd examples/advanced
go run main.go
```

Ou compile primeiro:

```bash
go build -o advanced-example main.go
./advanced-example
```

## Saída Esperada

O exemplo processará 5 tipos diferentes de erro, demonstrando:

1. **Inicialização**: Verificação de dependências
2. **Processamento**: Middlewares em cadeia
3. **Métricas**: Contadores por tipo
4. **Circuit Breaker**: Estado e falhas
5. **Audit**: Trail completo
6. **I18n**: Tradução para múltiplos locales
7. **Finalização**: Cleanup e estatísticas

## Casos de Uso Reais

### 1. Sistema de E-commerce
```go
// Middleware para tracking de transações
middlewares.RegisterGlobalMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
    if err.Type() == interfaces.BusinessError {
        transactionService.RecordError(ctx, err.Code())
    }
    return next(err)
})
```

### 2. API Gateway
```go
// Hook para rate limiting global
hooks.RegisterGlobalErrorHook(func(ctx context.Context, err interfaces.DomainErrorInterface) error {
    if rateLimiter.ShouldBlock(ctx, err) {
        // Aplicar penalidade
        rateLimiter.ApplyPenalty(ctx)
    }
    return nil
})
```

### 3. Sistema Bancário
```go
// Middleware para compliance financeira
middlewares.RegisterGlobalMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
    if isFinancialOperation(ctx) {
        complianceLogger.RecordFinancialError(ctx, err)
    }
    return next(err)
})
```

### 4. Sistema de Saúde
```go
// Hook para HIPAA compliance
hooks.RegisterGlobalErrorHook(func(ctx context.Context, err interfaces.DomainErrorInterface) error {
    if containsPHI(err) {
        hipaaLogger.RecordPHIAccess(ctx, err)
    }
    return nil
})
```

## Padrões Implementados

- **Observer Pattern**: Para hooks e notificações
- **Chain of Responsibility**: Para middlewares
- **Circuit Breaker**: Para resiliência
- **Strategy Pattern**: Para diferentes tipos de erro
- **Decorator Pattern**: Para enriquecimento de contexto

## Métricas e Observabilidade

O exemplo coleta e exibe:
- Contadores por tipo de erro
- Estado do circuit breaker
- Logs de audit com contexto
- Estatísticas de tradução
- Tempos de processamento
