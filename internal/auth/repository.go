package auth

import "context"

// RefreshTokenRepository persists session tokens by their hash.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteByFamilyID(ctx context.Context, familyID string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}
