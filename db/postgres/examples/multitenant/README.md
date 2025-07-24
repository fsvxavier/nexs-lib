# Exemplo de Multi-Tenancy

Este exemplo demonstra diferentes estratégias de multi-tenancy em PostgreSQL, permitindo que uma aplicação sirva múltiplos inquilinos (tenants) de forma segura e isolada.

## Funcionalidades Demonstradas

### 1. Schema-based Multi-Tenancy
- Um schema por tenant
- Isolamento completo de dados
- Mudança de contexto via search_path

### 2. Row-level Multi-Tenancy
- Coluna tenant_id nas tabelas
- Filtros automáticos por tenant
- Row Level Security (RLS)

### 3. Database-level Multi-Tenancy
- Conceitos e configuração
- Roteamento por tenant
- Vantagens e desvantagens

### 4. Tenant Isolation & Security
- Validação de tenant
- Sanitização de dados
- Auditoria de acesso

### 5. Tenant Management
- Listagem e estatísticas
- Operações de manutenção
- Monitoramento

## Estratégias de Multi-Tenancy

### Schema-based
```sql
-- Cada tenant tem seu próprio schema
CREATE SCHEMA tenant_empresa_a;
CREATE SCHEMA tenant_empresa_b;

-- Mudança de contexto
SET search_path TO tenant_empresa_a, public;
```

**Vantagens:**
- ✅ Bom isolamento de dados
- ✅ Backup individual por tenant
- ✅ Facilita migração de schema
- ✅ Queries simples

**Desvantagens:**
- ❌ Limite de schemas por banco
- ❌ Overhead de manutenção
- ❌ Dificuldade em relatórios cross-tenant

### Row-level
```sql
-- Coluna tenant_id nas tabelas
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    name VARCHAR(100),
    email VARCHAR(100)
);

-- Filtros automáticos
SELECT * FROM users WHERE tenant_id = $1;
```

**Vantagens:**
- ✅ Eficiente para muitos tenants
- ✅ Facilita relatórios cross-tenant
- ✅ Menor overhead de manutenção
- ✅ Escalabilidade horizontal

**Desvantagens:**
- ❌ Risco de vazamento de dados
- ❌ Complexidade de queries
- ❌ Backup/restore mais complexo

### Database-level
```sql
-- Bancos separados por tenant
nexs_tenant_empresa_a
nexs_tenant_empresa_b
nexs_tenant_empresa_c
```

**Vantagens:**
- ✅ Isolamento máximo
- ✅ Backup/restore individual
- ✅ Escalabilidade horizontal
- ✅ Configuração por tenant

**Desvantagens:**
- ❌ Alto custo de recursos
- ❌ Complexidade de deployment
- ❌ Overhead de conexões
- ❌ Relatórios cross-tenant impossíveis

## Como Executar

```bash
# Certifique-se de que o PostgreSQL está rodando
cd multitenant/
go run main.go
```

## Exemplo de Saída

```
=== Exemplo de Multi-Tenancy ===

1. Conectando ao banco...
2. Configurando estrutura multi-tenant...
   ✅ Estrutura multi-tenant configurada

3. Exemplo: Schema-based multi-tenancy...
   Criando schemas para cada tenant...
   ✅ Schema criado para Empresa A: tenant_empresa_a
   ✅ Schema criado para Empresa B: tenant_empresa_b
   ✅ Schema criado para Empresa C: tenant_empresa_c
   
   Inserindo dados específicos para cada tenant...
   ✅ Dados inseridos para schema tenant_empresa_a
   ✅ Dados inseridos para schema tenant_empresa_b
   ✅ Dados inseridos para schema tenant_empresa_c
   
   Demonstrando isolamento de dados...
   📊 Dados do tenant Empresa A:
     1. João Silva (joao@empresaa.com)
     2. Maria Santos (maria@empresaa.com)
   
   📊 Dados do tenant Empresa B:
     1. Ana Costa (ana@empresab.com)
     2. Pedro Oliveira (pedro@empresab.com)

4. Exemplo: Row-level multi-tenancy...
   Inserindo dados com tenant_id...
   ✅ Dados inseridos com tenant_id
   
   Demonstrando queries com filtro por tenant...
   📊 Usuários do tenant Empresa A (ID: 1):
     1. João Silva (joao@empresaa.com)
     2. Maria Santos (maria@empresaa.com)
   
   Demonstrando Row Level Security (RLS)...
   🔐 RLS configurado para tenant_id: 1
   📊 Consulta com RLS ativo:
     1. João Silva (joao@empresaa.com)
     2. Maria Santos (maria@empresaa.com)
   📊 Registros visíveis com RLS: 2

5. Exemplo: Database-level multi-tenancy...
   💡 Conceito: Database-level Multi-Tenancy
   📊 Configuração simulada de bancos por tenant:
   - tenant_empresa_a: nexs_tenant_empresa_a
   - tenant_empresa_b: nexs_tenant_empresa_b
   - tenant_empresa_c: nexs_tenant_empresa_c
```

## Implementação de Security

### Row Level Security (RLS)
```sql
-- Habilitar RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Criar política
CREATE POLICY tenant_isolation ON users
FOR ALL
USING (tenant_id = current_setting('app.current_tenant_id')::integer);

-- Configurar tenant atual
SET app.current_tenant_id = '1';
```

### Middleware de Validação
```go
func validateTenant(ctx context.Context, conn postgres.IConn, tenantID int) error {
    var exists bool
    var active bool
    
    err := conn.QueryRow(ctx, 
        "SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1), " +
        "COALESCE((SELECT active FROM tenants WHERE id = $1), false)",
        tenantID).Scan(&exists, &active)
    
    if err != nil {
        return fmt.Errorf("erro ao validar tenant: %w", err)
    }
    
    if !exists {
        return fmt.Errorf("tenant não existe")
    }
    
    if !active {
        return fmt.Errorf("tenant inativo")
    }
    
    return nil
}
```

## Casos de Uso

### 1. SaaS Applications
```go
// Roteamento por subdomínio
func getTenantFromSubdomain(host string) string {
    parts := strings.Split(host, ".")
    if len(parts) > 2 {
        return parts[0] // subdomain
    }
    return "default"
}

// Conexão por tenant
func getConnectionForTenant(tenantID string) (*postgres.Conn, error) {
    dsn := fmt.Sprintf("postgres://user:pass@localhost:5432/tenant_%s", tenantID)
    return postgres.Connect(context.Background(), dsn)
}
```

### 2. Enterprise Applications
```go
// Configuração por tenant
type TenantConfig struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Schema   string `json:"schema"`
    Features map[string]bool `json:"features"`
}

func configureTenantContext(conn postgres.IConn, config TenantConfig) error {
    // Configurar search_path
    _, err := conn.Exec(context.Background(), 
        fmt.Sprintf("SET search_path TO %s, public", config.Schema))
    return err
}
```

### 3. Multi-Client Platforms
```go
// Middleware para extração de tenant
func extractTenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := r.Header.Get("X-Tenant-ID")
        if tenantID == "" {
            http.Error(w, "Tenant ID required", http.StatusBadRequest)
            return
        }
        
        ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Considerações de Performance

### Schema-based
- **Conexões**: Pool por schema pode ser necessário
- **Queries**: Search_path pode ter overhead
- **Índices**: Mantidos separadamente por schema

### Row-level
- **Índices**: Incluir tenant_id em todos os índices
- **Queries**: Filtros por tenant_id são obrigatórios
- **RLS**: Pode ter overhead de validação

### Database-level
- **Conexões**: Pool separado por tenant
- **Recursos**: Isolamento completo de recursos
- **Escalabilidade**: Fácil distribuição por servidor

## Monitoramento e Métricas

### Métricas por Tenant
```go
type TenantMetrics struct {
    TenantID    int           `json:"tenant_id"`
    UserCount   int           `json:"user_count"`
    DataSize    int64         `json:"data_size_bytes"`
    QueryCount  int64         `json:"query_count"`
    LastActive  time.Time     `json:"last_active"`
}

func collectTenantMetrics(conn postgres.IConn) ([]TenantMetrics, error) {
    // Implementar coleta de métricas
}
```

### Alertas
- Uso excessivo de recursos por tenant
- Tentativas de acesso não autorizado
- Performance degradada
- Falhas de isolamento

## Backup e Disaster Recovery

### Schema-based
```bash
# Backup por schema
pg_dump -n tenant_empresa_a nexs_testdb > tenant_empresa_a_backup.sql

# Restore
psql nexs_testdb < tenant_empresa_a_backup.sql
```

### Row-level
```bash
# Backup filtrado por tenant
pg_dump --data-only --where="tenant_id=1" nexs_testdb > tenant_1_data.sql
```

### Database-level
```bash
# Backup de banco inteiro
pg_dump nexs_tenant_empresa_a > tenant_empresa_a_full_backup.sql
```

## Requisitos

- PostgreSQL rodando em `localhost:5432`
- Banco de dados `nexs_testdb`
- Usuário `nexs_user` com senha `nexs_password`
- Permissões para criar schemas e tabelas

## Migração entre Estratégias

### Schema-based → Row-level
```sql
-- Consolidar dados de múltiplos schemas
INSERT INTO shared_users (tenant_id, name, email)
SELECT 1, name, email FROM tenant_empresa_a.users
UNION ALL
SELECT 2, name, email FROM tenant_empresa_b.users;
```

### Row-level → Schema-based
```sql
-- Criar schemas e distribuir dados
CREATE SCHEMA tenant_empresa_a;
CREATE TABLE tenant_empresa_a.users AS
SELECT id, name, email FROM shared_users WHERE tenant_id = 1;
```

## Boas Práticas

1. **Sempre validar tenant_id** em todas as operações
2. **Usar índices compostos** incluindo tenant_id
3. **Implementar auditoria** de acesso cross-tenant
4. **Monitorar performance** por tenant
5. **Planejar estratégia de backup** adequada
6. **Testar isolamento** regularmente
7. **Documentar configuração** por tenant
