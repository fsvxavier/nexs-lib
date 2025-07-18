# Exemplo Básico de Logging

Este exemplo demonstra o uso básico do sistema de logging multi-provider.

## Executando o Exemplo

```bash
cd examples/basic
go run main.go
```

## Funcionalidades Demonstradas

### 1. Multi-Provider System
- **Slog**: Provider padrão da biblioteca Go
- **Zap**: Provider de alta performance (padrão do sistema)
- **Zerolog**: Provider otimizado para baixo consumo de memória
- Troca dinâmica entre providers

### 2. Provider Padrão (Zap)
- Configuração automática
- Formato JSON estruturado
- Performance otimizada (~240k logs/sec)
- Não requer configuração explícita

### 3. Logging Estruturado
- Campos tipados (String, Int, Bool, Float64, Duration, Time)
- Grupos de campos aninhados
- Context-aware logging
- Extração automática de trace_id, span_id, user_id, request_id

### 4. Diferentes Níveis de Log
- Debug, Info, Warn, Error
- Configuração por provider
- Filtragem por nível

### 5. Comparação de Providers
- Demonstração de cada provider
- Comparação de output
- Verificação de performance

### 6. Context-Aware Logging
- Extração automática de campos do contexto
- Logging com contexto enriquecido
- Rastreamento distribuído

## Saída Esperada

O exemplo produz logs em formato JSON estruturado para cada provider:

### Zap (Padrão)
```json
{"level":"info","time":"2025-07-18T10:30:45Z","msg":"Testando provider","provider":"zap"}
```

### Slog
```json
{"time":"2025-07-18T10:30:45Z","level":"INFO","msg":"Testando provider","provider":"slog"}
```

### Zerolog
```json
{"level":"info","time":"2025-07-18T10:30:45Z","message":"Testando provider","provider":"zerolog"}
```

## Performance

- **Zap**: ~240k logs/segundo (recomendado para alta performance)
- **Zerolog**: ~174k logs/segundo (recomendado para baixo consumo de memória)
- **Slog**: ~132k logs/segundo (compatibilidade com stdlib)
- Metadados do serviço

## Estrutura dos Logs

```json
{
  "time": "2024-01-15T10:30:45Z",
  "level": "INFO",
  "msg": "Aplicação iniciada",
  "service": "logger-example",
  "version": "1.0.0",
  "environment": "development",
  "component": "example",
  "status": "starting",
  "port": 8080
}
```

## Casos de Uso Comuns

1. **Logging de Aplicação**: Eventos importantes do ciclo de vida
2. **Logging de Requisições**: Tracking de requests HTTP
3. **Logging de Performance**: Medição de operações
4. **Logging de Erros**: Captura e contexto de erros
5. **Logging de Debugging**: Informações detalhadas para desenvolvimento
