# Phase 19 — Role & Permission Management

> HTTP administration for roles, permissions, and user-role assignments, with
> cache invalidation so privilege changes take effect immediately.

## Deliverables

### 19.1 Role management
- `internal/rbac/admin_repository.go` — PostgreSQL `Repository` with role CRUD,
  permission CRUD, role-permission and user-role assignment, user existence
  checks, shared pagination/search.
- Service (on `rbac.Service`) methods: `ListRoles`, `GetRole`, `CreateRole`,
  `UpdateRole`, `DeleteRole`, `GetRolePermissions`, `SetRolePermissions`.
- Guards: duplicate names → 409; the built-in `admin` role cannot be renamed or
  deleted.
- Routes (auth + `roles.view` / `roles.manage`):
  `GET|POST /roles`, `GET|PATCH|DELETE /roles/:id`,
  `GET /roles/:id/permissions`, `PUT /roles/:id/permissions`,
  `POST /roles/:id/users` (assign), `DELETE /roles/:id/users/:userId`.

### 19.2 Permission management
- Service methods: `ListPermissions`, `GetPermission`, `CreatePermission`,
  `UpdatePermission`, `DeletePermission`.
- Routes (auth + `permissions.view` / `permissions.manage`):
  `GET|POST /permissions`, `GET|PATCH|DELETE /permissions/:id`.
- Migration `000009` adds `updated_at` to `permissions` for API consistency.

### 19.3 User role visibility
- `GET /users/:id/roles` and `GET /users/:id/permissions` (requires
  `roles.view`) return a user's current grants.

### 19.4 Cache invalidation
- Role/permission edits and user-role assignment invalidate the shared cache
  (`Invalidate(userID)` for single users, `InvalidateAll` for model-wide
  changes) so the 5-minute memoization never serves stale grants.

## Verify
- DB-backed integration tests (`TEST_DATABASE_URL`, skip when unreachable):
  full role lifecycle, permission lifecycle, role↔permission assignment,
  user↔role assign/unassign.