# Exemplos - Decimal Module

Esta pasta contém exemplos práticos demonstrando diferentes aspectos e casos de uso do módulo decimal.

## 📁 Estrutura dos Exemplos

```
examples/
├── README.md           # Este arquivo
├── basic/              # Exemplo básico de uso
│   ├── main.go
│   └── README.md
├── providers/          # Comparação entre providers
│   ├── main.go
│   └── README.md
└── hooks/              # Hooks customizados avançados
    ├── main.go
    └── README.md
```

## 🚀 Como Executar

### Pré-requisitos
```bash
# Certifique-se de estar na pasta raiz do projeto
cd /mnt/e/go/src/github.com/fsvxavier/nexs-lib

# Instalar dependências (se necessário)
go mod tidy
```

### Executar Exemplos
```bash
# Exemplo básico
cd decimal/examples/basic
go run main.go

# Comparação de providers
cd ../providers
go run main.go

# Hooks avançados
cd ../hooks
go run main.go
```

## 📚 Guia de Exemplos

### 1. [Exemplo Básico](./basic/)
**Ideal para**: Iniciantes, primeiro contato com a biblioteca

**O que você aprenderá**:
- ✅ Criar e configurar um manager decimal
- ✅ Operações aritméticas básicas
- ✅ Comparações entre decimais
- ✅ Operações em lote (batch)
- ✅ Sistema básico de hooks
- ✅ Tratamento de erros

**Tempo estimado**: 5-10 minutos

### 2. [Comparação de Providers](./providers/)
**Ideal para**: Escolher o provider certo para seu caso de uso

**O que você aprenderá**:
- ✅ Diferenças entre Cockroach e Shopspring
- ✅ Quando usar cada provider
- ✅ Configurações específicas por provider
- ✅ Comparações de precisão e performance
- ✅ Switching de providers em runtime
- ✅ Casos de uso práticos

**Tempo estimado**: 10-15 minutos

### 3. [Hooks Avançados](./hooks/)
**Ideal para**: Implementar observabilidade e validações customizadas

**O que você aprenderá**:
- ✅ Implementar hooks customizados
- ✅ Sistema de auditoria completo
- ✅ Validações de negócio
- ✅ Coleta de métricas
- ✅ Logging estruturado
- ✅ Combinação de múltiplos hooks

**Tempo estimado**: 15-20 minutos

## 🎯 Trilha de Aprendizado Recomendada

### Para Iniciantes
1. **[Exemplo Básico](./basic/)** - Comece aqui
2. **[Comparação de Providers](./providers/)** - Entenda as opções
3. **[Hooks Avançados](./hooks/)** - Funcionalidades avançadas

### Para Usuários Experientes
1. **[Comparação de Providers](./providers/)** - Escolha otimizada
2. **[Hooks Avançados](./hooks/)** - Customização completa
3. **[Exemplo Básico](./basic/)** - Revisão rápida

### Para Casos de Uso Específicos
- **Sistemas Financeiros**: Básico → Providers (Cockroach) → Hooks (Auditoria)
- **E-commerce**: Básico → Providers (Shopspring) → Hooks (Métricas)
- **Analytics**: Providers (Performance) → Básico (Batch) → Hooks (Monitoramento)

## 💡 Dicas Gerais

### Configuração do Ambiente
```bash
# Certificar que está usando Go 1.21+
go version

# Verificar dependências
go mod verify

# Executar testes (opcional)
go test ./...
```

### Debugging
```bash
# Executar com verbose
go run -v main.go

# Build para análise
go build -v main.go
```

### Performance
```bash
# Benchmarks (na pasta raiz do projeto)
go test -bench=. -benchmem ./decimal/...

# Profiling (exemplo)
go test -cpuprofile=cpu.prof -bench=.
```

## 🔧 Personalizando os Exemplos

### Modificar Configurações
```go
// Em qualquer exemplo, você pode alterar a configuração
cfg := &config.Config{
    ProviderName:    "shopspring", // ou "cockroach"
    MaxPrecision:    20,           // ajustar precisão
    MaxExponent:     10,           // limites de expoente
    MinExponent:     -6,
    DefaultRounding: "RoundHalfUp", // modo de arredondamento
    HooksEnabled:    true,          // habilitar hooks
    Timeout:         30,            // timeout em segundos
}
```

### Adicionar Seus Próprios Testes
```go
// Criar função de teste personalizada
func testMyUseCase() {
    manager := decimal.NewManager(nil)
    
    // Seus testes aqui
    a, _ := manager.NewFromString("123.45")
    b, _ := manager.NewFromString("67.89")
    
    result, _ := a.Add(b)
    fmt.Printf("Meu resultado: %s\n", result.String())
}

// Adicionar na função main
func main() {
    // ... exemplos existentes
    
    fmt.Println("\n=== Meu Caso de Uso ===")
    testMyUseCase()
}
```

## 📊 Comparação Rápida

| Aspecto | Básico | Providers | Hooks |
|---------|--------|-----------|-------|
| **Complexidade** | Baixa | Média | Alta |
| **Tempo** | 5-10 min | 10-15 min | 15-20 min |
| **Foco** | Uso geral | Performance | Observabilidade |
| **Pré-requisitos** | Nenhum | Básico | Providers |

## 🚨 Problemas Comuns

### Erro de Import
```bash
# Se houver erro de import, execute:
go mod tidy
go clean -modcache
```

### Erro de Compilação
```bash
# Verificar versão do Go
go version # Deve ser 1.21+

# Limpar cache
go clean -cache
```

### Performance Inesperada
- Verifique se está usando o provider correto
- Desabilite hooks se não precisar
- Ajuste precisão para suas necessidades

## 📞 Próximos Passos

Após executar os exemplos:

1. **Leia a documentação principal**: [`../README.md`](../README.md)
2. **Explore os testes**: [`../decimal_test.go`](../decimal_test.go)
3. **Veja os benchmarks**: [`../benchmark_test.go`](../benchmark_test.go)
4. **Consulte NEXT_STEPS**: [`../NEXT_STEPS.md`](../NEXT_STEPS.md)

## 🤝 Contribuindo

Quer adicionar um exemplo? Veja nosso guia de contribuição:

1. Crie uma pasta com nome descritivo
2. Inclua `main.go` e `README.md`
3. Documente o caso de uso específico
4. Adicione entrada neste README
5. Teste thoroughly antes de submeter

## 📝 Feedback

Se você:
- ✅ Encontrou os exemplos úteis
- ❌ Teve dificuldades com algum exemplo
- 💡 Tem sugestões de melhorias
- 🆕 Quer propor novos exemplos

Por favor, abra uma issue no repositório!

---

**Boa sorte explorando o módulo decimal! 🎉**
