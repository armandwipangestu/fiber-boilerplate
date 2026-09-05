package rbac

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
	"github.com/jackc/pgx/v5/pgconn"
)

// Repository provides persistence for role & permission management. Lookups
// (HasPermission/HasRole) live in rbac.Service; this handles administration.
type Repository interface {
	CreateRole(ctx context.Context, name, description string) (*Role, error)
	ListRoles(ctx context.Context, query ListQuery) ([]Role, int64, error)
	GetRole(ctx context.Context, id string) (*Role, error)
	UpdateRole(ctx context.Context, id, name, description string) (*Role, error)
	DeleteRole(ctx context.Context, id string) error

	SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error)

	AssignRole(ctx context.Context, userID, roleID string) error
	UnassignRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]Role, error)
	GetUserPermissions(ctx context.Context, userID string) ([]Permission, error)
	UserExists(ctx context.Context, userID string) (bool, error)

	CreatePermission(ctx context.Context, name, description string) (*Permission, error)
	ListPermissions(ctx context.Context, query ListQuery) ([]Permission, int64, error)
	GetPermission(ctx context.Context, id string) (*Permission, error)
	UpdatePermission(ctx context.Context, id, name, description string) (*Permission, error)
	DeletePermission(ctx context.Context, id string) error
}

const roleColumns = "id, name, description, created_at, updated_at"
const permissionColumns = "id, name, description, created_at, updated_at"

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository builds the administration repository.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateRole(ctx context.Context, name, description string) (*Role, error) {
	query := `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING ` + roleColumns
	var role Role
	err := r.db.QueryRowContext(ctx, query, name, description).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if isUniqueViolation(err) {
		return nil, pkg.Conflict("role name already exists")
	}
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}
	return &role, nil
}

func (r *postgresRepository) ListRoles(ctx context.Context, query ListQuery) ([]Role, int64, error) {
	where, args := buildWhere("", query.Search, "name")

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM roles`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count roles: %w", err)
	}

	sqlQuery := `SELECT ` + roleColumns + ` FROM roles` + where +
		` ORDER BY name ASC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, query.PerPage, query.Offset())

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	roles := []Role{}
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, total, rows.Err()
}

func (r *postgresRepository) GetRole(ctx context.Context, id string) (*Role, error) {
	var role Role
	err := r.db.QueryRowContext(ctx, `SELECT `+roleColumns+` FROM roles WHERE id = $1`, id).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pkg.NotFound("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	return &role, nil
}

func (r *postgresRepository) UpdateRole(ctx context.Context, id, name, description string) (*Role, error) {
	query := `UPDATE roles SET name = $1, description = $2, updated_at = NOW()
	          WHERE id = $3 RETURNING ` + roleColumns
	var role Role
	err := r.db.QueryRowContext(ctx, query, name, description, id).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pkg.NotFound("role not found")
	}
	if isUniqueViolation(err) {
		return nil, pkg.Conflict("role name already exists")
	}
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}
	return &role, nil
}

func (r *postgresRepository) DeleteRole(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return pkg.NotFound("role not found")
	}
	return nil
}

func (r *postgresRepository) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, roleID).Scan(&exists); err != nil {
		return fmt.Errorf("check role: %w", err)
	}
	if !exists {
		return pkg.NotFound("role not found")
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
		return fmt.Errorf("clear role permissions: %w", err)
	}
	if len(permissionIDs) == 0 {
		return tx.Commit()
	}
	for _, permID := range permissionIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			roleID, permID); err != nil {
			return fmt.Errorf("assign permission: %w", err)
		}
	}
	return tx.Commit()
}

func (r *postgresRepository) GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	query := `SELECT p.id, p.name, p.description, p.created_at, p.updated_at
	          FROM permissions p
	          JOIN role_permissions rp ON rp.permission_id = p.id
	          WHERE rp.role_id = $1
	          ORDER BY p.name`
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}
	defer rows.Close()

	perms := []Permission{}
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role permissions: %w", err)
	}
	return perms, nil
}

func (r *postgresRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, roleID)
	if err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

func (r *postgresRepository) UnassignRole(ctx context.Context, userID, roleID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`, userID, roleID)
	if err != nil {
		return fmt.Errorf("unassign role: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return pkg.NotFound("role assignment not found")
	}
	return nil
}

func (r *postgresRepository) GetUserRoles(ctx context.Context, userID string) ([]Role, error) {
	query := `SELECT r.id, r.name, r.description, r.created_at, r.updated_at
	          FROM roles r
	          JOIN user_roles ur ON ur.role_id = r.id
	          WHERE ur.user_id = $1
	          ORDER BY r.name`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}
	defer rows.Close()

	roles := []Role{}
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *postgresRepository) GetUserPermissions(ctx context.Context, userID string) ([]Permission, error) {
	query := `SELECT DISTINCT p.id, p.name, p.description, p.created_at, p.updated_at
	          FROM permissions p
	          JOIN role_permissions rp ON rp.permission_id = p.id
	          JOIN user_roles ur ON ur.role_id = rp.role_id
	          WHERE ur.user_id = $1
	          ORDER BY p.name`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user permissions: %w", err)
	}
	defer rows.Close()

	perms := []Permission{}
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *postgresRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check user: %w", err)
	}
	return exists, nil
}

func (r *postgresRepository) CreatePermission(ctx context.Context, name, description string) (*Permission, error) {
	query := `INSERT INTO permissions (name, description) VALUES ($1, $2) RETURNING ` + permissionColumns
	var p Permission
	err := r.db.QueryRowContext(ctx, query, name, description).
		Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if isUniqueViolation(err) {
		return nil, pkg.Conflict("permission name already exists")
	}
	if err != nil {
		return nil, fmt.Errorf("create permission: %w", err)
	}
	return &p, nil
}

func (r *postgresRepository) ListPermissions(ctx context.Context, query ListQuery) ([]Permission, int64, error) {
	where, args := buildWhere("", query.Search, "name")

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM permissions`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count permissions: %w", err)
	}

	sqlQuery := `SELECT ` + permissionColumns + ` FROM permissions` + where +
		` ORDER BY name ASC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, query.PerPage, query.Offset())

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list permissions: %w", err)
	}
	defer rows.Close()

	perms := []Permission{}
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, p)
	}
	return perms, total, rows.Err()
}

func (r *postgresRepository) GetPermission(ctx context.Context, id string) (*Permission, error) {
	var p Permission
	err := r.db.QueryRowContext(ctx, `SELECT `+permissionColumns+` FROM permissions WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pkg.NotFound("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get permission: %w", err)
	}
	return &p, nil
}

func (r *postgresRepository) UpdatePermission(ctx context.Context, id, name, description string) (*Permission, error) {
	query := `UPDATE permissions SET name = $1, description = $2, updated_at = NOW()
	          WHERE id = $3 RETURNING ` + permissionColumns
	var p Permission
	err := r.db.QueryRowContext(ctx, query, name, description, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pkg.NotFound("permission not found")
	}
	if isUniqueViolation(err) {
		return nil, pkg.Conflict("permission name already exists")
	}
	if err != nil {
		return nil, fmt.Errorf("update permission: %w", err)
	}
	return &p, nil
}

func (r *postgresRepository) DeletePermission(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM permissions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete permission: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return pkg.NotFound("permission not found")
	}
	return nil
}

// ListQuery carries the shared pagination + search parameters.
type ListQuery struct {
	Page    int
	PerPage int
	Search  string
}

// Offset returns the LIMIT offset for the current page.
func (q ListQuery) Offset() int {
	if q.Page < 1 {
		q.Page = 1
	}
	return (q.Page - 1) * q.PerPage
}

func buildWhere(base, search, column string) (string, []any) {
	if search == "" {
		return base, []any{}
	}
	return base + ` WHERE ` + column + ` ILIKE '%' || $1 || '%'`, []any{search}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
