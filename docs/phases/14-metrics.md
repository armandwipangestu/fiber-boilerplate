# Phase 14 — Metrics (Prometheus)

> Scope: Prometheus scrape endpoint and automatic middleware instrumentation.

## Deliverables

### 14.1 Metrics registry (`internal/metrics/metrics.go`)
- `http_requests_total` (Counter: method, path, status).
- `http_request_duration_seconds` (Histogram: method, path).
- `http_requests_in_flight` (Gauge).
- Auto-registered via `promauto`; `Handler()` exposes the scrape handler.

### 14.2 Metrics middleware (`internal/middleware/metrics.go`)
- Increments the in-flight gauge before the handler, times the request,
  then records duration + total and decrements the gauge.
- **Path normalization**: UUID path segments are replaced with `:id`, so
  `/users/123e4567-...` labels as `/users/:id` — bounded cardinality.
- Error status derived from `fiber.Error` codes (404/422/etc. reported
  correctly, not masked as 200).

### 14.3 Wiring
- Global (`app.Use`) after logging, and a `/metrics` route served through
  Fiber's `adaptor.HTTPHandler`. Prometheus already scrapes
  `host.docker.internal:8080/metrics` (see `observability/prometheus/`).

## How to Reproduce

```bash
DATABASE_URL='postgres://postgres:postgres@localhost:5433/fiber_boilerplate?sslmode=disable' \
JWT_SECRET=dev PORT=8080 go run cmd/app/main.go
curl -s localhost:8080/ping
curl -s localhost:8080/metrics | grep http_requests_total
```

## Verify
- Middleware test asserts `/ping` counts and `/users/:id` normalization.
- E2E: `http_requests_total{method="GET",path="/ping",status="200"} 3` and
  `http_requests_total{method="GET",path="/users/:id",status="200"} 1`.