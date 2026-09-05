package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

// NewAuthMiddleware guards routes with a Bearer access token. The user id is
// stored in Fiber locals and read via pkg.GetUserID.
func NewAuthMiddleware(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return pkg.UnauthorizedResponse(c, "missing authorization header")
		}

		token, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || token == "" {
			return pkg.UnauthorizedResponse(c, "invalid authorization header format")
		}

		claims, err := auth.ValidateToken(token, cfg)
		if err != nil {
			return pkg.UnauthorizedResponse(c, "invalid or expired token")
		}
		if claims.Type != auth.TokenTypeAccess {
			return pkg.UnauthorizedResponse(c, "a refresh token cannot be used here")
		}

		pkg.SetUserID(c, claims.Subject)

		return c.Next()
	}
}
