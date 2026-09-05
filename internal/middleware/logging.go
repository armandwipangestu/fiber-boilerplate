package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

// NewLoggingMiddleware emits structured start + complete entries per request,
// including latency, status, client IP, and the correlated request id.
func NewLoggingMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		method := c.Method()
		path := c.Path()
		query := string(c.Request().URI().QueryString())
		requestID := pkg.GetRequestID(c)

		logger.Debug("request started",
			"method", method,
			"path", path,
			"query", query,
			"client_ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
			"request_id", requestID,
		)

		err := c.Next()

		status := c.Response().StatusCode()
		latency := time.Since(start)

		logger.Info("request completed",
			"method", method,
			"path", path,
			"query", query,
			"status", status,
			"latency_ms", float64(latency.Microseconds())/1000.0,
			"bytes_out", c.Response().Header.ContentLength(),
			"error", errToString(err),
			"request_id", requestID,
		)

		return err
	}
}

func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
