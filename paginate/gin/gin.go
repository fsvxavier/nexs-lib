package gin

import (
	"context"

	page "github.com/fsvxavier/nexs-lib/paginate"
	"github.com/gin-gonic/gin"
)

// GinRequest é um adaptador que implementa a interface HttpRequest para Gin
type GinRequest struct {
	ctx *gin.Context
}

// Query implementa o método Query da interface HttpRequest
func (gr *GinRequest) Query(key string) string {
	return gr.ctx.Query(key)
}

// QueryParam implementa o método QueryParam da interface HttpRequest
func (gr *GinRequest) QueryParam(key string) string {
	return gr.ctx.Query(key)
}

// NewGinRequest cria um novo adaptador para Gin
func NewGinRequest(ctx *gin.Context) *GinRequest {
	return &GinRequest{ctx: ctx}
}

// Parse analisa os parâmetros de paginação de uma requisição Gin
func Parse(ctx context.Context, ginCtx *gin.Context, sortable ...string) (*page.Metadata, error) {
	req := NewGinRequest(ginCtx)
	return page.ParseFromRequest(ctx, req, sortable...)
}
