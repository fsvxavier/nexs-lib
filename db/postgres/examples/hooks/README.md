# Exemplo de Sistema de Hooks

Este exemplo demonstra o poderoso sistema de hooks do PostgreSQL provider, que permite interceptar e customizar operações em tempo real.

## Funcionalidades Demonstradas

### 1. Hooks Básicos
- Logging automático de queries
- Medição de tempo de execução
- Interceptação de operações

### 2. Hooks de Performance
- Coleta de métricas em tempo real
- Detecção de queries lentas
- Análise de padrões suspeitos

### 3. Hooks de Auditoria
- Log completo de auditoria
- Detecção de operações sensíveis
- Validação de segurança

### 4. Hooks de Tratamento de Erros
- Categorização automática de erros
- Estatísticas de falhas
- Lógica de retry conceitual

### 5. Hooks Customizados
- Cache conceitual
- Rate limiting
- Métricas específicas

## Tipos de Hooks Suportados

### Hooks de Operação
- `BeforeQuery` / `AfterQuery`: Interceptam queries
- `BeforeExec` / `AfterExec`: Interceptam operações de modificação
- `BeforeTransaction` / `AfterTransaction`: Interceptam transações

### Hooks de Conexão
- `BeforeConnection` / `AfterConnection`: Interceptam conexões
- `BeforeAcquire` / `AfterAcquire`: Interceptam aquisição de conexões

### Hooks de Erro
- `OnError`: Interceptam erros para tratamento customizado

### Hooks Customizados
- `CustomHookBase + N`: Hooks definidos pelo usuário

## Como Executar

```bash
# Certifique-se de que o PostgreSQL está rodando
cd hooks/
go run main.go
```

## Exemplo de Saída

```
=== Exemplo de Sistema de Hooks ===

1. Conectando ao banco...
2. Configurando sistema de hooks...

3. Exemplo: Hooks básicos...
   Registrando hooks básicos...
   ✅ Hooks registrados com sucesso
   Testando hooks com queries...
   🔍 [LOG] Executando query: SELECT 1 as test
   ⏱️ [TIMING] query levou 2ms
   ✅ Query 1 executada com sucesso

4. Exemplo: Hooks de performance...
   🐌 [SLOW] Query lenta detectada: 155ms (threshold: 100ms)
   📊 Métricas de Performance:
   - Total de queries: 4
   - Tempo total: 200ms
   - Tempo médio: 50ms
   - Queries lentas: 1
```

## Casos de Uso

### 1. Observabilidade
```go
// Hook para métricas
metricsHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
    if ctx.Duration > 0 {
        metrics.RecordQuery(ctx.Operation, ctx.Duration)
    }
    return &postgres.HookResult{Continue: true}
}
```

### 2. Segurança
```go
// Hook para validação de segurança
securityHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
    if containsSecurityRisk(ctx.Query) {
        return &postgres.HookResult{
            Continue: false,
            Error: errors.New("query blocked by security policy"),
        }
    }
    return &postgres.HookResult{Continue: true}
}
```

### 3. Cache
```go
// Hook para cache
cacheHook := func(ctx *postgres.ExecutionContext) *postgres.HookResult {
    if isSelectQuery(ctx.Query) {
        if result := cache.Get(ctx.Query); result != nil {
            return &postgres.HookResult{
                Continue: false,
                Data: map[string]interface{}{"cached_result": result},
            }
        }
    }
    return &postgres.HookResult{Continue: true}
}
```

## Vantagens dos Hooks

- **Flexibilidade**: Customização sem modificar código core
- **Observabilidade**: Visibilidade completa do sistema
- **Segurança**: Validação e controle de acesso
- **Performance**: Monitoramento e otimização automática
- **Auditoria**: Rastreamento completo de operações

## Considerações de Performance

- Hooks são executados de forma síncrona
- Mantenha lógica de hooks simples e rápida
- Use hooks assíncronos para operações custosas
- Considere o overhead de múltiplos hooks

## Requisitos

- PostgreSQL rodando em `localhost:5432`
- Banco de dados `nexs_testdb`
- Usuário `nexs_user` com senha `nexs_password`
