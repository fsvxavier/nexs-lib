package pgx

import (
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
)

// NewConnForTest cria uma nova conexão para testes
// Esta função não é destinada a uso em produção, apenas para testes
func NewConnForTest(conn *pgx.Conn) common.IConn {
	return &Conn{
		conn:   conn,
		config: common.DefaultConfig(),
	}
}

// NewMockConnForTest cria uma nova conexão para testes com mock
// Esta função não é destinada a uso em produção, apenas para testes
func NewMockConnForTest(mock pgxmock.PgxConnIface) common.IConn {
	return &Conn{
		mockConn: mock,
		config:   common.DefaultConfig(),
	}
}

// NewPoolConnForTest cria uma nova conexão de pool para testes
// Esta função não é destinada a uso em produção, apenas para testes
func NewPoolConnForTest(poolConn *pgxpool.Conn) common.IConn {
	return &Conn{
		poolConn: poolConn,
		config:   common.DefaultConfig(),
	}
}

// NewTransactionForTest cria uma nova transação para testes
// Esta função não é destinada a uso em produção, apenas para testes
func NewTransactionForTest(tx pgx.Tx) common.ITransaction {
	return &Transaction{
		tx:     tx,
		config: common.DefaultConfig(),
	}
}
