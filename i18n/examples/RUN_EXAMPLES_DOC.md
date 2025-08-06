# 🚀 Script run_examples.sh - Documentação

## Visão Geral

O script `run_examples.sh` é uma ferramenta automatizada para testar todos os exemplos do módulo i18n. Ele executa cada exemplo em sequência, verifica a sintaxe, compila os projetos e relata os resultados de forma organizada.

## ✨ Funcionalidades

- **✅ Teste Automatizado**: Executa todos os 9 exemplos automaticamente
- **🔧 Configuração Automática**: Configura go.mod com dependências locais
- **📊 Relatório Detalhado**: Mostra estatísticas de sucesso/falha
- **🧹 Limpeza Automática**: Remove arquivos temporários após execução
- **⚡ Otimizado**: Usa timeouts e testes de compilação para exemplos web
- **🎨 Interface Colorida**: Output com cores para melhor visualização

## 🚀 Como Usar

### Execução Básica

```bash
cd /path/to/nexs-lib/i18n/examples
./run_examples.sh
```

### Opções Disponíveis

```bash
# Mostrar ajuda
./run_examples.sh --help

# Execução silenciosa (sem output detalhado)
./run_examples.sh --quiet

# Execução em modo verbose (debug)
./run_examples.sh --verbose
```

## 📋 Tipos de Testes

### 1. Exemplos Básicos
- **basic_json**: Teste de execução completa (5s timeout)
- **basic_yaml**: Teste de execução completa (5s timeout)

### 2. Exemplos Avançados
- **advanced**: Teste com hooks e configurações avançadas (10s timeout)
- **middleware_demo**: Teste de middlewares customizados (8s timeout)
- **performance_demo**: Benchmarks e otimizações (15s timeout)

### 3. Aplicações Web
- **web_app_gin**: Teste de compilação (Gin framework)
- **api_rest_echo**: Teste de compilação (Echo framework)

### 4. Microserviços
- **microservice**: Teste de compilação da API HTTP

### 5. Ferramentas CLI
- **cli_tool**: Teste não-interativo com comando `stats`

## 🔧 Como Funciona

### 1. Setup Automático
```bash
# O script automaticamente:
- Verifica se Go está instalado
- Verifica dependências (curl, lsof)
- Configura ambiente de teste
```

### 2. Configuração de Módulos
```bash
# Para cada exemplo, o script:
- Cria go.mod se necessário
- Adiciona replace para módulo local
- Instala dependências específicas
- Executa go mod tidy
```

### 3. Testes Diferenciados

#### Exemplos Normais
- Verifica sintaxe com `go vet`
- Executa com timeout configurado
- Captura tempo de execução

#### Aplicações Web
- Verifica sintaxe com `go vet`
- Testa apenas compilação (evita conflitos de porta)
- Instala dependências específicas (Gin/Echo)

#### CLI Tools
- Executa comando específico não-interativo
- Timeout reduzido para evitar interação manual

### 4. Limpeza Automática
```bash
# Após cada exemplo:
- Remove go.mod criado temporariamente
- Remove go.sum gerado
- Mata processos órfãos
```

## 📊 Interpretando os Resultados

### ✅ Status de Sucesso
- **SUCCESS**: Exemplo executado com sucesso
- **SUCCESS (compilation test)**: Aplicação web compilou corretamente

### ❌ Status de Erro
- **SYNTAX_ERROR**: Erro de sintaxe no código
- **TIMEOUT_OR_ERROR**: Timeout ou erro durante execução
- **BUILD_ERROR**: Falha na compilação
- **PORT_IN_USE**: Porta necessária já está em uso
- **DIRECTORY_NOT_FOUND**: Diretório do exemplo não encontrado
- **MAIN_GO_NOT_FOUND**: Arquivo main.go não encontrado

### 📈 Relatório Final

```bash
==================================
📊 Execution Summary
==================================

Total examples: 9
Successful: 9
Failed: 0

🎉 All examples executed successfully!
```

## 🐛 Solução de Problemas

### Erro: "Go não está instalado"
```bash
# Instale Go:
sudo apt install golang-go  # Ubuntu/Debian
brew install go             # macOS
```

### Erro: "SYNTAX_ERROR"
```bash
# Verifique manualmente:
cd exemplo_com_erro
go vet ./...
go build main.go
```

### Erro: "PORT_IN_USE"
```bash
# Encontre e mate processos na porta:
lsof -Pi :8080 -sTCP:LISTEN
kill <PID>
```

### Exemplo específico falhando
```bash
# Teste manual:
cd exemplo_específico
go mod init test_module
echo "replace github.com/fsvxavier/nexs-lib => ../../.." >> go.mod
go mod edit -require github.com/fsvxavier/nexs-lib@v0.0.0
go mod tidy
go run main.go
```

## 🔍 Debug e Desenvolvimento

### Modo Verbose
```bash
# Para debug detalhado:
./run_examples.sh --verbose
```

### Teste Manual de Função Específica
```bash
# Edite o script e comente outros testes:
# Mantenha apenas o exemplo problemático
```

### Modificar Timeouts
```bash
# No script, ajuste os valores:
run_example "advanced" "Advanced Features" 20  # Aumentar timeout
```

## 📝 Estrutura de Logs

### Durante Execução
```bash
🔸 Running Basic JSON...        # Início do teste
✅ Basic JSON - SUCCESS (1s)    # Resultado positivo
❌ Basic YAML - SYNTAX_ERROR (0s) # Resultado negativo
```

### Relatório de Falhas
```bash
Failed examples:
  - Example Name

💡 Tips for failures:
  - SYNTAX_ERROR: Check Go syntax with 'go vet'
  - TIMEOUT_OR_ERROR: May need manual interaction
```

## 🚀 Integração CI/CD

O script é ideal para integração em pipelines:

```yaml
# .github/workflows/test.yml
- name: Test i18n Examples
  run: |
    cd i18n/examples
    ./run_examples.sh
```

## 📋 Requisitos do Sistema

- **Go 1.21+**: Linguagem de programação
- **Bash**: Shell Unix/Linux
- **curl**: Opcional, para testes web (se disponível)
- **lsof**: Opcional, para verificação de portas (se disponível)
- **timeout**: Comando para timeouts (geralmente disponível)

## 🤝 Contribuindo

Para melhorar o script:

1. **Adicionar novos exemplos**: Edite a função `main()`
2. **Modificar timeouts**: Ajuste os valores por tipo de exemplo
3. **Melhorar detecção de erros**: Adicione novos status de erro
4. **Otimizar performance**: Reduza tempos de setup/cleanup

---

**📅 Última atualização**: Agosto 2025  
**👨‍💻 Mantido por**: fsvxavier  
**🔗 Repositório**: nexs-lib/i18n
