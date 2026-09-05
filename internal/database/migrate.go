package database

import (
	"errors"
	"fmt"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies migration commands ("up" | "down" | "version").
// The source files are read from the ./migrations directory.
func RunMigrations(cfg config.Config, direction string) error {
	m, err := migrate.New("file://migrations", buildMigrateURL(cfg))
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate up: %w", err)
		}
		fmt.Println("migrations applied (up)")
	case "down":
		if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate down: %w", err)
		}
		fmt.Println("migration reverted (down 1 step)")
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			return fmt.Errorf("migrate version: %w", err)
		}
		fmt.Printf("current version: %d (dirty=%v)\n", v, dirty)
	default:
		return fmt.Errorf("unknown migration command %q (use up|down|version)", direction)
	}
	return nil
}

// buildMigrateURL adapts the configured URL for golang-migrate's pgx driver.
func buildMigrateURL(cfg config.Config) string {
	switch cfg.DatabaseDriver {
	case "postgres", "pgx":
		return "pgx5://" + trimScheme(cfg.DatabaseURL)
	default:
		return cfg.DatabaseURL
	}
}

func trimScheme(url string) string {
	schemes := []string{"postgres://", "postgresql://"}
	for _, s := range schemes {
		if len(url) > len(s) && url[:len(s)] == s {
			return url[len(s):]
		}
	}
	return url
}
