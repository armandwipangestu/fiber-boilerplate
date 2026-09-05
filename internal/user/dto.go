package user

import "time"

// CreateUserRequest is the payload for creating a user.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// UpdateUserRequest uses pointers so omitted fields are not overwritten.
type UpdateUserRequest struct {
	Email    *string `json:"email" validate:"omitempty,email"`
	Name     *string `json:"name" validate:"omitempty,min=2,max=100"`
	Password *string `json:"password" validate:"omitempty,min=8,max=72"`
}

// ListUsersQuery are the query params for listing users.
type ListUsersQuery struct {
	Page    int    `query:"page"`
	PerPage int    `query:"per_page"`
	Search  string `query:"search"`
	SortBy  string `query:"sort_by"`
	SortDir string `query:"sort_dir"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListUsersResponse struct {
	Items []UserResponse `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}

// ListEntry is what the repository list query returns.
type ListEntry struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
