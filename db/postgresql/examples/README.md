# Examples - PostgreSQL Module

Esta pasta contém exemplos práticos abrangentes que demonstram como usar o módulo PostgreSQL em diferentes cenários reais.

## 📚 Exemplos Disponíveis

### 1. 🔄 [Basic Operations](basic_operations/)
**Operações CRUD fundamentais e conceitos básicos**

Demonstra:
- Conexão com banco de dados
- Operações INSERT, SELECT, UPDATE, DELETE
- Transações básicas
- Tratamento de erros
- Configuração de conexão

**Ideal para**: Iniciantes que querem entender os conceitos básicos

### 2. 🔌 [Connection Pool Management](connection_pool/)
**Gerenciamento avançado de pools de conexão**

Demonstra:
- Configuração de pool com diferentes tamanhos
- Monitoramento de estatísticas em tempo real
- Health checks automáticos
- Teste de carga com múltiplos workers
- Cenários de alta concorrência

**Ideal para**: Aplicações que precisam otimizar performance e monitoramento

### 3. 💳 [Advanced Transactions](advanced_transactions/)
**Transações complexas e cenários empresariais**

Demonstra:
- Savepoints e rollbacks parciais
- Diferentes níveis de isolamento
- Sistema bancário completo
- Retry automático em falhas
- Transações longas com timeout

**Ideal para**: Sistemas financeiros e aplicações críticas

### 4. 📦 [Batch Operations](batch_operations/)
**Operações em lote para grandes volumes de dados**

Demonstra:
- Inserções massivas otimizadas
- Comparação de performance entre estratégias
- Monitoramento de uso de memória
- Processamento de grandes datasets
- Benchmarks automáticos

**Ideal para**: ETL, importação de dados e processamento em lote

## 🚀 Como Executar os Exemplos

### Pré-requisitos

1. **Go 1.21+** instalado
2. **PostgreSQL 12+** rodando
3. **Banco de dados de teste** (exemplo: `example`)

### Configuração Rápida com Docker

```bash
# Iniciar PostgreSQL em container
docker run --name postgres-nexs \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=example \
  -p 5432:5432 \
  -d postgres:15

# Aguardar o banco inicializar
sleep 10
```

### Variáveis de Ambiente (Opcional)

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=example
export DB_USER=postgres
export DB_PASSWORD=password
```

### Executar Exemplo Específico

```bash
# Navegar para o exemplo desejado
cd examples/basic_operations

# Baixar dependências
go mod tidy

# Executar exemplo
go run .
```

## 📋 Estrutura dos Exemplos

Cada exemplo segue uma estrutura padronizada:

```
example_name/
├── README.md          # Documentação específica
├── go.mod            # Dependências do Go
├── main.go           # Arquivo principal
├── models.go         # Estruturas de dados (quando aplicável)
├── *.go             # Arquivos auxiliares específicos
```

## 🎯 Escolhendo o Exemplo Certo

| Se você quer... | Use este exemplo |
|------------------|------------------|
| Aprender o básico | [Basic Operations](basic_operations/) |
| Otimizar performance | [Connection Pool](connection_pool/) |
| Implementar transações complexas | [Advanced Transactions](advanced_transactions/) |
| Processar grandes volumes | [Batch Operations](batch_operations/) |

## 🔧 Configurações Comuns

### Configuração Básica
```go
cfg := config.NewConfig(
    config.WithHost("localhost"),
    config.WithPort(5432),
    config.WithDatabase("example"),
    config.WithUsername("postgres"),
    config.WithPassword("password"),
)
```

### Configuração para Alta Performance
```go
cfg := config.NewConfig(
    config.WithHost("localhost"),
    config.WithDatabase("example"),
    config.WithUsername("postgres"),
    config.WithPassword("password"),
    config.WithMaxConns(50),           // Pool grande
    config.WithMinConns(10),           // Conexões sempre ativas
    config.WithQueryTimeout(30*time.Second),
    config.WithConnectTimeout(10*time.Second),
)
```

### Configuração para Desenvolvimento
```go
cfg := config.NewConfig(
    config.WithHost("localhost"),
    config.WithDatabase("example"),
    config.WithUsername("postgres"),
    config.WithPassword("password"),
    config.WithMaxConns(5),            // Pool pequeno
    config.WithMinConns(1),            // Mínimo de conexões
    config.WithTracingEnabled(true),   // Debug habilitado
)
```

## 🐛 Troubleshooting

### Erro de Conexão
```
❌ Erro ao conectar com banco: connection refused
```
**Solução**: Verificar se PostgreSQL está rodando na porta correta

### Erro de Autenticação
```
❌ Erro ao conectar: password authentication failed
```
**Solução**: Verificar usuário e senha nas variáveis de ambiente

### Erro de Pool Esgotado
```
❌ Erro ao adquirir conexão: pool timeout
```
**Solução**: Aumentar `MaxConns` ou verificar vazamentos de conexão

### Banco de Dados Não Existe
```
❌ Erro ao conectar: database "example" does not exist
```
**Solução**: Criar o banco ou ajustar o nome na configuração

## 📊 Benchmarks e Performance

### Resultados Esperados (Hardware de Referência)

| Operação | Taxa Típica | Observações |
|----------|-------------|-------------|
| Inserção Individual | ~1,000/seg | Baseline |
| Batch Insert (500) | ~15,000/seg | 15x mais rápido |
| Transação Preparada | ~8,000/seg | Boa para ACID |
| Pool Connections | ~20,000/seg | Com pool otimizado |

*Resultados podem variar baseado em hardware, rede e configuração do banco*

## 🛡️ Boas Práticas Demonstradas

### Gerenciamento de Conexões
- ✅ Sempre fazer `defer conn.Release(ctx)`
- ✅ Usar timeouts apropriados
- ✅ Configurar pool adequadamente
- ❌ Não esquecer de liberar conexões

### Tratamento de Erros
- ✅ Verificar todos os erros
- ✅ Fazer rollback em transações
- ✅ Log adequado de erros
- ❌ Não ignorar erros silenciosamente

### Performance
- ✅ Usar batch operations para volumes grandes
- ✅ Prepared statements para queries repetitivas
- ✅ Monitorar estatísticas do pool
- ❌ Não fazer SELECT N+1

### Segurança
- ✅ Sempre usar parâmetros ($1, $2, etc)
- ✅ Validar entrada de dados
- ✅ Usar HTTPS em produção
- ❌ Nunca concatenar SQL diretamente

## 🔄 Próximos Passos

Após executar os exemplos:

1. **Adapte para seu caso de uso**: Modifique os exemplos para sua aplicação
2. **Teste em ambiente real**: Execute com dados reais do seu sistema
3. **Otimize configurações**: Ajuste pools e timeouts para sua carga
4. **Implemente monitoramento**: Use as métricas em produção
5. **Contribua**: Ajude a melhorar os exemplos

## 🤝 Contribuindo

Encontrou algum problema ou tem sugestões?

1. Abra uma [issue](https://github.com/fsvxavier/nexs-lib/issues)
2. Proponha melhorias via Pull Request
3. Compartilhe seu caso de uso

## 📚 Documentação Adicional

- [README Principal](../README.md) - Documentação completa do módulo
- [Next Steps](../NEXT_STEPS.md) - Roadmap e próximas funcionalidades
- [API Reference](../interfaces.go) - Referência das interfaces
