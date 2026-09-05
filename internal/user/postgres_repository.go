package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// postgresRepository stores users in a PostgreSQL-compatible database.
type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository builds a Repository backed by *sql.DB.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

const userColumns = "id, email, name, password_hash, avatar_url, created_at, updated_at"

func (r *postgresRepository) Create(ctx context.Context, u *Domain) (*Domain, error) {
	query := `INSERT INTO users (email, name, password_hash)
	          VALUES ($1, $2, $3)
	          RETURNING ` + userColumns

	var created Domain
	var avatar sql.NullString
	err := r.db.QueryRowContext(ctx, query, u.Email, u.Name, u.PasswordHash).
		Scan(&created.ID, &created.Email, &created.Name, &created.PasswordHash, &avatar, &created.CreatedAt, &created.UpdatedAt)
	created.AvatarKey = avatar.String
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &created, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*Domain, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`

	var u Domain
	var avatar sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &avatar, &u.CreatedAt, &u.UpdatedAt)
	u.AvatarKey = avatar.String
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*Domain, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE email = $1`

	var u Domain
	var avatar sql.NullString
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &avatar, &u.CreatedAt, &u.UpdatedAt)
	u.AvatarKey = avatar.String
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *postgresRepository) List(ctx context.Context, query ListUsersQuery) ([]*Domain, int64, error) {
	where := ""
	args := []any{}
	where, args = buildSearch(where, args, query.Search)

	countQuery := `SELECT COUNT(*) FROM users` + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	} else if !validSortColumn(sortBy) {
		sortBy = "created_at"
	}
	sortDir := strings.ToUpper(query.SortDir)
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}

	listQuery := `SELECT ` + userColumns + ` FROM users` + where +
		fmt.Sprintf(" ORDER BY %s %s LIMIT %d OFFSET %d", sortBy, sortDir, query.PerPage, (query.Page-1)*query.PerPage)

	rows, err := r.db.QueryContext(ctx, listQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := []*Domain{}
	for rows.Next() {
		var u Domain
		var avatar sql.NullString
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &avatar, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		u.AvatarKey = avatar.String
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}

	return users, total, nil
}

func (r *postgresRepository) Update(ctx context.Context, u *Domain) (*Domain, error) {
	query := `UPDATE users
	          SET email = $1, name = $2, password_hash = $3, updated_at = NOW()
	          WHERE id = $4
	          RETURNING ` + userColumns

	var updated Domain
	var avatar sql.NullString
	err := r.db.QueryRowContext(ctx, query, u.Email, u.Name, u.PasswordHash, u.ID).
		Scan(&updated.ID, &updated.Email, &updated.Name, &updated.PasswordHash, &avatar, &updated.CreatedAt, &updated.UpdatedAt)
	updated.AvatarKey = avatar.String
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &updated, nil
}

func (r *postgresRepository) UpdateAvatar(ctx context.Context, id, key string) (*Domain, error) {
	query := `UPDATE users
	          SET avatar_url = $1, updated_at = NOW()
	          WHERE id = $2
	          RETURNING ` + userColumns

	var updated Domain
	var avatar sql.NullString
	err := r.db.QueryRowContext(ctx, query, nullString(key), id).
		Scan(&updated.ID, &updated.Email, &updated.Name, &updated.PasswordHash, &avatar, &updated.CreatedAt, &updated.UpdatedAt)
	updated.AvatarKey = avatar.String
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update user avatar: %w", err)
	}
	return &updated, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user rows: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *postgresRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("exists user by email: %w", err)
	}
	return exists, nil
}

func buildSearch(where string, args []any, search string) (string, []any) {
	if search == "" {
		return where, args
	}
	if where == "" {
		where = " WHERE "
	} else {
		where += " AND "
	}
	placeholder := fmt.Sprintf("$%d", len(args)+1)
	where += "(email ILIKE " + placeholder + " OR name ILIKE " + placeholder + ")"
	args = append(args, "%"+search+"%")
	return where, args
}

func validSortColumn(col string) bool {
	switch col {
	case "email", "name", "created_at", "updated_at":
		return true
	default:
		return false
	}
}
