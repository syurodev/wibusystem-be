// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wibusystem/internal/modules/identity/domain"
	"wibusystem/internal/modules/identity/repository"
)

// UserRepository is the PostgreSQL implementation of repository.UserRepository.
type UserRepository struct {
	db     *pgxpool.Pool
	schema string
}

// NewUserRepository creates a new PostgreSQL user repository.
func NewUserRepository(db *pgxpool.Pool, schema string) *UserRepository {
	if schema == "" {
		schema = "identity"
	}
	return &UserRepository{
		db:     db,
		schema: schema,
	}
}

// Create creates a new user in the database.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return repository.WrapRepositoryError(err, "invalid user data", "INVALID_USER_DATA")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.users (
			id, email, email_verified, password_hash, display_name, avatar_url,
			status, last_login_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`, r.schema)

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.Status,
		user.LastLoginAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return repository.ErrUserAlreadyExists
		}
		return repository.WrapRepositoryError(err, "failed to create user", "CREATE_FAILED")
	}

	return nil
}

// GetByID retrieves a user by their ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := fmt.Sprintf(`
		SELECT id, email, email_verified, password_hash, display_name, avatar_url,
			   status, last_login_at, created_at, updated_at, deleted_at
		FROM %s.users
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerified,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Status,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get user by ID", "GET_FAILED")
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	query := fmt.Sprintf(`
		SELECT id, email, email_verified, password_hash, display_name, avatar_url,
			   status, last_login_at, created_at, updated_at, deleted_at
		FROM %s.users
		WHERE email = $1 AND deleted_at IS NULL
	`, r.schema)

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerified,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Status,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get user by email", "GET_FAILED")
	}

	return user, nil
}

// Update updates an existing user's information.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return repository.WrapRepositoryError(err, "invalid user data", "INVALID_USER_DATA")
	}

	user.UpdatedAt = time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET email = $2,
			email_verified = $3,
			password_hash = $4,
			display_name = $5,
			avatar_url = $6,
			status = $7,
			last_login_at = $8,
			updated_at = $9
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.Status,
		user.LastLoginAt,
		user.UpdatedAt,
	)

	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update user", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// Delete soft-deletes a user by their ID.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET deleted_at = $2,
			status = $3,
			updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query, id, now, domain.UserStatusInactive, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to delete user", "DELETE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// HardDelete permanently deletes a user from the database.
func (r *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s.users WHERE id = $1`, r.schema)

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to hard delete user", "HARD_DELETE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// List retrieves a paginated list of users based on the provided filter.
func (r *UserRepository) List(ctx context.Context, filter repository.UserListFilter) ([]*domain.User, int, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []any{}
	argCount := 1

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *filter.Status)
		argCount++
	}

	if filter.EmailVerified != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("email_verified = $%d", argCount))
		args = append(args, *filter.EmailVerified)
		argCount++
	}

	if filter.EmailContains != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("email ILIKE $%d", argCount))
		args = append(args, "%"+filter.EmailContains+"%")
		argCount++
	}

	if filter.DisplayNameContains != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("display_name ILIKE $%d", argCount))
		args = append(args, "%"+filter.DisplayNameContains+"%")
		argCount++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.users %s`, r.schema, whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "failed to count users", "COUNT_FAILED")
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		order := "DESC"
		if filter.SortOrder == "asc" {
			order = "ASC"
		}
		switch filter.SortBy {
		case "created_at", "email", "display_name", "status":
			orderBy = fmt.Sprintf("%s %s", filter.SortBy, order)
		}
	}

	// Build LIMIT and OFFSET
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	// Query users
	query := fmt.Sprintf(`
		SELECT id, email, email_verified, password_hash, display_name, avatar_url,
			   status, last_login_at, created_at, updated_at, deleted_at
		FROM %s.users
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, r.schema, whereClause, orderBy, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "failed to list users", "LIST_FAILED")
	}
	defer rows.Close()

	users := []*domain.User{}
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.EmailVerified,
			&user.PasswordHash,
			&user.DisplayName,
			&user.AvatarURL,
			&user.Status,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, 0, repository.WrapRepositoryError(err, "failed to scan user", "SCAN_FAILED")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "error iterating users", "ITERATION_FAILED")
	}

	return users, total, nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	query := fmt.Sprintf(`
		SELECT EXISTS(SELECT 1 FROM %s.users WHERE email = $1 AND deleted_at IS NULL)
	`, r.schema)

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check email existence", "EXISTS_CHECK_FAILED")
	}

	return exists, nil
}

// ExistsByID checks if a user with the given ID exists.
func (r *UserRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`
		SELECT EXISTS(SELECT 1 FROM %s.users WHERE id = $1 AND deleted_at IS NULL)
	`, r.schema)

	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check ID existence", "EXISTS_CHECK_FAILED")
	}

	return exists, nil
}

// CountAll returns the total number of users.
func (r *UserRepository) CountAll(ctx context.Context, includeDeleted bool) (int, error) {
	whereClause := ""
	if !includeDeleted {
		whereClause = "WHERE deleted_at IS NULL"
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.users %s`, r.schema, whereClause)

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count users", "COUNT_FAILED")
	}

	return count, nil
}

// CountByStatus returns the number of users with a specific status.
func (r *UserRepository) CountByStatus(ctx context.Context, status domain.UserStatus) (int, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s.users WHERE status = $1 AND deleted_at IS NULL
	`, r.schema)

	var count int
	err := r.db.QueryRow(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count users by status", "COUNT_FAILED")
	}

	return count, nil
}

// UpdateLastLogin updates the user's last login timestamp.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET last_login_at = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query, id, now, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update last login", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// UpdatePassword updates the user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	now := time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query, id, passwordHash, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update password", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// VerifyEmail marks a user's email as verified.
func (r *UserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET email_verified = true,
			status = CASE WHEN status = 'pending_verification' THEN 'active' ELSE status END,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query, id, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to verify email", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// UpdateStatus updates a user's status.
func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	now := time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET status = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query, id, status, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update status", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// Search searches for users by email or display name.
func (r *UserRepository) Search(ctx context.Context, query string, limit int) ([]*domain.User, error) {
	if limit <= 0 {
		limit = 10
	}

	searchQuery := fmt.Sprintf(`
		SELECT id, email, email_verified, password_hash, display_name, avatar_url,
			   status, last_login_at, created_at, updated_at, deleted_at
		FROM %s.users
		WHERE deleted_at IS NULL
		  AND (email ILIKE $1 OR display_name ILIKE $1)
		ORDER BY
			CASE WHEN email = $2 THEN 0 ELSE 1 END,
			created_at DESC
		LIMIT $3
	`, r.schema)

	pattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, pattern, query, limit)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to search users", "SEARCH_FAILED")
	}
	defer rows.Close()

	users := []*domain.User{}
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.EmailVerified,
			&user.PasswordHash,
			&user.DisplayName,
			&user.AvatarURL,
			&user.Status,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan user", "SCAN_FAILED")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.WrapRepositoryError(err, "error iterating users", "ITERATION_FAILED")
	}

	return users, nil
}

// GetByIDs retrieves multiple users by their IDs.
func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.User, error) {
	if len(ids) == 0 {
		return []*domain.User{}, nil
	}

	query := fmt.Sprintf(`
		SELECT id, email, email_verified, password_hash, display_name, avatar_url,
			   status, last_login_at, created_at, updated_at, deleted_at
		FROM %s.users
		WHERE id = ANY($1) AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, r.schema)

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to get users by IDs", "GET_FAILED")
	}
	defer rows.Close()

	users := []*domain.User{}
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.EmailVerified,
			&user.PasswordHash,
			&user.DisplayName,
			&user.AvatarURL,
			&user.Status,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan user", "SCAN_FAILED")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.WrapRepositoryError(err, "error iterating users", "ITERATION_FAILED")
	}

	return users, nil
}
