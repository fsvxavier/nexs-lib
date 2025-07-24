package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"

	"github.com/dock-tech/isis-golang-lib/httpserver/apierrors"
)

func ContentTypeMiddleware(method string, allowedContentTypes ...string) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		if method != ctx.Method() {
			return ctx.Next()
		}

		contentType := ctx.Get("Content-Type")
		if lo.Contains(allowedContentTypes, contentType) {
			return ctx.Next()
		}

		ctx.Status(http.StatusUnsupportedMediaType)
		err := apierrors.NewDockApiError(
			http.StatusUnsupportedMediaType,
			"415",
			"Unsupported Media Type",
		)

		err.SetId(ctx.Get("Trace-Id"))

		return ctx.JSON(err)
	}
}
