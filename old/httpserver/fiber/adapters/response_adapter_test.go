package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
	"github.com/dock-tech/isis-golang-lib/httpserver/apierrors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestStatusCodeString(t *testing.T) {

	os.Setenv("PREFIX_CODE_STRING", "DOCKAPI-")

	assert.Equal(t, "400", statusCodeString(400))
	assert.Equal(t, "500", statusCodeString(500))
	assert.Equal(t, "DOCKAPI-404", statusCodeString(404))
}

func TestProcessHTTPSuccess(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

	// Mock response body
	responseBody := `{"key":"value"}`
	ctx.Response().SetBody([]byte(responseBody))

	// Mock headers
	ctx.Request().Header.Set("transaction_uuid", "test-transaction-uuid")
	ctx.Request().Header.Set("client_id", "test-client-id")

	err := processHTTPSuccess(ctx)
	assert.NoError(t, err)

	// Check if the response body was unmarshalled correctly
	responsePayload := make(map[string]any)
	err = json.Unmarshal(ctx.Response().Body(), &responsePayload)
	assert.NoError(t, err)
	assert.Equal(t, "value", responsePayload["key"])

	// Check if the log message was created correctly
	transactionUUID := ctx.Get("transaction_uuid")
	clientID := ctx.Get("client_id")
	assert.Equal(t, "test-transaction-uuid", transactionUUID)
	assert.Equal(t, "test-client-id", clientID)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
}

func TestProcessHTTPError(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

	tests := []struct {
		errorStc       error
		name           string
		expectedError  string
		expectedStatus int
	}{
		{
			name:           "InvalidEntityError",
			errorStc:       &domainerrors.InvalidEntityError{EntityName: "TestEntity", Details: map[string][]string{"field": {"error"}}},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Bad Request",
		},
		{
			name:           "InvalidSchemaError",
			errorStc:       &domainerrors.InvalidSchemaError{Details: map[string][]string{"field": {"error"}}},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Bad Request",
		},
		{
			name:           "UnsupportedMediaTypeError",
			errorStc:       &domainerrors.UnsupportedMediaTypeError{},
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedError:  "Unsupported media type",
		},
		{
			name:           "UsecaseError",
			errorStc:       &domainerrors.UsecaseError{},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Unprocessable Entity",
		},
		{
			name:           "NotFoundError",
			errorStc:       &domainerrors.NotFoundError{},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Not Found",
		},
		{
			name:           "RepositoryError",
			errorStc:       &domainerrors.RepositoryError{Description: "Repo error", InternalError: fmt.Errorf("internal error")},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Unprocessable Entity",
		},
		{
			name:           "ServerError",
			errorStc:       &domainerrors.ServerError{InternalError: fmt.Errorf("internal error")},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Internal server error",
		},
		{
			name:           "ExternalIntegrationError",
			errorStc:       &domainerrors.ExternalIntegrationError{Code: 500, Data: []byte(`{"inner_error": "error"}`)},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Unable to complete request",
		},
		{
			name:           "UnprocessableEntity",
			errorStc:       &domainerrors.UnprocessableEntity{},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Unprocessable Entity",
		},
		{
			name:           "FiberError",
			errorStc:       &fiber.Error{Code: 400, Message: "Fiber error"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Fiber error",
		},
		{
			name:           "DefaultError",
			errorStc:       fmt.Errorf("default error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx.Request().Header.Set("transaction_uuid", "test-transaction-uuid")
			ctx.Request().Header.Set("client_id", "test-client-id")

			res := ControllerResponse{Error: tt.errorStc}
			err := processHTTPError(ctx, res)
			assert.NoError(t, err)

			var payload apierrors.DockApiError
			err = json.Unmarshal(ctx.Response().Body(), &payload)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, ctx.Response().StatusCode())
		})
	}
}
func TestResponseAdapter_Success(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

	// Mock response body
	responseBody := `{"foo":"bar"}`
	ctx.Response().SetBody([]byte(responseBody))
	ctx.Request().Header.Set("client_id", "test-client-id")
	ctx.Request().Header.Set("Uuid", "test-uuid")

	res := ControllerResponse{
		Data:       map[string]string{"foo": "bar"},
		Error:      nil,
		StatusCode: http.StatusOK,
	}

	err := ResponseAdapter(ctx, res)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
}

func TestResponseAdapter_Error(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

	ctx.Request().Header.Set("client_id", "test-client-id")
	ctx.Request().Header.Set("Uuid", "test-uuid")

	res := ControllerResponse{
		Error:      &domainerrors.InvalidEntityError{EntityName: "TestEntity", Details: map[string][]string{"field": {"error"}}},
		StatusCode: http.StatusBadRequest,
	}

	err := ResponseAdapter(ctx, res)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, ctx.Response().StatusCode())
}
