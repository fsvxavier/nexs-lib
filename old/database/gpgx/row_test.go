package gpgx_test

import (
	"errors"
	"testing"

	"github.com/dock-tech/isis-golang-lib/database/gpgx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPgxRow is a mock implementation of pgx.Row
type MockPgxRow struct {
	mock.Mock
}

func (m *MockPgxRow) Scan(dest ...any) error {
	args := m.Called(dest)
	return args.Error(0)
}

func TestNewPgxRow(t *testing.T) {
	// Arrange
	mockRow := new(MockPgxRow)

	// Act
	row := gpgx.NewPgxRow(mockRow)

	// Assert
	assert.NotNil(t, row, "NewPgxRow should return a non-nil IRow")
}

func TestPgxRow_Scan_Success(t *testing.T) {
	// Arrange
	mockRow := new(MockPgxRow)
	mockRow.On("Scan", mock.Anything).Return(nil)

	row := gpgx.NewPgxRow(mockRow)

	var result string

	// Act
	err := row.Scan(&result)

	// Assert
	assert.NoError(t, err, "Scan should not return an error when successful")
	mockRow.AssertExpectations(t)
}

func TestPgxRow_Scan_Error(t *testing.T) {
	// Arrange
	expectedErr := errors.New("scan error")
	mockRow := new(MockPgxRow)
	mockRow.On("Scan", mock.Anything).Return(expectedErr)

	row := gpgx.NewPgxRow(mockRow)

	var result string

	// Act
	err := row.Scan(&result)

	// Assert
	assert.Error(t, err, "Scan should return an error when the underlying row returns an error")
	assert.Equal(t, expectedErr, err, "The error should be passed through")
	mockRow.AssertExpectations(t)
}
