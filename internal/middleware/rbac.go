package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
	"github.com/armandwipangestu/fiber-boilerplate/internal/rbac"
)

// RequirePermission guards a route with a single required permission.
func RequirePermission(svc *rbac.Service, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := pkg.GetUserID(c)
		if userID == "" {
			return pkg.UnauthorizedResponse(c, "authentication required")
		}

		allowed, err := svc.HasPermission(c.Context(), userID, permission)
		if err != nil {
			return pkg.Error(c, err)
		}
		if !allowed {
			return pkg.ForbiddenResponse(c, "insufficient permissions")
		}
		return c.Next()
	}
}

// RequireAnyPermission passes when the user holds at least one permission.
func RequireAnyPermission(svc *rbac.Service, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := pkg.GetUserID(c)
		if userID == "" {
			return pkg.UnauthorizedResponse(c, "authentication required")
		}

		allowed, err := svc.HasAnyPermission(c.Context(), userID, permissions...)
		if err != nil {
			return pkg.Error(c, err)
		}
		if !allowed {
			return pkg.ForbiddenResponse(c, "insufficient permissions")
		}
		return c.Next()
	}
}

// RequireAllPermissions passes only when the user holds every permission.
func RequireAllPermissions(svc *rbac.Service, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := pkg.GetUserID(c)
		if userID == "" {
			return pkg.UnauthorizedResponse(c, "authentication required")
		}

		for _, perm := range permissions {
			allowed, err := svc.HasPermission(c.Context(), userID, perm)
			if err != nil {
				return pkg.Error(c, err)
			}
			if !allowed {
				return pkg.ForbiddenResponse(c, "insufficient permissions")
			}
		}
		return c.Next()
	}
}
