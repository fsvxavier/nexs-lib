package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
)

func main() {
	fmt.Println("=== Exemplo Pool de Conexões PostgreSQL ===")

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Configurar DSN do banco
	dsn := getEnvOrDefault("NEXS_DB_DSN", "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb")

	// Demonstrar diferentes configurações de pool (simplificado)
	examples := []struct {
		name string
		fn   func(context.Context, string) error
	}{
		{"Pool Básico", demonstrateBasicPool},
		{"Pool Configurado", demonstrateConfiguredPool},
		{"Pool com Timeout", demonstratePoolTimeout},
	}

	for _, example := range examples {
		fmt.Printf("\n%s\n", example.name)
		fmt.Printf("%s\n", strings.Repeat("=", len(example.name)))

		if err := example.fn(ctx, dsn); err != nil {
			log.Printf("Erro no exemplo %s: %v", example.name, err)
		} else {
			fmt.Printf("✓ %s concluído com sucesso\n", example.name)
		}
	}

	fmt.Println("\n=== Exemplos de pool concluídos ===")
}

// demonstrateBasicPool demonstra uso básico de pool
func demonstrateBasicPool(ctx context.Context, dsn string) error {
	fmt.Println("  Criando pool básico...")

	// Criar pool básico
	pool, err := postgres.ConnectPool(ctx, dsn)
	if err != nil {
		return fmt.Errorf("erro ao criar pool: %w", err)
	}

	// Adquirir conexão
	fmt.Println("  Adquirindo conexão...")
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}

	// Executar query simples
	var result string
	err = conn.QueryRow(ctx, "SELECT 'Pool básico funcionando!' as message").Scan(&result)
	if err != nil {
		conn.Release()
		return fmt.Errorf("erro ao executar query: %w", err)
	}

	fmt.Printf("  Resultado: %s\n", result)

	// Liberar conexão
	conn.Release()

	// Fechar pool de forma assíncrona
	go func() {
		time.Sleep(50 * time.Millisecond)
		pool.Close()
	}()

	return nil
}

// demonstrateConfiguredPool demonstra pool com configuração avançada
func demonstrateConfiguredPool(ctx context.Context, dsn string) error {
	fmt.Println("  Criando pool configurado...")

	// Criar configuração personalizada
	cfg := postgres.NewConfigWithOptions(
		dsn,
		postgres.WithMaxConns(10),
		postgres.WithMinConns(2),
		postgres.WithMaxConnLifetime(10*time.Minute),
		postgres.WithMaxConnIdleTime(5*time.Minute),
	)

	// Criar pool com configuração
	pool, err := postgres.ConnectPoolWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("erro ao criar pool configurado: %w", err)
	}

	// Adquirir conexão
	fmt.Println("  Adquirindo conexão do pool configurado...")
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}

	// Executar query
	var result string
	err = conn.QueryRow(ctx, "SELECT 'Pool configurado funcionando!' as message").Scan(&result)
	if err != nil {
		conn.Release()
		return fmt.Errorf("erro ao executar query: %w", err)
	}

	fmt.Printf("  Resultado: %s\n", result)

	// Liberar conexão
	conn.Release()

	// Fechar pool de forma assíncrona
	go func() {
		time.Sleep(50 * time.Millisecond)
		pool.Close()
	}()

	return nil
}

// demonstratePoolTimeout demonstra pool com timeout
func demonstratePoolTimeout(ctx context.Context, dsn string) error {
	fmt.Println("  Criando pool com timeout...")

	// Criar pool com timeout curto
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := postgres.ConnectPool(timeoutCtx, dsn)
	if err != nil {
		return fmt.Errorf("erro ao criar pool: %w", err)
	}

	// Adquirir conexão
	fmt.Println("  Adquirindo conexão com timeout...")
	conn, err := pool.Acquire(timeoutCtx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}

	// Executar query
	var result string
	err = conn.QueryRow(timeoutCtx, "SELECT 'Pool com timeout funcionando!' as message").Scan(&result)
	if err != nil {
		conn.Release()
		return fmt.Errorf("erro ao executar query: %w", err)
	}

	fmt.Printf("  Resultado: %s\n", result)

	// Liberar conexão
	conn.Release()

	// Fechar pool de forma assíncrona
	go func() {
		time.Sleep(50 * time.Millisecond)
		pool.Close()
	}()

	return nil
}

// getEnvOrDefault retorna o valor da variável de ambiente ou um valor padrão
func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}
