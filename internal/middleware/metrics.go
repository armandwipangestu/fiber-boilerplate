package middleware

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/metrics"
)

var uuidSegments = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// NewMetricsMiddleware records request counts, latency, and the in-flight
// gauge, normalizing dynamic path segments (UUIDs) to a stable `:id` label so
// metric cardinality stays bounded.
func NewMetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		metrics.RequestsInFlight.Inc()

		start := time.Now()
		err := c.Next()

		metrics.RequestsInFlight.Dec()

		path := uuidSegments.ReplaceAllString(c.Path(), ":id")
		status := c.Response().StatusCode()
		if err != nil {
			var fe *fiber.Error
			if errors.As(err, &fe) {
				status = fe.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		metrics.RequestsTotal.WithLabelValues(c.Method(), path, strconv.Itoa(status)).Inc()
		metrics.RequestDuration.WithLabelValues(c.Method(), path).Observe(time.Since(start).Seconds())

		return err
	}
}
