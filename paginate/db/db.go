package db

import (
	"context"
	"strconv"

	page "github.com/fsvxavier/nexs-lib/paginate"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
)

// CountTotal executa uma consulta para contar o total de registros
func CountTotal(ctx context.Context, querier interface{}, countQuery string, args ...interface{}) (int, error) {
	var total int

	// Verifica qual tipo de interface está sendo fornecida
	switch q := querier.(type) {
	case QueryRowExecutor:
		row := q.QueryRow(ctx, countQuery, args...)
		err := row.Scan(&total)
		if err != nil {
			dbErr := domainerrors.NewDatabaseError("Erro ao contar registros", err)
			return 0, dbErr
		}
	default:
		return 0, domainerrors.NewBusinessError("INVALID_QUERIER", "O executor de consulta é inválido")
	}

	return total, nil
}

// QueryRowExecutor define uma interface para executar consultas que retornam uma única linha
type QueryRowExecutor interface {
	QueryRow(ctx context.Context, query string, args ...interface{}) RowScanner
}

// RowScanner define uma interface para scanear valores de uma linha
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// BuildPaginatedQuery constrói uma consulta paginada baseada nos metadados
func BuildPaginatedQuery(baseQuery string, metadata *page.Metadata) string {
	query := baseQuery

	// Adicionar ordenação se presente
	if metadata.Sort.Field != "" && metadata.Sort.Order != "" {
		query += " ORDER BY " + metadata.Sort.Field + " " + metadata.Sort.Order
	}

	// Adicionar paginação se presente
	if metadata.Page.RecordsPerPage > 0 && metadata.Page.CurrentPage >= 1 {
		offset := (metadata.Page.CurrentPage - 1) * metadata.Page.RecordsPerPage
		query += " LIMIT " + strconv.Itoa(metadata.Page.RecordsPerPage) + " OFFSET " + strconv.Itoa(offset)
	}

	return query
}

// BuildCountQuery constrói uma consulta para contagem total de registros
func BuildCountQuery(baseQuery string) string {
	return "SELECT COUNT(*) FROM (" + baseQuery + ") AS count_query"
}

// ExecuteQuery executa uma consulta usando a configuração de paginação
func ExecuteQuery(
	ctx context.Context,
	querier interface{},
	metadata *page.Metadata,
	baseQuery string,
	args []interface{},
	resultProcessor func(rows interface{}) (interface{}, error),
) (*page.Output, error) {
	// Construir consulta de contagem
	countQuery := BuildCountQuery(baseQuery)

	// Executar consulta de contagem
	total, err := CountTotal(ctx, querier, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Construir consulta paginada
	paginatedQuery := BuildPaginatedQuery(baseQuery, metadata)

	// Executar a consulta paginada com base no tipo de querier
	var rows interface{}
	var queryErr error

	switch q := querier.(type) {
	case QueryExecutor:
		rows, queryErr = q.Query(ctx, paginatedQuery, args...)
	default:
		return nil, domainerrors.NewBusinessError("INVALID_QUERIER", "O executor de consulta não implementa a interface QueryExecutor")
	}

	if queryErr != nil {
		return nil, domainerrors.NewDatabaseError("Erro ao executar consulta paginada", queryErr)
	}

	// Processar os resultados usando a função fornecida
	results, err := resultProcessor(rows)
	if err != nil {
		return nil, err
	}

	// Criar saída paginada com total
	return page.NewOutputWithTotal(ctx, results, total, metadata)
}

// QueryExecutor define uma interface para executar consultas
type QueryExecutor interface {
	Query(ctx context.Context, query string, args ...interface{}) (interface{}, error)
}
