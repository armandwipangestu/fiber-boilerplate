package rbac

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

// RouteOptions carries the middleware needed to mount RBAC management routes.
type RouteOptions struct {
	Auth                     fiber.Handler
	RequireRolesView         fiber.Handler
	RequireRolesManage       fiber.Handler
	RequirePermissionsView   fiber.Handler
	RequirePermissionsManage fiber.Handler
}

// Handler exposes role and permission management over HTTP.
type Handler struct {
	svc      *Service
	validate pkg.Validator
}

// NewHandler builds the RBAC management handler.
func NewHandler(svc *Service, validate pkg.Validator) *Handler {
	return &Handler{svc: svc, validate: validate}
}

// RegisterRoutes mounts the RBAC management endpoints.
func (h *Handler) RegisterRoutes(v1 fiber.Router, opts RouteOptions) {
	roles := v1.Group("/roles", opts.Auth)
	roles.Get("/", opts.RequireRolesView, h.ListRoles)
	roles.Post("/", opts.RequireRolesManage, h.CreateRole)
	roles.Get("/:id", opts.RequireRolesView, h.GetRole)
	roles.Patch("/:id", opts.RequireRolesManage, h.UpdateRole)
	roles.Delete("/:id", opts.RequireRolesManage, h.DeleteRole)
	roles.Get("/:id/permissions", opts.RequireRolesView, h.GetRolePermissions)
	roles.Put("/:id/permissions", opts.RequireRolesManage, h.SetRolePermissions)
	roles.Post("/:id/users", opts.RequireRolesManage, h.AssignRole)
	roles.Delete("/:id/users/:userId", opts.RequireRolesManage, h.UnassignRole)

	permissions := v1.Group("/permissions", opts.Auth)
	permissions.Get("/", opts.RequirePermissionsView, h.ListPermissions)
	permissions.Post("/", opts.RequirePermissionsManage, h.CreatePermission)
	permissions.Get("/:id", opts.RequirePermissionsView, h.GetPermission)
	permissions.Patch("/:id", opts.RequirePermissionsManage, h.UpdatePermission)
	permissions.Delete("/:id", opts.RequirePermissionsManage, h.DeletePermission)

	users := v1.Group("/users", opts.Auth)
	users.Get("/:id/roles", opts.RequireRolesView, h.GetUserRoles)
	users.Get("/:id/permissions", opts.RequireRolesView, h.GetUserPermissions)
}

// ListRoles handles GET /roles.
func (h *Handler) ListRoles(c *fiber.Ctx) error {
	query := parseListQuery(c)
	roles, meta, err := h.svc.ListRoles(c.Context(), query)
	if err != nil {
		return pkg.Error(c, err)
	}
	items := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		items = append(items, toRoleResponse(r))
	}
	return pkg.OK(c, ListRolesResponse{Items: items, Meta: meta})
}

// CreateRole handles POST /roles.
func (h *Handler) CreateRole(c *fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}
	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}
	role, err := h.svc.CreateRole(c.Context(), req.Name, req.Description)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Created(c, toRoleResponse(*role))
}

// GetRole handles GET /roles/:id.
func (h *Handler) GetRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid role id")
	}
	role, err := h.svc.GetRole(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, toRoleResponse(*role))
}

// UpdateRole handles PATCH /roles/:id.
func (h *Handler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid role id")
	}
	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}
	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}
	role, err := h.svc.UpdateRole(c.Context(), id, req.Name, req.Description)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, toRoleResponse(*role))
}

// DeleteRole handles DELETE /roles/:id.
func (h *Handler) DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid role id")
	}
	if err := h.svc.DeleteRole(c.Context(), id); err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Message(c, fiber.StatusOK, "role deleted")
}

// GetRolePermissions handles GET /roles/:id/permissions.
func (h *Handler) GetRolePermissions(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid role id")
	}
	role, err := h.svc.GetRole(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	perms, err := h.svc.GetRolePermissions(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	items := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		items = append(items, toPermissionResponse(p))
	}
	return pkg.OK(c, RolePermissionsResponse{Role: toRoleResponse(*role), Permissions: items})
}

// SetRolePermissions handles PUT /roles/:id/permissions.
func (h *Handler) SetRolePermissions(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid role id")
	}
	var req SetRolePermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}
	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}
	if err := h.svc.SetRolePermissions(c.Context(), id, req.PermissionIDs); err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Message(c, fiber.StatusOK, "role permissions updated")
}

// AssignRole handles POST /roles/:id/users.
func (h *Handler) AssignRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid role id")
	}
	var req AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		return pkg.BadRequestResponse(c, "invalid user id")
	}
	if err := h.svc.AssignRole(c.Context(), req.UserID, id); err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Message(c, fiber.StatusOK, "role assigned")
}

// UnassignRole handles DELETE /roles/:id/users/:userId.
func (h *Handler) UnassignRole(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Params("userId")
	if !validID(id) || !validID(userID) {
		return pkg.BadRequestResponse(c, "invalid role or user id")
	}
	if err := h.svc.UnassignRole(c.Context(), userID, id); err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Message(c, fiber.StatusOK, "role removed from user")
}

// ListPermissions handles GET /permissions.
func (h *Handler) ListPermissions(c *fiber.Ctx) error {
	query := parseListQuery(c)
	perms, meta, err := h.svc.ListPermissions(c.Context(), query)
	if err != nil {
		return pkg.Error(c, err)
	}
	items := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		items = append(items, toPermissionResponse(p))
	}
	return pkg.OK(c, ListPermissionsResponse{Items: items, Meta: meta})
}

// CreatePermission handles POST /permissions.
func (h *Handler) CreatePermission(c *fiber.Ctx) error {
	var req CreatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}
	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}
	perm, err := h.svc.CreatePermission(c.Context(), req.Name, req.Description)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Created(c, toPermissionResponse(*perm))
}

// GetPermission handles GET /permissions/:id.
func (h *Handler) GetPermission(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid permission id")
	}
	perm, err := h.svc.GetPermission(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, toPermissionResponse(*perm))
}

// UpdatePermission handles PATCH /permissions/:id.
func (h *Handler) UpdatePermission(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid permission id")
	}
	var req UpdatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return pkg.BadRequestResponse(c, "invalid request body")
	}
	if fields, err := h.validate.ValidateStruct(&req); err != nil {
		return pkg.ValidationResponse(c, "validation error", fields)
	}
	perm, err := h.svc.UpdatePermission(c.Context(), id, req.Name, req.Description)
	if err != nil {
		return pkg.Error(c, err)
	}
	return pkg.OK(c, toPermissionResponse(*perm))
}

// DeletePermission handles DELETE /permissions/:id.
func (h *Handler) DeletePermission(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid permission id")
	}
	if err := h.svc.DeletePermission(c.Context(), id); err != nil {
		return pkg.Error(c, err)
	}
	return pkg.Message(c, fiber.StatusOK, "permission deleted")
}

// GetUserRoles handles GET /users/:id/roles.
func (h *Handler) GetUserRoles(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid user id")
	}
	roles, err := h.svc.GetUserRoles(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	items := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		items = append(items, toRoleResponse(r))
	}
	return pkg.OK(c, UserRolesResponse{UserID: id, Roles: items})
}

// GetUserPermissions handles GET /users/:id/permissions.
func (h *Handler) GetUserPermissions(c *fiber.Ctx) error {
	id := c.Params("id")
	if !validID(id) {
		return pkg.BadRequestResponse(c, "invalid user id")
	}
	perms, err := h.svc.GetUserPermissionsDetail(c.Context(), id)
	if err != nil {
		return pkg.Error(c, err)
	}
	items := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		items = append(items, toPermissionResponse(p))
	}
	return pkg.OK(c, ListPermissionsResponse{Items: items, Meta: PaginationMeta{Page: 1, PerPage: len(items), TotalItems: int64(len(items)), TotalPages: 1}})
}

func parseListQuery(c *fiber.Ctx) ListQuery {
	return ListQuery{
		Page:    queryInt(c, "page", 1),
		PerPage: queryInt(c, "per_page", 20),
		Search:  c.Query("search"),
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
