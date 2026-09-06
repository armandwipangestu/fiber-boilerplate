# Phase 17 — Graceful Shutdown

> Ordered, signal-driven teardown so in-flight requests complete and pending
> spans flush before the process exits.

## Deliverables

### 17.1 Graceful shutdown in `cmd/app/main.go`
- `signal.NotifyContext` listens for `SIGINT`/`SIGTERM`.
- `app.Listen` runs in a goroutine; listen errors are surfaced on a channel.
- On signal, `shutdown(cfg, logger, app, tracerProvider, rdb, db)` tears down in
  order:
  1. **HTTP server** — `app.ShutdownWithTimeout` stops accepting connections and
     drains active requests.
  2. **Tracing** — `tracing.Shutdown` flushes buffered spans to the collector.
  3. **Metrics** — no-op (Prometheus registry is memory-only; step kept for
     ordering clarity).
  4. **Redis** — client closed.
  5. **Database** — connection pool closed.
  6. Exit code `0` after `shutdown complete` is logged.

### 17.2 Shutdown timeout config
- `SHUTDOWN_TIMEOUT` (default `30s`) bounds the total drain window and is reused
  by `ShutdownWithTimeout` (already present in config from earlier phase).

## Verify
- Run the app and send `SIGTERM`: logs show each teardown step in order, the
  process exits `0`, and in-flight requests are allowed to finish.