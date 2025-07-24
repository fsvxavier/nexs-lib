package adapters

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
	"github.com/dock-tech/isis-golang-lib/httpserver/apierrors"
	logs "github.com/dock-tech/isis-golang-lib/observability/logger"
	log "github.com/dock-tech/isis-golang-lib/observability/logger/zap"
	"github.com/gofiber/fiber/v2"
	json "github.com/json-iterator/go"
)

const (
	clientID = "Client-Id"
)

type (
	ControllerResponse struct {
		Data       any
		Error      error
		StatusCode int
	}

	ResponseError struct {
		Payload *apierrors.DockApiError
		TraceId string
		Status  int
	}
)

func ResponseAdapter(ctx *fiber.Ctx, res ControllerResponse) error {
	if res.Error != nil {
		return processHTTPError(ctx, res)
	} else {
		return processHTTPSuccess(ctx)
	}
}

func processHTTPSuccess(ctx *fiber.Ctx) (err error) {
	requestPayload := make(map[string]any)
	responsePayload := make(map[string]any)
	json.Unmarshal(ctx.Body(), &requestPayload)
	json.Unmarshal(ctx.Response().Body(), &responsePayload)

	traceId := ctx.Get("Uuid")

	logMessage := logs.LogMessage{
		ClientId:   ctx.Get(clientID),
		TraceID:    traceId,
		HTTPStatus: http.StatusOK,
		Request:    requestPayload,
		Response:   responsePayload,
	}

	fields := append(
		[]log.Field{},
		log.Reflect("data", logMessage),
	)
	log.Info(ctx.UserContext(), "Success", fields...)
	return err
}

func processHTTPError(ctx *fiber.Ctx, res ControllerResponse) (err error) {

	err = res.Error
	traceId := ctx.Get("Uuid")
	status := 0
	var payload *apierrors.DockApiError
	var logMessage logs.LogMessage
	message := ""
	switch err := err.(type) {
	case *domainerrors.InvalidEntityError:
		status = http.StatusBadRequest
		payload = apierrors.NewDockApiError(status, statusCodeString(status), "Bad Request")
		for attr, details := range err.Details {
			payload.AddErrorDetail(strings.ToLower(attr), details...)
		}
		message = fmt.Sprintf("%v: %v", err.Error(), err.EntityName)
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *payload,
			Entity:     err.EntityName,
		}
	case *domainerrors.InvalidSchemaError:
		status = http.StatusBadRequest
		payload = apierrors.NewDockApiError(status, statusCodeString(status), "Bad Request")
		for attr, details := range err.Details {
			payload.AddErrorDetail(strings.ToLower(attr), details...)
		}
		message = err.Error()
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *payload,
		}
	case *domainerrors.UnsupportedMediaTypeError:
		status = http.StatusUnsupportedMediaType
		payload = apierrors.NewDockApiError(status, statusCodeString(status), "Unsupported media type")
		message = fmt.Sprintf("%v", err.Error())
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *payload,
		}
	case *domainerrors.UsecaseError:
		status = http.StatusUnprocessableEntity
		payload = apierrors.NewDockApiError(status, err.Code, err.Error())
		message = err.Error()
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *payload,
		}
	case *domainerrors.NotFoundError:
		status = http.StatusNotFound
		payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Error())
		message = err.Error()
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *payload,
		}
	case *domainerrors.RepositoryError:
		status = http.StatusUnprocessableEntity
		payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Error())
		message = err.Description
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *apierrors.NewDockApiError(status, statusCodeString(status), err.InternalError.Error()),
		}
	case *domainerrors.ServerError:
		status = http.StatusInternalServerError
		payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Error())
		message = err.InternalError.Error()
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       err.Metadata,
			Error:      *payload,
		}
	case *domainerrors.ExternalIntegrationError:
		status = err.Code
		if status >= 0 && status <= 399 {
			status = 500
		}

		var payloadResp *apierrors.DockApiError

		switch status {
		case 403, 500, 502, 503, 504:
			status = http.StatusInternalServerError
			payload = apierrors.NewDockApiError(status, statusCodeString(status), "Unable to complete request")
		case 400, 404, 422, 429, 431:
			payload = apierrors.NewDockApiError(status, statusCodeString(status), err.Warn())
			json.Unmarshal(err.Data, &payloadResp)
			payload.InnerError.Description = payload.InnerError.Description + ": " + payloadResp.InnerError.Description
		default:
			payload = apierrors.NewDockApiError(status, statusCodeString(status), "Error to complete request")
		}
		message = fmt.Sprintf("%v: %v", err.Error(), err.Extra())
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Error:      *payload,
		}
	case *domainerrors.UnprocessableEntity:
		status = http.StatusUnprocessableEntity
		payload = apierrors.NewDockApiError(status, err.Code, res.Error.Error())
		message = err.Error()
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       err,
			Error:      *payload,
		}
	case *fiber.Error:
		status = err.Code
		payload = apierrors.NewDockApiError(status, fmt.Sprintf("%v", status), err.Message)
		message = err.Error()
		logMessage = logs.LogMessage{
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       err,
			Error:      *payload,
		}
	default:
		status = http.StatusInternalServerError
		payload = apierrors.NewDockApiError(status, statusCodeString(status), "Internal server error")
		message = err.Error()
		logMessage = logs.LogMessage{
			ClientId:   ctx.Get(clientID),
			TraceID:    traceId,
			HTTPStatus: status,
			Data:       err,
			Error:      *payload,
		}
	}
	payload.SetId(traceId)

	fields := append(
		[]log.Field{},
		log.Reflect("data", logMessage),
	)
	switch status {
	case 400, 404, 422, 429, 431:
		log.Warn(ctx.UserContext(), message, fields...)
	default:
		log.Error(ctx.UserContext(), message, fields...)
	}
	return ctx.Status(status).JSON(payload)
}

func (r *ResponseError) Error() string {
	return statusCodeString(r.Status)
}

func statusCodeString(code int) string {
	switch code {
	case 400, 500:
		return fmt.Sprintf(`%v`, code)
	default:
		return fmt.Sprintf(`%s%v`, os.Getenv("PREFIX_CODE_STRING"), code)
	}
}
