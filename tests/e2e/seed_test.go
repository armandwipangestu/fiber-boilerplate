package e2e

import (
	"net/http"
	"testing"
)

// TestDefaultAdminLogin proves the seeded default admin can sign in and is
// granted the admin role without any manual user creation.
func TestDefaultAdminLogin(t *testing.T) {
	resp, raw, err := api("POST", "/api/v1/auth/login", map[string]any{
		"email":    "admin@example.com",
		"password": "Admin-e2e!",
	}, "")
	if err != nil {
		t.Fatalf("login as seeded admin: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status %d: %s", resp.StatusCode, raw)
	}
	var result authResult
	if err := unwrap(raw, &result); err != nil {
		t.Fatalf("login decode: %v (%s)", err, raw)
	}
	if result.AccessToken == "" {
		t.Fatal("no access token returned for seeded admin")
	}

	// The seeded account carries the admin role, so it can list users — an
	// operation gated by the users.view permission that only admin owns.
	resp, raw, err = api("GET", "/api/v1/users", nil, result.AccessToken)
	if err != nil {
		t.Fatalf("list users as seeded admin: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list users status %d: %s", resp.StatusCode, raw)
	}
}
