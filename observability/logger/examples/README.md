# Exemplos do Sistema de Logging

Esta pasta contém exemplos práticos demonstrando todas as funcionalidades do sistema de logging multi-provider.

## 📁 Estrutura dos Exemplos

```
examples/
├── README.md              # Este arquivo
├── default-provider/      # Uso do provider padrão (Zap)
├── basic/                 # Funcionalidades básicas
├── advanced/              # Cenários avançados
├── multi-provider/        # Comparação de providers
└── benchmark/             # Análise de performance
```

## 🚀 Execução Rápida

```bash
# Testa todos os exemplos
make test

# Executa exemplos específicos
make default        # Provider padrão
make basic          # Exemplo básico
make advanced       # Exemplo avançado
make multi-provider # Comparação de providers
make benchmark      # Análise de performance
```

## 📖 Guia de Exemplos

### 1. 🎯 Default Provider
**📁 `default-provider/`**
- **Objetivo**: Demonstrar o uso sem configuração
- **Foco**: Zap como provider padrão
- **Complexidade**: ⭐ Iniciante

```bash
cd default-provider
go run main.go
```

**O que você aprenderá:**
- Como usar o logger sem configuração
- Verificar qual provider está ativo
- Logging básico com provider padrão
- Por que Zap é o padrão escolhido

### 2. 📚 Basic Usage
**📁 `basic/`**
- **Objetivo**: Funcionalidades básicas de todos os providers
- **Foco**: Comparação básica entre providers
- **Complexidade**: ⭐⭐ Básico

```bash
cd basic
go run main.go
```

**O que você aprenderá:**
- Usar todos os três providers
- Campos estruturados
- Diferentes níveis de log
- Context-aware logging
- Switching entre providers

### 3. 🔧 Advanced Usage
**📁 `advanced/`**
- **Objetivo**: Cenários avançados e integração com serviços
- **Foco**: Aplicações reais
- **Complexidade**: ⭐⭐⭐ Intermediário

```bash
cd advanced
go run main.go
```

**O que você aprenderá:**
- Integração com serviços
- Context propagation
- Error handling estruturado
- Benchmarking inline
- Padrões de uso em produção

### 4. 🔄 Multi-Provider
**📁 `multi-provider/`**
- **Objetivo**: Comparação completa entre providers
- **Foco**: Análise comparativa
- **Complexidade**: ⭐⭐⭐ Intermediário

```bash
cd multi-provider
go run main.go
```

**O que você aprenderá:**
- Diferenças entre providers
- Configuração avançada
- Análise de performance
- Quando usar cada provider
- Switching em runtime

### 5. 📊 Benchmark
**📁 `benchmark/`**
- **Objetivo**: Análise detalhada de performance
- **Foco**: Métricas e otimização
- **Complexidade**: ⭐⭐⭐⭐ Avançado

```bash
cd benchmark
go run main.go
```

**O que você aprenderá:**
- Benchmarks detalhados
- Análise de memória
- Métricas de CPU
- Recomendações por cenário
- Otimização de performance

## 🎯 Roteiro de Aprendizado

### Para Iniciantes
1. **Comece com**: `default-provider/`
2. **Continue com**: `basic/`
3. **Explore**: `multi-provider/`

### Para Usuários Intermediários
1. **Revise**: `basic/`
2. **Aprofunde**: `advanced/`
3. **Compare**: `multi-provider/`
4. **Otimize**: `benchmark/`

### Para Usuários Avançados
1. **Analise**: `benchmark/`
2. **Customize**: `advanced/`
3. **Implemente**: Seus próprios cenários

## 🎨 Outputs de Exemplo

### Default Provider (Zap)
```json
{"level":"info","time":"2025-07-18T10:30:45Z","msg":"Usando provider padrão"}
```

### Basic Comparison
```
=== Testando Provider: zap ===
{"level":"info","time":"2025-07-18T10:30:45Z","msg":"Testando provider","provider":"zap"}

=== Testando Provider: slog ===
{"time":"2025-07-18T10:30:45Z","level":"INFO","msg":"Testando provider","provider":"slog"}

=== Testando Provider: zerolog ===
{"level":"info","time":"2025-07-18T10:30:45Z","message":"Testando provider","provider":"zerolog"}
```

### Advanced Service Integration
```json
{"level":"info","time":"2025-07-18T10:30:45Z","trace_id":"abc123","user_id":"user456","msg":"Criando usuário","nome":"João","email":"joao@email.com"}
```

### Benchmark Results
```
=== Performance Comparison ===
┌─────────┬──────────────┬─────────────┬──────────────┐
│ Provider│ Logs/Second  │ Memory (MB) │ CPU Usage    │
├─────────┼──────────────┼─────────────┼──────────────┤
│ zap     │ 242,156      │ 145         │ 12%          │
│ zerolog │ 174,823      │ 98          │ 8%           │
│ slog    │ 132,445      │ 167         │ 15%          │
└─────────┴──────────────┴─────────────┴──────────────┘
```

## 🛠️ Ferramentas Úteis

### Makefile Commands
```bash
# Testa todos os exemplos
make test

# Executa exemplos específicos
make default        # Default provider
make basic          # Basic usage
make advanced       # Advanced usage
make multi-provider # Multi-provider comparison
make benchmark      # Performance benchmark

# Executa todos os exemplos
make examples

# Verificação completa
make check
```

### Script de Teste
```bash
# Executa script de teste manual
chmod +x test_examples.sh
./test_examples.sh
```

## 📝 Estrutura dos Arquivos

Cada exemplo contém:
- **`main.go`**: Código principal do exemplo
- **`README.md`**: Documentação detalhada
- **`go.mod`**: Dependências (quando necessário)

## 🔧 Configuração

### Pré-requisitos
- Go 1.21 ou superior
- Dependências instaladas (`go mod tidy`)

### Instalação
```bash
# Na raiz do projeto
go mod tidy

# Testa se tudo está funcionando
make test
```

## 📊 Comparação Rápida

| Exemplo | Complexidade | Foco | Tempo |
|---------|--------------|------|-------|
| default-provider | ⭐ | Simplicidade | 2 min |
| basic | ⭐⭐ | Fundamentos | 5 min |
| advanced | ⭐⭐⭐ | Integração | 10 min |
| multi-provider | ⭐⭐⭐ | Comparação | 8 min |
| benchmark | ⭐⭐⭐⭐ | Performance | 15 min |

## 🎯 Recomendações

### Para Projetos Novos
1. **Comece com**: `default-provider/`
2. **Use**: Zap como padrão
3. **Referência**: `basic/` para funcionalidades

### Para Migração
1. **Analise**: `benchmark/` para escolher provider
2. **Implemente**: Padrões do `advanced/`
3. **Compare**: `multi-provider/` para validar escolha

### Para Otimização
1. **Execute**: `benchmark/` regularmente
2. **Monitore**: Métricas de performance
3. **Ajuste**: Configuração baseada nos resultados

## 📚 Documentação Adicional

- **[README.md](../README.md)**: Documentação principal
- **[USAGE.md](../USAGE.md)**: Guia de uso detalhado
- **[FINAL_SUMMARY.md](../FINAL_SUMMARY.md)**: Resumo completo

## 🆘 Suporte

Se você encontrar problemas:

1. **Verifique os logs**: Execute `make test` para ver se há erros
2. **Consulte a documentação**: Cada exemplo tem README detalhado
3. **Teste individualmente**: Execute cada exemplo separadamente
4. **Verifique dependências**: Execute `go mod tidy`

## 🎉 Próximos Passos

Depois de explorar os exemplos:

1. **Escolha seu provider**: Baseado nos benchmarks
2. **Implemente em seu projeto**: Use os padrões dos exemplos
3. **Configure para produção**: Otimize baseado nos resultados
4. **Monitore**: Implemente métricas de performance
