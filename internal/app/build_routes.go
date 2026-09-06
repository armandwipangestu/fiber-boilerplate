package app

import (
	"io"
	"log/slog"

	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/cache"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/health"
	"github.com/armandwipangestu/fiber-boilerplate/internal/middleware"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
	"github.com/armandwipangestu/fiber-boilerplate/internal/rbac"
	"github.com/armandwipangestu/fiber-boilerplate/internal/server"
	"github.com/armandwipangestu/fiber-boilerplate/internal/storage"
	"github.com/armandwipangestu/fiber-boilerplate/internal/user"

	"github.com/gofiber/fiber/v2"
)

// BuildForRoutes assembles the Fiber app for introspection-only purposes
// (e.g. the `route:list` console command). Every dependency is constructed
// without opening the database, a Redis connection, a tracer, or any seeders,
// so listing routes never touches infrastructure. All repositories and
// handlers dereference their *sql.DB lazily per request, which makes this
// safe.
func BuildForRoutes(cfg config.Config) (*fiber.App, error) {
	// Route listing should not pollute the terminal with logs about caches,
	// storage backends and so on.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	cacheStore := cache.NewMemoryCache()
	healthHandler := health.NewHandler(nil, nil)
	store := storage.New(cfg, logger)

	userRepo := user.NewPostgresRepository(nil)
	validator := pkg.NewValidator()
	rbacSvc := rbac.NewService(nil, rbac.NewCache(cacheStore))
	rbacHandler := rbac.NewHandler(rbacSvc, validator)
	userSvc := user.NewServiceWithStorage(userRepo, rbacSvc, store)
	userHandler := user.NewHandler(userSvc, validator)

	sessionRepo := auth.NewPostgresRefreshTokenRepository(nil)
	authSvc := auth.NewService(userRepo, sessionRepo, cfg)
	authHandler := auth.NewHandler(authSvc, validator)

	return server.New(cfg, server.Dependencies{
		UserHandler:    userHandler,
		AuthHandler:    authHandler,
		HealthHandler:  healthHandler,
		RBACHandler:    rbacHandler,
		AuthMiddleware: middleware.NewAuthMiddleware(cfg),
		RBACService:    rbacSvc,
		Logger:         logger,
	}), nil
}
