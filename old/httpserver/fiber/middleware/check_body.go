package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dock-tech/isis-golang-lib/httpserver/apierrors"
	"github.com/gofiber/fiber/v2"
)

func CheckBody(ctx *fiber.Ctx) error {
	if ctx.Body() != nil {
		if json.Valid(ctx.Body()) {
			sBody := strings.TrimSpace(string(ctx.Body()))
			if strings.HasPrefix(sBody, "{") && strings.HasSuffix(sBody, "}") {
				return ctx.Next()
			}
		}

		ctx.Status(http.StatusUnsupportedMediaType)
		err := apierrors.NewDockApiError(
			http.StatusUnsupportedMediaType,
			"415",
			"Unsupported Media Type",
		)

		if setIdErr := err.SetId(ctx.Get("Trace-Id", ctx.Get("trace-Id", ctx.Get("trace-id", ctx.Get("Trace-id"))))); setIdErr != nil {
			return setIdErr
		}

		return ctx.JSON(err)
	}
	return ctx.Next()
}
