package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/armandwipangestu/fiber-boilerplate/internal/user"
)

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()
	users := newMockUserRepo()
	sessions := &mockSessionRepo{}
	svc := NewService(users, sessions, testConfig())

	res, err := svc.Register(ctx, RegisterRequest{
		Email:    "new@example.com",
		Name:     "New User",
		Password: "supersecret",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	assert.Equal(t, "new@example.com", res.User.Email)
	require.Len(t, sessions.tokens, 1)
	assert.NotEqual(t, res.RefreshToken, sessions.tokens[0].TokenHash)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	users := newMockUserRepo()
	users.exists = true
	svc := NewService(users, &mockSessionRepo{}, testConfig())

	_, err := svc.Register(ctx, RegisterRequest{
		Email:    "dupe@example.com",
		Name:     "Dupe",
		Password: "supersecret",
	})
	assert.True(t, errors.Is(err, user.ErrEmailExists))
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	users := newMockUserRepo()
	hash, _ := user.HashPassword("correctpassword")
	users.byEmail["ada@example.com"] = &user.Domain{ID: "a", Email: "ada@example.com", Name: "Ada", PasswordHash: hash}

	svc := NewService(users, &mockSessionRepo{}, testConfig())
	res, err := svc.Login(ctx, LoginRequest{Email: "ada@example.com", Password: "correctpassword"})
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	assert.Equal(t, "a", res.User.ID)
}

func TestLogin_WrongPassword(t *testing.T) {
	ctx := context.Background()
	users := newMockUserRepo()
	hash, _ := user.HashPassword("correctpassword")
	users.byEmail["ada@example.com"] = &user.Domain{ID: "a", Email: "ada@example.com", Name: "Ada", PasswordHash: hash}

	svc := NewService(users, &mockSessionRepo{}, testConfig())
	_, err := svc.Login(ctx, LoginRequest{Email: "ada@example.com", Password: "wrong"})
	assert.True(t, errors.Is(err, ErrInvalidCredentials))
}

func TestLogin_UnknownUser(t *testing.T) {
	ctx := context.Background()
	svc := NewService(newMockUserRepo(), &mockSessionRepo{}, testConfig())
	_, err := svc.Login(ctx, LoginRequest{Email: "ghost@example.com", Password: "x"})
	assert.True(t, errors.Is(err, ErrInvalidCredentials))
}

func TestRefresh_Success(t *testing.T) {
	ctx := context.Background()
	users := newMockUserRepo()
	users.byID["a"] = &user.Domain{ID: "a", Email: "ada@example.com", Name: "Ada"}

	sessions := &mockSessionRepo{}
	raw, family, _ := GenerateRefreshToken("a", testConfig())
	sessions.tokens = append(sessions.tokens, &RefreshToken{
		UserID:    "a",
		TokenHash: HashToken(raw),
		FamilyID:  family,
		ExpiresAt: time.Now().Add(time.Hour),
	})

	svc := NewService(users, sessions, testConfig())
	res, err := svc.Refresh(ctx, raw)
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	require.Len(t, sessions.tokens, 2) // revoked old + new
	assert.Equal(t, HashToken(raw), sessions.tokens[0].TokenHash)
	assert.NotNil(t, sessions.tokens[0].RevokedAt) // old marked revoked
	assert.NotNil(t, sessions.tokens[1])
}

// Reusing a revoked token must wipe the family and reject.
func TestRefresh_ReuseDetected(t *testing.T) {
	ctx := context.Background()
	users := newMockUserRepo()
	users.byID["a"] = &user.Domain{ID: "a", Email: "ada@example.com", Name: "Ada"}

	sessions := &mockSessionRepo{}
	raw, family, _ := GenerateRefreshToken("a", testConfig())
	sessions.tokens = append(sessions.tokens, &RefreshToken{
		UserID:    "a",
		TokenHash: HashToken(raw),
		FamilyID:  family,
		ExpiresAt: time.Now().Add(time.Hour),
	})

	svc := NewService(users, sessions, testConfig())
	_, err := svc.Refresh(ctx, raw) // first use: OK
	require.NoError(t, err)

	// Reuse the already-rotated token -> SESSION_REVOKED + family wiped.
	_, err = svc.Refresh(ctx, raw)
	assert.True(t, errors.Is(err, ErrSessionRevoked))
}

func TestRefresh_RevokedToken(t *testing.T) {
	ctx := context.Background()
	users := newMockUserRepo()
	now := time.Now()

	sessions := &mockSessionRepo{}
	raw, family, _ := GenerateRefreshToken("a", testConfig())
	sessions.tokens = append(sessions.tokens, &RefreshToken{
		UserID:    "a",
		TokenHash: HashToken(raw),
		FamilyID:  family,
		ExpiresAt: now.Add(time.Hour),
		RevokedAt: &now,
	})

	svc := NewService(users, sessions, testConfig())
	_, err := svc.Refresh(ctx, raw)
	assert.True(t, errors.Is(err, ErrSessionRevoked))
}

func TestRefresh_AccessTokenRejected(t *testing.T) {
	ctx := context.Background()
	users := newMockUserRepo()
	svc := NewService(users, &mockSessionRepo{}, testConfig())

	access, _ := GenerateAccessToken("a", testConfig())
	_, err := svc.Refresh(ctx, access)
	assert.True(t, errors.Is(err, ErrInvalidCredentials))
}

func TestLogout(t *testing.T) {
	ctx := context.Background()
	raw, _, _ := GenerateRefreshToken("a", testConfig())

	sessions := &mockSessionRepo{}
	sessions.tokens = append(sessions.tokens, &RefreshToken{TokenHash: HashToken(raw)})

	svc := NewService(newMockUserRepo(), sessions, testConfig())
	require.NoError(t, svc.Logout(ctx, raw))
	assert.Empty(t, sessions.tokens)
}

type mockSessionRepo struct {
	tokens []*RefreshToken
}

func (m *mockSessionRepo) Create(ctx context.Context, t *RefreshToken) error {
	m.tokens = append(m.tokens, t)
	return nil
}

func (m *mockSessionRepo) GetByTokenHash(ctx context.Context, hash string) (*RefreshToken, error) {
	for _, t := range m.tokens {
		if t.TokenHash == hash {
			return t, nil
		}
	}
	return nil, ErrRefreshTokenMissing
}

func (m *mockSessionRepo) RevokeByTokenHash(ctx context.Context, hash string) error {
	for _, t := range m.tokens {
		if t.TokenHash == hash {
			now := time.Now()
			t.RevokedAt = &now
			return nil
		}
	}
	return ErrRefreshTokenMissing
}

func (m *mockSessionRepo) DeleteByTokenHash(ctx context.Context, hash string) error {
	out := m.tokens[:0]
	for _, t := range m.tokens {
		if t.TokenHash != hash {
			out = append(out, t)
		}
	}
	m.tokens = out
	return nil
}

func (m *mockSessionRepo) DeleteByFamilyID(ctx context.Context, family string) error {
	out := m.tokens[:0]
	for _, t := range m.tokens {
		if t.FamilyID != family {
			out = append(out, t)
		}
	}
	m.tokens = out
	return nil
}

func (m *mockSessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	out := m.tokens[:0]
	for _, t := range m.tokens {
		if t.UserID != userID {
			out = append(out, t)
		}
	}
	m.tokens = out
	return nil
}

func (m *mockSessionRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

type mockUserRepo struct {
	byID    map[string]*user.Domain
	byEmail map[string]*user.Domain
	exists  bool
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		byID:    map[string]*user.Domain{},
		byEmail: map[string]*user.Domain{},
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.Domain) (*user.Domain, error) {
	u.ID = "new-id"
	m.byID[u.ID] = u
	m.byEmail[u.Email] = u
	return u, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*user.Domain, error) {
	if u, ok := m.byID[id]; ok {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*user.Domain, error) {
	if u, ok := m.byEmail[email]; ok {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

func (m *mockUserRepo) List(ctx context.Context, q user.ListUsersQuery) ([]*user.Domain, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *user.Domain) (*user.Domain, error) {
	return u, nil
}

func (m *mockUserRepo) UpdateAvatar(ctx context.Context, id, key string) (*user.Domain, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.exists {
		return true, nil
	}
	_, ok := m.byEmail[email]
	return ok, nil
}
