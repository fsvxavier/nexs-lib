package gorm

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"gorm.io/gorm"
)

// Transaction implementa a interface common.ITransaction usando GORM
type Transaction struct {
	db                 *gorm.DB
	multiTenantEnabled bool
	config             *common.Config
}

// Todos os métodos de IConn são implementados pela Transaction

// QueryOne executa uma consulta e escaneia o primeiro resultado no destino
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := t.db.WithContext(ctx).Raw(query, args...).Scan(dst)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar QueryOne na transação")
	}

	if result.RowsAffected == 0 {
		return common.ErrNoRows
	}

	return nil
}

// QueryAll executa uma consulta e escaneia todos os resultados no destino
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := t.db.WithContext(ctx).Raw(query, args...).Scan(dst)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar QueryAll na transação")
	}

	return nil
}

// QueryCount executa uma consulta que retorna uma contagem
func (t *Transaction) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int
	result := t.db.WithContext(ctx).Raw(query, args...).Scan(&count)
	if result.Error != nil {
		return nil, WrapError(result.Error, "falha ao executar QueryCount na transação")
	}

	return &count, nil
}

// Query executa uma consulta e retorna as linhas
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (common.IRows, error) {
	rows, err := t.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, WrapError(err, "falha ao executar Query na transação")
	}

	return &Rows{rows: rows}, nil
}

// QueryRow executa uma consulta e retorna uma única linha
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (common.IRow, error) {
	row := t.db.WithContext(ctx).Raw(query, args...).Row()
	if row.Err() != nil {
		return nil, WrapError(row.Err(), "falha ao executar QueryRow na transação")
	}

	return &Row{row: row}, nil
}

// Exec executa um comando SQL
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	result := t.db.WithContext(ctx).Exec(query, args...)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar Exec na transação")
	}

	return nil
}

// SendBatch envia um lote de consultas para execução
func (t *Transaction) SendBatch(ctx context.Context, batch common.IBatch) (common.IBatchResults, error) {
	// GORM já está em uma transação, então apenas executa as queries
	gormBatch, ok := batch.GetBatch().(*Batch)
	if !ok {
		return nil, errors.New("batch inválido: esperado um batch do tipo GORM")
	}

	results := make([]*gorm.DB, len(gormBatch.queries))
	for i, query := range gormBatch.queries {
		results[i] = t.db.WithContext(ctx).Exec(query.sql, query.args...)
		if results[i].Error != nil {
			return nil, WrapError(results[i].Error, fmt.Sprintf("falha ao executar query %d no batch", i))
		}
	}

	return &BatchResults{
		results: results,
		index:   0,
	}, nil
}

// Ping verifica a conexão com o banco de dados (não aplicável em transações, mas necessário para interface)
func (t *Transaction) Ping(ctx context.Context) error {
	// Em uma transação, não precisamos fazer ping
	return nil
}

// Close não é aplicável em transações, mas necessário para interface
func (t *Transaction) Close(ctx context.Context) error {
	// Em transações, Close não faz sentido (commit ou rollback encerram a transação)
	return nil
}

// BeginTransaction não é suportado dentro de uma transação
func (t *Transaction) BeginTransaction(ctx context.Context) (common.ITransaction, error) {
	return nil, errors.New("não é possível iniciar uma transação dentro de outra transação")
}

// BeginTransactionWithOptions não é suportado dentro de uma transação
func (t *Transaction) BeginTransactionWithOptions(ctx context.Context, opts *common.TxOptions) (common.ITransaction, error) {
	return nil, errors.New("não é possível iniciar uma transação dentro de outra transação")
}

// Commit confirma a transação
func (t *Transaction) Commit(ctx context.Context) error {
	if err := t.db.Commit().Error; err != nil {
		return WrapError(err, "falha ao fazer commit da transação")
	}
	return nil
}

// Rollback reverte a transação
func (t *Transaction) Rollback(ctx context.Context) error {
	if err := t.db.Rollback().Error; err != nil {
		return WrapError(err, "falha ao fazer rollback da transação")
	}
	return nil
}

// Model retorna o objeto DB do GORM para operações com modelos
func (t *Transaction) Model(model interface{}) *gorm.DB {
	return t.db.Model(model)
}

// Create cria um novo registro usando GORM
func (t *Transaction) Create(ctx context.Context, model interface{}) error {
	result := t.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao criar registro na transação")
	}
	return nil
}

// Find busca registros usando GORM
func (t *Transaction) Find(ctx context.Context, dest interface{}, conds ...interface{}) error {
	result := t.db.WithContext(ctx).Find(dest, conds...)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao buscar registros na transação")
	}
	return nil
}

// First busca o primeiro registro usando GORM
func (t *Transaction) First(ctx context.Context, dest interface{}, conds ...interface{}) error {
	result := t.db.WithContext(ctx).First(dest, conds...)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrNoRows
		}
		return WrapError(result.Error, "falha ao buscar primeiro registro na transação")
	}
	return nil
}

// Update atualiza registros usando GORM
func (t *Transaction) Update(ctx context.Context, model interface{}, columns map[string]interface{}) error {
	result := t.db.WithContext(ctx).Model(model).Updates(columns)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao atualizar registro na transação")
	}
	return nil
}

// Delete exclui registros usando GORM
func (t *Transaction) Delete(ctx context.Context, model interface{}, conds ...interface{}) error {
	result := t.db.WithContext(ctx).Delete(model, conds...)
	if result.Error != nil {
		return WrapError(result.Error, "falha ao excluir registro na transação")
	}
	return nil
}
