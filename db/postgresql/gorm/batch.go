package gorm

import (
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"gorm.io/gorm"
)

// BatchQuery representa uma consulta em lote
type BatchQuery struct {
	sql  string
	args []interface{}
}

// Batch implementa a interface common.IBatch usando GORM
type Batch struct {
	queries []BatchQuery
}

// NewBatch cria um novo lote de consultas
func NewBatch() *Batch {
	return &Batch{
		queries: make([]BatchQuery, 0),
	}
}

// Queue adiciona uma consulta ao lote
func (b *Batch) Queue(query string, arguments ...any) {
	b.queries = append(b.queries, BatchQuery{
		sql:  query,
		args: arguments,
	})
}

// GetBatch retorna o objeto de lote interno
func (b *Batch) GetBatch() interface{} {
	return b
}

// BatchResults implementa a interface common.IBatchResults usando GORM
type BatchResults struct {
	results []*gorm.DB
	index   int
}

// QueryOne escaneia o resultado atual em um destino
func (br *BatchResults) QueryOne(dst interface{}) error {
	if br.index >= len(br.results) {
		return common.ErrNoRows
	}

	result := br.results[br.index]
	br.index++

	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar QueryOne no resultado do batch")
	}

	return nil
}

// QueryAll escaneia todos os resultados em um destino
func (br *BatchResults) QueryAll(dst interface{}) error {
	if br.index >= len(br.results) {
		return common.ErrNoRows
	}

	result := br.results[br.index]
	br.index++

	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar QueryAll no resultado do batch")
	}

	return nil
}

// Exec executa o comando atual
func (br *BatchResults) Exec() error {
	if br.index >= len(br.results) {
		return common.ErrNoRows
	}

	result := br.results[br.index]
	br.index++

	if result.Error != nil {
		return WrapError(result.Error, "falha ao executar Exec no resultado do batch")
	}

	return nil
}

// Close fecha os resultados
func (br *BatchResults) Close() error {
	// Não há nada para fechar explicitamente no GORM
	return nil
}
