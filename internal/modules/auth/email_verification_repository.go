package auth

import (
	"context"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// emailVerificationRepository triển khai EmailVerificationRepository sử dụng pgx.
type emailVerificationRepository struct {
	pool *pgxpool.Pool
}

// NewEmailVerificationRepository tạo một instance mới của emailVerificationRepository.
func NewEmailVerificationRepository(pool *pgxpool.Pool) domain.EmailVerificationRepository {
	return &emailVerificationRepository{pool: pool}
}

// Create tạo một verification token mới.
func (r *emailVerificationRepository) Create(ctx context.Context, token *domain.EmailVerificationToken) error {
	query := `
		INSERT INTO identify.email_verification_tokens (id, user_id, token, expires_at, created_at)
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
func (r *emailVerificationRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM identify.email_verification_tokens
		WHERE token = $1
	`

	rows, err := r.pool.Query(ctx, query, tokenStr)
	if err != nil {
		return nil, err
	}

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.EmailVerificationToken])
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// MarkAsUsed đánh dấu token đã được sử dụng.
func (r *emailVerificationRepository) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE identify.email_verification_tokens
		SET used_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, tokenID)
	return err
}

// DeleteByUserID xóa tất cả tokens của một user.
func (r *emailVerificationRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM identify.email_verification_tokens
		WHERE user_id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// CleanupExpired xóa các tokens đã hết hạn.
func (r *emailVerificationRepository) CleanupExpired(ctx context.Context) (int, error) {
	query := `SELECT identify.cleanup_expired_verification_tokens()`

	var count int
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}
