# Phase 12 — CORS & Security Headers

> Scope: cross-origin access policy and baseline hardening headers.

## Deliverables

### 12.1 CORS (`internal/middleware/cors.go`)
- Built on Fiber's `cors` middleware, driven by `CORS_ALLOWED_ORIGINS`
  (comma-separated in `.env`, default `*`).
- Methods `GET,POST,PATCH,PUT,DELETE,OPTIONS`; exposes request id and rate
  limit headers to browsers; `MaxAge: 86400`.
- `AllowCredentials` is only enabled for explicit origins (never with a
  wildcard — the insecure combination Fiber refuses).

### 12.2 Security Headers (`internal/middleware/security.go`)
- `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
  `X-XSS-Protection: 1; mode=block`, a strict `Content-Security-Policy`,
  `Referrer-Policy`, `Permissions-Policy`, `Cross-Origin-Opener-Policy`.
- `Strict-Transport-Security` emitted automatically when the request is
  HTTPS or proxied (X-Forwarded-Proto: https).

### 12.3 Wiring
- Applied globally in `server.New` right after RequestID: security headers,
  then CORS, then rate limit, then throttle.

## How to Reproduce

```bash
rm -f /tmp/oh.txt
# in server: CORS_ALLOWED_ORIGINS=https://app.example.com
curl -sI -H "Origin: https://app.example.com" localhost:8080/ping | grep -i access-control-allow-origin
```

## Verify
- `go test ./internal/middleware/...` passes.
- Tests assert every hardening header is present, HSTS appears behind a
  proxy, and a CORS preflight returns allow-origin + max-age.