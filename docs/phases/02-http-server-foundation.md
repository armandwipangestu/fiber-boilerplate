# Phase 2 — HTTP Server & Middleware Foundation

> Scope: Fiber app setup, request id middleware, context helpers, centralized
> error type, and standardized response envelope.

## Deliverables

### 2.1 Server (`internal/server/server.go`)
- `New(cfg config.Config) *fiber.App` builds the Fiber app with:
  - App name from config
  - 10MB body limit
  - Request ID middleware (first)
  - `/api/v1` group scaffold (handlers added in later phases)
  - `/ping` liveness route returning `{"message":"pong"}`

### 2.2 Request ID Middleware (`internal/middleware/requestid.go`)
- Uses `X-Request-ID` header if present, otherwise generates a hex id.
- Stores it in `c.Locals("request_id", id)` and echoes it back in the response.

### 2.3 Context Helpers (`internal/pkg/context.go`)
- `GetRequestID` / `SetRequestID`
- `GetUserID` / `SetUserID` (used by auth middleware in Phase 9)

### 2.4 Application Errors (`internal/pkg/errors.go`)
- `AppError` struct (Code, Message, HTTPStatus, Internal, Details) with `Unwrap`.
- Constructors: `NewAppError`, `NewValidationError`, `BadRequest`, `Unauthorized`,
  `Forbidden`, `NotFound`, `Conflict`, `Internal`, `RateLimited`, `ServiceUnavailable`.
- Sentinel error vars: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, etc.

### 2.5 Response Helpers (`internal/pkg/response.go`)
- `SuccessResponse { success, data }` and `ErrorResponse { success, error }` envelopes.
- Helpers: `OK`, `Created`, `Message`, `BadRequestResponse`, `UnauthorizedResponse`,
  `ForbiddenResponse`, `NotFoundResponse`, `ConflictResponse`, `ValidationResponse`,
  `RateLimitedResponse`, `ServiceUnavailableResponse`, and the central `Error(c, err)`.
- `Error()` inspects the error chain via `AsAppError` and writes the proper status,
  falling back to `Internal` (500) for non-`AppError` values.

## How to Reproduce

```bash
# 1. Install dependencies
go get github.com/gofiber/fiber/v2
go get github.com/stretchr/testify

# 2. Create files listed in Deliverables

# 3. Run
DATABASE_URL=postgres://... JWT_SECRET=secret go run cmd/app/main.go
curl -i http://localhost:8080/ping
# expect 200 + X-Request-Id header
```

## Verify
- `go build ./...` passes.
- `go test ./internal/config/...` passes.
- `curl -i /ping` returns 200 with `X-Request-Id` header.
