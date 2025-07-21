# Simple Middleware Example

Este exemplo demonstra uso básico do sistema de middleware com configuração mínima.

## Funcionalidades

- **CORS**: Configurado para permitir todos os origins
- **Logging**: Logs simples no console
- **Rate Limiting**: 10 requests por minuto (para demonstração)

## Como Executar

```bash
cd examples/middleware/simple
go run main.go
```

## Testando

```bash
# Requisição básica
curl http://localhost:8080

# Testar rate limiting (execute rapidamente)
for i in {1..15}; do curl http://localhost:8080; echo; done
```

## Saída Esperada

```
Simple middleware example starting on :8080
Try: curl http://localhost:8080
Rate limit: 10 requests per minute

[15:30:45] GET / - 200 (1.234ms)
[15:30:46] GET / - 200 (987μs)
[15:30:47] GET / - 429 (123μs)  # Rate limit exceeded
```

Este exemplo é ideal para entender os conceitos básicos do sistema de middleware.
