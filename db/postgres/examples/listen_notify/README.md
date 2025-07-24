# Exemplo de Listen/Notify

Este exemplo demonstra o poderoso sistema LISTEN/NOTIFY do PostgreSQL para comunicação assíncrona em tempo real.

## Funcionalidades Demonstradas

### 1. Listen/Notify Básico
- Configuração de listener
- Envio de notificações
- Recebimento assíncrono

### 2. Múltiplos Canais
- Listening em vários canais simultâneos
- Roteamento por canal
- Estatísticas por canal

### 3. Notificações com Payload
- Payloads JSON estruturados
- Payloads grandes (até 8KB)
- Parsing e processamento

### 4. Sistema de Chat Simples
- Simulação de chat em tempo real
- Múltiplos usuários
- Mensagens do sistema

### 5. Monitoramento de Mudanças
- Triggers automáticos
- Detecção de INSERT/UPDATE/DELETE
- Notificações em tempo real

## Conceitos Fundamentais

### Listen/Notify
- **Pub/Sub nativo**: Sistema publish/subscribe do PostgreSQL
- **Tempo real**: Latência ultra-baixa (< 1ms)
- **Assíncrono**: Não bloqueia operações
- **Escalável**: Múltiplos listeners por canal

### Payloads
- **Tamanho**: Até 8KB por notificação
- **Formato**: String livre (JSON recomendado)
- **Estruturado**: Dados complexos via JSON

## Como Executar

```bash
# Certifique-se de que o PostgreSQL está rodando
cd listen_notify/
go run main.go
```

## Exemplo de Saída

```
=== Exemplo de Listen/Notify ===

1. Conectando ao banco...

2. Exemplo: Listen/Notify básico...
   Fazendo LISTEN no canal 'test_channel'...
   🎧 Aguardando notificações...
   📤 Enviando notificação 1: 'Primeira mensagem'
   📨 Notificação recebida: Canal='test_channel', Payload='Primeira mensagem', PID=12345
   ✅ Notificação 1 enviada com sucesso

3. Exemplo: Múltiplos canais...
   Configurando LISTEN para 4 canais...
   📤 [orders] Enviando: 'Nova ordem #1001'
   📨 [orders] Notificação #1: 'Nova ordem #1001'
   📊 Estatísticas por canal:
   - orders: 3 notificações
   - users: 2 notificações
   - products: 1 notificação
   - system: 2 notificações

4. Exemplo: Notificações com payload...
   📤 Enviando payload 1 (65 bytes)...
   📨 Payload JSON recebido:
       Canal: json_channel
       Payload: {"event": "user_login", "user_id": 123, "timestamp": "2025-01-01T10:00:00Z"}
```

## Casos de Uso

### 1. Invalidação de Cache
```go
// Trigger para invalidar cache
_, err := conn.Exec(ctx, `
    CREATE OR REPLACE FUNCTION invalidate_cache()
    RETURNS TRIGGER AS $$
    BEGIN
        PERFORM pg_notify('cache_invalidate', 'users');
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;
`)
```

### 2. Notificações em Tempo Real
```go
// Listener para notificações
go func() {
    for {
        notification, err := conn.WaitForNotification(ctx, time.Second)
        if err == nil {
            // Processar notificação
            handleNotification(notification)
        }
    }
}()
```

### 3. Sincronização de Microserviços
```go
// Notificar outros serviços
_, err := conn.Exec(ctx, `
    NOTIFY order_events, 
    '{"event": "order_created", "order_id": 123, "service": "payment"}'
`)
```

## Vantagens do Listen/Notify

- **Performance**: Sem overhead de polling
- **Eficiência**: Comunicação direta via PostgreSQL
- **Confiabilidade**: Garantias ACID do PostgreSQL
- **Simplicidade**: API nativa e intuitiva
- **Escalabilidade**: Múltiplos listeners por canal

## Considerações Importantes

### Limitações
- Payload máximo: 8KB
- Não persistente: Notificações são perdidas se não houver listeners
- Escopo de sessão: LISTEN é por conexão

### Boas Práticas
- Use JSON para payloads estruturados
- Implemente timeout adequado
- Gerencie conexões de listening separadamente
- Monitore performance de triggers

### Uso em Produção
- Implemente reconexão automática
- Use connection pooling adequado
- Monitore latência e throughput
- Considere load balancing para múltiplos listeners

## Requisitos

- PostgreSQL rodando em `localhost:5432`
- Banco de dados `nexs_testdb`
- Usuário `nexs_user` com senha `nexs_password`
- Permissões para criar funções e triggers

## Integração com Aplicações

### WebSockets
```go
// Ponte entre Listen/Notify e WebSocket
func bridgeToWebSocket(conn postgres.IConn, ws *websocket.Conn) {
    for {
        notification, err := conn.WaitForNotification(ctx, time.Second)
        if err == nil {
            ws.WriteJSON(notification)
        }
    }
}
```

### Message Queues
```go
// Integração com sistemas de mensageria
func forwardToQueue(notification *postgres.Notification, queue MessageQueue) {
    queue.Publish(notification.Channel, notification.Payload)
}
```
