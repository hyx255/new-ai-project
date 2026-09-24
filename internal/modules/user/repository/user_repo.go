package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"broadcast-platform/internal/modules/user/model"
	"broadcast-platform/internal/platform/database"
)

// UserRepository defines the interface for User persistence.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, page, pageSize int, includeDeleted bool) ([]*model.User, int, error)
	Update(ctx context.Context, user *model.User) error
	UpdatePassword(ctx context.Context, id string, passwordHash string, mustChangePassword bool, tokenVersion int) error
	UpdateStatus(ctx context.Context, id string, status model.UserStatus, tokenVersion int) error
	IncrementTokenVersion(ctx context.Context, id string) (int, error)
	SoftDelete(ctx context.Context, id string) error
}

type userRepository struct {
	db *database.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, username, password_hash, role, status, must_change_password, token_version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.PasswordHash, string(user.Role), string(user.Status),
		user.MustChangePassword, user.TokenVersion, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if isUniqueConstraintError(err) {
			return fmt.Errorf("create user: %w", ErrUsernameExists)
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	var roleStr, statusStr string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, status, must_change_password, token_version, deleted_at, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &roleStr, &statusStr,
		&user.MustChangePassword, &user.TokenVersion, &user.DeletedAt, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	user.Role = model.UserRole(roleStr)
	user.Status = model.UserStatus(statusStr)

	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	var roleStr, statusStr string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, status, must_change_password, token_version, deleted_at, created_at, updated_at
		 FROM users WHERE username = ?`, username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &roleStr, &statusStr,
		&user.MustChangePassword, &user.TokenVersion, &user.DeletedAt, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	user.Role = model.UserRole(roleStr)
	user.Status = model.UserStatus(statusStr)

	return user, nil
}

func (r *userRepository) List(ctx context.Context, page, pageSize int, includeDeleted bool) ([]*model.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Build query based on includeDeleted
	whereClause := ""
	if !includeDeleted {
		whereClause = "WHERE deleted_at IS NULL"
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// Query page
	query := fmt.Sprintf(
		`SELECT id, username, password_hash, role, status, must_change_password, token_version, deleted_at, created_at, updated_at
		 FROM users %s
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		var roleStr, statusStr string

		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &roleStr, &statusStr,
			&user.MustChangePassword, &user.TokenVersion, &user.DeletedAt, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}

		user.Role = model.UserRole(roleStr)
		user.Status = model.UserStatus(statusStr)
		users = append(users, user)
	}

	return users, total, rows.Err()
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET username = ?, role = ?, status = ?, must_change_password = ?, token_version = ?, updated_at = ?
		 WHERE id = ?`,
		user.Username, string(user.Role), string(user.Status), user.MustChangePassword, user.TokenVersion, user.UpdatedAt, user.ID,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("update user: %w", ErrUsernameExists)
		}
		return fmt.Errorf("update user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update user %s: %w", user.ID, ErrNotFound)
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id string, passwordHash string, mustChangePassword bool, tokenVersion int) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, must_change_password = ?, token_version = ?, updated_at = ? WHERE id = ?`,
		passwordHash, mustChangePassword, tokenVersion, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update password %s: %w", id, ErrNotFound)
	}
	return nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, id string, status model.UserStatus, tokenVersion int) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET status = ?, token_version = ?, updated_at = ? WHERE id = ?`,
		string(status), tokenVersion, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update status %s: %w", id, ErrNotFound)
	}
	return nil
}

func (r *userRepository) IncrementTokenVersion(ctx context.Context, id string) (int, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET token_version = token_version + 1, updated_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	if err != nil {
		return 0, fmt.Errorf("increment token version: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return 0, fmt.Errorf("increment token version %s: %w", id, ErrNotFound)
	}

	// Read new version
	var newVersion int
	if err := r.db.QueryRowContext(ctx, `SELECT token_version FROM users WHERE id = ?`, id).Scan(&newVersion); err != nil {
		return 0, fmt.Errorf("read new token version: %w", err)
	}

	return newVersion, nil
}

func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET deleted_at = ?, token_version = token_version + 1, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("soft delete user %s: %w", id, ErrNotFound)
	}
	return nil
}

// isUniqueConstraintError checks if an error is a unique constraint violation.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// SQLite
	if contains(errMsg, "UNIQUE constraint failed") {
		return true
	}
	// MySQL
	if contains(errMsg, "Duplicate entry") || contains(errMsg, "1062") {
		return true
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}