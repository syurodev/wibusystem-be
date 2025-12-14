package user

import (
	"context"
	"encoding/json"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// userRepository triển khai UserRepository sử dụng pgx
type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository tạo một instance mới của userRepository
func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &userRepository{pool: pool}
}

// GetByID lấy user từ database theo ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, email_verified, password_hash, full_name, avatar_url,
		       phone, status, created_at, updated_at, last_login_at, settings,
		       display_name, username, bio, is_verified, follower_count, works_count, last_content_updated_at
		FROM identify.users
		WHERE id = $1 AND status != 'deleted'
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByEmail lấy user từ database theo email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, email_verified, password_hash, full_name, avatar_url,
		       phone, status, created_at, updated_at, last_login_at, settings,
		       display_name, username, bio, is_verified, follower_count, works_count, last_content_updated_at
		FROM identify.users
		WHERE email = $1 AND status != 'deleted'
	`

	rows, err := r.pool.Query(ctx, query, email)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByUsername lấy user từ database theo username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, email, email_verified, password_hash, full_name, avatar_url,
		       phone, status, created_at, updated_at, last_login_at, settings,
		       display_name, username, bio, is_verified, follower_count, works_count, last_content_updated_at
		FROM identify.users
		WHERE LOWER(username) = LOWER($1) AND status != 'deleted'
	`

	rows, err := r.pool.Query(ctx, query, username)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Create tạo user mới trong database
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO identify.users (
			id, email, email_verified, password_hash, full_name,
			avatar_url, phone, status, settings
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
	`

	settingsJSON, err := json.Marshal(user.Settings)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Phone,
		user.Status,
		string(settingsJSON),
	)

	return err
}

// Update cập nhật thông tin user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE identify.users
		SET email = $2,
		    email_verified = $3,
		    password_hash = $4,
		    full_name = $5,
		    avatar_url = $6,
		    phone = $7,
		    status = $8,
		    settings = $9::jsonb,
		    display_name = $10,
		    username = $11,
		    bio = $12::jsonb,
		    updated_at = NOW()
		WHERE id = $1
	`

	settingsJSON, err := json.Marshal(user.Settings)
	if err != nil {
		return err
	}

	bioJSON, err := json.Marshal(user.Bio)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Phone,
		user.Status,
		string(settingsJSON),
		user.DisplayName,
		user.Username,
		string(bioJSON),
	)

	return err
}

// UpdateLastLogin cập nhật thời gian đăng nhập cuối
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE identify.users
		SET last_login_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// GetGlobalPermissions lấy danh sách global permissions của user
func (r *userRepository) GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `SELECT permission_name FROM identify.get_user_global_permissions($1)`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	return permissions, rows.Err()
}

// GetOrganizationPermissions lấy danh sách permissions của user trong organization
func (r *userRepository) GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	query := `
		SELECT identify.get_user_organization_permissions($1, $2)
	`

	rows, err := r.pool.Query(ctx, query, userID, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var result struct {
			PermissionName string `db:"permission_name"`
		}
		err := rows.Scan(&result.PermissionName)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, result.PermissionName)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetGlobalRoles lấy danh sách global roles của user
func (r *userRepository) GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT r.name
		FROM identify.user_global_roles ugr
		JOIN identify.roles r ON ugr.role_id = r.id
		WHERE ugr.user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// GetOrganizationRoles lấy danh sách roles của user trong organization
func (r *userRepository) GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	query := `
		SELECT r.name
		FROM identify.user_organization_roles uor
		JOIN identify.roles r ON uor.role_id = r.id
		WHERE uor.user_id = $1 AND uor.organization_id = $2
	`

	rows, err := r.pool.Query(ctx, query, userID, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}
