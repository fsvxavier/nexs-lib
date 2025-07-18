package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
)

func main() {
	fmt.Println("=== Exemplo de Transação com Begin() ===")

	// Configuração da conexão
	connectionString := "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Conectar ao banco
	fmt.Println("\n1. Conectando ao banco...")
	conn, err := postgres.Connect(ctx, connectionString)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer conn.Close(ctx)

	// 2. Criar tabela de teste
	fmt.Println("2. Criando tabela de teste...")
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			balance DECIMAL(10,2) NOT NULL DEFAULT 0.00
		)
	`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}

	// 3. Limpar dados anteriores
	fmt.Println("3. Limpando dados anteriores...")
	_, err = conn.Exec(ctx, "DELETE FROM accounts")
	if err != nil {
		log.Fatalf("Erro ao limpar dados: %v", err)
	}

	// 4. Inserir dados iniciais
	fmt.Println("4. Inserindo dados iniciais...")
	_, err = conn.Exec(ctx, `
		INSERT INTO accounts (name, balance) VALUES 
		('Alice', 1000.00),
		('Bob', 500.00)
	`)
	if err != nil {
		log.Fatalf("Erro ao inserir dados: %v", err)
	}

	// 5. Exemplo de transação com Begin() - Transferência de dinheiro
	fmt.Println("\n5. Exemplo de transação com Begin() - Transferência de dinheiro...")

	// Mostrar balances iniciais
	fmt.Println("   Balances iniciais:")
	rows, err := conn.Query(ctx, "SELECT name, balance FROM accounts ORDER BY name")
	if err != nil {
		log.Fatalf("Erro ao consultar balances: %v", err)
	}

	for rows.Next() {
		var name string
		var balance float64
		if err := rows.Scan(&name, &balance); err != nil {
			log.Fatalf("Erro ao ler balance: %v", err)
		}
		fmt.Printf("     %s: $%.2f\n", name, balance)
	}
	rows.Close()

	// Iniciar transação
	fmt.Println("   Iniciando transação...")
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("Erro ao iniciar transação: %v", err)
	}

	// Função para gerenciar a transação
	transferAmount := 150.00
	fromAccount := "Alice"
	toAccount := "Bob"

	fmt.Printf("   Transferindo $%.2f de %s para %s...\n", transferAmount, fromAccount, toAccount)

	// Operação dentro da transação
	success := func() bool {
		// Verificar se a conta de origem tem saldo suficiente
		var fromBalance float64
		row := tx.QueryRow(ctx, "SELECT balance FROM accounts WHERE name = $1", fromAccount)
		if err := row.Scan(&fromBalance); err != nil {
			log.Printf("Erro ao verificar saldo de %s: %v", fromAccount, err)
			return false
		}

		if fromBalance < transferAmount {
			log.Printf("Saldo insuficiente em %s: $%.2f < $%.2f", fromAccount, fromBalance, transferAmount)
			return false
		}

		// Debitar da conta de origem
		_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE name = $2", transferAmount, fromAccount)
		if err != nil {
			log.Printf("Erro ao debitar de %s: %v", fromAccount, err)
			return false
		}

		// Creditar na conta de destino
		_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE name = $2", transferAmount, toAccount)
		if err != nil {
			log.Printf("Erro ao creditar para %s: %v", toAccount, err)
			return false
		}

		return true
	}()

	if success {
		// Commit da transação
		fmt.Println("   Fazendo commit da transação...")
		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("Erro ao fazer commit: %v", err)
		}
		fmt.Println("   Transação commitada com sucesso!")
	} else {
		// Rollback da transação
		fmt.Println("   Fazendo rollback da transação...")
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("Erro ao fazer rollback: %v", err)
		}
		fmt.Println("   Transação cancelada!")
	}

	// 6. Mostrar balances finais
	fmt.Println("\n6. Balances finais:")
	rows, err = conn.Query(ctx, "SELECT name, balance FROM accounts ORDER BY name")
	if err != nil {
		log.Fatalf("Erro ao consultar balances finais: %v", err)
	}

	for rows.Next() {
		var name string
		var balance float64
		if err := rows.Scan(&name, &balance); err != nil {
			log.Fatalf("Erro ao ler balance final: %v", err)
		}
		fmt.Printf("     %s: $%.2f\n", name, balance)
	}
	rows.Close()

	// 7. Exemplo de transação com rollback
	fmt.Println("\n7. Exemplo de transação com rollback...")

	fmt.Println("   Iniciando transação que será cancelada...")
	tx2, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("Erro ao iniciar transação 2: %v", err)
	}

	// Fazer uma operação que será cancelada
	fmt.Println("   Tentando transferir $2000.00 (mais do que disponível)...")
	_, err = tx2.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE name = $2", 2000.00, "Alice")
	if err != nil {
		log.Printf("Erro na operação: %v", err)
	} else {
		fmt.Println("   Operação realizada (será cancelada)...")
	}

	// Fazer rollback
	fmt.Println("   Fazendo rollback da transação...")
	if err := tx2.Rollback(ctx); err != nil {
		log.Printf("Erro ao fazer rollback: %v", err)
	}
	fmt.Println("   Rollback realizado com sucesso!")

	// Verificar que os dados não foram alterados
	fmt.Println("   Verificando que os dados não foram alterados:")
	rows, err = conn.Query(ctx, "SELECT name, balance FROM accounts ORDER BY name")
	if err != nil {
		log.Fatalf("Erro ao consultar balances após rollback: %v", err)
	}

	for rows.Next() {
		var name string
		var balance float64
		if err := rows.Scan(&name, &balance); err != nil {
			log.Fatalf("Erro ao ler balance após rollback: %v", err)
		}
		fmt.Printf("     %s: $%.2f\n", name, balance)
	}
	rows.Close()

	// 8. Exemplo de transação com BeginTx() e opções
	fmt.Println("\n8. Exemplo de transação com BeginTx() e opções...")

	// Usar BeginTx com opções específicas
	fmt.Println("   Iniciando transação com nível de isolamento ReadCommitted...")
	tx3, err := conn.BeginTx(ctx, postgres.TxOptions{
		IsoLevel:   postgres.TxIsoLevelReadCommitted,
		AccessMode: postgres.TxAccessModeReadWrite,
	})
	if err != nil {
		log.Fatalf("Erro ao iniciar transação com opções: %v", err)
	}

	// Fazer uma operação simples
	fmt.Println("   Fazendo operação na transação com opções...")
	_, err = tx3.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE name = $2", 0.01, "Alice")
	if err != nil {
		log.Printf("Erro na operação: %v", err)
		tx3.Rollback(ctx)
	} else {
		fmt.Println("   Fazendo commit da transação com opções...")
		if err := tx3.Commit(ctx); err != nil {
			log.Fatalf("Erro ao fazer commit: %v", err)
		}
		fmt.Println("   Transação com opções commitada com sucesso!")
	}

	// 9. Limpeza
	fmt.Println("\n9. Limpando tabela de teste...")
	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS accounts")
	if err != nil {
		log.Printf("Erro ao limpar tabela: %v", err)
	}

	fmt.Println("\n=== Exemplo de Transação com Begin() - CONCLUÍDO ===")
}
