package middleware

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityHeaders_Present(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders())
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	resp, err := app.Test(requestPath("/"))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"))
	assert.Contains(t, resp.Header.Get("Content-Security-Policy"), "default-src 'self'")
	assert.Contains(t, resp.Header.Get("Referrer-Policy"), "strict-origin-when-cross-origin")
	assert.Contains(t, resp.Header.Get("Permissions-Policy"), "geolocation=()")
	assert.Equal(t, "same-origin", resp.Header.Get("Cross-Origin-Opener-Policy"))
}

func TestSecurityHeaders_HSTSBehindProxy(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders())
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Contains(t, resp.Header.Get("Strict-Transport-Security"), "max-age=31536000")
}

func TestCORS_AllowsConfiguredOrigin(t *testing.T) {
	app := fiber.New()
	app.Use(NewCORSMiddleware(cfgForTest()))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	req, _ := http.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "86400", resp.Header.Get("Access-Control-Max-Age"))
}
