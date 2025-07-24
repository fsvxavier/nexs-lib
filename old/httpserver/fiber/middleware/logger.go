package middleware

import (
	"io"
	"net/http"

	"github.com/dock-tech/isis-golang-lib/httpserver/fiber/adapters"
	"github.com/gofiber/fiber/v2"
)

func LoggerMiddleware(w io.Writer) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		// Capture any error returned by the handler
		err := ctx.Next()
		if err != nil {
			return err
		}

		statusCode := ctx.Response().StatusCode()
		if statusCode == http.StatusOK || statusCode == http.StatusCreated {
			return adapters.ResponseAdapter(
				ctx,
				adapters.ControllerResponse{Error: err},
			)
		}

		return nil
	}
}
