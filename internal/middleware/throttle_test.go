package middleware

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThrottle_UnderLimit(t *testing.T) {
	app := fiber.New()
	app.Use(NewThrottleMiddleware(ThrottleConfig{MaxConcurrent: 4, Timeout: 100 * time.Millisecond}))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	resp, err := app.Test(requestPath("/"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestThrottle_Overloaded(t *testing.T) {
	app := fiber.New()
	app.Use(NewThrottleMiddleware(ThrottleConfig{MaxConcurrent: 1, Timeout: 50 * time.Millisecond}))

	var wg sync.WaitGroup
	results := make(chan int, 5)

	// Slow handler to hold the single slot.
	app.Get("/slow", func(c *fiber.Ctx) error {
		time.Sleep(300 * time.Millisecond)
		return c.SendStatus(http.StatusOK)
	})
	// Fast handler with the same pool.
	app.Get("/fast", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	// Occupy the only slot.
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := app.Test(requestPath("/slow"))
		if err == nil {
			results <- resp.StatusCode
		}
	}()
	time.Sleep(20 * time.Millisecond)

	// This request must be rejected because the pool stays busy > Timeout.
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := app.Test(requestPath("/fast"))
		if err == nil {
			results <- resp.StatusCode
		}
	}()

	wg.Wait()
	close(results)

	codes := []int{}
	for c := range results {
		codes = append(codes, c)
	}
	assert.Contains(t, codes, http.StatusOK, "slow request should complete")
	assert.Contains(t, codes, http.StatusServiceUnavailable, "throttled request should be 503")
}

func requestPath(path string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	return req
}
