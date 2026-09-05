# Phase 5 — User Feature Domain

> Scope: the user domain model, DTOs, repository contract, PostgreSQL
> implementation, and password hashing.

## Deliverables

### 5.1 Model (`internal/user/model.go`)
- `Domain` struct: ID, Email, Name, PasswordHash, CreatedAt, UpdatedAt.

### 5.2 DTOs (`internal/user/dto.go`)
- `CreateUserRequest` — `email`, `name`, `password` with `validate` tags.
- `UpdateUserRequest` — pointer fields so PATCH only touches provided fields.
- `ListUsersQuery` — `page`, `per_page`, `search`, `sort_by`, `sort_dir`.
- Response types: `UserResponse`, `ListUsersResponse`, `PaginationMeta`.

### 5.3 Repository Contract (`internal/user/repository.go`)
- Interface exposing Create / GetByID / GetByEmail / List / Update / Delete / ExistsByEmail.

### 5.4 PostgreSQL Implementation (`internal/user/postgres_repository.go`)
- Raw SQL with `RETURNING` for inserts/updates.
- `GetByID`/`GetByEmail` map `sql.ErrNoRows` → `ErrUserNotFound`.
- `List` uses whitelisted sort columns and an `ILIKE` search with correct placeholder indexing.
- `ExistsByEmail` used by the service for 409 detection.

### 5.5 Feature Errors (`internal/user/errors.go`)
- `ErrUserNotFound` (404), `ErrEmailExists` (409), `ErrCannotSelfDelete` (422)
  — all share the `*pkg.AppError` envelope so handlers respond uniformly.

### 5.6 Password Hashing (`internal/user/password.go`)
- `HashPassword` via bcrypt (default cost).
- `VerifyPassword` returns bool (constant-time compare).

## How to Reproduce

```bash
go get golang.org/x/crypto
# Create files above, then:
go build ./internal/user/...
go test ./internal/user/...
```

## Verify
- `go test ./internal/user/...` passes (hash verify round-trip, search builder, sort whitelist).