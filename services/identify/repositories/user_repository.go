package repositories

import (
	"context"
	"fmt"
	"time"

	m "wibusystem/pkg/common/model"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines CRUD and listing operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *m.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*m.User, error)
	GetByEmail(ctx context.Context, email string) (*m.User, error)
	GetByUsername(ctx context.Context, username string) (*m.User, error)
	Update(ctx context.Context, user *m.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*m.User, int64, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a Postgres-backed UserRepository.
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

// Create inserts a new user and returns the assigned ID and timestamps.
func (r *userRepository) Create(ctx context.Context, user *m.User) error {
	query := `
		INSERT INTO users (email, username, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	var username, displayName *string
	if user.Username != "" {
		username = &user.Username
	}
	if user.DisplayName != "" {
		displayName = &user.DisplayName
	}

	err := r.pool.QueryRow(ctx, query, user.Email, username, displayName).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID returns a user by ID or an error if not found.
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*m.User, error) {
	query := `
		SELECT id, email, username, display_name, created_at, last_login_at
		FROM users
		WHERE id = $1
	`

	user := &m.User{}
	var username, displayName *string
	var lastLoginAt *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&username,
		&displayName,
		&user.CreatedAt,
		&lastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Handle nullable fields from the database (username, display_name, last_login_at).
	if username != nil {
		user.Username = *username
	}
	if displayName != nil {
		user.DisplayName = *displayName
	}
	if lastLoginAt != nil {
		user.LastLoginAt = lastLoginAt
	}

	return user, nil
}

// GetByEmail returns a user by email.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*m.User, error) {
	query := `
		SELECT id, email, username, display_name, created_at, last_login_at
		FROM users
		WHERE email = $1
	`

	user := &m.User{}
	var username, displayName *string
	var lastLoginAt *time.Time

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&username,
		&displayName,
		&user.CreatedAt,
		&lastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Handle nullable fields
	if username != nil {
		user.Username = *username
	}
	if displayName != nil {
		user.DisplayName = *displayName
	}
	if lastLoginAt != nil {
		user.LastLoginAt = lastLoginAt
	}

	return user, nil
}

// GetByUsername returns a user by username.
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*m.User, error) {
	query := `
		SELECT id, email, username, display_name, created_at, last_login_at
		FROM users
		WHERE username = $1
	`

	user := &m.User{}
	var usernameResult, displayName *string
	var lastLoginAt *time.Time

	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&usernameResult,
		&displayName,
		&user.CreatedAt,
		&lastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	// Handle nullable fields
	if usernameResult != nil {
		user.Username = *usernameResult
	}
	if displayName != nil {
		user.DisplayName = *displayName
	}
	if lastLoginAt != nil {
		user.LastLoginAt = lastLoginAt
	}

	return user, nil
}

// Update modifies username/display_name for a user by ID.
func (r *userRepository) Update(ctx context.Context, user *m.User) error {
	query := `
		UPDATE users
		SET username = $2, display_name = $3
		WHERE id = $1
	`

	var username, displayName *string
	if user.Username != "" {
		username = &user.Username
	}
	if user.DisplayName != "" {
		displayName = &user.DisplayName
	}

	result, err := r.pool.Exec(ctx, query, user.ID, username, displayName)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with ID %d not found", user.ID)
	}

	return nil
}

// Delete removes a user by ID.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with ID %d not found", id)
	}

	return nil
}

// List returns paginated users and the total user count.
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*m.User, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM users`
	var total int64
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get users with pagination
	query := `
		SELECT id, email, username, display_name, created_at, last_login_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*m.User
	for rows.Next() {
		user := &m.User{}
		var username, displayName *string
		var lastLoginAt *time.Time

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&username,
			&displayName,
			&user.CreatedAt,
			&lastLoginAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		// Handle nullable fields
		if username != nil {
			user.Username = *username
		}
		if displayName != nil {
			user.DisplayName = *displayName
		}
		if lastLoginAt != nil {
			user.LastLoginAt = lastLoginAt
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

// UpdateLastLogin sets the last_login_at timestamp to NOW() for the given user.
func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with ID %d not found", id)
	}

	return nil
}
