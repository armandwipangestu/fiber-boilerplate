package user

import "time"

// Domain is the core user model used across the feature.
type Domain struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	AvatarKey    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
