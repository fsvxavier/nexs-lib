# gRPC & Message Queue Example

Este exemplo demonstra uso avançado do tracer com gRPC e sistemas de mensageria.

## Recursos Demonstrados

- **gRPC Tracing**: Client e Server spans para gRPC calls
- **Message Queue Tracing**: Producer e Consumer spans
- **Context Propagation**: Entre gRPC services e message queues
- **Async Processing**: Background workers com tracing
- **Error Handling**: Retry logic e circuit breaker patterns
- **Performance Monitoring**: Latency, throughput e error rates

## Arquitetura

```
Client → gRPC API → Message Queue → Background Workers
                 ↓
            Database Operations
```

## Como Executar

```bash
# Instalar dependências (simuladas)
# go mod tidy

# Executar servidor
go run main.go

# Em outro terminal, executar cliente
go run client.go
```

## Componentes

### gRPC Server
- Processamento de requisições síncronas
- Validação e transformação de dados
- Publicação em message queue

### Message Queue
- Processamento assíncrono
- Delivery guarantees
- Dead letter queues

### Background Workers
- Processamento de jobs
- Batch operations
- Scheduled tasks
