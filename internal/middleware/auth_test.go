package middleware

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

func cfgForTest() config.Config {
	return config.Config{
		AppName:               "fiber-boilerplate",
		JWTSecret:             "test-secret-that-is-long-enough-0123456789abcdef",
		JWTAccessTokenExpiry:  15 * 60 * 1e9,
		JWTRefreshTokenExpiry: 720 * 3600 * 1e9,
	}
}

func buildProtectedApp() *fiber.App {
	app := fiber.New()
	app.Get("/protected", NewAuthMiddleware(cfgForTest()), func(c *fiber.Ctx) error {
		return c.JSON(map[string]any{"user_id": pkg.GetUserID(c)})
	})
	return app
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	app := buildProtectedApp()

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_ValidAccessToken(t *testing.T) {
	app := buildProtectedApp()

	tok, err := auth.GenerateAccessToken("user-123", cfgForTest())
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_RefreshTokenRejected(t *testing.T) {
	app := buildProtectedApp()

	tok, _, err := auth.GenerateRefreshToken("user-123", cfgForTest())
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_GarbageToken(t *testing.T) {
	app := buildProtectedApp()

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.jwt")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
