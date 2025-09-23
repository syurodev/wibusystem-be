package interfaces

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// LoginResult represents the result of a successful login
type LoginResult struct {
	User    *m.User     `json:"user"`
	Tenants []*m.Tenant `json:"tenants"`
}

// ProfileResult represents user profile with tenants
type ProfileResult struct {
	User    *m.User     `json:"user"`
	Tenants []*m.Tenant `json:"tenants"`
}

// AuthServiceInterface defines the contract for authentication operations
type AuthServiceInterface interface {
	// Register creates a new user account with optional password
	Register(ctx context.Context, req d.CreateUserRequest) (*m.User, error)

	// Login authenticates user with email/password and returns user info with tenants
	Login(ctx context.Context, req d.LoginRequest) (*LoginResult, error)

	// ChangePassword updates user's password after verifying current password
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error

	// GetProfile retrieves user profile with associated tenants
	GetProfile(ctx context.Context, userID uuid.UUID) (*ProfileResult, error)

	// SetUserSession creates a session for the authenticated user
	SetUserSession(c *gin.Context, userID uuid.UUID) error

	// ClearUserSession clears the user session
	ClearUserSession(c *gin.Context)

	// RefreshToken refreshes access token using refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResult, error)
}

// TokenResult represents token refresh response
type TokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}