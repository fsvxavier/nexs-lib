package nethttp

import (
	"context"
	"net/http"

	page "github.com/fsvxavier/nexs-lib/paginate"
)

// NetHttpRequest é um adaptador que implementa a interface HttpRequest para net/http
type NetHttpRequest struct {
	req *http.Request
}

// Query implementa o método Query da interface HttpRequest
func (nr *NetHttpRequest) Query(key string) string {
	return nr.req.URL.Query().Get(key)
}

// QueryParam implementa o método QueryParam da interface HttpRequest
func (nr *NetHttpRequest) QueryParam(key string) string {
	return nr.req.URL.Query().Get(key)
}

// NewNetHttpRequest cria um novo adaptador para net/http
func NewNetHttpRequest(req *http.Request) *NetHttpRequest {
	return &NetHttpRequest{req: req}
}

// Parse analisa os parâmetros de paginação de uma requisição net/http
func Parse(ctx context.Context, req *http.Request, sortable ...string) (*page.Metadata, error) {
	adapter := NewNetHttpRequest(req)
	return page.ParseFromRequest(ctx, adapter, sortable...)
}
