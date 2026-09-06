package console

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
	"github.com/armandwipangestu/fiber-boilerplate/internal/seed"
)

// FreshOptions tunes the db:fresh command.
type FreshOptions struct {
	Yes    bool // skip the destructive confirmation prompt
	NoSeed bool // skip seeding after re-applying migrations
}

// FreshDB drops the schema, re-applies every migration and seeds the database.
func FreshDB(cfg config.Config, logger *slog.Logger, opts FreshOptions) error {
	if !isPostgres(cfg) {
		return errors.New("db:fresh only supports the postgres driver")
	}
	if !opts.Yes {
		if !confirm("This drops the entire public schema. Continue? [y/N]: ") {
			return errAborted
		}
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(),
		`DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`); err != nil {
		return fmt.Errorf("drop schema: %w", err)
	}
	printl("dropped public schema")

	if err := database.RunMigrations(cfg, "up"); err != nil {
		return err
	}

	if !opts.NoSeed {
		ran, err := seed.RunAll(context.Background(), db, cfg, logger)
		if err != nil {
			return err
		}
		if len(ran) == 0 {
			printl("seeded: none (nothing configured)")
		} else {
			printf("seeded: %v\n", ran)
		}
	}

	printl("database is fresh")
	return nil
}

func isPostgres(cfg config.Config) bool {
	switch cfg.DatabaseDriver {
	case "postgres", "pgx":
		return true
	}
	return false
}
