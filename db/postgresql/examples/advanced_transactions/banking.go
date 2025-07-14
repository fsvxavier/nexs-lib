package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
)

// Account representa uma conta bancária
type Account struct {
	ID      int     `json:"id"`
	Number  string  `json:"number"`
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
	UserID  int     `json:"user_id"`
}

// Transaction representa uma transação bancária
type BankTransaction struct {
	ID          int       `json:"id"`
	FromAccount int       `json:"from_account"`
	ToAccount   int       `json:"to_account"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// BankingService simula um serviço bancário com transações complexas
type BankingService struct {
	pool postgresql.IPool
}

// NewBankingService cria um novo serviço bancário
func NewBankingService(pool postgresql.IPool) *BankingService {
	return &BankingService{pool: pool}
}

// CreateBankingTables cria as tabelas necessárias para o sistema bancário
func (bs *BankingService) CreateBankingTables(ctx context.Context) error {
	conn, err := bs.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	queries := []string{
		`DROP TABLE IF EXISTS bank_transactions CASCADE`,
		`DROP TABLE IF EXISTS accounts CASCADE`,
		`
		CREATE TABLE accounts (
			id SERIAL PRIMARY KEY,
			number VARCHAR(20) UNIQUE NOT NULL,
			balance DECIMAL(12,2) NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
			status VARCHAR(10) DEFAULT 'active' CHECK (status IN ('active', 'blocked', 'closed')),
			user_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`
		CREATE TABLE bank_transactions (
			id SERIAL PRIMARY KEY,
			from_account INTEGER REFERENCES accounts(id),
			to_account INTEGER REFERENCES accounts(id),
			amount DECIMAL(12,2) NOT NULL CHECK (amount > 0),
			type VARCHAR(20) NOT NULL CHECK (type IN ('transfer', 'deposit', 'withdrawal')),
			status VARCHAR(15) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`
		CREATE INDEX idx_accounts_number ON accounts(number);
		CREATE INDEX idx_accounts_user_id ON accounts(user_id);
		CREATE INDEX idx_transactions_from_account ON bank_transactions(from_account);
		CREATE INDEX idx_transactions_to_account ON bank_transactions(to_account);
		CREATE INDEX idx_transactions_status ON bank_transactions(status);
		`,
	}

	for _, query := range queries {
		if err := conn.Exec(ctx, query); err != nil {
			return fmt.Errorf("erro ao executar query: %w", err)
		}
	}

	fmt.Println("✅ Tabelas bancárias criadas com sucesso!")
	return nil
}

// CreateInitialAccounts cria contas iniciais para teste
func (bs *BankingService) CreateInitialAccounts(ctx context.Context) error {
	conn, err := bs.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	accounts := []Account{
		{Number: "ACC-001", Balance: 10000.00, Status: "active", UserID: 1},
		{Number: "ACC-002", Balance: 5000.00, Status: "active", UserID: 2},
		{Number: "ACC-003", Balance: 15000.00, Status: "active", UserID: 3},
		{Number: "ACC-004", Balance: 2500.00, Status: "active", UserID: 4},
		{Number: "ACC-005", Balance: 8000.00, Status: "active", UserID: 5},
	}

	for _, account := range accounts {
		if err := conn.Exec(ctx,
			"INSERT INTO accounts (number, balance, status, user_id) VALUES ($1, $2, $3, $4)",
			account.Number, account.Balance, account.Status, account.UserID); err != nil {
			return fmt.Errorf("erro ao inserir conta %s: %w", account.Number, err)
		}
		fmt.Printf("   ✅ Conta criada: %s (Saldo: R$ %.2f)\n", account.Number, account.Balance)
	}

	return nil
}

// Transfer executa uma transferência bancária com transação robusta
func (bs *BankingService) Transfer(ctx context.Context, fromAccountID, toAccountID int, amount float64, description string) error {
	// Usar timeout específico para a transação
	txCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conn, err := bs.pool.Acquire(txCtx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(txCtx)

	// Iniciar transação com nível de isolamento serializable
	tx, err := conn.BeginTransactionWithOptions(txCtx, postgresql.TxOptions{
		IsoLevel: postgresql.IsoLevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}

	// Função para rollback automático em caso de erro
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(txCtx)
			panic(p)
		}
	}()

	fmt.Printf("🔄 Iniciando transferência: Conta %d → Conta %d (R$ %.2f)\n",
		fromAccountID, toAccountID, amount)

	// Savepoint antes de verificar saldos
	if err := tx.Savepoint(txCtx, "before_balance_check"); err != nil {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("erro ao criar savepoint: %w", err)
	}

	// Verificar se conta de origem existe e tem saldo suficiente (com lock)
	var fromBalance float64
	var fromStatus string
	row, _ := tx.QueryRow(txCtx,
		"SELECT balance, status FROM accounts WHERE id = $1 FOR UPDATE",
		fromAccountID)

	if err := row.Scan(&fromBalance, &fromStatus); err != nil {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("conta de origem não encontrada: %w", err)
	}

	if fromStatus != "active" {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("conta de origem não está ativa: %s", fromStatus)
	}

	if fromBalance < amount {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("saldo insuficiente: R$ %.2f disponível, R$ %.2f necessário",
			fromBalance, amount)
	}

	// Verificar se conta de destino existe e está ativa (com lock)
	var toStatus string
	toRow, _ := tx.QueryRow(txCtx,
		"SELECT status FROM accounts WHERE id = $1 FOR UPDATE",
		toAccountID)

	if err := toRow.Scan(&toStatus); err != nil {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("conta de destino não encontrada: %w", err)
	}

	if toStatus != "active" {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("conta de destino não está ativa: %s", toStatus)
	}

	// Savepoint antes das atualizações
	if err := tx.Savepoint(txCtx, "before_updates"); err != nil {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("erro ao criar savepoint de atualização: %w", err)
	}

	// Registrar transação bancária
	var transactionID int
	txRow, _ := tx.QueryRow(txCtx,
		`INSERT INTO bank_transactions (from_account, to_account, amount, type, description, status) 
		 VALUES ($1, $2, $3, 'transfer', $4, 'pending') RETURNING id`,
		fromAccountID, toAccountID, amount, description)

	if err := txRow.Scan(&transactionID); err != nil {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("erro ao registrar transação: %w", err)
	}

	fmt.Printf("   📝 Transação registrada: ID %d\n", transactionID)

	// Simular processamento complexo (pode falhar)
	if rand.Float64() < 0.1 { // 10% de chance de falha simulada
		fmt.Printf("   ❌ Falha simulada durante processamento\n")

		// Rollback para savepoint antes das atualizações
		if err := tx.RollbackToSavepoint(txCtx, "before_updates"); err != nil {
			_ = tx.Rollback(txCtx)
			return fmt.Errorf("erro ao fazer rollback para savepoint: %w", err)
		}

		// Marcar transação como falhada
		if err := tx.Exec(txCtx,
			"UPDATE bank_transactions SET status = 'failed' WHERE id = $1",
			transactionID); err != nil {
			_ = tx.Rollback(txCtx)
			return fmt.Errorf("erro ao marcar transação como falhada: %w", err)
		}

		if err := tx.Commit(txCtx); err != nil {
			return fmt.Errorf("erro ao confirmar rollback: %w", err)
		}

		return fmt.Errorf("falha simulada no processamento da transação")
	}

	// Debitar da conta de origem
	if err := tx.Exec(txCtx,
		"UPDATE accounts SET balance = balance - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		amount, fromAccountID); err != nil {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("erro ao debitar conta de origem: %w", err)
	}

	fmt.Printf("   💸 Debitado R$ %.2f da conta %d\n", amount, fromAccountID)

	// Creditar na conta de destino
	if err := tx.Exec(txCtx,
		"UPDATE accounts SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		amount, toAccountID); err != nil {

		// Tentar rollback para savepoint antes das atualizações
		if rollbackErr := tx.RollbackToSavepoint(txCtx, "before_updates"); rollbackErr != nil {
			_ = tx.Rollback(txCtx)
			return fmt.Errorf("erro ao creditar e rollback falhou: %w (original: %w)", rollbackErr, err)
		}

		// Marcar transação como falhada
		_ = tx.Exec(txCtx,
			"UPDATE bank_transactions SET status = 'failed' WHERE id = $1",
			transactionID)

		_ = tx.Commit(txCtx)
		return fmt.Errorf("erro ao creditar conta de destino: %w", err)
	}

	fmt.Printf("   💰 Creditado R$ %.2f na conta %d\n", amount, toAccountID)

	// Marcar transação como completada
	if err := tx.Exec(txCtx,
		"UPDATE bank_transactions SET status = 'completed' WHERE id = $1",
		transactionID); err != nil {
		_ = tx.Rollback(txCtx)
		return fmt.Errorf("erro ao marcar transação como completada: %w", err)
	}

	// Confirmar transação
	if err := tx.Commit(txCtx); err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	fmt.Printf("   ✅ Transferência completada com sucesso! (ID: %d)\n", transactionID)
	return nil
}

// GetAccountBalance obtém o saldo atual de uma conta
func (bs *BankingService) GetAccountBalance(ctx context.Context, accountID int) (float64, error) {
	conn, err := bs.pool.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	var balance float64
	row, _ := conn.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1", accountID)
	if err := row.Scan(&balance); err != nil {
		return 0, fmt.Errorf("erro ao consultar saldo: %w", err)
	}

	return balance, nil
}

// GetTransactionHistory obtém histórico de transações
func (bs *BankingService) GetTransactionHistory(ctx context.Context, accountID int, limit int) ([]BankTransaction, error) {
	conn, err := bs.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	rows, err := conn.Query(ctx,
		`SELECT id, from_account, to_account, amount, type, status, description, created_at 
		 FROM bank_transactions 
		 WHERE from_account = $1 OR to_account = $1 
		 ORDER BY created_at DESC LIMIT $2`,
		accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar histórico: %w", err)
	}
	defer rows.Close()

	var transactions []BankTransaction
	for rows.Next() {
		var tx BankTransaction
		if err := rows.Scan(&tx.ID, &tx.FromAccount, &tx.ToAccount, &tx.Amount,
			&tx.Type, &tx.Status, &tx.Description, &tx.CreatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear transação: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// PrintAccountSummary imprime resumo das contas
func (bs *BankingService) PrintAccountSummary(ctx context.Context) error {
	conn, err := bs.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("erro ao adquirir conexão: %w", err)
	}
	defer conn.Release(ctx)

	rows, err := conn.Query(ctx,
		"SELECT id, number, balance, status FROM accounts ORDER BY id")
	if err != nil {
		return fmt.Errorf("erro ao consultar contas: %w", err)
	}
	defer rows.Close()

	fmt.Println("\n💳 Resumo das Contas:")
	fmt.Println("┌────────┬─────────┬────────────┬─────────┐")
	fmt.Println("│   ID   │ Número  │   Saldo    │ Status  │")
	fmt.Println("├────────┼─────────┼────────────┼─────────┤")

	for rows.Next() {
		var account Account
		if err := rows.Scan(&account.ID, &account.Number, &account.Balance, &account.Status); err != nil {
			return fmt.Errorf("erro ao escanear conta: %w", err)
		}
		fmt.Printf("│ %6d │ %7s │ R$ %8.2f │ %7s │\n",
			account.ID, account.Number, account.Balance, account.Status)
	}

	fmt.Println("└────────┴─────────┴────────────┴─────────┘")
	return nil
}
