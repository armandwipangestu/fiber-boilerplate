package e2e

import (
	"net/http"
	"testing"
)

func TestUserCRUD(t *testing.T) {
	token := adminToken(t, uniqueEmail("admin"))

	name := "Created User"
	createResp, raw, err := api("POST", "/api/v1/users", map[string]any{
		"name": name, "email": uniqueEmail("crud"), "password": "Passw0rd!",
	}, token)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create user status %d: %s", createResp.StatusCode, raw)
	}
	var created struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := unwrap(raw, &created); err != nil {
		t.Fatalf("create decode: %v (%s)", err, raw)
	}

	// List is paginated and contains the new user.
	listResp, raw, err := api("GET", "/api/v1/users?page=1&per_page=10", nil, token)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list users status %d: %s", listResp.StatusCode, raw)
	}

	// Get by id.
	getResp, raw, err := api("GET", "/api/v1/users/"+created.ID, nil, token)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get user status %d: %s", getResp.StatusCode, raw)
	}

	// Update.
	updateResp, raw, err := api("PATCH", "/api/v1/users/"+created.ID, map[string]any{
		"name": "Renamed",
	}, token)
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update user status %d: %s", updateResp.StatusCode, raw)
	}
	var updated struct {
		Name string `json:"name"`
	}
	if err := unwrap(raw, &updated); err != nil {
		t.Fatalf("update decode: %v", err)
	}
	if updated.Name != "Renamed" {
		t.Fatalf("update did not change name: %s", raw)
	}

	// Delete.
	delResp, raw, err := api("DELETE", "/api/v1/users/"+created.ID, nil, token)
	if err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if delResp.StatusCode != http.StatusOK {
		t.Fatalf("delete user status %d: %s", delResp.StatusCode, raw)
	}

	// Gone after deletion.
	goneResp, raw, err := api("GET", "/api/v1/users/"+created.ID, nil, token)
	if err != nil {
		t.Fatalf("get deleted user: %v", err)
	}
	if goneResp.StatusCode != http.StatusNotFound {
		t.Fatalf("deleted user should 404, got %d: %s", goneResp.StatusCode, raw)
	}
}

func TestUserRequiresAdmin(t *testing.T) {
	userToken, _ := register(t, uniqueEmail("regular"), "Regular", "Passw0rd!")

	resp, raw, err := api("POST", "/api/v1/users", map[string]any{
		"name": "Sneaky", "email": uniqueEmail("sneaky"), "password": "Passw0rd!",
	}, userToken)
	if err != nil {
		t.Fatalf("create as regular user: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("regular user create should 403, got %d: %s", resp.StatusCode, raw)
	}
}

func TestUnauthorized(t *testing.T) {
	resp, raw, err := api("GET", "/api/v1/users?page=1&per_page=10", nil, "")
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token should 401, got %d: %s", resp.StatusCode, raw)
	}
}
