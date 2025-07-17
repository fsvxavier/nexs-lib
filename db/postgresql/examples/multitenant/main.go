package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// TenantConfig represents configuration for a specific tenant
type TenantConfig struct {
	TenantID    string
	SchemaName  string
	DatabaseURL string
	Settings    map[string]interface{}
}

// TenantManager manages multi-tenant operations
type TenantManager struct {
	provider interfaces.PostgreSQLProvider
	tenants  map[string]*TenantConfig
	pools    map[string]interfaces.IPool
}

func NewTenantManager(provider interfaces.PostgreSQLProvider) *TenantManager {
	return &TenantManager{
		provider: provider,
		tenants:  make(map[string]*TenantConfig),
		pools:    make(map[string]interfaces.IPool),
	}
}

func (tm *TenantManager) RegisterTenant(tenantConfig *TenantConfig) error {
	tm.tenants[tenantConfig.TenantID] = tenantConfig
	return nil
}

func (tm *TenantManager) GetPoolForTenant(ctx context.Context, tenantID string) (interfaces.IPool, error) {
	if pool, exists := tm.pools[tenantID]; exists {
		return pool, nil
	}

	tenantConfig, exists := tm.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant %s not registered", tenantID)
	}

	// Create pool configuration for tenant
	cfg := postgresql.NewDefaultConfig(tenantConfig.DatabaseURL)

	// Create pool for tenant
	pool, err := tm.provider.NewPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool for tenant %s: %w", tenantID, err)
	}

	tm.pools[tenantID] = pool
	return pool, nil
}

func (tm *TenantManager) CloseAll() {
	for tenantID, pool := range tm.pools {
		fmt.Printf("🔒 Closing pool for tenant: %s\n", tenantID)
		pool.Close()
	}
}

func main() {
	// Multi-tenant PostgreSQL provider example
	ctx := context.Background()

	// Create base configuration
	cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

	if defaultCfg, ok := cfg.(*config.DefaultConfig); ok {
		err := defaultCfg.ApplyOptions(
			postgresql.WithMaxConns(20),
			postgresql.WithMinConns(5),
			postgresql.WithMultiTenant(true),
		)
		if err != nil {
			log.Fatalf("Failed to apply configuration: %v", err)
		}
	}

	// Create provider
	provider, err := postgresql.NewPGXProvider()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Example 1: Schema-based multi-tenancy
	if err := demonstrateSchemaBasedMultiTenancy(ctx, provider, cfg); err != nil {
		log.Printf("Schema-based multi-tenancy example failed: %v", err)
	}

	// Example 2: Database-based multi-tenancy
	if err := demonstrateDatabaseBasedMultiTenancy(ctx, provider); err != nil {
		log.Printf("Database-based multi-tenancy example failed: %v", err)
	}

	// Example 3: Tenant isolation and security
	if err := demonstrateTenantIsolation(ctx, provider, cfg); err != nil {
		log.Printf("Tenant isolation example failed: %v", err)
	}

	// Example 4: Tenant management operations
	if err := demonstrateTenantManagement(ctx, provider); err != nil {
		log.Printf("Tenant management example failed: %v", err)
	}

	// Example 5: Cross-tenant operations
	if err := demonstrateCrossTenantOperations(ctx, provider, cfg); err != nil {
		log.Printf("Cross-tenant operations example failed: %v", err)
	}

	fmt.Println("Multi-tenant examples completed!")
}

func demonstrateSchemaBasedMultiTenancy(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("=== Schema-based Multi-tenancy Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Schema-based multi-tenancy would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	// Define tenant schemas
	tenants := []struct {
		id     string
		schema string
		name   string
	}{
		{"tenant_001", "company_acme", "ACME Corporation"},
		{"tenant_002", "company_globex", "Globex Corporation"},
		{"tenant_003", "company_initech", "Initech"},
	}

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Create schemas for each tenant
		fmt.Println("🏢 Creating tenant schemas...")

		for _, tenant := range tenants {
			fmt.Printf("  Creating schema for %s (%s)...\n", tenant.name, tenant.id)

			// Create schema
			_, err := conn.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", tenant.schema))
			if err != nil {
				fmt.Printf("    ❌ Failed to create schema %s: %v\n", tenant.schema, err)
				continue
			}

			// Set search path to tenant schema
			_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", tenant.schema))
			if err != nil {
				fmt.Printf("    ❌ Failed to set search path: %v\n", err)
				continue
			}

			// Create tenant-specific tables
			_, err = conn.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS users (
					id SERIAL PRIMARY KEY,
					name TEXT NOT NULL,
					email TEXT UNIQUE NOT NULL,
					created_at TIMESTAMP DEFAULT NOW()
				)
			`)
			if err != nil {
				fmt.Printf("    ❌ Failed to create users table: %v\n", err)
				continue
			}

			_, err = conn.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS orders (
					id SERIAL PRIMARY KEY,
					user_id INTEGER REFERENCES users(id),
					total DECIMAL(10,2) NOT NULL,
					status TEXT DEFAULT 'pending',
					created_at TIMESTAMP DEFAULT NOW()
				)
			`)
			if err != nil {
				fmt.Printf("    ❌ Failed to create orders table: %v\n", err)
				continue
			}

			// Insert sample data for tenant
			_, err = conn.Exec(ctx,
				"INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4)",
				fmt.Sprintf("John Doe (%s)", tenant.name),
				fmt.Sprintf("john@%s.com", strings.ToLower(tenant.schema)),
				fmt.Sprintf("Jane Smith (%s)", tenant.name),
				fmt.Sprintf("jane@%s.com", strings.ToLower(tenant.schema)),
			)
			if err != nil {
				fmt.Printf("    ❌ Failed to insert sample users: %v\n", err)
				continue
			}

			// Insert sample orders
			_, err = conn.Exec(ctx,
				"INSERT INTO orders (user_id, total, status) VALUES (1, 100.50, 'completed'), (2, 75.25, 'pending')")
			if err != nil {
				fmt.Printf("    ❌ Failed to insert sample orders: %v\n", err)
				continue
			}

			fmt.Printf("    ✅ Successfully set up tenant: %s\n", tenant.name)
		}

		// Demonstrate querying data from different tenants
		fmt.Println("\n📊 Querying data from different tenant schemas...")

		for _, tenant := range tenants {
			fmt.Printf("\n  📋 Data for %s (schema: %s):\n", tenant.name, tenant.schema)

			// Set search path for this tenant
			_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", tenant.schema))
			if err != nil {
				fmt.Printf("    ❌ Failed to set search path: %v\n", err)
				continue
			}

			// Query users
			rows, err := conn.Query(ctx, "SELECT name, email FROM users ORDER BY name")
			if err != nil {
				fmt.Printf("    ❌ Failed to query users: %v\n", err)
				continue
			}

			fmt.Printf("    Users:\n")
			for rows.Next() {
				var name, email string
				if err := rows.Scan(&name, &email); err != nil {
					fmt.Printf("      ❌ Failed to scan user: %v\n", err)
					continue
				}
				fmt.Printf("      - %s (%s)\n", name, email)
			}
			rows.Close()

			// Query orders with user names
			rows, err = conn.Query(ctx, `
				SELECT u.name, o.total, o.status 
				FROM orders o 
				JOIN users u ON u.id = o.user_id 
				ORDER BY o.created_at
			`)
			if err != nil {
				fmt.Printf("    ❌ Failed to query orders: %v\n", err)
				continue
			}

			fmt.Printf("    Orders:\n")
			for rows.Next() {
				var userName, status string
				var total float64
				if err := rows.Scan(&userName, &total, &status); err != nil {
					fmt.Printf("      ❌ Failed to scan order: %v\n", err)
					continue
				}
				fmt.Printf("      - %s: $%.2f (%s)\n", userName, total, status)
			}
			rows.Close()
		}

		// Reset search path
		_, err = conn.Exec(ctx, "SET search_path TO public")
		if err != nil {
			fmt.Printf("⚠️  Failed to reset search path: %v\n", err)
		}

		return nil
	})
}

func demonstrateDatabaseBasedMultiTenancy(ctx context.Context, provider interfaces.PostgreSQLProvider) error {
	fmt.Println("\n=== Database-based Multi-tenancy Example ===")

	// Create tenant manager
	tenantManager := NewTenantManager(provider)
	defer tenantManager.CloseAll()

	// Register tenants with separate databases
	tenants := []*TenantConfig{
		{
			TenantID:    "acme_corp",
			SchemaName:  "public",
			DatabaseURL: "postgres://user:password@localhost:5432/acme_db",
			Settings:    map[string]interface{}{"region": "us-east-1", "tier": "premium"},
		},
		{
			TenantID:    "globex_corp",
			SchemaName:  "public",
			DatabaseURL: "postgres://user:password@localhost:5432/globex_db",
			Settings:    map[string]interface{}{"region": "us-west-2", "tier": "standard"},
		},
		{
			TenantID:    "initech",
			SchemaName:  "public",
			DatabaseURL: "postgres://user:password@localhost:5432/initech_db",
			Settings:    map[string]interface{}{"region": "eu-west-1", "tier": "basic"},
		},
	}

	// Register all tenants
	fmt.Println("🏢 Registering tenants...")
	for _, tenant := range tenants {
		err := tenantManager.RegisterTenant(tenant)
		if err != nil {
			fmt.Printf("  ❌ Failed to register tenant %s: %v\n", tenant.TenantID, err)
			continue
		}
		fmt.Printf("  ✅ Registered tenant: %s (DB: %s, Region: %s, Tier: %s)\n",
			tenant.TenantID,
			extractDBName(tenant.DatabaseURL),
			tenant.Settings["region"],
			tenant.Settings["tier"])
	}

	// Demonstrate operations on different tenant databases
	fmt.Println("\n📊 Performing operations on tenant databases...")

	for _, tenant := range tenants {
		fmt.Printf("\n  🏢 Working with tenant: %s\n", tenant.TenantID)

		pool, err := tenantManager.GetPoolForTenant(ctx, tenant.TenantID)
		if err != nil {
			fmt.Printf("    ❌ Failed to get pool for tenant %s: %v\n", tenant.TenantID, err)
			continue
		}

		err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			// Create tenant-specific tables
			_, err := conn.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS tenant_info (
					id SERIAL PRIMARY KEY,
					tenant_id TEXT NOT NULL,
					region TEXT NOT NULL,
					tier TEXT NOT NULL,
					created_at TIMESTAMP DEFAULT NOW()
				)
			`)
			if err != nil {
				return fmt.Errorf("failed to create tenant_info table: %w", err)
			}

			// Insert tenant information
			_, err = conn.Exec(ctx,
				"INSERT INTO tenant_info (tenant_id, region, tier) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
				tenant.TenantID, tenant.Settings["region"], tenant.Settings["tier"])
			if err != nil {
				return fmt.Errorf("failed to insert tenant info: %w", err)
			}

			// Query tenant information
			var dbTenantID, region, tier string
			var createdAt time.Time
			row := conn.QueryRow(ctx, "SELECT tenant_id, region, tier, created_at FROM tenant_info WHERE tenant_id = $1", tenant.TenantID)
			err = row.Scan(&dbTenantID, &region, &tier, &createdAt)
			if err != nil {
				return fmt.Errorf("failed to query tenant info: %w", err)
			}

			fmt.Printf("    ✅ Tenant Info: ID=%s, Region=%s, Tier=%s, Created=%s\n",
				dbTenantID, region, tier, createdAt.Format("2006-01-02 15:04:05"))

			return nil
		})

		if err != nil {
			fmt.Printf("    ❌ Operations failed for tenant %s: %v\n", tenant.TenantID, err)
		} else {
			fmt.Printf("    ✅ Operations completed for tenant %s\n", tenant.TenantID)
		}
	}

	return nil
}

func demonstrateTenantIsolation(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Tenant Isolation and Security Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Tenant isolation example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Demonstrate Row Level Security (RLS)
		fmt.Println("🔒 Setting up Row Level Security...")

		// Create a shared table with tenant_id column
		_, err := conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS shared_documents (
				id SERIAL PRIMARY KEY,
				tenant_id TEXT NOT NULL,
				title TEXT NOT NULL,
				content TEXT,
				created_at TIMESTAMP DEFAULT NOW()
			)
		`)
		if err != nil {
			fmt.Printf("  ❌ Failed to create shared_documents table: %v\n", err)
			return err
		}

		// Enable RLS on the table
		_, err = conn.Exec(ctx, "ALTER TABLE shared_documents ENABLE ROW LEVEL SECURITY")
		if err != nil {
			fmt.Printf("  ⚠️  Failed to enable RLS (may already be enabled): %v\n", err)
		} else {
			fmt.Printf("  ✅ Enabled Row Level Security on shared_documents\n")
		}

		// Create RLS policy (using simplified approach for demo)
		_, err = conn.Exec(ctx, `
			CREATE POLICY tenant_isolation ON shared_documents
			FOR ALL TO PUBLIC
			USING (tenant_id = current_setting('app.current_tenant', true))
		`)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to create RLS policy (may already exist): %v\n", err)
		} else {
			fmt.Printf("  ✅ Created tenant isolation policy\n")
		}

		// Insert sample data for different tenants
		fmt.Println("\n📝 Inserting sample documents for different tenants...")

		documents := []struct {
			tenantID string
			title    string
			content  string
		}{
			{"acme", "ACME Company Policy", "This is ACME's internal policy document."},
			{"acme", "ACME Product Specs", "Technical specifications for ACME products."},
			{"globex", "Globex Strategy 2024", "Strategic planning document for Globex."},
			{"globex", "Globex Financial Report", "Q4 financial report for Globex Corporation."},
			{"initech", "Initech TPS Reports", "TPS report templates and guidelines."},
		}

		for _, doc := range documents {
			_, err := conn.Exec(ctx,
				"INSERT INTO shared_documents (tenant_id, title, content) VALUES ($1, $2, $3)",
				doc.tenantID, doc.title, doc.content)
			if err != nil {
				fmt.Printf("  ❌ Failed to insert document for %s: %v\n", doc.tenantID, err)
				continue
			}
			fmt.Printf("  ✅ Inserted document for %s: %s\n", doc.tenantID, doc.title)
		}

		// Demonstrate tenant isolation by querying with different tenant contexts
		fmt.Println("\n🔍 Demonstrating tenant isolation...")

		tenantIDs := []string{"acme", "globex", "initech"}

		for _, tenantID := range tenantIDs {
			fmt.Printf("\n  🏢 Querying as tenant: %s\n", tenantID)

			// Set current tenant (this would normally be done through application context)
			_, err := conn.Exec(ctx, fmt.Sprintf("SET app.current_tenant = '%s'", tenantID))
			if err != nil {
				fmt.Printf("    ❌ Failed to set current tenant: %v\n", err)
				continue
			}

			// Query documents (should only return documents for current tenant)
			rows, err := conn.Query(ctx, "SELECT title, content FROM shared_documents ORDER BY created_at")
			if err != nil {
				fmt.Printf("    ❌ Failed to query documents: %v\n", err)
				continue
			}

			fmt.Printf("    Documents accessible to %s:\n", tenantID)
			documentCount := 0
			for rows.Next() {
				var title, content string
				if err := rows.Scan(&title, &content); err != nil {
					fmt.Printf("      ❌ Failed to scan document: %v\n", err)
					continue
				}
				documentCount++
				fmt.Printf("      - %s\n", title)
			}
			rows.Close()

			if documentCount == 0 {
				fmt.Printf("    ⚠️  No documents found (RLS may not be fully configured)\n")
			} else {
				fmt.Printf("    ✅ Found %d documents for tenant %s\n", documentCount, tenantID)
			}
		}

		// Reset tenant setting
		_, err = conn.Exec(ctx, "RESET app.current_tenant")
		if err != nil {
			fmt.Printf("  ⚠️  Failed to reset tenant setting: %v\n", err)
		}

		return nil
	})
}

func demonstrateTenantManagement(ctx context.Context, provider interfaces.PostgreSQLProvider) error {
	fmt.Println("\n=== Tenant Management Operations Example ===")

	// Create tenant manager
	tenantManager := NewTenantManager(provider)
	defer tenantManager.CloseAll()

	// Demonstrate tenant lifecycle management
	fmt.Println("🔄 Demonstrating tenant lifecycle management...")

	// 1. Provision new tenant
	newTenant := &TenantConfig{
		TenantID:    "newcorp",
		SchemaName:  "newcorp_schema",
		DatabaseURL: "postgres://user:password@localhost:5432/testdb",
		Settings: map[string]interface{}{
			"region":        "us-central-1",
			"tier":          "trial",
			"max_users":     10,
			"storage_limit": "1GB",
		},
	}

	fmt.Printf("  📋 Provisioning new tenant: %s\n", newTenant.TenantID)
	err := tenantManager.RegisterTenant(newTenant)
	if err != nil {
		fmt.Printf("    ❌ Failed to provision tenant: %v\n", err)
	} else {
		fmt.Printf("    ✅ Successfully provisioned tenant: %s\n", newTenant.TenantID)
	}

	// 2. Initialize tenant database
	pool, err := tenantManager.GetPoolForTenant(ctx, newTenant.TenantID)
	if err != nil {
		fmt.Printf("    ❌ Failed to get pool for new tenant: %v\n", err)
		return err
	}

	err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		fmt.Printf("  🏗️  Initializing database schema for tenant: %s\n", newTenant.TenantID)

		// Create schema
		_, err := conn.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", newTenant.SchemaName))
		if err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}

		// Set search path
		_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", newTenant.SchemaName))
		if err != nil {
			return fmt.Errorf("failed to set search path: %w", err)
		}

		// Create tenant configuration table
		_, err = conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS tenant_config (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				updated_at TIMESTAMP DEFAULT NOW()
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create tenant_config table: %w", err)
		}

		// Insert tenant settings
		for key, value := range newTenant.Settings {
			_, err = conn.Exec(ctx,
				"INSERT INTO tenant_config (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()",
				key, fmt.Sprintf("%v", value))
			if err != nil {
				return fmt.Errorf("failed to insert config %s: %w", key, err)
			}
		}

		fmt.Printf("    ✅ Database schema initialized for tenant: %s\n", newTenant.TenantID)
		return nil
	})

	if err != nil {
		fmt.Printf("    ❌ Failed to initialize tenant database: %v\n", err)
		return err
	}

	// 3. Tenant configuration management
	fmt.Printf("  ⚙️  Managing tenant configuration...\n")

	err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Set search path
		_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", newTenant.SchemaName))
		if err != nil {
			return err
		}

		// Update tenant settings
		_, err = conn.Exec(ctx,
			"UPDATE tenant_config SET value = $1, updated_at = NOW() WHERE key = $2",
			"standard", "tier")
		if err != nil {
			return fmt.Errorf("failed to update tier: %w", err)
		}

		_, err = conn.Exec(ctx,
			"INSERT INTO tenant_config (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()",
			"upgrade_date", time.Now().Format("2006-01-02"))
		if err != nil {
			return fmt.Errorf("failed to insert upgrade_date: %w", err)
		}

		// Query current configuration
		rows, err := conn.Query(ctx, "SELECT key, value, updated_at FROM tenant_config ORDER BY key")
		if err != nil {
			return fmt.Errorf("failed to query config: %w", err)
		}
		defer rows.Close()

		fmt.Printf("    Current configuration for %s:\n", newTenant.TenantID)
		for rows.Next() {
			var key, value string
			var updatedAt time.Time
			if err := rows.Scan(&key, &value, &updatedAt); err != nil {
				return fmt.Errorf("failed to scan config: %w", err)
			}
			fmt.Printf("      %s: %s (updated: %s)\n", key, value, updatedAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	})

	if err != nil {
		fmt.Printf("    ❌ Failed to manage tenant configuration: %v\n", err)
	} else {
		fmt.Printf("    ✅ Tenant configuration managed successfully\n")
	}

	// 4. Tenant metrics and monitoring
	fmt.Printf("  📊 Collecting tenant metrics...\n")

	err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Set search path
		_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", newTenant.SchemaName))
		if err != nil {
			return err
		}

		// Get schema size (simplified)
		var schemaName string
		var tableCount int
		row := conn.QueryRow(ctx, `
			SELECT schemaname, COUNT(*) as table_count
			FROM pg_tables 
			WHERE schemaname = $1
			GROUP BY schemaname
		`, newTenant.SchemaName)

		err = row.Scan(&schemaName, &tableCount)
		if err != nil {
			// Schema may not have tables yet
			tableCount = 0
		}

		fmt.Printf("    📈 Metrics for tenant %s:\n", newTenant.TenantID)
		fmt.Printf("      - Schema: %s\n", newTenant.SchemaName)
		fmt.Printf("      - Tables: %d\n", tableCount)
		fmt.Printf("      - Region: %s\n", newTenant.Settings["region"])
		fmt.Printf("      - Tier: %s\n", newTenant.Settings["tier"])

		return nil
	})

	if err != nil {
		fmt.Printf("    ❌ Failed to collect metrics: %v\n", err)
	} else {
		fmt.Printf("    ✅ Metrics collected successfully\n")
	}

	return nil
}

func demonstrateCrossTenantOperations(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Cross-tenant Operations Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Cross-tenant operations example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Demonstrate cross-tenant reporting and analytics
		fmt.Println("📊 Demonstrating cross-tenant reporting...")

		// Create a global reporting view that aggregates data across tenants
		_, err := conn.Exec(ctx, `
			CREATE OR REPLACE VIEW global_tenant_summary AS
			SELECT 
				'acme' as tenant_id,
				COUNT(*) as user_count,
				SUM(CASE WHEN status = 'completed' THEN total ELSE 0 END) as completed_revenue
			FROM company_acme.users u
			LEFT JOIN company_acme.orders o ON u.id = o.user_id
			
			UNION ALL
			
			SELECT 
				'globex' as tenant_id,
				COUNT(*) as user_count,
				SUM(CASE WHEN status = 'completed' THEN total ELSE 0 END) as completed_revenue
			FROM company_globex.users u
			LEFT JOIN company_globex.orders o ON u.id = o.user_id
			
			UNION ALL
			
			SELECT 
				'initech' as tenant_id,
				COUNT(*) as user_count,
				SUM(CASE WHEN status = 'completed' THEN total ELSE 0 END) as completed_revenue
			FROM company_initech.users u
			LEFT JOIN company_initech.orders o ON u.id = o.user_id
		`)
		if err != nil {
			fmt.Printf("  ❌ Failed to create global summary view: %v\n", err)
			fmt.Println("  ℹ️  This is expected if tenant schemas don't exist")
		} else {
			fmt.Printf("  ✅ Created global tenant summary view\n")

			// Query the global summary
			rows, err := conn.Query(ctx, "SELECT tenant_id, user_count, COALESCE(completed_revenue, 0) FROM global_tenant_summary ORDER BY tenant_id")
			if err != nil {
				fmt.Printf("  ❌ Failed to query global summary: %v\n", err)
			} else {
				fmt.Printf("  📈 Global Tenant Summary:\n")
				var totalUsers int
				var totalRevenue float64

				for rows.Next() {
					var tenantID string
					var userCount int
					var revenue float64
					if err := rows.Scan(&tenantID, &userCount, &revenue); err != nil {
						fmt.Printf("    ❌ Failed to scan summary: %v\n", err)
						continue
					}
					fmt.Printf("    %s: %d users, $%.2f revenue\n", tenantID, userCount, revenue)
					totalUsers += userCount
					totalRevenue += revenue
				}
				rows.Close()

				fmt.Printf("  📊 Totals: %d users across all tenants, $%.2f total revenue\n", totalUsers, totalRevenue)
			}
		}

		// Demonstrate tenant migration/backup operations
		fmt.Println("\n💾 Demonstrating tenant backup operations...")

		// Create a backup table structure
		_, err = conn.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS tenant_backups (
				id SERIAL PRIMARY KEY,
				tenant_id TEXT NOT NULL,
				backup_type TEXT NOT NULL,
				backup_path TEXT,
				created_at TIMESTAMP DEFAULT NOW(),
				status TEXT DEFAULT 'pending'
			)
		`)
		if err != nil {
			fmt.Printf("  ❌ Failed to create backup table: %v\n", err)
		} else {
			fmt.Printf("  ✅ Created tenant backup tracking table\n")
		}

		// Simulate backup operations for each tenant
		tenantIDs := []string{"acme", "globex", "initech"}
		for _, tenantID := range tenantIDs {
			backupPath := fmt.Sprintf("/backups/%s_%s.sql", tenantID, time.Now().Format("20060102_150405"))

			_, err = conn.Exec(ctx,
				"INSERT INTO tenant_backups (tenant_id, backup_type, backup_path, status) VALUES ($1, $2, $3, $4)",
				tenantID, "full", backupPath, "completed")
			if err != nil {
				fmt.Printf("    ❌ Failed to record backup for %s: %v\n", tenantID, err)
				continue
			}

			fmt.Printf("    ✅ Backup recorded for tenant %s: %s\n", tenantID, backupPath)
		}

		// Query backup status
		rows, err := conn.Query(ctx, `
			SELECT tenant_id, backup_type, backup_path, status, created_at 
			FROM tenant_backups 
			ORDER BY created_at DESC 
			LIMIT 10
		`)
		if err != nil {
			fmt.Printf("  ❌ Failed to query backup status: %v\n", err)
		} else {
			fmt.Printf("  📋 Recent Backup Operations:\n")
			for rows.Next() {
				var tenantID, backupType, backupPath, status string
				var createdAt time.Time
				if err := rows.Scan(&tenantID, &backupType, &backupPath, &status, &createdAt); err != nil {
					fmt.Printf("    ❌ Failed to scan backup: %v\n", err)
					continue
				}
				fmt.Printf("    %s [%s]: %s backup to %s (%s)\n",
					createdAt.Format("2006-01-02 15:04:05"), tenantID, backupType, backupPath, status)
			}
			rows.Close()
		}

		return nil
	})
}

// Helper function to extract database name from URL
func extractDBName(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}
