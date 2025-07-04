package echo

import (
	"context"

	page "github.com/fsvxavier/nexs-lib/paginate"
	"github.com/labstack/echo/v4"
)

// EchoRequest é um adaptador que implementa a interface HttpRequest para Echo
type EchoRequest struct {
	ctx echo.Context
}

// Query implementa o método Query da interface HttpRequest
func (er *EchoRequest) Query(key string) string {
	return er.ctx.QueryParam(key)
}

// QueryParam implementa o método QueryParam da interface HttpRequest
func (er *EchoRequest) QueryParam(key string) string {
	return er.ctx.QueryParam(key)
}

// NewEchoRequest cria um novo adaptador para Echo
func NewEchoRequest(ctx echo.Context) *EchoRequest {
	return &EchoRequest{ctx: ctx}
}

// Parse analisa os parâmetros de paginação de uma requisição Echo
func Parse(ctx context.Context, echoCtx echo.Context, sortable ...string) (*page.Metadata, error) {
	req := NewEchoRequest(echoCtx)
	return page.ParseFromRequest(ctx, req, sortable...)
}
