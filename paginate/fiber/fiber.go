package fiber

import (
	"context"

	page "github.com/fsvxavier/nexs-lib/paginate"
	"github.com/gofiber/fiber/v2"
)

// FiberRequest é um adaptador que implementa a interface HttpRequest para Fiber
type FiberRequest struct {
	ctx *fiber.Ctx
}

// Query implementa o método Query da interface HttpRequest
func (fr *FiberRequest) Query(key string) string {
	return fr.ctx.Query(key)
}

// QueryParam implementa o método QueryParam da interface HttpRequest
func (fr *FiberRequest) QueryParam(key string) string {
	return fr.ctx.Query(key)
}

// NewFiberRequest cria um novo adaptador para Fiber
func NewFiberRequest(ctx *fiber.Ctx) *FiberRequest {
	return &FiberRequest{ctx: ctx}
}

// Parse analisa os parâmetros de paginação de uma requisição Fiber
func Parse(ctx context.Context, fiberCtx *fiber.Ctx, sortable ...string) (*page.Metadata, error) {
	req := NewFiberRequest(fiberCtx)
	return page.ParseFromRequest(ctx, req, sortable...)
}
