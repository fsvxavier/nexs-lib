package gpgx

import (
	"github.com/jackc/pgx/v5"
)

type pgxBatch struct {
	batch *pgx.Batch
}

func NewPgxBatch() IBatch {
	return pgxBatch{
		batch: new(pgx.Batch),
	}
}

func (pb pgxBatch) Queue(query string, arguments ...any) {
	pb.batch.Queue(query, arguments...)
}

func (pb pgxBatch) getBatch() *pgx.Batch {
	return pb.batch
}
