package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

// Token type claim values.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token has expired")
	ErrWrongTokenType = errors.New("token type mismatch")
)

// TokenClaims embeds registered claims plus app-specific fields. The user id
// lives in Subject (claim "sub") to avoid duplicate json tags.
type TokenClaims struct {
	Type     string `json:"type"`
	FamilyID string `json:"family_id,omitempty"`
	jwt.RegisteredClaims
}

func (c *TokenClaims) SubjectID() string { return c.Subject }

// GenerateAccessToken signs a short-lived access token for a user.
func GenerateAccessToken(userID string, cfg config.Config) (string, error) {
	now := time.Now()
	claims := TokenClaims{
		Type: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.AppName,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{"api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWTAccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	return signToken(claims, cfg.JWTSecret)
}

// GenerateRefreshToken signs a long-lived refresh token carrying a family id
// used to detect token-reuse rotation (new login -> new family).
func GenerateRefreshToken(userID string, cfg config.Config) (string, string, error) {
	now := time.Now()
	familyID := uuid.NewString()
	claims := TokenClaims{
		Type:     TokenTypeRefresh,
		FamilyID: familyID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.AppName,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{"api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWTRefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	token, err := signToken(claims, cfg.JWTSecret)
	if err != nil {
		return "", "", err
	}
	return token, familyID, nil
}

// ValidateToken parses and verifies a signed token with clock-skew tolerance
// (30 seconds) and a 30-second leeway bound.
func ValidateToken(tokenString string, cfg config.Config) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(cfg.JWTSecret), nil
	},
		jwt.WithIssuer(cfg.AppName),
		jwt.WithAudience("api"),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func signToken(claims TokenClaims, secret string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
