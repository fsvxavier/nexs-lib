package adapters

import (
	"context"
	"testing"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestParseRequest(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	t.Run("valid request", func(t *testing.T) {
		body := `{"key1":"value1","key2":"value2"}`
		req := fasthttp.AcquireRequest()
		req.SetBody([]byte(body))
		req.Header.SetContentType("application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{
			Request: *req,
		})
		defer app.ReleaseCtx(ctx)

		var receiver map[string]interface{}
		fields, err := ParseRequest(ctx, &receiver)

		assert.NoError(t, err)
		assert.NotNil(t, fields)
		assert.Equal(t, "value1", receiver["key1"])
		assert.Equal(t, "value2", receiver["key2"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		body := `invalid json`
		req := fasthttp.AcquireRequest()
		req.SetBody([]byte(body))
		req.Header.SetContentType("application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{
			Request: *req,
		})
		defer app.ReleaseCtx(ctx)

		var receiver map[string]interface{}
		fields, err := ParseRequest(ctx, &receiver)

		assert.Error(t, err)
		assert.Nil(t, fields)
		assert.IsType(t, &domainerrors.UnsupportedMediaTypeError{}, err)
	})

	t.Run("empty request body", func(t *testing.T) {
		req := fasthttp.AcquireRequest()
		req.SetBody([]byte(``))
		req.Header.SetContentType("application/json")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{
			Request: *req,
		})
		defer app.ReleaseCtx(ctx)

		var receiver map[string]interface{}
		fields, err := ParseRequest(ctx, &receiver)

		assert.Error(t, err)
		assert.Nil(t, fields)
		assert.IsType(t, &domainerrors.UnsupportedMediaTypeError{}, err)
	})

	t.Run("request with headers", func(t *testing.T) {
		body := `{"key1":"value1"}`
		req := fasthttp.AcquireRequest()
		req.SetBody([]byte(body))
		req.Header.SetContentType("application/json")
		req.Header.Set("X-Custom-Header", "custom-value")

		ctx := app.AcquireCtx(&fasthttp.RequestCtx{
			Request: *req,
		})
		defer app.ReleaseCtx(ctx)

		var receiver map[string]interface{}
		fields, err := ParseRequest(ctx, &receiver)

		assert.NoError(t, err)
		assert.NotNil(t, fields)
		assert.Equal(t, "value1", receiver["key1"])
		assert.Equal(t, "custom-value", fields["x_custom_header"])
	})
}

func TestParseAdditionalDataFields(t *testing.T) {
	t.Run("additional_data present", func(t *testing.T) {
		dataMap := map[string]interface{}{
			"key1": "value1",
			"additional_data": map[string]interface{}{
				"key2": "value2",
				"key3": "value3",
			},
		}

		parseAdditionalDataFields(dataMap)

		assert.Equal(t, "value1", dataMap["key1"])
		assert.Equal(t, "value2", dataMap["key2"])
		assert.Equal(t, "value3", dataMap["key3"])
		_, exists := dataMap["additional_data"]
		assert.False(t, exists)
	})

	t.Run("additional_data not present", func(t *testing.T) {
		dataMap := map[string]interface{}{
			"key1": "value1",
		}

		parseAdditionalDataFields(dataMap)

		assert.Equal(t, "value1", dataMap["key1"])
		_, exists := dataMap["additional_data"]
		assert.False(t, exists)
	})

	t.Run("additional_data is not a map", func(t *testing.T) {
		dataMap := map[string]interface{}{
			"key1":            "value1",
			"additional_data": "not a map",
		}

		parseAdditionalDataFields(dataMap)

		assert.Equal(t, "value1", dataMap["key1"])
		assert.Equal(t, "not a map", dataMap["additional_data"])
	})
}
func TestReadRequestToMap(t *testing.T) {
	t.Run("valid request body", func(t *testing.T) {
		body := `{"key1":"value1","key2":"value2"}`
		req := fasthttp.AcquireRequest()
		req.SetBody([]byte(body))
		req.Header.SetContentType("application/json")

		ctx := context.Background()
		fields, err := readRequestToMap(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, fields)
		assert.Equal(t, "value1", fields["key1"])
		assert.Equal(t, "value2", fields["key2"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		body := `invalid json`
		req := fasthttp.AcquireRequest()
		req.SetBody([]byte(body))
		req.Header.SetContentType("application/json")

		ctx := context.Background()
		fields, err := readRequestToMap(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, fields)
	})

	t.Run("empty request body", func(t *testing.T) {
		req := fasthttp.AcquireRequest()
		req.SetBody([]byte(``))
		req.Header.SetContentType("application/json")

		ctx := context.Background()
		fields, err := readRequestToMap(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, fields)
	})

	t.Run("request with additional_data", func(t *testing.T) {
		body := `{"key1":"value1","additional_data":{"key2":"value2","key3":"value3"}}`
		req := fasthttp.AcquireRequest()
		req.SetBody([]byte(body))
		req.Header.SetContentType("application/json")

		ctx := context.Background()
		fields, err := readRequestToMap(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, fields)
		assert.Equal(t, "value1", fields["key1"])
		assert.Equal(t, "value2", fields["key2"])
		assert.Equal(t, "value3", fields["key3"])
		_, exists := fields["additional_data"]
		assert.False(t, exists)
	})
}
