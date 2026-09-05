package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/armandwipangestu/fiber-boilerplate/internal/app"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
	"github.com/armandwipangestu/fiber-boilerplate/internal/logging"
)

// @title Fiber Boilerplate API
// @version 1.0.0
// @description A production-grade Fiber (Go) API boilerplate with auth, RBAC, storage, and observability.
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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

	resources, err := app.Build(logger, *cfg)
	if err != nil {
		logger.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

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
		errCh <- resources.App.Listen(addr)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := resources.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown error", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func formatPort(port int) string {
	if port == 0 {
		return "8080"
	}
	return fmt.Sprintf("%d", port)
}
