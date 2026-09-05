package pkg

import "github.com/gofiber/fiber/v2"

const requestIDCtxKey = "request_id"
const userIDCtxKey = "user_id"
const traceIDCtxKey = "trace_id"
const spanIDCtxKey = "span_id"

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

// GetTraceID returns the OpenTelemetry trace id stored by the tracing
// middleware. Empty when tracing is disabled.
func GetTraceID(c *fiber.Ctx) string {
	if v := c.Locals(traceIDCtxKey); v != nil {
		return v.(string)
	}
	return ""
}

// SetTraceID stores the OpenTelemetry trace id into the Fiber locals.
func SetTraceID(c *fiber.Ctx, id string) {
	c.Locals(traceIDCtxKey, id)
}

// GetSpanID returns the OpenTelemetry span id stored by the tracing
// middleware. Empty when tracing is disabled.
func GetSpanID(c *fiber.Ctx) string {
	if v := c.Locals(spanIDCtxKey); v != nil {
		return v.(string)
	}
	return ""
}

// SetSpanID stores the OpenTelemetry span id into the Fiber locals.
func SetSpanID(c *fiber.Ctx, id string) {
	c.Locals(spanIDCtxKey, id)
}
