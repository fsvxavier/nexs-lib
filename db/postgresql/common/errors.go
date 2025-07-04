package common

import (
	"errors"
	"strings"
)

const (
	// PostgreSQLUniqueViolationCode representa o código de erro para violação de restrição unique
	PostgreSQLUniqueViolationCode = "23505"
)

var (
	// ErrNoRows é retornado quando uma consulta não retorna linhas
	ErrNoRows = errors.New("no rows in result set")

	// ErrNoTransaction é retornado quando uma operação de transação é tentada sem uma transação ativa
	ErrNoTransaction = errors.New("no transaction taking place")

	// ErrNoConnection é retornado quando uma operação é tentada em uma conexão já fechada
	ErrNoConnection = errors.New("connection already closed")

	// ErrInvalidNestedTransaction é retornado quando uma transação é iniciada dentro de outra transação
	ErrInvalidNestedTransaction = errors.New("transaction already in progress")

	// ErrInvalidOperation é retornado quando uma operação inválida é tentada
	ErrInvalidOperation = errors.New("invalid operation")
)

// IsEmptyResultError verifica se um erro indica ausência de resultados
func IsEmptyResultError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoRows) ||
		strings.Contains(err.Error(), "no rows in result set") ||
		strings.Contains(err.Error(), "no row was found")
}

// IsDuplicateKeyError verifica se um erro indica violação de chave única
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), PostgreSQLUniqueViolationCode)
}

// PostgreSQLError encapsula erros específicos do PostgreSQL
type PostgreSQLError struct {
	Message string
	Code    string
}

// NewPostgreSQLError cria um novo erro de PostgreSQL
func NewPostgreSQLError(message string, code string) *PostgreSQLError {
	return &PostgreSQLError{
		Message: message,
		Code:    code,
	}
}

// Error implementa a interface error
func (pe PostgreSQLError) Error() string {
	return pe.Message
}

// IsUniqueViolation verifica se o erro é uma violação de chave única
func (pe PostgreSQLError) IsUniqueViolation() bool {
	return pe.Code == PostgreSQLUniqueViolationCode
}
