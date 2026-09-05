# fiber-boilerplate

A production-grade [Fiber](https://gofiber.io) (Go) API boilerplate with
JWT auth, RBAC, object storage, caching, and full observability — ready to
be forked and filled in with your domain.

![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-blue)
![CI](https://github.com/armandwipangestu/fiber-boilerplate/actions/workflows/ci.yml/badge.svg)

---

## Features

**Auth & security**
- JWT access tokens + rotating refresh tokens stored as an **httpOnly**
  cookie, with family-wide revocation on token reuse
- Role-based access control (RBAC) with per-route permission middleware
- bcrypt password hashing, rate limiting, request throttling, CORS, and
  security headers

**Data & storage**
- PostgreSQL (pgx) with a connection pool and `golang-migrate` migrations
- Redis-backed cache with automatic fallback to an in-memory cache
- Object storage: AWS S3 or MinIO via the AWS SDK, with a local-disk
  fallback for development

**Correctness & DX**
- Structured JSON logging with PII/secret sanitization and log rotation
- Prometheus metrics, OpenTelemetry tracing (OTLP), health/readiness probes
- OpenAPI/Swagger docs served in development
- Graceful shutdown, throttling under load
- Unit + end-to-end HTTP test suites, Docker, and a GitHub Actions pipeline

## Tech stack

|            |                                                                             |
| ---------- | --------------------------------------------------------------------------- |
| Language   | Go 1.26                                                                    |
| Web        | [Fiber v2](https://github.com/gofiber/fiber)                                |
| Database   | PostgreSQL (pgx) — MySQL/SQLite supported by the driver layer               |
| Cache      | Redis (`go-redis`) with in-memory fallback                                  |
| Storage    | AWS SDK v2 (S3 / MinIO) with local-disk fallback                            |
| Auth       | `golang-jwt/jwt/v5`, bcrypt                                                 |
| Obs        | Prometheus, OpenTelemetry OTLP, Grafana/Tempo/Loki/Alloy stack              |
| Docs       | Swag (OpenAPI 3), served via `gofiber/contrib/swagger`                      |
| Infra      | Docker, docker-compose, Taskfile, GitHub Actions, semantic-release          |

## Architecture

```
cmd/server           entrypoint + graceful shutdown + migrate subcommand
internal/app         composition root (Build/Shutdown), shared by server & tests
internal/config      env-driven configuration
internal/server      Fiber app, middleware stack, route mounting
internal/{auth,user,rbac}   domains (model → repository → service → handler → routes)
internal/{role,permission}  RBAC domain models
internal/{storage,cache,database}  infrastructure adapters
internal/{logging,metrics,tracing,health,middleware,pkg}  cross-cutting concerns
migrations           SQL migrations (golang-migrate)
tests/e2e            full HTTP end-to-end suite
docs                PRD, phase-by-phase dev log, Swagger output
observability        compose configs for the Grafana/Loki/Alloy stack
```

## Quickstart

Prerequisites: **Docker** (with compose), **Go 1.26+**.

```bash
# 1. Start the infra (PostgreSQL, Redis, MinIO, observability stack)
docker compose up -d

# 2. Point at it and start
cp .env.example .env
go run ./cmd/server            # or: task dev
```

A few minutes later, in a browser:

| Service          | URL                          |
| ---------------- | ---------------------------- |
| API              | http://localhost:8080        |
| Swagger docs     | http://localhost:8080/swagger |
| MinIO console    | http://localhost:9001        |
| Prometheus       | http://localhost:9090        |
| Grafana          | http://localhost:3002        |

### Migrations

```bash
go run ./cmd/server migrate up        # apply         (or: task migrate-up)
go run ./cmd/server migrate down      # revert 1 step
go run ./cmd/server migrate version   # current state
```

## Configuration

All settings are environment variables, documented in `.env.example`. The
only required values are `DATABASE_URL` and `JWT_SECRET`.

| Variable                    | Default                | Purpose                          |
| --------------------------- | ---------------------- | -------------------------------- |
| `APP_ENV`                   | `development`          | toggles dev-only behaviour       |
| `APP_PORT` / `APP_HOST`     | `8080` / `127.0.0.1`   | listen address                   |
| `DATABASE_URL`              | —                      | **required** DSN                 |
| `REDIS_URL`                 | *(empty)*              | empty → in-memory cache          |
| `JWT_SECRET`                | —                      | **required** signing key         |
| `JWT_ACCESS_TOKEN_EXPIRY`   | `15m`                  | access token TTL                 |
| `JWT_REFRESH_TOKEN_EXPIRY`  | `720h`                 | refresh token TTL                |
| `S3_BUCKET` (+ keys/endpoint)| *(empty)*              | set to enable S3/MinIO           |
| `STORAGE_PATH`              | `storage`              | local fallback avatar storage    |
| `PUBLIC_URL`                | `http://localhost:8080`| public base for fallback files   |
| `OTEL_ENABLED`              | `false`                | OTLP trace export                |
| `SWAGGER_ENABLED`           | `true`                 | serve Swagger UI (dev only)      |
| `RATE_LIMIT_ENABLED`        | `true`                 | per-IP rate limiting             |
| `SHUTDOWN_TIMEOUT`          | `30s`                  | graceful drain window            |

## API reference

OpenAPI docs are generated from source annotations and served at
`/swagger` in development. Regenerate them with `task swagger`.

Health probes: `GET /health/live` (liveness — always 200), `GET /health/ready`
and `GET /health` (readiness — 503 when dependencies are down). Metrics:
`GET /metrics`.

### Auth — `POST /api/v1/auth`

| Endpoint                    | Public |
| --------------------------- | ------ |
| `POST /register`            | ✓      |
| `POST /login`               | ✓      |
| `POST /refresh`             | ✓ (cookie) |
| `POST /logout`              | ✓ (cookie) |

`register`/`login` return the access token in the body and set the refresh
token as an httpOnly cookie scoped to `/api/v1/auth`.

### Users — `GET/POST/PATCH/DELETE /api/v1/users[/:id]`

Protected; each route maps to a permission
(`users.view` / `users.create` / `users.update` / `users.delete`).
Avatar endpoints: `POST /:id/avatar` (multipart upload) and
`DELETE /:id/avatar`.

### RBAC management — `/api/v1/roles` and `/api/v1/permissions`

| Endpoint                                    | Permission             |
| ------------------------------------------- | ---------------------- |
| `GET/POST /roles` · `GET/PATCH/DELETE /roles/:id` | `roles.view` / `roles.manage` |
| `GET /roles/:id/permissions`                        | `roles.view`                  |
| `POST /roles/:id/users` · `DELETE /roles/:id/users/:userId` | `roles.manage`      |
| `GET/POST /permissions` · `GET/PATCH/DELETE /permissions/:id` | `permissions.view` / `permissions.manage` |
| `GET /users/:id/roles` · `GET /users/:id/permissions` | `roles.view`          |

The built-in **admin** role cannot be renamed or deleted.

### Seeded roles & permissions (migration `000007`)

```text
admin  → all permissions
user   → users.view, roles.view, permissions.view

permissions: users.view|create|update|delete, roles.view|manage, permissions.view|manage
```

## Security model

- **JWT**: access tokens are short-lived bearer tokens; refresh tokens are
  opaque and stored hashed. Reusing a rotated refresh token revokes the
  entire family (session-hijack mitigation).
- **RBAC**: permission checks are served by a cache (Redis or in-memory,
  5-min TTL) with automatic invalidation on role/permission changes.
- **Headers**: CORS allow-list, `X-Content-Type-Options`, CSP, HSTS, etc.
- **Logs**: secrets (passwords, JWTs, card numbers, SSNs) are sanitized
  before they reach any sink.

## Object storage

When `S3_BUCKET` (and credentials) are set, files go to S3 or MinIO
(`S3_ENDPOINT` + `S3_FORCE_PATH_STYLE=true` for MinIO). Otherwise they are
written under `STORAGE_PATH` and served from `/uploads`, with URLs built
from `PUBLIC_URL`.

## Testing

```bash
task test        # unit:  go test ./cmd/... ./internal/...
task test-e2e    # e2e:   boots the real server against a Postgres DB
task lint        # go vet
```

- The E2E suite (`tests/e2e`) creates a dedicated `fiber_boilerplate_e2e`
  database, runs migrations, and exercises auth, user CRUD, and RBAC over
  HTTP. Point it elsewhere with `TEST_DATABASE_URL`.
- Live integration checks (cache, RBAC) auto-skip when Redis/Postgres at
  `localhost:6379/15` / `localhost:5433` are unreachable.

## Docker & task runner

```bash
docker build -t fiber-boilerplate .   # multi-stage, non-root, <120 MB
docker compose up -d --build          # full stack incl. the API service

task docker-up      # same as above
task docker-down
task build          # binary into ./bin
task fmt / migrate-* / swagger        # everyday dev plumbing
```

The production image runs as an unprivileged user with a healthcheck probe
on `/health/live` and bundles `migrations/` for in-container `migrate`.

## CI/CD

`.github/workflows/ci.yml` runs on every PR to `main`/`staging`:

- **Build, vet & test** — build, `go vet`, strict `gofmt` check, unit tests
- **End-to-end** — `postgres:17` + `redis:7` service containers, then the
  HTTP E2E suite
- **Vulnerability scan** — `govulncheck` over all packages
- **Docker build** — Buildx image build with layer caching

The existing semantic-release pipeline (`release.yml`) versions and ships
the image to GHCR/Docker Hub on merge. The full roadmap log lives in
`docs/phases` — one document per feature phase.

## License

MIT.