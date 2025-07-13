package nethttp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
	jsonEncDec "github.com/dock-tech/isis-golang-lib/json"
	"github.com/dock-tech/isis-golang-lib/observability/tracer"
)

var (
	targetApiServerError = func(target string, statusCode int) error {
		return fmt.Errorf("Integration error with target API - %s - Target Status Code - %d", target, statusCode)
	}
	parseTargetRequestError = func(target string, statusCode int) error {
		return fmt.Errorf("Error when trying to parser response from the target API - %s - Target Status Code - %d", target, statusCode)
	}
)

func WrapResponseErrors(ctx context.Context, response *Response, attribute string) error {
	ctxs, span := tracer.StartSpanFromContext(ctx, "wrapErrors")
	defer span.Finish()

	body := make(map[string]interface{})
	err := jsonEncDec.DecodeReader(bytes.NewReader(response.Body), &body)
	if err != nil {
		return &domainerrors.ServerError{
			InternalError: parseTargetRequestError(response.Request.BaseURL, response.StatusCode),
		}
	}

	switch response.StatusCode {
	case http.StatusBadRequest:
		description := getDescriptionFromBody(ctxs, body)
		if description == nil {
			return &domainerrors.ServerError{
				InternalError: parseTargetRequestError(response.Request.BaseURL, response.StatusCode),
			}
		}
		details := make(map[string][]string)
		details[attribute] = []string{
			*description,
		}
		return &domainerrors.BadRequestError{
			Details: details,
		}
	case http.StatusUnprocessableEntity:
		description := getDescriptionFromBody(ctxs, body)
		if description == nil {
			return &domainerrors.ServerError{
				InternalError: parseTargetRequestError(response.Request.BaseURL, response.StatusCode),
			}
		}
		return &domainerrors.UsecaseError{
			Description: *description,
		}
	case http.StatusNotFound:
		description := getDescriptionFromBody(ctxs, body)
		if description == nil {
			return &domainerrors.ServerError{
				InternalError: parseTargetRequestError(response.Request.BaseURL, response.StatusCode),
			}
		}
		return &domainerrors.NotFoundError{}
	default:
		return &domainerrors.ServerError{
			InternalError: targetApiServerError(response.Request.BaseURL, response.StatusCode),
		}
	}
}

func getDescriptionFromBody(ctx context.Context, body map[string]interface{}) *string {
	_, span := tracer.StartSpanFromContext(ctx, "getDescriptionFromBody")
	defer span.Finish()

	description, ok := body["error"].(map[string]interface{})["description"].(string)
	if !ok {
		return nil
	}
	return &description
}
