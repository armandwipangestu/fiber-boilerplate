package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

// ThrottleConfig controls the concurrency limiter.
type ThrottleConfig struct {
	MaxConcurrent int
	Timeout       time.Duration
}

// NewThrottleMiddleware caps concurrent in-flight requests using a channel
// semaphore. When the pool is exhausted for longer than Timeout, the request
// is rejected with 503 + Retry-After.
func NewThrottleMiddleware(cfg ThrottleConfig) fiber.Handler {
	sem := make(chan struct{}, cfg.MaxConcurrent)

	return func(c *fiber.Ctx) error {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			return c.Next()
		default:
			// Pool full: wait a short grace period before rejecting.
			timer := time.NewTimer(cfg.Timeout)
			defer timer.Stop()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				return c.Next()
			case <-timer.C:
				c.Set("Retry-After", strconv.Itoa(int(cfg.Timeout.Seconds())+1))
				return pkg.ServiceUnavailableResponse(c, "server is overloaded, please retry shortly")
			case <-c.Context().Done():
				return c.Context().Err()
			}
		}
	}
}
