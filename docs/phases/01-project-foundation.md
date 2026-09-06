# Phase 1 — Project Foundation

> Scope: initialize the Go module, core directory structure, environment config, and the application entry point.

## Deliverables

### 1.1 Go Module
- `go mod init github.com/armandwipangestu/fiber-boilerplate`

### 1.2 Directory Structure
```
cmd/app/          → application entry point
internal/config/     → environment configuration
internal/server/     → HTTP server setup (later phases)
internal/database/   → database connection (later phases)
internal/cache/      → cache layer (later phases)
internal/{middleware,logging,metrics,tracing,health}/
internal/{auth,user,role,permission,rbac,pkg}/
migrations/          → SQL migrations
docs/swagger/        → generated OpenAPI docs
tests/{e2e,fixtures}/→ integration tests
scripts/             → helper scripts
logs/                → runtime log files
```

### 1.3 Configuration (`internal/config/config.go`)
- **`Config` struct** — every field needed by the whole app (app, database, redis, jwt, rate limit, cors, logging, otel, swagger).
- **`Load()`** — reads env vars (with sensible defaults), auto-loads `.env` file for local dev, then validates required fields (`DATABASE_URL`, `JWT_SECRET`).
- **Typed getters** — `getEnvInt`, `getEnvBool`, `getEnvDuration`, `getEnvFloat`, `getEnvSlice`.

### 1.4 Environment Template (`.env.example`)
- Documents all config variables with commented defaults.

### 1.5 Entry Point (`cmd/app/main.go`)
- Loads config, prints startup info. HTTP server wired in later phases.

## How to Reproduce

```bash
# 1. Initialize module
go mod init github.com/armandwipangestu/fiber-boilerplate

# 2. Create directories
mkdir -p cmd/app
mkdir -p internal/{config,server,database,cache,middleware,logging,metrics,tracing,health}
mkdir -p internal/{auth,user,role,permission,rbac,pkg}

# 3. Install Fiber (used from Phase 2 onward)
go get github.com/gofiber/fiber/v2

# 4. Run
DATABASE_URL=postgres://postgres:postgres@localhost:5432/fiber_boilerplate JWT_SECRET=secret go run cmd/app/main.go
```

## Verify
- `go build ./...` passes.
- `go run cmd/app/main.go` starts and prints config info.
- `go test ./internal/config/...` passes validation rules.
