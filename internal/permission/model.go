package permission

import "time"

// Permission is an individual capability grantable to a role.
type Permission struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}
