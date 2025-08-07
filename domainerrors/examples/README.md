# Exemplos - Domain Errors

Esta pasta contém exemplos práticos demonstrando diferentes aspectos e casos de uso do sistema de domain errors.

## Estrutura dos Exemplos

### 📁 basic/
Exemplo básico demonstrando funcionalidades fundamentais:
- Criação de diferentes tipos de erro
- Uso de metadados
- Serialização JSON
- Mapeamento para códigos HTTP
- Stack traces
- Encadeamento de erros (wrapping)

**Funcionalidades**: Introdução aos conceitos básicos
**Arquivo**: `basic/main.go`
**README**: `basic/README.md`

### 📁 global/
Exemplo de hooks e middlewares globais:
- Registro de hooks globais (start, stop, error, i18n)
- Middlewares globais para processamento
- Sistema de tradução automática
- Estatísticas de hooks e middlewares
- Processamento em cadeia

**Funcionalidades**: Hooks, middlewares, i18n global
**Arquivo**: `global/main.go`
**README**: `global/README.md`

### 📁 advanced/
Exemplo avançado com padrões empresariais:
- Sistema de métricas thread-safe
- Audit trail para compliance
- Circuit breaker pattern
- Classificação de erros por criticidade
- Context enrichment
- Rate limiting
- Observabilidade completa

**Funcionalidades**: Métricas, audit, circuit breaker, observability
**Arquivo**: `advanced/main.go`
**README**: `advanced/README.md`

### 📁 outros/
Casos de uso práticos e integrações:
- Validação de formulários complexos
- Sistema bancário com transações
- API REST com tratamento de erros
- Autenticação multi-modal
- Integração com serviços externos
- Sistema de cache com fallback

**Funcionalidades**: Validação, transações, REST API, auth, cache
**Arquivo**: `outros/main.go`
**README**: `outros/README.md`

## Como Executar os Exemplos

Cada exemplo pode ser executado individualmente:

```bash
# Exemplo básico
cd basic && go run main.go

# Exemplo global
cd global && go run main.go

# Exemplo avançado
cd advanced && go run main.go

# Outros casos de uso
cd outros && go run main.go
```

Ou compile primeiro:

```bash
cd [exemplo]
go build -o example main.go
./example
```

## Progressão Recomendada

1. **basic/**: Comece aqui para entender os conceitos fundamentais
2. **global/**: Aprenda sobre hooks e middlewares globais
3. **advanced/**: Explore padrões empresariais avançados
4. **outros/**: Veja casos de uso práticos e integrações

## Dependências

Todos os exemplos dependem apenas do módulo principal:
```go
import "github.com/fsvxavier/nexs-lib/domainerrors"
```

Alguns exemplos específicos também importam subpacotes:
```go
import (
    "github.com/fsvxavier/nexs-lib/domainerrors/hooks"
    "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
    "github.com/fsvxavier/nexs-lib/domainerrors/middlewares"
)
```

## Funcionalidades por Exemplo

| Funcionalidade | basic | global | advanced | outros |
|---------------|-------|--------|----------|--------|
| Criação de erros | ✅ | ✅ | ✅ | ✅ |
| Metadados | ✅ | ✅ | ✅ | ✅ |
| Stack traces | ✅ | ✅ | ✅ | ✅ |
| JSON serialization | ✅ | ✅ | ✅ | ✅ |
| HTTP mapping | ✅ | ✅ | ✅ | ✅ |
| Hooks globais | ❌ | ✅ | ✅ | ❌ |
| Middlewares globais | ❌ | ✅ | ✅ | ❌ |
| I18n | ❌ | ✅ | ✅ | ❌ |
| Métricas | ❌ | ❌ | ✅ | ❌ |
| Audit trail | ❌ | ❌ | ✅ | ❌ |
| Circuit breaker | ❌ | ❌ | ✅ | ❌ |
| Validação complexa | ❌ | ❌ | ❌ | ✅ |
| Sistema bancário | ❌ | ❌ | ❌ | ✅ |
| API REST | ❌ | ❌ | ❌ | ✅ |
| Autenticação | ❌ | ❌ | ❌ | ✅ |
| Cache + Fallback | ❌ | ❌ | ❌ | ✅ |

## Padrões Demonstrados

### Design Patterns
- **Observer Pattern**: Hooks para notificações
- **Chain of Responsibility**: Middlewares em cadeia
- **Strategy Pattern**: Diferentes tipos de erro
- **Decorator Pattern**: Enriquecimento de contexto
- **Circuit Breaker**: Resiliência de sistema

### Architectural Patterns
- **Error Handling**: Tratamento consistente
- **Context Propagation**: Contexto através da aplicação
- **Audit Trail**: Rastreabilidade completa
- **Observability**: Métricas e monitoring
- **Fallback Strategy**: Recuperação de falhas

### Enterprise Patterns
- **Domain Driven Design**: Erros como parte do domínio
- **Clean Architecture**: Separação de responsabilidades
- **CQRS**: Diferentes tratamentos para command/query
- **Event Sourcing**: Auditoria através de eventos

## Casos de Uso Cobertos

### Web Applications
- Validação de formulários
- APIs REST
- Autenticação/autorização
- Tratamento de erros HTTP

### Microservices
- Integração entre serviços
- Circuit breakers
- Rate limiting
- Distributed tracing

### Financial Systems
- Transações bancárias
- Validação de negócio
- Audit compliance
- Risk management

### Enterprise Systems
- Sistema de métricas
- Audit trail
- Context enrichment
- Multi-tenancy

## Métricas de Cobertura

Cada exemplo cobre diferentes aspectos:
- **basic/**: ~30% das funcionalidades (fundamentos)
- **global/**: ~50% das funcionalidades (hooks + middlewares)
- **advanced/**: ~80% das funcionalidades (empresarial)
- **outros/**: ~60% das funcionalidades (casos práticos)

## Próximos Passos

Após estudar os exemplos, você pode:
1. Implementar o sistema em seu projeto
2. Customizar os tipos de erro para seu domínio
3. Criar hooks específicos para suas necessidades
4. Implementar middlewares customizados
5. Integrar com suas ferramentas de observabilidade

## Contribuição

Para adicionar novos exemplos:
1. Crie uma nova pasta com nome descritivo
2. Implemente o exemplo em `main.go`
3. Crie um `README.md` detalhado
4. Atualize este README principal
5. Execute os testes para garantir funcionamento
