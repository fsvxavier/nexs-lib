# Validator Schema Examples

Este diretório contém exemplos abrangentes demonstrando todas as funcionalidades do pacote `validator/schema`.

## 🚀 Como Executar

```bash
go run main.go
```

## 📚 Exemplos Incluídos

### 1. **Basic Validation Examples**
- Validação de campos obrigatórios
- Validação de email
- Validação de comprimento de strings

### 2. **Struct Validation Examples**
- Validação de structs com tags
- Casos válidos e inválidos
- Tratamento de múltiplos erros por campo

### 3. **Fluent Builder API Examples**
- API fluente para construção de regras
- Validação de strings complexas
- Validação de números com ranges
- Validação de datas com períodos

### 4. **JSON Schema Validation Examples**
- Validação de objetos com schemas JSON
- Schemas com formatos customizados
- Tratamento de erros de schema

### 5. **Custom Validation Rules Examples**
- Regras de validação personalizadas
- Validadores de formato customizados
- Implementação de lógica de negócio específica

### 6. **Format Validators Examples**
Demonstração de todos os validadores de formato built-in:
- `date_time` - Diversos formatos de data/hora
- `iso_8601_date` - Datas ISO 8601
- `text_match` - Texto com letras, underscore e espaços
- `text_match_with_number` - Texto incluindo números
- `strong_name` - Identificadores válidos
- `json_number` - Números JSON
- `decimal` - Decimais genéricos
- `decimal_by_factor_of_8` - Decimais com 8 casas
- `empty_string` - Strings vazias

### 7. **Advanced Schema Validation**
- **Validação Condicional**: Uso de `if/then/else`
- **Objetos Aninhados**: Schemas complexos com múltiplos níveis
- **Validação de Arrays**: Constraints de tamanho, unicidade, tipos

### 8. **Context and Performance Examples**
- **Timeouts**: Validação com tempo limite
- **Cancelamento**: Interrupção de validações
- **Reutilização**: Performance com validators reutilizáveis
- **Benchmarks**: Medição de tempo de execução

### 9. **Domain Error Integration Examples**
- Integração com sistema de erros do nexs-lib
- Análise detalhada de resultados
- Métodos utilitários de ValidationResult

### 10. **Batch Validation Examples**
- Validação em lote de múltiplos registros
- Estatísticas de validação
- Performance em processamento bulk
- Relatórios detalhados de erros

## 🎯 Casos de Uso Demonstrados

- **APIs REST**: Validação de entrada de dados
- **Processamento em Lote**: Validação de múltiplos registros
- **Validação Condicional**: Regras dependentes de contexto
- **Formatos Personalizados**: Validação de dados específicos do domínio
- **Performance**: Otimização para alta throughput

## 📊 Performance

Os exemplos incluem medições de performance que mostram:

- **Validação Individual**: ~50µs por validação
- **Validação em Batch**: 10+ validações por milissegundo
- **Throughput**: 1000 validações em ~100ms
- **Memory**: Uso eficiente com reutilização de validators

## 🔧 Extensão

Use estes exemplos como base para:

1. **Implementar validações específicas** do seu domínio
2. **Integrar com frameworks web** (Gin, Echo, Fiber, etc.)
3. **Criar pipelines de validação** em batch
4. **Desenvolver APIs robustas** com validação completa
5. **Otimizar performance** em aplicações de alta escala

## 📝 Notas

- Todos os exemplos são funcionais e testados
- Códigos comentados explicam conceitos importantes
- Estrutura modular permite estudo individual de cada tópico
- Exemplos progressivos do básico ao avançado
