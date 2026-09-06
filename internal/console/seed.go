package console

import (
	"context"
	"log/slog"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
	"github.com/armandwipangestu/fiber-boilerplate/internal/seed"
)

// Seed runs every registered seeder, or a single one when name is non-empty.
// The command fails fast when the migrations have not been applied yet.
func Seed(cfg config.Config, logger *slog.Logger, name string) error {
	db, err := database.NewDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	if name != "" {
		changed, err := seed.RunOne(ctx, db, cfg, logger, name)
		if err != nil {
			return err
		}
		if changed {
			printf("seed %s: done\n", name)
		} else if name == "admin" && cfg.DefaultAdminEmail == "" {
			printf("seed %s: nothing to do (set DEFAULT_ADMIN_EMAIL/DEFAULT_ADMIN_PASSWORD)\n", name)
		} else {
			printf("seed %s: nothing to do\n", name)
		}
		return nil
	}

	ran, err := seed.RunAll(ctx, db, cfg, logger)
	if err != nil {
		return err
	}
	if len(ran) == 0 {
		printl("seeded: none (nothing configured)")
		return nil
	}
	printf("seeded: %v\n", ran)
	return nil
}
