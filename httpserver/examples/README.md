# HTTP Server Examples

Este diretório contém exemplos práticos de uso da biblioteca HTTP Server.

## Estrutura dos Exemplos

### 📁 basic/
Exemplo básico demonstrando:
- Criação simples de servidor
- Registro de rota básica
- Uso do provider Fiber padrão

**Como executar:**
```bash
cd examples/basic
go run main.go
```

**Teste:**
```bash
curl http://localhost:8080/
```

### 📁 advanced/
Exemplo avançado demonstrando:
- Configuração detalhada com Builder pattern
- Observer customizado para logging
- Middleware de logging
- Múltiplas rotas com parâmetros
- Endpoint de health check
- Graceful shutdown com sinais

**Como executar:**
```bash
cd examples/advanced
go run main.go
```

**Teste:**
```bash
# Rota básica
curl http://localhost:8080/hello

# Rota com parâmetro
curl http://localhost:8080/hello/world

# Health check
curl http://localhost:8080/health
```

## Conceitos Demonstrados

### 1. **Provider Independence**
Ambos os exemplos usam o provider Fiber, mas podem facilmente ser trocados para outros providers (quando implementados) apenas mudando o nome do provider na criação do servidor.

### 2. **Observer Pattern**
O exemplo avançado mostra como implementar um observer customizado para logging de eventos do ciclo de vida do servidor.

### 3. **Middleware**
Demonstra como registrar middleware customizado que funciona nativamente com o provider escolhido.

### 4. **Configuration Builder**
Mostra o uso do padrão Builder para configuração flexível e validada.

### 5. **Type Safety**
Os handlers são tipados especificamente para o provider (fiber.Ctx), garantindo type safety em tempo de compilação.

### 6. **Statistics**
O endpoint /health demonstra como acessar estatísticas em tempo real do servidor.

## Executando os Exemplos

### Pré-requisitos
```bash
go mod tidy
```

### Executar Exemplo Básico
```bash
cd examples/basic
go run main.go
```

### Executar Exemplo Avançado
```bash
cd examples/advanced
go run main.go
```

## Próximos Exemplos (Planejados)

- [ ] **nethttp/** - Exemplo usando provider net/http
- [ ] **gin/** - Exemplo usando provider Gin (quando implementado)
- [ ] **echo/** - Exemplo usando provider Echo (quando implementado)
- [ ] **middleware/** - Exemplo focado em middleware customizado
- [ ] **tls/** - Exemplo com HTTPS/TLS
- [ ] **metrics/** - Exemplo com métricas Prometheus
