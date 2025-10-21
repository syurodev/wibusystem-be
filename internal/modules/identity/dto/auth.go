// Package dto contains Data Transfer Objects for the Identity module.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	Password    string  `json:"password" binding:"required,min=8,max=72"`
	DisplayName *string `json:"display_name,omitempty" binding:"omitempty,max=255"`
}

// RegisterResponse represents a user registration response.
type RegisterResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me,omitempty"`
}

// LoginResponse represents a user login response.
type LoginResponse struct {
	User         UserResponse `json:"user"`
	SessionToken string       `json:"session_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Message      string       `json:"message"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	SessionToken string `json:"session_token,omitempty"`
}

// LogoutResponse represents a logout response.
type LogoutResponse struct {
	Message string `json:"message"`
}

// RefreshSessionRequest represents a session refresh request.
type RefreshSessionRequest struct {
	SessionToken string `json:"session_token" binding:"required"`
}

// RefreshSessionResponse represents a session refresh response.
type RefreshSessionResponse struct {
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Message      string    `json:"message"`
}

// VerifyEmailRequest represents an email verification request.
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// VerifyEmailResponse represents an email verification response.
type VerifyEmailResponse struct {
	Message string `json:"message"`
}

// ForgotPasswordRequest represents a forgot password request.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPasswordResponse represents a forgot password response.
type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// ResetPasswordRequest represents a password reset request.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// ResetPasswordResponse represents a password reset response.
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
}

// ChangePasswordResponse represents a password change response.
type ChangePasswordResponse struct {
	Message string `json:"message"`
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"email_verified"`
	DisplayName   *string    `json:"display_name,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Status        string     `json:"status"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// UpdateProfileRequest represents a profile update request.
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty" binding:"omitempty,max=255"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// UpdateProfileResponse represents a profile update response.
type UpdateProfileResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}

// GetProfileResponse represents a get profile response.
type GetProfileResponse struct {
	User UserResponse `json:"user"`
}

// DeleteAccountRequest represents an account deletion request.
type DeleteAccountRequest struct {
	Password        string `json:"password" binding:"required"`
	ConfirmDeletion bool   `json:"confirm_deletion" binding:"required"`
}

// DeleteAccountResponse represents an account deletion response.
type DeleteAccountResponse struct {
	Message string `json:"message"`
}

// ListSessionsResponse represents a list of user sessions.
type ListSessionsResponse struct {
	Sessions []SessionResponse `json:"sessions"`
	Total    int               `json:"total"`
}

// SessionResponse represents a session in API responses.
type SessionResponse struct {
	ID             uuid.UUID `json:"id"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
	Browser        string    `json:"browser,omitempty"`
	OS             string    `json:"os,omitempty"`
	Device         string    `json:"device,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsActive       bool      `json:"is_active"`
	IsCurrent      bool      `json:"is_current,omitempty"`
}

// RevokeSessionRequest represents a session revocation request.
type RevokeSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// RevokeSessionResponse represents a session revocation response.
type RevokeSessionResponse struct {
	Message string `json:"message"`
}

// RevokeAllSessionsResponse represents a response for revoking all sessions.
type RevokeAllSessionsResponse struct {
	RevokedCount int    `json:"revoked_count"`
	Message      string `json:"message"`
}
