package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
	"github.com/armandwipangestu/fiber-boilerplate/internal/health"
	"github.com/armandwipangestu/fiber-boilerplate/internal/logging"
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

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// CLI subcommand: migrate
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: server migrate <up|down|version>")
			os.Exit(1)
		}
		if err := database.RunMigrations(*cfg, os.Args[2]); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	logger := logging.NewLogger(*cfg)
	slog.SetDefault(logger)

	db, err := database.NewDatabase(*cfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	var rdb *redis.Client
	if cfg.RedisURL != "" {
		opts, err2 := redis.ParseURL(cfg.RedisURL)
		if err2 != nil {
			logger.Error("failed to parse redis url", "error", err2)
			os.Exit(1)
		}
		rdb = redis.NewClient(opts)
	}

	var tracerProvider *sdktrace.TracerProvider
	if cfg.OTELEnabled {
		tracerProvider, err = tracing.Init(cfg.AppName, cfg.OTELEndpoint, cfg.OTELSampleRate)
		if err != nil {
			logger.Error("failed to initialize tracing, continuing without it", "error", err)
		} else {
			logger.Info("tracing enabled",
				"endpoint", cfg.OTELEndpoint,
				"sample_rate", cfg.OTELSampleRate,
			)
		}
	}

	healthHandler := health.NewHandler(db, rdb)

	store := storage.New(*cfg, logger)

	userRepo := user.NewPostgresRepository(db)
	rbacSvc := rbac.NewService(db, rbac.NewInMemoryCache())
	userSvc := user.NewServiceWithStorage(userRepo, rbacSvc, store)
	validator := pkg.NewValidator()
	userHandler := user.NewHandler(userSvc, validator)

	sessionRepo := auth.NewPostgresRefreshTokenRepository(db)
	authSvc := auth.NewService(userRepo, sessionRepo, *cfg)
	authHandler := auth.NewHandler(authSvc, validator)

	app := server.New(*cfg, server.Dependencies{
		UserHandler:    userHandler,
		AuthHandler:    authHandler,
		HealthHandler:  healthHandler,
		AuthMiddleware: middleware.NewAuthMiddleware(*cfg),
		RBACService:    rbacSvc,
		Logger:         logger,
	})

	addr := cfg.AppHost + ":" + formatPort(cfg.AppPort)
	logger.Info("server starting",
		"app", cfg.AppName,
		"env", cfg.AppEnv,
		"addr", addr,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listen(addr)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		if err := shutdown(*cfg, logger, app, tracerProvider, rdb, db); err != nil {
			logger.Error("shutdown error", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

// shutdown drains in-flight work and closes resources in a fixed order:
// HTTP server → tracing → metrics → redis → database.
func shutdown(cfg config.Config, logger *slog.Logger, app *fiber.App, tracerProvider *sdktrace.TracerProvider, rdb *redis.Client, db *sql.DB) error {
	timeout := cfg.ShutdownTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 1. HTTP server: stop accepting new connections, drain active requests.
	logger.Info("shutting down http server", "timeout", timeout.String())
	if err := app.ShutdownWithTimeout(timeout); err != nil {
		logger.Warn("http server shutdown did not complete cleanly", "error", err)
	}

	// 2. Tracing: flush any pending spans.
	if tracerProvider != nil {
		logger.Info("shutting down tracing")
		if err := tracing.Shutdown(ctx, tracerProvider); err != nil {
			logger.Warn("tracing shutdown error", "error", err)
		}
	}

	// 3. Metrics: the Prometheus registry is memory-only; nothing to flush.

	// 4. Redis.
	if rdb != nil {
		logger.Info("shutting down redis")
		if err := rdb.Close(); err != nil {
			logger.Warn("redis close error", "error", err)
		}
	}

	// 5. Database.
	logger.Info("shutting down database")
	if err := db.Close(); err != nil {
		logger.Warn("database close error", "error", err)
	}

	logger.Info("shutdown complete")
	return nil
}

func formatPort(port int) string {
	if port == 0 {
		return "8080"
	}
	return fmt.Sprintf("%d", port)
}
