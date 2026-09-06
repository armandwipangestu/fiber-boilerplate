# Phase 8 — Authentication (JWT + Refresh Tokens)

> Scope: JWT signing/validation, session repository, and the full auth flows:
> register → login → refresh (rotation) → logout.

## Deliverables

### 8.1 JWT Utility (`internal/auth/jwt.go`)
- `TokenClaims`: `type` (access|refresh), `family_id`, embedded
  `jwt.RegisteredClaims`; the user id lives in `sub` (Subject) to avoid
  duplicate json tags.
- `GenerateAccessToken` / `GenerateRefreshToken` (returns family id) / `ValidateToken`.
- Validation enforces HS256, issuer, audience, 30s leeway. Expired → `ErrExpiredToken`.

### 8.2 Session Model (`internal/auth/model.go`) & Repository
- `RefreshToken`: ID, UserID, TokenHash (sha-256 hex), FamilyID, ExpiresAt, RevokedAt.
- `RefreshTokenRepository` interface + Postgres impl on the existing
  `refresh_tokens` migration. Only token hashes are stored, never raw JWT.

### 8.3 Auth Service (`internal/auth/service.go`)
- `Register`: duplicate email → 409; bcrypt hash; create user; issue pair.
- `Login`: generic `INVALID_CREDENTIALS` (401) for both unknown user and bad
  password (no user enumeration).
- `Refresh`: validate token → load session → if revoked/expired wipe the whole
  family (`SESSION_REVOKED`) → revoke consumed token → issue fresh pair.
- `Logout`: delete session row by hash.

### 8.4 Handler & Routes (`internal/auth/handler.go`, `RegisterRoutes`)
- `POST /api/v1/auth/{register,login,refresh,logout}` — all public.
- Refresh token delivered in an httpOnly, SameSite=Lax cookie scoped to
  `/api/v1/auth`; cleared on logout (204).

### 8.5 Tests
- `jwt_test.go`: round-trips, expired rejection, invalid signature, leeway.
- `service_test.go`: register success/duplicate, login success/wrong-password,
  refresh success, revoked-token rejection, **reuse detection** (family wipe).

## How to Reproduce

```bash
go get github.com/golang-jwt/jwt/v5
DATABASE_DRIVER=postgres DATABASE_URL="postgres://postgres:postgres@localhost:5433/fiber_boilerplate?sslmode=disable" \
JWT_SECRET=devsecret APP_ENV=development go run cmd/app/main.go

curl -c /tmp/c.txt -X POST localhost:8080/api/v1/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","name":"You","password":"supersecret"}'
curl -c /tmp/c.txt -b /tmp/c.txt -X POST localhost:8080/api/v1/auth/refresh
```

## Verify
- `go test ./internal/auth/...` passes.
- Manual: register 201 + cookie; login 200; wrong password 401; refresh rotates
  (old token → `SESSION_REVOKED` + family wipe on reuse); logout 204 + clears cookie.