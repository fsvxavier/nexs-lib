# Exemplo 2: Integração com Fiber Framework

Este exemplo demonstra a integração completa do módulo de paginação com o framework Fiber, incluindo:

- API REST completa com paginação
- Diferentes endpoints com filtros
- Tratamento de erros personalizado
- Interface web para testes
- Múltiplos cenários de uso

## Pré-requisitos

```bash
go mod init pagination-fiber-example
go get github.com/gofiber/fiber/v2
```

## Como executar

```bash
cd examples/02-fiber-integration
go run main.go
```

O servidor iniciará na porta `:3000`.

## Endpoints da API

### 🌐 Interface Web
- **GET /** - Página com documentação e links de teste

### 📦 Produtos
- **GET /api/products** - Lista produtos com paginação
- **GET /api/products/in-stock** - Lista apenas produtos em estoque
- **GET /api/products/error-demo** - Demonstra tratamento de erros

### ℹ️ Informações
- **GET /api/info** - Informações da API e parâmetros

## Parâmetros de Query

| Parâmetro | Descrição | Padrão | Limite |
|-----------|-----------|---------|--------|
| `page` | Número da página | 1 | - |
| `limit` | Registros por página | 5 | 50 |
| `sort` | Campo de ordenação | id | id, name, price, category |
| `order` | Ordem de classificação | asc | asc, desc |

## Exemplos de Uso

### Listagem básica
```bash
curl "http://localhost:3000/api/products"
```

### Paginação específica
```bash
curl "http://localhost:3000/api/products?page=2&limit=3"
```

### Ordenação por preço (decrescente)
```bash
curl "http://localhost:3000/api/products?sort=price&order=desc"
```

### Produtos em estoque
```bash
curl "http://localhost:3000/api/products/in-stock?limit=2"
```

### Teste de erro (campo de ordenação inválido)
```bash
curl "http://localhost:3000/api/products/error-demo?sort=invalid_field"
```

## Resposta da API

```json
{
  "content": [
    {
      "id": 1,
      "name": "Smartphone Galaxy",
      "description": "Smartphone Android",
      "price": 899.99,
      "category": "Electronics",
      "in_stock": true
    }
  ],
  "metadata": {
    "current_page": 1,
    "records_per_page": 5,
    "total_pages": 3,
    "total_records": 15,
    "next": 2,
    "sort_field": "id",
    "sort_order": "asc"
  }
}
```

## Tratamento de Erros

### Parâmetros inválidos
```json
{
  "error": "Invalid pagination parameters",
  "details": "[INVALID_SORT_FIELD] sort field must be one of: [id, name, price, category]"
}
```

### Limite excedido
```json
{
  "error": "Invalid pagination parameters", 
  "details": "[LIMIT_TOO_LARGE] limit cannot exceed 50"
}
```

## Conceitos demonstrados

- ✅ **Integração com Fiber** usando provider específico
- ✅ **API REST completa** com múltiplos endpoints
- ✅ **Diferentes filtros** (produtos em estoque)
- ✅ **Tratamento de erros** robusto
- ✅ **Interface web** para testes
- ✅ **Documentação automática** via endpoint `/api/info`
- ✅ **Múltiplos cenários** de paginação

## Estrutura do Código

### ProductRepository
Simula um repositório de dados com 15 produtos de exemplo.

### Endpoints
1. `/api/products` - Listagem principal com paginação completa
2. `/api/products/in-stock` - Filtro personalizado
3. `/api/products/error-demo` - Demonstração de validação
4. `/api/info` - Documentação automática
5. `/` - Interface web para testes

### Middleware
- **Logger** - Log de requisições
- **CORS** - Suporte a requisições cross-origin

## Próximos passos

Após entender este exemplo, veja:
- `03-custom-config` - Configuração personalizada avançada
- `04-error-handling` - Tratamento de erros mais robusto
- `05-database-integration` - Integração com banco de dados real
