# 📚 Exemplos do Módulo de Paginação

Esta pasta contém exemplos práticos e progressivos que demonstram todas as funcionalidades do módulo de paginação da `nexs-lib`.

## 🗂️ Estrutura dos Exemplos

### 1. [**01-basic-usage**](./01-basic-usage/) - Uso Básico ⭐
**Ideal para iniciantes**

Demonstra o uso fundamental do módulo:
- Configuração básica do serviço
- Parse de parâmetros HTTP
- Construção de queries SQL
- Criação de resposta paginada
- Navegação entre páginas

```bash
cd 01-basic-usage && go run main.go
```

**Conceitos:** Configuração, Parse, Query Building, Response

---

### 2. [**02-fiber-integration**](./02-fiber-integration/) - Integração com Fiber 🌐
**Para APIs REST**

API completa usando framework Fiber:
- Endpoints REST com paginação
- Interface web para testes
- Tratamento de erros personalizado
- Múltiplos cenários de filtros
- Documentação automática

```bash
cd 02-fiber-integration && go run main.go
# Acesse: http://localhost:3000
```

**Conceitos:** Fiber Provider, API REST, Error Handling, Web Interface

---

### 3. [**03-custom-config**](./03-custom-config/) - Configuração Personalizada ⚙️
**Para casos avançados**

Personalização completa do comportamento:
- Providers customizados
- Validações de negócio específicas
- Configurações por contexto (mobile/web/API)
- Parsing de parâmetros alternativos
- Auto-correção de configurações

```bash
cd 03-custom-config && go run main.go
```

**Conceitos:** Custom Providers, Validation Rules, Multi-Context Config

---

### 4. [**04-error-handling**](./04-error-handling/) - Tratamento de Erros ❌
**Para robustez**

Tratamento robusto de erros e recuperação:
- Detecção automática de erros
- Formatação para APIs
- Mensagens amigáveis para usuários
- Recuperação graceful com fallbacks
- Sugestões automáticas de correção

```bash
cd 04-error-handling && go run main.go
```

**Conceitos:** Domain Errors, Recovery, User-Friendly Messages, Fallbacks

---

### 5. [**05-database-integration**](./05-database-integration/) - Integração com Banco 🗄️
**Para aplicações reais**

Integração completa com PostgreSQL:
- Conexão e configuração otimizada
- Queries com LIMIT/OFFSET
- Filtros complexos e busca textual
- Consultas agregadas paginadas
- Proteção contra SQL injection

```bash
cd 05-database-integration && go run main.go
# Requer PostgreSQL rodando
```

**Conceitos:** SQL Integration, Performance, Security, Aggregation

---

## 🎯 Guia de Escolha

### 👋 **Iniciando com Paginação**
→ Comece com `01-basic-usage`

### 🌐 **Criando uma API REST**
→ Vá para `02-fiber-integration`

### 🔧 **Customizações Específicas**
→ Explore `03-custom-config`

### 🛡️ **Aplicação Robusta**
→ Estude `04-error-handling`

### 🏢 **Sistema de Produção**
→ Implemente `05-database-integration`

## 📋 Pré-requisitos por Exemplo

### Todos os Exemplos
```bash
go mod init exemplo-paginacao
```

### 02-fiber-integration
```bash
go get github.com/gofiber/fiber/v2
```

### 05-database-integration
```bash
go get github.com/lib/pq
# PostgreSQL rodando
```

## 🚀 Execução Rápida

Para testar todos os exemplos básicos:

```bash
# Exemplo básico
cd 01-basic-usage && go run main.go

# Configuração personalizada  
cd ../03-custom-config && go run main.go

# Tratamento de erros
cd ../04-error-handling && go run main.go
```

Para exemplos que requerem serviços externos:

```bash
# API Fiber (requer porta 3000 livre)
cd 02-fiber-integration && go run main.go &
curl "http://localhost:3000/api/products?page=2&limit=3"

# PostgreSQL (requer banco configurado)
cd 05-database-integration && go run main.go
```

## 📖 Conceitos Progressivos

### Nível 1: Fundamentos
- **Parse** de parâmetros URL
- **Validação** básica
- **Query building** automático
- **Response** estruturada

### Nível 2: Integração
- **Framework** web (Fiber)
- **Error handling** robusto
- **API** endpoints
- **Documentation**

### Nível 3: Customização
- **Custom providers**
- **Business rules**
- **Multi-context** config
- **Advanced validation**

### Nível 4: Produção
- **Database** integration
- **Performance** optimization
- **Security** (SQL injection)
- **Monitoring** ready

## 🎨 Padrões Demonstrados

### Repository Pattern
```go
type UserRepository struct {
    db *sql.DB
    paginationService *pagination.PaginationService
}
```

### Dependency Injection
```go
service := pagination.NewPaginationServiceWithProviders(
    config, parser, validator, queryBuilder, calculator,
)
```

### Error Wrapping
```go
if err != nil {
    return nil, fmt.Errorf("failed to parse params: %w", err)
}
```

### Configuration by Convention
```go
cfg := config.NewDefaultConfig()
cfg.DefaultLimit = 20
cfg.MaxLimit = 500
```

## 🔧 Troubleshooting

### Erro de Compilação
```bash
# Verificar go.mod
go mod tidy

# Reinstalar dependências
go clean -modcache
go mod download
```

### PostgreSQL Connection
```bash
# Verificar se está rodando
sudo systemctl status postgresql

# Testar conexão
psql -h localhost -p 5432 -U postgres -d pagination_test
```

### Porta em Uso (Fiber)
```bash
# Verificar porta 3000
lsof -i :3000

# Matar processo se necessário
kill -9 $(lsof -t -i :3000)
```

## 🏆 Próximos Passos

Após dominar os exemplos:

1. **Leia a documentação** completa do módulo
2. **Implemente** em seu projeto real
3. **Customize** conforme suas necessidades
4. **Contribua** com melhorias
5. **Compartilhe** sua experiência

## 💡 Dicas de Performance

- Use **índices** apropriados no banco
- Configure **limites** conservadores
- Implemente **cache** para contagens
- Monitore **queries** lentas
- Use **connection pooling**

---

**Criado em:** 27 de Julho de 2025  
**Última atualização:** 27 de Julho de 2025  
**Autor:** nexs-lib team
