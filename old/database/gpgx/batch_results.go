package gpgx

import (
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type pgxBatchResults struct {
	batchResults pgx.BatchResults
}

func NewPgxBatchResults(batchResults pgx.BatchResults) IBatchResults {
	return pgxBatchResults{
		batchResults: batchResults,
	}
}

func (p pgxBatchResults) Close() {
	p.batchResults.Close()
}

func (p pgxBatchResults) Exec() error {
	_, err := p.batchResults.Exec()
	return err
}

func (p pgxBatchResults) QueryAll(dst interface{}) error {
	rows, err := p.batchResults.Query()
	if err != nil {
		return err
	}
	return pgxscan.ScanAll(dst, rows)
}

func (p pgxBatchResults) QueryOne(dst interface{}) error {
	return p.batchResults.QueryRow().Scan(dst)
}
