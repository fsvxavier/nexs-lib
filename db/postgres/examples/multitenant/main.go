package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres"
)

func main() {
	fmt.Println("=== Exemplo de Multi-Tenancy ===")

	// Configuração da conexão
	connectionString := "postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Conectar ao banco
	fmt.Println("\n1. Conectando ao banco...")
	conn, err := postgres.Connect(ctx, connectionString)
	if err != nil {
		log.Printf("💡 Exemplo de Multi-Tenancy seria executado com banco real: %v", err)
		demonstrateMultiTenancyConceptually()
		return
	}
	defer conn.Close(ctx)

	// 2. Configurar estrutura multi-tenant
	fmt.Println("2. Configurando estrutura multi-tenant...")
	if err := setupMultiTenantStructure(ctx, conn); err != nil {
		log.Fatalf("Erro ao configurar estrutura: %v", err)
	}

	// 3. Exemplo: Schema-based multi-tenancy
	fmt.Println("\n3. Exemplo: Schema-based multi-tenancy...")
	if err := demonstrateSchemaBasedMultiTenancy(ctx, conn); err != nil {
		log.Printf("Erro no exemplo schema-based: %v", err)
	}

	// 4. Exemplo: Row-level multi-tenancy
	fmt.Println("\n4. Exemplo: Row-level multi-tenancy...")
	if err := demonstrateRowLevelMultiTenancy(ctx, conn); err != nil {
		log.Printf("Erro no exemplo row-level: %v", err)
	}

	// 5. Exemplo: Database-level multi-tenancy
	fmt.Println("\n5. Exemplo: Database-level multi-tenancy...")
	if err := demonstrateDatabaseLevelMultiTenancy(ctx, conn); err != nil {
		log.Printf("Erro no exemplo database-level: %v", err)
	}

	// 6. Exemplo: Tenant isolation e security
	fmt.Println("\n6. Exemplo: Tenant isolation e security...")
	if err := demonstrateTenantIsolationSecurity(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de isolamento: %v", err)
	}

	// 7. Exemplo: Tenant management
	fmt.Println("\n7. Exemplo: Tenant management...")
	if err := demonstrateTenantManagement(ctx, conn); err != nil {
		log.Printf("Erro no exemplo de gerenciamento: %v", err)
	}

	// 8. Limpeza
	fmt.Println("\n8. Limpando recursos...")
	if err := cleanupMultiTenantResources(ctx, conn); err != nil {
		log.Printf("Erro na limpeza: %v", err)
	}

	fmt.Println("\n=== Exemplo de Multi-Tenancy - CONCLUÍDO ===")
}

func demonstrateMultiTenancyConceptually() {
	fmt.Println("\n🎯 Demonstração Conceitual de Multi-Tenancy")
	fmt.Println("==========================================")

	fmt.Println("\n💡 Conceitos fundamentais:")
	fmt.Println("  - Multi-tenancy permite múltiplos inquilinos (tenants) em uma aplicação")
	fmt.Println("  - Isolamento de dados entre tenants é fundamental")
	fmt.Println("  - Diferentes estratégias oferecem diferentes níveis de isolamento")
	fmt.Println("  - Performance e custo variam conforme a estratégia escolhida")

	fmt.Println("\n🏗️ Estratégias de Multi-Tenancy:")
	fmt.Println("  1. Schema-based: Um schema por tenant")
	fmt.Println("  2. Row-level: Coluna tenant_id nas tabelas")
	fmt.Println("  3. Database-level: Um banco por tenant")
	fmt.Println("  4. Híbrida: Combinação de estratégias")

	fmt.Println("\n⚖️ Comparação de estratégias:")
	fmt.Println("  Schema-based:")
	fmt.Println("    ✅ Bom isolamento")
	fmt.Println("    ✅ Facilita backup por tenant")
	fmt.Println("    ❌ Overhead de manutenção")
	fmt.Println("    ❌ Limite de schemas")

	fmt.Println("  Row-level:")
	fmt.Println("    ✅ Eficiente para muitos tenants")
	fmt.Println("    ✅ Fácil manutenção")
	fmt.Println("    ❌ Risco de vazamento de dados")
	fmt.Println("    ❌ Complexidade de queries")

	fmt.Println("  Database-level:")
	fmt.Println("    ✅ Isolamento máximo")
	fmt.Println("    ✅ Backup/restore individual")
	fmt.Println("    ❌ Alto custo de recursos")
	fmt.Println("    ❌ Complexidade de deployment")

	fmt.Println("\n🛠️ Casos de uso:")
	fmt.Println("  - SaaS applications")
	fmt.Println("  - Multi-client platforms")
	fmt.Println("  - Enterprise applications")
	fmt.Println("  - Cloud services")
}

func setupMultiTenantStructure(ctx context.Context, conn postgres.IConn) error {
	// Criar tabela de tenants
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tenants (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			schema_name VARCHAR(100) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT NOW(),
			active BOOLEAN DEFAULT true,
			settings JSONB DEFAULT '{}'
		)
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela de tenants: %w", err)
	}

	// Limpar dados anteriores
	_, err = conn.Exec(ctx, "DELETE FROM tenants")
	if err != nil {
		return fmt.Errorf("erro ao limpar tenants: %w", err)
	}

	fmt.Println("   ✅ Estrutura multi-tenant configurada")
	return nil
}

func demonstrateSchemaBasedMultiTenancy(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Schema-based Multi-Tenancy ===")

	// Definir tenants
	tenants := []struct {
		name       string
		schemaName string
	}{
		{"Empresa A", "tenant_empresa_a"},
		{"Empresa B", "tenant_empresa_b"},
		{"Empresa C", "tenant_empresa_c"},
	}

	// Criar schemas para cada tenant
	fmt.Println("   Criando schemas para cada tenant...")
	for _, tenant := range tenants {
		// Registrar tenant
		_, err := conn.Exec(ctx,
			"INSERT INTO tenants (name, schema_name) VALUES ($1, $2)",
			tenant.name, tenant.schemaName)
		if err != nil {
			return fmt.Errorf("erro ao registrar tenant %s: %w", tenant.name, err)
		}

		// Criar schema
		_, err = conn.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", tenant.schemaName))
		if err != nil {
			return fmt.Errorf("erro ao criar schema %s: %w", tenant.schemaName, err)
		}

		// Criar tabelas no schema do tenant
		_, err = conn.Exec(ctx, fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.users (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100) NOT NULL UNIQUE,
				created_at TIMESTAMP DEFAULT NOW()
			)
		`, tenant.schemaName))
		if err != nil {
			return fmt.Errorf("erro ao criar tabela users para %s: %w", tenant.name, err)
		}

		fmt.Printf("   ✅ Schema criado para %s: %s\n", tenant.name, tenant.schemaName)
	}

	// Inserir dados específicos para cada tenant
	fmt.Println("   Inserindo dados específicos para cada tenant...")

	tenantData := map[string][]struct {
		name  string
		email string
	}{
		"tenant_empresa_a": {
			{"João Silva", "joao@empresaa.com"},
			{"Maria Santos", "maria@empresaa.com"},
		},
		"tenant_empresa_b": {
			{"Pedro Oliveira", "pedro@empresab.com"},
			{"Ana Costa", "ana@empresab.com"},
		},
		"tenant_empresa_c": {
			{"Carlos Lima", "carlos@empresac.com"},
			{"Fernanda Rocha", "fernanda@empresac.com"},
		},
	}

	for schemaName, users := range tenantData {
		for _, user := range users {
			_, err := conn.Exec(ctx,
				fmt.Sprintf("INSERT INTO %s.users (name, email) VALUES ($1, $2)", schemaName),
				user.name, user.email)
			if err != nil {
				return fmt.Errorf("erro ao inserir usuário %s no schema %s: %w",
					user.name, schemaName, err)
			}
		}
		fmt.Printf("   ✅ Dados inseridos para schema %s\n", schemaName)
	}

	// Demonstrar isolamento: consultar dados de cada tenant
	fmt.Println("   Demonstrando isolamento de dados...")
	for _, tenant := range tenants {
		fmt.Printf("   📊 Dados do tenant %s:\n", tenant.name)

		rows, err := conn.Query(ctx,
			fmt.Sprintf("SELECT name, email FROM %s.users ORDER BY name", tenant.schemaName))
		if err != nil {
			return fmt.Errorf("erro ao consultar dados do tenant %s: %w", tenant.name, err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var name, email string
			if err := rows.Scan(&name, &email); err != nil {
				return fmt.Errorf("erro ao ler dados do tenant %s: %w", tenant.name, err)
			}
			count++
			fmt.Printf("     %d. %s (%s)\n", count, name, email)
		}

		if count == 0 {
			fmt.Printf("     (Nenhum usuário encontrado)\n")
		}
	}

	// Demonstrar mudança de contexto de tenant
	fmt.Println("   Demonstrando mudança de contexto de tenant...")
	if err := demonstrateSchemaContextSwitching(ctx, conn); err != nil {
		return fmt.Errorf("erro na mudança de contexto: %w", err)
	}

	return nil
}

func demonstrateSchemaContextSwitching(ctx context.Context, conn postgres.IConn) error {
	// Simular mudança de contexto usando search_path
	tenantSchemas := []string{"tenant_empresa_a", "tenant_empresa_b", "tenant_empresa_c"}

	for _, schema := range tenantSchemas {
		fmt.Printf("   🔄 Mudando contexto para schema: %s\n", schema)

		// Definir search_path para o tenant
		_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schema))
		if err != nil {
			return fmt.Errorf("erro ao definir search_path: %w", err)
		}

		// Consultar dados usando contexto do tenant
		rows, err := conn.Query(ctx, "SELECT COUNT(*) FROM users")
		if err != nil {
			return fmt.Errorf("erro ao consultar usuários do tenant %s: %w", schema, err)
		}

		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				rows.Close()
				return fmt.Errorf("erro ao ler contagem: %w", err)
			}
			fmt.Printf("   📊 Usuários encontrados no contexto: %d\n", count)
		}
		rows.Close() // Fechar rows imediatamente após uso
	}

	// Restaurar search_path padrão
	_, err := conn.Exec(ctx, "SET search_path TO public")
	if err != nil {
		return fmt.Errorf("erro ao restaurar search_path: %w", err)
	}

	return nil
}

func demonstrateRowLevelMultiTenancy(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Row-level Multi-Tenancy ===")

	// Criar tabela com coluna tenant_id
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS shared_users (
			id SERIAL PRIMARY KEY,
			tenant_id INTEGER NOT NULL REFERENCES tenants(id),
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(tenant_id, email)
		)
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela shared_users: %w", err)
	}

	// Obter IDs dos tenants
	tenantIDs := make(map[string]int)
	rows, err := conn.Query(ctx, "SELECT id, name FROM tenants ORDER BY id")
	if err != nil {
		return fmt.Errorf("erro ao consultar tenants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return fmt.Errorf("erro ao ler tenant: %w", err)
		}
		tenantIDs[name] = id
	}

	// Inserir dados com tenant_id
	fmt.Println("   Inserindo dados com tenant_id...")

	userData := []struct {
		tenantName string
		name       string
		email      string
	}{
		{"Empresa A", "João Silva", "joao@empresaa.com"},
		{"Empresa A", "Maria Santos", "maria@empresaa.com"},
		{"Empresa B", "Pedro Oliveira", "pedro@empresab.com"},
		{"Empresa B", "Ana Costa", "ana@empresab.com"},
		{"Empresa C", "Carlos Lima", "carlos@empresac.com"},
		{"Empresa C", "Fernanda Rocha", "fernanda@empresac.com"},
	}

	for _, user := range userData {
		tenantID, exists := tenantIDs[user.tenantName]
		if !exists {
			return fmt.Errorf("tenant %s não encontrado", user.tenantName)
		}

		_, err := conn.Exec(ctx,
			"INSERT INTO shared_users (tenant_id, name, email) VALUES ($1, $2, $3)",
			tenantID, user.name, user.email)
		if err != nil {
			return fmt.Errorf("erro ao inserir usuário %s: %w", user.name, err)
		}
	}

	fmt.Println("   ✅ Dados inseridos com tenant_id")

	// Demonstrar queries com filtro por tenant
	fmt.Println("   Demonstrando queries com filtro por tenant...")

	for tenantName, tenantID := range tenantIDs {
		fmt.Printf("   📊 Usuários do tenant %s (ID: %d):\n", tenantName, tenantID)

		rows, err := conn.Query(ctx,
			"SELECT name, email FROM shared_users WHERE tenant_id = $1 ORDER BY name",
			tenantID)
		if err != nil {
			return fmt.Errorf("erro ao consultar usuários do tenant %s: %w", tenantName, err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var name, email string
			if err := rows.Scan(&name, &email); err != nil {
				return fmt.Errorf("erro ao ler usuário: %w", err)
			}
			count++
			fmt.Printf("     %d. %s (%s)\n", count, name, email)
		}

		if count == 0 {
			fmt.Printf("     (Nenhum usuário encontrado)\n")
		}
	}

	// Demonstrar Row Level Security (RLS)
	fmt.Println("   Demonstrando Row Level Security (RLS)...")
	if err := demonstrateRowLevelSecurity(ctx, conn); err != nil {
		return fmt.Errorf("erro ao demonstrar RLS: %w", err)
	}

	return nil
}

func demonstrateRowLevelSecurity(ctx context.Context, conn postgres.IConn) error {
	// Habilitar RLS na tabela
	_, err := conn.Exec(ctx, "ALTER TABLE shared_users ENABLE ROW LEVEL SECURITY")
	if err != nil {
		return fmt.Errorf("erro ao habilitar RLS: %w", err)
	}

	// Criar política de RLS
	_, err = conn.Exec(ctx, `
		CREATE POLICY tenant_isolation ON shared_users
		FOR ALL
		USING (tenant_id = current_setting('app.current_tenant_id')::integer)
	`)
	if err != nil {
		// Política pode já existir, ignorar erro
		fmt.Printf("   ⚠️ Política RLS: %v\n", err)
	}

	// Simular configuração de tenant atual
	tenantID := 1 // Empresa A
	_, err = conn.Exec(ctx, fmt.Sprintf("SET app.current_tenant_id = '%d'", tenantID))
	if err != nil {
		return fmt.Errorf("erro ao definir tenant atual: %w", err)
	}

	fmt.Printf("   🔐 RLS configurado para tenant_id: %d\n", tenantID)
	fmt.Println("   📊 Consulta com RLS ativo:")

	// Consultar dados com RLS ativo
	rows, err := conn.Query(ctx, "SELECT name, email FROM shared_users ORDER BY name")
	if err != nil {
		return fmt.Errorf("erro ao consultar com RLS: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var name, email string
		if err := rows.Scan(&name, &email); err != nil {
			return fmt.Errorf("erro ao ler com RLS: %w", err)
		}
		count++
		fmt.Printf("     %d. %s (%s)\n", count, name, email)
	}

	fmt.Printf("   📊 Registros visíveis com RLS: %d\n", count)

	// Limpar configuração
	_, err = conn.Exec(ctx, "RESET app.current_tenant_id")
	if err != nil {
		return fmt.Errorf("erro ao resetar tenant_id: %w", err)
	}

	return nil
}

func demonstrateDatabaseLevelMultiTenancy(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Database-level Multi-Tenancy ===")

	// Demonstrar conceito (não criaremos DBs reais)
	fmt.Println("   💡 Conceito: Database-level Multi-Tenancy")
	fmt.Println("   - Cada tenant tem seu próprio banco de dados")
	fmt.Println("   - Isolamento máximo entre tenants")
	fmt.Println("   - Backup/restore individual por tenant")
	fmt.Println("   - Escalabilidade horizontal por tenant")

	// Simular configuração de connection strings por tenant
	tenantDatabases := map[string]string{
		"tenant_empresa_a": "nexs_tenant_empresa_a",
		"tenant_empresa_b": "nexs_tenant_empresa_b",
		"tenant_empresa_c": "nexs_tenant_empresa_c",
	}

	fmt.Println("   📊 Configuração simulada de bancos por tenant:")
	for tenant, dbName := range tenantDatabases {
		fmt.Printf("   - %s: %s\n", tenant, dbName)
	}

	// Demonstrar roteamento de tenant
	fmt.Println("   🔄 Simulando roteamento de tenant...")
	currentTenant := "tenant_empresa_a"
	targetDB := tenantDatabases[currentTenant]

	fmt.Printf("   Cliente requisita acesso ao tenant: %s\n", currentTenant)
	fmt.Printf("   Sistema roteia para banco: %s\n", targetDB)
	fmt.Printf("   Connection string: postgres://user:pass@localhost:5432/%s\n", targetDB)

	// Demonstrar vantagens e desvantagens
	fmt.Println("   ✅ Vantagens:")
	fmt.Println("     - Isolamento completo")
	fmt.Println("     - Backup individual")
	fmt.Println("     - Escalabilidade horizontal")
	fmt.Println("     - Configuração por tenant")

	fmt.Println("   ❌ Desvantagens:")
	fmt.Println("     - Alto custo de recursos")
	fmt.Println("     - Complexidade de gerenciamento")
	fmt.Println("     - Overhead de conexões")
	fmt.Println("     - Dificuldade em relatórios cross-tenant")

	return nil
}

func demonstrateTenantIsolationSecurity(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Tenant Isolation & Security ===")

	// Demonstrar validação de tenant
	fmt.Println("   🔐 Validação de tenant:")

	// Simular middleware de validação
	fmt.Println("   Simulando middleware de validação...")
	requestTenantID := 1 // Vindo do JWT token ou header

	// Validar se tenant existe e está ativo
	var tenantExists bool
	var tenantActive bool
	var tenantName string

	err := conn.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1), "+
			"COALESCE((SELECT active FROM tenants WHERE id = $1), false), "+
			"COALESCE((SELECT name FROM tenants WHERE id = $1), '')",
		requestTenantID).Scan(&tenantExists, &tenantActive, &tenantName)

	if err != nil {
		return fmt.Errorf("erro ao validar tenant: %w", err)
	}

	if !tenantExists {
		fmt.Printf("   ❌ Tenant ID %d não existe\n", requestTenantID)
		return fmt.Errorf("tenant inválido")
	}

	if !tenantActive {
		fmt.Printf("   ❌ Tenant '%s' está inativo\n", tenantName)
		return fmt.Errorf("tenant inativo")
	}

	fmt.Printf("   ✅ Tenant '%s' (ID: %d) validado com sucesso\n", tenantName, requestTenantID)

	// Demonstrar sanitização de dados
	fmt.Println("   🧹 Sanitização de dados:")
	fmt.Println("   - Filtros automáticos por tenant_id")
	fmt.Println("   - Validação de permissões")
	fmt.Println("   - Prevenção de SQL injection")

	// Simular query segura
	fmt.Println("   Executando query segura...")
	userEmail := "joao@empresaa.com"

	var userID int
	var userName string
	err = conn.QueryRow(ctx,
		"SELECT id, name FROM shared_users WHERE tenant_id = $1 AND email = $2",
		requestTenantID, userEmail).Scan(&userID, &userName)

	if err != nil {
		fmt.Printf("   ❌ Usuário não encontrado ou sem permissão\n")
	} else {
		fmt.Printf("   ✅ Usuário encontrado: %s (ID: %d)\n", userName, userID)
	}

	// Demonstrar auditoria
	fmt.Println("   📋 Auditoria de acesso:")
	fmt.Printf("   - Tenant: %s (ID: %d)\n", tenantName, requestTenantID)
	fmt.Printf("   - Usuário acessado: %s\n", userEmail)
	fmt.Printf("   - Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("   - IP: 192.168.1.100 (simulado)\n")

	return nil
}

func demonstrateTenantManagement(ctx context.Context, conn postgres.IConn) error {
	fmt.Println("=== Tenant Management ===")

	// Listar todos os tenants
	fmt.Println("   📋 Listando todos os tenants:")
	rows, err := conn.Query(ctx, `
		SELECT id, name, schema_name, created_at, active 
		FROM tenants 
		ORDER BY created_at
	`)
	if err != nil {
		return fmt.Errorf("erro ao listar tenants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, schemaName string
		var createdAt time.Time
		var active bool

		if err := rows.Scan(&id, &name, &schemaName, &createdAt, &active); err != nil {
			return fmt.Errorf("erro ao ler tenant: %w", err)
		}

		status := "✅ Ativo"
		if !active {
			status = "❌ Inativo"
		}

		fmt.Printf("   %d. %s (%s) - %s - Criado: %s\n",
			id, name, schemaName, status, createdAt.Format("2006-01-02 15:04:05"))
	}

	// Demonstrar estatísticas por tenant
	fmt.Println("   📊 Estatísticas por tenant:")

	// Estatísticas de schema-based
	fmt.Println("   Schema-based tenants:")
	for _, schema := range []string{"tenant_empresa_a", "tenant_empresa_b", "tenant_empresa_c"} {
		var count int
		err := conn.QueryRow(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s.users", schema)).Scan(&count)
		if err != nil {
			fmt.Printf("     %s: Erro ao consultar (%v)\n", schema, err)
		} else {
			fmt.Printf("     %s: %d usuários\n", schema, count)
		}
	}

	// Estatísticas de row-level
	fmt.Println("   Row-level tenants:")
	statRows, err := conn.Query(ctx, `
		SELECT t.name, COUNT(su.id) as user_count
		FROM tenants t
		LEFT JOIN shared_users su ON t.id = su.tenant_id
		GROUP BY t.id, t.name
		ORDER BY t.name
	`)
	if err != nil {
		return fmt.Errorf("erro ao consultar estatísticas: %w", err)
	}
	defer statRows.Close()

	for statRows.Next() {
		var tenantName string
		var userCount int
		if err := statRows.Scan(&tenantName, &userCount); err != nil {
			return fmt.Errorf("erro ao ler estatísticas: %w", err)
		}
		fmt.Printf("     %s: %d usuários\n", tenantName, userCount)
	}

	// Demonstrar operações de manutenção
	fmt.Println("   🔧 Operações de manutenção:")
	fmt.Println("   - Backup automático por tenant")
	fmt.Println("   - Limpeza de dados antigos")
	fmt.Println("   - Monitoramento de uso")
	fmt.Println("   - Migração de schema")

	return nil
}

func cleanupMultiTenantResources(ctx context.Context, conn postgres.IConn) error {
	// Limpar dados
	_, err := conn.Exec(ctx, "DELETE FROM shared_users")
	if err != nil {
		fmt.Printf("   ❌ Erro ao limpar shared_users: %v\n", err)
	}

	// Desabilitar RLS
	_, err = conn.Exec(ctx, "ALTER TABLE shared_users DISABLE ROW LEVEL SECURITY")
	if err != nil {
		fmt.Printf("   ❌ Erro ao desabilitar RLS: %v\n", err)
	}

	// Remover política RLS
	_, err = conn.Exec(ctx, "DROP POLICY IF EXISTS tenant_isolation ON shared_users")
	if err != nil {
		fmt.Printf("   ❌ Erro ao remover política RLS: %v\n", err)
	}

	// Remover schemas
	schemas := []string{"tenant_empresa_a", "tenant_empresa_b", "tenant_empresa_c"}
	for _, schema := range schemas {
		_, err = conn.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
		if err != nil {
			fmt.Printf("   ❌ Erro ao remover schema %s: %v\n", schema, err)
		}
	}

	// Remover tabelas
	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS shared_users")
	if err != nil {
		fmt.Printf("   ❌ Erro ao remover shared_users: %v\n", err)
	}

	_, err = conn.Exec(ctx, "DROP TABLE IF EXISTS tenants")
	if err != nil {
		fmt.Printf("   ❌ Erro ao remover tenants: %v\n", err)
	}

	fmt.Println("   ✅ Recursos multi-tenant limpos")
	return nil
}
