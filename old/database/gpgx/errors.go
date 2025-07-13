package gpgx

import (
	"fmt"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
)

var (
	DatabaseQueryError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database query", entity)
	}
	isEmptyResultError = func(err error) bool {
		if err == nil {
			return false
		}
		if pgErr := NewPgError(err.Error()); pgErr.IsEmptyResult() {
			return pgErr.IsEmptyResult()
		}
		return false
	}
	DatabaseQueryCountError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database count query", entity)
	}
	DatabaseQueryRowsError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database rows query", entity)
	}
	DatabaseQueryRowError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database row query", entity)
	}
	DatabaseExecQueryError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database exec query", entity)
	}
	DatabaseQueryRowsScanError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan rows", entity)
	}
	DatabaseQueryRowScanError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan row", entity)
	}
	DatabaseQueryCountScanError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan count", entity)
	}
	DatabaseQueryCountRowsError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database count rows query", entity)
	}
	DatabaseQueryRowsCountError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database rows count query", entity)
	}
	DatabaseQueryRowCountError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database row count query", entity)
	}
	DatabaseQueryRowCountScanError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan row count", entity)
	}
	DatabaseQueryRowsCountScanError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan rows count", entity)
	}
	DatabaseQueryCountRowsScanError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan count rows", entity)
	}
	DatabaseQueryRowsCountRowsError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database rows count rows query", entity)
	}
	DatabaseQueryRowCountRowsError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to run a database row count rows query", entity)
	}
	DatabaseQueryRowScanRowsError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan row rows", entity)
	}
	DatabaseQueryRowsScanRowsError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan rows rows", entity)
	}
	DatabaseQueryCountScanRowsError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan count rows", entity)
	}
	ScanQueryError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Unable to scan values", entity)
	}
	DatabaseInsertQueryError = func(entity string) string {
		return fmt.Sprintf("Error inserting %s: Unable to run a database query", entity)
	}
	DatabaseNotFoundError = func(entity string) string {
		return fmt.Sprintf("Error retrieving %s: Value not found in database", entity)
	}
	DatabaseUpdateQueryError = func(entity string) string {
		return fmt.Sprintf("Error updating %s: Unable to run a database query", entity)
	}
)

func WrapErrors(entity string, errType func(string) string, err error) error {
	switch {
	case IsEmptyResultError(err):
		return &domainerrors.UnprocessableEntity{Description: DatabaseNotFoundError(entity)}
	case err != nil:
		return &domainerrors.RepositoryError{InternalError: err, Description: errType(entity)}
	default:
		return nil
	}
}

type NotConnectedError struct{}

func (nce NotConnectedError) Error() string {
	return "not connected"
}

type DbError struct {
	Message string
}

func (e DbError) Error() string {
	return e.Message
}
