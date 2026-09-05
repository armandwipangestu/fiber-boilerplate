# Phase 9 — Auth Middleware

> Scope: guarding protected routes with Bearer access tokens.

## Deliverables

### 9.1 JWT Middleware (`internal/middleware/auth.go`)
- `NewAuthMiddleware(cfg config.Config) fiber.Handler`
- Reads `Authorization: Bearer <token>`, validates via `auth.ValidateToken`,
  rejects refresh tokens by claim type, stores the user id in locals
  (`pkg.SetUserID`) for downstream handlers and RBAC checks.
- 401 responses: `missing authorization header`, `invalid or expired token`, or
  `a refresh token cannot be used here`.

### 9.2 Wiring
- `server.Dependencies.AuthMiddleware` is injected by the composition root.
- `user.Handler.RegisterRoutes` now applies the middleware to the whole
  `/users` group.
- Auth routes (`/api/v1/auth/*`) remain public, `/api/v1/users/*` are protected.

### 9.3 Tests (`internal/middleware/auth_test.go`)
- Missing header → 401.
- Valid access token → 200 (user id available in handler).
- Refresh token → 401 (claim type guard).
- Garbage token → 401.

## How to Reproduce

```bash
# token from POST /api/v1/auth/login, then:
curl -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/users
```

## Verify
- `go test ./internal/middleware/...` passes.
- Manual curl: `/users` returns 401 without token, 200 with access token,
  401 with a refresh token.