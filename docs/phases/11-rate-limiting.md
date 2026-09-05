# Phase 11 — Rate Limiting & Throttling

> Scope: per-client request budgets (429) and a concurrency cap (503).

## Deliverables

### 11.1 Rate Limit Middleware (`internal/middleware/ratelimit.go`)
- `NewRateLimitMiddleware(cfg config.Config) (fiber.Handler, error)` using
  `github.com/ulule/limiter/v3`.
- In-memory store (Redis store lands in Phase 18 behind `REDIS_URL`).
- Keyed by IP when anonymous, by user id when authenticated.
- Sets `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`.
- On exhaustion: 429 `RATE_LIMITED` + `Retry-After`.
- Gates via `RATE_LIMIT_ENABLED` / `RATE_LIMIT_REQUESTS` /
  `RATE_LIMIT_EXPIRATION`.

### 11.2 Throttle Middleware (`internal/middleware/throttle.go`)
- `ThrottleConfig{MaxConcurrent, Timeout}` and `NewThrottleMiddleware`.
- Channel semaphore; when the pool stays full past Timeout the request is
  rejected with 503 `SERVICE_UNAVAILABLE` + `Retry-After`.
- Honors request cancellation (`c.Context().Done()`).

### 11.3 Wiring
- Rate limit + throttle applied globally in `server.New` (after RequestID).
- Defaults: 100 req/60s per client, 1000 concurrent, 5s timeout.

### 11.4 Tests (`internal/middleware/ratelimit_test.go`, `throttle_test.go`)
- Rate limit: 2 allowed, then 429 with headers; disabled → nil handler.
- Throttle: under limit passes; overloaded pool → 503 while a slow handler
  holds the slot.

## How to Reproduce

```bash
RATE_LIMIT_ENABLED=true RATE_LIMIT_REQUESTS=3 RATE_LIMIT_EXPIRATION=60s go run cmd/server/main.go
for i in 1 2 3 4; do curl -i localhost:8080/ping | head -1; done
```

## Verify
- `go test ./internal/middleware/...` passes.
- Manual: request 4+ returns 429 with `X-RateLimit-*` and `Retry-After`.