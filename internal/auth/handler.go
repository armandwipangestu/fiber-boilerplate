package auth

import (
	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
	"github.com/armandwipangestu/fiber-boilerplate/internal/user"
)

const refreshCookieName = "refresh_token"

// Handler exposes the auth flows over HTTP.
type Handler struct {
	svc      *Service
	validate pkg.Validator
}

// NewHandler builds the auth handler.
func NewHandler(svc *Service, validate pkg.Validator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// RegisterRoutes mounts public auth endpoints.
func (h *Handler) RegisterRoutes(v1 fiber.Router) {
	authGroup := v1.Group("/auth")

	authGroup.Post("/register", h.Register)
	authGroup.Post("/login", h.Login)
	authGroup.Post("/refresh", h.Refresh)
	authGroup.Post("/logout", h.Logout)
}

// Register handles POST /auth/register.
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}
	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}

	result, err := h.svc.Register(c.Context(), req)
	if err != nil {
		return pkg.Error(c, err)
	}
	h.setRefreshCookie(c, result.RefreshToken)
	return pkg.Created(c, toAuthResponse(result))
}

// Login handles POST /auth/login.
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}
	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}

	result, err := h.svc.Login(c.Context(), req)
	if err != nil {
		return pkg.Error(c, err)
	}
	h.setRefreshCookie(c, result.RefreshToken)
	return pkg.OK(c, toAuthResponse(result))
}

// Refresh handles POST /auth/refresh.
func (h *Handler) Refresh(c *fiber.Ctx) error {
	token := c.Cookies(refreshCookieName)
	result, err := h.svc.Refresh(c.Context(), token)
	if err != nil {
		return pkg.Error(c, err)
	}
	h.setRefreshCookie(c, result.RefreshToken)
	return pkg.OK(c, toAuthResponse(result))
}

// Logout handles POST /auth/logout.
func (h *Handler) Logout(c *fiber.Ctx) error {
	token := c.Cookies(refreshCookieName)
	if err := h.svc.Logout(c.Context(), token); err != nil {
		return pkg.Error(c, err)
	}
	clearRefreshCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) setRefreshCookie(c *fiber.Ctx, token string) {
	cookie := new(fiber.Cookie)
	cookie.Name = refreshCookieName
	cookie.Value = token
	cookie.Path = "/api/v1/auth"
	cookie.HTTPOnly = true
	cookie.Secure = false // set true behind TLS; dev only
	cookie.SameSite = "Lax"
	cookie.MaxAge = int(h.svc.cfg.JWTRefreshTokenExpiry.Seconds())
	c.Cookie(cookie)
}

func clearRefreshCookie(c *fiber.Ctx) {
	cookie := new(fiber.Cookie)
	cookie.Name = refreshCookieName
	cookie.Value = ""
	cookie.Path = "/api/v1/auth"
	cookie.HTTPOnly = true
	cookie.MaxAge = -1
	c.Cookie(cookie)
}

func toAuthResponse(r *TokenResult) AuthResponse {
	return AuthResponse{
		TokenType:   "bearer",
		AccessToken: r.AccessToken,
		ExpiresIn:   int64(r.ExpiresIn.Seconds()),
		User: user.UserResponse{
			ID:        r.User.ID,
			Email:     r.User.Email,
			Name:      r.User.Name,
			CreatedAt: r.User.CreatedAt,
			UpdatedAt: r.User.UpdatedAt,
		},
	}
}
