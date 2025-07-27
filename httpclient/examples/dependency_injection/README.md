# Dependency Injection Examples

Este diretório contém exemplos práticos demonstrando padrões de injeção de dependências com o cliente HTTP nexs-lib.

## 📋 Exemplos Disponíveis

### 1. Injeção de Cliente HTTP em Serviços
Demonstra como injetar clientes HTTP em serviços usando padrão de dependency injection.

### 2. Cliente Nomeado Reutilizável
Mostra como criar e reutilizar clientes HTTP nomeados através do ClientManager.

### 3. Múltiplos Clientes para Diferentes Serviços
Exemplifica gerenciamento de múltiplos clientes HTTP para diferentes APIs externas.

### 4. Health Checks e Métricas
Demonstra monitoramento de saúde e coleta de métricas dos clientes injetados.

### 5. Shutdown Graceful
Mostra como fazer limpeza adequada de recursos ao finalizar a aplicação.

## 🚀 Como Executar

```bash
cd httpclient/examples/dependency_injection
go run main.go
```

## 🔧 Padrões de Dependency Injection

### Service Layer Pattern
- **O que é**: Encapsula lógica de negócio com dependências injetadas
- **Benefício**: Facilita testes unitários e mocking
- **Exemplo**: APIService com cliente HTTP injetado

### Client Manager Pattern
- **O que é**: Gerenciador centralizado de clientes HTTP
- **Benefício**: Reutilização de conexões e configurações
- **Exemplo**: Singleton manager com clientes nomeados

### Named Clients Pattern
- **O que é**: Clientes identificados por nome para diferentes serviços
- **Benefício**: Configurações específicas por serviço
- **Exemplo**: "api-client", "github-client", "payment-client"

## 📊 Benefícios

### Reutilização de Recursos:
- ✅ **Pool de Conexões**: Conexões TCP reutilizadas
- ✅ **Configuração Centralizada**: Uma configuração por serviço
- ✅ **Memory Efficiency**: Instâncias únicas de clientes
- ✅ **Performance**: Menos overhead de criação

### Testabilidade:
- ✅ **Mock Injection**: Fácil substituição para testes
- ✅ **Interface Based**: Contratos bem definidos
- ✅ **Isolation**: Testes independentes
- ✅ **Coverage**: Melhor cobertura de testes

## 🏗️ Como Usar

### Criando um Serviço com DI
```go
type APIService struct {
    httpClient interfaces.Client
}

func NewAPIService(client interfaces.Client) *APIService {
    return &APIService{
        httpClient: client,
    }
}

func (s *APIService) GetUser(ctx context.Context, id string) (*User, error) {
    resp, err := s.httpClient.Get(ctx, fmt.Sprintf("/users/%s", id))
    if err != nil {
        return nil, err
    }
    
    var user User
    if err := json.Unmarshal(resp.Body, &user); err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

### Usando Cliente Nomeado
```go
// Criar cliente nomeado (reutilizável)
client, err := httpclient.NewNamed(
    "my-api",
    interfaces.ProviderNetHTTP,
    "https://api.example.com",
)

// Reutilizar cliente existente
sameClient, err := httpclient.NewNamed(
    "my-api", // Mesmo nome retorna a mesma instância
    interfaces.ProviderNetHTTP,
    "https://api.example.com",
)
```

### Configuração de Aplicação
```go
type Application struct {
    userService    *UserService
    orderService   *OrderService
    clientManager  *httpclient.ClientManager
}

func NewApplication() (*Application, error) {
    manager := httpclient.GetManager()
    
    // Cliente para API de usuários
    userClient, err := httpclient.NewNamed(
        "user-api",
        interfaces.ProviderNetHTTP,
        "https://users.api.com",
    )
    if err != nil {
        return nil, err
    }
    
    // Cliente para API de pedidos
    orderClient, err := httpclient.NewNamed(
        "order-api",
        interfaces.ProviderNetHTTP,
        "https://orders.api.com",
    )
    if err != nil {
        return nil, err
    }
    
    return &Application{
        userService:   NewUserService(userClient),
        orderService:  NewOrderService(orderClient),
        clientManager: manager,
    }, nil
}
```

## 🔍 Gerenciamento de Clientes

### ClientManager Features:
- **Registry**: Registro centralizado de clientes nomeados
- **Lifecycle**: Gerenciamento de ciclo de vida
- **Health Monitoring**: Monitoramento de saúde
- **Metrics Collection**: Coleta de métricas agregadas

### Operações Disponíveis:
```go
manager := httpclient.GetManager()

// Listar clientes registrados
names := manager.ListClients()

// Obter cliente por nome
client, exists := manager.GetClient("my-api")

// Verificar saúde de todos os clientes
healthMap := manager.HealthCheck()

// Shutdown graceful de todos os clientes
err := manager.Shutdown()
```

## 📈 Métricas e Monitoramento

### Métricas por Cliente:
```go
metrics := client.GetMetrics()

fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
fmt.Printf("Successful requests: %d\n", metrics.SuccessfulRequests)
fmt.Printf("Failed requests: %d\n", metrics.FailedRequests)
fmt.Printf("Average latency: %v\n", metrics.AverageLatency)
fmt.Printf("Error rate: %.2f%%\n", metrics.ErrorRate)
```

### Health Checks:
```go
// Check individual client
healthy := client.IsHealthy()

// Check all clients
healthMap := manager.HealthCheck()
for name, healthy := range healthMap {
    fmt.Printf("Client %s: %v\n", name, healthy)
}
```

## 🧪 Testabilidade

### Interface Mocking:
```go
type MockClient struct{}

func (m *MockClient) Get(ctx context.Context, endpoint string) (*interfaces.Response, error) {
    return &interfaces.Response{
        StatusCode: 200,
        Body:       []byte(`{"id": 1, "name": "Test User"}`),
    }, nil
}

// Inject mock in tests
service := NewAPIService(&MockClient{})
```

### Test Setup:
```go
func TestUserService(t *testing.T) {
    mockClient := &MockClient{}
    service := NewUserService(mockClient)
    
    user, err := service.GetUser(context.Background(), "123")
    assert.NoError(t, err)
    assert.Equal(t, "Test User", user.Name)
}
```

## 🔗 Integração com Frameworks

### Gin Integration:
```go
func SetupRoutes(userService *UserService) *gin.Engine {
    r := gin.Default()
    
    r.GET("/users/:id", func(c *gin.Context) {
        user, err := userService.GetUser(c, c.Param("id"))
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, user)
    })
    
    return r
}
```

### Wire (Google) Integration:
```go
//go:build wireinject
// +build wireinject

func InitializeApplication() (*Application, error) {
    wire.Build(
        httpclient.NewNamed,
        NewUserService,
        NewOrderService,
        NewApplication,
    )
    return &Application{}, nil
}
```

## 🔧 Configurações Avançadas

### Cliente com Middleware:
```go
client, err := httpclient.NewNamed("api", provider, baseURL)
client.AddMiddleware(&LoggingMiddleware{})
client.AddMiddleware(&AuthMiddleware{token: token})
```

### Cliente com Configuração Customizada:
```go
config := &interfaces.Config{
    Timeout:         30 * time.Second,
    MaxIdleConns:    50,
    IdleConnTimeout: 90 * time.Second,
}

client, err := httpclient.NewNamedWithConfig("api", provider, config)
```

## 💡 Casos de Uso

### 1. Microserviços Architecture
```go
// Diferentes clientes para diferentes serviços
userClient := httpclient.NewNamed("user-service", ...)
orderClient := httpclient.NewNamed("order-service", ...)
paymentClient := httpclient.NewNamed("payment-service", ...)
```

### 2. Multi-tenant Applications
```go
// Cliente por tenant
tenantClient := httpclient.NewNamed(
    fmt.Sprintf("tenant-%s", tenantID),
    provider,
    fmt.Sprintf("https://%s.api.com", tenantID),
)
```

### 3. Environment-specific Clients
```go
// Cliente baseado no ambiente
var baseURL string
switch os.Getenv("ENV") {
case "production":
    baseURL = "https://api.prod.com"
case "staging":
    baseURL = "https://api.staging.com"
default:
    baseURL = "https://api.dev.com"
}

client := httpclient.NewNamed("api", provider, baseURL)
```

## 📚 Referências

- [Dependency Injection Patterns](https://martinfowler.com/articles/injection.html)
- [Go Wire - Dependency Injection](https://github.com/google/wire)
- [Testing with Mocks](https://blog.golang.org/using-go-modules)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## 🤝 Integração

Este padrão funciona bem com:
- **Web Frameworks**: Gin, Echo, Fiber
- **DI Containers**: Wire, Dig, FX
- **Testing**: Testify, GoMock
- **Monitoring**: Prometheus, OpenTelemetry
