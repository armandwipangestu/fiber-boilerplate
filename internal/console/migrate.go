package console

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
)

// MigrateUp applies all pending migrations.
func MigrateUp(cfg config.Config) error {
	return database.RunMigrations(cfg, "up")
}

// MigrateDown reverts the most recent migration.
func MigrateDown(cfg config.Config) error {
	return database.RunMigrations(cfg, "down")
}

// MigrateVersion prints the current migration version.
func MigrateVersion(cfg config.Config) error {
	return database.RunMigrations(cfg, "version")
}

// MigrateStatus prints every known migration and whether it is applied,
// current or pending.
func MigrateStatus(cfg config.Config) error {
	statuses, current, dirty, err := database.MigrationList(cfg)
	if err != nil {
		return err
	}
	if len(statuses) == 0 {
		printl("no migrations found")
		return nil
	}

	for _, s := range statuses {
		name := fmt.Sprintf("%06d_%s", s.Version, s.Identifier)
		printf("%-8s  %s\n", label(s.State), name)
	}

	if dirty {
		printf("\ncurrent version: %d (dirty=%v)\n", current, dirty)
		printl("warning: the database is in a dirty state; resolve it before migrating")
	} else if current == 0 {
		printl("\ncurrent version: none")
	} else {
		printf("\ncurrent version: %d (dirty=%v)\n", current, dirty)
	}
	return nil
}

// MigrateReset reverts every migration (down) and re-applies them (up). It
// requires explicit confirmation unless yes is set.
func MigrateReset(cfg config.Config, yes bool) error {
	if !yes {
		if !confirm("This reverts and re-applies all migrations. Continue? [y/N]: ") {
			return errAborted
		}
	}
	if err := database.ResetMigrations(cfg); err != nil {
		return err
	}
	printl("migrations reset (down + up)")
	return nil
}

func label(state database.MigrationState) string {
	switch state {
	case database.MigrationApplied:
		return "applied"
	case database.MigrationCurrent:
		return "current"
	default:
		return "pending"
	}
}

var migrationPrefix = regexp.MustCompile(`^(\d+)`)

// nextMigrationNumber returns the next six-digit migration version based on
// the highest numeric prefix found in the migrations directory.
func nextMigrationNumber(root string) int {
	dir := filepath.Join(root, "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 1
	}
	highest := 0
	for _, e := range entries {
		m := migrationPrefix.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		if n, err := strconv.Atoi(m[1]); err == nil && n > highest {
			highest = n
		}
	}
	return highest + 1
}

// sortedMigrations retrieves migration status for use by tests.
func migrationNumbers(root string) []int {
	dir := filepath.Join(root, "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var nums []int
	for _, e := range entries {
		m := migrationPrefix.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err == nil {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	return nums
}
