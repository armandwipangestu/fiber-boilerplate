package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/armandwipangestu/fiber-boilerplate/internal/auth"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
	"github.com/armandwipangestu/fiber-boilerplate/internal/health"
	"github.com/armandwipangestu/fiber-boilerplate/internal/logging"
	"github.com/armandwipangestu/fiber-boilerplate/internal/middleware"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
	"github.com/armandwipangestu/fiber-boilerplate/internal/rbac"
	"github.com/armandwipangestu/fiber-boilerplate/internal/server"
	"github.com/armandwipangestu/fiber-boilerplate/internal/user"

	"github.com/redis/go-redis/v9"
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
	defer db.Close()

	var rdb *redis.Client
	if cfg.RedisURL != "" {
		opts, err2 := redis.ParseURL(cfg.RedisURL)
		if err2 != nil {
			logger.Error("failed to parse redis url", "error", err2)
			os.Exit(1)
		}
		rdb = redis.NewClient(opts)
		defer rdb.Close()
	}

	healthHandler := health.NewHandler(db, rdb)

	userRepo := user.NewPostgresRepository(db)
	rbacSvc := rbac.NewService(db, rbac.NewInMemoryCache())
	userSvc := user.NewService(userRepo, rbacSvc)
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

	if err := app.Listen(addr); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func formatPort(port int) string {
	if port == 0 {
		return "8080"
	}
	return fmt.Sprintf("%d", port)
}
