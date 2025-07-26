# Exemplo Avançado - Hooks Customizados

Este exemplo demonstra como implementar e usar hooks customizados para auditoria, validação, métricas e logging avançado.

## Como Executar

```bash
cd examples/hooks
go run main.go
```

## Conceitos Demonstrados

### 1. Sistema de Hooks
O módulo decimal suporta três tipos de hooks:
- **PreHook**: Executado antes das operações
- **PostHook**: Executado após operações bem-sucedidas
- **ErrorHook**: Executado quando ocorrem erros

### 2. Hooks Customizados Implementados

#### Hook de Auditoria (`CustomAuditHook`)
- **Propósito**: Registrar todas as operações para auditoria
- **Funcionalidades**:
  - Log de timestamp para cada operação
  - Registro de argumentos e resultados
  - Medição de tempo de execução
  - Trilha completa de auditoria

#### Hook de Validação (`CustomValidationHook`)
- **Propósito**: Validar valores dentro de limites específicos
- **Funcionalidades**:
  - Validação de valores mínimos e máximos
  - Verificação de argumentos e resultados
  - Prevenção de operações com valores inválidos
  - Mensagens de erro descritivas

#### Hook de Métricas (`CustomMetricsHook`)
- **Propósito**: Coletar métricas de performance
- **Funcionalidades**:
  - Contagem de operações por tipo
  - Medição de tempo de execução
  - Cálculo de tempo médio por operação
  - Relatórios de performance

## Implementação de Hooks Customizados

### Interface de Pre-Hook
```go
type PreHook interface {
    Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error)
}
```

### Interface de Post-Hook
```go
type PostHook interface {
    Execute(ctx context.Context, operation string, result interface{}, err error) error
}
```

### Interface de Error-Hook
```go
type ErrorHook interface {
    Execute(ctx context.Context, operation string, err error) error
}
```

## Casos de Uso Práticos

### 1. Auditoria para Compliance
```go
type AuditEntry struct {
    Timestamp time.Time
    Operation string
    Args      []interface{}
    Result    interface{}
    Error     error
    Duration  time.Duration
}
```

**Benefícios:**
- Rastreabilidade completa
- Compliance com regulamentações
- Debugging avançado
- Análise forense

### 2. Validação de Negócio
```go
func (h *CustomValidationHook) Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
    // Validações customizadas baseadas em regras de negócio
    // Exemplo: limites de transação, moedas permitidas, etc.
}
```

**Benefícios:**
- Prevenção de valores inválidos
- Aplicação de regras de negócio
- Segurança de dados
- Consistência do sistema

### 3. Monitoramento de Performance
```go
type MetricsHook struct {
    operationCounts map[string]int
    operationTimes  map[string]time.Duration
}
```

**Benefícios:**
- Identificação de gargalos
- Otimização de performance
- Alertas proativos
- Análise de tendências

## Configuração e Uso

### Habilitando Hooks
```go
cfg := &config.Config{
    ProviderName: "shopspring",
    HooksEnabled: true,  // Necessário para ativar hooks
}

manager := decimal.NewManager(cfg)
```

### Registrando Hooks Customizados
```go
// Exemplo conceitual - implementação real requereria extensão da API
hookManager := manager.GetHookManager()
hookManager.RegisterPreHook(NewCustomAuditHook())
hookManager.RegisterPostHook(NewCustomValidationHook())
hookManager.RegisterErrorHook(NewCustomMetricsHook())
```

## Saída Esperada

### Auditoria
```
Demonstrando hooks de auditoria (simulado)
Criado decimal: 100.50
Criado decimal: 25.75
Resultado da soma: 126.25
Resultado da divisão: 3.902912621359223

Log de auditoria simulado:
- NewFromString("100.50") -> sucesso
- NewFromString("25.75") -> sucesso
- Add(100.50, 25.75) -> 126.25
- Div(100.50, 25.75) -> 3.902912621359223
```

### Validação
```
Demonstrando hooks de validação (simulado)
Testando validação com limites: -1000 a 1000

Teste 1: Valores dentro dos limites
Valor válido: 500
Valor válido: 300
Soma válida: 800

Teste 2: Valores fora dos limites
Tentando criar valor 2000 (acima do limite)...
VALIDAÇÃO: Erro - valor 2000 está acima do máximo permitido 1000
```

### Métricas
```
Métricas simuladas:
NewFromString: 5 operações, tempo médio: 1.2µs
Add: 4 operações, tempo médio: 0.8µs
Mul: 4 operações, tempo médio: 0.9µs
Div: 4 operações, tempo médio: 1.5µs
Sum: 1 operação, tempo médio: 3.2µs
```

### Hooks Combinados
```
Simulando execução com múltiplos hooks:
- Hook de Auditoria: ATIVO
- Hook de Validação: ATIVO (limites: -10000 a 10000)
- Hook de Métricas: ATIVO
- Hook de Logging: ATIVO

[AUDIT] NewFromString("1000.00") registrado
[VALIDATION] Valor 1000.00 está dentro dos limites
[METRICS] NewFromString executado em 1.1µs
[LOGGING] Decimal criado com sucesso: 1000.00
```

## Padrões de Implementação

### 1. Context Usage
```go
func (h *CustomHook) Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
    // Use context para passar dados entre hooks
    start := time.Now()
    ctx = context.WithValue(ctx, "start_time", start)
    return nil, nil
}
```

### 2. Error Handling
```go
func (h *CustomHook) ExecuteError(ctx context.Context, operation string, err error) error {
    // Log estruturado de erros
    log.WithFields(log.Fields{
        "operation": operation,
        "error":     err,
        "timestamp": time.Now(),
    }).Error("Operation failed")
    return nil
}
```

### 3. Performance Monitoring
```go
func (h *MetricsHook) Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
    h.operationCounts[operation]++
    ctx = context.WithValue(ctx, "metrics_start", time.Now())
    return nil, nil
}
```

## Extensões Futuras

### 1. Hook Registry
- Sistema de registro dinâmico de hooks
- Priorização de execução
- Configuração via arquivo

### 2. Conditional Hooks
- Hooks baseados em condições
- Filtros por operação ou valor
- Ativação dinâmica

### 3. Async Hooks
- Processamento assíncrono
- Hooks não bloqueantes
- Queue de eventos

### 4. Distributed Hooks
- Hooks distribuídos
- Coleta centralizada de métricas
- Auditoria cross-service

## Próximos Passos

- Veja o exemplo básico para começar
- Explore o exemplo de providers para comparações
- Consulte a documentação de arquitetura no README principal
- Implemente seus próprios hooks customizados

## Considerações de Performance

- **Hooks custam performance**: Use apenas quando necessário
- **Async quando possível**: Para operações de I/O
- **Context é sua ferramenta**: Para passar dados entre hooks
- **Teste thoroughly**: Hooks podem quebrar operações
