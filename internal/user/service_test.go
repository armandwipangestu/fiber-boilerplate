package user

import (
	"context"
	"errors"
	"testing"

	"github.com/armandwipangestu/fiber-boilerplate/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	svc := NewService(mock, nil)

	u, err := svc.Create(ctx, CreateUserRequest{Email: "new@example.com", Name: "Ada", Password: "supersecret"})
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", u.Email)
	assert.True(t, VerifyPassword(u.PasswordHash, "supersecret"))
	assert.NotEqual(t, "supersecret", u.PasswordHash)
}

func TestService_Create_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	mock.exists = true
	svc := NewService(mock, nil)

	_, err := svc.Create(ctx, CreateUserRequest{Email: "ada@example.com", Name: "Ada", Password: "supersecret"})
	assert.True(t, errors.Is(err, ErrEmailExists))
}

func TestService_Update_Partial(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()

	svc := NewService(mock, nil)

	name := "Ada Lovelace"
	u, err := svc.Update(ctx, mock.users[0].ID, UpdateUserRequest{Name: &name})
	require.NoError(t, err)
	assert.Equal(t, "Ada Lovelace", u.Name)
	assert.Equal(t, "ada@example.com", u.Email)

	pass := "newpassword"
	u2, err := svc.Update(ctx, mock.users[0].ID, UpdateUserRequest{Password: &pass})
	require.NoError(t, err)
	assert.True(t, VerifyPassword(u2.PasswordHash, "newpassword"))
}

func TestService_Update_NotOwnedEmail(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()

	mock.existingOther = true
	svc := NewService(mock, nil)

	other := "other@example.com"
	_, err := svc.Update(ctx, mock.users[0].ID, UpdateUserRequest{Email: &other})
	assert.True(t, errors.Is(err, ErrEmailExists))
}

func TestService_Delete_Self(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	svc := NewService(mock, nil)

	err := svc.Delete(ctx, "a", "a")
	assert.True(t, errors.Is(err, ErrCannotSelfDelete))
}

func TestService_Delete_Forbidden(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	svc := NewService(mock, nil) // nil rbac -> fail closed

	err := svc.Delete(ctx, "actor", "target")
	assert.True(t, errors.Is(err, ErrNoPermission))
}

func TestService_Delete_RBACDenies(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	svc := NewService(mock, stubRBAC{allowed: false})

	err := svc.Delete(ctx, "actor", "target")
	assert.True(t, errors.Is(err, ErrNoPermission))
}

func TestService_Delete_Success(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	svc := NewService(mock, stubRBAC{allowed: true})

	err := svc.Delete(ctx, "actor", "target")
	assert.NoError(t, err)
	assert.True(t, mock.deleted)
}

func TestService_List_Defaults(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	svc := NewService(mock, nil)

	users, meta, err := svc.List(ctx, ListUsersQuery{})
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 20, meta.PerPage)
	assert.Equal(t, 1, meta.TotalPages)
}

func TestService_SetAvatar_StoresAndCleansOld(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	mock.users[0].AvatarKey = "avatars/old.png"
	stub := &stubStorage{key: "avatars/new.png"}
	svc := NewServiceWithStorage(mock, nil, stub)

	u, err := svc.SetAvatar(ctx, "a", storage.UploadOptions{Filename: "pic.png", ContentType: "image/png", Data: []byte("x")})
	require.NoError(t, err)
	assert.Equal(t, "avatars/new.png", u.AvatarKey)
	assert.Equal(t, "avatars/old.png", stub.deleted, "old object should be cleaned up")
}

func TestService_SetAvatar_WithoutStorage(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	svc := NewService(mock, nil)

	_, err := svc.SetAvatar(ctx, "a", storage.UploadOptions{Filename: "pic.png"})
	assert.Error(t, err)
}

func TestService_ClearAvatar_RemovesObject(t *testing.T) {
	ctx := context.Background()
	mock := newMockRepo()
	mock.users[0].AvatarKey = "avatars/pic.png"
	stub := &stubStorage{}
	svc := NewServiceWithStorage(mock, nil, stub)

	u, err := svc.ClearAvatar(ctx, "a")
	require.NoError(t, err)
	assert.Equal(t, "", u.AvatarKey)
	assert.Equal(t, "avatars/pic.png", stub.deleted)
}

func TestService_ResolveURL_EmptyKey(t *testing.T) {
	svc := NewService(nil, nil)
	assert.Equal(t, "", svc.ResolveURL(""))
}

type stubStorage struct {
	key     string
	deleted string
}

func (s *stubStorage) Upload(ctx context.Context, opts storage.UploadOptions) (string, error) {
	if s.key != "" {
		return s.key, nil
	}
	return "avatars/" + opts.Filename, nil
}

func (s *stubStorage) Delete(ctx context.Context, key string) error {
	s.deleted = key
	return nil
}

func (s *stubStorage) GetURL(key string) string {
	return "https://cdn.example/" + key
}

type stubRBAC struct {
	allowed bool
}

func (s stubRBAC) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	return s.allowed, nil
}

type mockRepo struct {
	users         []*Domain
	exists        bool
	existingOther bool
	deleted       bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users: []*Domain{
			{ID: "a", Email: "ada@example.com", Name: "Ada", PasswordHash: "hash"},
			{ID: "b", Email: "grace@example.com", Name: "Grace", PasswordHash: "hash"},
		},
	}
}

func (m *mockRepo) Create(ctx context.Context, u *Domain) (*Domain, error) {
	u.ID = "new"
	return u, nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Domain, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*Domain, error) {
	if m.existingOther {
		return &Domain{ID: "other", Email: email, Name: "Other"}, nil
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockRepo) List(ctx context.Context, q ListUsersQuery) ([]*Domain, int64, error) {
	return m.users, int64(len(m.users)), nil
}

func (m *mockRepo) Update(ctx context.Context, u *Domain) (*Domain, error) {
	return u, nil
}

func (m *mockRepo) UpdateAvatar(ctx context.Context, id, key string) (*Domain, error) {
	for _, u := range m.users {
		if u.ID == id {
			u.AvatarKey = key
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockRepo) Delete(ctx context.Context, id string) error {
	m.deleted = true
	return nil
}

func (m *mockRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.exists || m.existingOther {
		return true, nil
	}
	for _, u := range m.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}
