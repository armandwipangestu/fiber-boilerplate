package console

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

func TestFeatureNameNormalization(t *testing.T) {
	tests := []struct {
		in            string
		id, pkg, tbl  string
		plural, upper string
	}{
		{"user", "User", "user", "users", "Users", "USER"},
		{"User", "User", "user", "users", "Users", "USER"},
		{"book", "Book", "book", "books", "Books", "BOOK"},
		{"category", "Category", "category", "categories", "Categories", "CATEGORY"},
		{"post", "Post", "post", "posts", "Posts", "POST"},
		{"user_profile", "UserProfile", "user_profile", "user_profiles", "UserProfiles", "USER_PROFILE"},
		{"RolePermission", "RolePermission", "role_permission", "role_permissions", "RolePermissions", "ROLE_PERMISSION"},
	}
	for _, tt := range tests {
		got, err := newFeatureName(tt.in)
		if err != nil {
			t.Fatalf("newFeatureName(%q): %v", tt.in, err)
		}
		if got.ID != tt.id || got.Pkg != tt.pkg || got.Table != tt.tbl ||
			got.Plural != tt.plural || got.Upper != tt.upper {
			t.Errorf("newFeatureName(%q) = %+v, want ID=%s pkg=%s table=%s plural=%s upper=%s",
				tt.in, got, tt.id, tt.pkg, tt.tbl, tt.plural, tt.upper)
		}
	}
}

func TestFeatureNameRejectsEmpty(t *testing.T) {
	if _, err := newFeatureName("  "); err == nil {
		t.Fatal("expected an error for an empty name")
	}
}

func TestNextMigrationNumber(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "migrations"), 0o755)
	os.WriteFile(filepath.Join(dir, "migrations", "000007_seed_rbac.up.sql"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "migrations", "000012_created.up.sql"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "migrations", "readme.md"), []byte("x"), 0o644)
	if got := nextMigrationNumber(dir); got != 13 {
		t.Fatalf("nextMigrationNumber = %d, want 13", got)
	}

	empty := t.TempDir()
	os.MkdirAll(filepath.Join(empty, "migrations"), 0o755)
	if got := nextMigrationNumber(empty); got != 1 {
		t.Fatalf("nextMigrationNumber(empty) = %d, want 1", got)
	}
}

// TestRouteList exercises the route:list logic without touching a database.
func TestRouteList(t *testing.T) {
	var buf strings.Builder
	oldOut := out
	out = &buf
	t.Cleanup(func() { out = oldOut })

	cfg := config.Config{AppName: "test"}
	if err := RouteList(cfg, false); err != nil {
		t.Fatalf("RouteList: %v", err)
	}
	for path, wantMethod := range map[string]string{
		"/ping":              "GET",
		"/version":           "GET",
		"/api/v1/users/":     "POST",
		"/api/v1/users/:id":  "GET",
		"/health/live":       "GET",
		"/api/v1/auth/login": "POST",
	} {
		found := false
		for _, line := range strings.Split(buf.String(), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), wantMethod) && strings.Contains(line, path) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("route %s (%s) missing; got:\n%s", wantMethod, path, buf.String())
		}
	}
	if strings.Contains(buf.String(), "HEAD") {
		t.Errorf("HEAD routes should be hidden from route list:\n%s", buf.String())
	}
}

func TestRouteListJSON(t *testing.T) {
	var buf strings.Builder
	oldOut := out
	out = &buf
	t.Cleanup(func() { out = oldOut })

	cfg := config.Config{AppName: "test"}
	if err := RouteList(cfg, true); err != nil {
		t.Fatalf("RouteList(json): %v", err)
	}
	if !strings.Contains(buf.String(), `"method": "GET"`) {
		t.Errorf("expected JSON route output, got:\n%s", buf.String())
	}
}
