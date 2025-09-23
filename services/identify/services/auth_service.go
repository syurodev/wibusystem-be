package services

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/services/interfaces"
	"wibusystem/services/identify/session"
)

// AuthService implements authentication-related business logic
type AuthService struct {
	repos         *repositories.Repositories
	userService   interfaces.UserServiceInterface
	credService   interfaces.CredentialServiceInterface
	tenantService interfaces.TenantServiceInterface
	sessionMgr    *session.Manager
}

// NewAuthService creates a new auth service
func NewAuthService(
	repos *repositories.Repositories,
	userService interfaces.UserServiceInterface,
	credService interfaces.CredentialServiceInterface,
	tenantService interfaces.TenantServiceInterface,
	sessionMgr *session.Manager,
) interfaces.AuthServiceInterface {
	return &AuthService{
		repos:         repos,
		userService:   userService,
		credService:   credService,
		tenantService: tenantService,
		sessionMgr:    sessionMgr,
	}
}

// Register creates a new user account with optional password
func (s *AuthService) Register(ctx context.Context, req d.CreateUserRequest) (*m.User, error) {
	// Create user through user service
	user, err := s.userService.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create password credential if provided
	if req.Password != "" {
		_, err := s.credService.CreatePasswordCredential(ctx, user.ID, req.Email, req.Password)
		if err != nil {
			// Rollback user creation if credential creation fails
			_ = s.userService.DeleteUser(ctx, user.ID)
			return nil, fmt.Errorf("failed to create password credential: %w", err)
		}
	}

	return user, nil
}

// Login authenticates user with email/password and returns user info with tenants
func (s *AuthService) Login(ctx context.Context, req d.LoginRequest) (*interfaces.LoginResult, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	// Get user by email
	user, err := s.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Get password credential
	credential, err := s.credService.GetPasswordCredential(ctx, user.ID)
	if err != nil || credential.SecretHash == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := s.credService.VerifyPassword(*credential.SecretHash, req.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login timestamps
	_ = s.userService.UpdateLastLogin(ctx, user.ID)
	_ = s.repos.Credential.UpdateLastUsed(ctx, credential.ID)

	// Get user tenants
	tenants, err := s.tenantService.GetUserTenants(ctx, user.ID)
	if err != nil {
		// Don't fail login if tenant retrieval fails, just return empty slice
		tenants = []*m.Tenant{}
	}

	return &interfaces.LoginResult{
		User:    user,
		Tenants: tenants,
	}, nil
}

// ChangePassword updates user's password after verifying current password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	if currentPassword == "" || newPassword == "" {
		return fmt.Errorf("current password and new password are required")
	}

	// Get current password credential
	credential, err := s.credService.GetPasswordCredential(ctx, userID)
	if err != nil || credential.SecretHash == nil {
		return fmt.Errorf("no password is currently set for this user")
	}

	// Verify current password
	if err := s.credService.VerifyPassword(*credential.SecretHash, currentPassword); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password
	if err := s.credService.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("invalid new password: %w", err)
	}

	// Update password credential
	if err := s.credService.UpdatePasswordCredential(ctx, credential.ID, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// GetProfile retrieves user profile with associated tenants
func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*interfaces.ProfileResult, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	// Get user
	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Get user tenants
	tenants, err := s.tenantService.GetUserTenants(ctx, userID)
	if err != nil {
		// Don't fail if tenant retrieval fails, just return empty slice
		tenants = []*m.Tenant{}
	}

	return &interfaces.ProfileResult{
		User:    user,
		Tenants: tenants,
	}, nil
}

// SetUserSession creates a session for the authenticated user
func (s *AuthService) SetUserSession(c *gin.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	if err := s.sessionMgr.Set(c, userID.String()); err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}

	return nil
}

// ClearUserSession clears the user session
func (s *AuthService) ClearUserSession(c *gin.Context) {
	s.sessionMgr.Clear(c)
}

// RefreshToken refreshes access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*interfaces.TokenResult, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	// TODO: Implement actual token refresh logic with OAuth2 provider
	// This is a placeholder implementation
	// In a real implementation, you would:
	// 1. Validate the refresh token
	// 2. Check if it's not expired or revoked
	// 3. Generate new access token
	// 4. Optionally rotate refresh token

	return &interfaces.TokenResult{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}