package paginate

import (
	"context"
	"strconv"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/validator/schema"
)

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

func (s Set[T]) Remove(v T) {
	delete(s, v)
}

func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]

	return ok
}

// HttpRequest representa uma interface genérica para pedidos HTTP
type HttpRequest interface {
	Query(key string) string
	QueryParam(key string) string
}

// ParseFromRequest analisa os parâmetros de paginação de uma requisição HTTP
func ParseFromRequest(ctx context.Context, req HttpRequest, sortable ...string) (*Metadata, error) {
	// Verifica qual método de obtenção de parâmetros está disponível
	var getParam func(key string) string

	if req != nil {
		// Tenta Query primeiro
		if value := req.Query("page"); value != "" {
			getParam = req.Query
		} else if value := req.QueryParam("page"); value != "" {
			getParam = req.QueryParam
		} else {
			// Nenhum método disponível ou nenhum parâmetro fornecido
			return NewMetadata(), nil
		}
	} else {
		return NewMetadata(), nil
	}

	request := make(map[string]interface{})
	validationErr := domainerrors.NewValidationError("Erro de validação nos parâmetros de paginação", nil)

	// Processa parâmetro de página
	pageStr := getParam("page")
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			validationErr.WithField("page", "O valor deve ser um número inteiro")
			return nil, validationErr
		} else if page <= 0 {
			validationErr.WithField("page", "O valor deve ser maior que zero")
			return nil, validationErr
		}
		request["page"] = page
	}

	// Processa parâmetro de limite
	limitStr := getParam("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			validationErr.WithField("limit", "O valor deve ser um número inteiro")
			return nil, validationErr
		}

		if limit <= 0 {
			validationErr.WithField("limit", "O valor deve ser maior que zero")
			return nil, validationErr
		}

		if limit > 150 {
			limit = 150
		}
		request["limit"] = limit
	}

	// Verificar campos de ordenação permitidos
	allowedFields := Set[string]{}
	for _, field := range sortable {
		allowedFields.Add(field)
	}

	// Processar parâmetros de ordenação
	sort, order := getParam("sort"), getParam("order")
	if sort != "" && len(sortable) > 0 && !allowedFields.Contains(sort) {
		validationErr.WithField("sort", "Campo de ordenação inválido")
		return nil, validationErr
	}

	request["sort"], request["order"] = sort, order

	// Validar contra o esquema JSON
	schemaValidator := schema.NewJSONSchemaValidator()
	result := schemaValidator.ValidateSchema(ctx, request, Schema)
	if !result.Valid {
		// Converter erros de validação para erro de domínio
		validationErr := domainerrors.NewValidationError("Erro de validação nos parâmetros de paginação", nil)
		for field, errors := range result.Errors {
			for _, errorMsg := range errors {
				validationErr.WithField(field, errorMsg)
			}
		}
		for _, globalError := range result.GlobalErrors {
			validationErr.WithField("_global", globalError)
		}
		return nil, validationErr
	}

	// Criar e retornar o metadado
	var page, limit int
	if v, ok := request["page"].(int); ok {
		page = v
	}
	if v, ok := request["limit"].(int); ok {
		limit = v
	}

	return NewMetadata(
		WithPage(page),
		WithLimit(limit),
		WithSort(sort),
		WithOrder(order),
	), nil
}
