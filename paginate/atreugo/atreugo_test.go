package atreugo

import (
	"context"
	"testing"

	"github.com/savsgio/atreugo/v11"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestAtreugoRequest_Query(t *testing.T) {
	// Criar um contexto de requisição FastHTTP
	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?page=2&limit=10&sort=name&order=desc")

	// Criar um contexto Atreugo
	atreugoCtx := atreugo.AcquireRequestCtx(&ctx)

	// Criar um adaptador
	req := NewAtreugoRequest(atreugoCtx)

	// Testar Query
	assert.Equal(t, "2", req.Query("page"))
	assert.Equal(t, "10", req.Query("limit"))
	assert.Equal(t, "name", req.Query("sort"))
	assert.Equal(t, "desc", req.Query("order"))
	assert.Equal(t, "", req.Query("unknown"))
}

func TestAtreugoRequest_QueryParam(t *testing.T) {
	// Criar um contexto de requisição FastHTTP
	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?page=2&limit=10")

	// Criar um contexto Atreugo
	atreugoCtx := atreugo.AcquireRequestCtx(&ctx)

	// Criar um adaptador
	req := NewAtreugoRequest(atreugoCtx)

	// Testar QueryParam
	assert.Equal(t, "2", req.QueryParam("page"))
	assert.Equal(t, "10", req.QueryParam("limit"))
	assert.Equal(t, "", req.QueryParam("unknown"))
}

func TestParse(t *testing.T) {
	// Criar um contexto de requisição FastHTTP
	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?page=2&limit=10&sort=name&order=desc")

	// Criar um contexto Atreugo
	atreugoCtx := atreugo.AcquireRequestCtx(&ctx)

	// Executar Parse
	metadata, err := Parse(context.Background(), atreugoCtx, "id", "name", "email")

	// Verificar resultados
	assert.NoError(t, err)
	assert.Equal(t, 2, metadata.Page.CurrentPage)
	assert.Equal(t, 10, metadata.Page.RecordsPerPage)
	assert.Equal(t, "name", metadata.Sort.Field)
	assert.Equal(t, "desc", metadata.Sort.Order)

	// Testar com campo de ordenação inválido
	ctx = fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/users?sort=invalid")
	atreugoCtx = atreugo.AcquireRequestCtx(&ctx)

	_, err = Parse(context.Background(), atreugoCtx, "id", "name")
	assert.Error(t, err)
}
