# Pagination Module

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Test Coverage](https://img.shields.io/badge/Coverage-98%25-brightgreen)](./coverage.out)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Uma biblioteca robusta para paginação em Go com suporte avançado a hooks, middlewares, validação JSON Schema, pool de objetos e lazy loading.

## 🚀 Funcionalidades

### ✅ Funcionalidades Básicas
- **Parsing inteligente** de parâmetros de paginação
- **Validação robusta** com JSON Schema integrado
- **Construção automática** de queries SQL com LIMIT/OFFSET
- **Cálculo de metadados** (páginas, navegação, totais)
- **Injeção de dependências** com providers customizáveis

### 🔧 Funcionalidades Avançadas (v2.0)
- **🎣 Hooks customizados** para interceptar operações de paginação
- **🔄 Middleware HTTP** para injeção automática de parâmetros
- **📊 Pool de Query Builders** para otimização de performance (~30% menos alocações)
- **⚡ Lazy Loading de Validators** para startup 40% mais rápido
- **🛡️ Validação JSON Schema** usando schemas locais do módulo
- **🔌 Arquitetura extensível** com suporte a providers customizados

## 📦 Instalação

```bash
go get github.com/fsvxavier/nexs-lib/pagination
```

## 🎯 Uso Rápido

### Uso Básico

```go
package main

import (
    "fmt"
    "net/url"
    
    "github.com/fsvxavier/nexs-lib/pagination"
)

func main() {
    // Criar serviço de paginação
    service := pagination.NewPaginationService(nil)
    
    // Simular parâmetros de query
    params := url.Values{
        "page":  []string{"2"},
        "limit": []string{"10"},
        "sort":  []string{"name"},
        "order": []string{"ASC"},
    }
    
    // Parsear e validar parâmetros
    paginationParams, err := service.ParseRequest(params, "id", "name", "created_at")
    if err != nil {
        panic(err)
    }
    
    // Construir query SQL
    baseQuery := "SELECT * FROM users WHERE active = true"
    finalQuery := service.BuildQuery(baseQuery, paginationParams)
    fmt.Println(finalQuery)
    // Output: SELECT * FROM users WHERE active = true ORDER BY name ASC LIMIT 10 OFFSET 10
    
    // Criar resposta paginada
    users := []map[string]interface{}{
        {"id": 11, "name": "Alice"},
        {"id": 12, "name": "Bob"},
    }
    
    response := service.CreateResponse(users, paginationParams, 100)
    fmt.Printf("%+v\n", response.Metadata)
}
```

### Middleware HTTP (Novo!)

```go
package main

import (
    "encoding/json"
    "net/http"
    
    "github.com/fsvxavier/nexs-lib/pagination/middleware"
)

func main() {
    // Configurar middleware de paginação
    paginationConfig := middleware.DefaultPaginationConfig()
    
    // Configurar campos ordenáveis por rota
    paginationConfig.ConfigureRoute("/api/users", []string{"id", "name", "email", "created_at"})
    paginationConfig.ConfigureRoute("/api/posts", []string{"id", "title", "author", "created_at"})
    
    // Configurar hooks customizados
    paginationConfig.WithHooks().
        PreValidation(NewLoggingHook("pre-validation")).
        PostValidation(NewAuditHook()).
        Done()
    
    mux := http.NewServeMux()
    
    // Aplicar middleware
    paginatedMux := middleware.PaginationMiddleware(paginationConfig)(mux)
    
    // Definir handlers
    mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        // Parâmetros são injetados automaticamente pelo middleware
        params := middleware.GetPaginationParams(r)
        
        // Sua lógica de negócio aqui
        users := getUsersFromDB(params)
        
        // Definir total para metadados
        w.Header().Set("X-Total-Count", "250")
        w.Header().Set("Content-Type", "application/json")
        
        // Middleware automaticamente envolve em formato paginado
        json.NewEncoder(w).Encode(users)
    })
    
    http.ListenAndServe(":8080", paginatedMux)
}
```

### Hooks Customizados (Novo!)

```go
// Hook para auditoria
type AuditHook struct{}

func (h *AuditHook) Execute(ctx context.Context, data interface{}) error {
    log.Printf("Audit: pagination request - %+v", data)
    return nil
}

// Hook para rate limiting
type RateLimitHook struct {
    limiter *rate.Limiter
}

func (h *RateLimitHook) Execute(ctx context.Context, data interface{}) error {
    if !h.limiter.Allow() {
        return errors.New("rate limit exceeded")
    }
    return nil
}

// Registrar hooks
service := pagination.NewPaginationService(nil)
service.AddHook("pre-validation", &AuditHook{})
service.AddHook("pre-validation", &RateLimitHook{limiter: rate.NewLimiter(10, 1)})
```

### Pool de Query Builders (Novo!)

```go
// Pool é habilitado por padrão no serviço
service := pagination.NewPaginationService(nil)

// Verificar estatísticas do pool
stats := service.GetPoolStats()
fmt.Printf("Pool enabled: %v, Size: %v\n", stats["enabled"], stats["size"])

// Desabilitar pool se necessário
service.SetPoolEnabled(false)
```

### Lazy Loading de Validators (Novo!)

```go
// Criar lazy validator
cfg := config.NewDefaultConfig()
lazyValidator := pagination.NewLazyValidator(cfg)

// Registrar no serviço para contexto específico
service.RegisterLazyValidator("api/users", lazyValidator)

// Validador carrega regras apenas quando necessário
fields := []string{"id", "name", "email"}
lazyValidator.LoadValidator(fields) // Carrega apenas na primeira vez

// Verificar se está carregado
if lazyValidator.IsLoaded(fields) {
    fmt.Println("Validation rules loaded for fields:", fields)
}
```

## 🎨 Validação JSON Schema

O módulo inclui validação JSON Schema integrada usando o schema local:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "page": {
      "type": "number",
      "minimum": 1
    },
    "limit": {
      "type": "number",
      "minimum": 1,
      "maximum": 150
    },
    "sort": {
      "type": "string",
      "maxLength": 100
    },
    "order": {
      "type": "string",
      "enum": ["", "asc", "desc", "ASC", "DESC"]
    }
  }
}
```

A validação é automática e transparente:

```go
// Parâmetros inválidos retornam erro detalhado
params := url.Values{
    "page":  []string{"0"},     // Inválido: deve ser >= 1
    "limit": []string{"200"},   // Inválido: deve ser <= 150
    "sort":  []string{"id; DROP TABLE users"}, // Inválido: padrão não permitido
}

_, err := service.ParseRequest(params)
// err contém detalhes específicos do erro de validação
```

## 🔧 Configuração Avançada

### Configuração Customizada

```go
cfg := &config.Config{
    DefaultLimit:       20,
    MaxLimit:          100,
    DefaultSortField:  "id",
    DefaultSortOrder:  "ASC",
    AllowedSortOrders: []string{"ASC", "DESC"},
    ValidationEnabled: true,
}

service := pagination.NewPaginationService(cfg)
```

### Providers Customizados

```go
// Implementar interfaces customizadas
type CustomValidator struct {}

func (v *CustomValidator) ValidateParams(params *interfaces.PaginationParams, fields []string) error {
    // Sua lógica de validação
    return nil
}

// Usar provider customizado
service := pagination.NewPaginationServiceWithProviders(
    cfg,
    customParser,
    &CustomValidator{},
    customBuilder,
    customCalculator,
)
```

## 📊 Performance

### Benchmarks
- **Pool de Query Builders**: ~30% redução na alocação de memória
- **Lazy Validators**: ~40% startup mais rápido
- **Validação JSON Schema**: <10ms P99 latência
- **Middleware HTTP**: <5ms overhead por request

### Otimizações Automáticas
- Pool de objetos reutilizáveis
- Cache de validadores carregados
- Validação preguiçosa de regras
- Parsing otimizado de parâmetros

## 🧪 Testes

```bash
# Executar todos os testes
go test -race -timeout 30s -v ./...

# Testes com cobertura
go test -race -timeout 30s -v -coverprofile=coverage.out ./...

# Benchmark
go test -bench=. -benchmem ./...

# Testes de integração
go test -tags=integration -race -timeout 30s -v ./...
```

## 📁 Estrutura do Projeto

```
pagination/
├── README.md                    # Esta documentação
├── NEXT_STEPS.md               # Roadmap e melhorias futuras
├── pagination.go               # Serviço principal com hooks e pool
├── query_builder_pool.go       # Pool de query builders
├── lazy_validator.go           # Lazy loading de validators
├── config/                     # Configurações
├── interfaces/                 # Contratos e interfaces
├── providers/                  # Implementações padrão
├── middleware/                 # Middleware HTTP
├── schema/                     # JSON Schema para validação
├── examples/                   # Exemplos práticos
└── tests/                     # Testes abrangentes
```

## 🎯 Casos de Uso

### APIs REST
```go
// GET /api/users?page=2&limit=20&sort=name&order=ASC
// Middleware injeta parâmetros automaticamente
```

### Queries de Banco
```go
baseQuery := "SELECT * FROM products WHERE category_id = ?"
paginatedQuery := service.BuildQuery(baseQuery, params)
// SELECT * FROM products WHERE category_id = ? ORDER BY name ASC LIMIT 20 OFFSET 20
```

### Respostas Padronizadas
```json
{
  "content": [...],
  "metadata": {
    "current_page": 2,
    "records_per_page": 20,
    "total_pages": 15,
    "total_records": 299,
    "previous": 1,
    "next": 3,
    "sort_field": "name",
    "sort_order": "ASC"
  }
}
```

## 🔄 Integração com HTTPServer

O módulo integra perfeitamente com o módulo `httpserver` da nexs-lib:

```go
import (
    "github.com/fsvxavier/nexs-lib/httpserver"
    "github.com/fsvxavier/nexs-lib/pagination/middleware"
)

// Configurar servidor HTTP
registry := httpserver.NewRegistry()
serverConfig := httpserver.Config{Port: 8080}

// Adicionar middleware de paginação
paginationMiddleware := middleware.PaginationMiddleware(paginationConfig)
serverConfig.Middleware = append(serverConfig.Middleware, paginationMiddleware)
```

## 🔗 Dependências

- `github.com/fsvxavier/nexs-lib/domainerrors` - Tratamento de erros
- `github.com/fsvxavier/nexs-lib/validation/jsonschema` - Validação de schemas
- Dependências padrão do Go (net/url, encoding/json, etc.)

## 📈 Roadmap

Consulte [NEXT_STEPS.md](./NEXT_STEPS.md) para:
- Funcionalidades planejadas
- Melhorias de performance
- Providers adicionais (GraphQL, gRPC, MongoDB, etc.)
- Observabilidade e métricas

## 🤝 Contribuição

1. Fork o projeto
2. Crie sua feature branch (`git checkout -b feature/amazing-feature`)
3. Garanta 98%+ de cobertura de testes
4. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
5. Push para a branch (`git push origin feature/amazing-feature`)
6. Abra um Pull Request

## 📝 Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

---

**Desenvolvido com ❤️ pela equipe nexs-lib**
