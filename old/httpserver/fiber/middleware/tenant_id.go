package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func TenantIdMiddleware(ctx *fiber.Ctx) error {
	clientID := ctx.Get("client-id", ctx.Get("Client-Id", ctx.Get("client_id")))

	c := ctx.UserContext()
	//lint:ignore SA1029 ignore this!
	c = context.WithValue(c, "tenant_id", clientID)
	ctx.SetUserContext(c)

	return ctx.Next()
}
