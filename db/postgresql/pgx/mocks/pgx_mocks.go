package mocks

import (
	"context"

	"github.com/pashagolub/pgxmock/v4"
)

// GetMock retorna um novo mock para testes de pgx
func GetMock() (pgxmock.PgxConnIface, error) {
	return pgxmock.NewConn()
}

// GetMockWithExpectations retorna um mock com expectativas comuns já configuradas
func GetMockWithExpectations() (pgxmock.PgxConnIface, error) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		return nil, err
	}

	// Configura expectativa para ping
	mock.ExpectPing().WillReturnError(nil)

	return mock, nil
}

// CloseConn fecha a conexão mock
func CloseConn(mock pgxmock.PgxConnIface) {
	mock.Close(context.Background())
}
