package server

import (
	"log/slog"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/health"
	"github.com/armandwipangestu/fiber-boilerplate/internal/metrics"
	"github.com/armandwipangestu/fiber-boilerplate/internal/middleware"
	"github.com/armandwipangestu/fiber-boilerplate/internal/rbac"
	"github.com/armandwipangestu/fiber-boilerplate/internal/user"
)

// Dependencies injected into the server by the composition root (main).
type Dependencies struct {
	UserHandler    *user.Handler
	AuthHandler    *auth.Handler
	HealthHandler  *health.Handler
	RBACHandler    *rbac.Handler
	AuthMiddleware fiber.Handler
	RBACService    *rbac.Service
	Logger         *slog.Logger
}

// New builds and configures the Fiber application with core middleware.
func New(cfg config.Config, deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:           cfg.AppName,
		EnablePrintRoutes: false,
		BodyLimit:         10 * 1024 * 1024, // 10MB
	})

	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	app.Use(middleware.RequestID())
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.NewCORSMiddleware(cfg))
	app.Use(middleware.NewTracingMiddleware())
	app.Use(middleware.NewLoggingMiddleware(logger))
	app.Use(middleware.NewMetricsMiddleware())

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
	if deps.HealthHandler != nil {
		app.Get("/health", deps.HealthHandler.Health)
		app.Get("/health/live", deps.HealthHandler.Live)
		app.Get("/health/ready", deps.HealthHandler.Ready)
	}

	// Prometheus scrape endpoint
	app.Get("/metrics", adaptor.HTTPHandler(metrics.Handler()))

	// Local fallback storage: files stored under STORAGE_PATH when S3 is not
	// configured are served from /uploads.
	storageRoot := cfg.StoragePath
	if storageRoot != "" {
		app.Static("/uploads", storageRoot)
	}

	// Swagger UI (development only).
	if cfg.SwaggerEnabled && cfg.AppEnv == "development" {
		swaggerCfg := swagger.Config{
			BasePath: "/",
			FilePath: "./docs/swagger/swagger.json",
			Path:     "swagger",
			Title:    cfg.AppName,
		}
		app.Use(swagger.New(swaggerCfg))
	}

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

	if deps.RBACHandler != nil {
		deps.RBACHandler.RegisterRoutes(v1, rbac.RouteOptions{
			Auth:                     deps.AuthMiddleware,
			RequireRolesView:         middleware.RequirePermission(deps.RBACService, "roles.view"),
			RequireRolesManage:       middleware.RequirePermission(deps.RBACService, "roles.manage"),
			RequirePermissionsView:   middleware.RequirePermission(deps.RBACService, "permissions.view"),
			RequirePermissionsManage: middleware.RequirePermission(deps.RBACService, "permissions.manage"),
		})
	}

	return app
}
