package middleware

import (
	"encoding/json"
	"os"
	"strings"

	logs "github.com/dock-tech/isis-golang-lib/observability/logger"
	log "github.com/dock-tech/isis-golang-lib/observability/logger/zap"
	"github.com/gofiber/fiber/v2"
)

const (
	ADDITIONAL_DATA_KEY = "additional_data"
	clientID            = "Client-Id"
)

var (
	sensitiveData []string = strings.Split(os.Getenv("LOG_SENSITIVE_DATA"), ",")
)

func RequestLoggerMiddleware(ctx *fiber.Ctx) error {
	var requestPayload any

	if ctx.Body() != nil {
		err := json.Unmarshal(ctx.Body(), &requestPayload)
		if err != nil {
			return err
		}
	}

	traceId := ctx.Get("Uuid", ctx.Get("Trace-Id", ctx.Get("Transaction-Uuid")))

	logMessage := logs.LogMessage{
		ClientId:        ctx.Get("client-id", ctx.Get("Client-Id", ctx.Get("client_id"))),
		RequestIp:       ctx.IP(),
		Request:         requestPayload,
		TraceID:         traceId,
		TransactionUUID: traceId,
	}

	params := ctx.AllParams()
	if len(params) > 0 {
		logMessage.RequestParams = ctx.AllParams()
	}

	fields := append(
		[]log.Field{},
		log.Reflect("data", logMessage),
	)
	log.Info(ctx.UserContext(), "Request", fields...)

	return ctx.Next()
}

func removeSensitiveData(data map[string]interface{}) {
	additionalData, additionalDataExists := data[ADDITIONAL_DATA_KEY].(map[string]interface{})

	for _, key := range sensitiveData {
		if additionalDataExists {
			delete(additionalData, key)
		}

		delete(data, key)
	}
}
