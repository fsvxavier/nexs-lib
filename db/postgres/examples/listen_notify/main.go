package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
)

func main() {
	fmt.Println("=== Exemplo de Listen/Notify ===")

	// Configuração da conexão
	connectionString := "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Conectar ao banco
	fmt.Println("\n1. Conectando ao banco...")
	conn, err := postgres.Connect(ctx, connectionString)
	if err != nil {
		log.Printf("💡 Exemplo de Listen/Notify seria executado com banco real: %v", err)
		demonstrateListenNotifyConceptually()
		return
	}
	defer conn.Close(ctx)

	// 2. Exemplo: Listen/Notify básico
	fmt.Println("\n2. Exemplo: Listen/Notify básico...")
	if err := demonstrateBasicListenNotify(ctx, conn); err != nil {
		log.Printf("Erro no exemplo básico: %v", err)
	}

	// 3. Exemplo: Múltiplos canais
	fmt.Println("\n3. Exemplo: Múltiplos canais...")
	if err := demonstrateMultipleChannels(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de múltiplos canais: %v", err)
	}

	// 4. Exemplo: Notificações com payload
	fmt.Println("\n4. Exemplo: Notificações com payload...")
	if err := demonstrateNotificationsWithPayload(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de payload: %v", err)
	}

	// 5. Exemplo: Sistema de chat simples
	fmt.Println("\n5. Exemplo: Sistema de chat simples...")
	if err := demonstrateSimpleChat(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de chat: %v", err)
	}

	// 6. Exemplo: Monitoramento de mudanças
	fmt.Println("\n6. Exemplo: Monitoramento de mudanças...")
	if err := demonstrateChangeMonitoring(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de monitoramento: %v", err)
	}

	fmt.Println("\n=== Exemplo de Listen/Notify - CONCLUÍDO ===")
}

func demonstrateListenNotifyConceptually() {
	fmt.Println("\n🎯 Demonstração Conceitual de Listen/Notify")
	fmt.Println("==========================================")

	fmt.Println("\n💡 Conceitos fundamentais:")
	fmt.Println("  - LISTEN/NOTIFY é um sistema pub/sub nativo do PostgreSQL")
	fmt.Println("  - Permite comunicação assíncrona entre sessões")
	fmt.Println("  - Ideal para notificações em tempo real")
	fmt.Println("  - Suporta payloads de até 8KB")

	fmt.Println("\n🔄 Como funciona:")
	fmt.Println("  1. Sessão A executa LISTEN 'canal'")
	fmt.Println("  2. Sessão B executa NOTIFY 'canal', 'mensagem'")
	fmt.Println("  3. Sessão A recebe a notificação instantaneamente")
	fmt.Println("  4. Múltiplas sessões podem escutar o mesmo canal")

	fmt.Println("\n🛠️ Casos de uso comuns:")
	fmt.Println("  - Invalidação de cache")
	fmt.Println("  - Sistemas de chat em tempo real")
	fmt.Println("  - Notificações de mudanças de dados")
	fmt.Println("  - Sincronização entre microserviços")
	fmt.Println("  - Triggers de reprocessamento")

	fmt.Println("\n⚡ Vantagens:")
	fmt.Println("  - 🚀 Latência ultra-baixa (< 1ms)")
	fmt.Println("  - 🔋 Eficiência: não há polling")
	fmt.Println("  - 📡 Escalabilidade: múltiplos listeners")
	fmt.Println("  - 🎯 Confiabilidade: garantida pelo PostgreSQL")
}

func demonstrateBasicListenNotify(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Listen/Notify Básico ===")

	// Criar duas conexões: uma para listen e outra para notify
	fmt.Println("   Criando conexão para listening...")
	listenerConn, err := postgres.Connect(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable")
	if err != nil {
		return fmt.Errorf("erro ao criar conexão listener: %w", err)
	}
	defer listenerConn.Close(ctx)

	// Configurar canal de teste
	channelName := "test_channel"
	fmt.Printf("   Fazendo LISTEN no canal '%s'...\n", channelName)

	err = listenerConn.Listen(ctx, channelName)
	if err != nil {
		return fmt.Errorf("erro ao fazer LISTEN: %w", err)
	}

	// Função para receber notificações
	done := make(chan bool)
	go func() {
		fmt.Println("   🎧 Aguardando notificações...")
		for {
			select {
			case <-done:
				return
			default:
				notification, err := listenerConn.WaitForNotification(ctx, 2*time.Second)
				if err != nil {
					if err == context.DeadlineExceeded {
						// Timeout normal, continua aguardando
						continue
					}
					fmt.Printf("   ❌ Erro ao aguardar notificação: %v\n", err)
					return
				}

				fmt.Printf("   📨 Notificação recebida: Canal='%s', Payload='%s', PID=%d\n",
					notification.Channel, notification.Payload, notification.PID)
			}
		}
	}()

	// Dar tempo para o listener se configurar
	time.Sleep(100 * time.Millisecond)

	// Enviar notificações
	fmt.Println("   Enviando notificações...")
	notifications := []string{
		"Primeira mensagem",
		"Segunda mensagem",
		"Terceira mensagem",
	}

	for i, message := range notifications {
		fmt.Printf("   📤 Enviando notificação %d: '%s'\n", i+1, message)
		_, err := conn.Exec(ctx, fmt.Sprintf("NOTIFY %s, '%s'", channelName, message))
		if err != nil {
			fmt.Printf("   ❌ Erro ao enviar notificação %d: %v\n", i+1, err)
		} else {
			fmt.Printf("   ✅ Notificação %d enviada com sucesso\n", i+1)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Aguardar processamento
	time.Sleep(1 * time.Second)

	// Parar goroutine
	close(done)
	time.Sleep(100 * time.Millisecond)

	// Parar de escutar
	fmt.Printf("   Parando de escutar canal '%s'...\n", channelName)
	err = listenerConn.Unlisten(ctx, channelName)
	if err != nil {
		return fmt.Errorf("erro ao fazer UNLISTEN: %w", err)
	}

	fmt.Println("   ✅ Exemplo básico concluído")
	return nil
}

func demonstrateMultipleChannels(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Múltiplos Canais ===")

	// Criar conexão para listening
	listenerConn, err := postgres.Connect(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable")
	if err != nil {
		return fmt.Errorf("erro ao criar conexão listener: %w", err)
	}
	defer listenerConn.Close(ctx)

	// Configurar múltiplos canais
	channels := []string{"orders", "users", "products", "system"}

	fmt.Printf("   Configurando LISTEN para %d canais...\n", len(channels))
	for _, channel := range channels {
		err := listenerConn.Listen(ctx, channel)
		if err != nil {
			return fmt.Errorf("erro ao fazer LISTEN no canal %s: %w", channel, err)
		}
		fmt.Printf("   ✅ Listening em canal '%s'\n", channel)
	}

	// Contador de notificações por canal
	notificationCount := make(map[string]int)
	for _, channel := range channels {
		notificationCount[channel] = 0
	}

	// Função para receber notificações
	done := make(chan bool)
	go func() {
		fmt.Println("   🎧 Aguardando notificações em múltiplos canais...")
		for {
			select {
			case <-done:
				return
			default:
				notification, err := listenerConn.WaitForNotification(ctx, 2*time.Second)
				if err != nil {
					if err == context.DeadlineExceeded {
						// Timeout normal, continua aguardando
						continue
					}
					fmt.Printf("   ❌ Erro ao aguardar notificação: %v\n", err)
					return
				}

				notificationCount[notification.Channel]++
				fmt.Printf("   📨 [%s] Notificação #%d: '%s'\n",
					notification.Channel, notificationCount[notification.Channel], notification.Payload)
			}
		}
	}()

	// Dar tempo para o listener se configurar
	time.Sleep(200 * time.Millisecond)

	// Enviar notificações para diferentes canais
	fmt.Println("   Enviando notificações para diferentes canais...")

	notifications := []struct {
		channel string
		message string
	}{
		{"orders", "Nova ordem #1001"},
		{"users", "Usuário João logou"},
		{"products", "Produto em falta: Notebook"},
		{"system", "Sistema atualizado"},
		{"orders", "Nova ordem #1002"},
		{"users", "Usuário Maria logou"},
		{"orders", "Ordem #1001 processada"},
		{"system", "Backup concluído"},
	}

	for i, notif := range notifications {
		fmt.Printf("   📤 [%s] Enviando: '%s'\n", notif.channel, notif.message)
		_, err := conn.Exec(ctx, fmt.Sprintf("NOTIFY %s, '%s'", notif.channel, notif.message))
		if err != nil {
			fmt.Printf("   ❌ Erro ao enviar notificação %d: %v\n", i+1, err)
		}
		time.Sleep(300 * time.Millisecond)
	}

	// Aguardar processamento
	time.Sleep(1 * time.Second)

	// Parar goroutine
	close(done)
	time.Sleep(100 * time.Millisecond)

	// Mostrar estatísticas
	fmt.Println("\n   📊 Estatísticas por canal:")
	for _, channel := range channels {
		fmt.Printf("   - %s: %d notificações\n", channel, notificationCount[channel])
	}

	// Parar de escutar todos os canais
	fmt.Println("   Parando de escutar todos os canais...")
	for _, channel := range channels {
		err := listenerConn.Unlisten(ctx, channel)
		if err != nil {
			fmt.Printf("   ❌ Erro ao fazer UNLISTEN no canal %s: %v\n", channel, err)
		}
	}

	fmt.Println("   ✅ Exemplo de múltiplos canais concluído")
	return nil
}

func demonstrateNotificationsWithPayload(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Notificações com Payload ===")

	// Criar conexão para listening
	listenerConn, err := postgres.Connect(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable")
	if err != nil {
		return fmt.Errorf("erro ao criar conexão listener: %w", err)
	}
	defer listenerConn.Close(ctx)

	channelName := "json_channel"
	fmt.Printf("   Fazendo LISTEN no canal '%s'...\n", channelName)

	err = listenerConn.Listen(ctx, channelName)
	if err != nil {
		return fmt.Errorf("erro ao fazer LISTEN: %w", err)
	}

	// Função para receber notificações
	done := make(chan bool)
	go func() {
		fmt.Println("   🎧 Aguardando notificações com payload JSON...")
		for {
			select {
			case <-done:
				return
			default:
				notification, err := listenerConn.WaitForNotification(ctx, 3*time.Second)
				if err != nil {
					if err == context.DeadlineExceeded {
						// Timeout normal, continua aguardando
						continue
					}
					fmt.Printf("   ❌ Erro ao aguardar notificação: %v\n", err)
					return
				}

				fmt.Printf("   📨 Payload JSON recebido:\n")
				fmt.Printf("       Canal: %s\n", notification.Channel)
				fmt.Printf("       PID: %d\n", notification.PID)
				fmt.Printf("       Payload: %s\n", notification.Payload)

				// Em aplicação real, você faria parse do JSON aqui
				if len(notification.Payload) > 0 {
					fmt.Printf("       Tamanho: %d bytes\n", len(notification.Payload))
				}
			}
		}
	}()

	// Dar tempo para o listener se configurar
	time.Sleep(100 * time.Millisecond)

	// Enviar notificações com payloads estruturados
	fmt.Println("   Enviando notificações com payloads JSON...")

	jsonPayloads := []string{
		`{"event": "user_login", "user_id": 123, "timestamp": "2025-01-01T10:00:00Z"}`,
		`{"event": "order_created", "order_id": 456, "amount": 199.99, "currency": "USD"}`,
		`{"event": "product_updated", "product_id": 789, "changes": ["price", "description"]}`,
		`{"event": "system_alert", "level": "warning", "message": "High CPU usage detected"}`,
	}

	for i, payload := range jsonPayloads {
		fmt.Printf("   📤 Enviando payload %d (%d bytes)...\n", i+1, len(payload))
		_, err := conn.Exec(ctx, fmt.Sprintf("NOTIFY %s, '%s'", channelName, payload))
		if err != nil {
			fmt.Printf("   ❌ Erro ao enviar payload %d: %v\n", i+1, err)
		} else {
			fmt.Printf("   ✅ Payload %d enviado com sucesso\n", i+1)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Teste com payload grande
	fmt.Println("   Testando payload grande...")
	largePayload := `{"event": "bulk_import", "data": [` +
		`{"id": 1, "name": "Item 1", "description": "Descrição detalhada do item 1"},` +
		`{"id": 2, "name": "Item 2", "description": "Descrição detalhada do item 2"},` +
		`{"id": 3, "name": "Item 3", "description": "Descrição detalhada do item 3"}` +
		`], "total": 3, "timestamp": "2025-01-01T11:00:00Z"}`

	fmt.Printf("   📤 Enviando payload grande (%d bytes)...\n", len(largePayload))
	_, err = conn.Exec(ctx, fmt.Sprintf("NOTIFY %s, '%s'", channelName, largePayload))
	if err != nil {
		fmt.Printf("   ❌ Erro ao enviar payload grande: %v\n", err)
	} else {
		fmt.Printf("   ✅ Payload grande enviado com sucesso\n")
	}

	// Aguardar processamento
	time.Sleep(1 * time.Second)

	// Parar goroutine
	close(done)
	time.Sleep(100 * time.Millisecond)

	// Parar de escutar
	err = listenerConn.Unlisten(ctx, channelName)
	if err != nil {
		return fmt.Errorf("erro ao fazer UNLISTEN: %w", err)
	}

	fmt.Println("   ✅ Exemplo de payloads concluído")
	return nil
}

func demonstrateSimpleChat(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Sistema de Chat Simples ===")

	// Criar conexão para listening
	listenerConn, err := postgres.Connect(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable")
	if err != nil {
		return fmt.Errorf("erro ao criar conexão listener: %w", err)
	}
	defer listenerConn.Close(ctx)

	chatChannel := "chat_room"
	fmt.Printf("   Entrando na sala de chat '%s'...\n", chatChannel)

	err = listenerConn.Listen(ctx, chatChannel)
	if err != nil {
		return fmt.Errorf("erro ao fazer LISTEN: %w", err)
	}

	// Simulação de usuário ouvindo mensagens
	done := make(chan bool)
	go func() {
		fmt.Println("   💬 Aguardando mensagens do chat...")
		for {
			select {
			case <-done:
				return
			default:
				notification, err := listenerConn.WaitForNotification(ctx, 2*time.Second)
				if err != nil {
					if err == context.DeadlineExceeded {
						// Timeout normal, continua aguardando
						continue
					}
					fmt.Printf("   ❌ Erro ao aguardar mensagem: %v\n", err)
					return
				}

				fmt.Printf("   💬 Nova mensagem: %s\n", notification.Payload)
			}
		}
	}()

	// Dar tempo para o listener se configurar
	time.Sleep(100 * time.Millisecond)

	// Simular conversação
	fmt.Println("   Simulando conversação...")

	messages := []struct {
		user    string
		message string
	}{
		{"Alice", "Olá pessoal!"},
		{"Bob", "Oi Alice! Como vai?"},
		{"Charlie", "Bom dia a todos!"},
		{"Alice", "Tudo bem, obrigada! E vocês?"},
		{"Bob", "Tudo ótimo aqui"},
		{"Charlie", "Alguém viu o relatório de ontem?"},
		{"Alice", "Sim, está na pasta compartilhada"},
		{"System", "Backup automático concluído"},
	}

	for i, msg := range messages {
		chatMessage := fmt.Sprintf("[%s] %s", msg.user, msg.message)
		fmt.Printf("   📤 Enviando: %s\n", chatMessage)

		_, err := conn.Exec(ctx, fmt.Sprintf("NOTIFY %s, '%s'", chatChannel, chatMessage))
		if err != nil {
			fmt.Printf("   ❌ Erro ao enviar mensagem %d: %v\n", i+1, err)
		}
		time.Sleep(800 * time.Millisecond)
	}

	// Aguardar processamento
	time.Sleep(1 * time.Second)

	// Parar goroutine
	close(done)
	time.Sleep(100 * time.Millisecond)

	// Sair da sala
	fmt.Printf("   Saindo da sala de chat '%s'...\n", chatChannel)
	err = listenerConn.Unlisten(ctx, chatChannel)
	if err != nil {
		return fmt.Errorf("erro ao fazer UNLISTEN: %w", err)
	}

	fmt.Println("   ✅ Sistema de chat simples concluído")
	return nil
}

func demonstrateChangeMonitoring(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Monitoramento de Mudanças ===")

	// Criar tabela para monitoramento
	fmt.Println("   Criando tabela para monitoramento...")
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS monitored_table (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			value INTEGER NOT NULL,
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela: %w", err)
	}

	// Limpar dados anteriores
	_, err = conn.Exec(ctx, "DELETE FROM monitored_table")
	if err != nil {
		return fmt.Errorf("erro ao limpar tabela: %w", err)
	}

	// Criar função de trigger (se não existir)
	fmt.Println("   Criando função de trigger...")
	_, err = conn.Exec(ctx, `
		CREATE OR REPLACE FUNCTION notify_change()
		RETURNS TRIGGER AS $$
		BEGIN
			IF TG_OP = 'INSERT' THEN
				PERFORM pg_notify('table_changes', 
					json_build_object('operation', 'INSERT', 'id', NEW.id, 'name', NEW.name, 'value', NEW.value)::text);
				RETURN NEW;
			ELSIF TG_OP = 'UPDATE' THEN
				PERFORM pg_notify('table_changes', 
					json_build_object('operation', 'UPDATE', 'id', NEW.id, 'name', NEW.name, 'value', NEW.value)::text);
				RETURN NEW;
			ELSIF TG_OP = 'DELETE' THEN
				PERFORM pg_notify('table_changes', 
					json_build_object('operation', 'DELETE', 'id', OLD.id, 'name', OLD.name, 'value', OLD.value)::text);
				RETURN OLD;
			END IF;
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar função de trigger: %w", err)
	}

	// Criar trigger
	fmt.Println("   Criando trigger...")
	_, err = conn.Exec(ctx, `
		DROP TRIGGER IF EXISTS change_trigger ON monitored_table;
		CREATE TRIGGER change_trigger
		AFTER INSERT OR UPDATE OR DELETE ON monitored_table
		FOR EACH ROW EXECUTE FUNCTION notify_change();
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar trigger: %w", err)
	}

	// Criar conexão para listening
	listenerConn, err := postgres.Connect(ctx, "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable")
	if err != nil {
		return fmt.Errorf("erro ao criar conexão listener: %w", err)
	}
	defer listenerConn.Close(ctx)

	changeChannel := "table_changes"
	fmt.Printf("   Monitorando mudanças no canal '%s'...\n", changeChannel)

	err = listenerConn.Listen(ctx, changeChannel)
	if err != nil {
		return fmt.Errorf("erro ao fazer LISTEN: %w", err)
	}

	// Contador de mudanças
	changeCount := 0

	// Função para receber notificações de mudanças
	done := make(chan bool)
	go func() {
		fmt.Println("   🔍 Aguardando mudanças na tabela...")
		for {
			select {
			case <-done:
				return
			default:
				notification, err := listenerConn.WaitForNotification(ctx, 2*time.Second)
				if err != nil {
					if err == context.DeadlineExceeded {
						// Timeout normal, continua aguardando
						continue
					}
					fmt.Printf("   ❌ Erro ao aguardar mudanças: %v\n", err)
					return
				}

				changeCount++
				fmt.Printf("   🔄 Mudança #%d detectada: %s\n", changeCount, notification.Payload)
			}
		}
	}()

	// Dar tempo para o listener se configurar
	time.Sleep(100 * time.Millisecond)

	// Simular mudanças na tabela
	fmt.Println("   Simulando mudanças na tabela...")

	// INSERT
	fmt.Println("   📝 Inserindo registros...")
	_, err = conn.Exec(ctx, "INSERT INTO monitored_table (name, value) VALUES ('Item 1', 10)")
	if err != nil {
		return fmt.Errorf("erro ao inserir registro 1: %w", err)
	}

	_, err = conn.Exec(ctx, "INSERT INTO monitored_table (name, value) VALUES ('Item 2', 20)")
	if err != nil {
		return fmt.Errorf("erro ao inserir registro 2: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// UPDATE
	fmt.Println("   ✏️ Atualizando registros...")
	_, err = conn.Exec(ctx, "UPDATE monitored_table SET value = 15 WHERE name = 'Item 1'")
	if err != nil {
		return fmt.Errorf("erro ao atualizar registro: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// DELETE
	fmt.Println("   🗑️ Deletando registros...")
	_, err = conn.Exec(ctx, "DELETE FROM monitored_table WHERE name = 'Item 2'")
	if err != nil {
		return fmt.Errorf("erro ao deletar registro: %w", err)
	}

	// Aguardar processamento
	time.Sleep(1 * time.Second)

	// Parar goroutine
	close(done)
	time.Sleep(100 * time.Millisecond)

	// Parar de escutar
	err = listenerConn.Unlisten(ctx, changeChannel)
	if err != nil {
		return fmt.Errorf("erro ao fazer UNLISTEN: %w", err)
	}

	// Limpeza
	fmt.Println("   Limpando recursos...")
	_, err = conn.Exec(ctx, "DROP TRIGGER IF EXISTS change_trigger ON monitored_table")
	if err != nil {
		fmt.Printf("   ❌ Erro ao remover trigger: %v\n", err)
	}

	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS monitored_table")
	if err != nil {
		fmt.Printf("   ❌ Erro ao remover tabela: %v\n", err)
	}

	fmt.Printf("   📊 Total de mudanças detectadas: %d\n", changeCount)
	fmt.Println("   ✅ Monitoramento de mudanças concluído")
	return nil
}
