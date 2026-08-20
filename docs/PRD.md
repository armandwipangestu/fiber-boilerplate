# Product Requirements Document: Production-Ready Go REST API Boilerplate

**Version:** 1.0.0
**Status:** Draft
**Date:** 2026-08-20
**Author:** opencode
**Reviewers:** TBD

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Goals](#2-goals)
3. [Non-Goals](#3-non-goals)
4. [Architecture Overview](#4-architecture-overview)
5. [Technology Stack](#5-technology-stack)
6. [Project Structure](#6-project-structure)
7. [Request Lifecycle](#7-request-lifecycle)
8. [DTO Conventions](#8-dto-conventions)
9. [Validation](#9-validation)
10. [Error Handling](#10-error-handling)
11. [Logging](#11-logging)
12. [Authentication](#12-authentication)
13. [RBAC Authorization](#13-rbac-authorization)
14. [OpenAPI](#14-openapi)
15. [Rate Limiting](#15-rate-limiting)
16. [CORS](#16-cors)
17. [Prometheus Metrics](#17-prometheus-metrics)
18. [OpenTelemetry Distributed Tracing](#18-opentelemetry-distributed-tracing)
19. [Health Checks](#19-health-checks)
20. [Database Architecture](#20-database-architecture)
21. [Repository Pattern](#21-repository-pattern)
22. [Transactions](#22-transactions)
23. [Configuration](#23-configuration)
24. [Graceful Shutdown](#24-graceful-shutdown)
25. [API Versioning](#25-api-versioning)
26. [Pagination](#26-pagination)
27. [Testing Strategy](#27-testing-strategy)
28. [Security](#28-security)
29. [Developer Experience](#29-developer-experience)
30. [Docker](#30-docker)
31. [CI/CD](#31-cicd)
32. [Example: User Feature](#32-example-user-feature)
33. [Observability Architecture](#33-observability-architecture)
34. [Scalability Strategy](#34-scalability-strategy)
35. [Architecture Decision Records](#35-architecture-decision-records)
36. [Implementation Phases](#36-implementation-phases)
37. [Acceptance Criteria](#37-acceptance-criteria)
38. [Risks and Trade-offs](#38-risks-and-trade-offs)
39. [Future Extensions](#39-future-extensions)
40. [Log Sanitization and Sensitive Data Masking](#40-log-sanitization-and-sensitive-data-masking)
41. [Application-Level Throttling](#41-application-level-throttling)
42. [Caching Mechanism](#42-caching-mechanism)

---

## 1. Executive Summary

This PRD defines a reusable, production-ready Go REST API boilerplate built on the Fiber framework. The boilerplate serves as the standardized foundation for all future backend services within the organization.

It provides battle-tested implementations for HTTP serving, authentication (JWT with refresh token rotation), role-based access control (RBAC), structured logging with rotation, distributed tracing, Prometheus metrics, database access with multi-DB support, migration management, rate limiting (local and distributed), and comprehensive testing patterns.

The architecture is pragmatic and feature-oriented — not a dogmatic Clean Architecture or DDD implementation. Every abstraction must justify its existence. A developer should be able to trace the full request lifecycle by reading code, not by navigating 15 layers of interfaces.

The boilerplate targets PostgreSQL as the primary database, with adapter support for MySQL, MariaDB, and SQLite.

---

## 2. Goals

| Goal | Description |
|------|-------------|
| **Standardization** | Provide a single starting point for all new Go backend services. |
| **Production readiness** | Include security, observability, and operational concerns from day one. |
| **Maintainability** | Clean architecture with clear boundaries, no hidden magic. |
| **Scalability** | Stateless design, connection pooling, distributed rate limiting. |
| **Developer experience** | Minimal setup, good tooling, comprehensive documentation. |
| **Testability** | Both unit and E2E tests with real infrastructure. |
| **Extensibility** | Easy to add new features without restructuring the project. |
| **Multi-database** | Support PostgreSQL, MySQL, MariaDB, SQLite through adapter pattern. |

---

## 3. Non-Goals

| Non-Goal | Reason |
|----------|--------|
| Full application framework | This is a boilerplate, not a framework. |
| Laravel clone in Go | Go has its own idioms. |
| Dogmatic Clean Architecture | Pragmatic architecture over religious patterns. |
| Generic CRUD generator | Each feature has its own conventions. |
| Giant shared package (`pkg/`) | Packages grow organically from real needs. |
| DI framework | Explicit constructors are clear and traceable. |
| Message broker integration | Add when the use case requires it. |
| GraphQL | Out of scope for a REST API boilerplate. |
| gRPC | Out of scope for this boilerplate. |
| Kubernetes-specific manifests | Deployment orchestration is environment-specific. |
| Every feature using every abstraction | Not every feature needs caching, tracing spans, or metrics. |

---

## 4. Architecture Overview

### 4.1 Layered Architecture

```text
HTTP / Transport Layer (Fiber handlers, middleware)
            │
            ▼
Application / Service / Use Case Layer (business logic orchestration)
            │
            ▼
Domain Layer (entities, value objects, domain errors)
            │
            ▼
Repository Interfaces (defined in domain, implemented in infrastructure)
            │
            ▼
Infrastructure Layer (database drivers, Redis, external APIs, file system)
```

### 4.2 Dependency Rules

1. **Dependencies point inward.** HTTP → Service → Domain. Never the reverse.
2. **Domain has no external dependencies.** No imports from `net/http`, `database/sql`, Fiber, or any infrastructure package.
3. **Repository interfaces live near the domain.** Implementations live in infrastructure.
4. **Service layer orchestrates.** It owns transaction boundaries, calls repositories, and coordinates multiple domain operations.
5. **HTTP layer is thin.** It parses requests, validates input, calls the service layer, and formats responses. No business logic.

### 4.3 Feature-Oriented Organization

The project is organized primarily by **feature**, not by technical layer. Each feature module is self-contained.

```text
internal/
├── auth/           # Authentication feature
├── user/           # User management feature
├── role/           # Role management feature
├── permission/     # Permission management feature
└── ...
```

Each feature module contains:

```text
feature/
├── handler.go          # HTTP handlers (Fiber)
├── handler_test.go     # Handler tests
├── service.go          # Business logic
├── service_test.go     # Service tests
├── repository.go       # Repository interface
├── model.go            # Domain model / entity
├── dto.go              # Request/response DTOs
├── routes.go           # Route registration
└── errors.go           # Feature-specific errors (optional)
```

### 4.4 Shared / Cross-Cutting Concerns

Shared infrastructure lives in dedicated packages:

```text
internal/
├── config/         # Configuration loading
├── server/         # Fiber server setup, shutdown
├── middleware/      # HTTP middleware (auth, rate limit, CORS, etc.)
├── database/       # Database connection, migration runner
├── cache/          # Redis client wrapper
├── logging/        # Logger initialization
├── metrics/        # Prometheus metrics
├── tracing/        # OpenTelemetry setup
├── health/         # Health check handlers
└── pkg/            # Small, reusable utilities (only when needed)
```

### 4.5 When a Component Belongs Where

| Location | Belongs | Does NOT Belong |
|----------|---------|-----------------|
| `cmd/` | Application entry point only | Business logic, HTTP handlers |
| `internal/feature/` | Feature-specific handlers, services, repos, DTOs, models | Cross-cutting concerns, shared utilities |
| `internal/config/` | Configuration structs, loading, validation | Business logic |
| `internal/server/` | Fiber app setup, route mounting, shutdown | Individual route handlers |
| `internal/middleware/` | HTTP middleware | Business logic |
| `internal/database/` | DB connection, migrations, transaction helpers | Domain logic |
| `internal/pkg/` | Small, truly reusable utilities with no feature coupling | Feature-specific code |

### 4.6 Circular Dependency Prevention

- `internal/feature/` packages must not import each other directly. Cross-feature coordination goes through the service layer or shared infrastructure.
- `internal/` packages must not import from `cmd/`.
- `internal/pkg/` must not import from any `internal/` package.
- Use `go vet` and `go mod graph` to verify dependency direction in CI.

---

## 5. Technology Stack

### 5.1 Core Dependencies

| Dependency | Selection | Problem Solved | Alternatives | Trade-offs | Coupling |
|------------|-----------|----------------|--------------|------------|----------|
| **Language** | Go 1.22+ | Performance, simplicity, ecosystem | Rust, Node.js, Python | Go is the org's backend language | N/A |
| **HTTP Framework** | Fiber v2 | High-performance HTTP framework, familiar to team | Gin, Echo, Chi, net/http | Fiber uses fasthttp (not net/http), some stdlib middleware won't work directly. Team already uses Fiber. | Medium — handlers depend on `*fiber.Ctx`, but service layer does not |
| **Primary Database** | PostgreSQL | Feature-rich, JSON support, extensions, proven reliability | MySQL, MariaDB | Slightly more setup than SQLite, but production-grade | Low — behind repository interfaces |
| **Additional DB Support** | MySQL, MariaDB, SQLite | Flexibility for different deployment targets | CockroachDB, MongoDB | SQLite not suitable for production multi-instance deployments | Low — adapter pattern |
| **SQL Builder** | `sqlx` | Type-safe SQL with struct scanning, no ORM magic | GORM, sqlc | `sqlx` is a thin wrapper over `database/sql`. GORM adds magic and leaky abstractions. `sqlc` generates code from SQL. | Low — used only in repository implementations |
| **Migration Tool** | `golang-migrate/migrate` | Versioned, reversible SQL migrations | goose, atlas | Supports all 4 DBs, CLI + library, battle-tested | Low — run as CLI or programmatically |
| **Redis Client** | `go-redis/redis/v9` | Caching, distributed rate limiting, session storage | redigo, rueidis | Most maintained, feature-rich Redis client | Low — behind interface |
| **JWT** | `golang-jwt/jwt/v5` | Token generation and validation | jose, custom | De facto standard for Go JWT | Low — behind interface |
| **Password Hashing** | `golang.org/x/crypto/bcrypt` | Secure password hashing | argon2, scrypt | Well-understood, widely supported | Low — utility function |
| **Validation** | `go-playground/validator/v10` | Struct-based request validation | ozzo-validation, custom | Most widely used, supports custom tags, nested validation | Low — used in DTO layer |
| **Structured Logging** | `slog` (stdlib) | Structured, leveled logging | zerolog, zap, logrus | Stdlib since Go 1.21, zero dependencies | Low — behind interface |
| **Log Rotation** | `lumberjack.v2` | Log file rotation by size and age | Custom rotation | Standard, simple, reliable | Low — used in logger init |
| **Metrics** | `prometheus/client_golang` | Prometheus-compatible metrics | opentelemetry/metrics | Most mature. OTEL metrics is newer but less battle-tested. | Low — metrics package |
| **Tracing** | OpenTelemetry SDK + OTLP exporter | Distributed tracing | Jaeger, Zipkin | CNCF standard, vendor-neutral | Low — behind OTEL API |
| **Rate Limiting** | `ulule/limiter` | HTTP rate limiting | `didip/tollbooth`, custom | Supports Redis and in-memory, middleware-ready | Low — middleware |
| **CORS** | Fiber built-in middleware | CORS configuration | `rs/cors`, custom | Fiber has built-in CORS middleware | Low — middleware config |
| **API Docs** | `swaggo/swag` | OpenAPI/Swagger from annotations | oapi-codegen, Fern | Code-first, annotations near handlers | Low — annotations |
| **Testcontainers** | `testcontainers-go` | E2E test infrastructure | docker-compose scripts | Manages Docker containers programmatically, auto-cleanup | Low — test only |
| **Configuration** | `envconfig` | Environment variable loading | `caarlos0/env` | Simple, explicit | Low — config layer |

### 5.2 Dependency Coupling Strategy

**Low coupling:** Most dependencies are used only in specific layers and can be replaced without touching business logic:
- Database driver → repository implementation only
- Redis → cache/rate-limit implementation only
- Fiber → HTTP handler layer only
- Logging → logger initialization only
- Tracing → tracing initialization only

**Medium coupling:** Some dependencies touch multiple layers but behind interfaces:
- JWT → auth service + auth middleware
- Validator → DTO layer

**Key principle:** The domain layer and service layer must have zero infrastructure imports. Swap `sqlx` for `sqlc` by only changing repository implementations. Swap Fiber for Chi by only rewriting HTTP handlers.

---

## 6. Project Structure

### 6.1 Complete Project Structure

```text
fiber-boilerplate/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── server/
│   │   ├── server.go
│   │   └── shutdown.go
│   ├── database/
│   │   ├── database.go
│   │   ├── migrate.go
│   │   └── transaction.go
│   ├── cache/
│   │   └── redis.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── rbac.go
│   │   ├── ratelimit.go
│   │   ├── cors.go
│   │   ├── requestid.go
│   │   ├── logging.go
│   │   ├── metrics.go
│   │   └── tracing.go
│   ├── logging/
│   │   ├── logger.go
│   │   └── logger_test.go
│   ├── metrics/
│   │   └── metrics.go
│   ├── tracing/
│   │   └── tracing.go
│   ├── health/
│   │   └── health.go
│   ├── auth/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── repository.go
│   │   ├── jwt.go
│   │   ├── jwt_test.go
│   │   ├── dto.go
│   │   ├── model.go
│   │   └── routes.go
│   ├── user/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── repository.go
│   │   ├── dto.go
│   │   ├── model.go
│   │   ├── routes.go
│   │   └── errors.go
│   ├── role/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── repository.go
│   │   ├── dto.go
│   │   ├── model.go
│   │   └── routes.go
│   ├── permission/
│   │   ├── handler.go
│   │   ├── handler_test.go
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── repository.go
│   │   ├── dto.go
│   │   ├── model.go
│   │   └── routes.go
│   ├── rbac/
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── cache.go
│   └── pkg/
│       ├── response.go
│       ├── pagination.go
│       ├── context.go
│       └── errors.go
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_roles.up.sql
│   ├── 000002_create_roles.down.sql
│   ├── 000003_create_permissions.up.sql
│   ├── 000003_create_permissions.down.sql
│   ├── 000004_create_user_roles.up.sql
│   ├── 000004_create_user_roles.down.sql
│   ├── 000005_create_role_permissions.up.sql
│   ├── 000005_create_role_permissions.down.sql
│   ├── 000006_create_refresh_tokens.up.sql
│   └── 000006_create_refresh_tokens.down.sql
├── docs/
│   ├── PRD.md
│   └── swagger/
│       └── docs.go
├── tests/
│   ├── e2e/
│   │   ├── setup_test.go
│   │   ├── auth_test.go
│   │   ├── user_test.go
│   │   └── health_test.go
│   └── fixtures/
├── scripts/
│   ├── migrate-create.sh
│   └── seed.go
├── .env.example
├── .gitignore
├── Taskfile.yml
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

### 6.2 Directory Responsibilities

| Directory | Responsibility | Belongs | Does NOT Belong |
|-----------|---------------|---------|-----------------|
| `cmd/server/` | Application entry point, dependency wiring, startup | `main.go` only | Business logic, configuration structs |
| `internal/config/` | Config struct, loading from env, validation | Config definition and loading | Business logic, HTTP handlers |
| `internal/server/` | Fiber app creation, middleware registration, route mounting, shutdown | Server setup only | Individual feature logic |
| `internal/database/` | DB connection pool, migration execution, transaction helper | Database infrastructure | Domain logic, feature-specific queries |
| `internal/cache/` | Redis client wrapper | Cache infrastructure | Business logic |
| `internal/middleware/` | HTTP middleware functions | Cross-cutting HTTP concerns | Business logic |
| `internal/logging/` | Logger initialization, configuration | Logger setup | Log usage (use `slog` directly) |
| `internal/metrics/` | Prometheus metric definitions, registration | Metrics infrastructure | Business logic |
| `internal/tracing/` | OpenTelemetry setup | Tracing infrastructure | Business logic |
| `internal/health/` | Health check handlers and dependency checking | Health endpoints | Business logic |
| `internal/auth/` | Authentication feature | Auth feature code | Other feature code |
| `internal/user/` | User management feature | User feature code | Auth, role, permission code |
| `internal/role/` | Role management feature | Role feature code | User, permission code |
| `internal/permission/` | Permission management feature | Permission feature code | User, role code |
| `internal/rbac/` | RBAC checking logic, permission caching | Authorization concern | Feature-specific logic |
| `internal/pkg/` | Small reusable utilities | Generic helpers with no feature coupling | Feature-specific code |
| `migrations/` | SQL migration files | Database schema changes | Application code |
| `tests/e2e/` | End-to-end integration tests | Integration test code | Unit tests |
| `scripts/` | Development and operational scripts | CLI helpers | Application code |

### 6.3 Dependency Direction

```text
cmd/server
    ├── imports: internal/config
    ├── imports: internal/server
    ├── imports: internal/database
    ├── imports: internal/cache
    ├── imports: internal/logging
    ├── imports: internal/metrics
    ├── imports: internal/tracing
    └── imports: internal/<feature> (for route registration)

internal/server
    ├── imports: internal/middleware
    ├── imports: internal/health
    ├── imports: internal/<feature>/routes
    └── imports: internal/config

internal/<feature>
    ├── imports: internal/pkg (utilities only)
    └── does NOT import other internal/<feature> packages

internal/middleware
    ├── imports: internal/auth (for JWT validation)
    ├── imports: internal/rbac (for permission checking)
    ├── imports: internal/pkg (context helpers)
    └── imports: internal/config

internal/database
    └── imports: third-party DB drivers only

internal/pkg
    └── imports: standard library only
```

---

## 7. Request Lifecycle

### 7.1 Full Request Flow

```text
1. HTTP Request arrives at Fiber
2. Request ID middleware assigns/propagates request ID
3. Logging middleware logs incoming request
4. Tracing middleware creates/propagates span
5. Metrics middleware starts timer
6. Rate limit middleware checks limits
7. CORS middleware handles preflight
8. Auth middleware validates JWT (if protected route)
9. RBAC middleware checks permissions (if required)
10. Handler parses request into operation-specific DTO
11. Handler validates DTO
12. Handler calls service method
13. Service executes business logic
14. Service calls repository (within transaction if needed)
15. Repository executes SQL query
16. Repository returns domain model
17. Service processes result, applies business rules
18. Service returns result to handler
19. Handler maps domain model to response DTO
20. Handler sends JSON response
21. Metrics middleware records latency and status
22. Logging middleware logs completed request with duration
23. Tracing middleware ends span
```

### 7.2 Context Propagation

The following values are propagated through `context.Context` across all layers:

- `request_id` — unique identifier for the request
- `trace_id` — OpenTelemetry trace identifier
- `span_id` — OpenTelemetry span identifier
- `user_id` — authenticated user ID (set by auth middleware)
- `user_roles` — authenticated user's roles (set by auth middleware)

Context is the **only** acceptable mechanism for cross-cutting concern propagation. Global variables, goroutine-local storage, and package-level state are forbidden.

---

## 8. DTO Conventions

### 8.1 Naming Convention

DTOs are **operation/use-case oriented**, not entity-oriented.

```text
✅ CreateUserRequest
✅ UpdateUserRequest
✅ ListUsersQuery
✅ LoginRequest
✅ RegisterRequest
✅ RefreshTokenRequest
✅ ChangePasswordRequest

❌ UserRequest       (ambiguous — which operation?)
❌ AuthRequest       (ambiguous — login? register? refresh?)
❌ GenericRequest    (no meaning)
```

### 8.2 Request DTO Types

| Type | Use | Location |
|------|-----|----------|
| **Path parameters** | Resource identification (`:id`) | Parsed in handler via `c.Params()` |
| **Query parameters** | Filtering, sorting, pagination | Dedicated query struct per list operation |
| **Request body** | Create, update, complex operations | Dedicated request struct per operation |

### 8.3 Request Body Convention

```go
// Create user — operation-specific
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8,max=72"`
}

// Update user — partial update semantics
type UpdateUserRequest struct {
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
    Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
}

// List users — query parameters
type ListUsersQuery struct {
    Page    int    `query:"page" validate:"omitempty,min=1"`
    PerPage int    `query:"per_page" validate:"omitempty,min=1,max=100"`
    Sort    string `query:"sort" validate:"omitempty,oneof=name email created_at"`
    Order   string `query:"order" validate:"omitempty,oneof=asc desc"`
    Search  string `query:"search" validate:"omitempty,max=255"`
}
```

### 8.4 PUT vs PATCH Semantics

- **PUT** (`/api/v1/users/:id`): Full replacement. All fields are required.
- **PATCH** (`/api/v1/users/:id`): Partial update. Only provided fields are updated. Use pointer types (`*string`, `*int`) with `omitempty`.

**Recommendation:** Use PATCH for updates by default.

### 8.5 Response DTO Convention

```go
type UserResponse struct {
    ID        string `json:"id"`
    Email     string `json:"email"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type ListUsersResponse struct {
    Data []UserResponse `json:"data"`
    Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

### 8.6 DTO-to-Domain Mapping

Handlers are responsible for mapping between DTOs and domain models. This mapping happens in the handler layer, not in the service layer.

```go
func (h *Handler) Create(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return response.BadRequest(c, "Invalid request body")
    }

    if err := h.validator.Struct(req); err != nil {
        return response.ValidationError(c, err)
    }

    user := &user.Domain{
        Email:    req.Email,
        Name:     req.Name,
        Password: req.Password,
    }

    created, err := h.service.Create(c.Context(), user)
    if err != nil {
        return response.Error(c, err)
    }

    return response.Created(c, toUserResponse(created))
}

func toUserResponse(u *user.Domain) UserResponse {
    return UserResponse{
        ID:        u.ID,
        Email:     u.Email,
        Name:      u.Name,
        CreatedAt: u.CreatedAt.Format(time.RFC3339),
        UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
    }
}
```

### 8.7 Rules

1. Request DTOs must never embed database models directly.
2. Response DTOs must never expose password hashes or sensitive fields.
3. DTOs belong in the feature package, not in a shared `dto/` package.
4. Each operation gets its own DTO.
5. Timestamps in responses use ISO 8601 / RFC 3339 format.

---

## 9. Validation

### 9.1 Mechanism

Use `go-playground/validator/v10` for struct-based validation with custom tags.

### 9.2 Validation Rules

```go
type CreateUserRequest struct {
    Email    string `json:"email"    validate:"required,email"`
    Name     string `json:"name"     validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8,max=72"`
    RoleID   string `json:"role_id"  validate:"omitempty,uuid"`
}
```

### 9.3 Custom Validation

```go
v := validator.New()
v.RegisterValidation("password_strength", validatePasswordStrength)
```

### 9.4 Validation Error Response

```json
{
    "success": false,
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Request validation failed",
        "fields": {
            "email": ["must be a valid email address"],
            "password": ["must be at least 8 characters"]
        }
    }
}
```

### 9.5 Validation Error Propagation

```text
Handler parses request body into DTO
    │
    ▼
Handler calls validator.Struct(dto)
    │
    ▼
validator returns validation.ValidationErrors
    │
    ▼
Handler maps errors to field-level message map
    │
    ▼
Handler returns response.ValidationError(c, err)
    │
    ▼
Response returns HTTP 422 with structured error body
```

Validation errors are **always** caught and formatted at the handler layer. They never propagate to the service layer.

### 9.6 HTTP Status Code

Validation errors return **HTTP 422 Unprocessable Entity**.

Rationale: 400 is for malformed requests (invalid JSON). 422 is for semantically invalid data (valid JSON, invalid field values).

---

## 10. Error Handling

### 10.1 Application Error Type

```go
type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    HTTPStatus int    `json:"-"`
    Internal   error  `json:"-"`
    Details    any    `json:"-"`
}
```

### 10.2 Error Categories

| Category | Code | HTTP Status | Example |
|----------|------|-------------|---------|
| Validation error | `VALIDATION_ERROR` | 422 | Invalid field value |
| Authentication error | `UNAUTHORIZED` | 401 | Missing or invalid JWT |
| Authorization error | `FORBIDDEN` | 403 | Insufficient permissions |
| Not found | `NOT_FOUND` | 404 | Resource does not exist |
| Conflict | `CONFLICT` | 409 | Duplicate email |
| Business rule violation | `BUSINESS_ERROR` | 422 | Cannot delete active user |
| Database error | `INTERNAL_ERROR` | 500 | Query failed (logged internally) |
| External service error | `UPSTREAM_ERROR` | 502 | Third-party API timeout |
| Rate limited | `RATE_LIMITED` | 429 | Too many requests |
| Internal error | `INTERNAL_ERROR` | 500 | Unexpected failure |

### 10.3 Error Construction

```go
func NewAppError(code string, message string, httpStatus int, internal error) *AppError {
    return &AppError{Code: code, Message: message, HTTPStatus: httpStatus, Internal: internal}
}

func NotFound(resource string) *AppError {
    return NewAppError("NOT_FOUND", resource+" not found", 404, nil)
}

func Conflict(message string) *AppError {
    return NewAppError("CONFLICT", message, 409, nil)
}

func Unauthorized(message string) *AppError {
    return NewAppError("UNAUTHORIZED", message, 401, nil)
}

func Forbidden(message string) *AppError {
    return NewAppError("FORBIDDEN", message, 403, nil)
}

func Internal(err error) *AppError {
    return NewAppError("INTERNAL_ERROR", "An internal error occurred", 500, err)
}
```

### 10.4 Error Wrapping

Use `fmt.Errorf("context: %w", err)` for internal error wrapping. Preserves the error chain for logging while keeping internal details hidden from API responses.

```go
// In repository
users, err := r.db.SelectContext(ctx, query, args...)
if err != nil {
    return nil, fmt.Errorf("user repository: list: %w", err)
}

// In service
users, err := r.repo.List(ctx, query)
if err != nil {
    return nil, fmt.Errorf("user service: list: %w", err)
}
```

### 10.5 Error Propagation Through Layers

```text
Repository returns: fmt.Errorf("user repository: get by id: %w", err)
    │
    ▼
Service returns: fmt.Errorf("user service: get: %w", err)
    │
    ▼
Handler checks: errors.Is(err, domain.ErrNotFound)
    │
    ▼
If AppError: return response.Error(c, appErr)
If wrapped error: log internal error, return generic 500 to client
```

### 10.6 Production vs Development Error Response

**Production:**
```json
{
    "success": false,
    "error": {
        "code": "INTERNAL_ERROR",
        "message": "An internal error occurred"
    }
}
```

**Development (when `APP_ENV=development`):**
```json
{
    "success": false,
    "error": {
        "code": "INTERNAL_ERROR",
        "message": "An internal error occurred",
        "internal": "user repository: get by id: connection refused"
    }
}
```

### 10.7 Error Logging

- Validation errors: **WARN** level
- Authentication/authorization errors: **WARN** level
- Not found errors: **DEBUG** level
- Conflict errors: **INFO** level
- Database/external errors: **ERROR** level with full error chain
- Internal errors: **ERROR** level with full error chain

### 10.8 Rules

1. Internal error details must never be exposed to API clients in production.
2. Every error from repository/service must be logged before returning to the handler.
3. Handlers must not log errors already logged by the service layer.
4. Use structured logging fields for error context.

---

## 11. Logging

### 11.1 Logger

Use Go's standard library `log/slog` with a custom handler for structured, leveled logging.

**Why slog:** It's stdlib since Go 1.21, zero external dependencies, structured by default, supports custom handlers. zerolog and zap are faster, but the performance difference is negligible for most APIs, and stdlib means one fewer dependency.

### 11.2 Log Levels

| Level | Color (Console) | Usage |
|-------|-----------------|-------|
| `FATAL` | RED | Application cannot continue, will exit |
| `ERROR` | RED | Something failed, requires attention |
| `WARN` | YELLOW | Unexpected but not critical |
| `INFO` | GREEN | Normal operational events |
| `DEBUG` | BLUE | Detailed diagnostic information |
| `TRACE` | GRAY | Extremely verbose, step-by-step |

### 11.3 Structured Log Fields

Every log entry includes:

```json
{
    "time": "2026-08-20T14:30:00.123Z",
    "level": "INFO",
    "msg": "Request completed",
    "service": "fiber-boilerplate",
    "request_id": "req_abc123",
    "trace_id": "trace_xyz789",
    "span_id": "span_def456",
    "client_ip": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "method": "POST",
    "path": "/api/v1/users",
    "status": 201,
    "latency_ms": 45.2,
    "file": "user/handler.go",
    "function": "Create",
    "line": 42
}
```

### 11.4 Source Identification

The logger captures source location using runtime caller information:

```go
func sourceInfo(depth int) (file string, function string, line int) {
    pc, file, line, ok := runtime.Caller(depth)
    if !ok {
        return "", "", 0
    }
    funcName := runtime.FuncForPC(pc).Name()
    return file, funcName, line
}
```

### 11.5 Output Targets

The logger writes to **both** console and file simultaneously.

#### Console Output
Human-readable, colorized output for development. Uses `slog.HandlerOptions` with a custom text handler.

#### File Output
JSON-formatted structured logs written to `.log` files.

### 11.6 Log Rotation

Implement using `lumberjack.v2`:

```go
lumberjackLogger := &lumberjack.Logger{
    Filename:   config.LogFile,      // e.g., "logs/app.log"
    MaxSize:    config.LogMaxSize,   // megabytes, e.g., 100
    MaxBackups: config.LogMaxBackups, // number of old files, e.g., 7
    MaxAge:     config.LogMaxAge,     // days, e.g., 14
    Compress:   config.LogCompress,   // gzip compression, e.g., true
}
```

#### Daily Rotation Strategy

Use **size-based rotation as primary** with daily log file naming:

```text
logs/
├── app-2026-08-20.log
├── app-2026-08-19.log
└── ...
```

When the daily file exceeds `MaxSize`, lumberjack creates a backup (e.g., `app-2026-08-20.log.1`).

### 11.7 Compression

**Recommendation: gzip**

For log rotation compression, use plain **gzip** (`.gz`), not `.tar.gz`:

| Format | Pros | Cons |
|--------|------|------|
| `.gz` | Simple, streaming, no temp files, standard on Linux, one file = one archive | Cannot archive multiple files |
| `.tar.gz` | Can archive multiple files, standard for tarballs | Overkill for single log files, requires temp file |

Since lumberjack compresses individual backup files, **gzip** is the right choice. Lumberjack uses gzip by default when `Compress: true`.

### 11.8 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Minimum log level (trace, debug, info, warn, error, fatal) |
| `LOG_FORMAT` | `json` | Log format (`json` or `console`) |
| `LOG_OUTPUT` | `both` | Output target (`stdout`, `file`, `both`) |
| `LOG_FILE` | `logs/app.log` | Log file path |
| `LOG_MAX_SIZE` | `100` | Max file size in MB before rotation |
| `LOG_MAX_BACKUPS` | `7` | Max number of old log files |
| `LOG_MAX_AGE` | `14` | Max age of old log files in days |
| `LOG_COMPRESS` | `true` | Compress rotated files with gzip |
| `SERVICE_NAME` | `fiber-boilerplate` | Service name included in every log |

### 11.9 Environment-Specific Behavior

| Setting | Development | Staging | Production |
|---------|-------------|---------|------------|
| `LOG_LEVEL` | `debug` | `info` | `info` |
| `LOG_FORMAT` | `console` | `json` | `json` |
| `LOG_OUTPUT` | `stdout` | `both` | `both` |
| `LOG_MAX_SIZE` | `50` | `100` | `200` |
| `LOG_MAX_BACKUPS` | `3` | `7` | `14` |
| `LOG_MAX_AGE` | `3` | `7` | `30` |
| `LOG_COMPRESS` | `false` | `true` | `true` |

### 11.10 Sensitive Data in Logs

The following must **never** appear in log output:

- Passwords (raw or hashed)
- JWT tokens
- Refresh tokens
- Cookies
- Authorization headers
- API keys
- Database connection strings (credentials portion)
- Private keys

Implement a `SensitiveString` type that masks values in log output:

```go
type SensitiveString struct {
    value string
}

func (s SensitiveString) String() string {
    return "[REDACTED]"
}

func (s SensitiveString) MarshalJSON() ([]byte, error) {
    return json.Marshal("[REDACTED]")
}
```

---

## 12. Authentication

### 12.1 JWT Token Architecture

Use **Access Token + Refresh Token** pattern:

| Token | Lifetime | Storage | Purpose |
|-------|----------|---------|---------|
| Access Token | 15 minutes | Client memory (variable, not localStorage) | API authentication |
| Refresh Token | 7 days | HttpOnly Secure cookie | Token renewal |

### 12.2 Signing

- Algorithm: **HS256** (HMAC-SHA256)
- Secret: Configurable via `JWT_SECRET` environment variable
- Minimum secret length: 32 bytes (256 bits)

### 12.3 Access Token Claims

```json
{
    "sub": "user_id",
    "iss": "fiber-boilerplate",
    "aud": "fiber-boilerplate",
    "iat": 1692547200,
    "exp": 1692548100,
    "nbf": 1692547200,
    "jti": "unique_token_id",
    "type": "access"
}
```

### 12.4 Refresh Token Claims

```json
{
    "sub": "user_id",
    "iss": "fiber-boilerplate",
    "aud": "fiber-boilerplate",
    "iat": 1692547200,
    "exp": 1693152000,
    "nbf": 1692547200,
    "jti": "unique_token_id",
    "type": "refresh",
    "family": "token_family_id"
}
```

### 12.5 Token Family

Each refresh token belongs to a **token family**. When a refresh token is used:

1. The old refresh token is deleted from the database.
2. A new refresh token is created with the **same family ID**.
3. The new refresh token is returned.

If a refresh token is reused (detected by `jti` already being deleted), **all tokens in that family are revoked**. This prevents refresh token theft from going undetected.

### 12.6 Clock Skew

Allow 30 seconds of clock skew when validating token `exp` and `nbf` claims.

### 12.7 Token Validation Flow

```text
Access Token Validation:
1. Extract from Authorization header: "Bearer <token>"
2. Parse and validate signature (HS256)
3. Validate issuer, audience, expiration
4. Check token type claim is "access"
5. Extract user_id from "sub" claim
6. Attach user to context

Refresh Token Validation:
1. Extract from HttpOnly cookie
2. Parse and validate signature (HS256)
3. Validate issuer, audience, expiration
4. Check token type claim is "refresh"
5. Look up token in database by jti
6. Verify token family matches
7. Check user still exists and is active
```

### 12.8 Refresh Token Database Storage

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    family_id UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
```

Store a **hash** of the refresh token, not the token itself. If the database is compromised, the attacker cannot use the tokens.

### 12.9 Secret Rotation

The JWT secret should be rotated periodically:

1. Support multiple active signing keys via `JWT_SECRET_PREVIOUS` environment variable.
2. During rotation period, validation checks both current and previous keys.
3. After all tokens issued with the old key have expired, remove the previous key.

### 12.10 Token Revocation

- **Logout:** Delete the refresh token from database, clear cookie.
- **Change password:** Delete all refresh tokens for the user.
- **Admin revoke:** Delete all refresh tokens for a user.
- **Family compromise:** Delete all tokens in the family.

---

## 13. Refresh Token Security

### 13.1 Cookie Configuration

```go
cookie := &fiber.Cookie{
    Name:     "refresh_token",
    Value:    refreshTokenString,
    Path:     "/api/v1/auth",
    MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds
    HTTPOnly: true,
    Secure:   true,
    SameSite: fiber.CookieSameSiteStrictMode,
}
```

### 13.2 Why HttpOnly

- Prevents JavaScript from accessing the token via `document.cookie`
- Mitigates XSS attacks — even if an attacker injects malicious script, they cannot steal the refresh token
- The refresh token is only sent automatically by the browser in HTTP requests to the specified path

### 13.3 Why SameSite=Strict

- Prevents CSRF attacks by not sending the cookie on cross-origin requests
- The refresh token is never sent when a user navigates to the site from an external link
- Trade-off: if the application needs to receive the cookie on navigations from external links, use `Lax` instead

**Recommendation:** Use `Strict` for maximum security. If OAuth callback flows from external providers are needed, use `Lax` for the refresh token path and implement CSRF protection separately.

### 13.4 CSRF Considerations

Since `SameSite=Strict` prevents CSRF for the refresh token, the main CSRF risk is on the **access token**. Access tokens are sent via the `Authorization` header, not cookies, so they are not automatically included in cross-origin requests.

**If access tokens are stored in cookies** (not recommended), CSRF protection is required. This boilerplate stores access tokens in application memory on the client side.

### 13.5 HTTPS Requirements

- `Secure: true` ensures the cookie is only sent over HTTPS
- In development, when running on `localhost`, set `Secure: false` via configuration
- The `.env.example` file must document this distinction

### 13.6 Token Reuse Detection

```text
Token A (family X) used to refresh
    → Token A deleted from DB
    → Token B (family X) created and returned

Later: Token A used again
    → DB lookup: Token A not found
    → REUSE DETECTED
    → Delete ALL tokens with family X
    → Return 401 Unauthorized
    → Log security event at WARN level
    → Invalidate user session
```

### 13.7 Logout Behavior

```text
POST /api/v1/auth/logout
    │
    ▼
Extract refresh token from cookie
    │
    ▼
Delete refresh token from database
    │
    ▼
Clear cookie from browser
    │
    ▼
Return 204 No Content
```

### 13.8 Client-Side Token Management

For browser-based clients:

```text
Access Token:
    → Store in memory (JavaScript variable)
    → Send via Authorization header
    → Not persisted anywhere
    → Lost on page refresh (re-authenticate via refresh token)

Refresh Token:
    → Stored in HttpOnly cookie
    → Sent automatically by browser to /api/v1/auth/* endpoints
    → Not accessible via JavaScript
```

**For non-browser API clients:** The refresh token can optionally be returned in the response body instead of a cookie. Document this as an alternative via a query parameter or header flag (e.g., `X-Token-Storage: body`).

---

## 14. RBAC Authorization

### 14.1 Model

Inspired by Laravel Spatie Permission:

```text
User ──M:N── Role ──M:N── Permission
```

### 14.2 Database Schema

```sql
-- Roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User-Role (many-to-many)
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Role-Permission (many-to-many)
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
```

### 14.3 Permission Naming Convention

```text
<resource>.<action>

Examples:
users.view
users.create
users.update
users.delete
roles.view
roles.create
roles.update
roles.delete
permissions.view
orders.view
orders.create
orders.process
```

### 14.4 Permission Checking API

```go
// In middleware
RequirePermission("users.create")

// In service layer
if !h.rbac.HasPermission(ctx, userID, "users.create") {
    return nil, domain.ErrForbidden
}

// Multiple permissions (any of)
RequireAnyPermission("users.view", "users.admin")

// All permissions required
RequireAllPermissions("users.view", "users.export")
```

### 14.5 Middleware Implementation

```go
func RequirePermission(permission string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := auth.GetUserID(c)
        if userID == "" {
            return response.Unauthorized(c, "Authentication required")
        }

        hasPermission, err := rbacService.HasPermission(c.Context(), userID, permission)
        if err != nil {
            return response.InternalError(c)
        }

        if !hasPermission {
            return response.Forbidden(c, "Insufficient permissions")
        }

        return c.Next()
    }
}
```

### 14.6 Service-Layer Authorization

Middleware alone is insufficient. Business logic must also enforce authorization:

```go
func (s *Service) DeleteUser(ctx context.Context, actorID, targetID string) error {
    if !s.rbac.HasPermission(ctx, actorID, "users.delete") {
        return domain.ErrForbidden
    }

    // Business rule: cannot delete yourself
    if actorID == targetID {
        return domain.NewAppError("BUSINESS_ERROR", "Cannot delete your own account", 422, nil)
    }

    return s.repo.Delete(ctx, targetID)
}
```

### 14.7 Caching Strategy

Permission lookups are cached in Redis to avoid repeated database queries:

```text
Cache key:  rbac:permissions:{user_id}
Cache value: JSON array of permission strings
TTL:        5 minutes
```

### 14.8 Cache Invalidation

| Event | Action |
|-------|--------|
| Role permissions updated | Invalidate all users with that role |
| User role changed | Invalidate that user's permissions |
| User deleted | Invalidate that user's permissions |
| Permission created/deleted | Invalidate all roles (rare, typically at setup) |

### 14.9 Super-Admin Behavior

A configurable super-admin role bypasses all permission checks:

```go
func (s *Service) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
    if s.isSuperAdmin(ctx, userID) {
        return true, nil
    }

    perms, err := s.cache.GetPermissions(ctx, userID)
    if err != nil || perms == nil {
        perms, err = s.repo.GetUserPermissions(ctx, userID)
        if err != nil {
            return false, fmt.Errorf("rbac: get permissions: %w", err)
        }
        s.cache.SetPermissions(ctx, userID, perms)
    }

    for _, p := range perms {
        if p == permission || p == "*" {
            return true, nil
        }
    }
    return false, nil
}
```

The super-admin role is defined by `SUPER_ADMIN_ROLE` configuration (default: `"super-admin"`).

---

## 15. OpenAPI / Swagger

### 15.1 Approach

**Code-first** with `swaggo/swag` annotations.

**Why code-first:**
- Annotations live directly above handler functions, always in sync with implementation
- Generated documentation is always up-to-date (generated at build time)
- No separate OpenAPI file to maintain
- Trade-off: annotations can be verbose, but they're self-documenting

**Alternatives considered:**
- **OpenAPI-first** (write spec, generate server): Better for API contract enforcement across teams, but adds ceremony and tooling complexity for a boilerplate
- **oapi-codegen**: Generates Go types from OpenAPI spec. Good, but adds a build step
- **Fern**: Commercial tooling, adds dependency

### 15.2 OpenAPI Version

OpenAPI 3.0.x (swaggo generates 3.0 by default)

### 15.3 Annotation Example

```go
// @Summary      Create a new user
// @Description  Creates a new user with the provided information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body CreateUserRequest true "User creation payload"
// @Success      201 {object} UserResponse
// @Failure      422 {object} ValidationErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/users [post]
// @Security     BearerAuth
```

### 15.4 Authentication Scheme

```go
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"
```

### 15.5 Error Schemas

```go
// @model ErrorResponse
type ErrorResponse struct {
    Success bool   `json:"success" example:"false"`
    Error struct {
        Code    string `json:"code" example:"NOT_FOUND"`
        Message string `json:"message" example:"Resource not found"`
    } `json:"error"`
}

// @model ValidationErrorResponse
type ValidationErrorResponse struct {
    Success bool   `json:"success" example:"false"`
    Error struct {
        Code    string              `json:"code" example:"VALIDATION_ERROR"`
        Message string              `json:"message" example:"Request validation failed"`
        Fields  map[string][]string `json:"fields"`
    } `json:"error"`
}
```

### 15.6 Generation Workflow

```bash
swag init -g cmd/server/main.go -o docs/swagger
```

This generates:
- `docs/swagger/docs.go`
- `docs/swagger/swagger.json`
- `docs/swagger/swagger.yaml`

Swagger UI is served at `/swagger/index.html` in development mode only.

### 15.7 CI Validation

```bash
swag init -g cmd/server/main.go -o docs/swagger
git diff --exit-code docs/swagger/  # fail if generated files changed
```

---

## 16. Rate Limiting

### 16.1 Two Modes

#### Distributed Mode (Redis)

```text
App Instance 1 ─┐
App Instance 2 ─┼── Redis (shared counter)
App Instance 3 ─┘
```

Use `ulule/limiter` with Redis store. Counter is shared across all instances.

#### Local Mode (In-Memory)

```text
App Instance ── In-memory counter
```

Use `ulule/limiter` with in-memory store. Counter is per-instance only.

**Warning:** In distributed deployments with local rate limiting, each instance allows the full rate limit independently. A deployment with 3 instances effectively allows 3x the intended rate limit per client. This is documented and configurable.

### 16.2 Rate Limit Algorithm

Token bucket algorithm (default for `ulule/limiter`).

### 16.3 Default Limits

| Route Pattern | Rate | Burst | Notes |
|---------------|------|-------|-------|
| `POST /api/v1/auth/login` | 5/min | 5 | Prevent brute force |
| `POST /api/v1/auth/register` | 3/min | 3 | Prevent spam |
| `POST /api/v1/auth/refresh` | 10/min | 10 | Prevent token refresh abuse |
| `GET /api/v1/*` | 100/min | 100 | General API read |
| `POST /api/v1/*` | 30/min | 30 | General API write |
| `GET /metrics` | 30/min | 30 | Metrics endpoint |
| `GET /health/*` | 60/min | 60 | Health checks |

### 16.4 Key Design

**Per-IP (unauthenticated):**
```text
ratelimit:{route}:{ip_hash}
```

**Per-User (authenticated):**
```text
ratelimit:{route}:{user_id}
```

### 16.5 Redis Key TTL

TTL = rate window + 10% buffer. For 1-minute window, TTL = 66 seconds.

### 16.6 Failure Behavior

| Scenario | Behavior |
|----------|----------|
| Redis available | Use Redis rate limiter |
| Redis unavailable (local mode) | Fall back to in-memory rate limiter |
| Redis unavailable (distributed mode) | Log warning, allow all requests (fail open) |

**Fail open** is critical. Rate limiting should never cause downtime.

### 16.7 Rate Limit Response

```json
{
    "success": false,
    "error": {
        "code": "RATE_LIMITED",
        "message": "Rate limit exceeded. Try again in 45 seconds."
    }
}
```

Headers:
```text
X-RateLimit-Limit: 30
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1692547260
Retry-After: 45
```

### 16.8 Memory Cleanup

In-memory rate limiters must clean up expired entries periodically. `ulule/limiter` handles this internally with a built-in cleanup goroutine.

---

## 17. CORS

### 17.1 Configuration

```go
cors.Config{
    AllowOrigins:     "https://example.com,https://admin.example.com",
    AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
    ExposeHeaders:    "X-Request-ID,X-RateLimit-Limit,X-RateLimit-Remaining",
    AllowCredentials: true,
    MaxAge:           86400, // 24 hours
}
```

### 17.2 Rules

1. **Never use `AllowOrigins: "*"` when `AllowCredentials: true`.** This is forbidden by the CORS specification.
2. Use explicit origin whitelist.
3. Configure per-environment via `CORS_ALLOWED_ORIGINS` environment variable (comma-separated).
4. `OPTIONS` preflight requests must not require authentication.

### 17.3 Environment Configuration

| Environment | Allowed Origins |
|-------------|----------------|
| Development | `http://localhost:3000,http://localhost:5173` |
| Staging | `https://staging.example.com` |
| Production | `https://example.com,https://admin.example.com` |

---

## 18. Prometheus Metrics

### 18.1 Endpoint

```text
GET /metrics
```

### 18.2 Middleware Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | method, path, status | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | method, path | Request latency |
| `http_requests_in_flight` | Gauge | — | Currently active requests |
| `http_request_size_bytes` | Histogram | method | Request body size |
| `http_response_size_bytes` | Histogram | method | Response body size |

### 18.3 Application Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `app_info` | Gauge | version, go_version | Application metadata |
| `app_uptime_seconds` | Gauge | — | Application uptime |
| `app_errors_total` | Counter | code | Application errors by code |

### 18.4 Database Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `db_connections_open` | Gauge | driver | Open database connections |
| `db_connections_in_use` | Gauge | driver | Connections in use |
| `db_connections_idle` | Gauge | driver | Idle connections |
| `db_query_duration_seconds` | Histogram | driver | Query execution time |

### 18.5 Redis Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `redis_connections_open` | Gauge | — | Open Redis connections |
| `redis_operation_duration_seconds` | Histogram | operation | Redis operation latency |
| `redis_errors_total` | Counter | operation | Redis operation errors |

### 18.6 Go Runtime Metrics

`prometheus/client_golang` automatically registers Go runtime collectors:
- `go_goroutines`
- `go_memstats_alloc_bytes`
- `go_gc_duration_seconds`
- `process_cpu_seconds_total`
- `process_resident_memory_bytes`

### 18.7 Safe vs Unsafe Labels

**Safe labels (low cardinality):**
- `method`: GET, POST, PUT, PATCH, DELETE (bounded set)
- `status`: 200, 201, 400, 401, 403, 404, 429, 500 (bounded set)
- `path`: API versioned routes (`/api/v1/users`) — use route templates, not actual paths with IDs

**Unsafe labels (high cardinality — NEVER use):**
- `user_id` — unbounded, creates a new time series per user
- `request_id` — unique per request
- `query_string` — unbounded
- `ip_address` — unbounded
- `authorization` — unbounded, sensitive

### 18.8 Path Normalization

Normalize paths before using as metric labels:

```text
/api/v1/users/abc-123-def     → /api/v1/users/:id
/api/v1/users/abc-123-def/roles → /api/v1/users/:id/roles
```

---

## 19. OpenTelemetry Distributed Tracing

### 19.1 Architecture

```text
HTTP Request
    ↓
Fiber → OTel HTTP middleware (creates root span)
    ↓
Handler → creates child span
    ↓
Service → creates child span
    ↓
Repository → creates child span (DB query)
    ↓
OTLP Exporter → Collector → Backend (Jaeger, Tempo, etc.)
```

### 19.2 Automatic Instrumentation

Use `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` and Fiber OTel middleware for automatic span creation on HTTP requests.

### 19.3 Manual Instrumentation

Create spans for operations not covered by automatic instrumentation:

```go
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
    ctx, span := tracer.Start(ctx, "repository.user.get_by_id",
        trace.WithAttributes(
            attribute.String("user.id", id),
        ),
    )
    defer span.End()

    // ... database query ...

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    return user, nil
}
```

### 19.4 Span Definitions

| Span Name | Layer | Attributes |
|-----------|-------|------------|
| `http.request` | Fiber middleware | method, path, status |
| `handler.<feature>.<operation>` | Handler | feature, operation |
| `service.<feature>.<operation>` | Service | feature, operation |
| `repository.<entity>.<operation>` | Repository | entity, operation, db.system |
| `redis.<operation>` | Cache | operation, key pattern |
| `http.external.<host>` | External call | host, method, path |

### 19.5 Context Propagation

Trace context is propagated via `context.Context`:

```go
ctx = otel.TraceContextWithSpanContext(ctx, span.SpanContext())
user, err := service.GetUser(ctx, id)
```

### 19.6 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_ENDPOINT` | `localhost:4318` | OTLP HTTP collector endpoint |
| `OTEL_SERVICE_NAME` | `fiber-boilerplate` | Service name in traces |
| `OTEL_ENVIRONMENT` | `development` | Environment tag |
| `OTEL_SAMPLE_RATE` | `1.0` (dev), `0.1` (prod) | Sample rate (0.0 to 1.0) |

### 19.7 Sampling Strategy

- **Development:** Sample all requests (`sample_rate: 1.0`)
- **Staging:** Sample 50% of requests (`sample_rate: 0.5`)
- **Production:** Sample 10% of requests, always sample errors and slow requests (`sample_rate: 0.1`)

### 19.8 Error Recording

```go
if err != nil {
    span.RecordError(err,
        trace.WithAttributes(
            attribute.String("error.type", reflect.TypeOf(err).String()),
        ),
    )
    span.SetStatus(codes.Error, err.Error())
}
```

### 19.9 Sensitive Data in Spans

**Must NOT include:**
- Passwords, JWT tokens, refresh tokens, cookies, authorization headers
- API keys, database connection strings (credentials)
- Personal identifiable information (PII) beyond entity IDs

**Safe to include:**
- Entity IDs (user_id, order_id)
- HTTP method and status
- Query names and patterns
- Error types (not full error messages)
- Latency

### 19.10 Graceful Shutdown

```go
func (t *Tracer) Shutdown(ctx context.Context) error {
    return t.tracerProvider.Shutdown(ctx)
}
```

Called during application shutdown with a timeout context. Ensures all pending spans are exported before exit.

---

## 20. Health Checks

### 20.1 Endpoints

| Endpoint | Purpose | Dependencies Checked |
|----------|---------|---------------------|
| `GET /health/live` | Liveness probe | None (always 200 if process is running) |
| `GET /health/ready` | Readiness probe | Database, Redis (if configured) |
| `GET /health` | Combined health | All dependencies |

### 20.2 Liveness vs Readiness

**Liveness (`/health/live`):**
- Indicates the process is alive and can accept traffic
- Should always return 200 if the process is running
- Used by orchestrators to determine if the container should be restarted
- Must NOT check external dependencies

**Readiness (`/health/ready`):**
- Indicates the application is ready to serve requests
- Checks critical dependencies (database, Redis)
- Used by orchestrators to determine if the container should receive traffic
- Can return 503 if a critical dependency is unavailable

### 20.3 Response Format

```json
{
    "status": "healthy",
    "checks": {
        "database": {
            "status": "healthy",
            "latency_ms": 2
        },
        "redis": {
            "status": "healthy",
            "latency_ms": 1
        }
    }
}
```

### 20.4 Rules

- Health endpoints must not expose sensitive information
- Health endpoints must be accessible without authentication
- Health endpoints must have low overhead (simple ping queries)
- Health endpoints should be on a separate port (configurable) to avoid rate limiting

---

## 21. Database Architecture

### 21.1 Multi-Database Support

| Database | Driver Package | Notes |
|----------|---------------|-------|
| PostgreSQL | `github.com/jackc/pgx/v5` | Primary target, most features |
| MySQL | `github.com/go-sql-driver/mysql` | Full support |
| MariaDB | `github.com/go-sql-driver/mysql` | Uses MySQL driver |
| SQLite | `github.com/mattn/go-sqlite3` | Dev/testing only, CGO required |

### 21.2 Driver Selection

```go
func NewDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
    var db *sql.DB
    var err error

    switch cfg.Driver {
    case "postgres":
        db, err = sql.Open("pgx", cfg.URL)
    case "mysql", "mariadb":
        db, err = sql.Open("mysql", cfg.URL)
    case "sqlite":
        db, err = sql.Open("sqlite3", cfg.URL+"?_journal_mode=WAL")
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
    }

    if err != nil {
        return nil, fmt.Errorf("database open: %w", err)
    }

    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("database ping: %w", err)
    }

    return db, nil
}
```

### 21.3 Connection Pool Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `MaxOpenConns` | 25 | Maximum open connections |
| `MaxIdleConns` | 10 | Maximum idle connections |
| `ConnMaxLifetime` | 5 minutes | Maximum connection lifetime |
| `ConnMaxIdleTime` | 3 minutes | Maximum idle connection time |

### 21.4 Query Timeout

All database queries must respect context cancellation:

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

err := r.db.QueryRowContext(ctx, query, args...).Scan(&dest)
```

### 21.5 Transaction Helper

```go
func WithTransaction(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }

    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        }
    }()

    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("rollback: %v (original: %w)", rbErr, err)
        }
        return err
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit: %w", err)
    }

    return nil
}
```

### 21.6 Migration Strategy

Use `golang-migrate/migrate` for versioned SQL migrations.

**Migration naming convention:**
```text
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_roles.up.sql
├── 000002_create_roles.down.sql
└── ...
```

**Rules:**
- Migration files are numbered sequentially with zero-padded 6-digit prefixes
- Each migration has an `.up.sql` and `.down.sql` counterpart
- Migrations are **immutable** after they have been applied to any environment
- Never modify an applied migration — create a new one instead
- Use raw SQL, not ORM-generated schema
- Migrations must be idempotent where possible (use `IF NOT EXISTS`)
- Down migrations must be provided for rollback capability

### 21.7 Database-Specific Differences

| Feature | PostgreSQL | MySQL/MariaDB | SQLite |
|---------|------------|---------------|--------|
| UUID generation | `gen_random_uuid()` | Application-generated | Application-generated |
| JSON column | `jsonb` | `json` | `text` |
| Array type | Native arrays | Not supported | Not supported |
| Full-text search | `tsvector` | `FULLTEXT` | FTS5 |
| `UPSERT` | `ON CONFLICT DO UPDATE` | `ON DUPLICATE KEY UPDATE` | `ON CONFLICT DO UPDATE` |
| `RETURNING` clause | Supported | Not supported | Supported |

When database-specific features are needed, document which database the migration targets.

---

## 22. Repository Pattern

### 22.1 Interface Design

Repository interfaces are designed around **business needs**, not generic CRUD:

```go
type Repository interface {
    Create(ctx context.Context, user *Domain) (*Domain, error)
    GetByID(ctx context.Context, id string) (*Domain, error)
    GetByEmail(ctx context.Context, email string) (*Domain, error)
    List(ctx context.Context, query ListQuery) ([]*Domain, int, error)
    Update(ctx context.Context, user *Domain) (*Domain, error)
    Delete(ctx context.Context, id string) error
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}
```

### 22.2 Implementation

```go
type postgresRepository struct {
    db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
    return &postgresRepository{db: db}
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*Domain, error) {
    ctx, span := tracer.Start(ctx, "repository.user.get_by_id")
    defer span.End()

    query := `
        SELECT id, email, name, password_hash, created_at, updated_at
        FROM users
        WHERE id = $1`

    var user Domain
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID,
        &user.Email,
        &user.Name,
        &user.PasswordHash,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domain.ErrNotFound
        }
        return nil, fmt.Errorf("user repository: get by id: %w", err)
    }

    return &user, nil
}
```

### 22.3 Rules

1. Repository interfaces live in the feature package (e.g., `internal/user/repository.go`).
2. Implementations live alongside the interface or in a database-specific sub-package for multi-DB scenarios.
3. Repositories must not contain business logic.
4. Repositories handle database-specific concerns (query building, scanning).
5. Do not create generic `Repository[T]` interfaces unless all entities share the exact same access patterns.
6. Each repository method creates its own tracing span.
7. All errors from `database/sql` are wrapped with context before returning.

### 22.4 Multi-Database Implementations

```text
internal/user/
├── repository.go          # Interface definition
├── model.go               # Domain model
├── postgres_repository.go  # PostgreSQL implementation
├── mysql_repository.go     # MySQL implementation
└── sqlite_repository.go    # SQLite implementation
```

Select implementation at runtime based on configuration:

```go
func NewRepository(db *sql.DB, driver string) user.Repository {
    switch driver {
    case "postgres":
        return user.NewPostgresRepository(db)
    case "mysql", "mariadb":
        return user.NewMySQLRepository(db)
    case "sqlite":
        return user.NewSQLiteRepository(db)
    default:
        panic("unsupported driver: " + driver)
    }
}
```

---

## 23. Transactions

### 23.1 Transaction Boundaries

Transactions are managed at the **service layer**. Repositories do not manage transactions — they receive a `*sql.Tx` or `*sql.DB` and operate accordingly.

### 23.2 Transaction-Aware Repository

```go
type txKey struct{}

func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
    return context.WithValue(ctx, txKey{}, tx)
}

func (r *postgresRepository) exec(ctx context.Context) DBExecutor {
    if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
        return tx
    }
    return r.db
}
```

### 23.3 Service-Layer Transaction Example

```go
func (s *Service) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
    return database.WithTransaction(ctx, s.db, func(tx *sql.Tx) error {
        txCtx := database.WithTx(ctx, tx)

        if err := s.orderRepo.Create(txCtx, order); err != nil {
            return fmt.Errorf("create order: %w", err)
        }

        for _, item := range order.Items {
            if err := s.orderItemRepo.Create(txCtx, item); err != nil {
                return fmt.Errorf("create order item: %w", err)
            }
        }

        if err := s.inventoryRepo.DecrementStock(txCtx, order.Items); err != nil {
            return fmt.Errorf("decrement stock: %w", err)
        }

        return nil
    })
}
```

### 23.4 Rules

1. Always use `WithTransaction` from `internal/database/transaction.go`.
2. Never call `tx.Begin()` or `tx.Commit()` directly in service code.
3. Transaction context must be passed to all repositories within the transaction.
4. If any operation fails, the entire transaction is rolled back.
5. Transactions should be as short as possible to reduce lock contention.
6. Do not make external HTTP calls within a transaction.

---

## 24. Configuration

### 24.1 Configuration Source

Environment variables, loaded via `envconfig`. Local development uses `.env` files.

### 24.2 Complete Configuration Schema

```go
type Config struct {
    // Application
    AppEnv       string `env:"APP_ENV" default:"development"`
    AppName      string `env:"APP_NAME" default:"fiber-boilerplate"`
    AppVersion   string `env:"APP_VERSION" default:"dev"`
    HTTPHost     string `env:"HTTP_HOST" default:"0.0.0.0"`
    HTTPPort     int    `env:"HTTP_PORT" default:"8080"`

    // Database
    DatabaseDriver    string        `env:"DATABASE_DRIVER" default:"postgres"`
    DatabaseURL       string        `env:"DATABASE_URL" required:"true"`
    DBMaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" default:"25"`
    DBMaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" default:"10"`
    DBConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" default:"5m"`
    DBConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" default:"3m"`

    // Redis
    RedisURL string `env:"REDIS_URL" default:""`

    // JWT
    JWTSecret      string        `env:"JWT_SECRET" required:"true"`
    JWTAccessTTL   time.Duration `env:"JWT_ACCESS_TTL" default:"15m"`
    JWTRefreshTTL  time.Duration `env:"JWT_REFRESH_TTL" default:"168h"`

    // CORS
    CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" default:"http://localhost:3000"`

    // Logging
    LogLevel      string `env:"LOG_LEVEL" default:"info"`
    LogFormat     string `env:"LOG_FORMAT" default:"json"`
    LogOutput     string `env:"LOG_OUTPUT" default:"both"`
    LogFile       string `env:"LOG_FILE" default:"logs/app.log"`
    LogMaxSize    int    `env:"LOG_MAX_SIZE" default:"100"`
    LogMaxBackups int    `env:"LOG_MAX_BACKUPS" default:"7"`
    LogMaxAge     int    `env:"LOG_MAX_AGE" default:"14"`
    LogCompress   bool   `env:"LOG_COMPRESS" default:"true"`

    // OpenTelemetry
    OTELEndpoint    string  `env:"OTEL_ENDPOINT" default:"localhost:4318"`
    OTELServiceName string  `env:"OTEL_SERVICE_NAME" default:"fiber-boilerplate"`
    OTELEnvironment string  `env:"OTEL_ENVIRONMENT" default:"development"`
    OTelSampleRate  float64 `env:"OTEL_SAMPLE_RATE" default:"1.0"`

    // Rate Limiting
    RateLimitEnabled      bool `env:"RATE_LIMIT_ENABLED" default:"true"`
    RateLimitRedisEnabled bool `env:"RATE_LIMIT_REDIS_ENABLED" default:"false"`

    // RBAC
    SuperAdminRole string `env:"SUPER_ADMIN_ROLE" default:"super-admin"`
}
```

### 24.3 .env.example

```bash
# Application
APP_ENV=development
APP_NAME=fiber-boilerplate
APP_VERSION=dev
HTTP_HOST=0.0.0.0
HTTP_PORT=8080

# Database
DATABASE_DRIVER=postgres
DATABASE_URL=postgres://postgres:postgres@localhost:5432/fiber_boilerplate?sslmode=disable

# Redis (optional for local development)
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=change-me-to-a-secure-random-string-at-least-32-bytes
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Logging
LOG_LEVEL=debug
LOG_FORMAT=console
LOG_OUTPUT=stdout
LOG_FILE=logs/app.log
LOG_MAX_SIZE=100
LOG_MAX_BACKUPS=7
LOG_MAX_AGE=14
LOG_COMPRESS=false

# OpenTelemetry
OTEL_ENDPOINT=localhost:4318
OTEL_SERVICE_NAME=fiber-boilerplate
OTEL_ENVIRONMENT=development
OTEL_SAMPLE_RATE=1.0

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REDIS_ENABLED=false

# RBAC
SUPER_ADMIN_ROLE=super-admin
```

### 24.4 Secrets

- `JWT_SECRET` must never be hardcoded or committed to version control.
- Use `.env` for local development (gitignored).
- Use environment variables or secret managers (Vault, AWS Secrets Manager) in production.
- The `.env.example` file must contain placeholder values, not real secrets.

---

## 25. Graceful Shutdown

### 25.1 Shutdown Sequence

```text
SIGINT / SIGTERM received
    │
    ▼
1. Stop accepting new HTTP requests (Fiber.Shutdown)
    │
    ▼
2. Wait for active requests to complete (with timeout)
    │
    ▼
3. Shutdown background workers (if any)
    │
    ▼
4. Flush OpenTelemetry spans and metrics
    │
    ▼
5. Close Redis connection
    │
    ▼
6. Close database connection pool
    │
    ▼
7. Flush and close log files
    │
    ▼
8. Exit with code 0
```

### 25.2 Implementation

```go
func main() {
    cfg := config.Load()
    app := server.New(cfg)

    go func() {
        if err := app.Listen(fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)); err != nil {
            log.Error("server error", "error", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    shutdownSteps := []struct {
        name string
        fn   func(ctx context.Context) error
    }{
        {"fiber", func(ctx context.Context) error { return app.Shutdown() }},
        {"tracing", tracing.Shutdown},
        {"metrics", metrics.Flush},
        {"redis", cache.Close},
        {"database", database.Close},
        {"logger", logging.Close},
    }

    for _, step := range shutdownSteps {
        log.Info("shutting down", "component", step.name)
        if err := step.fn(ctx); err != nil {
            log.Error("shutdown error", "component", step.name, "error", err)
        }
    }

    log.Info("server stopped")
}
```

### 25.3 Configurable Timeout

The shutdown timeout is configurable via `SHUTDOWN_TIMEOUT` environment variable (default: `30s`).

---

## 26. API Versioning

### 26.1 Strategy

URL path-based versioning:

```text
/api/v1/users
/api/v1/auth/login
/api/v2/users (when breaking changes occur)
```

### 26.2 Version Prefix

All API routes are mounted under `/api/v{version}`.

### 26.3 Breaking Changes

When a breaking change is introduced:

1. Create a new version (e.g., `/api/v2/`).
2. Maintain the old version for a deprecation period.
3. Return `Sunset` header and `Deprecation` header on deprecated version endpoints.
4. Document migration guide.

### 26.4 Non-Breaking Changes

Non-breaking changes (new fields in response, new optional request fields, new endpoints) are added to the current version without incrementing.

---

## 27. Pagination

### 27.1 Query Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `page` | `1` | Page number (1-indexed) |
| `per_page` | `20` | Items per page (max: 100) |
| `sort` | `created_at` | Sort field |
| `order` | `desc` | Sort direction (`asc` or `desc`) |
| `search` | — | Full-text search query (feature-specific) |

### 27.2 Response Format

```json
{
    "data": [
        {"id": "abc", "name": "Alice"},
        {"id": "def", "name": "Bob"}
    ],
    "meta": {
        "page": 1,
        "per_page": 20,
        "total": 150,
        "total_pages": 8
    }
}
```

### 27.3 Cursor-Based Pagination

Evaluate cursor-based pagination for datasets exceeding 10,000 records or for real-time feeds. For the boilerplate, offset-based pagination is sufficient.

Cursor-based pagination can be added as an extension:

```text
GET /api/v1/users?cursor=abc123&per_page=20
```

Response:

```json
{
    "data": [...],
    "meta": {
        "next_cursor": "def456",
        "has_more": true
    }
}
```

---

## 28. Testing Strategy

### 28.1 Unit Tests

**Location:** Same package as the code being tested (`*_test.go` files).

**What to test:**
- Service/business logic with mock repositories
- Validation logic
- JWT generation and validation
- Error handling
- Utility functions
- DTO mapping

**What NOT to mock unnecessarily:**
- Standard library functions
- Simple data transformations
- Configuration loading (use real config in tests)

**Mock strategy:** Use interface-based mocks. For simple cases, hand-written fakes. For complex cases, use `mockgen` or `moq`.

```go
type mockRepository struct {
    createFn  func(ctx context.Context, user *Domain) (*Domain, error)
    getByIDFn func(ctx context.Context, id string) (*Domain, error)
}

func (m *mockRepository) Create(ctx context.Context, user *Domain) (*Domain, error) {
    return m.createFn(ctx, user)
}

func TestService_Create(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        repo := &mockRepository{
            createFn: func(ctx context.Context, user *Domain) (*Domain, error) {
                user.ID = "generated-id"
                return user, nil
            },
        }
        svc := NewService(repo, nil, nil)

        user, err := svc.Create(context.Background(), &Domain{
            Email: "test@example.com",
            Name:  "Test User",
        })

        require.NoError(t, err)
        assert.Equal(t, "generated-id", user.ID)
    })
}
```

### 28.2 E2E Tests

**Location:** `tests/e2e/`

**Infrastructure:** Testcontainers for real database and Redis instances.

```go
func TestMain(m *testing.M) {
    pgContainer, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready").WithStartupTimeout(30*time.Second),
        ),
    )

    redisContainer, err := redis.Run(ctx,
        "redis:7-alpine",
    )

    // Run migrations, start application, run tests, cleanup
    os.Exit(code)
}
```

**Test scenarios:**

```text
Auth:
    POST /api/v1/auth/register — success, validation error, duplicate email
    POST /api/v1/auth/login — success, wrong password, nonexistent user
    POST /api/v1/auth/refresh — success, expired token, revoked token, reuse detection
    POST /api/v1/auth/logout — success, clear cookie

User:
    GET /api/v1/users — list with pagination, filtering, sorting
    POST /api/v1/users — success, validation error, unauthorized, forbidden
    GET /api/v1/users/:id — success, not found
    PATCH /api/v1/users/:id — success, partial update, not found
    DELETE /api/v1/users/:id — success, not found, cannot delete self

RBAC:
    Access admin-only endpoint with admin role — success
    Access admin-only endpoint with user role — forbidden
    Access user endpoint with no auth — unauthorized

Rate Limiting:
    Exceed rate limit — 429 response with Retry-After header

Health:
    GET /health/live — 200
    GET /health/ready — 200 with dependency status
```

### 28.3 Test Naming Convention

```go
func TestService_Create(t *testing.T) {
    t.Run("success_creates_user_and_returns_domain", func(t *testing.T) {})
    t.Run("error_duplicate_email_returns_conflict", func(t *testing.T) {})
    t.Run("error_invalid_email_returns_validation_error", func(t *testing.T) {})
}
```

Pattern: `Test_<Type>_<Method>` for test functions, `<scenario>_<expected_behavior>` for subtests.

---

## 29. Security

### 29.1 Baseline Security Requirements

| Concern | Implementation |
|---------|---------------|
| JWT security | HS256 with 256-bit secret, short-lived access tokens, refresh token rotation |
| Password hashing | bcrypt with cost 12 |
| Cookie security | HttpOnly, Secure, SameSite=Strict |
| CORS | Explicit origin whitelist, no wildcard with credentials |
| CSRF | SameSite=Strict cookies, Authorization header for access tokens |
| Rate limiting | Per-IP and per-user rate limits, authentication endpoint throttling |
| SQL injection | Parameterized queries via `sqlx`/`database/sql` |
| Input validation | Struct-based validation with `validator/v10` |
| Security headers | X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, Content-Security-Policy, Strict-Transport-Security |
| Request size limits | Max request body size: 10MB (configurable) |
| Timeout | HTTP request timeout: 30s, idle timeout: 120s |
| Secret management | Environment variables, never hardcoded, never logged |
| Error leakage | Internal errors return generic messages in production |
| Logging sensitive data | SensitiveString type, structured redaction |
| Tracing sensitive data | No PII in span attributes |
| Dependency scanning | `govulncheck` in CI, Dependabot for dependency updates |

### 29.2 Security Headers Middleware

```go
func SecurityHeaders() fiber.Handler {
    return func(c *fiber.Ctx) error {
        c.Set("X-Content-Type-Options", "nosniff")
        c.Set("X-Frame-Options", "DENY")
        c.Set("X-XSS-Protection", "0")
        c.Set("Content-Security-Policy", "default-src 'self'")
        c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
        if c.Get("X-Forwarded-Proto") == "https" || c.IsHTTPS() {
            c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
        }
        return c.Next()
    }
}
```

### 29.3 Request Size Limit

```go
app.Use(func(c *fiber.Ctx) error {
    c.Context().SetBodyMaxSize(10 * 1024 * 1024) // 10MB
    return c.Next()
})
```

---

## 30. Developer Experience

### 30.1 Task Runner

Use `Taskfile.yml` (go-task) over Makefile.

**Why Taskfile:**
- Native Go tool, installed via `go install github.com/go-task/task/v3/cmd/task@latest`
- YAML format, more readable than Makefile syntax
- Better dependency management between tasks
- Cross-platform (Linux, macOS, Windows)

### 30.2 Available Commands

```yaml
version: "3"

tasks:
  dev:
    desc: "Start development environment"
    cmds:
      - docker compose up -d postgres redis
      - sleep 2
      - task migrate-up
      - go run cmd/server/main.go

  build:
    desc: "Build the application"
    cmds:
      - go build -o bin/server cmd/server/main.go

  test:
    desc: "Run all tests"
    cmds:
      - go test ./...

  test-unit:
    desc: "Run unit tests only"
    cmds:
      - go test ./internal/...

  test-e2e:
    desc: "Run E2E tests"
    cmds:
      - go test ./tests/e2e/...

  test-coverage:
    desc: "Run tests with coverage report"
    cmds:
      - go test -coverprofile=coverage.out ./...
      - go tool cover -html=coverage.out -o coverage.html

  lint:
    desc: "Run linter"
    cmds:
      - golangci-lint run

  fmt:
    desc: "Format code"
    cmds:
      - gofmt -s -w .
      - goimports -w .

  migrate-up:
    desc: "Run pending migrations"
    cmds:
      - go run cmd/server/main.go migrate up

  migrate-down:
    desc: "Rollback last migration"
    cmds:
      - go run cmd/server/main.go migrate down 1

  migrate-create:
    desc: "Create new migration"
    cmds:
      - ./scripts/migrate-create.sh {{.CLI_ARGS}}

  swagger:
    desc: "Generate OpenAPI docs"
    cmds:
      - swag init -g cmd/server/main.go -o docs/swagger

  docker-up:
    desc: "Start all Docker services"
    cmds:
      - docker compose up -d

  docker-down:
    desc: "Stop all Docker services"
    cmds:
      - docker compose down

  docker-build:
    desc: "Build production Docker image"
    cmds:
      - docker build -t fiber-boilerplate .

  generate:
    desc: "Run go generate"
    cmds:
      - go generate ./...

  vuln:
    desc: "Check for known vulnerabilities"
    cmds:
      - govulncheck ./...
```

### 30.3 Docker Compose (Local Development)

```yaml
services:
  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: fiber_boilerplate
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    ports:
      - "4318:4318"
    volumes:
      - ./otel-collector-config.yaml:/etc/otelcol-contrib/config.yaml

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    volumes:
      - grafana_data:/var/lib/grafana

  swagger-ui:
    image: swaggerapi/swagger-ui:latest
    ports:
      - "8081:8080"
    environment:
      SWAGGER_JSON: /docs/swagger.json
    volumes:
      - ./docs/swagger:/usr/share/nginx/html/docs

volumes:
  postgres_data:
  redis_data:
  grafana_data:
```

### 30.4 New Developer Onboarding

```bash
# 1. Clone the repository
git clone https://github.com/arman/fiber-boilerplate.git

# 2. Install dependencies
go mod download
go install github.com/go-task/task/v3/cmd/task@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest

# 3. Copy environment file
cp .env.example .env

# 4. Start infrastructure
task docker-up

# 5. Run migrations
task migrate-up

# 6. Start development server
task dev

# 7. Open Swagger UI
open http://localhost:8081
```

---

## 31. Docker

### 31.1 Production Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server cmd/server/main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -s /bin/sh -D appuser

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

RUN mkdir -p /app/logs && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

ENTRYPOINT ["/app/server"]
```

### 31.2 Image Optimization

- **Multi-stage build:** Build in `golang:1.22-alpine`, run in `alpine:3.19`
- **Non-root user:** Runs as `appuser:appgroup` (UID 1001)
- **CGO_ENABLED=0:** Static binary, no C dependencies
- **`-ldflags="-s -w"`:** Strip debug info and symbol table, smaller binary
- **Minimal base image:** Alpine (~5MB) vs Debian (~120MB)
- **No secrets in image:** All configuration via environment variables

### 31.3 Deployment Considerations

| Scenario | Consideration |
|----------|---------------|
| Single instance | Local rate limiting is fine, in-memory state is sufficient |
| Multiple instances | Redis required for distributed rate limiting, shared refresh token storage |
| Container orchestration | Health checks configured, graceful shutdown via SIGTERM, stateless design |
| Serverless/Cloud Run | In-memory rate limiting scales per instance, cold start considerations |

---

## 32. CI/CD

### 32.1 Pipeline

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4

  test-unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Unit tests
        run: go test -race -coverprofile=coverage.out ./internal/...

  test-e2e:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: E2E tests
        run: go test -race ./tests/e2e/...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable
          REDIS_URL: redis://localhost:6379

  build:
    runs-on: ubuntu-latest
    needs: [lint, test-unit, test-e2e]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Build
        run: go build -o /dev/null cmd/server/main.go

  swagger-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Install swag
        run: go install github.com/swaggo/swag/cmd/swag@latest
      - name: Generate docs
        run: swag init -g cmd/server/main.go -o docs/swagger
      - name: Check for changes
        run: git diff --exit-code docs/swagger/

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - name: Run govulncheck
        run: govulncheck ./...

  docker:
    runs-on: ubuntu-latest
    needs: [lint, test-unit, test-e2e, build]
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker image
        run: docker build -t fiber-boilerplate:${{ github.sha }} .
```

### 32.2 Pipeline Order

```text
Format check → Lint → Unit Test → E2E Test → Build → Security Scan → Docker Build
```

All steps must pass before Docker build. E2E tests use GitHub Actions service containers for PostgreSQL and Redis.

---

## 33. Example: User Feature

### 33.1 Domain Model

```go
package user

import "time"

type Domain struct {
    ID           string
    Email        string
    Name         string
    PasswordHash string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 33.2 Request DTOs

```go
package user

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8,max=72"`
}

type UpdateUserRequest struct {
    Email *string `json:"email,omitempty" validate:"omitempty,email"`
    Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
}

type ListUsersQuery struct {
    Page    int    `query:"page" validate:"omitempty,min=1"`
    PerPage int    `query:"per_page" validate:"omitempty,min=1,max=100"`
    Sort    string `query:"sort" validate:"omitempty,oneof=name email created_at"`
    Order   string `query:"order" validate:"omitempty,oneof=asc desc"`
    Search  string `query:"search" validate:"omitempty,max=255"`
}
```

### 33.3 Response DTOs

```go
type UserResponse struct {
    ID        string `json:"id"`
    Email     string `json:"email"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type ListUsersResponse struct {
    Data []UserResponse `json:"data"`
    Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

### 33.4 Repository Interface

```go
type Repository interface {
    Create(ctx context.Context, user *Domain) (*Domain, error)
    GetByID(ctx context.Context, id string) (*Domain, error)
    GetByEmail(ctx context.Context, email string) (*Domain, error)
    List(ctx context.Context, query ListQuery) ([]*Domain, int, error)
    Update(ctx context.Context, user *Domain) (*Domain, error)
    Delete(ctx context.Context, id string) error
    ExistsByEmail(ctx context.Context, email string) (bool, error)
}
```

### 33.5 PostgreSQL Repository

```go
type postgresRepository struct {
    db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
    return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, user *Domain) (*Domain, error) {
    ctx, span := tracer.Start(ctx, "repository.user.create")
    defer span.End()

    query := `
        INSERT INTO users (email, name, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`

    err := r.db.QueryRowContext(ctx, query,
        user.Email, user.Name, user.PasswordHash,
    ).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("user repository: create: %w", err)
    }

    return user, nil
}

func (r *postgresRepository) List(ctx context.Context, query ListQuery) ([]*Domain, int, error) {
    ctx, span := tracer.Start(ctx, "repository.user.list")
    defer span.End()

    where := "WHERE 1=1"
    args := []any{}
    argIdx := 1

    if query.Search != "" {
        where += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx)
        args = append(args, "%"+query.Search+"%")
        argIdx++
    }

    countQuery := "SELECT COUNT(*) FROM users " + where
    var total int
    if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
        return nil, 0, fmt.Errorf("user repository: list count: %w", err)
    }

    order := "created_at DESC"
    if query.Sort != "" {
        direction := "DESC"
        if query.Order == "asc" {
            direction = "ASC"
        }
        order = query.Sort + " " + direction
    }

    listQuery := fmt.Sprintf(`
        SELECT id, email, name, created_at, updated_at
        FROM users %s
        ORDER BY %s
        LIMIT $%d OFFSET $%d`, where, order, argIdx, argIdx+1)
    args = append(args, query.PerPage, (query.Page-1)*query.PerPage)

    rows, err := r.db.QueryContext(ctx, listQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("user repository: list: %w", err)
    }
    defer rows.Close()

    var users []*Domain
    for rows.Next() {
        var u Domain
        if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, 0, fmt.Errorf("user repository: list scan: %w", err)
        }
        users = append(users, &u)
    }

    return users, total, nil
}
```

### 33.6 Service

```go
package user

import (
    "context"
    "fmt"

    "golang.org/x/crypto/bcrypt"
)

type Service struct {
    repo Repository
    rbac RBACChecker
}

type RBACChecker interface {
    HasPermission(ctx context.Context, userID, permission string) (bool, error)
}

func NewService(repo Repository, rbac RBACChecker) *Service {
    return &Service{repo: repo, rbac: rbac}
}

func (s *Service) Create(ctx context.Context, req *CreateUserRequest) (*Domain, error) {
    exists, err := s.repo.ExistsByEmail(ctx, req.Email)
    if err != nil {
        return nil, fmt.Errorf("user service: check email: %w", err)
    }
    if exists {
        return nil, NewAppError("CONFLICT", "Email already exists", 409, nil)
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
    if err != nil {
        return nil, fmt.Errorf("user service: hash password: %w", err)
    }

    user := &Domain{
        Email:        req.Email,
        Name:         req.Name,
        PasswordHash: string(hash),
    }

    created, err := s.repo.Create(ctx, user)
    if err != nil {
        return nil, fmt.Errorf("user service: create: %w", err)
    }

    return created, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Domain, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("user service: get by id: %w", err)
    }
    return user, nil
}

func (s *Service) List(ctx context.Context, query ListQuery) ([]*Domain, int, error) {
    if query.Page == 0 {
        query.Page = 1
    }
    if query.PerPage == 0 {
        query.PerPage = 20
    }

    users, total, err := s.repo.List(ctx, query)
    if err != nil {
        return nil, 0, fmt.Errorf("user service: list: %w", err)
    }

    return users, total, nil
}

func (s *Service) Update(ctx context.Context, id string, req *UpdateUserRequest) (*Domain, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("user service: get for update: %w", err)
    }

    if req.Email != nil {
        exists, err := s.repo.ExistsByEmail(ctx, *req.Email)
        if err != nil {
            return nil, fmt.Errorf("user service: check email: %w", err)
        }
        if exists {
            return nil, NewAppError("CONFLICT", "Email already exists", 409, nil)
        }
        user.Email = *req.Email
    }

    if req.Name != nil {
        user.Name = *req.Name
    }

    updated, err := s.repo.Update(ctx, user)
    if err != nil {
        return nil, fmt.Errorf("user service: update: %w", err)
    }

    return updated, nil
}

func (s *Service) Delete(ctx context.Context, actorID, targetID string) error {
    if !s.rbac.HasPermission(ctx, actorID, "users.delete") {
        return NewAppError("FORBIDDEN", "Insufficient permissions", 403, nil)
    }

    if actorID == targetID {
        return NewAppError("BUSINESS_ERROR", "Cannot delete your own account", 422, nil)
    }

    if err := s.repo.Delete(ctx, targetID); err != nil {
        return fmt.Errorf("user service: delete: %w", err)
    }

    return nil
}
```

### 33.7 Handler

```go
package user

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

type Handler struct {
    service   *Service
    validator Validator
}

type Validator interface {
    Struct(s any) error
}

func NewHandler(service *Service, validator Validator) *Handler {
    return &Handler{service: service, validator: validator}
}

func (h *Handler) Create(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return response.BadRequest(c, "Invalid request body")
    }

    if err := h.validator.Struct(req); err != nil {
        return response.ValidationError(c, err)
    }

    user, err := h.service.Create(c.Context(), &req)
    if err != nil {
        return response.Error(c, err)
    }

    return response.Created(c, toUserResponse(user))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
    id := c.Params("id")
    if _, err := uuid.Parse(id); err != nil {
        return response.BadRequest(c, "Invalid user ID")
    }

    user, err := h.service.GetByID(c.Context(), id)
    if err != nil {
        return response.Error(c, err)
    }

    return response.OK(c, toUserResponse(user))
}

func (h *Handler) List(c *fiber.Ctx) error {
    var query ListUsersQuery
    if err := c.QueryParser(&query); err != nil {
        return response.BadRequest(c, "Invalid query parameters")
    }

    if err := h.validator.Struct(query); err != nil {
        return response.ValidationError(c, err)
    }

    users, total, err := h.service.List(c.Context(), ListUsersQuery{
        Page:    query.Page,
        PerPage: query.PerPage,
        Sort:    query.Sort,
        Order:   query.Order,
        Search:  query.Search,
    })
    if err != nil {
        return response.Error(c, err)
    }

    return response.OK(c, ListUsersResponse{
        Data: toUserResponseList(users),
        Meta: PaginationMeta{
            Page:       query.Page,
            PerPage:    query.PerPage,
            Total:      total,
            TotalPages: (total + query.PerPage - 1) / query.PerPage,
        },
    })
}

func (h *Handler) Update(c *fiber.Ctx) error {
    id := c.Params("id")
    if _, err := uuid.Parse(id); err != nil {
        return response.BadRequest(c, "Invalid user ID")
    }

    var req UpdateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return response.BadRequest(c, "Invalid request body")
    }

    if err := h.validator.Struct(req); err != nil {
        return response.ValidationError(c, err)
    }

    user, err := h.service.Update(c.Context(), id, &req)
    if err != nil {
        return response.Error(c, err)
    }

    return response.OK(c, toUserResponse(user))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
    id := c.Params("id")
    if _, err := uuid.Parse(id); err != nil {
        return response.BadRequest(c, "Invalid user ID")
    }

    actorID := auth.GetUserID(c)

    if err := h.service.Delete(c.Context(), actorID, id); err != nil {
        return response.Error(c, err)
    }

    return c.SendStatus(fiber.StatusNoContent)
}
```

### 33.8 Routes

```go
package user

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(api fiber.Router, h *Handler, authMiddleware fiber.Handler) {
    users := api.Group("/users")
    users.Use(authMiddleware)

    users.Get("", h.List)
    users.Post("", h.Create)
    users.Get("/:id", h.GetByID)
    users.Patch("/:id", h.Update)
    users.Delete("/:id", h.Delete)
}
```

### 33.9 Migration

```sql
-- 000001_create_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);
```

```sql
-- 000001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

### 33.10 RBAC Permissions

```sql
INSERT INTO permissions (name, description) VALUES
    ('users.view', 'View users'),
    ('users.create', 'Create users'),
    ('users.update', 'Update users'),
    ('users.delete', 'Delete users');

INSERT INTO roles (name, description) VALUES
    ('admin', 'Administrator with full access'),
    ('user', 'Regular user');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'user' AND p.name = 'users.view';
```

### 33.11 OpenAPI Definition

```go
// @Summary      List users
// @Description  Returns a paginated list of users
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        page     query int    false "Page number"     default(1)
// @Param        per_page query int    false "Items per page"  default(20)
// @Param        sort     query string false "Sort field"      Enums(name, email, created_at)
// @Param        order    query string false "Sort direction"  Enums(asc, desc)
// @Param        search   query string false "Search query"
// @Success      200 {object} ListUsersResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Router       /api/v1/users [get]
// @Security     BearerAuth
```

### 33.12 Unit Tests

```go
package user

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestService_Create_Success(t *testing.T) {
    repo := &mockRepository{
        existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
            return false, nil
        },
        createFn: func(ctx context.Context, user *Domain) (*Domain, error) {
            user.ID = "generated-id"
            return user, nil
        },
    }

    svc := NewService(repo, &mockRBAC{})

    user, err := svc.Create(context.Background(), &CreateUserRequest{
        Email:    "test@example.com",
        Name:     "Test User",
        Password: "securePassword123",
    })

    require.NoError(t, err)
    assert.Equal(t, "generated-id", user.ID)
    assert.Equal(t, "test@example.com", user.Email)
}

func TestService_Create_DuplicateEmail(t *testing.T) {
    repo := &mockRepository{
        existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
            return true, nil
        },
    }

    svc := NewService(repo, &mockRBAC{})

    _, err := svc.Create(context.Background(), &CreateUserRequest{
        Email:    "existing@example.com",
        Name:     "Test User",
        Password: "securePassword123",
    })

    require.Error(t, err)
    assert.Contains(t, err.Error(), "already exists")
}

func TestService_GetByID_NotFound(t *testing.T) {
    repo := &mockRepository{
        getByIDFn: func(ctx context.Context, id string) (*Domain, error) {
            return nil, ErrNotFound
        },
    }

    svc := NewService(repo, &mockRBAC{})

    _, err := svc.GetByID(context.Background(), "nonexistent-id")

    require.Error(t, err)
    assert.ErrorIs(t, err, ErrNotFound)
}
```

### 33.13 E2E Tests

```go
package e2e

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUser_Create_Success(t *testing.T) {
    body := map[string]string{
        "email":    "newuser@example.com",
        "name":     "New User",
        "password": "securePassword123",
    }
    jsonBody, _ := json.Marshal(body)

    req, _ := http.NewRequest("POST", baseURL+"/api/v1/users", bytes.NewReader(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+adminToken)

    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var response map[string]any
    json.NewDecoder(resp.Body).Decode(&response)
    assert.Equal(t, true, response["success"])
}

func TestUser_Create_Validation(t *testing.T) {
    body := map[string]string{
        "email":    "invalid-email",
        "name":     "",
        "password": "short",
    }
    jsonBody, _ := json.Marshal(body)

    req, _ := http.NewRequest("POST", baseURL+"/api/v1/users", bytes.NewReader(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+adminToken)

    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestUser_List_RequiresAuth(t *testing.T) {
    req, _ := http.NewRequest("GET", baseURL+"/api/v1/users", nil)

    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
```

---

## 34. Observability Architecture

### 34.1 Three Pillars Integration

```text
Request
   │
   ├── request_id
   ├── trace_id
   └── span_id
          │
          ├── logs (slog with trace_id and span_id fields)
          ├── metrics (Prometheus with route and status labels)
          └── traces (OTel spans with log and metric correlation)
```

### 34.2 Correlation Example

A single request to `POST /api/v1/users` produces:

**Log entry:**
```json
{
    "time": "2026-08-20T14:30:00.123Z",
    "level": "INFO",
    "msg": "Request completed",
    "request_id": "req_abc123",
    "trace_id": "trace_xyz789",
    "span_id": "span_def456",
    "method": "POST",
    "path": "/api/v1/users",
    "status": 201,
    "latency_ms": 45.2
}
```

**Metric data point:**
```text
http_requests_total{method="POST", path="/api/v1/users", status="201"} 1
http_request_duration_seconds{method="POST", path="/api/v1/users"} 0.0452
```

**Trace span:**
```json
{
    "trace_id": "trace_xyz789",
    "span_id": "span_def456",
    "operation": "POST /api/v1/users",
    "duration_ms": 45.2,
    "status": "OK",
    "attributes": {
        "http.method": "POST",
        "http.path": "/api/v1/users",
        "http.status_code": 201,
        "user.id": "user_abc123"
    }
}
```

### 34.3 Debugging Flow

```text
1. User reports slow request
2. Find trace by trace_id from Grafana/Jaeger
3. See latency breakdown across services
4. Find logs by trace_id → see detailed error context
5. Check metrics by route → see if it's a pattern or isolated
```

---

## 35. Scalability Strategy

### 35.1 Stateless Design

The application is stateless. All state lives in:
- PostgreSQL (domain data, refresh tokens)
- Redis (rate limits, permission cache, cache layer)

This enables horizontal scaling by simply adding more instances behind a load balancer.

### 35.2 Connection Pooling

Database and Redis connections are pooled per-instance:
- PostgreSQL: 25 open, 10 idle (configurable)
- Redis: default pool size (10 * NumCPU)

### 35.3 Rate Limiting Considerations

| Deployment | Rate Limit Strategy | Notes |
|------------|--------------------|----|
| Single instance | Local in-memory | Simple, no Redis dependency |
| Multiple instances | Redis distributed | Shared counter across instances |
| Multiple instances (without Redis) | Local per-instance | Each instance allows full rate |

**Do not silently allow local rate limiting to behave as distributed.** When `RATE_LIMIT_REDIS_ENABLED=false` and multiple instances are running, log a warning at startup.

### 35.4 Refresh Token Storage

Refresh tokens are stored in PostgreSQL. Multiple instances share the same database, so refresh token operations are consistent across instances.

### 35.5 Horizontal Scaling Checklist

- [ ] All state externalized to PostgreSQL/Redis
- [ ] No in-process caching of critical data (or cache is Redis-backed)
- [ ] Rate limiting uses Redis
- [ ] Health checks pass for new instances
- [ ] Graceful shutdown drains connections
- [ ] Load balancer uses health check endpoint

---

## 36. Architecture Decision Records

### ADR-001: Fiber as HTTP Framework

**Context:** Need a high-performance HTTP framework for Go REST APIs.
**Decision:** Use Fiber v2.
**Alternatives:** Gin, Echo, Chi, net/http with chi router.
**Trade-offs:** Fiber uses fasthttp, not net/http. Some stdlib-compatible middleware won't work directly. Performance is excellent. Team already uses Fiber.
**Consequences:** Team productivity is high. Some middleware may need adaptation. Migration away from Fiber would require rewriting HTTP layer only.

### ADR-002: Feature-Oriented Architecture

**Context:** Need clear code organization for growing codebases.
**Decision:** Organize by feature (auth, user, role) rather than by technical layer.
**Alternatives:** Strict layered architecture, DDD with aggregates.
**Trade-offs:** Feature modules may have some code duplication. But each feature is self-contained, easy to navigate, and easy to remove.
**Consequences:** New developers can find all code for a feature in one directory. Cross-feature dependencies require explicit interfaces.

### ADR-003: SQL Migrations with golang-migrate

**Context:** Need versioned, reversible database schema management.
**Decision:** Use `golang-migrate/migrate` with raw SQL files.
**Alternatives:** goose, atlas, ORM-generated migrations.
**Trade-offs:** Raw SQL requires understanding the target database. But it provides full control, works with all databases, and is the most transparent approach.
**Consequences:** Migrations are database-specific but fully transparent. No ORM coupling.

### ADR-004: Database Access Strategy

**Context:** Need database access that works across PostgreSQL, MySQL, MariaDB, SQLite.
**Decision:** Use `sqlx` as a thin wrapper over `database/sql`.
**Alternatives:** GORM, sqlc, bare `database/sql`.
**Trade-offs:** `sqlx` adds struct scanning and named parameters without ORM magic. GORM is easier but adds leaky abstractions.
**Consequences:** Repository implementations are slightly more verbose than GORM but fully transparent and maintainable.

### ADR-005: JWT Access/Refresh Token Architecture

**Context:** Need stateless authentication with token revocation capability.
**Decision:** Short-lived access tokens (15 min) + long-lived refresh tokens (7 days) stored in HttpOnly cookies and database.
**Alternatives:** Session-based auth, long-lived access tokens, opaque tokens.
**Trade-offs:** More complex than sessions but enables stateless API authentication. Database storage for refresh tokens adds a dependency but enables revocation and reuse detection.
**Consequences:** Token rotation and reuse detection provide strong security.

### ADR-006: RBAC Architecture

**Context:** Need role-based access control for multi-tenant services.
**Decision:** Spatie-inspired RBAC with User → Roles → Permissions model, Redis caching.
**Alternatives:** Casbin, OPA, custom ACL.
**Trade-offs:** Simpler than Casbin/OPA but sufficient for most use cases. Permission caching reduces database load but adds cache invalidation complexity.
**Consequences:** RBAC is easy to understand and extend.

### ADR-007: Logging Architecture

**Context:** Need structured, leveled logging with rotation.
**Decision:** `slog` (stdlib) with `lumberjack.v2` for rotation, plus sensitive data sanitization.
**Alternatives:** zerolog, zap, logrus.
**Trade-offs:** `slog` is stdlib (zero dependencies) but slightly slower than zerolog/zap. For most APIs, the difference is negligible.
**Consequences:** Zero external logging dependencies. Rotation is handled automatically.

### ADR-008: Prometheus Metrics

**Context:** Need application metrics for monitoring.
**Decision:** `prometheus/client_golang` for direct Prometheus exposition.
**Alternatives:** OpenTelemetry metrics.
**Trade-offs:** `prometheus/client_golang` is the most mature.
**Consequences:** Metrics are directly scrapable by Prometheus.

### ADR-009: OpenTelemetry Tracing

**Context:** Need distributed tracing across services.
**Decision:** OpenTelemetry SDK with OTLP exporter.
**Alternatives:** Jaeger client directly, Zipkin.
**Trade-offs:** OTEL is the CNCF standard and vendor-neutral.
**Consequences:** Traces are exportable to any OTEL-compatible backend.

### ADR-010: Distributed Rate Limiting

**Context:** Need rate limiting that works across multiple instances.
**Decision:** `ulule/limiter` with Redis backend (distributed) or in-memory backend (local).
**Alternatives:** Custom token bucket, `didip/tollbooth`.
**Trade-offs:** `ulule/limiter` provides both Redis and in-memory backends with middleware integration.
**Consequences:** Rate limiting degrades gracefully when Redis is unavailable.

### ADR-011: Multi-Database Support

**Context:** Different deployment environments may require different databases.
**Decision:** Support PostgreSQL (primary), MySQL, MariaDB, SQLite through adapter pattern.
**Alternatives:** Single database, ORM-based abstraction.
**Trade-offs:** More code to maintain (multiple repository implementations). But each implementation is transparent.
**Consequences:** Repository interfaces are the abstraction boundary.

### ADR-012: Log Sanitization

**Context:** Sensitive data (passwords, tokens, credit cards, SSN) must never appear in logs.
**Decision:** Regex-based pattern matching with automatic masking at the slog handler level.
**Alternatives:** Per-field manual redaction, structured logging only (no raw string logging).
**Trade-offs:** Regex adds slight overhead per log call. But it catches accidentally logged sensitive data across all log levels and formats.
**Consequences:** All log output is automatically scanned and sensitive patterns are masked before writing.

### ADR-013: Application-Level Throttling

**Context:** Rate limiting controls total request count, but does not control concurrency or processing speed.
**Decision:** Semaphore-based concurrency limiter at the middleware level, separate from rate limiting.
**Alternatives:** Worker pool pattern, token bucket for processing speed.
**Trade-offs:** Adds a middleware layer but provides defense-in-depth against traffic spikes and slow clients.
**Consequences:** Protects backend resources from being overwhelmed even when rate limits are generous.

### ADR-014: In-Memory Cache Layer

**Context:** Frequently accessed small data (config, permissions, feature flags) needs fast reads without Redis dependency.
**Decision:** Go `sync.RWMutex`-based in-memory cache with optional Redis backing for distributed scenarios.
**Alternatives:** Redis-only caching, bigcache, ristretto.
**Trade-offs:** In-memory is fastest but not shared across instances. Redis is shared but adds network latency and dependency.
**Consequences:** Developers can choose per-use-case whether in-memory, Redis, or both is appropriate.

---

## 37. Implementation Phases

### Phase 1: Foundation (Week 1)

- [ ] Project setup (`go.mod`, directory structure)
- [ ] Configuration loading (`internal/config`)
- [ ] Fiber server setup (`internal/server`)
- [ ] Request ID middleware
- [ ] Logging setup (`internal/logging`)
- [ ] Error handling (`internal/pkg/errors.go`)
- [ ] Response helpers (`internal/pkg/response.go`)
- [ ] Graceful shutdown
- [ ] Health check endpoints
- [ ] Docker Compose (PostgreSQL, Redis)
- [ ] Taskfile.yml
- [ ] `.env.example`
- [ ] Basic CI pipeline (lint, build)

### Phase 2: Database & Auth (Week 2)

- [ ] Database connection (`internal/database`)
- [ ] Migration system
- [ ] User migration
- [ ] User repository (PostgreSQL)
- [ ] User service
- [ ] User handler
- [ ] User routes
- [ ] Validation
- [ ] JWT utilities (`internal/auth/jwt.go`)
- [ ] Auth handler (register, login, refresh, logout)
- [ ] Auth middleware
- [ ] Refresh token migration and repository
- [ ] Cookie management

### Phase 3: RBAC & Security (Week 3)

- [ ] Role and Permission migrations
- [ ] RBAC service (`internal/rbac`)
- [ ] Permission caching in Redis
- [ ] RBAC middleware
- [ ] Security headers middleware
- [ ] CORS middleware
- [ ] Rate limiting middleware (local + Redis)
- [ ] Request size limits

### Phase 4: Observability (Week 3-4)

- [ ] Prometheus metrics (`internal/metrics`)
- [ ] Metrics middleware
- [ ] OpenTelemetry setup (`internal/tracing`)
- [ ] Tracing middleware
- [ ] Request logging middleware
- [ ] Log rotation configuration
- [ ] Sensitive data redaction (log sanitization)
- [ ] Observability correlation (request_id, trace_id, span_id in logs)

### Phase 5: Advanced Features (Week 4)

- [ ] Application-level throttling middleware
- [ ] In-memory cache layer (`internal/cache/local.go`)
- [ ] Cache middleware/helper
- [ ] Cache-aside pattern helpers
- [ ] Cache invalidation hooks

### Phase 6: Documentation & Testing (Week 4-5)

- [ ] OpenAPI annotations on all endpoints
- [ ] Swagger UI integration
- [ ] Swagger generation in CI
- [ ] Unit tests for all services
- [ ] Unit tests for validators
- [ ] E2E test infrastructure (Testcontainers)
- [ ] E2E tests for auth flow
- [ ] E2E tests for user CRUD
- [ ] E2E tests for RBAC
- [ ] E2E tests for rate limiting
- [ ] E2E tests for health checks

### Phase 7: Production Readiness (Week 5)

- [ ] Production Dockerfile
- [ ] Security scan in CI (`govulncheck`)
- [ ] Dependabot configuration
- [ ] README documentation
- [ ] Example feature documentation
- [ ] CI pipeline finalization
- [ ] Performance testing baseline
- [ ] Final review and hardening

---

## 38. Acceptance Criteria

### 39.1 Functional

- [ ] Application starts and serves HTTP on configured port
- [ ] All CRUD operations for User feature work correctly
- [ ] JWT authentication flow works (register → login → access protected endpoint → refresh → logout)
- [ ] RBAC permissions are enforced at both middleware and service layers
- [ ] Rate limiting works in both local and distributed modes
- [ ] Application-level throttling limits concurrency correctly
- [ ] In-memory cache returns correct data and respects TTL
- [ ] Health checks return correct status
- [ ] Database migrations run up and down correctly
- [ ] CORS headers are set correctly
- [ ] OpenAPI docs are generated and accessible at `/swagger/index.html`
- [ ] All unit tests pass
- [ ] All E2E tests pass

### 39.2 Non-Functional

- [ ] Application starts in under 2 seconds
- [ ] Health check responds in under 100ms
- [ ] Logging output is structured JSON in production mode
- [ ] Log rotation works (file size and age)
- [ ] Sensitive data is automatically masked in all log output
- [ ] Graceful shutdown completes within configured timeout
- [ ] Docker image is under 30MB
- [ ] No secrets are hardcoded or logged
- [ ] CI pipeline passes (lint, test, build, security scan)
- [ ] Code coverage is above 70% for unit tests
- [ ] All linter checks pass

### 39.3 Developer Experience

- [ ] New developer can start the project with `task dev` after cloning
- [ ] All Taskfile commands work as documented
- [ ] `.env.example` contains all required variables with documentation
- [ ] Code follows consistent conventions
- [ ] Request lifecycle is traceable by reading code
- [ ] Adding a new feature follows a clear, documented pattern

---

## 39. Risks and Trade-offs

### 39.1 Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Fiber uses fasthttp, not net/http | Some stdlib middleware incompatible | Document known incompatibilities. Service layer is framework-agnostic. |
| SQLite not suitable for production | Misuse in multi-instance deployments | Document clearly: SQLite is for dev/testing only. |
| Single JWT secret | Compromise affects all services | Support key rotation via `JWT_SECRET_PREVIOUS`. Document rotation procedure. |
| In-memory cache not shared across instances | Stale data between instances | Use Redis for data that must be consistent across instances. Document trade-off. |
| Log sanitization regex overhead | Slight latency per log call | Measure overhead in benchmarks. Use compiled regex patterns. Disable in extreme performance scenarios. |
| Throttling under sudden traffic spikes | May reject legitimate requests | Configure burst capacity. Monitor throttle metrics. Adjust limits based on real traffic patterns. |

### 39.2 Operational Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Redis failure | Rate limiting degrades to local mode, cache misses | Fail open for rate limiting. Cache falls back to database/source. |
| Database failure | All write operations fail | Health check reports unhealthy. Graceful degradation for read-only cached data. |
| Log disk full | Application may hang on log writes | Configure log rotation with max size. Monitor disk usage. |
| Clock skew across instances | JWT validation failures | Allow 30 seconds of clock skew. Use NTP. |

### 39.3 Decision Trade-offs

| Decision | Chosen | Trade-off |
|----------|--------|-----------|
| slog over zerolog/zap | slog | Slightly slower, but zero dependencies and stdlib |
| sqlx over GORM | sqlx | More verbose, but transparent and maintainable |
| Feature-oriented over layer-oriented | Feature-oriented | Some code duplication across features, but better navigation |
| Code-first OpenAPI over OpenAPI-first | Code-first | Annotations can be verbose, but always in sync |
| In-memory cache over Redis-only | Both | More code, but no Redis required for simple use cases |

---

## 40. Future Extensions

- **WebSocket support** for real-time features
- **Background job system** with retry and backoff
- **Multi-tenancy** support
- **API key authentication** for machine-to-machine communication
- **OAuth2 provider** integration (Google, GitHub, etc.)
- **Email service** integration
- **File upload/storage** abstraction
- **Internationalization (i18n)** for error messages
- **GraphQL** endpoint alongside REST
- **gRPC** service definitions
- **Feature flags** system
- **Audit logging** for compliance
- **Data encryption at rest** for sensitive fields
- **Circuit breaker** pattern for external service calls
- **Retry with exponential backoff** for transient failures
- **Load shedding** for overload protection

---

## 41. Log Sanitization and Sensitive Data Masking

### 43.1 Problem

Sensitive information can accidentally end up in log output through:
- Structured log fields containing sensitive values
- Error messages wrapping sensitive data (e.g., `fmt.Errorf("failed to process card %s: %w", cardNumber, err)`)
- Request/response body logging
- Manual developer mistakes (e.g., `log.Info("user login", "password", password)`)

### 43.2 Supported Sensitive Patterns

The sanitizer must detect and mask the following patterns:

| Data Type | Pattern | Mask Example |
|-----------|---------|--------------|
| **Credit card numbers** | 13-19 digit sequences (Visa, Mastercard, Amex, etc.) | `4111-1111-1111-1111` → `4111-****-****-1111` |
| **CVV/CVC** | 3-4 digit codes after card context | `CVV: 123` → `CVV: ***` |
| **SSN (US)** | `XXX-XX-XXXX` or `XXXXXXXXX` | `123-45-6789` → `***-**-6789` |
| **Passwords** | Key-value patterns: `password=...`, `"password":"..."`, `password: ...` | `password=mysecret` → `password=***` |
| **Access tokens** | Bearer tokens, JWT patterns (`eyJ...`) | `Bearer eyJhbG...` → `Bearer ***` |
| **Refresh tokens** | UUID-like tokens, opaque token strings in auth context | `refresh_token=abc123` → `refresh_token=***` |
| **API keys** | `sk-...`, `ak_...`, `key=...` patterns | `sk-proj-abc123` → `sk-proj-***` |
| **Private keys** | PEM-encoded private key blocks | `-----BEGIN PRIVATE KEY-----` → `[REDACTED_PRIVATE_KEY]` |
| **Database URLs** | Connection strings with credentials | `postgres://user:pass@host` → `postgres://user:***@host` |
| **Email addresses** (optional) | `user@domain.com` | `u***@domain.com` (configurable) |
| **IP addresses** (optional) | IPv4/IPv6 | Configurable per environment |
| **Authorization headers** | `Authorization: ...` values | `Authorization: ***` |

### 43.3 Architecture

```text
Log Message
    │
    ▼
Sanitizer Handler (wraps slog.Handler)
    │
    ├── Scan log message string for sensitive patterns
    ├── Scan all string attribute values for sensitive patterns
    ├── Apply masking rules
    │
    ▼
Underlying slog.Handler (console or file)
```

### 43.4 Implementation

```go
// internal/logging/sanitizer.go
package logging

import (
    "log/slog"
    "regexp"
    "strings"
)

// Sensitive patterns compiled once at init
var sensitivePatterns = []struct {
    Pattern *regexp.Regexp
    Mask    func(matched string) string
}{
    // Credit card numbers (Visa, Mastercard, Amex, Discover, etc.)
    {
        Pattern: regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`),
        Mask: func(s string) string {
            if len(s) <= 4 { return "***" }
            return s[:4] + strings.Repeat("*", len(s)-4)
        },
    },
    // CVV/CVC (3-4 digits near card context)
    {
        Pattern: regexp.MustCompile(`(?i)(cvv|cvc|security.?code)\s*[:=]\s*\d{3,4}`),
        Mask: func(s string) string {
            idx := strings.IndexAny(s, "0123456789")
            if idx < 0 { return s }
            return s[:idx] + "***"
        },
    },
    // SSN (US Social Security Number)
    {
        Pattern: regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
        Mask: func(s string) string {
            parts := strings.Split(s, "-")
            if len(parts) == 3 {
                return "***-**-" + parts[2]
            }
            return "***"
        },
    },
    // Passwords in key-value patterns
    {
        Pattern: regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*\S+`),
        Mask: func(s string) string {
            eqIdx := strings.IndexAny(s, ":=")
            if eqIdx < 0 { return "***" }
            return s[:eqIdx+1] + " ***"
        },
    },
    // JWT tokens (eyJ...)
    {
        Pattern: regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]*`),
        Mask: func(s string) string { return "***" },
    },
    // Bearer tokens
    {
        Pattern: regexp.MustCompile(`(?i)(Bearer\s+)[A-Za-z0-9_\-.]+`),
        Mask: func(s string) string {
            idx := strings.Index(s, " ") + strings.Index(strings.ToLower(s), "bearer")
            if idx < 0 { return "Bearer ***" }
            spaceIdx := strings.Index(s[idx:], " ")
            if spaceIdx < 0 { return "Bearer ***" }
            return s[:idx+spaceIdx+1] + "***"
        },
    },
    // API keys (sk-, ak_, key=, api_key=)
    {
        Pattern: regexp.MustCompile(`(?i)(sk-|ak_|api.?key.?=.?)[A-Za-z0-9_\-]{8,}`),
        Mask: func(s string) string {
            eqIdx := strings.IndexAny(s, "=-")
            if eqIdx < 0 { return "***" }
            prefix := s[:eqIdx+1]
            return prefix + "***"
        },
    },
    // Private keys (PEM)
    {
        Pattern: regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE KEY-----[\s\S]*?-----END\s+(RSA\s+)?PRIVATE KEY-----`),
        Mask: func(s string) string { return "[REDACTED_PRIVATE_KEY]" },
    },
    // Database connection strings with credentials
    {
        Pattern: regexp.MustCompile(`(postgres|mysql|mongodb|redis):\/\/[^:]+:[^@]+@`),
        Mask: func(s string) string {
            atIdx := strings.LastIndex(s, "@")
            colonIdx := strings.Index(s, "://")
            if colonIdx < 0 || atIdx < 0 { return s }
            credStart := s[colonIdx+3:]
            credColon := strings.Index(credStart, ":")
            if credColon < 0 { return s }
            return s[:colonIdx+3+credColon] + ":***@" + s[atIdx+1:]
        },
    },
    // Refresh tokens (in key-value context)
    {
        Pattern: regexp.MustCompile(`(?i)(refresh.?token)\s*[:=]\s*\S+`),
        Mask: func(s string) string {
            eqIdx := strings.IndexAny(s, ":=")
            if eqIdx < 0 { return "***" }
            return s[:eqIdx+1] + " ***"
        },
    },
    // Access tokens (in key-value context)
    {
        Pattern: regexp.MustCompile(`(?i)(access.?token)\s*[:=]\s*\S+`),
        Mask: func(s string) string {
            eqIdx := strings.IndexAny(s, ":=")
            if eqIdx < 0 { return "***" }
            return s[:eqIdx+1] + " ***"
        },
    },
}

// SanitizeString applies all masking rules to a string
func SanitizeString(s string) string {
    for _, p := range sensitivePatterns {
        s = p.Pattern.ReplaceAllStringFunc(s, p.Mask)
    }
    return s
}

// SanitizeHandler wraps a slog.Handler and sanitizes all output
type SanitizeHandler struct {
    inner slog.Handler
}

func NewSanitizeHandler(inner slog.Handler) *SanitizeHandler {
    return &SanitizeHandler{inner: inner}
}

func (h *SanitizeHandler) Handle(ctx context.Context, record slog.Record) error {
    // Sanitize the message
    record.Message = SanitizeString(record.Message)

    // Sanitize all attribute values
    var sanitizedAttrs []slog.Attr
    record.Attrs(func(a slog.Attr) bool {
        if a.Value.Kind() == slog.KindString {
            a.Value = slog.StringValue(SanitizeString(a.Value.String()))
        }
        sanitizedAttrs = append(sanitizedAttrs, a)
        return true
    })

    // Rebuild record with sanitized attrs
    record = slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
    for _, attr := range sanitizedAttrs {
        record.AddAttrs(attr)
    }

    return h.inner.Handle(ctx, record)
}

func (h *SanitizeHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.inner.Enabled(ctx, level)
}

func (h *SanitizeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return &SanitizeHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *SanitizeHandler) WithGroup(name string) slog.Handler {
    return &SanitizeHandler{inner: h.inner.WithGroup(name)}
}
```

### 43.5 Usage in Logger Initialization

```go
func NewLogger(cfg config.Config) *slog.Logger {
    var handler slog.Handler

    if cfg.LogFormat == "console" {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: parseLevel(cfg.LogLevel),
        })
    } else {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: parseLevel(cfg.LogLevel),
        })
    }

    // Wrap with sanitizer
    handler = NewSanitizeHandler(handler)

    // Add source information
    handler = NewSourceHandler(handler)

    return slog.New(handler).With(
        "service", cfg.AppName,
    )
}
```

### 43.6 Custom SensitiveType for Structured Logging

For values that should **always** be masked, use a custom type:

```go
// internal/logging/sensitive.go
package logging

import "encoding/json"

// SensitiveString is a string type that is always masked in log output.
// Use this for passwords, tokens, credit card numbers, etc.
type SensitiveString struct {
    value string
}

func NewSensitiveString(v string) SensitiveString {
    return SensitiveString{value: v}
}

// String returns a masked representation.
func (s SensitiveString) String() string {
    return "[REDACTED]"
}

// MarshalJSON returns a masked JSON value.
func (s SensitiveString) MarshalJSON() ([]byte, error) {
    return json.Marshal("[REDACTED]")
}

// Value returns the original value for internal use only.
// Never call this in log output or API responses.
func (s SensitiveString) Value() string {
    return s.value
}
```

Usage:

```go
// In service/handler — the password is logged as [REDACTED]
log.Info("user registration",
    "email", user.Email,
    "password", logging.NewSensitiveString(req.Password), // always masked
)
```

### 43.7 Environment-Specific Behavior

| Environment | Sanitization | Rationale |
|-------------|-------------|-----------|
| Development | Enabled (but may disable for debugging) | Developers may need to see raw values during debugging |
| Staging | Enabled | Staging may connect to real databases |
| Production | **Always enabled, cannot be disabled** | Never risk leaking sensitive data |

### 43.8 Testing Sanitization

```go
func TestSanitizeString_CreditCard(t *testing.T) {
    input := "Processing card 4111111111111111"
    result := SanitizeString(input)
    assert.NotContains(t, result, "4111111111111111")
    assert.Contains(t, result, "4111")
}

func TestSanitizeString_Password(t *testing.T) {
    input := `{"password":"mysecretpassword123"}`
    result := SanitizeString(input)
    assert.NotContains(t, result, "mysecretpassword123")
    assert.Contains(t, result, "***")
}

func TestSanitizeString_JWT(t *testing.T) {
    input := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
    result := SanitizeString(input)
    assert.NotContains(t, result, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
}

func TestSanitizeString_SSN(t *testing.T) {
    input := "SSN: 123-45-6789"
    result := SanitizeString(input)
    assert.Equal(t, "SSN: ***-**-6789", result)
}

func TestSanitizeString_DatabaseURL(t *testing.T) {
    input := "postgres://admin:secretpass@db.example.com:5432/mydb"
    result := SanitizeString(input)
    assert.NotContains(t, result, "secretpass")
    assert.Contains(t, result, "postgres://admin:***@db.example.com:5432/mydb")
}

func TestSanitizeString_NoFalsePositives(t *testing.T) {
    input := "User 42 created successfully with name John"
    result := SanitizeString(input)
    assert.Equal(t, input, result) // no sanitization needed
}
```

### 43.9 Performance Considerations

- Regex patterns are compiled once at package init, not per log call.
- Sanitization adds ~1-5μs per log call (measured on typical hardware).
- For high-throughput scenarios (>100k logs/sec), consider disabling sanitization at the handler level and relying on `SensitiveString` types only.
- The sanitizer only processes string values — non-string attributes pass through unchanged.

### 43.10 Rules

1. All log output must pass through the sanitizer handler in production.
2. Use `SensitiveString` for values that are **always** sensitive (passwords, tokens).
3. The sanitizer is a safety net, not a replacement for developer discipline.
4. Never log raw passwords, tokens, credit card numbers, SSNs, or API keys.
5. Test sanitization patterns regularly to ensure coverage.
6. Document any new sensitive patterns added to the sanitizer.

---

## 42. Application-Level Throttling

### 43.1 Problem

Rate limiting controls the **total number of requests** over a time window, but does not control:
- **Concurrency** — how many requests are being processed simultaneously
- **Processing speed** — how fast requests are being handled
- **Resource exhaustion** — slow clients tying up goroutines

Throttling provides a second layer of protection beyond rate limiting.

### 43.2 Throttling vs Rate Limiting

| Aspect | Rate Limiting | Throttling |
|--------|--------------|------------|
| **Controls** | Total request count per time window | Concurrent processing slots |
| **Algorithm** | Token bucket / fixed window | Semaphore / worker pool |
| **Scope** | Per-IP, per-user, per-route | Per-route, global, per-user |
| **When triggered** | Too many requests over time | Too many requests being processed at once |
| **Response** | HTTP 429 with Retry-After | HTTP 503 with Retry-After or HTTP 429 |
| **Redis support** | Yes (distributed) | Optional (distributed semaphore) |

### 43.3 Architecture

```text
Request arrives
    │
    ▼
Rate Limit Middleware (controls total requests per window)
    │
    ▼
Throttle Middleware (controls concurrent processing)
    │
    ▼
Handler processes request
    │
    ▼
Release throttle slot
```

### 43.4 Throttling Modes

#### Mode 1: Global Concurrency Limit

Limit total concurrent requests across all routes:

```yaml
throttle:
  enabled: true
  max_concurrent: 100
  timeout: 30s
  mode: global
```

#### Mode 2: Per-Route Concurrency Limit

Limit concurrent requests per route or route group:

```yaml
throttle:
  enabled: true
  routes:
    "/api/v1/auth/login":
      max_concurrent: 10
      timeout: 10s
    "/api/v1/orders":
      max_concurrent: 50
      timeout: 30s
    "/api/v1/*":
      max_concurrent: 200
      timeout: 30s
  mode: per-route
```

#### Mode 3: Per-User Concurrency Limit

Limit concurrent requests per authenticated user:

```yaml
throttle:
  enabled: true
  max_per_user: 10
  timeout: 30s
  mode: per-user
```

### 43.5 Implementation

```go
// internal/middleware/throttle.go
package middleware

import (
    "context"
    "sync"
    "time"

    "github.com/gofiber/fiber/v2"
)

type ThrottleConfig struct {
    Enabled       bool
    MaxConcurrent int
    Timeout       time.Duration
    KeyFunc       func(c *fiber.Ctx) string // for per-user/per-route limiting
}

type ThrottleMiddleware struct {
    config   ThrottleConfig
    semaphores sync.Map // key → *Semaphore
}

type Semaphore struct {
    mu      sync.Mutex
    slots   chan struct{}
    max     int
    timeout time.Duration
}

func NewSemaphore(max int, timeout time.Duration) *Semaphore {
    return &Semaphore{
        slots:   make(chan struct{}, max),
        max:     max,
        timeout: timeout,
    }
}

func (s *Semaphore) Acquire(ctx context.Context) error {
    select {
    case s.slots <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(s.timeout):
        return context.DeadlineExceeded
    }
}

func (s *Semaphore) Release() {
    <-s.slots
}

func NewThrottleMiddleware(config ThrottleConfig) *ThrottleMiddleware {
    return &ThrottleMiddleware{config: config}
}

func (t *ThrottleMiddleware) Handle(c *fiber.Ctx) error {
    if !t.config.Enabled {
        return c.Next()
    }

    key := "global"
    if t.config.KeyFunc != nil {
        key = t.config.KeyFunc(c)
    }

    sem := t.getOrCreate(key)

    ctx, cancel := context.WithTimeout(c.Context(), t.config.Timeout)
    defer cancel()

    if err := sem.Acquire(ctx); err != nil {
        c.Set("Retry-After", fmt.Sprintf("%d", int(t.config.Timeout.Seconds())))
        return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
            "success": false,
            "error": fiber.Map{
                "code":    "THROTTLED",
                "message": "Too many concurrent requests. Please try again later.",
            },
        })
    }
    defer sem.Release()

    return c.Next()
}

func (t *ThrottleMiddleware) getOrCreate(key string) *Semaphore {
    if val, ok := t.semaphores.Load(key); ok {
        return val.(*Semaphore)
    }

    sem := NewSemaphore(t.config.MaxConcurrent, t.config.Timeout)
    actual, _ := t.semaphores.LoadOrStore(key, sem)
    return actual.(*Semaphore)
}
```

### 43.6 Redis-Based Distributed Throttling

For multi-instance deployments, use Redis for shared concurrency tracking:

```go
type RedisThrottle struct {
    client    *redis.Client
    key       string
    maxSlots  int
    timeout   time.Duration
}

func (r *RedisThrottle) Acquire(ctx context.Context) error {
    lockKey := "throttle:" + r.key

    for {
        // Try to decrement the counter
        val, err := r.client.Decr(ctx, lockKey).Result()
        if err != nil {
            return fmt.Errorf("throttle redis decr: %w", err)
        }

        if val >= 0 {
            // Successfully acquired a slot
            r.client.Expire(ctx, lockKey, r.timeout)
            return nil
        }

        // No slots available — increment back and wait
        r.client.Incr(ctx, lockKey)

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(50 * time.Millisecond):
            // Retry
        }
    }
}

func (r *RedisThrottle) Release(ctx context.Context) {
    r.client.Incr(ctx, "throttle:"+r.key)
}
```

### 43.7 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `THROTTLE_ENABLED` | `true` | Enable/disable throttling |
| `THROTTLE_MAX_CONCURRENT` | `100` | Max concurrent requests (global mode) |
| `THROTTLE_MAX_PER_USER` | `10` | Max concurrent requests per user (per-user mode) |
| `THROTTLE_TIMEOUT` | `30s` | Max wait time for a throttle slot |
| `THROTTLE_MODE` | `global` | Throttling mode: `global`, `per-route`, `per-user` |
| `THROTTLE_REDIS_ENABLED` | `false` | Use Redis for distributed throttling |

### 43.8 Throttle Response

```json
{
    "success": false,
    "error": {
        "code": "THROTTLED",
        "message": "Too many concurrent requests. Please try again later."
    }
}
```

Headers:
```text
Retry-After: 30
X-Throttle-Limit: 100
X-Throttle-Remaining: 0
```

### 43.9 Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `throttle_acquired_total` | Counter | key, mode | Throttle slots acquired |
| `throttle_rejected_total` | Counter | key, mode | Throttle slots rejected |
| `throttle_active` | Gauge | key, mode | Currently active throttle slots |
| `throttle_wait_duration_seconds` | Histogram | key, mode | Time spent waiting for a throttle slot |

### 43.10 Relationship with Rate Limiting

```text
Request Flow:
    │
    ├── Rate Limiting: "Have you exceeded your request quota?"
    │       │
    │       ├── Yes → HTTP 429
    │       └── No  → Continue
    │
    └── Throttling: "Is the system currently overloaded?"
            │
            ├── Yes → HTTP 503 (with Retry-After)
            └── No  → Process request
```

Both middleware layers operate independently. A request must pass both checks to be processed.

### 43.11 Rules

1. Throttling is separate from rate limiting — do not conflate the two.
2. Throttling protects backend resources; rate limiting protects against abuse.
3. Always return `Retry-After` header when throttling.
4. Log throttled requests at WARN level with request_id and throttle key.
5. Monitor throttle metrics to tune limits based on real traffic.
6. In per-user mode, unauthenticated requests use IP-based throttling.
7. Throttling must not cause cascading failures — fail fast with clear error message.

---

## 43. Caching Mechanism

### 43.1 Problem

Certain data is:
- **Small in volume** (user permissions, feature flags, configuration, lookup tables)
- **High in access frequency** (read on almost every request)
- **Expensive to compute or fetch** (database queries, external API calls)

A caching layer reduces database load and improves response times for frequently accessed data.

### 43.2 Caching Strategy

```text
Application Code
    │
    ▼
Cache Interface
    │
    ├── In-Memory Cache (default, zero-dependency)
    │       │
    │       └── sync.RWMutex + map[string]CacheEntry
    │
    └── Redis Cache (optional, distributed)
            │
            └── go-redis client
```

### 43.3 Cache Interface

```go
// internal/cache/cache.go
package cache

import "context"

type Cache interface {
    Get(ctx context.Context, key string, dest any) error
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    DeletePattern(ctx context.Context, pattern string) error
    Close() error
}
```

### 43.4 In-Memory Cache

```go
// internal/cache/memory.go
package cache

import (
    "context"
    "encoding/json"
    "sync"
    "time"
)

type MemoryCache struct {
    mu      sync.RWMutex
    entries map[string]memoryEntry
    stopCh  chan struct{}
}

type memoryEntry struct {
    data      []byte
    expiresAt time.Time
}

func NewMemoryCache() *MemoryCache {
    c := &MemoryCache{
        entries: make(map[string]memoryEntry),
        stopCh:  make(chan struct{}),
    }
    go c.cleanup()
    return c
}

func (c *MemoryCache) Get(ctx context.Context, key string, dest any) error {
    c.mu.RLock()
    entry, exists := c.entries[key]
    c.mu.RUnlock()

    if !exists || time.Now().After(entry.expiresAt) {
        return ErrCacheMiss
    }

    return json.Unmarshal(entry.data, dest)
}

func (c *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("cache marshal: %w", err)
    }

    c.mu.Lock()
    c.entries[key] = memoryEntry{
        data:      data,
        expiresAt: time.Now().Add(ttl),
    }
    c.mu.Unlock()

    return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
    c.mu.Lock()
    delete(c.entries, key)
    c.mu.Unlock()
    return nil
}

func (c *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
    // Convert glob pattern to regex-like matching
    re := globToRegex(pattern)

    c.mu.Lock()
    for key := range c.entries {
        if re.MatchString(key) {
            delete(c.entries, key)
        }
    }
    c.mu.Unlock()
    return nil
}

func (c *MemoryCache) Close() error {
    close(c.stopCh)
    return nil
}

// cleanup periodically removes expired entries
func (c *MemoryCache) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            c.mu.Lock()
            now := time.Now()
            for key, entry := range c.entries {
                if now.After(entry.expiresAt) {
                    delete(c.entries, key)
                }
            }
            c.mu.Unlock()
        case <-c.stopCh:
            return
        }
    }
}
```

### 43.5 Redis Cache

```go
// internal/cache/redis.go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
    return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        if errors.Is(err, redis.Nil) {
            return ErrCacheMiss
        }
        return fmt.Errorf("redis cache get: %w", err)
    }

    return json.Unmarshal(data, dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("redis cache marshal: %w", err)
    }

    if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
        return fmt.Errorf("redis cache set: %w", err)
    }

    return nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
    var cursor uint64
    for {
        keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil {
            return fmt.Errorf("redis cache scan: %w", err)
        }

        if len(keys) > 0 {
            if err := c.client.Del(ctx, keys...).Err(); err != nil {
                return fmt.Errorf("redis cache del: %w", err)
            }
        }

        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }

    return nil
}

func (c *RedisCache) Close() error {
    return c.client.Close()
}
```

### 43.6 Cache Factory

```go
// internal/cache/factory.go
package cache

func NewCache(cfg config.Config) (Cache, error) {
    // If Redis is configured and enabled, use Redis cache
    if cfg.RedisURL != "" && cfg.CacheRedisEnabled {
        client := redis.NewClient(&redis.Options{
            Addr:         cfg.RedisURL,
            PoolSize:     10,
            MinIdleConns: 5,
        })

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := client.Ping(ctx).Err(); err != nil {
            log.Warn("redis unavailable, falling back to in-memory cache", "error", err)
            return NewMemoryCache(), nil
        }

        log.Info("using redis cache", "url", cfg.RedisURL)
        return NewRedisCache(client), nil
    }

    log.Info("using in-memory cache")
    return NewMemoryCache(), nil
}
```

### 43.7 Cache-Aside Pattern Helper

```go
// internal/cache/aside.go
package cache

import (
    "context"
    "time"
)

// GetOrLoad retrieves from cache, or loads from source and caches the result.
func GetOrLoad[T any](ctx context.Context, c Cache, key string, ttl time.Duration, load func() (T, error)) (T, error) {
    var dest T

    // Try cache first
    err := c.Get(ctx, key, &dest)
    if err == nil {
        return dest, nil
    }

    // Cache miss — load from source
    value, err := load()
    if err != nil {
        return zero[T](), err
    }

    // Store in cache (best-effort, don't fail the request)
    _ = c.Set(ctx, key, value, ttl)

    return value, nil
}

func zero[T any]() T {
    var v T
    return v
}
```

Usage:

```go
// In service or repository
permissions, err := cache.GetOrLoad(ctx, s.cache, "rbac:permissions:"+userID, 5*time.Minute, func() ([]string, error) {
    return s.repo.GetUserPermissions(ctx, userID)
})
```

### 43.8 Use Cases

| Use Case | Cache Key | TTL | In-Memory | Redis | Notes |
|----------|-----------|-----|-----------|-------|-------|
| RBAC permissions | `rbac:permissions:{user_id}` | 5 min | ✅ | ✅ | Must be invalidated on role change |
| Feature flags | `flags:{flag_name}` | 1 min | ✅ | ✅ | Low TTL for fast propagation |
| User profile (frequent read) | `user:profile:{user_id}` | 2 min | ✅ | ✅ | Invalidate on update |
| Lookup tables | `lookup:{table_name}` | 10 min | ✅ | ✅ | Rarely change |
| Rate limit counters | `ratelimit:{route}:{key}` | 1 min | ❌ | ✅ | Must be shared across instances |
| Session data | `session:{session_id}` | 30 min | ❌ | ✅ | Must be shared across instances |
| API response cache | `api:response:{method}:{path}:{hash}` | 30 sec | ✅ | ❌ | Short-lived, instance-specific |

### 43.9 Cache Invalidation

```go
// Invalidation helpers
func InvalidateUserCache(ctx context.Context, c Cache, userID string) error {
    // Invalidate user profile
    _ = c.Delete(ctx, "user:profile:"+userID)

    // Invalidate RBAC permissions
    _ = c.Delete(ctx, "rbac:permissions:"+userID)

    // Invalidate any user-specific API responses
    _ = c.DeletePattern(ctx, "api:response:*:user:"+userID+":*")

    return nil
}

func InvalidateRolePermissions(ctx context.Context, c Cache, roleID string) error {
    // This requires knowing which users have this role
    // Typically handled by the RBAC service
    _ = c.DeletePattern(ctx, "rbac:permissions:*")

    return nil
}
```

### 43.10 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CACHE_ENABLED` | `true` | Enable/disable caching |
| `CACHE_TYPE` | `memory` | Cache type: `memory` or `redis` |
| `CACHE_DEFAULT_TTL` | `5m` | Default TTL for cached entries |
| `CACHE_MAX_ENTRIES` | `10000` | Max entries in memory cache (0 = unlimited) |
| `CACHE_REDIS_ENABLED` | `false` | Use Redis for distributed caching |
| `CACHE_REDIS_URL` | — | Redis connection URL for cache |

### 43.11 Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `cache_hits_total` | Counter | cache_type, key_prefix | Cache hits |
| `cache_misses_total` | Counter | cache_type, key_prefix | Cache misses |
| `cache_hit_ratio` | Gauge | cache_type | Cache hit ratio (computed) |
| `cache_entries` | Gauge | cache_type | Current number of cached entries |
| `cache_memory_bytes` | Gauge | — | Memory used by in-memory cache |
| `cache_operation_duration_seconds` | Histogram | cache_type, operation | Cache operation latency |

### 43.12 Rules

1. Cache is a **performance optimization**, not a source of truth. Always handle cache misses gracefully.
2. Use short TTLs for data that changes frequently. Use longer TTLs for stable data.
3. Always implement cache invalidation when data changes.
4. Never cache sensitive data (passwords, tokens, PII) in Redis without encryption.
5. Monitor cache hit ratio — below 50% suggests the cache configuration needs tuning.
6. In-memory cache is per-instance — do not use for data that must be consistent across instances.
7. Redis cache adds network latency (~1-5ms) — use in-memory for ultra-low-latency needs.
8. Use `DeletePattern` sparingly — it scans all keys and can be expensive.
9. Always set a TTL — never cache indefinitely.
10. Test cache behavior: verify hit, miss, expiration, and invalidation scenarios.
