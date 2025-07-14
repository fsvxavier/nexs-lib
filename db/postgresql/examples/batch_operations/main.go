package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/pgx"
)

func main() {
	ctx := context.Background()

	fmt.Println("📦 Demonstração de Operações em Lote (Batch Operations)")
	fmt.Println("======================================================")

	// Criar provedor PGX
	provider := pgx.NewProvider()
	defer func() {
		if err := provider.Close(); err != nil {
			log.Printf("Erro ao fechar provedor: %v", err)
		}
	}()

	// Configurar banco de dados otimizado para operações em lote
	cfg := config.NewConfig(
		config.WithHost(getEnv("DB_HOST", "localhost")),
		config.WithPort(getEnvInt("DB_PORT", 5432)),
		config.WithDatabase(getEnv("DB_NAME", "example")),
		config.WithUsername(getEnv("DB_USER", "postgres")),
		config.WithPassword(getEnv("DB_PASSWORD", "password")),
		config.WithMaxConns(20),
		config.WithMinConns(5),
		config.WithConnectTimeout(30*time.Second),
		config.WithQueryTimeout(60*time.Second), // Timeout maior para operações em lote
		config.WithMaxConnLifetime(1*time.Hour),
		config.WithMaxConnIdleTime(15*time.Minute),
	)

	// Criar pool de conexões
	pool, err := provider.CreatePool(ctx, cfg)
	if err != nil {
		log.Fatalf("Erro ao criar pool: %v", err)
	}
	defer pool.Close()

	// Testar conexão
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Erro ao conectar com banco: %v", err)
	}

	fmt.Println("✅ Conectado ao PostgreSQL com sucesso!")

	// Criar estrutura de dados para testes
	if err := createBatchTables(ctx, pool); err != nil {
		log.Fatalf("Erro ao criar tabelas: %v", err)
	}

	// Demonstrar diferentes cenários de batch operations
	scenarios := []struct {
		name        string
		recordCount int
		description string
	}{
		{
			name:        "Pequeno Lote",
			recordCount: 1000,
			description: "Inserção de 1.000 registros",
		},
		{
			name:        "Lote Médio",
			recordCount: 10000,
			description: "Inserção de 10.000 registros",
		},
		{
			name:        "Lote Grande",
			recordCount: 50000,
			description: "Inserção de 50.000 registros",
		},
	}

	// Executar cenários
	for i, scenario := range scenarios {
		fmt.Printf("\n🎯 Cenário %d: %s\n", i+1, scenario.name)
		fmt.Printf("📝 %s\n", scenario.description)
		fmt.Println("─────────────────────────────────────────")

		if err := demonstrateBatchScenario(ctx, pool, scenario.recordCount); err != nil {
			log.Printf("❌ Erro no cenário %d: %v", i+1, err)
			continue
		}

		// Limpar dados entre cenários
		if err := cleanupBatchData(ctx, pool); err != nil {
			log.Printf("⚠️ Erro na limpeza: %v", err)
		}

		// Aguardar entre cenários
		if i < len(scenarios)-1 {
			fmt.Printf("\n⏳ Aguardando 3 segundos antes do próximo cenário...\n")
			time.Sleep(3 * time.Second)
		}
	}

	// Demonstração final com diferentes estratégias de inserção
	fmt.Printf("\n🏆 Demonstração Final: Comparação de Estratégias\n")
	fmt.Println("═══════════════════════════════════════════════════")

	if err := compareInsertStrategies(ctx, pool); err != nil {
		log.Fatalf("❌ Erro na comparação de estratégias: %v", err)
	}

	fmt.Println("\n🎉 Demonstração de operações em lote concluída!")
}

// createBatchTables cria as tabelas necessárias para os testes
func createBatchTables(ctx context.Context, pool postgresql.IPool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	queries := []string{
		`DROP TABLE IF EXISTS batch_orders CASCADE`,
		`DROP TABLE IF EXISTS batch_products CASCADE`,
		`DROP TABLE IF EXISTS batch_customers CASCADE`,
		`
		CREATE TABLE batch_customers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(150) UNIQUE NOT NULL,
			phone VARCHAR(20),
			city VARCHAR(50),
			country VARCHAR(50),
			registration_date DATE DEFAULT CURRENT_DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`
		CREATE TABLE batch_products (
			id SERIAL PRIMARY KEY,
			sku VARCHAR(50) UNIQUE NOT NULL,
			name VARCHAR(200) NOT NULL,
			category VARCHAR(50),
			price DECIMAL(10,2) NOT NULL,
			cost DECIMAL(10,2) NOT NULL,
			stock INTEGER DEFAULT 0,
			weight DECIMAL(8,3),
			dimensions VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`
		CREATE TABLE batch_orders (
			id SERIAL PRIMARY KEY,
			customer_id INTEGER REFERENCES batch_customers(id),
			product_id INTEGER REFERENCES batch_products(id),
			quantity INTEGER NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			total_price DECIMAL(12,2) NOT NULL,
			order_date DATE DEFAULT CURRENT_DATE,
			status VARCHAR(20) DEFAULT 'pending',
			shipping_address TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`
		CREATE INDEX idx_batch_customers_email ON batch_customers(email);
		CREATE INDEX idx_batch_customers_city ON batch_customers(city);
		CREATE INDEX idx_batch_products_sku ON batch_products(sku);
		CREATE INDEX idx_batch_products_category ON batch_products(category);
		CREATE INDEX idx_batch_orders_customer_id ON batch_orders(customer_id);
		CREATE INDEX idx_batch_orders_product_id ON batch_orders(product_id);
		CREATE INDEX idx_batch_orders_order_date ON batch_orders(order_date);
		CREATE INDEX idx_batch_orders_status ON batch_orders(status);
		`,
	}

	for _, query := range queries {
		if err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("erro ao executar query: %w", err)
		}
	}

	fmt.Println("✅ Tabelas para batch operations criadas com sucesso!")
	return nil
}

// demonstrateBatchScenario demonstra um cenário específico de batch operations
func demonstrateBatchScenario(ctx context.Context, pool postgresql.IPool, recordCount int) error {
	// Gerar dados
	fmt.Printf("🔧 Gerando %d registros de teste...\n", recordCount)
	customers := generateCustomers(recordCount)
	products := generateProducts(recordCount / 2) // Metade dos produtos vs clientes

	// Performance monitor
	monitor := NewPerformanceMonitor()

	// Inserir clientes em lote
	fmt.Printf("👥 Inserindo %d clientes em lote...\n", len(customers))
	if err := insertCustomersBatch(ctx, pool, customers, monitor); err != nil {
		return fmt.Errorf("erro ao inserir clientes: %w", err)
	}

	// Inserir produtos em lote
	fmt.Printf("📦 Inserindo %d produtos em lote...\n", len(products))
	if err := insertProductsBatch(ctx, pool, products, monitor); err != nil {
		return fmt.Errorf("erro ao inserir produtos: %w", err)
	}

	// Gerar pedidos baseados nos clientes e produtos inseridos
	fmt.Printf("🛒 Gerando e inserindo pedidos...\n")
	orders := generateOrders(recordCount*2, len(customers), len(products)) // Mais pedidos que clientes

	if err := insertOrdersBatch(ctx, pool, orders, monitor); err != nil {
		return fmt.Errorf("erro ao inserir pedidos: %w", err)
	}

	// Exibir estatísticas de performance
	monitor.PrintSummary()

	// Verificar dados inseridos
	if err := verifyInsertedData(ctx, pool); err != nil {
		return fmt.Errorf("erro na verificação: %w", err)
	}

	return nil
}

// insertCustomersBatch insere clientes usando batch operations
func insertCustomersBatch(ctx context.Context, pool postgresql.IPool, customers []Customer, monitor *PerformanceMonitor) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	start := time.Now()

	// Criar batch
	batch := &simpleBatch{}

	for _, customer := range customers {
		batch.Queue(
			"INSERT INTO batch_customers (name, email, phone, city, country) VALUES ($1, $2, $3, $4, $5)",
			customer.Name, customer.Email, customer.Phone, customer.City, customer.Country)
	}

	// Executar batch
	results, err := conn.SendBatch(ctx, batch)
	if err != nil {
		return fmt.Errorf("erro ao executar batch: %w", err)
	}
	defer results.Close()

	// Processar resultados
	successCount := 0
	errorCount := 0

	for i := 0; i < batch.Len(); i++ {
		if err := results.Exec(); err != nil {
			errorCount++
			log.Printf("Erro no item %d do batch: %v", i, err)
		} else {
			successCount++
		}
	}

	duration := time.Since(start)

	monitor.AddOperation("customers_batch_insert", successCount, duration)

	fmt.Printf("   ✅ Clientes inseridos: %d | Erros: %d | Tempo: %v\n",
		successCount, errorCount, duration)
	fmt.Printf("   📊 Taxa: %.2f registros/segundo\n",
		float64(successCount)/duration.Seconds())

	return nil
}

// insertProductsBatch insere produtos usando batch operations
func insertProductsBatch(ctx context.Context, pool postgresql.IPool, products []Product, monitor *PerformanceMonitor) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	start := time.Now()

	batch := &simpleBatch{}

	for _, product := range products {
		batch.Queue(
			"INSERT INTO batch_products (sku, name, category, price, cost, stock, weight, dimensions) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			product.SKU, product.Name, product.Category, product.Price,
			product.Cost, product.Stock, product.Weight, product.Dimensions)
	}

	results, err := conn.SendBatch(ctx, batch)
	if err != nil {
		return fmt.Errorf("erro ao executar batch: %w", err)
	}
	defer results.Close()

	successCount := 0
	errorCount := 0

	for i := 0; i < batch.Len(); i++ {
		if err := results.Exec(); err != nil {
			errorCount++
			log.Printf("Erro no item %d do batch: %v", i, err)
		} else {
			successCount++
		}
	}

	duration := time.Since(start)

	monitor.AddOperation("products_batch_insert", successCount, duration)

	fmt.Printf("   ✅ Produtos inseridos: %d | Erros: %d | Tempo: %v\n",
		successCount, errorCount, duration)
	fmt.Printf("   📊 Taxa: %.2f registros/segundo\n",
		float64(successCount)/duration.Seconds())

	return nil
}

// insertOrdersBatch insere pedidos usando batch operations
func insertOrdersBatch(ctx context.Context, pool postgresql.IPool, orders []Order, monitor *PerformanceMonitor) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	start := time.Now()

	batch := &simpleBatch{}

	for _, order := range orders {
		batch.Queue(
			"INSERT INTO batch_orders (customer_id, product_id, quantity, unit_price, total_price, status, shipping_address) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			order.CustomerID, order.ProductID, order.Quantity, order.UnitPrice,
			order.TotalPrice, order.Status, order.ShippingAddress)
	}

	results, err := conn.SendBatch(ctx, batch)
	if err != nil {
		return fmt.Errorf("erro ao executar batch: %w", err)
	}
	defer results.Close()

	successCount := 0
	errorCount := 0

	for i := 0; i < batch.Len(); i++ {
		if err := results.Exec(); err != nil {
			errorCount++
			log.Printf("Erro no item %d do batch: %v", i, err)
		} else {
			successCount++
		}
	}

	duration := time.Since(start)

	monitor.AddOperation("orders_batch_insert", successCount, duration)

	fmt.Printf("   ✅ Pedidos inseridos: %d | Erros: %d | Tempo: %v\n",
		successCount, errorCount, duration)
	fmt.Printf("   📊 Taxa: %.2f registros/segundo\n",
		float64(successCount)/duration.Seconds())

	return nil
}

// compareInsertStrategies compara diferentes estratégias de inserção
func compareInsertStrategies(ctx context.Context, pool postgresql.IPool) error {
	testSize := 5000
	fmt.Printf("🔬 Comparando estratégias com %d registros cada:\n", testSize)

	customers := generateCustomers(testSize)

	strategies := []struct {
		name string
		fn   func(context.Context, postgresql.IPool, []Customer) (time.Duration, error)
	}{
		{"Inserção Individual", insertCustomersIndividual},
		{"Batch Operations", insertCustomersBatchOptimized},
		{"Transação com Prepared Statement", insertCustomersTransaction},
	}

	results := make(map[string]time.Duration)

	for i, strategy := range strategies {
		fmt.Printf("\n%d. Testando: %s\n", i+1, strategy.name)

		// Limpar dados antes do teste
		if err := cleanupBatchData(ctx, pool); err != nil {
			return err
		}

		duration, err := strategy.fn(ctx, pool, customers)
		if err != nil {
			log.Printf("   ❌ Erro na estratégia %s: %v", strategy.name, err)
			continue
		}

		results[strategy.name] = duration

		fmt.Printf("   ⏱️ Tempo: %v\n", duration)
		fmt.Printf("   📊 Taxa: %.2f registros/segundo\n",
			float64(testSize)/duration.Seconds())
	}

	// Exibir comparação final
	fmt.Printf("\n🏆 Comparação Final:\n")
	fmt.Println("┌────────────────────────────────────┬────────────┬─────────────────┐")
	fmt.Println("│ Estratégia                         │ Tempo      │ Registros/seg   │")
	fmt.Println("├────────────────────────────────────┼────────────┼─────────────────┤")

	for strategy, duration := range results {
		rate := float64(testSize) / duration.Seconds()
		fmt.Printf("│ %-34s │ %10v │ %15.2f │\n", strategy, duration, rate)
	}

	fmt.Println("└────────────────────────────────────┴────────────┴─────────────────┘")

	return nil
}

// verifyInsertedData verifica os dados inseridos
func verifyInsertedData(ctx context.Context, pool postgresql.IPool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	tables := []string{"batch_customers", "batch_products", "batch_orders"}

	fmt.Printf("\n📊 Verificação dos dados inseridos:\n")

	for _, table := range tables {
		var count int
		row, _ := conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
		if err := row.Scan(&count); err != nil {
			return fmt.Errorf("erro ao contar registros em %s: %w", table, err)
		}

		fmt.Printf("   📋 %s: %d registros\n", table, count)
	}

	return nil
}

// cleanupBatchData limpa os dados dos testes
func cleanupBatchData(ctx context.Context, pool postgresql.IPool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	queries := []string{
		"TRUNCATE batch_orders CASCADE",
		"TRUNCATE batch_products CASCADE",
		"TRUNCATE batch_customers CASCADE",
		"ALTER SEQUENCE batch_customers_id_seq RESTART WITH 1",
		"ALTER SEQUENCE batch_products_id_seq RESTART WITH 1",
		"ALTER SEQUENCE batch_orders_id_seq RESTART WITH 1",
	}

	for _, query := range queries {
		if err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("erro na limpeza: %w", err)
		}
	}

	return nil
}

// simpleBatch implementação simples de IBatch para o exemplo
type simpleBatch struct {
	queries []string
	args    [][]interface{}
}

func (b *simpleBatch) Queue(query string, args ...interface{}) {
	b.queries = append(b.queries, query)
	b.args = append(b.args, args)
}

func (b *simpleBatch) Len() int {
	return len(b.queries)
}

func (b *simpleBatch) Clear() {
	b.queries = b.queries[:0]
	b.args = b.args[:0]
}

// Funções utilitárias
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d"); err == nil && intValue == 1 {
			var result int
			fmt.Sscanf(value, "%d", &result)
			return result
		}
	}
	return defaultValue
}
