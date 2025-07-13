package pagination

import (
	"fmt"
	"strconv"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
	"github.com/dock-tech/isis-golang-lib/set"
	"github.com/dock-tech/isis-golang-lib/validator/schema"
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type (
	Pagination struct {
		Previous       int `json:"previous,omitempty"`
		Next           int `json:"next,omitempty"`
		RecordsPerPage int `json:"records_per_page,omitempty"`
		CurrentPage    int `json:"current_page,omitempty"`
		TotalPages     int `json:"total_pages,omitempty"`
	}

	Sort struct {
		Field string `json:"field,omitempty"`
		Order string `json:"order,omitempty"`
	}

	Metadata struct {
		Sort       Sort `json:"sort,omitempty"`
		query      string
		Pagination Pagination `json:"pagination,omitempty"`
	}
)

func ParseMetadata(ctxFiber *fiber.Ctx, sortable ...string) (metadata *Metadata, err error) {
	request := make(map[string]interface{})
	// convert limit and page to int, and validate each one as a min value
	var page, limit int

	dae := domainerrors.InvalidEntityError{}
	dae.Details = make(map[string][]string)

	pageStr := ctxFiber.Query("page")
	if pageStr != "" {
		page, err = strconv.Atoi(ctxFiber.Query("page"))
		if err != nil {
			dae.Details["page"] = []string{"INVALID_DATA_TYPE"}
			return nil, &dae
		} else if page <= 0 {
			dae.Details["page"] = []string{"INVALID_VALUE"}
			return nil, &dae
		}

		request["page"] = page
	}

	limitStr := ctxFiber.Query("limit")
	if limitStr != "" {
		limit, err = strconv.Atoi(ctxFiber.Query("limit"))
		if err != nil {
			dae.Details["limit"] = []string{"INVALID_DATA_TYPE"}
			return nil, &dae
		}

		if limit > 150 {
			limit = 150
		}
		request["limit"] = limit
	}

	allowedValues := set.Set[string]{}
	for idx := range sortable {
		allowedValues.Add(sortable[idx])
	}

	sort, order := ctxFiber.Query("sort"), ctxFiber.Query("order")
	if sort != "" && !allowedValues.Contains(sort) {
		dae.Details["sort"] = []string{"INVALID_VALUE"}
		return nil, &dae
	}

	request["sort"], request["order"] = sort, order
	if err = schema.Validate(request, paginationSchema); err != nil {
		return nil, err
	}

	return NewMetadata(page, limit, sort, order), nil
}

func SetQuery(metadata *Metadata, query string) {
	metadata.query = query
}

func GetQuery(metadata *Metadata) string {
	return metadata.query
}

func PreparePagination(metadata *Metadata) *Metadata {
	// order by
	if metadata.Sort.Field != "" && metadata.Sort.Order != "" {
		metadata.query += " ORDER BY " + metadata.Sort.Field + " " + metadata.Sort.Order
	}

	// pagination
	if metadata.Pagination.RecordsPerPage > 0 && metadata.Pagination.CurrentPage >= 1 {
		metadata.query += fmt.Sprintf(` LIMIT %d OFFSET %d`, metadata.Pagination.RecordsPerPage, metadata.Pagination.CurrentPage-1)
	}
	return metadata
}

func NewMetadata(page, limit int, sortField, order string) *Metadata {
	metadata := &Metadata{
		Pagination: Pagination{
			CurrentPage:    page,
			RecordsPerPage: limit,
		},
		Sort: Sort{
			Field: sortField,
			Order: order,
		},
	}

	if metadata.Pagination.RecordsPerPage == 0 || metadata.Pagination.RecordsPerPage > 150 {
		metadata.Pagination.RecordsPerPage = 150
	}

	if metadata.Sort.Field == "" {
		metadata.Sort.Field = "id"
	}

	if metadata.Sort.Order == "" {
		metadata.Sort.Order = "asc"
	}

	return metadata
}

type PaginatedOutput struct {
	Content  any       `json:"content"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

func NewPaginatedOutput(body any, pagination *Metadata) *PaginatedOutput {
	b, err := json.Marshal(body)
	if err != nil || len(b) == 0 || string(b) == "null" {
		body = make([]any, 0)
	}

	output := &PaginatedOutput{
		Content:  body,
		Metadata: pagination,
	}

	return output
}

func NewPaginatedOutputWithTotal(body any, total *int, pagination *Metadata) (*PaginatedOutput, error) {
	b, err := json.Marshal(body)
	if err != nil || len(b) == 0 || string(b) == "null" {
		body = make([]any, 0)
	}

	output := &PaginatedOutput{
		Content:  body,
		Metadata: pagination,
	}

	output.Metadata.Pagination.CalculateNextPreviousPage(total)

	if output.Metadata.Pagination.CurrentPage > output.Metadata.Pagination.TotalPages {
		dae := domainerrors.InvalidEntityError{}
		dae.Details = make(map[string][]string)
		dae.Details["page"] = []string{"INVALID_VALUE"}
		return nil, &dae

	}

	return output, nil
}

func (p *Pagination) CalculationTotalPage(totalData *int) int {
	if totalData == nil || *totalData == 0 {
		return 0
	}

	totalPages := *totalData / p.RecordsPerPage
	if *totalData%p.RecordsPerPage > 0 {
		totalPages++
	}

	return totalPages
}

// Calculator the Next/Previous Page
func (p *Pagination) CalculateNextPreviousPage(totalData *int) {

	p.TotalPages = p.CalculationTotalPage(totalData)

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
