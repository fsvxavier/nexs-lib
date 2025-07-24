package middleware

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TODO Verificar os testes, se estão corretos, e se o middleware está funcionando como esperado.
func TestCheckBody(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Valid JSON body",
			body:           `{"key": "value"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON body",
			body:           `{"key": "value"`,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Empty body",
			body:           "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-JSON body",
			body:           "plain text",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})
			app.Use(CheckBody)
			app.Post("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req, _ := http.NewRequest("POST", "/test", strings.NewReader(tt.body))
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
