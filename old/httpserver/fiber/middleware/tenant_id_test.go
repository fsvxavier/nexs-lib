package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestTenantIdMiddleware(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(TenantIdMiddleware)

	app.Get("/", func(ctx *fiber.Ctx) error {
		tenantID := ctx.UserContext().Value("tenant_id").(string)
		return ctx.SendString(tenantID)
	})

	tests := []struct {
		name       string
		clientID   string
		expectedID string
	}{
		{
			name:       "Client-Id header present",
			clientID:   "12345",
			expectedID: "12345",
		},
		{
			name:       "Client-Id header absent",
			clientID:   "",
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Client-Id", tt.clientID)

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedID, resp.Request.Header.Get("Client-Id"))
		})
	}
}
