package user

import "context"

// Repository is the storage abstraction for the user feature.
type Repository interface {
	Create(ctx context.Context, u *Domain) (*Domain, error)
	GetByID(ctx context.Context, id string) (*Domain, error)
	GetByEmail(ctx context.Context, email string) (*Domain, error)
	List(ctx context.Context, query ListUsersQuery) ([]*Domain, int64, error)
	Update(ctx context.Context, u *Domain) (*Domain, error)
	// UpdateAvatar stores/replaces the avatar storage key; an empty key clears it.
	UpdateAvatar(ctx context.Context, id, key string) (*Domain, error)
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
