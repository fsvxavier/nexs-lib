# Exemplo Completo - Hooks + Middlewares

Este exemplo demonstra o uso completo e integrado dos sistemas de hooks e middlewares da biblioteca `nexs-lib/httpserver`.

## 📋 O que este exemplo demonstra

- **Sistema Completo de Hooks**: Todos os 7 tipos de hooks trabalhando em conjunto
- **Sistema Completo de Middlewares**: Logging avançado e autenticação multi-método
- **Integração Hooks + Middlewares**: Como combinar ambos para máximo monitoramento
- **API Completa**: Endpoints públicos, protegidos e administrativos
- **Métricas Avançadas**: Coleta e exposição de métricas detalhadas

## 🚀 Como executar

```bash
cd httpserver/examples/complete
go run main.go
```

O servidor iniciará na porta 8080.

## 🏗️ Arquitetura do Exemplo

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP Request  │───▶│   Middlewares    │───▶│     Hooks       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │                         │
                              ▼                         ▼
                       ┌──────────────┐         ┌──────────────┐
                       │   Logging    │         │ Monitoring   │
                       │     Auth     │         │   Metrics    │
                       └──────────────┘         └──────────────┘
```

## 🧪 Testando os endpoints

### Endpoints Públicos

```bash
# Página inicial com informações do sistema
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health

# Área pública
curl http://localhost:8080/public

# Documentação da API
curl http://localhost:8080/docs
```

### Endpoints Protegidos - Basic Auth

```bash
# Lista de usuários
curl -u admin:admin123 http://localhost:8080/api/users

# Criar novo usuário
curl -u admin:admin123 -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"Novo Usuario","role":"user"}' \
  http://localhost:8080/api/users

# Perfil do usuário atual
curl -u user:user123 http://localhost:8080/api/profile

# Estatísticas em tempo real
curl -u developer:dev123 http://localhost:8080/api/stats

# Teste de rota lenta (>500ms = alerta)
curl -u admin:admin123 http://localhost:8080/api/slow

# Simular erro para testing
curl -u admin:admin123 http://localhost:8080/api/error
```

### Endpoints Protegidos - API Key

```bash
# Usando API key com acesso limitado
curl -H "X-API-Key: api-key-123" http://localhost:8080/api/users

# Usando API key administrativa
curl -H "X-API-Key: admin-key-456" http://localhost:8080/admin/dashboard

# Métricas completas do sistema
curl -H "X-API-Key: admin-key-456" http://localhost:8080/metrics
```

### Área Administrativa

```bash
# Dashboard administrativo
curl -u admin:admin123 http://localhost:8080/admin/dashboard

# Logs do sistema
curl -u admin:admin123 http://localhost:8080/admin/logs
```

## 📊 Funcionalidades Avançadas

### 1. Sistema de Hooks Completo

| Hook | Função | Métricas Coletadas |
|------|--------|-------------------|
| **StartHook** | Inicialização do servidor | Contagem de starts, tempo de inicialização |
| **StopHook** | Parada do servidor | Contagem de stops, uptime, shutdown graceful |
| **RequestHook** | Rastreamento de requisições | Total, ativas, pico concorrente, tamanho |
| **ResponseHook** | Rastreamento de respostas | Tempo de resposta, códigos de status, tamanho |
| **ErrorHook** | Monitoramento de erros | Total de erros, tipos, threshold alerts |
| **RouteInHook** | Entrada em rotas | Contagem por rota, patterns |
| **RouteOutHook** | Saída de rotas | Duração, alertas de latência |

### 2. Sistema de Middlewares Avançado

#### Logging Middleware
- **Logs Estruturados**: Formato JSON para facilitar parsing
- **Filtragem Inteligente**: Skip de health checks e arquivos estáticos
- **Sanitização**: Remove dados sensíveis automaticamente
- **Controle de Tamanho**: Limita logs de body grandes

#### Authentication Middleware
- **Multi-método**: Basic Auth + API Keys
- **Usuários Configuráveis**: Diferentes níveis de acesso
- **Filtros de Rota**: Rotas públicas vs protegidas
- **Tokens com Roles**: Controle granular de permissões

### 3. Métricas em Tempo Real

Acesse `/metrics` para ver:

```json
{
  "timestamp": 1659456789,
  "hooks": {
    "registered": 7,
    "list": ["start", "stop", "request", "response", "error", "route-in", "route-out"]
  },
  "requests": {
    "total": 42,
    "active": 3,
    "max_concurrent": 8,
    "total_size": 15360,
    "average_size": 366
  },
  "server": {
    "start_count": 1,
    "stop_count": 0,
    "is_running": true
  }
}
```

## 🔐 Sistema de Autenticação

### Credenciais Basic Auth

| Usuário | Senha | Permissões |
|---------|-------|------------|
| admin | admin123 | Acesso total (API + Admin) |
| user | user123 | Acesso à API |
| developer | dev123 | Acesso à API + Stats |

### API Keys

| Key | Acesso | Descrição |
|-----|--------|-----------|
| api-key-123 | Read-only | Acesso limitado à API |
| admin-key-456 | Full access | Acesso total incluindo admin |

## 📁 Estrutura de Rotas

```
/                     - Landing page (público)
/health              - Health check (público)
/public              - Área pública (público)
/docs                - Documentação (público)

/api/                - API principal (protegido)
  ├── users          - CRUD de usuários
  ├── profile        - Perfil do usuário
  ├── stats          - Estatísticas em tempo real
  ├── slow           - Teste de performance
  └── error          - Teste de erro

/admin/              - Área administrativa (protegido)
  ├── dashboard      - Dashboard principal
  └── logs           - Logs do sistema

/metrics             - Métricas completas (protegido)
```

## 🔍 Logs Detalhados

Durante a execução, você verá logs abrangentes incluindo:

```
🚀 Exemplo Completo - Hooks + Middlewares
✅ 7 hooks registrados
✅ 2 middlewares configurados
🌟 Servidor completo iniciado na porta 8080

[INFO] Request received (ID: 0xc000...) (hook: request-monitor)
[INFO] Entering route: GET /api/users (hook: route-in-monitor)
[INFO] Processing request: GET /api/users (middleware: logging)
[INFO] Authentication successful: admin (middleware: auth)
[INFO] Route GET /api/users completed in 105ms (hook: route-out-monitor)
[INFO] Response sent (Request ID: 0xc000..., Duration: 105ms) (hook: response-monitor)
```

## 🎯 Casos de Uso Demonstrados

1. **Monitoramento de Performance**: Hook de RouteOut detecta rotas lentas
2. **Rastreamento de Erros**: ErrorHook com thresholds automáticos
3. **Métricas de Carga**: RequestHook mostra picos de concorrência
4. **Auditoria de Acesso**: Logging de todas as tentativas de auth
5. **Health Monitoring**: Métricas de uptime e disponibilidade

## 🛠️ Personalizações Avançadas

### Adicionar Novos Hooks

```go
customHook := hooks.NewCustomHook("my-hook")
customHook.SetMetricsEnabled(true)
hookManager.RegisterHook("custom", customHook)
```

### Adicionar Novos Middlewares

```go
corsMiddleware := middlewares.NewCORSMiddleware(2)
middlewareManager.AddMiddleware(corsMiddleware)
```

### Configurar Alertas

```go
errorHook.SetErrorThreshold(10)  // Alertar após 10 erros
routeOutHook.SetSlowThreshold(1 * time.Second)  // Alertar rotas > 1s
```

## 📈 Monitoramento em Produção

Este exemplo pode ser usado como base para:

- **APM Integration**: Conectar com Datadog, New Relic, etc.
- **Metrics Export**: Prometheus, Grafana
- **Log Aggregation**: ELK Stack, Splunk
- **Alerting**: PagerDuty, Slack notifications

## 🎓 Próximos Passos

Após dominar este exemplo completo:

1. Explore [exemplos específicos por framework](../)
2. Implemente métricas customizadas
3. Adicione integração com sistemas de monitoramento
4. Configure alertas automatizados
5. Implemente circuit breakers e rate limiting
