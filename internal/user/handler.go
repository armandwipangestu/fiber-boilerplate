package user

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

// Handler exposes the user feature over HTTP.
type Handler struct {
	svc      *Service
	validate pkg.Validator
}

// NewHandler builds the user handler.
func NewHandler(svc *Service, validate pkg.Validator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// RegisterRoutes mounts the user endpoints under the given router group.
// All users routes require a valid access token.
func (h *Handler) RegisterRoutes(v1 fiber.Router, authMW fiber.Handler) {
	users := v1.Group("/users", authMW)

	users.Post("/", h.Create)
	users.Get("/", h.List)
	users.Get("/:id", h.GetByID)
	users.Patch("/:id", h.Update)
	users.Delete("/:id", h.Delete)
}

// Create handles POST /users.
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}

	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}

	u, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Created(c, toResponse(u))
}

// List handles GET /users.
func (h *Handler) List(c *fiber.Ctx) error {
	query := ListUsersQuery{
		Page:    queryInt(c, "page", 1),
		PerPage: queryInt(c, "per_page", 20),
		Search:  c.Query("search"),
		SortBy:  c.Query("sort_by"),
		SortDir: c.Query("sort_dir"),
	}

	users, meta, err := h.svc.List(c.Context(), query)
	if err != nil {
		return pkg.Error(c, err)
	}

	items := make([]UserResponse, 0, len(users))
	for _, u := range users {
		items = append(items, toResponse(u))
	}
	return pkg.OK(c, ListUsersResponse{Items: items, Meta: meta})
}

// GetByID handles GET /users/:id.
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid user id")
	}

	u, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, toResponse(u))
}

// Update handles PATCH /users/:id.
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid user id")
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}

	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}

	u, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, toResponse(u))
}

// Delete handles DELETE /users/:id.
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid user id")
	}

	err := h.svc.Delete(c.Context(), pkg.GetUserID(c), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Message(c, fiber.StatusOK, "user deleted")
}

func toResponse(u *Domain) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func validID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func queryInt(c *fiber.Ctx, key string, fallback int) int {
	if v, err := strconv.Atoi(c.Query(key)); err == nil {
		return v
	}
	return fallback
}
