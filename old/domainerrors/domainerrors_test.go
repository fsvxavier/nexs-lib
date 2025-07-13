package domainerrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExternalIntegrationError_Error(t *testing.T) {
	err := &ExternalIntegrationError{
		InternalError: errors.New("internal error"),
	}
	assert.Equal(t, "internal error", err.Error())

	err = &ExternalIntegrationError{}
	assert.Equal(t, "integration error", err.Error())
}

func TestExternalIntegrationError_Warn(t *testing.T) {
	err := &ExternalIntegrationError{
		InternalError: errors.New("internal warn"),
	}
	assert.Equal(t, "internal warn", err.Warn())

	err = &ExternalIntegrationError{}
	assert.Equal(t, "integration warn", err.Warn())
}

func TestExternalIntegrationError_Extra(t *testing.T) {
	err := &ExternalIntegrationError{
		Data: []byte(`{"error": {"description": "extra error"}}`),
	}
	assert.Equal(t, "extra error", err.Extra())
}

func TestInvalidEntityError_Error(t *testing.T) {
	err := &InvalidEntityError{}
	assert.Equal(t, "invalid entity", err.Error())
}

func TestInvalidSchemaError_Error(t *testing.T) {
	err := &InvalidSchemaError{}
	assert.Equal(t, "Bad Request", err.Error())
}

func TestUsecaseError_Error(t *testing.T) {
	err := &UsecaseError{
		Description: "usecase error",
	}
	assert.Equal(t, "usecase error", err.Error())
}

func TestRepositoryError_Error(t *testing.T) {
	err := &RepositoryError{
		Description: "repository error",
	}
	assert.Equal(t, "repository error", err.Error())
}

func TestServerError_Error(t *testing.T) {
	err := &ServerError{
		Description: "server error",
	}
	assert.Equal(t, "server error", err.Error())
}

func TestNotFoundError_Error(t *testing.T) {
	err := &NotFoundError{
		Description: "not found error",
	}
	assert.Equal(t, "not found error", err.Error())
}

func TestUnsupportedMediaTypeError_Error(t *testing.T) {
	err := &UnsupportedMediaTypeError{}
	assert.Equal(t, "unsupported media type", err.Error())
}

func TestUnprocessableEntity_Error(t *testing.T) {
	err := &UnprocessableEntity{
		Description: "unprocessable entity",
	}
	assert.Equal(t, "unprocessable entity", err.Error())
}

func TestUnprocessableEntity_StatusCode(t *testing.T) {
	err := &UnprocessableEntity{}
	assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode())
}

func TestNewInvalidEntityError(t *testing.T) {
	details := map[string][]string{"field": {"error"}}
	entity := struct{}{}
	err := NewInvalidEntityError(details, entity)
	assert.Equal(t, "", err.EntityName)
	assert.Equal(t, details, err.Details)
}

func TestNewUnprocessableEntityError(t *testing.T) {
	description := "unprocessable entity"
	err := NewUnprocessableEntityError(description)
	assert.Equal(t, description, err.Description)
}

func TestNewInternalServerError(t *testing.T) {
	description := "internal server error"
	err := NewInternalServerError(description)
	assert.Equal(t, description, err.Description)
}

func TestNewNotFoundError(t *testing.T) {
	description := "not found error"
	err := NewNotFoundError(description)
	assert.Equal(t, description, err.Description)
}

func TestNewBadRequestError(t *testing.T) {
	description := "bad request error"
	err := NewBadRequestError(description)
	assert.Equal(t, description, err.Description)
}

func TestParseError(t *testing.T) {
	err := errors.New(`pq: duplicate key value violates unique constraint "users_pkey" (SQLSTATE 23505)`)
	code := parseError(err)
	assert.Equal(t, "23505", code)

	err = errors.New("some other error")
	code = parseError(err)
	assert.Equal(t, "", code)
}

func TestFindError(t *testing.T) {
	err := errors.New(`pq: duplicate key value violates unique constraint "users_pkey" (SQLSTATE 23505)`)
	dockError := findError(err)
	assert.Equal(t, "Duplicate key value violates unique constraint", dockError.Description)
	assert.Equal(t, http.StatusUnprocessableEntity, dockError.StatusCode)

	err = errors.New("some other error")
	dockError = findError(err)
	assert.Equal(t, "Internal server error", dockError.Description)
	assert.Equal(t, http.StatusInternalServerError, dockError.StatusCode)
}

func TestHandleDatabaseError(t *testing.T) {
	err := errors.New(`pq: duplicate key value violates unique constraint "users_pkey" (SQLSTATE 23505)`)
	dockError := HandleDatabaseError(err)
	assert.Equal(t, "Duplicate key value violates unique constraint", dockError.Description)
	assert.Equal(t, http.StatusUnprocessableEntity, dockError.StatusCode)

	err = errors.New("some other error")
	dockError = HandleDatabaseError(err)
	assert.Equal(t, "Internal server error", dockError.Description)
	assert.Equal(t, http.StatusInternalServerError, dockError.StatusCode)
}
func TestForbiddenError_Error(t *testing.T) {
	err := &ForbiddenError{}
	assert.Equal(t, "Forbidden", err.Error())
}

func TestUnauthorizedError_Error(t *testing.T) {
	err := &UnauthorizedError{}
	assert.Equal(t, "Unauthorized", err.Error())
}

func TestRepositoryError(t *testing.T) {
	err := &RepositoryError{Description: "erro repo"}
	if err.Error() != "erro repo" {
		t.Errorf("esperado 'erro repo', obtido '%s'", err.Error())
	}
}

func TestExternalIntegrationError(t *testing.T) {
	err := &ExternalIntegrationError{InternalError: errors.New("fail")}
	if err.Error() != "fail" {
		t.Errorf("esperado 'fail', obtido '%s'", err.Error())
	}
}

func TestBadRequestError(t *testing.T) {
	err := &BadRequestError{}
	if err.Error() != "Bad request" {
		t.Errorf("esperado 'Bad request', obtido '%s'", err.Error())
	}
}

func TestInvalidEntityError(t *testing.T) {
	err := &InvalidEntityError{}
	if err.Error() != "invalid entity" {
		t.Errorf("esperado 'invalid entity', obtido '%s'", err.Error())
	}
}

func TestInvalidSchemaError(t *testing.T) {
	err := &InvalidSchemaError{}
	if err.Error() != "Bad Request" {
		t.Errorf("esperado 'Bad Request', obtido '%s'", err.Error())
	}
}

func TestUnsupportedMediaTypeError(t *testing.T) {
	err := &UnsupportedMediaTypeError{}
	if err.Error() != "unsupported media type" {
		t.Errorf("esperado 'unsupported media type', obtido '%s'", err.Error())
	}
}

func TestForbiddenError(t *testing.T) {
	err := &ForbiddenError{}
	if err.Error() != "Forbidden" {
		t.Errorf("esperado 'Forbidden', obtido '%s'", err.Error())
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := &UnauthorizedError{}
	if err.Error() != "Unauthorized" {
		t.Errorf("esperado 'Unauthorized', obtido '%s'", err.Error())
	}
}

func TestUsecaseError(t *testing.T) {
	err := &UsecaseError{Description: "negócio"}
	if err.Error() != "negócio" {
		t.Errorf("esperado 'negócio', obtido '%s'", err.Error())
	}
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{Description: "não achei"}
	if err.Error() != "não achei" {
		t.Errorf("esperado 'não achei', obtido '%s'", err.Error())
	}
}

func TestServerError(t *testing.T) {
	err := &ServerError{Description: "erro interno"}
	if err.Error() != "erro interno" {
		t.Errorf("esperado 'erro interno', obtido '%s'", err.Error())
	}
}

func TestUnprocessableEntity(t *testing.T) {
	err := &UnprocessableEntity{Description: "inválido"}
	if err.Error() != "inválido" {
		t.Errorf("esperado 'inválido', obtido '%s'", err.Error())
	}
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{}
	if err.Error() != "Timeout Exceeded" {
		t.Errorf("esperado 'Timeout Exceeded', obtido '%s'", err.Error())
	}
}

func TestErrTargetServiceUnavailable(t *testing.T) {
	err := &ErrTargetServiceUnavailable{}
	if err.Error() != "Target Service Unavailable" {
		t.Errorf("esperado 'Target Service Unavailable', obtido '%s'", err.Error())
	}
}
