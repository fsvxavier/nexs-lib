# Exemplos de Middlewares

Esta pasta contém exemplos práticos demonstrando como usar os middlewares da biblioteca nexs-lib.

## Exemplos Disponíveis

### `complete_example.go`

Exemplo completo demonstrando o uso de todos os novos middlewares:

- **Body Validator**: Valida JSON em requisições POST
- **Content Type**: Garante Content-Type correto para APIs JSON
- **Tenant ID**: Extrai tenant ID de headers para aplicações multi-tenant
- **Trace ID**: Gera/extrai IDs de rastreamento para observabilidade
- **Error Handler**: Captura panics e formata erros de forma padronizada

#### Como executar

```bash
# Construir o exemplo
go build -o complete_example complete_example.go

# Executar
./complete_example
```

O servidor será iniciado em `http://localhost:8080` com os seguintes endpoints:

- `GET /health` - Health check simples
- `GET /users` - Lista usuários (demonstra extração de contexto)
- `POST /users/create` - Cria usuário (demonstra validação de body)
- `GET /panic` - Demonstra tratamento de panics

#### Exemplos de requisições

**1. Health check (sem validações):**
```bash
curl http://localhost:8080/health
```

**2. Listar usuários (com trace e tenant):**
```bash
curl -H "X-Trace-ID: trace-123" \
     -H "X-Tenant-ID: tenant-abc" \
     http://localhost:8080/users
```

**3. Criar usuário (validação completa):**
```bash
curl -X POST http://localhost:8080/users/create \
     -H "Content-Type: application/json" \
     -H "X-Tenant-ID: tenant-abc" \
     -H "X-Trace-ID: trace-456" \
     -d '{"name":"João Silva","email":"joao@example.com"}'
```

**4. Testar Content-Type inválido:**
```bash
curl -X POST http://localhost:8080/users/create \
     -H "Content-Type: text/plain" \
     -H "X-Tenant-ID: tenant-abc" \
     -d '{"name":"João Silva","email":"joao@example.com"}'
```

**5. Testar JSON inválido:**
```bash
curl -X POST http://localhost:8080/users/create \
     -H "Content-Type: application/json" \
     -H "X-Tenant-ID: tenant-abc" \
     -d '{"name":"João Silva","email":}'
```

**6. Demonstrar error handler:**
```bash
curl http://localhost:8080/panic
```

#### Respostas esperadas

Todas as respostas incluem informações de contexto:

```json
{
  "success": true,
  "data": { ... },
  "trace_id": "generated-or-extracted-trace-id",
  "tenant": "tenant-from-header-or-default"
}
```

Em caso de erro:

```json
{
  "success": false,
  "error": "Error message",
  "status": 400,
  "trace_id": "trace-id",
  "timestamp": "2025-01-21T00:00:00Z"
}
```

## Observações sobre Ordem de Middlewares

No exemplo, os middlewares são aplicados em ordem específica usando wrapping manual:

1. **Error Handler** (outermost) - Captura todos os erros e panics
2. **Trace ID** - Gera/extrai ID cedo para rastreamento
3. **Tenant ID** - Identifica tenant
4. **Content Type** - Valida Content-Type antes de processar body
5. **Body Validator** (innermost) - Valida JSON do corpo da requisição

Esta ordem garante que:
- Erros são sempre capturados e formatados corretamente
- Trace IDs estão disponíveis em todo o pipeline
- Validações ocorrem na ordem lógica (Content-Type → Body)

## Desenvolvimento

Para adicionar novos exemplos:

1. Crie um novo arquivo `.go` nesta pasta
2. Importe os middlewares necessários
3. Siga o padrão dos exemplos existentes
4. Documente no README

## Testes

Para testar os exemplos automaticamente, você pode usar o script de testes:

```bash
# Executar todos os testes dos middlewares
go test -v ../...

# Executar testes com cobertura
go test -cover ../...
```
