package repository

import (
	"context"
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
		       phone, status, created_at, updated_at, last_login_at
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
		       phone, status, created_at, updated_at, last_login_at
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

// Create tạo user mới trong database
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO identify.users (
			id, email, email_verified, password_hash, full_name,
			avatar_url, phone, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Phone,
		user.Status,
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
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Phone,
		user.Status,
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
