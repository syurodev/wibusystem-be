package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// AttestationType định nghĩa loại attestation trong WebAuthn
type AttestationType string

const (
	AttestationTypeNone     AttestationType = "none"
	AttestationTypeIndirect AttestationType = "indirect"
	AttestationTypeDirect   AttestationType = "direct"
)

// SessionType định nghĩa loại WebAuthn session
type SessionType string

const (
	SessionTypeRegistration   SessionType = "registration"
	SessionTypeAuthentication SessionType = "authentication"
)

// WebAuthnCredential là domain model cho WebAuthn credential
type WebAuthnCredential struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	CredentialID   string // Base64URL encoded credential ID
	PublicKey      []byte // Public key trong binary format
	AttestationType AttestationType
	AAGUID         []byte // Authenticator AAGUID (16 bytes)
	SignCount      int32
	Transports     []string // usb, nfc, ble, internal, hybrid
	BackupEligible bool
	BackupState    bool
	CredentialName *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastUsedAt     *time.Time
}

// WebAuthnSession là domain model cho WebAuthn session (temporary)
type WebAuthnSession struct {
	ID          uuid.UUID
	UserID      *uuid.UUID // NULL for registration
	Challenge   string     // Base64URL encoded challenge
	SessionType SessionType
	UserAgent   *string
	IPAddress   *string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// WebAuthnCredentialRepository định nghĩa interface cho việc truy cập dữ liệu WebAuthn credentials
type WebAuthnCredentialRepository interface {
	// GetByID lấy credential theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*WebAuthnCredential, error)

	// GetByCredentialID lấy credential theo credential ID
	GetByCredentialID(ctx context.Context, credentialID string) (*WebAuthnCredential, error)

	// GetByUserID lấy tất cả credentials của một user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*WebAuthnCredential, error)

	// Create tạo credential mới
	Create(ctx context.Context, credential *WebAuthnCredential) error

	// Update cập nhật credential (sign count, last used, etc.)
	Update(ctx context.Context, credential *WebAuthnCredential) error

	// Delete xóa credential
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateSignCount cập nhật sign count và last used time
	UpdateSignCount(ctx context.Context, credentialID string, signCount int32) error
}

// WebAuthnSessionRepository định nghĩa interface cho việc truy cập dữ liệu WebAuthn sessions
type WebAuthnSessionRepository interface {
	// GetByChallenge lấy session theo challenge
	GetByChallenge(ctx context.Context, challenge string) (*WebAuthnSession, error)

	// Create tạo session mới
	Create(ctx context.Context, session *WebAuthnSession) error

	// Delete xóa session
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteExpired xóa các sessions đã hết hạn
	DeleteExpired(ctx context.Context) error
}
