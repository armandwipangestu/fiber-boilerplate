package rbac

import (
	"context"
	"strings"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

const builtinAdminRole = "admin"

// PaginationMeta describes a page of a paginated list.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// --- Roles ---

func (s *Service) ListRoles(ctx context.Context, query ListQuery) ([]Role, PaginationMeta, error) {
	roles, total, err := s.admin.ListRoles(ctx, query)
	if err != nil {
		return nil, PaginationMeta{}, err
	}
	return roles, paginationMeta(query, total), nil
}

func (s *Service) GetRole(ctx context.Context, id string) (*Role, error) {
	return s.admin.GetRole(ctx, id)
}

func (s *Service) CreateRole(ctx context.Context, name, description string) (*Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, pkg.BadRequest("role name is required")
	}
	return s.admin.CreateRole(ctx, name, description)
}

func (s *Service) UpdateRole(ctx context.Context, id string, name, description *string) (*Role, error) {
	role, err := s.admin.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		name = cleanPtr(name)
		if *name == "" {
			return nil, pkg.BadRequest("role name is required")
		}
		role.Name = *name
	}
	if description != nil {
		role.Description = strings.TrimSpace(*description)
	}
	if role.Name == builtinAdminRole {
		return nil, pkg.BadRequest("the built-in admin role cannot be renamed")
	}
	updated, err := s.admin.UpdateRole(ctx, role.ID, role.Name, role.Description)
	if err != nil {
		return nil, err
	}
	s.cache.InvalidateAll()
	return updated, nil
}

func (s *Service) DeleteRole(ctx context.Context, id string) error {
	role, err := s.admin.GetRole(ctx, id)
	if err != nil {
		return err
	}
	if role.Name == builtinAdminRole {
		return pkg.BadRequest("the built-in admin role cannot be deleted")
	}
	if err := s.admin.DeleteRole(ctx, id); err != nil {
		return err
	}
	s.cache.InvalidateAll()
	return nil
}

func (s *Service) GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	return s.admin.GetRolePermissions(ctx, roleID)
}

func (s *Service) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	if _, err := s.admin.GetRole(ctx, roleID); err != nil {
		return err
	}
	if err := s.admin.SetRolePermissions(ctx, roleID, permissionIDs); err != nil {
		return err
	}
	s.cache.InvalidateAll()
	return nil
}

// --- Permissions ---

func (s *Service) ListPermissions(ctx context.Context, query ListQuery) ([]Permission, PaginationMeta, error) {
	perms, total, err := s.admin.ListPermissions(ctx, query)
	if err != nil {
		return nil, PaginationMeta{}, err
	}
	return perms, paginationMeta(query, total), nil
}

func (s *Service) GetPermission(ctx context.Context, id string) (*Permission, error) {
	return s.admin.GetPermission(ctx, id)
}

func (s *Service) CreatePermission(ctx context.Context, name, description string) (*Permission, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, pkg.BadRequest("permission name is required")
	}
	return s.admin.CreatePermission(ctx, name, description)
}

func (s *Service) UpdatePermission(ctx context.Context, id string, name, description *string) (*Permission, error) {
	perm, err := s.admin.GetPermission(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		name = cleanPtr(name)
		if *name == "" {
			return nil, pkg.BadRequest("permission name is required")
		}
		perm.Name = *name
	}
	if description != nil {
		perm.Description = strings.TrimSpace(*description)
	}
	updated, err := s.admin.UpdatePermission(ctx, perm.ID, perm.Name, perm.Description)
	if err != nil {
		return nil, err
	}
	s.cache.InvalidateAll()
	return updated, nil
}

func (s *Service) DeletePermission(ctx context.Context, id string) error {
	if err := s.admin.DeletePermission(ctx, id); err != nil {
		return err
	}
	s.cache.InvalidateAll()
	return nil
}

// --- User-role assignment ---

func (s *Service) AssignRole(ctx context.Context, userID, roleID string) error {
	exists, err := s.admin.UserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return pkg.NotFound("user not found")
	}
	if _, err := s.admin.GetRole(ctx, roleID); err != nil {
		return err
	}
	if err := s.admin.AssignRole(ctx, userID, roleID); err != nil {
		return err
	}
	s.cache.Invalidate(userID)
	return nil
}

func (s *Service) UnassignRole(ctx context.Context, userID, roleID string) error {
	if err := s.admin.UnassignRole(ctx, userID, roleID); err != nil {
		return err
	}
	s.cache.Invalidate(userID)
	return nil
}

func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]Role, error) {
	return s.admin.GetUserRoles(ctx, userID)
}

// GetUserPermissionsDetail returns the full permission records for a user.
func (s *Service) GetUserPermissionsDetail(ctx context.Context, userID string) ([]Permission, error) {
	return s.admin.GetUserPermissions(ctx, userID)
}

func cleanPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	return &v
}

func paginationMeta(query ListQuery, total int64) PaginationMeta {
	page := query.Page
	perPage := query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	pages := int(total / int64(perPage))
	if total%int64(perPage) != 0 {
		pages++
	}
	return PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: pages,
	}
}
