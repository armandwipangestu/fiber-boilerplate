# Phase 10 — RBAC (Roles, Permissions, Middleware)

> Scope: role/permission models, a cache-accelerated permission service,
> route-level permission middleware, and seeding already shipped in
> migration `000007`.

## Deliverables

### 10.1 Models
- `internal/role/model.go` — `Role{ID, Name, Description, CreatedAt, UpdatedAt}`.
- `internal/permission/model.go` — `Permission{ID, Name, Description, CreatedAt}`.

### 10.2 RBAC Service (`internal/rbac/service.go`)
- `HasPermission(ctx, userID, permission)` — cache first, DB on miss.
- `HasRole(ctx, userID, role)`, `GetUserPermissions(ctx, userID)`,
  `HasAnyPermission`, `IsSuperAdmin`.
- SQL joins `user_roles` → `roles` → `role_permissions` → `permissions`.

### 10.3 Permission Cache (`internal/rbac/cache.go`)
- `Cache` interface: `Get/Set Permissions`, `Get/Set Roles`, `Invalidate`,
  `InvalidateAll`.
- In-memory impl with a 5-minute TTL; keys `rbac:permissions:{user_id}`.
- Phase 18 layers Redis behind the same interface.

### 10.4 RBAC Middleware (`internal/middleware/rbac.go`)
- `RequirePermission(svc, perm)` → 403 when not granted.
- `RequireAnyPermission(svc, perms...)`, `RequireAllPermissions(svc, perms...)`.

### 10.5 Route Wiring
- `RouteOptions` in the user handler holds per-action permission handlers.
- `/api/v1/users`:
  - POST requires `users.create`
  - GET (list / by id) requires `users.view`
  - PATCH requires `users.update`
  - DELETE requires `users.delete`
- User service still re-checks `users.delete` defensively on `Delete`.

### 10.6 Seed Data
- Migration `000007` seeds roles `admin`/`user`, 8 permissions, admin gets all,
  user gets view-only. Already applied in Phase 4.

## How to Reproduce

```bash
# register/login a user, then:
# 1. without any role -> every /users 403
# 2. INSERT INTO user_roles (user_id, role_id) SELECT u.id, r.id FROM users u, roles r WHERE u.email='you@x.dev' AND r.name='admin';
#    (server restart or cache invalidation required) -> 200
```

## Verify
- `go test ./internal/rbac/...` passes.
- Manual: a user with no roles → 403; an admin → 200 on list/delete;
  self-delete → 422; cache invalidated when roles change via role management.