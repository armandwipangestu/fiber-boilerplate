package seed

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/user"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func seedDB(t *testing.T) *sql.DB {
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
	return db
}

func TestDefaultAdminSeededOnceAndIdempotent(t *testing.T) {
	db := seedDB(t)
	ctx := context.Background()

	email := fmt.Sprintf("admin+%d@example.com", time.Now().UnixNano())
	password := "Admin-Str0ng!"
	cfg := config.Config{DefaultAdminEmail: email, DefaultAdminPassword: password}

	// First run creates the account and grants the admin role.
	created, err := DefaultAdmin(ctx, db, cfg, nil)
	if err != nil {
		t.Fatalf("first DefaultAdmin: %v", err)
	}
	if !created {
		t.Fatal("expected account to be created on first seed")
	}
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM users WHERE email = $1`, email) })

	var hash string
	if err := db.QueryRow(`SELECT password_hash FROM users WHERE email = $1`, email).Scan(&hash); err != nil {
		t.Fatalf("read seeded user: %v", err)
	}
	if !user.VerifyPassword(hash, password) {
		t.Fatal("seeded password does not verify")
	}
	if user.VerifyPassword(hash, "wrong") {
		t.Fatal("wrong password verified")
	}

	var assigned bool
	if err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			JOIN users u ON u.id = ur.user_id
			WHERE u.email = $1 AND r.name = 'admin'
		)`, email).Scan(&assigned); err != nil {
		t.Fatalf("verify role: %v", err)
	}
	if !assigned {
		t.Fatal("admin role was not assigned to seeded user")
	}

	// Second run must be a no-op that keeps the existing password hash.
	created, err = DefaultAdmin(ctx, db, cfg, nil)
	if err != nil {
		t.Fatalf("second DefaultAdmin: %v", err)
	}
	if created {
		t.Fatal("second seed must not create the account again")
	}
	var hash2 string
	if err := db.QueryRow(`SELECT password_hash FROM users WHERE email = $1`, email).Scan(&hash2); err != nil {
		t.Fatalf("read user after second seed: %v", err)
	}
	if hash2 != hash {
		t.Fatal("existing account password was overwritten by re-seed")
	}
}

func TestDefaultAdminSkipsWhenNotConfigured(t *testing.T) {
	db := seedDB(t)
	cfg := config.Config{}
	created, err := DefaultAdmin(context.Background(), db, cfg, nil)
	if err != nil {
		t.Fatalf("DefaultAdmin with empty config: %v", err)
	}
	if created {
		t.Fatal("empty config must not create anything")
	}
}
