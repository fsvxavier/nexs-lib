# Builder Pattern Examples

Este exemplo demonstra o uso avançado do Builder Pattern para construção fluente de erros.

## 🎯 Objetivo

Mostrar as capacidades avançadas do Builder Pattern:
- Construção fluente e flexível
- Configuração complexa de erros
- Integração com contexto
- Severidade e categorização
- Detalhes HTTP específicos
- Otimizações de performance

## 🚀 Como Executar

```bash
go run main.go
```

## 📝 Funcionalidades Demonstradas

### 1. Builder Simples
```go
err := domainerrors.NewBuilder().
    WithCode("USR001").
    WithMessage("User not found").
    WithType(string(types.ErrorTypeNotFound)).
    Build()
```

### 2. Builder Complexo
```go
err := domainerrors.NewBuilder().
    WithCode("API001").
    WithMessage("Request processing failed").
    WithType(string(types.ErrorTypeBadRequest)).
    WithSeverity(interfaces.Severity(types.SeverityHigh)).
    WithCategory(interfaces.CategoryTechnical).
    WithDetail("endpoint", "/api/v1/users").
    WithTag("api").
    WithStatusCode(400).
    WithHeader("Content-Type", "application/json").
    Build()
```

### 3. Integração com Context
- Extração automática de valores do contexto
- Propagação de request IDs
- Metadata contextual

### 4. Severidade e Categorização
- `SeverityCritical`, `SeverityHigh`, `SeverityMedium`, `SeverityLow`
- `CategoryBusiness`, `CategoryTechnical`, `CategoryInfrastructure`, `CategorySecurity`

### 5. Detalhes HTTP
- Status codes personalizados
- Headers HTTP customizados
- Rate limiting headers
- Retry-After headers

### 6. Padrões de Chaining
- Step-by-step building
- Fluent chaining
- Conditional building

### 7. Otimizações de Performance
- Object pooling
- Memory efficient operations
- Thread-safe operations
- High-frequency error creation

## 🔧 Estrutura do Código

- `simpleBuilderExample()` - Builder básico
- `complexBuilderExample()` - Builder avançado
- `builderWithContext()` - Integração com context
- `builderWithSeverityAndCategory()` - Classificação
- `builderWithHTTPDetails()` - Detalhes HTTP
- `builderChaining()` - Padrões de chaining
- `performanceOptimizedBuilder()` - Otimizações

## ⚡ Performance

- **Object Pooling**: Reutilização de objetos para reduzir GC pressure
- **Thread Safety**: Todas as operações são thread-safe
- **Memory Efficient**: Otimizado para uso mínimo de memória
- **High Throughput**: Suporta alta frequência de criação de erros

## 📊 Próximos Passos

Veja outros exemplos:
- [Error Stacking](../error-stacking/) - Empilhamento e wrapping
- [Validation](../validation/) - Erros de validação avançados
- [Factory Usage](../factory-usage/) - Uso de factories
