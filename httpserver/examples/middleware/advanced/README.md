# Advanced Middleware Example

Este exemplo demonstra o uso completo do sistema de middleware com todas as funcionalidades implementadas.

## Funcionalidades Demonstradas

### 1. Health Checks
- **Liveness Probes**: `/health/live`
- **Readiness Probes**: `/health/ready` 
- **Comprehensive Health**: `/health`

### 2. Middleware Chain Completa
- **CORS**: Configurado para permitir todos os origins (desenvolvimento)
- **Logging**: Logs estruturados de todas as requisições
- **Timeout**: 30 segundos por requisição
- **Rate Limiting**: 100 requests/minuto com Token Bucket
- **Bulkhead**: Isolamento por tipo de endpoint
- **Retry**: Até 2 tentativas para falhas transitórias
- **Compression**: Compressão automática de respostas

### 3. Endpoints de Demonstração
- `GET /api/users` - Lista de usuários
- `GET /api/orders` - Lista de pedidos  
- `GET /api/heavy-task` - Tarefa pesada (demonstra bulkhead)
- `GET /api/flaky-endpoint` - Endpoint instável (demonstra retry)

## Como Executar

```bash
# A partir da pasta do exemplo
go run main.go
```

## Testando as Funcionalidades

### Health Checks

```bash
# Health check completo
curl http://localhost:8080/health

# Liveness probe
curl http://localhost:8080/health/live

# Readiness probe  
curl http://localhost:8080/health/ready
```

### Rate Limiting

```bash
# Testar rate limiting (execute múltiplas vezes rapidamente)
for i in {1..10}; do curl http://localhost:8080/api/users; done
```

### Bulkhead Pattern

```bash
# Testar isolamento de recursos
# Abra múltiplas conexões para heavy-task
curl http://localhost:8080/api/heavy-task &
curl http://localhost:8080/api/heavy-task &
curl http://localhost:8080/api/users  # Deve funcionar normalmente
```

### Retry Policies

```bash
# Testar retry automático
curl http://localhost:8080/api/flaky-endpoint
curl http://localhost:8080/api/flaky-endpoint
curl http://localhost:8080/api/flaky-endpoint  # Deve ter sucesso
```

### Compression

```bash
# Testar compressão (observe o header Content-Encoding)
curl -H "Accept-Encoding: gzip" -v http://localhost:8080/api/users
```

### CORS

```bash
# Testar CORS preflight
curl -X OPTIONS \
  -H "Origin: https://example.com" \
  -H "Access-Control-Request-Method: GET" \
  -v http://localhost:8080/api/users
```

## Logs de Exemplo

O sistema gera logs estruturados como este:

```json
{
  "correlation_id": "a1b2c3d4e5f6",
  "method": "GET",
  "path": "/api/users",
  "query": "",
  "remote_addr": "127.0.0.1:54321",
  "user_agent": "curl/7.68.0",
  "status_code": 200,
  "response_size": 156,
  "duration": "1.234ms",
  "timestamp": "2025-01-20T10:30:45Z"
}
```

## Headers HTTP de Exemplo

### Rate Limiting Headers
```
X-Rate-Limit-Limit: 100
X-Rate-Limit-Remaining: 99
X-Rate-Limit-Reset: 1642680645
```

### CORS Headers
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Vary: Origin
```

### Compression Headers
```
Content-Encoding: gzip
Vary: Accept-Encoding
```

### Correlation ID
```
X-Correlation-ID: a1b2c3d4e5f6
```

## Monitoramento

Este exemplo demonstra como integrar:
- Health checks para Kubernetes
- Logs estruturados para centralização
- Headers HTTP para debugging
- Métricas de performance
- Isolation pattern para resiliência

## Configuração para Produção

Para produção, ajuste:

1. **CORS**: Restringir origins específicos
2. **Rate Limiting**: Ajustar limites por ambiente
3. **Timeouts**: Valores conservadores
4. **Bulkhead**: Recursos por criticidade
5. **Logging**: Reduzir verbosidade
6. **Health Checks**: Checks reais de dependências

## Arquivos de Configuração

O exemplo pode ser estendido com:
- Configuração via environment variables
- Arquivos YAML/JSON de configuração
- Integration com service discovery
- Métricas Prometheus
- Tracing distribuído
