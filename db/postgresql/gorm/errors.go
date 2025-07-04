package gorm

import (
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
)

// WrapError é uma função auxiliar para envolver erros com mensagens adicionais
// (implementação específica para o provider GORM)
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Erros específicos do GORM
var (
	// ErrNoRows é um erro que indica que nenhum registro foi encontrado
	ErrNoRows = common.ErrNoRows

	// ErrNotGormConnection é retornado quando a conexão não é uma conexão GORM
	ErrNotGormConnection = errors.New("a conexão fornecida não é uma conexão GORM")

	// ErrNotGormTransaction é retornado quando a transação não é uma transação GORM
	ErrNotGormTransaction = errors.New("a transação fornecida não é uma transação GORM")
)
