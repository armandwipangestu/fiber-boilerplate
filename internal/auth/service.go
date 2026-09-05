package auth

import (
	"context"
	"errors"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
	"github.com/armandwipangestu/fiber-boilerplate/internal/user"
)

var (
	ErrInvalidCredentials = pkg.NewAppError("INVALID_CREDENTIALS", "email or password is incorrect", 401, nil)
	ErrSessionNotFound    = pkg.NewAppError("SESSION_NOT_FOUND", "session not found", 401, nil)
	ErrSessionRevoked     = pkg.NewAppError("SESSION_REVOKED", "session has been revoked", 401, nil)
	ErrSessionExpired     = pkg.NewAppError("SESSION_EXPIRED", "session has expired", 401, nil)
)

// TokenResult is returned by successful auth flows.
type TokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
	User         *user.Domain
}

// Service orchestrates authentication flows.
type Service struct {
	users    user.Repository
	sessions RefreshTokenRepository
	cfg      config.Config
}

// NewService builds the auth service.
func NewService(users user.Repository, sessions RefreshTokenRepository, cfg config.Config) *Service {
	return &Service{users: users, sessions: sessions, cfg: cfg}
}

// Register creates a user (duplicate email -> 409) and issues a token pair.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*TokenResult, error) {
	exists, err := s.users.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, user.ErrEmailExists
	}

	hash, err := user.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	created, err := s.users.Create(ctx, &user.Domain{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: hash,
	})
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, created)
}

// Login verifies credentials and issues a token pair.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenResult, error) {
	u, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.VerifyPassword(u.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, u)
}

// Refresh rotates a refresh token. Reuse of a revoked token revokes the whole
// family to limit session-hijack damage.
func (s *Service) Refresh(ctx context.Context, rawToken string) (*TokenResult, error) {
	claims, err := ValidateToken(rawToken, s.cfg)
	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			return nil, ErrSessionExpired
		}
		return nil, ErrInvalidCredentials
	}
	if claims.Type != TokenTypeRefresh {
		return nil, ErrInvalidCredentials
	}

	hash := HashToken(rawToken)
	stored, err := s.sessions.GetByTokenHash(ctx, hash)
	if errors.Is(err, ErrRefreshTokenMissing) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	if stored.RevokedAt != nil || stored.ExpiresAt.Before(time.Now()) {
		// Reused or stale token: revoke the entire family.
		_ = s.sessions.DeleteByFamilyID(ctx, stored.FamilyID)
		return nil, ErrSessionRevoked
	}

	u, err := s.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	// Rotation: revoke the consumed token so its reuse triggers family wipe.
	if err := s.sessions.RevokeByTokenHash(ctx, hash); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, u)
}

// Logout deletes the presented refresh token.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return s.sessions.DeleteByTokenHash(ctx, HashToken(rawToken))
}

func (s *Service) issueTokens(ctx context.Context, u *user.Domain) (*TokenResult, error) {
	access, err := GenerateAccessToken(u.ID, s.cfg)
	if err != nil {
		return nil, err
	}

	refresh, familyID, err := GenerateRefreshToken(u.ID, s.cfg)
	if err != nil {
		return nil, err
	}

	err = s.sessions.Create(ctx, &RefreshToken{
		UserID:    u.ID,
		TokenHash: HashToken(refresh),
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(s.cfg.JWTRefreshTokenExpiry),
	})
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    s.cfg.JWTAccessTokenExpiry,
		User:         u,
	}, nil
}
