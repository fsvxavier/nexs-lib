# Basic PostgreSQL Provider Example

Este exemplo demonstra o uso básico do provider PostgreSQL de forma genérica e driver-agnóstica.

## 📋 Funcionalidades Demonstradas

- ✅ **Conexão Básica**: Criação e gerenciamento de conexões simples
- ✅ **Queries Simples**: Execução de consultas SELECT básicas
- ✅ **Queries Parametrizadas**: Uso de parâmetros em consultas
- ✅ **Transações Básicas**: Gerenciamento de transações simples
- ✅ **Health Check**: Verificação de saúde da conexão
- ✅ **Estatísticas Básicas**: Monitoramento básico de conexões

## 🚀 Pré-requisitos

1. **PostgreSQL Database**:
   ```bash
   # Usando Docker
   docker run --name postgres-basic \
     -e POSTGRES_USER=user \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=testdb \
     -p 5432:5432 -d postgres:15
   ```

2. **Dependências Go**:
   ```bash
   go mod tidy
   ```

## ⚙️ Configuração

1. **Atualize a string de conexão** no arquivo `main.go`:
   ```go
   cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")
   ```

2. **Configure as variáveis de ambiente** (opcional):
   ```bash
   export POSTGRES_HOST=localhost
   export POSTGRES_PORT=5432
   export POSTGRES_USER=user
   export POSTGRES_PASSWORD=password
   export POSTGRES_DB=testdb
   ```

## 🏃‍♂️ Executando o Exemplo

```bash
# Na pasta do exemplo
cd examples/basic

# Executar o exemplo
go run main.go
```

## 📊 Saída Esperada

```
=== Basic Connection Example ===
✅ Connection health check passed
📊 Connection Stats - Total Queries: 0, Failed Queries: 0

=== Simple Query Example ===
✅ Query result: number=1, greeting=Hello
✅ Parameterized query result: sum=30

=== Basic Transaction Example ===
✅ Created user with ID: 1
✅ Committing transaction
Basic examples completed!
```

## 📝 Conceitos Demonstrados

### 1. Configuração Básica
```go
cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")
provider, err := postgresql.NewPGXProvider()
```

### 2. Conexão Simples
```go
conn, err := provider.NewConn(ctx, cfg)
defer conn.Close(ctx)
```

### 3. Query Parametrizada
```go
row := conn.QueryRow(ctx, "SELECT $1 + $2 as sum", 10, 20)
```

### 4. Transação Básica
```go
tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
    IsoLevel:   interfaces.TxIsoLevelReadCommitted,
    AccessMode: interfaces.TxAccessModeReadWrite,
})
```

## 🔧 Troubleshooting

### Erro de Conexão
```
connection refused
```
**Solução**: Verifique se o PostgreSQL está rodando e as credenciais estão corretas.

### Timeout de Conexão
```
context deadline exceeded
```
**Solução**: Aumente o timeout ou verifique a conectividade de rede.

## 📚 Próximos Passos

Após entender este exemplo básico, explore:

1. **[Pool Example](../pool/)**: Gerenciamento de pools de conexão
2. **[Transaction Example](../transaction/)**: Transações avançadas
3. **[Advanced Example](../advanced/)**: Funcionalidades avançadas como hooks e middleware

## 🐛 Logs de Depuração

Para habilitar logs detalhados, defina:
```bash
export LOG_LEVEL=debug
```

Isso mostrará informações detalhadas sobre:
- Estabelecimento de conexões
- Execução de queries
- Gerenciamento de transações
- Estatísticas de performance
