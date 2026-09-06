package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/console"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
)

// consoleCfg points the console commands at an isolated, throwaway database
// created next to the e2e database so destructive commands (reset, fresh)
// never touch shared data.
func consoleCfg(t *testing.T, url string) config.Config {
	t.Helper()
	return config.Config{
		AppName:              "fiber-boilerplate-console-test",
		AppEnv:               "test",
		DatabaseDriver:       "postgres",
		DatabaseURL:          url,
		JWTSecret:            "console-test-secret",
		DefaultAdminEmail:    "admin@console-test",
		DefaultAdminPassword: "Console-Test!",
	}
}

// newConsoleDB creates a scratch database and returns its config plus a
// cleanup that drops the database again.
func newConsoleDB(t *testing.T) config.Config {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5433/fiber_boilerplate_e2e?sslmode=disable"
	}

	// Derive a maintenance URL pointing at the same server's "postgres" db.
	base := "fiber_boilerplate_e2e"
	maintURL := strings.Replace(url, "/"+base, "/postgres", 1)
	name := fmt.Sprintf("fiber_boilerplate_console_%d", time.Now().UnixNano())
	if !validDBIdentifier(name) {
		t.Fatalf("invalid database name %q", name)
	}

	mdb, err := database.OpenConn(maintURL)
	if err != nil {
		t.Fatalf("connect maintenance db: %v", err)
	}
	if _, err := mdb.Exec(`CREATE DATABASE ` + name); err != nil {
		t.Fatalf("create console db: %v", err)
	}
	t.Cleanup(func() {
		// Terminate lingering connections before dropping.
		_, _ = mdb.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = ` + "'" + name + "'")
		_, _ = mdb.Exec(`DROP DATABASE ` + name)
		_ = mdb.Close()
	})

	scratchURL := replaceDBName(url, name)
	return consoleCfg(t, scratchURL)
}

func replaceDBName(url, name string) string {
	idx := strings.LastIndexByte(url, '/')
	base := url
	if idx >= 0 {
		base = url[:idx]
	}
	return base + "/" + name
}

func validDBIdentifier(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return s != ""
}

// capture runs fn while fd 1 (stdout) is redirected into a pipe, so both the
// console package's cached os.Stdout handle and direct os.Stdout writes are
// captured. This is safe because the e2e suite runs sequentially.
func capture(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	saved, err := syscall.Dup(1)
	if err != nil {
		return "", err
	}
	r, w, err := os.Pipe()
	if err != nil {
		syscall.Close(saved)
		return "", err
	}
	if err := syscall.Dup2(int(w.Fd()), 1); err != nil {
		syscall.Close(saved)
		return "", err
	}

	runErr := fn()

	_ = syscall.Dup2(saved, 1)
	syscall.Close(saved)
	if err := w.Close(); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), runErr
}

func TestConsoleMigrationLifecycle(t *testing.T) {
	cfg := newConsoleDB(t)

	// A pristine database reports no current version.
	out, err := capture(t, func() error { return console.MigrateStatus(cfg) })
	if err != nil {
		t.Fatalf("status on pristine db: %v", err)
	}
	if !strings.Contains(out, "current version: none") {
		t.Fatalf("expected pristine status output, got:\n%s", out)
	}

	if err := console.MigrateUp(cfg); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	out, err = capture(t, func() error { return console.MigrateStatus(cfg) })
	if err != nil {
		t.Fatalf("status after up: %v", err)
	}
	if strings.Contains(out, "pending") {
		t.Fatalf("expected no pending migrations after up, got:\n%s", out)
	}
	if !strings.Contains(out, "current version: 9") {
		t.Fatalf("expected current version 9, got:\n%s", out)
	}

	// Down reverts exactly one step.
	if err := console.MigrateDown(cfg); err != nil {
		t.Fatalf("migrate down: %v", err)
	}
	var dirty bool
	var version int
	if err := consoleMigrationVersion(cfg, &version, &dirty); err != nil {
		t.Fatalf("check version after down: %v", err)
	}
	if version != 8 || dirty {
		t.Fatalf("expected version 8 after one down, got %d (dirty=%v)", version, dirty)
	}

	// Reset brings the database back to the latest migration.
	if err := console.MigrateReset(cfg, true); err != nil {
		t.Fatalf("migrate reset: %v", err)
	}
	if err := consoleMigrationVersion(cfg, &version, &dirty); err != nil {
		t.Fatalf("check version after reset: %v", err)
	}
	if version != 9 || dirty {
		t.Fatalf("expected version 9 after reset, got %d (dirty=%v)", version, dirty)
	}
}

func TestConsoleSeedAndFresh(t *testing.T) {
	cfg := newConsoleDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := console.MigrateUp(cfg); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	out, err := capture(t, func() error { return console.Seed(cfg, logger, "") })
	if err != nil {
		t.Fatalf("db:seed: %v", err)
	}
	if !strings.Contains(out, "seeded: [admin]") {
		t.Fatalf("expected seeded admin, got:\n%s", out)
	}

	// db:fresh drops the schema, re-migrates and seeds again.
	if err := console.FreshDB(cfg, logger, console.FreshOptions{Yes: true}); err != nil {
		t.Fatalf("db:fresh: %v", err)
	}

	// Seeded admin survives the fresh cycle.
	db, err := database.NewDatabase(cfg)
	if err != nil {
		t.Fatalf("open console db: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = $1`, "admin@console-test").Scan(&count); err != nil {
		t.Fatalf("query fresh admin: %v", err)
	} else if count != 1 {
		t.Fatalf("expected 1 admin after fresh, got %d", count)
	}

	// Migrations were re-applied.
	out, err = capture(t, func() error { return console.MigrateStatus(cfg) })
	if err != nil {
		t.Fatalf("status after fresh: %v", err)
	}
	if !strings.Contains(out, "current version: 9") {
		t.Fatalf("expected version 9 after fresh, got:\n%s", out)
	}
}

func TestConfigCheck(t *testing.T) {
	cfg := newConsoleDB(t)

	// ConfigCheck reads environment variables, so point them at the scratch DB.
	t.Setenv("DATABASE_URL", cfg.DatabaseURL)
	t.Setenv("JWT_SECRET", cfg.JWTSecret)
	t.Setenv("DEFAULT_ADMIN_EMAIL", cfg.DefaultAdminEmail)
	t.Setenv("DEFAULT_ADMIN_PASSWORD", cfg.DefaultAdminPassword)
	t.Setenv("REDIS_URL", "")

	out, err := capture(t, func() error { return console.ConfigCheck() })
	if err != nil {
		t.Fatalf("config:check: %v", err)
	}
	for _, want := range []string{"database: ok", "check: ok", "JWT_SECRET", "******"} {
		if !strings.Contains(out, want) {
			t.Errorf("config:check output missing %q; got:\n%s", want, out)
		}
	}
}

func TestRouteListAsCommand(t *testing.T) {
	cfg := newConsoleDB(t)

	var rows []struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}
	out, err := capture(t, func() error {
		old := os.Getenv("DATABASE_URL")
		os.Setenv("DATABASE_URL", cfg.DatabaseURL)
		defer os.Setenv("DATABASE_URL", old)
		return console.RouteList(cfg, true)
	})
	if err != nil {
		t.Fatalf("route:list json: %v", err)
	}
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("route:list json decode: %v (%s)", err, out)
	}
	got := map[string]bool{}
	for _, r := range rows {
		got[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{"GET /ping", "POST /api/v1/users/", "GET /api/v1/users/:id", "POST /api/v1/auth/login"} {
		if !got[want] {
			t.Errorf("route list missing %s", want)
		}
	}
}

// consoleMigrationVersion reads the golang-migrate version directly.
func consoleMigrationVersion(cfg config.Config, version *int, dirty *bool) error {
	db, err := database.NewDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.QueryRow(`SELECT version, dirty FROM schema_migrations WHERE version IS NOT NULL`).Scan(version, dirty)
}
