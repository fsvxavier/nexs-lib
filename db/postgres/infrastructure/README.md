# NEXS-LIB Infrastructure

![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)

Infraestrutura Docker completa para desenvolvimento, testes e exemplos da biblioteca NEXS-LIB PostgreSQL.

## 🏗️ Arquitetura da Infraestrutura

```
┌─────────────────────────────────────────────────────────────┐
│                    NEXS-LIB Infrastructure                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   PostgreSQL    │    │   PostgreSQL    │                │
│  │    Primary      │◄───┤    Replica 1    │                │
│  │   (port 5432)   │    │   (port 5433)   │                │
│  └─────────────────┘    └─────────────────┘                │
│           │                                                 │
│           │              ┌─────────────────┐                │
│           └──────────────┤   PostgreSQL    │                │
│                          │    Replica 2    │                │
│                          │   (port 5434)   │                │
│                          └─────────────────┘                │
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │     Redis       │    │     PgAdmin     │                │
│  │   (port 6379)   │    │   (port 8080)   │                │
│  └─────────────────┘    └─────────────────┘                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Estrutura da Infraestrutura

```
infrastructure/
├── README.md                         # Este arquivo
├── manage.sh                         # Script de gerenciamento
├── docker/                           # Configurações Docker
│   ├── docker-compose.yml            # Orquestração dos serviços
│   └── postgres/                     # Configurações PostgreSQL
│       ├── primary/                  # Configurações do primary
│       │   ├── postgresql.conf       # Configuração do PostgreSQL
│       │   └── pg_hba.conf           # Autenticação
│       └── scripts/                  # Scripts de inicialização
│           └── super-simple-replica.sh # Script de setup das replicas
└── database/                         # Scripts de banco de dados
    └── init/                         # Scripts de inicialização
        ├── 01_init_replication.sql   # Configuração de replicação
        ├── 02_schema.sql             # Schema principal
        ├── 03_sample_data.sql        # Dados de exemplo
        └── 04_examples_setup.sql     # Setup específico para exemplos
```

## 🎯 Banco de Dados - Estrutura para Exemplos

O banco de dados foi estruturado especificamente para suportar todos os exemplos da biblioteca NEXS-LIB:

### 📊 Tabelas Principais

#### **Operações Básicas e Batch**
- `products` - Produtos para exemplos de batch, transações e operações básicas
- `accounts` - Contas para exemplos de transação e transferências

#### **Operações COPY**
- `copy_test` - Tabela otimizada para operações de COPY FROM/TO
- Inclui diversos tipos de dados (texto, numérico, data, boolean)

#### **Multi-Tenancy**
- `tenants` - Gerenciamento de inquilinos/tenants
- `shared_users` - Usuários compartilhados para row-level security
- Schemas separados: `tenant_empresa_a`, `tenant_empresa_b`, `tenant_empresa_c`

#### **LISTEN/NOTIFY**
- `chat_messages` - Sistema de chat em tempo real
- `monitored_table` - Tabela monitorada para notificações de mudanças

#### **Réplicas e Performance**
- `replica_test` - Testes de replicação
- `performance_test` - Dados para testes de performance
- `audit_log` - Log de auditoria para hooks

### 🔧 Funcionalidades Especiais

#### **Funções Utilitárias**
```sql
-- Gerar dados de teste para operações batch
SELECT generate_batch_test_data(1000);

-- Gerar dados para operações COPY
SELECT generate_copy_test_data(5000);

-- Simular transações entre contas
SELECT simulate_account_transactions(100);

-- Popular canais de chat
SELECT populate_chat_channels();

-- Configurar dados de teste para tenants
SELECT setup_tenant_test_data();

-- Resetar todos os dados de exemplo
SELECT reset_example_data();

-- Obter estatísticas das tabelas
SELECT * FROM get_example_table_stats();
```

#### **Views para Análise**
```sql
-- Resumo de operações batch
SELECT * FROM batch_operation_summary;

-- Resumo de multi-tenancy
SELECT * FROM multi_tenant_summary;

-- Métricas de performance
SELECT * FROM performance_metrics;

-- Estatísticas de produtos
SELECT * FROM product_stats;

-- Resumo de contas
SELECT * FROM account_summary;

-- Estatísticas de tenants
SELECT * FROM tenant_stats;
```

#### **Triggers e Automação**
- **Audit Triggers**: Registro automático de mudanças
- **LISTEN/NOTIFY**: Notificações em tempo real
- **Updated_at**: Atualização automática de timestamps
- **Row Level Security**: Isolamento por tenant

### 📋 Dados de Exemplo

#### **Produtos (20 registros)**
```sql
-- Exemplos: Laptop Gaming, Mouse Wireless, Keyboard Mechanical, etc.
-- Categorias: Electronics, Office, Furniture
-- Preços: Variados de $9.99 a $1299.99
```

#### **Contas (10 registros)**
```sql
-- Exemplos: Alice Johnson ($1000), Bob Smith ($500), etc.
-- Balances variados para testes de transação
```

#### **Dados COPY (15 registros base)**
```sql
-- Funcionários com departamentos, salários, datas de contratação
-- Departamentos: Engineering, Marketing, Sales, HR, Finance
```

#### **Multi-Tenancy (5 tenants)**
```sql
-- Empresa A, B, C + Test Company + Demo Corp
-- Usuários em schemas separados e tabela compartilhada
```

#### **Chat (10+ mensagens)**
```sql
-- Canais: general, tech, support, random, notifications
-- Usuários: admin, user1, developer1, support1, etc.
```
    └── init/                         # Scripts de inicialização
        ├── 01_init_replication.sql   # Configuração de replicação
        ├── 02_schema.sql             # Esquema do banco
        └── 03_sample_data.sql        # Dados de exemplo
```

## 🐳 Serviços Docker

### PostgreSQL Primary
- **Imagem**: `postgres:15`
- **Porta**: `5432`
- **Função**: Banco principal (leitura/escrita)
- **Configuração**: WAL ativado para replicação
- **Health Check**: Integrado

### PostgreSQL Replica 1 & 2 ✅ OTIMIZADO
- **Imagem**: `postgres:15`
- **Portas**: `5433` (Replica 1) e `5434` (Replica 2)
- **Função**: Réplicas de leitura (somente leitura)
- **Replicação**: Streaming replication automática
- **Dependência**: postgres-primary
- **Inicialização**: Script `super-simple-replica.sh` otimizado
- **Configuração**: Automática via `pg_basebackup -R`
- **Health Check**: 120s start_period para inicialização robusta
- **Status**: ✅ Sem warnings de "not ready yet"

### Redis
- **Imagem**: `redis:7-alpine`
- **Porta**: `6379`
- **Função**: Cache e sessões
- **Persistência**: Configurada

### PgAdmin
- **Imagem**: `dpage/pgadmin4:latest`
- **Porta**: `8080`
- **Função**: Interface web para administração
- **Credenciais**: admin@nexs.com / admin123

## 🚀 Início Rápido

### Pré-requisitos
```bash
# Verificar se Docker está instalado
docker --version

# Verificar se docker-compose está instalado
docker-compose --version

# Verificar se Go está instalado
go version
```

### Iniciar Infraestrutura
```bash
# Navegar para o diretório do projeto
cd /path/to/nexs-lib

# Iniciar todos os serviços
sudo ./infrastructure/manage.sh start

# Verificar status
sudo ./infrastructure/manage.sh status
```

## 📋 Comandos do Script de Gerenciamento

### Comandos Básicos

```bash
# Iniciar infraestrutura
sudo ./infrastructure/manage.sh start

# Parar infraestrutura
sudo ./infrastructure/manage.sh stop

# Reiniciar infraestrutura
sudo ./infrastructure/manage.sh restart

# Verificar status
sudo ./infrastructure/manage.sh status

# Ajuda
sudo ./infrastructure/manage.sh help
```

### Comandos de Desenvolvimento

```bash
# Executar testes
sudo ./infrastructure/manage.sh test

# Executar exemplos
sudo ./infrastructure/manage.sh example basic
sudo ./infrastructure/manage.sh example replicas
sudo ./infrastructure/manage.sh example advanced
sudo ./infrastructure/manage.sh example pool

# Ver logs
sudo ./infrastructure/manage.sh logs
sudo ./infrastructure/manage.sh logs postgres-primary
sudo ./infrastructure/manage.sh logs postgres-replica1
```

### Comandos de Manutenção

```bash
# Resetar banco de dados (cuidado!)
sudo ./infrastructure/manage.sh reset

# Ver logs de um serviço específico
sudo ./infrastructure/manage.sh logs [serviço]
```

## 🔧 Configurações

### Variáveis de Ambiente

Após iniciar a infraestrutura, as seguintes variáveis são automaticamente configuradas:

```bash
# Banco principal
NEXS_DB_HOST=localhost
NEXS_DB_PORT=5432
NEXS_DB_NAME=nexs_testdb
NEXS_DB_USER=nexs_user
NEXS_DB_PASSWORD=nexs_password
NEXS_DB_DSN="postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"

# Réplicas
NEXS_DB_REPLICA1_HOST=localhost
NEXS_DB_REPLICA1_PORT=5433
NEXS_DB_REPLICA2_HOST=localhost
NEXS_DB_REPLICA2_PORT=5434
NEXS_DB_REPLICA_DSN="postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb"
```

### Informações de Conexão

| Serviço | Host | Porta | Usuário | Senha | Banco |
|---------|------|-------|---------|-------|-------|
| Primary | localhost | 5432 | nexs_user | nexs_password | nexs_testdb |
| Replica 1 | localhost | 5433 | nexs_user | nexs_password | nexs_testdb |
| Replica 2 | localhost | 5434 | nexs_user | nexs_password | nexs_testdb |
| Redis | localhost | 6379 | - | - | - |
| PgAdmin | localhost | 8080 | admin@nexs.com | admin123 | - |

### Configuração PostgreSQL

#### Primary (Mestre)
```sql
-- postgresql.conf (Configuração otimizada v2.0.0)
wal_level = replica
max_wal_senders = 10              # ⬆️ Aumentado de 3 para 10
max_replication_slots = 10
synchronous_commit = on
archive_mode = on
archive_command = 'test ! -f /var/lib/postgresql/data/archive/%f && cp %p /var/lib/postgresql/data/archive/%f'
```

#### Replica (Escravos)
```bash
# ✅ Configuração automática via super-simple-replica.sh
# O script utiliza pg_basebackup com flag -R para configuração automática
# Não requer configuração manual de postgresql.conf ou pg_hba.conf

# Comando executado automaticamente:
pg_basebackup -h postgres-primary -D /var/lib/postgresql/data -U replicator -R -W
```

#### Timeouts e Health Checks (v2.0.0)
```yaml
# docker-compose.yml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U nexs_user -d nexs_testdb"]
  interval: 30s
  timeout: 10s
  retries: 5
  start_period: 120s  # ⬆️ Aumentado de 60s para 120s

# manage.sh
POSTGRES_TIMEOUT=60   # ⬆️ Aumentado de 30s para 60s
REPLICA_TIMEOUT=120   # ⬆️ Aumentado de 60s para 120s
```

## 🗄️ Esquema de Banco de Dados

### Tabelas Principais

#### users
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### products
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock_quantity INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### orders
```sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### order_items
```sql
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id),
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL
);
```

### Funcionalidades Avançadas

#### Audit Log
```sql
CREATE TABLE audit_log (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(255) NOT NULL,
    operation VARCHAR(10) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    user_id INTEGER,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Multi-tenancy
```sql
-- Esquemas para diferentes tenants
CREATE SCHEMA tenant_1;
CREATE SCHEMA tenant_2;

-- Tabelas específicas por tenant
CREATE TABLE tenant_1.tenant_data (
    id SERIAL PRIMARY KEY,
    data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Performance Testing
```sql
CREATE TABLE performance_test (
    id SERIAL PRIMARY KEY,
    test_data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🧪 Testes

### Executar Testes Automatizados

```bash
# Executar todos os testes
sudo ./infrastructure/manage.sh test

# Executar testes específicos
cd db/postgres
go test -v -race -timeout 30s ./...

# Executar benchmarks
go test -bench=. -benchmem ./...
```

### Validar Replicação

```bash
# Conectar ao primary
psql -h localhost -p 5432 -U nexs_user -d nexs_testdb

# Inserir dados no primary
INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com');

# Conectar à replica
psql -h localhost -p 5433 -U nexs_user -d nexs_testdb

# Verificar se os dados foram replicados
SELECT * FROM users WHERE email = 'test@example.com';
```

### Teste de Failover

```bash
# Parar primary
sudo docker-compose -f infrastructure/docker/docker-compose.yml stop postgres-primary

# Executar exemplo de replica (deve funcionar)
sudo ./infrastructure/manage.sh example replicas

# Reiniciar primary
sudo docker-compose -f infrastructure/docker/docker-compose.yml start postgres-primary
```

## 📊 Monitoramento

### Logs em Tempo Real

```bash
# Todos os serviços
sudo ./infrastructure/manage.sh logs

# Serviço específico
sudo ./infrastructure/manage.sh logs postgres-primary
sudo ./infrastructure/manage.sh logs postgres-replica1
sudo ./infrastructure/manage.sh logs redis
sudo ./infrastructure/manage.sh logs pgadmin
```

### Métricas de Replicação

```sql
-- No primary: verificar replicas conectadas
SELECT * FROM pg_stat_replication;

-- Na replica: verificar lag de replicação
SELECT * FROM pg_stat_wal_receiver;
```

### Health Checks

```bash
# Verificar se todos os serviços estão saudáveis
sudo docker-compose -f infrastructure/docker/docker-compose.yml ps

# Verificar conectividade do banco
pg_isready -h localhost -p 5432 -U nexs_user -d nexs_testdb
pg_isready -h localhost -p 5433 -U nexs_user -d nexs_testdb
pg_isready -h localhost -p 5434 -U nexs_user -d nexs_testdb
```

## 🔧 Troubleshooting

### Problemas Comuns

#### 1. Docker não está rodando
```bash
# Erro: "Docker is not running or accessible"
# Solução:
sudo systemctl start docker

# Ou adicionar usuário ao grupo docker
sudo usermod -aG docker $USER
# Fazer logout/login
```

#### 2. Porta em uso
```bash
# Erro: "Port 5432 is already in use"
# Solução: verificar processos usando a porta
sudo lsof -i :5432
sudo kill -9 <PID>

# Ou usar portas diferentes no docker-compose.yml
```

#### 3. Replica não sincroniza ✅ CORRIGIDO
```bash
# ✅ Problema resolvido na v2.0.0
# Anteriormente: "[WARNING] Replica 1 is not ready yet, but continuing..."
# Solução implementada:
# - Script simplificado super-simple-replica.sh
# - Timeouts aumentados para 60s/120s
# - max_wal_senders configurado para 10
# - Health checks aprimorados

# Para verificar se as replicas estão funcionando:
sudo ./infrastructure/manage.sh status

# Verificar logs se houver problemas
sudo ./infrastructure/manage.sh logs postgres-replica1
sudo ./infrastructure/manage.sh logs postgres-replica2

# Verificar configuração de replicação
psql -h localhost -p 5432 -U nexs_user -d nexs_testdb -c "SELECT * FROM pg_stat_replication;"
```

#### 4. Banco não inicializa
```bash
# Verificar logs
sudo ./infrastructure/manage.sh logs postgres-primary

# Resetar banco
sudo ./infrastructure/manage.sh reset
```

### Comandos de Debug

```bash
# Verificar containers
sudo docker ps -a

# Verificar networks
sudo docker network ls

# Verificar volumes
sudo docker volume ls

# Logs detalhados
sudo docker-compose -f infrastructure/docker/docker-compose.yml logs --tail=50 postgres-primary

# Conectar ao container
sudo docker exec -it nexs-postgres-primary bash
```

### Limpeza Completa

```bash
# Parar e remover tudo
sudo ./infrastructure/manage.sh stop

# Remover volumes (perde dados!)
sudo docker-compose -f infrastructure/docker/docker-compose.yml down -v

# Remover imagens
sudo docker rmi postgres:15 redis:7-alpine dpage/pgadmin4:latest

# Limpeza geral do Docker
sudo docker system prune -a
```

## 🛡️ Segurança

### Credenciais Padrão

⚠️ **Atenção**: As credenciais padrão são para desenvolvimento apenas!

```bash
# PostgreSQL
POSTGRES_USER=nexs_user
POSTGRES_PASSWORD=nexs_password
POSTGRES_REPLICATION_USER=replicator
POSTGRES_REPLICATION_PASSWORD=replicator_password

# PgAdmin
PGADMIN_DEFAULT_EMAIL=admin@nexs.com
PGADMIN_DEFAULT_PASSWORD=admin123
```

### Para Produção

1. **Alterar todas as senhas**
2. **Usar variáveis de ambiente**
3. **Configurar SSL/TLS**
4. **Implementar firewall**
5. **Usar secrets do Docker**

### Configuração SSL

```yaml
# docker-compose.yml
environment:
  POSTGRES_SSL_MODE: require
  POSTGRES_SSL_CERT: /path/to/cert.pem
  POSTGRES_SSL_KEY: /path/to/key.pem
  POSTGRES_SSL_CA: /path/to/ca.pem
```

## 🚀 Performance

### Configurações Otimizadas

#### PostgreSQL
```sql
-- postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB
max_connections = 200
```

#### Redis
```conf
# redis.conf
maxmemory 128mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### Benchmarks

```bash
# Teste de performance PostgreSQL
pgbench -i -s 50 -h localhost -p 5432 -U nexs_user nexs_testdb
pgbench -c 10 -j 2 -t 1000 -h localhost -p 5432 -U nexs_user nexs_testdb

# Teste de performance Redis
redis-benchmark -h localhost -p 6379 -t set,get -n 100000
```

## 🔄 Backup e Restore

### Backup Automático

```bash
# Criar backup
sudo docker exec nexs-postgres-primary pg_dump -U nexs_user nexs_testdb > backup.sql

# Backup com compressão
sudo docker exec nexs-postgres-primary pg_dump -U nexs_user -Fc nexs_testdb > backup.dump
```

### Restore

```bash
# Restaurar de SQL
cat backup.sql | sudo docker exec -i nexs-postgres-primary psql -U nexs_user nexs_testdb

# Restaurar de dump
sudo docker exec -i nexs-postgres-primary pg_restore -U nexs_user -d nexs_testdb backup.dump
```

## 📈 Próximos Passos

### ✅ Melhorias Recentes (v2.0.0)

1. **Replicação PostgreSQL Otimizada**
   - ✅ Correção de warnings de replica não responsiva
   - ✅ Script de inicialização simplificado (`super-simple-replica.sh`)
   - ✅ Timeouts ajustados para 60s/120s
   - ✅ Configuração `max_wal_senders = 10` para melhor performance
   - ✅ Health checks aprimorados com período de inicialização estendido

2. **Infraestrutura Limpa e Otimizada**
   - ✅ Remoção de scripts obsoletos de replica
   - ✅ Configuração simplificada sem arquivos desnecessários
   - ✅ Estrutura de diretórios otimizada
   - ✅ Documentação atualizada

3. **Monitoramento Aprimorado**
   - ✅ Logs detalhados no script de gerenciamento
   - ✅ Verificação de status de replicas melhorada
   - ✅ Mensagens de erro mais informativas

### Melhorias Planejadas

1. **Monitoring Stack**
   - Prometheus + Grafana
   - Métricas de performance
   - Alertas automáticos

2. **Backup Automatizado**
   - Backups regulares
   - Retenção configurável
   - Restore automático

3. **Load Balancer**
   - HAProxy/Nginx
   - Balanceamento de replicas
   - Health checks

4. **Segurança Avançada**
   - Certificados SSL
   - Autenticação via LDAP
   - Audit logging

5. **Escalabilidade**
   - Kubernetes support
   - Auto-scaling
   - Sharding

## 📞 Suporte

### Documentação
- [Docker Compose](https://docs.docker.com/compose/)
- [PostgreSQL](https://www.postgresql.org/docs/)
- [Redis](https://redis.io/documentation)
- [PgAdmin](https://www.pgadmin.org/docs/)

### Contato
- **Issues**: [GitHub Issues](https://github.com/fsvxavier/nexs-lib/issues)
- **Maintainer**: @fsvxavier
- **Email**: Para suporte técnico, abra uma issue no GitHub

---

**Versão**: 2.0.0 ✅ ATUAL  
**Última Atualização**: 17 de julho de 2025  
**Compatibilidade**: Docker 20+, PostgreSQL 15+, Redis 7+  
**Status**: ✅ Replicação PostgreSQL funcionando sem warnings  
**Replicas**: ✅ Ambas as replicas (5433 e 5434) operacionais  

### 🎯 Melhorias v2.0.0
- ✅ Correção de warnings de replica
- ✅ Scripts de inicialização otimizados
- ✅ Timeouts ajustados
- ✅ Configuração simplificada
- ✅ Health checks aprimorados
- ✅ Documentação atualizada

🚀 **Pronto para começar?** Execute `./infrastructure/manage.sh start` e explore a infraestrutura!
