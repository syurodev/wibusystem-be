package services

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	m "wibusystem/pkg/common/model"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/services/interfaces"
)

// CredentialService implements credential-related business logic
type CredentialService struct {
	repos      *repositories.Repositories
	bcryptCost int
}

// NewCredentialService creates a new credential service
func NewCredentialService(repos *repositories.Repositories, bcryptCost int) interfaces.CredentialServiceInterface {
	if bcryptCost < 10 || bcryptCost > 15 {
		bcryptCost = 12 // Default bcrypt cost
	}
	return &CredentialService{
		repos:      repos,
		bcryptCost: bcryptCost,
	}
}

// HashPassword generates a bcrypt hash from a plain text password
func (s *CredentialService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword compares a plain text password with a hashed password
func (s *CredentialService) VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" || password == "" {
		return fmt.Errorf("hashed password and password cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}

	return nil
}

// CreatePasswordCredential creates a new password credential for a user
func (s *CredentialService) CreatePasswordCredential(ctx context.Context, userID uuid.UUID, email, password string) (*m.Credential, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	if err := s.ValidatePassword(password); err != nil {
		return nil, err
	}

	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	credential := &m.Credential{
		UserID:     userID,
		Type:       m.AuthTypePassword,
		Identifier: &email,
		SecretHash: &hashedPassword,
	}

	if err := s.repos.Credential.Create(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to create password credential: %w", err)
	}

	return credential, nil
}

// UpdatePasswordCredential updates an existing password credential
func (s *CredentialService) UpdatePasswordCredential(ctx context.Context, credentialID uuid.UUID, newPassword string) error {
	if credentialID == uuid.Nil {
		return fmt.Errorf("credential ID cannot be nil")
	}

	if err := s.ValidatePassword(newPassword); err != nil {
		return err
	}

	credential, err := s.repos.Credential.GetByID(ctx, credentialID)
	if err != nil {
		return fmt.Errorf("failed to get credential: %w", err)
	}

	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	credential.SecretHash = &hashedPassword
	if err := s.repos.Credential.Update(ctx, credential); err != nil {
		return fmt.Errorf("failed to update password credential: %w", err)
	}

	return nil
}

// GetPasswordCredential retrieves password credential for a user
func (s *CredentialService) GetPasswordCredential(ctx context.Context, userID uuid.UUID) (*m.Credential, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	credential, err := s.repos.Credential.GetByUserIDAndType(ctx, userID, m.AuthTypePassword)
	if err != nil {
		return nil, fmt.Errorf("failed to get password credential: %w", err)
	}

	return credential, nil
}

// ValidatePassword checks if password meets security requirements
func (s *CredentialService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var missing []string
	if !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if !hasNumber {
		missing = append(missing, "number")
	}
	if !hasSpecial {
		missing = append(missing, "special character")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain at least one: %s", strings.Join(missing, ", "))
	}

	return nil
}