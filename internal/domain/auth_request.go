package domain

import (
	"context"
	"time"

	"github.com/ory/fosite"
)

// AuthRequestRepository định nghĩa interface cho việc lưu trữ authorization requests tạm thời
// Các requests này cần được lưu trong quá trình user login và consent
type AuthRequestRepository interface {
	// SaveAuthRequest lưu authorization request vào storage với TTL
	SaveAuthRequest(ctx context.Context, requestID string, ar fosite.AuthorizeRequester, ttl time.Duration) error

	// GetAuthRequest lấy authorization request từ storage
	GetAuthRequest(ctx context.Context, requestID string) (fosite.AuthorizeRequester, error)

	// DeleteAuthRequest xóa authorization request khỏi storage
	DeleteAuthRequest(ctx context.Context, requestID string) error

	// SaveAuthRequestWithUserID lưu authorization request cùng với userID (sau khi login)
	SaveAuthRequestWithUserID(ctx context.Context, requestID string, ar fosite.AuthorizeRequester, userID string, ttl time.Duration) error

	// GetUserIDFromAuthRequest lấy userID từ stored authorization request
	GetUserIDFromAuthRequest(ctx context.Context, requestID string) (string, error)
}
