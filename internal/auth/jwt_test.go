package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		AppName:               "fiber-boilerplate",
		JWTSecret:             "test-secret-that-is-long-enough-0123456789abcdef",
		JWTAccessTokenExpiry:  15 * time.Minute,
		JWTRefreshTokenExpiry: 720 * time.Hour,
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	cfg := testConfig()
	userID := "4d4fdb0f-6adc-4822-8176-dbde5999bb72"

	tok, err := GenerateAccessToken(userID, cfg)
	require.NoError(t, err)

	claims, err := ValidateToken(tok, cfg)
	require.NoError(t, err)
	assert.Equal(t, TokenTypeAccess, claims.Type)
	assert.Equal(t, userID, claims.SubjectID())
	assert.Equal(t, userID, claims.Subject)
	assert.NotEmpty(t, claims.ID) // jti present
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	cfg := testConfig()
	userID := "4d4fdb0f-6adc-4822-8176-dbde5999bb72"

	tok, family, err := GenerateRefreshToken(userID, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, family)

	claims, err := ValidateToken(tok, cfg)
	require.NoError(t, err)
	assert.Equal(t, TokenTypeRefresh, claims.Type)
	assert.Equal(t, family, claims.FamilyID)
	assert.Equal(t, userID, claims.SubjectID())
}

func TestExpiredTokenRejected(t *testing.T) {
	cfg := testConfig()
	cfg.JWTAccessTokenExpiry = -1 * time.Minute

	tok, err := GenerateAccessToken("user-1", cfg)
	require.NoError(t, err)

	_, err = ValidateToken(tok, cfg)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestInvalidSignatureRejected(t *testing.T) {
	cfg := testConfig()

	tok, err := GenerateAccessToken("user-1", cfg)
	require.NoError(t, err)

	other := cfg
	other.JWTSecret = "different-secret-value-9876543210abcdef"
	_, err = ValidateToken(tok, other)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestClockSkewLeeway(t *testing.T) {
	cfg := testConfig()
	// Token valid for future not-before by <1s; leeway of 30s should accept.
	tok, err := GenerateAccessToken("user-1", cfg)
	require.NoError(t, err)

	_, err = ValidateToken(tok, cfg)
	assert.NoError(t, err)
}
