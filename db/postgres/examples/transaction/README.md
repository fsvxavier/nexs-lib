# Exemplo de Transação

Este exemplo demonstra como usar transações no PostgreSQL usando o método `Begin()` da biblioteca nexs-lib.

## Visão Geral

O exemplo mostra diferentes cenários de uso de transações:

1. **Transação básica com Begin()** - Transferência de dinheiro entre contas
2. **Rollback de transação** - Cancelamento de operações
3. **Transação com BeginTx()** - Usando opções específicas de isolamento

## Funcionalidades Demonstradas

### 1. Transação Básica
- Inicia uma transação com `conn.Begin(ctx)`
- Executa múltiplas operações SQL dentro da transação
- Faz commit se todas as operações são bem-sucedidas
- Faz rollback se alguma operação falha

### 2. Controle de Transação
- Verificação de saldo antes da transferência
- Operações de débito e crédito atômicas
- Tratamento de erros com rollback automático

### 3. Transação com Opções
- Uso de `BeginTx()` com opções específicas
- Configuração de nível de isolamento (`ReadCommitted`)
- Configuração de modo de acesso (`ReadWrite`)

## Pré-requisitos

1. PostgreSQL rodando localmente
2. Banco de dados `nexs_testdb` criado
3. Usuário `nexs_user` com senha `nexs_password`

### Configuração do Banco

```sql
-- Criar usuário
CREATE USER nexs_user WITH PASSWORD 'nexs_password';

-- Criar banco
CREATE DATABASE nexs_testdb OWNER nexs_user;

-- Conceder permissões
GRANT ALL PRIVILEGES ON DATABASE nexs_testdb TO nexs_user;
```

## Como Executar

```bash
cd /home/fabricioxavier/go/src/github.com/fsvxavier/nexs-lib/db/postgres/examples/transaction
go run main.go
```

## Estrutura do Exemplo

### Tabela de Teste
```sql
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    balance DECIMAL(10,2) NOT NULL DEFAULT 0.00
)
```

### Dados Iniciais
- Alice: $1000.00
- Bob: $500.00

## Fluxo de Execução

1. **Configuração inicial**: Conexão com banco e criação da tabela
2. **Inserção de dados**: Criação de contas com saldos iniciais
3. **Transação de transferência**: Transfere $150.00 de Alice para Bob
4. **Verificação de saldo**: Confirma se há saldo suficiente
5. **Commit da transação**: Se tudo OK, confirma as alterações
6. **Exemplo de rollback**: Tenta transferir valor maior que o disponível
7. **Transação com opções**: Demonstra uso de BeginTx com configurações específicas
8. **Limpeza**: Remove a tabela de teste

## Saída Esperada

```
=== Exemplo de Transação com Begin() ===

1. Conectando ao banco...
2. Criando tabela de teste...
3. Limpando dados anteriores...
4. Inserindo dados iniciais...

5. Exemplo de transação com Begin() - Transferência de dinheiro...
   Balances iniciais:
     Alice: $1000.00
     Bob: $500.00
   Iniciando transação...
   Transferindo $150.00 de Alice para Bob...
   Fazendo commit da transação...
   Transação commitada com sucesso!

6. Balances finais:
     Alice: $850.00
     Bob: $650.00

7. Exemplo de transação com rollback...
   Iniciando transação que será cancelada...
   Tentando transferir $2000.00 (mais do que disponível)...
   Fazendo rollback da transação...
   Rollback realizado com sucesso!
   Verificando que os dados não foram alterados:
     Alice: $850.00
     Bob: $650.00

8. Exemplo de transação com BeginTx() e opções...
   Iniciando transação com nível de isolamento ReadCommitted...
   Fazendo operação na transação com opções...
   Fazendo commit da transação com opções...
   Transação com opções commitada com sucesso!

9. Limpando tabela de teste...

=== Exemplo de Transação com Begin() - CONCLUÍDO ===
```

## Conceitos Importantes

### Transações ACID
- **Atomicidade**: Todas as operações são executadas ou nenhuma é
- **Consistência**: Os dados permanecem em estado válido
- **Isolamento**: Transações não interferem umas nas outras
- **Durabilidade**: Mudanças commitadas são permanentes

### Níveis de Isolamento
- `ReadUncommitted`: Permite leitura de dados não commitados
- `ReadCommitted`: Só permite leitura de dados commitados (padrão)
- `RepeatableRead`: Garante leituras consistentes durante a transação
- `Serializable`: Mais alto nível de isolamento

### Modos de Acesso
- `ReadOnly`: Transação apenas para leitura
- `ReadWrite`: Transação para leitura e escrita (padrão)

## Tratamento de Erros

O exemplo demonstra:
- Verificação de saldo antes da transferência
- Rollback automático em caso de erro
- Tratamento adequado de erros de conexão e SQL

## Arquivos Relacionados

- `main.go`: Código principal do exemplo
- `README.md`: Esta documentação
- `../../../postgres.go`: Implementação da biblioteca
- `../../../interfaces/interfaces.go`: Interfaces utilizadas
