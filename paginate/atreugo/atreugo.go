package atreugo

import (
	"context"

	page "github.com/fsvxavier/nexs-lib/paginate"
	"github.com/savsgio/atreugo/v11"
)

// AtreugoRequest é um adaptador que implementa a interface HttpRequest para Atreugo
type AtreugoRequest struct {
	ctx *atreugo.RequestCtx
}

// Query implementa o método Query da interface HttpRequest
func (ar *AtreugoRequest) Query(key string) string {
	return string(ar.ctx.QueryArgs().Peek(key))
}

// QueryParam implementa o método QueryParam da interface HttpRequest
func (ar *AtreugoRequest) QueryParam(key string) string {
	return string(ar.ctx.QueryArgs().Peek(key))
}

// NewAtreugoRequest cria um novo adaptador para Atreugo
func NewAtreugoRequest(ctx *atreugo.RequestCtx) *AtreugoRequest {
	return &AtreugoRequest{ctx: ctx}
}

// Parse analisa os parâmetros de paginação de uma requisição Atreugo
func Parse(ctx context.Context, atreugoCtx *atreugo.RequestCtx, sortable ...string) (*page.Metadata, error) {
	req := NewAtreugoRequest(atreugoCtx)
	return page.ParseFromRequest(ctx, req, sortable...)
}
