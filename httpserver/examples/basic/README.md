# 🚀 Exemplo Básico - HTTP Server

Este exemplo demonstra o uso mais simples da biblioteca `nexs-lib/httpserver` com o provider Fiber.

## 📋 Funcionalidades

- ✅ Configuração mínima de servidor
- ✅ Registro de rota simples
- ✅ Provider Fiber (padrão)
- ✅ Resposta JSON básica

## 🎯 Objetivo

Demonstrar como criar um servidor HTTP funcional com o mínimo de código possível.

## 🔧 Como Executar

### Pré-requisitos
```bash
go mod tidy
```

### Execução
```bash
cd basic
go run main.go
```

### Teste
```bash
curl http://localhost:8080/
```

**Resposta esperada:**
```json
{
  "message": "Hello, World!",
  "status": "success"
}
```

## 📊 Arquitetura

```
HTTP Request → Fiber Router → JSON Response
```

## 💡 Conceitos Demonstrados

1. **Configuração Básica**: `config.NewBaseConfig()`
2. **Criação de Servidor**: `httpserver.CreateServerWithConfig()`
3. **Registro de Rota**: `server.RegisterRoute()`
4. **Inicialização**: `server.Start()`

## 🎓 Para Quem é Este Exemplo

- **Iniciantes** que querem entender o básico
- **Desenvolvedores** buscando implementação mínima
- **Prototipagem** rápida de APIs

## 🔗 Próximos Passos

Após dominar este exemplo, continue com:
1. `gin/` - Framework Gin com hooks
2. `hooks-basic/` - Conceitos de monitoramento
3. `middlewares-basic/` - Autenticação e logging

## 🏗️ Estrutura do Código

```go
// 1. Configuração
cfg := config.NewBaseConfig()

// 2. Criação do servidor
server, err := httpserver.CreateServerWithConfig("fiber", cfg)

// 3. Registro de rotas
server.RegisterRoute("GET", "/", handler)

// 4. Inicialização
server.Start(ctx)
```

## 📈 Performance

- **Overhead**: Mínimo (~1ms)
- **Memória**: ~5-10MB base
- **CPU**: <1% uso idle

## 🐛 Troubleshooting

### Porta em uso
```bash
# Verificar processos na porta 8080
lsof -i :8080

# Matar processo se necessário
kill -9 <PID>
```

### Dependências
```bash
# Baixar dependências
go mod download
```

---

*Este é o exemplo mais simples da biblioteca nexs-lib/httpserver*
