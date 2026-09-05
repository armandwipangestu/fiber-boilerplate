package pkg

import "github.com/gofiber/fiber/v2"

const requestIDCtxKey = "request_id"
const userIDCtxKey = "user_id"

// GetRequestID returns the request id stored by the RequestID middleware.
func GetRequestID(c *fiber.Ctx) string {
	if v := c.Locals(requestIDCtxKey); v != nil {
		return v.(string)
	}
	return ""
}

// SetRequestID stores a request id into the Fiber locals.
func SetRequestID(c *fiber.Ctx, id string) {
	c.Locals(requestIDCtxKey, id)
}

// GetUserID returns the authenticated user id stored by the auth middleware.
func GetUserID(c *fiber.Ctx) string {
	if v := c.Locals(userIDCtxKey); v != nil {
		return v.(string)
	}
	return ""
}

// SetUserID stores the authenticated user id into the Fiber locals.
func SetUserID(c *fiber.Ctx, id string) {
	c.Locals(userIDCtxKey, id)
}
