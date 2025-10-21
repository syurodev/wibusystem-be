// Package service contains business logic and use cases for the Identity module.
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"

	"wibusystem/internal/modules/identity/domain"
	"wibusystem/internal/modules/identity/dto"
	"wibusystem/internal/modules/identity/repository"
)

// authServiceImpl is the implementation of AuthService.
type authServiceImpl struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	// TODO: Add email service for sending verification/reset emails
	// emailService EmailService
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
) AuthService {
	return &authServiceImpl{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// Register registers a new user account.
func (s *authServiceImpl) Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error) {
	// Validate email format
	if err := domain.ValidateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Create new user with validated data
	user, err := domain.NewUser(req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set display name if provided
	if req.DisplayName != nil && *req.DisplayName != "" {
		if err := user.UpdateDisplayName(*req.DisplayName); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// TODO: Send verification email
	// if err := s.SendVerificationEmail(ctx, user.ID); err != nil {
	//     // Log error but don't fail registration
	//     log.Printf("Failed to send verification email: %v", err)
	// }

	return user, nil
}

// Login authenticates a user and creates a session.
func (s *authServiceImpl) Login(ctx context.Context, req dto.LoginRequest) (*domain.User, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return nil, "", ErrUserNotFound
	}

	// Verify password
	if !user.VerifyPassword(req.Password) {
		return nil, "", ErrInvalidCredentials
	}

	// Check user status
	switch user.Status {
	case domain.UserStatusSuspended:
		return nil, "", ErrUserSuspended
	case domain.UserStatusInactive:
		return nil, "", ErrUserNotActive
	}

	// Determine session duration based on "remember me"
	duration := domain.DefaultSessionDuration
	if req.RememberMe {
		duration = domain.ExtendedSessionDuration
	}

	// Create session
	session, sessionToken, err := domain.NewSession(user.ID, duration, nil, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	// Save session to database
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to save session: %w", err)
	}

	// Update last login timestamp
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
		// The session is already created, so login is successful
	}

	return user, sessionToken, nil
}

// Logout logs out a user by revoking their session.
func (s *authServiceImpl) Logout(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return ErrInvalidInput
	}

	// Hash the token to find the session
	// Note: We need to hash it the same way as when creating the session
	// For now, we'll use the token directly to find by token hash
	// In production, implement proper token hashing lookup

	// Get session by token
	session, err := s.getSessionByToken(ctx, sessionToken)
	if err != nil {
		if err == ErrSessionNotFound {
			// Session already doesn't exist, consider logout successful
			return nil
		}
		return err
	}

	// Revoke the session
	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

// RefreshSession refreshes an existing session.
func (s *authServiceImpl) RefreshSession(ctx context.Context, sessionToken string) (string, error) {
	// Validate current session
	session, err := s.ValidateSession(ctx, sessionToken)
	if err != nil {
		return "", err
	}

	// Check if session should be refreshed
	if !session.ShouldRefresh() {
		// Session is still good, return the same token
		return sessionToken, nil
	}

	// Extend the session expiration
	newExpiration := time.Now().UTC().Add(domain.DefaultSessionDuration)
	if err := s.sessionRepo.ExtendExpiration(ctx, session.ID, newExpiration); err != nil {
		return "", fmt.Errorf("failed to extend session: %w", err)
	}

	// In a more secure implementation, we would:
	// 1. Generate a new token
	// 2. Create a new session with the new token
	// 3. Revoke the old session
	// For now, we just extend the existing session and return the same token

	return sessionToken, nil
}

// VerifyEmail verifies a user's email address using a verification token.
func (s *authServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement token verification
	// 1. Decode and validate token
	// 2. Extract user ID from token
	// 3. Check if token is expired
	// 4. Mark user's email as verified

	// For now, return not implemented
	return fmt.Errorf("email verification not yet implemented")
}

// SendVerificationEmail sends an email verification link to the user.
func (s *authServiceImpl) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement email sending
	// 1. Get user by ID
	// 2. Generate verification token
	// 3. Send email with verification link

	// For now, return not implemented
	return fmt.Errorf("send verification email not yet implemented")
}

// RequestPasswordReset initiates a password reset flow.
func (s *authServiceImpl) RequestPasswordReset(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// For security, don't reveal if email exists
		// Return success even if user not found
		if err == repository.ErrUserNotFound {
			return nil
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Don't send reset email to deleted or suspended users
	if user.IsDeleted() || user.Status == domain.UserStatusSuspended {
		// Return success for security (don't reveal account status)
		return nil
	}

	// TODO: Generate reset token and send email
	// 1. Generate secure reset token with expiration
	// 2. Store token (in cache or database)
	// 3. Send email with reset link

	return nil
}

// ResetPassword resets a user's password using a reset token.
func (s *authServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) error {
	// TODO: Implement password reset
	// 1. Validate and decode token
	// 2. Check if token is expired
	// 3. Get user ID from token
	// 4. Update user's password
	// 5. Invalidate the reset token
	// 6. Optionally revoke all existing sessions

	return fmt.Errorf("password reset not yet implemented")
}

// ChangePassword changes a user's password (when they know the current password).
func (s *authServiceImpl) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if !user.VerifyPassword(currentPassword) {
		return ErrPasswordMismatch
	}

	// Validate and hash new password
	if err := user.ChangePassword(newPassword); err != nil {
		return fmt.Errorf("%w: %v", ErrWeakPassword, err)
	}

	// Update password in database
	if err := s.userRepo.UpdatePassword(ctx, userID, user.PasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Optionally revoke all other sessions (force re-login on all devices)
	// This is a security best practice when password changes
	// Commenting out for now as it might be too aggressive
	// _, _ = s.sessionRepo.RevokeAllByUser(ctx, userID)

	return nil
}

// ValidateSession validates a session token and returns the session.
func (s *authServiceImpl) ValidateSession(ctx context.Context, sessionToken string) (*domain.Session, error) {
	if sessionToken == "" {
		return nil, ErrInvalidSession
	}

	// Get session by token
	session, err := s.getSessionByToken(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	// Check if session is valid
	if !session.IsValid() {
		if session.Revoked {
			return nil, ErrSessionRevoked
		}
		if session.IsExpired() {
			return nil, ErrSessionExpired
		}
		return nil, ErrInvalidSession
	}

	// Update last accessed timestamp (async to avoid blocking)
	go func() {
		// Use background context to avoid cancellation
		bgCtx := context.Background()
		_ = s.sessionRepo.UpdateLastAccessed(bgCtx, session.ID)
	}()

	return session, nil
}

// ValidateSessionAndGetUser validates a session and returns both session and user.
func (s *authServiceImpl) ValidateSessionAndGetUser(ctx context.Context, sessionToken string) (*domain.Session, *domain.User, error) {
	// Validate session
	session, err := s.ValidateSession(ctx, sessionToken)
	if err != nil {
		return nil, nil, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, nil, ErrUserNotFound
		}
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if user.IsDeleted() {
		return nil, nil, ErrUserNotFound
	}
	if !user.IsActive() {
		return nil, nil, ErrUserNotActive
	}

	return session, user, nil
}

// GetUserFromSession retrieves the user associated with a session token.
func (s *authServiceImpl) GetUserFromSession(ctx context.Context, sessionToken string) (*domain.User, error) {
	_, user, err := s.ValidateSessionAndGetUser(ctx, sessionToken)
	return user, err
}

// RevokeSession revokes a specific session.
func (s *authServiceImpl) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessionRepo.Revoke(ctx, sessionID); err != nil {
		if err == repository.ErrSessionNotFound {
			return ErrSessionNotFound
		}
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user.
func (s *authServiceImpl) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := s.sessionRepo.RevokeAllByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to revoke sessions: %w", err)
	}
	return count, nil
}

// RevokeAllUserSessionsExcept revokes all sessions for a user except the specified one.
func (s *authServiceImpl) RevokeAllUserSessionsExcept(ctx context.Context, userID, exceptSessionID uuid.UUID) (int, error) {
	count, err := s.sessionRepo.RevokeAllExcept(ctx, userID, exceptSessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to revoke sessions: %w", err)
	}
	return count, nil
}

// UpdateSessionActivity updates the last accessed timestamp for a session.
func (s *authServiceImpl) UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessionRepo.UpdateLastAccessed(ctx, sessionID); err != nil {
		if err == repository.ErrSessionNotFound {
			return ErrSessionNotFound
		}
		return fmt.Errorf("failed to update session activity: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions from the database.
func (s *authServiceImpl) CleanupExpiredSessions(ctx context.Context) (int, error) {
	count, err := s.sessionRepo.DeleteExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return count, nil
}

// GetUserSessions retrieves all sessions for a user.
func (s *authServiceImpl) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	sessions, err := s.sessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	return sessions, nil
}

// GetUserActiveSessions retrieves only active sessions for a user.
func (s *authServiceImpl) GetUserActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	sessions, err := s.sessionRepo.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	return sessions, nil
}

// GetSession retrieves a single session by its ID.
func (s *authServiceImpl) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if err == repository.ErrSessionNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

// Helper methods
// getSessionByToken retrieves a session by its token.
// This is a helper method that handles token verification.
func (s *authServiceImpl) getSessionByToken(ctx context.Context, sessionToken string) (*domain.Session, error) {
	// In a production system, we would:
	// 1. Hash the token
	// 2. Look up session by token hash in database
	// 3. Verify the token against the stored hash

	// For now, we'll implement a simple approach:
	// We need to iterate through sessions or use a token->session mapping
	// This is NOT production-ready and needs optimization

	// TODO: Implement proper token-to-session lookup
	// Option 1: Store token hash in session and lookup by hash
	// Option 2: Use Redis for token->session_id mapping
	// Option 3: Encode session ID in the token (JWT-style)

	// Temporary implementation: This won't work with bcrypt hashed tokens
	// We need to rethink the token storage strategy
	return nil, fmt.Errorf("getSessionByToken not fully implemented - needs token lookup strategy")
}

// generateVerificationToken generates a secure verification token.
func generateVerificationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateResetToken generates a secure password reset token.
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
