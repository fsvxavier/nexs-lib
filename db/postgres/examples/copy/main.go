package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
	pgxprovider "github.com/fsvxavier/nexs-lib/db/postgres/providers/pgx"
	"github.com/jackc/pgx/v5"
)

func main() {
	fmt.Println("=== Exemplo de Copy Operations ===")

	// Configuração da conexão
	connectionString := "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Conectar ao banco
	fmt.Println("\n1. Conectando ao banco...")
	conn, err := postgres.Connect(ctx, connectionString)
	if err != nil {
		log.Printf("💡 Exemplo de Copy Operations seria executado com banco real: %v", err)
		demonstrateCopyConceptually()
		return
	}
	defer conn.Close(ctx)

	// 2. Criar tabela de teste
	fmt.Println("2. Criando tabela de teste...")
	err = createTestTable(ctx, conn)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}

	// 3. Exemplo: COPY FROM básico
	fmt.Println("\n3. Exemplo: COPY FROM básico...")
	if err := demonstrateCopyFrom(ctx, conn); err != nil {
		log.Printf("Erro no exemplo COPY FROM: %v", err)
	}

	// 4. Exemplo: COPY TO básico
	fmt.Println("\n4. Exemplo: COPY TO básico...")
	if err := demonstrateCopyTo(ctx, conn); err != nil {
		log.Printf("Erro no exemplo COPY TO: %v", err)
	}

	// 5. Exemplo: COPY FROM com dados grandes
	fmt.Println("\n5. Exemplo: COPY FROM com dados grandes...")
	if err := demonstrateBulkCopyFrom(ctx, conn); err != nil {
		log.Printf("Erro no exemplo bulk COPY FROM: %v", err)
	}

	// 6. Exemplo: Performance comparison
	fmt.Println("\n6. Exemplo: Comparação de performance...")
	if err := demonstratePerformanceComparison(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de performance: %v", err)
	}

	// 7. Exemplo: Tratamento de erros
	fmt.Println("\n7. Exemplo: Tratamento de erros...")
	if err := demonstrateErrorHandling(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de tratamento de erros: %v", err)
	}

	// 8. Limpeza
	fmt.Println("\n8. Limpando recursos...")
	if err := cleanupResources(ctx, conn); err != nil {
		log.Printf("Erro na limpeza: %v", err)
	}

	fmt.Println("\n=== Exemplo de Copy Operations - CONCLUÍDO ===")
}

func demonstrateCopyConceptually() {
	fmt.Println("\n🎯 Demonstração Conceitual de Copy Operations")
	fmt.Println("=============================================")

	fmt.Println("\n💡 Conceitos fundamentais:")
	fmt.Println("  - COPY é o método mais eficiente para transferir dados em massa")
	fmt.Println("  - Bypassa muito do overhead de SQL normal")
	fmt.Println("  - Ideal para ETL e operações de data warehouse")
	fmt.Println("  - Suporta formatos CSV, TSV, Binary e Custom")

	fmt.Println("\n🔄 Tipos de operações:")
	fmt.Println("  - COPY FROM: Importa dados de uma fonte externa")
	fmt.Println("  - COPY TO: Exporta dados para um destino externo")
	fmt.Println("  - COPY ... FROM STDIN: Importa dados de input stream")
	fmt.Println("  - COPY ... TO STDOUT: Exporta dados para output stream")

	fmt.Println("\n⚡ Vantagens do COPY:")
	fmt.Println("  - 📈 Performance: 10-100x mais rápido que INSERTs individuais")
	fmt.Println("  - 💾 Eficiência: Menor uso de memória e CPU")
	fmt.Println("  - 🔄 Streaming: Processar dados maiores que a memória")
	fmt.Println("  - 🛡️ Confiabilidade: Operações transacionais")

	fmt.Println("\n🛠️ Casos de uso:")
	fmt.Println("  - Importação de CSV grandes")
	fmt.Println("  - Backup e restore de dados")
	fmt.Println("  - ETL (Extract, Transform, Load)")
	fmt.Println("  - Migração de dados entre sistemas")
	fmt.Println("  - Sincronização de dados")
}

func createTestTable(ctx context.Context, conn postgres.IConn) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS copy_test (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL,
			age INTEGER NOT NULL,
			salary DECIMAL(10,2) NOT NULL,
			department VARCHAR(50) NOT NULL,
			hire_date DATE NOT NULL,
			active BOOLEAN NOT NULL DEFAULT true
		)
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela: %w", err)
	}

	// Limpar dados anteriores
	_, err = conn.Exec(ctx, "DELETE FROM copy_test")
	if err != nil {
		return fmt.Errorf("erro ao limpar tabela: %w", err)
	}

	fmt.Println("   ✅ Tabela criada com sucesso")
	return nil
}

func demonstrateCopyFrom(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== COPY FROM Básico ===")

	// Criar dados de teste
	testData := [][]interface{}{
		{"João Silva", "joao@email.com", 30, 5000.00, "TI", "2023-01-15", true},
		{"Maria Santos", "maria@email.com", 28, 4500.00, "RH", "2023-02-20", true},
		{"Pedro Oliveira", "pedro@email.com", 35, 6000.00, "Vendas", "2023-03-10", true},
		{"Ana Costa", "ana@email.com", 32, 5500.00, "TI", "2023-04-05", true},
		{"Carlos Lima", "carlos@email.com", 29, 4800.00, "Marketing", "2023-05-12", true},
	}

	// Criar CopyFromSource usando PGX
	fmt.Printf("   Preparando %d registros para COPY FROM...\n", len(testData))

	// Usar CopyFromRows do PGX
	pgxCopySource := pgx.CopyFromRows(testData)
	copySource := pgxprovider.NewCopyFromSource(pgxCopySource)

	// Executar COPY FROM
	fmt.Println("   Executando COPY FROM...")
	startTime := time.Now()

	rowsAffected, err := conn.CopyFrom(ctx, "copy_test",
		[]string{"name", "email", "age", "salary", "department", "hire_date", "active"},
		copySource)

	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("erro no COPY FROM: %w", err)
	}

	fmt.Printf("   ✅ COPY FROM concluído em %v\n", duration)
	fmt.Printf("   📊 Linhas inseridas: %d\n", rowsAffected)

	// Verificar dados inseridos
	fmt.Println("   Verificando dados inseridos...")
	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM copy_test").Scan(&count)
	if err != nil {
		return fmt.Errorf("erro ao verificar contagem: %w", err)
	}

	fmt.Printf("   📊 Total de registros na tabela: %d\n", count)

	// Mostrar alguns registros
	fmt.Println("   Primeiros 3 registros:")
	rows, err := conn.Query(ctx, "SELECT name, email, department FROM copy_test ORDER BY id LIMIT 3")
	if err != nil {
		return fmt.Errorf("erro ao consultar registros: %w", err)
	}
	defer rows.Close()

	i := 1
	for rows.Next() {
		var name, email, department string
		if err := rows.Scan(&name, &email, &department); err != nil {
			return fmt.Errorf("erro ao ler registro: %w", err)
		}
		fmt.Printf("     %d. %s (%s) - %s\n", i, name, email, department)
		i++
	}

	return nil
}

func demonstrateCopyTo(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== COPY TO Básico ===")

	// Criar CopyToWriter
	fmt.Println("   Preparando COPY TO...")
	copyWriter := &TestCopyToWriter{
		rows: make([][]interface{}, 0),
	}

	// Executar COPY TO
	fmt.Println("   Executando COPY TO...")
	startTime := time.Now()

	err := conn.CopyTo(ctx, copyWriter, "SELECT name, email, department, salary FROM copy_test ORDER BY name")

	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("erro no COPY TO: %w", err)
	}

	fmt.Printf("   ✅ COPY TO concluído em %v\n", duration)
	fmt.Printf("   📊 Linhas exportadas: %d\n", len(copyWriter.rows))

	// Mostrar dados exportados
	fmt.Println("   Dados exportados:")
	for i, row := range copyWriter.rows {
		if i >= 5 { // Limitar exibição
			fmt.Printf("     ... e mais %d registros\n", len(copyWriter.rows)-i)
			break
		}
		fmt.Printf("     %d. %s (%s) - %s - $%.2f\n",
			i+1, row[0], row[1], row[2], row[3])
	}

	return nil
}

func demonstrateBulkCopyFrom(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== COPY FROM com Dados Grandes ===")

	// Gerar dados grandes
	bulkSize := 1000
	fmt.Printf("   Gerando %d registros para teste de bulk...\n", bulkSize)

	bulkData := make([][]interface{}, bulkSize)
	for i := 0; i < bulkSize; i++ {
		bulkData[i] = []interface{}{
			fmt.Sprintf("Usuário %d", i+1),
			fmt.Sprintf("user%d@email.com", i+1),
			25 + (i % 40),             // Idade entre 25-65
			3000.00 + float64(i%5000), // Salário entre 3000-8000
			getDepartment(i % 5),      // Departamento rotativo
			"2023-01-01",              // Data padrão
			i%10 != 0,                 // 90% ativos
		}
	}

	// Criar CopyFromSource usando PGX para dados grandes
	fmt.Printf("   Preparando %d registros para COPY FROM...\n", bulkSize)

	// Usar CopyFromRows do PGX
	pgxCopySource := pgx.CopyFromRows(bulkData)
	copySource := pgxprovider.NewCopyFromSource(pgxCopySource)

	// Executar COPY FROM
	fmt.Println("   Executando COPY FROM em massa...")
	startTime := time.Now()

	rowsAffected, err := conn.CopyFrom(ctx, "copy_test",
		[]string{"name", "email", "age", "salary", "department", "hire_date", "active"},
		copySource)

	duration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("erro no COPY FROM em massa: %w", err)
	}

	fmt.Printf("   ✅ COPY FROM em massa concluído em %v\n", duration)
	fmt.Printf("   📊 Linhas inseridas: %d\n", rowsAffected)

	// Calcular performance
	if duration > 0 {
		ratePerSecond := float64(rowsAffected) / duration.Seconds()
		fmt.Printf("   📈 Taxa de inserção: %.0f linhas/segundo\n", ratePerSecond)
	}

	// Verificar total na tabela
	var totalCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM copy_test").Scan(&totalCount)
	if err != nil {
		return fmt.Errorf("erro ao verificar contagem total: %w", err)
	}

	fmt.Printf("   📊 Total de registros na tabela: %d\n", totalCount)

	// Estatísticas por departamento
	fmt.Println("   Estatísticas por departamento:")
	rows, err := conn.Query(ctx, `
		SELECT department, COUNT(*), AVG(salary) 
		FROM copy_test 
		GROUP BY department 
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return fmt.Errorf("erro ao consultar estatísticas: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var department string
		var count int
		var avgSalary float64
		if err := rows.Scan(&department, &count, &avgSalary); err != nil {
			return fmt.Errorf("erro ao ler estatísticas: %w", err)
		}
		fmt.Printf("     %s: %d funcionários, salário médio: $%.2f\n",
			department, count, avgSalary)
	}

	return nil
}

func demonstratePerformanceComparison(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Comparação de Performance ===")

	// Limpar tabela para teste
	_, err := conn.Exec(ctx, "DELETE FROM copy_test")
	if err != nil {
		return fmt.Errorf("erro ao limpar tabela: %w", err)
	}

	// Preparar dados de teste
	testSize := 500
	testData := make([][]interface{}, testSize)
	for i := 0; i < testSize; i++ {
		testData[i] = []interface{}{
			fmt.Sprintf("Teste %d", i+1),
			fmt.Sprintf("teste%d@email.com", i+1),
			25 + (i % 40),
			3000.00 + float64(i%2000),
			getDepartment(i % 5),
			"2023-01-01",
			true,
		}
	}

	// Teste 1: INSERT individual
	fmt.Printf("   Teste 1: INSERT individual (%d registros)...\n", testSize)
	startTime := time.Now()

	for i, data := range testData {
		_, err := conn.Exec(ctx, `
			INSERT INTO copy_test (name, email, age, salary, department, hire_date, active) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, data...)
		if err != nil {
			return fmt.Errorf("erro no INSERT %d: %w", i+1, err)
		}
	}

	insertDuration := time.Since(startTime)
	fmt.Printf("   ⏱️ INSERT individual: %v\n", insertDuration)

	// Verificar contagem
	var count1 int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM copy_test").Scan(&count1)
	if err != nil {
		return fmt.Errorf("erro ao verificar contagem: %w", err)
	}

	// Limpar para próximo teste
	_, err = conn.Exec(ctx, "DELETE FROM copy_test")
	if err != nil {
		return fmt.Errorf("erro ao limpar tabela: %w", err)
	}

	// Teste 2: COPY FROM
	fmt.Printf("   Teste 2: COPY FROM (%d registros)...\n", testSize)

	// Usar CopyFromRows do PGX
	pgxCopySource := pgx.CopyFromRows(testData)
	copySource := pgxprovider.NewCopyFromSource(pgxCopySource)

	startTime = time.Now()

	_, err = conn.CopyFrom(ctx, "copy_test",
		[]string{"name", "email", "age", "salary", "department", "hire_date", "active"},
		copySource)

	copyDuration := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("erro no COPY FROM: %w", err)
	}

	fmt.Printf("   ⏱️ COPY FROM: %v\n", copyDuration)

	// Verificar contagem
	var count2 int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM copy_test").Scan(&count2)
	if err != nil {
		return fmt.Errorf("erro ao verificar contagem: %w", err)
	}

	// Análise de performance
	fmt.Println("\n   📊 Análise de Performance:")
	fmt.Printf("   - INSERT individual: %v (%d registros)\n", insertDuration, count1)
	fmt.Printf("   - COPY FROM: %v (%d registros)\n", copyDuration, count2)

	if copyDuration > 0 && insertDuration > 0 {
		speedup := float64(insertDuration) / float64(copyDuration)
		fmt.Printf("   - Speedup: %.2fx mais rápido\n", speedup)

		insertRate := float64(count1) / insertDuration.Seconds()
		copyRate := float64(count2) / copyDuration.Seconds()

		fmt.Printf("   - INSERT rate: %.0f registros/segundo\n", insertRate)
		fmt.Printf("   - COPY rate: %.0f registros/segundo\n", copyRate)
	}

	return nil
}

func demonstrateErrorHandling(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Tratamento de Erros ===")

	// Teste 1: Dados inválidos
	fmt.Println("   Teste 1: Dados inválidos...")

	invalidData := [][]interface{}{
		{"João Silva", "joao@email.com", 30, 5000.00, "TI", "2023-01-15", true},
		{"Maria Santos", "email_invalido", -5, 4500.00, "RH", "data_invalida", true}, // Dados inválidos
		{"Pedro Oliveira", "pedro@email.com", 35, 6000.00, "Vendas", "2023-03-10", true},
	}

	// Usar CopyFromRows do PGX
	pgxCopySource := pgx.CopyFromRows(invalidData)
	copySource := pgxprovider.NewCopyFromSource(pgxCopySource)

	_, err := conn.CopyFrom(ctx, "copy_test",
		[]string{"name", "email", "age", "salary", "department", "hire_date", "active"},
		copySource)

	if err != nil {
		fmt.Printf("   ✅ Erro esperado capturado: %v\n", err)
	} else {
		fmt.Println("   ❌ Era esperado um erro com dados inválidos")
	}

	// Teste 2: Tabela inexistente
	fmt.Println("   Teste 2: Tabela inexistente...")

	validData := [][]interface{}{
		{"João Silva", "joao@email.com", 30, 5000.00, "TI", "2023-01-15", true},
	}

	// Usar CopyFromRows do PGX
	pgxCopySource2 := pgx.CopyFromRows(validData)
	copySource2 := pgxprovider.NewCopyFromSource(pgxCopySource2)

	_, err = conn.CopyFrom(ctx, "tabela_inexistente",
		[]string{"name", "email", "age", "salary", "department", "hire_date", "active"},
		copySource2)

	if err != nil {
		fmt.Printf("   ✅ Erro esperado capturado: %v\n", err)
	} else {
		fmt.Println("   ❌ Era esperado um erro com tabela inexistente")
	}

	// Teste 3: Colunas incorretas
	fmt.Println("   Teste 3: Colunas incorretas...")

	// Usar CopyFromRows do PGX
	pgxCopySource3 := pgx.CopyFromRows(validData)
	copySource3 := pgxprovider.NewCopyFromSource(pgxCopySource3)

	_, err = conn.CopyFrom(ctx, "copy_test",
		[]string{"coluna_inexistente", "outra_coluna_inexistente"},
		copySource3)

	if err != nil {
		fmt.Printf("   ✅ Erro esperado capturado: %v\n", err)
	} else {
		fmt.Println("   ❌ Era esperado um erro com colunas incorretas")
	}

	fmt.Println("   ✅ Tratamento de erros concluído")
	return nil
}

func cleanupResources(ctx context.Context, conn postgres.IConn) error {
	_, err := conn.Exec(ctx, "DROP TABLE IF EXISTS copy_test")
	if err != nil {
		return fmt.Errorf("erro ao remover tabela: %w", err)
	}

	fmt.Println("   ✅ Recursos limpos com sucesso")
	return nil
}

// Estruturas de suporte

type TestCopyFromSource struct {
	data  [][]interface{}
	index int
}

func (s *TestCopyFromSource) Next() bool {
	return s.index < len(s.data)
}

func (s *TestCopyFromSource) Values() ([]interface{}, error) {
	if s.index >= len(s.data) {
		return nil, fmt.Errorf("no more data")
	}
	values := s.data[s.index]
	s.index++
	return values, nil
}

func (s *TestCopyFromSource) Err() error {
	return nil
}

type TestCopyToWriter struct {
	rows [][]interface{}
}

func (w *TestCopyToWriter) Write(row []interface{}) error {
	w.rows = append(w.rows, row)
	return nil
}

func (w *TestCopyToWriter) Close() error {
	return nil
}

// Função utilitária

func getDepartment(index int) string {
	departments := []string{"TI", "RH", "Vendas", "Marketing", "Financeiro"}
	return departments[index%len(departments)]
}
