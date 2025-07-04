package fasthttp

import (
	"context"

	page "github.com/fsvxavier/nexs-lib/paginate"
	"github.com/valyala/fasthttp"
)

// FastHTTPRequest é um adaptador que implementa a interface HttpRequest para FastHTTP
type FastHTTPRequest struct {
	ctx *fasthttp.RequestCtx
}

// Query implementa o método Query da interface HttpRequest
func (fr *FastHTTPRequest) Query(key string) string {
	return string(fr.ctx.QueryArgs().Peek(key))
}

// QueryParam implementa o método QueryParam da interface HttpRequest
func (fr *FastHTTPRequest) QueryParam(key string) string {
	return string(fr.ctx.QueryArgs().Peek(key))
}

// NewFastHTTPRequest cria um novo adaptador para FastHTTP
func NewFastHTTPRequest(ctx *fasthttp.RequestCtx) *FastHTTPRequest {
	return &FastHTTPRequest{ctx: ctx}
}

// Parse analisa os parâmetros de paginação de uma requisição FastHTTP
func Parse(ctx context.Context, fasthttpCtx *fasthttp.RequestCtx, sortable ...string) (*page.Metadata, error) {
	req := NewFastHTTPRequest(fasthttpCtx)
	return page.ParseFromRequest(ctx, req, sortable...)
}
