package rbac

import "time"

type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListRolesResponse struct {
	Items []RoleResponse `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}

type RolePermissionsResponse struct {
	Role        RoleResponse         `json:"role"`
	Permissions []PermissionResponse `json:"permissions"`
}

type PermissionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListPermissionsResponse struct {
	Items []PermissionResponse `json:"items"`
	Meta  PaginationMeta       `json:"meta"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=50"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type UpdatePermissionRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

type SetRolePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" validate:"required"`
}

type AssignRoleRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type UserRolesResponse struct {
	UserID string         `json:"user_id"`
	Roles  []RoleResponse `json:"roles"`
}

func toRoleResponse(r Role) RoleResponse {
	return RoleResponse{ID: r.ID, Name: r.Name, Description: r.Description, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func toPermissionResponse(p Permission) PermissionResponse {
	return PermissionResponse{ID: p.ID, Name: p.Name, Description: p.Description, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
}
