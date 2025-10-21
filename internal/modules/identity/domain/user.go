// Package domain contains the core business entities and logic for the Identity module.
package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserStatus represents the status of a user account.
type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusInactive            UserStatus = "inactive"
	UserStatusSuspended           UserStatus = "suspended"
	UserStatusPendingVerification UserStatus = "pending_verification"
)

// Valid checks if the UserStatus is valid.
func (s UserStatus) Valid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended, UserStatusPendingVerification:
		return true
	default:
		return false
	}
}

// String returns the string representation of UserStatus.
func (s UserStatus) String() string {
	return string(s)
}

// Scan implements the sql.Scanner interface.
func (s *UserStatus) Scan(value any) error {
	if value == nil {
		*s = UserStatusPendingVerification
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan UserStatus: %v", value)
	}

	*s = UserStatus(str)
	if !s.Valid() {
		return fmt.Errorf("invalid UserStatus: %s", str)
	}

	return nil
}

// Value implements the driver.Valuer interface.
func (s UserStatus) Value() (driver.Value, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("invalid UserStatus: %s", s)
	}
	return string(s), nil
}

// User represents a user account in the system.
type User struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"email_verified"`
	PasswordHash  string     `json:"-"` // Never expose password hash in JSON
	DisplayName   *string    `json:"display_name,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Status        UserStatus `json:"status"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Domain errors
var (
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrEmailTooLong        = errors.New("email exceeds maximum length (255 characters)")
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong     = errors.New("password exceeds maximum length (72 characters)")
	ErrPasswordTooWeak     = errors.New("password must contain at least one uppercase, one lowercase, and one number")
	ErrInvalidPasswordHash = errors.New("invalid password hash")
	ErrInvalidUserStatus   = errors.New("invalid user status")
	ErrUserDeleted         = errors.New("user has been deleted")
	ErrUserNotActive       = errors.New("user is not active")
	ErrDisplayNameTooLong  = errors.New("display name exceeds maximum length (255 characters)")
	ErrInvalidUserID       = errors.New("invalid user ID")
)

// NewUser creates a new User with the given email and password.
// It validates the inputs and hashes the password.
func NewUser(email, password string) (*User, error) {
	// Validate email
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	// Validate and hash password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := &User{
		ID:            uuid.New(),
		Email:         strings.ToLower(strings.TrimSpace(email)),
		EmailVerified: false,
		PasswordHash:  passwordHash,
		Status:        UserStatusPendingVerification,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return user, nil
}

// ValidateEmail validates an email address.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return ErrInvalidEmail
	}

	if len(email) > 255 {
		return ErrEmailTooLong
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePassword validates a password according to security requirements.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	if len(password) > 72 {
		return ErrPasswordTooLong
	}

	hasUpper := false
	hasLower := false
	hasNumber := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return ErrPasswordTooWeak
	}

	return nil
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies a password against the stored hash.
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ChangePassword changes the user's password.
// It validates the new password and updates the hash.
func (u *User) ChangePassword(newPassword string) error {
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now().UTC()

	return nil
}

// VerifyEmail marks the user's email as verified.
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now().UTC()

	// Auto-activate user after email verification
	if u.Status == UserStatusPendingVerification {
		u.Status = UserStatusActive
	}
}

// Activate activates a user account.
func (u *User) Activate() error {
	if u.IsDeleted() {
		return ErrUserDeleted
	}

	u.Status = UserStatusActive
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// Deactivate deactivates a user account.
func (u *User) Deactivate() error {
	if u.IsDeleted() {
		return ErrUserDeleted
	}

	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// Suspend suspends a user account.
func (u *User) Suspend() error {
	if u.IsDeleted() {
		return ErrUserDeleted
	}

	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateLastLogin updates the last login timestamp.
func (u *User) UpdateLastLogin() {
	now := time.Now().UTC()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// UpdateDisplayName updates the user's display name.
func (u *User) UpdateDisplayName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		u.DisplayName = nil
		u.UpdatedAt = time.Now().UTC()
		return nil
	}

	if len(name) > 255 {
		return ErrDisplayNameTooLong
	}

	u.DisplayName = &name
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateAvatarURL updates the user's avatar URL.
func (u *User) UpdateAvatarURL(url string) {
	url = strings.TrimSpace(url)

	if url == "" {
		u.AvatarURL = nil
	} else {
		u.AvatarURL = &url
	}

	u.UpdatedAt = time.Now().UTC()
}

// Delete soft-deletes the user.
func (u *User) Delete() {
	now := time.Now().UTC()
	u.DeletedAt = &now
	u.Status = UserStatusInactive
	u.UpdatedAt = now
}

// IsDeleted returns true if the user is soft-deleted.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// IsActive returns true if the user is active and not deleted.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && !u.IsDeleted()
}

// CanLogin returns true if the user can login.
func (u *User) CanLogin() bool {
	return u.IsActive() && u.EmailVerified
}

// Validate validates the user entity.
func (u *User) Validate() error {
	if u.ID == uuid.Nil {
		return ErrInvalidUserID
	}

	if err := ValidateEmail(u.Email); err != nil {
		return err
	}

	if u.PasswordHash == "" {
		return ErrInvalidPasswordHash
	}

	if !u.Status.Valid() {
		return ErrInvalidUserStatus
	}

	if u.DisplayName != nil && len(*u.DisplayName) > 255 {
		return ErrDisplayNameTooLong
	}

	return nil
}

// GetDisplayNameOrEmail returns the display name if set, otherwise the email.
func (u *User) GetDisplayNameOrEmail() string {
	if u.DisplayName != nil && *u.DisplayName != "" {
		return *u.DisplayName
	}
	return u.Email
}

// Clone creates a deep copy of the User.
func (u *User) Clone() *User {
	clone := *u

	if u.DisplayName != nil {
		name := *u.DisplayName
		clone.DisplayName = &name
	}

	if u.AvatarURL != nil {
		url := *u.AvatarURL
		clone.AvatarURL = &url
	}

	if u.LastLoginAt != nil {
		timestamp := *u.LastLoginAt
		clone.LastLoginAt = &timestamp
	}

	if u.DeletedAt != nil {
		timestamp := *u.DeletedAt
		clone.DeletedAt = &timestamp
	}

	return &clone
}
