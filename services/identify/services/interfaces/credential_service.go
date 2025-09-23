package interfaces

import (
	"context"

	"github.com/google/uuid"
	m "wibusystem/pkg/common/model"
)

// CredentialServiceInterface defines the contract for credential operations
type CredentialServiceInterface interface {
	// HashPassword generates a bcrypt hash from a plain text password
	HashPassword(password string) (string, error)

	// VerifyPassword compares a plain text password with a hashed password
	VerifyPassword(hashedPassword, password string) error

	// CreatePasswordCredential creates a new password credential for a user
	CreatePasswordCredential(ctx context.Context, userID uuid.UUID, email, password string) (*m.Credential, error)

	// UpdatePasswordCredential updates an existing password credential
	UpdatePasswordCredential(ctx context.Context, credentialID uuid.UUID, newPassword string) error

	// GetPasswordCredential retrieves password credential for a user
	GetPasswordCredential(ctx context.Context, userID uuid.UUID) (*m.Credential, error)

	// ValidatePassword checks if password meets security requirements
	ValidatePassword(password string) error
}