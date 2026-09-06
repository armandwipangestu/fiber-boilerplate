# Phase 25 — Artisan-style CLI (`./app`)

> Scope: the binary doubles as a Laravel-style command runner. `migrate *`,
> `db:seed` / `db:fresh`, `config:check`, `route:list` and `make:*` codegen are
> available as `./app <command>`, while a bare `./app` still serves the API.
> The entrypoint moves from `cmd/server` to `cmd/app`.

## Deliverables

### 25.1 Binary rename `cmd/server` → `cmd/app`
- `git mv cmd/server cmd/app`; `Taskfile.yml` (`dev`/`build` → `go run ./cmd/app`,
  `-o bin/app`), `Dockerfile` (build `./cmd/app`, `ENTRYPOINT ["/app/app"]`),
  `.github/workflows/release.yml` (build `./cmd/app`; release artifact names
  unchanged) and `docs/*` updated. `swag init` now targets `cmd/app/main.go`.

### 25.2 CLI framework
- `github.com/spf13/cobra v1.10.2`. `cmd/app/main.go` builds a root command:
  no-arg = `runServer` (unchanged graceful-shutdown server); `--version`/`-v`
  and `version` print the stamped (or `dev-<commit>`) version.
- Business logic lives in a new `internal/console` package (cobra stays the
  shell; commands are plain funcs callable from tests).

### 25.3 Migrations (`internal/console/migrate.go`, `internal/database`)
- `migrate up|down|version` as before (`RunMigrations`).
- `migrate status`: per-file table (`applied`/`current`/`pending`, dirty flag)
  via new `database.MigrationList`; pristine DB reports `current version: none`.
- `migrate reset`: `m.Down()` then `m.Up()`, tolerating `ErrNoChange`; destructive
  commands (`down`, `reset`) ask for confirmation, `--yes` skips it.
- Config split: `config.LoadCLI()` requires only `DATABASE_URL` for DB-only
  commands; `config.Load()` (server, `config:check`) keeps validating `JWT_SECRET`.

### 25.4 Seeding (`internal/seed` registry, `internal/console/seed.go`)
- `Seeder{Name, Run}` registry with `Register`/`Registered`/`RunAll`/`RunOne`;
  the admin seeder is registered by `init`.
- `db:seed` (all) and `db:seed <name>` (one), printing `seeded: [admin]`.
- `db:fresh` (postgres): `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`
  → migrate up → seed unless `--no-seed`; guarded by `--yes`.

### 25.5 Diagnostics
- `config:check`: full `config.Load`, a settings table (secrets masked), live
  probes for Postgres and Redis, non-zero exit on problems.
- `route:list` (`--json`): new `app.BuildForRoutes` assembles an uncompiled
  Fiber app (nil DB, memory cache, no seed) so no connection is opened;
  routes are printed via tabwriter (HEAD/empty filtered, sorted) or indented JSON.

### 25.6 Codegen (`internal/console/make.go`, `templates/`)
- `make:model|dto|repository|service|handler|feature <name>` write gofmt'd,
  domain-pattern files (model → dto → repository → interface → service →
  handler → errors) into `internal/<feature>/`, mirroring the existing
  `auth`/`user` structure. `make:feature` also scaffolds the next-numbered
  migration (`000010_create_<table>.up/.down.sql`) and prints a wiring checklist
  (mount in `internal/server`, append `x.create/view/update/delete` permission
  rows to `migrations/000007_seed_rbac.up.sql`).
- `writeGenerated` runs `go/format.Source`; existing files are never
  overwritten without `--force`.

## Verify
- `internal/console` unit tests: name normalization, next migration number,
  route list (table + JSON), codegen file set + Go parses + overwrite guard.
- `tests/e2e/console_test.go`: throwaway Postgres per test — status on pristine
  DB, `up`→current version 9, `down`→8, `reset`→9; `db:seed` both orders;
  `db:fresh` keeps seeded admin; `config:check` ok; `route:list --json` decodes.
- Smoke: `go run ./cmd/app --version` → `dev`, `--help` lists all commands.