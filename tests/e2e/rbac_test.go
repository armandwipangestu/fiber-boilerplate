package e2e

import (
	"net/http"
	"testing"
)

func TestRBACAdminAccess(t *testing.T) {
	token := adminToken(t, uniqueEmail("rbac-admin"))

	// List roles is gated by roles.view.
	resp, raw, err := api("GET", "/api/v1/roles", nil, token)
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list roles status %d: %s", resp.StatusCode, raw)
	}

	// Create a role, then expose its permissions, then delete it.
	resp, raw, err = api("POST", "/api/v1/roles", map[string]any{
		"name": "e2e-role", "description": "created by e2e",
	}, token)
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create role status %d: %s", resp.StatusCode, raw)
	}
	var role struct {
		ID string `json:"id"`
	}
	if err := unwrap(raw, &role); err != nil {
		t.Fatalf("create role decode: %v", err)
	}

	resp, raw, err = api("DELETE", "/api/v1/roles/"+role.ID, nil, token)
	if err != nil {
		t.Fatalf("delete role: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete role status %d: %s", resp.StatusCode, raw)
	}
}

func TestRBACForbiddenForRegularUser(t *testing.T) {
	userToken, _ := register(t, uniqueEmail("rbac-regular"), "Regular", "Passw0rd!")

	resp, raw, err := api("GET", "/api/v1/roles", nil, userToken)
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("regular user roles should 403, got %d: %s", resp.StatusCode, raw)
	}
}