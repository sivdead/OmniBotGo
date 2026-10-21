package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	bearerPrefix = "Bearer "

	jsonKeyCode    = "code"
	jsonKeyMessage = "message"
	jsonKeyData    = "data"
)

// RequireAPIKey enforces Authorization: Bearer <apiKey> on admin API routes.
// When apiKey is empty, all requests are rejected so the admin surface cannot
// run unauthenticated by misconfiguration.
func RequireAPIKey(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if apiKey == "" {
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{
				jsonKeyCode:    http.StatusServiceUnavailable,
				jsonKeyMessage: "admin API key is not configured (set HTTP_API_KEY)",
				jsonKeyData:    nil,
			})
		}

		auth := c.Get(fiber.HeaderAuthorization)
		if !strings.HasPrefix(auth, bearerPrefix) {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				jsonKeyCode:    http.StatusUnauthorized,
				jsonKeyMessage: "missing or invalid Authorization bearer token",
				jsonKeyData:    nil,
			})
		}

		provided := strings.TrimPrefix(auth, bearerPrefix)
		if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				jsonKeyCode:    http.StatusUnauthorized,
				jsonKeyMessage: "invalid API key",
				jsonKeyData:    nil,
			})
		}

		return c.Next()
	}
}
