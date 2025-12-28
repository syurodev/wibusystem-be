// ============================================================================
// Email Verification Repository (Ent Implementation)
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
	"system/internal/ent/generated/emailverification"
)

// emailVerificationEntRepository triển khai EmailVerificationRepository sử dụng Ent
type emailVerificationEntRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewEmailVerificationEntRepository tạo một instance mới của emailVerificationEntRepository
func NewEmailVerificationEntRepository(client *ent.Client, db *sql.DB) domain.EmailVerificationRepository {
	return &emailVerificationEntRepository{client: client, db: db}
}

// Create tạo một verification token mới.
func (r *emailVerificationEntRepository) Create(ctx context.Context, token *domain.EmailVerificationToken) error {
	_, err := database.GetClientFromContext(ctx, r.client).EmailVerification.Create().
		SetID(token.ID).
		SetUserID(token.UserID).
		SetToken(token.Token).
		SetExpiresAt(token.ExpiresAt).
		SetCreatedAt(token.CreatedAt).
		Save(ctx)
	return err
}

// GetByToken lấy token theo token string.
func (r *emailVerificationEntRepository) GetByToken(ctx context.Context, tokenStr string) (*domain.EmailVerificationToken, error) {
	ev, err := database.GetClientFromContext(ctx, r.client).EmailVerification.Query().
		Where(emailverification.TokenEQ(tokenStr)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.EmailVerificationToken{
		ID:        ev.ID,
		UserID:    ev.UserID,
		Token:     ev.Token,
		ExpiresAt: ev.ExpiresAt,
		UsedAt:    ev.UsedAt,
		CreatedAt: ev.CreatedAt,
	}, nil
}

// MarkAsUsed đánh dấu token đã được sử dụng.
func (r *emailVerificationEntRepository) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).EmailVerification.UpdateOneID(tokenID).
		SetUsedAt(now).
		Save(ctx)
	return err
}

// DeleteByUserID xóa tất cả tokens của một user.
func (r *emailVerificationEntRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).EmailVerification.Delete().
		Where(emailverification.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

// CleanupExpired xóa các tokens đã hết hạn.
// Note: Sử dụng PostgreSQL function cho hiệu quả
func (r *emailVerificationEntRepository) CleanupExpired(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT identify.cleanup_expired_verification_tokens()").Scan(&count)
	return count, err
}
