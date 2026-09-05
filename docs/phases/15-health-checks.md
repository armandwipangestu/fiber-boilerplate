# Phase 15 — Health Checks

> Scope: liveness and readiness endpoints for orchestrators and load balancers.

## Deliverables

### 15.1 Health handler (`internal/health/health.go`)
- `Handler` holds the database and an optional Redis client.
- `Live` → always `200 {"status":"ok"}` (process is up).
- `Ready` → pings the DB (and Redis when configured, 3s budget) and returns
  `200` with per-dependency results, or `503` when any dependency fails.
- Structured response: `{"status":"ready","checks":[{name,status,error?}]}`.

### 15.2 Routes (`internal/server/server.go`)
- `GET /health` (alias of ready), `GET /health/live`, `GET /health/ready`.
- Registered with **no** auth — these must stay reachable for probes.
- Redis client is opt-in from `REDIS_URL` (absent → skipped from checks).

## How to Reproduce

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5433/fiber_boilerplate?sslmode=disable' \
REDIS_URL='redis://localhost:6379/0' JWT_SECRET=dev PORT=8080 go run cmd/server/main.go
curl localhost:8080/health/live
curl localhost:8080/health/ready
```

## Verify
- Unit tests: live always 200; ready is 503 when DB is missing or unreachable.
- E2E: ready returns 200 with `database: ok` and `redis: ok`.