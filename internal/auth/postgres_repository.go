package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// ErrRefreshTokenMissing is returned by GetByTokenHash when no session row
// matches. It is a plain error; the service maps it to ErrSessionNotFound.
var ErrRefreshTokenMissing = errors.New("refresh token not found")

type postgresRefreshRepo struct {
	db *sql.DB
}

// NewPostgresRefreshTokenRepository builds a RefreshTokenRepository on *sql.DB.
func NewPostgresRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return &postgresRefreshRepo{db: db}
}

// HashToken returns the hex sha-256 digest of a raw refresh token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (r *postgresRefreshRepo) Create(ctx context.Context, t *RefreshToken) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at, revoked_at)
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, t.UserID, t.TokenHash, t.FamilyID, t.ExpiresAt, nil)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *postgresRefreshRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, family_id, expires_at, revoked_at, created_at
	          FROM refresh_tokens WHERE token_hash = $1`

	var t RefreshToken
	var revoked sql.NullTime
	err := r.db.QueryRowContext(ctx, query, tokenHash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &t.FamilyID, &t.ExpiresAt, &revoked, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRefreshTokenMissing
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	if revoked.Valid {
		t.RevokedAt = &revoked.Time
	}
	return &t, nil
}

func (r *postgresRefreshRepo) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return ErrRefreshTokenMissing
	}
	return nil
}

func (r *postgresRefreshRepo) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete refresh token by hash: %w", err)
	}
	return nil
}

func (r *postgresRefreshRepo) DeleteByFamilyID(ctx context.Context, familyID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE family_id = $1`, familyID)
	if err != nil {
		return fmt.Errorf("delete refresh token family: %w", err)
	}
	return nil
}

func (r *postgresRefreshRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete refresh tokens by user: %w", err)
	}
	return nil
}

func (r *postgresRefreshRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at < $1`, time.Now())
	if err != nil {
		return fmt.Errorf("delete expired refresh tokens: %w", err)
	}
	return nil
}
