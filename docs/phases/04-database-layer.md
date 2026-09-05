# Phase 4 — Database Layer & Migrations

> Scope: multi-driver connection (postgres/mysql/sqlite), transaction helper,
> SQL migrations with `golang-migrate`, and local dev infrastructure.

## Deliverables

### 4.1 Connection (`internal/database/database.go`)
- `NewDatabase(cfg)` opens the configured driver (`postgres`→pgx, `mysql`,
  `sqlite`), applies pool settings, pings with a 5s timeout.

### 4.2 Transaction Helper (`internal/database/transaction.go`)
- `WithTransaction(ctx, db, fn)` — commits on nil, rolls back on error **and on panic**.

### 4.3 Migration Runner (`internal/database/migrate.go`)
- `RunMigrations(cfg, dir)` supports `up` | `down` | `version`.
- `migrate.New("file://migrations", "pgx5://...")` for postgres.
- Gracefully ignores `migrate.ErrNoChange`.

### 4.4 CLI Subcommand
- `go run cmd/server/main.go migrate up|down|version` (added in main.go).

### 4.5 Migrations (`migrations/`)
| # | Up | Purpose |
|---|----|---------|
| 1 | users | core user table (UUID, email unique) |
| 2 | roles | role table |
| 3 | permissions | permission table |
| 4 | user_roles | many-to-many users↔roles |
| 5 | role_permissions | many-to-many roles↔permissions |
| 6 | refresh_tokens | rotating refresh-token store (family tracking, revocation) |
| 7 | seed_rbac | inserts `admin`/`user` roles + 8 permissions, admin gets all |

### 4.6 Local Infra (`docker-compose.yml` + `observability/`)
- `db-fiber-boilerplate` (Postgres 17, host port **5433**) + pgAdmin (5051)
- `redis-fiber-boilerplate` (Redis 7, 6379)
- OTel Collector, Tempo, Prometheus, Loki, Alloy, Grafana (3002)
  — provisioned datasources (Prometheus/Tempo/Loki) + a Go metrics dashboard.

## How to Reproduce

```bash
docker compose up -d db-fiber-boilerplate redis-fiber-boilerplate

DATABASE_DRIVER=postgres \
DATABASE_URL="postgres://postgres:postgres@localhost:5433/fiber_boilerplate?sslmode=disable" \
JWT_SECRET=devsecret \
go run cmd/server/main.go migrate up
```

## Verify
- `\dt` shows 6 tables + `schema_migrations`.
- `migrate version` → `7 (dirty=false)`.
- `go test ./internal/...` passes; `go vet ./internal/... ./cmd/...` clean.