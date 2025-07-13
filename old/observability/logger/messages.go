package logger

import "github.com/dock-tech/isis-golang-lib/httpserver/apierrors"

type (
	LogMessage struct {
		RequestHeader   any                    `json:"request_header,omitempty"`
		RequestParams   any                    `json:"request_params,omitempty"`
		Request         any                    `json:"request,omitempty"`
		RequestIp       string                 `json:"request_ip,omitempty"`
		Response        any                    `json:"response,omitempty"`
		Data            interface{}            `json:"data,omitempty"`
		TraceID         string                 `json:"trace_id,omitempty"`
		TransactionUUID string                 `json:"transaction_uuid,omitempty"`
		ClientId        string                 `json:"client_id,omitempty"`
		Entity          string                 `json:"entity,omitempty"`
		Error           apierrors.DockApiError `json:"error_data,omitempty"`
		HTTPStatus      int                    `json:"http_status,omitempty"`
	}
)
