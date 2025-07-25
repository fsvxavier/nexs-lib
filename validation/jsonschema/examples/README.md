# Exemplos de Uso - JSON Schema Validation

Esta pasta contém exemplos práticos demonstrando como usar a biblioteca de validação JSON Schema em diferentes cenários.

## 📁 Estrutura dos Exemplos

### 1. [Basic](./basic/) - Uso Básico
Demonstra funcionalidades básicas da biblioteca:
- Validação simples com configuração padrão
- Uso de diferentes providers
- Validação com arquivos de schema
- Validações múltiplas com schemas registrados

**Como executar:**
```bash
cd examples/basic
go run basic_validation.go
```

### 2. [Hooks](./hooks/) - Sistema de Hooks
Mostra como usar o sistema de hooks para estender funcionalidades:
- Hooks de pré-validação (normalização de dados, logging)
- Hooks de pós-validação (enriquecimento de erros, resumos)
- Hooks de erro (filtros, notificações)
- Pipeline completo com múltiplos hooks

**Como executar:**
```bash
cd examples/hooks
go run hooks_example.go
```

### 3. [Checks](./checks/) - Validações Customizadas
Exemplifica o uso de checks adicionais:
- Validação de campos obrigatórios
- Verificação de constraints de enum
- Validação de datas
- Combinação de múltiplos checks

**Como executar:**
```bash
cd examples/checks
go run checks_example.go
```

### 4. [Providers](./providers/) - Diferentes Engines
Compara o uso de diferentes providers de validação:
- kaptinlin/jsonschema (padrão, performance)
- xeipuuv/gojsonschema (compatibilidade)
- santhosh-tekuri/jsonschema (funcionalidades avançadas)
- Comparação de performance entre providers
- Formatos customizados

**Como executar:**
```bash
cd examples/providers
go run providers_example.go
```

### 5. [Migration](./migration/) - Migração do Código Legacy
Demonstra como migrar do `_old/validator` para o novo módulo:
- Compatibilidade com funções legacy
- Migração para a nova API
- Funcionalidades aprimoradas disponíveis
- Comparação lado a lado

**Como executar:**
```bash
cd examples/migration
go run migration_example.go
```

### 6. [Real World](./real_world/) - Cenários Reais
Exemplos práticos de uso em aplicações reais:
- Validação de requisições de API REST
- Validação de arquivos de configuração
- Validação de produtos e-commerce
- Registro de usuários com regras complexas

**Como executar:**
```bash
cd examples/real_world
go run real_world_examples.go
```

## 🚀 Executando Todos os Exemplos

Para executar todos os exemplos de uma vez:

```bash
# Na raiz do projeto
./run_all_examples.sh
```

Ou execute individualmente:

```bash
# Exemplo básico
cd examples/basic && go run basic_validation.go

# Hooks
cd examples/hooks && go run hooks_example.go

# Checks customizados
cd examples/checks && go run checks_example.go

# Diferentes providers
cd examples/providers && go run providers_example.go

# Migração
cd examples/migration && go run migration_example.go

# Cenários reais
cd examples/real_world && go run real_world_examples.go
```

## 📋 Pré-requisitos

- Go 1.23 ou superior
- Módulo `github.com/fsvxavier/nexs-lib` configurado
- Dependências instaladas (`go mod download`)

## 🔧 Configuração

Certifique-se de que o módulo está corretamente configurado:

```bash
# Na raiz do projeto nexs-lib
go mod tidy
go mod download
```

## 📖 Guia de Aprendizado

Recomendamos seguir os exemplos nesta ordem para melhor compreensão:

1. **Basic** - Entenda os conceitos fundamentais
2. **Providers** - Explore diferentes engines de validação
3. **Hooks** - Aprenda a estender funcionalidades
4. **Checks** - Implemente validações customizadas
5. **Migration** - Veja como migrar código existente
6. **Real World** - Aplique em cenários práticos

## 🎯 Casos de Uso por Exemplo

| Exemplo | Melhor Para |
|---------|-------------|
| **Basic** | Primeiros passos, validações simples |
| **Hooks** | Logging, normalização, monitoramento |
| **Checks** | Regras de negócio específicas |
| **Providers** | Otimização de performance, compatibilidade |
| **Migration** | Atualização de código legacy |
| **Real World** | APIs, configs, e-commerce, registros |

## 💡 Dicas de Performance

### Escolha do Provider
- **kaptinlin/jsonschema**: Melhor performance geral
- **xeipuuv/gojsonschema**: Máxima compatibilidade
- **santhosh-tekuri/jsonschema**: Funcionalidades avançadas

### Otimizações
- Registre schemas frequentemente usados uma única vez
- Use hooks apenas quando necessário
- Configure error limiting para prevenir ataques
- Cache validadores em aplicações de alta performance

## 🔍 Debugging

Para debugar problemas de validação:

1. **Use logging hooks** para ver dados de entrada
2. **Ative validation summary** para métricas
3. **Configure error enrichment** para contexto adicional
4. **Teste com providers diferentes** para compatibilidade

## 📚 Recursos Adicionais

- [Documentação Completa](../README.md)
- [Próximos Passos](../NEXT_STEPS.md)
- [JSON Schema Specification](https://json-schema.org/)
- [Guia de Performance](../docs/performance.md) (quando disponível)

## 🤝 Contribuindo

Encontrou um bug nos exemplos ou tem uma ideia para um novo exemplo?

1. Abra uma issue descrevendo o problema/sugestão
2. Faça um fork do repositório
3. Crie um branch para seu exemplo (`git checkout -b example/nome-do-exemplo`)
4. Adicione o exemplo seguindo a estrutura existente
5. Teste o exemplo (`go run exemplo.go`)
6. Submeta um pull request

## 📄 Licença

Estes exemplos estão sob a mesma licença do projeto principal (MIT).
