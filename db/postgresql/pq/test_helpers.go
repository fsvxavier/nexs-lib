package pq

import (
	"database/sql"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
)

// NewConnForTest cria uma nova conexão para testes
// Esta função não é destinada a uso em produção, apenas para testes
func NewConnForTest(conn *sql.Conn) common.IConn {
	return &Conn{
		conn:   conn,
		config: common.DefaultConfig(),
	}
}

// NewTransactionForTest cria uma nova transação para testes
// Esta função não é destinada a uso em produção, apenas para testes
func NewTransactionForTest(tx *sql.Tx) common.ITransaction {
	return &Transaction{
		tx:     tx,
		config: common.DefaultConfig(),
	}
}

// NewPoolForTest cria um novo pool para testes
// Esta função não é destinada a uso em produção, apenas para testes
func NewPoolForTest(db *sql.DB) common.IPool {
	return &Pool{
		db:     db,
		config: common.DefaultConfig(),
	}
}
