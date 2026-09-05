// Package app is the composition root: it builds all services, wires them into
// the Fiber application, and owns their lifecycle. main and the E2E suite both
// use it.
package app

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/cache"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
	"github.com/armandwipangestu/fiber-boilerplate/internal/health"
	"github.com/armandwipangestu/fiber-boilerplate/internal/middleware"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
	"github.com/armandwipangestu/fiber-boilerplate/internal/rbac"
	"github.com/armandwipangestu/fiber-boilerplate/internal/server"
	"github.com/armandwipangestu/fiber-boilerplate/internal/storage"
	"github.com/armandwipangestu/fiber-boilerplate/internal/tracing"
	"github.com/armandwipangestu/fiber-boilerplate/internal/user"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Resources holds the running application and its owned dependencies.
type Resources struct {
	App    *fiber.App
	DB     *sql.DB
	Redis  *redis.Client
	Cache  cache.Cache
	Tracer *sdktrace.TracerProvider
	Logger *slog.Logger
	cfg    config.Config
}

// Build constructs the Fiber application and every service it depends on.
func Build(logger *slog.Logger, cfg config.Config) (*Resources, error) {
	db, err := database.NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	var rdb *redis.Client
	if cfg.RedisURL != "" {
		opts, err2 := redis.ParseURL(cfg.RedisURL)
		if err2 != nil {
			db.Close()
			return nil, err2
		}
		rdb = redis.NewClient(opts)
	}

	cacheStore, err := cache.NewCache(context.Background(), cfg.RedisURL, logger)
	if err != nil {
		db.Close()
		if rdb != nil {
			rdb.Close()
		}
		return nil, err
	}

	var tracerProvider *sdktrace.TracerProvider
	if cfg.OTELEnabled {
		tracerProvider, err = tracing.Init(cfg.AppName, cfg.OTELEndpoint, cfg.OTELSampleRate)
		if err != nil {
			logger.Error("failed to initialize tracing, continuing without it", "error", err)
			tracerProvider = nil
		}
	}

	healthHandler := health.NewHandler(db, rdb)
	store := storage.New(cfg, logger)

	userRepo := user.NewPostgresRepository(db)
	validator := pkg.NewValidator()
	rbacSvc := rbac.NewService(db, rbac.NewCache(cacheStore))
	rbacHandler := rbac.NewHandler(rbacSvc, validator)
	userSvc := user.NewServiceWithStorage(userRepo, rbacSvc, store)
	userHandler := user.NewHandler(userSvc, validator)

	sessionRepo := auth.NewPostgresRefreshTokenRepository(db)
	authSvc := auth.NewService(userRepo, sessionRepo, cfg)
	authHandler := auth.NewHandler(authSvc, validator)

	app := server.New(cfg, server.Dependencies{
		UserHandler:    userHandler,
		AuthHandler:    authHandler,
		HealthHandler:  healthHandler,
		RBACHandler:    rbacHandler,
		AuthMiddleware: middleware.NewAuthMiddleware(cfg),
		RBACService:    rbacSvc,
		Logger:         logger,
	})

	return &Resources{
		App:    app,
		DB:     db,
		Redis:  rdb,
		Cache:  cacheStore,
		Tracer: tracerProvider,
		Logger: logger,
		cfg:    cfg,
	}, nil
}

// Shutdown drains in-flight requests and closes resources in a fixed order:
// HTTP server → tracing → metrics (no-op) → cache → redis → database.
func (r *Resources) Shutdown(ctx context.Context) error {
	timeout := r.cfg.ShutdownTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	logger := r.Logger

	logger.Info("shutting down http server", "timeout", timeout.String())
	if err := r.App.ShutdownWithTimeout(timeout); err != nil {
		logger.Warn("http server shutdown did not complete cleanly", "error", err)
	}

	if r.Tracer != nil {
		logger.Info("shutting down tracing")
		if err := tracing.Shutdown(ctx, r.Tracer); err != nil {
			logger.Warn("tracing shutdown error", "error", err)
		}
	}

	if r.Cache != nil {
		if err := r.Cache.Close(); err != nil {
			logger.Warn("cache close error", "error", err)
		}
	}
	if r.Redis != nil {
		if err := r.Redis.Close(); err != nil {
			logger.Warn("redis close error", "error", err)
		}
	}
	if r.DB != nil {
		if err := r.DB.Close(); err != nil {
			logger.Warn("database close error", "error", err)
		}
	}
	return nil
}
