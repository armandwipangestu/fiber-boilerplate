package user

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
	"github.com/armandwipangestu/fiber-boilerplate/internal/storage"
)

// RouteOptions carries the middleware needed to mount user routes.
type RouteOptions struct {
	Auth          fiber.Handler
	RequireCreate fiber.Handler
	RequireView   fiber.Handler
	RequireUpdate fiber.Handler
	RequireDelete fiber.Handler
}

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
// All users routes require a valid access token plus a matching permission.
func (h *Handler) RegisterRoutes(v1 fiber.Router, opts RouteOptions) {
	users := v1.Group("/users", opts.Auth)

	users.Post("/", opts.RequireCreate, h.Create)
	users.Get("/", opts.RequireView, h.List)
	users.Get("/:id", opts.RequireView, h.GetByID)
	users.Patch("/:id", opts.RequireUpdate, h.Update)
	users.Delete("/:id", opts.RequireDelete, h.Delete)

	users.Post("/:id/avatar", opts.RequireUpdate, h.UploadAvatar)
	users.Delete("/:id/avatar", opts.RequireUpdate, h.ClearAvatar)
}

// Create handles POST /users.
// @Summary Create a user (admin)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User"
// @Success 201 {object} UserResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Router /users [post]
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
	return pkg.Created(c, h.resolveAvatar(u))
}

// List handles GET /users.
// @Summary List users with pagination
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param search query string false "Search by email or name"
// @Success 200 {object} ListUsersResponse
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Router /users [get]
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
		items = append(items, h.resolveAvatar(u))
	}
	return pkg.OK(c, ListUsersResponse{Items: items, Meta: meta})
}

// GetByID handles GET /users/:id.
// @Summary Get a user by ID
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /users/{id} [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid user id")
	}

	u, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, h.resolveAvatar(u))
}

// Update handles PATCH /users/:id.
// @Summary Update a user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "Fields to update"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /users/{id} [patch]
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
	return pkg.OK(c, h.resolveAvatar(u))
}

// Delete handles DELETE /users/:id.
// @Summary Delete a user
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /users/{id} [delete]
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

// UploadAvatar handles POST /users/:id/avatar (multipart field "avatar").
// @Summary Upload or replace a user avatar
// @Tags Users
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "User ID"
// @Param avatar formData file true "Avatar image (max 10MB)"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Router /users/{id}/avatar [post]
func (h *Handler) UploadAvatar(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid user id")
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return pkg.BadRequestResponse(c, "multipart field 'avatar' is required")
	}

	if file.Size > 10*1024*1024 {
		return pkg.BadRequestResponse(c, "avatar exceeds 10MB limit")
	}

	fh, err := file.Open()
	if err != nil {
		return pkg.BadRequestResponse(c, "could not read uploaded file")
	}
	defer fh.Close()

	data := make([]byte, file.Size)
	if _, err := fh.Read(data); err != nil {
		return pkg.BadRequestResponse(c, "could not read uploaded file")
	}

	u, err := h.svc.SetAvatar(c.Context(), id, storage.UploadOptions{
		Filename:    file.Filename,
		ContentType: file.Header.Get("Content-Type"),
		Data:        data,
	})
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, h.resolveAvatar(u))
}

// ClearAvatar handles DELETE /users/:id/avatar.
// @Summary Remove a user avatar
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Router /users/{id}/avatar [delete]
func (h *Handler) ClearAvatar(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid user id")
	}

	u, err := h.svc.ClearAvatar(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, toResponse(u))
}

func toResponse(u *Domain) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarKey,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// resolveAvatar maps a user's avatar storage key to a public URL using the
// handler's service.
func (h *Handler) resolveAvatar(u *Domain) UserResponse {
	resp := toResponse(u)
	resp.AvatarURL = h.svc.ResolveURL(u.AvatarKey)
	return resp
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
