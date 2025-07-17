# Multi-tenant PostgreSQL Provider Example

Este exemplo demonstra implementações avançadas de multi-tenancy com PostgreSQL, incluindo isolamento por schema, por database, Row Level Security (RLS), e operações cross-tenant.

## 📋 Funcionalidades Demonstradas

- ✅ **Schema-based Multi-tenancy**: Isolamento usando schemas PostgreSQL
- ✅ **Database-based Multi-tenancy**: Isolamento usando bancos separados
- ✅ **Row Level Security (RLS)**: Isolamento a nível de linha
- ✅ **Tenant Management**: Provisionamento e gestão de tenants
- ✅ **Cross-tenant Operations**: Relatórios e operações agregadas
- ✅ **Tenant Isolation**: Segurança e isolamento de dados
- ✅ **Backup & Migration**: Operações de backup por tenant

## 🚀 Pré-requisitos

1. **PostgreSQL Database**:
   ```bash
   # Usando Docker - banco principal
   docker run --name postgres-multitenant \
     -e POSTGRES_USER=user \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=testdb \
     -p 5432:5432 -d postgres:15

   # Criar bancos adicionais para database-based tenancy
   docker exec -it postgres-multitenant psql -U user -d testdb -c "CREATE DATABASE acme_db;"
   docker exec -it postgres-multitenant psql -U user -d testdb -c "CREATE DATABASE globex_db;"
   docker exec -it postgres-multitenant psql -U user -d testdb -c "CREATE DATABASE initech_db;"
   ```

2. **Dependências Go**:
   ```bash
   go mod tidy
   ```

## ⚙️ Configuração

1. **Atualize as strings de conexão** no arquivo `main.go`:
   ```go
   cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")
   ```

2. **Habilite multi-tenancy**:
   ```go
   postgresql.WithMultiTenant(true)
   ```

## 🏃‍♂️ Executando o Exemplo

```bash
# Na pasta do exemplo
cd examples/multitenant

# Executar o exemplo
go run main.go
```

## 📊 Saída Esperada

```
=== Schema-based Multi-tenancy Example ===
🏢 Creating tenant schemas...
  Creating schema for ACME Corporation (tenant_001)...
    ✅ Successfully set up tenant: ACME Corporation
  Creating schema for Globex Corporation (tenant_002)...
    ✅ Successfully set up tenant: Globex Corporation
  Creating schema for Initech (tenant_003)...
    ✅ Successfully set up tenant: Initech

📊 Querying data from different tenant schemas...

  📋 Data for ACME Corporation (schema: company_acme):
    Users:
      - Jane Smith (ACME Corporation) (jane@company_acme.com)
      - John Doe (ACME Corporation) (john@company_acme.com)
    Orders:
      - John Doe (ACME Corporation): $100.50 (completed)
      - Jane Smith (ACME Corporation): $75.25 (pending)

  📋 Data for Globex Corporation (schema: company_globex):
    Users:
      - Jane Smith (Globex Corporation) (jane@company_globex.com)
      - John Doe (Globex Corporation) (john@company_globex.com)
    Orders:
      - John Doe (Globex Corporation): $100.50 (completed)
      - Jane Smith (Globex Corporation): $75.25 (pending)

  📋 Data for Initech (schema: company_initech):
    Users:
      - Jane Smith (Initech) (jane@company_initech.com)
      - John Doe (Initech) (john@company_initech.com)
    Orders:
      - John Doe (Initech): $100.50 (completed)
      - Jane Smith (Initech): $75.25 (pending)

=== Database-based Multi-tenancy Example ===
🏢 Registering tenants...
  ✅ Registered tenant: acme_corp (DB: acme_db, Region: us-east-1, Tier: premium)
  ✅ Registered tenant: globex_corp (DB: globex_db, Region: us-west-2, Tier: standard)
  ✅ Registered tenant: initech (DB: initech_db, Region: eu-west-1, Tier: basic)

📊 Performing operations on tenant databases...

  🏢 Working with tenant: acme_corp
    ✅ Tenant Info: ID=acme_corp, Region=us-east-1, Tier=premium, Created=2024-01-15 14:30:25
    ✅ Operations completed for tenant acme_corp

  🏢 Working with tenant: globex_corp
    ✅ Tenant Info: ID=globex_corp, Region=us-west-2, Tier=standard, Created=2024-01-15 14:30:25
    ✅ Operations completed for tenant globex_corp

  🏢 Working with tenant: initech
    ✅ Tenant Info: ID=initech, Region=eu-west-1, Tier=basic, Created=2024-01-15 14:30:25
    ✅ Operations completed for tenant initech

=== Tenant Isolation and Security Example ===
🔒 Setting up Row Level Security...
  ✅ Enabled Row Level Security on shared_documents
  ✅ Created tenant isolation policy

📝 Inserting sample documents for different tenants...
  ✅ Inserted document for acme: ACME Company Policy
  ✅ Inserted document for acme: ACME Product Specs
  ✅ Inserted document for globex: Globex Strategy 2024
  ✅ Inserted document for globex: Globex Financial Report
  ✅ Inserted document for initech: Initech TPS Reports

🔍 Demonstrating tenant isolation...

  🏢 Querying as tenant: acme
    Documents accessible to acme:
      - ACME Company Policy
      - ACME Product Specs
    ✅ Found 2 documents for tenant acme

  🏢 Querying as tenant: globex
    Documents accessible to globex:
      - Globex Strategy 2024
      - Globex Financial Report
    ✅ Found 2 documents for tenant globex

  🏢 Querying as tenant: initech
    Documents accessible to initech:
      - Initech TPS Reports
    ✅ Found 1 documents for tenant initech

=== Tenant Management Operations Example ===
🔄 Demonstrating tenant lifecycle management...
  📋 Provisioning new tenant: newcorp
    ✅ Successfully provisioned tenant: newcorp
  🏗️  Initializing database schema for tenant: newcorp
    ✅ Database schema initialized for tenant: newcorp
  ⚙️  Managing tenant configuration...
    Current configuration for newcorp:
      max_users: 10 (updated: 2024-01-15 14:30:25)
      region: us-central-1 (updated: 2024-01-15 14:30:25)
      storage_limit: 1GB (updated: 2024-01-15 14:30:25)
      tier: standard (updated: 2024-01-15 14:30:26)
      upgrade_date: 2024-01-15 (updated: 2024-01-15 14:30:26)
    ✅ Tenant configuration managed successfully
  📊 Collecting tenant metrics...
    📈 Metrics for tenant newcorp:
      - Schema: newcorp_schema
      - Tables: 1
      - Region: us-central-1
      - Tier: trial
    ✅ Metrics collected successfully

=== Cross-tenant Operations Example ===
📊 Demonstrating cross-tenant reporting...
  ✅ Created global tenant summary view
  📈 Global Tenant Summary:
    acme: 2 users, $100.50 revenue
    globex: 2 users, $100.50 revenue
    initech: 2 users, $100.50 revenue
  📊 Totals: 6 users across all tenants, $301.50 total revenue

💾 Demonstrating tenant backup operations...
  ✅ Created tenant backup tracking table
    ✅ Backup recorded for tenant acme: /backups/acme_20240115_143026.sql
    ✅ Backup recorded for tenant globex: /backups/globex_20240115_143026.sql
    ✅ Backup recorded for tenant initech: /backups/initech_20240115_143026.sql
  📋 Recent Backup Operations:
    2024-01-15 14:30:26 [initech]: full backup to /backups/initech_20240115_143026.sql (completed)
    2024-01-15 14:30:26 [globex]: full backup to /backups/globex_20240115_143026.sql (completed)
    2024-01-15 14:30:26 [acme]: full backup to /backups/acme_20240115_143026.sql (completed)

🔒 Closing pool for tenant: acme_corp
🔒 Closing pool for tenant: globex_corp
🔒 Closing pool for tenant: initech
🔒 Closing pool for tenant: newcorp
Multi-tenant examples completed!
```

## 📝 Conceitos Demonstrados

### 1. Schema-based Multi-tenancy
```go
// Criar schema por tenant
_, err := conn.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", tenant.schema))

// Definir search path para isolamento
_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", tenant.schema))
```

**Vantagens:**
- Uma única database
- Fácil backup e manutenção
- Queries cross-tenant simples

**Desvantagens:**
- Isolamento limitado
- Risco de vazamento de dados

### 2. Database-based Multi-tenancy
```go
type TenantConfig struct {
    TenantID    string
    DatabaseURL string
    Settings    map[string]interface{}
}

// Pool separado por tenant
pool, err := tenantManager.GetPoolForTenant(ctx, tenantID)
```

**Vantagens:**
- Isolamento total
- Backup independente
- Escalabilidade horizontal

**Desvantagens:**
- Maior complexidade
- Recursos duplicados
- Queries cross-tenant complexas

### 3. Row Level Security (RLS)
```go
// Habilitar RLS
_, err = conn.Exec(ctx, "ALTER TABLE shared_documents ENABLE ROW LEVEL SECURITY")

// Criar política de isolamento
_, err = conn.Exec(ctx, `
    CREATE POLICY tenant_isolation ON shared_documents
    FOR ALL TO PUBLIC
    USING (tenant_id = current_setting('app.current_tenant', true))
`)

// Definir contexto do tenant
_, err := conn.Exec(ctx, fmt.Sprintf("SET app.current_tenant = '%s'", tenantID))
```

**Vantagens:**
- Isolamento a nível de linha
- Flexibilidade máxima
- Uma tabela para todos os tenants

**Desvantagens:**
- Complexidade de configuração
- Performance pode ser impactada
- Requer cuidado com índices

### 4. Tenant Manager
```go
type TenantManager struct {
    provider interfaces.PostgreSQLProvider
    tenants  map[string]*TenantConfig
    pools    map[string]interfaces.IPool
}

func (tm *TenantManager) GetPoolForTenant(ctx context.Context, tenantID string) (interfaces.IPool, error) {
    // Lazy loading de pools por tenant
}
```

### 5. Cross-tenant Operations
```go
// View agregada cross-tenant
CREATE OR REPLACE VIEW global_tenant_summary AS
SELECT 'acme' as tenant_id, COUNT(*) as user_count FROM company_acme.users
UNION ALL
SELECT 'globex' as tenant_id, COUNT(*) as user_count FROM company_globex.users
```

## 🏗️ Padrões de Arquitetura

### 1. Schema-based Pattern
```
Database: app_db
├── Schema: tenant_001
│   ├── users
│   ├── orders
│   └── products
├── Schema: tenant_002
│   ├── users
│   ├── orders
│   └── products
└── Schema: public
    ├── tenant_configs
    └── global_settings
```

### 2. Database-based Pattern
```
┌─ Database: tenant_001_db
│  ├── users
│  ├── orders
│  └── products
├─ Database: tenant_002_db
│  ├── users
│  ├── orders
│  └── products
└─ Database: shared_db
   ├── tenant_configs
   └── global_metrics
```

### 3. Hybrid Pattern
```
Database: app_db
├── Schema: tenant_001 (small tenants)
├── Schema: tenant_002 (small tenants)
└── External DB: enterprise_tenant_db (large tenant)
```

## 🔧 Configuração de Tenants

### Tenant Configuration
```go
type TenantConfig struct {
    TenantID    string                 // Identificador único
    SchemaName  string                 // Nome do schema (schema-based)
    DatabaseURL string                 // URL do banco (database-based)
    Settings    map[string]interface{} // Configurações específicas
}
```

### Configurações Típicas
```go
Settings: map[string]interface{}{
    "region":         "us-east-1",      // Região geográfica
    "tier":          "premium",         // Nível de serviço
    "max_users":     1000,              // Limite de usuários
    "storage_limit": "10GB",            // Limite de armazenamento
    "features":      []string{"api", "analytics"}, // Features habilitadas
}
```

## 📊 Monitoramento Multi-tenant

### Métricas por Tenant
- **Storage Usage**: Uso de armazenamento por tenant
- **Query Performance**: Performance de queries por tenant
- **Connection Usage**: Uso de conexões por tenant
- **Error Rates**: Taxa de erro por tenant

### Alertas
- Limite de armazenamento atingido
- Performance degradada
- Erro rate alta
- Uso excessivo de conexões

## 🔒 Segurança Multi-tenant

### Best Practices
1. **Sempre validar tenant_id** nos parâmetros de entrada
2. **Usar RLS quando possível** para isolamento automático
3. **Criptografar dados sensíveis** por tenant
4. **Auditar operações cross-tenant**
5. **Implementar rate limiting** por tenant

### Validação de Tenant
```go
func validateTenantAccess(userTenantID, requestedTenantID string) error {
    if userTenantID != requestedTenantID {
        return errors.New("unauthorized tenant access")
    }
    return nil
}
```

## 🚀 Performance Multi-tenant

### Schema-based Optimization
- Índices específicos por schema
- Particionamento por tenant_id
- Connection pooling compartilhado

### Database-based Optimization
- Pools dedicados por tenant
- Scaling independente
- Backup/restore otimizado

## 📚 Próximos Passos

Após dominar multi-tenancy, explore:

1. **[Performance Example](../performance/)**: Otimização de performance
2. **[Production Example](../production/)**: Configuração para produção
3. **[Monitoring Example](../monitoring/)**: Monitoramento avançado

## 🔍 Debugging Multi-tenant

Para debug detalhado:
```bash
export LOG_LEVEL=debug
export MULTITENANT_DEBUG=true
export RLS_DEBUG=true
export TENANT_METRICS=true
```

Logs incluirão:
- Mudanças de contexto de tenant
- Execução de políticas RLS
- Métricas por tenant
- Operações cross-tenant
