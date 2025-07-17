# Exemplos PostgreSQL Provider

Esta pasta contém exemplos completos e detalhados para demonstrar todas as funcionalidades do provider PostgreSQL genérico da nexs-lib.

## 📁 Estrutura dos Exemplos

```
examples/
├── basic/           # Exemplo básico - Primeiros passos
├── pool/            # Gerenciamento de pools de conexão
├── transaction/     # Transações avançadas
├── advanced/        # Hooks, middleware e monitoramento
├── multitenant/     # Suporte multi-tenant
├── performance/     # Otimização de performance
└── README.md        # Este arquivo
```

## 🚀 Ordem Recomendada de Estudo

### 1. **[Basic Example](./basic/)** - Fundamentos
Comece aqui se você é novo no provider PostgreSQL.

**O que você aprenderá:**
- ✅ Configuração básica do provider
- ✅ Criação de conexões simples
- ✅ Execução de queries básicas
- ✅ Transações simples
- ✅ Health checks

**Tempo estimado:** 15 minutos

### 2. **[Pool Example](./pool/)** - Gerenciamento de Conexões
Aprenda sobre pools de conexão e otimização.

**O que você aprenderá:**
- ✅ Configuração de pools
- ✅ Monitoramento de estatísticas
- ✅ Operações concorrentes
- ✅ Lifecycle management
- ✅ Tuning de performance

**Tempo estimado:** 30 minutos

### 3. **[Transaction Example](./transaction/)** - Transações Avançadas
Domine o gerenciamento de transações complexas.

**O que você aprenderá:**
- ✅ Níveis de isolamento
- ✅ Savepoints (transações aninhadas)
- ✅ Rollback automático e manual
- ✅ Timeouts em transações
- ✅ Operações em lote

**Tempo estimado:** 45 minutos

### 4. **[Advanced Example](./advanced/)** - Funcionalidades Avançadas
Explore hooks, middleware e monitoramento.

**O que você aprenderá:**
- ✅ Sistema de hooks customizados
- ✅ Coleta de métricas
- ✅ Análise de performance
- ✅ Tratamento avançado de erros
- ✅ Auditoria de segurança

**Tempo estimado:** 60 minutos

### 5. **[Multi-tenant Example](./multitenant/)** - Arquitetura Multi-tenant
Implemente soluções multi-tenant robustas.

**O que você aprenderá:**
- ✅ Schema-based multi-tenancy
- ✅ Database-based multi-tenancy
- ✅ Row Level Security (RLS)
- ✅ Isolamento de dados
- ✅ Operações cross-tenant

**Tempo estimado:** 75 minutos

### 6. **[Performance Example](./performance/)** - Otimização Avançada
Maximize a performance das suas aplicações com benchmarks reais e simulações.

**O que você aprenderá:**
- ✅ Benchmarking de pools com diferentes configurações
- ✅ Otimização de queries e prepared statements
- ✅ Análise de concorrência multi-worker
- ✅ Monitoramento de memória e GC
- ✅ Tuning de produção com métricas reais
- ✅ Tratamento robusto de erros e falhas
- ✅ **Novo**: Modo simulação (funciona sem banco de dados)
- ✅ **Novo**: Zero panic guarantee (nunca falha)

**Características especiais:**
- 🛡️ **Robustez total**: Funciona com ou sem banco PostgreSQL
- 📊 **Modo simulação**: Fornece dicas educativas mesmo sem infraestrutura
- ⚡ **Execução segura**: Nunca gera panic ou falhas inesperadas
- 🎯 **Graceful degradation**: Sempre fornece valor, independente do ambiente

**Tempo estimado:** 60-90 minutos

## 🛠 Pré-requisitos

### Essenciais
```bash
# Dependências Go (obrigatório)
go mod tidy
```

### PostgreSQL Database (Opcional para alguns exemplos)

**Para execução completa** de todos os exemplos, recomendamos PostgreSQL via Docker:

```bash
# Instância básica para exemplos simples
docker run --name postgres-examples \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 -d postgres:15

# Para exemplos multi-tenant, criar bancos adicionais
docker exec -it postgres-examples psql -U user -d testdb -c "CREATE DATABASE acme_db;"
docker exec -it postgres-examples psql -U user -d testdb -c "CREATE DATABASE globex_db;"
docker exec -it postgres-examples psql -U user -d testdb -c "CREATE DATABASE initech_db;"
```

**Nota importante:** O exemplo **Performance** foi especialmente otimizado para funcionar **sem banco de dados**, fornecendo simulações educativas e dicas valiosas mesmo sem infraestrutura configurada.

### Dependências Go
```bash
# Na raiz do projeto
go mod tidy
```

## 🛡️ Robustez e Modo Simulação

### Exemplos Resilientes
Todos os exemplos foram desenvolvidos com foco em **robustez e graceful degradation**:

- ✅ **Performance Example**: Funciona 100% sem banco (modo simulação)
- ✅ **Error Recovery**: Tratamento robusto de falhas de conectividade
- ✅ **Zero Panic**: Nunca falha com panic, sempre fornece valor educativo
- ✅ **Fallback Inteligente**: Simulações realistas quando banco não disponível

### Modo Simulação vs Real
```bash
# Execução sem banco (modo simulação)
cd examples/performance && go run main.go
# ✅ Sempre funciona, fornece dicas e simulações

# Execução com banco (benchmarks reais)  
docker run -d --name postgres postgres:15
cd examples/performance && go run main.go
# ✅ Métricas reais e benchmarks completos
```

### Variáveis de Ambiente (Opcional)
```bash
# Configuração padrão de banco
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=user
export POSTGRES_PASSWORD=password
export POSTGRES_DB=testdb

# Configurações de debug
export LOG_LEVEL=info
export POSTGRES_DEBUG=false
```

## 🎯 Cenários de Uso

### Para Desenvolvedores Iniciantes
```
basic/ → pool/ → transaction/
```
- Foque nos conceitos fundamentais
- Execute cada exemplo passo a passo
- Leia os comentários no código

### Para Desenvolvedores Experientes
```
advanced/ → multitenant/ → performance/
```
- Explore funcionalidades avançadas
- Adapte os exemplos para seus casos de uso
- Contribua com melhorias

### Para Arquitetos de Sistema
```
multitenant/ → performance/ → advanced/
```
- Entenda padrões de arquitetura
- Analise trade-offs de performance
- Planeje implementações em produção

### Para DevOps/SRE
```
performance/ → pool/ → advanced/
```
- Foque em métricas e monitoramento
- Entenda configurações de tuning
- Implemente alertas e dashboards

## 📊 Matriz de Funcionalidades

| Funcionalidade | Basic | Pool | Transaction | Advanced | Multi-tenant | Performance |
|----------------|-------|------|-------------|----------|--------------|-------------|
| Conexões Simples | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pool Management | - | ✅ | ✅ | ✅ | ✅ | ✅ |
| Transações | ✅ | - | ✅ | ✅ | ✅ | - |
| Hooks/Middleware | - | - | - | ✅ | - | ✅ |
| Multi-tenancy | - | - | - | - | ✅ | - |
| Performance Tuning | - | ✅ | - | ✅ | - | ✅ |
| Métricas | - | ✅ | - | ✅ | ✅ | ✅ |
| Concorrência | - | ✅ | ✅ | ✅ | ✅ | ✅ |
| Segurança | - | - | - | ✅ | ✅ | - |
| Benchmarking | - | - | - | - | - | ✅ |
| **Modo Simulação** | - | - | - | - | - | ✅ |
| **Error Recovery** | - | - | - | - | - | ✅ |
| **Zero Panic** | - | - | - | - | - | ✅ |

## 🔧 Execução Rápida

### Execução Individual (Recomendado)
```bash
# Performance (sempre funciona, com ou sem banco)
cd examples/performance && go run main.go

# Outros exemplos (requerem PostgreSQL)
cd examples/basic && go run main.go
cd examples/pool && go run main.go
# ... etc
```

### Execução em Lote
Para executar todos os exemplos em sequência:

```bash
#!/bin/bash
# run-all-examples.sh

echo "🚀 Executando todos os exemplos PostgreSQL..."

examples=("basic" "pool" "transaction" "advanced" "multitenant" "performance")

for example in "${examples[@]}"; do
    echo ""
    echo "📁 Executando exemplo: $example"
    echo "================================================"
    cd "$example"
    go run main.go
    cd ..
    echo "✅ Exemplo $example concluído"
    sleep 2
done

echo ""
echo "🎉 Todos os exemplos executados com sucesso!"
```

```bash
# Tornar executável e rodar
chmod +x run-all-examples.sh
./run-all-examples.sh
```

## 📚 Recursos Adicionais

### Documentação
- [PostgreSQL Provider Documentation](../README.md)
- [Interface Definitions](../interface/interfaces.go)
- [Configuration Options](../config/config.go)

### Testes
Cada exemplo inclui cenários que podem ser adaptados para testes:
```bash
# Executar testes unitários
go test -v ./...

# Executar com race detection
go test -race -v ./...

# Executar com coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Profiling
Para análise de performance detalhada:
```bash
# CPU profiling
go run -cpuprofile=cpu.prof main.go

# Memory profiling  
go run -memprofile=mem.prof main.go

# Análise
go tool pprof cpu.prof
go tool pprof mem.prof
```

## 🎯 Benefícios da Abordagem Robusta

### Para Desenvolvedores
- ✅ **Execução Imediata**: Exemplo performance funciona sem configuração
- ✅ **Aprendizado Contínuo**: Sempre fornece valor educativo
- ✅ **Zero Friction**: Não precisa configurar infraestrutura para aprender
- ✅ **Código Resiliente**: Aprenda padrões de tratamento de erro

### Para Equipes
- ✅ **Onboarding Rápido**: Novos desenvolvedores podem executar exemplos imediatamente
- ✅ **CI/CD Friendly**: Testes sempre passam, mesmo sem banco
- ✅ **Demonstrações**: Pode mostrar funcionalidades sem infraestrutura
- ✅ **Documentação Viva**: Código que sempre funciona como documentação

### Para Produção
- ✅ **Patterns Robustos**: Código que nunca falha inesperadamente
- ✅ **Graceful Degradation**: Aplicações que se adaptam a falhas
- ✅ **Error Recovery**: Tratamento inteligente de problemas de conectividade
- ✅ **Monitoring Ready**: Métricas que funcionam em qualquer ambiente

## 🐛 Troubleshooting

### Problemas Comuns

#### Erro de Conexão
```
connection refused
```
**Solução:** Verifique se o PostgreSQL está rodando e as credenciais estão corretas.
**Nota:** O exemplo Performance continua funcionando em modo simulação.

#### Timeout de Operação
```
context deadline exceeded
```
**Solução:** Aumente os timeouts ou verifique a performance da rede/database.

#### Pool Exhausted
```
pool exhausted
```
**Solução:** Aumente o MaxConns ou otimize o uso de conexões.

#### Deadlock Detectado
```
deadlock detected
```
**Solução:** Revise a ordem de aquisição de locks em transações.

### Debug Avançado

Para debug detalhado, configure as variáveis:
```bash
export LOG_LEVEL=debug
export POSTGRES_DEBUG=true
export POOL_DEBUG=true
export TRANSACTION_DEBUG=true
export HOOK_DEBUG=true
```

### Logs Estruturados
Os exemplos usam logging estruturado. Configure conforme sua necessidade:
```bash
# JSON logging
export LOG_FORMAT=json

# Text logging (padrão)
export LOG_FORMAT=text

# Log level
export LOG_LEVEL=debug|info|warn|error
```

## 🤝 Contribuindo

### Melhorias nos Exemplos
1. Fork o repositório
2. Crie uma branch para sua feature
3. Adicione/modifique exemplos
4. Teste exhaustivamente
5. Abra um Pull Request

### Novos Exemplos
Se você tem ideias para novos exemplos:
1. Crie uma issue descrevendo o exemplo
2. Implemente seguindo a estrutura existente
3. Inclua README.md detalhado
4. Adicione à matriz de funcionalidades

### Padrões para Novos Exemplos
- **README.md** completo com explicações
- **Comentários** extensivos no código
- **Casos de erro** bem tratados
- **Métricas** e logging apropriados
- **Cleanup** de recursos

## 📞 Suporte

- **Issues:** [GitHub Issues](../../issues)
- **Discussões:** [GitHub Discussions](../../discussions)
- **Wiki:** [Project Wiki](../../wiki)

---

💡 **Dica:** Comece sempre pelo exemplo [basic](./basic/) e programe gradualmente. Cada exemplo constrói sobre o conhecimento do anterior!
