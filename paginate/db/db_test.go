package db

import (
	"context"
	"testing"

	page "github.com/fsvxavier/nexs-lib/paginate"
	"github.com/stretchr/testify/assert"
)

// MockRow implementa a interface RowScanner para testes
type MockRow struct {
	value int
	err   error
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}
	*dest[0].(*int) = m.value
	return nil
}

// MockQueryRowExecutor implementa a interface QueryRowExecutor para testes
type MockQueryRowExecutor struct {
	row       *MockRow
	lastQuery string
	lastArgs  []interface{}
}

func (m *MockQueryRowExecutor) QueryRow(ctx context.Context, query string, args ...interface{}) RowScanner {
	m.lastQuery = query
	m.lastArgs = args
	return m.row
}

// MockQueryExecutor implementa a interface QueryExecutor para testes
type MockQueryExecutor struct {
	rows      interface{}
	err       error
	lastQuery string
	lastArgs  []interface{}
}

func (m *MockQueryExecutor) Query(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	m.lastQuery = query
	m.lastArgs = args
	return m.rows, m.err
}

func TestBuildPaginatedQuery(t *testing.T) {
	tests := []struct {
		name      string
		baseQuery string
		metadata  *page.Metadata
		wantQuery string
	}{
		{
			name:      "Query completa com ordenação e paginação",
			baseQuery: "SELECT * FROM users WHERE active = true",
			metadata: page.NewMetadata(
				page.WithPage(2),
				page.WithLimit(15),
				page.WithSort("name"),
				page.WithOrder("desc"),
			),
			wantQuery: "SELECT * FROM users WHERE active = true ORDER BY name desc LIMIT 15 OFFSET 15",
		},
		{
			name:      "Query apenas com ordenação",
			baseQuery: "SELECT * FROM users",
			metadata: page.NewMetadata(
				page.WithSort("id"),
				page.WithOrder("asc"),
			),
			wantQuery: "SELECT * FROM users ORDER BY id asc",
		},
		{
			name:      "Query apenas com paginação",
			baseQuery: "SELECT * FROM users",
			metadata: page.NewMetadata(
				page.WithPage(3),
				page.WithLimit(10),
			),
			wantQuery: "SELECT * FROM users ORDER BY id asc LIMIT 10 OFFSET 20",
		},
		{
			name:      "Query sem ordenação nem paginação",
			baseQuery: "SELECT * FROM users",
			metadata:  &page.Metadata{},
			wantQuery: "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BuildPaginatedQuery(tt.baseQuery, tt.metadata)
			assert.Equal(t, tt.wantQuery, query)
		})
	}
}

func TestBuildCountQuery(t *testing.T) {
	baseQuery := "SELECT * FROM users WHERE active = true"
	expected := "SELECT COUNT(*) FROM (SELECT * FROM users WHERE active = true) AS count_query"

	result := BuildCountQuery(baseQuery)
	assert.Equal(t, expected, result)
}

func TestCountTotal(t *testing.T) {
	ctx := context.Background()

	// Caso de sucesso
	mockRow := &MockRow{value: 42, err: nil}
	mockExecutor := &MockQueryRowExecutor{row: mockRow}

	total, err := CountTotal(ctx, mockExecutor, "SELECT COUNT(*) FROM users", 1, true)
	assert.NoError(t, err)
	assert.Equal(t, 42, total)
	assert.Equal(t, "SELECT COUNT(*) FROM users", mockExecutor.lastQuery)
	assert.Equal(t, []interface{}{1, true}, mockExecutor.lastArgs)

	// Teste com executor não suportado
	_, err = CountTotal(ctx, "string executor", "SELECT COUNT(*)", 1)
	assert.Error(t, err)
}
