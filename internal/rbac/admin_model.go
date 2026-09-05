package rbac

import "time"

// Role is a named set of permissions assignable to users.
type Role struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permission is an individual capability a role can grant.
type Permission struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
