package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/armandwipangestu/fiber-boilerplate/internal/app"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/console"
	"github.com/armandwipangestu/fiber-boilerplate/internal/logging"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
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
	if err := newRootCmd().Execute(); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

// newRootCmd builds the cobra tree. Running the binary without a subcommand
// starts the API server; the artisan subcommands (migrate, make:*, db:*,
// config:check, route:list) are thin shells over the internal/console package.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "app",
		Short:         "fiber-boilerplate API server and artisan console",
		Long:          "Run the API server (no arguments) or artisan-style console commands: migrate, make:*, db:seed, db:fresh, config:check, route:list.",
		RunE:          runServer,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.Version = pkg.EffectiveVersion()
	root.SetVersionTemplate("{{.Version}}\n")

	root.AddCommand(
		versionCmd(),
		migrateCmd(),
		makeCmd("make:model", "Scaffold a domain model", console.MakeModel),
		makeCmd("make:dto", "Scaffold DTOs for a feature", console.MakeDTO),
		makeCmd("make:repository", "Scaffold a repository interface + Postgres implementation", console.MakeRepository),
		makeCmd("make:service", "Scaffold a service and its business errors", console.MakeService),
		makeCmd("make:handler", "Scaffold an HTTP handler with Swagger annotations", console.MakeHandler),
		makeCmd("make:feature", "Scaffold a complete feature (+ migration)", console.MakeFeature),
		dbSeedCmd(),
		dbFreshCmd(),
		configCheckCmd(),
		routeListCmd(),
	)
	return root
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the build version and exit",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), pkg.EffectiveVersion())
		},
	}
}

func migrateCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database migrations",
		Long:  "Manage database migrations: up, down, status, reset or version.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadCLI()
			if err != nil {
				return err
			}
			switch args[0] {
			case "up":
				return console.MigrateUp(*cfg)
			case "down":
				return console.MigrateDown(*cfg)
			case "status":
				return console.MigrateStatus(*cfg)
			case "reset":
				return console.MigrateReset(*cfg, yes)
			case "version":
				return console.MigrateVersion(*cfg)
			default:
				return fmt.Errorf("unknown migrate command %q (use up|down|status|reset|version)", args[0])
			}
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "skip the confirmation prompt (for 'reset')")
	return cmd
}

type makeFunc func(name string, force bool) error

func makeCmd(use, short string, run makeFunc) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   use + " <name>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(args[0], force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files")
	return cmd
}

func dbSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "db:seed [name]",
		Short: "Seed baseline data (no arg seeds everything)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadCLI()
			if err != nil {
				return err
			}
			logger := logging.NewLogger(*cfg)
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			return console.Seed(*cfg, logger, name)
		},
	}
}

func dbFreshCmd() *cobra.Command {
	var yes, noSeed bool
	cmd := &cobra.Command{
		Use:   "db:fresh",
		Short: "Drop the schema, re-apply migrations and seed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadCLI()
			if err != nil {
				return err
			}
			logger := logging.NewLogger(*cfg)
			return console.FreshDB(*cfg, logger, console.FreshOptions{Yes: yes, NoSeed: noSeed})
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "skip the destructive confirmation prompt")
	cmd.Flags().BoolVar(&noSeed, "no-seed", false, "skip seeding after migrating")
	return cmd
}

func configCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config:check",
		Short: "Validate configuration and probe dependencies",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return console.ConfigCheck()
		},
	}
}

func routeListCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "route:list",
		Short: "List every registered HTTP route",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadCLI()
			if err != nil {
				return err
			}
			return console.RouteList(*cfg, jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "emit routes as JSON")
	return cmd
}

// runServer starts the API server. Dependencies are composed by app.Build and
// torn down on SIGINT/SIGTERM.
func runServer(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.NewLogger(*cfg)
	slog.SetDefault(logger)

	resources, err := app.Build(logger, *cfg)
	if err != nil {
		return err
	}

	addr := cfg.AppHost + ":" + formatPort(cfg.AppPort)
	logger.Info("server starting",
		"app", cfg.AppName,
		"env", cfg.AppEnv,
		"version", pkg.EffectiveVersion(),
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
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return resources.Shutdown(shutdownCtx)
	}
}

func formatPort(port int) string {
	if port == 0 {
		return "8080"
	}
	return fmt.Sprintf("%d", port)
}
