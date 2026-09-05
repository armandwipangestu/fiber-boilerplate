package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
	"github.com/armandwipangestu/fiber-boilerplate/internal/logging"
	"github.com/armandwipangestu/fiber-boilerplate/internal/server"
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

	app := server.New(*cfg)

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
