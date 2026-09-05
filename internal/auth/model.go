package auth

import "time"

// RefreshToken is a stored session token. Only the sha-256 hash of the raw
// JWT is persisted, never the token itself.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	FamilyID  string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}
