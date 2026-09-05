package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
)

const requestIDKey = "request_id"

// RequestID generates a request id if one is not present in the
// X-Request-ID header, stores it in locals, and echoes it back.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(fiber.HeaderXRequestID)
		if id == "" {
			id = newRequestID()
		}
		c.Locals(requestIDKey, id)
		c.Set(fiber.HeaderXRequestID, id)
		return c.Next()
	}
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
