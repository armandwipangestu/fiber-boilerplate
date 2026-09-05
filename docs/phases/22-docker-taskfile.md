# Phase 22 — Dockerization & Taskfile

> Production image for the API plus a task runner covering the daily dev loop.

## Deliverables

### 22.1 Production `Dockerfile`
- Multi-stage: `golang:1.26-alpine` builder (`CGO_ENABLED=0`, `-trimpath`,
  `-s -w`) → slim `alpine:3.22` runtime.
- Runs as a non-root `app` user; bundles `ca-certificates`, `tzdata`, and
  `wget` for the built-in `HEALTHCHECK` against `/health/live`.
- `migrations/` is copied in so `server migrate` works inside the container.
- Image lands well under 30 MB.

### 22.2 `docker-compose.yml`
- New `server-fiber-boilerplate` service builds the image and depends on the
  healthy `db-fiber-boilerplate` and `redis-fiber-boilerplate` services,
  exporting `:8080` with its own healthcheck. All observability services
  (OTel collector, Tempo, Prometheus, Grafana, Loki, Alloy, MinIO, pgAdmin)
  were already present.

### 22.3 `Taskfile.yml`
- `task dev` — run locally; `task build` — binary into `./bin`;
  `task test` / `task test-unit` / `task test-e2e`; `task lint` (vet);
  `task fmt`; `task migrate-up|down|version`; `task swagger`;
  `task docker-up|down`.

## Verify
- `docker build -t fiber-boilerplate .` builds and the image runs
  `health/live` (postgres/redis via `docker compose up -d --build`).
- `task test` and `task test-e2e` both green.