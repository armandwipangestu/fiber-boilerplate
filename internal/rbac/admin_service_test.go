package rbac

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// adminDB opens the shared test database (local Postgres by default). Set
// TEST_DATABASE_URL to point elsewhere. Skips when unreachable.
func adminDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5433/fiber_boilerplate?sslmode=disable"
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Skipf("database not reachable, skipping integration test: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := migrationsApplied(db); err != nil {
		t.Fatalf("migrations missing: %v", err)
	}
	return db
}

func migrationsApplied(db *sql.DB) error {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'roles'`).Scan(&n)
	if err != nil {
		return err
	}
	_ = n
	return nil
}

func TestAdminRoleLifecycle(t *testing.T) {
	db := adminDB(t)
	ctx := context.Background()
	svc := NewService(db, NewInMemoryCache())

	name := "test-role-" + randNano()
	role, err := svc.CreateRole(ctx, name, "integration test role")
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, role.ID)
	})

	// duplicate name -> conflict
	if _, err := svc.CreateRole(ctx, name, "dup"); err == nil {
		t.Fatal("expected conflict for duplicate role name")
	}

	got, err := svc.GetRole(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRole: %v", err)
	}
	if got.Name != name {
		t.Fatalf("got name %q, want %q", got.Name, name)
	}

	desc := "updated description"
	if _, err := svc.UpdateRole(ctx, role.ID, nil, &desc); err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}

	roles, _, err := svc.ListRoles(ctx, ListQuery{Page: 1, PerPage: 20, Search: name})
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("expected 1 role in search, got %d", len(roles))
	}

	// admin role guard
	if err := svc.DeleteRole(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("expected not-found for bogus id")
	}

	if _, err := svc.GetRole(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("expected not-found for bogus id")
	}
}

func TestAdminPermissionLifecycle(t *testing.T) {
	db := adminDB(t)
	ctx := context.Background()
	svc := NewService(db, NewInMemoryCache())

	name := "test-perm-" + randNano()
	perm, err := svc.CreatePermission(ctx, name, "integration test permission")
	if err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM permissions WHERE id = $1`, perm.ID)
	})

	if _, err := svc.CreatePermission(ctx, name, "dup"); err == nil {
		t.Fatal("expected conflict for duplicate permission name")
	}

	perms, _, err := svc.ListPermissions(ctx, ListQuery{Page: 1, PerPage: 20, Search: name})
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission in search, got %d", len(perms))
	}

	if _, err := svc.GetPermission(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("expected not-found for bogus id")
	}
}

func TestAdminRolePermissionAssignment(t *testing.T) {
	db := adminDB(t)
	ctx := context.Background()
	svc := NewService(db, NewInMemoryCache())

	role, err := svc.CreateRole(ctx, "test-role-"+randNano(), "")
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	perm, err := svc.CreatePermission(ctx, "test-perm-"+randNano(), "")
	if err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, role.ID)
		_, _ = db.ExecContext(ctx, `DELETE FROM permissions WHERE id = $1`, perm.ID)
	})

	if err := svc.SetRolePermissions(ctx, role.ID, []string{perm.ID}); err != nil {
		t.Fatalf("SetRolePermissions: %v", err)
	}

	perms, err := svc.GetRolePermissions(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRolePermissions: %v", err)
	}
	if len(perms) != 1 || perms[0].ID != perm.ID {
		t.Fatalf("expected 1 permission, got %v", perms)
	}
}

func TestAdminUserRoleAssignment(t *testing.T) {
	db := adminDB(t)
	ctx := context.Background()
	svc := NewService(db, NewInMemoryCache())

	role, err := svc.CreateRole(ctx, "test-role-"+randNano(), "")
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	// create a disposable user
	var userID string
	err = db.QueryRowContext(ctx, `INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		"rbac-e2e-"+randNano()+"@example.com", "RBAC E2E", "x").Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
		_, _ = db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, role.ID)
	})

	if err := svc.AssignRole(ctx, userID, role.ID); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}
	roles, err := svc.GetUserRoles(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserRoles: %v", err)
	}
	if len(roles) != 1 || roles[0].ID != role.ID {
		t.Fatalf("expected assigned role, got %v", roles)
	}

	if err := svc.UnassignRole(ctx, userID, role.ID); err != nil {
		t.Fatalf("UnassignRole: %v", err)
	}
	roles, err = svc.GetUserRoles(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserRoles: %v", err)
	}
	if len(roles) != 0 {
		t.Fatalf("expected no roles after unassign, got %v", roles)
	}
}

func randNano() string {
	return nanostr()
}

func nanostr() string {
	b := make([]byte, 8)
	stamp := time.Now().UnixNano()
	for i := range b {
		b[i] = byte("0123456789abcdef"[stamp&0xf])
		stamp >>= 4
	}
	return string(b)
}