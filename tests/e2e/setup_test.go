package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/app"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
)

var (
	baseURL string
	client  *http.Client
	res     *app.Resources
)

// TestMain boots the full application against a dedicated database, applies
// migrations, runs the tests, then tears everything down. Set
// TEST_DATABASE_URL to override the default local database.
func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/fiber_boilerplate_e2e?sslmode=disable"
	}
	if err := ensureDatabase(dbURL); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: %v\n", err)
		os.Exit(1)
	}

	cfg := config.Config{
		AppName:                 "fiber-boilerplate-e2e",
		AppEnv:                  "test",
		AppPort:                 0,
		AppHost:                 "127.0.0.1",
		ShutdownTimeout:         5 * time.Second,
		DatabaseDriver:          "postgres",
		DatabaseURL:             dbURL,
		DatabaseMaxOpenConns:    10,
		DatabaseMaxIdleConns:    5,
		DatabaseConnMaxIdleTime: time.Minute,
		JWTSecret:               "e2e-test-secret",
		JWTAccessTokenExpiry:    15 * time.Minute,
		JWTRefreshTokenExpiry:   24 * time.Hour,
		RateLimitEnabled:        false,
		StoragePath:             os.TempDir() + "/fiber-e2e-storage",
		LogLevel:                "warn",
		DefaultAdminEmail:       "admin@example.com",
		DefaultAdminPassword:    "Admin-e2e!",
	}

	if err := database.RunMigrations(cfg, "up"); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: migrations failed: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var err error
	res, err = app.Build(logger, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: build failed: %v\n", err)
		os.Exit(1)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: listen failed: %v\n", err)
		os.Exit(1)
	}
	go func() {
		if err := res.App.Listener(ln); err != nil {
			fmt.Fprintf(os.Stderr, "e2e: serve failed: %v\n", err)
			os.Exit(1)
		}
	}()
	baseURL = "http://" + ln.Addr().String()

	client = &http.Client{Timeout: 10 * time.Second}
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: cookie jar: %v\n", err)
		os.Exit(1)
	}
	client.Jar = jar

	code := m.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = res.Shutdown(ctx)
	os.Exit(code)
}

// ensureDatabase creates the target database when missing by connecting to the
// same server's maintenance database.
func ensureDatabase(dbURL string) error {
	// Swap the path/database segment for the maintenance "postgres" db.
	maintURL := strings.Replace(dbURL, "/fiber_boilerplate_e2e", "/postgres", 1)
	dbName := "fiber_boilerplate_e2e"

	mdb, err := database.OpenConn(maintURL)
	if err != nil {
		return fmt.Errorf("connect maintenance db: %w", err)
	}
	defer mdb.Close()

	var exists bool
	if err := mdb.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`, dbName).Scan(&exists); err != nil {
		return fmt.Errorf("check database: %w", err)
	}
	if !exists {
		// CREATE DATABASE cannot run inside a transaction/parameterized query.
		if _, err := mdb.Exec(`CREATE DATABASE ` + dbName); err != nil {
			return fmt.Errorf("create database: %w", err)
		}
	}
	return nil
}

// api performs an HTTP request and returns the raw response plus body.
func api(method, path string, payload any, token string) (*http.Response, []byte, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	return resp, raw, err
}

// unwrap decodes the standard API envelope {"success":bool,"data":...}.
func unwrap(raw []byte, out any) error {
	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	if out != nil && len(envelope.Data) > 0 {
		return json.Unmarshal(envelope.Data, out)
	}
	return nil
}

type authResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	User        struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"user"`
}

// register logs a new user in and returns its token + user id.
func register(t *testing.T, email, name, password string) (string, string) {
	t.Helper()
	resp, raw, err := api("POST", "/api/v1/auth/register", map[string]any{
		"email": email, "name": name, "password": password,
	}, "")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register status %d: %s", resp.StatusCode, raw)
	}
	var result authResult
	if err := unwrap(raw, &result); err != nil {
		t.Fatalf("register decode: %v (%s)", err, raw)
	}
	return result.AccessToken, result.User.ID
}

// assignRole grants a named role to a user directly in the DB.
func assignRole(t *testing.T, userID, roleName string) {
	t.Helper()
	if res == nil {
		t.Fatal("app not booted")
	}
	var roleID string
	err := res.DB.QueryRow(`SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID)
	if err != nil {
		t.Fatalf("find role %q: %v", roleName, err)
	}
	if _, err := res.DB.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, roleID); err != nil {
		t.Fatalf("assign role: %v", err)
	}
}

// adminToken registers a user, grants it the admin role, logs in, and returns
// the access token.
func adminToken(t *testing.T, email string) string {
	t.Helper()
	password := "Sup3rSecret!"
	_, id := register(t, email, "Admin", password)
	assignRole(t, id, "admin")

	resp, raw, err := api("POST", "/api/v1/auth/login", map[string]any{
		"email": email, "password": password,
	}, "")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin login status %d: %s", resp.StatusCode, raw)
	}
	var result authResult
	if err := unwrap(raw, &result); err != nil {
		t.Fatalf("admin login decode: %v", err)
	}
	return result.AccessToken
}
