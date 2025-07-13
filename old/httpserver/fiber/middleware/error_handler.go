package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/dock-tech/isis-golang-lib/httpserver/fiber/adapters"
)

func ApplicationErrorHandler(ctx *fiber.Ctx, err error) error {
	return adapters.ResponseAdapter(
		ctx,
		adapters.ControllerResponse{Error: err},
	)
}
