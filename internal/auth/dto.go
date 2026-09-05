package auth

import "github.com/armandwipangestu/fiber-boilerplate/internal/user"

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest carries no body; the token comes from the httpOnly cookie.

type AuthResponse struct {
	TokenType   string            `json:"token_type"`
	AccessToken string            `json:"access_token"`
	ExpiresIn   int64             `json:"expires_in"`
	User        user.UserResponse `json:"user"`
}
