// Package service contains business logic and use cases for the Identity module.
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"wibusystem/internal/modules/identity/domain"
	"wibusystem/internal/modules/identity/dto"
)

// AuthService defines the interface for authentication operations.
type AuthService interface {
	// Register registers a new user account.
	// Returns the created user and an error if registration fails.
	Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error)

	// Login authenticates a user and creates a session.
	// Returns the user, session token, and error if login fails.
	Login(ctx context.Context, req dto.LoginRequest) (*domain.User, string, error)

	// Logout logs out a user by revoking their session.
	// Returns an error if logout fails.
	Logout(ctx context.Context, sessionToken string) error

	// RefreshSession refreshes an existing session.
	// Returns a new session token and error if refresh fails.
	RefreshSession(ctx context.Context, sessionToken string) (string, error)

	// VerifyEmail verifies a user's email address using a verification token.
	// Returns an error if verification fails.
	VerifyEmail(ctx context.Context, token string) error

	// SendVerificationEmail sends an email verification link to the user.
	// Returns an error if sending fails.
	SendVerificationEmail(ctx context.Context, userID uuid.UUID) error

	// RequestPasswordReset initiates a password reset flow.
	// Sends a password reset email to the user.
	// Returns an error if the request fails.
	RequestPasswordReset(ctx context.Context, email string) error

	// ResetPassword resets a user's password using a reset token.
	// Returns an error if reset fails.
	ResetPassword(ctx context.Context, token, newPassword string) error

	// ChangePassword changes a user's password (when they know the current password).
	// Returns an error if change fails.
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error

	// ValidateSession validates a session token and returns the session.
	// Returns an error if validation fails (expired, revoked, or not found).
	ValidateSession(ctx context.Context, sessionToken string) (*domain.Session, error)

	// ValidateSessionAndGetUser validates a session and returns both session and user.
	// Returns an error if validation fails.
	ValidateSessionAndGetUser(ctx context.Context, sessionToken string) (*domain.Session, *domain.User, error)

	// GetUserFromSession retrieves the user associated with a session token.
	// Returns an error if the session is invalid or user not found.
	GetUserFromSession(ctx context.Context, sessionToken string) (*domain.User, error)

	// RevokeSession revokes a specific session.
	// Returns an error if revocation fails.
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error

	// RevokeAllUserSessions revokes all sessions for a user.
	// Returns the number of sessions revoked and an error if it fails.
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) (int, error)

	// RevokeAllUserSessionsExcept revokes all sessions for a user except the specified one.
	// Useful for "logout from all other devices" functionality.
	// Returns the number of sessions revoked and an error if it fails.
	RevokeAllUserSessionsExcept(ctx context.Context, userID, exceptSessionID uuid.UUID) (int, error)

	// UpdateSessionActivity updates the last accessed timestamp for a session.
	// Should be called on each authenticated request.
	// Returns an error if update fails.
	UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error

	// CleanupExpiredSessions removes expired sessions from the database.
	// Should be run periodically (e.g., daily cron job).
	// Returns the number of sessions cleaned up and an error if it fails.
	CleanupExpiredSessions(ctx context.Context) (int, error)

	// GetUserSessions retrieves all sessions for a user.
	// Returns a list of sessions and an error if retrieval fails.
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)

	// GetUserActiveSessions retrieves only active (not expired, not revoked) sessions for a user.
	// Returns a list of active sessions and an error if retrieval fails.
	GetUserActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)

	// GetSession retrieves a single session by its ID.
	GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
}

// Service errors
var (
	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrEmailAlreadyExists is returned when trying to register with an existing email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserNotActive is returned when a user account is not active.
	ErrUserNotActive = errors.New("user account is not active")

	// ErrUserSuspended is returned when a user account is suspended.
	ErrUserSuspended = errors.New("user account is suspended")

	// ErrEmailNotVerified is returned when attempting an action that requires verified email.
	ErrEmailNotVerified = errors.New("email not verified")

	// ErrInvalidVerificationToken is returned when email verification token is invalid.
	ErrInvalidVerificationToken = errors.New("invalid verification token")

	// ErrInvalidResetToken is returned when password reset token is invalid.
	ErrInvalidResetToken = errors.New("invalid reset token")

	// ErrResetTokenExpired is returned when password reset token has expired.
	ErrResetTokenExpired = errors.New("reset token has expired")

	// ErrSessionNotFound is returned when a session cannot be found.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired is returned when a session has expired.
	ErrSessionExpired = errors.New("session has expired")

	// ErrSessionRevoked is returned when a session has been revoked.
	ErrSessionRevoked = errors.New("session has been revoked")

	// ErrInvalidSession is returned when a session is invalid.
	ErrInvalidSession = errors.New("invalid session")

	// ErrPasswordMismatch is returned when the current password doesn't match.
	ErrPasswordMismatch = errors.New("current password is incorrect")

	// ErrWeakPassword is returned when a password doesn't meet strength requirements.
	ErrWeakPassword = errors.New("password does not meet strength requirements")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrOperationFailed is returned when an operation fails for unknown reasons.
	ErrOperationFailed = errors.New("operation failed")

	// ErrUnauthorized is returned when the user is not authorized to perform an action.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user doesn't have permission.
	ErrForbidden = errors.New("forbidden")
)

// VerificationToken represents an email verification token.
type VerificationToken struct {
	UserID    uuid.UUID
	Email     string
	Token     string
	ExpiresAt int64 // Unix timestamp
}

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	UserID    uuid.UUID
	Email     string
	Token     string
	ExpiresAt int64 // Unix timestamp
}
