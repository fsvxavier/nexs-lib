package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestLoggerMiddleware_StatusOKAndCreated(t *testing.T) {
	app := fiber.New()
	called := false

	app.Use(LoggerMiddleware(io.Discard))
	app.Get("/ok", func(c *fiber.Ctx) error {
		c.Status(http.StatusOK)
		called = true
		return nil
	})
	app.Get("/created", func(c *fiber.Ctx) error {
		c.Status(http.StatusCreated)
		called = true
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if !called {
		t.Error("handler was not called for /ok")
	}

	called = false
	req = httptest.NewRequest(http.MethodGet, "/created", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if !called {
		t.Error("handler was not called for /created")
	}
}

func TestLoggerMiddleware_ErrorPropagation(t *testing.T) {
	app := fiber.New()
	app.Use(LoggerMiddleware(io.Discard))
	app.Get("/fail", func(c *fiber.Ctx) error {
		return fiber.ErrBadRequest
	})

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}
}

func TestLoggerMiddleware_NonOKStatus(t *testing.T) {
	app := fiber.New()
	called := false

	app.Use(LoggerMiddleware(io.Discard))
	app.Get("/notfound", func(c *fiber.Ctx) error {
		c.Status(http.StatusNotFound)
		called = true
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	if !called {
		t.Error("handler was not called for /notfound")
	}
}
