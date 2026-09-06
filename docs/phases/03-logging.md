# Phase 3 — Logging

> Scope: structured logging with `log/slog`, sensitive-data sanitization,
> caller source info, and log rotation.

## Deliverables

### 3.1 Logger Init (`internal/logging/logger.go`)
- `NewLogger(cfg)` returns a `*slog.Logger`.
- **Output**: `stdout`, `file`, or `both` (controlled by `LOG_OUTPUT`).
- **Format**: text handler in development, JSON handler otherwise.
- **Level**: from `LOG_LEVEL` (debug/info/warn/error).
- **Service tag**: every record carries `service=<AppName>`.
- Rotation via `lumberjack` (size-based).

### 3.2 Source Handler (`internal/logging/source.go`)
- Wraps an inner handler and appends `file`, `function`, `line` to each record
  using `runtime.Caller`.

### 3.3 Sanitizer (`internal/logging/sanitizer.go`)
- `SanitizeString(s)` masks: credit cards, CVV, SSN, JWTs, refresh tokens,
  API keys (`sk_`/`pk_`/`rk_`/`ak_`), `Bearer <token>`, private keys, DB URLs,
  password/token field assignments.
- `SanitizeHandler` wraps a handler and sanitizes every message + string attr.
- **Only active when `AppEnv != development`** (dev is trusted, prod scrubbed).
- **Order matters**: SSN/credit-card patterns run before the broad CVV rule.

### 3.4 SensitiveString (`internal/logging/sensitive.go`)
- `SensitiveString` never leaks its value via `String()`, `GoString()`,
  `MarshalJSON()` — always `"[REDACTED]"`. Use it for secrets held in structs
  that may be logged. Access the real value via `.Value()` only when needed.

### 3.5 Rotation (`internal/logging/daily.go`)
- Size-based rotation handled by `lumberjack`.
- `dailyRotator` additionally rolls files per-day as `app-YYYY-MM-DD.log`.

## How to Reproduce

```bash
go get gopkg.in/natefinch/lumberjack.v2
# Create files listed above, then:
DATABASE_URL=postgres://... JWT_SECRET=secret \
LOG_OUTPUT=both APP_ENV=development \
go run cmd/app/main.go
# logs appear on stdout and in logs/app.log
```

## Verify
- `go test ./internal/logging/...` passes (sanitize + source + rotator tests).
- In dev mode: `[service=fiber-boilerplate]` present on each line.
- Set `APP_ENV=production` and log a JWT — it appears as `[REDACTED]`.