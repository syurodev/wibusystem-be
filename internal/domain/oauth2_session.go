package domain

import (
	"context"
	"time"
)

// OAuth2SessionRepository định nghĩa interface cho việc truy cập dữ liệu session OAuth2.
// Các session này bao gồm refresh token, và có thể mở rộng cho các loại khác.
type OAuth2SessionRepository interface {
	// CreateSession lưu một session mới vào database.
	CreateSession(ctx context.Context, signature string, requestID string, sessionType string, sessionData []byte, expiresAt time.Time, clientID string, subject string) error

	// GetSessionBySignature lấy một session từ database bằng signature của nó.
	GetSessionBySignature(ctx context.Context, signature string) ([]byte, error)

	// DeleteSessionBySignature xóa một session khỏi database.
	DeleteSessionBySignature(ctx context.Context, signature string) error

	// RevokeAllUserSessions đánh dấu inactive tất cả sessions của một user (theo subject_id).
	RevokeAllUserSessions(ctx context.Context, subjectID string) error

	// GetActiveSessionsBySubject lấy tất cả sessions đang active của một user.
	GetActiveSessionsBySubject(ctx context.Context, subjectID string) ([]string, error)
}
