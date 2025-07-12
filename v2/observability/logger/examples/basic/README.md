# Exemplo Básico - Logger v2

Este exemplo demonstra o uso básico do sistema de logging v2 da nexs-lib.

## Funcionalidades Demonstradas

- ✅ Configuração básica do logger
- ✅ Registro de providers padrão
- ✅ Diferentes níveis de log (Debug, Info, Warn, Error)
- ✅ Logging com formatação (printf-style)
- ✅ Verificação de níveis habilitados
- ✅ Mudança dinâmica de nível de log
- ✅ Ciclo de vida do logger (flush, close)

## Como Executar

```bash
cd basic
go run main.go
```

## Output Esperado

O exemplo produzirá logs em formato console mostrando:
1. Logs em diferentes níveis
2. Logs formatados com variáveis
3. Verificação de níveis
4. Comportamento após mudança de nível

## Código Principal

```go
// Configuração básica
config := logger.DefaultConfig()
config.ServiceName = "basic-example"
config.Format = interfaces.ConsoleFormat
config.Level = interfaces.DebugLevel

// Criação do logger
factory := logger.NewFactory()
factory.RegisterDefaultProviders()
logger, err := factory.CreateLogger("basic", config)

// Uso básico
logger.Info(ctx, "Aplicação iniciada com sucesso")
logger.Warnf(ctx, "Usuário %d executou operação: %s", userID, operation)

// Verificação de nível
if logger.IsLevelEnabled(interfaces.DebugLevel) {
    logger.Debug(ctx, "Esta mensagem será processada")
}

// Mudança dinâmica
logger.SetLevel(interfaces.WarnLevel)
```

## Conceitos Aprendidos

1. **Configuração Simples**: Como configurar um logger com valores padrão
2. **Factory Pattern**: Uso da factory para criação de loggers
3. **Níveis de Log**: Como usar diferentes níveis apropriadamente
4. **Formatação**: Logging com interpolação de variáveis
5. **Performance**: Verificação de nível antes de processar logs
6. **Lifecycle**: Importância do flush e close

## Próximos Passos

Após dominar este exemplo, prossiga para:
- [Structured Logging](../structured/) - Para logs mais ricos
- [Context-Aware](../context-aware/) - Para aplicações distribuídas
- [Async](../async/) - Para aplicações de alta performance
