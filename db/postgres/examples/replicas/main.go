package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
)

func main() {
	fmt.Println("=== Sistema de Read Replicas PostgreSQL ===")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Configuração das conexões
	fmt.Println("\n1. Configuração das conexões...")

	primaryDSN := getEnvOrDefault("NEXS_PRIMARY_DSN", "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")
	replica1DSN := getEnvOrDefault("NEXS_REPLICA1_DSN", "postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb")
	replica2DSN := getEnvOrDefault("NEXS_REPLICA2_DSN", "postgres://nexs_user:nexs_password@localhost:5434/nexs_testdb")

	fmt.Printf("   Primary: %s\n", primaryDSN)
	fmt.Printf("   Replica 1: %s\n", replica1DSN)
	fmt.Printf("   Replica 2: %s\n", replica2DSN)

	// Demonstrar uso real com conexões
	fmt.Println("\n2. Demonstrando uso real com conexões...")
	if err := demonstrateRealUsage(ctx, primaryDSN, replica1DSN, replica2DSN); err != nil {
		log.Printf("Erro na demonstração: %v", err)
	}

	// Demonstrar verificação de conectividade
	fmt.Println("\n3. Demonstrando verificação de conectividade...")
	if err := demonstrateConnectivityCheck(ctx, primaryDSN, replica1DSN, replica2DSN); err != nil {
		log.Printf("Erro na verificação de conectividade: %v", err)
	}

	fmt.Println("\n=== Demonstração de Read Replicas concluída ===")
}

func demonstrateRealUsage(ctx context.Context, primaryDSN, replica1DSN, replica2DSN string) error {
	fmt.Println("   Criando conexões com primary e réplicas...")

	// Conectar ao primary
	primaryConn, err := postgres.Connect(ctx, primaryDSN)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao primary: %w", err)
	}
	defer primaryConn.Close(ctx)

	// Tentar conectar às réplicas (podem falhar se não existirem)
	replica1Conn, err := postgres.Connect(ctx, replica1DSN)
	if err != nil {
		fmt.Printf("   Aviso: Não foi possível conectar à replica1: %v\n", err)
		replica1Conn = nil
	}
	if replica1Conn != nil {
		defer replica1Conn.Close(ctx)
	}

	replica2Conn, err := postgres.Connect(ctx, replica2DSN)
	if err != nil {
		fmt.Printf("   Aviso: Não foi possível conectar à replica2: %v\n", err)
		replica2Conn = nil
	}
	if replica2Conn != nil {
		defer replica2Conn.Close(ctx)
	}

	// Operações de escrita no primary
	fmt.Println("   Testando operações de escrita no primary...")

	// Criar tabela
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS replica_test (
			id SERIAL PRIMARY KEY,
			message TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err = primaryConn.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela: %w", err)
	}

	// Inserir dados
	insertSQL := "INSERT INTO replica_test (message) VALUES ($1), ($2), ($3)"
	_, err = primaryConn.Exec(ctx, insertSQL, "Mensagem 1", "Mensagem 2", "Mensagem 3")
	if err != nil {
		return fmt.Errorf("erro ao inserir dados: %w", err)
	}

	fmt.Println("   ✓ Dados inseridos no primary com sucesso")

	// Aguardar replicação (reduzido para exemplo)
	fmt.Println("   Aguardando replicação...")
	time.Sleep(100 * time.Millisecond)

	// Testar leitura das réplicas
	if replica1Conn != nil {
		fmt.Println("   Testando leitura da replica1...")
		var count int
		err = replica1Conn.QueryRow(ctx, "SELECT COUNT(*) FROM replica_test").Scan(&count)
		if err != nil {
			fmt.Printf("   Aviso: Erro ao ler da replica1: %v\n", err)
		} else {
			fmt.Printf("   ✓ Replica1: %d registros encontrados\n", count)
		}
	}

	if replica2Conn != nil {
		fmt.Println("   Testando leitura da replica2...")
		var count int
		err = replica2Conn.QueryRow(ctx, "SELECT COUNT(*) FROM replica_test").Scan(&count)
		if err != nil {
			fmt.Printf("   Aviso: Erro ao ler da replica2: %v\n", err)
		} else {
			fmt.Printf("   ✓ Replica2: %d registros encontrados\n", count)
		}
	}

	// Limpar dados
	_, err = primaryConn.Exec(ctx, "DROP TABLE IF EXISTS replica_test")
	if err != nil {
		return fmt.Errorf("erro ao limpar tabela: %w", err)
	}

	fmt.Println("   ✓ Demonstração de uso real concluída")
	return nil
}

func demonstrateConnectivityCheck(ctx context.Context, primaryDSN, replica1DSN, replica2DSN string) error {
	fmt.Println("   Verificando conectividade de todas as conexões...")

	// Testar conexões
	connections := []struct {
		name string
		dsn  string
	}{
		{"Primary", primaryDSN},
		{"Replica1", replica1DSN},
		{"Replica2", replica2DSN},
	}

	for _, connInfo := range connections {
		fmt.Printf("   Testando %s...\n", connInfo.name)

		conn, err := postgres.Connect(ctx, connInfo.dsn)
		if err != nil {
			fmt.Printf("   ❌ %s: Falha na conexão: %v\n", connInfo.name, err)
			continue
		}

		var version string
		err = conn.QueryRow(ctx, "SELECT version()").Scan(&version)
		if err != nil {
			fmt.Printf("   ❌ %s: Falha na query: %v\n", connInfo.name, err)
		} else {
			fmt.Printf("   ✓ %s: Conectado com sucesso\n", connInfo.name)
		}

		conn.Close(ctx)
	}

	fmt.Println("   ✓ Verificação de conectividade concluída")
	return nil
}

func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}
