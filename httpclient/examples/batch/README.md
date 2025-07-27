# Batch Operations Examples

Este diretório contém exemplos práticos demonstrando operações em lote (batch) com o cliente HTTP nexs-lib.

## 📋 Exemplos Disponíveis

### 1. Batch Simples
Demonstra como agrupar múltiplas requisições e executá-las em paralelo para melhor performance.

### 2. Batch Complexo
Mostra requisições para diferentes endpoints com processamento personalizado dos resultados.

### 3. Requisições Customizadas
Exemplifica batch com diferentes métodos HTTP (GET, POST, PUT, DELETE) e configurações específicas.

### 4. Comparação de Performance
Compara performance entre requisições sequenciais vs. batch paralelo.

### 5. Tratamento de Erros
Demonstra como lidar com erros em operações batch e análise de resultados.

### 6. Batch Grande
Testa performance com grande volume de requisições simultâneas.

## 🚀 Como Executar

```bash
cd httpclient/examples/batch
go run main.go
```

## 🔧 Funcionalidades Batch

### Paralelização
- **O que é**: Execução simultânea de múltiplas requisições HTTP
- **Benefício**: Reduz tempo total de execução significativamente
- **Exemplo**: 10 requisições em ~100ms vs. 1000ms sequencial

### Gerenciamento de Recursos
- **Pool de Conexões**: Reutilização eficiente de conexões TCP
- **Controle de Concorrência**: Evita sobrecarga do servidor
- **Memory Management**: Otimização do uso de memória

### Tratamento de Erros
- **Isolamento**: Falha em uma requisição não afeta as outras
- **Retry Inteligente**: Apenas requisições falhadas são re-executadas
- **Estatísticas**: Relatórios detalhados de sucesso vs. falha

## 📊 Performance

### Benefícios do Batch:
- ✅ **Paralelização**: 5-10x mais rápido que requisições sequenciais
- ✅ **Eficiência de Rede**: Melhor utilização de banda e conexões
- ✅ **Timeout Granular**: Controle individual de timeout por requisição
- ✅ **Balanceamento**: Distribuição inteligente de carga

### Cenários Ideais:
- 🎯 Fetch de dados de múltiplas APIs
- 🎯 Validação de múltiplos recursos
- 🎯 Operações CRUD em lote
- 🎯 Sincronização de dados

## 🏗️ Como Usar

### Batch Básico
```go
batch := client.Batch()
batch.Add("GET", "/users/1", nil)
batch.Add("GET", "/users/2", nil)
batch.Add("GET", "/users/3", nil)

results, err := batch.Execute(ctx)
```

### Batch com Configurações
```go
batch := client.Batch()
batch.Add("POST", "/posts", map[string]interface{}{
    "title": "New Post",
    "body": "Content here",
})
batch.Add("PUT", "/posts/1", updateData)
batch.Add("DELETE", "/posts/2", nil)

results, err := batch.Execute(ctx)
```

### Processamento de Resultados
```go
for i, response := range results {
    if response == nil {
        fmt.Printf("Request %d failed: no response\n", i+1)
        continue
    }
    
    if response.StatusCode >= 400 {
        fmt.Printf("Request %d error: %d\n", i+1, response.StatusCode)
        continue
    }
    
    // Process successful response
    fmt.Printf("Request %d success: %d bytes\n", i+1, len(response.Body))
}
```

## 🔍 Análise de Performance

O exemplo inclui métricas detalhadas:

- **Tempo Total**: Duração da operação batch completa
- **Tempo Médio**: Tempo médio por requisição
- **Taxa de Sucesso**: Percentual de requisições bem-sucedidas
- **Throughput**: Requisições por segundo
- **Comparação**: Batch vs. sequencial

## 📈 Otimizações

### Configurações Recomendadas:
```go
// Para APIs rápidas
config := &interfaces.Config{
    Timeout:           5 * time.Second,
    MaxIdleConns:      50,
    IdleConnTimeout:   30 * time.Second,
    DisableKeepAlives: false,
}

// Para operações em lote grandes
config := &interfaces.Config{
    Timeout:           30 * time.Second,
    MaxIdleConns:      100,
    IdleConnTimeout:   60 * time.Second,
    DisableKeepAlives: false,
}
```

## 🛡️ Tratamento de Erros

### Tipos de Erro:
- **Erro de Rede**: Conexão falhou, timeout
- **Erro HTTP**: Status 4xx/5xx
- **Erro de Parse**: JSON inválido, formato incorreto

### Estratégias:
- **Fail Fast**: Para operações críticas
- **Best Effort**: Para operações que podem ter falhas parciais
- **Retry Seletivo**: Apenas para erros temporários

## 🔗 Endpoints de Teste

Os exemplos utilizam [JSONPlaceholder](https://jsonplaceholder.typicode.com/):

- `/users` - Lista de usuários
- `/posts` - Posts do blog
- `/comments` - Comentários
- `/albums` - Álbuns de fotos
- `/photos` - Fotos

## 🤝 Integração

### Com Middleware:
```go
client.AddMiddleware(&LoggingMiddleware{})
client.AddMiddleware(&AuthMiddleware{token: "abc123"})

batch := client.Batch() // Middleware aplicado automaticamente
```

### Com Hooks:
```go
client.AddHook(&MetricsHook{})
client.AddHook(&AuditHook{})

batch := client.Batch() // Hooks chamados para cada requisição
```

## 💡 Casos de Uso

### 1. Sincronização de Dados
```go
// Buscar dados de múltiplas fontes
batch.Add("GET", "/api/users", nil)
batch.Add("GET", "/api/products", nil) 
batch.Add("GET", "/api/orders", nil)
```

### 2. Validação em Lote
```go
// Validar múltiplos IDs
for _, id := range userIDs {
    batch.Add("HEAD", fmt.Sprintf("/users/%d", id), nil)
}
```

### 3. Operações CRUD
```go
// Criar múltiplos recursos
for _, item := range items {
    batch.Add("POST", "/items", item)
}
```

## 📚 Referências

- [Concurrent Programming in Go](https://blog.golang.org/concurrency-is-not-parallelism)
- [HTTP Connection Pooling](https://developer.mozilla.org/en-US/docs/Web/HTTP/Connection_management_in_HTTP_1.x)
- [Best Practices for API Clients](https://github.com/microsoft/api-guidelines)
