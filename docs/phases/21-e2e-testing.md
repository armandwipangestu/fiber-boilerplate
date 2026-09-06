# Phase 21 — End-to-End Testing

> Full-stack HTTP tests that boot the real application against a real
> PostgreSQL database and exercise the public API over the network.

## Deliverables

### 21.1 Testable composition root
- New `internal/app` package is the single wiring point: it owns every
  dependency (DB, Redis, cache, tracer, storage, repos, handlers) and
  exposes `Build(logger, cfg) (*Resources, error)` plus a `Shutdown(ctx)`
  that tears down in the same fixed order as before.
- `cmd/app/main.go` is now a thin shell: load config → handle the
  `migrate` subcommand → `app.Build` → signal loop → `app.Shutdown`.

### 21.2 Migration resolution
- `database.RunMigrations` now locates `migrations/` by walking up from the
  caller's file instead of relying on the process cwd, so it works from both
  `cmd/app` and `tests/e2e`.

### 21.3 The E2E suite (`tests/e2e`)
- `setup_test.go`: `TestMain` creates the dedicated `fiber_boilerplate_e2e`
  database if missing, runs `migrate up`, boots the app on an ephemeral port,
  and tears down. Override the target with `TEST_DATABASE_URL`.
- `auth_test.go` — register → login → refresh (httpOnly cookie) → logout →
  refresh-reuse-rejected, duplicate email 409, wrong password 401.
- `user_test.go` — admin-driven user CRUD (create → list → get → update →
  delete → 404), regular user rejected with 403, anonymous rejected 401.
- `rbac_test.go` — admin can list/create/delete roles, regular user is 403.
- Helpers: `api()` JSON client with a cookie jar, `unwrap()` envelope
  decoding, `register`/`adminToken`/`assignRole`.

## Verify
- `TEST_DATABASE_URL=postgres://... go test ./tests/e2e/... -count=1` (8 tests).
- Full suite still green: `go test ./cmd/... ./internal/... -count=1`.