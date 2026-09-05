// Package seed bootstraps idempotent baseline data required to start using
// the API: e.g. an initial admin account that owns every permission.
package seed

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"

	"golang.org/x/crypto/bcrypt"
)

const defaultAdminName = "Administrator"

// DefaultAdmin provisions the account configured via DEFAULT_ADMIN_EMAIL /
// DEFAULT_ADMIN_PASSWORD and grants it the 'admin' role, which already
// carries every permission (see migration 000007_seed_rbac).
//
// It is a no-op when the environment variables are not set, and idempotent:
// an existing account is never modified, and role assignment is ON CONFLICT.
// Runs after migrations, so it fails fast if the schema is missing. Returns
// true when the account was created by this call.
func DefaultAdmin(ctx context.Context, db *sql.DB, cfg config.Config, logger *slog.Logger) (bool, error) {
	if cfg.DefaultAdminEmail == "" || cfg.DefaultAdminPassword == "" {
		return false, nil
	}
	if logger == nil {
		logger = slog.Default()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("hash default admin password: %w", err)
	}

	timeout := 15 * time.Second
	txCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tx, err := db.BeginTx(txCtx, nil)
	if err != nil {
		return false, fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback()

	created, err := ensureAdminUser(txCtx, tx, cfg.DefaultAdminEmail, defaultAdminName, string(hash))
	if err != nil {
		return created, err
	}

	if err := tx.Commit(); err != nil {
		return created, fmt.Errorf("commit seed tx: %w", err)
	}

	if created {
		logger.Info("seeded default admin user",
			"email", cfg.DefaultAdminEmail,
			"warning", "change DEFAULT_ADMIN_PASSWORD in production",
		)
	}
	return created, nil
}

func ensureAdminUser(ctx context.Context, tx *sql.Tx, email, name, hash string) (bool, error) {
	var userID string
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM users WHERE email = $1`, email,
	).Scan(&userID)
	if err == nil {
		return false, nil
	}
	if err != sql.ErrNoRows {
		return false, fmt.Errorf("lookup default admin: %w", err)
	}

	if err := tx.QueryRowContext(ctx,
		`INSERT INTO users (email, name, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		email, name, hash,
	).Scan(&userID); err != nil {
		return false, fmt.Errorf("insert default admin: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO user_roles (user_id, role_id)
		 SELECT $1, id FROM roles WHERE name = 'admin'
		 ON CONFLICT DO NOTHING`,
		userID,
	); err != nil {
		return true, fmt.Errorf("assign default admin role: %w", err)
	}

	return true, nil
}
