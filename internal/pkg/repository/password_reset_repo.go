package repository

import (
	"context"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// passwordResetRepository triển khai PasswordResetRepository sử dụng pgx.
type passwordResetRepository struct {
	pool *pgxpool.Pool
}

// NewPasswordResetRepository tạo một instance mới của passwordResetRepository.
func NewPasswordResetRepository(pool *pgxpool.Pool) domain.PasswordResetRepository {
	return &passwordResetRepository{pool: pool}
}

// Create tạo một password reset token mới.
func (r *passwordResetRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO identify.password_reset_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)

	return err
}

// GetByToken lấy token theo token string.
func (r *passwordResetRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM identify.password_reset_tokens
		WHERE token = $1
	`

	rows, err := r.pool.Query(ctx, query, tokenStr)
	if err != nil {
		return nil, err
	}

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.PasswordResetToken])
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// MarkAsUsed đánh dấu token đã được sử dụng.
func (r *passwordResetRepository) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE identify.password_reset_tokens
		SET used_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, tokenID)
	return err
}

// DeleteByUserID xóa tất cả tokens của một user.
func (r *passwordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM identify.password_reset_tokens
		WHERE user_id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// CleanupExpired xóa các tokens đã hết hạn.
func (r *passwordResetRepository) CleanupExpired(ctx context.Context) (int, error) {
	query := `SELECT identify.cleanup_expired_password_reset_tokens()`

	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}
