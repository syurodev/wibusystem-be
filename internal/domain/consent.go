package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// ConsentMethod định nghĩa phương thức consent
type ConsentMethod string

const (
	ConsentMethodExplicit   ConsentMethod = "explicit"   // User clicked "Allow"
	ConsentMethodImplicit   ConsentMethod = "implicit"   // Trusted first-party app
	ConsentMethodRemembered ConsentMethod = "remembered" // Previous consent
)

// OAuth2Consent đại diện cho quyền mà user đã cấp cho một OAuth2 client
type OAuth2Consent struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ClientID      uuid.UUID
	GrantedScopes []string
	Revoked       bool
	GrantedAt     time.Time
	RevokedAt     *time.Time
	LastUsedAt    *time.Time
	ExpiresAt     *time.Time
	ConsentMethod ConsentMethod
	IPAddress     *string
	UserAgent     *string
}

// ConsentRepository định nghĩa interface cho việc quản lý consents
type ConsentRepository interface {
	// GetActiveConsent lấy active consent của user cho một client
	GetActiveConsent(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (*OAuth2Consent, error)

	// CreateConsent tạo consent mới
	CreateConsent(ctx context.Context, consent *OAuth2Consent) error

	// UpdateConsent cập nhật consent (mainly for last_used_at)
	UpdateConsent(ctx context.Context, consent *OAuth2Consent) error

	// RevokeConsent thu hồi consent của user cho một client
	RevokeConsent(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) error

	// RevokeAllUserConsents thu hồi tất cả consents của một user
	RevokeAllUserConsents(ctx context.Context, userID uuid.UUID) (int, error)

	// GetUserConsents lấy tất cả consents của một user
	GetUserConsents(ctx context.Context, userID uuid.UUID, includeRevoked bool) ([]*OAuth2Consent, error)
}
