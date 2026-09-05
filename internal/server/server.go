package server

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/middleware"
	"github.com/armandwipangestu/fiber-boilerplate/internal/rbac"
	"github.com/armandwipangestu/fiber-boilerplate/internal/user"
)

// Dependencies injected into the server by the composition root (main).
type Dependencies struct {
	UserHandler    *user.Handler
	AuthHandler    *auth.Handler
	AuthMiddleware fiber.Handler
	RBACService    *rbac.Service
}

// New builds and configures the Fiber application with core middleware.
func New(cfg config.Config, deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:           cfg.AppName,
		EnablePrintRoutes: false,
		BodyLimit:         10 * 1024 * 1024, // 10MB
	})

	app.Use(middleware.RequestID())

	if rateLimit, err := middleware.NewRateLimitMiddleware(cfg); err != nil {
		panic(err) // misconfiguration, fail fast
	} else if rateLimit != nil {
		app.Use(rateLimit)
	}

	app.Use(middleware.NewThrottleMiddleware(middleware.ThrottleConfig{
		MaxConcurrent: 1000,
		Timeout:       5 * time.Second,
	}))

	// Health & liveness
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(map[string]any{"message": "pong"})
	})

	api := app.Group("/api")
	v1 := api.Group("/v1")

	if deps.UserHandler != nil {
		deps.UserHandler.RegisterRoutes(v1, user.RouteOptions{
			Auth:          deps.AuthMiddleware,
			RequireCreate: middleware.RequirePermission(deps.RBACService, "users.create"),
			RequireView:   middleware.RequirePermission(deps.RBACService, "users.view"),
			RequireUpdate: middleware.RequirePermission(deps.RBACService, "users.update"),
			RequireDelete: middleware.RequirePermission(deps.RBACService, "users.delete"),
		})
	}
	if deps.AuthHandler != nil {
		deps.AuthHandler.RegisterRoutes(v1)
	}

	return app
}
