package domain

import (
	"context"
	"time"
)

// UserSession đại diện cho một session của user sau khi đăng nhập
type UserSession struct {
	SessionID string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionRepository định nghĩa interface cho việc quản lý user sessions
type SessionRepository interface {
	// CreateSession tạo một session mới cho user
	CreateSession(ctx context.Context, sessionID string, userID string, ttl time.Duration) error

	// GetSession lấy thông tin session và trả về userID
	GetSession(ctx context.Context, sessionID string) (string, error)

	// DeleteSession xóa session (logout)
	DeleteSession(ctx context.Context, sessionID string) error

	// RefreshSession gia hạn thời gian sống của session
	RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) error
}
