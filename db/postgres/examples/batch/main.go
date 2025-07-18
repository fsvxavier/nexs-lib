package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
	pgxprovider "github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx"
)

func main() {
	fmt.Println("=== Exemplo de Operações Batch ===")

	// Configuração da conexão
	connectionString := "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Conectar ao banco
	fmt.Println("\n1. Conectando ao banco...")
	conn, err := postgres.Connect(ctx, connectionString)
	if err != nil {
		log.Printf("💡 Exemplo de batch seria executado com banco real: %v", err)
		demonstrateBatchConceptually()
		return
	}
	defer conn.Close(ctx)

	// 2. Criar tabela de teste
	fmt.Println("2. Criando tabela de teste...")
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			category VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}

	// 3. Limpar dados anteriores
	fmt.Println("3. Limpando dados anteriores...")
	_, err = conn.Exec(ctx, "DELETE FROM products")
	if err != nil {
		log.Fatalf("Erro ao limpar dados: %v", err)
	}

	// 4. Exemplo: Operações batch básicas
	fmt.Println("\n4. Exemplo: Operações batch básicas...")
	if err := demonstrateBasicBatch(ctx, conn); err != nil {
		log.Printf("Erro no exemplo básico: %v", err)
	}

	// 5. Exemplo: Batch com transação
	fmt.Println("\n5. Exemplo: Batch com transação...")
	if err := demonstrateBatchWithTransaction(ctx, conn); err != nil {
		log.Printf("Erro no exemplo com transação: %v", err)
	}

	// 6. Exemplo: Comparação de performance
	fmt.Println("\n6. Exemplo: Comparação de performance...")
	if err := demonstratePerformanceComparison(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de performance: %v", err)
	}

	// 7. Exemplo: Tratamento de erros em batch
	fmt.Println("\n7. Exemplo: Tratamento de erros em batch...")
	if err := demonstrateErrorHandling(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de tratamento de erros: %v", err)
	}

	// 8. Limpeza
	fmt.Println("\n8. Limpando tabela de teste...")
	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS products")
	if err != nil {
		log.Printf("Erro ao limpar tabela: %v", err)
	}

	fmt.Println("\n=== Exemplo de Operações Batch - CONCLUÍDO ===")
}

func demonstrateBatchConceptually() {
	fmt.Println("\n🎯 Demonstração Conceitual de Operações Batch")
	fmt.Println("============================================")

	fmt.Println("\n💡 Conceitos fundamentais:")
	fmt.Println("  - Batch agrupa múltiplas operações em uma única requisição")
	fmt.Println("  - Reduz latência de rede e overhead de comunicação")
	fmt.Println("  - Melhora significativamente a performance para operações em massa")
	fmt.Println("  - Permite transações atômicas em múltiplas operações")

	fmt.Println("\n⚡ Vantagens dos Batches:")
	fmt.Println("  - 📈 Performance: 5-10x mais rápido que operações individuais")
	fmt.Println("  - 🔄 Atomicidade: Todas as operações ou nenhuma")
	fmt.Println("  - 🌐 Redução de latência: Menos round-trips ao banco")
	fmt.Println("  - 💾 Eficiência de memória: Melhor uso de buffers")

	fmt.Println("\n🛠️ Tipos de operações suportadas:")
	fmt.Println("  - INSERT em massa")
	fmt.Println("  - UPDATE múltiplos")
	fmt.Println("  - DELETE em lote")
	fmt.Println("  - Queries SELECT múltiplas")
	fmt.Println("  - Operações mistas (INSERT + UPDATE + DELETE)")

	fmt.Println("\n📊 Exemplo de Performance:")
	fmt.Println("  - Inserção individual: 1000 registros em ~15s")
	fmt.Println("  - Inserção em batch: 1000 registros em ~2s")
	fmt.Println("  - Speedup: 7.5x mais rápido")
}

func demonstrateBasicBatch(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Operações Batch Básicas ===")

	// Criar um batch
	batch := pgxprovider.NewBatch()
	if batch == nil {
		return fmt.Errorf("falha ao criar batch")
	}

	// Adicionar operações ao batch
	products := []struct {
		name     string
		price    float64
		category string
	}{
		{"Notebook Dell", 2500.00, "Eletrônicos"},
		{"Mouse Logitech", 150.00, "Periféricos"},
		{"Teclado Mecânico", 350.00, "Periféricos"},
		{"Monitor 4K", 1200.00, "Eletrônicos"},
		{"Webcam HD", 200.00, "Eletrônicos"},
	}

	fmt.Printf("   Adicionando %d produtos ao batch...\n", len(products))
	for _, p := range products {
		batch.Queue(
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			p.name, p.price, p.category,
		)
	}

	fmt.Printf("   Batch criado com %d operações\n", batch.Len())

	// Executar batch
	fmt.Println("   Executando batch...")
	startTime := time.Now()
	results := conn.SendBatch(ctx, batch)
	defer results.Close()

	// Processar resultados
	successful := 0
	failed := 0

	for i := 0; i < len(products); i++ {
		_, err := results.Exec()
		if err != nil {
			fmt.Printf("   ❌ Erro ao inserir produto %d: %v\n", i+1, err)
			failed++
		} else {
			successful++
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("   ✅ Batch concluído em %v\n", duration)
	fmt.Printf("   📊 Resultado: %d sucessos, %d falhas\n", successful, failed)

	// Verificar resultados
	fmt.Println("   Verificando produtos inseridos:")
	rows, err := conn.Query(ctx, "SELECT name, price, category FROM products ORDER BY name")
	if err != nil {
		return fmt.Errorf("erro ao consultar produtos: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var name, category string
		var price float64
		if err := rows.Scan(&name, &price, &category); err != nil {
			return fmt.Errorf("erro ao ler produto: %w", err)
		}
		count++
		fmt.Printf("     %d. %s - $%.2f (%s)\n", count, name, price, category)
	}

	return nil
}

func demonstrateBatchWithTransaction(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Batch com Transação ===")

	// Limpar dados anteriores
	_, err := conn.Exec(ctx, "DELETE FROM products")
	if err != nil {
		return fmt.Errorf("erro ao limpar dados: %w", err)
	}

	// Iniciar transação
	fmt.Println("   Iniciando transação...")
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}

	// Função para gerenciar commit/rollback
	var commitTx = true
	defer func() {
		if commitTx {
			if err := tx.Commit(ctx); err != nil {
				fmt.Printf("   ❌ Erro ao fazer commit: %v\n", err)
			} else {
				fmt.Println("   ✅ Transação commitada com sucesso")
			}
		} else {
			if err := tx.Rollback(ctx); err != nil {
				fmt.Printf("   ❌ Erro ao fazer rollback: %v\n", err)
			} else {
				fmt.Println("   🔄 Transação cancelada (rollback)")
			}
		}
	}()

	// Criar batch para operações mistas
	batch := pgxprovider.NewBatch()

	// Adicionar operações diversas
	operations := []struct {
		desc  string
		query string
		args  []interface{}
	}{
		{
			"Inserir produto premium",
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			[]interface{}{"MacBook Pro", 8000.00, "Eletrônicos"},
		},
		{
			"Inserir produto médio",
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			[]interface{}{"iPad Air", 3000.00, "Eletrônicos"},
		},
		{
			"Inserir produto básico",
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			[]interface{}{"AirPods", 1200.00, "Eletrônicos"},
		},
	}

	fmt.Printf("   Adicionando %d operações ao batch...\n", len(operations))
	for _, op := range operations {
		batch.Queue(op.query, op.args...)
	}

	// Executar batch na transação
	fmt.Println("   Executando batch na transação...")
	startTime := time.Now()
	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	// Processar resultados
	successful := 0
	for _, op := range operations {
		cmdTag, err := results.Exec()
		if err != nil {
			fmt.Printf("   ❌ Erro em '%s': %v\n", op.desc, err)
			commitTx = false
			return nil
		}

		fmt.Printf("   ✅ %s: %d linhas afetadas\n", op.desc, cmdTag.RowsAffected())
		successful++
	}

	duration := time.Since(startTime)
	fmt.Printf("   ⏱️ Batch executado em %v\n", duration)
	fmt.Printf("   📊 %d/%d operações bem-sucedidas\n", successful, len(operations))

	// Verificar dentro da transação
	fmt.Println("   Verificando dados dentro da transação:")
	rows, err := tx.Query(ctx, "SELECT COUNT(*) FROM products")
	if err != nil {
		commitTx = false
		return fmt.Errorf("erro ao contar produtos: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			commitTx = false
			return fmt.Errorf("erro ao ler contagem: %w", err)
		}
		fmt.Printf("   📊 Total de produtos na transação: %d\n", count)
	}

	return nil
}

func demonstratePerformanceComparison(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Comparação de Performance ===")

	// Limpar dados
	_, err := conn.Exec(ctx, "DELETE FROM products")
	if err != nil {
		return fmt.Errorf("erro ao limpar dados: %w", err)
	}

	// Preparar dados de teste
	testData := make([]struct {
		name     string
		price    float64
		category string
	}, 100)

	for i := 0; i < 100; i++ {
		testData[i] = struct {
			name     string
			price    float64
			category string
		}{
			name:     fmt.Sprintf("Produto %d", i+1),
			price:    float64(100 + i*10),
			category: "Teste",
		}
	}

	// Teste 1: Inserções individuais
	fmt.Println("\n   Teste 1: Inserções individuais...")
	startTime := time.Now()
	for i, p := range testData {
		_, err := conn.Exec(ctx,
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			p.name, p.price, p.category,
		)
		if err != nil {
			fmt.Printf("   ❌ Erro ao inserir produto %d: %v\n", i+1, err)
		}
	}
	individualDuration := time.Since(startTime)

	// Verificar contagem
	var count1 int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count1)
	if err != nil {
		return fmt.Errorf("erro ao contar produtos: %w", err)
	}

	fmt.Printf("   ⏱️ Inserções individuais: %v (%d registros)\n", individualDuration, count1)

	// Limpar para próximo teste
	_, err = conn.Exec(ctx, "DELETE FROM products")
	if err != nil {
		return fmt.Errorf("erro ao limpar dados: %w", err)
	}

	// Teste 2: Inserção em batch
	fmt.Println("\n   Teste 2: Inserção em batch...")
	startTime = time.Now()

	batch := pgxprovider.NewBatch()
	for _, p := range testData {
		batch.Queue(
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			p.name, p.price, p.category,
		)
	}

	results := conn.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(testData); i++ {
		_, err := results.Exec()
		if err != nil {
			fmt.Printf("   ❌ Erro no batch item %d: %v\n", i+1, err)
		}
	}

	batchDuration := time.Since(startTime)

	// Verificar contagem
	var count2 int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count2)
	if err != nil {
		return fmt.Errorf("erro ao contar produtos: %w", err)
	}

	fmt.Printf("   ⏱️ Inserção em batch: %v (%d registros)\n", batchDuration, count2)

	// Análise de performance
	fmt.Println("\n   📊 Análise de Performance:")
	if individualDuration > 0 && batchDuration > 0 {
		speedup := float64(individualDuration) / float64(batchDuration)
		fmt.Printf("   🚀 Speedup: %.2fx mais rápido\n", speedup)

		individualOPS := float64(len(testData)) / individualDuration.Seconds()
		batchOPS := float64(len(testData)) / batchDuration.Seconds()

		fmt.Printf("   📈 Individual: %.1f ops/sec\n", individualOPS)
		fmt.Printf("   📈 Batch: %.1f ops/sec\n", batchOPS)
	}

	return nil
}

func demonstrateErrorHandling(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Tratamento de Erros em Batch ===")

	// Limpar dados
	_, err := conn.Exec(ctx, "DELETE FROM products")
	if err != nil {
		return fmt.Errorf("erro ao limpar dados: %w", err)
	}

	// Criar batch com operações válidas e inválidas
	batch := pgxprovider.NewBatch()

	operations := []struct {
		desc  string
		query string
		args  []interface{}
		valid bool
	}{
		{
			"Inserir produto válido 1",
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			[]interface{}{"Produto Válido 1", 100.00, "Categoria"},
			true,
		},
		{
			"Inserir produto com erro (preço negativo)",
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			[]interface{}{"Produto Inválido", -100.00, "Categoria"},
			false, // Pode gerar erro se houver constraint
		},
		{
			"Inserir produto válido 2",
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			[]interface{}{"Produto Válido 2", 200.00, "Categoria"},
			true,
		},
		{
			"Query inválida",
			"INSERT INTO tabela_inexistente (campo) VALUES ($1)",
			[]interface{}{"valor"},
			false,
		},
		{
			"Inserir produto válido 3",
			"INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
			[]interface{}{"Produto Válido 3", 300.00, "Categoria"},
			true,
		},
	}

	fmt.Printf("   Adicionando %d operações ao batch (algumas com erro)...\n", len(operations))
	for _, op := range operations {
		batch.Queue(op.query, op.args...)
	}

	// Executar batch
	fmt.Println("   Executando batch com tratamento de erros...")
	results := conn.SendBatch(ctx, batch)
	defer results.Close()

	// Processar resultados com tratamento de erro
	successful := 0
	failed := 0

	for _, op := range operations {
		cmdTag, err := results.Exec()
		if err != nil {
			fmt.Printf("   ❌ Erro em '%s': %v\n", op.desc, err)
			failed++
		} else {
			fmt.Printf("   ✅ Sucesso '%s': %d linhas afetadas\n", op.desc, cmdTag.RowsAffected())
			successful++
		}
	}

	// Resumo do tratamento de erros
	fmt.Printf("\n   📊 Resumo do Tratamento de Erros:\n")
	fmt.Printf("   ✅ Operações bem-sucedidas: %d\n", successful)
	fmt.Printf("   ❌ Operações com falha: %d\n", failed)
	fmt.Printf("   📈 Taxa de sucesso: %.1f%%\n", float64(successful)/float64(len(operations))*100)

	// Verificar dados inseridos
	fmt.Println("\n   Verificando dados inseridos com sucesso:")
	rows, err := conn.Query(ctx, "SELECT name, price FROM products ORDER BY name")
	if err != nil {
		return fmt.Errorf("erro ao consultar produtos: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var name string
		var price float64
		if err := rows.Scan(&name, &price); err != nil {
			return fmt.Errorf("erro ao ler produto: %w", err)
		}
		count++
		fmt.Printf("     %d. %s - $%.2f\n", count, name, price)
	}

	fmt.Printf("   📊 Total de produtos inseridos: %d\n", count)

	return nil
}
