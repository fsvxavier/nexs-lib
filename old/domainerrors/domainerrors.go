package domainerrors

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"

	json "github.com/json-iterator/go"
)

var jsonSTD = json.ConfigCompatibleWithStandardLibrary

var ErrNoRows = errors.New("no rows in result set")

// A type to encapsulate use case errors.
type (
	// Errors related to saving or querying data from a database.
	RepositoryError struct {
		InternalError error  `json:"internal_error,omitempty"`
		Description   string `json:"description,omitempty"`
	}

	// A type that encapsulates errors resulting from external services.
	ExternalIntegrationError struct {
		InternalError error
		Metadata      map[string]any `json:"metadata,omitempty"`
		Data          []byte         `json:"data,omitempty"`
		Code          int            `json:"code,omitempty"`
	}

	BadRequestError struct {
		Details map[string][]string `json:"details"`
	}

	// A type to encapsulate validation errors.
	InvalidEntityError struct {
		Details    map[string][]string `json:"details,omitempty"`
		EntityName string              `json:"entity,omitempty"`
	}

	InvalidSchemaError struct {
		Details map[string][]string `json:"details,omitempty"`
	}

	UnsupportedMediaTypeError struct{}

	ForbiddenError struct{}

	UnauthorizedError struct{}

	// Errors related to business rules.
	UsecaseError struct {
		Code        string `json:"code,omitempty"`
		Description string `json:"description,omitempty"`
	}

	NotFoundError struct {
		Description string `json:"description,omitempty"`
	}

	ServerError struct {
		InternalError error
		Metadata      map[string]any `json:"metadata,omitempty"`
		Description   string         `json:"description,omitempty"`
	}

	UnprocessableEntity struct {
		Code        string `json:"code,omitempty"`
		Description string `json:"description,omitempty"`
	}

	TimeoutError struct{}

	ErrTargetServiceUnavailable struct {
		InternalError error
	}
)

func (err *ExternalIntegrationError) Error() string {
	if err.InternalError == nil {
		return "integration error"
	}
	return err.InternalError.Error()
}

func (err *ExternalIntegrationError) Warn() string {
	if err.InternalError == nil {
		return "integration warn"
	}
	return err.InternalError.Error()
}

func (err *ExternalIntegrationError) Extra() string {
	type DockApiError struct {
		Error map[string]any
	}

	var dockError DockApiError
	jsonSTD.Unmarshal(err.Data, &dockError)
	return fmt.Sprintf("%v", dockError.Error["description"])
}

func (*InvalidEntityError) Error() string {
	return "invalid entity"
}

func (s *InvalidSchemaError) Error() string {
	return "Bad Request"
}

func (s *BadRequestError) Error() string {
	return "Bad request"
}

func (u *UsecaseError) Error() string {
	return u.Description
}

func (d *RepositoryError) Error() string {
	return d.Description
}

func (d *ServerError) Error() string {
	return d.Description
}

func (d *NotFoundError) Error() string {
	return d.Description
}

func (d *UnsupportedMediaTypeError) Error() string {
	return "unsupported media type"
}

func (d *ForbiddenError) Error() string {
	return "Forbidden"
}

func (d *UnauthorizedError) Error() string {
	return "Unauthorized"
}

func (d *UnprocessableEntity) Error() string {
	return d.Description
}

func (t *TimeoutError) Error() string {
	return "Timeout Exceeded"
}

func (t *ErrTargetServiceUnavailable) Error() string {
	return "Target Service Unavailable"
}

func (d *UnprocessableEntity) StatusCode() int {
	return http.StatusUnprocessableEntity
}

func NewInvalidEntityError(details map[string][]string, entity any) *InvalidEntityError {
	return &InvalidEntityError{
		Details:    details,
		EntityName: reflect.TypeOf(entity).Name(),
	}
}

func NewUnprocessableEntityError(description string) *UnprocessableEntity {
	return &UnprocessableEntity{
		Description: description,
	}
}

func NewInternalServerError(description string) *ServerError {
	return &ServerError{
		Description: description,
	}
}

func NewNotFoundError(description string) *NotFoundError {
	return &NotFoundError{
		Description: description,
	}
}

func NewBadRequestError(description string) *NotFoundError {
	return &NotFoundError{
		Description: description,
	}
}

type DockError struct {
	Description string `json:"description,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
}

var (
	dockErrors = map[string]DockError{
		"23503": {
			Description: "Dock API error",
			StatusCode:  http.StatusUnprocessableEntity,
		},
		"23505": {
			Description: "Duplicate key value violates unique constraint",
			StatusCode:  http.StatusUnprocessableEntity,
		},
		"500": {
			Description: "Internal server error",
			StatusCode:  http.StatusInternalServerError,
		},
	}

	sqlErrRegex = regexp.MustCompile(`^(.*)\(SQLSTATE (.*)\).*$`)
)

func parseError(err error) string {
	match := sqlErrRegex.FindStringSubmatch(err.Error())
	if len(match) == 3 {
		fmt.Printf("Message: %s / Code: %s\n", match[1], match[2])
		return match[2]
	}

	return ""
}

func findError(err error) DockError {
	dockError, ok := dockErrors[parseError(err)]
	if !ok {
		return dockErrors["500"]
	}
	return dockError
}

func HandleDatabaseError(err error) DockError {
	return findError(err)
}
