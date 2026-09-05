package rbac

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const superAdminRole = "admin"

// Service answers role and permission lookups with cache-accelerated DB hits.
type Service struct {
	db    *sql.DB
	cache Cache
}

// NewService builds the RBAC service.
func NewService(db *sql.DB, cache Cache) *Service {
	return &Service{db: db, cache: cache}
}

// HasPermission reports whether a user holds the given permission.
func (s *Service) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	if permission == "" {
		return true, nil
	}

	perms := s.cache.GetPermissions(userID)
	if perms == nil {
		var err error
		perms, err = s.GetUserPermissions(ctx, userID)
		if err != nil {
			return false, err
		}
		s.cache.SetPermissions(userID, perms)
	}

	for _, p := range perms {
		if p == permission {
			return true, nil
		}
	}
	return false, nil
}

// HasRole reports whether a user holds the given role.
func (s *Service) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	roles := s.cache.GetRoles(userID)
	if roles == nil {
		var err error
		roles, err = s.getUserRoles(ctx, userID)
		if err != nil {
			return false, err
		}
		s.cache.SetRoles(userID, roles)
	}

	for _, r := range roles {
		if r == roleName {
			return true, nil
		}
	}
	return false, nil
}

// GetUserPermissions returns the distinct permission names for a user.
func (s *Service) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	if perms := s.cache.GetPermissions(userID); perms != nil {
		return perms, nil
	}

	query := `
		SELECT DISTINCT p.name
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		JOIN role_permissions rp ON rp.role_id = r.id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE u.id = $1
		ORDER BY p.name`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user permissions: %w", err)
	}
	defer rows.Close()

	perms := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate permissions: %w", err)
	}

	s.cache.SetPermissions(userID, perms)
	return perms, nil
}

func (s *Service) getUserRoles(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT r.name
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE u.id = $1
		ORDER BY r.name`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}
	defer rows.Close()

	roles := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, name)
	}
	return roles, rows.Err()
}

// IsSuperAdmin reports whether the user holds the admin role.
func (s *Service) IsSuperAdmin(ctx context.Context, userID string) (bool, error) {
	ok, _ := s.HasRole(ctx, userID, superAdminRole)
	// A super-admin inherits every permission without a DB round trip beyond
	// the role lookup; the admin role is the source of truth.
	return ok, nil
}

// HasAnyPermission is a convenience combining multiple permission checks.
func (s *Service) HasAnyPermission(ctx context.Context, userID string, permissions ...string) (bool, error) {
	perms := s.cache.GetPermissions(userID)
	if perms == nil {
		var err error
		perms, err = s.GetUserPermissions(ctx, userID)
		if err != nil {
			return false, err
		}
	}
	have := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		have[p] = struct{}{}
	}
	for _, p := range permissions {
		if _, ok := have[p]; ok {
			return true, nil
		}
	}
	return false, nil
}

// SanitizePermission normalizes permission input for logs/messages.
func SanitizePermission(p string) string {
	return strings.TrimSpace(p)
}
