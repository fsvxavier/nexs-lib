# Exemplo 5: Integração com Banco de Dados

Este exemplo demonstra integração completa do módulo de paginação com PostgreSQL, incluindo:

- Conexão e configuração de banco
- Queries otimizadas com LIMIT/OFFSET
- Contagem automática de registros
- Filtros complexos e busca textual
- Consultas agregadas paginadas
- Mapeamento seguro de campos

## Pré-requisitos

### PostgreSQL
```bash
# Instalar PostgreSQL (Ubuntu/Debian)
sudo apt-get install postgresql postgresql-contrib

# Iniciar serviço
sudo systemctl start postgresql

# Criar banco de teste
sudo -u postgres createdb pagination_test
```

### Driver Go
```bash
go mod init pagination-db-example
go get github.com/lib/pq
```

## Como executar

```bash
cd examples/05-database-integration

# Configurar variáveis de ambiente (opcional)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=pagination_test

go run main.go
```

## Estrutura do Banco

### Tabela `users`
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    department VARCHAR(100) NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Dados de Teste
O exemplo insere automaticamente 20 usuários de teste distribuídos entre departamentos:
- Engineering (6 usuários)
- Marketing (3 usuários) 
- Sales (3 usuários)
- HR (3 usuários)
- Finance (3 usuários)

## Funcionalidades Demonstradas

### 1. 📄 Paginação Básica

```go
params := url.Values{
    "page":  []string{"1"},
    "limit": []string{"5"},
    "sort":  []string{"name"},
    "order": []string{"asc"},
}

response, err := repo.GetUsers(params)
```

**Query gerada:**
```sql
SELECT id, name, email, department, active, created_at, updated_at 
FROM users 
WHERE active = true 
ORDER BY name asc 
LIMIT 5 OFFSET 0
```

### 2. 🏢 Filtro por Departamento

```go
response, err := repo.GetUsersByDepartment("Engineering", params)
```

**Query gerada:**
```sql
SELECT id, name, email, department, active, created_at, updated_at 
FROM users 
WHERE active = true AND department = $1 
ORDER BY created_at desc 
LIMIT 10 OFFSET 0
```

### 3. 🔍 Busca Textual

```go
response, err := repo.SearchUsers("silva", params)
```

**Query gerada:**
```sql
SELECT id, name, email, department, active, created_at, updated_at 
FROM users 
WHERE active = true 
AND (name ILIKE $1 OR email ILIKE $1 OR department ILIKE $1) 
ORDER BY name asc 
LIMIT 5 OFFSET 0
```

### 4. 📊 Consultas Agregadas

```go
response, err := repo.GetUserStats(params)
```

**Query gerada:**
```sql
SELECT 
    department,
    COUNT(*) as user_count,
    AVG(EXTRACT(DAYS FROM NOW() - created_at)) as avg_created_days
FROM users 
WHERE active = true 
GROUP BY department 
HAVING COUNT(*) > 0 
ORDER BY user_count desc 
LIMIT 10 OFFSET 0
```

## Saída do Exemplo

```
🗄️  Exemplos de Integração com Banco de Dados - Módulo de Paginação
===================================================================

📡 Conectando ao banco de dados...
✅ Conexão com banco estabelecida
✅ Tabela já possui 20 registros

=== 1. Paginação Básica ===
🔍 Executando query: SELECT id, name, email, department, active, created_at, updated_at FROM users WHERE active = true ORDER BY name asc LIMIT 5 OFFSET 0
📊 Executando count: SELECT COUNT(*) FROM (SELECT id, name, email, department, active, created_at, updated_at FROM users WHERE active = true) AS count_query
📈 Total de registros encontrados: 20
👥 Usuários carregados: 5
📄 Página 1 de 4
👥 Usuários nesta página: 5
📊 Total de usuários: 20
   1. Alice Silva <alice@company.com> [Engineering]
   2. Bob Santos <bob@company.com> [Engineering]
   3. Carol Oliveira <carol@company.com> [Marketing]
   ... e mais 2 usuários

=== 2. Filtro por Departamento ===
🏢 Buscando usuários do departamento: Engineering
🏢 Usuários do departamento Engineering: 6
   • Paul Dias <paul@company.com>
   • Liam Cardoso <liam@company.com>
   • Ivy Ferreira <ivy@company.com>
   • Eva Lima <eva@company.com>
   • Bob Santos <bob@company.com>
   • Alice Silva <alice@company.com>

=== 3. Busca Textual ===
🔎 Buscando por: silva
🔍 Resultados para 'silva': 1
   ✅ Alice Silva <alice@company.com> [Engineering]
```

## UserRepository - Estrutura Completa

### Inicialização
```go
func NewUserRepository(db *sql.DB) *UserRepository {
    cfg := config.NewDefaultConfig()
    cfg.DefaultLimit = 20
    cfg.MaxLimit = 500
    cfg.DefaultSortField = "created_at"
    cfg.DefaultSortOrder = "desc"
    
    return &UserRepository{
        db: db,
        paginationService: pagination.NewPaginationService(cfg),
    }
}
```

### Métodos Implementados

#### `GetUsers(params url.Values)` 
- Listagem geral com paginação
- Validação de campos ordenáveis
- Contagem automática de registros

#### `GetUsersByDepartment(department, params)`
- Filtro específico por departamento  
- Queries parametrizadas (proteção SQL injection)
- Contagem ajustada ao filtro

#### `SearchUsers(searchTerm, params)`
- Busca textual em múltiplos campos
- Uso de ILIKE para PostgreSQL
- Pattern matching com wildcards

#### `GetUserStats(params)`
- Consultas agregadas com GROUP BY
- Cálculos de médias e contagens
- Paginação de resultados agregados

## Conceitos Demonstrados

- ✅ **SQL Injection Protection** - Queries parametrizadas
- ✅ **Performance Otimizada** - COUNT separado da query principal
- ✅ **Mapeamento Seguro** - Whitelist de campos ordenáveis
- ✅ **Filtros Complexos** - WHERE com múltiplas condições
- ✅ **Busca Full-Text** - ILIKE com wildcards
- ✅ **Consultas Agregadas** - GROUP BY com paginação
- ✅ **Error Handling** - Tratamento de erros SQL
- ✅ **Resource Management** - Defer rows.Close()

## Integração com bibliotecas do projeto

### Domain Errors
```go
import "github.com/fsvxavier/nexs-lib/domainerrors"

// Erros de validação são automaticamente convertidos
_, err := r.paginationService.ParseRequest(params, sortableFields...)
if err != nil {
    return nil, fmt.Errorf("invalid pagination parameters: %w", err)
}
```

### PostgreSQL Driver
```go
import _ "github.com/lib/pq"

// Conexão otimizada para PostgreSQL
db, err := sql.Open("postgres", dsn)
```

## Otimizações de Performance

### 1. 🎯 Campos Indexados
```sql
CREATE INDEX idx_users_active ON users(active);
CREATE INDEX idx_users_department ON users(department);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_name ON users(name);
```

### 2. 📊 Query de Count Otimizada
- Count separado da query principal
- Evita carregar dados desnecessários
- Cache de contagem para queries frequentes

### 3. 🔍 Busca Otimizada
- Índices GIN para busca textual
- LIMIT aplicado corretamente
- Prepared statements para queries repetidas

## Configuração de Produção

### Environment Variables
```bash
export DB_HOST=production-db-host
export DB_PORT=5432
export DB_USER=app_user
export DB_PASSWORD=secure_password
export DB_NAME=production_db
export DB_SSLMODE=require
export DB_MAX_CONNECTIONS=20
export DB_MAX_IDLE=5
export DB_MAX_LIFETIME=1h
```

### Connection Pool
```go
db.SetMaxOpenConns(20)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(time.Hour)
```

## Próximos Passos

Após entender integração com banco, veja:
- `06-performance-optimization` - Otimizações avançadas
- `07-middleware-advanced` - Middleware para APIs
- `08-cursor-pagination` - Paginação baseada em cursor
