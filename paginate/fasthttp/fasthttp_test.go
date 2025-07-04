package fasthttp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestFastHTTPRequest_Query(t *testing.T) {
	// Criar um contexto de requisição FastHTTP
	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?page=3&limit=15&sort=id&order=asc")

	// Criar um adaptador
	req := NewFastHTTPRequest(&ctx)

	// Testar Query
	assert.Equal(t, "3", req.Query("page"))
	assert.Equal(t, "15", req.Query("limit"))
	assert.Equal(t, "id", req.Query("sort"))
	assert.Equal(t, "asc", req.Query("order"))
	assert.Equal(t, "", req.Query("unknown"))
}

func TestFastHTTPRequest_QueryParam(t *testing.T) {
	// Criar um contexto de requisição FastHTTP
	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?page=3&limit=15")

	// Criar um adaptador
	req := NewFastHTTPRequest(&ctx)

	// Testar QueryParam
	assert.Equal(t, "3", req.QueryParam("page"))
	assert.Equal(t, "15", req.QueryParam("limit"))
	assert.Equal(t, "", req.QueryParam("unknown"))
}

func TestParse(t *testing.T) {
	// Criar um contexto de requisição FastHTTP
	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?page=3&limit=15&sort=created_at&order=desc")

	// Executar Parse
	metadata, err := Parse(context.Background(), &ctx, "id", "name", "created_at")

	// Verificar resultados
	assert.NoError(t, err)
	assert.Equal(t, 3, metadata.Page.CurrentPage)
	assert.Equal(t, 15, metadata.Page.RecordsPerPage)
	assert.Equal(t, "created_at", metadata.Sort.Field)
	assert.Equal(t, "desc", metadata.Sort.Order)

	// Testar com página inválida
	ctx = fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?page=-1")

	_, err = Parse(context.Background(), &ctx, "id", "name")
	assert.Error(t, err)

	// Testar com limite inválido
	ctx = fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?limit=abc")

	_, err = Parse(context.Background(), &ctx, "id", "name")
	assert.Error(t, err)
}
