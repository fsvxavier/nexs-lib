# Nexs Observability infraestructure

Infraestrutura Docker completa para testes e desenvolvimento dos componentes de observabilidade (logger e tracer) da Nexs Library.

## 📋 Índice

- [Visão Geral](#visão-geral)
- [Serviços Incluídos](#serviços-incluídos)
- [Pré-requisitos](#pré-requisitos)
- [Instalação e Uso](#instalação-e-uso)
- [Configuração](#configuração)
- [URLs dos Serviços](#urls-dos-serviços)
- [Exemplos de Uso](#exemplos-de-uso)
- [Solução de Problemas](#solução-de-problemas)
- [Contribuição](#contribuição)

## 🎯 Visão Geral

Esta infraestrutura fornece um ambiente completo de observabilidade para desenvolvimento e testes, incluindo:

- **Tracing**: Jaeger, Tempo, OpenTelemetry Collector
- **Logging**: Elasticsearch, Logstash, Fluentd, Kibana
- **Metrics**: Prometheus, Grafana
- **Databases**: PostgreSQL, MongoDB, Redis, RabbitMQ

## 🔧 Serviços Incluídos

### Tracing Stack
| Serviço | Versão | Porta | Descrição |
|---------|---------|-------|-----------|
| Jaeger | latest | 16686 | UI para visualização de traces |
| Tempo | latest | 3200 | Backend de tracing da Grafana |
| OTEL Collector | latest | 4317/4318 | Coletor OpenTelemetry |

### Logging Stack
| Serviço | Versão | Porta | Descrição |
|---------|---------|-------|-----------|
| Elasticsearch | 8.11.0 | 9200 | Motor de busca para logs |
| Logstash | 8.11.0 | 5044 | Processamento de logs |
| Fluentd | latest | 24224 | Coletor de logs |
| Kibana | 8.11.0 | 5601 | UI para visualização de logs |

### Metrics Stack
| Serviço | Versão | Porta | Descrição |
|---------|---------|-------|-----------|
| Prometheus | latest | 9090 | Coleta e armazenamento de métricas |
| Grafana | latest | 3000 | Dashboards e visualizações |

### Databases
| Serviço | Versão | Porta | Descrição |
|---------|---------|-------|-----------|
| PostgreSQL | 15 | 5432 | Banco relacional para testes |
| MongoDB | 7.0 | 27017 | Banco NoSQL para testes |
| Redis | 7-alpine | 6379 | Cache em memória |
| RabbitMQ | 3-management | 5672/15672 | Message broker |

## 📋 Pré-requisitos

- Docker >= 20.10.0
- Docker Compose >= 2.0.0
- Go >= 1.23.0 (para desenvolvimento)
- curl (para health checks)
- 8GB RAM disponível (recomendado)
- 10GB espaço em disco

## 🚀 Instalação e Uso

### Início Rápido

```bash
# Clone o repositório (se ainda não fez)
git clone <repository-url>
cd nexs-lib/observability/infraestructure

# Validar ambiente
make validate-env

# Iniciar toda a stack
make infra-up

# Verificar status
make infra-status

# Ver URLs dos serviços
make infra-urls
```

### Comandos do Makefile

```bash
# Gerenciamento da infraestrutura
make infra-up           # Iniciar todos os serviços
make infra-down         # Parar todos os serviços
make infra-restart      # Reiniciar todos os serviços
make infra-status       # Status dos serviços
make infra-logs         # Ver logs (SERVICE=nome opcional)
make infra-clean        # Limpar volumes
make infra-reset        # Reset completo

# Grupos específicos
make infra-up GROUP=tracer     # Apenas serviços de tracing
make infra-up GROUP=logger     # Apenas serviços de logging
make infra-up GROUP=metrics    # Apenas serviços de métricas
make infra-up GROUP=databases  # Apenas bancos de dados

# Desenvolvimento
make dev-setup          # Setup completo para desenvolvimento
make dev-tracer         # Ambiente para tracer
make dev-logger         # Ambiente para logger

# Testes
make test-integration   # Testes de integração
make test-examples      # Testar exemplos
make perf-test          # Testes de performance

# Monitoramento
make monitor-traces     # Abrir Jaeger UI
make monitor-metrics    # Abrir Grafana
make monitor-logs       # Abrir Kibana
```

### Script de Gerenciamento

```bash
# Usando o script diretamente
./manage.sh up          # Iniciar tudo
./manage.sh down        # Parar tudo
./manage.sh status      # Status
./manage.sh health      # Health check
./manage.sh urls        # Mostrar URLs
./manage.sh logs jaeger # Logs do Jaeger
```

## ⚙️ Configuração

### Variáveis de Ambiente

Principais variáveis configuráveis no `docker-compose.yml`:

```yaml
# Elasticsearch
ELASTIC_PASSWORD=nexs123
ES_JAVA_OPTS=-Xms1g -Xmx1g

# PostgreSQL
POSTGRES_PASSWORD=nexs123

# MongoDB
MONGO_INITDB_ROOT_PASSWORD=nexs123

# RabbitMQ
RABBITMQ_DEFAULT_PASS=nexs123
```

### Arquivos de Configuração

- `otel-collector-config.yaml`: Configuração do OpenTelemetry Collector
- `tempo.yaml`: Configuração do Tempo
- `prometheus.yml`: Configuração do Prometheus
- `logstash.conf`: Pipeline do Logstash
- `fluentd.conf`: Configuração do Fluentd
- `grafana/`: Dashboards e datasources do Grafana

## 🌐 URLs dos Serviços

### UIs Web
- **Jaeger**: http://localhost:16686
- **Grafana**: http://localhost:3000 (admin/nexs123)
- **Kibana**: http://localhost:5601
- **Prometheus**: http://localhost:9090
- **RabbitMQ**: http://localhost:15672 (guest/nexs123)

### APIs e Endpoints
- **Elasticsearch**: http://localhost:9200
- **PostgreSQL**: localhost:5432
- **MongoDB**: localhost:27017
- **Redis**: localhost:6379
- **OTEL gRPC**: localhost:4317
- **OTEL HTTP**: localhost:4318

## 📚 Exemplos de Uso

### Teste Manual com curl

```bash
# Verificar Elasticsearch
curl -u elastic:nexs123 http://localhost:9200/_cluster/health

# Verificar Prometheus
curl http://localhost:9090/api/v1/targets

# Verificar OTEL Collector
curl http://localhost:13133/

# Verificar Jaeger
curl http://localhost:16686/api/services
```

### Desenvolvimento com Tracer

```bash
# Iniciar apenas serviços necessários para tracer
make dev-tracer

# Executar exemplo
cd ../tracer/examples/datadog
go run main.go

# Verificar traces no Jaeger
open http://localhost:16686
```

### Desenvolvimento com Logger

```bash
# Iniciar apenas serviços necessários para logger
make dev-logger

# Executar testes com logs
cd ../logger
go test -v

# Verificar logs no Kibana
open http://localhost:5601
```

## 🔍 Solução de Problemas

### Problemas Comuns

#### Erro de Memória
```bash
# Verificar uso de memória
docker stats

# Aumentar limite de memória virtual para Elasticsearch
sudo sysctl -w vm.max_map_count=262144
```

#### Serviços não inicializam
```bash
# Verificar logs
make infra-logs SERVICE=elasticsearch

# Verificar portas ocupadas
sudo netstat -tlnp | grep :9200

# Limpar volumes corrompidos
make infra-clean
```

#### Health checks falhando
```bash
# Verificar conectividade
make infra-health

# Aguardar mais tempo para inicialização
sleep 30 && make infra-status
```

### Logs e Debug

```bash
# Ver logs de todos os serviços
make infra-logs

# Logs de um serviço específico
make infra-logs SERVICE=jaeger

# Debug do Docker Compose
docker-compose -f docker-compose.yml config

# Verificar recursos do sistema
docker system df
docker system events
```

### Reset Completo

```bash
# Reset completo em caso de problemas
make infra-down
make infra-clean
docker system prune -f
make infra-up
```

## 📊 Dashboards Pré-configurados

### Grafana Dashboards

1. **Nexs Tracer Overview**: Métricas gerais de tracing
2. **Nexs Logger Overview**: Métricas de logging
3. **infraestructure Health**: Status da infraestrutura
4. **Performance Metrics**: Métricas de performance

### Kibana Index Patterns

- `nexs-logs-*`: Logs da aplicação
- `nexs-traces-*`: Traces convertidos para logs
- `infraestructure-*`: Logs da infraestrutura

## 🔄 CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.23'
      
      - name: Start infraestructure
        run: |
          cd observability/infraestructure
          make infra-up
          sleep 30
      
      - name: Run Tests
        run: make test-integration
      
      - name: Cleanup
        run: |
          cd observability/infraestructure
          make infra-down
```

## 🤝 Contribuição

### Adicionando Novos Serviços

1. Editar `docker-compose.yml`
2. Adicionar configurações necessárias
3. Atualizar `manage.sh` se necessário
4. Documentar no README
5. Adicionar ao Makefile

### Estrutura de Arquivos

```
infraestructure/
├── docker-compose.yml         # Definição dos serviços
├── manage.sh                  # Script de gerenciamento
├── Makefile                   # Comandos de automação
├── README.md                  # Esta documentação
├── configs/                   # Configurações dos serviços
│   ├── otel-collector-config.yaml
│   ├── tempo.yaml
│   ├── prometheus.yml
│   ├── logstash.conf
│   └── fluentd.conf
├── grafana/                   # Configurações do Grafana
│   ├── provisioning/
│   └── dashboards/
└── init/                      # Scripts de inicialização
    ├── postgres/
    └── mongodb/
```

## 📝 Licença

Este projeto está licenciado sob a mesma licença do projeto principal Nexs Library.

## 🆘 Suporte

Para problemas ou dúvidas:
1. Verificar esta documentação
2. Verificar logs: `make infra-logs`
3. Abrir issue no repositório
4. Consultar documentação oficial dos serviços

---

**Nota**: Esta infraestrutura é destinada para desenvolvimento e testes. Para produção, consulte as melhores práticas de cada serviço individual.
