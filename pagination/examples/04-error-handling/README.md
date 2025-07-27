# Exemplo 4: Tratamento Avançado de Erros

Este exemplo demonstra como implementar tratamento robusto de erros no módulo de paginação, incluindo:

- Detecção e classificação automática de erros
- Formatação de erros para APIs
- Recuperação graceful com fallbacks
- Mensagens amigáveis para usuários
- Sugestões automáticas de correção
- Encadeamento de operações com tratamento de falhas

## Como executar

```bash
cd examples/04-error-handling
go run main.go
```

## O que o exemplo demonstra

### 1. ❌ Cenários de Erro Comuns

#### Parâmetros de Página Inválidos
- Valores não numéricos (`page=abc`)
- Números negativos (`page=-1`)  
- Zero (`page=0`)

#### Parâmetros de Limite Inválidos
- Valores não numéricos (`limit=xyz`)
- Números negativos (`limit=-10`)
- Zero (`limit=0`)
- Excesso do máximo (`limit=500`)

#### Ordenação Inválida
- Campos não permitidos (`sort=password`)
- Ordens inválidas (`order=random`)

### 2. 🔄 Recuperação Graceful

```go
// Parâmetros vazios → usa valores padrão
params := url.Values{}
result, err := service.ParseRequest(params)
// ✅ result.Page = 1, result.Limit = 50, result.SortField = "id"

// Parâmetros parciais → preenche faltantes
params := url.Values{"page": []string{"2"}}
result, err := service.ParseRequest(params)
// ✅ Usa page=2, mas limit e sort são padrão
```

### 3. 📋 Formatação de Erros para API

#### Estrutura de Resposta
```json
{
  "error": "Validation Failed",
  "code": "INVALID_PAGE_PARAMETER", 
  "message": "page must be a positive integer",
  "details": {
    "type": "ValidationError",
    "field": "page",
    "suggestion": "Use um número inteiro positivo para a página (ex: page=1)"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-12345"
}
```

#### Códigos de Erro Padronizados
- `INVALID_PAGE_PARAMETER` - Página inválida
- `INVALID_LIMIT_PARAMETER` - Limite inválido  
- `LIMIT_TOO_LARGE` - Limite excede máximo
- `INVALID_SORT_FIELD` - Campo de ordenação inválido
- `INVALID_SORT_ORDER` - Ordem de classificação inválida

### 4. 👤 Mensagens Amigáveis

#### Para Usuários Finais
```
❌ "A primeira página é a número 1. Tente novamente com page=1."
❌ "Muitos resultados solicitados. O máximo permitido é 100 registros por página."
❌ "Não é possível ordenar por este campo. Campos disponíveis: id, name, email."
```

#### Para Desenvolvedores
```
🔧 "Page parameter must be >= 1"
🔧 "Limit exceeds maximum allowed value" 
🔧 "Sort field not in allowed list"
```

### 5. 💡 Sugestões Automáticas

O sistema analiza erros e sugere correções:

```go
// Entrada problemática: page=-5&limit=abc&sort=invalid&order=random
// Saída sugerida: page=1&limit=10&sort=id&order=asc
```

### 6. 🔗 Encadeamento com Fallbacks

```go
// 1. Tenta parse dos parâmetros
params, err := service.ParseRequest(userParams)
if err != nil {
    // 2. Usa parâmetros padrão como fallback
    fallbackParams, _ := service.ParseRequest(defaultParams)
    // 3. Continua operação com fallback
    query := service.BuildQuery("SELECT * FROM users", fallbackParams)
}
```

## Saída do Exemplo

```
❌ Exemplos de Tratamento de Erros - Módulo de Paginação
========================================================

=== 1. Demonstração de Cenários de Erro ===

🧪 Teste 1: Página Inválida - Texto
📝 Descrição: Parâmetro page com valor não numérico
🔗 Parâmetros: page=abc
❌ Erro capturado: [INVALID_PAGE_PARAMETER] page must be a positive integer
📋 Resposta da API:
   {
     "error": "Validation Failed",
     "code": "INVALID_PAGE_PARAMETER",
     "message": "page must be a positive integer",
     "details": {
       "type": "ValidationError", 
       "field": "page",
       "suggestion": "Use um número inteiro positivo para a página (ex: page=1)"
     },
     "timestamp": "2024-01-01T12:00:00Z",
     "request_id": "req-12345"
   }
✅ Erro corresponde ao esperado: INVALID_PAGE_PARAMETER

=== 2. Demonstração de Recuperação de Erros ===

🔄 Teste de Recuperação 1: Parâmetros Vazios
📝 Deve usar valores padrão quando nenhum parâmetro é fornecido
🔗 Parâmetros: 
✅ Recuperação bem-sucedida:
   Página: 1
   Limite: 50
   Campo de ordenação: id
   Ordem: asc
```

## Conceitos Demonstrados

- ✅ **Domain Errors** - Uso da biblioteca `domainerrors` do projeto
- ✅ **Códigos Estruturados** - Erros com códigos máquina-legíveis
- ✅ **Mensagens Contextuais** - Diferentes níveis de detalhamento
- ✅ **Recuperação Automática** - Fallbacks inteligentes
- ✅ **Sugestões Dinâmicas** - Análise de erros com correções
- ✅ **Logging Estruturado** - Request IDs e timestamps
- ✅ **Operações Encadeadas** - Continuar processamento após falhas

## Integração com Domain Errors

O exemplo utiliza a biblioteca `domainerrors` do projeto:

```go
import "github.com/fsvxavier/nexs-lib/domainerrors"

// O módulo de paginação retorna DomainError
if domainErr, ok := err.(*domainerrors.DomainError); ok {
    // Acesso a código, tipo, mensagem estruturados
    code := domainErr.Code
    message := domainErr.Message
    errorType := domainErr.Type
}
```

## Casos de Uso Práticos

### 1. 🌐 API REST
- Retornar erros HTTP estruturados
- Logs detalhados para debugging
- Códigos de erro para clientes

### 2. 📱 Aplicação Mobile
- Mensagens simplificadas para usuários
- Retry automático com parâmetros corrigidos
- Offline handling

### 3. 🔧 Sistema Interno
- Logs técnicos detalhados
- Alertas para erros recorrentes
- Métricas de qualidade

### 4. 🎯 Interface de Admin
- Validação em tempo real
- Sugestões de correção
- Histórico de erros

## Próximos Passos

Após entender tratamento de erros, veja:
- `05-database-integration` - Integração com PostgreSQL
- `06-performance-optimization` - Otimizações de performance
- `07-middleware-advanced` - Middleware avançado
