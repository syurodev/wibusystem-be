// ============================================================================
// Password Reset Repository (Ent Implementation)
// ============================================================================

package auth

import (
	"context"
	"database/sql"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/passwordreset"
)

// passwordResetEntRepository triển khai PasswordResetRepository sử dụng Ent
type passwordResetEntRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewPasswordResetEntRepository tạo một instance mới của passwordResetEntRepository
func NewPasswordResetEntRepository(client *ent.Client, db *sql.DB) domain.PasswordResetRepository {
	return &passwordResetEntRepository{client: client, db: db}
}

// Create tạo một password reset token mới.
func (r *passwordResetEntRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	_, err := database.GetClientFromContext(ctx, r.client).PasswordReset.Create().
		SetID(token.ID).
		SetUserID(token.UserID).
		SetToken(token.Token).
		SetExpiresAt(token.ExpiresAt).
		SetCreatedAt(token.CreatedAt).
		Save(ctx)
	return err
}

// GetByToken lấy token theo token string.
func (r *passwordResetEntRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.PasswordResetToken, error) {
	pr, err := database.GetClientFromContext(ctx, r.client).PasswordReset.Query().
		Where(passwordreset.TokenEQ(tokenStr)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.PasswordResetToken{
		ID:        pr.ID,
		UserID:    pr.UserID,
		Token:     pr.Token,
		ExpiresAt: pr.ExpiresAt,
		UsedAt:    pr.UsedAt,
		CreatedAt: pr.CreatedAt,
	}, nil
}

// MarkAsUsed đánh dấu token đã được sử dụng.
func (r *passwordResetEntRepository) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).PasswordReset.UpdateOneID(tokenID).
		SetUsedAt(now).
		Save(ctx)
	return err
}

// DeleteByUserID xóa tất cả tokens của một user.
func (r *passwordResetEntRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).PasswordReset.Delete().
		Where(passwordreset.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

// CleanupExpired xóa các tokens đã hết hạn.
// Note: Sử dụng PostgreSQL function cho hiệu quả
func (r *passwordResetEntRepository) CleanupExpired(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT identify.cleanup_expired_password_reset_tokens()").Scan(&count)
	return count, err
}
