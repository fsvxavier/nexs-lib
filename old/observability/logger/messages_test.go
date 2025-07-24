package logger

import (
	"testing"

	"github.com/dock-tech/isis-golang-lib/httpserver/apierrors"
	"github.com/stretchr/testify/assert"
)

func TestLogMessage(t *testing.T) {
	t.Run("should create log message with all fields", func(t *testing.T) {
		expected := LogMessage{
			RequestHeader:   map[string]string{"content-type": "application/json"},
			RequestParams:   map[string]string{"id": "123"},
			Request:         map[string]string{"name": "test"},
			RequestIp:       "127.0.0.1",
			Response:        map[string]string{"status": "success"},
			Data:            map[string]string{"key": "value"},
			TraceID:         "trace-123",
			TransactionUUID: "trans-123",
			ClientId:        "client-123",
			Entity:          "test-entity",
			Error:           apierrors.DockApiError{},
			HTTPStatus:      200,
		}

		msg := LogMessage{
			RequestHeader:   expected.RequestHeader,
			RequestParams:   expected.RequestParams,
			Request:         expected.Request,
			RequestIp:       expected.RequestIp,
			Response:        expected.Response,
			Data:            expected.Data,
			TraceID:         expected.TraceID,
			TransactionUUID: expected.TransactionUUID,
			ClientId:        expected.ClientId,
			Entity:          expected.Entity,
			Error:           expected.Error,
			HTTPStatus:      expected.HTTPStatus,
		}

		assert.Equal(t, expected, msg)
	})
}
