# Phase 13 — Logging Middleware

> Scope: structured request lifecycle logging correlated via `X-Request-Id`.
> (The tracing middleware from the plan is implemented in Phase 16 where the
> OpenTelemetry SDK is wired.)

## Deliverables

### 13.1 Request Logging (`internal/middleware/logging.go`)
- `NewLoggingMiddleware(logger *slog.Logger) fiber.Handler`.
- Start entry (Debug): method, path, query, client IP, user agent, request id.
- Complete entry (Info): adds `status`, `latency_ms`, `bytes_out`, `error`.
- Uses the shared `pkg.GetRequestID` so logs correlate with trace ids later.

### 13.2 Wiring
- Added `Dependencies.Logger` to `server.New`; the global logger created in
  `main.go` is injected and the middleware runs for every request.

## How to Reproduce

```bash
go run cmd/server/main.go
curl localhost:8080/ping   # LOG_LEVEL=debug to also see "request started"
```

## Verify
- `go test ./internal/middleware/...` passes.
- Test asserts both `request started` and `request completed` (with status and
  latency) are emitted for a 200 request.