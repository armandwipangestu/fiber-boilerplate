# Phase 6 — User Feature Service & Handler

> Scope: business rules (Service), HTTP handlers, route wiring, and the
> composition root that connects everything.

## Deliverables

### 6.1 Service (`internal/user/service.go`)
- `Service{repo, rbac}` — `rbac` is the `RBACChecker` interface, nil-safe.
- `Create`: duplicate-email check (409) → bcrypt hash → persist.
- `Update`: PATCH semantics via `*string` fields; re-checks email ownership.
- `List`: page/per_page defaults (page≥1, per_page 1..100) + `PaginationMeta`.
- `Delete`: self-delete guard (422) → rbac `users.delete` (403 when nil) → delete.

### 6.2 Handler (`internal/user/handler.go`)
- `Handler{svc, validate}` backed by `pkg.Validator` (go-playground/validator).
- Routes under `POST/GET /users`, `GET/PATCH/DELETE /users/:id`.
- `:id` validated as UUID → 400 otherwise.
- Request body parsed → validated → `pkg.ValidationResponse` (422 + field map).
- Business errors surface through the central `pkg.Error` responder.

### 6.3 Wiring (`cmd/server/main.go`)
- `database.NewDatabase` (driver name `postgres` → sql driver `pgx`).
- Compose repo → service → handler → `server.New(server.Dependencies{UserHandler})`.
- `server.New` now takes a `Dependencies` struct and registers user routes on `/api/v1`.

## How to Reproduce

```bash
# start Postgres (docker compose), then:
DATABASE_DRIVER=postgres DATABASE_URL="postgres://postgres:postgres@localhost:5433/fiber_boilerplate?sslmode=disable" \
JWT_SECRET=devsecret APP_ENV=development go run cmd/server/main.go

curl -X POST localhost:8080/api/v1/users -H 'Content-Type: application/json' \
  -d '{"email":"ada@lovelace.dev","name":"Ada Lovelace","password":"supersecret"}'
```

## Verify
- `go test ./internal/user/...` passes (service rules with mock repo + stub RBAC).
- Manual curl: create 201, duplicate 409, validation 422 w/ field map,
  get/list 200 w/ meta, patch 200 updates `updated_at`, bad UUID 400.