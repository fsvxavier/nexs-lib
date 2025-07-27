# Exemplo 1: Uso Básico da Paginação

Este exemplo demonstra o uso básico do módulo de paginação, incluindo:

- Configuração do serviço
- Parse de parâmetros de query
- Construção de queries SQL
- Criação de resposta paginada
- Navegação entre páginas

## Como executar

```bash
cd examples/01-basic-usage
go run main.go
```

## O que o exemplo faz

1. **Configura o serviço** com parâmetros padrão
2. **Simula parâmetros HTTP** como `/api/users?page=2&limit=5&sort=name&order=desc`
3. **Faz parse** dos parâmetros com validação
4. **Constrói queries SQL** automaticamente com ORDER BY e LIMIT/OFFSET
5. **Cria resposta paginada** com metadados de navegação
6. **Exibe informações** de navegação (página anterior/próxima)

## Saída esperada

```json
{
  "content": [
    {
      "id": 1,
      "name": "Alice Johnson",
      "email": "alice@example.com",
      "active": true
    },
    // ... mais usuários
  ],
  "metadata": {
    "current_page": 2,
    "records_per_page": 5,
    "total_pages": 5,
    "total_records": 23,
    "previous": 1,
    "next": 3,
    "sort_field": "name",
    "sort_order": "desc"
  }
}
```

## Conceitos demonstrados

- ✅ **Configuração básica** do serviço de paginação
- ✅ **Parse de parâmetros** de query HTTP
- ✅ **Validação** de campos de ordenação
- ✅ **Construção automática** de queries SQL
- ✅ **Metadados de navegação** (próxima/anterior)
- ✅ **Resposta estruturada** em JSON

## Próximos passos

Após entender este exemplo básico, veja:
- `02-fiber-integration` - Integração com Fiber framework
- `03-custom-config` - Configuração personalizada
- `04-error-handling` - Tratamento de erros avançado
