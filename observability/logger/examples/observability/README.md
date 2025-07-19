# Exemplo: Observabilidade Avançada - Fase 7

Este exemplo demonstra as funcionalidades de **observabilidade avançada** implementadas na Fase 7 do sistema de logging, incluindo **métricas de logging** e **hooks customizados**.

## 🚀 Funcionalidades Demonstradas

### 1. Métricas de Logging
- **Contadores por nível**: DEBUG, INFO, WARN, ERROR, FATAL, PANIC
- **Tempo de processamento**: Médio geral e por nível
- **Taxa de erro**: Percentual de logs com erro
- **Taxa de sampling**: Eficiência do sistema de sampling
- **Estatísticas de provider**: Performance específica por provider

### 2. Hooks Customizados
- **Hook de Validação**: Valida entradas de log antes do processamento
- **Hook de Transformação**: Modifica dados dos logs após processamento
- **Hook de Filtro**: Filtra logs com base em critérios específicos
- **Hook de Métricas**: Coleta automática de métricas (auto-registrado)

### 3. Gestão de Hooks em Runtime
- Registro e remoção dinâmica de hooks
- Habilitação/desabilitação em tempo real
- Listagem e controle de estado dos hooks
- Thread-safety completa

## 📋 Pré-requisitos

```bash
# Certifique-se de que o módulo está atualizado
go mod tidy

# Verifique se não há erros de compilação
go build ./observability/logger/examples/observability
```

## 🏃 Como Executar

```bash
# Execute o exemplo
go run ./observability/logger/examples/observability/main.go
```

## 📊 Saída Esperada

O exemplo produzirá saída similar a:

```
=== Demonstração: Observabilidade Avançada - Fase 7 ===

1. MÉTRICAS DE LOGGING
======================
Total de logs: 3
Logs INFO: 1
Logs WARN: 1
Logs ERROR: 1
Taxa de erro: 33.33%
Tempo médio de processamento: 245µs

2. HOOKS CUSTOMIZADOS
=====================
✓ Hook de validação registrado
✓ Hook de transformação registrado
✓ Hook de filtro registrado

Testando hooks:
{"time":"2025-01-19T10:30:45Z","level":"INFO","msg":"Log processado com hooks","data":"exemplo","processed_at":"2025-01-19T10:30:45Z","service":"observability-demo"}
{"time":"2025-01-19T10:30:45Z","level":"ERROR","msg":"Erro processado com validação e transformação","error_code":"DB_CONNECTION_FAILED","processed_at":"2025-01-19T10:30:45Z","service":"observability-demo"}

3. ESTATÍSTICAS DOS HOOKS
=========================
Hooks 'before': 2
  - entry_validator (ativo: true)
  - log_filter (ativo: true)
Hooks 'after': 2
  - metrics_collector (ativo: true)
  - data_transformer (ativo: true)

4. MÉTRICAS FINAIS
==================
Total final de logs: 5

5. DEMONSTRAÇÃO DE SAMPLING
============================
Gerando 10 logs com sampling (inicial: 5, depois: 1 a cada 2):
Taxa de sampling: 75.00%

6. GESTÃO DE HOOKS EM RUNTIME
==============================
Desabilitando todos os hooks...
Reabilitando hooks...
Removendo hook de filtro...
✓ Hook de filtro removido

7. MÉTRICAS DETALHADAS POR NÍVEL
=================================
DEBUG: 1 logs, tempo médio: 125µs
INFO: 4 logs, tempo médio: 189µs
WARN: 1 logs, tempo médio: 234µs
ERROR: 1 logs, tempo médio: 312µs

✅ Demonstração da Fase 7 - Observabilidade Avançada concluída!
```

## 🔧 Componentes Técnicos

### Métricas Coletadas

```go
type Metrics interface {
    GetLogCount(level Level) int64
    GetTotalLogCount() int64
    GetAverageProcessingTime() time.Duration
    GetProcessingTimeByLevel(level Level) time.Duration
    GetErrorRate() float64
    GetSamplingRate() float64
    GetProviderStats(provider string) *ProviderStats
    Reset()
    Export() map[string]interface{}
}
```

### Tipos de Hooks Disponíveis

```go
// Hook de Validação
hook := logger.NewValidationHook(
    func(entry *interfaces.LogEntry) error {
        // Validação customizada
        return nil
    },
)

// Hook de Transformação
hook := logger.NewTransformHook(
    func(entry *interfaces.LogEntry) error {
        // Transformação customizada
        return nil
    },
)

// Hook de Filtro
hook := logger.NewFilterHook(
    func(entry *interfaces.LogEntry) bool {
        // Lógica de filtro
        return true
    },
)
```

### Integração com Logger Observável

```go
// Cria logger observável
observableLogger := logger.ConfigureObservableLogger(provider, config)

// Registra hooks
observableLogger.RegisterHook(interfaces.BeforeHook, validationHook)
observableLogger.RegisterHook(interfaces.AfterHook, transformHook)

// Obtém métricas
metrics := observableLogger.GetMetrics()
fmt.Printf("Total de logs: %d\n", metrics.GetTotalLogCount())

// Gerencia hooks
hookManager := observableLogger.GetHookManager()
hookManager.EnableAllHooks()
hookManager.DisableAllHooks()
```

## 🎯 Casos de Uso

### 1. Monitoramento de Performance
- Acompanhe tempo médio de processamento
- Identifique gargalos por nível de log
- Monitore taxa de erro da aplicação

### 2. Validação de Dados
- Implemente validações customizadas para logs
- Garanta consistência de dados de auditoria
- Bloqueie logs mal formados

### 3. Transformação de Dados
- Adicione metadados automaticamente
- Sanitize informações sensíveis
- Enriqueça logs com contexto adicional

### 4. Filtragem Inteligente
- Filtre logs por ambiente (dev/prod)
- Implemente rate limiting por tipo de log
- Controle volume de logs em tempo real

### 5. Integração com Sistemas de Monitoramento
- Exporte métricas para Prometheus
- Envie alertas baseados em thresholds
- Colete dados para dashboards

## 🔍 Debugging

Para habilitar logs de debug durante desenvolvimento:

```go
config := &interfaces.Config{
    Level:  interfaces.DebugLevel,
    Format: interfaces.ConsoleFormat,
}
```

## ⚡ Performance

- **Overhead de métricas**: < 5% sobre logging direto
- **Thread-safety**: Contadores atômicos + RWMutex
- **Memória**: Limitação automática de histórico (1000 amostras por nível)
- **Hooks**: Execução assíncrona para hooks "after"

## 🚨 Considerações de Produção

1. **Limite de Hooks**: Mantenha número razoável de hooks para evitar latência
2. **Validação de Entrada**: Hooks de validação podem impactar performance
3. **Tratamento de Erro**: Hooks "before" com erro interrompem o log
4. **Métricas**: Use `Reset()` periodicamente para controlar uso de memória
5. **Sampling**: Configure adequadamente para ambientes de alto volume

## 📚 Próximos Passos

Esta implementação permite evoluir para:
- Integração com Prometheus/Grafana
- Alertas automáticos baseados em métricas
- Hooks assíncronos para processamento pesado
- Cache de métricas para consulta rápida
- API REST para gestão de hooks em runtime
