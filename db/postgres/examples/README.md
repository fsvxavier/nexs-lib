# Exemplos PostgreSQL - NEXS-LIB

Esta pasta contém exemplos práticos de uso da biblioteca NEXS-LIB para PostgreSQL, demonstrando diferentes funcionalidades e padrões de uso.

## Estrutura dos Exemplos

```
db/postgres/examples/
├── basic/          # Exemplos básicos de conexão e uso
├── replicas/       # Exemplos com read replicas
├── advanced/       # Exemplos avançados com múltiplas funcionalidades
├── pool/           # Exemplos de pool de conexões
└── README.md       # Este arquivo
```

## Visão Geral dos Exemplos

### 1. Basic (Básico)
**Localização**: `basic/`
**Funcionalidades**:
- Conexão simples com PostgreSQL
- Pool de conexões básico
- Queries e transações simples
- Configuração básica

**Quando usar**: Primeiros passos com a biblioteca, conceitos fundamentais.

### 2. Replicas (Read Replicas)
**Localização**: `replicas/`
**Funcionalidades**:
- Configuração de read replicas
- Load balancing entre réplicas
- Failover automático
- Uso em cenários reais

**Quando usar**: Aplicações com alta demanda de leitura, necessidade de escalabilidade.

### 3. Advanced (Avançado)
**Localização**: `advanced/`
**Funcionalidades**:
- Pool management avançado
- Transações complexas
- Operações batch
- Operações concorrentes
- Tratamento de erros
- Multi-tenancy
- LISTEN/NOTIFY
- Testes de performance

**Quando usar**: Aplicações complexas, alta concorrência, recursos avançados.

### 4. Pool (Pool de Conexões)
**Localização**: `pool/`
**Funcionalidades**:
- Configuração detalhada de pools
- Métricas e monitoramento
- Timeouts e limites
- Lifecycle management
- Testes de carga

**Quando usar**: Otimização de performance, monitoramento, tuning de aplicações.

## Como Executar os Exemplos

### Usando Docker (Recomendado)

A forma mais fácil de executar os exemplos é usando o Docker, que fornece um ambiente PostgreSQL completo com primary + replicas.

```bash
# 1. Iniciar a infraestrutura
./infraestructure/manage.sh start

# 2. Executar exemplos específicos
./infraestructure/manage.sh example basic      # Exemplo básico
./infraestructure/manage.sh example replicas   # Exemplo com replicas
./infraestructure/manage.sh example advanced   # Exemplo avançado
./infraestructure/manage.sh example pool       # Exemplo de pool

# 3. Parar a infraestrutura
./infraestructure/manage.sh stop
```

### Execução Manual

Se preferir executar manualmente, primeiro configure o ambiente:

```bash
# Configurar variáveis de ambiente
export NEXS_DB_DSN="postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"
export NEXS_DB_REPLICA_DSN="postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb"

# Executar exemplo específico
cd db/postgres/examples/basic
go run main.go
```

## Pré-requisitos

### Para Docker (Recomendado)
- Docker
- Docker Compose
- Go 1.19+

### Para Execução Manual
- PostgreSQL 12+
- Go 1.19+
- Banco de dados `nexs_testdb` configurado
- Usuário `nexs_user` com permissões adequadas

## Infraestrutura Docker

A infraestrutura Docker inclui:

| Serviço | Porta | Descrição |
|---------|-------|-----------|
| postgres-primary | 5432 | Banco principal (leitura/escrita) |
| postgres-replica1 | 5433 | Réplica 1 (somente leitura) |
| postgres-replica2 | 5434 | Réplica 2 (somente leitura) |
| redis | 6379 | Cache Redis |
| pgadmin | 8080 | Interface web PgAdmin |

### Comandos de Infraestrutura

```bash
# Verificar status
./infraestructure/manage.sh status

# Ver logs
./infraestructure/manage.sh logs [serviço]

# Resetar banco (cuidado!)
./infraestructure/manage.sh reset

# Executar testes
./infraestructure/manage.sh test
```

## Guia de Aprendizado

### Sequência Recomendada

1. **Comece com Basic**: Entenda conceitos fundamentais
2. **Explore Pool**: Aprenda sobre otimização de conexões
3. **Teste Replicas**: Implemente escalabilidade de leitura
4. **Avance para Advanced**: Use recursos complexos

### Por Nível de Experiência

#### Iniciante
- `basic/` - Conceitos fundamentais
- `pool/` - Otimização básica

#### Intermediário
- `replicas/` - Escalabilidade
- `advanced/` - Recursos avançados (parte 1)

#### Avançado
- `advanced/` - Recursos avançados (completo)
- Personalização dos exemplos

## Estrutura de Cada Exemplo

Cada exemplo contém:

```
exemplo/
├── main.go       # Código principal
├── README.md     # Documentação detalhada
└── [outros arquivos específicos]
```

## Configuração de Conexão

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|---------|
| `NEXS_DB_DSN` | DSN do banco principal | `postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb` |
| `NEXS_DB_REPLICA_DSN` | DSN da réplica | `postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb` |

### Configuração Avançada

```go
cfg := postgres.NewConfigWithOptions(
    dsn,
    postgres.WithMaxConns(20),
    postgres.WithMinConns(5),
    postgres.WithMaxConnLifetime(30*time.Minute),
    postgres.WithMaxConnIdleTime(10*time.Minute),
)
```

## Troubleshooting

### Problemas Comuns

1. **Erro de Conexão**
   ```bash
   # Verificar se a infraestrutura está rodando
   ./infraestructure/manage.sh status
   
   # Verificar logs
   ./infraestructure/manage.sh logs postgres-primary
   ```

2. **Banco não Inicializado**
   ```bash
   # Resetar e recriar banco
   ./infraestructure/manage.sh reset
   ./infraestructure/manage.sh start
   ```

3. **Exemplo não Encontrado**
   ```bash
   # Verificar exemplos disponíveis
   ls -la db/postgres/examples/
   ```

### Logs e Debugging

```bash
# Logs gerais
./infraestructure/manage.sh logs

# Logs específicos
./infraestructure/manage.sh logs postgres-primary
./infraestructure/manage.sh logs postgres-replica1

# Status detalhado
./infraestructure/manage.sh status
```

## Contribuindo

### Adicionando Novos Exemplos

1. Crie nova pasta em `db/postgres/examples/`
2. Adicione `main.go` com exemplo
3. Crie `README.md` detalhado
4. Atualize este README principal
5. Teste com Docker

### Padrões de Código

- Use comentários em português
- Documente cada função
- Inclua tratamento de erros
- Use configurações do ambiente
- Mantenha código limpo e legível

### Estrutura do README de Exemplo

```markdown
# Título do Exemplo

## Funcionalidades Demonstradas
- Lista de funcionalidades

## Como Executar
- Instruções Docker
- Execução manual

## Configuração
- Variáveis de ambiente
- Opções de configuração

## Saída Esperada
- Exemplo de output

## Próximos Passos
- Melhorias possíveis
```

## Recursos Adicionais

### Documentação
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Go Database/SQL Tutorial](https://golang.org/pkg/database/sql/)

### Ferramentas
- [PgAdmin](http://localhost:8080) - Interface web (quando Docker rodando)
- [pgx Documentation](https://github.com/jackc/pgx)

### Observabilidade
- Métricas de conexão
- Logs estruturados
- Monitoramento de performance

## Próximos Passos

1. **Adicionar Exemplos**:
   - Streaming de dados
   - Prepared statements
   - Connection health checks
   - Retry policies

2. **Melhorar Infraestrutura**:
   - Monitoring stack
   - Load balancer
   - Backup automation

3. **Adicionar Testes**:
   - Testes de integração
   - Testes de performance
   - Testes de failover

4. **Documentação**:
   - Guias de migração
   - Best practices
   - Performance tuning

---

**Dúvidas?** Consulte os READMEs individuais de cada exemplo ou abra uma issue no repositório.
