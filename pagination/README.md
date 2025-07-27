# Pagination Module

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/fsvxavier/nexs-lib/pagination)](https://goreportcard.com/report/github.com/fsvxavier/nexs-lib/pagination)

Um módulo robusto e extensível para paginação em aplicações Go, com suporte nativo ao Fiber e arquitetura baseada em interfaces para máxima flexibilidade.

## ✨ Características

- **🏗️ Arquitetura Modular**: Baseada em interfaces com inversão de dependências
- **🔧 Configurável**: Configuração flexível com valores padrão sensatos
- **🚀 Performance**: Otimizado para alta performance com baixa alocação de memória
- **🧪 Testado**: Cobertura de testes > 98% com testes unitários, integração e benchmarks
- **🔒 Seguro**: Validação robusta de parâmetros e proteção contra SQL injection
- **📦 Fiber Ready**: Suporte nativo ao framework Fiber
- **🔄 Extensível**: Interface plugável para diferentes providers
- **📊 Metadados Ricos**: Informações completas de navegação (anterior, próximo, total)

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/pagination
```

## 🚀 Uso Rápido

### Uso Básico com Fiber

```go
package main

import (
    "github.com/fsvxavier/nexs-lib/pagination/providers/fiber"
    "github.com/gofiber/fiber/v2"
)

func main() {
    app := fiber.New()
    
    app.Get("/users", func(c *fiber.Ctx) error {
        // Parse parâmetros de paginação
        params, err := fiber.ParseMetadata(c, "id", "name", "created_at")
        if err != nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }
        
        // Construir query SQL
        service := fiber.NewFiberPaginationService(nil)
        baseQuery := "SELECT * FROM users WHERE active = true"
        query := service.BuildQuery(baseQuery, params)
        
        // Executar query (pseudo-código)
        users, totalRecords := executeQuery(query, service.BuildCountQuery(baseQuery))
        
        // Criar resposta paginada
        response := service.CreateResponse(users, params, totalRecords)
        
        return c.JSON(response)
    })
    
    app.Listen(":3000")
}
```

### Uso com Service Personalizado

```go
package main

import (
    "net/url"
    "github.com/fsvxavier/nexs-lib/pagination"
    "github.com/fsvxavier/nexs-lib/pagination/config"
)

func main() {
    // Configuração personalizada
    cfg := &config.Config{
        DefaultLimit:     25,
        MaxLimit:         100,
        DefaultSortField: "created_at",
        DefaultSortOrder: "desc",
    }
    
    // Criar serviço
    service := pagination.NewPaginationService(cfg)
    
    // Parse parâmetros de URL
    params := url.Values{
        "page":  []string{"2"},
        "limit": []string{"50"},
        "sort":  []string{"name"},
        "order": []string{"asc"},
    }
    
    paginationParams, err := service.ParseRequest(params, "id", "name", "created_at")
    if err != nil {
        // Tratar erro de validação
        return
    }
    
    // Construir queries
    baseQuery := "SELECT * FROM products WHERE category_id = $1"
    query := service.BuildQuery(baseQuery, paginationParams)
    countQuery := service.BuildCountQuery(baseQuery)
    
    // Criar resposta
    response := service.CreateResponse(products, paginationParams, totalCount)
}
```

## 🔧 Configuração

### Configuração Padrão

```go
cfg := config.NewDefaultConfig()
// DefaultLimit: 50
// MaxLimit: 150  
// DefaultSortField: "id"
// DefaultSortOrder: "asc"
// AllowedSortOrders: ["asc", "desc", "ASC", "DESC"]
// ValidationEnabled: true
// StrictMode: false
```

### Configuração Personalizada

```go
cfg := &config.Config{
    DefaultLimit:      25,
    MaxLimit:          200,
    DefaultSortField:  "created_at",
    DefaultSortOrder:  "desc",
    AllowedSortOrders: []string{"asc", "desc"},
    ValidationEnabled: true,
    StrictMode:        true,
}
```

## 📊 Parâmetros de Query

| Parâmetro | Tipo     | Descrição                    | Exemplo     |
|-----------|----------|------------------------------|-------------|
| `page`    | int      | Número da página (≥ 1)       | `?page=2`   |
| `limit`   | int      | Registros por página (1-150) | `?limit=25` |
| `sort`    | string   | Campo para ordenação         | `?sort=name`|
| `order`   | string   | Direção (asc/desc)           | `?order=desc`|

### Exemplos de URLs

```
GET /api/users?page=1&limit=20
GET /api/users?page=2&limit=50&sort=created_at&order=desc  
GET /api/products?sort=price&order=asc
```

## 📋 Resposta JSON

```json
{
  "content": [
    {"id": 1, "name": "User 1", "email": "user1@example.com"},
    {"id": 2, "name": "User 2", "email": "user2@example.com"}
  ],
  "metadata": {
    "current_page": 2,
    "records_per_page": 20,
    "total_pages": 5,
    "total_records": 100,
    "previous": 1,
    "next": 3,
    "sort_field": "created_at",
    "sort_order": "desc"
  }
}
```

## 🏗️ Arquitetura

### Interfaces Principais

```go
// Parser de parâmetros de request
type RequestParser interface {
    ParsePaginationParams(params url.Values) (*PaginationParams, error)
}

// Validador de parâmetros
type Validator interface {
    ValidateParams(params *PaginationParams, sortableFields []string) error
}

// Construtor de queries SQL
type QueryBuilder interface {
    BuildQuery(baseQuery string, params *PaginationParams) string
    BuildCountQuery(baseQuery string) string
}

// Calculadora de metadados
type PaginationCalculator interface {
    CalculateMetadata(params *PaginationParams, totalRecords int) *PaginationMetadata
}
```

### Providers Disponíveis

- **StandardRequestParser**: Parser padrão para url.Values
- **StandardValidator**: Validador com regras configuráveis
- **StandardQueryBuilder**: Constructor de SQL queries
- **StandardPaginationCalculator**: Calculadora de metadados
- **FiberRequestParser**: Parser específico para Fiber context

## 🔒 Validação e Segurança

### Validação de Parâmetros

```go
// Campos ordenáveis permitidos
sortableFields := []string{"id", "name", "email", "created_at"}

params, err := service.ParseRequest(urlParams, sortableFields...)
if err != nil {
    // erro de validação - parâmetros inválidos
}
```

### Proteção SQL Injection

O módulo constrói queries usando parâmetros seguros:

```go
// ✅ Seguro - usa parâmetros validados
query := "SELECT * FROM users ORDER BY name desc LIMIT 10 OFFSET 20"

// ❌ Evitado - injeção direta de parâmetros não validados
```

## 🧪 Testes

### Executar Testes

```bash
# Testes unitários
go test -v -race -timeout 30s ./...

# Testes com cobertura
go test -v -race -timeout 30s -coverprofile=coverage.out ./...

# Visualizar cobertura
go tool cover -html=coverage.out

# Benchmarks
go test -bench=. -benchmem ./...
```

### Cobertura de Testes

- ✅ **Cobertura Total**: > 98%
- ✅ **Testes Unitários**: Todos os providers e funcionalidades
- ✅ **Testes de Integração**: Fiber integration
- ✅ **Benchmarks**: Performance e alocação de memória
- ✅ **Testes de Edge Cases**: Validação de limites e erros

## 📈 Performance

### Benchmarks (Go 1.21)

```
BenchmarkPaginationService_ParseRequest-8         1000000    1203 ns/op    456 B/op    8 allocs/op
BenchmarkPaginationService_BuildQuery-8           2000000     654 ns/op    128 B/op    3 allocs/op
BenchmarkPaginationService_CreateResponse-8        500000    2456 ns/op    892 B/op   12 allocs/op
```

### Otimizações

- ✅ Parsing eficiente de parâmetros
- ✅ Pool de strings para construção de queries
- ✅ Mínima alocação de memória
- ✅ Validação lazy quando desabilitada
- ✅ Cache de configuração

## 🔄 Extensibilidade

### Provider Personalizado

```go
// Implementar interfaces customizadas
type CustomRequestParser struct {
    // implementação customizada
}

func (p *CustomRequestParser) ParsePaginationParams(params url.Values) (*interfaces.PaginationParams, error) {
    // lógica personalizada
}

// Usar no serviço
service := pagination.NewPaginationServiceWithProviders(
    cfg,
    &CustomRequestParser{},
    providers.NewStandardValidator(cfg),
    providers.NewStandardQueryBuilder(),
    providers.NewStandardPaginationCalculator(),
)
```

### Hooks Personalizados

```go
// Middleware para logging
func LoggingMiddleware(next func(*interfaces.PaginationParams) error) func(*interfaces.PaginationParams) error {
    return func(params *interfaces.PaginationParams) error {
        log.Printf("Pagination request: page=%d, limit=%d", params.Page, params.Limit)
        return next(params)
    }
}
```

## 📚 Dependências

### Dependências Internas
- `github.com/fsvxavier/nexs-lib/domainerrors` - Tratamento de erros
- `github.com/fsvxavier/nexs-lib/validation/jsonschema` - Validação JSON Schema

### Dependências Externas
- `github.com/gofiber/fiber/v2` - Framework web (opcional, apenas para provider Fiber)
- `github.com/stretchr/testify` - Testes (dev dependency)

## 🔄 Migração da Versão Antiga

### Antes (v1)

```go
// Código antigo
metadata, err := pagination.ParseMetadata(ctxFiber, "id", "name")
output := pagination.NewPaginatedOutput(data, metadata)
```

### Depois (v2)

```go
// Código novo
params, err := fiber.ParseMetadata(c, "id", "name")
service := fiber.NewFiberPaginationService(nil)
response := service.CreateResponse(data, params, totalRecords)
```

### Compatibilidade

O módulo mantém funções de compatibilidade para facilitar a migração:

```go
// Funções legacy ainda funcionam
metadata, err := fiber.ParseMetadata(c, "id", "name")
output := fiber.NewPaginatedOutput(data, metadata)
```

## 📋 TODO / Roadmap

- [ ] Provider para GraphQL
- [ ] Provider para gRPC  
- [ ] Suporte a cursor-based pagination
- [ ] Cache de metadados
- [ ] Métricas de performance
- [ ] Provider para MongoDB
- [ ] Suporte a filtros dinâmicos
- [ ] Templates de query personalizáveis

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### Guidelines

- ✅ Manter cobertura de testes > 98%
- ✅ Seguir convenções Go (gofmt, golint, go vet)
- ✅ Documentar funções públicas
- ✅ Adicionar benchmarks para novas funcionalidades
- ✅ Testar edge cases

## 📄 Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 👥 Autores

- **Nexs Team** - *Desenvolvimento inicial* - [fsvxavier](https://github.com/fsvxavier)

## 🙏 Agradecimentos

- Framework Fiber pela inspiração na API
- Comunidade Go pelas melhores práticas
- Contribuidores e reviewers
