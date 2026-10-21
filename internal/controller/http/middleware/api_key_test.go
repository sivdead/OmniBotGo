package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRequireAPIKey_RejectsWhenUnset(t *testing.T) {
	app := fiber.New()
	app.Get("/api/v1/channels", RequireAPIKey(""), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestRequireAPIKey_RejectsMissingAndWrongKey(t *testing.T) {
	app := fiber.New()
	app.Get("/api/v1/channels", RequireAPIKey("secret-key"), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAPIKey_AllowsValidBearer(t *testing.T) {
	app := fiber.New()
	app.Get("/api/v1/channels", RequireAPIKey("secret-key"), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	req.Header.Set("Authorization", "Bearer secret-key")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "ok", string(body))
}
