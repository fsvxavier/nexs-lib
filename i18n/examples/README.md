# I18n Examples

Este diretório contém exemplos práticos e completos demonstrando o uso da biblioteca i18n em diferentes cenários e arquiteturas.

## 📖 Índice de Exemplos

### 🔸 Exemplos Básicos

1. **[basic_json/](basic_json/)** - Uso básico com provider JSON
   - Configuração simples
   - Traduções com parâmetros
   - Chaves aninhadas
   - Comportamento de fallback

2. **[basic_yaml/](basic_yaml/)** - Uso básico com provider YAML
   - Configuração com YAML
   - Mesmas funcionalidades do JSON
   - Formato YAML para traduções

### 🔸 Exemplos Avançados

3. **[advanced/](advanced/)** - Funcionalidades avançadas com hooks
   - Sistema de hooks (logging, metrics)
   - Configuração avançada
   - Monitoramento de performance
   - Lifecycle management

4. **[middleware_demo/](middleware_demo/)** - Demonstração de middlewares
   - Cache middleware customizado
   - Rate limiting middleware
   - Cadeia de middlewares
   - Implementação de interfaces

5. **[performance_demo/](performance_demo/)** - Otimizações de performance
   - Benchmarks abrangentes
   - Análise de memória
   - Performance concorrente
   - String pooling e interning

### 🔸 Aplicações Web

6. **[web_app_gin/](web_app_gin/)** - Aplicação web completa com Gin
   - Framework Gin integrado
   - Middleware de i18n
   - Endpoints internacionalizados
   - Gestão de idiomas via query/header

7. **[api_rest_echo/](api_rest_echo/)** - API REST com Echo framework
   - API RESTful completa
   - CRUD internacionalizado
   - Responses padronizadas
   - Validação multilíngue

### 🔸 Microserviços

8. **[microservice/](microservice/)** - Microserviço i18n
   - Arquitetura de microserviço
   - API HTTP para traduções
   - Health checks
   - Métricas e monitoramento
   - Documentação da API

### 🔸 Ferramentas CLI

9. **[cli_tool/](cli_tool/)** - Ferramenta de linha de comando
   - Interface CLI interativa
   - Comandos de tradução
   - Validação de arquivos
   - Estatísticas de uso

## 🚀 Como Executar os Exemplos

### Pré-requisitos

```bash
# Certifique-se de que está no diretório raiz do módulo i18n
cd /mnt/e/go/src/github.com/fsvxavier/nexs-lib/i18n

# Instale as dependências
go mod tidy
```

### ⚡ Executar Todos os Exemplos (Recomendado)

```bash
# Execute o script automático que testa todos os exemplos
cd examples
./run_examples.sh

# Opções disponíveis:
./run_examples.sh --help     # Mostrar ajuda
./run_examples.sh --quiet    # Execução silenciosa
./run_examples.sh --verbose  # Execução detalhada
```

### 📋 Executando os Exemplos Individualmente

#### 1. Exemplos Básicos

```bash
# JSON básico
cd examples/basic_json
go run main.go

# YAML básico
cd examples/basic_yaml
go run main.go
```

#### 2. Exemplos Avançados

```bash
# Funcionalidades avançadas
cd examples/advanced
go run main.go

# Demonstração de middlewares
cd examples/middleware_demo
go run main.go

# Demo de performance
cd examples/performance_demo
go run main.go
```

#### 3. Aplicações Web

```bash
# Aplicação Gin (requer: go get github.com/gin-gonic/gin)
cd examples/web_app_gin
go mod init web_app_gin
go get github.com/gin-gonic/gin
go get github.com/fsvxavier/nexs-lib/i18n
go run main.go
# Acesse: http://localhost:8080

# API Echo (requer: go get github.com/labstack/echo/v4)
cd examples/api_rest_echo
go mod init api_rest_echo
go get github.com/labstack/echo/v4
go get github.com/fsvxavier/nexs-lib/i18n
go run main.go
# Acesse: http://localhost:8080
```

#### 4. Microserviço

```bash
cd examples/microservice
go mod init microservice_i18n
go get github.com/fsvxavier/nexs-lib/i18n
go run main.go
# Acesse: http://localhost:8080/health
```

#### 5. CLI Tool

```bash
cd examples/cli_tool
go mod init cli_tool
go get github.com/fsvxavier/nexs-lib/i18n

# Modo interativo
go run main.go -interactive

# Comandos específicos
go run main.go -cmd translate -lang pt
go run main.go -cmd list-keys
go run main.go -cmd validate -verbose
```

## 📋 Funcionalidades Demonstradas

### Por Exemplo

| Exemplo | JSON | YAML | Hooks | Middlewares | Cache | Rate Limit | Web | API | CLI | Performance |
|---------|------|------|-------|-------------|-------|------------|-----|-----|-----|-------------|
| basic_json | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| basic_yaml | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| advanced | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| middleware_demo | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ |
| performance_demo | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| web_app_gin | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| api_rest_echo | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| microservice | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| cli_tool | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ |

### Recursos Específicos

- **🔧 Configuração**: Todos os exemplos mostram diferentes configurações
- **🌍 Idiomas**: Suporte para en, pt, es (alguns incluem fr, de)
- **📝 Parâmetros**: Interpolação de variáveis nas traduções
- **🔄 Fallback**: Comportamento de fallback entre idiomas
- **📊 Monitoramento**: Health checks, métricas, logging
- **⚡ Performance**: Benchmarks, otimizações, análise de memória
- **🔗 Integração**: Frameworks web populares do Go

## 🛠️ Customização

### Adicionando Novos Idiomas

```go
cfg, err := config.NewConfigBuilder().
    WithSupportedLanguages("en", "pt", "es", "fr", "de", "it").
    WithDefaultLanguage("en").
    // ... outras configurações
    Build()
```

### Criando Arquivos de Tradução

Estrutura recomendada:
```
translations/
├── en.json
├── pt.json
├── es.json
└── fr.json
```

Formato JSON:
```json
{
  "app": {
    "title": "My Application",
    "description": "Welcome to {{appName}}"
  },
  "user": {
    "greeting": "Hello {{name}}!"
  }
}
```

### Integrando com Seu Projeto

1. **Copie o exemplo mais próximo** ao seu caso de uso
2. **Adapte a configuração** para suas necessidades
3. **Crie seus arquivos de tradução** seguindo a estrutura
4. **Integre o provider** no seu código existente

## 📈 Métricas de Performance

### Benchmarks Típicos (valores aproximados)

- **Traduções simples**: ~50ns por operação
- **Com cache**: ~20ns por operação (cache hit)
- **Batch operations**: ~2.6K traduções/ms
- **Throughput concorrente**: 100K+ ops/segundo
- **Uso de memória**: <2MB para 50K traduções

### Otimizações Implementadas

- ✅ String pooling e interning
- ✅ Cache em memória com TTL
- ✅ Operações em lote otimizadas
- ✅ Lazy loading de idiomas
- ✅ Pool de objetos reutilizáveis

## 🔍 Troubleshooting

### Problemas Comuns

1. **"Provider not found"**
   - Verifique se o provider foi registrado corretamente
   - Confirme o nome usado no `CreateProvider()`

2. **"Translation key not found"**
   - Verifique se o arquivo de tradução existe
   - Confirme a estrutura das chaves aninhadas
   - Ative o fallback para idioma padrão

3. **Performance lenta**
   - Ative o cache na configuração
   - Use batch operations para múltiplas traduções
   - Considere usar providers otimizados

4. **Erro de arquivo não encontrado**
   - Verifique o caminho do diretório de traduções
   - Confirme o padrão de nomes dos arquivos
   - Verifique permissões de leitura

### Debug e Logging

```go
// Ativar logs detalhados
cfg, err := config.NewConfigBuilder().
    WithStrictMode(true).  // Erros mais rigorosos
    // ... outras configurações
    Build()

// Verificar saúde do provider
if err := provider.Health(ctx); err != nil {
    log.Printf("Provider unhealthy: %v", err)
}

// Estatísticas de uso
fmt.Printf("Total translations: %d\n", provider.GetTranslationCount())
fmt.Printf("Loaded languages: %v\n", provider.GetLoadedLanguages())
```

## 🤝 Contribuindo

Para adicionar novos exemplos:

1. Crie um novo diretório em `examples/`
2. Siga a estrutura dos exemplos existentes
3. Inclua documentação inline
4. Adicione uma seção neste README.md
5. Teste com diferentes cenários

## 📚 Próximos Passos

Após explorar os exemplos, consulte:

- **[RUN_EXAMPLES_DOC.md](RUN_EXAMPLES_DOC.md)** - Documentação completa do script de testes
- **[../README.md](../README.md)** - Documentação principal do módulo
- **[../NEXT_STEPS.md](../NEXT_STEPS.md)** - Roadmap e próximas features
- **[../hooks/](../hooks/)** - Sistema de hooks avançado
- **[../middlewares/](../middlewares/)** - Middlewares disponíveis
- **[../providers/](../providers/)** - Providers de dados

---

**🚀 Happy coding with i18n!**

*Última atualização: Agosto 2025*
