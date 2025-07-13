# Microservices Example

Este exemplo demonstra como usar o tracer em um ambiente de microserviços com propagação de contexto entre serviços.

## Arquitetura

```
User → Gateway → Auth Service → User Service → Database
                              → Order Service → Payment Service
```

## Recursos Demonstrados

- **Context Propagation**: Propagação de trace entre múltiplos serviços
- **Different Span Kinds**: Server, Client, Internal, Producer, Consumer
- **Service Communication**: HTTP, gRPC, Message Queue
- **Error Handling**: Retry logic, circuit breaker, fallback
- **Business Metrics**: Performance SLAs, error rates, latency percentiles

## Como Executar

```bash
# Executar todos os serviços
go run .

# Testar endpoints
curl http://localhost:8081/api/user/123
curl http://localhost:8081/api/orders/456
curl http://localhost:8081/health
```

## Serviços

### Gateway (Port 8081)
- Entry point para todas as requisições
- Load balancing e routing
- Authentication middleware

### Auth Service (Port 8082)
- Validação de tokens
- User permission checks
- Session management

### User Service (Port 8083)
- User profile management
- CRUD operations
- Database integration

### Order Service (Port 8084)
- Order processing
- Integration with payment service
- Inventory checks

### Payment Service (Port 8085)
- Payment processing
- External payment gateway integration
- Transaction logging
