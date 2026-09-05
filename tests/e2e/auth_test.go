package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
}

func TestAuthFlow(t *testing.T) {
	email := uniqueEmail("flow")
	password := "Passw0rd!"
	regResp, raw, err := api("POST", "/api/v1/auth/register", map[string]any{
		"email": email, "name": "Flow Tester", "password": password,
	}, "")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if regResp.StatusCode != http.StatusCreated {
		t.Fatalf("register status %d: %s", regResp.StatusCode, raw)
	}

	loginResp, raw, err := api("POST", "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	}, "")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status %d: %s", loginResp.StatusCode, raw)
	}
	var result authResult
	if err := unwrap(raw, &result); err != nil {
		t.Fatalf("login decode: %v (%s)", err, raw)
	}
	if result.AccessToken == "" || result.User.ID == "" {
		t.Fatalf("login did not return token/user: %s", raw)
	}

	// Refresh rotates the token pair using the httpOnly cookie.
	refreshResp, raw, err := api("POST", "/api/v1/auth/refresh", nil, "")
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshResp.StatusCode != http.StatusOK {
		t.Fatalf("refresh status %d: %s", refreshResp.StatusCode, raw)
	}

	// Logout revokes the refresh token family.
	logoutResp, raw, err := api("POST", "/api/v1/auth/logout", nil, "")
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if logoutResp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout status %d: %s", logoutResp.StatusCode, raw)
	}

	// Refreshing after logout is rejected.
	reuseResp, raw, err := api("POST", "/api/v1/auth/refresh", nil, "")
	if err != nil {
		t.Fatalf("refresh after logout: %v", err)
	}
	if reuseResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("refresh after logout should be 401, got %d: %s", reuseResp.StatusCode, raw)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	email := uniqueEmail("dup")
	password := "Passw0rd!"

	resp, raw, err := api("POST", "/api/v1/auth/register", map[string]any{
		"email": email, "name": "First", "password": password,
	}, "")
	if err != nil {
		t.Fatalf("register 1: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register 1 status %d: %s", resp.StatusCode, raw)
	}

	resp2, raw2, err := api("POST", "/api/v1/auth/register", map[string]any{
		"email": email, "name": "Second", "password": password,
	}, "")
	if err != nil {
		t.Fatalf("register 2: %v", err)
	}
	if resp2.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate register status %d: %s", resp2.StatusCode, raw2)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	email := uniqueEmail("badpass")
	password := "Passw0rd!"

	if _, raw, err := api("POST", "/api/v1/auth/register", map[string]any{
		"email": email, "name": "Pass", "password": password,
	}, ""); err != nil {
		t.Fatalf("register: %v", err)
	} else if len(raw) == 0 {
		t.Fatal("no response")
	}

	resp, raw, err := api("POST", "/api/v1/auth/login", map[string]any{
		"email": email, "password": "wrong-password",
	}, "")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password login status %d: %s", resp.StatusCode, raw)
	}
}