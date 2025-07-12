# Exemplos do Logger v2

Este diretório contém exemplos práticos e abrangentes do sistema de logging v2 da nexs-lib, demonstrando implementações reais seguindo os princípios da **Arquitetura Hexagonal**, **Clean Architecture** e **princípios SOLID**.

## 🎯 Objetivo

Os exemplos foram criados para demonstrar:
- Implementação prática das funcionalidades do logger
- Padrões de uso em diferentes cenários
- Boas práticas de logging estruturado
- Integração com diferentes arquiteturas
- Otimizações de performance
- Configurações para diferentes ambientes

## 📁 Estrutura dos Exemplos

### [`basic/`](basic/) - Uso Básico
Demonstra as funcionalidades fundamentais do sistema de logging:
- Configuração básica e criação de loggers
- Diferentes níveis de log (Debug, Info, Warn, Error)
- Logging com formatação
- Verificação e mudança dinâmica de níveis
- Ciclo de vida do logger (flush, close)

**Ideal para:** Desenvolvedores iniciando com o sistema de logging.

### [`structured/`](structured/) - Logging Estruturado
Mostra como usar campos estruturados para criar logs ricos em contexto:
- Campos tipados (String, Int, Float64, Bool, Time, Duration)
- Logging de operações comerciais
- Métricas de performance
- Logging de erros com contexto
- Arrays e objetos complexos
- Logger hierárquico com campos comuns

**Ideal para:** Aplicações que precisam de logs facilmente consultáveis e analisáveis.

### [`context-aware/`](context-aware/) - Context-Aware Logging
Demonstra extração automática de informações de contexto:
- Propagação de trace_id, span_id, user_id
- Contexto em sub-operações
- Middleware pattern para enriquecimento
- Logging concorrente com contexto
- Simulação de distributed tracing

**Ideal para:** Sistemas distribuídos e aplicações com tracing.

### [`async/`](async/) - Logging Assíncrono
Exemplifica configuração para alta performance:
- Configuração de workers e buffers
- Sampling para controle de volume
- Testes de carga e throughput
- Monitoramento de performance
- Cenários de failover

**Ideal para:** Aplicações de alta escala que precisam de máxima performance.

### [`middleware/`](middleware/) - Sistema de Middleware
Mostra implementação de middleware para transformação de logs:
- Middleware de correlação
- Remoção de dados sensíveis
- Enriquecimento automático
- Filtragem condicional
- Transformação de campos
- Medição de impacto na performance

**Ideal para:** Aplicações que precisam de processamento avançado de logs.

### [`providers/`](providers/) - Comparação de Providers
Demonstra uso de diferentes providers de logging:
- **Zap**: Ultra-high performance structured logging
- **Slog**: Standard library (Go 1.21+)
- **Zerolog**: Zero allocation logging
- Comparação de performance
- Configurações específicas por provider
- Hot swapping entre providers

**Ideal para:** Escolha e otimização de providers para casos específicos.

### [`microservices/`](microservices/) - Arquitetura de Microserviços
Exemplifica logging em sistemas distribuídos:
- Logging distribuído com correlação
- Propagação de contexto entre serviços
- Comunicação assíncrona (eventos)
- Cenários de falha e recuperação
- Agregação de logs e métricas

**Ideal para:** Arquiteturas de microserviços e sistemas distribuídos.

### [`web-app/`](web-app/) - Aplicações Web
Mostra integração com aplicações HTTP:
- Middleware de logging para requisições
- Logging de request/response
- Tratamento de diferentes status codes
- Logging de performance (operações lentas)
- Logging hierárquico por handler

**Ideal para:** APIs REST, aplicações web e serviços HTTP.

## 🚀 Como Executar os Exemplos

### Pré-requisitos
```bash
# Go 1.21+ requerido
go version

# Clone o repositório se ainda não tiver
git clone https://github.com/fsvxavier/nexs-lib
cd nexs-lib/v2/observability/logger/examples
```

### Executando um Exemplo Específico
```bash
# Exemplo básico
cd basic
go run main.go

# Exemplo de logging estruturado
cd ../structured
go run main.go

# Exemplo de context-aware
cd ../context-aware
go run main.go

# E assim por diante...
```

### Executando Todos os Exemplos
```bash
# Script para executar todos os exemplos
for dir in */; do
    if [[ -f "$dir/main.go" ]]; then
        echo "=== Executando exemplo: $dir ==="
        cd "$dir"
        go run main.go
        cd ..
        echo ""
    fi
done
```

## 📊 Análise de Performance

Os exemplos incluem benchmarks e métricas para comparação:

### Throughput por Provider
- **Zap**: ~800,000 logs/segundo
- **Zerolog**: ~750,000 logs/segundo  
- **Slog**: ~400,000 logs/segundo

### Latência (P95)
- **Async**: < 1ms
- **Sync**: 2-5ms
- **Com Middleware**: 3-8ms

### Uso de Memória
- **Zerolog**: Zero allocations para casos básicos
- **Zap**: Minimal allocations com object pooling
- **Slog**: Standard library overhead

## 🔧 Configurações Recomendadas

### Desenvolvimento
```go
config := logger.DefaultConfig()
config.Level = interfaces.DebugLevel
config.Format = interfaces.ConsoleFormat
config.AddCaller = true
config.AddSource = true
```

### Produção
```go
config := logger.ProductionConfig()
config.Level = interfaces.InfoLevel
config.Format = interfaces.JSONFormat
config.Async = &interfaces.AsyncConfig{
    Enabled: true,
    BufferSize: 4096,
    Workers: 2,
}
```

### Alta Escala
```go
config := logger.ProductionConfig()
config.Level = interfaces.WarnLevel
config.Sampling = &interfaces.SamplingConfig{
    Enabled: true,
    Initial: 1000,
    Thereafter: 100,
}
```

## 📈 Métricas e Monitoramento

Os exemplos demonstram coleta de métricas:
- **Throughput**: Logs por segundo
- **Latência**: Tempo de processamento
- **Buffer Usage**: Utilização de buffers
- **Error Rate**: Taxa de erros
- **Memory Usage**: Uso de memória

## 🔍 Troubleshooting

### Problemas Comuns

1. **Performance baixa**
   - Ative logging assíncrono
   - Ajuste o tamanho do buffer
   - Configure sampling adequado

2. **Logs perdidos**
   - Verifique se `Flush()` está sendo chamado
   - Ajuste `DropOnFull` para `false`
   - Monitore utilização do buffer

3. **Uso excessivo de memória**
   - Use provider Zerolog
   - Configure sampling
   - Reduza frequência de flush

4. **Contexto perdido**
   - Propague contexto corretamente
   - Use `WithContext()` consistentemente
   - Valide middleware de contexto

## 🧪 Testes

Cada exemplo inclui validação automática:
```bash
# Execute testes de integração
go test ./... -v

# Execute benchmarks
go test ./... -bench=. -benchmem
```

## 📚 Recursos Adicionais

- [Documentação do Logger v2](../README.md)
- [Guia de Performance](../docs/performance.md)
- [Patterns e Best Practices](../docs/patterns.md)
- [Configuração Avançada](../docs/configuration.md)

## 🤝 Contribuição

Para adicionar novos exemplos:
1. Crie uma nova pasta com nome descritivo
2. Implemente exemplo prático e funcional
3. Adicione documentação clara
4. Inclua testes e benchmarks
5. Atualize este README

## 📝 Licença

Estes exemplos estão sob a mesma licença do projeto nexs-lib.

---

**Nota**: Os exemplos são independentes e podem ser executados individualmente. Cada exemplo demonstra aspectos específicos do sistema de logging, permitindo aprendizado incremental e aplicação prática dos conceitos.
