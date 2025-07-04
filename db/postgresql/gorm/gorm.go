package gorm

import (
	"context"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"gorm.io/gorm"
)

// ModelWithDB combina um model GORM com uma instância GORM DB
func ModelWithDB(conn common.IConn, model interface{}) (*gorm.DB, error) {
	// Verifica se a conexão é uma conexão GORM
	gormConn, ok := conn.(*Conn)
	if !ok {
		gormTx, ok := conn.(*Transaction)
		if !ok {
			return nil, ErrNotGormConnection
		}
		return gormTx.db.Model(model), nil
	}

	return gormConn.db.Model(model), nil
}

// GormTx extrai o objeto de transação GORM de uma transação comum
func GormTx(tx common.ITransaction) (*gorm.DB, error) {
	gormTx, ok := tx.(*Transaction)
	if !ok {
		return nil, ErrNotGormTransaction
	}
	return gormTx.db, nil
}

// GormDB extrai o objeto DB GORM de uma conexão comum
func GormDB(conn common.IConn) (*gorm.DB, error) {
	// Verifica se a conexão é uma conexão GORM
	gormConn, ok := conn.(*Conn)
	if !ok {
		gormTx, ok := conn.(*Transaction)
		if !ok {
			return nil, ErrNotGormConnection
		}
		return gormTx.db, nil
	}

	return gormConn.db, nil
}

// AutoMigrate executa a migração automática do GORM para os modelos fornecidos
func AutoMigrate(ctx context.Context, conn common.IConn, models ...interface{}) error {
	db, err := GormDB(conn)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).AutoMigrate(models...)
}
