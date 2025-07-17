package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func main() {
	// Advanced transaction management example
	ctx := context.Background()

	// Create configuration
	cfg := postgresql.NewDefaultConfig("postgres://user:password@localhost:5432/testdb")

	if defaultCfg, ok := cfg.(*config.DefaultConfig); ok {
		err := defaultCfg.ApplyOptions(
			postgresql.WithMaxConns(10),
			postgresql.WithMinConns(2),
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

	// Example 1: Basic transaction
	if err := basicTransactionExample(ctx, provider, cfg); err != nil {
		log.Printf("Basic transaction example failed: %v", err)
	}

	// Example 2: Transaction isolation levels
	if err := isolationLevelsExample(ctx, provider, cfg); err != nil {
		log.Printf("Isolation levels example failed: %v", err)
	}

	// Example 3: Nested transactions (savepoints)
	if err := nestedTransactionsExample(ctx, provider, cfg); err != nil {
		log.Printf("Nested transactions example failed: %v", err)
	}

	// Example 4: Transaction rollback scenarios
	if err := rollbackScenariosExample(ctx, provider, cfg); err != nil {
		log.Printf("Rollback scenarios example failed: %v", err)
	}

	// Example 5: Transaction with timeout
	if err := transactionTimeoutExample(ctx, provider, cfg); err != nil {
		log.Printf("Transaction timeout example failed: %v", err)
	}

	// Example 6: Bulk operations in transaction
	if err := bulkOperationsExample(ctx, provider, cfg); err != nil {
		log.Printf("Bulk operations example failed: %v", err)
	}

	fmt.Println("Transaction examples completed!")
}

func basicTransactionExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("=== Basic Transaction Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Transaction example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Begin transaction with default options
		tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
			IsoLevel:   interfaces.TxIsoLevelReadCommitted,
			AccessMode: interfaces.TxAccessModeReadWrite,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Success flag to determine commit/rollback
		var success bool
		defer func() {
			if success {
				fmt.Println("✅ Committing transaction")
				if err := tx.Commit(ctx); err != nil {
					fmt.Printf("❌ Commit failed: %v\n", err)
				}
			} else {
				fmt.Println("🔄 Rolling back transaction")
				if err := tx.Rollback(ctx); err != nil {
					fmt.Printf("❌ Rollback failed: %v\n", err)
				}
			}
		}()

		// Create tables
		_, err = tx.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS accounts (
				id SERIAL PRIMARY KEY,
				name TEXT NOT NULL,
				balance DECIMAL(10,2) DEFAULT 0
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create accounts table: %w", err)
		}

		// Insert initial data
		_, err = tx.Exec(ctx, "INSERT INTO accounts (name, balance) VALUES ($1, $2), ($3, $4)",
			"Alice", 1000.00, "Bob", 500.00)
		if err != nil {
			return fmt.Errorf("failed to insert accounts: %w", err)
		}

		// Query inserted data
		rows, err := tx.Query(ctx, "SELECT name, balance FROM accounts ORDER BY name")
		if err != nil {
			return fmt.Errorf("failed to query accounts: %w", err)
		}
		defer rows.Close()

		fmt.Println("📊 Account balances:")
		for rows.Next() {
			var name string
			var balance float64
			if err := rows.Scan(&name, &balance); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
			fmt.Printf("  %s: $%.2f\n", name, balance)
		}

		success = true
		return nil
	})
}

func isolationLevelsExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Isolation Levels Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Isolation levels example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	isolationLevels := []struct {
		level interfaces.TxIsoLevel
		name  string
	}{
		{interfaces.TxIsoLevelReadUncommitted, "READ UNCOMMITTED"},
		{interfaces.TxIsoLevelReadCommitted, "READ COMMITTED"},
		{interfaces.TxIsoLevelRepeatableRead, "REPEATABLE READ"},
		{interfaces.TxIsoLevelSerializable, "SERIALIZABLE"},
	}

	for _, iso := range isolationLevels {
		fmt.Printf("🔒 Testing isolation level: %s\n", iso.name)

		err := pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
			tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
				IsoLevel:   iso.level,
				AccessMode: interfaces.TxAccessModeReadOnly,
			})
			if err != nil {
				return fmt.Errorf("failed to begin transaction with %s: %w", iso.name, err)
			}
			defer tx.Rollback(ctx)

			// Test read operation
			var count int
			row := tx.QueryRow(ctx, "SELECT COUNT(*) FROM accounts")
			if err := row.Scan(&count); err != nil {
				fmt.Printf("  ❌ Failed to query with %s: %v\n", iso.name, err)
			} else {
				fmt.Printf("  ✅ Successfully queried with %s: %d records\n", iso.name, count)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("  ❌ %s transaction failed: %v\n", iso.name, err)
		}
	}

	return nil
}

func nestedTransactionsExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Nested Transactions (Savepoints) Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Nested transactions example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		// Main transaction
		tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
			IsoLevel:   interfaces.TxIsoLevelReadCommitted,
			AccessMode: interfaces.TxAccessModeReadWrite,
		})
		if err != nil {
			return fmt.Errorf("failed to begin main transaction: %w", err)
		}

		var success bool
		defer func() {
			if success {
				fmt.Println("✅ Committing main transaction")
				tx.Commit(ctx)
			} else {
				fmt.Println("🔄 Rolling back main transaction")
				tx.Rollback(ctx)
			}
		}()

		// Create savepoint 1
		fmt.Println("💾 Creating savepoint 'sp1'")
		_, err = tx.Exec(ctx, "SAVEPOINT sp1")
		if err != nil {
			return fmt.Errorf("failed to create savepoint sp1: %w", err)
		}

		// Operations within savepoint 1
		_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + 100 WHERE name = 'Alice'")
		if err != nil {
			return fmt.Errorf("failed to update Alice's balance: %w", err)
		}
		fmt.Println("✅ Updated Alice's balance (+$100)")

		// Create savepoint 2
		fmt.Println("💾 Creating savepoint 'sp2'")
		_, err = tx.Exec(ctx, "SAVEPOINT sp2")
		if err != nil {
			return fmt.Errorf("failed to create savepoint sp2: %w", err)
		}

		// Operations within savepoint 2 (will be rolled back)
		_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - 999999 WHERE name = 'Bob'")
		if err != nil {
			return fmt.Errorf("failed to update Bob's balance: %w", err)
		}
		fmt.Println("⚠️  Updated Bob's balance (-$999999) - this will be rolled back")

		// Rollback to savepoint 2
		fmt.Println("🔄 Rolling back to savepoint 'sp2'")
		_, err = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT sp2")
		if err != nil {
			return fmt.Errorf("failed to rollback to savepoint sp2: %w", err)
		}

		// Release savepoint 2
		fmt.Println("🗑️  Releasing savepoint 'sp2'")
		_, err = tx.Exec(ctx, "RELEASE SAVEPOINT sp2")
		if err != nil {
			return fmt.Errorf("failed to release savepoint sp2: %w", err)
		}

		// Operations after rollback
		_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + 50 WHERE name = 'Bob'")
		if err != nil {
			return fmt.Errorf("failed to update Bob's balance correctly: %w", err)
		}
		fmt.Println("✅ Updated Bob's balance (+$50)")

		// Show final balances
		rows, err := tx.Query(ctx, "SELECT name, balance FROM accounts ORDER BY name")
		if err != nil {
			return fmt.Errorf("failed to query final balances: %w", err)
		}
		defer rows.Close()

		fmt.Println("📊 Final balances after savepoint operations:")
		for rows.Next() {
			var name string
			var balance float64
			if err := rows.Scan(&name, &balance); err != nil {
				return fmt.Errorf("failed to scan final balance: %w", err)
			}
			fmt.Printf("  %s: $%.2f\n", name, balance)
		}

		success = true
		return nil
	})
}

func rollbackScenariosExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Rollback Scenarios Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Rollback scenarios example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	// Scenario 1: Automatic rollback on error
	fmt.Println("🔄 Scenario 1: Automatic rollback on constraint violation")

	err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
			IsoLevel:   interfaces.TxIsoLevelReadCommitted,
			AccessMode: interfaces.TxAccessModeReadWrite,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx) // Will always rollback

		// This should cause a constraint violation
		_, err = tx.Exec(ctx, "INSERT INTO accounts (name, balance) VALUES ('', -1000)")
		if err != nil {
			fmt.Printf("  ✅ Expected error occurred: %v\n", err)
			return nil // Return nil to show successful rollback
		}

		fmt.Println("  ⚠️  Constraint violation didn't occur as expected")
		return nil
	})

	if err != nil {
		fmt.Printf("  ❌ Scenario 1 failed: %v\n", err)
	}

	// Scenario 2: Manual rollback decision
	fmt.Println("\n🔄 Scenario 2: Manual rollback based on business logic")

	err = pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
			IsoLevel:   interfaces.TxIsoLevelReadCommitted,
			AccessMode: interfaces.TxAccessModeReadWrite,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Business logic check
		var aliceBalance float64
		row := tx.QueryRow(ctx, "SELECT balance FROM accounts WHERE name = 'Alice'")
		if err := row.Scan(&aliceBalance); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to get Alice's balance: %w", err)
		}

		// Try to transfer more than available
		transferAmount := aliceBalance + 100

		if transferAmount > aliceBalance {
			fmt.Printf("  ⚠️  Insufficient funds: trying to transfer $%.2f, but only $%.2f available\n",
				transferAmount, aliceBalance)
			fmt.Println("  🔄 Rolling back due to business rule violation")
			tx.Rollback(ctx)
			return nil
		}

		// This won't execute due to the rollback above
		_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE name = 'Alice'", transferAmount)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to transfer: %w", err)
		}

		tx.Commit(ctx)
		return nil
	})

	if err != nil {
		fmt.Printf("  ❌ Scenario 2 failed: %v\n", err)
	}

	return nil
}

func transactionTimeoutExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Transaction Timeout Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Transaction timeout example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	fmt.Println("⏰ Starting transaction with 2-second timeout...")

	err = pool.AcquireFunc(timeoutCtx, func(conn interfaces.IConn) error {
		tx, err := conn.BeginTx(timeoutCtx, interfaces.TxOptions{
			IsoLevel:   interfaces.TxIsoLevelReadCommitted,
			AccessMode: interfaces.TxAccessModeReadWrite,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(timeoutCtx)

		// Simulate long-running operation
		fmt.Println("💤 Simulating long operation (3 seconds)...")
		select {
		case <-time.After(3 * time.Second):
			// This should not execute due to timeout
			fmt.Println("  ⚠️  Long operation completed unexpectedly")
		case <-timeoutCtx.Done():
			fmt.Printf("  ✅ Transaction timed out as expected: %v\n", timeoutCtx.Err())
			return timeoutCtx.Err()
		}

		return nil
	})

	if err != nil {
		fmt.Printf("✅ Transaction timeout handled correctly: %v\n", err)
	}

	return nil
}

func bulkOperationsExample(ctx context.Context, provider interfaces.PostgreSQLProvider, cfg interfaces.Config) error {
	fmt.Println("\n=== Bulk Operations in Transaction Example ===")

	pool, err := provider.NewPool(ctx, cfg)
	if err != nil {
		fmt.Printf("Note: Bulk operations example would require actual database: %v\n", err)
		return nil
	}
	defer pool.Close()

	return pool.AcquireFunc(ctx, func(conn interfaces.IConn) error {
		tx, err := conn.BeginTx(ctx, interfaces.TxOptions{
			IsoLevel:   interfaces.TxIsoLevelReadCommitted,
			AccessMode: interfaces.TxAccessModeReadWrite,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		var success bool
		defer func() {
			if success {
				fmt.Println("✅ Committing bulk transaction")
				tx.Commit(ctx)
			} else {
				fmt.Println("🔄 Rolling back bulk transaction")
				tx.Rollback(ctx)
			}
		}()

		// Create test table for bulk operations
		_, err = tx.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS transactions (
				id SERIAL PRIMARY KEY,
				account_name TEXT NOT NULL,
				amount DECIMAL(10,2) NOT NULL,
				transaction_type TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT NOW()
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create transactions table: %w", err)
		}

		// Prepare statement for bulk inserts
		fmt.Println("📝 Using prepared statements for bulk inserts...")
		preparedSQL := "INSERT INTO transactions (account_name, amount, transaction_type) VALUES ($1, $2, $3)"

		// Bulk insert operations
		fmt.Println("📝 Performing bulk insert operations...")
		transactions := []struct {
			account string
			amount  float64
			txType  string
		}{
			{"Alice", 100.00, "deposit"},
			{"Bob", 50.00, "deposit"},
			{"Alice", -25.00, "withdrawal"},
			{"Bob", -10.00, "withdrawal"},
			{"Alice", 75.00, "deposit"},
		}

		for i, txData := range transactions {
			_, err = tx.Exec(ctx, preparedSQL, txData.account, txData.amount, txData.txType)
			if err != nil {
				return fmt.Errorf("failed to insert transaction %d: %w", i, err)
			}
			fmt.Printf("  ✅ Inserted transaction %d: %s $%.2f (%s)\n",
				i+1, txData.account, txData.amount, txData.txType)
		}

		// Query inserted transactions
		rows, err := tx.Query(ctx, `
			SELECT account_name, SUM(amount) as total
			FROM transactions 
			GROUP BY account_name 
			ORDER BY account_name
		`)
		if err != nil {
			return fmt.Errorf("failed to query transaction summary: %w", err)
		}
		defer rows.Close()

		fmt.Println("📊 Transaction summary:")
		for rows.Next() {
			var account string
			var total float64
			if err := rows.Scan(&account, &total); err != nil {
				return fmt.Errorf("failed to scan summary: %w", err)
			}
			fmt.Printf("  %s: $%.2f\n", account, total)
		}

		success = true
		return nil
	})
}
