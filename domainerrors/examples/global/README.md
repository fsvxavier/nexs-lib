# Exemplo Global - Domain Errors

Este exemplo demonstra como usar hooks e middlewares globais do sistema de domainerrors para processamento de erros em toda a aplicação.

## Funcionalidades Demonstradas

### 1. Hooks Globais
- **Start Hook**: Executado quando o sistema inicia
- **Stop Hook**: Executado quando o sistema para
- **Error Hook**: Executado sempre que um erro é processado
- **I18n Hook**: Executado para processamento de internacionalização

### 2. Middlewares Globais
- **Middleware Geral**: Processa todos os erros adicionando metadados
- **Middleware I18n**: Traduz mensagens de erro baseado no locale

### 3. Fluxo de Processamento
1. Hooks de início são executados
2. Um erro de validação é criado
3. Middlewares globais processam o erro
4. Hooks de erro são executados
5. Middleware de i18n traduz para diferentes locales
6. Hooks de i18n são executados para cada locale
7. Estatísticas são exibidas
8. Hooks de parada são executados

## Como Executar

```bash
cd examples/global
go run main.go
```

Ou compile primeiro:

```bash
go build -o global-example main.go
./global-example
```

## Conceitos Importantes

### Hooks vs Middlewares
- **Hooks**: Executam ações paralelas sem modificar o erro (logging, notificações)
- **Middlewares**: Modificam e transformam o erro durante o processamento

### Registro Global
- Hooks e middlewares são registrados globalmente usando `init()`
- São aplicados automaticamente a todos os erros processados
- Útil para funcionalidades transversais como logging, audit, metrics

### Processamento de I18n
- Middlewares de i18n criam novas instâncias com mensagens traduzidas
- Mantêm metadados e contexto original
- Hooks de i18n executam ações complementares para cada locale

## Saída Esperada

O exemplo produzirá saída similar a:
```
=== Exemplo de Hooks e Middlewares Globais ===

1. Executando hooks de início:
🚀 Global Start Hook: Sistema iniciando...

2. Criando erro para demonstrar middlewares:
Erro original: Campo obrigatório não informado

3. Executando middlewares globais:
🔧 Global Middleware: Processando erro VALIDATION_ERROR

[... continuação da saída ...]
```

## Casos de Uso Reais

### Logging Global
```go
hooks.RegisterGlobalErrorHook(func(ctx context.Context, err interfaces.DomainErrorInterface) error {
    logger.Error("Error occurred", 
        zap.String("code", err.Code()),
        zap.String("message", err.Error()),
        zap.Any("metadata", err.Metadata()),
    )
    return nil
})
```

### Metrics e Monitoring
```go
middlewares.RegisterGlobalMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
    metrics.IncrementErrorCounter(err.Type().String(), err.Code())
    return next(err)
})
```

### Audit Trail
```go
hooks.RegisterGlobalErrorHook(func(ctx context.Context, err interfaces.DomainErrorInterface) error {
    auditService.LogError(ctx, err.Code(), err.Error(), err.Metadata())
    return nil
})
```
