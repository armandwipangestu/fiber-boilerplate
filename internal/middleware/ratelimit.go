package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

// NewRateLimitMiddleware enforces a per-client request budget and adds the
// standard X-RateLimit-* headers. An in-memory store is used until the Redis
// store lands in Phase 18.
func NewRateLimitMiddleware(cfg config.Config) (fiber.Handler, error) {
	if !cfg.RateLimitEnabled {
		return nil, nil
	}

	store := memory.NewStore()
	rate := limiter.Rate{
		Period: cfg.RateLimitExpiration,
		Limit:  int64(cfg.RateLimitRequests),
	}
	instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))

	return func(c *fiber.Ctx) error {
		key := c.IP()
		if userID := pkg.GetUserID(c); userID != "" {
			key = userID
		}

		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		context, err := instance.Get(ctx, key)
		if err != nil {
			return pkg.Error(c, err)
		}

		c.Set("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			retryAfter := time.Until(time.Unix(context.Reset, 0))
			if retryAfter < 0 {
				retryAfter = 0
			}
			c.Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			return pkg.RateLimitedResponse(c, "too many requests, please slow down")
		}

		return c.Next()
	}, nil
}
