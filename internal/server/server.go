package server

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/docs"
	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/health"
	"github.com/armandwipangestu/fiber-boilerplate/internal/metrics"
	"github.com/armandwipangestu/fiber-boilerplate/internal/middleware"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
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

	// Build/release version (stamped by the release pipeline, falls back to
	// the VCS revision for dev builds).
	app.Get("/version", versionHandler(cfg))
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

	// Swagger UI (development only). The spec is embedded in the binary, so it
	// works regardless of the working directory (go run, dev shell, download).
	if cfg.SwaggerEnabled && cfg.AppEnv == "development" {
		spec, err := docs.SwaggerFS.ReadFile("swagger/swagger.json")
		if err != nil {
			logger.Error("embedded swagger spec missing", "error", err)
		} else {
			swaggerCfg := swagger.Config{
				BasePath:    "/",
				FilePath:    "docs/swagger/swagger.json",
				FileContent: spec,
				Path:        "swagger",
				Title:       cfg.AppName,
			}
			app.Use(relaxCSPForSwagger())
			app.Use(swagger.New(swaggerCfg))
		}
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

// versionHandler reports the running build's app name and version.
// @Summary Show build version
// @Description Returns the app name and the build version (stamped from the
// @Description release pipeline, falls back to a dev-<commit> identifier).
// @Tags System
// @Produce json
// @Success 200 {object} map[string]any
// @Router /version [get]
func versionHandler(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(map[string]any{"app": cfg.AppName, "version": pkg.EffectiveVersion()})
	}
}

// relaxCSPForSwagger narrows the hardened Content-Security-Policy to the
// Swagger UI page only: that page loads its assets from the unpkg CDN and
// ships inline <script>/<style>, both of which the API-wide "default-src
// 'self'" policy blocks. Every other route keeps the strict header.
func relaxCSPForSwagger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		p := c.Path()
		if p == "/swagger" || strings.HasPrefix(p, "/docs/swagger/") {
			c.Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' https://unpkg.com; "+
					"style-src 'self' 'unsafe-inline' https://unpkg.com; "+
					"img-src 'self' data: https://unpkg.com; "+
					"connect-src 'self'; "+
					"font-src 'self' data:; "+
					"frame-ancestors 'none'; base-uri 'self'")
		}
		return c.Next()
	}
}
