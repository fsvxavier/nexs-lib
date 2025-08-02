# Exemplo Básico - Sistema de Hooks

Este exemplo demonstra o uso básico do sistema de hooks da biblioteca `nexs-lib/httpserver`.

## 📋 O que este exemplo demonstra

- **Hooks de Ciclo de Vida**: StartHook e StopHook para monitorar início e parada do servidor
- **Hook de Requisições**: RequestHook para rastrear todas as requisições HTTP
- **Hook de Erros**: ErrorHook para capturar e monitorar erros
- **Integração com Gin**: Como integrar os hooks com o framework Gin

## 🚀 Como executar

```bash
cd httpserver/examples/hooks-basic
go run main.go
```

O servidor iniciará na porta 8080.

## 🧪 Testando os endpoints

```bash
# Página inicial
curl http://localhost:8080/

# Health check
curl http://localhost:8080/health

# Lista de usuários (com delay simulado)
curl http://localhost:8080/users

# Simular erro
curl http://localhost:8080/error

# Ver métricas dos hooks
curl http://localhost:8080/metrics
```

## 📊 Funcionalidades demonstradas

### 1. Hooks de Ciclo de Vida
- **StartHook**: Registra quando o servidor inicia
- **StopHook**: Registra quando o servidor para e calcula estatísticas de uptime

### 2. Hook de Requisições
- Conta total de requisições
- Rastreia requisições ativas
- Calcula tamanho médio das requisições
- Monitora pico de requisições concorrentes

### 3. Hook de Erros
- Captura e categoriza erros
- Alerta quando limite de erros é excedido
- Mantém histórico de erros recentes

### 4. Métricas
- Número total de hooks registrados
- Contador de inicializações do servidor
- Contador de paradas do servidor
- Total de requisições processadas
- Requisições ativas no momento

## 🔍 Logs produzidos

Durante a execução, você verá logs detalhados mostrando:
- Registro de hooks
- Início do servidor
- Cada requisição recebida
- Erros simulados
- Métricas de shutdown

## 📖 Conceitos importantes

### Hook Manager
O `HookManager` é responsável por:
- Registrar hooks
- Gerenciar o ciclo de vida dos hooks
- Fornecer interface unificada para todos os hooks

### Integração com Middleware
Os hooks são integrados através de middleware personalizado que:
- Intercepta requisições antes do processamento
- Chama hooks apropriados em cada etapa
- Captura informações de resposta

### Thread Safety
Todos os hooks são thread-safe e podem ser usados em ambientes concorrentes sem problemas.

## 🎯 Próximos passos

Após entender este exemplo básico, você pode explorar:
- [Exemplo com Middlewares](../middlewares-basic/) - Adiciona autenticação e logging
- [Exemplo Avançado](../advanced/) - Combina hooks e middlewares com configurações complexas
- [Exemplos específicos por framework](../) - Gin, Echo, FastHTTP, etc.
