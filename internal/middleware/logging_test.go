package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware_EmitsStartAndComplete(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	app := fiber.New()
	app.Use(NewLoggingMiddleware(logger))
	app.Get("/ok", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	resp, err := app.Test(mustRequest(t, "/ok"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	out := buf.String()
	assert.Contains(t, out, "request started")
	assert.Contains(t, out, `msg="request completed"`)
	assert.Contains(t, out, "status=200")
	assert.Contains(t, out, "latency_ms=")
	assert.Contains(t, out, "method=GET")
	assert.Contains(t, out, "path=/ok")
}

func mustRequest(t *testing.T, path string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, path, nil)
	require.NoError(t, err)
	return req
}
