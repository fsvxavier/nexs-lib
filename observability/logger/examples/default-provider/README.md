# Exemplo Default Provider

Este exemplo demonstra o uso do provider padrão do sistema de logging (Zap).

## Executando o Exemplo

```bash
cd examples/default-provider
go run main.go
```

## Funcionalidades Demonstradas

### 1. Provider Padrão (Zap)
- **Configuração automática**: Zap é configurado automaticamente quando registrado
- **Sem configuração explícita**: Funciona imediatamente após importar os providers
- **Formato JSON**: Saída estruturada por padrão
- **Performance otimizada**: ~240k logs/segundo

### 2. Auto-Registration
- **Import side-effect**: Providers são registrados automaticamente
- **Zap como padrão**: Primeiro provider zap registrado torna-se padrão
- **Verificação de provider**: Função para verificar qual provider está ativo

### 3. Uso Simples
- **Interface única**: Usa a mesma interface para todos os providers
- **Logging direto**: Não requer configuração prévia
- **Campos estruturados**: Suporte completo a campos tipados

### 4. Verificação de Estado
- **Lista de providers**: Mostra todos os providers disponíveis
- **Provider atual**: Identifica qual provider está ativo
- **Confirmação**: Verifica se zap é realmente o padrão

## Código de Exemplo

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/fsvxavier/nexs-lib/observability/logger"
    
    // Auto-registration dos providers
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zap"
    _ "github.com/fsvxavier/nexs-lib/observability/logger/providers/zerolog"
)

func main() {
    ctx := context.Background()
    
    // Verifica qual provider está sendo usado
    currentProvider := logger.GetCurrentProviderName()
    fmt.Printf("Provider atual (padrão): %s\n", currentProvider)
    
    // Usa o logger sem configurar explicitamente
    logger.Info(ctx, "Mensagem de teste com provider padrão")
}
```

## Saída Esperada

### Informações do Sistema
```
Providers disponíveis: [slog zap zerolog]
Provider atual (padrão): zap
```

### Logs do Zap
```json
{"level":"info","time":"2025-07-18T10:30:45Z","msg":"Mensagem de teste com provider padrão"}
{"level":"info","time":"2025-07-18T10:30:45Z","msg":"Teste com campos estruturados","provider":"zap","zap_default":true,"test_number":42}
{"level":"warn","time":"2025-07-18T10:30:45Z","msg":"Mensagem de warning"}
{"level":"error","time":"2025-07-18T10:30:45Z","msg":"Mensagem de erro"}
```

## Por que Zap como Padrão?

### 1. Performance Superior
- **~240k logs/segundo**: Mais rápido que slog (~132k) e zerolog (~174k)
- **Baixa latência**: Otimizado para aplicações de alta performance
- **Eficiência de memória**: Alocações mínimas durante logging

### 2. Adoção na Comunidade
- **Amplamente usado**: Padrão em muitas aplicações Go
- **Maturidade**: Biblioteca estável e bem testada
- **Ecossistema**: Boa integração com outras ferramentas

### 3. Flexibilidade
- **Configuração rica**: Muitas opções de configuração
- **Múltiplos formatos**: JSON, Console, Custom
- **Sampling**: Controle de volume de logs

### 4. Compatibilidade
- **Interface slog**: Mantém compatibilidade com a interface padrão
- **Migração fácil**: Transição suave de outros loggers
- **Estrutura consistente**: Campos estruturados padronizados

## Vantagens do Provider Padrão

### ✅ Simplicidade
- Não requer configuração explícita
- Funciona imediatamente após importação
- Interface consistente

### ✅ Performance
- Otimizado para produção
- Baixo overhead
- Alta taxa de throughput

### ✅ Confiabilidade
- Biblioteca madura e testada
- Usado em aplicações críticas
- Suporte ativo da comunidade

### ✅ Flexibilidade
- Pode ser trocado por outros providers
- Configuração opcional disponível
- Suporte a diferentes formatos

## Quando Usar

### ✅ Recomendado para:
- Aplicações novas
- Sistemas de alta performance
- Produção com alta carga
- Quando não há preferência específica

### ⚠️ Considere alternativas quando:
- Restrições extremas de memória (use zerolog)
- Dependência apenas da stdlib (use slog)
- Integrações específicas de outros providers

## Próximos Passos

1. **Configuração personalizada**: Veja `examples/multi-provider/` para configurações avançadas
2. **Comparação de providers**: Execute `examples/benchmark/` para comparar performance
3. **Uso em serviços**: Veja `examples/advanced/` para integração com serviços
4. **Logging básico**: Veja `examples/basic/` para funcionalidades básicas
