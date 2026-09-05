package user

import (
	"context"
	"errors"
	"math"

	"github.com/armandwipangestu/fiber-boilerplate/internal/storage"
)

// RBACChecker is implemented by the rbac service and injected to keep the
// user feature decoupled from the RBAC internals.
type RBACChecker interface {
	HasPermission(ctx context.Context, userID string, permission string) (bool, error)
}

// Storage is the object-storage abstraction the user feature depends on. It
// is implemented by internal/storage (S3 with local fallback).
type Storage interface {
	Upload(ctx context.Context, opts storage.UploadOptions) (string, error)
	Delete(ctx context.Context, key string) error
	GetURL(key string) string
}

// Service contains the user feature's business rules.
type Service struct {
	repo    Repository
	rbac    RBACChecker
	storage Storage
}

// NewService builds the user service.
func NewService(repo Repository, rbac RBACChecker) *Service {
	return &Service{repo: repo, rbac: rbac}
}

// NewServiceWithStorage builds the user service with object-storage support.
func NewServiceWithStorage(repo Repository, rbac RBACChecker, storage Storage) *Service {
	return &Service{repo: repo, rbac: rbac, storage: storage}
}

// Create validates uniqueness, hashes the password, and persists the user.
func (s *Service) Create(ctx context.Context, req CreateUserRequest) (*Domain, error) {
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &Domain{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: hash,
	})
}

// GetByID fetches a single user.
func (s *Service) GetByID(ctx context.Context, id string) (*Domain, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a page of users plus pagination metadata.
func (s *Service) List(ctx context.Context, query ListUsersQuery) ([]*Domain, PaginationMeta, error) {
	q := query
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 || q.PerPage > 100 {
		q.PerPage = 20
	}

	users, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(q.PerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	return users, PaginationMeta{
		Page:       q.Page,
		PerPage:    q.PerPage,
		TotalItems: total,
		TotalPages: totalPages,
	}, nil
}

// Update applies only the provided fields.
func (s *Service) Update(ctx context.Context, id string, req UpdateUserRequest) (*Domain, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != nil {
		current.Email = *req.Email
		exists, err := s.repo.ExistsByEmail(ctx, current.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			// The same user updating their own (unchanged) email should not 409.
			owner, err := s.repo.GetByEmail(ctx, current.Email)
			if err == nil && owner.ID != current.ID {
				return nil, ErrEmailExists
			}
		}
	}
	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.Password != nil {
		hash, err := HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		current.PasswordHash = hash
	}

	return s.repo.Update(ctx, current)
}

// Delete removes a user after RBAC and self-deletion checks.
func (s *Service) Delete(ctx context.Context, actorID, targetID string) error {
	if actorID == targetID {
		return ErrCannotSelfDelete
	}

	if s.rbac == nil {
		return ErrNoPermission
	}
	allowed, err := s.rbac.HasPermission(ctx, actorID, "users.delete")
	if err != nil {
		return err
	}
	if !allowed {
		return ErrNoPermission
	}

	return s.repo.Delete(ctx, targetID)
}

// AvatarKey returns the stored avatar storage key ("" when unset).
func (s *Service) AvatarKey(u *Domain) string {
	return u.AvatarKey
}

// ResolveURL turns an avatar storage key into a public URL.
func (s *Service) ResolveURL(key string) string {
	if key == "" || s.storage == nil {
		return ""
	}
	return s.storage.GetURL(key)
}

// SetAvatar replaces the avatar for id. The previous object is deleted from
// storage before the new key is persisted.
func (s *Service) SetAvatar(ctx context.Context, id string, opts storage.UploadOptions) (*Domain, error) {
	if s.storage == nil {
		return nil, errors.New("storage is not configured")
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	opts.Folder = "avatars"
	key, err := s.storage.Upload(ctx, opts)
	if err != nil {
		return nil, err
	}

	if current.AvatarKey != "" && current.AvatarKey != key {
		_ = s.storage.Delete(ctx, current.AvatarKey) // best-effort cleanup
	}

	return s.repo.UpdateAvatar(ctx, id, key)
}

// ClearAvatar removes the avatar. The stored object is deleted and the column
// set to NULL.
func (s *Service) ClearAvatar(ctx context.Context, id string) (*Domain, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if current.AvatarKey != "" {
		_ = s.storage.Delete(ctx, current.AvatarKey) // best-effort cleanup
	}

	return s.repo.UpdateAvatar(ctx, id, "")
}

// IsNotFound reports whether err is the feature's not-found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
