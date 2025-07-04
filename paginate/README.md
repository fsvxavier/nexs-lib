# Biblioteca de Paginação (Page)

Esta biblioteca fornece uma solução abrangente para paginação, ordenação e manipulação de consultas para APIs REST e interfaces de banco de dados. Ela segue os padrões das bibliotecas `edomain`, `httpservers` e `db` para fornecer uma experiência consistente e robusta.

## 🚀 Características

- **Interface unificada** para diferentes frameworks HTTP (Fiber, Echo, Gin, net/http, Atreugo, FastHTTP)
- **Adaptadores específicos** com parsing automático de parâmetros HTTP
- **Integração com banco de dados** para consultas paginadas eficientes
- **Paginação de slices em memória** com funções centralizadas para camada de serviço
- **Cálculo automático de índices** sem necessidade de implementação manual
- **Validação robusta de parâmetros** com JSON Schema
- **Tratamento adequado de erros** usando edomain
- **Configuração flexível** com padrão de options
- **Suporte a tipos genéricos** através de conversão automática de slices

## Estrutura

```
page/
  ├── page.go         # Estruturas e funções principais
  ├── schema.go       # Esquema JSON para validação
  ├── parse.go        # Análise de parâmetros HTTP
  ├── fiber/          # Adaptador para Fiber
  ├── echo/           # Adaptador para Echo
  ├── gin/            # Adaptador para Gin
  ├── nethttp/        # Adaptador para net/http
  ├── atreugo/        # Adaptador para Atreugo
  ├── fasthttp/       # Adaptador para FastHTTP
  ├── db/             # Utilitários para banco de dados
  └── examples/       # Exemplos de uso
```

## Uso Básico

### Extração de Parâmetros HTTP

#### Fiber

```go
import (
    "context"
    "github.com/gofiber/fiber/v2"
    "github.com/fsvxavier/nexs-lib/paginate"
    pagefiber "github.com/fsvxavier/nexs-lib/paginate/fiber"
)

func ListUsers(c *fiber.Ctx) error {
    ctx := context.Background()
    
    // Extrair parâmetros de paginação da requisição Fiber
    // Campos permitidos para ordenação: "id", "name", "created_at"
    metadata, err := pagefiber.Parse(ctx, c, "id", "name", "created_at")
    if err != nil {
        return c.Status(400).JSON(err)
    }
    
    // ... implementação do serviço ...
    
    return c.JSON(result)
}
```

#### Atreugo

```go
import (
    "context"
    "github.com/savsgio/atreugo/v11"
    "github.com/fsvxavier/nexs-lib/paginate"
    pageAtreugo "github.com/fsvxavier/nexs-lib/paginate/atreugo"
)

func ListUsers(ctx *atreugo.RequestCtx) error {
    appCtx := context.Background()
    
    // Extrair parâmetros de paginação da requisição Atreugo
    // Campos permitidos para ordenação: "id", "name", "created_at"
    metadata, err := pageAtreugo.Parse(appCtx, ctx, "id", "name", "created_at")
    if err != nil {
        return ctx.JSONResponse(map[string]interface{}{"error": err.Error()}, 400)
    }
    
    // ... implementação do serviço ...
    
    return ctx.JSONResponse(result, 200)
}
```

#### FastHTTP

```go
import (
    "context"
    "encoding/json"
    "github.com/valyala/fasthttp"
    "github.com/fsvxavier/nexs-lib/paginate"
    pageFastHTTP "github.com/fsvxavier/nexs-lib/paginate/fasthttp"
)

func HandleUsers(ctx *fasthttp.RequestCtx) {
    appCtx := context.Background()
    
    // Extrair parâmetros de paginação da requisição FastHTTP
    // Campos permitidos para ordenação: "id", "name", "created_at"
    metadata, err := pageFastHTTP.Parse(appCtx, ctx, "id", "name", "created_at")
    if err != nil {
        ctx.SetStatusCode(400)
        response := map[string]interface{}{"error": err.Error()}
        jsonResponse, _ := json.Marshal(response)
        ctx.SetBody(jsonResponse)
        return
    }
    
    // ... implementação do serviço ...
    
    // Retornar resultado paginado
    jsonResponse, _ := json.Marshal(result)
    ctx.SetContentType("application/json")
    ctx.SetBody(jsonResponse)
}
```

### Consulta Paginada no Banco de Dados

```go
import (
    "context"
    "github.com/fsvxavier/nexs-lib/paginate"
    "github.com/fsvxavier/nexs-lib/paginate/db"
)

func GetUsers(ctx context.Context, metadata *page.Metadata, filters map[string]interface{}) (*page.Output, error) {
    // Base da consulta
    baseQuery := "SELECT id, name, email FROM users WHERE is_active = true"
    args := []interface{}{}
    
    // Adicionar filtros dinâmicos
    if name, ok := filters["name"].(string); ok && name != "" {
        baseQuery += " AND name ILIKE $1"
        args = append(args, "%"+name+"%")
    }
    
    // Processador de resultados (converte as linhas para objetos)
    resultProcessor := func(rows interface{}) (interface{}, error) {
        // ... implementação específica para processar resultados ...
        return users, nil
    }
    
    // Executar consulta paginada
    return db.ExecuteQuery(
        ctx,
        sqlExecutor, // Sua implementação de interface de banco de dados
        metadata,
        baseQuery,
        args,
        resultProcessor,
    )
}
```

### Paginação de Slices em Memória

Para cenários onde os dados já estão carregados na memória (como em serviços de cache, microserviços ou APIs intermediárias), a biblioteca oferece funções centralizadas que eliminam a necessidade de implementar manualmente os cálculos de paginação em cada serviço.

#### ✨ Função Centralizada - `ApplyPaginationToSlice`

A função mais simples e recomendada para paginação de slices:

```go
import (
    "context"
    "github.com/fsvxavier/nexs-lib/paginate"
)

type UserService struct {
    users []User // Dados em memória
}

func (s *UserService) GetPaginatedUsers(ctx context.Context, pageNum, limit int, sortField, sortOrder string) (*page.Output, error) {
    // Criar metadados de paginação
    metadata := page.NewMetadata(
        page.WithPage(pageNum),
        page.WithLimit(limit),
        page.WithSort(sortField),
        page.WithOrder(sortOrder),
    )
    
    // 🎯 Uma única linha resolve toda a paginação!
    return page.ApplyPaginationToSlice(ctx, s.users, metadata)
}
```

**Benefícios:**
- ✅ **Cálculo automático** de índices de início e fim
- ✅ **Validação automática** de páginas inválidas
- ✅ **Tratamento de listas vazias** (página 1 sempre válida)
- ✅ **Metadados completos** (total de páginas, próxima/anterior)
- ✅ **Suporte a qualquer tipo** de slice (structs, primitivos, maps)
- ✅ **Tratamento de erros** integrado com edomain

#### 🔧 Função de Baixo Nível - `PaginationIndices`

Para casos onde você precisa apenas dos índices para implementar lógica customizada:

```go
import (
    "fmt"
    "github.com/fsvxavier/nexs-lib/paginate"
)

func PaginateData(data []MyType, pageNum, limit int) ([]MyType, error) {
    // Criar metadados
    metadata := page.NewMetadata(
        page.WithPage(pageNum),
        page.WithLimit(limit),
    )
    
    // Calcular índices de início e fim
    startIndex, endIndex, isValid := page.PaginationIndices(metadata, len(data))
    
    // Verificar se a página é válida
    if !isValid {
        return nil, fmt.Errorf("página inválida: %d", pageNum)
    }
    
    // Aplicar paginação manualmente
    if len(data) == 0 {
        return []MyType{}, nil
    }
    
    return data[startIndex:endIndex], nil
}
```

#### 📊 Comparação: Antes vs Depois

**❌ Código Manual (Antes):**
```go
func (s *Service) ListItems(ctx context.Context, page, limit int) (*page.Output, error) {
    total := len(s.items)
    
    // Calcular índices manualmente
    startIndex := (page - 1) * limit
    if startIndex >= total {
        return nil, errors.New("página inválida")
    }
    
    endIndex := startIndex + limit
    if endIndex > total {
        endIndex = total
    }
    
    // Aplicar paginação
    var paginated []Item
    if startIndex < total {
        paginated = s.items[startIndex:endIndex]
    } else {
        paginated = []Item{}
    }
    
    // Calcular metadados manualmente
    totalPages := (total + limit - 1) / limit
    
    metadata := &page.Metadata{
        Page: page.Page{
            CurrentPage:    page,
            RecordsPerPage: limit,
            TotalPages:     totalPages,
            // ... mais cálculos manuais
        },
        TotalData: total,
    }
    
    return page.NewOutput(paginated, metadata), nil
}
```

**✅ Com Função Centralizada (Depois):**
```go
func (s *Service) ListItems(ctx context.Context, page, limit int) (*page.Output, error) {
    metadata := page.NewMetadata(
        page.WithPage(page),
        page.WithLimit(limit),
    )
    
    // 🎯 Uma linha faz tudo!
    return page.ApplyPaginationToSlice(ctx, s.items, metadata)
}
```

#### 🎭 Suporte a Diferentes Tipos de Slices

A função `ApplyPaginationToSlice` suporta automaticamente diversos tipos:

```go
// Structs customizados
type Product struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
products := []Product{{1, "Produto A"}, {2, "Produto B"}}
result, _ := page.ApplyPaginationToSlice(ctx, products, metadata)

// Maps
items := []map[string]interface{}{
    {"id": 1, "name": "Item 1"},
    {"id": 2, "name": "Item 2"},
}
result, _ := page.ApplyPaginationToSlice(ctx, items, metadata)

// Primitivos
numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
result, _ := page.ApplyPaginationToSlice(ctx, numbers, metadata)

// Strings
names := []string{"Alice", "Bob", "Charlie", "Diana"}
result, _ := page.ApplyPaginationToSlice(ctx, names, metadata)
```

#### 🛡️ Tratamento de Casos Especiais

```go
// Lista vazia - Página 1 sempre válida
emptyList := []Item{}
result, err := page.ApplyPaginationToSlice(ctx, emptyList, metadata)
// ✅ Sucesso: retorna lista vazia com metadados corretos

// Página inválida - Erro claro
metadata := page.NewMetadata(page.WithPage(999), page.WithLimit(10))
result, err := page.ApplyPaginationToSlice(ctx, smallList, metadata)
// ❌ Erro: "Página solicitada é maior que o total de páginas disponíveis"

// Tipo inválido - Erro de validação  
invalidData := "not a slice"
result, err := page.ApplyPaginationToSlice(ctx, invalidData, metadata)
// ❌ Erro: "O parâmetro fornecido não é uma slice válida"
```

## Integração com Outros Frameworks

A biblioteca é projetada para ser extensível. Para adicionar suporte a outros frameworks HTTP, basta criar um adaptador que implemente a interface `HttpRequest`.

```go
type MyFrameworkRequest struct {
    // campos específicos do framework
}

// Query implementa a interface HttpRequest
func (r *MyFrameworkRequest) Query(key string) string {
    // implementação específica para obter parâmetros de consulta
    return ""
}

// QueryParam implementa a interface HttpRequest
func (r *MyFrameworkRequest) QueryParam(key string) string {
    return r.Query(key)
}
```

## 📋 Funções Principais

### 🔧 Funções de Configuração e Criação

| Função | Descrição | Exemplo |
|--------|-----------|---------|
| `NewMetadata(opts ...Option)` | Cria metadados de paginação com valores padrão ou personalizados | `NewMetadata(WithPage(1), WithLimit(10))` |
| `WithPage(page int)` | Define a página atual (mínimo: 1) | `WithPage(2)` |
| `WithLimit(limit int)` | Define registros por página (máximo: 150) | `WithLimit(25)` |
| `WithSort(field string)` | Define campo de ordenação | `WithSort("created_at")` |
| `WithOrder(order string)` | Define direção da ordenação (asc/desc) | `WithOrder("desc")` |
| `WithQuery(query string)` | Define query SQL base | `WithQuery("SELECT * FROM users")` |

### 🎯 Funções de Paginação de Slices (Novas!)

| Função | Descrição | Retorno |
|--------|-----------|---------|
| `ApplyPaginationToSlice(ctx, slice, metadata)` | **Função principal** - Aplica paginação completa a qualquer slice | `(*Output, error)` |
| `PaginationIndices(metadata, totalItems)` | Calcula índices de início e fim para paginação manual | `(startIndex, endIndex, isValid)` |

### 📤 Funções de Saída

| Função | Descrição | Uso |
|--------|-----------|-----|
| `NewOutput(content, metadata)` | Cria saída paginada simples | Para dados já processados |
| `NewOutputWithTotal(ctx, content, totalData, metadata)` | Cria saída com cálculo automático de metadados | Para dados com total conhecido |

### 🌐 Funções de Parsing HTTP

| Framework | Função | Descrição |
|-----------|--------|-----------|
| **Fiber** | `fiber.Parse(ctx, c, allowedFields...)` | Extrai parâmetros de `fiber.Ctx` |
| **Echo** | `echo.Parse(ctx, c, allowedFields...)` | Extrai parâmetros de `echo.Context` |
| **Gin** | `gin.Parse(ctx, c, allowedFields...)` | Extrai parâmetros de `gin.Context` |
| **net/http** | `nethttp.Parse(ctx, r, allowedFields...)` | Extrai parâmetros de `http.Request` |
| **Atreugo** | `atreugo.Parse(ctx, c, allowedFields...)` | Extrai parâmetros de `atreugo.RequestCtx` |
| **FastHTTP** | `fasthttp.Parse(ctx, c, allowedFields...)` | Extrai parâmetros de `fasthttp.RequestCtx` |

### 🗄️ Funções de Banco de Dados

| Função | Descrição | Uso |
|--------|-----------|-----|
| `db.ExecuteQuery(ctx, executor, metadata, baseQuery, args, resultProcessor)` | Executa consulta paginada | Para queries com paginação automática |
| `db.ExecuteTotalQuery(ctx, executor, baseQuery, args)` | Conta total de registros | Para obter total sem paginação |

## 📚 Exemplos Práticos

### 🚀 Início Rápido - Paginação de Slice

```go
package main

import (
    "context"
    "fmt"
    "github.com/fsvxavier/nexs-lib/paginate"
)

func main() {
    ctx := context.Background()
    
    // Dados de exemplo
    items := []map[string]interface{}{
        {"id": 1, "name": "Item 1"},
        {"id": 2, "name": "Item 2"},
        {"id": 3, "name": "Item 3"},
        {"id": 4, "name": "Item 4"},
        {"id": 5, "name": "Item 5"},
    }
    
    // Configurar paginação (página 2, 2 itens por página)
    metadata := page.NewMetadata(
        page.WithPage(2),
        page.WithLimit(2),
    )
    
    // Aplicar paginação
    result, err := page.ApplyPaginationToSlice(ctx, items, metadata)
    if err != nil {
        panic(err)
    }
    
    // Resultado
    fmt.Printf("Página: %d\n", result.Metadata.Page.CurrentPage)           // 2
    fmt.Printf("Total de itens: %d\n", result.Metadata.TotalData)          // 5
    fmt.Printf("Total de páginas: %d\n", result.Metadata.Page.TotalPages)  // 3
    fmt.Printf("Itens retornados: %d\n", len(result.Content.([]interface{}))) // 2
}
```

### 🌐 Exemplos por Framework

Veja exemplos completos de uso no diretório `examples/`:

| Arquivo | Descrição | Framework |
|---------|-----------|-----------|
| `slice_pagination_example.go` | **🆕 Exemplo principal** - Paginação de slices em serviços | Independente |
| `httpservers_agnóstico.go` | **🆕 Atualizado** - Uso das novas funções centralizadas | httpservers |
| `fiber_example.go` | Integração com Fiber | Fiber |
| `echo_example.go` | Integração com Echo | Echo |  
| `gin_example.go` | Integração com Gin | Gin |
| `nethttp_example.go` | Integração com net/http | net/http |
| `atreugo_example.go` | Integração com Atreugo | Atreugo |
| `fasthttp_example.go` | Integração com FastHTTP | FastHTTP |
| `httpservers_example.go` | Uso com httpservers | httpservers |

### 🔗 Exemplo de Integração Completa

```go
// Definir serviço
type ProductService struct {
    products []Product
}

func (s *ProductService) ListProducts(ctx context.Context, pageNum, limit int) (*page.Output, error) {
    metadata := page.NewMetadata(
        page.WithPage(pageNum),
        page.WithLimit(limit),
        page.WithSort("name"),
        page.WithOrder("asc"),
    )
    
    return page.ApplyPaginationToSlice(ctx, s.products, metadata)
}

// Usar em handler HTTP (exemplo com Fiber)
func (h *Handler) GetProducts(c *fiber.Ctx) error {
    ctx := context.Background()
    
    // Extrair parâmetros automaticamente
    metadata, err := pagefiber.Parse(ctx, c, "id", "name", "created_at")
    if err != nil {
        return c.Status(400).JSON(err)
    }
    
    // Usar serviço
    result, err := h.productService.ListProducts(ctx, 
        metadata.Page.CurrentPage, 
        metadata.Page.RecordsPerPage)
    if err != nil {
        return c.Status(500).JSON(err)
    }
    
    return c.JSON(result)
}
```

### 🎯 Casos de Uso

| Cenário | Função Recomendada | Vantagem |
|---------|-------------------|----------|
| **Paginação simples em serviço** | `ApplyPaginationToSlice` | Uma linha resolve tudo |
| **Integração com HTTP** | `framework.Parse` + `ApplyPaginationToSlice` | Parsing automático + paginação |
| **Lógica customizada** | `PaginationIndices` | Controle total dos índices |
| **Consultas de BD** | `db.ExecuteQuery` | Paginação otimizada no BD |

## 🚀 Vantagens da Nova Implementação

### ✅ Antes vs Depois

| Aspecto | Implementação Manual | Com Biblioteca |
|---------|---------------------|----------------|
| **Linhas de código** | ~30 linhas | 1 linha |
| **Cálculo de índices** | Manual, propenso a erros | Automático e testado |
| **Validação de páginas** | Implementação customizada | Validação robusta integrada |
| **Tratamento de listas vazias** | Lógica adicional necessária | Tratado automaticamente |
| **Metadados** | Cálculo manual de páginas/navegação | Gerado automaticamente |
| **Tipos suportados** | Apenas tipos específicos | Qualquer tipo de slice |
| **Tratamento de erros** | Errors básicos | Integração com edomain |
| **Testabilidade** | Testes para cada implementação | Testes centralizados |

### 🛡️ Robustez e Confiabilidade

- ✅ **Testado extensivamente** com 16+ cenários de teste
- ✅ **Tratamento de edge cases** (listas vazias, páginas inválidas)
- ✅ **Validação de tipos** automática com mensagens claras
- ✅ **Integração com edomain** para tratamento consistente de erros
- ✅ **Performance otimizada** sem cópias desnecessárias de dados

## 🔧 Parâmetros HTTP Suportados

A biblioteca automaticamente reconhece e processa os seguintes parâmetros de query string:

| Parâmetro | Descrição | Valor Padrão | Exemplo |
|-----------|-----------|--------------|---------|
| `page` | Número da página (mínimo: 1) | 1 | `?page=2` |
| `limit` | Itens por página (máximo: 150) | 150 | `?limit=10` |
| `sort` | Campo para ordenação | "id" | `?sort=name` |
| `order` | Direção da ordenação | "asc" | `?order=desc` |

**Exemplo de URL completa:**
```
GET /api/users?page=2&limit=25&sort=created_at&order=desc
```

## 🏗️ Estrutura da Resposta

Todas as funções de paginação retornam um objeto `Output` padronizado:

```json
{
  "content": [
    {"id": 1, "name": "Item 1"},
    {"id": 2, "name": "Item 2"}
  ],
  "metadata": {
    "pagination": {
      "current_page": 2,
      "records_per_page": 10,
      "total_pages": 5,
      "previous": 1,
      "next": 3
    },
    "total_data": 42,
    "sort": {
      "field": "id",
      "order": "asc"
    }
  }
}
```

## 🚦 Status e Roadmap

### ✅ Funcionalidades Implementadas

- [x] Adaptadores para todos os principais frameworks Go HTTP
- [x] Paginação centralizada de slices em memória  
- [x] Cálculo automático de índices de paginação
- [x] Integração com banco de dados
- [x] Validação robusta de parâmetros
- [x] Tratamento de erros com edomain
- [x] Suporte a tipos genéricos de slices
- [x] Documentação completa e exemplos

### 🔄 Melhorias Futuras

- [ ] Cache automático de resultados paginados
- [ ] Suporte a cursor-based pagination
- [ ] Métricas de performance integradas
- [ ] Adaptadores para mais frameworks (iris, revel, etc.)
- [ ] Filtros dinâmicos integrados

## 📞 Suporte e Contribuição

### 🐛 Reportar Issues

Para reportar bugs ou solicitar funcionalidades, abra uma issue no repositório principal.

### 🤝 Contribuindo

1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

### 📋 Requisitos para Contribuição

- Testes unitários para novas funcionalidades
- Documentação atualizada
- Compatibilidade com Go 1.19+
- Seguir os padrões da biblioteca edomain

## 📄 Licença

Esta biblioteca segue a mesma licença do projeto principal isis-golang-lib.

---

**Desenvolvido com ❤️ para simplificar paginação em aplicações Go**
