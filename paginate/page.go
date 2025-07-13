package paginate

import (
	"context"
	"fmt"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Page representa a estrutura base de paginação
type Page struct {
	Previous       int `json:"previous,omitempty"`
	Next           int `json:"next,omitempty"`
	RecordsPerPage int `json:"records_per_page,omitempty"`
	CurrentPage    int `json:"current_page,omitempty"`
	TotalPages     int `json:"total_pages,omitempty"`
}

// Sort representa a estrutura de ordenação
type Sort struct {
	Field string `json:"field,omitempty"`
	Order string `json:"order,omitempty"`
}

// Metadata combina paginação e ordenação
type Metadata struct {
	Sort      Sort `json:"sort,omitempty"`
	Query     string
	Page      Page `json:"pagination,omitempty"`
	TotalData int  `json:"total_data,omitempty"`
}

// Output representa a saída paginada
type Output struct {
	Content  any      `json:"content"`
	Metadata Metadata `json:"metadata,omitempty"`
}

// Option define o padrão de funções de opções para configuração
type Option func(*Metadata)

// NewMetadata cria uma nova instância de Metadata com valores padrão
func NewMetadata(opts ...Option) *Metadata {
	metadata := &Metadata{
		Page: Page{
			CurrentPage:    1,
			RecordsPerPage: 150,
		},
		Sort: Sort{
			Field: "id",
			Order: "asc",
		},
	}

	// Aplicar opções
	for _, opt := range opts {
		opt(metadata)
	}

	return metadata
}

// WithPage configura a página atual
func WithPage(page int) Option {
	return func(m *Metadata) {
		if page > 0 {
			m.Page.CurrentPage = page
		}
	}
}

// WithLimit configura o limite de registros por página
func WithLimit(limit int) Option {
	return func(m *Metadata) {
		if limit > 0 {
			if limit > 150 {
				limit = 150
			}
			m.Page.RecordsPerPage = limit
		}
	}
}

// WithSort configura o campo de ordenação
func WithSort(field string) Option {
	return func(m *Metadata) {
		if field != "" {
			m.Sort.Field = field
		}
	}
}

// WithOrder configura a direção da ordenação
func WithOrder(order string) Option {
	return func(m *Metadata) {
		if order != "" {
			m.Sort.Order = order
		}
	}
}

// WithQuery configura a query SQL base
func WithQuery(query string) Option {
	return func(m *Metadata) {
		m.Query = query
	}
}

// GetQuery retorna a query armazenada
func (m *Metadata) GetQuery() string {
	return m.Query
}

// SetQuery define a query
func (m *Metadata) SetQuery(query string) {
	m.Query = query
}

// PreparePagination prepara a query com ordenação e paginação
func (m *Metadata) PreparePagination() *Metadata {
	// Ordenação
	if m.Sort.Field != "" && m.Sort.Order != "" {
		m.Query += " ORDER BY " + m.Sort.Field + " " + m.Sort.Order
	}

	// Paginação
	if m.Page.RecordsPerPage > 0 && m.Page.CurrentPage >= 1 {
		offset := (m.Page.CurrentPage - 1) * m.Page.RecordsPerPage
		m.Query += fmt.Sprintf(` LIMIT %d OFFSET %d`, m.Page.RecordsPerPage, offset)
	}
	return m
}

// CalculationTotalPages calcula o total de páginas
func (p *Page) CalculationTotalPages(totalData int) int {
	if totalData == 0 {
		return 0
	}

	totalPages := totalData / p.RecordsPerPage
	if totalData%p.RecordsPerPage > 0 {
		totalPages++
	}

	return totalPages
}

// CalculateNextPreviousPage calcula as páginas anterior e próxima
func (p *Page) CalculateNextPreviousPage(totalData int) {
	p.TotalPages = p.CalculationTotalPages(totalData)

	switch {
	case p.CurrentPage == 1 && p.TotalPages > 1:
		p.Next = p.CurrentPage + 1
	case p.CurrentPage < p.TotalPages && p.CurrentPage > 1:
		p.Previous = p.CurrentPage - 1
		p.Next = p.CurrentPage + 1
	case p.CurrentPage == p.TotalPages:
		p.Previous = p.CurrentPage - 1
		p.Next = 0
	case p.CurrentPage > p.TotalPages:
		p.Previous = p.CurrentPage - 1
		p.Next = 0
	}
}

// PaginationIndices calcula e retorna os índices de início e fim para paginação de slices baseado nos metadados
// Retorna o índice inicial (inclusive) e final (exclusivo) para uso com slices, e se a página solicitada é válida
func PaginationIndices(metadata *Metadata, totalItems int) (int, int, bool) {
	if totalItems == 0 {
		return 0, 0, true
	}

	startIndex := (metadata.Page.CurrentPage - 1) * metadata.Page.RecordsPerPage
	endIndex := startIndex + metadata.Page.RecordsPerPage

	if startIndex >= totalItems {
		// Página inválida (fora dos limites)
		return 0, 0, false
	}

	// Garantir que o endIndex não ultrapasse o tamanho total
	if endIndex > totalItems {
		endIndex = totalItems
	}

	return startIndex, endIndex, true
}

// ApplyPaginationToSlice aplica paginação a uma slice e retorna os dados paginados com metadados atualizados
// Suporta qualquer tipo de slice através de interface{}
func ApplyPaginationToSlice(ctx context.Context, slice interface{}, metadata *Metadata) (*Output, error) {
	// Converte a slice para []interface{} para manipulação genérica
	data, ok := toInterfaceSlice(slice)
	if !ok {
		return nil, domainerrors.New("INVALID_TYPE", "O parâmetro fornecido não é uma slice válida").WithType(domainerrors.ErrorTypeInternal)
	}

	totalItems := len(data)

	// Para listas vazias, retorna imediatamente uma lista vazia com metadados
	if totalItems == 0 {
		return NewOutputWithTotal(ctx, []interface{}{}, 0, metadata)
	}

	startIndex, endIndex, isValid := PaginationIndices(metadata, totalItems)

	if !isValid {
		validationErr := domainerrors.NewValidationError("Página solicitada é maior que o total de páginas disponíveis", nil)
		totalPages := metadata.Page.CalculationTotalPages(totalItems)
		validationErr.WithField("page", fmt.Sprintf("Valor %d é inválido. Total de páginas: %d",
			metadata.Page.CurrentPage, totalPages))
		return nil, validationErr
	}

	// Aplicar a paginação (slice[startIndex:endIndex])
	pagedData := data[startIndex:endIndex]

	// Criar saída com os dados paginados e metadados
	return NewOutputWithTotal(ctx, pagedData, totalItems, metadata)
}

// toInterfaceSlice converte uma slice de qualquer tipo para []interface{}
func toInterfaceSlice(slice interface{}) ([]interface{}, bool) {
	// Usa reflection para verificar e converter diferentes tipos de slices
	switch s := slice.(type) {
	case []interface{}:
		return s, true
	case []map[string]interface{}:
		result := make([]interface{}, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result, true
	case []string:
		result := make([]interface{}, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result, true
	case []int:
		result := make([]interface{}, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result, true
	case []int64:
		result := make([]interface{}, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result, true
	case []float64:
		result := make([]interface{}, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result, true
	default:
		// Tenta converter usando json
		bytes, err := json.Marshal(slice)
		if err != nil {
			return nil, false
		}

		var result []interface{}
		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, false
		}

		return result, true
	}
}

// NewOutput cria uma nova saída paginada
func NewOutput(content any, metadata *Metadata) *Output {
	b, err := json.Marshal(content)
	if err != nil || len(b) == 0 || string(b) == "null" {
		content = make([]any, 0)
	}

	output := &Output{
		Content:  content,
		Metadata: *metadata,
	}

	return output
}

// NewOutputWithTotal cria uma saída paginada com cálculo de total
func NewOutputWithTotal(ctx context.Context, content any, totalData int, metadata *Metadata) (*Output, error) {
	b, err := json.Marshal(content)
	if err != nil || len(b) == 0 || string(b) == "null" {
		content = make([]any, 0)
	}

	metadata.TotalData = totalData
	metadata.Page.CalculateNextPreviousPage(totalData)

	// Para listas vazias, apenas a página 1 é válida
	if totalData == 0 && metadata.Page.CurrentPage > 1 {
		validationErr := domainerrors.NewValidationError("Página solicitada é maior que o total de páginas disponíveis", nil)
		validationErr.WithField("page", fmt.Sprintf("Valor %d é inválido. Para dados vazios, apenas a página 1 é válida",
			metadata.Page.CurrentPage))
		return nil, validationErr
	}

	// Para listas com dados, verificar se a página atual está dentro do limite
	if totalData > 0 && metadata.Page.CurrentPage > metadata.Page.TotalPages {
		validationErr := domainerrors.NewValidationError("Página solicitada é maior que o total de páginas disponíveis", nil)
		validationErr.WithField("page", fmt.Sprintf("Valor %d é inválido. Total de páginas: %d",
			metadata.Page.CurrentPage, metadata.Page.TotalPages))
		return nil, validationErr
	}

	output := &Output{
		Content:  content,
		Metadata: *metadata,
	}

	return output, nil
}
