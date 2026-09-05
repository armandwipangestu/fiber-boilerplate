package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

// NewCORSMiddleware configures cross-origin access from cfg.CORSAllowedOrigins.
func NewCORSMiddleware(cfg config.Config) fiber.Handler {
	origins := cfg.CORSAllowedOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	return cors.New(cors.Config{
		AllowOrigins:     joinOrigins(origins),
		AllowMethods:     "GET,POST,PATCH,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,If-None-Match",
		ExposeHeaders:    "X-Request-Id,Content-Length,ETag,X-RateLimit-Limit,X-RateLimit-Remaining,X-RateLimit-Reset",
		AllowCredentials: !containsWildcard(origins), // wildcard + credentials is insecure
		MaxAge:           86400,
	})
}

// containsWildcard reports whether any configured origin is a wildcard.
func containsWildcard(origins []string) bool {
	for _, o := range origins {
		if o == "*" || o == "null" {
			return true
		}
	}
	return false
}

// joinOrigins converts the config slice into Fiber's comma-separated format.
func joinOrigins(origins []string) string {
	out := ""
	for i, o := range origins {
		if i > 0 {
			out += ","
		}
		out += o
	}
	return out
}
