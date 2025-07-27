package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	_ "github.com/lib/pq" // Driver PostgreSQL
)

// User representa um usuário no banco de dados
type User struct {
	ID         int       `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	Department string    `json:"department" db:"department"`
	Active     bool      `json:"active" db:"active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// DepartmentStats representa estatísticas por departamento
type DepartmentStats struct {
	Department     string  `json:"department"`
	UserCount      int     `json:"user_count"`
	AvgCreatedDays float64 `json:"avg_created_days"`
}

// UserRepository gerencia operações de usuários no banco
type UserRepository struct {
	db                *sql.DB
	paginationService *pagination.PaginationService
}

func NewUserRepository(db *sql.DB) *UserRepository {
	// Configurar paginação otimizada para banco
	cfg := config.NewDefaultConfig()
	cfg.DefaultLimit = 20
	cfg.MaxLimit = 500
	cfg.DefaultSortField = "created_at"
	cfg.DefaultSortOrder = "desc"

	return &UserRepository{
		db:                db,
		paginationService: pagination.NewPaginationService(cfg),
	}
}

// GetUsers retorna usuários paginados do banco de dados
func (r *UserRepository) GetUsers(params url.Values) (*interfaces.PaginatedResponse, error) {
	// Campos permitidos para ordenação (mapeamento seguro)
	sortableFields := []string{"id", "name", "email", "department", "created_at", "updated_at"}

	// Parse e validação de parâmetros
	paginationParams, err := r.paginationService.ParseRequest(params, sortableFields...)
	if err != nil {
		return nil, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Query base para buscar usuários
	baseQuery := `
		SELECT id, name, email, department, active, created_at, updated_at 
		FROM users 
		WHERE active = true`

	// Construir query completa com paginação
	finalQuery := r.paginationService.BuildQuery(baseQuery, paginationParams)
	countQuery := r.paginationService.BuildCountQuery(baseQuery)

	fmt.Printf("🔍 Executando query: %s\n", finalQuery)
	fmt.Printf("📊 Executando count: %s\n", countQuery)

	// Executar query de contagem
	var totalRecords int
	err = r.db.QueryRow(countQuery).Scan(&totalRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	fmt.Printf("📈 Total de registros encontrados: %d\n", totalRecords)

	// Executar query principal
	rows, err := r.db.Query(finalQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Processar resultados
	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Department,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}

	fmt.Printf("👥 Usuários carregados: %d\n", len(users))

	// Criar resposta paginada
	response := r.paginationService.CreateResponse(users, paginationParams, totalRecords)
	return response, nil
}

// GetUsersByDepartment retorna usuários filtrados por departamento
func (r *UserRepository) GetUsersByDepartment(department string, params url.Values) (*interfaces.PaginatedResponse, error) {
	sortableFields := []string{"id", "name", "email", "created_at"}

	paginationParams, err := r.paginationService.ParseRequest(params, sortableFields...)
	if err != nil {
		return nil, err
	}

	// Query com filtro por departamento
	baseQuery := `
		SELECT id, name, email, department, active, created_at, updated_at 
		FROM users 
		WHERE active = true AND department = $1`

	// Construir queries
	finalQuery := r.paginationService.BuildQuery(baseQuery, paginationParams)
	countQuery := r.paginationService.BuildCountQuery(baseQuery)

	fmt.Printf("🏢 Buscando usuários do departamento: %s\n", department)

	// Count com parâmetro
	var totalRecords int
	err = r.db.QueryRow(countQuery, department).Scan(&totalRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Query principal com parâmetro
	rows, err := r.db.Query(finalQuery, department)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Department,
			&user.Active, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return r.paginationService.CreateResponse(users, paginationParams, totalRecords), nil
}

// SearchUsers implementa busca com filtros complexos
func (r *UserRepository) SearchUsers(searchTerm string, params url.Values) (*interfaces.PaginatedResponse, error) {
	sortableFields := []string{"id", "name", "email", "department", "created_at"}

	paginationParams, err := r.paginationService.ParseRequest(params, sortableFields...)
	if err != nil {
		return nil, err
	}

	// Query com busca textual (usando ILIKE para PostgreSQL)
	baseQuery := `
		SELECT id, name, email, department, active, created_at, updated_at 
		FROM users 
		WHERE active = true 
		AND (name ILIKE $1 OR email ILIKE $1 OR department ILIKE $1)`

	searchPattern := "%" + searchTerm + "%"

	finalQuery := r.paginationService.BuildQuery(baseQuery, paginationParams)
	countQuery := r.paginationService.BuildCountQuery(baseQuery)

	fmt.Printf("🔎 Buscando por: %s\n", searchTerm)

	// Executar queries
	var totalRecords int
	err = r.db.QueryRow(countQuery, searchPattern).Scan(&totalRecords)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(finalQuery, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Department,
			&user.Active, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return r.paginationService.CreateResponse(users, paginationParams, totalRecords), nil
}

// GetUserStats retorna estatísticas agregadas com paginação
func (r *UserRepository) GetUserStats(params url.Values) (*interfaces.PaginatedResponse, error) {
	sortableFields := []string{"department", "user_count", "avg_created_days"}

	paginationParams, err := r.paginationService.ParseRequest(params, sortableFields...)
	if err != nil {
		return nil, err
	}

	// Query agregada complexa
	baseQuery := `
		SELECT 
			department,
			COUNT(*) as user_count,
			AVG(EXTRACT(DAYS FROM NOW() - created_at)) as avg_created_days
		FROM users 
		WHERE active = true 
		GROUP BY department 
		HAVING COUNT(*) > 0`

	finalQuery := r.paginationService.BuildQuery(baseQuery, paginationParams)
	countQuery := r.paginationService.BuildCountQuery(baseQuery)

	fmt.Printf("📊 Calculando estatísticas por departamento\n")

	// Para queries agregadas, o count é diferente
	var totalRecords int
	err = r.db.QueryRow(countQuery).Scan(&totalRecords)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(finalQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []DepartmentStats
	for rows.Next() {
		var stat DepartmentStats
		err := rows.Scan(&stat.Department, &stat.UserCount, &stat.AvgCreatedDays)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return r.paginationService.CreateResponse(stats, paginationParams, totalRecords), nil
}

func createTestDatabase() (*sql.DB, error) {
	// Conexão de exemplo (ajuste conforme sua configuração)
	// Em produção, use variáveis de ambiente
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=pagination_test sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Testar conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func setupTestData(db *sql.DB) error {
	// Criar tabela se não existir
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		department VARCHAR(100) NOT NULL,
		active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Verificar se já tem dados
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		fmt.Printf("✅ Tabela já possui %d registros\n", count)
		return nil
	}

	// Inserir dados de teste
	fmt.Println("📝 Inserindo dados de teste...")

	testUsers := []struct {
		name       string
		email      string
		department string
	}{
		{"Alice Silva", "alice@company.com", "Engineering"},
		{"Bob Santos", "bob@company.com", "Engineering"},
		{"Carol Oliveira", "carol@company.com", "Marketing"},
		{"David Costa", "david@company.com", "Sales"},
		{"Eva Lima", "eva@company.com", "Engineering"},
		{"Frank Pereira", "frank@company.com", "HR"},
		{"Grace Moura", "grace@company.com", "Marketing"},
		{"Henry Alves", "henry@company.com", "Sales"},
		{"Ivy Ferreira", "ivy@company.com", "Engineering"},
		{"Jack Barbosa", "jack@company.com", "Finance"},
		{"Kelly Rocha", "kelly@company.com", "Marketing"},
		{"Liam Cardoso", "liam@company.com", "Engineering"},
		{"Mia Teixeira", "mia@company.com", "HR"},
		{"Noah Ribeiro", "noah@company.com", "Sales"},
		{"Olivia Gomes", "olivia@company.com", "Finance"},
		{"Paul Dias", "paul@company.com", "Engineering"},
		{"Quinn Souza", "quinn@company.com", "Marketing"},
		{"Ruby Martins", "ruby@company.com", "Sales"},
		{"Sam Nascimento", "sam@company.com", "HR"},
		{"Tina Cavalcanti", "tina@company.com", "Finance"},
	}

	for _, user := range testUsers {
		_, err := db.Exec(
			"INSERT INTO users (name, email, department) VALUES ($1, $2, $3)",
			user.name, user.email, user.department,
		)
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %w", user.name, err)
		}
	}

	fmt.Printf("✅ Inseridos %d usuários de teste\n", len(testUsers))
	return nil
}

func demonstrateBasicPagination(repo *UserRepository) {
	fmt.Println("\n=== 1. Paginação Básica ===")

	// Buscar primeira página
	params := url.Values{
		"page":  []string{"1"},
		"limit": []string{"5"},
		"sort":  []string{"name"},
		"order": []string{"asc"},
	}

	response, err := repo.GetUsers(params)
	if err != nil {
		fmt.Printf("❌ Erro: %v\n", err)
		return
	}

	fmt.Printf("📄 Página %d de %d\n", response.Metadata.CurrentPage, response.Metadata.TotalPages)
	fmt.Printf("� Total de usuários: %d\n", response.Metadata.TotalRecords)

	// Type assertion segura
	if users, ok := response.Content.([]User); ok {
		fmt.Printf("� Usuários nesta página: %d\n", len(users))

		// Mostrar alguns usuários
		for i, user := range users {
			if i < 3 { // Mostrar apenas os 3 primeiros
				fmt.Printf("   %d. %s <%s> [%s]\n", user.ID, user.Name, user.Email, user.Department)
			}
		}
		if len(users) > 3 {
			fmt.Printf("   ... e mais %d usuários\n", len(users)-3)
		}
	} else {
		fmt.Printf("   ⚠️  Tipo inesperado de dados retornados\n")
	}
}

func demonstrateDepartmentFilter(repo *UserRepository) {
	fmt.Println("\n=== 2. Filtro por Departamento ===")

	params := url.Values{
		"page":  []string{"1"},
		"limit": []string{"10"},
		"sort":  []string{"created_at"},
		"order": []string{"desc"},
	}

	response, err := repo.GetUsersByDepartment("Engineering", params)
	if err != nil {
		fmt.Printf("❌ Erro: %v\n", err)
		return
	}

	fmt.Printf("🏢 Usuários do departamento Engineering: %d\n", response.Metadata.TotalRecords)

	if users, ok := response.Content.([]User); ok {
		for _, user := range users {
			fmt.Printf("   • %s <%s>\n", user.Name, user.Email)
		}
	} else {
		fmt.Printf("   ⚠️  Tipo inesperado de dados retornados\n")
	}
}

func demonstrateSearch(repo *UserRepository) {
	fmt.Println("\n=== 3. Busca Textual ===")

	params := url.Values{
		"page":  []string{"1"},
		"limit": []string{"5"},
		"sort":  []string{"name"},
	}

	// Buscar por "silva"
	response, err := repo.SearchUsers("silva", params)
	if err != nil {
		fmt.Printf("❌ Erro: %v\n", err)
		return
	}

	fmt.Printf("🔍 Resultados para 'silva': %d\n", response.Metadata.TotalRecords)

	if users, ok := response.Content.([]User); ok {
		for _, user := range users {
			fmt.Printf("   ✅ %s <%s> [%s]\n", user.Name, user.Email, user.Department)
		}
	} else {
		fmt.Printf("   ⚠️  Tipo inesperado de dados retornados\n")
	}
}

func demonstrateAggregation(repo *UserRepository) {
	fmt.Println("\n=== 4. Consultas Agregadas ===")

	params := url.Values{
		"page":  []string{"1"},
		"limit": []string{"10"},
		"sort":  []string{"user_count"},
		"order": []string{"desc"},
	}

	response, err := repo.GetUserStats(params)
	if err != nil {
		fmt.Printf("❌ Erro: %v\n", err)
		return
	}

	fmt.Printf("📊 Estatísticas por departamento:\n")

	// Type assertion para o tipo correto
	if stats, ok := response.Content.([]DepartmentStats); ok {
		for _, stat := range stats {
			fmt.Printf("   📈 %s: %d usuários (média %.1f dias)\n",
				stat.Department, stat.UserCount, stat.AvgCreatedDays)
		}
	} else {
		fmt.Printf("   ⚠️  Tipo inesperado de dados retornados\n")
	}
}

func main() {
	fmt.Println("🗄️  Exemplos de Integração com Banco de Dados - Módulo de Paginação")
	fmt.Println("===================================================================")

	// Conectar ao banco
	fmt.Println("\n📡 Conectando ao banco de dados...")
	db, err := createTestDatabase()
	if err != nil {
		log.Printf("❌ Erro ao conectar ao banco: %v", err)
		log.Println("💡 Certifique-se de que o PostgreSQL está rodando e configure a DSN corretamente")
		log.Println("💡 Exemplo de DSN: host=localhost port=5432 user=postgres password=postgres dbname=pagination_test sslmode=disable")
		return
	}
	defer db.Close()

	fmt.Println("✅ Conexão com banco estabelecida")

	// Configurar dados de teste
	if err := setupTestData(db); err != nil {
		log.Fatalf("❌ Erro ao configurar dados de teste: %v", err)
	}

	// Criar repositório
	repo := NewUserRepository(db)

	// Executar demonstrações
	demonstrateBasicPagination(repo)
	demonstrateDepartmentFilter(repo)
	demonstrateSearch(repo)
	demonstrateAggregation(repo)

	fmt.Println("\n🎉 Todos os exemplos de integração com banco foram executados!")
	fmt.Println()
	fmt.Println("💡 Principais aprendizados:")
	fmt.Println("   • Paginação otimizada com LIMIT/OFFSET")
	fmt.Println("   • Queries de contagem automáticas")
	fmt.Println("   • Filtros complexos com parâmetros")
	fmt.Println("   • Busca textual com ILIKE")
	fmt.Println("   • Consultas agregadas paginadas")
	fmt.Println("   • Mapeamento seguro de campos de ordenação")
}
