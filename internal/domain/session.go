package domain

import (
	"context"
	"time"
)

// UserSession đại diện cho một session của user sau khi đăng nhập
type UserSession struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	ClientOS  string    `json:"client_os"`
	Browser   string    `json:"browser"`
	Device    string    `json:"device"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}

// SessionRepository định nghĩa interface cho việc quản lý user sessions
type SessionRepository interface {
	// CreateSession tạo một session mới cho user với metadata
	CreateSession(ctx context.Context, session *UserSession, ttl time.Duration) error

	// GetSession lấy thông tin session
	GetSession(ctx context.Context, sessionID string) (*UserSession, error)

	// GetUserSessions lấy danh sách tất cả session của user
	GetUserSessions(ctx context.Context, userID string) ([]*UserSession, error)

	// DeleteSession xóa session (logout)
	DeleteSession(ctx context.Context, sessionID string) error

	// DeleteUserSessions xóa tất cả session của user (logout all)
	DeleteUserSessions(ctx context.Context, userID string) error

	// RefreshSession gia hạn thời gian sống của session
	RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) error
}
