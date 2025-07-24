# Health Check Middleware Example

Este exemplo demonstra o uso completo do middleware de health checks da nexs-lib.

## 🏥 Sobre Health Checks

Os health checks são essenciais para monitoramento e orquestração de serviços. Este middleware implementa os três tipos principais de probes:

- **Liveness Probe** - Verifica se a aplicação está funcionando
- **Readiness Probe** - Verifica se a aplicação está pronta para receber tráfego
- **Startup Probe** - Verifica se a aplicação iniciou corretamente

## 🚀 Executando o Exemplo

```bash
go run main.go
```

O servidor iniciará na porta `:8080` com os seguintes endpoints disponíveis:

## 📍 Endpoints

### Health Check Endpoints

| Endpoint | Descrição | Tipo |
|----------|-----------|------|
| `GET /health` | Status geral de saúde | All checks |
| `GET /health/live` | Liveness probe | Liveness only |
| `GET /health/ready` | Readiness probe | Readiness only |
| `GET /health/startup` | Startup probe | Startup only |

### API Endpoints

| Endpoint | Descrição |
|----------|-----------|
| `GET /api/hello` | Endpoint de exemplo |

## 🔍 Health Checks Implementados

### 1. Ping Check (Liveness)
```go
registry.Register("ping", simpleCheck("Ping check passed"),
    health.WithType(health.CheckTypeLiveness))
```
- Verifica se a aplicação responde
- Sempre retorna sucesso

### 2. Database Check (Readiness)
```go
registry.Register("database", databaseCheck(),
    health.WithType(health.CheckTypeReadiness),
    health.WithCritical(true))
```
- Simula verificação de conexão com banco de dados
- Marcado como crítico
- Inclui metadados (conexões ativas, latência)

### 3. External API Check (Liveness)
```go
registry.Register("external-api", health.URLCheck("https://httpbin.org/status/200"),
    health.WithType(health.CheckTypeLiveness),
    health.WithCritical(false))
```
- Verifica conectividade com API externa
- Não crítico (falha não afeta status geral)

### 4. Memory Check (Liveness)
```go
registry.Register("memory", health.MemoryCheck(1024),
    health.WithType(health.CheckTypeLiveness))
```
- Verifica uso de memória (limite: 1GB)
- Automático baseado em métricas do sistema

### 5. Disk Space Check (Liveness)
```go
registry.Register("disk", health.DiskSpaceCheck("/tmp", 1),
    health.WithType(health.CheckTypeLiveness))
```
- Verifica espaço em disco (mínimo: 1GB)
- Monitora diretório `/tmp`

## 🧪 Testando

### Teste Básico
```bash
# Verificar saúde geral
curl http://localhost:8080/health

# Verificar liveness
curl http://localhost:8080/health/live

# Verificar readiness
curl http://localhost:8080/health/ready
```

### Respostas Esperadas

#### Resposta Saudável (`/health`)
```json
{
  "status": "healthy",
  "timestamp": "2025-07-20T15:04:05Z",
  "checks": {
    "ping": {
      "status": "healthy",
      "message": "Ping check passed",
      "type": "liveness"
    },
    "database": {
      "status": "healthy",
      "message": "Database connection successful",
      "type": "readiness",
      "critical": true,
      "metadata": {
        "connection_time": "100ms",
        "active_connections": 10,
        "max_connections": 100
      }
    },
    "external-api": {
      "status": "healthy",
      "message": "External API reachable",
      "type": "liveness",
      "critical": false
    }
  }
}
```

#### Resposta com Falha
```json
{
  "status": "unhealthy",
  "timestamp": "2025-07-20T15:04:05Z",
  "checks": {
    "database": {
      "status": "unhealthy",
      "message": "Database connection failed",
      "type": "readiness",
      "critical": true,
      "metadata": {
        "error": "connection timeout"
      }
    }
  }
}
```

## 🔧 Configuração

### Adicionando Novos Health Checks

```go
// Health check customizado
registry.Register("my-service", func(ctx context.Context) interfaces.HealthCheckResult {
    // Sua lógica de verificação aqui
    return interfaces.HealthCheckResult{
        Status:  "healthy",
        Message: "Service is running",
        Metadata: map[string]interface{}{
            "version": "1.0.0",
            "uptime": "1h30m",
        },
    }
}, health.WithType(health.CheckTypeReadiness))
```

### Opções de Configuração

- `health.WithType(type)` - Define o tipo do check
- `health.WithCritical(bool)` - Define se é crítico
- `health.WithInterval(duration)` - Intervalo de execução
- `health.WithTimeout(duration)` - Timeout do check

## 📊 Monitoramento

### Kubernetes Integration

```yaml
# Deployment com health checks
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Docker Health Check

```dockerfile
# Dockerfile com health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health/live || exit 1
```

## 🔍 Troubleshooting

### Problemas Comuns

**Health check sempre falha:**
- Verifique se o serviço está realmente rodando
- Confirme as dependências (banco, APIs externas)
- Verifique logs de erro nos metadados

**Timeout nos checks:**
- Aumente o timeout dos checks
- Verifique conectividade de rede
- Otimize as consultas de verificação

**Status inconsistente:**
- Verifique se os checks são determinísticos
- Evite checks que dependem de condições externas variáveis
- Use cache para checks custosos

## 📚 Referências

- [Health Check Patterns](https://microservices.io/patterns/observability/health-check-api.html)
- [Kubernetes Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Docker Health Checks](https://docs.docker.com/engine/reference/builder/#healthcheck)
