package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

func TestRateLimit_Enforcement(t *testing.T) {
	cfg := config.Config{
		RateLimitEnabled:    true,
		RateLimitRequests:   2,
		RateLimitExpiration: time.Minute,
	}

	rl, err := NewRateLimitMiddleware(cfg)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(rl)
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	var firstLimit, firstReset string
	for i := 0; i < 4; i++ {
		req := requestPath("/")
		resp, err := app.Test(req)
		require.NoError(t, err)
		if i == 0 {
			firstLimit = resp.Header.Get("X-RateLimit-Limit")
			firstReset = resp.Header.Get("X-RateLimit-Reset")
		}
		if i < 2 {
			assert.Equal(t, http.StatusOK, resp.StatusCode, "requests %d should pass", i)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "request %d should be limited", i)
			assert.NotEmpty(t, resp.Header.Get("Retry-After"), "Retry-After present on 429")
		}
		resp.Body.Close()
	}

	assert.Equal(t, "2", firstLimit)
	assert.NotEmpty(t, firstReset)
}

func TestRateLimit_Disabled(t *testing.T) {
	cfg := config.Config{RateLimitEnabled: false}
	rl, err := NewRateLimitMiddleware(cfg)
	require.NoError(t, err)
	assert.Nil(t, rl, "disabled limiter returns nil handler")
}
