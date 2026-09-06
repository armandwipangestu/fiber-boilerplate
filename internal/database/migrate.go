package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationState classifies a single migration relative to the database.
type MigrationState string

const (
	// MigrationApplied means the migration version is below the current one.
	MigrationApplied MigrationState = "applied"
	// MigrationCurrent means the migration version equals the current one.
	MigrationCurrent MigrationState = "current"
	// MigrationPending means the migration version has not been applied yet.
	MigrationPending MigrationState = "pending"
)

// MigrationStatus describes the state of a known migration.
type MigrationStatus struct {
	Version    uint
	Identifier string
	State      MigrationState
}

// RunMigrations applies migration commands ("up" | "down" | "version").
// The source files are read from the ./migrations directory relative to the
// project root.
func RunMigrations(cfg config.Config, direction string) error {
	m, err := migrate.New("file://"+migrationsDir(), buildMigrateURL(cfg))
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
		return fmt.Errorf("unknown migration command %q (use up|down|version|status|reset)", direction)
	}
	return nil
}

// MigrationList returns every known migration with its state, together with
// the current version and dirty flag read from the database. A pristine
// database (no schema_migrations table) yields an empty current version.
func MigrationList(cfg config.Config) (statuses []MigrationStatus, current uint, dirty bool, err error) {
	src, err := (&file.File{}).Open("file://" + migrationsDir())
	if err != nil {
		return nil, 0, false, fmt.Errorf("migrations source: %w", err)
	}
	defer src.Close()

	m, err := migrate.New("file://"+migrationsDir(), buildMigrateURL(cfg))
	if err != nil {
		return nil, 0, false, fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	current, dirty, err = m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			current, dirty = 0, false
		} else {
			return nil, 0, false, fmt.Errorf("migrate version: %w", err)
		}
	}
	applied := current > 0

	version, ferr := src.First()
	for ferr == nil {
		ident := migrationIdentifier(src, version)
		state := MigrationPending
		switch {
		case applied && version == current:
			state = MigrationCurrent
		case applied && version < current:
			state = MigrationApplied
		}
		statuses = append(statuses, MigrationStatus{
			Version:    version,
			Identifier: ident,
			State:      state,
		})

		version, ferr = src.Next(version)
	}

	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Version < statuses[j].Version })
	return statuses, current, dirty, nil
}

// ResetMigrations reverts every migration (down) and re-applies them (up).
func ResetMigrations(cfg config.Config) error {
	m, err := migrate.New("file://"+migrationsDir(), buildMigrateURL(cfg))
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate reset (down): %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate reset (up): %w", err)
	}
	return nil
}

func migrationIdentifier(src source.Driver, version uint) string {
	rc, ident, err := src.ReadUp(version)
	if err != nil {
		return ""
	}
	rc.Close()
	return ident
}

// migrationsDir locates the migrations folder by walking up from the caller's
// file (works for both cmd/app/main.go and tests) until a directory named
// migrations is found.
func migrationsDir() string {
	_, file, _, ok := runtime.Caller(0)
	dir := filepath.Dir(file)
	for i := 0; i < 8 && ok; i++ {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// Fall back to the classic relative path from the current directory.
	return "migrations"
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
