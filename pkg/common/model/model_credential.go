package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuthType represents the authentication type enum
type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeOAuth    AuthType = "oauth"
	AuthTypeOIDC     AuthType = "oidc"
	AuthTypeSAML     AuthType = "saml"
	AuthTypeWebAuthn AuthType = "webauthn"
	AuthTypeTOTP     AuthType = "totp"
	AuthTypePasskey  AuthType = "passkey"
)

// Suppress unused constant warnings - these are defined for future authentication methods
var (
	_ = AuthTypeOAuth
	_ = AuthTypeOIDC
	_ = AuthTypeSAML
	_ = AuthTypeWebAuthn
	_ = AuthTypeTOTP
	_ = AuthTypePasskey
)

// Value implements the driver.Valuer interface for database storage
func (at *AuthType) Value() (driver.Value, error) {
	return string(*at), nil
}

// Scan implements the sql.Scanner interface for database scanning
func (at *AuthType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch s := value.(type) {
	case string:
		*at = AuthType(s)
	case []byte:
		*at = AuthType(s)
	default:
		return fmt.Errorf("cannot scan %T into AuthType", value)
	}
	return nil
}

// Credential stores all authentication data and identify provider links
type Credential struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	Type              AuthType   `json:"type" db:"type"`
	Provider          *string    `json:"provider,omitempty" db:"provider"`
	Identifier        *string    `json:"identifier,omitempty" db:"identifier"`
	SecretHash        *string    `json:"-" db:"secret_hash"` // Never expose in JSON
	PublicKey         *string    `json:"public_key,omitempty" db:"public_key"`
	SignCount         *int       `json:"sign_count,omitempty" db:"sign_count"`
	AttestationAAGUID *string    `json:"attestation_aaguid,omitempty" db:"attestation_aaguid"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
}
