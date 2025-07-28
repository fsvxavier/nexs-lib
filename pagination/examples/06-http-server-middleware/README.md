# HTTP Server Middleware Example

Este exemplo demonstra como usar o middleware de paginação com servidores HTTP padrão do Go.

## Funcionalidades Demonstradas

- ✅ Middleware automático de paginação
- ✅ Configuração de campos ordenáveis por rota
- ✅ Hooks personalizados para logging e auditoria
- ✅ Resposta paginada automática
- ✅ Tratamento de erros personalizado
- ✅ Headers de paginação automáticos

## Como Executar

```bash
cd examples/06-http-server-middleware
go run main.go
```

O servidor iniciará na porta 8080.

## Endpoints Disponíveis

### GET /api/users
Retorna uma lista paginada de usuários.

**Parâmetros de Query:**
- `page` (int): Número da página (padrão: 1)
- `limit` (int): Itens por página (padrão: 50, máximo: 150)
- `sort` (string): Campo para ordenação (id, name, email, created_at)
- `order` (string): Direção da ordenação (asc, desc)

**Exemplo:**
```bash
curl "http://localhost:8080/api/users?page=2&limit=3&sort=name&order=desc"
```

### GET /api/posts
Endpoint de exemplo para demonstrar configuração por rota.

**Parâmetros de Query:**
- `page` (int): Número da página (padrão: 1)
- `limit` (int): Itens por página (padrão: 50, máximo: 150)
- `sort` (string): Campo para ordenação (id, title, created_at)
- `order` (string): Direção da ordenação (asc, desc)

**Exemplo:**
```bash
curl "http://localhost:8080/api/posts?page=1&limit=5&sort=title&order=asc"
```

### GET /health
Endpoint de health check (sem paginação).

## Resposta de Exemplo

```json
{
  "content": [
    {
      "id": 1,
      "name": "Alice Johnson",
      "email": "alice@example.com",
      "created_at": "2024-01-01T10:00:00Z"
    },
    {
      "id": 2,
      "name": "Bob Smith",
      "email": "bob@example.com",
      "created_at": "2024-01-02T10:00:00Z"
    }
  ],
  "metadata": {
    "current_page": 1,
    "records_per_page": 2,
    "total_pages": 5,
    "total_records": 10,
    "next": 2,
    "sort_field": "id",
    "sort_order": "asc"
  }
}
```

## Headers HTTP

O middleware adiciona automaticamente headers informativos:

```
X-Pagination-Page: 1
X-Pagination-Limit: 50
X-Pagination-Sort: name
X-Pagination-Order: asc
```

## Hooks Personalizados

Este exemplo demonstra como implementar hooks para:

- **Logging**: Registra todas as operações de paginação
- **Auditoria**: Monitora acessos e parâmetros utilizados
- **Performance**: Mede tempo de execução das operações

## Configuração Avançada

```go
// Configurar campos ordenáveis por rota
paginationConfig.ConfigureRoute("/api/users", []string{"id", "name", "email", "created_at"})

// Adicionar caminhos que devem pular a paginação
paginationConfig.AddSkipPath("/health")
paginationConfig.AddSkipPath("/metrics")

// Configurar hooks personalizados
paginationConfig.WithHooks().
    PreValidation(NewLoggingHook("pre-validation")).
    PostValidation(NewLoggingHook("post-validation")).
    Done()
```

## Tratamento de Erros

O middleware trata automaticamente:

- Parâmetros inválidos (page, limit)
- Campos de ordenação não permitidos
- Valores fora dos limites configurados

Erros retornam status HTTP 400 com detalhes:

```json
{
  "error": {
    "message": "page must be greater than 0",
    "type": "pagination_validation_error"
  }
}
```
