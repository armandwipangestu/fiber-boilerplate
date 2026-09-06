package seed

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

// Seeder provisions a single piece of baseline data in an idempotent way.
type Seeder struct {
	// Name identifies the seeder for `db:seed <name>`.
	Name string
	// Run executes the seeding. It returns true when it changed data.
	Run func(ctx context.Context, db *sql.DB, cfg config.Config, logger *slog.Logger) (bool, error)
}

// registry holds every registered seeder, keyed by name.
var registry = map[string]Seeder{}

// Register makes a seeder available to the `db:seed` console command.
func Register(seeder Seeder) {
	if seeder.Name == "" {
		panic("seed: registering a seeder with an empty name")
	}
	if _, exists := registry[seeder.Name]; exists {
		panic(fmt.Sprintf("seed: seeder %q already registered", seeder.Name))
	}
	registry[seeder.Name] = seeder
}

// Registered returns the names of all registered seeders.
func Registered() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// RunAll executes every registered seeder and returns the names that ran.
func RunAll(ctx context.Context, db *sql.DB, cfg config.Config, logger *slog.Logger) ([]string, error) {
	var ran []string
	for _, seeder := range registry {
		changed, err := seeder.Run(ctx, db, cfg, logger)
		if err != nil {
			return ran, fmt.Errorf("seed %s: %w", seeder.Name, err)
		}
		if changed {
			ran = append(ran, seeder.Name)
		}
	}
	return ran, nil
}

// RunOne executes a single seeder by name.
func RunOne(ctx context.Context, db *sql.DB, cfg config.Config, logger *slog.Logger, name string) (bool, error) {
	seeder, ok := registry[name]
	if !ok {
		return false, fmt.Errorf("unknown seeder %q (registered: %v)", name, Registered())
	}
	return seeder.Run(ctx, db, cfg, logger)
}

func init() {
	Register(Seeder{
		Name: "admin",
		Run: func(ctx context.Context, db *sql.DB, cfg config.Config, logger *slog.Logger) (bool, error) {
			return DefaultAdmin(ctx, db, cfg, logger)
		},
	})
}
