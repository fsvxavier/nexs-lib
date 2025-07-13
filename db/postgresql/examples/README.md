# Exemplos de Uso - PostgreSQL Provider

Esta pasta contém exemplos práticos de como usar o módulo PostgreSQL com diferentes drivers.

## Estrutura dos Exemplos

```
examples/
├── basic/          # Uso básico com factory pattern
├── pgx/           # Recursos específicos do PGX
├── gorm/          # Uso com GORM ORM
└── pq/            # Uso com lib/pq
```

## Configuração do Ambiente

### PostgreSQL Local (Docker)

```bash
# Iniciar PostgreSQL para testes
docker run --name postgres-test \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  -d postgres:15

# Verificar se está rodando
docker ps | grep postgres-test
```

### Variáveis de Ambiente

Crie um arquivo `.env` ou exporte as variáveis:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=testdb
export DB_USER=postgres
export DB_PASSWORD=password
export DB_SSL_MODE=disable
```

## Executando os Exemplos

### 1. Exemplo Básico
Demonstra o uso do factory pattern e operações básicas.

```bash
cd basic/
go run main.go
```

**Features demonstradas:**
- Factory pattern para criação de providers
- Operações CRUD básicas
- Pool de conexões
- Transações simples

### 2. Exemplo PGX
Recursos avançados específicos do driver PGX.

```bash
cd pgx/
go run main.go
```

**Features demonstradas:**
- Operações em lote (batch)
- Transações com opções
- Pool statistics
- Recursos avançados do PGX

### 3. Exemplo GORM
Uso com GORM ORM através da interface unificada.

```bash
cd gorm/
go run main.go
```

**Features demonstradas:**
- Integração com GORM
- Operações ORM-style
- Migrações (conceitual)
- Mapeamento de estruturas

### 4. Exemplo lib/pq
Uso com o driver clássico lib/pq.

```bash
cd pq/
go run main.go
```

**Features demonstradas:**
- Compatibilidade com database/sql
- Operações row-level
- Transações padrão
- Scanning de resultados

## Dependências

Certifique-se de ter as dependências instaladas:

```bash
# Dependências básicas
go mod download

# Para PGX
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool

# Para GORM  
go get gorm.io/gorm
go get gorm.io/driver/postgres

# Para lib/pq
go get github.com/lib/pq
```

## Estrutura dos Dados

Todos os exemplos usam uma estrutura simples de usuário:

```go
type User struct {
    ID    int    `json:"id" db:"id" gorm:"primaryKey"`
    Name  string `json:"name" db:"name" gorm:"size:100;not null"`
    Email string `json:"email" db:"email" gorm:"size:100;uniqueIndex;not null"`
}
```

## Outputs Esperados

### Sucesso
```
=== Example 1: Pool Operations ===
Connection ping successful
Pool stats - Max: 25, Total: 1, Idle: 1

=== Example 2: CRUD Operations ===
Table created successfully
User inserted successfully
Found user: {ID:1 Name:John Doe Email:john@example.com}
Found 1 users
Total users: 1

=== Example 3: Transaction Operations ===
Transaction committed successfully

=== Example completed ===
```

### Em caso de erro de conexão
```
Failed to connect: dial tcp [::1]:5432: connect: connection refused
```

**Solução**: Verifique se o PostgreSQL está rodando e as configurações estão corretas.

## Testando sem PostgreSQL

Se você não tem PostgreSQL disponível, pode usar os mocks:

```go
import "github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx/mocks"

func ExampleWithMocks() {
    mockProvider := &mocks.MockDatabaseProvider{}
    mockPool := &mocks.MockPool{}
    mockConn := &mocks.MockConn{}
    
    // Configurar mocks
    mockProvider.On("Pool").Return(mockPool)
    mockPool.On("Acquire", mock.Anything).Return(mockConn, nil)
    
    // Usar nos exemplos...
}
```

## Troubleshooting

### Erro: "driver: bad connection"
- Verifique se o PostgreSQL está rodando
- Confirme as credenciais de conexão
- Teste a conectividade: `telnet localhost 5432`

### Erro: "relation does not exist"
- As tabelas são criadas automaticamente nos exemplos
- Verifique se tem permissões para criar tabelas
- Confira se está conectando no banco correto

### Erro: "too many connections"
- Ajuste `MaxOpenConns` na configuração
- Verifique se está fechando conexões adequadamente
- Use `defer conn.Release(ctx)` sempre

## Performance Tips

1. **Pool de Conexões**: Ajuste baseado na sua carga
```go
config.WithMaxOpenConns(25)
config.WithMaxIdleConns(5)
```

2. **Transações**: Use para operações relacionadas
```go
tx, _ := conn.BeginTransaction(ctx)
// múltiplas operações
tx.Commit(ctx)
```

3. **Batch Operations**: Para inserções em massa (PGX)
```go
batch := pgx.NewBatch()
// adicionar operações
conn.SendBatch(ctx, batch)
```

## Contribuindo

Para adicionar novos exemplos:

1. Crie um novo diretório em `examples/`
2. Adicione `main.go` com exemplo funcional
3. Documente as features demonstradas
4. Teste com PostgreSQL real
5. Abra um Pull Request

---

**Última Atualização**: Janeiro 2025  
**Dificuldade**: Iniciante a Intermediário  
**Tempo Estimado**: 30-60 minutos por exemplo
