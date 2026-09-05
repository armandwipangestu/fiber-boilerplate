# Implementation TODO — Step-by-Step Guide

> Each phase builds on the previous one. Complete a phase before moving to the next.
> Mark items with `[x]` as you finish them.

---

## Phase 1 — Project Foundation

### 1.1 Initialize Go Module

```bash
go mod init github.com/armandwipangestu/fiber-boilerplate
```

**Verify:** `go.mod` exists with correct module path.

### 1.2 Create Directory Structure

```bash
mkdir -p cmd/server
mkdir -p internal/{config,server,database,cache,middleware,logging,metrics,tracing,health}
mkdir -p internal/{auth,user,role,permission,rbac,pkg}
mkdir -p migrations
mkdir -p docs/swagger
mkdir -p tests/{e2e,fixtures}
mkdir -p scripts
mkdir -p logs
```

**Verify:** Directory tree matches PRD Section 6.

### 1.3 Create .env.example

Create `docs/../.env.example` (project root) with all variables from PRD Section 24.3.

**Verify:** All config variables documented with sensible defaults.

### 1.4 Create .gitignore

Add to existing `.gitignore`:

```gitignore
# Go
bin/
*.exe
*.test
*.out
coverage.out
coverage.html

# Environment
.env

# Logs
logs/

# IDE
.idea/
.vscode/
*.swp

# Binary
server
```

**Verify:** `bin/`, `.env`, `logs/` are ignored.

### 1.5 Create Configuration Package

**File:** `internal/config/config.go`

Step by step:
1. Define `Config` struct with all fields from PRD Section 24.2
2. Implement `Load()` function that reads from environment variables
3. Add validation for required fields (`DATABASE_URL`, `JWT_SECRET`)
4. Add defaults for optional fields
5. Implement `loadDotEnv()` helper to parse `.env` file for local dev

**Verify:** `config.Load()` returns populated `Config` struct. Test with `APP_ENV=development`.

### 1.6 Create Config Tests

**File:** `internal/config/config_test.go`

Step by step:
1. Test default values
2. Test required field validation (missing `DATABASE_URL`)
3. Test environment variable override
4. Test `.env` file loading

**Verify:** `go test ./internal/config/...` passes.

### 1.7 Create Application Entry Point

**File:** `cmd/server/main.go`

Step by step:
1. Import `internal/config`
2. Call `config.Load()`
3. Print loaded config (for now, just `fmt.Printf`)
4. Start Fiber server on configured host:port
5. Add signal handling for `SIGINT`/`SIGTERM`

**Verify:** `go run cmd/server/main.go` starts and listens on port 8080.

### 1.8 Install Fiber

```bash
go get github.com/gofiber/fiber/v2
```

**Verify:** `go.sum` contains fiber dependency.

---

## Phase 2 — HTTP Server & Middleware Foundation

### 2.1 Create Fiber Server Setup

**File:** `internal/server/server.go`

Step by step:
1. Create `New(cfg config.Config) *fiber.App` function
2. Configure Fiber with:
   - `AppName` from config
   - `DisableHeaderNormalizing: false`
   - `BodyLimit: 10 * 1024 * 1024` (10MB)
3. Return the Fiber app (middleware registration comes later)

**Verify:** Server creates without error, app instance is returned.

### 2.2 Create Request ID Middleware

**File:** `internal/middleware/requestid.go`

Step by step:
1. Create middleware that generates UUID if no `X-Request-ID` header
2. Store request ID in Fiber context via `c.Locals("request_id", id)`
3. Set `X-Request-ID` response header
4. Make request ID available via `internal/pkg/context.go` helper

**Verify:** Every request gets a `X-Request-ID` header in response.

### 2.3 Create Context Helpers

**File:** `internal/pkg/context.go`

Step by step:
1. Define context key types (unexported): `type contextKey struct{}`
2. Create `GetRequestID(c *fiber.Ctx) string` helper
3. Create `SetRequestID(c *fiber.Ctx, id string)` helper
4. Create `GetUserID(c *fiber.Ctx) string` helper (for auth middleware later)

**Verify:** Helpers compile and return correct values from Fiber context.

### 2.4 Create Response Helpers

**File:** `internal/pkg/response.go`

Step by step:
1. Define standard response structs:
   - `SuccessResponse { Success bool, Data any }`
   - `ErrorResponse { Success bool, Error ErrorBody }`
   - `ErrorBody { Code string, Message string, Fields map[string][]string }`
2. Create helper functions:
   - `OK(c *fiber.Ctx, data any) error`
   - `Created(c *fiber.Ctx, data any) error`
   - `BadRequest(c *fiber.Ctx, message string) error`
   - `Unauthorized(c *fiber.Ctx, message string) error`
   - `Forbidden(c *fiber.Ctx, message string) error`
   - `NotFound(c *fiber.Ctx, message string) error`
   - `Conflict(c *fiber.Ctx, message string) error`
   - `ValidationError(c *fiber.Ctx, err error) error`
   - `Error(c *fiber.Ctx, err error) error`
   - `InternalServerError(c *fiber.Ctx) error`
3. `Error()` must check if err is `*AppError` and use its Code/Message/HTTPStatus
4. For non-AppError, return generic 500
5. Include `APP_ENV` check: if `development`, include `internal` error message

**Verify:** Response helpers return correct JSON structure and HTTP status codes.

### 2.5 Create Application Error Type

**File:** `internal/pkg/errors.go`

Step by step:
1. Define `AppError` struct (Code, Message, HTTPStatus, Internal, Details)
2. Implement `Error() string` method
3. Create constructors: `NewAppError`, `NotFound`, `Conflict`, `Unauthorized`, `Forbidden`, `Internal`, `TooManyRequests`
4. Define sentinel errors: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`

**Verify:** Error constructors produce correct AppError instances. `errors.Is` works with sentinels.

### 2.6 Wire Up Middleware in Server

Update `internal/server/server.go`:

Step by step:
1. Import middleware package
2. Add request ID middleware first
3. Add a test route: `GET /ping` returning `{"message": "pong"}`
4. Mount middleware on the app

**Verify:** `GET /ping` returns 200 with `X-Request-ID` header.

---

## Phase 3 — Logging

### 3.1 Create Logger Initialization

**File:** `internal/logging/logger.go`

Step by step:
1. Implement `NewLogger(cfg config.Config) *slog.Logger`
2. Create JSON handler for file/production output
3. Create text handler for console/development output
4. Add source information handler (wraps inner handler, adds `file`, `function`, `line`)
5. Add service name to every log entry
6. Support dual output (console + file simultaneously) using `io.MultiWriter`

**Verify:** Logger produces structured JSON logs to stdout and file.

### 3.2 Create Source Handler

**File:** `internal/logging/source.go`

Step by step:
1. Implement `SourceHandler` that wraps `slog.Handler`
2. In `Handle()`, use `runtime.Caller()` to get file, function, line
3. Add these as attributes to the log record
4. Skip logging frames (adjust caller depth)

**Verify:** Log output includes `file`, `function`, `line` fields.

### 3.3 Create Log Sanitizer

**File:** `internal/logging/sanitizer.go`

Step by step:
1. Define compiled regex patterns for all sensitive data types from PRD Section 41.2
2. Implement `SanitizeString(s string) string` that applies all patterns
3. Implement `SanitizeHandler` wrapping `slog.Handler`
4. In `Handle()`, sanitize `record.Message` and all string attributes
5. Test each pattern individually

**Verify:** Passwords, credit cards, JWTs, SSNs are all masked in output.

### 3.4 Create SensitiveString Type

**File:** `internal/logging/sensitive.go`

Step by step:
1. Define `SensitiveString` struct with `value string`
2. Implement `String() string` returning `"[REDACTED]"`
3. Implement `MarshalJSON()` returning `"[REDACTED]"`
4. Implement `Value() string` for internal use (document: never log this)

**Verify:** `SensitiveString("secret").String()` returns `"[REDACTED]"`.

### 3.5 Create Log Rotation

**File:** `internal/logging/rotate.go`

Step by step:
1. Install `gopkg.in/natefinish/lumberjack.v2`
2. Create `newRotatingWriter(cfg config.Config) io.Writer`
3. Configure `lumberjack.Logger` with MaxSize, MaxBackups, MaxAge, Compress
4. Return the writer

**Verify:** Log files rotate at configured size. Old files are cleaned up.

### 3.6 Create Daily Rotator

**File:** `internal/logging/daily.go`

Step by step:
1. Implement `DailyRotator` with `sync.Mutex`, current file handle, today's date
2. In `Write()`, check if date changed → rotate to new file
3. Format filename: `logs/app-2006-01-02.log`
4. Close previous file, open new one

**Verify:** New log file created each day. Filename contains date.

### 3.7 Integrate Logger with Server

Update `cmd/server/main.go`:

Step by step:
1. Create logger via `logging.NewLogger(cfg)`
2. Set as global `slog` default: `slog.SetDefault(logger)`
3. Replace `fmt.Printf` with `log.Info` calls
4. Log server startup with host/port

**Verify:** Server startup logs to both console and file with structured fields.

### 3.8 Create Logger Tests

**File:** `internal/logging/logger_test.go`

Step by step:
1. Test log level filtering (debug message not shown at info level)
2. Test JSON output format
3. Test console output format
4. Test sanitization masks sensitive data
5. Test source information is present

**Verify:** `go test ./internal/logging/...` passes.

---

## Phase 4 — Database Layer

### 4.1 Create Database Connection

**File:** `internal/database/database.go`

Step by step:
1. Implement `NewDatabase(cfg config.Config) (*sql.DB, error)`
2. Switch on `cfg.Driver`: `postgres` → `pgx`, `mysql` → `mysql`, `sqlite` → `sqlite3`
3. Configure connection pool: MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime
4. Ping with context timeout (5s)
5. Return configured `*sql.DB`

**Verify:** Database connects and pings successfully with PostgreSQL.

### 4.2 Install Database Drivers

```bash
go get github.com/jackc/pgx/v5
go get github.com/go-sql-driver/mysql
go get github.com/mattn/go-sqlite3
go get github.com/jmoiron/sqlx
```

**Verify:** `go.mod` contains all driver dependencies.

### 4.3 Create Transaction Helper

**File:** `internal/database/transaction.go`

Step by step:
1. Define context key type: `type txKey struct{}`
2. Implement `WithTx(ctx context.Context, tx *sql.Tx) context.Context`
3. Implement `GetTx(ctx context.Context) (*sql.Tx, bool)`
4. Implement `WithTransaction(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error`
5. Handle rollback on error, commit on success
6. Handle panic recovery with rollback

**Verify:** Transaction commits on success, rolls back on error, rolls back on panic.

### 4.4 Create Migration Runner

**File:** `internal/database/migrate.go`

Step by step:
1. Install `github.com/golang-migrate/migrate/v4`
2. Implement `RunMigrations(cfg config.Config, direction string) error`
3. Support `up`, `down`, `create` commands
4. Use `migrate.New()` with database URL
5. Handle "no change" gracefully

**Verify:** Migrations run up and down without error.

### 4.5 Create First Migration

**File:** `migrations/000001_create_users.up.sql`

```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
```

**File:** `migrations/000001_create_users.down.sql`

```sql
DROP TABLE IF EXISTS users;
```

**Verify:** `migrate up` creates `users` table. `migrate down` drops it.

### 4.6 Create Remaining Migrations

Create migration files for:
- `000002_create_roles.sql`
- `000003_create_permissions.sql`
- `000004_create_user_roles.sql`
- `000005_create_role_permissions.sql`
- `000006_create_refresh_tokens.sql`

**Verify:** All migrations run up and down cleanly.

### 4.7 Add Migration Subcommand

Update `cmd/server/main.go`:

Step by step:
1. Check if `os.Args[1] == "migrate"`
2. If so, call `database.RunMigrations(cfg, os.Args[2])`
3. Exit after migration completes

**Verify:** `go run cmd/server/main.go migrate up` runs migrations.

---

## Phase 5 — User Feature (Domain)

### 5.1 Create User Domain Model

**File:** `internal/user/model.go`

Step by step:
1. Define `Domain` struct with fields: ID, Email, Name, PasswordHash, CreatedAt, UpdatedAt
2. Use `time.Time` for timestamps (not string)
3. Keep it simple — no methods on the model yet

**Verify:** Model compiles with correct field types.

### 5.2 Create User DTOs

**File:** `internal/user/dto.go`

Step by step:
1. Define `CreateUserRequest` with json + validate tags
2. Define `UpdateUserRequest` with pointer types for partial update
3. Define `ListUsersQuery` with query tags
4. Define `UserResponse` and `ListUsersResponse`
5. Define `PaginationMeta`
6. Add `validate` tags for all fields

**Verify:** DTOs compile with correct tags. `validator.Struct()` works on them.

### 5.3 Create User Repository Interface

**File:** `internal/user/repository.go`

Step by step:
1. Define `Repository` interface with methods:
   - `Create(ctx, *Domain) (*Domain, error)`
   - `GetByID(ctx, string) (*Domain, error)`
   - `GetByEmail(ctx, string) (*Domain, error)`
   - `List(ctx, ListQuery) ([]*Domain, int, error)`
   - `Update(ctx, *Domain) (*Domain, error)`
   - `Delete(ctx, string) error`
   - `ExistsByEmail(ctx, string) (bool, error)`

**Verify:** Interface compiles. No implementation yet.

### 5.4 Create User PostgreSQL Repository

**File:** `internal/user/postgres_repository.go`

Step by step:
1. Define `postgresRepository` struct with `*sql.DB`
2. Implement `NewPostgresRepository(db *sql.DB) Repository`
3. Implement each method:
   - `Create`: INSERT with RETURNING
   - `GetByID`: SELECT by id, handle `sql.ErrNoRows` → `ErrNotFound`
   - `GetByEmail`: SELECT by email
   - `List`: SELECT with WHERE, ORDER BY, LIMIT, OFFSET + COUNT query
   - `Update`: UPDATE with RETURNING
   - `Delete`: DELETE by id
   - `ExistsByEmail`: SELECT 1 WHERE email = ...
4. Add tracing spans to each method
5. Wrap errors with context: `fmt.Errorf("user repository: method: %w", err)`

**Verify:** Repository compiles. Can create/read/update/delete users via SQL.

### 5.5 Create User Errors

**File:** `internal/user/errors.go`

Step by step:
1. Define `ErrEmailExists = NewAppError("CONFLICT", "Email already exists", 409, nil)`
2. Define `ErrCannotDeleteSelf = NewAppError("BUSINESS_ERROR", "Cannot delete your own account", 422, nil)`

**Verify:** Error variables are usable with `errors.Is`.

---

## Phase 6 — User Feature (Service & Handler)

### 6.1 Create User Service

**File:** `internal/user/service.go`

Step by step:
1. Define `Service` struct with `Repository` and `RBACChecker` fields
2. Define `RBACChecker` interface: `HasPermission(ctx, userID, permission string) (bool, error)`
3. Implement `NewService(repo Repository, rbac RBACChecker) *Service`
4. Implement methods:
   - `Create`: check email exists → hash password → repo.Create
   - `GetByID`: repo.GetByID
   - `List`: set defaults (page=1, perPage=20) → repo.List
   - `Update`: repo.GetByID → apply changes → repo.Update
   - `Delete`: check RBAC → check not self → repo.Delete

**Verify:** Service compiles. Business logic is correct (email uniqueness, password hashing, RBAC check).

### 6.2 Create User Service Tests

**File:** `internal/user/service_test.go`

Step by step:
1. Create mock repository struct with function fields
2. Create mock RBAC checker
3. Write test cases:
   - `TestService_Create_Success`
   - `TestService_Create_DuplicateEmail`
   - `TestService_GetByID_Success`
   - `TestService_GetByID_NotFound`
   - `TestService_Delete_Success`
   - `TestService_Delete_CannotDeleteSelf`
   - `TestService_Delete_Forbidden`

**Verify:** `go test ./internal/user/...` passes all tests.

### 6.3 Create User Handler

**File:** `internal/user/handler.go`

Step by step:
1. Define `Handler` struct with `*Service` and `Validator`
2. Define `Validator` interface: `Struct(s any) error`
3. Implement `NewHandler(service *Service, validator Validator) *Handler`
4. Implement methods:
   - `Create`: BodyParser → validate → service.Create → toUserResponse → 201
   - `GetByID`: Parse params → validate UUID → service.GetByID → toUserResponse
   - `List`: QueryParser → validate → service.List → toUserResponseList + meta
   - `Update`: Parse params + BodyParser → validate → service.Update → toUserResponse
   - `Delete`: Parse params → validate → service.Delete → 204
5. Implement `toUserResponse(*Domain) UserResponse` mapping function
6. Implement `toUserResponseList([]*Domain) []UserResponse`

**Verify:** Handler compiles. Request parsing and response formatting work.

### 6.4 Create User Routes

**File:** `internal/user/routes.go`

Step by step:
1. Define `RegisterRoutes(api fiber.Router, h *Handler, authMiddleware fiber.Handler)`
2. Create `/users` group with auth middleware
3. Mount: `GET ""`, `POST ""`, `GET "/:id"`, `PATCH "/:id"`, `DELETE "/:id"`

**Verify:** Routes register without error.

### 6.5 Wire User Feature in Server

Update `internal/server/server.go`:

Step by step:
1. Import user package
2. Create user repository, service, handler (dependency injection)
3. Create a placeholder auth middleware (just `c.Next()`)
4. Call `user.RegisterRoutes(api, handler, authMiddleware)`

**Verify:** `GET /api/v1/users` returns empty list. `POST /api/v1/users` creates a user.

---

## Phase 7 — Validation

### 7.1 Install Validator

```bash
go get github.com/go-playground/validator/v10
```

### 7.2 Create Validator Instance

**File:** `internal/pkg/validator.go`

Step by step:
1. Create `New() *validator.Validate` function
2. Register custom validators if needed (e.g., `password_strength`)
3. Return configured validator

**Verify:** `validator.Struct(CreateUserRequest{...})` returns validation errors for invalid input.

### 7.3 Integrate Validation in Handler

Update user handler to use validator:

Step by step:
1. After `BodyParser`, call `h.validator.Struct(&req)`
2. If error, call `response.ValidationError(c, err)`
3. `ValidationError` must format errors as field-level map

**Verify:** Invalid request body returns 422 with structured field errors.

### 7.4 Create Validation Error Formatter

**File:** `internal/pkg/validation.go`

Step by step:
1. Implement `FormatValidationErrors(err error) map[string][]string`
2. Map `validator.ValidationErrors` to field name → list of messages
3. Handle custom message overrides if needed

**Verify:** Formatter produces correct field→message mapping.

---

## Phase 8 — Authentication

### 8.1 Create JWT Utility

**File:** `internal/auth/jwt.go`

Step by step:
1. Install `github.com/golang-jwt/jwt/v5`
2. Define `TokenClaims` struct embedding `jwt.RegisteredClaims` + custom fields (Type, Family)
3. Implement `GenerateAccessToken(userID string, cfg config.Config) (string, error)`
4. Implement `GenerateRefreshToken(userID string, familyID string, cfg config.Config) (string, error)`
5. Implement `ValidateToken(tokenString string, cfg config.Config) (*TokenClaims, error)`
6. Set issuer, audience, expiration from config
7. Generate unique `jti` for each token

**Verify:** Generate → validate round-trip works. Expired tokens are rejected.

### 8.2 Create JWT Tests

**File:** `internal/auth/jwt_test.go`

Step by step:
1. Test access token generation and validation
2. Test refresh token generation and validation
3. Test expired token rejection
4. Test invalid signature rejection
5. Test clock skew handling (30s)

**Verify:** All JWT tests pass.

### 8.3 Create Refresh Token Model

**File:** `internal/auth/model.go`

Step by step:
1. Define `RefreshToken` struct: ID, UserID, TokenHash, FamilyID, ExpiresAt, CreatedAt, RevokedAt

**Verify:** Model compiles.

### 8.4 Create Refresh Token Repository

**File:** `internal/auth/repository.go`

Step by step:
1. Define `RefreshTokenRepository` interface:
   - `Create(ctx, *RefreshToken) error`
   - `GetByTokenHash(ctx, string) (*RefreshToken, error)`
   - `DeleteByTokenHash(ctx, string) error`
   - `DeleteByFamilyID(ctx, string) error`
   - `DeleteByUserID(ctx, string) error`
   - `DeleteExpired(ctx) error`
2. Implement PostgreSQL version
3. Store hash of token, not raw token

**Verify:** Repository compiles. CRUD operations work.

### 8.5 Create Auth DTOs

**File:** `internal/auth/dto.go`

Step by step:
1. Define `RegisterRequest`: Email, Name, Password
2. Define `LoginRequest`: Email, Password
3. Define `RefreshRequest` (empty — token from cookie)
4. Define `AuthResponse`: AccessToken, User (UserResponse)
5. Define `MessageResponse`: Message string

**Verify:** DTOs compile with correct tags.

### 8.6 Create Auth Service

**File:** `internal/auth/service.go`

Step by step:
1. Define `Service` struct with UserRepository, RefreshTokenRepository, JWT config
2. Implement `Register`: validate email unique → hash password → create user → generate tokens → store refresh token → return
3. Implement `Login`: find user by email → compare password → generate tokens → store refresh token → return
4. Implement `Refresh`: validate refresh token from DB → check reuse → delete old → generate new → store → return
5. Implement `Logout`: delete refresh token from DB → return

**Verify:** Full auth flow works: register → login → refresh → logout.

### 8.7 Create Auth Service Tests

**File:** `internal/auth/service_test.go`

Step by step:
1. Mock user repository and refresh token repository
2. Test register success and duplicate email
3. Test login success and wrong password
4. Test refresh success and revoked token
5. Test refresh token reuse detection

**Verify:** All auth service tests pass.

### 8.8 Create Auth Handler

**File:** `internal/auth/handler.go`

Step by step:
1. Implement `Register`: BodyParser → validate → service.Register → set refresh cookie → 201
2. Implement `Login`: BodyParser → validate → service.Login → set refresh cookie → 200
3. Implement `Refresh`: extract cookie → service.Refresh → set new cookie → 200
4. Implement `Logout`: extract cookie → service.Logout → clear cookie → 204
5. Implement `setRefreshCookie(c *fiber.Ctx, token string)` with HttpOnly, Secure, SameSite, Path

**Verify:** Auth endpoints return correct responses and set cookies properly.

### 8.9 Create Auth Routes

**File:** `internal/auth/routes.go`

Step by step:
1. Define `RegisterRoutes(api fiber.Router, h *Handler)`
2. Mount: `POST "/register"`, `POST "/login"`, `POST "/refresh"`, `POST "/logout"`
3. No auth middleware on these routes

**Verify:** Auth routes register without error.

---

## Phase 9 — Auth Middleware

### 9.1 Create JWT Authentication Middleware

**File:** `internal/middleware/auth.go`

Step by step:
1. Create `NewAuthMiddleware(cfg config.Config) fiber.Handler`
2. Extract `Authorization` header
3. Split "Bearer <token>"
4. Validate token via `auth.ValidateToken()`
5. Extract user_id from claims
6. Store in context: `c.Locals("user_id", userID)`
7. Return 401 if token missing/invalid/expired

**Verify:** Protected routes return 401 without valid token, 200 with valid token.

### 9.2 Wire Auth Middleware

Update `internal/server/server.go`:

Step by step:
1. Create auth middleware instance
2. Pass to feature route registrations
3. Auth routes should NOT use auth middleware

**Verify:** `GET /api/v1/users` returns 401. `POST /api/v1/auth/login` works without token.

---

## Phase 10 — RBAC

### 10.1 Create Role & Permission Models

**File:** `internal/role/model.go`

Step by step:
1. Define `Role` struct: ID, Name, Description, CreatedAt, UpdatedAt

**File:** `internal/permission/model.go`

Step by step:
1. Define `Permission` struct: ID, Name, Description, CreatedAt

### 10.2 Create RBAC Service

**File:** `internal/rbac/service.go`

Step by step:
1. Define `Service` struct with database connection and cache
2. Implement `HasPermission(ctx, userID, permission string) (bool, error)`
3. Check super-admin first
4. Check cache → if miss, load from DB → cache result
5. Implement `HasRole(ctx, userID, role string) (bool, error)`
6. Implement `GetUserPermissions(ctx, userID) ([]string, error)`

**Verify:** Permission check works correctly.

### 10.3 Create RBAC Permission Cache

**File:** `internal/rbac/cache.go`

Step by step:
1. Define `PermissionCache` interface: `Get`, `Set`, `Invalidate`, `InvalidateAll`
2. Implement in-memory version (for now)
3. Cache key: `rbac:permissions:{user_id}`
4. TTL: 5 minutes

**Verify:** Cache stores and retrieves permissions. Invalidation works.

### 10.4 Create RBAC Middleware

**File:** `internal/middleware/rbac.go`

Step by step:
1. Create `RequirePermission(permission string) fiber.Handler`
2. Extract user_id from context
3. Call `rbacService.HasPermission()`
4. Return 403 if not permitted
5. Create `RequireAnyPermission(permissions ...string) fiber.Handler`
6. Create `RequireAllPermissions(permissions ...string) fiber.Handler`

**Verify:** Middleware blocks unauthorized access, allows authorized.

### 10.5 Seed RBAC Data

**File:** `migrations/000007_seed_rbac.sql`

Step by step:
1. Insert default roles: `admin`, `user`
2. Insert permissions: `users.view`, `users.create`, `users.update`, `users.delete`
3. Assign all permissions to admin role
4. Assign `users.view` to user role

**Verify:** Roles and permissions exist in database after migration.

---

## Phase 11 — Rate Limiting & Throttling

### 11.1 Create Rate Limiting Middleware

**File:** `internal/middleware/ratelimit.go`

Step by step:
1. Install `github.com/ulule/limiter/v2`
2. Create `NewRateLimitMiddleware(cfg config.Config) fiber.Handler`
3. If Redis configured: use Redis store
4. If not: use in-memory store
5. Configure per-route rates from config
6. Set rate limit headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
7. Return 429 with `Retry-After` header when exceeded

**Verify:** Rate limiting works. Returns 429 after limit exceeded.

### 11.2 Create Throttling Middleware

**File:** `internal/middleware/throttle.go`

Step by step:
1. Create `ThrottleConfig` struct with MaxConcurrent, Timeout, Mode
2. Create `Semaphore` with channel-based slots
3. Create `NewThrottleMiddleware(config ThrottleConfig) fiber.Handler`
4. Acquire slot → process → release
5. Return 503 with `Retry-After` if timeout

**Verify:** Throttling limits concurrent requests. Returns 503 when overloaded.

### 11.3 Create Rate Limit Response

**File:** `internal/pkg/errors.go` (add to existing)

Step by step:
1. Add `RateLimited(message string) *AppError` constructor
2. Ensure `Error()` returns 429 status

**Verify:** Rate limit response has correct JSON structure.

---

## Phase 12 — CORS & Security Headers

### 12.1 Create CORS Middleware

**File:** `internal/middleware/cors.go`

Step by step:
1. Use Fiber's built-in CORS middleware
2. Configure from `cfg.CORSAllowedOrigins` (comma-separated)
3. Set AllowMethods, AllowHeaders, ExposeHeaders
4. AllowCredentials: true
5. MaxAge: 86400

**Verify:** CORS preflight returns correct headers. Credentials work with explicit origins.

### 12.2 Create Security Headers Middleware

**File:** `internal/middleware/security.go`

Step by step:
1. Create `SecurityHeaders() fiber.Handler`
2. Set: X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, Content-Security-Policy, Referrer-Policy, Permissions-Policy
3. If HTTPS: set Strict-Transport-Security

**Verify:** All security headers present in response.

---

## Phase 13 — Logging Middleware

### 13.1 Create Request Logging Middleware

**File:** `internal/middleware/logging.go`

Step by step:
1. Create `NewLoggingMiddleware(logger *slog.Logger) fiber.Handler`
2. Before handler: log incoming request (method, path, client_ip, user_agent)
3. After handler: log completed request with status, latency_ms, request_id
4. Use structured fields

**Verify:** Every request produces two log entries (start + complete).

### 13.2 Create Tracing Middleware

**File:** `internal/middleware/tracing.go`

Step by step:
1. Create span from incoming request context
2. Set span attributes: method, path, status
3. End span after handler completes
4. Store trace_id and span_id in context for logging

**Verify:** Traces are created for each request (check with OTEL collector).

---

## Phase 14 — Metrics

### 14.1 Create Prometheus Metrics

**File:** `internal/metrics/metrics.go`

Step by step:
1. Install `github.com/prometheus/client_golang`
2. Define metrics:
   - `http_requests_total` (Counter, labels: method, path, status)
   - `http_request_duration_seconds` (Histogram, labels: method, path)
   - `http_requests_in_flight` (Gauge)
3. Implement `Register()` to register all metrics
4. Implement `Handler()` that returns Fiber handler for `/metrics`

**Verify:** `/metrics` endpoint returns Prometheus-formatted metrics.

### 14.2 Create Metrics Middleware

**File:** `internal/middleware/metrics.go`

Step by step:
1. Create `NewMetricsMiddleware() fiber.Handler`
2. Increment in_flight before handler
3. Start timer
4. After handler: record duration, increment request count, decrement in_flight
5. Normalize path labels (replace UUIDs with `:id`)

**Verify:** Metrics increment with each request. Path labels are normalized.

---

## Phase 15 — Health Checks

### 15.1 Create Health Check Handler

**File:** `internal/health/health.go`

Step by step:
1. Create `Handler` struct with db, cache dependencies
2. Implement `Live(c *fiber.Ctx) error` → always 200
3. Implement `Ready(c *fiber.Ctx) error` → check db ping, redis ping
4. Implement `Health(c *fiber.Ctx) error` → combined check
5. Return structured response with status and check results

**Verify:** `/health/live` always 200. `/health/ready` returns dependency status.

### 15.2 Register Health Routes

Update `internal/server/server.go`:

Step by step:
1. Create health handler
2. Mount: `GET /health`, `GET /health/live`, `GET /health/ready`
3. No auth middleware

**Verify:** Health endpoints work without authentication.

---

## Phase 16 — OpenTelemetry Tracing

### 16.1 Create Tracing Setup

**File:** `internal/tracing/tracing.go`

Step by step:
1. Install `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/sdk`, `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`
2. Implement `Init(serviceName, endpoint, sampleRate string) (*sdktrace.TracerProvider, error)`
3. Configure OTLP HTTP exporter
4. Configure sampler based on sample rate
5. Set as global tracer provider
6. Return provider for shutdown

**Verify:** Traces export to OTEL collector.

### 16.2 Create Tracing Shutdown

Step by step:
1. Implement `Shutdown(ctx context.Context) error`
2. Call `tracerProvider.Shutdown(ctx)`
3. Wire into graceful shutdown sequence

**Verify:** Pending spans are flushed on shutdown.

---

## Phase 17 — Graceful Shutdown

### 17.1 Implement Graceful Shutdown

Update `cmd/server/main.go`:

Step by step:
1. Listen for `SIGINT`/`SIGTERM` via `signal.Notify`
2. On signal: stop Fiber server
3. Shutdown in order: Fiber → Tracing → Metrics → Redis → Database → Logger
4. Use timeout context (30s default)
5. Log each shutdown step
6. Exit with code 0

**Verify:** Server shuts down cleanly. Active requests complete before exit.

### 17.2 Add Shutdown Timeout Config

Add `SHUTDOWN_TIMEOUT` env var (default `30s`).

**Verify:** Timeout is configurable.

---

## Phase 18 — Cache Layer

### 18.1 Create Cache Interface

**File:** `internal/cache/cache.go`

Step by step:
1. Define `Cache` interface: `Get`, `Set`, `Delete`, `DeletePattern`, `Close`
2. Define `ErrCacheMiss` sentinel error

**Verify:** Interface compiles.

### 18.2 Create In-Memory Cache

**File:** `internal/cache/memory.go`

Step by step:
1. Implement `MemoryCache` with `sync.RWMutex` + `map[string]memoryEntry`
2. Implement `Get`, `Set`, `Delete`, `DeletePattern`, `Close`
3. Add background cleanup goroutine (every 1 minute)
4. Handle TTL expiration

**Verify:** Cache stores and retrieves values. Expired entries are cleaned up.

### 18.3 Create Redis Cache

**File:** `internal/cache/redis.go`

Step by step:
1. Implement `RedisCache` with `*redis.Client`
2. Implement `Get`, `Set`, `Delete`, `DeletePattern`, `Close`
3. Use JSON marshal/unmarshal for values

**Verify:** Redis cache stores and retrieves values.

### 18.4 Create Cache Factory

**File:** `internal/cache/factory.go`

Step by step:
1. Implement `NewCache(cfg config.Config) (Cache, error)`
2. If Redis URL configured and reachable: use Redis
3. Otherwise: use in-memory
4. Log which cache is being used

**Verify:** Factory returns correct cache type based on config.

### 18.5 Create Cache-Aside Helper

**File:** `internal/cache/aside.go`

Step by step:
1. Implement `GetOrLoad[T any](ctx, cache, key, ttl, loadFn) (T, error)`
2. Try cache first
3. On miss: call loadFn → store result → return
4. Handle cache errors gracefully (log, don't fail request)

**Verify:** Cache-aside pattern works: first call loads, second call hits cache.

---

## Phase 19 — Role & Permission Management

### 19.1 Create Role Repository & Service

**File:** `internal/role/repository.go`, `internal/role/service.go`

Step by step:
1. Define repository interface and PostgreSQL implementation
2. Create service with CRUD operations
3. Create handler and routes

**Verify:** Role CRUD works.

### 19.2 Create Permission Repository & Service

**File:** `internal/permission/repository.go`, `internal/permission/service.go`

Step by step:
1. Define repository interface and PostgreSQL implementation
2. Create service with CRUD operations
3. Create handler and routes

**Verify:** Permission CRUD works.

---

## Phase 20 — API Documentation

### 20.1 Install Swag

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 20.2 Add OpenAPI Annotations

Step by step:
1. Add `// @Summary`, `// @Tags`, `// @Router` to each handler
2. Add `// @Security BearerAuth` to protected endpoints
3. Add `// @Failure` for error responses

### 20.3 Generate Swagger Docs

```bash
swag init -g cmd/server/main.go -o docs/swagger
```

**Verify:** `/swagger/index.html` shows API documentation.

### 20.4 Serve Swagger UI

Update server to serve Swagger in development mode only:

```go
if cfg.AppEnv == "development" {
    app.Get("/swagger/*", swagger.HandlerDefault)
}
```

**Verify:** Swagger UI accessible at `/swagger/index.html` in dev mode.

---

## Phase 21 — Testing

### 21.1 Unit Test Setup

Step by step:
1. Install testify: `go get github.com/stretchr/testify`
2. Create mock implementations for all repository interfaces
3. Write tests for each service method

### 21.2 E2E Test Setup

**File:** `tests/e2e/setup_test.go`

Step by step:
1. Install testcontainers: `go get github.com/testcontainers/testcontainers-go`
2. Create `TestMain` that starts PostgreSQL and Redis containers
3. Run migrations on test database
4. Start application server
5. Cleanup after tests

### 21.3 Auth E2E Tests

**File:** `tests/e2e/auth_test.go`

Step by step:
1. Test register → login → access protected endpoint → refresh → logout flow
2. Test duplicate email registration
3. Test wrong password login
4. Test expired token refresh
5. Test refresh token reuse detection

### 21.4 User E2E Tests

**File:** `tests/e2e/user_test.go`

Step by step:
1. Test create user (admin)
2. Test list users with pagination
3. Test get user by ID
4. Test update user (PATCH)
5. Test delete user
6. Test unauthorized access
7. Test forbidden access

---

## Phase 22 — Docker

### 22.1 Create Production Dockerfile

**File:** `Dockerfile`

Step by step:
1. Multi-stage build: builder + runtime
2. Builder: `golang:1.22-alpine`, copy source, build binary
3. Runtime: `alpine:3.19`, copy binary + migrations
4. Non-root user, healthcheck, expose port

**Verify:** `docker build -t fiber-boilerplate .` succeeds. Image is <30MB.

### 22.2 Create docker-compose.yml

**File:** `docker-compose.yml`

Step by step:
1. PostgreSQL service with volume
2. Redis service with volume
3. OTEL Collector service
4. Prometheus service with config
5. Grafana service with datasource config
6. Swagger UI service

**Verify:** `docker compose up -d` starts all services.

---

## Phase 23 — CI/CD

### 23.1 Create GitHub Actions Workflow

**File:** `.github/workflows/ci.yml`

Step by step:
1. Lint job: `golangci-lint`
2. Unit test job: `go test ./internal/...`
3. E2E test job: with PostgreSQL and Redis service containers
4. Build job: `go build`
5. Swagger check job: generate + diff
6. Security scan job: `govulncheck`
7. Docker build job: only on main

**Verify:** CI pipeline runs successfully on push.

---

## Phase 24 — Taskfile

### 24.1 Create Taskfile.yml

**File:** `Taskfile.yml`

Step by step:
1. Define all tasks from PRD Section 30.2
2. `dev`, `build`, `test`, `test-unit`, `test-e2e`, `lint`, `fmt`
3. `migrate-up`, `migrate-down`, `migrate-create`
4. `swagger`, `docker-up`, `docker-down`

**Verify:** `task dev` starts the full development environment.

---

## Completion Checklist

- [ ] All phases complete
- [ ] All unit tests pass
- [ ] All E2E tests pass
- [ ] CI pipeline green
- [ ] Docker image builds
- [ ] Swagger docs generated
- [ ] README written
- [ ] `.env.example` complete
- [ ] Code coverage > 70%
