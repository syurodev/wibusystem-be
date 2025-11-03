package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// EmailVerificationToken là domain model cho email verification token.
type EmailVerificationToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

// EmailVerificationRepository định nghĩa interface cho việc quản lý verification tokens.
type EmailVerificationRepository interface {
	// Create tạo một verification token mới
	Create(ctx context.Context, token *EmailVerificationToken) error

	// GetByToken lấy token theo token string
	GetByToken(ctx context.Context, token string) (*EmailVerificationToken, error)

	// MarkAsUsed đánh dấu token đã được sử dụng
	MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error

	// DeleteByUserID xóa tất cả tokens của một user (khi tạo token mới)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// CleanupExpired xóa các tokens đã hết hạn
	CleanupExpired(ctx context.Context) (int, error)
}
